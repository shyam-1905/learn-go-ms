package model

// DTOs (Data Transfer Objects) - these are used for API requests/responses
// They're separate from the User model to control what data is exposed

// RegisterRequest represents the data sent when a user registers
// Notice: no password_hash here - we only accept plain password from client
type RegisterRequest struct {
	// Email must be provided (we'll validate this)
	Email string `json:"email" binding:"required"`

	// Password is the plain text password (we'll hash it before storing)
	Password string `json:"password" binding:"required"`

	// Name is the user's display name
	Name string `json:"name" binding:"required"`
}

// LoginRequest represents the data sent when a user logs in
type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is what we send back after successful registration or login
// It includes the user info and a JWT token
type AuthResponse struct {
	// UserID is the unique identifier for the user
	UserID string `json:"user_id"`

	// Email of the authenticated user
	Email string `json:"email"`

	// Name of the user
	Name string `json:"name"`

	// Token is the JWT token that the client will use for authenticated requests
	Token string `json:"token"`
}

// ValidateResponse is what we send back when validating a token
type ValidateResponse struct {
	// Valid indicates if the token is valid
	Valid bool `json:"valid"`

	// UserID is extracted from the token if valid
	UserID string `json:"user_id,omitempty"`

	// Email is extracted from the token if valid
	Email string `json:"email,omitempty"`
}
