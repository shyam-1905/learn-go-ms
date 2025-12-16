# Terraform Infrastructure

This directory contains Terraform modules for deploying the complete infrastructure on AWS EKS.

## Structure

```
terraform/
├── main.tf              # Root module (orchestrates all modules)
├── variables.tf          # Root variables
├── outputs.tf           # Root outputs
├── terraform.tfvars.example  # Example variables file
├── networking/           # VPC, subnets, NAT, IGW
├── eks/                  # EKS cluster, node groups
├── rds/                  # PostgreSQL databases
├── s3/                   # S3 bucket for receipts
├── messaging/            # SNS topics, SQS queues
├── ecr/                  # ECR repositories
├── iam/                  # IAM roles for IRSA
├── alb-ingress/          # ALB Ingress Controller
├── argocd/               # ArgoCD installation
└── logging/              # CloudWatch Logs, Fluent Bit
```

## Quick Start

1. **Copy example variables file:**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. **Edit terraform.tfvars:**
   - Set `git_repo_url` to your Git repository
   - Adjust other values as needed

3. **Initialize Terraform:**
   ```bash
   terraform init
   ```

4. **Plan the deployment:**
   ```bash
   terraform plan
   ```

5. **Apply (creates all infrastructure):**
   ```bash
   terraform apply
   ```

6. **Configure kubectl:**
   ```bash
   aws eks update-kubeconfig --name expense-tracker-cluster --region us-east-1
   ```

## Deployment Order

Terraform automatically handles dependencies, but the order is:
1. Networking (VPC, subnets)
2. EKS (cluster, nodes)
3. RDS (databases)
4. S3, SNS/SQS, ECR (AWS resources)
5. IAM (roles)
6. ALB Ingress Controller
7. ArgoCD
8. Logging (CloudWatch Log Groups)

## Important Outputs

After `terraform apply`, check outputs:
```bash
terraform output
```

Key outputs:
- `configure_kubectl` - Command to configure kubectl
- `argocd_admin_password` - ArgoCD admin password
- `ecr_repository_urls` - ECR URLs for pushing images
- `database_secrets_arns` - Secrets Manager ARNs for DB credentials

## Cost Estimation

Approximate monthly costs:
- EKS cluster: ~$73
- 2x t3.medium nodes: ~$60
- 3x db.t3.micro RDS: ~$36
- NAT Gateways (3): ~$96
- S3, SNS, SQS: ~$5-10
- **Total: ~$270-280/month**

For cost savings:
- Use SPOT instances for nodes
- Use single NAT Gateway
- Use smaller RDS instances
- Disable Multi-AZ

## Cleanup

To destroy all infrastructure:
```bash
terraform destroy
```

**Warning:** This will delete everything! Make sure you have backups.

