package model

import "time"

// DTOs (Data Transfer Objects) - used for API requests/responses

// UploadReceiptRequest represents the data sent when uploading a receipt
// This is handled as multipart/form-data, not JSON
// Fields: file (required), expense_id (optional)
type UploadReceiptRequest struct {
	// File is the receipt file (image or PDF)
	// This comes from multipart form data
	File interface{} `json:"-"` // Not serialized, handled separately

	// ExpenseID is optional - can link receipt to expense during upload
	ExpenseID *string `json:"expense_id,omitempty"`
}

// LinkReceiptRequest represents the data sent when linking a receipt to an expense
type LinkReceiptRequest struct {
	// ExpenseID is the UUID of the expense to link to
	ExpenseID string `json:"expense_id" binding:"required"`
}

// ReceiptResponse is what we send back after creating/updating/getting a receipt
type ReceiptResponse struct {
	ID           string     `json:"id"`
	UserID       string     `json:"user_id"`
	ExpenseID    *string    `json:"expense_id,omitempty"`
	FileName     string     `json:"file_name"`
	FileURL      string     `json:"file_url"` // Presigned URL
	FileSize     int64      `json:"file_size"`
	MimeType     string     `json:"mime_type"`
	MerchantName *string    `json:"merchant_name,omitempty"`
	ReceiptDate  *time.Time `json:"receipt_date,omitempty"`
	TotalAmount  *string    `json:"total_amount,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// ListReceiptsRequest represents query parameters for listing receipts
// These come from URL query parameters, not JSON body
type ListReceiptsRequest struct {
	// ExpenseID filter (optional) - get receipts for a specific expense
	ExpenseID string

	// Page number for pagination (default: 1)
	Page int

	// Limit is items per page (default: 20, max: 100)
	Limit int
}

// ListReceiptsResponse contains the list of receipts and pagination info
type ListReceiptsResponse struct {
	Receipts []ReceiptResponse `json:"receipts"`
	Total    int               `json:"total"` // Total number of receipts (before pagination)
	Page     int               `json:"page"`  // Current page number
	Limit    int               `json:"limit"` // Items per page
	Pages    int               `json:"pages"` // Total number of pages
}
