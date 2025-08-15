package services

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// Service interfaces - Services struct is defined in services.go

// AuthService handles authentication and authorization
type AuthService interface {
	// Authentication
	Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error)
	Register(ctx context.Context, req *domain.RegisterRequest) (*domain.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.LoginResponse, error)
	Logout(ctx context.Context, sessionID uuid.UUID) error
	
	// Password management
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, token, newPassword string) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) error
	
	// Email verification
	SendEmailVerification(ctx context.Context, userID uuid.UUID) error
	VerifyEmail(ctx context.Context, token string) error
	
	// Two-factor authentication
	EnableTwoFactor(ctx context.Context, userID uuid.UUID) (*TwoFactorSetup, error)
	DisableTwoFactor(ctx context.Context, userID uuid.UUID, password string) error
	VerifyTwoFactor(ctx context.Context, userID uuid.UUID, token string) error
	
	// Session management
	GetUserSessions(ctx context.Context, userID uuid.UUID) ([]*domain.UserSession, error)
	RevokeSession(ctx context.Context, userID, sessionID uuid.UUID) error
	RevokeAllSessions(ctx context.Context, userID uuid.UUID) error
}

// TenantService handles tenant management
type TenantService interface {
	// CRUD operations
	CreateTenant(ctx context.Context, req *domain.CreateTenantRequest) (*domain.EnhancedTenant, error)
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.EnhancedTenant, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, req *domain.UpdateTenantRequest) (*domain.EnhancedTenant, error)
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	ListTenants(ctx context.Context, filter *TenantFilter) (*domain.PaginatedResponse, error)
	
	// Tenant operations
	SuspendTenant(ctx context.Context, tenantID uuid.UUID, reason string) error
	ActivateTenant(ctx context.Context, tenantID uuid.UUID) error
	
	// Settings and configuration
	GetTenantSettings(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error)
	UpdateTenantSettings(ctx context.Context, tenantID uuid.UUID, settings map[string]interface{}) error
	
	// Usage and billing
	GetTenantUsage(ctx context.Context, tenantID uuid.UUID) (*TenantUsage, error)
	GetTenantBilling(ctx context.Context, tenantID uuid.UUID) (*TenantBilling, error)
	
	// White-label features
	GetTenantBranding(ctx context.Context, tenantID uuid.UUID) (*TenantBranding, error)
	UpdateTenantBranding(ctx context.Context, tenantID uuid.UUID, branding *TenantBranding) error
}

// UserService handles user management
type UserService interface {
	// CRUD operations
	CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.EnhancedUser, error)
	GetUser(ctx context.Context, userID uuid.UUID) (*domain.EnhancedUser, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, req *domain.UpdateUserRequest) (*domain.EnhancedUser, error)
	DeleteUser(ctx context.Context, userID uuid.UUID) error
	ListUsers(ctx context.Context, filter *UserFilter) (*domain.PaginatedResponse, error)
	
	// User operations
	ActivateUser(ctx context.Context, userID uuid.UUID) error
	DeactivateUser(ctx context.Context, userID uuid.UUID) error
	ResetUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error
	
	// Permissions
	UpdateUserPermissions(ctx context.Context, userID uuid.UUID, permissions []string) error
	GetUserPermissions(ctx context.Context, userID uuid.UUID) ([]string, error)
	
	// Profile management
	UpdateProfile(ctx context.Context, userID uuid.UUID, profile *UserProfile) error
	UploadAvatar(ctx context.Context, userID uuid.UUID, avatarData []byte, contentType string) (string, error)
}

// CustomerService handles customer management
type CustomerService interface {
	// CRUD operations
	CreateCustomer(ctx context.Context, req *domain.CreateCustomerRequest) (*domain.EnhancedCustomer, error)
	GetCustomer(ctx context.Context, customerID uuid.UUID) (*domain.EnhancedCustomer, error)
	UpdateCustomer(ctx context.Context, customerID uuid.UUID, req *CustomerUpdateRequest) (*domain.EnhancedCustomer, error)
	DeleteCustomer(ctx context.Context, customerID uuid.UUID) error
	ListCustomers(ctx context.Context, filter *CustomerFilter) (*domain.PaginatedResponse, error)
	
	// Search and filtering
	SearchCustomers(ctx context.Context, query string, filter *CustomerFilter) (*domain.PaginatedResponse, error)
	
	// Related data
	GetCustomerProperties(ctx context.Context, customerID uuid.UUID) ([]*domain.EnhancedProperty, error)
	GetCustomerJobs(ctx context.Context, customerID uuid.UUID, filter *JobFilter) (*domain.PaginatedResponse, error)
	GetCustomerInvoices(ctx context.Context, customerID uuid.UUID, filter *InvoiceFilter) (*domain.PaginatedResponse, error)
	GetCustomerQuotes(ctx context.Context, customerID uuid.UUID, filter *QuoteFilter) (*domain.PaginatedResponse, error)
	
	// Customer analytics
	GetCustomerSummary(ctx context.Context, customerID uuid.UUID) (*CustomerSummary, error)
}

// PropertyService handles property management
type PropertyService interface {
	// CRUD operations
	CreateProperty(ctx context.Context, req *domain.CreatePropertyRequest) (*domain.EnhancedProperty, error)
	GetProperty(ctx context.Context, propertyID uuid.UUID) (*domain.EnhancedProperty, error)
	UpdateProperty(ctx context.Context, propertyID uuid.UUID, req *PropertyUpdateRequest) (*domain.EnhancedProperty, error)
	DeleteProperty(ctx context.Context, propertyID uuid.UUID) error
	ListProperties(ctx context.Context, filter *PropertyFilter) (*domain.PaginatedResponse, error)
	
	// Geographic operations
	GetNearbyProperties(ctx context.Context, lat, lng float64, radiusMiles float64) ([]*domain.EnhancedProperty, error)
	SearchProperties(ctx context.Context, query string, filter *PropertyFilter) (*domain.PaginatedResponse, error)
	
	// Related data
	GetPropertyJobs(ctx context.Context, propertyID uuid.UUID, filter *JobFilter) (*domain.PaginatedResponse, error)
	GetPropertyQuotes(ctx context.Context, propertyID uuid.UUID, filter *QuoteFilter) (*domain.PaginatedResponse, error)
	
	// Property analytics
	GetPropertyValue(ctx context.Context, propertyID uuid.UUID) (*PropertyValuation, error)
}

// ServiceService handles service catalog management
type ServiceService interface {
	// CRUD operations
	CreateService(ctx context.Context, req *ServiceCreateRequest) (*domain.Service, error)
	GetService(ctx context.Context, serviceID uuid.UUID) (*domain.Service, error)
	UpdateService(ctx context.Context, serviceID uuid.UUID, req *ServiceUpdateRequest) (*domain.Service, error)
	DeleteService(ctx context.Context, serviceID uuid.UUID) error
	ListServices(ctx context.Context, filter *ServiceFilter) (*domain.PaginatedResponse, error)
	
	// Categories
	GetServiceCategories(ctx context.Context) ([]string, error)
	GetServicesByCategory(ctx context.Context, category string) ([]*domain.Service, error)
	
	// Pricing
	CalculateServicePrice(ctx context.Context, serviceID uuid.UUID, quantity float64, propertyDetails *PropertyDetails) (*ServicePricing, error)
}

// JobService handles job/work order management
type JobService interface {
	// CRUD operations
	CreateJob(ctx context.Context, req *domain.CreateJobRequest) (*domain.EnhancedJob, error)
	GetJob(ctx context.Context, jobID uuid.UUID) (*domain.EnhancedJob, error)
	UpdateJob(ctx context.Context, jobID uuid.UUID, req *domain.UpdateJobRequest) (*domain.EnhancedJob, error)
	DeleteJob(ctx context.Context, jobID uuid.UUID) error
	ListJobs(ctx context.Context, filter *JobFilter) (*domain.PaginatedResponse, error)
	
	// Job lifecycle
	StartJob(ctx context.Context, jobID uuid.UUID, startDetails *JobStartDetails) error
	CompleteJob(ctx context.Context, jobID uuid.UUID, completionDetails *JobCompletionDetails) error
	CancelJob(ctx context.Context, jobID uuid.UUID, reason string) error
	
	// Assignment
	AssignJob(ctx context.Context, jobID, userID uuid.UUID) error
	UnassignJob(ctx context.Context, jobID uuid.UUID) error
	AssignJobToCrew(ctx context.Context, jobID, crewID uuid.UUID) error
	
	// Services
	UpdateJobServices(ctx context.Context, jobID uuid.UUID, services []*JobServiceUpdate) error
	GetJobServices(ctx context.Context, jobID uuid.UUID) ([]*domain.JobService, error)
	
	// Media and documentation
	UploadJobPhotos(ctx context.Context, jobID uuid.UUID, photos []*JobPhoto) error
	AddJobSignature(ctx context.Context, jobID uuid.UUID, signature string) error
	
	// Scheduling
	GetJobSchedule(ctx context.Context, filter *ScheduleFilter) ([]*ScheduledJob, error)
	GetJobCalendar(ctx context.Context, startDate, endDate time.Time) ([]*CalendarEvent, error)
	
	// Recurring jobs
	CreateRecurringJob(ctx context.Context, req *RecurringJobRequest) (*RecurringJobSeries, error)
	
	// Route optimization
	OptimizeJobRoute(ctx context.Context, jobIDs []uuid.UUID, date time.Time) (*RouteOptimization, error)
}

// QuoteService handles quote management
type QuoteService interface {
	// CRUD operations
	CreateQuote(ctx context.Context, req *QuoteCreateRequest) (*domain.Quote, error)
	GetQuote(ctx context.Context, quoteID uuid.UUID) (*domain.Quote, error)
	UpdateQuote(ctx context.Context, quoteID uuid.UUID, req *QuoteUpdateRequest) (*domain.Quote, error)
	DeleteQuote(ctx context.Context, quoteID uuid.UUID) error
	ListQuotes(ctx context.Context, filter *QuoteFilter) (*domain.PaginatedResponse, error)
	
	// Quote lifecycle
	ApproveQuote(ctx context.Context, quoteID uuid.UUID) error
	RejectQuote(ctx context.Context, quoteID uuid.UUID, reason string) error
	ConvertQuoteToJob(ctx context.Context, quoteID uuid.UUID) (*domain.EnhancedJob, error)
	
	// Quote generation
	GenerateQuotePDF(ctx context.Context, quoteID uuid.UUID) ([]byte, error)
	SendQuote(ctx context.Context, quoteID uuid.UUID, sendOptions *QuoteSendOptions) error
	
	// AI-powered features
	GenerateQuoteFromDescription(ctx context.Context, req *QuoteGenerationRequest) (*domain.Quote, error)
}

// InvoiceService handles invoice management
type InvoiceService interface {
	// CRUD operations
	CreateInvoice(ctx context.Context, req *InvoiceCreateRequest) (*domain.Invoice, error)
	GetInvoice(ctx context.Context, invoiceID uuid.UUID) (*domain.Invoice, error)
	UpdateInvoice(ctx context.Context, invoiceID uuid.UUID, req *InvoiceUpdateRequest) (*domain.Invoice, error)
	DeleteInvoice(ctx context.Context, invoiceID uuid.UUID) error
	ListInvoices(ctx context.Context, filter *InvoiceFilter) (*domain.PaginatedResponse, error)
	
	// Invoice operations
	SendInvoice(ctx context.Context, invoiceID uuid.UUID, sendOptions *InvoiceSendOptions) error
	GenerateInvoicePDF(ctx context.Context, invoiceID uuid.UUID) ([]byte, error)
	
	// Payment tracking
	GetInvoicePayments(ctx context.Context, invoiceID uuid.UUID) ([]*domain.Payment, error)
	MarkInvoiceAsPaid(ctx context.Context, invoiceID uuid.UUID, paymentID uuid.UUID) error
	
	// Overdue management
	GetOverdueInvoices(ctx context.Context) ([]*domain.Invoice, error)
	SendOverdueReminders(ctx context.Context) error
	
	// Automation
	CreateInvoiceFromJob(ctx context.Context, jobID uuid.UUID) (*domain.Invoice, error)
}

// PaymentService handles payment processing
type PaymentService interface {
	// Payment processing
	ProcessPayment(ctx context.Context, req *PaymentProcessRequest) (*domain.Payment, error)
	RefundPayment(ctx context.Context, paymentID uuid.UUID, amount float64) (*PaymentRefund, error)
	
	// Payment methods
	GetPaymentMethods(ctx context.Context, customerID uuid.UUID) ([]*PaymentMethod, error)
	AddPaymentMethod(ctx context.Context, customerID uuid.UUID, method *PaymentMethod) error
	DeletePaymentMethod(ctx context.Context, methodID uuid.UUID) error
	
	// Webhooks
	HandleStripeWebhook(ctx context.Context, payload []byte, signature string) error
	
	// Payment analytics
	GetPaymentSummary(ctx context.Context, filter *PaymentFilter) (*PaymentSummary, error)
}

// EquipmentService handles equipment management
type EquipmentService interface {
	// CRUD operations
	CreateEquipment(ctx context.Context, req *EquipmentCreateRequest) (*domain.Equipment, error)
	GetEquipment(ctx context.Context, equipmentID uuid.UUID) (*domain.Equipment, error)
	UpdateEquipment(ctx context.Context, equipmentID uuid.UUID, req *EquipmentUpdateRequest) (*domain.Equipment, error)
	DeleteEquipment(ctx context.Context, equipmentID uuid.UUID) error
	ListEquipment(ctx context.Context, filter *EquipmentFilter) (*domain.PaginatedResponse, error)
	
	// Availability
	GetAvailableEquipment(ctx context.Context, startDate, endDate time.Time) ([]*domain.Equipment, error)
	CheckEquipmentAvailability(ctx context.Context, equipmentID uuid.UUID, startDate, endDate time.Time) (bool, error)
	
	// Maintenance
	ScheduleMaintenance(ctx context.Context, equipmentID uuid.UUID, req *MaintenanceScheduleRequest) error
	GetMaintenanceHistory(ctx context.Context, equipmentID uuid.UUID) ([]*MaintenanceRecord, error)
	GetUpcomingMaintenance(ctx context.Context) ([]*MaintenanceSchedule, error)
}

// CrewService handles crew management
type CrewService interface {
	// CRUD operations
	CreateCrew(ctx context.Context, req *CrewCreateRequest) (*domain.Crew, error)
	GetCrew(ctx context.Context, crewID uuid.UUID) (*domain.Crew, error)
	UpdateCrew(ctx context.Context, crewID uuid.UUID, req *CrewUpdateRequest) (*domain.Crew, error)
	DeleteCrew(ctx context.Context, crewID uuid.UUID) error
	ListCrews(ctx context.Context, filter *CrewFilter) (*domain.PaginatedResponse, error)
	
	// Crew membership
	AddCrewMember(ctx context.Context, crewID, userID uuid.UUID, role string) error
	RemoveCrewMember(ctx context.Context, crewID, userID uuid.UUID) error
	GetCrewMembers(ctx context.Context, crewID uuid.UUID) ([]*CrewMemberDetails, error)
	
	// Availability and scheduling
	GetCrewSchedule(ctx context.Context, crewID uuid.UUID, startDate, endDate time.Time) ([]*CrewScheduleEntry, error)
	GetAvailableCrews(ctx context.Context, startDate, endDate time.Time) ([]*domain.Crew, error)
	
	// Performance tracking
	GetCrewPerformance(ctx context.Context, crewID uuid.UUID, filter *PerformanceFilter) (*CrewPerformance, error)
}

// NotificationService handles notifications
type NotificationService interface {
	// Send notifications
	SendNotification(ctx context.Context, req *NotificationRequest) error
	SendBulkNotification(ctx context.Context, req *BulkNotificationRequest) error
	
	// Get notifications
	GetUserNotifications(ctx context.Context, userID uuid.UUID, filter *NotificationFilter) (*domain.PaginatedResponse, error)
	GetUnreadCount(ctx context.Context, userID uuid.UUID) (int, error)
	
	// Mark as read
	MarkNotificationRead(ctx context.Context, notificationID uuid.UUID) error
	MarkAllNotificationsRead(ctx context.Context, userID uuid.UUID) error
	
	// Templates
	CreateNotificationTemplate(ctx context.Context, template *NotificationTemplate) error
	GetNotificationTemplate(ctx context.Context, templateID string) (*NotificationTemplate, error)
}

// WebhookService handles webhook management
type WebhookService interface {
	// CRUD operations
	CreateWebhook(ctx context.Context, req *WebhookCreateRequest) (*domain.Webhook, error)
	GetWebhook(ctx context.Context, webhookID uuid.UUID) (*domain.Webhook, error)
	UpdateWebhook(ctx context.Context, webhookID uuid.UUID, req *WebhookUpdateRequest) (*domain.Webhook, error)
	DeleteWebhook(ctx context.Context, webhookID uuid.UUID) error
	ListWebhooks(ctx context.Context, filter *WebhookFilter) (*domain.PaginatedResponse, error)
	
	// Webhook operations
	TestWebhook(ctx context.Context, webhookID uuid.UUID) error
	GetWebhookDeliveries(ctx context.Context, webhookID uuid.UUID, filter *DeliveryFilter) (*domain.PaginatedResponse, error)
	
	// Event handling
	TriggerWebhook(ctx context.Context, event string, data interface{}) error
	GetWebhookEvents(ctx context.Context) ([]string, error)
	
	// Retry and failure handling
	RetryFailedDelivery(ctx context.Context, deliveryID uuid.UUID) error
}

// AuditService handles audit logging
type AuditService interface {
	// Audit logging
	LogAction(ctx context.Context, req *AuditLogRequest) error
	
	// Retrieve audit logs
	GetAuditLogs(ctx context.Context, filter *AuditFilter) (*domain.PaginatedResponse, error)
	GetUserActivity(ctx context.Context, userID uuid.UUID, filter *ActivityFilter) (*domain.PaginatedResponse, error)
	
	// Export
	ExportAuditLogs(ctx context.Context, filter *AuditFilter, format string) ([]byte, error)
	
	// Compliance
	GetComplianceReport(ctx context.Context, startDate, endDate time.Time) (*ComplianceReport, error)
}

// ReportService handles reporting and analytics
type ReportService interface {
	// Dashboard
	GetDashboardData(ctx context.Context, timeRange *TimeRange) (*DashboardData, error)
	
	// Financial reports
	GetRevenueReport(ctx context.Context, filter *RevenueFilter) (*RevenueReport, error)
	GetProfitLossReport(ctx context.Context, filter *ProfitLossFilter) (*ProfitLossReport, error)
	
	// Operational reports
	GetJobsReport(ctx context.Context, filter *JobReportFilter) (*JobsReport, error)
	GetCustomersReport(ctx context.Context, filter *CustomerReportFilter) (*CustomersReport, error)
	GetPerformanceReport(ctx context.Context, filter *PerformanceReportFilter) (*PerformanceReport, error)
	
	// Export
	ExportReport(ctx context.Context, reportType string, filter interface{}, format string) ([]byte, error)
}

// StorageService wraps the go-storage package
type StorageService interface {
	Upload(ctx context.Context, path string, data []byte, contentType string) (string, error)
	Download(ctx context.Context, path string) ([]byte, error)
	Delete(ctx context.Context, path string) error
	GetURL(ctx context.Context, path string) (string, error)
	GetSignedURL(ctx context.Context, path string, expiry time.Duration) (string, error)
}

// LLMService wraps the go-llm package
type LLMService interface {
	GenerateText(ctx context.Context, prompt string, options *LLMOptions) (string, error)
	AnalyzeImage(ctx context.Context, imageData []byte, prompt string) (string, error)
	GenerateQuote(ctx context.Context, req *QuoteGenerationRequest) (*QuoteGenerationResponse, error)
	AnalyzeJobDescription(ctx context.Context, description string) (*JobAnalysis, error)
	GenerateJobSummary(ctx context.Context, jobID uuid.UUID) (string, error)
}

// CommunicationService wraps the go-comms package
type CommunicationService interface {
	SendEmail(ctx context.Context, req *EmailRequest) error
	SendSMS(ctx context.Context, req *SMSRequest) error
	SendPushNotification(ctx context.Context, req *PushNotificationRequest) error
	
	// Template-based communications
	SendTemplatedEmail(ctx context.Context, template string, data interface{}, to []string) error
	SendTemplatedSMS(ctx context.Context, template string, data interface{}, to []string) error
}

// ScheduleService handles scheduling and calendar operations
type ScheduleService interface {
	// Schedule optimization
	OptimizeSchedule(ctx context.Context, req *ScheduleOptimizationRequest) (*ScheduleOptimizationResult, error)
	
	// Route optimization
	OptimizeRoute(ctx context.Context, jobs []*domain.EnhancedJob, startLocation *Location) (*RouteOptimization, error)
	
	// Availability checking
	CheckAvailability(ctx context.Context, req *AvailabilityRequest) (*AvailabilityResponse, error)
	
	// Calendar integration
	SyncWithExternalCalendar(ctx context.Context, userID uuid.UUID, calendarType string) error
	GetCalendarEvents(ctx context.Context, userID uuid.UUID, startDate, endDate time.Time) ([]*CalendarEvent, error)
}

// BillingService handles billing and subscription operations  
type BillingService interface {
	// Subscription management
	GetByTenantID(ctx context.Context, tenantID uuid.UUID) (*BillingSubscription, error)
	CreateSubscription(ctx context.Context, req *CreateSubscriptionRequest) (*BillingSubscription, error)
	UpdateSubscription(ctx context.Context, subscriptionID uuid.UUID, req *UpdateSubscriptionRequest) (*BillingSubscription, error)
	CancelSubscription(ctx context.Context, subscriptionID uuid.UUID) error
}

// MetricsService handles metrics collection and analysis
type MetricsService interface {
	// Basic metrics
	RecordMetric(ctx context.Context, name string, value float64, tags map[string]string) error
	GetMetrics(ctx context.Context, filter *MetricsFilter) (*MetricsResponse, error)
}

// MonitoringService handles system monitoring and health checks
type MonitoringService interface {
	// Health checks
	GetServiceHealthStatus(ctx context.Context) ([]ServiceHealth, error)
	GetDatabaseHealth(ctx context.Context) (*DatabaseHealth, error)  
	GetCacheHealth(ctx context.Context) (*CacheHealth, error)
	GetExternalServiceHealth(ctx context.Context) ([]ExternalServiceHealth, error)
	GetPerformanceOverview(ctx context.Context) (*PerformanceOverview, error)
	GetSecurityOverview(ctx context.Context) (*SecurityOverview, error)
	GetSystemUptime(ctx context.Context) (time.Duration, error)
}

// DatabaseService handles database operations
type DatabaseService interface {
	// Connection management
	GetConnection(ctx context.Context) (DatabaseConnection, error)
	CloseConnection(ctx context.Context, conn DatabaseConnection) error
}