# Notification Service Environment Variables
# Run this script before starting the notification service
# Usage: .\set-env.ps1

# NOTE: Do NOT commit real secrets to this file.
# Replace the placeholder values below with your own credentials locally,
# or, better, set them outside of source control.

# AWS Configuration
$env:AWS_REGION = "us-east-1"
$env:AWS_ACCESS_KEY_ID = "YOUR_AWS_ACCESS_KEY_ID"
$env:AWS_SECRET_ACCESS_KEY = "YOUR_AWS_SECRET_ACCESS_KEY"

# SQS Queue URLs (for consuming events)
$env:EXPENSE_EVENTS_QUEUE_URL = "https://sqs.us-east-1.amazonaws.com/ACCOUNT_ID/expense-events-queue"
$env:RECEIPT_EVENTS_QUEUE_URL = "https://sqs.us-east-1.amazonaws.com/ACCOUNT_ID/receipt-events-queue"
$env:AUTH_EVENTS_QUEUE_URL = "https://sqs.us-east-1.amazonaws.com/ACCOUNT_ID/auth-events-queue"

# SNS Topic ARN (for sending email notifications)
$env:NOTIFICATION_EMAIL_TOPIC_ARN = "arn:aws:sns:us-east-1:ACCOUNT_ID:notification-email-topic"

# Server Configuration
$env:SERVER_PORT = "8083"

Write-Host "Environment variables set for notification-service" -ForegroundColor Green
Write-Host "Note: Update the queue URLs and topic ARN after running setup-aws.go" -ForegroundColor Yellow
