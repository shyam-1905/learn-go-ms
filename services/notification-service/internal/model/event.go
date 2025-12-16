package model

import "time"

// Event represents a notification event from other services
type Event struct {
	EventType string                 `json:"event_type"`
	UserID    string                 `json:"user_id"`
	UserEmail string                 `json:"user_email"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Event types constants
const (
	EventTypeExpenseCreated  = "expense.created"
	EventTypeExpenseUpdated  = "expense.updated"
	EventTypeReceiptUploaded = "receipt.uploaded"
	EventTypeReceiptLinked   = "receipt.linked"
	EventTypeUserRegistered  = "user.registered"
)

// ExpenseCreatedData represents data for expense.created event
type ExpenseCreatedData struct {
	ExpenseID   string `json:"expense_id"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	Category    string `json:"category"`
	ExpenseDate string `json:"expense_date"`
}

// ExpenseUpdatedData represents data for expense.updated event
type ExpenseUpdatedData struct {
	ExpenseID   string `json:"expense_id"`
	Amount      string `json:"amount"`
	Description string `json:"description"`
	Category    string `json:"category"`
	ExpenseDate string `json:"expense_date"`
}

// ReceiptUploadedData represents data for receipt.uploaded event
type ReceiptUploadedData struct {
	ReceiptID string `json:"receipt_id"`
	FileName  string `json:"file_name"`
	FileSize  int64  `json:"file_size"`
	MimeType  string `json:"mime_type"`
	ExpenseID string `json:"expense_id,omitempty"`
}

// ReceiptLinkedData represents data for receipt.linked event
type ReceiptLinkedData struct {
	ReceiptID string `json:"receipt_id"`
	ExpenseID string `json:"expense_id"`
	FileName  string `json:"file_name"`
}

// UserRegisteredData represents data for user.registered event
type UserRegisteredData struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Name   string `json:"name"`
}
