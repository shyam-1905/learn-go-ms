# ArgoCD Module
# Installs ArgoCD on the EKS cluster using Helm
# Creates ArgoCD Project and Application for GitOps deployment

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
# Kubernetes Namespace for ArgoCD
# ============================================================================
resource "kubernetes_namespace" "argocd" {
  metadata {
    name = "argocd"
    labels = {
      name = "argocd"
      "app.kubernetes.io/name" = "argocd"
    }
  }
}

# ============================================================================
# Helm Release for ArgoCD
# ============================================================================
# Install ArgoCD using the official Helm chart
resource "helm_release" "argocd" {
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  namespace  = kubernetes_namespace.argocd.metadata[0].name
  version    = "5.51.6"  # Use latest stable version
  
  # Wait for installation to complete
  wait    = true
  timeout = 600
  
  # Optional: Custom values file
  # values = [file("${path.module}/values.yaml")]
  
  # Set some common values
  set {
    name  = "server.service.type"
    value = "ClusterIP"  # Use ClusterIP, access via port-forward or Ingress
  }
  
  set {
    name  = "configs.params.server.insecure"
    value = "true"  # Allow HTTP access (for development)
  }
  
  depends_on = [kubernetes_namespace.argocd]
}

# ============================================================================
# ArgoCD Project
# ============================================================================
# Project defines which repositories and namespaces ArgoCD can deploy to
resource "kubernetes_manifest" "argocd_project" {
  depends_on = [helm_release.argocd]
  
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "AppProject"
    metadata = {
      name      = var.project_name
      namespace = "argocd"
    }
    spec = {
      description = "Expense Tracker Application Project"
      
      # Source repositories (Git repos ArgoCD can pull from)
      sourceRepos = var.git_repo_url != "" ? [var.git_repo_url] : ["*"]
      
      # Destinations (where ArgoCD can deploy)
      destinations = [
        {
          namespace = var.target_namespace
          server    = "https://kubernetes.default.svc"
        }
      ]
      
      # Cluster resources ArgoCD can manage
      clusterResourceWhitelist = [
        {
          group = "*"
          kind  = "*"
        }
      ]
      
      # Namespace resources ArgoCD can manage
      namespaceResourceWhitelist = [
        {
          group = "*"
          kind  = "*"
        }
      ]
    }
  }
}

# ============================================================================
# ArgoCD Application
# ============================================================================
# Application defines what to deploy and how
resource "kubernetes_manifest" "argocd_application" {
  depends_on = [kubernetes_manifest.argocd_project]
  
  manifest = {
    apiVersion = "argoproj.io/v1alpha1"
    kind       = "Application"
    metadata = {
      name      = "${var.project_name}-app"
      namespace = "argocd"
      finalizers = ["resources-finalizer.argocd.argoproj.io"]
    }
    spec = {
      project = var.project_name
      
      # Source: Git repository
      source = {
        repoURL        = var.git_repo_url
        targetRevision = var.git_branch
        path           = var.kustomize_path
      }
      
      # Destination: Kubernetes cluster and namespace
      destination = {
        server    = "https://kubernetes.default.svc"
        namespace = var.target_namespace
      }
      
      # Sync policy: automatic sync with self-heal
      syncPolicy = {
        automated = {
          prune    = true   # Delete resources not in Git
          selfHeal = true   # Automatically fix drift
        }
        syncOptions = ["CreateNamespace=true"]  # Create namespace if it doesn't exist
        retry = {
          limit = 5
          backoff = {
            duration    = "5s"
            factor      = 2
            maxDuration = "3m"
          }
        }
      }
    }
  }
}

# ============================================================================
# Data Source: ArgoCD Admin Password
# ============================================================================
# Retrieve the initial admin password from Kubernetes secret
data "kubernetes_secret" "argocd_admin_password" {
  depends_on = [helm_release.argocd]
  
  metadata {
    name      = "argocd-initial-admin-secret"
    namespace = "argocd"
  }
}

