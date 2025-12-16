package repository

import (
	"context"
	"database/sql"
	"expense-tracker/auth-service/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepository implements UserRepository using PostgreSQL
// This struct holds the database connection pool
type PostgresUserRepository struct {
	// pool is the database connection pool
	// pgxpool.Pool manages multiple database connections efficiently
	// Connection pools reuse connections instead of creating new ones each time
	pool *pgxpool.Pool
}

// NewPostgresUserRepository creates a new PostgreSQL repository
// This is a constructor function - notice it returns the interface type!
// This is a Go best practice: return interfaces, not concrete types
func NewPostgresUserRepository(pool *pgxpool.Pool) UserRepository {
	return &PostgresUserRepository{
		pool: pool,
	}
}

// Create inserts a new user into the database
func (r *PostgresUserRepository) Create(ctx context.Context, user *model.User) error {
	// SQL query with placeholders ($1, $2, etc.) - this prevents SQL injection!
	// In Go, we use $1, $2 for PostgreSQL (MySQL uses ?)
	query := `
		INSERT INTO users (id, email, password_hash, name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	// Execute the query
	// ctx allows cancellation/timeout
	// Exec returns a result (we don't need it here) and an error
	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		// In Go, we return errors - no exceptions!
		// The caller will check if err != nil
		return err
	}

	return nil
}

// FindByEmail finds a user by email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	// SQL query to find user by email
	// WHERE deleted_at IS NULL means we only get active users (soft delete)
	query := `
		SELECT id, email, password_hash, name, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user model.User
	var deletedAt sql.NullTime // sql.NullTime handles nullable timestamps

	// QueryRow executes a query that returns at most one row
	// Scan copies the columns into the variables
	err := r.pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		// pgx.ErrNoRows is returned when no rows match
		if err == pgx.ErrNoRows {
			return nil, nil // User not found - return nil user, nil error
		}
		// Other database errors
		return nil, err
	}

	// Convert sql.NullTime to *time.Time
	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}

// FindByID finds a user by ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `
		SELECT id, email, password_hash, name, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user model.User
	var deletedAt sql.NullTime

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.CreatedAt,
		&user.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if deletedAt.Valid {
		user.DeletedAt = &deletedAt.Time
	}

	return &user, nil
}
