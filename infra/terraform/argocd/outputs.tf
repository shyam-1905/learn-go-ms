# ArgoCD Outputs

output "argocd_namespace" {
  description = "Namespace where ArgoCD is installed"
  value       = kubernetes_namespace.argocd.metadata[0].name
}

output "argocd_server_url" {
  description = "Internal URL of ArgoCD server"
  value       = "https://argocd-server.argocd.svc.cluster.local"
}

output "argocd_admin_password" {
  description = "Initial admin password for ArgoCD (base64 encoded)"
  value       = data.kubernetes_secret.argocd_admin_password.data.password
  sensitive   = true
}

output "argocd_project_name" {
  description = "Name of the ArgoCD Project"
  value       = var.project_name
}

output "argocd_application_name" {
  description = "Name of the ArgoCD Application"
  value       = "${var.project_name}-app"
}

# Instructions for accessing ArgoCD
output "argocd_access_instructions" {
  description = "Instructions for accessing ArgoCD UI"
  value = <<-EOT
    To access ArgoCD UI:
    
    1. Get admin password:
       terraform output -raw argocd_admin_password | base64 -d
    
    2. Port forward:
       kubectl port-forward svc/argocd-server -n argocd 8080:443
    
    3. Access UI:
       https://localhost:8080
       Username: admin
       Password: <from step 1>
  EOT
}

