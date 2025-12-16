# Variables for RDS Module

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

variable "vpc_id" {
  description = "ID of the VPC"
  type        = string
}

variable "private_subnet_ids" {
  description = "List of private subnet IDs for RDS"
  type        = list(string)
}

variable "eks_cluster_security_group_id" {
  description = "Security group ID of the EKS cluster (for RDS access)"
  type        = string
}

variable "db_username" {
  description = "Master username for databases"
  type        = string
  default     = "postgres"
  sensitive   = true
}

variable "db_instance_class" {
  description = "Instance class for RDS (e.g., db.t3.micro, db.t3.small)"
  type        = string
  default     = "db.t3.micro"  # Small instance for learning (can scale up later)
}

variable "postgres_version" {
  description = "PostgreSQL engine version"
  type        = string
  default     = "15.4"
}

variable "allocated_storage" {
  description = "Initial storage size in GB"
  type        = number
  default     = 20
}

variable "max_allocated_storage" {
  description = "Maximum storage size for autoscaling (0 = disabled)"
  type        = number
  default     = 100
}

variable "backup_retention_days" {
  description = "Number of days to retain backups"
  type        = number
  default     = 7
}

variable "enable_multi_az" {
  description = "Enable Multi-AZ deployment (high availability, costs 2x)"
  type        = bool
  default     = false  # Disable for cost savings in learning environment
}

variable "deletion_protection" {
  description = "Enable deletion protection"
  type        = bool
  default     = true  # Protect against accidental deletion
}
