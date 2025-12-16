# Quick Start Guide - RDS for Auth Service

This guide will walk you through setting up RDS PostgreSQL and connecting it to your auth-service in EKS.

## 🚀 Step-by-Step Setup

### Step 1: Prerequisites Check

```bash
# Check Terraform version
terraform version  # Should be >= 1.0

# Check AWS CLI
aws --version

# Verify AWS credentials
aws sts get-caller-identity

# Check kubectl (for EKS)
kubectl version --client
```

### Step 2: Configure Terraform Variables

```bash
cd infra/terraform/rds

# Copy example variables file
cp terraform.tfvars.example terraform.tfvars

# Edit with your values
# Windows: notepad terraform.tfvars
# Linux/Mac: nano terraform.tfvars
```

**Important variables to set:**
- `vpc_name` or `vpc_id`: Your EKS VPC
- `eks_cluster_security_group_name`: EKS cluster security group
- `project_name`: Should match your project

### Step 3: Find Your EKS Security Groups

```bash
# Get your EKS cluster name
aws eks list-clusters

# Get cluster security group
aws eks describe-cluster --name <cluster-name> --query 'cluster.resourcesVpcConfig.clusterSecurityGroupId' --output text

# Get node group security groups
aws eks describe-nodegroup --cluster-name <cluster-name> --nodegroup-name <nodegroup-name> --query 'nodegroup.resources.remoteAccessSecurityGroup' --output text
```

### Step 4: Initialize and Apply Terraform

```bash
# Initialize Terraform (downloads providers)
terraform init

# Review what will be created
terraform plan

# Apply configuration (creates RDS)
terraform apply
```

Type `yes` when prompted. This takes about 10-15 minutes.

### Step 5: Get RDS Connection Details

```bash
# Get RDS endpoint
terraform output rds_endpoint

# Get secret ARN
terraform output master_credentials_secret_arn
```

### Step 6: Retrieve Database Credentials

**Option A: Using the helper script (PowerShell)**
```powershell
cd scripts
.\get-db-credentials.ps1 master
```

**Option B: Using AWS CLI directly**
```bash
aws secretsmanager get-secret-value \
  --secret-id expense-tracker/rds/master-credentials \
  --query SecretString --output text | jq -r .
```

### Step 7: Create Kubernetes Secret

```bash
# Get the values from Step 6
RDS_ENDPOINT=$(terraform output -raw rds_endpoint)
DB_PASSWORD=$(aws secretsmanager get-secret-value \
  --secret-id expense-tracker/rds/master-credentials \
  --query SecretString --output text | jq -r .password)

# Create Kubernetes secret
kubectl create secret generic auth-db-credentials \
  --from-literal=DB_HOST=$RDS_ENDPOINT \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_USER=postgres \
  --from-literal=DB_PASSWORD=$DB_PASSWORD \
  --from-literal=DB_NAME=auth_db

# Create JWT secret (generate a strong secret!)
JWT_SECRET=$(openssl rand -base64 32)
kubectl create secret generic auth-jwt-secret \
  --from-literal=JWT_SECRET=$JWT_SECRET
```

### Step 8: Run Database Migration

**Option A: Using a Kubernetes Job**

Create `migrations/job.yaml`:
```yaml
apiVersion: batch/v1
kind: Job
metadata:
  name: auth-db-migration
spec:
  template:
    spec:
      containers:
      - name: migrate
        image: postgres:15
        command: ["psql"]
        args:
        - "-h"
        - "$(DB_HOST)"
        - "-U"
        - "$(DB_USER)"
        - "-d"
        - "$(DB_NAME)"
        - "-f"
        - "/migrations/001_create_users_table.sql"
        envFrom:
        - secretRef:
            name: auth-db-credentials
      restartPolicy: Never
```

**Option B: Run manually from your machine**

```bash
# Install psql (if not installed)
# Windows: Download from https://www.postgresql.org/download/windows/
# Mac: brew install postgresql
# Linux: sudo apt-get install postgresql-client

# Get connection details
RDS_ENDPOINT=$(terraform output -raw rds_endpoint)
DB_PASSWORD=$(aws secretsmanager get-secret-value \
  --secret-id expense-tracker/rds/master-credentials \
  --query SecretString --output text | jq -r .password)

# Run migration
PGPASSWORD=$DB_PASSWORD psql -h $RDS_ENDPOINT -U postgres -d auth_db -f ../../services/auth-service/migrations/001_create_users_table.sql
```

### Step 9: Update Auth Service Deployment

Update your `deployments/auth-service/deployment.yaml` to use the secrets:

```yaml
env:
- name: DB_HOST
  valueFrom:
    secretKeyRef:
      name: auth-db-credentials
      key: DB_HOST
- name: DB_PORT
  valueFrom:
    secretKeyRef:
      name: auth-db-credentials
      key: DB_PORT
# ... (see EKS_INTEGRATION.md for full example)
```

### Step 10: Deploy and Test

```bash
# Apply deployment
kubectl apply -f deployments/auth-service/deployment.yaml

# Check pods
kubectl get pods -l app=auth-service

# Check logs
kubectl logs -l app=auth-service

# Test health endpoint
kubectl port-forward svc/auth-service 8080:8080
curl http://localhost:8080/health
```

## ✅ Verification Checklist

- [ ] RDS instance is running
- [ ] Security group allows traffic from EKS
- [ ] Kubernetes secrets created
- [ ] Database migration completed
- [ ] Auth service pods are running
- [ ] Health endpoint returns 200
- [ ] Can register a new user
- [ ] Can login with credentials

## 🔧 Common Issues

### Issue: Terraform can't find VPC

**Solution**: 
- Check VPC name tag matches your VPC
- Or provide `vpc_id` directly in `terraform.tfvars`

### Issue: RDS creation fails

**Solution**:
- Check you have sufficient AWS service quotas
- Verify subnet group has subnets in at least 2 AZs (for Multi-AZ)
- Check IAM permissions

### Issue: Can't connect from EKS

**Solution**:
- Verify security group allows port 5432 from EKS security groups
- Check RDS is in private subnets
- Verify route tables allow communication

### Issue: Migration fails

**Solution**:
- Verify connection string is correct
- Check database exists
- Ensure user has CREATE TABLE permissions

## 📚 Next Steps

1. Set up monitoring and alerts
2. Configure automated backups
3. Enable Multi-AZ for production
4. Set up connection pooling (PgBouncer)
5. Implement read replicas for scaling

## 🆘 Getting Help

- Check logs: `kubectl logs -l app=auth-service`
- Check RDS status: AWS Console → RDS
- Check security groups: AWS Console → EC2 → Security Groups
- Review Terraform state: `terraform show`
