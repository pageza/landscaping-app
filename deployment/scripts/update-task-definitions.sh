#!/bin/bash

set -e

ENVIRONMENT=$1
IMAGE_TAG=$2

if [[ -z "$ENVIRONMENT" || -z "$IMAGE_TAG" ]]; then
    echo "Usage: $0 <environment> <image_tag>"
    exit 1
fi

REGISTRY="ghcr.io/landscaping-app/landscaping-app"
CLUSTER_NAME="landscaping-${ENVIRONMENT}"

echo "Updating task definitions for environment: $ENVIRONMENT"
echo "Using image tag: $IMAGE_TAG"

# Update API task definition
cat > api-task-definition.json << EOF
{
  "family": "landscaping-api-${ENVIRONMENT}",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/landscaping-task-role",
  "containerDefinitions": [
    {
      "name": "landscaping-api",
      "image": "${REGISTRY}/api:${IMAGE_TAG}",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "essential": true,
      "environment": [
        {
          "name": "ENV",
          "value": "${ENVIRONMENT}"
        },
        {
          "name": "API_PORT",
          "value": "8080"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:ACCOUNT:secret:landscaping/${ENVIRONMENT}/database-url"
        },
        {
          "name": "REDIS_URL",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:ACCOUNT:secret:landscaping/${ENVIRONMENT}/redis-url"
        },
        {
          "name": "JWT_SECRET",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:ACCOUNT:secret:landscaping/${ENVIRONMENT}/jwt-secret"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/landscaping-api-${ENVIRONMENT}",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "/api --health-check"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
EOF

# Register new task definition for API
aws ecs register-task-definition --cli-input-json file://api-task-definition.json

# Update Worker task definition
cat > worker-task-definition.json << EOF
{
  "family": "landscaping-worker-${ENVIRONMENT}",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/landscaping-task-role",
  "containerDefinitions": [
    {
      "name": "landscaping-worker",
      "image": "${REGISTRY}/worker:${IMAGE_TAG}",
      "essential": true,
      "environment": [
        {
          "name": "ENV",
          "value": "${ENVIRONMENT}"
        },
        {
          "name": "WORKER_CONCURRENCY",
          "value": "10"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:ACCOUNT:secret:landscaping/${ENVIRONMENT}/database-url"
        },
        {
          "name": "REDIS_URL",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:ACCOUNT:secret:landscaping/${ENVIRONMENT}/redis-url"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/landscaping-worker-${ENVIRONMENT}",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "/worker --health-check"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
EOF

# Register new task definition for Worker
aws ecs register-task-definition --cli-input-json file://worker-task-definition.json

# Update Web task definition
cat > web-task-definition.json << EOF
{
  "family": "landscaping-web-${ENVIRONMENT}",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/landscaping-task-role",
  "containerDefinitions": [
    {
      "name": "landscaping-web",
      "image": "${REGISTRY}/web:${IMAGE_TAG}",
      "portMappings": [
        {
          "containerPort": 8081,
          "protocol": "tcp"
        }
      ],
      "essential": true,
      "environment": [
        {
          "name": "ENV",
          "value": "${ENVIRONMENT}"
        },
        {
          "name": "WEB_PORT",
          "value": "8081"
        },
        {
          "name": "API_BASE_URL",
          "value": "https://api.landscaping-app.com"
        }
      ],
      "secrets": [
        {
          "name": "SESSION_SECRET",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:ACCOUNT:secret:landscaping/${ENVIRONMENT}/session-secret"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/landscaping-web-${ENVIRONMENT}",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      },
      "healthCheck": {
        "command": ["CMD-SHELL", "/web-server --health-check"],
        "interval": 30,
        "timeout": 5,
        "retries": 3,
        "startPeriod": 60
      }
    }
  ]
}
EOF

# Register new task definition for Web
aws ecs register-task-definition --cli-input-json file://web-task-definition.json

# Migration task definition
cat > migrate-task-definition.json << EOF
{
  "family": "landscaping-migrate-${ENVIRONMENT}",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "256",
  "memory": "512",
  "executionRoleArn": "arn:aws:iam::ACCOUNT:role/ecsTaskExecutionRole",
  "taskRoleArn": "arn:aws:iam::ACCOUNT:role/landscaping-task-role",
  "containerDefinitions": [
    {
      "name": "landscaping-migrate",
      "image": "${REGISTRY}/api:${IMAGE_TAG}",
      "command": ["/api", "migrate", "up"],
      "essential": true,
      "environment": [
        {
          "name": "ENV",
          "value": "${ENVIRONMENT}"
        }
      ],
      "secrets": [
        {
          "name": "DATABASE_URL",
          "valueFrom": "arn:aws:secretsmanager:us-east-1:ACCOUNT:secret:landscaping/${ENVIRONMENT}/database-url"
        }
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/landscaping-migrate-${ENVIRONMENT}",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
EOF

# Register new task definition for Migration
aws ecs register-task-definition --cli-input-json file://migrate-task-definition.json

echo "Task definitions updated successfully!"

# Clean up temporary files
rm -f api-task-definition.json worker-task-definition.json web-task-definition.json migrate-task-definition.json