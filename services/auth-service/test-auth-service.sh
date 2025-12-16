#!/bin/bash
# Bash script to test auth-service endpoints
# Usage: ./test-auth-service.sh

BASE_URL="http://localhost:8080"

echo "🧪 Testing Auth Service"
echo "======================"
echo ""

# 1. Health Check
echo "1. Testing Health Check..."
if curl -s -X GET $BASE_URL/health | jq .; then
    echo "✅ Health check passed"
else
    echo "❌ Health check failed"
fi
echo ""

# 2. Register User
echo "2. Registering new user..."
REGISTER_RESPONSE=$(curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "testpassword123",
    "name": "Test User"
  }')

echo "$REGISTER_RESPONSE" | jq .

# Extract token
TOKEN=$(echo $REGISTER_RESPONSE | jq -r '.token')
if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
    echo "✅ Registration successful"
    echo "Token: $TOKEN"
else
    echo "❌ Registration failed"
fi
echo ""

# 3. Try duplicate registration
echo "3. Testing duplicate registration (should fail)..."
DUPLICATE_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "password123",
    "name": "Duplicate"
  }')

HTTP_CODE=$(echo "$DUPLICATE_RESPONSE" | tail -n1)
BODY=$(echo "$DUPLICATE_RESPONSE" | sed '$d')

echo "$BODY" | jq .
if [ "$HTTP_CODE" = "409" ]; then
    echo "✅ Correctly rejected duplicate (Status: $HTTP_CODE)"
else
    echo "❌ Should have returned 409 but got $HTTP_CODE"
fi
echo ""

# 4. Login
echo "4. Testing login..."
LOGIN_RESPONSE=$(curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "testpassword123"
  }')

echo "$LOGIN_RESPONSE" | jq .

LOGIN_TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')
if [ "$LOGIN_TOKEN" != "null" ] && [ -n "$LOGIN_TOKEN" ]; then
    echo "✅ Login successful"
    echo "Token: $LOGIN_TOKEN"
else
    echo "❌ Login failed"
fi
echo ""

# 5. Validate Token
echo "5. Validating token..."
if [ -n "$LOGIN_TOKEN" ]; then
    VALIDATE_RESPONSE=$(curl -s -X GET $BASE_URL/auth/validate \
      -H "Authorization: Bearer $LOGIN_TOKEN")
    
    echo "$VALIDATE_RESPONSE" | jq .
    
    IS_VALID=$(echo $VALIDATE_RESPONSE | jq -r '.valid')
    if [ "$IS_VALID" = "true" ]; then
        echo "✅ Token validation successful"
    else
        echo "❌ Token validation failed"
    fi
else
    echo "⚠️  Skipping - no token available"
fi
echo ""

# 6. Test invalid login
echo "6. Testing invalid login (should fail)..."
INVALID_LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "wrongpassword"
  }')

HTTP_CODE=$(echo "$INVALID_LOGIN_RESPONSE" | tail -n1)
BODY=$(echo "$INVALID_LOGIN_RESPONSE" | sed '$d')

echo "$BODY" | jq .
if [ "$HTTP_CODE" = "401" ]; then
    echo "✅ Correctly rejected invalid credentials (Status: $HTTP_CODE)"
else
    echo "❌ Should have returned 401 but got $HTTP_CODE"
fi
echo ""

# 7. Test missing Authorization header
echo "7. Testing validate without token (should fail)..."
NO_AUTH_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET $BASE_URL/auth/validate)

HTTP_CODE=$(echo "$NO_AUTH_RESPONSE" | tail -n1)
BODY=$(echo "$NO_AUTH_RESPONSE" | sed '$d')

echo "$BODY" | jq .
if [ "$HTTP_CODE" = "401" ]; then
    echo "✅ Correctly rejected request without token (Status: $HTTP_CODE)"
else
    echo "❌ Should have returned 401 but got $HTTP_CODE"
fi
echo ""

echo "✅ All tests completed!"
