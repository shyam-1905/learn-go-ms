package model

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the system
// In Go, structs are like classes - they group related data together
type User struct {
	// ID is a UUID (Universally Unique Identifier)
	// We use a pointer to uuid.UUID, but actually we'll use string for simplicity
	ID string `json:"id" db:"id"`

	// Email is the user's email address (unique identifier)
	// The `json:"email"` tag tells Go how to serialize this field to JSON
	// The `db:"email"` tag tells our database library which column to map to
	Email string `json:"email" db:"email"`

	// PasswordHash stores the bcrypt-hashed password
	// Notice: we NEVER store plain passwords!
	// The `json:"-"` means this field is excluded from JSON responses (security!)
	PasswordHash string `json:"-" db:"password_hash"`

	// Name is the user's display name
	Name string `json:"name" db:"name"`

	// CreatedAt tracks when the user was created
	// time.Time is Go's built-in time type
	CreatedAt time.Time `json:"created_at" db:"created_at"`

	// UpdatedAt tracks when the user was last updated
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// DeletedAt is for soft deletes (nullable, so we use *time.Time)
	// If nil, user is active. If set, user is deleted.
	DeletedAt *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// NewUser creates a new User with generated ID and timestamps
// This is a constructor function - a common Go pattern
func NewUser(email, passwordHash, name string) *User {
	// uuid.New() generates a new UUID
	// .String() converts it to a string format
	now := time.Now()
	return &User{
		ID:           uuid.New().String(),
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
		CreatedAt:    now,
		UpdatedAt:    now,
		DeletedAt:    nil, // nil means not deleted
	}
}
