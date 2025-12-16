# Authentication Architecture - Centralized Token Validation

## Overview

The expense-service now uses **centralized authentication** by calling auth-service to validate JWT tokens, instead of validating tokens locally with a shared secret.

## Architecture Change

### Before (Shared Secret Approach)
- Service had its own JWT validation logic
- Required sharing `JWT_SECRET` with auth-service
- Risk of secret mismatch causing "invalid signature" errors
- No single source of truth for authentication

### After (Centralized Validation)
- Service calls auth-service `/auth/validate` endpoint
- Auth-service is the single source of truth for token validation
- No need to share JWT_SECRET
- Better separation of concerns

## How It Works

1. Client sends request with JWT token to expense-service
2. Service middleware extracts token from Authorization header
3. Service calls auth-service `/auth/validate` endpoint with the token
4. Auth-service validates token and returns user_id and email
5. Service attaches user info to request context and proceeds

## Configuration Changes

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

The service now has an `AuthClient` that:
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
2. **No Secret Sharing**: Expense-service doesn't need JWT_SECRET
3. **Easier Updates**: Token validation logic changes only in auth-service
4. **Better Security**: Secrets are centralized in auth-service
5. **Consistent Validation**: All services use the same validation logic

## Environment Variables

```bash
AUTH_SERVICE_URL=http://localhost:8080  # URL of auth-service
```

## Testing

Make sure auth-service is running before starting expense-service:

```bash
# Terminal 1: Start auth-service
cd services/auth-service
go run cmd/main.go

# Terminal 2: Start expense-service
cd services/expense-service
. .\set-env.ps1
go run cmd/main.go
```

## Troubleshooting

### "Invalid signature" Error
- **Old Issue**: JWT_SECRET mismatch
- **New Solution**: Ensure auth-service is running and AUTH_SERVICE_URL is correct

### "Failed to call auth-service" Error
- Check that auth-service is running
- Verify AUTH_SERVICE_URL is correct
- Check network connectivity between services
