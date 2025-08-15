#!/bin/bash

# Redis Backup Script for Landscaping App
# This script creates backups of Redis data

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/../configs/backup.conf"
LOG_FILE="/var/log/redis-backup.log"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
else
    echo "Configuration file not found: $CONFIG_FILE"
    exit 1
fi

# Default values if not set in config
BACKUP_DIR="${BACKUP_DIR:-/backups/redis}"
RETENTION_DAYS="${RETENTION_DAYS:-7}"
AWS_S3_BUCKET="${AWS_S3_BUCKET:-}"
ENCRYPTION_KEY="${ENCRYPTION_KEY:-}"
SLACK_WEBHOOK_URL="${SLACK_WEBHOOK_URL:-}"

# Redis connection details
REDIS_HOST="${REDIS_HOST:-localhost}"
REDIS_PORT="${REDIS_PORT:-6379}"
REDIS_PASSWORD="${REDIS_PASSWORD:-}"
REDIS_DB="${REDIS_DB:-0}"

# Ensure backup directory exists
mkdir -p "$BACKUP_DIR"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Error handling
error_exit() {
    log "ERROR: $1"
    send_notification "❌ Redis Backup Failed" "$1" "danger"
    exit 1
}

# Success notification
success_notification() {
    local message="$1"
    log "SUCCESS: $message"
    send_notification "✅ Redis Backup Success" "$message" "good"
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
    
    # Check if redis-cli is available
    if ! command -v redis-cli &> /dev/null; then
        error_exit "redis-cli is not installed or not in PATH"
    fi
    
    # Check Redis connectivity
    local redis_cmd="redis-cli -h $REDIS_HOST -p $REDIS_PORT"
    if [[ -n "$REDIS_PASSWORD" ]]; then
        redis_cmd="$redis_cmd -a $REDIS_PASSWORD"
    fi
    
    if ! $redis_cmd ping &>/dev/null; then
        error_exit "Cannot connect to Redis server"
    fi
    
    log "Prerequisites check passed"
}

# Get Redis info
get_redis_info() {
    local redis_cmd="redis-cli -h $REDIS_HOST -p $REDIS_PORT"
    if [[ -n "$REDIS_PASSWORD" ]]; then
        redis_cmd="$redis_cmd -a $REDIS_PASSWORD"
    fi
    
    $redis_cmd info
}

# Create RDB backup
create_rdb_backup() {
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local backup_file="${BACKUP_DIR}/redis_rdb_${timestamp}.rdb"
    local temp_file="${backup_file}.tmp"
    
    log "Starting RDB backup..."
    
    local redis_cmd="redis-cli -h $REDIS_HOST -p $REDIS_PORT"
    if [[ -n "$REDIS_PASSWORD" ]]; then
        redis_cmd="$redis_cmd -a $REDIS_PASSWORD"
    fi
    
    # Get database size info
    local db_info=$($redis_cmd info keyspace)
    log "Redis keyspace info: $db_info"
    
    # Trigger BGSAVE
    local start_time=$(date +%s)
    
    if ! $redis_cmd bgsave; then
        error_exit "Failed to trigger BGSAVE"
    fi
    
    # Wait for BGSAVE to complete
    log "Waiting for BGSAVE to complete..."
    local max_wait=1800  # 30 minutes
    local elapsed=0
    
    while [[ $elapsed -lt $max_wait ]]; do
        local last_save=$($redis_cmd lastsave)
        sleep 2
        local new_last_save=$($redis_cmd lastsave)
        
        if [[ "$last_save" != "$new_last_save" ]]; then
            log "BGSAVE completed"
            break
        fi
        
        elapsed=$((elapsed + 2))
        
        if [[ $((elapsed % 60)) -eq 0 ]]; then
            log "Still waiting for BGSAVE... (${elapsed}s elapsed)"
        fi
    done
    
    if [[ $elapsed -ge $max_wait ]]; then
        error_exit "BGSAVE timeout after ${max_wait}s"
    fi
    
    # Get the RDB file path from Redis
    local rdb_filename=$($redis_cmd config get dbfilename | tail -1)
    local rdb_dir=$($redis_cmd config get dir | tail -1)
    local source_rdb="${rdb_dir}/${rdb_filename}"
    
    # Copy and compress the RDB file
    if [[ -f "$source_rdb" ]]; then
        cp "$source_rdb" "$temp_file"
        gzip "$temp_file"
        mv "${temp_file}.gz" "${backup_file}.gz"
        
        local end_time=$(date +%s)
        local duration=$((end_time - start_time))
        local backup_size=$(du -h "${backup_file}.gz" | cut -f1)
        
        log "RDB backup completed in ${duration}s, size: $backup_size"
        echo "${backup_file}.gz"
    else
        error_exit "Source RDB file not found: $source_rdb"
    fi
}

# Create AOF backup (if AOF is enabled)
create_aof_backup() {
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local backup_file="${BACKUP_DIR}/redis_aof_${timestamp}.aof.gz"
    
    local redis_cmd="redis-cli -h $REDIS_HOST -p $REDIS_PORT"
    if [[ -n "$REDIS_PASSWORD" ]]; then
        redis_cmd="$redis_cmd -a $REDIS_PASSWORD"
    fi
    
    # Check if AOF is enabled
    local aof_enabled=$($redis_cmd config get appendonly | tail -1)
    
    if [[ "$aof_enabled" != "yes" ]]; then
        log "AOF is not enabled, skipping AOF backup"
        return 0
    fi
    
    log "Starting AOF backup..."
    
    # Get AOF file path
    local aof_filename=$($redis_cmd config get appendfilename | tail -1)
    local aof_dir=$($redis_cmd config get dir | tail -1)
    local source_aof="${aof_dir}/${aof_filename}"
    
    # Trigger AOF rewrite for consistent backup
    $redis_cmd bgrewriteaof
    
    # Wait for AOF rewrite to complete
    log "Waiting for AOF rewrite to complete..."
    local max_wait=900  # 15 minutes
    local elapsed=0
    
    while [[ $elapsed -lt $max_wait ]]; do
        local aof_rewrite_in_progress=$($redis_cmd info persistence | grep aof_rewrite_in_progress | cut -d: -f2 | tr -d '\r')
        
        if [[ "$aof_rewrite_in_progress" == "0" ]]; then
            log "AOF rewrite completed"
            break
        fi
        
        sleep 5
        elapsed=$((elapsed + 5))
        
        if [[ $((elapsed % 60)) -eq 0 ]]; then
            log "Still waiting for AOF rewrite... (${elapsed}s elapsed)"
        fi
    done
    
    if [[ $elapsed -ge $max_wait ]]; then
        error_exit "AOF rewrite timeout after ${max_wait}s"
    fi
    
    # Copy and compress the AOF file
    if [[ -f "$source_aof" ]]; then
        gzip -c "$source_aof" > "$backup_file"
        local backup_size=$(du -h "$backup_file" | cut -f1)
        log "AOF backup completed, size: $backup_size"
        echo "$backup_file"
    else
        log "Source AOF file not found: $source_aof"
        return 1
    fi
}

# Create JSON dump backup
create_json_backup() {
    local timestamp=$(date '+%Y%m%d_%H%M%S')
    local backup_file="${BACKUP_DIR}/redis_dump_${timestamp}.json.gz"
    local temp_file="${backup_file%%.gz}"
    
    log "Starting JSON dump backup..."
    
    local redis_cmd="redis-cli -h $REDIS_HOST -p $REDIS_PORT"
    if [[ -n "$REDIS_PASSWORD" ]]; then
        redis_cmd="$redis_cmd --no-auth-warning -a $REDIS_PASSWORD"
    fi
    
    # Create JSON dump
    {
        echo "{"
        local first=true
        
        # Get all keys
        $redis_cmd --scan | while IFS= read -r key; do
            if [[ "$first" == "true" ]]; then
                first=false
            else
                echo ","
            fi
            
            # Get key type
            local key_type=$($redis_cmd type "$key")
            
            # Output key with proper JSON escaping
            printf '  "%s": {' "$(echo "$key" | sed 's/"/\\"/g')"
            printf '"type": "%s", "value": ' "$key_type"
            
            case "$key_type" in
                "string")
                    printf '"%s"' "$($redis_cmd get "$key" | sed 's/"/\\"/g')"
                    ;;
                "list")
                    printf '['
                    $redis_cmd lrange "$key" 0 -1 | while IFS= read -r value; do
                        printf '"%s"' "$(echo "$value" | sed 's/"/\\"/g')"
                        # Add comma for all but last element (simplified)
                    done
                    printf ']'
                    ;;
                "set")
                    printf '['
                    $redis_cmd smembers "$key" | while IFS= read -r value; do
                        printf '"%s"' "$(echo "$value" | sed 's/"/\\"/g')"
                    done
                    printf ']'
                    ;;
                "hash")
                    printf '{'
                    $redis_cmd hgetall "$key" | while IFS= read -r field && IFS= read -r value; do
                        printf '"%s": "%s"' "$(echo "$field" | sed 's/"/\\"/g')" "$(echo "$value" | sed 's/"/\\"/g')"
                    done
                    printf '}'
                    ;;
                *)
                    printf 'null'
                    ;;
            esac
            
            printf '}'
        done
        
        echo ""
        echo "}"
    } > "$temp_file"
    
    # Compress the JSON file
    gzip "$temp_file"
    
    local backup_size=$(du -h "$backup_file" | cut -f1)
    log "JSON dump backup completed, size: $backup_size"
    echo "$backup_file"
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

# Upload to S3
upload_to_s3() {
    local backup_file="$1"
    
    if [[ -z "$AWS_S3_BUCKET" ]]; then
        return 0
    fi
    
    log "Uploading backup to S3..."
    local s3_key="redis-backups/$(date '+%Y/%m')/$(basename "$backup_file")"
    
    if aws s3 cp "$backup_file" "s3://$AWS_S3_BUCKET/$s3_key" --storage-class STANDARD_IA; then
        log "Successfully uploaded to s3://$AWS_S3_BUCKET/$s3_key"
    else
        error_exit "Failed to upload backup to S3"
    fi
}

# Clean up old backups
cleanup_old_backups() {
    log "Cleaning up old backups..."
    
    # Clean local backups older than retention period
    find "$BACKUP_DIR" -name "redis_*" -type f -mtime +$RETENTION_DAYS -delete
    local deleted_count=$(find "$BACKUP_DIR" -name "redis_*" -type f -mtime +$RETENTION_DAYS -print | wc -l)
    log "Deleted $deleted_count old backups"
    
    log "Cleanup completed"
}

# Main backup function
main() {
    log "Starting Redis backup process"
    
    # Trap to ensure cleanup on exit
    trap 'log "Backup script interrupted"; exit 1' INT TERM
    
    check_prerequisites
    
    # Get Redis info for logging
    log "Redis server info:"
    get_redis_info | head -10
    
    local backup_files=()
    local total_size=0
    
    # Create RDB backup
    local rdb_backup=$(create_rdb_backup)
    backup_files+=("$rdb_backup")
    rdb_backup=$(encrypt_backup "$rdb_backup")
    upload_to_s3 "$rdb_backup"
    
    # Create AOF backup if enabled
    if aof_backup=$(create_aof_backup); then
        backup_files+=("$aof_backup")
        aof_backup=$(encrypt_backup "$aof_backup")
        upload_to_s3 "$aof_backup"
    fi
    
    # Create JSON dump backup
    local json_backup=$(create_json_backup)
    backup_files+=("$json_backup")
    json_backup=$(encrypt_backup "$json_backup")
    upload_to_s3 "$json_backup"
    
    # Clean up old backups
    cleanup_old_backups
    
    # Generate backup report
    local total_backups=$(find "$BACKUP_DIR" -name "redis_*" -type f | wc -l)
    local message="Redis backup completed successfully!\nBackup files: ${#backup_files[@]}\nTotal backups: $total_backups"
    
    success_notification "$message"
    
    log "Redis backup process completed successfully"
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi