# PowerShell script to test auth-service endpoints
# Usage: .\test-auth-service.ps1

$BaseUrl = "http://localhost:8080"

Write-Host "🧪 Testing Auth Service" -ForegroundColor Cyan
Write-Host "======================" -ForegroundColor Cyan
Write-Host ""

# 1. Health Check
Write-Host "1. Testing Health Check..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "$BaseUrl/health" -Method GET
    $health | ConvertTo-Json
    Write-Host "✅ Health check passed" -ForegroundColor Green
} catch {
    Write-Host "❌ Health check failed: $_" -ForegroundColor Red
}
Write-Host ""

# 2. Register User
Write-Host "2. Registering new user..." -ForegroundColor Yellow
$registerBody = @{
    email = "testuser@example.com"
    password = "testpassword123"
    name = "Test User"
} | ConvertTo-Json

try {
    $registerResponse = Invoke-RestMethod -Uri "$BaseUrl/auth/register" `
        -Method POST `
        -ContentType "application/json" `
        -Body $registerBody

    $registerResponse | ConvertTo-Json
    $TOKEN = $registerResponse.token
    Write-Host "✅ Registration successful" -ForegroundColor Green
    Write-Host "Token: $TOKEN" -ForegroundColor Gray
} catch {
    Write-Host "❌ Registration failed: $_" -ForegroundColor Red
}
Write-Host ""

# 3. Try duplicate registration
Write-Host "3. Testing duplicate registration (should fail)..." -ForegroundColor Yellow
try {
    $duplicate = Invoke-RestMethod -Uri "$BaseUrl/auth/register" `
        -Method POST `
        -ContentType "application/json" `
        -Body $registerBody
    Write-Host "❌ Should have failed but didn't!" -ForegroundColor Red
    $duplicate | ConvertTo-Json
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "✅ Correctly rejected duplicate (Status: $statusCode)" -ForegroundColor Green
}
Write-Host ""

# 4. Login
Write-Host "4. Testing login..." -ForegroundColor Yellow
$loginBody = @{
    email = "testuser@example.com"
    password = "testpassword123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "$BaseUrl/auth/login" `
        -Method POST `
        -ContentType "application/json" `
        -Body $loginBody

    $loginResponse | ConvertTo-Json
    $LOGIN_TOKEN = $loginResponse.token
    Write-Host "✅ Login successful" -ForegroundColor Green
    Write-Host "Token: $LOGIN_TOKEN" -ForegroundColor Gray
} catch {
    Write-Host "❌ Login failed: $_" -ForegroundColor Red
}
Write-Host ""

# 5. Validate Token
Write-Host "5. Validating token..." -ForegroundColor Yellow
if ($LOGIN_TOKEN) {
    $headers = @{
        "Authorization" = "Bearer $LOGIN_TOKEN"
    }

    try {
        $validateResponse = Invoke-RestMethod -Uri "$BaseUrl/auth/validate" `
            -Method GET `
            -Headers $headers

        $validateResponse | ConvertTo-Json
        if ($validateResponse.valid) {
            Write-Host "✅ Token validation successful" -ForegroundColor Green
        } else {
            Write-Host "❌ Token validation failed" -ForegroundColor Red
        }
    } catch {
        Write-Host "❌ Token validation error: $_" -ForegroundColor Red
    }
} else {
    Write-Host "⚠️  Skipping - no token available" -ForegroundColor Yellow
}
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
    Write-Host "❌ Should have failed but didn't!" -ForegroundColor Red
    $invalidLogin | ConvertTo-Json
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "✅ Correctly rejected invalid credentials (Status: $statusCode)" -ForegroundColor Green
}
Write-Host ""

# 7. Test missing Authorization header
Write-Host "7. Testing validate without token (should fail)..." -ForegroundColor Yellow
try {
    $noAuth = Invoke-RestMethod -Uri "$BaseUrl/auth/validate" -Method GET
    Write-Host "❌ Should have failed but didn't!" -ForegroundColor Red
    $noAuth | ConvertTo-Json
} catch {
    $statusCode = $_.Exception.Response.StatusCode.value__
    Write-Host "✅ Correctly rejected request without token (Status: $statusCode)" -ForegroundColor Green
}
Write-Host ""

Write-Host "✅ All tests completed!" -ForegroundColor Green
