# Authentication Architecture - Centralized Token Validation

## Overview

The receipt-service and expense-service now use **centralized authentication** by calling auth-service to validate JWT tokens, instead of validating tokens locally with a shared secret.

## Architecture Change

### Before (Shared Secret Approach)
- Each service had its own JWT validation logic
- Required sharing `JWT_SECRET` across all services
- Risk of secret mismatch causing "invalid signature" errors
- No single source of truth for authentication

### After (Centralized Validation)
- All services call auth-service `/auth/validate` endpoint
- Auth-service is the single source of truth for token validation
- No need to share JWT_SECRET with other services
- Better separation of concerns

## How It Works

```
┌─────────────┐         ┌──────────────┐         ┌─────────────┐
│   Client    │────────▶│ Receipt/Exp │────────▶│ Auth Service│
│             │  Token  │   Service    │  HTTP   │             │
└─────────────┘         └──────────────┘  Call   └─────────────┘
                              │                    │
                              │                    │
                              ▼                    ▼
                        Validate Token      Return User Info
```

1. Client sends request with JWT token to receipt-service or expense-service
2. Service middleware extracts token from Authorization header
3. Service calls auth-service `/auth/validate` endpoint with the token
4. Auth-service validates token and returns user_id and email
5. Service attaches user info to request context and proceeds

## Configuration Changes

### Receipt Service

**Old Configuration:**
```bash
JWT_SECRET=your-secret-key
```

**New Configuration:**
```bash
AUTH_SERVICE_URL=http://localhost:8080
```

### Expense Service

**Old Configuration:**
```bash
JWT_SECRET=your-secret-key
```

**New Configuration:**
```bash
AUTH_SERVICE_URL=http://localhost:8080
```

## Implementation Details

### Auth Client Service

Both services now have an `AuthClient` that:
- Makes HTTP calls to auth-service
- Handles timeouts (5 seconds)
- Parses validation responses
- Returns user_id and email on success

**File:** `internal/service/auth_client.go`

### Updated Middleware

The authentication middleware now:
- Uses `AuthClient` instead of `JWTService`
- Calls auth-service for validation
- Handles network errors gracefully

**File:** `internal/middleware/auth_middleware.go`

## Benefits

1. **Single Source of Truth**: Auth-service is the only service that validates tokens
2. **No Secret Sharing**: Other services don't need JWT_SECRET
3. **Easier Updates**: Token validation logic changes only in auth-service
4. **Better Security**: Secrets are centralized in auth-service
5. **Consistent Validation**: All services use the same validation logic

## Trade-offs

1. **Network Latency**: Each request requires an HTTP call to auth-service
   - Mitigation: 5-second timeout, auth-service should be fast
2. **Dependency**: Services depend on auth-service being available
   - Mitigation: Proper error handling and monitoring

## Environment Variables

### Receipt Service
```bash
AUTH_SERVICE_URL=http://localhost:8080  # URL of auth-service
```

### Expense Service
```bash
AUTH_SERVICE_URL=http://localhost:8080  # URL of auth-service
```

## Testing

Make sure auth-service is running before starting receipt-service or expense-service:

```bash
# Terminal 1: Start auth-service
cd services/auth-service
go run cmd/main.go

# Terminal 2: Start expense-service
cd services/expense-service
. .\set-env.ps1
go run cmd/main.go

# Terminal 3: Start receipt-service
cd services/receipt-service
. .\set-env.ps1
go run cmd/main.go
```

## Troubleshooting

### "Invalid signature" Error
- **Old Issue**: JWT_SECRET mismatch between services
- **New Solution**: Ensure auth-service is running and AUTH_SERVICE_URL is correct

### "Failed to call auth-service" Error
- Check that auth-service is running
- Verify AUTH_SERVICE_URL is correct
- Check network connectivity between services

### Timeout Errors
- Auth-service might be slow or unavailable
- Check auth-service logs
- Verify auth-service health endpoint
