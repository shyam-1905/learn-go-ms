package model

import (
	"time"

	"github.com/google/uuid"
)

// Expense represents an expense entry in the system
// This is the core entity for expense tracking
type Expense struct {
	// ID is a UUID primary key
	ID string `json:"id" db:"id"`

	// UserID is the UUID of the user who owns this expense
	// This comes from the auth-service (users table)
	// We don't have a foreign key constraint since it's in a different database
	UserID string `json:"user_id" db:"user_id"`

	// Amount is the expense amount (using string for JSON, will parse to decimal)
	// DECIMAL(10,2) in database for precise currency handling
	Amount string `json:"amount" db:"amount"`

	// Description describes what the expense was for
	Description string `json:"description" db:"description"`

	// Category categorizes the expense (e.g., "Food", "Transport", "Entertainment")
	Category string `json:"category" db:"category"`

	// ExpenseDate is when the expense occurred
	// DATE type in database
	ExpenseDate time.Time `json:"expense_date" db:"expense_date"`

	// CreatedAt tracks when the expense was created
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt tracks when the expense was last updated
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// DeletedAt is for soft deletes (nullable)
	// If nil, expense is active. If set, expense is deleted.
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// NewExpense creates a new Expense with generated ID and timestamps
func NewExpense(userID, amount, description, category string, expenseDate time.Time) *Expense {
	now := time.Now()
	return &Expense{
		ID:          uuid.New().String(),
		UserID:      userID,
		Amount:      amount,
		Description: description,
		Category:    category,
		ExpenseDate: expenseDate,
		CreatedAt:   now,
		UpdatedAt:   now,
		DeletedAt:   nil,
	}
}
