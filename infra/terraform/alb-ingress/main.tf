# ALB Ingress Controller Module
# Installs AWS Load Balancer Controller (formerly ALB Ingress Controller)
# This allows Kubernetes Ingress resources to create AWS Application Load Balancers

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
}

# ============================================================================
# IAM Role for AWS Load Balancer Controller
# ============================================================================
# The controller needs IAM permissions to create and manage ALBs

resource "aws_iam_role" "alb_controller" {
  name = "${var.project_name}-aws-load-balancer-controller"
  
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Federated = var.oidc_provider_arn
        }
        Action = "sts:AssumeRoleWithWebIdentity"
        Condition = {
          StringEquals = {
            "${replace(var.oidc_provider_url, "https://", "")}:sub" = "system:serviceaccount:kube-system:aws-load-balancer-controller"
            "${replace(var.oidc_provider_url, "https://", "")}:aud" = "sts.amazonaws.com"
          }
        }
      }
    ]
  })
  
  tags = {
    Name        = "${var.project_name}-alb-controller-role"
    Environment = var.environment
  }
}

# Attach AWS managed policy for Load Balancer Controller
resource "aws_iam_role_policy_attachment" "alb_controller" {
  role       = aws_iam_role.alb_controller.name
  policy_arn = "arn:aws:iam::aws:policy/ElasticLoadBalancingFullAccess"
}

# ============================================================================
# Kubernetes Service Account
# ============================================================================
# Service account for the Load Balancer Controller with IRSA annotation
resource "kubernetes_service_account" "alb_controller" {
  metadata {
    name      = "aws-load-balancer-controller"
    namespace = "kube-system"
    annotations = {
      "eks.amazonaws.com/role-arn" = aws_iam_role.alb_controller.arn
    }
    labels = {
      "app.kubernetes.io/name"       = "aws-load-balancer-controller"
      "app.kubernetes.io/component"  = "controller"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

# ============================================================================
# Helm Release for AWS Load Balancer Controller
# ============================================================================
# Install the controller using the official Helm chart
resource "helm_release" "aws_load_balancer_controller" {
  name       = "aws-load-balancer-controller"
  repository = "https://aws.github.io/eks-charts"
  chart      = "aws-load-balancer-controller"
  namespace  = "kube-system"
  version    = "1.6.0"  # Use latest stable version
  
  # Wait for installation
  wait = true
  timeout = 600
  
  # Values for Helm chart
  set {
    name  = "clusterName"
    value = var.cluster_name
  }
  
  set {
    name  = "serviceAccount.create"
    value = "false"  # We create it via Terraform for IRSA
  }
  
  set {
    name  = "serviceAccount.name"
    value = kubernetes_service_account.alb_controller.metadata[0].name
  }
  
  set {
    name  = "region"
    value = var.aws_region
  }
  
  set {
    name  = "vpcId"
    value = var.vpc_id
  }
  
  # Enable webhook for admission controller
  set {
    name  = "enableServiceMutatorWebhook"
    value = "false"  # Disable if not needed
  }
  
  depends_on = [
    kubernetes_service_account.alb_controller,
    aws_iam_role_policy_attachment.alb_controller
  ]
}

