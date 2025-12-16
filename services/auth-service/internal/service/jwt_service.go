package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles JWT token generation and validation
// JWT (JSON Web Token) is a way to securely transmit information between parties
// It consists of three parts: Header.Payload.Signature
type JWTService struct {
	// secretKey is used to sign and verify tokens
	// NEVER expose this key! It should come from environment variables
	secretKey []byte

	// tokenExpiration is how long tokens are valid (e.g., 24 hours)
	tokenExpiration time.Duration
}

// Claims represents the data stored in the JWT token
// jwt.RegisteredClaims includes standard fields like "exp" (expiration), "iat" (issued at), etc.
type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// NewJWTService creates a new JWT service
// secretKey: the secret used to sign tokens (from environment variable)
// tokenExpiration: how long tokens are valid (e.g., 24 * time.Hour)
func NewJWTService(secretKey string, tokenExpiration time.Duration) *JWTService {
	return &JWTService{
		secretKey:       []byte(secretKey), // Convert string to []byte for signing
		tokenExpiration: tokenExpiration,
	}
}

// GenerateToken creates a new JWT token for a user
// This token will be sent to the client and used for authenticated requests
func (s *JWTService) GenerateToken(userID, email string) (string, error) {
	// Create the expiration time
	expirationTime := time.Now().Add(s.tokenExpiration)

	// Create the claims (the data in the token)
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			// ExpiresAt: when the token expires
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			// IssuedAt: when the token was created
			IssuedAt: jwt.NewNumericDate(time.Now()),
			// Issuer: who issued the token (optional, but good practice)
			Issuer: "auth-service",
		},
	}

	// Create the token with the signing method and claims
	// HS256 is a symmetric algorithm (uses the same secret to sign and verify)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Sign the token with our secret key
	// This creates the signature part of the JWT
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
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
