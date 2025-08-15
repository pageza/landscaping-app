package services

import (
	// TODO: Add imports as needed
	// "context"
	// "fmt"

	"github.com/pageza/landscaping-app/backend/internal/config"
	// TODO: Re-enable when needed
	// "github.com/pageza/landscaping-app/backend/internal/domain"
	// "github.com/pageza/landscaping-app/backend/internal/repository" // Temporarily commented to break import cycle
)

// Services holds all business services
type Services struct {
	Auth         AuthService
	Tenant       TenantService
	User         UserService
	Customer     CustomerService
	Property     PropertyService
	Service      ServiceService
	Job          JobService
	Quote        QuoteService
	Invoice      InvoiceService
	Payment      PaymentService
	Equipment    EquipmentService
	Crew         CrewService
	Notification NotificationService
	Webhook      WebhookService
	Audit        AuditService
	Report       ReportService
	Storage      StorageService
	LLM          LLMService
	Communication CommunicationService
	Schedule     ScheduleService
	// File and Email services not yet defined
}

// NewServices creates a new services instance
// func NewServices(repos *repository.Repositories, config *config.Config) *Services { // Temporarily commented
func NewServices(config *config.Config) *Services { // Simplified version without repositories
	// TODO: Initialize external service clients when integrations are available
	// For now, set to nil to prevent compilation errors
	
	return &Services{
		// Auth:      NewAuthService(repos, config), // Temporarily commented - requires repos
		// User:      NewUserService(repos), // Temporarily commented - requires repos
		// Customer:  NewCustomerService(repos), // Temporarily commented - requires repos
		// Property:  NewPropertyService(repos), // Temporarily commented - requires repos
		// Service:   NewServiceService(repos), // Temporarily commented - requires repos
		// Job:       NewJobService(repos), // Temporarily commented - requires repos
		// Invoice:   NewInvoiceService(repos), // Temporarily commented - requires repos
		// Payment:   NewPaymentService(repos, config), // Temporarily commented - requires repos
		// Equipment: NewEquipmentService(repos), // Temporarily commented - requires repos
		// File:      NewFileService(repos, storageService), // Temporarily commented - requires repos
		// Email:     nil, // TODO: Implement when email service is available
		Storage:   nil, // TODO: Implement when storage service is available
		LLM:       nil, // TODO: Implement when LLM service is available
	}
}
