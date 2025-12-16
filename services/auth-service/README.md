# Auth Service

A complete authentication microservice built with Go, featuring JWT-based authentication, user management, and PostgreSQL integration.

## 🎓 What You've Learned

This service demonstrates:

- **Go Structs & Tags**: Using struct tags for JSON and database mapping
- **Interfaces**: Clean architecture with repository pattern
- **Error Handling**: Explicit error returns (no exceptions)
- **Context**: Request cancellation and timeouts
- **HTTP Handlers**: REST API endpoints
- **Middleware**: JWT authentication middleware
- **Database**: PostgreSQL with connection pooling
- **Security**: Bcrypt password hashing, JWT tokens
- **Configuration**: Environment variable management

## 📁 Project Structure

```
auth-service/
├── cmd/
│   └── main.go              # Application entry point
├── internal/
│   ├── model/               # Data models (User, DTOs)
│   ├── repository/          # Data access layer
│   ├── service/             # Business logic (Auth, JWT)
│   ├── handler/             # HTTP handlers
│   ├── middleware/          # JWT middleware
│   └── config/              # Configuration management
├── migrations/              # Database migration scripts
└── go.mod                   # Go module dependencies
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database (local or AWS RDS)

### Quick Start: Testing with AWS RDS Demo Database

**For quick testing**, you can create a demo RDS database in AWS and test locally:

1. **Create RDS in AWS Console** (see `TESTING_LOCALLY.md` for detailed steps)
2. **Get connection details**: endpoint, username, password
3. **Set environment variables** (see below)
4. **Run migrations**: `psql -h <endpoint> -U postgres -d auth_db -f migrations/001_create_users_table.sql`
5. **Run service**: `go run cmd/main.go`

📖 **Full guide**: See [TESTING_LOCALLY.md](./TESTING_LOCALLY.md)

### 1. Set Up Environment Variables

Create a `.env` file or set these environment variables:

```bash
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=auth_db

# JWT Configuration (IMPORTANT: Use a strong secret in production!)
JWT_SECRET=your-super-secret-jwt-key
JWT_EXPIRATION_HOURS=24

# Server Configuration
SERVER_PORT=8080
```

### 2. Run Database Migration

Connect to your PostgreSQL database and run:

```bash
psql -U postgres -d auth_db -f migrations/001_create_users_table.sql
```

Or manually execute the SQL in `migrations/001_create_users_table.sql`.

### 3. Run the Service

```bash
# From the auth-service directory
go run cmd/main.go
```

The server will start on port 8080 (or your configured port).

## 📡 API Endpoints

### Health Check
```http
GET /health
```

### Register User
```http
POST /auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123",
  "name": "John Doe"
}
```

**Response:**
```json
{
  "user_id": "uuid-here",
  "email": "user@example.com",
  "name": "John Doe",
  "token": "jwt-token-here"
}
```

### Login
```http
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepassword123"
}
```

**Response:** Same as register (includes JWT token)

### Validate Token
```http
GET /auth/validate
Authorization: Bearer <your-jwt-token>
```

**Response:**
```json
{
  "valid": true,
  "user_id": "uuid-here",
  "email": "user@example.com"
}
```

## 🧪 Testing with cURL

### Register a user:
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","name":"Test User"}'
```

### Login:
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

### Validate token (replace TOKEN with actual token):
```bash
curl -X GET http://localhost:8080/auth/validate \
  -H "Authorization: Bearer TOKEN"
```

## 🏗️ Architecture

This service follows **Clean Architecture** principles:

1. **Models** (`internal/model/`): Data structures
2. **Repository** (`internal/repository/`): Data access (database operations)
3. **Service** (`internal/service/`): Business logic
4. **Handler** (`internal/handler/`): HTTP layer (API endpoints)
5. **Middleware** (`internal/middleware/`): Cross-cutting concerns (auth)

Each layer depends only on inner layers, making the code:
- **Testable**: Easy to mock dependencies
- **Maintainable**: Clear separation of concerns
- **Flexible**: Easy to swap implementations

## 🔐 Security Features

- **Password Hashing**: Bcrypt with automatic salting
- **JWT Tokens**: Secure token-based authentication
- **SQL Injection Prevention**: Parameterized queries
- **Soft Deletes**: Users marked as deleted, not removed
- **Token Expiration**: Configurable token lifetime

## 📚 Key Go Concepts Used

- **Structs**: Data modeling
- **Interfaces**: Abstraction and testability
- **Error Handling**: Explicit error returns
- **Context**: Request lifecycle management
- **Goroutines**: Concurrent server handling
- **Channels**: Signal handling
- **Pointers**: Efficient data passing
- **Tags**: Metadata for serialization

## 🎯 Next Steps

Try these enhancements:

1. Add email validation
2. Implement password strength requirements
3. Add rate limiting
4. Create protected routes using the middleware
5. Add logging (use `log` or `zap`)
6. Write unit tests
7. Add request validation library
8. Implement refresh tokens

## 📖 Learning Resources

- [Go Tour](https://go.dev/tour/)
- [Effective Go](https://go.dev/doc/effective_go)
- [Go by Example](https://gobyexample.com/)
