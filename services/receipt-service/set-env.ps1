# PowerShell script to set environment variables for receipt-service
# Run this script before starting the service: . .\set-env.ps1

# NOTE: Do NOT commit real secrets to this file.
# Replace the placeholder values below with your own credentials locally,
# or better yet, set them outside of source control (e.g. in your shell profile or CI secrets).

# Database configuration (same RDS instance as other services)
$env:DB_HOST = "your-db-hostname.rds.amazonaws.com"
$env:DB_PORT = "5432"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "CHANGE_ME_DB_PASSWORD"
$env:DB_NAME = "receipt_db"

# Auth Service configuration (URL of auth-service for token validation)
$env:AUTH_SERVICE_URL = "http://localhost:8080"

# AWS S3 configuration
$env:AWS_REGION = "us-east-1"
$env:AWS_ACCESS_KEY_ID = "YOUR_AWS_ACCESS_KEY_ID"
$env:AWS_SECRET_ACCESS_KEY = "YOUR_AWS_SECRET_ACCESS_KEY"
# S3 Bucket Name - must be globally unique across all AWS accounts
$env:S3_BUCKET_NAME = "your-receipt-service-bucket-name"

# AWS SNS configuration for event publishing
$env:RECEIPT_EVENTS_TOPIC_ARN = "arn:aws:sns:us-east-1:ACCOUNT_ID:receipt-events-topic"

# Server configuration
$env:SERVER_PORT = "8082"

Write-Host "Environment variables set for receipt-service!" -ForegroundColor Green
Write-Host ""
Write-Host "Database: $env:DB_NAME on $env:DB_HOST" -ForegroundColor Cyan
Write-Host "Server Port: $env:SERVER_PORT" -ForegroundColor Cyan
Write-Host "AWS Region: $env:AWS_REGION" -ForegroundColor Cyan
Write-Host "S3 Bucket: $env:S3_BUCKET_NAME" -ForegroundColor Cyan
Write-Host ""
Write-Host "You can now run:" -ForegroundColor Yellow
Write-Host "  go run cmd/main.go" -ForegroundColor White
