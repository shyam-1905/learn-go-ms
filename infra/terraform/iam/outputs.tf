# IAM Role ARNs for IRSA
output "auth_service_role_arn" {
  description = "ARN of IAM role for auth-service"
  value       = aws_iam_role.auth_service.arn
}

output "expense_service_role_arn" {
  description = "ARN of IAM role for expense-service"
  value       = aws_iam_role.expense_service.arn
}

output "receipt_service_role_arn" {
  description = "ARN of IAM role for receipt-service"
  value       = aws_iam_role.receipt_service.arn
}

output "notification_service_role_arn" {
  description = "ARN of IAM role for notification-service"
  value       = aws_iam_role.notification_service.arn
}

output "external_secrets_role_arn" {
  description = "ARN of IAM role for External Secrets Operator (if enabled)"
  value       = var.enable_external_secrets ? aws_iam_role.external_secrets[0].arn : null
}

