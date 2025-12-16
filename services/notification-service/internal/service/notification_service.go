package service

import (
	"context"
	"expense-tracker/notification-service/internal/model"
	"fmt"
	"log"
)

// NotificationService handles processing events and sending notifications
type NotificationService struct {
	snsPublisher    *SNSPublisher
	templateService *TemplateService
}

// NewNotificationService creates a new notification service
func NewNotificationService(snsPublisher *SNSPublisher, templateService *TemplateService) *NotificationService {
	return &NotificationService{
		snsPublisher:    snsPublisher,
		templateService: templateService,
	}
}

// ProcessEvent processes an event and sends the appropriate notification
func (s *NotificationService) ProcessEvent(ctx context.Context, event *model.Event) error {
	log.Printf("Processing event: %s for user %s", event.EventType, event.UserID)

	// Determine notification type and render template
	var templateName string
	var subject string
	var templateData interface{}

	switch event.EventType {
	case model.EventTypeExpenseCreated:
		templateName = "expense_created"
		subject = "New Expense Added"
		templateData = s.buildExpenseCreatedData(event)

	case model.EventTypeExpenseUpdated:
		templateName = "expense_updated"
		subject = "Expense Updated"
		templateData = s.buildExpenseUpdatedData(event)

	case model.EventTypeReceiptUploaded:
		templateName = "receipt_uploaded"
		subject = "Receipt Uploaded"
		templateData = s.buildReceiptUploadedData(event)

	case model.EventTypeReceiptLinked:
		templateName = "receipt_linked"
		subject = "Receipt Linked to Expense"
		templateData = s.buildReceiptLinkedData(event)

	case model.EventTypeUserRegistered:
		templateName = "user_registered"
		subject = "Welcome to Expense Tracker!"
		templateData = s.buildUserRegisteredData(event)

	default:
		return fmt.Errorf("unknown event type: %s", event.EventType)
	}

	// Render email template
	body, err := s.templateService.Render(templateName, templateData)
	if err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	// Create notification
	notification := &model.Notification{
		To:      event.UserEmail,
		Subject: subject,
		Body:    body,
	}

	// Send email via SNS
	if err := s.snsPublisher.SendEmail(ctx, notification); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	log.Printf("Successfully sent notification for event %s to %s", event.EventType, event.UserEmail)
	return nil
}

// buildExpenseCreatedData builds template data for expense created event
func (s *NotificationService) buildExpenseCreatedData(event *model.Event) map[string]interface{} {
	data := make(map[string]interface{})

	// Extract expense data
	if expenseData, ok := event.Data["expense_id"].(string); ok {
		data["ExpenseID"] = expenseData
	}
	if amount, ok := event.Data["amount"].(string); ok {
		data["Amount"] = amount
	}
	if description, ok := event.Data["description"].(string); ok {
		data["Description"] = description
	}
	if category, ok := event.Data["category"].(string); ok {
		data["Category"] = category
	}
	if expenseDate, ok := event.Data["expense_date"].(string); ok {
		data["ExpenseDate"] = expenseDate
	}

	// Add user info
	data["UserEmail"] = event.UserEmail
	data["Content"] = fmt.Sprintf(
		"<h2>New Expense Added</h2><p>You've added a new expense:</p><ul><li><strong>Amount:</strong> $%s</li><li><strong>Description:</strong> %s</li><li><strong>Category:</strong> %s</li><li><strong>Date:</strong> %s</li></ul>",
		data["Amount"], data["Description"], data["Category"], data["ExpenseDate"],
	)

	return data
}

// buildExpenseUpdatedData builds template data for expense updated event
func (s *NotificationService) buildExpenseUpdatedData(event *model.Event) map[string]interface{} {
	data := make(map[string]interface{})

	if expenseID, ok := event.Data["expense_id"].(string); ok {
		data["ExpenseID"] = expenseID
	}
	if amount, ok := event.Data["amount"].(string); ok {
		data["Amount"] = amount
	}
	if description, ok := event.Data["description"].(string); ok {
		data["Description"] = description
	}
	if category, ok := event.Data["category"].(string); ok {
		data["Category"] = category
	}
	if expenseDate, ok := event.Data["expense_date"].(string); ok {
		data["ExpenseDate"] = expenseDate
	}

	data["UserEmail"] = event.UserEmail
	data["Content"] = fmt.Sprintf(
		"<h2>Expense Updated</h2><p>Your expense has been updated:</p><ul><li><strong>Amount:</strong> $%s</li><li><strong>Description:</strong> %s</li><li><strong>Category:</strong> %s</li><li><strong>Date:</strong> %s</li></ul>",
		data["Amount"], data["Description"], data["Category"], data["ExpenseDate"],
	)

	return data
}

// buildReceiptUploadedData builds template data for receipt uploaded event
func (s *NotificationService) buildReceiptUploadedData(event *model.Event) map[string]interface{} {
	data := make(map[string]interface{})

	if receiptID, ok := event.Data["receipt_id"].(string); ok {
		data["ReceiptID"] = receiptID
	}
	if fileName, ok := event.Data["file_name"].(string); ok {
		data["FileName"] = fileName
	}
	if fileSize, ok := event.Data["file_size"].(float64); ok {
		data["FileSize"] = int64(fileSize)
	}
	if mimeType, ok := event.Data["mime_type"].(string); ok {
		data["MimeType"] = mimeType
	}
	if expenseID, ok := event.Data["expense_id"].(string); ok {
		data["ExpenseID"] = expenseID
	}

	data["UserEmail"] = event.UserEmail

	expenseInfo := ""
	if expenseID, ok := data["ExpenseID"].(string); ok && expenseID != "" {
		expenseInfo = fmt.Sprintf("<li><strong>Linked to Expense:</strong> %s</li>", expenseID)
	}

	data["Content"] = fmt.Sprintf(
		"<h2>Receipt Uploaded</h2><p>You've uploaded a new receipt:</p><ul><li><strong>File Name:</strong> %s</li><li><strong>File Size:</strong> %d bytes</li><li><strong>Type:</strong> %s</li>%s</ul>",
		data["FileName"], data["FileSize"], data["MimeType"], expenseInfo,
	)

	return data
}

// buildReceiptLinkedData builds template data for receipt linked event
func (s *NotificationService) buildReceiptLinkedData(event *model.Event) map[string]interface{} {
	data := make(map[string]interface{})

	if receiptID, ok := event.Data["receipt_id"].(string); ok {
		data["ReceiptID"] = receiptID
	}
	if expenseID, ok := event.Data["expense_id"].(string); ok {
		data["ExpenseID"] = expenseID
	}
	if fileName, ok := event.Data["file_name"].(string); ok {
		data["FileName"] = fileName
	}

	data["UserEmail"] = event.UserEmail
	data["Content"] = fmt.Sprintf(
		"<h2>Receipt Linked to Expense</h2><p>Your receipt has been linked to an expense:</p><ul><li><strong>Receipt:</strong> %s</li><li><strong>Expense ID:</strong> %s</li></ul>",
		data["FileName"], data["ExpenseID"],
	)

	return data
}

// buildUserRegisteredData builds template data for user registered event
func (s *NotificationService) buildUserRegisteredData(event *model.Event) map[string]interface{} {
	data := make(map[string]interface{})

	if userID, ok := event.Data["user_id"].(string); ok {
		data["UserID"] = userID
	}
	if email, ok := event.Data["email"].(string); ok {
		data["Email"] = email
	}
	if name, ok := event.Data["name"].(string); ok {
		data["Name"] = name
	}

	data["UserEmail"] = event.UserEmail
	data["Content"] = fmt.Sprintf(
		"<h2>Welcome to Expense Tracker!</h2><p>Hi %s,</p><p>Thank you for registering with Expense Tracker. You can now start tracking your expenses and receipts.</p><p>Get started by creating your first expense!</p>",
		data["Name"],
	)

	return data
}
