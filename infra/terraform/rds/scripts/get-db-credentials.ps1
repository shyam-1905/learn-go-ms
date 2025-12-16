# PowerShell script to retrieve RDS credentials from AWS Secrets Manager
# Usage: .\get-db-credentials.ps1 [master|app]

param(
    [Parameter(Position=0)]
    [ValidateSet("master", "app")]
    [string]$SecretType = "master"
)

$ProjectName = if ($env:PROJECT_NAME) { $env:PROJECT_NAME } else { "expense-tracker" }

if ($SecretType -eq "master") {
    $SecretName = "${ProjectName}/rds/master-credentials"
} else {
    $SecretName = "${ProjectName}/rds/app-credentials"
}

Write-Host "Retrieving ${SecretType} credentials from Secrets Manager..." -ForegroundColor Green
Write-Host "Secret: ${SecretName}" -ForegroundColor Cyan
Write-Host ""

try {
    # Get secret value
    $SecretResponse = aws secretsmanager get-secret-value `
        --secret-id $SecretName `
        --query SecretString `
        --output text

    if (-not $SecretResponse) {
        Write-Host "Error: Could not retrieve secret" -ForegroundColor Red
        exit 1
    }

    # Parse JSON
    $SecretJson = $SecretResponse | ConvertFrom-Json

    Write-Host "Credentials:" -ForegroundColor Yellow
    Write-Host "  Host: $($SecretJson.host)"
    Write-Host "  Port: $($SecretJson.port)"
    Write-Host "  Database: $($SecretJson.dbname)"
    Write-Host "  Username: $($SecretJson.username)"
    Write-Host "  Password: ********" -ForegroundColor Gray

    Write-Host ""
    Write-Host "Connection string:" -ForegroundColor Yellow
    Write-Host "postgres://$($SecretJson.username):***@$($SecretJson.host):$($SecretJson.port)/$($SecretJson.dbname)" -ForegroundColor Cyan

} catch {
    Write-Host "Error: $_" -ForegroundColor Red
    exit 1
}
