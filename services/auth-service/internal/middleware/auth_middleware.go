package middleware

import (
	"context"
	"expense-tracker/auth-service/internal/service"
	"net/http"
	"strings"
)

// Middleware is a function that wraps an HTTP handler
// It can:
// 1. Execute code before the handler (e.g., validate token)
// 2. Execute code after the handler (e.g., log response)
// 3. Modify the request/response
// 4. Call the next handler or return early

// AuthMiddleware validates JWT tokens and attaches user info to request context
// This protects routes that require authentication
type AuthMiddleware struct {
	jwtService *service.JWTService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtService *service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth is a middleware function that validates JWT tokens
// It returns a function that wraps the actual handler
// This is a common Go pattern for middleware
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	// Return a new handler function
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return // Stop here, don't call next handler
		}

		// Parse "Bearer <token>"
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]

		// Validate the token
		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Attach user info to request context
		// Context is Go's way of passing request-scoped data
		// This allows handlers to access user info without parsing the token again
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		ctx = context.WithValue(ctx, "user_email", claims.Email)

		// Create a new request with the updated context
		r = r.WithContext(ctx)

		// Call the next handler (the actual route handler)
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
