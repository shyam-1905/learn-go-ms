package handler

import (
	"encoding/json"
	"expense-tracker/auth-service/internal/model"
	"expense-tracker/auth-service/internal/service"
	"net/http"
	"strings"
)

// AuthHandler handles HTTP requests for authentication
// In Go, HTTP handlers are functions that take http.ResponseWriter and *http.Request
// We use a struct to hold dependencies (the auth service)
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register handles user registration
// POST /auth/register
// Request body: { "email": "...", "password": "...", "name": "..." }
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Decode JSON request body
	var req model.RegisterRequest
	// json.NewDecoder reads from the request body
	// Decode parses JSON into our struct
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// If JSON is invalid, return 400 Bad Request
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields (simple validation)
	if req.Email == "" || req.Password == "" || req.Name == "" {
		respondWithError(w, http.StatusBadRequest, "Email, password, and name are required")
		return
	}

	// Call the auth service
	// r.Context() provides a context for cancellation/timeouts
	resp, err := h.authService.Register(r.Context(), &req)
	if err != nil {
		// Check error type and return appropriate status code
		if strings.Contains(err.Error(), "already exists") {
			respondWithError(w, http.StatusConflict, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to register user")
		return
	}

	// Return success response
	respondWithJSON(w, http.StatusCreated, resp)
}

// Login handles user login
// POST /auth/login
// Request body: { "email": "...", "password": "..." }
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		respondWithError(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	resp, err := h.authService.Login(r.Context(), &req)
	if err != nil {
		// Invalid credentials
		if strings.Contains(err.Error(), "invalid") {
			respondWithError(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to login")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// Validate handles token validation
// GET /auth/validate
// Headers: Authorization: Bearer <token>
func (h *AuthHandler) Validate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract token from Authorization header
	// Format: "Bearer <token>"
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		respondWithError(w, http.StatusUnauthorized, "Authorization header required")
		return
	}

	// Check if it starts with "Bearer "
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		respondWithError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}

	token := parts[1]

	// Validate the token
	resp, err := h.authService.ValidateToken(r.Context(), token)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to validate token")
		return
	}

	if !resp.Valid {
		respondWithError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// Health handles health check requests
// GET /health
func (h *AuthHandler) Health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// Helper functions for JSON responses

// respondWithJSON sends a JSON response
// w is the ResponseWriter (where we write the response)
// statusCode is the HTTP status code (200, 400, 500, etc.)
// payload is the data to send (will be converted to JSON)
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	// Set content type header
	w.Header().Set("Content-Type", "application/json")
	// Set status code
	w.WriteHeader(statusCode)
	// Encode payload to JSON and write to response
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// If encoding fails, log error (in production, use a logger)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
