# Messaging Module
# Creates SNS topics and SQS queues for event-driven architecture

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
# SNS Topics
# ============================================================================
# Topics for publishing events from services

# Auth Events Topic
resource "aws_sns_topic" "auth_events" {
  name = "${var.project_name}-auth-events-topic"
  
  tags = {
    Name        = "${var.project_name}-auth-events-topic"
    Environment = var.environment
    Service     = "auth-service"
  }
}

# Expense Events Topic
resource "aws_sns_topic" "expense_events" {
  name = "${var.project_name}-expense-events-topic"
  
  tags = {
    Name        = "${var.project_name}-expense-events-topic"
    Environment = var.environment
    Service     = "expense-service"
  }
}

# Receipt Events Topic
resource "aws_sns_topic" "receipt_events" {
  name = "${var.project_name}-receipt-events-topic"
  
  tags = {
    Name        = "${var.project_name}-receipt-events-topic"
    Environment = var.environment
    Service     = "receipt-service"
  }
}

# Notification Email Topic
resource "aws_sns_topic" "notification_email" {
  name = "${var.project_name}-notification-email-topic"
  
  tags = {
    Name        = "${var.project_name}-notification-email-topic"
    Environment = var.environment
    Service     = "notification-service"
  }
}

# ============================================================================
# SQS Queues
# ============================================================================
# Queues for consuming events (subscribed to SNS topics)

# Auth Events Queue
resource "aws_sqs_queue" "auth_events" {
  name                      = "${var.project_name}-auth-events-queue"
  message_retention_seconds = 345600  # 4 days
  receive_wait_time_seconds = 20      # Long polling
  
  tags = {
    Name        = "${var.project_name}-auth-events-queue"
    Environment = var.environment
    Service     = "notification-service"
  }
}

# Expense Events Queue
resource "aws_sqs_queue" "expense_events" {
  name                      = "${var.project_name}-expense-events-queue"
  message_retention_seconds = 345600
  receive_wait_time_seconds = 20
  
  tags = {
    Name        = "${var.project_name}-expense-events-queue"
    Environment = var.environment
    Service     = "notification-service"
  }
}

# Receipt Events Queue
resource "aws_sqs_queue" "receipt_events" {
  name                      = "${var.project_name}-receipt-events-queue"
  message_retention_seconds = 345600
  receive_wait_time_seconds = 20
  
  tags = {
    Name        = "${var.project_name}-receipt-events-queue"
    Environment = var.environment
    Service     = "notification-service"
  }
}

# ============================================================================
# SNS to SQS Subscriptions
# ============================================================================
# Subscribe SQS queues to SNS topics so events flow: Service → SNS → SQS → Notification Service

resource "aws_sns_topic_subscription" "auth_events_to_queue" {
  topic_arn = aws_sns_topic.auth_events.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.auth_events.arn
}

resource "aws_sns_topic_subscription" "expense_events_to_queue" {
  topic_arn = aws_sns_topic.expense_events.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.expense_events.arn
}

resource "aws_sns_topic_subscription" "receipt_events_to_queue" {
  topic_arn = aws_sns_topic.receipt_events.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.receipt_events.arn
}

# ============================================================================
# SQS Queue Policies
# ============================================================================
# Allow SNS topics to send messages to queues

resource "aws_sqs_queue_policy" "auth_events" {
  queue_url = aws_sqs_queue.auth_events.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action   = "sqs:SendMessage"
        Resource = aws_sqs_queue.auth_events.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.auth_events.arn
          }
        }
      }
    ]
  })
}

resource "aws_sqs_queue_policy" "expense_events" {
  queue_url = aws_sqs_queue.expense_events.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action   = "sqs:SendMessage"
        Resource = aws_sqs_queue.expense_events.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.expense_events.arn
          }
        }
      }
    ]
  })
}

resource "aws_sqs_queue_policy" "receipt_events" {
  queue_url = aws_sqs_queue.receipt_events.id
  
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "sns.amazonaws.com"
        }
        Action   = "sqs:SendMessage"
        Resource = aws_sqs_queue.receipt_events.arn
        Condition = {
          ArnEquals = {
            "aws:SourceArn" = aws_sns_topic.receipt_events.arn
          }
        }
      }
    ]
  })
}

