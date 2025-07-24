#!/bin/bash

# PostgreSQL Database Backup Script for Hub
# This script creates backups of the Hub PostgreSQL database

set -e

# Configuration
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="hub_backup_${TIMESTAMP}.sql"
RETENTION_DAYS="${RETENTION_DAYS:-7}"

# Database configuration from environment or defaults
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-hub}"
DB_USER="${DB_USER:-hub}"

# Create backup directory if it doesn't exist
mkdir -p "$BACKUP_DIR"

echo "Starting database backup..."
echo "Database: $DB_NAME"
echo "Host: $DB_HOST:$DB_PORT"
echo "Backup file: $BACKUP_DIR/$BACKUP_FILE"

# Create the backup
pg_dump \
  --host="$DB_HOST" \
  --port="$DB_PORT" \
  --username="$DB_USER" \
  --dbname="$DB_NAME" \
  --verbose \
  --clean \
  --no-owner \
  --no-privileges \
  --file="$BACKUP_DIR/$BACKUP_FILE"

# Compress the backup
gzip "$BACKUP_DIR/$BACKUP_FILE"
COMPRESSED_FILE="${BACKUP_FILE}.gz"

echo "Backup completed: $BACKUP_DIR/$COMPRESSED_FILE"

# Clean up old backups (keep only last N days)
echo "Cleaning up backups older than $RETENTION_DAYS days..."
find "$BACKUP_DIR" -name "hub_backup_*.sql.gz" -type f -mtime +$RETENTION_DAYS -delete

echo "Backup script completed successfully!"

# Optional: Upload to cloud storage (uncomment and configure as needed)
# aws s3 cp "$BACKUP_DIR/$COMPRESSED_FILE" s3://your-backup-bucket/database/
# az storage blob upload --file "$BACKUP_DIR/$COMPRESSED_FILE" --container-name backups --name "database/$COMPRESSED_FILE"