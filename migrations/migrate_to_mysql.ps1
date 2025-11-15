# PowerShell script untuk migrasi database ke MySQL via Docker
# Usage: .\migrate_to_mysql.ps1

$DOCKER_CONTAINER = "vla-mysql"
$DB_USER = "root"
$DB_PASS = "vlaroot"
$DB_NAME = "vla-db"
$SQL_FILE = "database\db.sql"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Migrasi Database ke MySQL via Docker" -ForegroundColor Cyan
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

# Check if SQL file exists
if (-not (Test-Path $SQL_FILE)) {
    Write-Host "Error: SQL file not found: $SQL_FILE" -ForegroundColor Red
    exit 1
}
Write-Host "SQL file found: $SQL_FILE" -ForegroundColor Green
Write-Host ""

# Check if database exists, if not create it
Write-Host "Checking database existence..." -ForegroundColor Yellow
$dbExists = docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS -e "SHOW DATABASES LIKE '$DB_NAME';" 2>&1 | Select-String -Pattern $DB_NAME
if (-not $dbExists) {
    Write-Host "Database '$DB_NAME' does not exist. Creating..." -ForegroundColor Yellow
    docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS -e "CREATE DATABASE \`$DB_NAME\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>&1 | Out-Null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Database created successfully!" -ForegroundColor Green
    } else {
        Write-Host "Error: Failed to create database!" -ForegroundColor Red
        exit 1
    }
} else {
    Write-Host "Database '$DB_NAME' already exists." -ForegroundColor Green
}
Write-Host ""

# Import SQL file
Write-Host "Importing SQL file..." -ForegroundColor Yellow
Write-Host "This may take a while depending on the file size..." -ForegroundColor Yellow

# Get absolute path of SQL file
$absolutePath = (Resolve-Path $SQL_FILE).Path
$containerPath = "/tmp/db.sql"

# Copy SQL file to container
Write-Host "Copying SQL file to container..." -ForegroundColor Yellow
docker cp $absolutePath "${DOCKER_CONTAINER}:${containerPath}" 2>&1 | Out-Null
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to copy SQL file to container!" -ForegroundColor Red
    exit 1
}

# Import SQL file using cat inside container (more reliable)
Write-Host "Executing SQL commands..." -ForegroundColor Yellow
$result = docker exec $DOCKER_CONTAINER bash -c "cat $containerPath | mysql -u $DB_USER -p$DB_PASS $DB_NAME" 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host "SQL file imported successfully!" -ForegroundColor Green
} else {
    # Try alternative: direct import
    Write-Host "Trying alternative method (direct import)..." -ForegroundColor Yellow
    Get-Content $absolutePath | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "SQL file imported successfully (alternative method)!" -ForegroundColor Green
    } else {
        Write-Host "Error: Failed to import SQL file!" -ForegroundColor Red
        Write-Host "Output: $result" -ForegroundColor Red
        Write-Host ""
        Write-Host "Please try manually:" -ForegroundColor Yellow
        Write-Host "  Get-Content $SQL_FILE | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME" -ForegroundColor Cyan
        exit 1
    }
}

# Clean up: Remove SQL file from container
docker exec $DOCKER_CONTAINER rm -f $containerPath 2>&1 | Out-Null

Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Migration completed successfully!" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Seed reward data (optional):" -ForegroundColor Yellow
Write-Host "   docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME < migrations\seed_rewards.sql" -ForegroundColor Cyan
Write-Host ""
Write-Host "2. Verify tables:" -ForegroundColor Yellow
Write-Host "   docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e 'SHOW TABLES;'" -ForegroundColor Cyan
Write-Host ""

