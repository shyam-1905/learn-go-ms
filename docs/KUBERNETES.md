# Kubernetes Guide

This guide explains Kubernetes concepts and how our microservices are deployed.

## What is Kubernetes?

Kubernetes (K8s) is an open-source container orchestration platform that automates deployment, scaling, and management of containerized applications.

### Key Concepts

- **Pod**: Smallest deployable unit (one or more containers)
- **Deployment**: Manages pod replicas and updates
- **Service**: Network abstraction for accessing pods
- **Namespace**: Virtual cluster for resource isolation
- **ConfigMap**: Non-sensitive configuration data
- **Secret**: Sensitive data (passwords, tokens)
- **Ingress**: External access to services

## Architecture Overview

```
Internet
   ↓
Ingress (ALB)
   ↓
Services (ClusterIP)
   ↓
Pods (Containers)
   ↓
Persistent Storage (RDS)
```

## Core Resources

### 1. Namespace

Logical grouping of resources:

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: expense-tracker
```

**Why**: Isolation, resource quotas, access control

### 2. Deployment

Manages pod replicas:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: auth-service
spec:
  replicas: 2
  selector:
    matchLabels:
      app: auth-service
  template:
    spec:
      containers:
      - name: auth-service
        image: <image>
        ports:
        - containerPort: 8080
```

**Key Features**:
- **Replicas**: Number of pod instances
- **Rolling Updates**: Zero-downtime deployments
- **Rollback**: Revert to previous version

### 3. Service

Exposes pods internally:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: auth-service
spec:
  selector:
    app: auth-service
  ports:
  - port: 8080
    targetPort: 8080
  type: ClusterIP
```

**Service Types**:
- **ClusterIP**: Internal access only (default)
- **NodePort**: External access via node IP
- **LoadBalancer**: Cloud provider load balancer

### 4. ConfigMap

Non-sensitive configuration:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: auth-service-config
data:
  server-port: "8080"
  aws-region: "us-east-1"
```

**Usage**: Environment variables, configuration files

### 5. Secret

Sensitive data:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: auth-service-secrets
type: Opaque
data:
  db-password: <base64-encoded>
```

**Best Practices**:
- Never commit secrets to Git
- Use external secret management (AWS Secrets Manager)
- Rotate secrets regularly

### 6. Ingress

External access with routing:

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: expense-tracker-ingress
spec:
  rules:
  - http:
      paths:
      - path: /auth
        pathType: Prefix
        backend:
          service:
            name: auth-service
            port:
              number: 8080
```

**Features**:
- Path-based routing
- TLS termination
- Load balancing

## Service Discovery

Services are accessible via DNS:

```
<service-name>.<namespace>.svc.cluster.local
```

**Example**:
```
auth-service.expense-tracker.svc.cluster.local:8080
```

**Benefits**:
- Automatic DNS resolution
- Load balancing across pods
- No hardcoded IPs

## Health Checks

### Liveness Probe

Detects if container is alive:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
```

**Action**: Restart container if unhealthy

### Readiness Probe

Detects if container is ready:

```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
```

**Action**: Remove from service endpoints if not ready

## Resource Management

### Requests

Minimum resources guaranteed:

```yaml
resources:
  requests:
    cpu: 100m      # 0.1 CPU
    memory: 128Mi  # 128 MiB
```

### Limits

Maximum resources allowed:

```yaml
resources:
  limits:
    cpu: 500m      # 0.5 CPU
    memory: 512Mi  # 512 MiB
```

**Why**: Prevent resource exhaustion, ensure fair sharing

## Security

### Service Account

Identity for pods:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: auth-service
  annotations:
    eks.amazonaws.com/role-arn: <IAM_ROLE_ARN>
```

**IRSA**: IAM Roles for Service Accounts (AWS EKS)

### Security Context

Pod and container security:

```yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  fsGroup: 1000
```

**Best Practices**:
- Run as non-root
- Read-only root filesystem (when possible)
- Drop unnecessary capabilities

## Kustomize

Manages environment-specific configurations:

### Base

```yaml
# k8s/base/kustomization.yaml
resources:
  - namespace.yaml
  - auth-service
```

### Overlay

```yaml
# k8s/overlays/production/kustomization.yaml
resources:
  - ../../base
patchesStrategicMerge:
  - replicas.yaml
```

**Benefits**:
- No templating needed
- Environment-specific configs
- Base + overlay pattern

## Common Commands

### View Resources

```bash
kubectl get pods -n expense-tracker
kubectl get svc -n expense-tracker
kubectl get deployments -n expense-tracker
```

### Describe Resource

```bash
kubectl describe pod <pod-name> -n expense-tracker
```

### View Logs

```bash
kubectl logs <pod-name> -n expense-tracker
kubectl logs -f <pod-name> -n expense-tracker  # Follow logs
```

### Execute Command

```bash
kubectl exec -it <pod-name> -n expense-tracker -- /bin/sh
```

### Port Forward

```bash
kubectl port-forward svc/auth-service 8080:8080 -n expense-tracker
```

### Apply Manifest

```bash
kubectl apply -f k8s/base/
```

### Delete Resource

```bash
kubectl delete deployment auth-service -n expense-tracker
```

## Troubleshooting

### Pod Not Starting

1. Check pod status: `kubectl get pods`
2. Describe pod: `kubectl describe pod <pod-name>`
3. View logs: `kubectl logs <pod-name>`
4. Check events: `kubectl get events --sort-by='.lastTimestamp'`

### Service Not Accessible

1. Verify service exists: `kubectl get svc`
2. Check endpoints: `kubectl get endpoints`
3. Verify selector matches pod labels
4. Test DNS resolution from pod

### Image Pull Errors

1. Check image name and tag
2. Verify image exists in registry
3. Check image pull secrets
4. Verify registry credentials

### Resource Limits

1. Check resource requests/limits
2. View node resources: `kubectl top nodes`
3. Check pod resources: `kubectl top pods`
4. Adjust limits if needed

## Best Practices

✅ Use Deployments, not Pods directly  
✅ Set resource requests and limits  
✅ Use health checks (liveness + readiness)  
✅ Run as non-root user  
✅ Use ConfigMaps for configuration  
✅ Use Secrets for sensitive data  
✅ Use namespaces for isolation  
✅ Label resources consistently  
✅ Use Kustomize for environment configs  
✅ Monitor resource usage  

## Additional Resources

- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Kubernetes Concepts](https://kubernetes.io/docs/concepts/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)
