# PowerShell script to set environment variables for expense-service
# Run this before starting the service: .\set-env.ps1

# NOTE: Do NOT commit real secrets to this file.
# Replace the placeholder values below with your own credentials locally,
# or, better, set them outside of source control.

# Database Configuration (AWS RDS)
$env:DB_HOST = "your-db-hostname.rds.amazonaws.com"
$env:DB_PORT = "5432"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "CHANGE_ME_DB_PASSWORD"
$env:DB_NAME = "expense_db"

# Auth Service Configuration
# URL of auth-service for token validation
$env:AUTH_SERVICE_URL = "http://localhost:8080"

# AWS SNS configuration for event publishing
$env:AWS_REGION = "us-east-1"
$env:AWS_ACCESS_KEY_ID = "YOUR_AWS_ACCESS_KEY_ID"
$env:AWS_SECRET_ACCESS_KEY = "YOUR_AWS_SECRET_ACCESS_KEY"
$env:EXPENSE_EVENTS_TOPIC_ARN = "arn:aws:sns:us-east-1:ACCOUNT_ID:expense-events-topic"

# Server Configuration
$env:SERVER_PORT = "8081"

Write-Host "Environment variables set!" -ForegroundColor Green
Write-Host ""
Write-Host "Database Host: $env:DB_HOST"
Write-Host "Database Port: $env:DB_PORT"
Write-Host "Database User: $env:DB_USER"
Write-Host "Database Name: $env:DB_NAME"
Write-Host "Server Port: $env:SERVER_PORT"
Write-Host ""
Write-Host "To start the service, run: go run cmd/main.go" -ForegroundColor Cyan
