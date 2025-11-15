#!/bin/bash
# Bash script untuk menambahkan tabel binary system ke database yang sudah ada
# Usage: ./migrations/add_binary_tables.sh

DOCKER_CONTAINER="vla-mysql"
DB_USER="root"
DB_PASS="vlaroot"
DB_NAME="vla-db"
SQL_FILE="migrations/add_binary_tables.sql"

echo "========================================="
echo "Menambahkan Tabel Binary System"
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
    echo "Please create the database first or run full migration."
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

# Check existing tables
echo "Checking existing tables..."
if docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW TABLES LIKE 'binary_nodes';" 2>/dev/null | grep -q "binary_nodes"; then
    echo "Warning: Table 'binary_nodes' already exists. Skipping creation..."
else
    echo "Table 'binary_nodes' does not exist. Will be created."
fi

if docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW TABLES LIKE 'rewards';" 2>/dev/null | grep -q "rewards"; then
    echo "Warning: Table 'rewards' already exists. Skipping creation..."
else
    echo "Table 'rewards' does not exist. Will be created."
fi

if docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW TABLES LIKE 'reward_progress';" 2>/dev/null | grep -q "reward_progress"; then
    echo "Warning: Table 'reward_progress' already exists. Skipping creation..."
else
    echo "Table 'reward_progress' does not exist. Will be created."
fi
echo ""

# Import SQL file
echo "Executing SQL migration..."
echo "This will add new tables without affecting existing data..."
echo ""

# Get absolute path of SQL file
ABSOLUTE_PATH=$(cd "$(dirname "$SQL_FILE")" && pwd)/$(basename "$SQL_FILE")

# Execute SQL file
cat "$ABSOLUTE_PATH" | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME 2>/dev/null

if [ $? -eq 0 ]; then
    echo "Migration completed successfully!"
else
    echo "Error: Failed to execute migration!"
    echo ""
    echo "Please try manually:"
    echo "  cat $SQL_FILE | docker exec -i $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME"
    exit 1
fi

echo ""
echo "Verifying tables..."
docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SHOW TABLES LIKE 'binary%'; SHOW TABLES LIKE 'reward%';" 2>/dev/null

echo ""
echo "Checking reward data..."
REWARD_COUNT=$(docker exec $DOCKER_CONTAINER mysql -u $DB_USER -p$DB_PASS $DB_NAME -e "SELECT COUNT(*) FROM rewards;" 2>/dev/null | tail -n 1)
if [ -n "$REWARD_COUNT" ] && [ "$REWARD_COUNT" -gt 0 ]; then
    echo "Rewards found: $REWARD_COUNT"
else
    echo "No rewards found. Data will be seeded."
fi

echo ""
echo "========================================="
echo "Migration completed successfully!"
echo "========================================="
echo ""
echo "New tables added:"
echo "  - binary_nodes"
echo "  - rewards"
echo "  - reward_progress"
echo ""

