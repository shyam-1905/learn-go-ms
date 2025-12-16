# Receipt Service API Testing Guide

This document provides comprehensive cURL commands to test all endpoints of the receipt-service.

## Prerequisites

1. **Auth Service Running**: The receipt-service requires JWT tokens from auth-service
2. **Receipt Service Running**: Start the service on port 8082 (default)
3. **AWS S3 Bucket**: Ensure the S3 bucket exists and credentials are configured
4. **Test File**: Have a sample image or PDF file ready for upload

## Step 1: Get JWT Token from Auth Service

First, register or login to get a JWT token:

```bash
# Register a new user (or use existing credentials)
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123",
    "name": "Test User"
  }'

# Login to get JWT token
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "password123"
  }'
```

**Save the token from the response** (you'll need it for all protected endpoints):
```bash
# Example response:
# {"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...","user":{"id":"...","email":"test@example.com"}}

# Set token as variable (PowerShell)
$TOKEN = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# Or (Bash)
export TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

## Step 2: Health Check

Test if the service is running:

```bash
curl -X GET http://localhost:8082/health
```

**Expected Response:**
```json
{
  "status": "healthy"
}
```

## Step 3: Upload Receipt

Upload a receipt file (image or PDF). The file must be:
- Maximum 10MB
- Format: JPEG, PNG, or PDF

### Upload Receipt (without expense_id)

```bash
# PowerShell
curl -X POST http://localhost:8082/receipts `
  -H "Authorization: Bearer $TOKEN" `
  -F "file=@/path/to/receipt.jpg"

# Bash
curl -X POST http://localhost:8082/receipts \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/receipt.jpg"
```

### Upload Receipt (with expense_id - link immediately)

```bash
# PowerShell
curl -X POST http://localhost:8082/receipts `
  -H "Authorization: Bearer $TOKEN" `
  -F "file=@/path/to/receipt.jpg" `
  -F "expense_id=YOUR_EXPENSE_ID_HERE"

# Bash
curl -X POST http://localhost:8082/receipts \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/receipt.jpg" \
  -F "expense_id=YOUR_EXPENSE_ID_HERE"
```

**Expected Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "expense_id": null,
  "file_name": "receipt.jpg",
  "file_url": "https://receipt-service-bucket.s3.amazonaws.com/...",
  "file_size": 245678,
  "mime_type": "image/jpeg",
  "merchant_name": null,
  "receipt_date": null,
  "total_amount": null,
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

**Save the receipt ID** from the response for subsequent tests.

## Step 4: Get Receipt by ID

Retrieve a specific receipt:

```bash
# PowerShell
curl -X GET http://localhost:8082/receipts/RECEIPT_ID_HERE `
  -H "Authorization: Bearer $TOKEN"

# Bash
curl -X GET http://localhost:8082/receipts/RECEIPT_ID_HERE \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "user_id": "123e4567-e89b-12d3-a456-426614174000",
  "expense_id": null,
  "file_name": "receipt.jpg",
  "file_url": "https://receipt-service-bucket.s3.amazonaws.com/...",
  "file_size": 245678,
  "mime_type": "image/jpeg",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

## Step 5: List All Receipts

Get all receipts for the authenticated user with pagination:

```bash
# PowerShell - List first page (default: 20 items)
curl -X GET "http://localhost:8082/receipts?page=1&limit=20" `
  -H "Authorization: Bearer $TOKEN"

# Bash
curl -X GET "http://localhost:8082/receipts?page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response:**
```json
{
  "receipts": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "file_name": "receipt.jpg",
      "file_url": "https://...",
      "file_size": 245678,
      "mime_type": "image/jpeg",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 20,
  "pages": 1
}
```

## Step 6: Get Receipts by Expense ID

Retrieve all receipts linked to a specific expense:

```bash
# PowerShell
curl -X GET "http://localhost:8082/receipts?expense_id=EXPENSE_ID_HERE" `
  -H "Authorization: Bearer $TOKEN"

# Bash
curl -X GET "http://localhost:8082/receipts?expense_id=EXPENSE_ID_HERE" \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response:**
```json
{
  "receipts": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "user_id": "123e4567-e89b-12d3-a456-426614174000",
      "expense_id": "EXPENSE_ID_HERE",
      "file_name": "receipt.jpg",
      "file_url": "https://...",
      "file_size": 245678,
      "mime_type": "image/jpeg",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "total": 1
}
```

## Step 7: Link Receipt to Expense

Link an existing receipt to an expense:

```bash
# PowerShell
curl -X PUT http://localhost:8082/receipts/RECEIPT_ID_HERE/link `
  -H "Authorization: Bearer $TOKEN" `
  -H "Content-Type: application/json" `
  -d '{
    "expense_id": "EXPENSE_ID_HERE"
  }'

# Bash
curl -X PUT http://localhost:8082/receipts/RECEIPT_ID_HERE/link \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "expense_id": "EXPENSE_ID_HERE"
  }'
```

**Expected Response:**
```json
{
  "message": "Receipt linked to expense successfully"
}
```

## Step 8: Delete Receipt

Soft delete a receipt (removes from S3 and marks as deleted):

```bash
# PowerShell
curl -X DELETE http://localhost:8082/receipts/RECEIPT_ID_HERE `
  -H "Authorization: Bearer $TOKEN"

# Bash
curl -X DELETE http://localhost:8082/receipts/RECEIPT_ID_HERE \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response:**
```json
{
  "message": "Receipt deleted successfully"
}
```

## Error Cases

### 1. Missing Authorization Header

```bash
curl -X GET http://localhost:8082/receipts
```

**Expected Response:** `401 Unauthorized`

### 2. Invalid Token

```bash
curl -X GET http://localhost:8082/receipts \
  -H "Authorization: Bearer invalid_token"
```

**Expected Response:** `401 Unauthorized`

### 3. File Too Large

```bash
# Upload a file larger than 10MB
curl -X POST http://localhost:8082/receipts \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/large_file.jpg"
```

**Expected Response:** `400 Bad Request` with error message about file size

### 4. Invalid File Type

```bash
# Upload a file that's not JPEG, PNG, or PDF
curl -X POST http://localhost:8082/receipts \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/path/to/document.txt"
```

**Expected Response:** `400 Bad Request` with error about file type

### 5. Receipt Not Found

```bash
curl -X GET http://localhost:8082/receipts/00000000-0000-0000-0000-000000000000 \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response:** `404 Not Found`

### 6. Accessing Another User's Receipt

```bash
# Try to access a receipt that belongs to a different user
curl -X GET http://localhost:8082/receipts/OTHER_USER_RECEIPT_ID \
  -H "Authorization: Bearer $TOKEN"
```

**Expected Response:** `404 Not Found` (for security, we don't reveal if it exists)

## Complete Test Flow Example

Here's a complete test flow combining all operations:

```bash
# 1. Get token
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' | jq -r '.token')

# 2. Upload receipt
RECEIPT_RESPONSE=$(curl -s -X POST http://localhost:8082/receipts \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@receipt.jpg")
RECEIPT_ID=$(echo $RECEIPT_RESPONSE | jq -r '.id')

# 3. Get receipt
curl -X GET http://localhost:8082/receipts/$RECEIPT_ID \
  -H "Authorization: Bearer $TOKEN"

# 4. List all receipts
curl -X GET "http://localhost:8082/receipts?page=1&limit=20" \
  -H "Authorization: Bearer $TOKEN"

# 5. Link to expense (if you have an expense_id)
EXPENSE_ID="your-expense-id-here"
curl -X PUT http://localhost:8082/receipts/$RECEIPT_ID/link \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"expense_id\":\"$EXPENSE_ID\"}"

# 6. Get receipts by expense
curl -X GET "http://localhost:8082/receipts?expense_id=$EXPENSE_ID" \
  -H "Authorization: Bearer $TOKEN"

# 7. Delete receipt
curl -X DELETE http://localhost:8082/receipts/$RECEIPT_ID \
  -H "Authorization: Bearer $TOKEN"
```

## Notes

1. **Presigned URLs**: The `file_url` in responses is a presigned S3 URL that expires after 1 hour. You'll need to call the API again to get a fresh URL.

2. **File Upload**: Make sure the file path is correct. Use absolute paths or paths relative to your current directory.

3. **Expense Integration**: To test linking receipts to expenses, you'll need to:
   - Create an expense using the expense-service
   - Get the expense ID
   - Use that ID when linking receipts

4. **S3 Bucket**: Ensure the S3 bucket exists and the AWS credentials have proper permissions:
   - `s3:PutObject` - for uploading files
   - `s3:GetObject` - for generating presigned URLs
   - `s3:DeleteObject` - for deleting files

5. **File Size Limit**: Maximum file size is 10MB. Larger files will be rejected.

6. **Supported Formats**: Only JPEG, PNG, and PDF files are accepted.

## Troubleshooting

- **Connection Refused**: Make sure the receipt-service is running on port 8082
- **401 Unauthorized**: Check that your JWT token is valid and not expired
- **S3 Errors**: Verify AWS credentials and S3 bucket permissions
- **File Upload Fails**: Check file size and format (must be JPEG, PNG, or PDF)
