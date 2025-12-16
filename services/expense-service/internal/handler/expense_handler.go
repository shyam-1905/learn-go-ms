package handler

import (
	"encoding/json"
	"expense-tracker/expense-service/internal/middleware"
	"expense-tracker/expense-service/internal/model"
	"expense-tracker/expense-service/internal/service"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// ExpenseHandler handles HTTP requests for expenses
type ExpenseHandler struct {
	expenseService *service.ExpenseService
}

// NewExpenseHandler creates a new expense handler
func NewExpenseHandler(expenseService *service.ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{
		expenseService: expenseService,
	}
}

// CreateExpense handles expense creation
// POST /expenses
func (h *ExpenseHandler) CreateExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user_id from context (set by auth middleware)
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Decode JSON request body
	var req model.CreateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call the expense service
	resp, err := h.expenseService.CreateExpense(r.Context(), userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "format") {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to create expense")
		return
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

// GetExpense handles getting a single expense
// GET /expenses/:id
func (h *ExpenseHandler) GetExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user_id from context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get expense ID from URL path
	vars := mux.Vars(r)
	expenseID := vars["id"]

	if expenseID == "" {
		respondWithError(w, http.StatusBadRequest, "Expense ID is required")
		return
	}

	// Call the expense service
	resp, err := h.expenseService.GetExpense(r.Context(), expenseID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get expense")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// ListExpenses handles listing expenses with filters and pagination
// GET /expenses?category=Food&start_date=2024-01-01&end_date=2024-01-31&page=1&limit=20
func (h *ExpenseHandler) ListExpenses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user_id from context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse query parameters
	filters := &model.ListExpensesRequest{
		Category:  r.URL.Query().Get("category"),
		StartDate: r.URL.Query().Get("start_date"),
		EndDate:   r.URL.Query().Get("end_date"),
	}

	// Parse pagination parameters
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err == nil && page > 0 {
			filters.Page = page
		} else {
			filters.Page = 1
		}
	} else {
		filters.Page = 1
	}

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err == nil && limit > 0 {
			if limit > 100 {
				limit = 100 // Max limit
			}
			filters.Limit = limit
		} else {
			filters.Limit = 20
		}
	} else {
		filters.Limit = 20
	}

	// Call the expense service
	resp, err := h.expenseService.ListExpenses(r.Context(), userID, filters)
	if err != nil {
		if strings.Contains(err.Error(), "format") {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to list expenses")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// UpdateExpense handles expense updates
// PUT /expenses/:id
func (h *ExpenseHandler) UpdateExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user_id from context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get expense ID from URL path
	vars := mux.Vars(r)
	expenseID := vars["id"]

	if expenseID == "" {
		respondWithError(w, http.StatusBadRequest, "Expense ID is required")
		return
	}

	// Decode JSON request body
	var req model.UpdateExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Call the expense service
	resp, err := h.expenseService.UpdateExpense(r.Context(), expenseID, userID, &req)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		if strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "format") {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to update expense")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// DeleteExpense handles expense deletion
// DELETE /expenses/:id
func (h *ExpenseHandler) DeleteExpense(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user_id from context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Get expense ID from URL path
	vars := mux.Vars(r)
	expenseID := vars["id"]

	if expenseID == "" {
		respondWithError(w, http.StatusBadRequest, "Expense ID is required")
		return
	}

	// Call the expense service
	err := h.expenseService.DeleteExpense(r.Context(), expenseID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete expense")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Expense deleted successfully",
	})
}

// GetSummary handles expense summary by category
// GET /expenses/summary?start_date=2024-01-01&end_date=2024-01-31
func (h *ExpenseHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user_id from context
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse query parameters
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	var startDatePtr, endDatePtr *string
	if startDate != "" {
		startDatePtr = &startDate
	}
	if endDate != "" {
		endDatePtr = &endDate
	}

	// Call the expense service
	resp, err := h.expenseService.GetExpenseSummary(r.Context(), userID, startDatePtr, endDatePtr)
	if err != nil {
		if strings.Contains(err.Error(), "format") {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get expense summary")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// Health handles health check requests
// GET /health
func (h *ExpenseHandler) Health(w http.ResponseWriter, r *http.Request) {
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
func respondWithJSON(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
	respondWithJSON(w, statusCode, map[string]string{
		"error": message,
	})
}
