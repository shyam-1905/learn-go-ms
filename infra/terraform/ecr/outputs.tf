# ECR Repository URLs
output "auth_service_repository_url" {
  description = "URL of auth-service ECR repository"
  value       = aws_ecr_repository.auth_service.repository_url
}

output "expense_service_repository_url" {
  description = "URL of expense-service ECR repository"
  value       = aws_ecr_repository.expense_service.repository_url
}

output "receipt_service_repository_url" {
  description = "URL of receipt-service ECR repository"
  value       = aws_ecr_repository.receipt_service.repository_url
}

output "notification_service_repository_url" {
  description = "URL of notification-service ECR repository"
  value       = aws_ecr_repository.notification_service.repository_url
}

# ECR Repository ARNs
output "auth_service_repository_arn" {
  description = "ARN of auth-service ECR repository"
  value       = aws_ecr_repository.auth_service.arn
}

output "expense_service_repository_arn" {
  description = "ARN of expense-service ECR repository"
  value       = aws_ecr_repository.expense_service.arn
}

output "receipt_service_repository_arn" {
  description = "ARN of receipt-service ECR repository"
  value       = aws_ecr_repository.receipt_service.arn
}

output "notification_service_repository_arn" {
  description = "ARN of notification-service ECR repository"
  value       = aws_ecr_repository.notification_service.arn
}

