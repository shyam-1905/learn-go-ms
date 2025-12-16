#!/bin/bash
# Script to set up Kubernetes secrets for auth-service from Terraform outputs
# Usage: ./setup-k8s-secrets.sh [namespace]

set -e

NAMESPACE=${1:-default}
PROJECT_NAME=${PROJECT_NAME:-expense-tracker}

echo "Setting up Kubernetes secrets for auth-service in namespace: ${NAMESPACE}"
echo ""

# Check if namespace exists, create if not
if ! kubectl get namespace ${NAMESPACE} &> /dev/null; then
    echo "Creating namespace: ${NAMESPACE}"
    kubectl create namespace ${NAMESPACE}
fi

# Get RDS endpoint from Terraform (assumes we're in terraform directory)
if [ -f "terraform.tfstate" ]; then
    RDS_ENDPOINT=$(terraform output -raw rds_endpoint 2>/dev/null || echo "")
    if [ -z "$RDS_ENDPOINT" ]; then
        echo "Error: Could not get RDS endpoint from Terraform"
        echo "Make sure you're in the terraform directory and have applied the configuration"
        exit 1
    fi
else
    echo "Error: terraform.tfstate not found"
    echo "Please run this script from the terraform directory after applying configuration"
    exit 1
fi

echo "RDS Endpoint: ${RDS_ENDPOINT}"

# Get credentials from Secrets Manager
echo "Retrieving credentials from AWS Secrets Manager..."
SECRET_JSON=$(aws secretsmanager get-secret-value \
    --secret-id "${PROJECT_NAME}/rds/master-credentials" \
    --query SecretString \
    --output text)

if [ -z "$SECRET_JSON" ]; then
    echo "Error: Could not retrieve secret from Secrets Manager"
    exit 1
fi

# Parse JSON (requires jq)
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required. Install it: https://stedolan.github.io/jq/download/"
    exit 1
fi

DB_USER=$(echo "$SECRET_JSON" | jq -r .username)
DB_PASSWORD=$(echo "$SECRET_JSON" | jq -r .password)
DB_NAME=$(echo "$SECRET_JSON" | jq -r .dbname)
DB_PORT=$(echo "$SECRET_JSON" | jq -r .port)

echo "Creating Kubernetes secret: auth-db-credentials"

# Delete existing secret if it exists
kubectl delete secret auth-db-credentials -n ${NAMESPACE} 2>/dev/null || true

# Create new secret
kubectl create secret generic auth-db-credentials \
    --from-literal=DB_HOST=${RDS_ENDPOINT} \
    --from-literal=DB_PORT=${DB_PORT} \
    --from-literal=DB_USER=${DB_USER} \
    --from-literal=DB_PASSWORD=${DB_PASSWORD} \
    --from-literal=DB_NAME=${DB_NAME} \
    -n ${NAMESPACE}

echo ""
echo "✅ Database secret created successfully!"

# Generate and create JWT secret
echo ""
echo "Generating JWT secret..."
JWT_SECRET=$(openssl rand -base64 32)

kubectl delete secret auth-jwt-secret -n ${NAMESPACE} 2>/dev/null || true

kubectl create secret generic auth-jwt-secret \
    --from-literal=JWT_SECRET=${JWT_SECRET} \
    -n ${NAMESPACE}

echo "✅ JWT secret created successfully!"

# Create ConfigMap for non-sensitive config
echo ""
echo "Creating ConfigMap: auth-service-config"

kubectl delete configmap auth-service-config -n ${NAMESPACE} 2>/dev/null || true

kubectl create configmap auth-service-config \
    --from-literal=SERVER_PORT=8080 \
    --from-literal=JWT_EXPIRATION_HOURS=24 \
    -n ${NAMESPACE}

echo "✅ ConfigMap created successfully!"

echo ""
echo "Summary:"
echo "  Namespace: ${NAMESPACE}"
echo "  Secrets created:"
echo "    - auth-db-credentials"
echo "    - auth-jwt-secret"
echo "  ConfigMap created:"
echo "    - auth-service-config"
echo ""
echo "Next steps:"
echo "  1. Update your deployment.yaml to reference these secrets"
echo "  2. Deploy the auth-service"
echo "  3. Run database migrations"
