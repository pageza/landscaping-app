package services

import (
	"github.com/pageza/landscaping-app/web/internal/config"
)

// Services holds all service dependencies for the web application
type Services struct {
	Auth     *AuthService
	API      *APIService
	Template *TemplateService
	WebSocket *WebSocketService
}

// NewServices initializes all services
func NewServices(cfg *config.Config) (*Services, error) {
	// Initialize API service first as others depend on it
	apiSvc := NewAPIService(cfg)

	// Initialize authentication service
	authSvc := NewAuthService(cfg, apiSvc)

	// Initialize template service
	templateSvc, err := NewTemplateService(cfg)
	if err != nil {
		return nil, err
	}

	// Initialize WebSocket service
	wsSvc := NewWebSocketService(cfg)

	return &Services{
		Auth:      authSvc,
		API:       apiSvc,
		Template:  templateSvc,
		WebSocket: wsSvc,
	}, nil
}