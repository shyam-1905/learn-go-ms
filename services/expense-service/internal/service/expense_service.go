package service

import (
	"context"
	"errors"
	"expense-tracker/expense-service/internal/model"
	"expense-tracker/expense-service/internal/repository"
	"log"
	"strconv"
	"time"
)

// ExpenseService handles expense business logic
type ExpenseService struct {
	expenseRepo    repository.ExpenseRepository
	eventPublisher *EventPublisher // Optional - can be nil if not configured
}

// NewExpenseService creates a new expense service
func NewExpenseService(expenseRepo repository.ExpenseRepository) *ExpenseService {
	return &ExpenseService{
		expenseRepo: expenseRepo,
	}
}

// SetEventPublisher sets the event publisher (optional)
func (s *ExpenseService) SetEventPublisher(publisher *EventPublisher) {
	s.eventPublisher = publisher
}

// CreateExpense creates a new expense for a user
func (s *ExpenseService) CreateExpense(ctx context.Context, userID string, req *model.CreateExpenseRequest) (*model.ExpenseResponse, error) {
	// Validate amount
	amount, err := strconv.ParseFloat(req.Amount, 64)
	if err != nil || amount <= 0 {
		return nil, errors.New("amount must be a positive number")
	}

	// Validate description
	if req.Description == "" {
		return nil, errors.New("description is required")
	}

	// Validate category
	if req.Category == "" {
		return nil, errors.New("category is required")
	}

	// Parse expense date
	expenseDate, err := time.Parse("2006-01-02", req.ExpenseDate)
	if err != nil {
		return nil, errors.New("expense_date must be in YYYY-MM-DD format")
	}

	// Validate date is not in the future (optional business rule)
	if expenseDate.After(time.Now()) {
		return nil, errors.New("expense_date cannot be in the future")
	}

	// Create expense
	expense := model.NewExpense(userID, req.Amount, req.Description, req.Category, expenseDate)

	// Save to database
	err = s.expenseRepo.Create(ctx, expense)
	if err != nil {
		return nil, err
	}

	// Publish event (non-blocking, async)
	if s.eventPublisher != nil {
		// Extract user email from context (set by auth middleware)
		userEmail := ""
		if emailVal := ctx.Value("user_email"); emailVal != nil {
			if email, ok := emailVal.(string); ok {
				userEmail = email
			}
		}

		event := &Event{
			EventType: "expense.created",
			UserID:    userID,
			UserEmail: userEmail,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"expense_id":   expense.ID,
				"amount":       expense.Amount,
				"description":  expense.Description,
				"category":     expense.Category,
				"expense_date": expense.ExpenseDate.Format("2006-01-02"),
			},
		}
		log.Printf("Publishing expense.created event for expense %s (user: %s, email: %s)", expense.ID, userID, userEmail)
		s.eventPublisher.PublishEventAsync(ctx, event)
	} else {
		log.Printf("WARNING: Event publisher not configured - expense.created event will not be published")
	}

	// Return response
	return &model.ExpenseResponse{
		ID:          expense.ID,
		UserID:      expense.UserID,
		Amount:      expense.Amount,
		Description: expense.Description,
		Category:    expense.Category,
		ExpenseDate: expense.ExpenseDate,
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	}, nil
}

// GetExpense retrieves a single expense by ID
// Verifies ownership (user can only access their own expenses)
func (s *ExpenseService) GetExpense(ctx context.Context, expenseID, userID string) (*model.ExpenseResponse, error) {
	expense, err := s.expenseRepo.FindByID(ctx, expenseID, userID)
	if err != nil {
		return nil, err
	}

	if expense == nil {
		return nil, errors.New("expense not found")
	}

	return &model.ExpenseResponse{
		ID:          expense.ID,
		UserID:      expense.UserID,
		Amount:      expense.Amount,
		Description: expense.Description,
		Category:    expense.Category,
		ExpenseDate: expense.ExpenseDate,
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	}, nil
}

// ListExpenses retrieves expenses for a user with optional filters and pagination
func (s *ExpenseService) ListExpenses(ctx context.Context, userID string, filters *model.ListExpensesRequest) (*model.ListExpensesResponse, error) {
	// Validate date format if provided
	if filters.StartDate != "" {
		_, err := time.Parse("2006-01-02", filters.StartDate)
		if err != nil {
			return nil, errors.New("start_date must be in YYYY-MM-DD format")
		}
	}

	if filters.EndDate != "" {
		_, err := time.Parse("2006-01-02", filters.EndDate)
		if err != nil {
			return nil, errors.New("end_date must be in YYYY-MM-DD format")
		}
	}

	// Validate date range
	if filters.StartDate != "" && filters.EndDate != "" {
		startDate, _ := time.Parse("2006-01-02", filters.StartDate)
		endDate, _ := time.Parse("2006-01-02", filters.EndDate)
		if startDate.After(endDate) {
			return nil, errors.New("start_date cannot be after end_date")
		}
	}

	// Get expenses from repository
	expenses, total, err := s.expenseRepo.FindByUserID(ctx, userID, filters)
	if err != nil {
		return nil, err
	}

	// Convert to response format
	expenseResponses := make([]model.ExpenseResponse, len(expenses))
	for i, exp := range expenses {
		expenseResponses[i] = model.ExpenseResponse{
			ID:          exp.ID,
			UserID:      exp.UserID,
			Amount:      exp.Amount,
			Description: exp.Description,
			Category:    exp.Category,
			ExpenseDate: exp.ExpenseDate,
			CreatedAt:   exp.CreatedAt,
			UpdatedAt:   exp.UpdatedAt,
		}
	}

	// Calculate pagination
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	page := filters.Page
	if page <= 0 {
		page = 1
	}

	pages := (total + limit - 1) / limit // Ceiling division

	return &model.ListExpensesResponse{
		Expenses: expenseResponses,
		Total:    total,
		Page:     page,
		Limit:    limit,
		Pages:    pages,
	}, nil
}

// UpdateExpense updates an existing expense
// Verifies ownership before updating
func (s *ExpenseService) UpdateExpense(ctx context.Context, expenseID, userID string, req *model.UpdateExpenseRequest) (*model.ExpenseResponse, error) {
	// Get existing expense (verifies ownership)
	expense, err := s.expenseRepo.FindByID(ctx, expenseID, userID)
	if err != nil {
		return nil, err
	}

	if expense == nil {
		return nil, errors.New("expense not found")
	}

	// Update fields if provided
	if req.Amount != nil {
		amount, err := strconv.ParseFloat(*req.Amount, 64)
		if err != nil || amount <= 0 {
			return nil, errors.New("amount must be a positive number")
		}
		expense.Amount = *req.Amount
	}

	if req.Description != nil {
		if *req.Description == "" {
			return nil, errors.New("description cannot be empty")
		}
		expense.Description = *req.Description
	}

	if req.Category != nil {
		if *req.Category == "" {
			return nil, errors.New("category cannot be empty")
		}
		expense.Category = *req.Category
	}

	if req.ExpenseDate != nil {
		expenseDate, err := time.Parse("2006-01-02", *req.ExpenseDate)
		if err != nil {
			return nil, errors.New("expense_date must be in YYYY-MM-DD format")
		}
		expense.ExpenseDate = expenseDate
	}

	// Update timestamp
	expense.UpdatedAt = time.Now()

	// Save to database
	err = s.expenseRepo.Update(ctx, expense)
	if err != nil {
		return nil, err
	}

	// Publish event (non-blocking, async)
	if s.eventPublisher != nil {
		// Extract user email from context (set by auth middleware)
		userEmail := ""
		if emailVal := ctx.Value("user_email"); emailVal != nil {
			if email, ok := emailVal.(string); ok {
				userEmail = email
			}
		}

		event := &Event{
			EventType: "expense.updated",
			UserID:    userID,
			UserEmail: userEmail,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"expense_id":   expense.ID,
				"amount":       expense.Amount,
				"description":  expense.Description,
				"category":     expense.Category,
				"expense_date": expense.ExpenseDate.Format("2006-01-02"),
			},
		}
		log.Printf("Publishing expense.updated event for expense %s (user: %s, email: %s)", expense.ID, userID, userEmail)
		s.eventPublisher.PublishEventAsync(ctx, event)
	} else {
		log.Printf("WARNING: Event publisher not configured - expense.updated event will not be published")
	}

	return &model.ExpenseResponse{
		ID:          expense.ID,
		UserID:      expense.UserID,
		Amount:      expense.Amount,
		Description: expense.Description,
		Category:    expense.Category,
		ExpenseDate: expense.ExpenseDate,
		CreatedAt:   expense.CreatedAt,
		UpdatedAt:   expense.UpdatedAt,
	}, nil
}

// DeleteExpense soft deletes an expense
// Verifies ownership before deleting
func (s *ExpenseService) DeleteExpense(ctx context.Context, expenseID, userID string) error {
	// Verify expense exists and belongs to user
	expense, err := s.expenseRepo.FindByID(ctx, expenseID, userID)
	if err != nil {
		return err
	}

	if expense == nil {
		return errors.New("expense not found")
	}

	// Soft delete
	return s.expenseRepo.Delete(ctx, expenseID, userID)
}

// GetExpenseSummary gets expense summary grouped by category
func (s *ExpenseService) GetExpenseSummary(ctx context.Context, userID string, startDate, endDate *string) (*model.ExpenseSummaryResponse, error) {
	// Validate date format if provided
	if startDate != nil && *startDate != "" {
		_, err := time.Parse("2006-01-02", *startDate)
		if err != nil {
			return nil, errors.New("start_date must be in YYYY-MM-DD format")
		}
	}

	if endDate != nil && *endDate != "" {
		_, err := time.Parse("2006-01-02", *endDate)
		if err != nil {
			return nil, errors.New("end_date must be in YYYY-MM-DD format")
		}
	}

	// Get summary from repository
	byCategory, total, err := s.expenseRepo.GetTotalByCategory(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Parse dates for response
	var start time.Time
	var end time.Time

	if startDate != nil && *startDate != "" {
		start, _ = time.Parse("2006-01-02", *startDate)
	} else {
		// Default to 30 days ago if not provided
		start = time.Now().AddDate(0, 0, -30)
	}

	if endDate != nil && *endDate != "" {
		end, _ = time.Parse("2006-01-02", *endDate)
	} else {
		end = time.Now()
	}

	return &model.ExpenseSummaryResponse{
		StartDate:  start,
		EndDate:    end,
		Total:      total,
		ByCategory: byCategory,
	}, nil
}
