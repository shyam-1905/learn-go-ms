package service

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles JWT token validation
// This service only validates tokens (doesn't generate them)
// Token generation is done by auth-service
type JWTService struct {
	// secretKey is used to verify tokens
	// MUST match the secret used by auth-service
	secretKey []byte
}

// Claims represents the data stored in the JWT token
// This must match the Claims structure in auth-service
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT validation service
// secretKey: the secret used to sign tokens (must match auth-service)
func NewJWTService(secretKey string) *JWTService {
	return &JWTService{
		secretKey: []byte(secretKey),
	}
}

// ValidateToken checks if a token is valid and extracts the claims
// Returns the claims if valid, or an error if invalid
func (s *JWTService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse the token
	// The function validates the signature and expiration automatically
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		// This prevents algorithm confusion attacks
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		// Return the secret key for verification
		return s.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Extract the claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
