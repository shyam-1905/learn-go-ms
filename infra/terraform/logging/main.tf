# Logging Module
# Creates CloudWatch Log Groups and IAM roles for Fluent Bit
# Fluent Bit will be deployed as a DaemonSet to collect logs from all pods

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
  }
}

# ============================================================================
# CloudWatch Log Groups
# ============================================================================
# One log group per service for organized log collection

resource "aws_cloudwatch_log_group" "auth_service" {
  name              = "/aws/eks/${var.project_name}/auth-service"
  retention_in_days = var.log_retention_days
  
  tags = {
    Name        = "${var.project_name}-auth-service-logs"
    Environment = var.environment
    Service     = "auth-service"
  }
}

resource "aws_cloudwatch_log_group" "expense_service" {
  name              = "/aws/eks/${var.project_name}/expense-service"
  retention_in_days = var.log_retention_days
  
  tags = {
    Name        = "${var.project_name}-expense-service-logs"
    Environment = var.environment
    Service     = "expense-service"
  }
}

resource "aws_cloudwatch_log_group" "receipt_service" {
  name              = "/aws/eks/${var.project_name}/receipt-service"
  retention_in_days = var.log_retention_days
  
  tags = {
    Name        = "${var.project_name}-receipt-service-logs"
    Environment = var.environment
    Service     = "receipt-service"
  }
}

resource "aws_cloudwatch_log_group" "notification_service" {
  name              = "/aws/eks/${var.project_name}/notification-service"
  retention_in_days = var.log_retention_days
  
  tags = {
    Name        = "${var.project_name}-notification-service-logs"
    Environment = var.environment
    Service     = "notification-service"
  }
}

# General cluster logs
resource "aws_cloudwatch_log_group" "cluster" {
  name              = "/aws/eks/${var.project_name}/cluster"
  retention_in_days = var.log_retention_days
  
  tags = {
    Name        = "${var.project_name}-cluster-logs"
    Environment = var.environment
  }
}

# ============================================================================
# IAM Role for Fluent Bit
# ============================================================================
# Fluent Bit needs permissions to write to CloudWatch Logs

resource "aws_iam_role" "fluent_bit" {
  name = "${var.project_name}-fluent-bit-role"
  
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
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:amazon-cloudwatch:fluent-bit"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })
  
  tags = {
    Name        = "${var.project_name}-fluent-bit-role"
    Environment = var.environment
  }
}

# Policy for Fluent Bit to write to CloudWatch Logs
resource "aws_iam_role_policy" "fluent_bit_logs" {
  name = "${var.project_name}-fluent-bit-logs-policy"
  role = aws_iam_role.fluent_bit.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents",
          "logs:DescribeLogStreams"
        ]
        Resource = [
          aws_cloudwatch_log_group.auth_service.arn,
          "${aws_cloudwatch_log_group.auth_service.arn}:*",
          aws_cloudwatch_log_group.expense_service.arn,
          "${aws_cloudwatch_log_group.expense_service.arn}:*",
          aws_cloudwatch_log_group.receipt_service.arn,
          "${aws_cloudwatch_log_group.receipt_service.arn}:*",
          aws_cloudwatch_log_group.notification_service.arn,
          "${aws_cloudwatch_log_group.notification_service.arn}:*",
          aws_cloudwatch_log_group.cluster.arn,
          "${aws_cloudwatch_log_group.cluster.arn}:*"
        ]
      }
    ]
  })
}

