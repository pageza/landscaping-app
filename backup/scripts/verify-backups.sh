#!/bin/bash

# Backup Verification Script
# Verifies the integrity and validity of database backups

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CONFIG_FILE="${SCRIPT_DIR}/../configs/backup.conf"
LOG_FILE="/var/log/backup-verify.log"

# Load configuration
if [[ -f "$CONFIG_FILE" ]]; then
    source "$CONFIG_FILE"
fi

# Default values
BACKUP_DIR="${BACKUP_DIR:-/backups}"
VERIFY_LATEST_ONLY="${VERIFY_LATEST_ONLY:-true}"
MAX_BACKUPS_TO_VERIFY="${MAX_BACKUPS_TO_VERIFY:-5}"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Error handling
error_exit() {
    log "ERROR: $1"
    exit 1
}

# Send notification
send_notification() {
    local title="$1"
    local message="$2"
    local color="${3:-good}"
    
    if [[ -n "${SLACK_WEBHOOK_URL:-}" ]]; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"attachments\":[{\"color\":\"$color\",\"title\":\"$title\",\"text\":\"$message\",\"footer\":\"$(hostname)\",\"ts\":$(date +%s)}]}" \
            "$SLACK_WEBHOOK_URL" &>/dev/null || log "Failed to send Slack notification"
    fi
}

# Verify PostgreSQL backup
verify_postgres_backup() {
    local backup_file="$1"
    local temp_db="verify_$(date +%s)"
    
    log "Verifying PostgreSQL backup: $(basename "$backup_file")"
    
    # Check file integrity first
    if [[ "$backup_file" =~ \.gz$ ]]; then
        if ! gzip -t "$backup_file"; then
            log "FAILED: Backup file is corrupted (gzip test failed)"
            return 1
        fi
    fi
    
    # Decrypt if needed
    local actual_backup="$backup_file"
    if [[ "$backup_file" =~ \.enc$ ]]; then
        if [[ -z "${ENCRYPTION_KEY:-}" ]]; then
            log "SKIPPED: Backup is encrypted but no encryption key available"
            return 0
        fi
        
        actual_backup="/tmp/verify_$(basename "$backup_file" .enc)"
        openssl enc -aes-256-cbc -d -in "$backup_file" -out "$actual_backup" -k "$ENCRYPTION_KEY"
        
        # Test decompression
        if [[ "$actual_backup" =~ \.gz$ ]]; then
            if ! gzip -t "$actual_backup"; then
                log "FAILED: Decrypted backup file is corrupted"
                rm -f "$actual_backup"
                return 1
            fi
        fi
    fi
    
    # Create temporary database for verification
    log "Creating temporary database: $temp_db"
    
    if ! psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d postgres -c "CREATE DATABASE $temp_db;" &>/dev/null; then
        log "FAILED: Cannot create temporary database"
        [[ "$actual_backup" != "$backup_file" ]] && rm -f "$actual_backup"
        return 1
    fi
    
    # Restore backup to temporary database
    local restore_success=true
    
    if [[ "$actual_backup" =~ \.gz$ ]]; then
        if ! gunzip -c "$actual_backup" | pg_restore -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "$temp_db" --no-owner --no-privileges &>/dev/null; then
            restore_success=false
        fi
    else
        if ! pg_restore -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "$temp_db" --no-owner --no-privileges "$actual_backup" &>/dev/null; then
            restore_success=false
        fi
    fi
    
    # Clean up temporary files
    [[ "$actual_backup" != "$backup_file" ]] && rm -f "$actual_backup"
    
    if [[ "$restore_success" == "false" ]]; then
        log "FAILED: Cannot restore backup to temporary database"
        psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d postgres -c "DROP DATABASE IF EXISTS $temp_db;" &>/dev/null
        return 1
    fi
    
    # Verify database contents
    local table_count=$(psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d "$temp_db" -t -c "SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs || echo "0")
    
    if [[ "$table_count" -eq 0 ]]; then
        log "FAILED: Restored database contains no tables"
        psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d postgres -c "DROP DATABASE IF EXISTS $temp_db;" &>/dev/null
        return 1
    fi
    
    # Clean up temporary database
    psql -h "${DB_HOST:-localhost}" -p "${DB_PORT:-5432}" -U "${DB_USER:-postgres}" -d postgres -c "DROP DATABASE $temp_db;" &>/dev/null
    
    log "SUCCESS: PostgreSQL backup verified successfully ($table_count tables)"
    return 0
}

# Verify Redis backup
verify_redis_backup() {
    local backup_file="$1"
    
    log "Verifying Redis backup: $(basename "$backup_file")"
    
    # Check file integrity first
    if [[ "$backup_file" =~ \.gz$ ]]; then
        if ! gzip -t "$backup_file"; then
            log "FAILED: Backup file is corrupted (gzip test failed)"
            return 1
        fi
    fi
    
    # Decrypt if needed
    local actual_backup="$backup_file"
    if [[ "$backup_file" =~ \.enc$ ]]; then
        if [[ -z "${ENCRYPTION_KEY:-}" ]]; then
            log "SKIPPED: Backup is encrypted but no encryption key available"
            return 0
        fi
        
        actual_backup="/tmp/verify_$(basename "$backup_file" .enc)"
        openssl enc -aes-256-cbc -d -in "$backup_file" -out "$actual_backup" -k "$ENCRYPTION_KEY"
    fi
    
    # Verify different backup types
    local verification_success=true
    
    if [[ "$actual_backup" =~ rdb ]]; then
        # RDB backup verification
        if [[ "$actual_backup" =~ \.gz$ ]]; then
            local temp_rdb="/tmp/verify_$(date +%s).rdb"
            if gunzip -c "$actual_backup" > "$temp_rdb"; then
                # Try to read RDB file header
                if ! head -c 9 "$temp_rdb" | grep -q "REDIS"; then
                    log "FAILED: RDB backup file does not have valid header"
                    verification_success=false
                fi
                rm -f "$temp_rdb"
            else
                verification_success=false
            fi
        fi
    elif [[ "$actual_backup" =~ json ]]; then
        # JSON backup verification
        if [[ "$actual_backup" =~ \.gz$ ]]; then
            if ! gunzip -c "$actual_backup" | jq empty &>/dev/null; then
                log "FAILED: JSON backup is not valid JSON"
                verification_success=false
            fi
        else
            if ! jq empty "$actual_backup" &>/dev/null; then
                log "FAILED: JSON backup is not valid JSON"
                verification_success=false
            fi
        fi
    fi
    
    # Clean up temporary files
    [[ "$actual_backup" != "$backup_file" ]] && rm -f "$actual_backup"
    
    if [[ "$verification_success" == "true" ]]; then
        log "SUCCESS: Redis backup verified successfully"
        return 0
    else
        log "FAILED: Redis backup verification failed"
        return 1
    fi
}

# Get backup files to verify
get_postgres_backups_to_verify() {
    local backups=()
    
    if [[ "$VERIFY_LATEST_ONLY" == "true" ]]; then
        # Get latest backup from each category
        for dir in daily weekly monthly; do
            local latest=$(find "$BACKUP_DIR/postgres/$dir" -name "*.sql.gz*" -type f -printf "%T@ %p\n" 2>/dev/null | sort -nr | head -1 | cut -d' ' -f2-)
            [[ -n "$latest" ]] && backups+=("$latest")
        done
    else
        # Get multiple recent backups
        local all_backups=($(find "$BACKUP_DIR/postgres" -name "*.sql.gz*" -type f -printf "%T@ %p\n" 2>/dev/null | sort -nr | head -$MAX_BACKUPS_TO_VERIFY | cut -d' ' -f2-))
        backups=("${all_backups[@]}")
    fi
    
    printf '%s\n' "${backups[@]}"
}

get_redis_backups_to_verify() {
    local backups=()
    
    if [[ "$VERIFY_LATEST_ONLY" == "true" ]]; then
        # Get latest RDB and JSON backups
        local latest_rdb=$(find "$BACKUP_DIR/redis" -name "redis_rdb_*.gz*" -type f -printf "%T@ %p\n" 2>/dev/null | sort -nr | head -1 | cut -d' ' -f2-)
        local latest_json=$(find "$BACKUP_DIR/redis" -name "redis_dump_*.json.gz*" -type f -printf "%T@ %p\n" 2>/dev/null | sort -nr | head -1 | cut -d' ' -f2-)
        
        [[ -n "$latest_rdb" ]] && backups+=("$latest_rdb")
        [[ -n "$latest_json" ]] && backups+=("$latest_json")
    else
        # Get multiple recent backups
        local all_backups=($(find "$BACKUP_DIR/redis" -name "redis_*.gz*" -type f -printf "%T@ %p\n" 2>/dev/null | sort -nr | head -$MAX_BACKUPS_TO_VERIFY | cut -d' ' -f2-))
        backups=("${all_backups[@]}")
    fi
    
    printf '%s\n' "${backups[@]}"
}

# Main verification function
main() {
    log "Starting backup verification process"
    
    local total_verified=0
    local total_failed=0
    local failed_backups=()
    
    # Verify PostgreSQL backups
    log "Verifying PostgreSQL backups..."
    while IFS= read -r backup_file; do
        if [[ -n "$backup_file" ]]; then
            if verify_postgres_backup "$backup_file"; then
                total_verified=$((total_verified + 1))
            else
                total_failed=$((total_failed + 1))
                failed_backups+=("$(basename "$backup_file")")
            fi
        fi
    done < <(get_postgres_backups_to_verify)
    
    # Verify Redis backups
    log "Verifying Redis backups..."
    while IFS= read -r backup_file; do
        if [[ -n "$backup_file" ]]; then
            if verify_redis_backup "$backup_file"; then
                total_verified=$((total_verified + 1))
            else
                total_failed=$((total_failed + 1))
                failed_backups+=("$(basename "$backup_file")")
            fi
        fi
    done < <(get_redis_backups_to_verify)
    
    # Generate report
    log "Backup verification completed"
    log "Total backups verified: $total_verified"
    log "Total backups failed: $total_failed"
    
    if [[ $total_failed -gt 0 ]]; then
        local failed_list=$(IFS=', '; echo "${failed_backups[*]}")
        log "Failed backups: $failed_list"
        
        local message="Backup verification completed with failures!\nVerified: $total_verified\nFailed: $total_failed\nFailed backups: $failed_list"
        send_notification "⚠️ Backup Verification Issues" "$message" "warning"
        
        exit 1
    else
        local message="All backups verified successfully!\nTotal verified: $total_verified"
        send_notification "✅ Backup Verification Success" "$message" "good"
        
        log "All backup verifications passed"
    fi
}

# Script entry point
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi