# Root Module Variables
# These variables control the entire infrastructure

# ============================================================================
# Project Configuration
# ============================================================================
variable "project_name" {
  description = "Name of the project (used for resource naming)"
  type        = string
  default     = "expense-tracker"
}

variable "environment" {
  description = "Environment name (production, staging, dev)"
  type        = string
  default     = "production"
}

variable "aws_region" {
  description = "AWS region to deploy resources"
  type        = string
  default     = "us-east-1"
}

# ============================================================================
# Networking Configuration
# ============================================================================
variable "vpc_cidr" {
  description = "CIDR block for the VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "public_subnet_cidrs" {
  description = "CIDR blocks for public subnets"
  type        = list(string)
  default     = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "private_subnet_cidrs" {
  description = "CIDR blocks for private subnets"
  type        = list(string)
  default     = ["10.0.11.0/24", "10.0.12.0/24", "10.0.13.0/24"]
}

variable "enable_nat_gateway" {
  description = "Enable NAT Gateway (costs ~$32/month per NAT)"
  type        = bool
  default     = true
}

# ============================================================================
# EKS Configuration
# ============================================================================
variable "kubernetes_version" {
  description = "Kubernetes version for EKS"
  type        = string
  default     = "1.28"
}

variable "node_instance_types" {
  description = "EC2 instance types for node group"
  type        = list(string)
  default     = ["t3.medium"]
}

variable "node_capacity_type" {
  description = "Capacity type: ON_DEMAND or SPOT"
  type        = string
  default     = "ON_DEMAND"
}

variable "node_desired_size" {
  description = "Desired number of nodes"
  type        = number
  default     = 2
}

variable "node_min_size" {
  description = "Minimum number of nodes"
  type        = number
  default     = 1
}

variable "node_max_size" {
  description = "Maximum number of nodes"
  type        = number
  default     = 4
}

variable "node_disk_size" {
  description = "Disk size in GB for each node"
  type        = number
  default     = 20
}

variable "ssh_key_name" {
  description = "EC2 SSH key name for node access (optional)"
  type        = string
  default     = null
}

variable "cluster_endpoint_public_access_cidrs" {
  description = "CIDR blocks allowed to access cluster endpoint (empty = all)"
  type        = list(string)
  default     = []
}

variable "log_retention_days" {
  description = "Number of days to retain CloudWatch logs"
  type        = number
  default     = 7
}

# ============================================================================
# RDS Configuration
# ============================================================================
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.micro"
}

variable "postgres_version" {
  description = "PostgreSQL version"
  type        = string
  default     = "15.4"
}

variable "allocated_storage" {
  description = "Initial RDS storage size in GB"
  type        = number
  default     = 20
}

variable "max_allocated_storage" {
  description = "Maximum RDS storage size for autoscaling"
  type        = number
  default     = 100
}

variable "backup_retention_days" {
  description = "Number of days to retain RDS backups"
  type        = number
  default     = 7
}

variable "enable_multi_az" {
  description = "Enable Multi-AZ for RDS (high availability, costs 2x)"
  type        = bool
  default     = false
}

variable "deletion_protection" {
  description = "Enable deletion protection for RDS"
  type        = bool
  default     = true
}

# ============================================================================
# ArgoCD Configuration
# ============================================================================
variable "git_repo_url" {
  description = "Git repository URL for ArgoCD"
  type        = string
  # Example: "https://github.com/username/learn-go-ms.git"
}

variable "git_branch" {
  description = "Git branch for ArgoCD to sync from"
  type        = string
  default     = "main"
}

variable "kustomize_path" {
  description = "Path to Kustomize overlay in Git repo"
  type        = string
  default     = "k8s/overlays/production"
}

# ============================================================================
# Kubernetes Configuration
# ============================================================================
variable "k8s_namespace" {
  description = "Kubernetes namespace for application deployment"
  type        = string
  default     = "expense-tracker"
}

# ============================================================================
# Optional Features
# ============================================================================
variable "enable_external_secrets" {
  description = "Enable External Secrets Operator"
  type        = bool
  default     = false
}

