# Landscaping SaaS Application

A comprehensive multi-tenant SaaS application for landscaping businesses, built with Go, HTMX/Templ, and Flutter.

## Features

- **Multi-tenant Architecture**: Secure tenant isolation with configurable isolation levels
- **Customer Management**: Complete customer and property management system
- **Job Scheduling**: Advanced job scheduling and tracking with real-time updates
- **Invoicing & Payments**: Integrated billing and payment processing with Stripe
- **Equipment Management**: Track and maintain equipment and assets
- **Mobile App**: Flutter mobile application for field workers
- **File Management**: Secure file upload and storage with S3-compatible backends
- **AI Integration**: LLM-powered job descriptions and content generation
- **Email Communications**: Automated email notifications and templates
- **Real-time Updates**: WebSocket support for live updates
- **Background Jobs**: Reliable background job processing with Redis
- **Comprehensive API**: RESTful API with OpenAPI documentation

## Architecture

### Backend (Go)
- **API Server**: REST API with Gorilla Mux router
- **Worker Service**: Background job processor
- **Database**: PostgreSQL with migration support
- **Cache**: Redis for caching and job queues
- **Storage**: S3-compatible object storage
- **Authentication**: JWT-based authentication with refresh tokens

### Frontend
- **Web App**: HTMX + Templ for server-side rendering
- **Mobile App**: Flutter for iOS and Android

### Infrastructure
- **Docker**: Containerized deployment
- **GitHub Actions**: CI/CD pipeline
- **Database Migrations**: Versioned schema management
- **Monitoring**: Prometheus metrics and health checks

## Project Structure

```
landscaping-app/
├── backend/                 # Go backend services
│   ├── cmd/                # Application entry points
│   │   ├── api/           # REST API server
│   │   ├── worker/        # Background job processor
│   │   └── migrate/       # Database migration tool
│   ├── internal/          # Internal packages
│   │   ├── config/        # Configuration management
│   │   ├── domain/        # Business entities and models
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── middleware/    # Authentication, logging, CORS
│   │   ├── repository/    # Database access layer
│   │   ├── services/      # Business logic layer
│   │   └── tenant/        # Multi-tenancy management
│   ├── migrations/        # SQL migration files
│   ├── pkg/              # Public packages
│   └── tests/            # Integration and unit tests
├── web/                   # HTMX + Templ frontend
├── mobile/               # Flutter mobile application
├── docker/               # Docker configuration files
├── .github/workflows/    # GitHub Actions CI/CD
├── scripts/              # Build and deployment scripts
└── docs/                # API and developer documentation
```

## Quick Start

### Prerequisites

- Go 1.18+
- Docker and Docker Compose
- PostgreSQL 15+
- Redis 7+
- Node.js 18+ (for web frontend)
- Flutter SDK (for mobile app)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/pageza/landscaping-app.git
   cd landscaping-app
   ```

2. **Setup environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Install development tools**
   ```bash
   make install-tools
   ```

4. **Start development services**
   ```bash
   make docker-up
   ```

5. **Run database migrations**
   ```bash
   make migrate-up
   ```

6. **Start the API server**
   ```bash
   make dev
   ```

7. **Start the worker (in another terminal)**
   ```bash
   make dev-worker
   ```

The API will be available at `http://localhost:8080`

### Docker Development

```bash
# Start all services
make docker-up

# View logs
make docker-logs

# Stop all services
make docker-down
```

## Configuration

Configuration is managed through environment variables. See `.env.example` for all available options.

### Key Configuration Areas

- **Database**: PostgreSQL connection settings
- **Redis**: Cache and job queue configuration  
- **JWT**: Authentication token settings
- **Email**: SMTP server configuration
- **Storage**: S3-compatible storage settings
- **Payments**: Stripe API keys
- **LLM**: OpenAI/Anthropic API keys
- **Multi-tenancy**: Tenant isolation level

## API Documentation

The API follows RESTful conventions with the following endpoints:

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/refresh` - Refresh token
- `POST /api/v1/auth/forgot-password` - Password reset request
- `POST /api/v1/auth/reset-password` - Password reset

### Resources
- `/api/v1/users` - User management
- `/api/v1/customers` - Customer management
- `/api/v1/properties` - Property management
- `/api/v1/services` - Service catalog
- `/api/v1/jobs` - Job scheduling and tracking
- `/api/v1/invoices` - Invoice management
- `/api/v1/payments` - Payment processing
- `/api/v1/equipment` - Equipment tracking
- `/api/v1/files` - File upload and management

## Database Schema

The application uses PostgreSQL with the following key entities:

- **Tenants**: Multi-tenant isolation
- **Users**: System users with role-based access
- **Customers**: Business customers
- **Properties**: Customer properties
- **Services**: Service catalog
- **Jobs**: Work orders and scheduling
- **Invoices**: Billing and invoicing
- **Payments**: Payment tracking
- **Equipment**: Asset management
- **File Attachments**: File storage metadata

## Multi-Tenancy

The application supports three levels of tenant isolation:

1. **Database Isolation**: Each tenant has a separate database
2. **Schema Isolation**: Tenants share a database but have separate schemas
3. **Row Isolation**: All tenants share database/schema, isolated by tenant_id

Configure via `TENANT_ISOLATION_LEVEL` environment variable.

## Background Jobs

Background jobs are processed using a Redis-based queue system:

- Email notifications
- Invoice generation
- Payment processing
- File processing
- Data exports
- Scheduled maintenance

## Testing

```bash
# Run unit tests
make test

# Run integration tests
make test-integration

# Generate coverage report
make coverage

# Run linter
make lint
```

## Deployment

### Production Build

```bash
# Build production binaries
make build-prod

# Build Docker images
make docker-build
```

### GitHub Actions

The project includes CI/CD workflows for:

- **Test**: Run tests on pull requests
- **Build**: Build and push Docker images
- **Deploy**: Deploy to production environments

### Environment Variables

Set the following secrets in your GitHub repository:

- `AWS_ACCESS_KEY_ID`
- `AWS_SECRET_ACCESS_KEY`
- `STRIPE_SECRET_KEY`
- `OPENAI_API_KEY`
- `SENTRY_DSN`

## Monitoring

### Health Checks

- `GET /health` - API health status
- Prometheus metrics on `:9090/metrics`
- Database connection monitoring
- Redis connection monitoring

### Logging

Structured JSON logging with configurable levels:
- `LOG_LEVEL`: debug, info, warn, error
- `LOG_FORMAT`: json, text

## Security

- JWT-based authentication
- Rate limiting per IP/user
- CORS protection
- SQL injection protection
- File upload validation
- Environment-based secrets
- Secure headers middleware

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Run the linter
6. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support and questions:
- Create an issue on GitHub
- Email: zap.freelance@gmail.com

## Roadmap

- [ ] Advanced reporting and analytics
- [ ] Mobile app push notifications
- [ ] GPS tracking for field workers
- [ ] Integration with accounting software
- [ ] Advanced scheduling algorithms
- [ ] Customer portal
- [ ] White-label support
- [ ] API rate limiting per tenant
- [ ] Advanced role-based permissions
- [ ] Audit logging
- [ ] Data export/import tools
- [ ] Advanced equipment maintenance scheduling