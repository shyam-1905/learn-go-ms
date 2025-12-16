# Networking Module

This Terraform module creates the VPC (Virtual Private Cloud) foundation for the EKS cluster.

## What it creates:

1. **VPC** - Isolated network in AWS (10.0.0.0/16)
2. **Public Subnets** - For ALB and NAT Gateways (3 subnets, one per AZ)
3. **Private Subnets** - For EKS nodes and RDS (3 subnets, one per AZ)
4. **Internet Gateway** - Allows public subnets to access internet
5. **NAT Gateways** - Allows private subnets to access internet (outbound only)
6. **Route Tables** - Routes traffic between subnets and gateways

## Usage:

```hcl
module "networking" {
  source = "./networking"
  
  project_name = "expense-tracker"
  environment  = "production"
  vpc_cidr     = "10.0.0.0/16"
  
  public_subnet_cidrs  = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
  private_subnet_cidrs = ["10.0.11.0/24", "10.0.12.0/24", "10.0.13.0/24"]
  
  enable_nat_gateway = true
}
```

## Cost Considerations:

- NAT Gateways cost ~$32/month each + data transfer
- For cost savings in dev/staging, set `enable_nat_gateway = false` or use single NAT Gateway

