# Expense Service

A complete expense tracking microservice built with Go, featuring CRUD operations, filtering, pagination, and JWT authentication integration.

## 🎓 What You've Learned

This service demonstrates:

- **JWT Authentication**: Validating tokens from auth-service
- **Context Usage**: Passing user_id through request context
- **Query Parameters**: Parsing URL query strings for filtering
- **Pagination**: LIMIT/OFFSET implementation
- **Date Filtering**: SQL date range queries
- **Ownership Validation**: Ensuring users only access their own data
- **Clean Architecture**: Repository, Service, Handler layers

## 📁 Project Structure

```
expense-service/
├── cmd/
│   ├── main.go              # Application entry point
│   └── migrate.go           # Database migration tool
├── internal/
│   ├── model/               # Data models (Expense, DTOs)
│   ├── repository/          # Data access layer
│   ├── service/             # Business logic (Expense, JWT validation)
│   ├── handler/             # HTTP handlers
│   ├── middleware/          # JWT middleware, logging
│   └── config/              # Configuration management
├── migrations/              # Database migration scripts
└── go.mod                   # Go module dependencies
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database (local or AWS RDS)
- Auth-service running (for JWT token generation)

### 1. Set Up Environment Variables

Set these environment variables (or use a script):

```bash
# Database Configuration
DB_HOST=your-rds-endpoint.rds.amazonaws.com
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=expense_db

# JWT Configuration (MUST match auth-service secret!)
JWT_SECRET=your-super-secret-jwt-key

# Server Configuration
SERVER_PORT=8081
```

### 2. Run Database Migration

```bash
# Set environment variables first
go run cmd/migrate.go
```

This will:
- Create the database if it doesn't exist
- Create the expenses table
- Create indexes for performance

### 3. Run the Service

```bash
go run cmd/main.go
```

The server will start on port 8081 (or your configured port).

## 🔐 Authentication Flow

1. **User authenticates with auth-service:**
   ```bash
   POST http://localhost:8080/auth/login
   Response: { token: "jwt-token", user_id: "uuid" }
   ```

2. **User calls expense-service with token:**
   ```bash
   POST http://localhost:8081/expenses
   Headers: Authorization: Bearer <jwt-token>
   Body: { amount: "100.50", description: "Lunch", category: "Food", expense_date: "2024-01-15" }
   ```

3. **Expense-service validates token:**
   - Extracts token from Authorization header
   - Validates JWT signature using shared secret
   - Extracts user_id from JWT claims
   - Attaches user_id to request context
   - Processes request with authenticated user_id

## 📡 API Endpoints

### Public Endpoints

```
GET /health
```

### Protected Endpoints (Require JWT Token)

#### Create Expense
```http
POST /expenses
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": "100.50",
  "description": "Lunch at restaurant",
  "category": "Food",
  "expense_date": "2024-01-15"
}
```

#### List Expenses
```http
GET /expenses?category=Food&start_date=2024-01-01&end_date=2024-01-31&page=1&limit=20
Authorization: Bearer <token>
```

**Query Parameters:**
- `category` - Filter by category (optional)
- `start_date` - Filter from date YYYY-MM-DD (optional)
- `end_date` - Filter to date YYYY-MM-DD (optional)
- `page` - Page number (default: 1)
- `limit` - Items per page (default: 20, max: 100)

#### Get Single Expense
```http
GET /expenses/:id
Authorization: Bearer <token>
```

#### Update Expense
```http
PUT /expenses/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "amount": "120.00",
  "description": "Updated description"
}
```

#### Delete Expense
```http
DELETE /expenses/:id
Authorization: Bearer <token>
```

#### Get Expense Summary
```http
GET /expenses/summary?start_date=2024-01-01&end_date=2024-01-31
Authorization: Bearer <token>
```

**Response:**
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

## 🧪 Testing with cURL

### 1. Get JWT Token from Auth Service

```bash
# Login to get token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  | jq -r '.token')
```

### 2. Create an Expense

```bash
curl -X POST http://localhost:8081/expenses \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "amount": "50.00",
    "description": "Coffee",
    "category": "Food",
    "expense_date": "2024-01-15"
  }'
```

### 3. List Expenses

```bash
curl -X GET "http://localhost:8081/expenses?category=Food&page=1&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

### 4. Get Expense Summary

```bash
curl -X GET "http://localhost:8081/expenses/summary?start_date=2024-01-01&end_date=2024-01-31" \
  -H "Authorization: Bearer $TOKEN"
```

## 🏗️ Architecture

This service follows **Clean Architecture** principles:

1. **Models** (`internal/model/`): Data structures
2. **Repository** (`internal/repository/`): Data access (database operations)
3. **Service** (`internal/service/`): Business logic
4. **Handler** (`internal/handler/`): HTTP layer (API endpoints)
5. **Middleware** (`internal/middleware/`): Cross-cutting concerns (auth, logging)

## 🔐 Security Features

- **JWT Authentication**: All endpoints (except /health) require valid JWT token
- **Ownership Validation**: Users can only access their own expenses
- **SQL Injection Prevention**: Parameterized queries
- **Soft Deletes**: Expenses marked as deleted, not removed
- **Input Validation**: Amount, date, and category validation

## 📚 Key Go Concepts Used

- **Context**: Passing user_id through request context
- **Query Parameters**: Parsing URL query strings
- **Pagination**: LIMIT/OFFSET SQL implementation
- **Date Parsing**: time.Parse for date validation
- **Decimal Handling**: String representation for currency
- **Ownership Checks**: Verifying user owns resource before operations

## 🎯 Next Steps

Try these enhancements:

1. Add batch expense creation (concurrency example)
2. Add expense categories enum/table
3. Add currency support
4. Add receipt_id field (for receipt-service integration)
5. Add expense tags
6. Add export functionality (CSV/JSON)
7. Add expense search by description

## 📖 Learning Resources

- [Go Context Package](https://pkg.go.dev/context)
- [Gorilla Mux Router](https://github.com/gorilla/mux)
- [PostgreSQL Date Functions](https://www.postgresql.org/docs/current/functions-datetime.html)
