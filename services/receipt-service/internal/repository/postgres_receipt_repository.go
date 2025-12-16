package repository

import (
	"context"
	"database/sql"
	"expense-tracker/receipt-service/internal/model"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresReceiptRepository implements ReceiptRepository using PostgreSQL
type PostgresReceiptRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresReceiptRepository creates a new PostgreSQL repository
func NewPostgresReceiptRepository(pool *pgxpool.Pool) ReceiptRepository {
	return &PostgresReceiptRepository{
		pool: pool,
	}
}

// Create inserts a new receipt into the database
func (r *PostgresReceiptRepository) Create(ctx context.Context, receipt *model.Receipt) error {
	query := `
		INSERT INTO receipts (id, user_id, expense_id, file_name, file_key, file_url, file_size, mime_type, 
		                      merchant_name, receipt_date, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
	`

	_, err := r.pool.Exec(ctx, query,
		receipt.ID,
		receipt.UserID,
		receipt.ExpenseID,
		receipt.FileName,
		receipt.FileKey,
		receipt.FileURL,
		receipt.FileSize,
		receipt.MimeType,
		receipt.MerchantName,
		receipt.ReceiptDate,
		receipt.TotalAmount,
		receipt.CreatedAt,
		receipt.UpdatedAt,
	)

	return err
}

// FindByID finds a receipt by ID and user ID
// This ensures ownership - users can only access their own receipts
func (r *PostgresReceiptRepository) FindByID(ctx context.Context, id, userID string) (*model.Receipt, error) {
	query := `
		SELECT id, user_id, expense_id, file_name, file_key, file_url, file_size, mime_type,
		       merchant_name, receipt_date, total_amount, created_at, updated_at, deleted_at
		FROM receipts
		WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL
	`

	var receipt model.Receipt
	var expenseID sql.NullString
	var merchantName sql.NullString
	var receiptDate sql.NullTime
	var totalAmount sql.NullString
	var deletedAt sql.NullTime

	err := r.pool.QueryRow(ctx, query, id, userID).Scan(
		&receipt.ID,
		&receipt.UserID,
		&expenseID,
		&receipt.FileName,
		&receipt.FileKey,
		&receipt.FileURL,
		&receipt.FileSize,
		&receipt.MimeType,
		&merchantName,
		&receiptDate,
		&totalAmount,
		&receipt.CreatedAt,
		&receipt.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Receipt not found
		}
		return nil, err
	}

	// Convert nullable fields
	if expenseID.Valid {
		receipt.ExpenseID = &expenseID.String
	}
	if merchantName.Valid {
		receipt.MerchantName = &merchantName.String
	}
	if receiptDate.Valid {
		receipt.ReceiptDate = &receiptDate.Time
	}
	if totalAmount.Valid {
		receipt.TotalAmount = &totalAmount.String
	}
	if deletedAt.Valid {
		receipt.DeletedAt = &deletedAt.Time
	}

	return &receipt, nil
}

// FindByExpenseID finds all receipts for a specific expense and user
func (r *PostgresReceiptRepository) FindByExpenseID(ctx context.Context, expenseID, userID string) ([]*model.Receipt, error) {
	query := `
		SELECT id, user_id, expense_id, file_name, file_key, file_url, file_size, mime_type,
		       merchant_name, receipt_date, total_amount, created_at, updated_at, deleted_at
		FROM receipts
		WHERE expense_id = $1 AND user_id = $2 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, expenseID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var receipts []*model.Receipt
	for rows.Next() {
		var receipt model.Receipt
		var expenseIDVal sql.NullString
		var merchantName sql.NullString
		var receiptDate sql.NullTime
		var totalAmount sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&receipt.ID,
			&receipt.UserID,
			&expenseIDVal,
			&receipt.FileName,
			&receipt.FileKey,
			&receipt.FileURL,
			&receipt.FileSize,
			&receipt.MimeType,
			&merchantName,
			&receiptDate,
			&totalAmount,
			&receipt.CreatedAt,
			&receipt.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, err
		}

		// Convert nullable fields
		if expenseIDVal.Valid {
			receipt.ExpenseID = &expenseIDVal.String
		}
		if merchantName.Valid {
			receipt.MerchantName = &merchantName.String
		}
		if receiptDate.Valid {
			receipt.ReceiptDate = &receiptDate.Time
		}
		if totalAmount.Valid {
			receipt.TotalAmount = &totalAmount.String
		}
		if deletedAt.Valid {
			receipt.DeletedAt = &deletedAt.Time
		}

		receipts = append(receipts, &receipt)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return receipts, nil
}

// FindByUserID finds all receipts for a user with optional filters and pagination
func (r *PostgresReceiptRepository) FindByUserID(ctx context.Context, userID string, filters *model.ListReceiptsRequest) ([]*model.Receipt, int, error) {
	// Build WHERE clause dynamically based on filters
	whereClause := "user_id = $1 AND deleted_at IS NULL"
	args := []interface{}{userID}
	argIndex := 2

	// Add expense_id filter if provided
	if filters.ExpenseID != "" {
		whereClause += fmt.Sprintf(" AND expense_id = $%d", argIndex)
		args = append(args, filters.ExpenseID)
		argIndex++
	}

	// Get total count (for pagination)
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM receipts WHERE %s", whereClause)
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
		SELECT id, user_id, expense_id, file_name, file_key, file_url, file_size, mime_type,
		       merchant_name, receipt_date, total_amount, created_at, updated_at, deleted_at
		FROM receipts
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var receipts []*model.Receipt
	for rows.Next() {
		var receipt model.Receipt
		var expenseID sql.NullString
		var merchantName sql.NullString
		var receiptDate sql.NullTime
		var totalAmount sql.NullString
		var deletedAt sql.NullTime

		err := rows.Scan(
			&receipt.ID,
			&receipt.UserID,
			&expenseID,
			&receipt.FileName,
			&receipt.FileKey,
			&receipt.FileURL,
			&receipt.FileSize,
			&receipt.MimeType,
			&merchantName,
			&receiptDate,
			&totalAmount,
			&receipt.CreatedAt,
			&receipt.UpdatedAt,
			&deletedAt,
		)
		if err != nil {
			return nil, 0, err
		}

		// Convert nullable fields
		if expenseID.Valid {
			receipt.ExpenseID = &expenseID.String
		}
		if merchantName.Valid {
			receipt.MerchantName = &merchantName.String
		}
		if receiptDate.Valid {
			receipt.ReceiptDate = &receiptDate.Time
		}
		if totalAmount.Valid {
			receipt.TotalAmount = &totalAmount.String
		}
		if deletedAt.Valid {
			receipt.DeletedAt = &deletedAt.Time
		}

		receipts = append(receipts, &receipt)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, err
	}

	return receipts, total, nil
}

// Update updates an existing receipt
func (r *PostgresReceiptRepository) Update(ctx context.Context, receipt *model.Receipt) error {
	query := `
		UPDATE receipts
		SET expense_id = $1,
		    file_url = $2,
		    merchant_name = $3,
		    receipt_date = $4,
		    total_amount = $5,
		    updated_at = $6
		WHERE id = $7 AND user_id = $8 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		receipt.ExpenseID,
		receipt.FileURL,
		receipt.MerchantName,
		receipt.ReceiptDate,
		receipt.TotalAmount,
		receipt.UpdatedAt,
		receipt.ID,
		receipt.UserID,
	)

	if err != nil {
		return err
	}

	// Check if any row was updated
	if result.RowsAffected() == 0 {
		return fmt.Errorf("receipt not found or access denied")
	}

	return nil
}

// Delete soft deletes a receipt
func (r *PostgresReceiptRepository) Delete(ctx context.Context, id, userID string) error {
	query := `
		UPDATE receipts
		SET deleted_at = $1
		WHERE id = $2 AND user_id = $3 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, now, id, userID)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("receipt not found or access denied")
	}

	return nil
}
