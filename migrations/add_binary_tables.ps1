# PowerShell script untuk menambahkan tabel binary system ke database yang sudah ada
# Usage: .\migrations\add_binary_tables.ps1

$DOCKER_CONTAINER = "vla-mysql"
$DB_USER = "root"
$DB_PASS = "vlaroot"
$DB_NAME = "vla-db"
$SQL_FILE = "migrations\add_binary_tables.sql"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Menambahkan Tabel Binary System" -ForegroundColor Cyan
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
    Write-Host "Please create the database first or run full migration." -ForegroundColor Yellow
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

# Check existing tables
Write-Host "Checking existing tables..." -ForegroundColor Yellow
$existingTables = docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW TABLES LIKE 'binary_nodes';" 2>&1 | Select-String -Pattern "binary_nodes"
if ($existingTables) {
    Write-Host "Warning: Table 'binary_nodes' already exists. Skipping creation..." -ForegroundColor Yellow
} else {
    Write-Host "Table 'binary_nodes' does not exist. Will be created." -ForegroundColor Green
}

$existingRewards = docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW TABLES LIKE 'rewards';" 2>&1 | Select-String -Pattern "rewards"
if ($existingRewards) {
    Write-Host "Warning: Table 'rewards' already exists. Skipping creation..." -ForegroundColor Yellow
} else {
    Write-Host "Table 'rewards' does not exist. Will be created." -ForegroundColor Green
}

$existingProgress = docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW TABLES LIKE 'reward_progress';" 2>&1 | Select-String -Pattern "reward_progress"
if ($existingProgress) {
    Write-Host "Warning: Table 'reward_progress' already exists. Skipping creation..." -ForegroundColor Yellow
} else {
    Write-Host "Table 'reward_progress' does not exist. Will be created." -ForegroundColor Green
}
Write-Host ""

# Import SQL file
Write-Host "Executing SQL migration..." -ForegroundColor Yellow
Write-Host "This will add new tables without affecting existing data..." -ForegroundColor Yellow
Write-Host ""

# Get absolute path of SQL file
$absolutePath = (Resolve-Path $SQL_FILE).Path

# Execute SQL file
Get-Content $absolutePath | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME 2>&1 | Out-Null

if ($LASTEXITCODE -eq 0) {
    Write-Host "Migration completed successfully!" -ForegroundColor Green
} else {
    Write-Host "Error: Failed to execute migration!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please try manually:" -ForegroundColor Yellow
    Write-Host "  Get-Content $SQL_FILE | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME" -ForegroundColor Cyan
    exit 1
}

Write-Host ""
Write-Host "Verifying tables..." -ForegroundColor Yellow
$tables = docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW TABLES LIKE 'binary%' OR SHOW TABLES LIKE 'reward%';" 2>&1
Write-Host $tables

Write-Host ""
Write-Host "Checking reward data..." -ForegroundColor Yellow
$rewardCount = docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SELECT COUNT(*) as count FROM rewards;" 2>&1 | Select-String -Pattern "\d+"
if ($rewardCount) {
    Write-Host "Rewards found: $rewardCount" -ForegroundColor Green
} else {
    Write-Host "No rewards found. Data will be seeded." -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Migration completed successfully!" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "New tables added:" -ForegroundColor Yellow
Write-Host "  - binary_nodes" -ForegroundColor Cyan
Write-Host "  - rewards" -ForegroundColor Cyan
Write-Host "  - reward_progress" -ForegroundColor Cyan
Write-Host ""

