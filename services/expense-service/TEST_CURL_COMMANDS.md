# Expense Service - cURL Test Commands

Complete set of cURL commands to test all endpoints of the expense-service.

## 🚀 Prerequisites

1. **Auth-service must be running** on `http://localhost:8080`
2. **Expense-service must be running** on `http://localhost:8081`
3. **Get a JWT token** from auth-service first

## 📋 Step 1: Get JWT Token

First, authenticate with auth-service to get a JWT token:

```bash
# Register a user (if not already registered)
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'

# Or login
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Save the token:**
```bash
# Linux/Mac
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')

echo "Token: $TOKEN"
```

**PowerShell (Windows):**
```powershell
$response = Invoke-RestMethod -Uri "http://localhost:8080/auth/login" `
  -Method POST `
  -ContentType "application/json" `
  -Body '{"email":"test@example.com","password":"password123"}'

$TOKEN = $response.token
Write-Host "Token: $TOKEN"
```

## 📋 Test Commands

### 1. Health Check

```bash
curl -X GET http://localhost:8081/health
```

**Expected Response:**
```json
{"status":"healthy"}
```

---

### 2. Create Expense

```bash
curl -X POST http://localhost:8081/expenses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "50.00",
    "description": "Lunch at restaurant",
    "category": "Food",
    "expense_date": "2024-01-15"
  }'
```

**Expected Response (201 Created):**
```json
{
  "id": "uuid-here",
  "user_id": "uuid-here",
  "amount": "50.00",
  "description": "Lunch at restaurant",
  "category": "Food",
  "expense_date": "2024-01-15T00:00:00Z",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Save expense ID for later:**
```bash
EXPENSE_ID=$(curl -s -X POST http://localhost:8081/expenses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"amount":"25.50","description":"Coffee","category":"Food","expense_date":"2024-01-15"}' \
  | jq -r '.id')
```

---

### 3. List Expenses

**Basic list (all expenses):**
```bash
curl -X GET http://localhost:8081/expenses \
  -H "Authorization: Bearer $TOKEN"
```

**With pagination:**
```bash
curl -X GET "http://localhost:8081/expenses?page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by category:**
```bash
curl -X GET "http://localhost:8081/expenses?category=Food" \
  -H "Authorization: Bearer $TOKEN"
```

**Filter by date range:**
```bash
curl -X GET "http://localhost:8081/expenses?start_date=2024-01-01&end_date=2024-01-31" \
  -H "Authorization: Bearer $TOKEN"
```

**Combined filters:**
```bash
curl -X GET "http://localhost:8081/expenses?category=Food&start_date=2024-01-01&end_date=2024-01-31&page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response:**
```json
{
  "expenses": [...],
  "total": 50,
  "page": 1,
  "limit": 20,
  "pages": 3
}
```

---

### 4. Get Single Expense

```bash
curl -X GET "http://localhost:8081/expenses/$EXPENSE_ID" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "id": "uuid-here",
  "user_id": "uuid-here",
  "amount": "50.00",
  "description": "Lunch at restaurant",
  "category": "Food",
  "expense_date": "2024-01-15T00:00:00Z",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

---

### 5. Update Expense

**Full update:**
```bash
curl -X PUT "http://localhost:8081/expenses/$EXPENSE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "60.00",
    "description": "Updated lunch description",
    "category": "Food",
    "expense_date": "2024-01-15"
  }'
```

**Partial update (only amount):**
```bash
curl -X PUT "http://localhost:8081/expenses/$EXPENSE_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "65.00"
  }'
```

---

### 6. Delete Expense

```bash
curl -X DELETE "http://localhost:8081/expenses/$EXPENSE_ID" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (200 OK):**
```json
{
  "message": "Expense deleted successfully"
}
```

---

### 7. Get Expense Summary

**Summary for all time:**
```bash
curl -X GET "http://localhost:8081/expenses/summary" \
  -H "Authorization: Bearer $TOKEN"
```

**Summary for date range:**
```bash
curl -X GET "http://localhost:8081/expenses/summary?start_date=2024-01-01&end_date=2024-01-31" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response:**
```json
{
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-31T00:00:00Z",
  "total": "1500.00",
  "by_category": [
    {
      "category": "Food",
      "total": "800.00",
      "count": 15
    },
    {
      "category": "Transport",
      "total": "700.00",
      "count": 10
    }
  ]
}
```

---

## 🧪 Error Cases

### Missing Authorization Header

```bash
curl -X GET http://localhost:8081/expenses
```

**Expected Response (401 Unauthorized):**
```
Authorization header required
```

### Invalid Token

```bash
curl -X GET http://localhost:8081/expenses \
  -H "Authorization: Bearer invalid.token.here"
```

**Expected Response (401 Unauthorized):**
```
Invalid or expired token
```

### Invalid Amount

```bash
curl -X POST http://localhost:8081/expenses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "-10.00",
    "description": "Test",
    "category": "Food",
    "expense_date": "2024-01-15"
  }'
```

**Expected Response (400 Bad Request):**
```json
{"error":"amount must be a positive number"}
```

### Invalid Date Format

```bash
curl -X POST http://localhost:8081/expenses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "50.00",
    "description": "Test",
    "category": "Food",
    "expense_date": "2024/01/15"
  }'
```

**Expected Response (400 Bad Request):**
```json
{"error":"expense_date must be in YYYY-MM-DD format"}
```

### Access Other User's Expense

```bash
# Try to access expense with different user's token
# This should fail even if you know the expense ID
curl -X GET "http://localhost:8081/expenses/other-user-expense-id" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response (404 Not Found):**
```json
{"error":"expense not found"}
```

---

## 🎯 Quick Reference

| Endpoint | Method | Auth Required | Description |
|----------|--------|---------------|-------------|
| `/health` | GET | No | Health check |
| `/expenses` | POST | Yes | Create expense |
| `/expenses` | GET | Yes | List expenses (with filters & pagination) |
| `/expenses/:id` | GET | Yes | Get single expense |
| `/expenses/:id` | PUT | Yes | Update expense |
| `/expenses/:id` | DELETE | Yes | Delete expense |
| `/expenses/summary` | GET | Yes | Get summary by category |

---

## 💡 Tips

1. **Always get token first** from auth-service before testing expense endpoints
2. **Use query parameters** for filtering and pagination
3. **Check ownership** - users can only access their own expenses
4. **Date format** must be YYYY-MM-DD
5. **Amount** must be positive number (as string)
6. **Check terminal logs** to see request/response details
