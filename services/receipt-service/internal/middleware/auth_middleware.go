package middleware

import (
	"context"
	"expense-tracker/receipt-service/internal/service"
	"net/http"
	"strings"
)

// AuthMiddleware validates JWT tokens by calling auth-service
// This centralizes authentication logic in auth-service
type AuthMiddleware struct {
	authClient *service.AuthClient
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(authClient *service.AuthClient) *AuthMiddleware {
	return &AuthMiddleware{
		authClient: authClient,
	}
}

// RequireAuth is a middleware function that validates JWT tokens
// It extracts the token from Authorization header, calls auth-service to validate it,
// and attaches user_id to the request context
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Parse "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate the token by calling auth-service
		userID, email, err := m.authClient.ValidateToken(r.Context(), token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Attach user info to request context
		// This allows handlers to access user_id without parsing the token again
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "user_email", email)

		// Create a new request with the updated context
		r = r.WithContext(ctx)

		// Call the next handler
		next(w, r)
	}
}

// GetUserID extracts the user ID from the request context
// This is a helper function for handlers to get user info
func GetUserID(ctx context.Context) string {
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		return ""
	}
	return userID
}

// GetUserEmail extracts the user email from the request context
func GetUserEmail(ctx context.Context) string {
	email, ok := ctx.Value("user_email").(string)
	if !ok {
		return ""
	}
	return email
}
