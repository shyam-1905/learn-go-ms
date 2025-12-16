variable "project_name" {
  description = "Name of the project (used for ArgoCD Project and Application)"
  type        = string
  default     = "expense-tracker"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "git_repo_url" {
  description = "Git repository URL for ArgoCD to sync from"
  type        = string
  # Example: "https://github.com/username/learn-go-ms.git"
}

variable "git_branch" {
  description = "Git branch to sync from"
  type        = string
  default     = "main"
}

variable "kustomize_path" {
  description = "Path to Kustomize overlay in Git repo"
  type        = string
  default     = "k8s/overlays/production"
}

variable "target_namespace" {
  description = "Kubernetes namespace where ArgoCD will deploy"
  type        = string
  default     = "expense-tracker"
}

