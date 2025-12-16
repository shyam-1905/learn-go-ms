# Root Module Outputs
# These outputs provide important information after infrastructure is created

# ============================================================================
# EKS Outputs
# ============================================================================
output "cluster_name" {
  description = "Name of the EKS cluster"
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.eks.cluster_endpoint
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data for the cluster"
  value       = module.eks.cluster_certificate_authority_data
  sensitive   = true
}

output "configure_kubectl" {
  description = "Command to configure kubectl"
  value       = "aws eks update-kubeconfig --name ${module.eks.cluster_name} --region ${var.aws_region}"
}

# ============================================================================
# RDS Outputs
# ============================================================================
output "auth_db_endpoint" {
  description = "RDS endpoint for auth database"
  value       = module.rds.auth_db_endpoint
}

output "expense_db_endpoint" {
  description = "RDS endpoint for expense database"
  value       = module.rds.expense_db_endpoint
}

output "receipt_db_endpoint" {
  description = "RDS endpoint for receipt database"
  value       = module.rds.receipt_db_endpoint
}

output "database_secrets_arns" {
  description = "ARNs of Secrets Manager secrets for databases"
  value = {
    auth_db    = module.rds.auth_db_secret_arn
    expense_db = module.rds.expense_db_secret_arn
    receipt_db = module.rds.receipt_db_secret_arn
  }
}

# ============================================================================
# S3 Outputs
# ============================================================================
output "s3_bucket_name" {
  description = "Name of the S3 bucket for receipts"
  value       = module.s3.bucket_name
}

output "s3_bucket_arn" {
  description = "ARN of the S3 bucket"
  value       = module.s3.bucket_arn
}

# ============================================================================
# Messaging Outputs
# ============================================================================
output "sns_topic_arns" {
  description = "ARNs of SNS topics"
  value = {
    auth_events         = module.messaging.auth_events_topic_arn
    expense_events      = module.messaging.expense_events_topic_arn
    receipt_events      = module.messaging.receipt_events_topic_arn
    notification_email  = module.messaging.notification_email_topic_arn
  }
}

output "sqs_queue_urls" {
  description = "URLs of SQS queues"
  value = {
    auth_events    = module.messaging.auth_events_queue_url
    expense_events = module.messaging.expense_events_queue_url
    receipt_events = module.messaging.receipt_events_queue_url
  }
}

# ============================================================================
# ECR Outputs
# ============================================================================
output "ecr_repository_urls" {
  description = "URLs of ECR repositories"
  value = {
    auth_service         = module.ecr.auth_service_repository_url
    expense_service      = module.ecr.expense_service_repository_url
    receipt_service      = module.ecr.receipt_service_repository_url
    notification_service = module.ecr.notification_service_repository_url
  }
}

# ============================================================================
# IAM Outputs (IRSA)
# ============================================================================
output "service_role_arns" {
  description = "ARNs of IAM roles for service accounts"
  value = {
    auth_service         = module.iam.auth_service_role_arn
    expense_service      = module.iam.expense_service_role_arn
    receipt_service      = module.iam.receipt_service_role_arn
    notification_service = module.iam.notification_service_role_arn
  }
}

# ============================================================================
# ArgoCD Outputs
# ============================================================================
output "argocd_admin_password" {
  description = "ArgoCD initial admin password"
  value       = module.argocd.argocd_admin_password
  sensitive   = true
}

output "argocd_server_url" {
  description = "Internal URL of ArgoCD server"
  value       = module.argocd.argocd_server_url
}

output "argocd_access_instructions" {
  description = "Instructions for accessing ArgoCD UI"
  value       = module.argocd.argocd_access_instructions
}

# ============================================================================
# Logging Outputs
# ============================================================================
output "cloudwatch_log_groups" {
  description = "CloudWatch Log Group names"
  value = {
    auth_service         = module.logging.auth_service_log_group
    expense_service      = module.logging.expense_service_log_group
    receipt_service      = module.logging.receipt_service_log_group
    notification_service = module.logging.notification_service_log_group
    cluster              = module.logging.cluster_log_group
  }
}

output "fluent_bit_role_arn" {
  description = "ARN of IAM role for Fluent Bit"
  value       = module.logging.fluent_bit_role_arn
}

# ============================================================================
# Networking Outputs
# ============================================================================
output "vpc_id" {
  description = "ID of the VPC"
  value       = module.networking.vpc_id
}

output "public_subnet_ids" {
  description = "IDs of public subnets"
  value       = module.networking.public_subnet_ids
}

output "private_subnet_ids" {
  description = "IDs of private subnets"
  value       = module.networking.private_subnet_ids
}

