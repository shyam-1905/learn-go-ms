# Auth Service - cURL Test Commands

Complete set of cURL commands to test all endpoints of the auth-service.

## 🚀 Prerequisites

1. Start the auth-service:
   ```powershell
   cd services/auth-service
   .\set-env.ps1
   go run cmd/main.go
   ```

2. The service should be running on `http://localhost:8080`

## 📋 Test Commands

### 1. Health Check

Test if the service is running:

```bash
curl -X GET http://localhost:8080/health
```

**Expected Response:**
```json
{"status":"healthy"}
```

**With verbose output (to see headers):**
```bash
curl -v -X GET http://localhost:8080/health
```

---

### 2. Register a New User

Create a new user account:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "securepassword123",
    "name": "Test User"
  }'
```

**Expected Response (201 Created):**
```json
{
  "user_id": "uuid-here",
  "email": "test@example.com",
  "name": "Test User",
  "token": "jwt-token-here"
}
```

**Save the token for later tests:**
```bash
# Save response to variable (Linux/Mac)
TOKEN=$(curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test2@example.com","password":"password123","name":"Test User 2"}' \
  | jq -r '.token')

echo "Token: $TOKEN"
```

**PowerShell (Windows):**
```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8080/auth/register" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"email":"test2@example.com","password":"password123","name":"Test User 2"}'

$TOKEN = $response.token
Write-Host "Token: $TOKEN"
```

---

### 3. Register with Invalid Data (Error Cases)

**Missing email:**
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "password": "password123",
    "name": "Test User"
  }'
```

**Expected Response (400 Bad Request):**
```json
{"error":"Email, password, and name are required"}
```

**Invalid JSON:**
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d 'invalid json'
```

**Expected Response (400 Bad Request):**
```json
{"error":"Invalid request body"}
```

---

### 4. Register Duplicate User (Error Case)

Try to register the same email twice:

```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Duplicate User"
  }'
```

**Expected Response (409 Conflict):**
```json
{"error":"user with this email already exists"}
```

---

### 5. Login with Valid Credentials

Login with the user you just created:

```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "securepassword123"
  }'
```

**Expected Response (200 OK):**
```json
{
  "user_id": "uuid-here",
  "email": "test@example.com",
  "name": "Test User",
  "token": "jwt-token-here"
}
```

**Save token for validation:**
```bash
# Linux/Mac
LOGIN_TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"securepassword123"}' \
  | jq -r '.token')

echo "Login Token: $LOGIN_TOKEN"
```

---

### 6. Login with Invalid Credentials (Error Cases)

**Wrong password:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "wrongpassword"
  }'
```

**Expected Response (401 Unauthorized):**
```json
{"error":"Invalid email or password"}
```

**Non-existent user:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "nonexistent@example.com",
    "password": "password123"
  }'
```

**Expected Response (401 Unauthorized):**
```json
{"error":"Invalid email or password"}
```

**Missing fields:**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com"
  }'
```

**Expected Response (400 Bad Request):**
```json
{"error":"Email and password are required"}
```

---

### 7. Validate Token

Validate a JWT token (replace `YOUR_TOKEN` with actual token):

```bash
curl -X GET http://localhost:8080/auth/validate \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**Example with saved token:**
```bash
# Linux/Mac (using token from login)
curl -X GET http://localhost:8080/auth/validate \
  -H "Authorization: Bearer $LOGIN_TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "valid": true,
  "user_id": "uuid-here",
  "email": "test@example.com"
}
```

**PowerShell (Windows):**
```powershell
$headers = @{
    "Authorization" = "Bearer $TOKEN"
}

Invoke-RestMethod -Uri "http://localhost:8080/auth/validate" `
  -Method GET `
  -Headers $headers
```

---

### 8. Validate Token - Error Cases

**Missing Authorization header:**
```bash
curl -X GET http://localhost:8080/auth/validate
```

**Expected Response (401 Unauthorized):**
```json
{"error":"Authorization header required"}
```

**Invalid token format:**
```bash
curl -X GET http://localhost:8080/auth/validate \
  -H "Authorization: InvalidFormat token"
```

**Expected Response (401 Unauthorized):**
```json
{"error":"Invalid authorization header format"}
```

**Expired or invalid token:**
```bash
curl -X GET http://localhost:8080/auth/validate \
  -H "Authorization: Bearer invalid.token.here"
```

**Expected Response (401 Unauthorized):**
```json
{"error":"Invalid token"}
```

---

## 🧪 Complete Test Script

### Bash Script (Linux/Mac)

Create `test-auth-service.sh`:

```bash
#!/bin/bash

BASE_URL="http://localhost:8080"

echo "🧪 Testing Auth Service"
echo "======================"
echo ""

# 1. Health Check
echo "1. Testing Health Check..."
curl -s -X GET $BASE_URL/health | jq .
echo ""
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
echo ""
echo "Token: $TOKEN"
echo ""
echo ""

# 3. Try duplicate registration
echo "3. Testing duplicate registration (should fail)..."
curl -s -X POST $BASE_URL/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "password123",
    "name": "Duplicate"
  }' | jq .
echo ""
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
echo ""
echo ""

# 5. Validate Token
echo "5. Validating token..."
curl -s -X GET $BASE_URL/auth/validate \
  -H "Authorization: Bearer $LOGIN_TOKEN" | jq .
echo ""
echo ""

# 6. Test invalid login
echo "6. Testing invalid login (should fail)..."
curl -s -X POST $BASE_URL/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser@example.com",
    "password": "wrongpassword"
  }' | jq .
echo ""
echo ""

echo "✅ All tests completed!"
```

**Make it executable and run:**
```bash
chmod +x test-auth-service.sh
./test-auth-service.sh
```

---

### PowerShell Script (Windows)

Create `test-auth-service.ps1`:

```powershell
$BaseUrl = "http://localhost:8080"

Write-Host "🧪 Testing Auth Service" -ForegroundColor Cyan
Write-Host "======================" -ForegroundColor Cyan
Write-Host ""

# 1. Health Check
Write-Host "1. Testing Health Check..." -ForegroundColor Yellow
$health = Invoke-RestMethod -Uri "$BaseUrl/health" -Method GET
$health | ConvertTo-Json
Write-Host ""

# 2. Register User
Write-Host "2. Registering new user..." -ForegroundColor Yellow
$registerBody = @{
    email = "testuser@example.com"
    password = "testpassword123"
    name = "Test User"
} | ConvertTo-Json

$registerResponse = Invoke-RestMethod -Uri "$BaseUrl/auth/register" `
    -Method POST `
    -ContentType "application/json" `
    -Body $registerBody

$registerResponse | ConvertTo-Json
$TOKEN = $registerResponse.token
Write-Host "Token: $TOKEN" -ForegroundColor Green
Write-Host ""

# 3. Try duplicate registration
Write-Host "3. Testing duplicate registration (should fail)..." -ForegroundColor Yellow
try {
    $duplicate = Invoke-RestMethod -Uri "$BaseUrl/auth/register" `
        -Method POST `
        -ContentType "application/json" `
        -Body $registerBody
    $duplicate | ConvertTo-Json
} catch {
    $_.Exception.Response | ConvertTo-Json
}
Write-Host ""

# 4. Login
Write-Host "4. Testing login..." -ForegroundColor Yellow
$loginBody = @{
    email = "testuser@example.com"
    password = "testpassword123"
} | ConvertTo-Json

$loginResponse = Invoke-RestMethod -Uri "$BaseUrl/auth/login" `
    -Method POST `
    -ContentType "application/json" `
    -Body $loginBody

$loginResponse | ConvertTo-Json
$LOGIN_TOKEN = $loginResponse.token
Write-Host ""

# 5. Validate Token
Write-Host "5. Validating token..." -ForegroundColor Yellow
$headers = @{
    "Authorization" = "Bearer $LOGIN_TOKEN"
}

$validateResponse = Invoke-RestMethod -Uri "$BaseUrl/auth/validate" `
    -Method GET `
    -Headers $headers

$validateResponse | ConvertTo-Json
Write-Host ""

# 6. Test invalid login
Write-Host "6. Testing invalid login (should fail)..." -ForegroundColor Yellow
$invalidLoginBody = @{
    email = "testuser@example.com"
    password = "wrongpassword"
} | ConvertTo-Json

try {
    $invalidLogin = Invoke-RestMethod -Uri "$BaseUrl/auth/login" `
        -Method POST `
        -ContentType "application/json" `
        -Body $invalidLoginBody
    $invalidLogin | ConvertTo-Json
} catch {
    Write-Host "Error (expected): $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

Write-Host "✅ All tests completed!" -ForegroundColor Green
```

**Run it:**
```powershell
.\test-auth-service.ps1
```

---

## 📊 Expected Log Output

When you run these commands, you should see logs in your terminal like:

```
➡️  [REQUEST] GET /health from 127.0.0.1:xxxxx
⬅️  [RESPONSE] GET /health - Status: 200 - Duration: 1.234ms

➡️  [REQUEST] POST /auth/register from 127.0.0.1:xxxxx
⬅️  [RESPONSE] POST /auth/register - Status: 201 - Duration: 45.678ms

➡️  [REQUEST] POST /auth/login from 127.0.0.1:xxxxx
⬅️  [RESPONSE] POST /auth/login - Status: 200 - Duration: 23.456ms

➡️  [REQUEST] GET /auth/validate from 127.0.0.1:xxxxx
⬅️  [RESPONSE] GET /auth/validate - Status: 200 - Duration: 12.345ms
```

---

## 🎯 Quick Reference

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/health` | GET | No | Health check |
| `/auth/register` | POST | No | Register new user |
| `/auth/login` | POST | No | Login user |
| `/auth/validate` | GET | Yes (Bearer token) | Validate JWT token |

---

## 💡 Tips

1. **Use `-v` flag for verbose output** to see HTTP headers:
   ```bash
   curl -v -X GET http://localhost:8080/health
   ```

2. **Use `jq` to format JSON responses** (install: `brew install jq` or `apt-get install jq`):
   ```bash
   curl -s http://localhost:8080/health | jq .
   ```

3. **Save tokens to variables** for easier testing:
   ```bash
   TOKEN="your-token-here"
   curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/auth/validate
   ```

4. **Test error cases** to ensure proper error handling

5. **Check the terminal logs** to see request/response details
