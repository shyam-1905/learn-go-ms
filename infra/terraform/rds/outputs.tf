# Outputs for RDS Module

# Auth DB Outputs
output "auth_db_endpoint" {
  description = "RDS endpoint for auth database"
  value       = aws_db_instance.auth_db.endpoint
  sensitive   = false
}

output "auth_db_port" {
  description = "Port for auth database"
  value       = aws_db_instance.auth_db.port
}

output "auth_db_secret_arn" {
  description = "ARN of Secrets Manager secret for auth database"
  value       = aws_secretsmanager_secret.auth_db.arn
}

# Expense DB Outputs
output "expense_db_endpoint" {
  description = "RDS endpoint for expense database"
  value       = aws_db_instance.expense_db.endpoint
}

output "expense_db_port" {
  description = "Port for expense database"
  value       = aws_db_instance.expense_db.port
}

output "expense_db_secret_arn" {
  description = "ARN of Secrets Manager secret for expense database"
  value       = aws_secretsmanager_secret.expense_db.arn
}

# Receipt DB Outputs
output "receipt_db_endpoint" {
  description = "RDS endpoint for receipt database"
  value       = aws_db_instance.receipt_db.endpoint
}

output "receipt_db_port" {
  description = "Port for receipt database"
  value       = aws_db_instance.receipt_db.port
}

output "receipt_db_secret_arn" {
  description = "ARN of Secrets Manager secret for receipt database"
  value       = aws_secretsmanager_secret.receipt_db.arn
}

# Common Outputs
output "db_subnet_group_name" {
  description = "Name of the DB subnet group"
  value       = aws_db_subnet_group.main.name
}

output "rds_security_group_id" {
  description = "Security group ID for RDS instances"
  value       = aws_security_group.rds.id
}
