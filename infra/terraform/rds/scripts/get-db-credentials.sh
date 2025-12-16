#!/bin/bash
# Script to retrieve RDS credentials from AWS Secrets Manager
# Usage: ./get-db-credentials.sh [master|app]

set -e

SECRET_TYPE=${1:-master}
PROJECT_NAME=${PROJECT_NAME:-expense-tracker}

if [ "$SECRET_TYPE" = "master" ]; then
    SECRET_NAME="${PROJECT_NAME}/rds/master-credentials"
elif [ "$SECRET_TYPE" = "app" ]; then
    SECRET_NAME="${PROJECT_NAME}/rds/app-credentials"
else
    echo "Usage: $0 [master|app]"
    exit 1
fi

echo "Retrieving ${SECRET_TYPE} credentials from Secrets Manager..."
echo "Secret: ${SECRET_NAME}"
echo ""

# Get secret value
SECRET_JSON=$(aws secretsmanager get-secret-value \
    --secret-id "${SECRET_NAME}" \
    --query SecretString \
    --output text)

if [ -z "$SECRET_JSON" ]; then
    echo "Error: Could not retrieve secret"
    exit 1
fi

# Parse and display (using jq if available)
if command -v jq &> /dev/null; then
    echo "Credentials:"
    echo "$SECRET_JSON" | jq .
    
    echo ""
    echo "Connection string:"
    HOST=$(echo "$SECRET_JSON" | jq -r .host)
    PORT=$(echo "$SECRET_JSON" | jq -r .port)
    DBNAME=$(echo "$SECRET_JSON" | jq -r .dbname)
    USERNAME=$(echo "$SECRET_JSON" | jq -r .username)
    
    echo "postgres://${USERNAME}:***@${HOST}:${PORT}/${DBNAME}"
else
    echo "$SECRET_JSON"
fi
