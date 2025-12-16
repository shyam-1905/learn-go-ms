# CloudWatch Log Groups
output "auth_service_log_group" {
  description = "CloudWatch Log Group name for auth-service"
  value       = aws_cloudwatch_log_group.auth_service.name
}

output "expense_service_log_group" {
  description = "CloudWatch Log Group name for expense-service"
  value       = aws_cloudwatch_log_group.expense_service.name
}

output "receipt_service_log_group" {
  description = "CloudWatch Log Group name for receipt-service"
  value       = aws_cloudwatch_log_group.receipt_service.name
}

output "notification_service_log_group" {
  description = "CloudWatch Log Group name for notification-service"
  value       = aws_cloudwatch_log_group.notification_service.name
}

output "cluster_log_group" {
  description = "CloudWatch Log Group name for cluster logs"
  value       = aws_cloudwatch_log_group.cluster.name
}

# IAM Role for Fluent Bit
output "fluent_bit_role_arn" {
  description = "ARN of IAM role for Fluent Bit"
  value       = aws_iam_role.fluent_bit.arn
}

