# ECR Module
# Creates Elastic Container Registry repositories for Docker images

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
# ECR Repositories
# ============================================================================
# One repository per microservice

# Auth Service Repository
resource "aws_ecr_repository" "auth_service" {
  name                 = "${var.project_name}/auth-service"
  image_tag_mutability = "MUTABLE"  # Allow overwriting tags
  
  image_scanning_configuration {
    scan_on_push = true  # Scan images for vulnerabilities
  }
  
  encryption_configuration {
    encryption_type = "AES256"
  }
  
  tags = {
    Name        = "${var.project_name}-auth-service"
    Environment = var.environment
    Service     = "auth-service"
  }
}

# Expense Service Repository
resource "aws_ecr_repository" "expense_service" {
  name                 = "${var.project_name}/expense-service"
  image_tag_mutability = "MUTABLE"
  
  image_scanning_configuration {
    scan_on_push = true
  }
  
  encryption_configuration {
    encryption_type = "AES256"
  }
  
  tags = {
    Name        = "${var.project_name}-expense-service"
    Environment = var.environment
    Service     = "expense-service"
  }
}

# Receipt Service Repository
resource "aws_ecr_repository" "receipt_service" {
  name                 = "${var.project_name}/receipt-service"
  image_tag_mutability = "MUTABLE"
  
  image_scanning_configuration {
    scan_on_push = true
  }
  
  encryption_configuration {
    encryption_type = "AES256"
  }
  
  tags = {
    Name        = "${var.project_name}-receipt-service"
    Environment = var.environment
    Service     = "receipt-service"
  }
}

# Notification Service Repository
resource "aws_ecr_repository" "notification_service" {
  name                 = "${var.project_name}/notification-service"
  image_tag_mutability = "MUTABLE"
  
  image_scanning_configuration {
    scan_on_push = true
  }
  
  encryption_configuration {
    encryption_type = "AES256"
  }
  
  tags = {
    Name        = "${var.project_name}/notification-service"
    Environment = var.environment
    Service     = "notification-service"
  }
}

# ============================================================================
# Lifecycle Policies
# ============================================================================
# Keep only the last N images to save storage costs

resource "aws_ecr_lifecycle_policy" "auth_service" {
  repository = aws_ecr_repository.auth_service.name
  
  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

resource "aws_ecr_lifecycle_policy" "expense_service" {
  repository = aws_ecr_repository.expense_service.name
  
  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

resource "aws_ecr_lifecycle_policy" "receipt_service" {
  repository = aws_ecr_repository.receipt_service.name
  
  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

resource "aws_ecr_lifecycle_policy" "notification_service" {
  repository = aws_ecr_repository.notification_service.name
  
  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection = {
          tagStatus   = "any"
          countType   = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      }
    ]
  })
}

