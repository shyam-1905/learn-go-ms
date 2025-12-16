# Integrating RDS with EKS - Auth Service

This guide explains how to connect your auth-service running in EKS to the RDS PostgreSQL database.

## 🎓 Learning Objectives

- **Kubernetes Secrets**: Storing sensitive data
- **ConfigMaps**: Storing non-sensitive configuration
- **Service Accounts**: Pod identity
- **IRSA**: IAM Roles for Service Accounts (for Secrets Manager access)
- **Environment Variables**: Passing config to containers

## 📋 Prerequisites

1. RDS instance created via Terraform
2. EKS cluster running
3. `kubectl` configured to access your cluster

## 🔐 Option 1: Using Kubernetes Secrets (Recommended for Learning)

### Step 1: Retrieve Credentials from Secrets Manager

```bash
# Get master credentials
aws secretsmanager get-secret-value \
  --secret-id expense-tracker/rds/master-credentials \
  --query SecretString --output text | jq -r .
```

### Step 2: Create Kubernetes Secret

```bash
# Create secret with database credentials
kubectl create secret generic auth-db-credentials \
  --from-literal=DB_HOST=<rds-endpoint> \
  --from-literal=DB_PORT=5432 \
  --from-literal=DB_USER=postgres \
  --from-literal=DB_PASSWORD=<password-from-secrets-manager> \
  --from-literal=DB_NAME=auth_db \
  --namespace=default
```

### Step 3: Create ConfigMap for Non-Sensitive Config

```bash
kubectl create configmap auth-service-config \
  --from-literal=SERVER_PORT=8080 \
  --from-literal=JWT_EXPIRATION_HOURS=24 \
  --namespace=default
```

### Step 4: Update Deployment to Use Secrets

Create or update `deployments/auth-service/deployment.yaml`:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
  namespace: default
spec:
  replicas: 2
  selector:
    matchLabels:
      app: auth-service
  template:
    metadata:
      labels:
        app: auth-service
    spec:
      containers:
      - name: auth-service
        image: <your-ecr-repo>/auth-service:latest
        ports:
        - containerPort: 8080
        env:
        # Database config from Secret
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
        - name: DB_USER
          valueFrom:
            secretKeyRef:
              name: auth-db-credentials
              key: DB_USER
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: auth-db-credentials
              key: DB_PASSWORD
        - name: DB_NAME
          valueFrom:
            secretKeyRef:
              name: auth-db-credentials
              key: DB_NAME
        # JWT config from ConfigMap
        - name: SERVER_PORT
          valueFrom:
            configMapKeyRef:
              name: auth-service-config
              key: SERVER_PORT
        - name: JWT_EXPIRATION_HOURS
          valueFrom:
            configMapKeyRef:
              name: auth-service-config
              key: JWT_EXPIRATION_HOURS
        # JWT Secret (should also be in Secrets Manager!)
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: auth-db-credentials  # Or create separate secret
              key: JWT_SECRET
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "256Mi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 5
```

## 🔐 Option 2: Using IRSA + Secrets Manager (Production Best Practice)

This approach uses IAM Roles for Service Accounts (IRSA) to let pods access Secrets Manager directly.

### Step 1: Create IAM Role for Service Account

```bash
# Create service account
kubectl create serviceaccount auth-service-sa -n default

# Annotate with IAM role (you'll create this role separately)
kubectl annotate serviceaccount auth-service-sa \
  eks.amazonaws.com/role-arn=arn:aws:iam::<account-id>:role/auth-service-secrets-role
```

### Step 2: Update Deployment to Use Service Account

```yaml
spec:
  serviceAccountName: auth-service-sa
  containers:
  - name: auth-service
    # ... rest of config
    env:
    - name: AWS_REGION
      value: us-east-1
    - name: SECRETS_MANAGER_SECRET
      value: expense-tracker/rds/master-credentials
```

### Step 3: Update Auth Service to Read from Secrets Manager

You'll need to add AWS SDK code to fetch secrets at startup:

```go
// In config/config.go, add function to fetch from Secrets Manager
func LoadFromSecretsManager(secretName string) (*Config, error) {
    // Use AWS SDK to fetch secret
    // Parse JSON and populate Config
}
```

## 🧪 Testing the Connection

### 1. Check Pod Logs

```bash
kubectl logs -l app=auth-service
```

Look for: "Database connection established!"

### 2. Test from Inside Pod

```bash
# Get pod name
kubectl get pods -l app=auth-service

# Exec into pod
kubectl exec -it <pod-name> -- sh

# Test database connection (if psql is installed)
psql -h $DB_HOST -U $DB_USER -d $DB_NAME
```

### 3. Test API Endpoint

```bash
# Port forward to local machine
kubectl port-forward svc/auth-service 8080:8080

# Test health endpoint
curl http://localhost:8080/health

# Test registration
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"test123","name":"Test User"}'
```

## 🔧 Troubleshooting

### Connection Refused

**Problem**: Pod can't connect to RDS

**Solutions**:
1. Check security group allows traffic from EKS nodes
2. Verify RDS is in private subnets
3. Check route tables

```bash
# Check security group rules
aws ec2 describe-security-groups --group-ids <rds-sg-id>

# Check if RDS is publicly accessible (should be false)
aws rds describe-db-instances --db-instance-identifier expense-tracker-postgres
```

### Authentication Failed

**Problem**: Wrong username/password

**Solutions**:
1. Verify secret values are correct
2. Check if password has special characters (may need URL encoding)
3. Retrieve fresh credentials from Secrets Manager

### DNS Resolution Failed

**Problem**: Can't resolve RDS endpoint

**Solutions**:
1. Check VPC DNS settings
2. Verify pod is in same VPC
3. Check CoreDNS is running in cluster

```bash
# Check CoreDNS
kubectl get pods -n kube-system | grep coredns
```

## 📚 Key Concepts Learned

### Kubernetes Secrets
- Store sensitive data (passwords, keys)
- Base64 encoded (not encrypted by default)
- Mounted as files or environment variables

### ConfigMaps
- Store non-sensitive configuration
- Can be mounted as files or env vars
- Easy to update without rebuilding images

### Service Accounts
- Identity for pods
- Can be bound to IAM roles (IRSA)
- Enables fine-grained permissions

### IRSA (IAM Roles for Service Accounts)
- Pods assume IAM roles
- No need to store AWS credentials
- Follows principle of least privilege

## 🎯 Next Steps

1. **Create the deployment YAML** using the template above
2. **Run database migrations** from a Job or init container
3. **Set up monitoring** (CloudWatch, Prometheus)
4. **Configure autoscaling** (HPA)
5. **Add service mesh** (Istio, Linkerd) for advanced networking

## 📖 Additional Resources

- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [IRSA Documentation](https://docs.aws.amazon.com/eks/latest/userguide/iam-roles-for-service-accounts.html)
- [AWS Secrets Manager](https://docs.aws.amazon.com/secretsmanager/)
