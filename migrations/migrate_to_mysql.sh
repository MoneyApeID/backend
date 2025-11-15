#!/bin/bash
# Bash script untuk migrasi database ke MySQL via Docker
# Usage: ./migrate_to_mysql.sh

DOCKER_CONTAINER="vla-mysql"
DB_USER="root"
DB_PASS="vlaroot"
DB_NAME="vla-db"
SQL_FILE="database/db.sql"

echo "========================================="
echo "Migrasi Database ke MySQL via Docker"
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

# Check if SQL file exists
if [ ! -f "$SQL_FILE" ]; then
    echo "Error: SQL file not found: $SQL_FILE"
    exit 1
fi
echo "SQL file found: $SQL_FILE"
echo ""

# Check if database exists, if not create it
echo "Checking database existence..."
if ! docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS -e "SHOW DATABASES LIKE '$DB_NAME';" 2>/dev/null | grep -q "$DB_NAME"; then
    echo "Database '$DB_NAME' does not exist. Creating..."
    docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS -e "CREATE DATABASE \`$DB_NAME\` DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "Database created successfully!"
    else
        echo "Error: Failed to create database!"
        exit 1
    fi
else
    echo "Database '$DB_NAME' already exists."
fi
echo ""

# Import SQL file
echo "Importing SQL file..."
echo "This may take a while depending on the file size..."

# Get absolute path of SQL file
ABSOLUTE_PATH=$(cd "$(dirname "$SQL_FILE")" && pwd)/$(basename "$SQL_FILE")
CONTAINER_PATH="/tmp/db.sql"

# Copy SQL file to container
echo "Copying SQL file to container..."
docker cp "$ABSOLUTE_PATH" "${DOCKER_CONTAINER}:${CONTAINER_PATH}" 2>/dev/null
if [ $? -ne 0 ]; then
    echo "Error: Failed to copy SQL file to container!"
    exit 1
fi

# Import SQL file using cat inside container (more reliable)
echo "Executing SQL commands..."
docker exec $DOCKER_CONTAINER bash -c "cat $CONTAINER_PATH | mysql -u $DB_USER -p$DB_PASS $DB_NAME" 2>/dev/null

if [ $? -eq 0 ]; then
    echo "SQL file imported successfully!"
else
    # Try alternative: direct import
    echo "Trying alternative method (direct import)..."
    cat "$ABSOLUTE_PATH" | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME 2>/dev/null
    
    if [ $? -eq 0 ]; then
        echo "SQL file imported successfully (alternative method)!"
    else
        echo "Error: Failed to import SQL file!"
        echo ""
        echo "Please try manually:"
        echo "  cat $SQL_FILE | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME"
        exit 1
    fi
fi

# Clean up: Remove SQL file from container
docker exec $DOCKER_CONTAINER rm -f $CONTAINER_PATH 2>/dev/null

echo ""
echo "========================================="
echo "Migration completed successfully!"
echo "========================================="
echo ""
echo "Next steps:"
echo "1. Seed reward data (optional):"
echo "   docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME < migrations/seed_rewards.sql"
echo ""
echo "2. Verify tables:"
echo "   docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e 'SHOW TABLES;'"
echo ""

