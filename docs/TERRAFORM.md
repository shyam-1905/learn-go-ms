# Terraform Guide

This guide explains Infrastructure as Code (IaC) concepts and our Terraform setup.

## What is Terraform?

Terraform is an open-source Infrastructure as Code tool that lets you define and provision cloud infrastructure using declarative configuration files.

### Key Concepts

- **Provider**: Plugin for cloud provider (AWS, Azure, GCP)
- **Resource**: Infrastructure component (EC2, RDS, S3)
- **Module**: Reusable configuration package
- **State**: Current state of infrastructure
- **Plan**: Preview of changes before applying

## Why Terraform?

### Benefits

1. **Version Control**: Infrastructure changes tracked in Git
2. **Reproducibility**: Same infrastructure every time
3. **Collaboration**: Team members can review changes
4. **Documentation**: Code documents infrastructure
5. **Rollback**: Revert to previous state

## Project Structure

```
infra/terraform/
├── main.tf              # Root module, orchestrates all modules
├── variables.tf         # Input variables
├── outputs.tf           # Output values
├── terraform.tfvars     # Variable values (not in Git)
├── networking/          # VPC, subnets, NAT
├── eks/                 # EKS cluster
├── rds/                 # PostgreSQL databases
├── s3/                  # S3 bucket
├── messaging/           # SNS/SQS
├── ecr/                 # Container registry
├── iam/                 # IAM roles and policies
├── alb-ingress/         # Load balancer controller
├── argocd/              # GitOps deployment
└── logging/             # CloudWatch logs
```

## Module Pattern

Each module follows this structure:

```
module-name/
├── main.tf      # Resources
├── variables.tf # Inputs
└── outputs.tf   # Outputs
```

### Example Module

```hcl
# variables.tf
variable "project_name" {
  type = string
}

# main.tf
resource "aws_s3_bucket" "main" {
  bucket = "${var.project_name}-bucket"
}

# outputs.tf
output "bucket_name" {
  value = aws_s3_bucket.main.id
}
```

## Root Module

`main.tf` orchestrates all modules:

```hcl
module "networking" {
  source = "./networking"
  # ... variables
}

module "eks" {
  source = "./eks"
  vpc_id = module.networking.vpc_id
  # ... other variables
  depends_on = [module.networking]
}
```

**Dependencies**: Modules reference outputs from other modules

## Variables

### Declaration

```hcl
variable "project_name" {
  description = "Name of the project"
  type        = string
  default     = "expense-tracker"
}
```

### Usage

```hcl
# In terraform.tfvars
project_name = "expense-tracker"

# In code
resource "aws_s3_bucket" "main" {
  bucket = "${var.project_name}-bucket"
}
```

## Outputs

Expose important values:

```hcl
output "cluster_name" {
  description = "EKS cluster name"
  value       = module.eks.cluster_name
}
```

**Usage**:
```bash
terraform output cluster_name
terraform output -json  # All outputs as JSON
```

## State Management

### Local State (Default)

Stored in `terraform.tfstate` file (not committed to Git)

### Remote State (Recommended)

Store in S3 with DynamoDB locking:

```hcl
terraform {
  backend "s3" {
    bucket         = "terraform-state-bucket"
    key            = "expense-tracker/terraform.tfstate"
    region         = "us-east-1"
    dynamodb_table = "terraform-state-lock"
    encrypt        = true
  }
}
```

**Benefits**:
- Team collaboration
- State locking (prevent conflicts)
- State versioning

## Common Commands

### Initialize

```bash
terraform init
```

Downloads providers and modules

### Plan

```bash
terraform plan
```

Preview changes without applying

### Apply

```bash
terraform apply
```

Create/update infrastructure

### Destroy

```bash
terraform destroy
```

Delete all resources

### Validate

```bash
terraform validate
```

Check configuration syntax

### Format

```bash
terraform fmt
```

Format configuration files

## Workflow

1. **Write Configuration**: Define resources in `.tf` files
2. **Initialize**: `terraform init`
3. **Plan**: `terraform plan` (review changes)
4. **Apply**: `terraform apply` (create infrastructure)
5. **Modify**: Update configuration as needed
6. **Plan Again**: Review new changes
7. **Apply**: Update infrastructure

## Best Practices

### 1. Use Modules

Break down into reusable modules:

```hcl
module "database" {
  source = "./rds"
  # ...
}
```

### 2. Use Variables

Avoid hardcoding values:

```hcl
# Bad
resource "aws_instance" "web" {
  instance_type = "t3.medium"
}

# Good
variable "instance_type" {
  default = "t3.medium"
}

resource "aws_instance" "web" {
  instance_type = var.instance_type
}
```

### 3. Use Outputs

Expose important values:

```hcl
output "database_endpoint" {
  value = aws_db_instance.main.endpoint
}
```

### 4. Version Control

- Commit `.tf` files to Git
- **Never** commit `.tfstate` files
- Use `.gitignore`:

```
*.tfstate
*.tfstate.*
.terraform/
```

### 5. Use Workspaces

Manage multiple environments:

```bash
terraform workspace new production
terraform workspace new staging
terraform workspace select production
```

### 6. Tag Resources

Add tags for cost tracking:

```hcl
resource "aws_instance" "web" {
  tags = {
    Environment = "production"
    Project     = "expense-tracker"
    ManagedBy   = "terraform"
  }
}
```

## Troubleshooting

### State Locked

```bash
# If apply fails due to lock
terraform force-unlock <LOCK_ID>
```

### State Drift

Infrastructure changed outside Terraform:

```bash
terraform refresh  # Update state
terraform plan     # See differences
```

### Import Existing Resources

```bash
terraform import aws_s3_bucket.main bucket-name
```

### Debugging

Enable verbose logging:

```bash
TF_LOG=DEBUG terraform apply
```

## Cost Management

### Estimate Costs

```bash
terraform plan -out=tfplan
terraform show -json tfplan | jq '.planned_values.root_module'
```

### Cost Optimization Tips

1. Use appropriate instance sizes
2. Enable auto-scaling
3. Use spot instances for non-critical workloads
4. Clean up unused resources
5. Monitor costs in AWS Cost Explorer

## Security

### Secrets Management

**Never** hardcode secrets:

```hcl
# Bad
variable "db_password" {
  default = "password123"
}

# Good - Use AWS Secrets Manager
data "aws_secretsmanager_secret_version" "db" {
  secret_id = "db-credentials"
}
```

### Least Privilege

Grant minimum required permissions:

```hcl
resource "aws_iam_role_policy" "example" {
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = ["s3:GetObject"]  # Specific actions only
      Resource = "arn:aws:s3:::bucket/*"
    }]
  })
}
```

## Additional Resources

- [Terraform Documentation](https://www.terraform.io/docs)
- [AWS Provider Docs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [Terraform Best Practices](https://www.terraform.io/docs/cloud/guides/recommended-practices/index.html)
