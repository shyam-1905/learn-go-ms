package model

// Notification represents a notification to be sent
type Notification struct {
	To      string // Email address
	Subject string // Email subject
	Body    string // Email body (HTML)
}

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeExpenseCreated  NotificationType = "expense_created"
	NotificationTypeExpenseUpdated  NotificationType = "expense_updated"
	NotificationTypeReceiptUploaded NotificationType = "receipt_uploaded"
	NotificationTypeReceiptLinked   NotificationType = "receipt_linked"
	NotificationTypeUserRegistered  NotificationType = "user_registered"
)
