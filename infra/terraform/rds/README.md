# RDS Module

This Terraform module creates PostgreSQL databases for the microservices.

## What it creates:

1. **3 RDS PostgreSQL Instances**:
   - `auth-db` for auth-service
   - `expense-db` for expense-service
   - `receipt-db` for receipt-service

2. **DB Subnet Group** - Private subnets for RDS
3. **Security Group** - Allows access from EKS cluster only
4. **Secrets Manager Secrets** - Stores database credentials securely
5. **IAM Role** - For RDS Enhanced Monitoring

## Usage:

```hcl
module "rds" {
  source = "./rds"
  
  project_name      = "expense-tracker"
  environment       = "production"
  vpc_id            = module.networking.vpc_id
  private_subnet_ids = module.networking.private_subnet_ids
  eks_cluster_security_group_id = module.eks.cluster_security_group_id
  
  db_instance_class = "db.t3.micro"
  allocated_storage = 20
  backup_retention_days = 7
}
```

## Accessing Credentials:

Credentials are stored in AWS Secrets Manager:
- `expense-tracker/auth-db/credentials`
- `expense-tracker/expense-db/credentials`
- `expense-tracker/receipt-db/credentials`

Retrieve with:
```bash
aws secretsmanager get-secret-value --secret-id expense-tracker/auth-db/credentials
```

## Cost Considerations:

- db.t3.micro: ~$0.017/hour (~$12/month per instance)
- Storage: ~$0.115/GB-month
- 3 instances: ~$36/month + storage
