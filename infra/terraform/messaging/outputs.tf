# SNS Topic ARNs
output "auth_events_topic_arn" {
  description = "ARN of auth events SNS topic"
  value       = aws_sns_topic.auth_events.arn
}

output "expense_events_topic_arn" {
  description = "ARN of expense events SNS topic"
  value       = aws_sns_topic.expense_events.arn
}

output "receipt_events_topic_arn" {
  description = "ARN of receipt events SNS topic"
  value       = aws_sns_topic.receipt_events.arn
}

output "notification_email_topic_arn" {
  description = "ARN of notification email SNS topic"
  value       = aws_sns_topic.notification_email.arn
}

# SQS Queue URLs
output "auth_events_queue_url" {
  description = "URL of auth events SQS queue"
  value       = aws_sqs_queue.auth_events.url
}

output "expense_events_queue_url" {
  description = "URL of expense events SQS queue"
  value       = aws_sqs_queue.expense_events.url
}

output "receipt_events_queue_url" {
  description = "URL of receipt events SQS queue"
  value       = aws_sqs_queue.receipt_events.url
}

