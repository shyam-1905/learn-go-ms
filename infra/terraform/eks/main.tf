# EKS Module
# This module creates the Kubernetes cluster on AWS (EKS)
# It includes: EKS cluster, node groups, OIDC provider, and addons

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
  }
}

# ============================================================================
# Data Sources
# ============================================================================

# Get current AWS caller identity (for IAM policies)
data "aws_caller_identity" "current" {}

# Get current AWS region
data "aws_region" "current" {}

# ============================================================================
# EKS Cluster
# ============================================================================
# The EKS cluster is the Kubernetes control plane managed by AWS
resource "aws_eks_cluster" "main" {
  name     = "${var.project_name}-cluster"
  role_arn = aws_iam_role.cluster.arn
  version  = var.kubernetes_version
  
  # VPC configuration
  vpc_config {
    subnet_ids              = var.private_subnet_ids
    endpoint_private_access = true   # Allow private API access
    endpoint_public_access  = true    # Allow public API access (can restrict with CIDR blocks)
    public_access_cidrs     = var.cluster_endpoint_public_access_cidrs
  }
  
  # Enable encryption at rest for Kubernetes secrets
  encryption_config {
    provider {
      key_arn = aws_kms_key.eks.arn
    }
    resources = ["secrets"]
  }
  
  # Enable logging for cluster activities
  enabled_cluster_log_types = ["api", "audit", "authenticator", "controllerManager", "scheduler"]
  
  # Ensure IAM role is created before cluster
  depends_on = [
    aws_iam_role_policy_attachment.cluster_AmazonEKSClusterPolicy,
    aws_cloudwatch_log_group.eks_cluster,
  ]
  
  tags = {
    Name        = "${var.project_name}-cluster"
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# ============================================================================
# CloudWatch Log Group for EKS
# ============================================================================
resource "aws_cloudwatch_log_group" "eks_cluster" {
  name              = "/aws/eks/${var.project_name}-cluster/cluster"
  retention_in_days = var.log_retention_days
  
  tags = {
    Name        = "${var.project_name}-eks-logs"
    Environment = var.environment
  }
}

# ============================================================================
# KMS Key for EKS Encryption
# ============================================================================
resource "aws_kms_key" "eks" {
  description             = "EKS cluster encryption key"
  deletion_window_in_days = 10
  enable_key_rotation     = true
  
  tags = {
    Name        = "${var.project_name}-eks-key"
    Environment = var.environment
  }
}

resource "aws_kms_alias" "eks" {
  name          = "alias/${var.project_name}-eks"
  target_key_id = aws_kms_key.eks.key_id
}

# ============================================================================
# OIDC Provider
# ============================================================================
# OIDC (OpenID Connect) provider allows EKS to integrate with IAM
# This enables IRSA (IAM Roles for Service Accounts)
data "tls_certificate" "eks" {
  url = aws_eks_cluster.main.identity[0].oidc[0].issuer
}

resource "aws_iam_openid_connect_provider" "eks" {
  client_id_list  = ["sts.amazonaws.com"]
  thumbprint_list = [data.tls_certificate.eks.certificates[0].sha1_fingerprint]
  url             = aws_eks_cluster.main.identity[0].oidc[0].issuer
  
  tags = {
    Name        = "${var.project_name}-eks-oidc"
    Environment = var.environment
  }
}

# ============================================================================
# EKS Node Group
# ============================================================================
# Node groups are the worker nodes that run your pods
resource "aws_eks_node_group" "main" {
  cluster_name    = aws_eks_cluster.main.name
  node_group_name = "${var.project_name}-node-group"
  node_role_arn   = aws_iam_role.node.arn
  subnet_ids      = var.private_subnet_ids
  
  # Instance configuration
  instance_types = var.node_instance_types
  capacity_type  = var.node_capacity_type  # ON_DEMAND or SPOT
  disk_size      = var.node_disk_size
  
  # Scaling configuration
  scaling_config {
    desired_size = var.node_desired_size
    min_size     = var.node_min_size
    max_size     = var.node_max_size
  }
  
  # Update configuration
  update_config {
    max_unavailable = 1  # Allow 1 node to be unavailable during updates
  }
  
  # Remote access (SSH) - optional, for debugging
  remote_access {
    ec2_ssh_key               = var.ssh_key_name
    source_security_group_ids = [aws_security_group.node_ssh.id]
  }
  
  # Ensure cluster and IAM roles are ready
  depends_on = [
    aws_iam_role_policy_attachment.node_AmazonEKSWorkerNodePolicy,
    aws_iam_role_policy_attachment.node_AmazonEKS_CNI_Policy,
    aws_iam_role_policy_attachment.node_AmazonEC2ContainerRegistryReadOnly,
    aws_eks_cluster.main,
  ]
  
  tags = {
    Name        = "${var.project_name}-node-group"
    Environment = var.environment
    ManagedBy   = "terraform"
  }
}

# ============================================================================
# EKS Addons
# ============================================================================
# Addons are essential components that run on the cluster

# VPC CNI - Network plugin for pods
resource "aws_eks_addon" "vpc_cni" {
  cluster_name = aws_eks_cluster.main.name
  addon_name   = "vpc-cni"
  
  tags = {
    Name        = "${var.project_name}-vpc-cni"
    Environment = var.environment
  }
}

# CoreDNS - DNS server for the cluster
resource "aws_eks_addon" "coredns" {
  cluster_name = aws_eks_cluster.main.name
  addon_name   = "coredns"
  
  depends_on = [aws_eks_node_group.main]
  
  tags = {
    Name        = "${var.project_name}-coredns"
    Environment = var.environment
  }
}

# kube-proxy - Network proxy for services
resource "aws_eks_addon" "kube_proxy" {
  cluster_name = aws_eks_cluster.main.name
  addon_name   = "kube-proxy"
  
  tags = {
    Name        = "${var.project_name}-kube-proxy"
    Environment = var.environment
  }
}

# ============================================================================
# Security Groups
# ============================================================================

# Security group for node SSH access (if remote access is enabled)
resource "aws_security_group" "node_ssh" {
  name_prefix = "${var.project_name}-node-ssh-"
  vpc_id      = var.vpc_id
  description = "Security group for EKS node SSH access"
  
  ingress {
    description = "SSH from VPC"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name        = "${var.project_name}-node-ssh-sg"
    Environment = var.environment
  }
}

