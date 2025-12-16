# IAM Module
# Creates IAM roles for service accounts (IRSA) for each microservice
# This allows services to access AWS services without hardcoded credentials

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
# IAM Role for Auth Service
# ============================================================================
# Allows auth-service to publish events to SNS
resource "aws_iam_role" "auth_service" {
  name = "${var.project_name}-auth-service-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:auth-service"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })
  
  tags = {
    Name        = "${var.project_name}-auth-service-role"
    Environment = var.environment
    Service     = "auth-service"
  }
}

# Policy for auth-service to publish to SNS
resource "aws_iam_role_policy" "auth_service_sns" {
  name = "${var.project_name}-auth-service-sns-policy"
  role = aws_iam_role.auth_service.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sns:Publish"
        ]
        Resource = var.auth_events_topic_arn
      }
    ]
  })
}

# ============================================================================
# IAM Role for Expense Service
# ============================================================================
resource "aws_iam_role" "expense_service" {
  name = "${var.project_name}-expense-service-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:expense-service"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })
  
  tags = {
    Name        = "${var.project_name}-expense-service-role"
    Environment = var.environment
    Service     = "expense-service"
  }
}

resource "aws_iam_role_policy" "expense_service_sns" {
  name = "${var.project_name}-expense-service-sns-policy"
  role = aws_iam_role.expense_service.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sns:Publish"
        ]
        Resource = var.expense_events_topic_arn
      }
    ]
  })
}

# ============================================================================
# IAM Role for Receipt Service
# ============================================================================
resource "aws_iam_role" "receipt_service" {
  name = "${var.project_name}-receipt-service-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:receipt-service"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })
  
  tags = {
    Name        = "${var.project_name}-receipt-service-role"
    Environment = var.environment
    Service     = "receipt-service"
  }
}

# Policy for receipt-service to access S3 and SNS
resource "aws_iam_role_policy" "receipt_service_s3_sns" {
  name = "${var.project_name}-receipt-service-s3-sns-policy"
  role = aws_iam_role.receipt_service.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:GetObject",
          "s3:PutObject",
          "s3:DeleteObject",
          "s3:ListBucket"
        ]
        Resource = [
          var.s3_bucket_arn,
          "${var.s3_bucket_arn}/*"
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "sns:Publish"
        ]
        Resource = var.receipt_events_topic_arn
      }
    ]
  })
}

# ============================================================================
# IAM Role for Notification Service
# ============================================================================
resource "aws_iam_role" "notification_service" {
  name = "${var.project_name}-notification-service-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:notification-service"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })
  
  tags = {
    Name        = "${var.project_name}-notification-service-role"
    Environment = var.environment
    Service     = "notification-service"
  }
}

# Policy for notification-service to access SQS and SNS
resource "aws_iam_role_policy" "notification_service_sqs_sns" {
  name = "${var.project_name}-notification-service-sqs-sns-policy"
  role = aws_iam_role.notification_service.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "sqs:ReceiveMessage",
          "sqs:DeleteMessage",
          "sqs:GetQueueAttributes",
          "sqs:GetQueueUrl"
        ]
        Resource = [
          var.auth_events_queue_arn,
          var.expense_events_queue_arn,
          var.receipt_events_queue_arn
        ]
      },
      {
        Effect = "Allow"
        Action = [
          "sns:Publish",
          "sns:Subscribe",
          "sns:Unsubscribe"
        ]
        Resource = var.notification_email_topic_arn
      }
    ]
  })
}

# ============================================================================
# IAM Role for External Secrets Operator (optional)
# ============================================================================
# Allows External Secrets Operator to read from Secrets Manager
resource "aws_iam_role" "external_secrets" {
  count = var.enable_external_secrets ? 1 : 0
  
  name = "${var.project_name}-external-secrets-role"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:${var.namespace}:external-secrets"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })
  
  tags = {
    Name        = "${var.project_name}-external-secrets-role"
    Environment = var.environment
  }
}

resource "aws_iam_role_policy" "external_secrets_secrets_manager" {
  count = var.enable_external_secrets ? 1 : 0
  
  name = "${var.project_name}-external-secrets-policy"
  role = aws_iam_role.external_secrets[0].id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue",
          "secretsmanager:DescribeSecret"
        ]
        Resource = var.secrets_manager_arns
      }
    ]
  })
}

