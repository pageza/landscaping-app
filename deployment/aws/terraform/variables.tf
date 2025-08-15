variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "us-east-1"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
  
  validation {
    condition     = contains(["development", "staging", "production"], var.environment)
    error_message = "Environment must be development, staging, or production."
  }
}

variable "project_name" {
  description = "Project name"
  type        = string
  default     = "landscaping-app"
}

variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

# Database variables
variable "db_instance_class" {
  description = "RDS instance class"
  type        = string
  default     = "db.t3.medium"
}

variable "db_allocated_storage" {
  description = "RDS allocated storage in GB"
  type        = number
  default     = 100
}

variable "db_max_allocated_storage" {
  description = "RDS maximum allocated storage in GB"
  type        = number
  default     = 1000
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "landscaping_prod"
}

variable "db_username" {
  description = "Database username"
  type        = string
  default     = "landscaping_user"
}

variable "db_password" {
  description = "Database password"
  type        = string
  sensitive   = true
}

variable "db_backup_retention_period" {
  description = "Database backup retention period in days"
  type        = number
  default     = 30
}

# Redis variables
variable "redis_node_type" {
  description = "ElastiCache node type"
  type        = string
  default     = "cache.t3.medium"
}

variable "redis_num_cache_nodes" {
  description = "Number of cache nodes"
  type        = number
  default     = 2
}

variable "redis_auth_token" {
  description = "Redis auth token"
  type        = string
  sensitive   = true
}

# ECS variables
variable "api_cpu" {
  description = "CPU units for API service"
  type        = number
  default     = 1024
}

variable "api_memory" {
  description = "Memory for API service"
  type        = number
  default     = 2048
}

variable "api_desired_count" {
  description = "Desired count for API service"
  type        = number
  default     = 2
}

variable "api_max_capacity" {
  description = "Maximum capacity for API service auto scaling"
  type        = number
  default     = 10
}

variable "api_min_capacity" {
  description = "Minimum capacity for API service auto scaling"
  type        = number
  default     = 2
}

variable "worker_cpu" {
  description = "CPU units for Worker service"
  type        = number
  default     = 1024
}

variable "worker_memory" {
  description = "Memory for Worker service"
  type        = number
  default     = 2048
}

variable "worker_desired_count" {
  description = "Desired count for Worker service"
  type        = number
  default     = 2
}

variable "worker_max_capacity" {
  description = "Maximum capacity for Worker service auto scaling"
  type        = number
  default     = 5
}

variable "worker_min_capacity" {
  description = "Minimum capacity for Worker service auto scaling"
  type        = number
  default     = 1
}

variable "web_cpu" {
  description = "CPU units for Web service"
  type        = number
  default     = 512
}

variable "web_memory" {
  description = "Memory for Web service"
  type        = number
  default     = 1024
}

variable "web_desired_count" {
  description = "Desired count for Web service"
  type        = number
  default     = 2
}

variable "web_max_capacity" {
  description = "Maximum capacity for Web service auto scaling"
  type        = number
  default     = 10
}

variable "web_min_capacity" {
  description = "Minimum capacity for Web service auto scaling"
  type        = number
  default     = 2
}

# Domain and SSL
variable "domain_name" {
  description = "Domain name for the application"
  type        = string
  default     = "landscaping-app.com"
}

variable "api_domain_name" {
  description = "API domain name"
  type        = string
  default     = "api.landscaping-app.com"
}

variable "certificate_arn" {
  description = "ARN of the SSL certificate"
  type        = string
  default     = ""
}

# Monitoring
variable "enable_container_insights" {
  description = "Enable CloudWatch Container Insights"
  type        = bool
  default     = true
}

variable "log_retention_in_days" {
  description = "CloudWatch logs retention in days"
  type        = number
  default     = 30
}

# Auto Scaling
variable "enable_auto_scaling" {
  description = "Enable auto scaling for ECS services"
  type        = bool
  default     = true
}

variable "scale_up_cpu_threshold" {
  description = "CPU threshold for scaling up"
  type        = number
  default     = 70
}

variable "scale_down_cpu_threshold" {
  description = "CPU threshold for scaling down"
  type        = number
  default     = 30
}

variable "scale_up_memory_threshold" {
  description = "Memory threshold for scaling up"
  type        = number
  default     = 80
}

variable "scale_down_memory_threshold" {
  description = "Memory threshold for scaling down"
  type        = number
  default     = 40
}

# Secrets
variable "jwt_secret" {
  description = "JWT secret key"
  type        = string
  sensitive   = true
}

variable "encryption_key" {
  description = "Encryption key for application data"
  type        = string
  sensitive   = true
}

variable "session_secret" {
  description = "Session secret for web application"
  type        = string
  sensitive   = true
}

# External Services
variable "openai_api_key" {
  description = "OpenAI API key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "stripe_secret_key" {
  description = "Stripe secret key"
  type        = string
  sensitive   = true
  default     = ""
}

variable "twilio_account_sid" {
  description = "Twilio Account SID"
  type        = string
  sensitive   = true
  default     = ""
}

variable "twilio_auth_token" {
  description = "Twilio Auth Token"
  type        = string
  sensitive   = true
  default     = ""
}

# Tags
variable "additional_tags" {
  description = "Additional tags to apply to all resources"
  type        = map(string)
  default     = {}
}

# Backup
variable "enable_backup" {
  description = "Enable automated backups"
  type        = bool
  default     = true
}

variable "backup_schedule" {
  description = "Backup schedule expression"
  type        = string
  default     = "cron(0 2 * * ? *)"  # Daily at 2 AM UTC
}

# Security
variable "enable_waf" {
  description = "Enable AWS WAF"
  type        = bool
  default     = true
}

variable "enable_shield" {
  description = "Enable AWS Shield Advanced"
  type        = bool
  default     = false
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access the application"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

# Cost Optimization
variable "use_spot_instances" {
  description = "Use Fargate Spot for cost optimization"
  type        = bool
  default     = false
}

variable "fargate_spot_percentage" {
  description = "Percentage of tasks to run on Fargate Spot"
  type        = number
  default     = 50
  
  validation {
    condition     = var.fargate_spot_percentage >= 0 && var.fargate_spot_percentage <= 100
    error_message = "Fargate Spot percentage must be between 0 and 100."
  }
}