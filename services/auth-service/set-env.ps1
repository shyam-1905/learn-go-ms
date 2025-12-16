# PowerShell script to set environment variables for auth-service
# Run this before starting the service: .\set-env.ps1

# NOTE: Do NOT commit real secrets to this file.
# Replace the placeholder values below with your own credentials locally,
# or, better, set them outside of source control.

# Database Configuration (AWS RDS)
$env:DB_HOST = "your-db-hostname.rds.amazonaws.com"
$env:DB_PORT = "5432"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "CHANGE_ME_DB_PASSWORD"
$env:DB_NAME = "auth_db"

# JWT Configuration
# Generate a strong secret: openssl rand -base64 32
$env:JWT_SECRET = "CHANGE_ME_JWT_SECRET"
$env:JWT_EXPIRATION_HOURS = "24"

# AWS SNS configuration for event publishing
$env:AWS_REGION = "us-east-1"
$env:AWS_ACCESS_KEY_ID = "YOUR_AWS_ACCESS_KEY_ID"
$env:AWS_SECRET_ACCESS_KEY = "YOUR_AWS_SECRET_ACCESS_KEY"
$env:AUTH_EVENTS_TOPIC_ARN = "arn:aws:sns:us-east-1:ACCOUNT_ID:auth-events-topic"

# Server Configuration
$env:SERVER_PORT = "8080"

Write-Host "Environment variables set!" -ForegroundColor Green
Write-Host ""
Write-Host "Database Host: $env:DB_HOST"
Write-Host "Database Port: $env:DB_PORT"
Write-Host "Database User: $env:DB_USER"
Write-Host "Database Name: $env:DB_NAME"
Write-Host ""
Write-Host "To start the service, run: go run cmd/main.go" -ForegroundColor Cyan
