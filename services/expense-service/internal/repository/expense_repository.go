package repository

import (
	"context"
	"expense-tracker/expense-service/internal/model"
)

// ExpenseRepository defines the interface for expense data operations
// This follows the repository pattern for clean architecture
type ExpenseRepository interface {
	// Create inserts a new expense into the database
	Create(ctx context.Context, expense *model.Expense) error

	// FindByID finds an expense by ID and user ID
	// This ensures users can only access their own expenses
	FindByID(ctx context.Context, id, userID string) (*model.Expense, error)

	// FindByUserID finds all expenses for a user with optional filters and pagination
	// Filters: category, startDate, endDate
	// Pagination: page, limit
	FindByUserID(ctx context.Context, userID string, filters *model.ListExpensesRequest) ([]*model.Expense, int, error)

	// Update updates an existing expense
	// Verifies ownership through userID
	Update(ctx context.Context, expense *model.Expense) error

	// Delete soft deletes an expense (sets deleted_at)
	// Verifies ownership through userID
	Delete(ctx context.Context, id, userID string) error

	// GetTotalByCategory gets expense totals grouped by category
	// Used for summary/aggregation queries
	GetTotalByCategory(ctx context.Context, userID string, startDate, endDate *string) ([]model.ExpenseSummaryItem, string, error)
}
