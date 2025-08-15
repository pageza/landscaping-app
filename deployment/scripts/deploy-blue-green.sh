#!/bin/bash

set -e

ENVIRONMENT=$1

if [[ -z "$ENVIRONMENT" ]]; then
    echo "Usage: $0 <environment>"
    exit 1
fi

CLUSTER_NAME="landscaping-${ENVIRONMENT}"
REGION="us-east-1"

echo "Starting blue-green deployment for environment: $ENVIRONMENT"

# Function to get current task definition ARN
get_current_task_def() {
    local service_name=$1
    aws ecs describe-services \
        --cluster "$CLUSTER_NAME" \
        --services "$service_name" \
        --query 'services[0].taskDefinition' \
        --output text
}

# Function to get latest task definition ARN
get_latest_task_def() {
    local family_name=$1
    aws ecs describe-task-definition \
        --task-definition "$family_name" \
        --query 'taskDefinition.taskDefinitionArn' \
        --output text
}

# Function to wait for service stability
wait_for_stability() {
    local service_name=$1
    echo "Waiting for $service_name to be stable..."
    
    aws ecs wait services-stable \
        --cluster "$CLUSTER_NAME" \
        --services "$service_name" \
        --region "$REGION"
    
    echo "$service_name is now stable"
}

# Function to check service health
check_service_health() {
    local service_name=$1
    local health_endpoint=$2
    
    echo "Checking health for $service_name..."
    
    # Get the service's load balancer target group
    local target_group_arn=$(aws elbv2 describe-target-groups \
        --names "landscaping-${service_name}-${ENVIRONMENT}-tg" \
        --query 'TargetGroups[0].TargetGroupArn' \
        --output text)
    
    # Check target health
    local healthy_targets=$(aws elbv2 describe-target-health \
        --target-group-arn "$target_group_arn" \
        --query 'length(TargetHealthDescriptions[?TargetHealth.State==`healthy`])' \
        --output text)
    
    if [[ "$healthy_targets" -gt 0 ]]; then
        echo "$service_name has $healthy_targets healthy targets"
        
        # Test the health endpoint if provided
        if [[ -n "$health_endpoint" ]]; then
            for i in {1..5}; do
                if curl -f "$health_endpoint" &>/dev/null; then
                    echo "$service_name health check passed"
                    return 0
                fi
                echo "Health check attempt $i failed, retrying in 10s..."
                sleep 10
            done
            echo "Health check failed after 5 attempts"
            return 1
        fi
        
        return 0
    else
        echo "$service_name has no healthy targets"
        return 1
    fi
}

# Function to rollback service
rollback_service() {
    local service_name=$1
    local previous_task_def=$2
    
    echo "Rolling back $service_name to $previous_task_def"
    
    aws ecs update-service \
        --cluster "$CLUSTER_NAME" \
        --service "$service_name" \
        --task-definition "$previous_task_def" \
        --region "$REGION"
    
    wait_for_stability "$service_name"
}

# Services to deploy
SERVICES=("landscaping-api" "landscaping-worker" "landscaping-web")

# Store current task definitions for potential rollback
declare -A PREVIOUS_TASK_DEFS
for service in "${SERVICES[@]}"; do
    PREVIOUS_TASK_DEFS[$service]=$(get_current_task_def "$service")
    echo "Current task definition for $service: ${PREVIOUS_TASK_DEFS[$service]}"
done

# Deploy new versions
DEPLOYED_SERVICES=()
for service in "${SERVICES[@]}"; do
    echo "Deploying $service..."
    
    # Get the latest task definition
    family_name="${service}-${ENVIRONMENT}"
    latest_task_def=$(get_latest_task_def "$family_name")
    
    echo "Updating $service to $latest_task_def"
    
    # Update the service
    aws ecs update-service \
        --cluster "$CLUSTER_NAME" \
        --service "$service" \
        --task-definition "$latest_task_def" \
        --region "$REGION"
    
    DEPLOYED_SERVICES+=("$service")
    
    # Wait for the service to be stable
    wait_for_stability "$service"
    
    # Health check based on service type
    case $service in
        "landscaping-api")
            health_url="https://api.landscaping-app.com/health"
            if [[ "$ENVIRONMENT" == "staging" ]]; then
                health_url="https://staging-api.landscaping-app.com/health"
            fi
            
            if ! check_service_health "$service" "$health_url"; then
                echo "Health check failed for $service, rolling back..."
                rollback_service "$service" "${PREVIOUS_TASK_DEFS[$service]}"
                exit 1
            fi
            ;;
            
        "landscaping-web")
            health_url="https://landscaping-app.com/health"
            if [[ "$ENVIRONMENT" == "staging" ]]; then
                health_url="https://staging.landscaping-app.com/health"
            fi
            
            if ! check_service_health "$service" "$health_url"; then
                echo "Health check failed for $service, rolling back..."
                rollback_service "$service" "${PREVIOUS_TASK_DEFS[$service]}"
                exit 1
            fi
            ;;
            
        "landscaping-worker")
            # Worker doesn't have HTTP health check, just verify it's running
            if ! check_service_health "$service"; then
                echo "Health check failed for $service, rolling back..."
                rollback_service "$service" "${PREVIOUS_TASK_DEFS[$service]}"
                exit 1
            fi
            ;;
    esac
    
    echo "$service deployed and healthy"
done

# Final system-wide health check
echo "Performing final system health check..."

# Test critical endpoints
ENDPOINTS=(
    "https://api.landscaping-app.com/health"
    "https://landscaping-app.com/health"
)

if [[ "$ENVIRONMENT" == "staging" ]]; then
    ENDPOINTS=(
        "https://staging-api.landscaping-app.com/health"
        "https://staging.landscaping-app.com/health"
    )
fi

for endpoint in "${ENDPOINTS[@]}"; do
    if ! curl -f "$endpoint" &>/dev/null; then
        echo "Final health check failed for $endpoint"
        
        # Rollback all services
        echo "Rolling back all services..."
        for service in "${DEPLOYED_SERVICES[@]}"; do
            rollback_service "$service" "${PREVIOUS_TASK_DEFS[$service]}"
        done
        
        exit 1
    fi
    echo "✓ $endpoint is healthy"
done

echo "Blue-green deployment completed successfully!"
echo "All services are running the new version and are healthy."

# Log deployment details
echo "Deployment completed at $(date)" >> "/tmp/deployment-${ENVIRONMENT}.log"
echo "Services deployed: ${DEPLOYED_SERVICES[*]}" >> "/tmp/deployment-${ENVIRONMENT}.log"