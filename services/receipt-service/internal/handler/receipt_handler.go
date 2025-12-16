package handler

import (
	"encoding/json"
	"expense-tracker/receipt-service/internal/middleware"
	"expense-tracker/receipt-service/internal/model"
	"expense-tracker/receipt-service/internal/service"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

// ReceiptHandler handles HTTP requests for receipts
type ReceiptHandler struct {
	receiptService *service.ReceiptService
}

// NewReceiptHandler creates a new receipt handler
func NewReceiptHandler(receiptService *service.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{
		receiptService: receiptService,
	}
}

// UploadReceipt handles receipt file upload
// POST /receipts
func (h *ReceiptHandler) UploadReceipt(w http.ResponseWriter, r *http.Request) {
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

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20) // 10MB
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Failed to parse multipart form")
		return
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "File is required")
		return
	}
	defer file.Close()

	// Get optional expense_id from form
	var expenseID *string
	if expenseIDStr := r.FormValue("expense_id"); expenseIDStr != "" {
		expenseID = &expenseIDStr
	}

	// Upload receipt
	resp, err := h.receiptService.UploadReceipt(
		r.Context(),
		userID,
		file,
		header.Filename,
		header.Size,
		expenseID,
	)
	if err != nil {
		// Log the actual error for debugging
		log.Printf("Error uploading receipt: %v", err)

		// Handle specific error types
		if strings.Contains(err.Error(), "size") || strings.Contains(err.Error(), "type") {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		// Check for S3/AWS errors
		if strings.Contains(err.Error(), "S3") || strings.Contains(err.Error(), "AWS") ||
			strings.Contains(err.Error(), "bucket") || strings.Contains(err.Error(), "credentials") ||
			strings.Contains(err.Error(), "NoSuchBucket") || strings.Contains(err.Error(), "AccessDenied") {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Storage error: %v", err))
			return
		}

		// Check for database errors
		if strings.Contains(err.Error(), "database") || strings.Contains(err.Error(), "connection") {
			respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Database error: %v", err))
			return
		}

		// Generic error with more detail
		respondWithError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to upload receipt: %v", err))
		return
	}

	respondWithJSON(w, http.StatusCreated, resp)
}

// GetReceipt handles getting a single receipt
// GET /receipts/:id
func (h *ReceiptHandler) GetReceipt(w http.ResponseWriter, r *http.Request) {
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

	// Get receipt ID from URL path
	vars := mux.Vars(r)
	receiptID := vars["id"]

	if receiptID == "" {
		respondWithError(w, http.StatusBadRequest, "Receipt ID is required")
		return
	}

	// Call the receipt service
	resp, err := h.receiptService.GetReceipt(r.Context(), receiptID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to get receipt")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// ListReceipts handles listing receipts with filters and pagination
// GET /receipts?expense_id=xxx&page=1&limit=20
func (h *ReceiptHandler) ListReceipts(w http.ResponseWriter, r *http.Request) {
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

	// Check if expense_id query parameter is provided
	expenseID := r.URL.Query().Get("expense_id")
	if expenseID != "" {
		// Get receipts by expense_id
		receipts, err := h.receiptService.GetReceiptsByExpense(r.Context(), expenseID, userID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Failed to get receipts")
			return
		}

		// Return as list response
		respondWithJSON(w, http.StatusOK, map[string]interface{}{
			"receipts": receipts,
			"total":    len(receipts),
		})
		return
	}

	// Parse query parameters for pagination
	filters := &model.ListReceiptsRequest{}

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

	// Call the receipt service
	resp, err := h.receiptService.ListReceipts(r.Context(), userID, filters)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to list receipts")
		return
	}

	respondWithJSON(w, http.StatusOK, resp)
}

// LinkReceipt handles linking a receipt to an expense
// PUT /receipts/:id/link
func (h *ReceiptHandler) LinkReceipt(w http.ResponseWriter, r *http.Request) {
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

	// Get receipt ID from URL path
	vars := mux.Vars(r)
	receiptID := vars["id"]

	if receiptID == "" {
		respondWithError(w, http.StatusBadRequest, "Receipt ID is required")
		return
	}

	// Decode JSON request body
	var req model.LinkReceiptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.ExpenseID == "" {
		respondWithError(w, http.StatusBadRequest, "expense_id is required")
		return
	}

	// Call the receipt service
	err := h.receiptService.LinkToExpense(r.Context(), receiptID, req.ExpenseID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to link receipt to expense")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Receipt linked to expense successfully",
	})
}

// DeleteReceipt handles receipt deletion
// DELETE /receipts/:id
func (h *ReceiptHandler) DeleteReceipt(w http.ResponseWriter, r *http.Request) {
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

	// Get receipt ID from URL path
	vars := mux.Vars(r)
	receiptID := vars["id"]

	if receiptID == "" {
		respondWithError(w, http.StatusBadRequest, "Receipt ID is required")
		return
	}

	// Call the receipt service
	err := h.receiptService.DeleteReceipt(r.Context(), receiptID, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			respondWithError(w, http.StatusNotFound, err.Error())
			return
		}
		respondWithError(w, http.StatusInternalServerError, "Failed to delete receipt")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"message": "Receipt deleted successfully",
	})
}

// Health handles health check requests
// GET /health
func (h *ReceiptHandler) Health(w http.ResponseWriter, r *http.Request) {
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
