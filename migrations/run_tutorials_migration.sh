#!/bin/bash
# Migration script untuk menambahkan tabel tutorials
# Usage: ./run_tutorials_migration.sh

# Container name MySQL
CONTAINER_NAME="vla-mysql"

# Database name (sesuaikan dengan .env Anda)
DB_NAME="vla-db"

# MySQL root user dan password (sesuaikan dengan .env Anda)
DB_USER="root"
DB_PASS="vlaroot"

# Path ke file migration
MIGRATION_FILE="./migrations/create_tutorials_table.sql"

echo "Running migration for tutorials table..."

# Import SQL file
docker exec -i ${CONTAINER_NAME} mysql -u${DB_USER} -p${DB_PASS} ${DB_NAME} < ${MIGRATION_FILE}

if [ $? -eq 0 ]; then
    echo "Migration berhasil! Tabel tutorials telah ditambahkan."
else
    echo "Migration gagal! Silakan cek error di atas."
    exit 1
fi

