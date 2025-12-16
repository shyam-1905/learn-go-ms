variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "expense-tracker"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "namespace" {
  description = "Kubernetes namespace where services are deployed"
  type        = string
  default     = "expense-tracker"
}

variable "oidc_provider_arn" {
  description = "ARN of the EKS OIDC provider"
  type        = string
}

variable "oidc_provider_url" {
  description = "URL of the EKS OIDC provider"
  type        = string
}

variable "auth_events_topic_arn" {
  description = "ARN of auth events SNS topic"
  type        = string
}

variable "expense_events_topic_arn" {
  description = "ARN of expense events SNS topic"
  type        = string
}

variable "receipt_events_topic_arn" {
  description = "ARN of receipt events SNS topic"
  type        = string
}

variable "notification_email_topic_arn" {
  description = "ARN of notification email SNS topic"
  type        = string
}

variable "s3_bucket_arn" {
  description = "ARN of S3 bucket for receipts"
  type        = string
}

variable "auth_events_queue_arn" {
  description = "ARN of auth events SQS queue"
  type        = string
}

variable "expense_events_queue_arn" {
  description = "ARN of expense events SQS queue"
  type        = string
}

variable "receipt_events_queue_arn" {
  description = "ARN of receipt events SQS queue"
  type        = string
}

variable "enable_external_secrets" {
  description = "Enable External Secrets Operator IAM role"
  type        = bool
  default     = false
}

variable "secrets_manager_arns" {
  description = "List of Secrets Manager ARNs for External Secrets Operator"
  type        = list(string)
  default     = []
}

