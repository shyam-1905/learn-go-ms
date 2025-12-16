# ALB Ingress Controller Module

This module installs the AWS Load Balancer Controller (formerly ALB Ingress Controller) on the EKS cluster.

## What it creates:

1. **IAM Role** - For the controller to manage ALBs
2. **Kubernetes Service Account** - With IRSA annotation
3. **Helm Release** - Installs the controller from official Helm chart

## Usage:

```hcl
module "alb_ingress" {
  source = "./alb-ingress"
  
  project_name      = "expense-tracker"
  cluster_name      = module.eks.cluster_name
  vpc_id            = module.networking.vpc_id
  oidc_provider_arn = module.eks.oidc_provider_arn
  oidc_provider_url = module.eks.oidc_provider_url
}
```

## After installation:

The controller will automatically create ALBs when you create Ingress resources in Kubernetes.

Verify installation:
```bash
kubectl get deployment -n kube-system aws-load-balancer-controller
```

