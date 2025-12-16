# EKS Module

This Terraform module creates an Amazon EKS (Elastic Kubernetes Service) cluster.

## What it creates:

1. **EKS Cluster** - Managed Kubernetes control plane
2. **Node Group** - EC2 instances that run your pods
3. **OIDC Provider** - For IRSA (IAM Roles for Service Accounts)
4. **EKS Addons** - VPC CNI, CoreDNS, kube-proxy
5. **IAM Roles** - For cluster and nodes
6. **KMS Key** - For encryption at rest
7. **CloudWatch Log Group** - For cluster logs

## Usage:

```hcl
module "eks" {
  source = "./eks"
  
  project_name      = "expense-tracker"
  environment       = "production"
  vpc_id            = module.networking.vpc_id
  vpc_cidr          = module.networking.vpc_cidr_block
  private_subnet_ids = module.networking.private_subnet_ids
  
  kubernetes_version = "1.28"
  node_instance_types = ["t3.medium"]
  node_desired_size  = 2
  node_min_size      = 1
  node_max_size      = 4
}
```

## After creation:

Configure kubectl:
```bash
aws eks update-kubeconfig --name expense-tracker-cluster --region us-east-1
```

Verify:
```bash
kubectl get nodes
```

## Cost Considerations:

- EKS cluster: ~$0.10/hour (~$73/month)
- Node instances: Depends on instance type and count
- t3.medium: ~$0.0416/hour (~$30/month per node)

