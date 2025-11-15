#!/bin/bash
# Bash script untuk menambahkan AUTO_INCREMENT pada semua tabel
# Usage: ./migrations/fix_auto_increment.sh

DOCKER_CONTAINER="vla-mysql"
DB_USER="root"
DB_PASS="vlaroot"
DB_NAME="vla-db"
SQL_FILE="migrations/fix_auto_increment.sql"

echo "========================================="
echo "Fix AUTO_INCREMENT untuk Semua Tabel"
echo "========================================="
echo ""

# Check if Docker container is running
echo "Checking Docker container..."
if ! docker ps --format "{{.Names}}" | grep -q "^${DOCKER_CONTAINER}$"; then
    echo "Error: Docker container '$DOCKER_CONTAINER' is not running!"
    echo "Please start the container first: docker start $DOCKER_CONTAINER"
    exit 1
fi
echo "Container is running"
echo ""

# Check if database exists
echo "Checking database existence..."
if ! docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS -e "SHOW DATABASES LIKE '$DB_NAME';" 2>/dev/null | grep -q "$DB_NAME"; then
    echo "Error: Database '$DB_NAME' does not exist!"
    exit 1
fi
echo "Database '$DB_NAME' found."
echo ""

# Check if SQL file exists
if [ ! -f "$SQL_FILE" ]; then
    echo "Error: SQL file not found: $SQL_FILE"
    exit 1
fi
echo "SQL file found: $SQL_FILE"
echo ""

# Execute SQL file
echo "Executing SQL commands to fix AUTO_INCREMENT..."
echo "This will add AUTO_INCREMENT to all tables that don't have it..."
echo ""

# Get absolute path of SQL file
ABSOLUTE_PATH=$(cd "$(dirname "$SQL_FILE")" && pwd)/$(basename "$SQL_FILE")

# Execute SQL file
cat "$ABSOLUTE_PATH" | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME 2>/dev/null

if [ $? -eq 0 ]; then
    echo "AUTO_INCREMENT fixed successfully!"
else
    echo "Error: Failed to execute SQL commands!"
    echo ""
    echo "Please try manually:"
    echo "  cat $SQL_FILE | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME"
    exit 1
fi

echo ""
echo "Verifying AUTO_INCREMENT on key tables..."
TABLES=("users" "admins" "transactions" "investments" "deposits")
for table in "${TABLES[@]}"; do
    if docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW CREATE TABLE \`$table\`;" 2>/dev/null | grep -q "AUTO_INCREMENT"; then
        echo "  ✓ $table has AUTO_INCREMENT"
    else
        echo "  ✗ $table missing AUTO_INCREMENT"
    fi
done

echo ""
echo "========================================="
echo "Fix completed!"
echo "========================================="
echo ""

