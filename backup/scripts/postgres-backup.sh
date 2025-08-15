#!/bin/bash

# PostgreSQL Backup Script for Landscaping App
# This script creates full and incremental backups of PostgreSQL database

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/../configs/backup.conf"
LOG_FILE="/var/log/postgres-backup.log"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Default values if not set in config
BACKUP_TYPE="${BACKUP_TYPE:-full}"
BACKUP_DIR="${BACKUP_DIR:-/backups/postgres}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
RETENTION_WEEKS="${RETENTION_WEEKS:-4}"
RETENTION_MONTHS="${RETENTION_MONTHS:-6}"
AWS_S3_BUCKET="${AWS_S3_BUCKET:-}"
ENCRYPTION_KEY="${ENCRYPTION_KEY:-}"
SLACK_WEBHOOK_URL="${SLACK_WEBHOOK_URL:-}"

# Database connection details
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-landscaping_prod}"
DB_USER="${DB_USER:-postgres}"
PGPASSWORD="${PGPASSWORD:-}"

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"/{daily,weekly,monthly}

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Error handling
error_exit() {
    log "ERROR: $1"
    send_notification "❌ PostgreSQL Backup Failed" "$1" "danger"
    exit 1
}

# Success notification
success_notification() {
    local message="$1"
    log "SUCCESS: $message"
    send_notification "✅ PostgreSQL Backup Success" "$message" "good"
}

# Send Slack notification
send_notification() {
    local title="$1"
    local message="$2"
    local color="${3:-good}"
    
    if [[ -n "$SLACK_WEBHOOK_URL" ]]; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"attachments\":[{\"color\":\"$color\",\"title\":\"$title\",\"text\":\"$message\",\"footer\":\"$(hostname)\",\"ts\":$(date +%s)}]}" \
            "$SLACK_WEBHOOK_URL" &>/dev/null || log "Failed to send Slack notification"
    fi
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if pg_dump is available
    if ! command -v pg_dump &> /dev/null; then
        error_exit "pg_dump is not installed or not in PATH"
    fi
    
    # Check if AWS CLI is available (if S3 backup is enabled)
    if [[ -n "$AWS_S3_BUCKET" ]] && ! command -v aws &> /dev/null; then
        error_exit "AWS CLI is not installed or not in PATH"
    fi
    
    # Check database connectivity
    if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" &>/dev/null; then
        error_exit "Cannot connect to PostgreSQL database"
    fi
    
    log "Prerequisites check passed"
}

# Create backup filename
get_backup_filename() {
    local backup_type="$1"
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local filename="${DB_NAME}_${backup_type}_${timestamp}"
    
    case "$backup_type" in
        "full")
            echo "${BACKUP_DIR}/daily/${filename}.sql.gz"
            ;;
        "weekly")
            echo "${BACKUP_DIR}/weekly/${filename}.sql.gz"
            ;;
        "monthly")
            echo "${BACKUP_DIR}/monthly/${filename}.sql.gz"
            ;;
        *)
            echo "${BACKUP_DIR}/daily/${filename}.sql.gz"
            ;;
    esac
}

# Encrypt backup file
encrypt_backup() {
    local input_file="$1"
    local output_file="${input_file}.enc"
    
    if [[ -n "$ENCRYPTION_KEY" ]]; then
        log "Encrypting backup file..."
        openssl enc -aes-256-cbc -salt -in "$input_file" -out "$output_file" -k "$ENCRYPTION_KEY"
        rm "$input_file"
        echo "$output_file"
    else
        echo "$input_file"
    fi
}

# Create database backup
create_backup() {
    local backup_type="$1"
    local backup_file=$(get_backup_filename "$backup_type")
    local temp_file="${backup_file}.tmp"
    
    log "Starting $backup_type backup to: $backup_file"
    
    # Get database size for monitoring
    local db_size=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT pg_size_pretty(pg_database_size('$DB_NAME'));" | xargs)
    log "Database size: $db_size"
    
    # Create backup with compression
    local start_time=$(date +%s)
    
    pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        --verbose \
        --format=custom \
        --no-password \
        --compress=9 \
        --lock-wait-timeout=300000 \
        --exclude-table-data='audit_logs' \
        --exclude-table-data='session_data' \
        | gzip > "$temp_file"
    
    if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
        rm -f "$temp_file"
        error_exit "pg_dump failed"
    fi
    
    # Move temp file to final location
    mv "$temp_file" "$backup_file"
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    local backup_size=$(du -h "$backup_file" | cut -f1)
    
    log "Backup completed in ${duration}s, size: $backup_size"
    
    # Encrypt backup if encryption key is provided
    backup_file=$(encrypt_backup "$backup_file")
    
    # Verify backup integrity
    if [[ "$backup_file" =~ \.enc$ ]]; then
        log "Skipping integrity check for encrypted backup"
    else
        log "Verifying backup integrity..."
        if ! gzip -t "$backup_file"; then
            error_exit "Backup integrity check failed"
        fi
        log "Backup integrity verified"
    fi
    
    echo "$backup_file"
}

# Upload to S3
upload_to_s3() {
    local backup_file="$1"
    
    if [[ -z "$AWS_S3_BUCKET" ]]; then
        return 0
    fi
    
    log "Uploading backup to S3..."
    local s3_key="postgres-backups/$(date '+%Y/%m')/$(basename "$backup_file")"
    
    if aws s3 cp "$backup_file" "s3://$AWS_S3_BUCKET/$s3_key" --storage-class STANDARD_IA; then
        log "Successfully uploaded to s3://$AWS_S3_BUCKET/$s3_key"
    else
        error_exit "Failed to upload backup to S3"
    fi
}

# Clean up old backups
cleanup_old_backups() {
    log "Cleaning up old backups..."
    
    # Clean daily backups older than retention period
    find "${BACKUP_DIR}/daily" -name "*.sql.gz*" -type f -mtime +$RETENTION_DAYS -delete
    find "${BACKUP_DIR}/daily" -name "*.sql.gz*" -type f -mtime +$RETENTION_DAYS -print | wc -l | xargs -I {} log "Deleted {} old daily backups"
    
    # Clean weekly backups
    find "${BACKUP_DIR}/weekly" -name "*.sql.gz*" -type f -mtime +$((RETENTION_WEEKS * 7)) -delete
    
    # Clean monthly backups
    find "${BACKUP_DIR}/monthly" -name "*.sql.gz*" -type f -mtime +$((RETENTION_MONTHS * 30)) -delete
    
    # Clean S3 backups if configured
    if [[ -n "$AWS_S3_BUCKET" ]]; then
        log "Cleaning old S3 backups..."
        # Use S3 lifecycle policies for automated cleanup
        aws s3api put-bucket-lifecycle-configuration \
            --bucket "$AWS_S3_BUCKET" \
            --lifecycle-configuration file://"${SCRIPT_DIR}/../configs/s3-lifecycle.json" || log "Failed to update S3 lifecycle policy"
    fi
    
    log "Cleanup completed"
}

# Backup database schema separately
backup_schema() {
    local schema_file="${BACKUP_DIR}/schema/schema_$(date '+%Y%m%d_%H%M%S').sql"
    
    mkdir -p "$(dirname "$schema_file")"
    
    log "Backing up database schema..."
    
    pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        --schema-only \
        --no-password > "$schema_file"
    
    if [[ $? -eq 0 ]]; then
        gzip "$schema_file"
        log "Schema backup created: ${schema_file}.gz"
    else
        error_exit "Schema backup failed"
    fi
}

# Main backup function
main() {
    local backup_type="${1:-$BACKUP_TYPE}"
    
    log "Starting PostgreSQL backup process (type: $backup_type)"
    
    # Trap to ensure cleanup on exit
    trap 'log "Backup script interrupted"; exit 1' INT TERM
    
    check_prerequisites
    
    # Create the backup
    local backup_file=$(create_backup "$backup_type")
    
    # Upload to S3 if configured
    upload_to_s3 "$backup_file"
    
    # Backup schema separately
    backup_schema
    
    # Clean up old backups
    cleanup_old_backups
    
    # Generate backup report
    local backup_size=$(du -h "$backup_file" | cut -f1)
    local total_backups=$(find "$BACKUP_DIR" -name "*.sql.gz*" -type f | wc -l)
    local message="Backup completed successfully!\nFile: $(basename "$backup_file")\nSize: $backup_size\nTotal backups: $total_backups"
    
    success_notification "$message"
    
    log "PostgreSQL backup process completed successfully"
}

# Handle different backup types based on day of week/month
determine_backup_type() {
    local day_of_week=$(date +%u)  # 1-7 (Monday-Sunday)
    local day_of_month=$(date +%d)
    
    if [[ "$day_of_month" == "01" ]]; then
        echo "monthly"
    elif [[ "$day_of_week" == "7" ]]; then  # Sunday
        echo "weekly"
    else
        echo "full"
    fi
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    BACKUP_TYPE=$(determine_backup_type)
    main "$BACKUP_TYPE"
fi