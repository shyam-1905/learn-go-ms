#!/bin/bash
# Bash script to set environment variables for auth-service
# Run this before starting the service: source set-env.sh

# Database Configuration (AWS RDS)
export DB_HOST="test-db.c8bs0qiu4gv4.us-east-1.rds.amazonaws.com"
export DB_PORT="5432"
export DB_USER="postgres"
export DB_PASSWORD="Adminpass123"
export DB_NAME="auth_db"

# JWT Configuration
# Generate a strong secret: openssl rand -base64 32
export JWT_SECRET="your-super-secret-jwt-key-change-this-in-production"
export JWT_EXPIRATION_HOURS="24"

# Server Configuration
export SERVER_PORT="8080"

echo "Environment variables set!"
echo ""
echo "Database Host: $DB_HOST"
echo "Database Port: $DB_PORT"
echo "Database User: $DB_USER"
echo "Database Name: $DB_NAME"
echo ""
echo "To start the service, run: go run cmd/main.go"
