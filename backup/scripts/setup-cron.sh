#!/bin/bash

# Setup cron jobs for automated backups
# Run this script to install backup cron jobs

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Create cron jobs
cat << 'EOF' > /tmp/backup-cron
# Landscaping App Database Backups
# PostgreSQL backup every day at 2 AM
0 2 * * * /home/hermes/Projects/landscaping-app/backup/scripts/postgres-backup.sh >> /var/log/postgres-backup-cron.log 2>&1

# Redis backup every 6 hours
0 */6 * * * /home/hermes/Projects/landscaping-app/backup/scripts/redis-backup.sh >> /var/log/redis-backup-cron.log 2>&1

# Cleanup old logs weekly
0 3 * * 0 find /var/log -name "*backup*.log" -mtime +30 -delete

# Test backup integrity monthly
0 4 1 * * /home/hermes/Projects/landscaping-app/backup/scripts/verify-backups.sh >> /var/log/backup-verify.log 2>&1
EOF

# Install cron jobs
crontab -l > /tmp/current-cron 2>/dev/null || touch /tmp/current-cron
cat /tmp/current-cron /tmp/backup-cron | crontab -

echo "Backup cron jobs installed successfully!"
echo "Current crontab:"
crontab -l

# Clean up
rm -f /tmp/backup-cron /tmp/current-cron