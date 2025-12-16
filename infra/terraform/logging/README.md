# Logging Module

This module creates CloudWatch Log Groups and IAM roles for centralized logging with Fluent Bit.

## What it creates:

1. **CloudWatch Log Groups** - One per service + cluster logs
2. **IAM Role** - For Fluent Bit to write to CloudWatch Logs
3. **Fluent Bit Manifest** - DaemonSet to collect logs (apply manually after Terraform)

## Usage:

```hcl
module "logging" {
  source = "./logging"
  
  project_name        = "expense-tracker"
  oidc_provider_arn   = module.eks.oidc_provider_arn
  oidc_provider_url   = module.eks.oidc_provider_url
  log_retention_days  = 7
}
```

## Deploying Fluent Bit:

After Terraform creates the IAM role:

1. Update `fluent-bit.yaml` with:
   - `<FLUENT_BIT_ROLE_ARN>` → Terraform output
   - `<AWS_REGION>` → Your region
   - `<PROJECT_NAME>` → Your project name

2. Apply the manifest:
   ```bash
   kubectl apply -f infra/terraform/logging/fluent-bit.yaml
   ```

## Viewing Logs:

In CloudWatch Console:
- `/aws/eks/expense-tracker/auth-service`
- `/aws/eks/expense-tracker/expense-service`
- etc.

Or use CloudWatch Logs Insights to query across all services.

