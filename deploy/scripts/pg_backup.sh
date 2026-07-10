#!/bin/bash
set -e

# Configuration
DB_NAME=${DB_NAME:-"marketing_agent"}
DB_USER=${DB_USER:-"postgres"}
DB_HOST=${DB_HOST:-"localhost"}
S3_BUCKET=${S3_BUCKET:-"tormentnexus-backups"}
BACKUP_DIR=${BACKUP_DIR:-"/tmp/db_backups"}
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
FILENAME="db_backup_${DB_NAME}_${TIMESTAMP}.sql.gz"

mkdir -p "$BACKUP_DIR"

echo "Starting database backup for ${DB_NAME} at ${TIMESTAMP}..."

# Dump and compress
PGPASSWORD="${DB_PASSWORD}" pg_dump -h "$DB_HOST" -U "$DB_USER" -d "$DB_NAME" | gzip > "${BACKUP_DIR}/${FILENAME}"

echo "Backup created: ${BACKUP_DIR}/${FILENAME}"

# Upload to S3 if configured
if [ -n "$AWS_ACCESS_KEY_ID" ] && [ -n "$AWS_SECRET_ACCESS_KEY" ]; then
    echo "Uploading to s3://${S3_BUCKET}/${FILENAME}..."
    aws s3 cp "${BACKUP_DIR}/${FILENAME}" "s3://${S3_BUCKET}/${FILENAME}"
    echo "Upload complete."
else
    echo "AWS credentials not configured. Skipping S3 upload."
fi

# Cleanup old local backups (keep last 7 days)
find "$BACKUP_DIR" -name "db_backup_*.sql.gz" -type f -mtime +7 -exec rm {} \;

echo "Backup process finished."
