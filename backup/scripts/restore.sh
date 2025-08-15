#!/bin/bash

# Database Restore Script for Landscaping App
# This script restores PostgreSQL and Redis backups

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/../configs/backup.conf"
LOG_FILE="/var/log/restore.log"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
fi

# Default values
BACKUP_DIR="${BACKUP_DIR:-/backups}"
RESTORE_TYPE="${1:-}"
BACKUP_FILE="${2:-}"

# Usage information
usage() {
    echo "Usage: $0 <restore_type> [backup_file]"
    echo ""
    echo "Restore types:"
    echo "  postgres-latest    - Restore latest PostgreSQL backup"
    echo "  postgres-file      - Restore specific PostgreSQL backup file"
    echo "  redis-latest       - Restore latest Redis backup"
    echo "  redis-file         - Restore specific Redis backup file"
    echo "  list-postgres      - List available PostgreSQL backups"
    echo "  list-redis         - List available Redis backups"
    echo ""
    echo "Examples:"
    echo "  $0 postgres-latest"
    echo "  $0 postgres-file /backups/postgres/daily/landscaping_prod_full_20240115_120000.sql.gz"
    echo "  $0 redis-latest"
    echo "  $0 list-postgres"
    exit 1
}

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Error handling
error_exit() {
    log "ERROR: $1"
    exit 1
}

# Warning prompt
confirm_restore() {
    local restore_type="$1"
    local target="$2"
    
    echo ""
    echo "⚠️  WARNING: Database Restore Operation ⚠️"
    echo "=============================================="
    echo "Restore Type: $restore_type"
    echo "Target: $target"
    echo "Current Time: $(date)"
    echo ""
    echo "This operation will:"
    echo "- OVERWRITE existing database data"
    echo "- May cause service downtime"
    echo "- Should be performed during maintenance window"
    echo ""
    read -p "Are you sure you want to proceed? (type 'YES' to confirm): " confirmation
    
    if [[ "$confirmation" != "YES" ]]; then
        echo "Restore operation cancelled."
        exit 0
    fi
}

# Decrypt backup file
decrypt_backup() {
    local encrypted_file="$1"
    local decrypted_file="${encrypted_file%.enc}"
    
    if [[ "$encrypted_file" =~ \.enc$ ]]; then
        if [[ -z "$ENCRYPTION_KEY" ]]; then
            error_exit "Backup is encrypted but no encryption key provided"
        fi
        
        log "Decrypting backup file..."
        openssl enc -aes-256-cbc -d -in "$encrypted_file" -out "$decrypted_file" -k "$ENCRYPTION_KEY"
        echo "$decrypted_file"
    else
        echo "$encrypted_file"
    fi
}

# Download from S3 if needed
download_from_s3() {
    local s3_path="$1"
    local local_file="$2"
    
    if [[ -n "$AWS_S3_BUCKET" ]] && [[ "$s3_path" =~ ^s3:// ]]; then
        log "Downloading backup from S3..."
        aws s3 cp "$s3_path" "$local_file"
        echo "$local_file"
    else
        echo "$s3_path"
    fi
}

# List PostgreSQL backups
list_postgres_backups() {
    echo "Available PostgreSQL backups:"
    echo "=============================="
    
    for dir in daily weekly monthly; do
        if [[ -d "$BACKUP_DIR/postgres/$dir" ]]; then
            echo ""
            echo "$dir backups:"
            find "$BACKUP_DIR/postgres/$dir" -name "*.sql.gz*" -type f -printf "%T+ %s %p\n" | sort -r | head -10 | while IFS= read -r line; do
                echo "  $line"
            done
        fi
    done
    
    echo ""
    echo "S3 backups (if configured):"
    if [[ -n "$AWS_S3_BUCKET" ]]; then
        aws s3 ls "s3://$AWS_S3_BUCKET/postgres-backups/" --recursive --human-readable --summarize | tail -20
    else
        echo "  S3 not configured"
    fi
}

# List Redis backups
list_redis_backups() {
    echo "Available Redis backups:"
    echo "========================"
    
    if [[ -d "$BACKUP_DIR/redis" ]]; then
        find "$BACKUP_DIR/redis" -name "redis_*" -type f -printf "%T+ %s %p\n" | sort -r | head -20 | while IFS= read -r line; do
            echo "  $line"
        done
    fi
    
    echo ""
    echo "S3 backups (if configured):"
    if [[ -n "$AWS_S3_BUCKET" ]]; then
        aws s3 ls "s3://$AWS_S3_BUCKET/redis-backups/" --recursive --human-readable --summarize | tail -20
    else
        echo "  S3 not configured"
    fi
}

# Get latest PostgreSQL backup
get_latest_postgres_backup() {
    local latest_backup=""
    
    # Check daily backups first
    if [[ -d "$BACKUP_DIR/postgres/daily" ]]; then
        latest_backup=$(find "$BACKUP_DIR/postgres/daily" -name "*.sql.gz*" -type f -printf "%T@ %p\n" | sort -nr | head -1 | cut -d' ' -f2-)
    fi
    
    if [[ -z "$latest_backup" ]]; then
        error_exit "No PostgreSQL backups found in $BACKUP_DIR/postgres/daily"
    fi
    
    echo "$latest_backup"
}

# Get latest Redis backup
get_latest_redis_backup() {
    local latest_backup=""
    
    if [[ -d "$BACKUP_DIR/redis" ]]; then
        latest_backup=$(find "$BACKUP_DIR/redis" -name "redis_rdb_*.gz*" -type f -printf "%T@ %p\n" | sort -nr | head -1 | cut -d' ' -f2-)
    fi
    
    if [[ -z "$latest_backup" ]]; then
        error_exit "No Redis backups found in $BACKUP_DIR/redis"
    fi
    
    echo "$latest_backup"
}

# Restore PostgreSQL backup
restore_postgres() {
    local backup_file="$1"
    
    log "Starting PostgreSQL restore from: $backup_file"
    
    # Validate backup file exists
    if [[ ! -f "$backup_file" ]]; then
        error_exit "Backup file not found: $backup_file"
    fi
    
    # Decrypt if needed
    backup_file=$(decrypt_backup "$backup_file")
    
    # Confirm restore operation
    confirm_restore "PostgreSQL" "$DB_NAME"
    
    # Check database connectivity
    if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" &>/dev/null; then
        error_exit "Cannot connect to PostgreSQL database"
    fi
    
    # Create a pre-restore backup
    local pre_restore_backup="${BACKUP_DIR}/postgres/pre-restore-$(date +%Y%m%d_%H%M%S).sql.gz"
    log "Creating pre-restore backup: $pre_restore_backup"
    
    pg_dump -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
        --format=custom --compress=9 | gzip > "$pre_restore_backup"
    
    # Terminate active connections
    log "Terminating active connections..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c \
        "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = '$DB_NAME' AND pid <> pg_backend_pid();"
    
    # Drop and recreate database
    log "Dropping and recreating database..."
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "DROP DATABASE IF EXISTS $DB_NAME;"
    psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d postgres -c "CREATE DATABASE $DB_NAME;"
    
    # Restore from backup
    log "Restoring database from backup..."
    local start_time=$(date +%s)
    
    if [[ "$backup_file" =~ \.gz$ ]]; then
        gunzip -c "$backup_file" | pg_restore -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" --verbose --no-owner --no-privileges
    else
        pg_restore -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" --verbose --no-owner --no-privileges "$backup_file"
    fi
    
    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    # Verify restore
    log "Verifying restore..."
    local table_count=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';" | xargs)
    
    log "PostgreSQL restore completed in ${duration}s"
    log "Restored database contains $table_count tables"
    log "Pre-restore backup saved to: $pre_restore_backup"
    
    # Clean up decrypted file if it was created
    if [[ "$backup_file" != "$1" ]]; then
        rm -f "$backup_file"
    fi
}

# Restore Redis backup
restore_redis() {
    local backup_file="$1"
    
    log "Starting Redis restore from: $backup_file"
    
    # Validate backup file exists
    if [[ ! -f "$backup_file" ]]; then
        error_exit "Backup file not found: $backup_file"
    fi
    
    # Decrypt if needed
    backup_file=$(decrypt_backup "$backup_file")
    
    # Confirm restore operation
    confirm_restore "Redis" "$REDIS_HOST:$REDIS_PORT"
    
    # Check Redis connectivity
    local redis_cmd="redis-cli -h $REDIS_HOST -p $REDIS_PORT"
    if [[ -n "$REDIS_PASSWORD" ]]; then
        redis_cmd="$redis_cmd -a $REDIS_PASSWORD"
    fi
    
    if ! $redis_cmd ping &>/dev/null; then
        error_exit "Cannot connect to Redis server"
    fi
    
    # Create pre-restore backup
    log "Creating pre-restore backup..."
    local pre_restore_backup="${BACKUP_DIR}/redis/pre-restore-$(date +%Y%m%d_%H%M%S).rdb.gz"
    $redis_cmd --rdb "$pre_restore_backup"
    
    # Get current key count
    local key_count_before=$($redis_cmd dbsize)
    log "Current key count: $key_count_before"
    
    # Determine backup type and restore accordingly
    if [[ "$backup_file" =~ rdb ]]; then
        restore_redis_rdb "$backup_file"
    elif [[ "$backup_file" =~ json ]]; then
        restore_redis_json "$backup_file"
    else
        error_exit "Unknown Redis backup format: $backup_file"
    fi
    
    # Verify restore
    local key_count_after=$($redis_cmd dbsize)
    log "Redis restore completed"
    log "Key count after restore: $key_count_after"
    log "Pre-restore backup saved to: $pre_restore_backup"
    
    # Clean up decrypted file if it was created
    if [[ "$backup_file" != "$1" ]]; then
        rm -f "$backup_file"
    fi
}

# Restore Redis RDB backup
restore_redis_rdb() {
    local backup_file="$1"
    
    log "Restoring Redis RDB backup..."
    
    # Stop Redis service
    log "Stopping Redis service..."
    docker stop landscaping_redis_prod || systemctl stop redis-server || service redis-server stop
    
    # Get Redis data directory
    local redis_cmd="redis-cli -h $REDIS_HOST -p $REDIS_PORT"
    if [[ -n "$REDIS_PASSWORD" ]]; then
        redis_cmd="$redis_cmd -a $REDIS_PASSWORD"
    fi
    
    # Start Redis temporarily to get config
    docker start landscaping_redis_prod || systemctl start redis-server || service redis-server start
    sleep 5
    
    local rdb_dir=$($redis_cmd config get dir | tail -1)
    local rdb_filename=$($redis_cmd config get dbfilename | tail -1)
    local rdb_path="${rdb_dir}/${rdb_filename}"
    
    # Stop Redis again
    docker stop landscaping_redis_prod || systemctl stop redis-server || service redis-server stop
    
    # Replace RDB file
    log "Replacing RDB file: $rdb_path"
    
    if [[ "$backup_file" =~ \.gz$ ]]; then
        gunzip -c "$backup_file" > "$rdb_path"
    else
        cp "$backup_file" "$rdb_path"
    fi
    
    # Start Redis service
    log "Starting Redis service..."
    docker start landscaping_redis_prod || systemctl start redis-server || service redis-server start
    
    # Wait for Redis to be ready
    local max_wait=60
    local elapsed=0
    
    while [[ $elapsed -lt $max_wait ]]; do
        if $redis_cmd ping &>/dev/null; then
            log "Redis is ready"
            break
        fi
        sleep 1
        elapsed=$((elapsed + 1))
    done
    
    if [[ $elapsed -ge $max_wait ]]; then
        error_exit "Redis failed to start after restore"
    fi
}

# Restore Redis JSON backup
restore_redis_json() {
    local backup_file="$1"
    
    log "Restoring Redis JSON backup..."
    
    local redis_cmd="redis-cli -h $REDIS_HOST -p $REDIS_PORT"
    if [[ -n "$REDIS_PASSWORD" ]]; then
        redis_cmd="$redis_cmd --no-auth-warning -a $REDIS_PASSWORD"
    fi
    
    # Clear existing data
    log "Clearing existing Redis data..."
    $redis_cmd flushall
    
    # Parse and restore JSON backup
    log "Restoring data from JSON backup..."
    
    if [[ "$backup_file" =~ \.gz$ ]]; then
        gunzip -c "$backup_file" | jq -r 'to_entries[] | "\(.key)|\(.value.type)|\(.value.value)"' | while IFS='|' read -r key type value; do
            case "$type" in
                "string")
                    $redis_cmd set "$key" "$value"
                    ;;
                "list")
                    echo "$value" | jq -r '.[]' | while IFS= read -r item; do
                        $redis_cmd lpush "$key" "$item"
                    done
                    ;;
                "set")
                    echo "$value" | jq -r '.[]' | while IFS= read -r item; do
                        $redis_cmd sadd "$key" "$item"
                    done
                    ;;
                "hash")
                    echo "$value" | jq -r 'to_entries[] | "\(.key) \(.value)"' | while read -r field val; do
                        $redis_cmd hset "$key" "$field" "$val"
                    done
                    ;;
            esac
        done
    else
        error_exit "JSON backup must be compressed (.gz)"
    fi
}

# Main function
main() {
    if [[ $# -eq 0 ]]; then
        usage
    fi
    
    case "$RESTORE_TYPE" in
        "postgres-latest")
            local backup_file=$(get_latest_postgres_backup)
            restore_postgres "$backup_file"
            ;;
        "postgres-file")
            if [[ -z "$BACKUP_FILE" ]]; then
                error_exit "Backup file path required for postgres-file restore"
            fi
            restore_postgres "$BACKUP_FILE"
            ;;
        "redis-latest")
            local backup_file=$(get_latest_redis_backup)
            restore_redis "$backup_file"
            ;;
        "redis-file")
            if [[ -z "$BACKUP_FILE" ]]; then
                error_exit "Backup file path required for redis-file restore"
            fi
            restore_redis "$BACKUP_FILE"
            ;;
        "list-postgres")
            list_postgres_backups
            ;;
        "list-redis")
            list_redis_backups
            ;;
        *)
            echo "Unknown restore type: $RESTORE_TYPE"
            usage
            ;;
    esac
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi