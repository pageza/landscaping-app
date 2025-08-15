#!/bin/bash

# SSL Certificate Setup Script using Let's Encrypt and Certbot
# Automates SSL certificate generation and renewal for the Landscaping App

set -e

# Configuration
DOMAIN="${DOMAIN:-landscaping-app.com}"
API_DOMAIN="${API_DOMAIN:-api.landscaping-app.com}"
EMAIL="${EMAIL:-admin@landscaping-app.com}"
WEBROOT_PATH="${WEBROOT_PATH:-/var/www/html}"
CERT_PATH="/etc/letsencrypt/live"
NGINX_CONF_DIR="${NGINX_CONF_DIR:-/etc/nginx/sites-available}"
DOCKER_COMPOSE_FILE="${DOCKER_COMPOSE_FILE:-}"

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Error handling
error_exit() {
    log "ERROR: $1"
    exit 1
}

# Check prerequisites
check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if running as root
    if [[ $EUID -ne 0 ]]; then
        error_exit "This script must be run as root"
    fi
    
    # Check if certbot is installed
    if ! command -v certbot &> /dev/null; then
        log "Installing certbot..."
        apt-get update
        apt-get install -y certbot python3-certbot-nginx
    fi
    
    # Check if domains resolve to this server
    local server_ip=$(curl -s https://ipecho.net/plain)
    local domain_ip=$(dig +short "$DOMAIN" | tail -n1)
    
    if [[ "$server_ip" != "$domain_ip" ]]; then
        log "WARNING: Domain $DOMAIN does not resolve to this server IP ($server_ip vs $domain_ip)"
        read -p "Continue anyway? (y/N): " -r
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            exit 1
        fi
    fi
    
    log "Prerequisites check passed"
}

# Setup initial nginx configuration for ACME challenge
setup_initial_nginx() {
    log "Setting up initial nginx configuration..."
    
    cat > "${NGINX_CONF_DIR}/default" << EOF
server {
    listen 80;
    listen [::]:80;
    
    server_name $DOMAIN $API_DOMAIN;
    
    # ACME challenge location
    location /.well-known/acme-challenge/ {
        root $WEBROOT_PATH;
        try_files \$uri =404;
    }
    
    # Redirect all other HTTP traffic to HTTPS
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}
EOF
    
    # Create webroot directory
    mkdir -p "$WEBROOT_PATH/.well-known/acme-challenge"
    chown -R www-data:www-data "$WEBROOT_PATH"
    
    # Test nginx configuration
    nginx -t
    systemctl reload nginx || docker exec nginx nginx -s reload 2>/dev/null || true
    
    log "Initial nginx configuration completed"
}

# Obtain SSL certificates
obtain_certificates() {
    log "Obtaining SSL certificates from Let's Encrypt..."
    
    # Create certificates for main domain and API subdomain
    certbot certonly \
        --webroot \
        --webroot-path="$WEBROOT_PATH" \
        --email "$EMAIL" \
        --agree-tos \
        --no-eff-email \
        --domains "$DOMAIN,$API_DOMAIN" \
        --non-interactive \
        --expand
    
    if [[ $? -eq 0 ]]; then
        log "SSL certificates obtained successfully"
    else
        error_exit "Failed to obtain SSL certificates"
    fi
}

# Setup production nginx configuration with SSL
setup_ssl_nginx() {
    log "Setting up production nginx configuration with SSL..."
    
    cat > "${NGINX_CONF_DIR}/landscaping-app-ssl" << EOF
# Rate limiting
limit_req_zone \$binary_remote_addr zone=api:10m rate=10r/s;
limit_req_zone \$binary_remote_addr zone=web:10m rate=5r/s;

# Upstream servers
upstream api_backend {
    least_conn;
    server api:8080 max_fails=3 fail_timeout=30s;
    server api:8080 max_fails=3 fail_timeout=30s backup;
    keepalive 32;
}

upstream web_backend {
    least_conn;
    server web:8081 max_fails=3 fail_timeout=30s;
    server web:8081 max_fails=3 fail_timeout=30s backup;
    keepalive 32;
}

# Redirect HTTP to HTTPS
server {
    listen 80;
    listen [::]:80;
    server_name $DOMAIN $API_DOMAIN;
    
    # ACME challenge location
    location /.well-known/acme-challenge/ {
        root $WEBROOT_PATH;
        try_files \$uri =404;
    }
    
    location / {
        return 301 https://\$server_name\$request_uri;
    }
}

# Main website HTTPS
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name $DOMAIN;
    
    # SSL Configuration
    ssl_certificate ${CERT_PATH}/${DOMAIN}/fullchain.pem;
    ssl_certificate_key ${CERT_PATH}/${DOMAIN}/privkey.pem;
    ssl_trusted_certificate ${CERT_PATH}/${DOMAIN}/chain.pem;
    
    # Modern SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-RSA-AES128-SHA:DHE-RSA-AES256-SHA;
    ssl_prefer_server_ciphers off;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_stapling on;
    ssl_stapling_verify on;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' https:; frame-ancestors 'none';" always;
    
    # Gzip compression
    gzip on;
    gzip_vary on;
    gzip_min_length 1024;
    gzip_types text/plain text/css text/xml text/javascript application/javascript application/xml+rss application/json;
    
    # Client settings
    client_max_body_size 20M;
    client_body_timeout 60s;
    
    # API endpoints
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        
        proxy_pass http://api_backend/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # Main web application
    location / {
        limit_req zone=web burst=10 nodelay;
        
        proxy_pass http://web_backend/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # Health check endpoint
    location /health {
        access_log off;
        return 200 "healthy\\n";
        add_header Content-Type text/plain;
    }
    
    # Static file caching
    location ~* \\.(css|js|png|jpg|jpeg|gif|ico|svg|woff|woff2|ttf|eot)$ {
        expires 1y;
        add_header Cache-Control "public, immutable";
        access_log off;
    }
    
    # ACME challenge location
    location /.well-known/acme-challenge/ {
        root $WEBROOT_PATH;
        try_files \$uri =404;
    }
}

# API subdomain HTTPS
server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name $API_DOMAIN;
    
    # SSL Configuration
    ssl_certificate ${CERT_PATH}/${DOMAIN}/fullchain.pem;
    ssl_certificate_key ${CERT_PATH}/${DOMAIN}/privkey.pem;
    ssl_trusted_certificate ${CERT_PATH}/${DOMAIN}/chain.pem;
    
    # SSL settings (same as above)
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES128-GCM-SHA256:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-RSA-AES128-SHA256:ECDHE-RSA-AES256-SHA384:ECDHE-RSA-AES128-SHA:ECDHE-RSA-AES256-SHA:DHE-RSA-AES128-SHA256:DHE-RSA-AES256-SHA256:DHE-RSA-AES128-SHA:DHE-RSA-AES256-SHA;
    ssl_prefer_server_ciphers off;
    ssl_session_timeout 1d;
    ssl_session_cache shared:SSL:50m;
    ssl_stapling on;
    ssl_stapling_verify on;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "DENY" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    # Client settings
    client_max_body_size 20M;
    client_body_timeout 60s;
    
    # All requests go to API
    location / {
        limit_req zone=api burst=30 nodelay;
        
        proxy_pass http://api_backend/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        proxy_cache_bypass \$http_upgrade;
        
        proxy_connect_timeout 30s;
        proxy_send_timeout 30s;
        proxy_read_timeout 30s;
    }
    
    # ACME challenge location
    location /.well-known/acme-challenge/ {
        root $WEBROOT_PATH;
        try_files \$uri =404;
    }
}
EOF
    
    # Enable the new configuration
    ln -sf "${NGINX_CONF_DIR}/landscaping-app-ssl" /etc/nginx/sites-enabled/
    rm -f /etc/nginx/sites-enabled/default
    
    # Test nginx configuration
    nginx -t
    systemctl reload nginx || docker exec nginx nginx -s reload 2>/dev/null || true
    
    log "SSL nginx configuration completed"
}

# Setup automatic renewal
setup_renewal() {
    log "Setting up automatic certificate renewal..."
    
    # Create renewal script
    cat > /usr/local/bin/renew-certs.sh << 'EOF'
#!/bin/bash
# Certificate renewal script

LOG_FILE="/var/log/letsencrypt-renewal.log"

echo "[$(date)] Starting certificate renewal check..." >> "$LOG_FILE"

# Renew certificates
certbot renew --quiet --no-self-upgrade >> "$LOG_FILE" 2>&1

# Check if renewal was successful
if [ $? -eq 0 ]; then
    echo "[$(date)] Certificate renewal check completed successfully" >> "$LOG_FILE"
    
    # Reload nginx if certificates were renewed
    if [ -n "$(find /etc/letsencrypt/live -name '*.pem' -newer /var/log/nginx/access.log 2>/dev/null)" ]; then
        echo "[$(date)] Reloading nginx due to certificate renewal" >> "$LOG_FILE"
        systemctl reload nginx || docker exec nginx nginx -s reload 2>/dev/null || true
    fi
else
    echo "[$(date)] Certificate renewal failed" >> "$LOG_FILE"
fi
EOF
    
    chmod +x /usr/local/bin/renew-certs.sh
    
    # Create cron job for automatic renewal (twice daily as recommended)
    cat > /etc/cron.d/certbot-renewal << EOF
# Automatic certificate renewal for Let's Encrypt
SHELL=/bin/bash
PATH=/usr/local/sbin:/usr/local/bin:/sbin:/bin:/usr/sbin:/usr/bin

# Renew certificates twice daily at random times
$(shuf -i 0-59 -n 1) $(shuf -i 0-5 -n 1) * * * root /usr/local/bin/renew-certs.sh
$(shuf -i 0-59 -n 1) $(shuf -i 12-17 -n 1) * * * root /usr/local/bin/renew-certs.sh
EOF
    
    log "Automatic certificate renewal setup completed"
}

# Setup Docker integration
setup_docker_integration() {
    if [[ -n "$DOCKER_COMPOSE_FILE" && -f "$DOCKER_COMPOSE_FILE" ]]; then
        log "Setting up Docker integration..."
        
        # Create docker-compose override for SSL
        cat > "$(dirname "$DOCKER_COMPOSE_FILE")/docker-compose.ssl.yml" << EOF
version: '3.8'

services:
  nginx:
    volumes:
      - /etc/letsencrypt:/etc/letsencrypt:ro
      - $WEBROOT_PATH:$WEBROOT_PATH
    ports:
      - "80:80"
      - "443:443"
    environment:
      - DOMAIN=$DOMAIN
      - API_DOMAIN=$API_DOMAIN
EOF
        
        log "Docker SSL integration setup completed"
    fi
}

# Test SSL configuration
test_ssl() {
    log "Testing SSL configuration..."
    
    # Wait for services to be ready
    sleep 10
    
    # Test main domain
    if curl -fsSL "https://$DOMAIN/health" &>/dev/null; then
        log "✅ Main domain SSL test passed: https://$DOMAIN"
    else
        log "❌ Main domain SSL test failed"
        return 1
    fi
    
    # Test API domain
    if curl -fsSL "https://$API_DOMAIN/health" &>/dev/null; then
        log "✅ API domain SSL test passed: https://$API_DOMAIN"
    else
        log "❌ API domain SSL test failed"
        return 1
    fi
    
    # Test SSL certificate
    local cert_expiry=$(openssl x509 -in "${CERT_PATH}/${DOMAIN}/cert.pem" -noout -dates | grep notAfter | cut -d= -f2)
    log "SSL certificate expires: $cert_expiry"
    
    # Test SSL rating (optional - requires external service)
    # log "SSL rating test: https://www.ssllabs.com/ssltest/analyze.html?d=${DOMAIN}"
    
    log "SSL configuration test completed"
}

# Backup certificates
backup_certificates() {
    log "Creating certificate backup..."
    
    local backup_dir="/opt/ssl-backups"
    local backup_file="${backup_dir}/letsencrypt-backup-$(date +%Y%m%d_%H%M%S).tar.gz"
    
    mkdir -p "$backup_dir"
    
    tar -czf "$backup_file" -C /etc letsencrypt/
    
    # Keep only last 10 backups
    find "$backup_dir" -name "letsencrypt-backup-*.tar.gz" -type f | sort | head -n -10 | xargs rm -f
    
    log "Certificate backup created: $backup_file"
}

# Main setup function
main() {
    log "Starting SSL certificate setup with Let's Encrypt..."
    
    check_prerequisites
    setup_initial_nginx
    obtain_certificates
    setup_ssl_nginx
    setup_renewal
    setup_docker_integration
    backup_certificates
    test_ssl
    
    log "🔒 SSL certificate setup completed successfully!"
    
    echo ""
    echo "Certificate locations:"
    echo "Certificate: ${CERT_PATH}/${DOMAIN}/fullchain.pem"
    echo "Private Key: ${CERT_PATH}/${DOMAIN}/privkey.pem"
    echo "Chain: ${CERT_PATH}/${DOMAIN}/chain.pem"
    echo ""
    echo "Automatic renewal is configured to run twice daily"
    echo "Manual renewal command: certbot renew"
    echo ""
    echo "Test your SSL configuration:"
    echo "https://www.ssllabs.com/ssltest/analyze.html?d=${DOMAIN}"
}

# Handle different operations
case "${1:-setup}" in
    "setup")
        main
        ;;
    "renew")
        /usr/local/bin/renew-certs.sh
        ;;
    "test")
        test_ssl
        ;;
    "backup")
        backup_certificates
        ;;
    *)
        echo "Usage: $0 [setup|renew|test|backup]"
        echo ""
        echo "Commands:"
        echo "  setup   - Initial SSL certificate setup (default)"
        echo "  renew   - Manually renew certificates"
        echo "  test    - Test SSL configuration"
        echo "  backup  - Backup certificates"
        echo ""
        echo "Environment variables:"
        echo "  DOMAIN              - Main domain (default: landscaping-app.com)"
        echo "  API_DOMAIN          - API subdomain (default: api.landscaping-app.com)"
        echo "  EMAIL               - Contact email (default: admin@landscaping-app.com)"
        echo "  WEBROOT_PATH        - Webroot for ACME challenge (default: /var/www/html)"
        echo "  DOCKER_COMPOSE_FILE - Docker compose file for integration"
        exit 1
        ;;
esac