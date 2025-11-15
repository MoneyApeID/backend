# Migration script untuk menambahkan tabel tutorials (PowerShell)
# Usage: .\run_tutorials_migration.ps1

# Container name MySQL
$CONTAINER_NAME = "vla-mysql"

# Database name (sesuaikan dengan .env Anda)
$DB_NAME = "vla-db"

# MySQL root user dan password (sesuaikan dengan .env Anda)
$DB_USER = "root"
$DB_PASS = "vlaroot"

# Path ke file migration
$MIGRATION_FILE = ".\migrations\create_tutorials_table.sql"

Write-Host "Running migration for tutorials table..." -ForegroundColor Yellow

# Import SQL file
Get-Content $MIGRATION_FILE | docker exec -i $CONTAINER_NAME mysql -u$DB_USER -p$DB_PASS $DB_NAME

if ($LASTEXITCODE -eq 0) {
    Write-Host "Migration berhasil! Tabel tutorials telah ditambahkan." -ForegroundColor Green
} else {
    Write-Host "Migration gagal! Silakan cek error di atas." -ForegroundColor Red
    exit 1
}

