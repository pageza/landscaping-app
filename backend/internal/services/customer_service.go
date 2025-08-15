package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// CustomerServiceImpl implements the CustomerService interface
type CustomerServiceImpl struct {
	customerRepo CustomerRepository
	propertyRepo PropertyRepository
	jobRepo      JobRepository
	invoiceRepo  InvoiceRepository
	quoteRepo    QuoteRepository
	auditService AuditService
	logger       *log.Logger
}

// CustomerRepository defines the interface for customer data access
type CustomerRepository interface {
	// CRUD operations
	Create(ctx context.Context, customer *domain.EnhancedCustomer) error
	GetByID(ctx context.Context, tenantID, customerID uuid.UUID) (*domain.EnhancedCustomer, error)
	Update(ctx context.Context, customer *domain.EnhancedCustomer) error
	Delete(ctx context.Context, tenantID, customerID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *CustomerFilter) ([]*domain.EnhancedCustomer, int64, error)
	
	// Search operations
	Search(ctx context.Context, tenantID uuid.UUID, query string, filter *CustomerFilter) ([]*domain.EnhancedCustomer, int64, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.EnhancedCustomer, error)
	GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*domain.EnhancedCustomer, error)
	
	// Analytics
	GetCustomerSummary(ctx context.Context, tenantID, customerID uuid.UUID) (*CustomerSummary, error)
	GetCustomerCount(ctx context.Context, tenantID uuid.UUID) (int64, error)
}

// PropertyRepository interface for customer-related property operations
type PropertyRepository interface {
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID) ([]*domain.EnhancedProperty, error)
}

// JobRepository interface for customer-related job operations  
type JobRepository interface {
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
}

// InvoiceRepository interface for customer-related invoice operations
type InvoiceRepository interface {
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *InvoiceFilter) ([]*domain.Invoice, int64, error)
}

// QuoteRepository interface for customer-related quote operations
type QuoteRepository interface {
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *QuoteFilter) ([]*domain.Quote, int64, error)
}

// NewCustomerService creates a new customer service instance
func NewCustomerService(
	customerRepo CustomerRepository,
	propertyRepo PropertyRepository,
	jobRepo JobRepository,
	invoiceRepo InvoiceRepository,
	quoteRepo QuoteRepository,
	auditService AuditService,
	logger *log.Logger,
) CustomerService {
	return &CustomerServiceImpl{
		customerRepo: customerRepo,
		propertyRepo: propertyRepo,
		jobRepo:      jobRepo,
		invoiceRepo:  invoiceRepo,
		quoteRepo:    quoteRepo,
		auditService: auditService,
		logger:       logger,
	}
}

// CreateCustomer creates a new customer
func (s *CustomerServiceImpl) CreateCustomer(ctx context.Context, req *domain.CreateCustomerRequest) (*domain.EnhancedCustomer, error) {
	// Get tenant ID from context
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Validate the request
	if err := s.validateCreateCustomerRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check for duplicate email
	if req.Email != nil && *req.Email != "" {
		existing, err := s.customerRepo.GetByEmail(ctx, tenantID, *req.Email)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("customer with email %s already exists", *req.Email)
		}
	}

	// Check for duplicate phone
	if req.Phone != nil && *req.Phone != "" {
		existing, err := s.customerRepo.GetByPhone(ctx, tenantID, *req.Phone)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("customer with phone %s already exists", *req.Phone)
		}
	}

	// Create customer entity
	customer := &domain.EnhancedCustomer{
		Customer: domain.Customer{
			ID:           uuid.New(),
			TenantID:     tenantID,
			FirstName:    req.FirstName,
			LastName:     req.LastName,
			Email:        req.Email,
			Phone:        req.Phone,
			AddressLine1: req.AddressLine1,
			AddressLine2: req.AddressLine2,
			City:         req.City,
			State:        req.State,
			ZipCode:      req.ZipCode,
			Country:      "US", // Default to US
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		CompanyName:            req.CompanyName,
		PreferredContactMethod: req.PreferredContactMethod,
		LeadSource:             req.LeadSource,
		CustomerType:           req.CustomerType,
		PaymentTerms:           30, // Default 30 days
	}

	// Save to database
	if err := s.customerRepo.Create(ctx, customer); err != nil {
		s.logger.Printf("Failed to create customer", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to create customer: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "customer.create",
		ResourceType: "customer",
		ResourceID:   &customer.ID,
		NewValues: map[string]interface{}{
			"first_name":    customer.FirstName,
			"last_name":     customer.LastName,
			"email":         customer.Email,
			"customer_type": customer.CustomerType,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Customer created successfully", "customer_id", customer.ID, "tenant_id", tenantID)
	return customer, nil
}

// GetCustomer retrieves a customer by ID
func (s *CustomerServiceImpl) GetCustomer(ctx context.Context, customerID uuid.UUID) (*domain.EnhancedCustomer, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		s.logger.Printf("Failed to get customer", "error", err, "customer_id", customerID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	return customer, nil
}

// UpdateCustomer updates an existing customer
func (s *CustomerServiceImpl) UpdateCustomer(ctx context.Context, customerID uuid.UUID, req *CustomerUpdateRequest) (*domain.EnhancedCustomer, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get existing customer
	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	// Store old values for audit
	oldValues := map[string]interface{}{
		"first_name": customer.FirstName,
		"last_name":  customer.LastName,
		"email":      customer.Email,
		"phone":      customer.Phone,
	}

	// Update fields
	if req.FirstName != nil {
		customer.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		customer.LastName = *req.LastName
	}
	if req.Email != nil {
		// Check for duplicate email
		if *req.Email != "" {
			existing, err := s.customerRepo.GetByEmail(ctx, tenantID, *req.Email)
			if err == nil && existing != nil && existing.ID != customerID {
				return nil, fmt.Errorf("customer with email %s already exists", *req.Email)
			}
		}
		customer.Email = req.Email
	}
	if req.Phone != nil {
		// Check for duplicate phone
		if *req.Phone != "" {
			existing, err := s.customerRepo.GetByPhone(ctx, tenantID, *req.Phone)
			if err == nil && existing != nil && existing.ID != customerID {
				return nil, fmt.Errorf("customer with phone %s already exists", *req.Phone)
			}
		}
		customer.Phone = req.Phone
	}
	if req.CompanyName != nil {
		customer.CompanyName = req.CompanyName
	}
	if req.AddressLine1 != nil {
		customer.AddressLine1 = req.AddressLine1
	}
	if req.AddressLine2 != nil {
		customer.AddressLine2 = req.AddressLine2
	}
	if req.City != nil {
		customer.City = req.City
	}
	if req.State != nil {
		customer.State = req.State
	}
	if req.ZipCode != nil {
		customer.ZipCode = req.ZipCode
	}
	if req.PreferredContactMethod != nil {
		customer.PreferredContactMethod = *req.PreferredContactMethod
	}
	if req.CustomerType != nil {
		customer.CustomerType = *req.CustomerType
	}
	if req.CreditLimit != nil {
		customer.CreditLimit = req.CreditLimit
	}
	if req.PaymentTerms != nil {
		customer.PaymentTerms = *req.PaymentTerms
	}
	if req.Notes != nil {
		customer.Notes = req.Notes
	}

	customer.UpdatedAt = time.Now()

	// Validate the updated customer
	if err := s.validateCustomer(customer); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to database
	if err := s.customerRepo.Update(ctx, customer); err != nil {
		s.logger.Printf("Failed to update customer", "error", err, "customer_id", customerID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update customer: %w", err)
	}

	// Log audit event
	newValues := map[string]interface{}{
		"first_name": customer.FirstName,
		"last_name":  customer.LastName,
		"email":      customer.Email,
		"phone":      customer.Phone,
	}

	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "customer.update",
		ResourceType: "customer",
		ResourceID:   &customer.ID,
		OldValues:    oldValues,
		NewValues:    newValues,
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Customer updated successfully", "customer_id", customerID, "tenant_id", tenantID)
	return customer, nil
}

// DeleteCustomer deletes a customer
func (s *CustomerServiceImpl) DeleteCustomer(ctx context.Context, customerID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get customer before deletion for audit log
	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		return fmt.Errorf("customer not found")
	}

	// Check if customer has active jobs
	jobs, _, err := s.jobRepo.GetByCustomerID(ctx, tenantID, customerID, &JobFilter{
		Status: "pending,scheduled,in_progress",
	})
	if err != nil {
		s.logger.Printf("Failed to check for active jobs", "error", err)
	} else if len(jobs) > 0 {
		return fmt.Errorf("cannot delete customer with active jobs")
	}

	// Soft delete by updating status
	customer.Status = "deleted"
	customer.UpdatedAt = time.Now()

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		s.logger.Printf("Failed to delete customer", "error", err, "customer_id", customerID, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "customer.delete",
		ResourceType: "customer",
		ResourceID:   &customer.ID,
		OldValues: map[string]interface{}{
			"status": "active",
		},
		NewValues: map[string]interface{}{
			"status": "deleted",
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Customer deleted successfully", "customer_id", customerID, "tenant_id", tenantID)
	return nil
}

// ListCustomers lists customers with filtering and pagination
func (s *CustomerServiceImpl) ListCustomers(ctx context.Context, filter *CustomerFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Set defaults
	if filter == nil {
		filter = &CustomerFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	customers, total, err := s.customerRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Printf("Failed to list customers", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       customers,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// SearchCustomers searches customers by query string
func (s *CustomerServiceImpl) SearchCustomers(ctx context.Context, query string, filter *CustomerFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Set defaults
	if filter == nil {
		filter = &CustomerFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	customers, total, err := s.customerRepo.Search(ctx, tenantID, query, filter)
	if err != nil {
		s.logger.Printf("Failed to search customers", "error", err, "query", query, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to search customers: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       customers,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetCustomerProperties retrieves properties for a customer
func (s *CustomerServiceImpl) GetCustomerProperties(ctx context.Context, customerID uuid.UUID) ([]*domain.EnhancedProperty, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	properties, err := s.propertyRepo.GetByCustomerID(ctx, tenantID, customerID)
	if err != nil {
		s.logger.Printf("Failed to get customer properties", "error", err, "customer_id", customerID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get customer properties: %w", err)
	}

	return properties, nil
}

// GetCustomerJobs retrieves jobs for a customer
func (s *CustomerServiceImpl) GetCustomerJobs(ctx context.Context, customerID uuid.UUID, filter *JobFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	// Set defaults
	if filter == nil {
		filter = &JobFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}

	jobs, total, err := s.jobRepo.GetByCustomerID(ctx, tenantID, customerID, filter)
	if err != nil {
		s.logger.Printf("Failed to get customer jobs", "error", err, "customer_id", customerID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get customer jobs: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       jobs,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetCustomerInvoices retrieves invoices for a customer
func (s *CustomerServiceImpl) GetCustomerInvoices(ctx context.Context, customerID uuid.UUID, filter *InvoiceFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	// Set defaults
	if filter == nil {
		filter = &InvoiceFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}

	invoices, total, err := s.invoiceRepo.GetByCustomerID(ctx, tenantID, customerID, filter)
	if err != nil {
		s.logger.Printf("Failed to get customer invoices", "error", err, "customer_id", customerID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get customer invoices: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       invoices,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetCustomerQuotes retrieves quotes for a customer
func (s *CustomerServiceImpl) GetCustomerQuotes(ctx context.Context, customerID uuid.UUID, filter *QuoteFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	// Set defaults
	if filter == nil {
		filter = &QuoteFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}

	quotes, total, err := s.quoteRepo.GetByCustomerID(ctx, tenantID, customerID, filter)
	if err != nil {
		s.logger.Printf("Failed to get customer quotes", "error", err, "customer_id", customerID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get customer quotes: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       quotes,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetCustomerSummary retrieves customer analytics and summary
func (s *CustomerServiceImpl) GetCustomerSummary(ctx context.Context, customerID uuid.UUID) (*CustomerSummary, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	summary, err := s.customerRepo.GetCustomerSummary(ctx, tenantID, customerID)
	if err != nil {
		s.logger.Printf("Failed to get customer summary", "error", err, "customer_id", customerID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get customer summary: %w", err)
	}

	return summary, nil
}

// Helper methods

func (s *CustomerServiceImpl) validateCreateCustomerRequest(req *domain.CreateCustomerRequest) error {
	if strings.TrimSpace(req.FirstName) == "" {
		return fmt.Errorf("first name is required")
	}
	if strings.TrimSpace(req.LastName) == "" {
		return fmt.Errorf("last name is required")
	}
	if req.Email != nil && *req.Email != "" {
		if !isValidEmail(*req.Email) {
			return fmt.Errorf("invalid email format")
		}
	}
	if req.PreferredContactMethod != "email" && req.PreferredContactMethod != "phone" {
		return fmt.Errorf("preferred contact method must be 'email' or 'phone'")
	}
	if req.CustomerType != "residential" && req.CustomerType != "commercial" {
		return fmt.Errorf("customer type must be 'residential' or 'commercial'")
	}

	// If preferred contact is email, email is required
	if req.PreferredContactMethod == "email" && (req.Email == nil || *req.Email == "") {
		return fmt.Errorf("email is required when preferred contact method is email")
	}

	// If preferred contact is phone, phone is required
	if req.PreferredContactMethod == "phone" && (req.Phone == nil || *req.Phone == "") {
		return fmt.Errorf("phone is required when preferred contact method is phone")
	}

	return nil
}

func (s *CustomerServiceImpl) validateCustomer(customer *domain.EnhancedCustomer) error {
	if strings.TrimSpace(customer.FirstName) == "" {
		return fmt.Errorf("first name is required")
	}
	if strings.TrimSpace(customer.LastName) == "" {
		return fmt.Errorf("last name is required")
	}
	if customer.Email != nil && *customer.Email != "" {
		if !isValidEmail(*customer.Email) {
			return fmt.Errorf("invalid email format")
		}
	}
	return nil
}

func isValidEmail(email string) bool {
	// Basic email validation
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// Context helpers - using functions from billing_service.go to avoid duplication