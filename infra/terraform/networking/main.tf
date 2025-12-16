# Networking Module
# This module creates the VPC foundation for the EKS cluster
# It includes: VPC, subnets, Internet Gateway, NAT Gateways, and route tables

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

# ============================================================================
# VPC
# ============================================================================
# The VPC (Virtual Private Cloud) is your isolated network in AWS
# CIDR 10.0.0.0/16 gives us 65,536 IP addresses (10.0.0.0 to 10.0.255.255)
resource "aws_vpc" "main" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true  # Required for EKS
  enable_dns_support   = true  # Required for EKS
  
  tags = {
    Name = "${var.project_name}-vpc"
    Environment = var.environment
    ManagedBy = "terraform"
  }
}

# ============================================================================
# Internet Gateway
# ============================================================================
# Allows resources in public subnets to access the internet
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  
  tags = {
    Name = "${var.project_name}-igw"
    Environment = var.environment
  }
}

# ============================================================================
# Availability Zones
# ============================================================================
# Get available AZs in the region for high availability
data "aws_availability_zones" "available" {
  state = "available"
}

# ============================================================================
# Public Subnets
# ============================================================================
# Public subnets have direct internet access via Internet Gateway
# Used for: ALB (Application Load Balancer), NAT Gateways
resource "aws_subnet" "public" {
  count = length(var.public_subnet_cidrs)
  
  vpc_id                  = aws_vpc.main.id
  cidr_block              = var.public_subnet_cidrs[count.index]
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  map_public_ip_on_launch = true  # Auto-assign public IPs
  
  tags = {
    Name = "${var.project_name}-public-subnet-${count.index + 1}"
    Environment = var.environment
    Type = "public"
    "kubernetes.io/role/elb" = "1"  # Tag for ALB Ingress Controller
  }
}

# ============================================================================
# Private Subnets
# ============================================================================
# Private subnets have no direct internet access
# Used for: EKS nodes, RDS databases (more secure)
resource "aws_subnet" "private" {
  count = length(var.private_subnet_cidrs)
  
  vpc_id            = aws_vpc.main.id
  cidr_block        = var.private_subnet_cidrs[count.index]
  availability_zone = data.aws_availability_zones.available.names[count.index]
  
  tags = {
    Name = "${var.project_name}-private-subnet-${count.index + 1}"
    Environment = var.environment
    Type = "private"
    "kubernetes.io/role/internal-elb" = "1"  # Tag for internal load balancers
  }
}

# ============================================================================
# Elastic IPs for NAT Gateways
# ============================================================================
# NAT Gateways need static public IPs
resource "aws_eip" "nat" {
  count = var.enable_nat_gateway ? length(aws_subnet.public) : 0
  
  domain = "vpc"
  
  tags = {
    Name = "${var.project_name}-nat-eip-${count.index + 1}"
    Environment = var.environment
  }
  
  depends_on = [aws_internet_gateway.main]
}

# ============================================================================
# NAT Gateways
# ============================================================================
# NAT Gateways allow resources in private subnets to access the internet
# (for downloading packages, updates, etc.) while keeping them private
resource "aws_nat_gateway" "main" {
  count = var.enable_nat_gateway ? length(aws_subnet.public) : 0
  
  allocation_id = aws_eip.nat[count.index].id
  subnet_id     = aws_subnet.public[count.index].id
  
  tags = {
    Name = "${var.project_name}-nat-gateway-${count.index + 1}"
    Environment = var.environment
  }
  
  depends_on = [aws_internet_gateway.main]
}

# ============================================================================
# Route Tables
# ============================================================================

# Public Route Table
# Routes traffic from public subnets to Internet Gateway
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  
  tags = {
    Name = "${var.project_name}-public-rt"
    Environment = var.environment
  }
}

# Associate public subnets with public route table
resource "aws_route_table_association" "public" {
  count = length(aws_subnet.public)
  
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# Private Route Tables
# Routes traffic from private subnets to NAT Gateway
resource "aws_route_table" "private" {
  count = var.enable_nat_gateway ? length(aws_subnet.private) : 0
  
  vpc_id = aws_vpc.main.id
  
  route {
    cidr_block     = "0.0.0.0/0"
    nat_gateway_id = aws_nat_gateway.main[count.index].id
  }
  
  tags = {
    Name = "${var.project_name}-private-rt-${count.index + 1}"
    Environment = var.environment
  }
}

# Associate private subnets with private route tables
resource "aws_route_table_association" "private" {
  count = var.enable_nat_gateway ? length(aws_subnet.private) : 0
  
  subnet_id      = aws_subnet.private[count.index].id
  route_table_id = aws_route_table.private[count.index].id
}

