#!/bin/bash

# Docker Swarm Deployment Script
# Deploys the Landscaping App using Docker Swarm for self-hosted environments

set -e

# Configuration
STACK_NAME="landscaping-app"
COMPOSE_FILE="docker-compose.swarm.yml"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Network configuration
NETWORK_NAME="${STACK_NAME}_overlay"
TRAEFIK_NETWORK="traefik-public"

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
    
    # Check if Docker is running
    if ! docker info &>/dev/null; then
        error_exit "Docker is not running"
    fi
    
    # Check if Docker Swarm is initialized
    if ! docker info | grep -q "Swarm: active"; then
        log "Initializing Docker Swarm..."
        docker swarm init
    fi
    
    # Check if Compose file exists
    if [[ ! -f "$SCRIPT_DIR/$COMPOSE_FILE" ]]; then
        error_exit "Compose file not found: $SCRIPT_DIR/$COMPOSE_FILE"
    fi
    
    log "Prerequisites check passed"
}

# Setup Docker networks
setup_networks() {
    log "Setting up Docker networks..."
    
    # Create overlay network for the application
    if ! docker network ls | grep -q "$NETWORK_NAME"; then
        docker network create \
            --driver overlay \
            --attachable \
            "$NETWORK_NAME"
        log "Created overlay network: $NETWORK_NAME"
    fi
    
    # Create Traefik public network if it doesn't exist
    if ! docker network ls | grep -q "$TRAEFIK_NETWORK"; then
        docker network create \
            --driver overlay \
            --attachable \
            "$TRAEFIK_NETWORK"
        log "Created Traefik network: $TRAEFIK_NETWORK"
    fi
}

# Setup Docker configs and secrets
setup_configs_and_secrets() {
    log "Setting up Docker configs and secrets..."
    
    # Create secrets from files if they exist
    local secrets_dir="$SCRIPT_DIR/secrets"
    
    if [[ -d "$secrets_dir" ]]; then
        for secret_file in "$secrets_dir"/*.txt; do
            if [[ -f "$secret_file" ]]; then
                local secret_name=$(basename "$secret_file" .txt)
                local docker_secret_name="${STACK_NAME}_${secret_name}"
                
                if ! docker secret ls | grep -q "$docker_secret_name"; then
                    docker secret create "$docker_secret_name" "$secret_file"
                    log "Created secret: $docker_secret_name"
                fi
            fi
        done
    fi
    
    # Create configs from files if they exist
    local configs_dir="$SCRIPT_DIR/configs"
    
    if [[ -d "$configs_dir" ]]; then
        for config_file in "$configs_dir"/*; do
            if [[ -f "$config_file" ]]; then
                local config_name=$(basename "$config_file")
                local docker_config_name="${STACK_NAME}_${config_name}"
                
                if ! docker config ls | grep -q "$docker_config_name"; then
                    docker config create "$docker_config_name" "$config_file"
                    log "Created config: $docker_config_name"
                fi
            fi
        done
    fi
}

# Deploy the stack
deploy_stack() {
    log "Deploying Docker stack: $STACK_NAME..."
    
    cd "$SCRIPT_DIR"
    
    # Deploy the stack
    docker stack deploy \
        --compose-file "$COMPOSE_FILE" \
        --with-registry-auth \
        "$STACK_NAME"
    
    log "Stack deployment initiated"
}

# Wait for services to be ready
wait_for_services() {
    log "Waiting for services to be ready..."
    
    local max_wait=300  # 5 minutes
    local elapsed=0
    
    while [[ $elapsed -lt $max_wait ]]; do
        local ready_services=0
        local total_services=0
        
        while IFS= read -r line; do
            if [[ "$line" =~ ^[[:space:]]*([0-9]+)/([0-9]+) ]]; then
                local current="${BASH_REMATCH[1]}"
                local desired="${BASH_REMATCH[2]}"
                total_services=$((total_services + 1))
                
                if [[ "$current" == "$desired" ]] && [[ "$current" -gt 0 ]]; then
                    ready_services=$((ready_services + 1))
                fi
            fi
        done < <(docker stack services "$STACK_NAME" --format "table {{.Replicas}}" | tail -n +2)
        
        if [[ $ready_services -eq $total_services ]] && [[ $total_services -gt 0 ]]; then
            log "All services are ready ($ready_services/$total_services)"
            return 0
        fi
        
        log "Waiting for services... ($ready_services/$total_services ready)"
        sleep 10
        elapsed=$((elapsed + 10))
    done
    
    error_exit "Timeout waiting for services to be ready"
}

# Health check
health_check() {
    log "Performing health check..."
    
    # Get the manager node IP
    local manager_ip=$(docker node ls --filter "role=manager" --format "{{.Hostname}}" | head -1)
    
    # Wait a bit for load balancer to be ready
    sleep 30
    
    # Check API health
    if curl -f "http://${manager_ip}/api/health" &>/dev/null; then
        log "✅ API service is healthy"
    else
        log "❌ API service health check failed"
        return 1
    fi
    
    # Check Web health
    if curl -f "http://${manager_ip}/health" &>/dev/null; then
        log "✅ Web service is healthy"
    else
        log "❌ Web service health check failed"
        return 1
    fi
    
    log "All health checks passed"
}

# Setup monitoring
deploy_monitoring() {
    log "Deploying monitoring stack..."
    
    local monitoring_compose="docker-compose.monitoring.yml"
    
    if [[ -f "$SCRIPT_DIR/$monitoring_compose" ]]; then
        docker stack deploy \
            --compose-file "$SCRIPT_DIR/$monitoring_compose" \
            --with-registry-auth \
            "${STACK_NAME}-monitoring"
        
        log "Monitoring stack deployed"
    else
        log "Monitoring compose file not found, skipping monitoring deployment"
    fi
}

# Deploy Traefik as reverse proxy
deploy_traefik() {
    log "Deploying Traefik reverse proxy..."
    
    local traefik_compose="docker-compose.traefik.yml"
    
    if [[ -f "$SCRIPT_DIR/$traefik_compose" ]]; then
        docker stack deploy \
            --compose-file "$SCRIPT_DIR/$traefik_compose" \
            traefik
        
        log "Traefik deployed"
    else
        log "Traefik compose file not found, skipping Traefik deployment"
    fi
}

# Setup backup
setup_backup() {
    log "Setting up backup cronjobs..."
    
    # Create backup directories
    mkdir -p /opt/landscaping-backups/{postgres,redis,volumes}
    
    # Setup backup cronjob
    cat > /tmp/swarm-backup-cron << 'EOF'
# Landscaping App Docker Swarm Backups
# PostgreSQL backup every day at 2 AM
0 2 * * * docker exec $(docker ps -q -f name=landscaping-app_postgres) pg_dump -U postgres landscaping_prod | gzip > /opt/landscaping-backups/postgres/backup_$(date +\%Y\%m\%d_\%H\%M\%S).sql.gz

# Redis backup every 6 hours
0 */6 * * * docker exec $(docker ps -q -f name=landscaping-app_redis) redis-cli --rdb - | gzip > /opt/landscaping-backups/redis/backup_$(date +\%Y\%m\%d_\%H\%M\%S).rdb.gz

# Volume backup weekly
0 3 * * 0 tar -czf /opt/landscaping-backups/volumes/volumes_$(date +\%Y\%m\%d).tar.gz -C /var/lib/docker/volumes/ .

# Cleanup old backups
0 4 * * * find /opt/landscaping-backups -name "*.gz" -mtime +7 -delete
EOF
    
    # Install backup cronjobs
    crontab -l > /tmp/current-cron 2>/dev/null || touch /tmp/current-cron
    cat /tmp/current-cron /tmp/swarm-backup-cron | sort -u | crontab -
    
    rm -f /tmp/swarm-backup-cron /tmp/current-cron
    
    log "Backup cronjobs installed"
}

# Rolling update
rolling_update() {
    local service_name="$1"
    local new_image="$2"
    
    if [[ -z "$service_name" || -z "$new_image" ]]; then
        error_exit "Usage: rolling_update <service_name> <new_image>"
    fi
    
    log "Performing rolling update for $service_name with image $new_image..."
    
    docker service update \
        --image "$new_image" \
        --update-parallelism 1 \
        --update-delay 30s \
        --update-failure-action rollback \
        "${STACK_NAME}_${service_name}"
    
    log "Rolling update completed for $service_name"
}

# Scale service
scale_service() {
    local service_name="$1"
    local replicas="$2"
    
    if [[ -z "$service_name" || -z "$replicas" ]]; then
        error_exit "Usage: scale_service <service_name> <replicas>"
    fi
    
    log "Scaling $service_name to $replicas replicas..."
    
    docker service scale "${STACK_NAME}_${service_name}=$replicas"
    
    log "Service $service_name scaled to $replicas replicas"
}

# Show stack status
show_status() {
    log "Stack Status:"
    echo ""
    
    echo "Services:"
    docker stack services "$STACK_NAME"
    echo ""
    
    echo "Tasks:"
    docker stack ps "$STACK_NAME" --no-trunc
    echo ""
    
    if docker stack ls | grep -q "${STACK_NAME}-monitoring"; then
        echo "Monitoring Services:"
        docker stack services "${STACK_NAME}-monitoring"
        echo ""
    fi
}

# Remove stack
remove_stack() {
    log "Removing Docker stack: $STACK_NAME..."
    
    # Remove main stack
    docker stack rm "$STACK_NAME"
    
    # Remove monitoring stack if exists
    if docker stack ls | grep -q "${STACK_NAME}-monitoring"; then
        docker stack rm "${STACK_NAME}-monitoring"
    fi
    
    # Wait for stacks to be removed
    log "Waiting for stacks to be removed..."
    sleep 30
    
    # Clean up networks
    docker network rm "$NETWORK_NAME" 2>/dev/null || true
    
    log "Stack removed successfully"
}

# Main deployment function
main() {
    log "Starting Docker Swarm deployment..."
    
    check_prerequisites
    setup_networks
    setup_configs_and_secrets
    
    # Deploy Traefik if needed
    deploy_traefik
    
    # Deploy main application stack
    deploy_stack
    
    # Wait for services
    wait_for_services
    
    # Deploy monitoring
    deploy_monitoring
    
    # Setup backups
    setup_backup
    
    # Health check
    if health_check; then
        log "🎉 Deployment completed successfully!"
        show_status
    else
        error_exit "Deployment failed health check"
    fi
}

# Handle different commands
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "update")
        if [[ $# -lt 3 ]]; then
            echo "Usage: $0 update <service_name> <new_image>"
            exit 1
        fi
        rolling_update "$2" "$3"
        ;;
    "scale")
        if [[ $# -lt 3 ]]; then
            echo "Usage: $0 scale <service_name> <replicas>"
            exit 1
        fi
        scale_service "$2" "$3"
        ;;
    "status")
        show_status
        ;;
    "health")
        health_check
        ;;
    "remove")
        remove_stack
        ;;
    *)
        echo "Usage: $0 [deploy|update|scale|status|health|remove]"
        echo ""
        echo "Commands:"
        echo "  deploy                    - Deploy the full stack"
        echo "  update <service> <image>  - Perform rolling update"
        echo "  scale <service> <count>   - Scale service"
        echo "  status                    - Show stack status"
        echo "  health                    - Perform health check"
        echo "  remove                    - Remove the stack"
        exit 1
        ;;
esac