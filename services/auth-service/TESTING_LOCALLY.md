# Testing Auth Service Locally with AWS RDS

This guide helps you test the auth-service locally using a demo RDS PostgreSQL database in AWS.

## 🎯 Goal

Test the auth-service on your local machine connected to a PostgreSQL database in AWS RDS.

## 📋 Step 1: Create Demo RDS Database in AWS Console

### Option A: AWS Console (Easiest)

1. **Go to AWS Console** → RDS → Databases → Create database

2. **Choose settings:**
   - **Engine**: PostgreSQL
   - **Version**: 15.4 (or latest)
   - **Template**: Free tier (if available) or Dev/Test
   - **DB instance identifier**: `expense-tracker-demo`
   - **Master username**: `postgres`
   - **Master password**: Create a strong password (save it!)

3. **Instance configuration:**
   - **DB instance class**: `db.t3.micro` (smallest/cheapest)
   - **Storage**: 20 GB (minimum)

4. **Connectivity:**
   - **VPC**: Choose default VPC (or your VPC)
   - **Public access**: ✅ **YES** (for local testing)
   - **VPC security group**: Create new or use existing
   - **Database port**: 5432

5. **Database authentication**: Password authentication

6. **Additional configuration:**
   - **Initial database name**: `auth_db`
   - **Backup retention**: 0 days (for demo - no backups needed)
   - **Enable Enhanced monitoring**: ❌ No (saves money)

7. **Click "Create database"** (takes 5-10 minutes)

### Option B: AWS CLI (Faster)

```bash
aws rds create-db-instance \
  --db-instance-identifier expense-tracker-demo \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --engine-version 15.4 \
  --master-username postgres \
  --master-user-password YourStrongPassword123! \
  --allocated-storage 20 \
  --db-name auth_db \
  --publicly-accessible \
  --backup-retention-period 0 \
  --no-multi-az
```

## 🔐 Step 2: Configure Security Group

After RDS is created, you need to allow your IP to connect:

1. **Go to EC2 Console** → Security Groups
2. **Find the security group** attached to your RDS instance
3. **Edit inbound rules** → Add rule:
   - **Type**: PostgreSQL
   - **Port**: 5432
   - **Source**: My IP (or `0.0.0.0/0` for testing - **not secure for production!**)

**Or via AWS CLI:**
```bash
# Get your public IP
MY_IP=$(curl -s https://checkip.amazonaws.com)

# Get security group ID from RDS
SG_ID=$(aws rds describe-db-instances \
  --db-instance-identifier expense-tracker-demo \
  --query 'DBInstances[0].VpcSecurityGroups[0].VpcSecurityGroupId' \
  --output text)

# Add rule
aws ec2 authorize-security-group-ingress \
  --group-id $SG_ID \
  --protocol tcp \
  --port 5432 \
  --cidr $MY_IP/32
```

## 📝 Step 3: Get Connection Details

Once the database is **available** (status = "Available"):

1. **Go to RDS Console** → Databases → Click your database
2. **Copy the endpoint** (e.g., `expense-tracker-demo.xxxxx.us-east-1.rds.amazonaws.com`)
3. **Note the port**: 5432
4. **You already have**: username (`postgres`) and password

## 🗄️ Step 4: Run Database Migration

### Install PostgreSQL Client (if needed)

**Windows:**
- Download from: https://www.postgresql.org/download/windows/
- Or use: `choco install postgresql`

**Mac:**
```bash
brew install postgresql
```

**Linux:**
```bash
sudo apt-get install postgresql-client
```

### Run Migration

```bash
# Set environment variables
export DB_HOST=expense-tracker-demo.xxxxx.us-east-1.rds.amazonaws.com
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=YourPasswordHere
export DB_NAME=auth_db

# Run migration
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_create_users_table.sql
```

**Windows PowerShell:**
```powershell
$env:PGPASSWORD="YourPasswordHere"
psql -h expense-tracker-demo.xxxxx.us-east-1.rds.amazonaws.com -U postgres -d auth_db -f migrations/001_create_users_table.sql
```

## 🚀 Step 5: Test Auth Service Locally

### Set Environment Variables

**Windows PowerShell:**
```powershell
$env:DB_HOST="expense-tracker-demo.xxxxx.us-east-1.rds.amazonaws.com"
$env:DB_PORT="5432"
$env:DB_USER="postgres"
$env:DB_PASSWORD="YourPasswordHere"
$env:DB_NAME="auth_db"
$env:JWT_SECRET="your-super-secret-jwt-key-change-this"
$env:JWT_EXPIRATION_HOURS="24"
$env:SERVER_PORT="8080"
```

**Linux/Mac:**
```bash
export DB_HOST=expense-tracker-demo.xxxxx.us-east-1.rds.amazonaws.com
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=YourPasswordHere
export DB_NAME=auth_db
export JWT_SECRET=your-super-secret-jwt-key-change-this
export JWT_EXPIRATION_HOURS=24
export SERVER_PORT=8080
```

### Run the Service

```bash
cd services/auth-service
go run cmd/main.go
```

You should see:
```
Connecting to database...
Database connection established!
Server starting on port 8080...
Server is running. Press Ctrl+C to stop.
```

## 🧪 Step 6: Test the API

### Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{"status":"healthy"}
```

### Register a User
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "securepassword123",
    "name": "Test User"
  }'
```

Expected response:
```json
{
  "user_id": "uuid-here",
  "email": "test@example.com",
  "name": "Test User",
  "token": "jwt-token-here"
}
```

### Login
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "test@example.com",
    "password": "securepassword123"
  }'
```

### Validate Token
```bash
# Replace TOKEN with the token from register/login
curl -X GET http://localhost:8080/auth/validate \
  -H "Authorization: Bearer TOKEN"
```

## 📝 Step 7: Create .env File (Optional)

Create a `.env` file in `services/auth-service/` for easier testing:

```env
DB_HOST=expense-tracker-demo.xxxxx.us-east-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=YourPasswordHere
DB_NAME=auth_db
JWT_SECRET=your-super-secret-jwt-key-change-this
JWT_EXPIRATION_HOURS=24
SERVER_PORT=8080
```

Then use a tool like `godotenv` to load it, or manually export before running.

## 🔧 Troubleshooting

### Connection Timeout

**Problem**: Can't connect to RDS

**Solutions**:
1. Check security group allows your IP
2. Verify RDS is publicly accessible
3. Check RDS status is "Available" (not "Creating")
4. Verify endpoint is correct

```bash
# Test connection
psql -h <endpoint> -U postgres -d auth_db
```

### Authentication Failed

**Problem**: Wrong password

**Solutions**:
1. Verify password is correct (no extra spaces)
2. Check username is `postgres`
3. Try resetting password in AWS Console

### Database Doesn't Exist

**Problem**: `database "auth_db" does not exist`

**Solutions**:
1. Check initial database name in RDS settings
2. Or create it manually:
```sql
CREATE DATABASE auth_db;
```

## 💰 Cost Considerations

- **db.t3.micro**: ~$15/month (or free tier eligible)
- **Storage**: ~$0.10/GB/month (20 GB = ~$2/month)
- **Total**: ~$17/month for demo

**To minimize costs:**
- Delete database when not testing
- Use free tier if eligible
- Set backup retention to 0

## 🧹 Cleanup

When done testing:

```bash
# Delete RDS instance
aws rds delete-db-instance \
  --db-instance-identifier expense-tracker-demo \
  --skip-final-snapshot
```

Or delete from AWS Console → RDS → Databases → Select → Actions → Delete

## ✅ Checklist

- [ ] RDS database created and available
- [ ] Security group allows your IP
- [ ] Database migration completed
- [ ] Environment variables set
- [ ] Auth service running
- [ ] Health endpoint works
- [ ] Can register a user
- [ ] Can login
- [ ] Can validate token

## 🎓 What You've Learned

- How to create RDS in AWS Console
- How to configure security groups
- How to connect from local machine
- How to run database migrations
- How to test the auth-service locally
- How to use environment variables in Go

## 🚀 Next Steps

Once testing works locally:
1. Set up EKS cluster
2. Use Terraform to create production RDS
3. Deploy auth-service to EKS
4. Connect EKS to RDS (private networking)
