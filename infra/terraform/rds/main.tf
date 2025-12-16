# RDS Module
# This module creates PostgreSQL databases for the microservices
# Creates 3 separate databases: auth-db, expense-db, receipt-db

terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.1"
    }
  }
}

# ============================================================================
# Random Password Generator
# ============================================================================
# Generate secure random passwords for each database
resource "random_password" "auth_db_password" {
  length  = 32
  special = true
}

resource "random_password" "expense_db_password" {
  length  = 32
  special = true
}

resource "random_password" "receipt_db_password" {
  length  = 32
  special = true
}

# ============================================================================
# DB Subnet Group
# ============================================================================
# RDS instances must be in a subnet group (private subnets)
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = var.private_subnet_ids
  
  tags = {
    Name        = "${var.project_name}-db-subnet-group"
    Environment = var.environment
  }
}

# ============================================================================
# Security Group for RDS
# ============================================================================
# Allows access from EKS cluster only
resource "aws_security_group" "rds" {
  name_prefix = "${var.project_name}-rds-"
  vpc_id      = var.vpc_id
  description = "Security group for RDS PostgreSQL databases"
  
  # Allow PostgreSQL (port 5432) from EKS cluster security group
  ingress {
    description     = "PostgreSQL from EKS cluster"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [var.eks_cluster_security_group_id]
  }
  
  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
  
  tags = {
    Name        = "${var.project_name}-rds-sg"
    Environment = var.environment
  }
}

# ============================================================================
# Secrets Manager Secrets
# ============================================================================
# Store database credentials securely in AWS Secrets Manager

# Auth DB Secret
resource "aws_secretsmanager_secret" "auth_db" {
  name        = "${var.project_name}/auth-db/credentials"
  description = "Credentials for auth-service database"
  
  tags = {
    Name        = "${var.project_name}-auth-db-secret"
    Environment = var.environment
  }
}

resource "aws_secretsmanager_secret_version" "auth_db" {
  secret_id = aws_secretsmanager_secret.auth_db.id
  secret_string = jsonencode({
    username = var.db_username
    password = random_password.auth_db_password.result
    engine   = "postgres"
    host     = aws_db_instance.auth_db.endpoint
    port     = aws_db_instance.auth_db.port
    dbname   = "auth_db"
  })
}

# Expense DB Secret
resource "aws_secretsmanager_secret" "expense_db" {
  name        = "${var.project_name}/expense-db/credentials"
  description = "Credentials for expense-service database"
  
  tags = {
    Name        = "${var.project_name}-expense-db-secret"
    Environment = var.environment
  }
}

resource "aws_secretsmanager_secret_version" "expense_db" {
  secret_id = aws_secretsmanager_secret.expense_db.id
  secret_string = jsonencode({
    username = var.db_username
    password = random_password.expense_db_password.result
    engine   = "postgres"
    host     = aws_db_instance.expense_db.endpoint
    port     = aws_db_instance.expense_db.port
    dbname   = "expense_db"
  })
}

# Receipt DB Secret
resource "aws_secretsmanager_secret" "receipt_db" {
  name        = "${var.project_name}/receipt-db/credentials"
  description = "Credentials for receipt-service database"
  
  tags = {
    Name        = "${var.project_name}-receipt-db-secret"
    Environment = var.environment
  }
}

resource "aws_secretsmanager_secret_version" "receipt_db" {
  secret_id = aws_secretsmanager_secret.receipt_db.id
  secret_string = jsonencode({
    username = var.db_username
    password = random_password.receipt_db_password.result
    engine   = "postgres"
    host     = aws_db_instance.receipt_db.endpoint
    port     = aws_db_instance.receipt_db.port
    dbname   = "receipt_db"
  })
}

# ============================================================================
# RDS Instances
# ============================================================================

# Auth Service Database
resource "aws_db_instance" "auth_db" {
  identifier = "${var.project_name}-auth-db"
  
  # Engine configuration
  engine         = "postgres"
  engine_version = var.postgres_version
  instance_class = var.db_instance_class
  
  # Database configuration
  db_name  = "auth_db"
  username = var.db_username
  password = random_password.auth_db_password.result
  
  # Storage configuration
  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_type         = "gp3"
  storage_encrypted     = true
  
  # Network configuration
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false  # Keep databases private
  
  # Backup configuration
  backup_retention_period = var.backup_retention_days
  backup_window          = "03:00-04:00"  # UTC
  maintenance_window     = "mon:04:00-mon:05:00"  # UTC
  
  # High availability (optional, costs more)
  multi_az = var.enable_multi_az
  
  # Deletion protection
  deletion_protection = var.deletion_protection
  skip_final_snapshot = !var.deletion_protection
  
  # Monitoring
  enabled_cloudwatch_logs_exports = ["postgresql"]
  monitoring_interval            = 60
  monitoring_role_arn            = aws_iam_role.rds_monitoring.arn
  
  tags = {
    Name        = "${var.project_name}-auth-db"
    Environment = var.environment
    Service     = "auth-service"
  }
}

# Expense Service Database
resource "aws_db_instance" "expense_db" {
  identifier = "${var.project_name}-expense-db"
  
  engine         = "postgres"
  engine_version = var.postgres_version
  instance_class = var.db_instance_class
  
  db_name  = "expense_db"
  username = var.db_username
  password = random_password.expense_db_password.result
  
  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_type         = "gp3"
  storage_encrypted     = true
  
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false
  
  backup_retention_period = var.backup_retention_days
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"
  
  multi_az = var.enable_multi_az
  
  deletion_protection = var.deletion_protection
  skip_final_snapshot = !var.deletion_protection
  
  enabled_cloudwatch_logs_exports = ["postgresql"]
  monitoring_interval            = 60
  monitoring_role_arn            = aws_iam_role.rds_monitoring.arn
  
  tags = {
    Name        = "${var.project_name}-expense-db"
    Environment = var.environment
    Service     = "expense-service"
  }
}

# Receipt Service Database
resource "aws_db_instance" "receipt_db" {
  identifier = "${var.project_name}-receipt-db"
  
  engine         = "postgres"
  engine_version = var.postgres_version
  instance_class = var.db_instance_class
  
  db_name  = "receipt_db"
  username = var.db_username
  password = random_password.receipt_db_password.result
  
  allocated_storage     = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_type         = "gp3"
  storage_encrypted     = true
  
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false
  
  backup_retention_period = var.backup_retention_days
  backup_window          = "03:00-04:00"
  maintenance_window     = "mon:04:00-mon:05:00"
  
  multi_az = var.enable_multi_az
  
  deletion_protection = var.deletion_protection
  skip_final_snapshot = !var.deletion_protection
  
  enabled_cloudwatch_logs_exports = ["postgresql"]
  monitoring_interval            = 60
  monitoring_role_arn            = aws_iam_role.rds_monitoring.arn
  
  tags = {
    Name        = "${var.project_name}-receipt-db"
    Environment = var.environment
    Service     = "receipt-service"
  }
}
