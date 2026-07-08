#!/bin/bash
# TormentNexus SalesBot - Automated Database Backup Script
# Automatically dumps target Postgres database schemas and uploads to S3/Blob storage

# Exit on any error
set -e

# Load Environment Variables from optional path or default deployment
ENV_FILE="/opt/marketing_agent/.env"
if [ -f "$ENV_FILE" ]; then
    export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

# Ensure Database URL is present
if [ -z "$DATABASE_URL" ]; then
    echo "ERROR: DATABASE_URL is not set in environment."
    exit 1
fi

# Configure Backup Parameters
BACKUP_DIR="/var/backups/salesbot"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_FILE="${BACKUP_DIR}/salesbot_backup_${TIMESTAMP}.sql.gz"

# Ensure local backup directory exists
mkdir -p "$BACKUP_DIR"

echo "Starting Postgres database dump..."
pg_dump "$DATABASE_URL" | gzip > "$BACKUP_FILE"
echo "Backup successfully written locally: $BACKUP_FILE"

# Retention policy: Remove local backups older than 7 days
find "$BACKUP_DIR" -name "salesbot_backup_*.sql.gz" -mtime +7 -delete

# Optional S3 Upload trigger if bucket parameter exists
if [ -n "$S3_BUCKET_NAME" ]; then
    echo "Uploading backup to AWS S3 bucket: $S3_BUCKET_NAME..."
    if command -v aws &> /dev/null; then
        aws s3 cp "$BACKUP_FILE" "s3://${S3_BUCKET_NAME}/backups/$(basename "$BACKUP_FILE")"
        echo "Successfully uploaded backup to S3."
    else
        echo "WARNING: AWS CLI is not installed. Skipping S3 upload."
    fi
fi

echo "Database backup sequence complete."
