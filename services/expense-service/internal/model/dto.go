package model

import "time"

// DTOs (Data Transfer Objects) - used for API requests/responses

// CreateExpenseRequest represents the data sent when creating an expense
type CreateExpenseRequest struct {
	// Amount must be a positive decimal number
	Amount string `json:"amount" binding:"required"`

	// Description of the expense
	Description string `json:"description" binding:"required"`

	// Category for the expense (e.g., "Food", "Transport", "Entertainment")
	Category string `json:"category" binding:"required"`

	// ExpenseDate is when the expense occurred (format: YYYY-MM-DD)
	ExpenseDate string `json:"expense_date" binding:"required"`
}

// UpdateExpenseRequest represents the data sent when updating an expense
// All fields are optional for partial updates
type UpdateExpenseRequest struct {
	Amount      *string `json:"amount,omitempty"`
	Description *string `json:"description,omitempty"`
	Category    *string `json:"category,omitempty"`
	ExpenseDate *string `json:"expense_date,omitempty"`
}

// ExpenseResponse is what we send back after creating/updating/getting an expense
type ExpenseResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Amount      string    `json:"amount"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	ExpenseDate time.Time `json:"expense_date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListExpensesRequest represents query parameters for listing expenses
// These come from URL query parameters, not JSON body
type ListExpensesRequest struct {
	// Category filter (optional)
	Category string

	// StartDate for date range filter (format: YYYY-MM-DD)
	StartDate string

	// EndDate for date range filter (format: YYYY-MM-DD)
	EndDate string

	// Page number for pagination (default: 1)
	Page int

	// Limit is items per page (default: 20, max: 100)
	Limit int
}

// ListExpensesResponse contains the list of expenses and pagination info
type ListExpensesResponse struct {
	Expenses []ExpenseResponse `json:"expenses"`
	Total    int               `json:"total"` // Total number of expenses (before pagination)
	Page     int               `json:"page"`  // Current page number
	Limit    int               `json:"limit"` // Items per page
	Pages    int               `json:"pages"` // Total number of pages
}

// ExpenseSummaryItem represents a category summary
type ExpenseSummaryItem struct {
	Category string `json:"category"`
	Total    string `json:"total"` // Total amount for this category
	Count    int    `json:"count"` // Number of expenses in this category
}

// ExpenseSummaryResponse contains expense summary grouped by category
type ExpenseSummaryResponse struct {
	StartDate  time.Time            `json:"start_date"`
	EndDate    time.Time            `json:"end_date"`
	Total      string               `json:"total"` // Grand total across all categories
	ByCategory []ExpenseSummaryItem `json:"by_category"`
}
