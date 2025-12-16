package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AuthClient handles communication with auth-service for token validation
// This centralizes authentication logic in auth-service
type AuthClient struct {
	authServiceURL string
	httpClient     *http.Client
}

// ValidateResponse represents the response from auth-service /auth/validate endpoint
type ValidateResponse struct {
	Valid  bool   `json:"valid"`
	UserID string `json:"user_id,omitempty"`
	Email  string `json:"email,omitempty"`
}

// NewAuthClient creates a new auth client
// authServiceURL: base URL of auth-service (e.g., "http://localhost:8080")
func NewAuthClient(authServiceURL string) *AuthClient {
	return &AuthClient{
		authServiceURL: authServiceURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second, // 5 second timeout for auth calls
		},
	}
}

// ValidateToken calls auth-service to validate a JWT token
// Returns user_id and email if token is valid, error otherwise
func (c *AuthClient) ValidateToken(ctx context.Context, token string) (userID, email string, err error) {
	// Build the request URL
	url := fmt.Sprintf("%s/auth/validate", c.authServiceURL)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set Authorization header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Content-Type", "application/json")

	// Make the HTTP call
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("failed to call auth-service: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode == http.StatusUnauthorized {
		return "", "", fmt.Errorf("invalid token")
	}

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("auth-service returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var validateResp ValidateResponse
	if err := json.Unmarshal(body, &validateResp); err != nil {
		return "", "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if token is valid
	if !validateResp.Valid {
		return "", "", fmt.Errorf("invalid token")
	}

	return validateResp.UserID, validateResp.Email, nil
}
