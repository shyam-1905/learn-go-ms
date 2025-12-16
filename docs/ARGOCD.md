# ArgoCD Guide

This guide explains GitOps concepts and how ArgoCD manages our deployments.

## What is GitOps?

GitOps is a methodology for managing infrastructure and applications using Git as the single source of truth. Changes to infrastructure or applications are made through Git commits, and an automated process syncs the Git state to the cluster.

### Key Principles

1. **Git as Source of Truth**: All configurations in Git
2. **Declarative**: Describe desired state, not how to achieve it
3. **Automated**: Changes automatically applied
4. **Observable**: Monitor sync status and health

## What is ArgoCD?

ArgoCD is a declarative, GitOps continuous delivery tool for Kubernetes. It monitors Git repositories and automatically syncs the desired state to Kubernetes clusters.

### Key Features

- **Automatic Sync**: Detects Git changes and applies them
- **Self-Healing**: Automatically fixes drift
- **Rollback**: Revert to previous Git commit
- **Multi-Environment**: Manage multiple clusters/environments
- **Web UI**: Visual interface for monitoring

## Architecture

```
Git Repository (Source of Truth)
        ↓
   ArgoCD Server
        ↓
   Kubernetes API
        ↓
   Pods/Services/etc.
```

## Core Concepts

### 1. Application

Represents a deployment from a Git repository:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: expense-tracker-app
spec:
  project: expense-tracker
  source:
    repoURL: https://github.com/user/repo.git
    targetRevision: main
    path: k8s/overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: expense-tracker
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
```

### 2. Project

Defines which repositories and namespaces ArgoCD can deploy to:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: AppProject
metadata:
  name: expense-tracker
spec:
  sourceRepos:
    - '*'
  destinations:
    - namespace: expense-tracker
      server: https://kubernetes.default.svc
```

### 3. Sync Policy

Controls how ArgoCD syncs:

- **Manual**: Require manual sync
- **Automated**: Auto-sync on Git changes
- **Prune**: Delete resources not in Git
- **Self-Heal**: Fix drift automatically

## Installation

ArgoCD is installed via Terraform using Helm:

```hcl
resource "helm_release" "argocd" {
  name       = "argocd"
  repository = "https://argoproj.github.io/argo-helm"
  chart      = "argo-cd"
  namespace  = "argocd"
  version    = "5.51.6"
}
```

## Accessing ArgoCD

### 1. Get Admin Password

```bash
terraform output -raw argocd_admin_password | base64 -d
```

### 2. Port Forward

```bash
kubectl port-forward svc/argocd-server -n argocd 8080:443
```

### 3. Access UI

- URL: `https://localhost:8080`
- Username: `admin`
- Password: (from step 1)

## Workflow

### 1. Make Changes

Edit Kubernetes manifests in Git:

```bash
git add k8s/base/auth-service/deployment.yaml
git commit -m "Update auth-service replicas"
git push
```

### 2. ArgoCD Detects Changes

ArgoCD polls Git repository (default: every 3 minutes)

### 3. Automatic Sync

If automated sync is enabled, ArgoCD applies changes automatically

### 4. Monitor Status

Check sync status in ArgoCD UI or CLI:

```bash
argocd app get expense-tracker-app
```

## Sync Status

### Healthy

All resources match Git state

### Syncing

Currently applying changes

### OutOfSync

Git has changes not yet applied

### Degraded

Resources exist but are unhealthy

### Unknown

Status cannot be determined

## Manual Operations

### Sync Application

```bash
argocd app sync expense-tracker-app
```

### Rollback

```bash
# List history
argocd app history expense-tracker-app

# Rollback to previous version
argocd app rollback expense-tracker-app <REVISION>
```

### Refresh

Force ArgoCD to check Git:

```bash
argocd app get expense-tracker-app --refresh
```

## Best Practices

### 1. Use Kustomize

Structure manifests with Kustomize:

```
k8s/
├── base/           # Base manifests
└── overlays/
    ├── production/ # Production overlay
    └── staging/    # Staging overlay
```

### 2. Enable Automated Sync

For continuous deployment:

```yaml
syncPolicy:
  automated:
    prune: true
    selfHeal: true
```

### 3. Use Sync Windows

Prevent syncs during maintenance:

```yaml
syncWindows:
  - kind: allow
    schedule: '* * * * *'
    duration: 1h
```

### 4. Health Checks

Ensure resources are healthy:

```yaml
syncPolicy:
  syncOptions:
    - CreateNamespace=true
    - PrunePropagationPolicy=foreground
```

### 5. Resource Hooks

Run jobs before/after sync:

```yaml
metadata:
  annotations:
    argocd.argoproj.io/hook: PreSync
    argocd.argoproj.io/hook-delete-policy: BeforeHookCreation
```

## Troubleshooting

### Application Out of Sync

1. Check Git repository access
2. Verify path is correct
3. Check Kustomize build errors
4. Review application events in UI

### Sync Fails

1. Check resource errors:
   ```bash
   argocd app get expense-tracker-app
   ```

2. View application logs:
   ```bash
   kubectl logs -n argocd -l app.kubernetes.io/name=argocd-application-controller
   ```

3. Check resource status:
   ```bash
   kubectl get all -n expense-tracker
   ```

### Self-Heal Not Working

1. Verify self-heal is enabled
2. Check if resource is managed by ArgoCD
3. Review sync policy settings

## CLI Installation

Install ArgoCD CLI:

```bash
# macOS
brew install argocd

# Linux
curl -sSL -o /usr/local/bin/argocd https://github.com/argoproj/argo-cd/releases/latest/download/argocd-linux-amd64
chmod +x /usr/local/bin/argocd
```

### Login

```bash
argocd login localhost:8080
```

## Advanced Features

### Multi-Cluster

Deploy to multiple clusters:

```yaml
destination:
  server: https://cluster2.example.com
  namespace: expense-tracker
```

### Application Sets

Manage multiple applications:

```yaml
apiVersion: argoproj.io/v1alpha1
kind: ApplicationSet
metadata:
  name: expense-tracker-services
spec:
  generators:
    - list:
        elements:
          - name: auth-service
          - name: expense-service
  template:
    metadata:
      name: '{{name}}'
    spec:
      # ... application spec
```

### RBAC

Control access:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: argocd-rbac-cm
data:
  policy.csv: |
    p, role:org-admin, applications, *, */*, allow
    g, alice, role:org-admin
```

## Additional Resources

- [ArgoCD Documentation](https://argo-cd.readthedocs.io/)
- [GitOps Principles](https://www.gitops.tech/)
- [ArgoCD Best Practices](https://argo-cd.readthedocs.io/en/stable/user-guide/best_practices/)
