# PowerShell script to check RDS connection and security group
# This helps diagnose connection issues

$RDS_ENDPOINT = "test-db.c8bs0qiu4gv4.us-east-1.rds.amazonaws.com"

Write-Host "Checking RDS connection..." -ForegroundColor Cyan
Write-Host ""

# Get your public IP
Write-Host "1. Getting your public IP address..." -ForegroundColor Yellow
try {
    $MyIP = (Invoke-WebRequest -Uri "https://checkip.amazonaws.com" -UseBasicParsing).Content.Trim()
    Write-Host "   Your IP: $MyIP" -ForegroundColor Green
} catch {
    Write-Host "   Could not determine your IP" -ForegroundColor Red
    $MyIP = "unknown"
}

Write-Host ""

# Check if we can resolve the DNS
Write-Host "2. Checking DNS resolution..." -ForegroundColor Yellow
try {
    $Resolved = [System.Net.Dns]::GetHostAddresses($RDS_ENDPOINT)
    Write-Host "   ✅ DNS resolved successfully" -ForegroundColor Green
    Write-Host "   IP Address: $($Resolved[0].IPAddressToString)" -ForegroundColor Gray
} catch {
    Write-Host "   ❌ Could not resolve DNS" -ForegroundColor Red
}

Write-Host ""

# Try to connect to port 5432
Write-Host "3. Testing connection to port 5432..." -ForegroundColor Yellow
try {
    $TCPClient = New-Object System.Net.Sockets.TcpClient
    $Connection = $TCPClient.BeginConnect($RDS_ENDPOINT, 5432, $null, $null)
    $Wait = $Connection.AsyncWaitHandle.WaitOne(5000, $false)
    
    if ($Wait) {
        $TCPClient.EndConnect($Connection)
        Write-Host "   ✅ Port 5432 is reachable!" -ForegroundColor Green
        $TCPClient.Close()
    } else {
        Write-Host "   ❌ Connection timeout - port 5432 is not accessible" -ForegroundColor Red
        Write-Host "   This usually means:" -ForegroundColor Yellow
        Write-Host "   - Security group doesn't allow your IP ($MyIP)" -ForegroundColor Yellow
        Write-Host "   - RDS is not publicly accessible" -ForegroundColor Yellow
        Write-Host "   - RDS is still starting up" -ForegroundColor Yellow
    }
} catch {
    Write-Host "   ❌ Connection failed: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "1. Go to AWS Console → RDS → Databases → Select your database" -ForegroundColor White
Write-Host "2. Check 'Publicly accessible' is set to 'Yes'" -ForegroundColor White
Write-Host "3. Go to EC2 → Security Groups → Find the security group attached to RDS" -ForegroundColor White
Write-Host "4. Edit inbound rules → Add rule:" -ForegroundColor White
Write-Host "   - Type: PostgreSQL" -ForegroundColor Gray
Write-Host "   - Port: 5432" -ForegroundColor Gray
Write-Host "   - Source: $MyIP/32 (or 0.0.0.0/0 for testing)" -ForegroundColor Gray
Write-Host ""
Write-Host "Or use AWS CLI:" -ForegroundColor Cyan
Write-Host "aws ec2 authorize-security-group-ingress --group-id <sg-id> --protocol tcp --port 5432 --cidr $MyIP/32" -ForegroundColor Gray
