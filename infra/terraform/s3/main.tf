# S3 Module
# Creates S3 bucket for storing receipt images

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ============================================================================
# S3 Bucket for Receipt Images
# ============================================================================
resource "aws_s3_bucket" "receipts" {
  bucket = "${var.project_name}-receipts-${var.environment}-${data.aws_caller_identity.current.account_id}"
  
  tags = {
    Name        = "${var.project_name}-receipts"
    Environment = var.environment
    Service     = "receipt-service"
  }
}

# ============================================================================
# S3 Bucket Versioning
# ============================================================================
# Enable versioning to keep multiple versions of files
resource "aws_s3_bucket_versioning" "receipts" {
  bucket = aws_s3_bucket.receipts.id
  
  versioning_configuration {
    status = "Enabled"
  }
}

# ============================================================================
# S3 Bucket Encryption
# ============================================================================
# Encrypt data at rest using AWS managed keys
resource "aws_s3_bucket_server_side_encryption_configuration" "receipts" {
  bucket = aws_s3_bucket.receipts.id
  
  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# ============================================================================
# S3 Bucket Public Access Block
# ============================================================================
# Block all public access (security best practice)
resource "aws_s3_bucket_public_access_block" "receipts" {
  bucket = aws_s3_bucket.receipts.id
  
  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# ============================================================================
# S3 Bucket Lifecycle Policy
# ============================================================================
# Automatically delete old versions and incomplete multipart uploads
resource "aws_s3_bucket_lifecycle_configuration" "receipts" {
  bucket = aws_s3_bucket.receipts.id
  
  rule {
    id     = "delete-old-versions"
    status = "Enabled"
    
    noncurrent_version_expiration {
      noncurrent_days = 30  # Delete versions older than 30 days
    }
  }
  
  rule {
    id     = "delete-incomplete-uploads"
    status = "Enabled"
    
    abort_incomplete_multipart_upload {
      days_after_initiation = 7  # Delete incomplete uploads after 7 days
    }
  }
}

# ============================================================================
# S3 Bucket Policy
# ============================================================================
# Allow receipt-service to read/write to the bucket
# This will be updated when IRSA is configured
resource "aws_s3_bucket_policy" "receipts" {
  bucket = aws_s3_bucket.receipts.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowReceiptServiceAccess"
        Effect = "Allow"
        Principal = {
          AWS = var.receipt_service_role_arn != "" ? var.receipt_service_role_arn : "*"
        }
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          aws_s3_bucket.receipts.arn,
          "${aws_s3_bucket.receipts.arn}/*"
        ]
      }
    ]
  })
}

# ============================================================================
# Data Sources
# ============================================================================
data "aws_caller_identity" "current" {}

