# PowerShell script untuk menambahkan AUTO_INCREMENT pada semua tabel
# Usage: .\migrations\fix_auto_increment.ps1

$DOCKER_CONTAINER = "vla-mysql"
$DB_USER = "root"
$DB_PASS = "vlaroot"
$DB_NAME = "vla-db"
$SQL_FILE = "migrations\fix_auto_increment.sql"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Fix AUTO_INCREMENT untuk Semua Tabel" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Docker container is running
Write-Host "Checking Docker container..." -ForegroundColor Yellow
$containerStatus = docker ps --filter "name=$DOCKER_CONTAINER" --format "{{.Status}}"
if (-not $containerStatus) {
    Write-Host "Error: Docker container '$DOCKER_CONTAINER' is not running!" -ForegroundColor Red
    Write-Host "Please start the container first: docker start $DOCKER_CONTAINER" -ForegroundColor Yellow
    exit 1
}
Write-Host "Container is running: $containerStatus" -ForegroundColor Green
Write-Host ""

# Check if database exists
Write-Host "Checking database existence..." -ForegroundColor Yellow
$dbExists = docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS -e "SHOW DATABASES LIKE '$DB_NAME';" 2>&1 | Select-String -Pattern $DB_NAME
if (-not $dbExists) {
    Write-Host "Error: Database '$DB_NAME' does not exist!" -ForegroundColor Red
    exit 1
}
Write-Host "Database '$DB_NAME' found." -ForegroundColor Green
Write-Host ""

# Check if SQL file exists
if (-not (Test-Path $SQL_FILE)) {
    Write-Host "Error: SQL file not found: $SQL_FILE" -ForegroundColor Red
    exit 1
}
Write-Host "SQL file found: $SQL_FILE" -ForegroundColor Green
Write-Host ""

# Execute SQL file
Write-Host "Executing SQL commands to fix AUTO_INCREMENT..." -ForegroundColor Yellow
Write-Host "This will add AUTO_INCREMENT to all tables that don't have it..." -ForegroundColor Yellow
Write-Host ""

# Get absolute path of SQL file
$absolutePath = (Resolve-Path $SQL_FILE).Path

# Execute SQL file
Get-Content $absolutePath | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME 2>&1 | Out-Null

if ($LASTEXITCODE -eq 0) {
    Write-Host "AUTO_INCREMENT fixed successfully!" -ForegroundColor Green
} else {
    Write-Host "Error: Failed to execute SQL commands!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please try manually:" -ForegroundColor Yellow
    Write-Host "  Get-Content $SQL_FILE | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME" -ForegroundColor Cyan
    exit 1
}

Write-Host ""
Write-Host "Verifying AUTO_INCREMENT on key tables..." -ForegroundColor Yellow
$tables = @("users", "admins", "transactions", "investments", "deposits")
foreach ($table in $tables) {
    $result = docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW CREATE TABLE \`$table\`;" 2>&1 | Select-String -Pattern "AUTO_INCREMENT"
    if ($result) {
        Write-Host "  ✓ $table has AUTO_INCREMENT" -ForegroundColor Green
    } else {
        Write-Host "  ✗ $table missing AUTO_INCREMENT" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Fix completed!" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

