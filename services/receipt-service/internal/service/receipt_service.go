package service

import (
	"context"
	"errors"
	"expense-tracker/receipt-service/internal/model"
	"expense-tracker/receipt-service/internal/repository"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"
)

// ReceiptService handles business logic for receipt operations
type ReceiptService struct {
	receiptRepo    repository.ReceiptRepository
	s3Service      *S3Service
	eventPublisher *EventPublisher // Optional - can be nil if not configured
}

// NewReceiptService creates a new receipt service
func NewReceiptService(receiptRepo repository.ReceiptRepository, s3Service *S3Service) *ReceiptService {
	return &ReceiptService{
		receiptRepo: receiptRepo,
		s3Service:   s3Service,
	}
}

// SetEventPublisher sets the event publisher (optional)
func (s *ReceiptService) SetEventPublisher(publisher *EventPublisher) {
	s.eventPublisher = publisher
}

// MaxFileSize is the maximum file size allowed (10MB)
const MaxFileSize = 10 * 1024 * 1024

// AllowedMimeTypes are the allowed file types
var AllowedMimeTypes = map[string]bool{
	"image/jpeg":      true,
	"image/jpg":       true,
	"image/png":       true,
	"application/pdf": true,
}

// UploadReceipt handles receipt file upload
// Validates file, uploads to S3, saves to database, and generates presigned URL
func (s *ReceiptService) UploadReceipt(ctx context.Context, userID string, file io.Reader, filename string, fileSize int64, expenseID *string) (*model.ReceiptResponse, error) {
	// Validate file size
	if fileSize > MaxFileSize {
		return nil, errors.New("file size exceeds maximum allowed size (10MB)")
	}

	if fileSize <= 0 {
		return nil, errors.New("file size must be greater than zero")
	}

	// Detect MIME type from filename extension
	mimeType := mime.TypeByExtension(filepath.Ext(filename))
	if mimeType == "" {
		// Try to detect from content (basic check)
		mimeType = "application/octet-stream"
	}

	// Normalize MIME type (handle variations like image/jpg)
	if mimeType == "image/jpg" {
		mimeType = "image/jpeg"
	}

	// Validate MIME type
	if !AllowedMimeTypes[mimeType] {
		return nil, fmt.Errorf("file type not allowed. Allowed types: JPEG, PNG, PDF")
	}

	// Create receipt record first to get ID
	receipt := model.NewReceipt(userID, filename, "", mimeType, fileSize)
	if expenseID != nil {
		receipt.ExpenseID = expenseID
	}

	// Generate S3 key
	s3Key := s.s3Service.GenerateFileKey(userID, receipt.ID, filename)
	receipt.FileKey = s3Key

	// Upload file to S3
	err := s.s3Service.UploadFile(ctx, file, fileSize, s3Key, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Generate presigned URL (valid for 1 hour)
	presignedURL, err := s.s3Service.GetPresignedURL(ctx, s3Key, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	receipt.FileURL = presignedURL

	// Save receipt to database
	receipt.CreatedAt = time.Now()
	receipt.UpdatedAt = time.Now()
	err = s.receiptRepo.Create(ctx, receipt)
	if err != nil {
		// If database save fails, try to clean up S3 file
		_ = s.s3Service.DeleteFile(ctx, s3Key)
		return nil, fmt.Errorf("failed to save receipt to database: %w", err)
	}

	// Publish event (non-blocking, async)
	if s.eventPublisher != nil {
		// Extract user email from context (set by auth middleware)
		userEmail := ""
		if emailVal := ctx.Value("user_email"); emailVal != nil {
			if email, ok := emailVal.(string); ok {
				userEmail = email
			}
		}

		eventData := map[string]interface{}{
			"receipt_id": receipt.ID,
			"file_name":  receipt.FileName,
			"file_size":  receipt.FileSize,
			"mime_type":  receipt.MimeType,
		}
		if expenseID != nil {
			eventData["expense_id"] = *expenseID
		}
		event := &Event{
			EventType: "receipt.uploaded",
			UserID:    userID,
			UserEmail: userEmail,
			Timestamp: time.Now(),
			Data:      eventData,
		}
		s.eventPublisher.PublishEventAsync(ctx, event)
	}

	// Return response
	return s.toReceiptResponse(receipt), nil
}

// GetReceipt retrieves a receipt by ID
func (s *ReceiptService) GetReceipt(ctx context.Context, receiptID, userID string) (*model.ReceiptResponse, error) {
	receipt, err := s.receiptRepo.FindByID(ctx, receiptID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipt: %w", err)
	}

	if receipt == nil {
		return nil, errors.New("receipt not found")
	}

	// Generate fresh presigned URL
	presignedURL, err := s.s3Service.GetPresignedURL(ctx, receipt.FileKey, 1*time.Hour)
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	receipt.FileURL = presignedURL

	return s.toReceiptResponse(receipt), nil
}

// GetReceiptsByExpense retrieves all receipts for a specific expense
func (s *ReceiptService) GetReceiptsByExpense(ctx context.Context, expenseID, userID string) ([]*model.ReceiptResponse, error) {
	receipts, err := s.receiptRepo.FindByExpenseID(ctx, expenseID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get receipts: %w", err)
	}

	// Generate presigned URLs for all receipts
	responses := make([]*model.ReceiptResponse, len(receipts))
	for i, receipt := range receipts {
		presignedURL, err := s.s3Service.GetPresignedURL(ctx, receipt.FileKey, 1*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to generate presigned URL for receipt %s: %w", receipt.ID, err)
		}
		receipt.FileURL = presignedURL
		responses[i] = s.toReceiptResponse(receipt)
	}

	return responses, nil
}

// ListReceipts lists receipts for a user with optional filters and pagination
func (s *ReceiptService) ListReceipts(ctx context.Context, userID string, filters *model.ListReceiptsRequest) (*model.ListReceiptsResponse, error) {
	receipts, total, err := s.receiptRepo.FindByUserID(ctx, userID, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list receipts: %w", err)
	}

	// Generate presigned URLs for all receipts
	responses := make([]model.ReceiptResponse, len(receipts))
	for i, receipt := range receipts {
		presignedURL, err := s.s3Service.GetPresignedURL(ctx, receipt.FileKey, 1*time.Hour)
		if err != nil {
			return nil, fmt.Errorf("failed to generate presigned URL for receipt %s: %w", receipt.ID, err)
		}
		receipt.FileURL = presignedURL
		responses[i] = *s.toReceiptResponse(receipt)
	}

	// Calculate pagination
	limit := filters.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	page := filters.Page
	if page <= 0 {
		page = 1
	}

	pages := (total + limit - 1) / limit // Ceiling division

	return &model.ListReceiptsResponse{
		Receipts: responses,
		Total:    total,
		Page:     page,
		Limit:    limit,
		Pages:    pages,
	}, nil
}

// LinkToExpense links a receipt to an expense
func (s *ReceiptService) LinkToExpense(ctx context.Context, receiptID, expenseID, userID string) error {
	// Verify receipt exists and belongs to user
	receipt, err := s.receiptRepo.FindByID(ctx, receiptID, userID)
	if err != nil {
		return fmt.Errorf("failed to get receipt: %w", err)
	}

	if receipt == nil {
		return errors.New("receipt not found")
	}

	// Update expense_id
	receipt.ExpenseID = &expenseID
	receipt.UpdatedAt = time.Now()

	err = s.receiptRepo.Update(ctx, receipt)
	if err != nil {
		return fmt.Errorf("failed to link receipt to expense: %w", err)
	}

	// Publish event (non-blocking, async)
	if s.eventPublisher != nil {
		// Extract user email from context (set by auth middleware)
		userEmail := ""
		if emailVal := ctx.Value("user_email"); emailVal != nil {
			if email, ok := emailVal.(string); ok {
				userEmail = email
			}
		}

		event := &Event{
			EventType: "receipt.linked",
			UserID:    userID,
			UserEmail: userEmail,
			Timestamp: time.Now(),
			Data: map[string]interface{}{
				"receipt_id": receiptID,
				"expense_id": expenseID,
				"file_name":  receipt.FileName,
			},
		}
		s.eventPublisher.PublishEventAsync(ctx, event)
	}

	return nil
}

// DeleteReceipt soft deletes a receipt and removes file from S3
func (s *ReceiptService) DeleteReceipt(ctx context.Context, receiptID, userID string) error {
	// Get receipt to get S3 key
	receipt, err := s.receiptRepo.FindByID(ctx, receiptID, userID)
	if err != nil {
		return fmt.Errorf("failed to get receipt: %w", err)
	}

	if receipt == nil {
		return errors.New("receipt not found")
	}

	// Soft delete in database
	err = s.receiptRepo.Delete(ctx, receiptID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete receipt: %w", err)
	}

	// Delete file from S3 (best effort - don't fail if S3 delete fails)
	_ = s.s3Service.DeleteFile(ctx, receipt.FileKey)

	return nil
}

// toReceiptResponse converts a Receipt model to ReceiptResponse DTO
func (s *ReceiptService) toReceiptResponse(receipt *model.Receipt) *model.ReceiptResponse {
	return &model.ReceiptResponse{
		ID:           receipt.ID,
		UserID:       receipt.UserID,
		ExpenseID:    receipt.ExpenseID,
		FileName:     receipt.FileName,
		FileURL:      receipt.FileURL,
		FileSize:     receipt.FileSize,
		MimeType:     receipt.MimeType,
		MerchantName: receipt.MerchantName,
		ReceiptDate:  receipt.ReceiptDate,
		TotalAmount:  receipt.TotalAmount,
		CreatedAt:    receipt.CreatedAt,
		UpdatedAt:    receipt.UpdatedAt,
	}
}
