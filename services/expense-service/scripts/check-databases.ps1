# PowerShell script to check databases and their schemas
# Verifies both auth_db and expense_db exist and have correct tables

$DB_HOST = "test-db.c8bs0qiu4gv4.us-east-1.rds.amazonaws.com"
$DB_PORT = "5432"
$DB_USER = "postgres"
$DB_PASSWORD = "Adminpass123"

Write-Host "Checking databases on RDS..." -ForegroundColor Cyan
Write-Host "Host: $DB_HOST" -ForegroundColor Gray
Write-Host ""

# Check if psql is available
if (-not (Get-Command psql -ErrorAction SilentlyContinue)) {
    Write-Host "⚠️  psql not found. Installing PostgreSQL client..." -ForegroundColor Yellow
    Write-Host "You can install it from: https://www.postgresql.org/download/windows/" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Or use the Go migration tool to verify:" -ForegroundColor Cyan
    Write-Host "  go run cmd/migrate.go" -ForegroundColor Gray
    exit 0
}

# Set password environment variable
$env:PGPASSWORD = $DB_PASSWORD

Write-Host "📊 Database List:" -ForegroundColor Yellow
$dbList = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d postgres -t -c "SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname;"
Write-Host $dbList
Write-Host ""

# Check auth_db
Write-Host "🔍 Checking auth_db..." -ForegroundColor Yellow
$authTables = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d auth_db -t -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;"
if ($authTables) {
    Write-Host "✅ auth_db exists" -ForegroundColor Green
    Write-Host "Tables:" -ForegroundColor Cyan
    $authTables.Trim() | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
    
    # Check users table structure
    Write-Host ""
    Write-Host "Users table columns:" -ForegroundColor Cyan
    $userColumns = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d auth_db -t -c "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'users' ORDER BY ordinal_position;"
    $userColumns.Trim() | ForEach-Object { Write-Host "  $_" -ForegroundColor Gray }
} else {
    Write-Host "❌ auth_db not found or empty" -ForegroundColor Red
}
Write-Host ""

# Check expense_db
Write-Host "🔍 Checking expense_db..." -ForegroundColor Yellow
$expenseTables = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d expense_db -t -c "SELECT table_name FROM information_schema.tables WHERE table_schema = 'public' ORDER BY table_name;"
if ($expenseTables) {
    Write-Host "✅ expense_db exists" -ForegroundColor Green
    Write-Host "Tables:" -ForegroundColor Cyan
    $expenseTables.Trim() | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
    
    # Check expenses table structure
    Write-Host ""
    Write-Host "Expenses table columns:" -ForegroundColor Cyan
    $expenseColumns = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d expense_db -t -c "SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'expenses' ORDER BY ordinal_position;"
    $expenseColumns.Trim() | ForEach-Object { Write-Host "  $_" -ForegroundColor Gray }
    
    # Check indexes
    Write-Host ""
    Write-Host "Indexes on expenses table:" -ForegroundColor Cyan
    $indexes = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d expense_db -t -c "SELECT indexname FROM pg_indexes WHERE tablename = 'expenses' ORDER BY indexname;"
    $indexes.Trim() | ForEach-Object { Write-Host "  - $_" -ForegroundColor Gray }
} else {
    Write-Host "❌ expense_db not found or empty" -ForegroundColor Red
}

Write-Host ""
Write-Host "✅ Database check completed!" -ForegroundColor Green
