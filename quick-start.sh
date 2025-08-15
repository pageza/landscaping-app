#!/bin/bash

echo "🌱 Landscaping SaaS Platform - Quick Start"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check prerequisites
check_prerequisites() {
    echo -e "${BLUE}Checking prerequisites...${NC}"
    
    if ! command -v docker &> /dev/null; then
        echo -e "${RED}❌ Docker is not installed. Please install Docker first.${NC}"
        exit 1
    fi
    
    if ! docker compose version &> /dev/null; then
        echo -e "${RED}❌ Docker Compose is not installed. Please install Docker Compose first.${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✅ All prerequisites met!${NC}"
}

# Setup environment
setup_environment() {
    echo -e "${BLUE}Setting up environment...${NC}"
    
    if [ ! -f "docker/.env.dev" ]; then
        echo -e "${YELLOW}📝 Creating environment file...${NC}"
        cp docker/.env.template docker/.env.dev 2>/dev/null || {
            echo -e "${YELLOW}📝 Creating default environment file...${NC}"
            cat > docker/.env.dev << EOF
# Database Configuration
POSTGRES_DB=landscaping_db
POSTGRES_USER=landscaping_user
POSTGRES_PASSWORD=secure_password_123

# Redis Configuration
REDIS_URL=redis://redis:6379

# JWT Configuration (change in production!)
JWT_SECRET=your_jwt_secret_that_should_be_at_least_32_characters_long
JWT_REFRESH_SECRET=your_refresh_secret_that_should_be_at_least_32_characters_long

# Application Configuration
APP_ENV=development
APP_PORT=8080
WEB_PORT=8081

# LLM Configuration (optional)
LLM_PROVIDER=ollama
OLLAMA_URL=http://host.docker.internal:11434

# Storage Configuration
STORAGE_PROVIDER=local
LOCAL_STORAGE_PATH=/app/uploads

# Email Configuration (optional)
EMAIL_PROVIDER=smtp
SMTP_HOST=localhost
SMTP_PORT=587
SMTP_USER=
SMTP_PASSWORD=

# SMS Configuration (optional)
SMS_PROVIDER=twilio
TWILIO_ACCOUNT_SID=
TWILIO_AUTH_TOKEN=
EOF
        }
    fi
    
    echo -e "${GREEN}✅ Environment configured!${NC}"
}

# Start services
start_services() {
    echo -e "${BLUE}Starting services...${NC}"
    
    cd docker
    
    # Pull latest images
    echo -e "${YELLOW}📥 Pulling Docker images...${NC}"
    docker compose -f docker-compose.dev.yml pull
    
    # Start services
    echo -e "${YELLOW}🚀 Starting all services...${NC}"
    docker compose -f docker-compose.dev.yml up -d
    
    # Wait for services to be healthy
    echo -e "${YELLOW}⏳ Waiting for services to start...${NC}"
    sleep 10
    
    # Check service status
    echo -e "${BLUE}📊 Service Status:${NC}"
    docker compose -f docker-compose.dev.yml ps
}

# Show access information
show_access_info() {
    echo ""
    echo -e "${GREEN}🎉 Landscaping SaaS Platform is now running!${NC}"
    echo -e "${GREEN}==========================================${NC}"
    echo ""
    echo -e "${BLUE}📱 Access Points:${NC}"
    echo -e "• Web Dashboard: ${YELLOW}http://localhost:8081${NC}"
    echo -e "• API Endpoints: ${YELLOW}http://localhost:8080${NC}"
    echo -e "• API Documentation: ${YELLOW}http://localhost:8080/docs${NC}"
    echo ""
    echo -e "${BLUE}💾 Database Access:${NC}"
    echo -e "• PostgreSQL: ${YELLOW}localhost:5432${NC}"
    echo -e "• Redis: ${YELLOW}localhost:6379${NC}"
    echo ""
    echo -e "${BLUE}🔧 Useful Commands:${NC}"
    echo -e "• View logs: ${YELLOW}docker compose -f docker/docker-compose.dev.yml logs -f${NC}"
    echo -e "• Stop services: ${YELLOW}docker compose -f docker/docker-compose.dev.yml down${NC}"
    echo -e "• Restart services: ${YELLOW}docker compose -f docker/docker-compose.dev.yml restart${NC}"
    echo ""
    echo -e "${BLUE}📚 Next Steps:${NC}"
    echo "1. Open http://localhost:8081 in your browser"
    echo "2. Create your first tenant/business account"
    echo "3. Explore the admin dashboard"
    echo "4. Check out the mobile app in the /mobile folder"
    echo ""
    echo -e "${GREEN}Happy landscaping! 🌱${NC}"
}

# Main execution
main() {
    check_prerequisites
    setup_environment
    start_services
    show_access_info
}

# Handle script interruption
trap 'echo -e "${RED}Script interrupted. Cleaning up...${NC}"; exit 1' INT

# Run main function
main