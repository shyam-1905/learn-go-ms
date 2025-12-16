package repository

import (
	"context"
	"database/sql"
	"expense-tracker/expense-service/internal/model"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresExpenseRepository implements ExpenseRepository using PostgreSQL
type PostgresExpenseRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresExpenseRepository creates a new PostgreSQL repository
func NewPostgresExpenseRepository(pool *pgxpool.Pool) ExpenseRepository {
	return &PostgresExpenseRepository{
		pool: pool,
	}
}

// Create inserts a new expense into the database
func (r *PostgresExpenseRepository) Create(ctx context.Context, expense *model.Expense) error {
	query := `
		INSERT INTO expenses (id, user_id, amount, description, category, expense_date, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.pool.Exec(ctx, query,
		expense.ID,
		expense.UserID,
		expense.Amount,
		expense.Description,
		expense.Category,
		expense.ExpenseDate,
		expense.CreatedAt,
		expense.UpdatedAt,
	)

	return err
}

// FindByID finds an expense by ID and user ID
// This ensures ownership - users can only access their own expenses
func (r *PostgresExpenseRepository) FindByID(ctx context.Context, id, userID string) (*model.Expense, error) {
	query := `
		SELECT id, user_id, amount, description, category, expense_date, created_at, updated_at, deleted_at
		FROM expenses
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var expense model.Expense
	var deletedAt sql.NullTime

	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&expense.ID,
		&expense.UserID,
		&expense.Amount,
		&expense.Description,
		&expense.Category,
		&expense.ExpenseDate,
		&expense.CreatedAt,
		&expense.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Expense not found
		}
		return nil, err
	}

	if deletedAt.Valid {
		expense.DeletedAt = &deletedAt.Time
	}

	return &expense, nil
}

// FindByUserID finds all expenses for a user with optional filters and pagination
func (r *PostgresExpenseRepository) FindByUserID(ctx context.Context, userID string, filters *model.ListExpensesRequest) ([]*model.Expense, int, error) {
	// Build WHERE clause dynamically based on filters
	whereClause := "user_id = $1 AND deleted_at IS NULL"
	args := []interface{}{userID}
	argIndex := 2

	// Add category filter
	if filters.Category != "" {
		whereClause += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, filters.Category)
		argIndex++
	}

	// Add date range filters
	if filters.StartDate != "" {
		whereClause += fmt.Sprintf(" AND expense_date >= $%d", argIndex)
		args = append(args, filters.StartDate)
		argIndex++
	}

	if filters.EndDate != "" {
		whereClause += fmt.Sprintf(" AND expense_date <= $%d", argIndex)
		args = append(args, filters.EndDate)
		argIndex++
	}

	// Get total count (for pagination)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM expenses WHERE %s", whereClause)
	var total int
	err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Apply pagination
	limit := filters.Limit
	if limit <= 0 {
		limit = 20 // Default
	}
	if limit > 100 {
		limit = 100 // Max
	}

	page := filters.Page
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Build SELECT query with pagination
	query := fmt.Sprintf(`
		SELECT id, user_id, amount, description, category, expense_date, created_at, updated_at, deleted_at
		FROM expenses
		WHERE %s
		ORDER BY expense_date DESC, created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var expenses []*model.Expense
	for rows.Next() {
		var expense model.Expense
		var deletedAt sql.NullTime

		err := rows.Scan(
			&expense.ID,
			&expense.UserID,
			&expense.Amount,
			&expense.Description,
			&expense.Category,
			&expense.ExpenseDate,
			&expense.CreatedAt,
			&expense.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		if deletedAt.Valid {
			expense.DeletedAt = &deletedAt.Time
		}

		expenses = append(expenses, &expense)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return expenses, total, nil
}

// Update updates an existing expense
func (r *PostgresExpenseRepository) Update(ctx context.Context, expense *model.Expense) error {
	// Update only non-nil fields
	query := `
		UPDATE expenses
		SET amount = $1,
		    description = $2,
		    category = $3,
		    expense_date = $4,
		    updated_at = $5
		WHERE id = $6 AND user_id = $7 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		expense.Amount,
		expense.Description,
		expense.Category,
		expense.ExpenseDate,
		expense.UpdatedAt,
		expense.ID,
		expense.UserID,
	)

	if err != nil {
		return err
	}

	// Check if any row was updated
	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense not found or access denied")
	}

	return nil
}

// Delete soft deletes an expense
func (r *PostgresExpenseRepository) Delete(ctx context.Context, id, userID string) error {
	query := `
		UPDATE expenses
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, now, id, userID)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("expense not found or access denied")
	}

	return nil
}

// GetTotalByCategory gets expense totals grouped by category
func (r *PostgresExpenseRepository) GetTotalByCategory(ctx context.Context, userID string, startDate, endDate *string) ([]model.ExpenseSummaryItem, string, error) {
	whereClause := "user_id = $1 AND deleted_at IS NULL"
	args := []interface{}{userID}
	argIndex := 2

	if startDate != nil && *startDate != "" {
		whereClause += fmt.Sprintf(" AND expense_date >= $%d", argIndex)
		args = append(args, *startDate)
		argIndex++
	}

	if endDate != nil && *endDate != "" {
		whereClause += fmt.Sprintf(" AND expense_date <= $%d", argIndex)
		args = append(args, *endDate)
		argIndex++
	}

	// Get totals by category
	query := fmt.Sprintf(`
		SELECT category, SUM(amount) as total, COUNT(*) as count
		FROM expenses
		WHERE %s
		GROUP BY category
		ORDER BY total DESC
	`, whereClause)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var items []model.ExpenseSummaryItem
	var grandTotal float64

	for rows.Next() {
		var item model.ExpenseSummaryItem
		var total float64

		err := rows.Scan(&item.Category, &total, &item.Count)
		if err != nil {
			return nil, "", err
		}

		// Format total as string with 2 decimal places
		item.Total = fmt.Sprintf("%.2f", total)
		grandTotal += total

		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, "", err
	}

	grandTotalStr := fmt.Sprintf("%.2f", grandTotal)
	return items, grandTotalStr, nil
}
