# RDS Terraform Module - Summary

## 📦 What We Created

A complete Terraform module to provision AWS RDS PostgreSQL for the auth-service, designed to work seamlessly with EKS.

### Files Created

```
infra/terraform/rds/
├── main.tf                    # Main Terraform configuration
├── variables.tf               # Input variables
├── outputs.tf                 # Output values
├── terraform.tfvars.example   # Example configuration
├── .gitignore                 # Git ignore rules
├── README.md                  # Detailed documentation
├── QUICKSTART.md              # Step-by-step setup guide
├── EKS_INTEGRATION.md        # How to connect from EKS
└── scripts/
    ├── get-db-credentials.sh      # Bash script to get credentials
    ├── get-db-credentials.ps1     # PowerShell script to get credentials
    └── setup-k8s-secrets.sh        # Automated K8s secret setup
```

## 🎓 Key Concepts You've Learned

### 1. Infrastructure as Code (IaC)
- **What**: Define infrastructure using code (Terraform)
- **Why**: Version control, repeatability, consistency
- **How**: Write `.tf` files, run `terraform apply`

### 2. AWS RDS
- **What**: Managed PostgreSQL database service
- **Why**: No need to manage database servers yourself
- **Features**: Automated backups, Multi-AZ, encryption, monitoring

### 3. VPC Networking
- **Subnets**: Private subnets for RDS (security)
- **Security Groups**: Firewall rules (allow port 5432 from EKS)
- **Route Tables**: Control network traffic flow

### 4. Secrets Management
- **AWS Secrets Manager**: Store credentials securely
- **Kubernetes Secrets**: Pass secrets to pods
- **Best Practice**: Never hardcode passwords!

### 5. Terraform Concepts
- **Resources**: Things to create (RDS, security groups)
- **Data Sources**: Query existing resources (VPC, subnets)
- **Variables**: Parameterize configuration
- **Outputs**: Expose important values
- **Providers**: AWS, random, etc.

## 🏗️ Architecture

```
┌─────────────────────────────────────────┐
│           AWS VPC                        │
│                                          │
│  ┌──────────────┐      ┌─────────────┐ │
│  │  EKS Cluster  │      │   RDS       │ │
│  │  (Public)    │─────▶│  (Private)  │ │
│  │              │      │             │ │
│  │  auth-service│      │ PostgreSQL  │ │
│  │  pods        │      │             │ │
│  └──────────────┘      └─────────────┘ │
│         │                    │          │
│         └────────────────────┘          │
│      Security Group Rules               │
│      (Port 5432 allowed)                 │
└─────────────────────────────────────────┘
```

## 🔐 Security Features

1. **Private Subnets**: RDS not publicly accessible
2. **Security Groups**: Only EKS can access RDS
3. **Encryption**: Data encrypted at rest
4. **Secrets Manager**: Credentials stored securely
5. **IAM Roles**: Fine-grained permissions (IRSA)

## 📊 Resources Created

When you run `terraform apply`, it creates:

1. **DB Subnet Group**: Tells RDS which subnets to use
2. **Security Group**: Firewall rules for RDS
3. **Parameter Group**: Database configuration settings
4. **RDS Instance**: The PostgreSQL database
5. **Secrets Manager Secrets**: Master and app credentials
6. **IAM Role**: For enhanced monitoring

## 🚀 Quick Commands

```bash
# Initialize Terraform
terraform init

# Plan changes
terraform plan

# Apply configuration
terraform apply

# Get outputs
terraform output

# Destroy everything (careful!)
terraform destroy
```

## 🔗 Integration Points

### With EKS
- Security groups allow EKS → RDS traffic
- Kubernetes secrets store connection details
- Pods connect using environment variables

### With Auth Service
- Reads `DB_HOST`, `DB_USER`, etc. from env vars
- Connects using `pgxpool` library
- Runs migrations on startup (or via Job)

## 📚 Learning Path

1. ✅ **Created Terraform config** - Infrastructure as Code
2. ✅ **Understood VPC networking** - Subnets, security groups
3. ✅ **Learned RDS concepts** - Managed databases
4. ✅ **Secrets management** - Secure credential storage
5. ⏭️ **Next: Deploy to EKS** - Connect service to database

## 🎯 Next Steps

1. **Run Terraform** to create RDS
2. **Set up Kubernetes secrets** using the scripts
3. **Deploy auth-service** to EKS
4. **Run migrations** to create tables
5. **Test the connection** and API endpoints

## 💡 Pro Tips

- **Start small**: Use `db.t3.micro` for dev
- **Enable Multi-AZ** for production
- **Use Secrets Manager** instead of hardcoding
- **Monitor costs**: RDS can be expensive
- **Set deletion protection** in production
- **Regular backups**: Configure retention period

## 🆘 Common Questions

**Q: How much does this cost?**
A: `db.t3.micro` is ~$15/month. Production instances cost more.

**Q: Can I use this for other services?**
A: Yes! Create separate databases or instances as needed.

**Q: How do I scale?**
A: Change `instance_class` variable and apply. Or use read replicas.

**Q: What about backups?**
A: Automated daily backups. Retention configurable (default: 7 days).

**Q: Can I access from my local machine?**
A: Not directly (private subnet). Use port forwarding from EKS pod or bastion host.

## 📖 Additional Resources

- [Terraform AWS Provider Docs](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)
- [AWS RDS User Guide](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/)
- [Kubernetes Secrets](https://kubernetes.io/docs/concepts/configuration/secret/)
- [EKS Best Practices](https://aws.github.io/aws-eks-best-practices/)
