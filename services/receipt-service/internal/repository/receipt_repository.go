package repository

import (
	"context"
	"expense-tracker/receipt-service/internal/model"
)

// ReceiptRepository defines the interface for receipt data operations
// This follows the repository pattern for clean architecture
type ReceiptRepository interface {
	// Create inserts a new receipt into the database
	Create(ctx context.Context, receipt *model.Receipt) error

	// FindByID finds a receipt by ID and user ID
	// This ensures users can only access their own receipts
	FindByID(ctx context.Context, id, userID string) (*model.Receipt, error)

	// FindByExpenseID finds all receipts for a specific expense and user
	// Returns receipts linked to the given expense_id
	FindByExpenseID(ctx context.Context, expenseID, userID string) ([]*model.Receipt, error)

	// FindByUserID finds all receipts for a user with optional filters and pagination
	// Filters: expense_id (optional)
	// Pagination: page, limit
	FindByUserID(ctx context.Context, userID string, filters *model.ListReceiptsRequest) ([]*model.Receipt, int, error)

	// Update updates an existing receipt
	// Verifies ownership through userID
	Update(ctx context.Context, receipt *model.Receipt) error

	// Delete soft deletes a receipt (sets deleted_at)
	// Verifies ownership through userID
	Delete(ctx context.Context, id, userID string) error
}
