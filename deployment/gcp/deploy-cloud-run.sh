#!/bin/bash

# Google Cloud Run Deployment Script
# Deploys the Landscaping App to Google Cloud Run

set -e

# Configuration
PROJECT_ID="${GCP_PROJECT_ID:-landscaping-app-prod}"
REGION="${GCP_REGION:-us-central1}"
IMAGE_REGISTRY="gcr.io"
SERVICE_ACCOUNT_EMAIL="${SERVICE_ACCOUNT_EMAIL:-landscaping-app@${PROJECT_ID}.iam.gserviceaccount.com}"

# Service configurations
API_SERVICE_NAME="landscaping-api"
WORKER_SERVICE_NAME="landscaping-worker"
WEB_SERVICE_NAME="landscaping-web"

# Image tags
IMAGE_TAG="${IMAGE_TAG:-latest}"
API_IMAGE="${IMAGE_REGISTRY}/${PROJECT_ID}/${API_SERVICE_NAME}:${IMAGE_TAG}"
WORKER_IMAGE="${IMAGE_REGISTRY}/${PROJECT_ID}/${WORKER_SERVICE_NAME}:${IMAGE_TAG}"
WEB_IMAGE="${IMAGE_REGISTRY}/${PROJECT_ID}/${WEB_SERVICE_NAME}:${IMAGE_TAG}"

# Secrets (stored in Google Secret Manager)
DATABASE_URL_SECRET="projects/${PROJECT_ID}/secrets/database-url/versions/latest"
REDIS_URL_SECRET="projects/${PROJECT_ID}/secrets/redis-url/versions/latest"
JWT_SECRET_SECRET="projects/${PROJECT_ID}/secrets/jwt-secret/versions/latest"
ENCRYPTION_KEY_SECRET="projects/${PROJECT_ID}/secrets/encryption-key/versions/latest"
SESSION_SECRET_SECRET="projects/${PROJECT_ID}/secrets/session-secret/versions/latest"

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
    
    # Check if gcloud is installed and authenticated
    if ! command -v gcloud &> /dev/null; then
        error_exit "gcloud CLI is not installed"
    fi
    
    # Check if authenticated
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | grep -q .; then
        error_exit "Not authenticated with gcloud. Run 'gcloud auth login'"
    fi
    
    # Set project
    gcloud config set project "$PROJECT_ID"
    
    log "Prerequisites check passed"
}

# Enable required APIs
enable_apis() {
    log "Enabling required Google Cloud APIs..."
    
    gcloud services enable \
        cloudbuild.googleapis.com \
        cloudrun.googleapis.com \
        cloudsql.googleapis.com \
        redis.googleapis.com \
        secretmanager.googleapis.com \
        monitoring.googleapis.com \
        logging.googleapis.com \
        compute.googleapis.com \
        vpcaccess.googleapis.com
    
    log "APIs enabled successfully"
}

# Create VPC connector for private resource access
create_vpc_connector() {
    log "Creating VPC connector..."
    
    local connector_name="landscaping-connector"
    
    # Check if connector already exists
    if gcloud compute networks vpc-access connectors describe "$connector_name" \
        --region="$REGION" &>/dev/null; then
        log "VPC connector already exists"
        return 0
    fi
    
    gcloud compute networks vpc-access connectors create "$connector_name" \
        --network="default" \
        --range="10.8.0.0/28" \
        --region="$REGION" \
        --min-instances=2 \
        --max-instances=10
    
    log "VPC connector created successfully"
}

# Deploy API service
deploy_api() {
    log "Deploying API service..."
    
    gcloud run deploy "$API_SERVICE_NAME" \
        --image="$API_IMAGE" \
        --platform="managed" \
        --region="$REGION" \
        --allow-unauthenticated \
        --service-account="$SERVICE_ACCOUNT_EMAIL" \
        --memory="2Gi" \
        --cpu="2" \
        --min-instances="1" \
        --max-instances="10" \
        --concurrency="80" \
        --timeout="300s" \
        --port="8080" \
        --vpc-connector="landscaping-connector" \
        --set-env-vars="ENV=production,API_PORT=8080" \
        --set-secrets="DATABASE_URL=${DATABASE_URL_SECRET},REDIS_URL=${REDIS_URL_SECRET},JWT_SECRET=${JWT_SECRET_SECRET},ENCRYPTION_KEY=${ENCRYPTION_KEY_SECRET}" \
        --labels="app=landscaping-app,service=api,environment=production"
    
    log "API service deployed successfully"
}

# Deploy Worker service
deploy_worker() {
    log "Deploying Worker service..."
    
    gcloud run deploy "$WORKER_SERVICE_NAME" \
        --image="$WORKER_IMAGE" \
        --platform="managed" \
        --region="$REGION" \
        --no-allow-unauthenticated \
        --service-account="$SERVICE_ACCOUNT_EMAIL" \
        --memory="2Gi" \
        --cpu="2" \
        --min-instances="1" \
        --max-instances="5" \
        --concurrency="10" \
        --timeout="3600s" \
        --vpc-connector="landscaping-connector" \
        --set-env-vars="ENV=production,WORKER_CONCURRENCY=10" \
        --set-secrets="DATABASE_URL=${DATABASE_URL_SECRET},REDIS_URL=${REDIS_URL_SECRET}" \
        --labels="app=landscaping-app,service=worker,environment=production"
    
    log "Worker service deployed successfully"
}

# Deploy Web service
deploy_web() {
    log "Deploying Web service..."
    
    gcloud run deploy "$WEB_SERVICE_NAME" \
        --image="$WEB_IMAGE" \
        --platform="managed" \
        --region="$REGION" \
        --allow-unauthenticated \
        --service-account="$SERVICE_ACCOUNT_EMAIL" \
        --memory="1Gi" \
        --cpu="1" \
        --min-instances="1" \
        --max-instances="10" \
        --concurrency="100" \
        --timeout="60s" \
        --port="8081" \
        --set-env-vars="ENV=production,WEB_PORT=8081,API_BASE_URL=https://${API_SERVICE_NAME}-${REGION}.a.run.app" \
        --set-secrets="SESSION_SECRET=${SESSION_SECRET_SECRET}" \
        --labels="app=landscaping-app,service=web,environment=production"
    
    log "Web service deployed successfully"
}

# Setup Cloud Load Balancer
setup_load_balancer() {
    log "Setting up Cloud Load Balancer..."
    
    local lb_name="landscaping-lb"
    local backend_service_api="landscaping-api-backend"
    local backend_service_web="landscaping-web-backend"
    local url_map="landscaping-url-map"
    local target_proxy="landscaping-target-proxy"
    local forwarding_rule="landscaping-forwarding-rule"
    
    # Create NEGs for Cloud Run services
    gcloud compute network-endpoint-groups create "$API_SERVICE_NAME-neg" \
        --region="$REGION" \
        --network-endpoint-type="serverless" \
        --cloud-run-service="$API_SERVICE_NAME" || true
    
    gcloud compute network-endpoint-groups create "$WEB_SERVICE_NAME-neg" \
        --region="$REGION" \
        --network-endpoint-type="serverless" \
        --cloud-run-service="$WEB_SERVICE_NAME" || true
    
    # Create backend services
    gcloud compute backend-services create "$backend_service_api" \
        --global \
        --load-balancing-scheme="EXTERNAL" \
        --protocol="HTTP" || true
    
    gcloud compute backend-services create "$backend_service_web" \
        --global \
        --load-balancing-scheme="EXTERNAL" \
        --protocol="HTTP" || true
    
    # Add NEGs to backend services
    gcloud compute backend-services add-backend "$backend_service_api" \
        --global \
        --network-endpoint-group="$API_SERVICE_NAME-neg" \
        --network-endpoint-group-region="$REGION"
    
    gcloud compute backend-services add-backend "$backend_service_web" \
        --global \
        --network-endpoint-group="$WEB_SERVICE_NAME-neg" \
        --network-endpoint-group-region="$REGION"
    
    # Create URL map
    gcloud compute url-maps create "$url_map" \
        --default-service="$backend_service_web" || true
    
    # Add path matcher for API
    gcloud compute url-maps add-path-matcher "$url_map" \
        --path-matcher-name="api-matcher" \
        --default-service="$backend_service_api" \
        --path-rules="/api/*=$backend_service_api"
    
    # Create target HTTP proxy
    gcloud compute target-http-proxies create "$target_proxy" \
        --url-map="$url_map" || true
    
    # Create forwarding rule
    gcloud compute forwarding-rules create "$forwarding_rule" \
        --global \
        --target-http-proxy="$target_proxy" \
        --ports="80" || true
    
    log "Load balancer setup completed"
}

# Setup monitoring and alerting
setup_monitoring() {
    log "Setting up monitoring and alerting..."
    
    # Create notification channel (replace with your email)
    local notification_channel=$(gcloud alpha monitoring channels create \
        --display-name="Landscaping App Alerts" \
        --type="email" \
        --channel-labels="email_address=devops@landscaping-app.com" \
        --format="value(name)" || echo "")
    
    if [[ -n "$notification_channel" ]]; then
        # Create alert policies
        cat > /tmp/api-uptime-policy.yaml << EOF
displayName: "API Service Down"
conditions:
  - displayName: "API Service Uptime Check"
    conditionThreshold:
      filter: 'resource.type="cloud_run_revision" resource.labels.service_name="${API_SERVICE_NAME}"'
      comparison: COMPARISON_GREATER_THAN
      thresholdValue: 0.1
      duration: 300s
      aggregations:
        - alignmentPeriod: 300s
          perSeriesAligner: ALIGN_RATE
          crossSeriesReducer: REDUCE_MEAN
          groupByFields:
            - resource.label.service_name
notificationChannels:
  - ${notification_channel}
enabled: true
EOF
        
        gcloud alpha monitoring policies create --policy-from-file=/tmp/api-uptime-policy.yaml
        
        rm -f /tmp/api-uptime-policy.yaml
    fi
    
    log "Monitoring setup completed"
}

# Health check
health_check() {
    log "Performing health check..."
    
    local api_url=$(gcloud run services describe "$API_SERVICE_NAME" \
        --region="$REGION" \
        --format="value(status.url)")
    
    local web_url=$(gcloud run services describe "$WEB_SERVICE_NAME" \
        --region="$REGION" \
        --format="value(status.url)")
    
    # Wait for services to be ready
    sleep 30
    
    # Check API health
    if curl -f "${api_url}/health" &>/dev/null; then
        log "✅ API service is healthy: ${api_url}/health"
    else
        error_exit "❌ API service health check failed"
    fi
    
    # Check Web health
    if curl -f "${web_url}/health" &>/dev/null; then
        log "✅ Web service is healthy: ${web_url}/health"
    else
        error_exit "❌ Web service health check failed"
    fi
    
    log "All health checks passed"
}

# Rollback function
rollback() {
    local service_name="$1"
    local previous_revision="$2"
    
    log "Rolling back $service_name to revision $previous_revision..."
    
    gcloud run services update-traffic "$service_name" \
        --to-revisions="$previous_revision=100" \
        --region="$REGION"
    
    log "Rollback completed for $service_name"
}

# Main deployment function
main() {
    log "Starting Google Cloud Run deployment..."
    
    # Trap to handle errors
    trap 'log "Deployment failed"; exit 1' ERR
    
    check_prerequisites
    enable_apis
    create_vpc_connector
    
    # Deploy services
    deploy_api
    deploy_worker
    deploy_web
    
    # Setup load balancer
    setup_load_balancer
    
    # Setup monitoring
    setup_monitoring
    
    # Health check
    health_check
    
    log "🎉 Deployment completed successfully!"
    
    # Output service URLs
    local api_url=$(gcloud run services describe "$API_SERVICE_NAME" \
        --region="$REGION" \
        --format="value(status.url)")
    
    local web_url=$(gcloud run services describe "$WEB_SERVICE_NAME" \
        --region="$REGION" \
        --format="value(status.url)")
    
    echo ""
    echo "Service URLs:"
    echo "API: $api_url"
    echo "Web: $web_url"
    echo ""
    echo "Load Balancer IP:"
    gcloud compute forwarding-rules describe "landscaping-forwarding-rule" \
        --global \
        --format="value(IPAddress)"
}

# Handle different deployment modes
case "${1:-deploy}" in
    "deploy")
        main
        ;;
    "rollback")
        if [[ $# -lt 3 ]]; then
            echo "Usage: $0 rollback <service_name> <revision>"
            exit 1
        fi
        rollback "$2" "$3"
        ;;
    "health-check")
        health_check
        ;;
    *)
        echo "Usage: $0 [deploy|rollback|health-check]"
        exit 1
        ;;
esac