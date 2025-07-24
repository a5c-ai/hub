#!/bin/bash

# PostgreSQL Database Restore Script for Hub
# This script restores the Hub PostgreSQL database from backup

set -e

# Check if backup file is provided
if [ $# -eq 0 ]; then
    echo "Usage: $0 <backup_file.sql.gz>"
    echo "Example: $0 ./backups/hub_backup_20240124_143000.sql.gz"
    exit 1
fi

BACKUP_FILE="$1"

# Check if backup file exists
if [ ! -f "$BACKUP_FILE" ]; then
    echo "Error: Backup file '$BACKUP_FILE' not found!"
    exit 1
fi

# Database configuration from environment or defaults
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-hub}"
DB_USER="${DB_USER:-hub}"

echo "Starting database restore..."
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "Backup file: $BACKUP_FILE"

# Warning message
echo "WARNING: This will completely replace the current database!"
read -p "Are you sure you want to continue? (yes/no): " confirmation

if [ "$confirmation" != "yes" ]; then
    echo "Restore cancelled."
    exit 0
fi

# Create temporary file for decompressed backup
TEMP_FILE=$(mktemp)
trap "rm -f $TEMP_FILE" EXIT

# Decompress backup file
echo "Decompressing backup file..."
gunzip -c "$BACKUP_FILE" > "$TEMP_FILE"

# Drop existing database connections
echo "Terminating existing database connections..."
psql \
  --host="$DB_HOST" \
  --port="$DB_PORT" \
  --username="$DB_USER" \
  --dbname="postgres" \
  --command="SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();" \
  || echo "Warning: Could not terminate all connections"

# Restore the database
echo "Restoring database..."
psql \
  --host="$DB_HOST" \
  --port="$DB_PORT" \
  --username="$DB_USER" \
  --dbname="$DB_NAME" \
  --file="$TEMP_FILE" \
  --verbose

echo "Database restore completed successfully!"
echo "Note: You may need to restart the Hub application."