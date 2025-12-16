# Deployment Guide

This guide walks you through deploying the Expense Tracker microservices platform to AWS EKS.

## Prerequisites

Before starting, ensure you have the following tools installed:

- **AWS CLI** (v2.x) - [Install Guide](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html)
- **kubectl** - [Install Guide](https://kubernetes.io/docs/tasks/tools/)
- **Terraform** (>= 1.0) - [Install Guide](https://learn.hashicorp.com/tutorials/terraform/install-cli)
- **Docker** - [Install Guide](https://docs.docker.com/get-docker/)
- **Helm** (v3.x) - [Install Guide](https://helm.sh/docs/intro/install/)

### AWS Account Setup

1. Create an AWS account if you don't have one
2. Configure AWS CLI credentials:
   ```bash
   aws configure
   ```
3. Ensure you have appropriate IAM permissions for:
   - EKS cluster creation
   - VPC, RDS, S3, SNS, SQS resource creation
   - IAM role and policy management
   - ECR repository creation

## Step-by-Step Deployment

### Step 1: Clone the Repository

```bash
git clone <your-repo-url>
cd learn-go-ms
```

### Step 2: Configure Terraform Variables

1. Copy the example variables file:
   ```bash
   cp infra/terraform/terraform.tfvars.example infra/terraform/terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your values:
   ```hcl
   project_name = "expense-tracker"
   environment  = "production"
   aws_region   = "us-east-1"
   
   # ArgoCD Configuration
   git_repo_url = "https://github.com/yourusername/learn-go-ms.git"
   git_branch   = "main"
   kustomize_path = "k8s/overlays/production"
   
   # EKS Configuration
   kubernetes_version = "1.28"
   node_instance_types = ["t3.medium"]
   node_desired_size = 2
   node_min_size = 1
   node_max_size = 4
   
   # RDS Configuration
   db_instance_class = "db.t3.micro"
   postgres_version = "15.4"
   ```

### Step 3: Initialize and Apply Terraform

1. Navigate to the Terraform directory:
   ```bash
   cd infra/terraform
   ```

2. Initialize Terraform:
   ```bash
   terraform init
   ```

3. Review the execution plan:
   ```bash
   terraform plan
   ```

4. Apply the infrastructure:
   ```bash
   terraform apply
   ```
   
   This will create:
   - VPC and networking components
   - EKS cluster and node groups
   - RDS databases (3 PostgreSQL instances)
   - S3 bucket for receipts
   - SNS topics and SQS queues
   - ECR repositories
   - IAM roles for service accounts (IRSA)
   - ALB Ingress Controller
   - ArgoCD installation
   - CloudWatch Log Groups

5. Save important outputs:
   ```bash
   terraform output > terraform-outputs.txt
   ```

### Step 4: Configure kubectl

After EKS cluster is created, configure kubectl:

```bash
aws eks update-kubeconfig --name expense-tracker-cluster --region us-east-1
```

Verify connection:
```bash
kubectl get nodes
```

### Step 5: Build and Push Docker Images

For each service, build and push to ECR:

#### Auth Service

```bash
# Get ECR repository URL
ECR_URL=$(terraform output -json ecr_repository_urls | jq -r '.auth_service')

# Login to ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_URL

# Build image
docker build -f docker/Dockerfile.auth-service -t $ECR_URL:latest .

# Push image
docker push $ECR_URL:latest
```

#### Expense Service

```bash
ECR_URL=$(terraform output -json ecr_repository_urls | jq -r '.expense_service')
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_URL
docker build -f docker/Dockerfile.expense-service -t $ECR_URL:latest .
docker push $ECR_URL:latest
```

#### Receipt Service

```bash
ECR_URL=$(terraform output -json ecr_repository_urls | jq -r '.receipt_service')
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_URL
docker build -f docker/Dockerfile.receipt-service -t $ECR_URL:latest .
docker push $ECR_URL:latest
```

#### Notification Service

```bash
ECR_URL=$(terraform output -json ecr_repository_urls | jq -r '.notification_service')
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_URL
docker build -f docker/Dockerfile.notification-service -t $ECR_URL:latest .
docker push $ECR_URL:latest
```

### Step 6: Create Kubernetes Secrets

Create secrets for each service with database credentials and JWT secrets:

#### Auth Service Secrets

```bash
# Get database credentials from Secrets Manager
DB_SECRET=$(aws secretsmanager get-secret-value --secret-id expense-tracker-auth-db-secret --query SecretString --output text)

# Extract values (requires jq)
DB_HOST=$(echo $DB_SECRET | jq -r '.host')
DB_PORT=$(echo $DB_SECRET | jq -r '.port')
DB_USER=$(echo $DB_SECRET | jq -r '.username')
DB_PASSWORD=$(echo $DB_SECRET | jq -r '.password')
DB_NAME=$(echo $DB_SECRET | jq -r '.dbname')

# Generate JWT secret
JWT_SECRET=$(openssl rand -base64 32)

# Create Kubernetes secret
kubectl create secret generic auth-service-secrets \
  --from-literal=db-host=$DB_HOST \
  --from-literal=db-port=$DB_PORT \
  --from-literal=db-user=$DB_USER \
  --from-literal=db-password=$DB_PASSWORD \
  --from-literal=db-name=$DB_NAME \
  --from-literal=jwt-secret=$JWT_SECRET \
  --namespace=expense-tracker
```

Repeat for `expense-service-secrets` and `receipt-service-secrets` using their respective secret ARNs.

### Step 7: Update ConfigMaps with Terraform Outputs

Update ConfigMaps with actual values from Terraform outputs:

```bash
# Get Terraform outputs
AUTH_TOPIC_ARN=$(terraform output -json sns_topic_arns | jq -r '.auth_events')
EXPENSE_TOPIC_ARN=$(terraform output -json sns_topic_arns | jq -r '.expense_events')
RECEIPT_TOPIC_ARN=$(terraform output -json sns_topic_arns | jq -r '.receipt_events')
NOTIFICATION_TOPIC_ARN=$(terraform output -json sns_topic_arns | jq -r '.notification_email')
S3_BUCKET=$(terraform output -raw s3_bucket_name)
EXPENSE_QUEUE=$(terraform output -json sqs_queue_urls | jq -r '.expense_events')
RECEIPT_QUEUE=$(terraform output -json sqs_queue_urls | jq -r '.receipt_events')
AUTH_QUEUE=$(terraform output -json sqs_queue_urls | jq -r '.auth_events')

# Update ConfigMaps
kubectl patch configmap auth-service-config -n expense-tracker --type merge -p "{\"data\":{\"auth-events-topic-arn\":\"$AUTH_TOPIC_ARN\"}}"
kubectl patch configmap expense-service-config -n expense-tracker --type merge -p "{\"data\":{\"expense-events-topic-arn\":\"$EXPENSE_TOPIC_ARN\"}}"
kubectl patch configmap receipt-service-config -n expense-tracker --type merge -p "{\"data\":{\"s3-bucket-name\":\"$S3_BUCKET\",\"receipt-events-topic-arn\":\"$RECEIPT_TOPIC_ARN\"}}"
kubectl patch configmap notification-service-config -n expense-tracker --type merge -p "{\"data\":{\"expense-events-queue-url\":\"$EXPENSE_QUEUE\",\"receipt-events-queue-url\":\"$RECEIPT_QUEUE\",\"auth-events-queue-url\":\"$AUTH_QUEUE\",\"notification-email-topic-arn\":\"$NOTIFICATION_TOPIC_ARN\"}}"
```

### Step 8: Deploy Fluent Bit for Logging

Deploy Fluent Bit DaemonSet for centralized logging:

```bash
# Get Fluent Bit role ARN
FLUENT_BIT_ROLE=$(terraform output -raw fluent_bit_role_arn)

# Replace placeholder in fluent-bit.yaml
sed "s|<FLUENT_BIT_ROLE_ARN>|$FLUENT_BIT_ROLE|g" infra/terraform/logging/fluent-bit.yaml | \
sed "s|<AWS_REGION>|us-east-1|g" | \
sed "s|<PROJECT_NAME>|expense-tracker|g" | \
kubectl apply -f -
```

### Step 9: Deploy Application via ArgoCD

ArgoCD should already be installed by Terraform. Access it:

1. Get ArgoCD admin password:
   ```bash
   terraform output -raw argocd_admin_password | base64 -d
   ```

2. Port forward to ArgoCD server:
   ```bash
   kubectl port-forward svc/argocd-server -n argocd 8080:443
   ```

3. Access ArgoCD UI:
   - URL: `https://localhost:8080`
   - Username: `admin`
   - Password: (from step 1)

4. Verify Application Sync:
   - The application should sync automatically
   - Check sync status in ArgoCD UI
   - If not syncing, manually sync from the UI

### Step 10: Run Database Migrations

Run migrations for each service:

```bash
# Auth service migration
kubectl run auth-migrate --image=<ECR_URL>/auth-service:latest --restart=Never \
  --env="DB_HOST=$DB_HOST" \
  --env="DB_PORT=$DB_PORT" \
  --env="DB_USER=$DB_USER" \
  --env="DB_PASSWORD=$DB_PASSWORD" \
  --env="DB_NAME=$DB_NAME" \
  --command -- ./migrate

# Repeat for expense-service and receipt-service
```

### Step 11: Verify Deployment

1. Check pod status:
   ```bash
   kubectl get pods -n expense-tracker
   ```

2. Check service status:
   ```bash
   kubectl get svc -n expense-tracker
   ```

3. Check ingress:
   ```bash
   kubectl get ingress -n expense-tracker
   ```

4. Get ALB URL:
   ```bash
   kubectl get ingress expense-tracker-ingress -n expense-tracker -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'
   ```

5. Test health endpoints:
   ```bash
   ALB_URL=$(kubectl get ingress expense-tracker-ingress -n expense-tracker -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')
   curl http://$ALB_URL/auth/health
   curl http://$ALB_URL/expenses/health
   curl http://$ALB_URL/receipts/health
   ```

## Post-Deployment

### Accessing Services

- **Auth Service**: `http://<ALB_URL>/auth/*`
- **Expense Service**: `http://<ALB_URL>/expenses/*`
- **Receipt Service**: `http://<ALB_URL>/receipts/*`

### Monitoring

- **CloudWatch Logs**: Check log groups in AWS Console
- **ArgoCD**: Monitor application sync status
- **Kubernetes**: Use `kubectl` commands to check pod logs

### Troubleshooting

1. **Pods not starting**: Check pod logs with `kubectl logs <pod-name> -n expense-tracker`
2. **Database connection issues**: Verify secrets and security groups
3. **Ingress not working**: Check ALB Ingress Controller logs
4. **ArgoCD sync issues**: Check ArgoCD application events

## Cleanup

To destroy all resources:

```bash
cd infra/terraform
terraform destroy
```

**Warning**: This will delete all infrastructure including databases and S3 buckets!
