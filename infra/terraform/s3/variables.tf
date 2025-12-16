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

variable "receipt_service_role_arn" {
  description = "IAM role ARN for receipt-service (for bucket policy, optional)"
  type        = string
  default     = ""
}

