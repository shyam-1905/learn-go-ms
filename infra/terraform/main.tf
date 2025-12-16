# Root Terraform Module
# This orchestrates all infrastructure modules
# Run: terraform init, terraform plan, terraform apply

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.23"
    }
    helm = {
      source  = "hashicorp/helm"
      version = "~> 2.11"
    }
  }
  
  # Optional: Use remote state backend (S3 + DynamoDB for locking)
  # Uncomment and configure for team collaboration
  # backend "s3" {
  #   bucket         = "your-terraform-state-bucket"
  #   key            = "expense-tracker/terraform.tfstate"
  #   region         = "us-east-1"
  #   dynamodb_table = "terraform-state-lock"
  #   encrypt        = true
  # }
}

# Configure AWS Provider
provider "aws" {
  region = var.aws_region
  
  default_tags {
    tags = {
      Project     = var.project_name
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

# Configure Kubernetes Provider (after EKS is created)
# This will be configured dynamically after EKS cluster is ready
provider "kubernetes" {
  host                   = module.eks.cluster_endpoint
  cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
  
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = [
      "eks",
      "get-token",
      "--cluster-name",
      module.eks.cluster_name
    ]
  }
}

# Configure Helm Provider (after EKS is created)
provider "helm" {
  kubernetes {
    host                   = module.eks.cluster_endpoint
    cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
    
    exec {
      api_version = "client.authentication.k8s.io/v1beta1"
      command     = "aws"
      args = [
        "eks",
        "get-token",
        "--cluster-name",
        module.eks.cluster_name
      ]
    }
  }
}

# ============================================================================
# Module: Networking
# ============================================================================
# Creates VPC, subnets, NAT Gateways, Internet Gateway
module "networking" {
  source = "./networking"
  
  project_name         = var.project_name
  environment          = var.environment
  vpc_cidr             = var.vpc_cidr
  public_subnet_cidrs  = var.public_subnet_cidrs
  private_subnet_cidrs = var.private_subnet_cidrs
  enable_nat_gateway   = var.enable_nat_gateway
}

# ============================================================================
# Module: EKS
# ============================================================================
# Creates Kubernetes cluster, node groups, OIDC provider
module "eks" {
  source = "./eks"
  
  project_name      = var.project_name
  environment       = var.environment
  vpc_id            = module.networking.vpc_id
  vpc_cidr          = module.networking.vpc_cidr_block
  private_subnet_ids = module.networking.private_subnet_ids
  
  kubernetes_version              = var.kubernetes_version
  node_instance_types             = var.node_instance_types
  node_capacity_type              = var.node_capacity_type
  node_desired_size               = var.node_desired_size
  node_min_size                   = var.node_min_size
  node_max_size                   = var.node_max_size
  node_disk_size                  = var.node_disk_size
  ssh_key_name                    = var.ssh_key_name
  cluster_endpoint_public_access_cidrs = var.cluster_endpoint_public_access_cidrs
  log_retention_days              = var.log_retention_days
  
  depends_on = [module.networking]
}

# ============================================================================
# Module: RDS
# ============================================================================
# Creates 3 PostgreSQL databases
module "rds" {
  source = "./rds"
  
  project_name      = var.project_name
  environment       = var.environment
  vpc_id            = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  eks_cluster_security_group_id = module.eks.cluster_security_group_id
  
  db_instance_class     = var.db_instance_class
  postgres_version      = var.postgres_version
  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  backup_retention_days = var.backup_retention_days
  enable_multi_az       = var.enable_multi_az
  deletion_protection   = var.deletion_protection
  
  depends_on = [module.networking, module.eks]
}

# ============================================================================
# Module: S3
# ============================================================================
# Creates S3 bucket for receipt images
module "s3" {
  source = "./s3"
  
  project_name            = var.project_name
  environment             = var.environment
  receipt_service_role_arn = ""  # Will be updated after IAM module
}

# ============================================================================
# Module: Messaging
# ============================================================================
# Creates SNS topics and SQS queues
module "messaging" {
  source = "./messaging"
  
  project_name = var.project_name
  environment  = var.environment
}

# ============================================================================
# Module: ECR
# ============================================================================
# Creates ECR repositories for Docker images
module "ecr" {
  source = "./ecr"
  
  project_name = var.project_name
  environment  = var.environment
}

# ============================================================================
# Module: IAM
# ============================================================================
# Creates IAM roles for service accounts (IRSA)
module "iam" {
  source = "./iam"
  
  project_name                  = var.project_name
  environment                   = var.environment
  namespace                     = var.k8s_namespace
  oidc_provider_arn             = module.eks.oidc_provider_arn
  oidc_provider_url             = module.eks.oidc_provider_url
  auth_events_topic_arn         = module.messaging.auth_events_topic_arn
  expense_events_topic_arn      = module.messaging.expense_events_topic_arn
  receipt_events_topic_arn      = module.messaging.receipt_events_topic_arn
  notification_email_topic_arn  = module.messaging.notification_email_topic_arn
  s3_bucket_arn                 = module.s3.bucket_arn
  auth_events_queue_arn         = "arn:aws:sqs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:${var.project_name}-auth-events-queue"
  expense_events_queue_arn       = "arn:aws:sqs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:${var.project_name}-expense-events-queue"
  receipt_events_queue_arn       = "arn:aws:sqs:${var.aws_region}:${data.aws_caller_identity.current.account_id}:${var.project_name}-receipt-events-queue"
  enable_external_secrets        = var.enable_external_secrets
  secrets_manager_arns           = [
    module.rds.auth_db_secret_arn,
    module.rds.expense_db_secret_arn,
    module.rds.receipt_db_secret_arn
  ]
  
  depends_on = [module.eks, module.messaging, module.s3, module.rds]
}

# ============================================================================
# Module: ALB Ingress Controller
# ============================================================================
# Installs AWS Load Balancer Controller
module "alb_ingress" {
  source = "./alb-ingress"
  
  project_name      = var.project_name
  environment       = var.environment
  cluster_name      = module.eks.cluster_name
  vpc_id            = module.networking.vpc_id
  aws_region        = var.aws_region
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_provider_url = module.eks.oidc_provider_url
  
  depends_on = [module.eks]
}

# ============================================================================
# Module: ArgoCD
# ============================================================================
# Installs ArgoCD for GitOps deployment
module "argocd" {
  source = "./argocd"
  
  project_name     = var.project_name
  environment      = var.environment
  git_repo_url     = var.git_repo_url
  git_branch       = var.git_branch
  kustomize_path   = var.kustomize_path
  target_namespace = var.k8s_namespace
  
  depends_on = [module.eks]
}

# ============================================================================
# Module: Logging
# ============================================================================
# Creates CloudWatch Log Groups and IAM roles for Fluent Bit
module "logging" {
  source = "./logging"
  
  project_name        = var.project_name
  environment         = var.environment
  oidc_provider_arn   = module.eks.oidc_provider_arn
  oidc_provider_url   = module.eks.oidc_provider_url
  log_retention_days  = var.log_retention_days
  
  depends_on = [module.eks]
}

# ============================================================================
# Data Sources
# ============================================================================
data "aws_caller_identity" "current" {}

