# Landscaping App Web Interface

A modern, responsive web application built with Go, HTMX, Templ, and Tailwind CSS for managing landscaping business operations.

## Features

### Core Functionality
- **Authentication System**: Login, registration, password reset with JWT tokens
- **Admin Dashboard**: Real-time KPIs, charts, and business overview
- **Customer Management**: Full CRUD operations with search and filtering
- **Customer Portal**: Self-service dashboard for clients
- **Real-time Updates**: WebSocket-powered notifications and live data
- **AI Assistant**: Integrated chat interface for business support

### Technical Features
- **Server-Side Rendering**: Go templates with HTMX for dynamic content
- **Responsive Design**: Mobile-first approach with Tailwind CSS
- **Progressive Enhancement**: Works without JavaScript, enhanced with it
- **Real-time Communication**: WebSocket integration for live updates
- **API Integration**: Seamless backend API consumption
- **Security**: CSRF protection, secure headers, input validation

## Architecture

### Technology Stack
- **Backend**: Go with Gorilla Mux router
- **Frontend**: HTMX + Templ templates + Tailwind CSS
- **JavaScript**: Alpine.js for client-side interactions
- **WebSockets**: Real-time bidirectional communication
- **Authentication**: JWT tokens with secure cookies

### Directory Structure
```
web/
├── cmd/
│   └── server/
│       └── main.go              # Web server entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Configuration management
│   ├── handlers/
│   │   ├── handlers.go          # Route setup and main handlers
│   │   ├── auth_handlers.go     # Authentication handlers
│   │   ├── dashboard_handlers.go # Dashboard and portal handlers
│   │   ├── customer_handlers.go  # Customer management handlers
│   │   ├── websocket_handlers.go # WebSocket handlers
│   │   └── placeholder_handlers.go # Placeholder handlers
│   ├── middleware/
│   │   └── middleware.go        # HTTP middleware
│   └── services/
│       ├── services.go          # Service initialization
│       ├── auth_service.go      # Authentication service
│       ├── api_service.go       # Backend API client
│       ├── template_service.go  # Template rendering
│       └── websocket_service.go # WebSocket management
├── static/
│   ├── css/
│   │   └── custom.css           # Custom styles
│   ├── js/
│   │   └── app.js              # Application JavaScript
│   └── images/                  # Static images
└── README.md                    # This file
```

## Getting Started

### Prerequisites
- Go 1.23.2 or later
- Backend API server running
- PostgreSQL database
- Redis (optional, for sessions)

### Installation

1. **Clone the repository** (if not already done)
   ```bash
   git clone <repository-url>
   cd landscaping-app
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Set environment variables**
   ```bash
   export WEB_PORT=3000
   export BACKEND_URL=http://localhost:8080
   export SESSION_SECRET=your-session-secret
   ```

4. **Run the web server**
   ```bash
   go run web/cmd/server/main.go
   ```

5. **Access the application**
   - Open http://localhost:3000 in your browser
   - Login with your credentials or register a new account

### Configuration

Environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `WEB_PORT` | `3000` | Port for the web server |
| `BACKEND_URL` | `http://localhost:8080` | Backend API URL |
| `SESSION_SECRET` | Required | Secret for session encryption |
| `STATIC_PATH` | `./web/static` | Path to static assets |
| `TEMPLATE_PATH` | `./web/templates` | Path to templates |
| `CSRF_SECRET` | Required | CSRF protection secret |
| `ENABLE_TLS` | `false` | Enable HTTPS |
| `TLS_CERT_PATH` | Empty | Path to TLS certificate |
| `TLS_KEY_PATH` | Empty | Path to TLS private key |

## Features Implemented

### ✅ Completed Features

1. **Web Server Infrastructure**
   - Go HTTP server with Gorilla Mux
   - Middleware for security, CORS, logging
   - Configuration management
   - Static file serving

2. **Authentication System**
   - Login/logout functionality
   - User registration
   - Password reset flow
   - JWT token management
   - Session handling

3. **Template System**
   - Server-side rendering with Go templates
   - Responsive layouts with Tailwind CSS
   - HTMX integration for dynamic content
   - Alpine.js for client-side interactions

4. **Admin Dashboard**
   - KPI cards with real-time data
   - Navigation sidebar
   - User menu and profile
   - Quick action buttons

5. **Customer Management**
   - Customer list with search and filtering
   - Customer creation form
   - Customer detail views
   - CRUD operations via API

6. **Customer Portal**
   - Customer-specific dashboard
   - Service history view
   - Billing information
   - Quote requests

7. **Real-time Features**
   - WebSocket connection management
   - Live notifications
   - Real-time dashboard updates
   - AI chat interface

8. **WebSocket Integration**
   - Bidirectional communication
   - Real-time notifications
   - Chat message handling
   - Connection management with reconnection

### 🚧 Placeholder Features (Structure Ready)

1. **Property Management**
   - Property CRUD operations
   - Map integration (placeholder)
   - Property-customer associations

2. **Job Management**
   - Job scheduling
   - Calendar view
   - Job status tracking
   - Photo upload

3. **Quote System**
   - Quote creation and management
   - PDF generation
   - Quote approval workflow

4. **Invoice System**
   - Invoice generation
   - Payment tracking
   - Automated billing

5. **Equipment Management**
   - Equipment tracking
   - Maintenance scheduling
   - Asset management

6. **Reporting**
   - Revenue reports
   - Performance analytics
   - Custom report generation

## API Integration

The web application integrates with the backend API through the `APIService`:

### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/forgot-password` - Password reset request
- `POST /api/v1/auth/reset-password` - Password reset
- `GET /api/v1/auth/me` - Get current user

### Customer Management
- `GET /api/v1/customers` - List customers
- `POST /api/v1/customers` - Create customer
- `GET /api/v1/customers/{id}` - Get customer
- `PUT /api/v1/customers/{id}` - Update customer
- `DELETE /api/v1/customers/{id}` - Delete customer

### Real-time Data
- `GET /api/v1/dashboard/stats` - Dashboard statistics
- `GET /api/v1/notifications` - User notifications

## WebSocket Events

### Client to Server
- `auth` - Authentication with JWT token
- `chat_message` - AI chat messages
- `typing` - Typing indicators

### Server to Client
- `welcome` - Connection confirmation
- `notification` - Real-time notifications
- `job_update` - Job status changes
- `dashboard_update` - Dashboard data updates
- `chat_response` - AI chat responses

## Security Features

1. **CSRF Protection**
   - CSRF tokens for form submissions
   - Secure cookie handling

2. **Security Headers**
   - Content Security Policy
   - X-Frame-Options
   - X-Content-Type-Options
   - Referrer-Policy

3. **Authentication**
   - JWT token validation
   - Session management
   - Role-based access control

4. **Input Validation**
   - Form validation
   - Sanitization
   - Error handling

## Development

### Running in Development
```bash
# Set development environment
export ENV=development
export DEBUG_SQL=true

# Run with hot reload (requires air)
air -c .air.toml

# Or run normally
go run web/cmd/server/main.go
```

### Building for Production
```bash
# Build the binary
go build -o bin/web-server web/cmd/server/main.go

# Run in production
export ENV=production
export SESSION_SECRET=secure-production-secret
./bin/web-server
```

### Testing
```bash
# Run tests
go test ./web/...

# Run with coverage
go test -cover ./web/...
```

## Deployment

### Docker Deployment
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o web-server web/cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/web-server .
COPY --from=builder /app/web/static ./web/static
CMD ["./web-server"]
```

### Environment Setup
- Set all required environment variables
- Configure TLS certificates for HTTPS
- Set up reverse proxy (nginx/Apache)
- Configure database connections
- Set up monitoring and logging

## Contributing

1. Follow Go conventions and best practices
2. Write tests for new functionality
3. Update documentation for new features
4. Use semantic commit messages
5. Ensure responsive design principles

## License

This project is part of the Landscaping App suite. See the main project LICENSE file for details.