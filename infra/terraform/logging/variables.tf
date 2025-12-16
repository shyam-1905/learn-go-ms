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

variable "oidc_provider_arn" {
  description = "ARN of the EKS OIDC provider"
  type        = string
}

variable "oidc_provider_url" {
  description = "URL of the EKS OIDC provider"
  type        = string
}

variable "log_retention_days" {
  description = "Number of days to retain logs in CloudWatch"
  type        = number
  default     = 7
}

