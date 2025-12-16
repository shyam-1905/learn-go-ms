package model

import (
	"time"

	"github.com/google/uuid"
)

// Receipt represents a receipt entry in the system
// Receipts are linked to expenses and stored in AWS S3
type Receipt struct {
	// ID is a UUID primary key
	ID string `json:"id" db:"id"`

	// UserID is the UUID of the user who owns this receipt
	// This comes from the auth-service (users table)
	UserID string `json:"user_id" db:"user_id"`

	// ExpenseID is the UUID of the associated expense (nullable)
	// Links to expense-service expenses table
	ExpenseID *string `json:"expense_id,omitempty" db:"expense_id"`

	// FileName is the original filename of the uploaded receipt
	FileName string `json:"file_name" db:"file_name"`

	// FileKey is the S3 object key (path in S3 bucket)
	FileKey string `json:"file_key" db:"file_key"`

	// FileURL is the S3 presigned URL or public URL
	// This is generated when needed and may expire
	FileURL string `json:"file_url" db:"file_url"`

	// FileSize is the file size in bytes
	FileSize int64 `json:"file_size" db:"file_size"`

	// MimeType is the MIME type of the file (e.g., "image/jpeg", "image/png", "application/pdf")
	MimeType string `json:"mime_type" db:"mime_type"`

	// MerchantName is the merchant/store name extracted from receipt (nullable)
	MerchantName *string `json:"merchant_name,omitempty" db:"merchant_name"`

	// ReceiptDate is the date from the receipt (nullable)
	ReceiptDate *time.Time `json:"receipt_date,omitempty" db:"receipt_date"`

	// TotalAmount is the total amount from the receipt (nullable)
	TotalAmount *string `json:"total_amount,omitempty" db:"total_amount"`

	// CreatedAt tracks when the receipt was created
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt tracks when the receipt was last updated
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// DeletedAt is for soft deletes (nullable)
	// If nil, receipt is active. If set, receipt is deleted.
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// NewReceipt creates a new Receipt with generated ID and timestamps
func NewReceipt(userID, fileName, fileKey, mimeType string, fileSize int64) *Receipt {
	now := time.Now()
	return &Receipt{
		ID:           uuid.New().String(),
		UserID:       userID,
		FileName:     fileName,
		FileKey:      fileKey,
		FileSize:     fileSize,
		MimeType:     mimeType,
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil,
		ExpenseID:    nil,
		MerchantName: nil,
		ReceiptDate:  nil,
		TotalAmount:  nil,
	}
}
