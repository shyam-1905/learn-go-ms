# ArgoCD Module

This module installs ArgoCD on the EKS cluster and configures it for GitOps deployment.

## What it creates:

1. **ArgoCD Namespace** - `argocd` namespace
2. **Helm Release** - ArgoCD installation via Helm
3. **ArgoCD Project** - Defines allowed repositories and namespaces
4. **ArgoCD Application** - Defines what to deploy from Git

## Usage:

```hcl
module "argocd" {
  source = "./argocd"
  
  project_name      = "expense-tracker"
  git_repo_url      = "https://github.com/username/learn-go-ms.git"
  git_branch        = "main"
  kustomize_path    = "k8s/overlays/production"
  target_namespace  = "expense-tracker"
}
```

## Accessing ArgoCD:

After Terraform apply:
```bash
# Get password
terraform output -raw argocd_admin_password | base64 -d

# Port forward
kubectl port-forward svc/argocd-server -n argocd 8080:443

# Access UI
# https://localhost:8080
# Username: admin
# Password: <from above>
```

## GitOps Workflow:

1. Push changes to Git repository
2. ArgoCD detects changes (polling or webhook)
3. ArgoCD syncs changes to Kubernetes cluster
4. Services are updated automatically

