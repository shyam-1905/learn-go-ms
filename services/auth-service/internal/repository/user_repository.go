package repository

import (
	"context"
	"expense-tracker/auth-service/internal/model"
)

// UserRepository defines the interface for user data operations
// In Go, interfaces are implicit - if a type has all the methods, it implements the interface!
// This is called "duck typing" - if it walks like a duck and quacks like a duck, it's a duck!
//
// Why use interfaces?
// 1. Makes testing easier (we can create mock implementations)
// 2. Allows swapping implementations (PostgreSQL, MySQL, MongoDB, etc.)
// 3. Follows clean architecture principles
type UserRepository interface {
	// Create inserts a new user into the database
	// ctx (context.Context) is used for cancellation and timeouts
	// Returns the created user or an error
	Create(ctx context.Context, user *model.User) error

	// FindByEmail finds a user by their email address
	// Returns the user and an error (error will be nil if user not found - we'll check)
	FindByEmail(ctx context.Context, email string) (*model.User, error)

	// FindByID finds a user by their ID
	FindByID(ctx context.Context, id string) (*model.User, error)
}
