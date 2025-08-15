package services

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// QuoteServiceImpl implements the QuoteService interface
type QuoteServiceImpl struct {
	quoteRepo           QuoteRepositoryFull
	customerRepo        CustomerRepository
	propertyRepo        PropertyRepositoryExtended
	serviceRepo         ServiceRepository
	auditService        AuditService
	communicationService CommunicationService
	llmService          LLMService
	storageService      StorageService
	logger              *log.Logger
}

// QuoteRepositoryFull defines the complete interface for quote data access
type QuoteRepositoryFull interface {
	// CRUD operations
	Create(ctx context.Context, quote *domain.Quote) error
	GetByID(ctx context.Context, tenantID, quoteID uuid.UUID) (*domain.Quote, error)
	Update(ctx context.Context, quote *domain.Quote) error
	Delete(ctx context.Context, tenantID, quoteID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *QuoteFilter) ([]*domain.Quote, int64, error)
	
	// Filtering operations
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *QuoteFilter) ([]*domain.Quote, int64, error)
	GetByPropertyID(ctx context.Context, tenantID, propertyID uuid.UUID, filter *QuoteFilter) ([]*domain.Quote, int64, error)
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Quote, error)
	
	// Quote services
	CreateQuoteService(ctx context.Context, quoteService *domain.QuoteService) error
	UpdateQuoteService(ctx context.Context, quoteService *domain.QuoteService) error
	DeleteQuoteService(ctx context.Context, quoteServiceID uuid.UUID) error
	GetQuoteServices(ctx context.Context, quoteID uuid.UUID) ([]*domain.QuoteService, error)
	
	// Quote numbering
	GetNextQuoteNumber(ctx context.Context, tenantID uuid.UUID) (string, error)
}

// NewQuoteService creates a new quote service instance
func NewQuoteService(
	quoteRepo QuoteRepositoryFull,
	customerRepo CustomerRepository,
	propertyRepo PropertyRepositoryExtended,
	serviceRepo ServiceRepository,
	auditService AuditService,
	communicationService CommunicationService,
	llmService LLMService,
	storageService StorageService,
	logger *log.Logger,
) QuoteService {
	return &QuoteServiceImpl{
		quoteRepo:            quoteRepo,
		customerRepo:         customerRepo,
		propertyRepo:         propertyRepo,
		serviceRepo:          serviceRepo,
		auditService:         auditService,
		communicationService: communicationService,
		llmService:           llmService,
		storageService:       storageService,
		logger:               logger,
	}
}

// CreateQuote creates a new quote
func (s *QuoteServiceImpl) CreateQuote(ctx context.Context, req *QuoteCreateRequest) (*domain.Quote, error) {
	// Get tenant ID from context
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Validate the request
	if err := s.validateCreateQuoteRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, tenantID, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	// Verify property exists and belongs to customer
	property, err := s.propertyRepo.GetByID(ctx, tenantID, req.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify property: %w", err)
	}
	if property == nil {
		return nil, fmt.Errorf("property not found")
	}
	if property.CustomerID != req.CustomerID {
		return nil, fmt.Errorf("property does not belong to the specified customer")
	}

	// Verify services exist
	if len(req.Services) == 0 {
		return nil, fmt.Errorf("at least one service is required")
	}

	serviceIDs := make([]uuid.UUID, len(req.Services))
	for i, svc := range req.Services {
		serviceIDs[i] = svc.ServiceID
	}

	services, err := s.serviceRepo.GetByIDs(ctx, tenantID, serviceIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to verify services: %w", err)
	}
	if len(services) != len(serviceIDs) {
		return nil, fmt.Errorf("one or more services not found")
	}

	// Generate quote number
	quoteNumber, err := s.quoteRepo.GetNextQuoteNumber(ctx, tenantID)
	if err != nil {
		s.logger.Printf("Failed to generate quote number", "error", err)
		quoteNumber = fmt.Sprintf("Q-%d", time.Now().Unix())
	}

	// Calculate totals
	subtotal := 0.0
	for _, svc := range req.Services {
		subtotal += svc.Quantity * svc.UnitPrice
	}

	taxRate := 0.08 // Default 8% tax rate - this could be configurable
	taxAmount := subtotal * taxRate
	totalAmount := subtotal + taxAmount

	// Set valid until date (30 days from now if not specified)
	validUntil := req.ValidUntil
	if validUntil == nil {
		futureDate := time.Now().AddDate(0, 0, 30)
		validUntil = &futureDate
	}

	// Create quote entity
	quote := &domain.Quote{
		ID:                 uuid.New(),
		TenantID:           tenantID,
		CustomerID:         req.CustomerID,
		PropertyID:         req.PropertyID,
		QuoteNumber:        quoteNumber,
		Title:              req.Title,
		Description:        req.Description,
		Subtotal:           subtotal,
		TaxRate:            taxRate,
		TaxAmount:          taxAmount,
		TotalAmount:        totalAmount,
		Status:             "draft",
		ValidUntil:         validUntil,
		TermsAndConditions: req.TermsAndConditions,
		Notes:              req.Notes,
		CreatedBy:          GetUserIDFromContext(ctx),
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	// Save quote to database
	if err := s.quoteRepo.Create(ctx, quote); err != nil {
		s.logger.Printf("Failed to create quote", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to create quote: %w", err)
	}

	// Create quote services
	for _, svcReq := range req.Services {
		quoteService := &domain.QuoteService{
			ID:          uuid.New(),
			QuoteID:     quote.ID,
			ServiceID:   svcReq.ServiceID,
			Quantity:    svcReq.Quantity,
			UnitPrice:   svcReq.UnitPrice,
			TotalPrice:  svcReq.Quantity * svcReq.UnitPrice,
			Description: svcReq.Description,
			CreatedAt:   time.Now(),
		}

		if err := s.quoteRepo.CreateQuoteService(ctx, quoteService); err != nil {
			s.logger.Printf("Failed to create quote service", "error", err, "quote_id", quote.ID, "service_id", svcReq.ServiceID)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "quote.create",
		ResourceType: "quote",
		ResourceID:   &quote.ID,
		NewValues: map[string]interface{}{
			"quote_number": quote.QuoteNumber,
			"customer_id":  quote.CustomerID,
			"property_id":  quote.PropertyID,
			"total_amount": quote.TotalAmount,
			"status":       quote.Status,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Quote created successfully", "quote_id", quote.ID, "quote_number", quoteNumber, "tenant_id", tenantID)
	return quote, nil
}

// GetQuote retrieves a quote by ID
func (s *QuoteServiceImpl) GetQuote(ctx context.Context, quoteID uuid.UUID) (*domain.Quote, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		s.logger.Printf("Failed to get quote", "error", err, "quote_id", quoteID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	if quote == nil {
		return nil, fmt.Errorf("quote not found")
	}

	return quote, nil
}

// UpdateQuote updates an existing quote
func (s *QuoteServiceImpl) UpdateQuote(ctx context.Context, quoteID uuid.UUID, req *QuoteUpdateRequest) (*domain.Quote, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get existing quote
	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	if quote == nil {
		return nil, fmt.Errorf("quote not found")
	}

	// Check if quote can be updated (only draft and pending quotes)
	if quote.Status != "draft" && quote.Status != "pending" {
		return nil, fmt.Errorf("cannot update quote with status: %s", quote.Status)
	}

	// Store old values for audit
	oldValues := map[string]interface{}{
		"title":        quote.Title,
		"total_amount": quote.TotalAmount,
		"valid_until":  quote.ValidUntil,
	}

	// Update fields
	if req.Title != nil {
		quote.Title = *req.Title
	}
	if req.Description != nil {
		quote.Description = req.Description
	}
	if req.ValidUntil != nil {
		quote.ValidUntil = req.ValidUntil
	}
	if req.TermsAndConditions != nil {
		quote.TermsAndConditions = req.TermsAndConditions
	}
	if req.Notes != nil {
		quote.Notes = req.Notes
	}

	// Update services if provided
	if len(req.Services) > 0 {
		// Verify services exist
		serviceIDs := make([]uuid.UUID, len(req.Services))
		for i, svc := range req.Services {
			serviceIDs[i] = svc.ServiceID
		}

		services, err := s.serviceRepo.GetByIDs(ctx, tenantID, serviceIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to verify services: %w", err)
		}
		if len(services) != len(serviceIDs) {
			return nil, fmt.Errorf("one or more services not found")
		}

		// Delete existing quote services
		existingServices, err := s.quoteRepo.GetQuoteServices(ctx, quoteID)
		if err != nil {
			s.logger.Printf("Failed to get existing quote services", "error", err)
		} else {
			for _, svc := range existingServices {
				if err := s.quoteRepo.DeleteQuoteService(ctx, svc.ID); err != nil {
					s.logger.Printf("Failed to delete quote service", "error", err, "service_id", svc.ID)
				}
			}
		}

		// Create new quote services and recalculate totals
		subtotal := 0.0
		for _, svcReq := range req.Services {
			totalPrice := svcReq.Quantity * svcReq.UnitPrice
			subtotal += totalPrice

			quoteService := &domain.QuoteService{
				ID:          uuid.New(),
				QuoteID:     quote.ID,
				ServiceID:   svcReq.ServiceID,
				Quantity:    svcReq.Quantity,
				UnitPrice:   svcReq.UnitPrice,
				TotalPrice:  totalPrice,
				Description: svcReq.Description,
				CreatedAt:   time.Now(),
			}

			if err := s.quoteRepo.CreateQuoteService(ctx, quoteService); err != nil {
				s.logger.Printf("Failed to create updated quote service", "error", err, "quote_id", quote.ID, "service_id", svcReq.ServiceID)
			}
		}

		// Update quote totals
		quote.Subtotal = subtotal
		quote.TaxAmount = subtotal * quote.TaxRate
		quote.TotalAmount = quote.Subtotal + quote.TaxAmount
	}

	quote.UpdatedAt = time.Now()

	// Save to database
	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		s.logger.Printf("Failed to update quote", "error", err, "quote_id", quoteID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update quote: %w", err)
	}

	// Log audit event
	newValues := map[string]interface{}{
		"title":        quote.Title,
		"total_amount": quote.TotalAmount,
		"valid_until":  quote.ValidUntil,
	}

	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "quote.update",
		ResourceType: "quote",
		ResourceID:   &quote.ID,
		OldValues:    oldValues,
		NewValues:    newValues,
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Quote updated successfully", "quote_id", quoteID, "tenant_id", tenantID)
	return quote, nil
}

// DeleteQuote deletes a quote (soft delete)
func (s *QuoteServiceImpl) DeleteQuote(ctx context.Context, quoteID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get quote before deletion for audit log
	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}
	if quote == nil {
		return fmt.Errorf("quote not found")
	}

	// Check if quote can be deleted (only draft quotes can be deleted)
	if quote.Status != "draft" {
		return fmt.Errorf("only draft quotes can be deleted")
	}

	// Soft delete by updating status
	quote.Status = "cancelled"
	quote.UpdatedAt = time.Now()

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		s.logger.Printf("Failed to delete quote", "error", err, "quote_id", quoteID, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete quote: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "quote.delete",
		ResourceType: "quote",
		ResourceID:   &quote.ID,
		OldValues: map[string]interface{}{
			"status": "draft",
		},
		NewValues: map[string]interface{}{
			"status": "cancelled",
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Quote deleted successfully", "quote_id", quoteID, "tenant_id", tenantID)
	return nil
}

// ListQuotes lists quotes with filtering and pagination
func (s *QuoteServiceImpl) ListQuotes(ctx context.Context, filter *QuoteFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
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
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	quotes, total, err := s.quoteRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Printf("Failed to list quotes", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to list quotes: %w", err)
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

// ApproveQuote approves a quote
func (s *QuoteServiceImpl) ApproveQuote(ctx context.Context, quoteID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get quote
	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}
	if quote == nil {
		return fmt.Errorf("quote not found")
	}

	// Check current status
	if quote.Status != "pending" && quote.Status != "draft" {
		return fmt.Errorf("quote must be in pending or draft status to approve")
	}

	// Update quote status
	quote.Status = "approved"
	quote.ApprovedAt = timePtr(time.Now())
	quote.ApprovedBy = GetUserIDFromContext(ctx)
	quote.UpdatedAt = time.Now()

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		s.logger.Printf("Failed to approve quote", "error", err, "quote_id", quoteID, "tenant_id", tenantID)
		return fmt.Errorf("failed to approve quote: %w", err)
	}

	// Send notification to customer if they have an email
	customer, err := s.customerRepo.GetByID(ctx, tenantID, quote.CustomerID)
	if err == nil && customer != nil && customer.Email != nil && *customer.Email != "" {
		if err := s.communicationService.SendEmail(ctx, &EmailRequest{
			To:      []string{*customer.Email},
			Subject: fmt.Sprintf("Quote %s Approved", quote.QuoteNumber),
			Body:    fmt.Sprintf("Your quote %s has been approved. Total amount: $%.2f", quote.QuoteNumber, quote.TotalAmount),
			IsHTML:  false,
		}); err != nil {
			s.logger.Printf("Failed to send quote approval email", "error", err, "quote_id", quoteID)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "quote.approve",
		ResourceType: "quote",
		ResourceID:   &quote.ID,
		NewValues: map[string]interface{}{
			"status":      "approved",
			"approved_at": quote.ApprovedAt,
			"approved_by": quote.ApprovedBy,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Quote approved successfully", "quote_id", quoteID, "tenant_id", tenantID)
	return nil
}

// RejectQuote rejects a quote
func (s *QuoteServiceImpl) RejectQuote(ctx context.Context, quoteID uuid.UUID, reason string) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get quote
	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}
	if quote == nil {
		return fmt.Errorf("quote not found")
	}

	// Check current status
	if quote.Status != "pending" && quote.Status != "draft" {
		return fmt.Errorf("quote must be in pending or draft status to reject")
	}

	// Update quote status
	quote.Status = "rejected"
	quote.UpdatedAt = time.Now()

	// Add rejection reason to notes
	rejectionNote := fmt.Sprintf("Quote rejected: %s", reason)
	if quote.Notes != nil {
		quote.Notes = stringPtr(*quote.Notes + "\n\n" + rejectionNote)
	} else {
		quote.Notes = &rejectionNote
	}

	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		s.logger.Printf("Failed to reject quote", "error", err, "quote_id", quoteID, "tenant_id", tenantID)
		return fmt.Errorf("failed to reject quote: %w", err)
	}

	// Send notification to customer if they have an email
	customer, err := s.customerRepo.GetByID(ctx, tenantID, quote.CustomerID)
	if err == nil && customer != nil && customer.Email != nil && *customer.Email != "" {
		if err := s.communicationService.SendEmail(ctx, &EmailRequest{
			To:      []string{*customer.Email},
			Subject: fmt.Sprintf("Quote %s Status Update", quote.QuoteNumber),
			Body:    fmt.Sprintf("Your quote %s has been updated. Reason: %s", quote.QuoteNumber, reason),
			IsHTML:  false,
		}); err != nil {
			s.logger.Printf("Failed to send quote rejection email", "error", err, "quote_id", quoteID)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "quote.reject",
		ResourceType: "quote",
		ResourceID:   &quote.ID,
		NewValues: map[string]interface{}{
			"status": "rejected",
			"reason": reason,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Quote rejected successfully", "quote_id", quoteID, "reason", reason, "tenant_id", tenantID)
	return nil
}

// ConvertQuoteToJob converts an approved quote to a job
func (s *QuoteServiceImpl) ConvertQuoteToJob(ctx context.Context, quoteID uuid.UUID) (*domain.EnhancedJob, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get quote
	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	if quote == nil {
		return nil, fmt.Errorf("quote not found")
	}

	// Check quote status
	if quote.Status != "approved" {
		return nil, fmt.Errorf("only approved quotes can be converted to jobs")
	}

	// Get quote services
	quoteServices, err := s.quoteRepo.GetQuoteServices(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote services: %w", err)
	}

	// TODO: Use quote services to create corresponding job services
	s.logger.Printf("Retrieved quote services for job creation - count: %d", len(quoteServices))

	// Create job from quote
	job := &domain.EnhancedJob{
		Job: domain.Job{
			ID:          uuid.New(),
			TenantID:    tenantID,
			CustomerID:  quote.CustomerID,
			PropertyID:  quote.PropertyID,
			Title:       quote.Title,
			Description: quote.Description,
			Status:      domain.JobStatusPending,
			Priority:    "medium", // Default priority
			TotalAmount: &quote.TotalAmount,
			Notes:       stringPtr(fmt.Sprintf("Created from quote %s", quote.QuoteNumber)),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		CrewSize: 1, // Default crew size
	}

	// This would typically use a job service to create the job
	// For now, we'll return a mock job
	s.logger.Printf("Quote converted to job successfully", "quote_id", quoteID, "job_id", job.ID)

	// Update quote status
	quote.Status = "converted"
	quote.UpdatedAt = time.Now()
	if err := s.quoteRepo.Update(ctx, quote); err != nil {
		s.logger.Printf("Failed to update quote status after conversion", "error", err)
	}

	return job, nil
}

// GenerateQuotePDF generates a PDF for the quote
func (s *QuoteServiceImpl) GenerateQuotePDF(ctx context.Context, quoteID uuid.UUID) ([]byte, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get quote
	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	if quote == nil {
		return nil, fmt.Errorf("quote not found")
	}

	// Get customer and property details
	customer, err := s.customerRepo.GetByID(ctx, tenantID, quote.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	property, err := s.propertyRepo.GetByID(ctx, tenantID, quote.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get property: %w", err)
	}

	// Get quote services
	quoteServices, err := s.quoteRepo.GetQuoteServices(ctx, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote services: %w", err)
	}

	// Generate PDF content (simplified HTML to PDF conversion)
	pdfContent := s.generateQuotePDFContent(quote, customer, property, quoteServices)

	// In a real implementation, you would use a PDF generation library like wkhtmltopdf or similar
	// For now, return the HTML content as bytes
	return []byte(pdfContent), nil
}

// SendQuote sends a quote to the customer
func (s *QuoteServiceImpl) SendQuote(ctx context.Context, quoteID uuid.UUID, sendOptions *QuoteSendOptions) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get quote
	quote, err := s.quoteRepo.GetByID(ctx, tenantID, quoteID)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}
	if quote == nil {
		return nil
	}

	// Get customer details
	customer, err := s.customerRepo.GetByID(ctx, tenantID, quote.CustomerID)
	if err != nil {
		return fmt.Errorf("failed to get customer: %w", err)
	}

	// Determine recipient email - use provided email or customer's email
	recipientEmail := sendOptions.Email
	if recipientEmail == "" && customer != nil && customer.Email != nil {
		recipientEmail = *customer.Email
	}
	if recipientEmail == "" {
		return fmt.Errorf("no email address available for customer")
	}

	// Prepare email
	subject := fmt.Sprintf("Quote %s", quote.QuoteNumber)
	if sendOptions.Subject != nil {
		subject = *sendOptions.Subject
	}

	body := fmt.Sprintf("Please find attached your quote %s for $%.2f", quote.QuoteNumber, quote.TotalAmount)
	if sendOptions.Message != nil {
		body = *sendOptions.Message
	}

	emailReq := &EmailRequest{
		To:      []string{recipientEmail},
		Subject: subject,
		Body:    body,
		IsHTML:  false,
	}

	// Add CC emails if provided
	if len(sendOptions.CCEmails) > 0 {
		emailReq.CC = sendOptions.CCEmails
	}

	// Generate and attach PDF if requested
	if sendOptions.IncludePDF {
		pdfData, err := s.GenerateQuotePDF(ctx, quoteID)
		if err != nil {
			s.logger.Printf("Failed to generate PDF for quote email", "error", err, "quote_id", quoteID)
		} else {
			attachment := Attachment{
				Name:        fmt.Sprintf("quote_%s.pdf", quote.QuoteNumber),
				ContentType: "application/pdf",
				Data:        pdfData,
			}
			emailReq.Attachments = []Attachment{attachment}
		}
	}

	// Send email
	if err := s.communicationService.SendEmail(ctx, emailReq); err != nil {
		s.logger.Printf("Failed to send quote email", "error", err, "quote_id", quoteID)
		return fmt.Errorf("failed to send quote email: %w", err)
	}

	// Update quote status to sent if it was draft
	if quote.Status == "draft" {
		quote.Status = "pending"
		quote.UpdatedAt = time.Now()
		if err := s.quoteRepo.Update(ctx, quote); err != nil {
			s.logger.Printf("Failed to update quote status after sending", "error", err)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "quote.send",
		ResourceType: "quote",
		ResourceID:   &quote.ID,
		NewValues: map[string]interface{}{
			"sent_to":     sendOptions.Email,
			"include_pdf": sendOptions.IncludePDF,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Quote sent successfully", "quote_id", quoteID, "email", sendOptions.Email)
	return nil
}

// GenerateQuoteFromDescription uses AI to generate a quote from a description
func (s *QuoteServiceImpl) GenerateQuoteFromDescription(ctx context.Context, req *QuoteGenerationRequest) (*domain.Quote, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify customer and property exist
	customer, err := s.customerRepo.GetByID(ctx, tenantID, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	property, err := s.propertyRepo.GetByID(ctx, tenantID, req.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify property: %w", err)
	}
	if property == nil {
		return nil, fmt.Errorf("property not found")
	}

	// Use LLM service to analyze the description and generate recommendations
	llmResponse, err := s.llmService.GenerateQuote(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI quote: %w", err)
	}

	// Convert LLM recommendations to quote services
	services := make([]QuoteServiceRequest, 0, len(llmResponse.RecommendedServices))
	for _, rec := range llmResponse.RecommendedServices {
		services = append(services, QuoteServiceRequest{
			ServiceID:   rec.ServiceID,
			Quantity:    rec.Quantity,
			UnitPrice:   rec.UnitPrice,
			Description: &rec.Reasoning,
		})
	}

	// Create quote request
	quoteReq := &QuoteCreateRequest{
		CustomerID:  req.CustomerID,
		PropertyID:  req.PropertyID,
		Title:       "AI Generated Quote",
		Description: &req.Description,
		Services:    services,
		Notes:       &llmResponse.Notes,
	}

	// Create the quote
	quote, err := s.CreateQuote(ctx, quoteReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create AI generated quote: %w", err)
	}

	s.logger.Printf("AI quote generated successfully", 
		"quote_id", quote.ID, 
		"confidence", llmResponse.Confidence,
		"estimated_total", llmResponse.EstimatedTotal)

	return quote, nil
}

// Helper methods

func (s *QuoteServiceImpl) validateCreateQuoteRequest(req *QuoteCreateRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("quote title is required")
	}
	if len(req.Services) == 0 {
		return fmt.Errorf("at least one service is required")
	}
	for i, svc := range req.Services {
		if svc.Quantity <= 0 {
			return fmt.Errorf("service %d: quantity must be greater than 0", i+1)
		}
		if svc.UnitPrice < 0 {
			return fmt.Errorf("service %d: unit price cannot be negative", i+1)
		}
	}
	if req.ValidUntil != nil && req.ValidUntil.Before(time.Now()) {
		return fmt.Errorf("valid until date cannot be in the past")
	}
	return nil
}

func (s *QuoteServiceImpl) generateQuotePDFContent(quote *domain.Quote, customer *domain.EnhancedCustomer, property *domain.EnhancedProperty, services []*domain.QuoteService) string {
	// Generate a simple HTML template for the quote
	var buf bytes.Buffer
	
	buf.WriteString("<html><head><title>Quote " + quote.QuoteNumber + "</title></head><body>")
	buf.WriteString("<h1>Quote " + quote.QuoteNumber + "</h1>")
	
	buf.WriteString("<h2>Customer Information</h2>")
	buf.WriteString("<p>" + customer.FirstName + " " + customer.LastName + "</p>")
	if customer.Email != nil {
		buf.WriteString("<p>Email: " + *customer.Email + "</p>")
	}
	
	buf.WriteString("<h2>Property Information</h2>")
	buf.WriteString("<p>" + property.Name + "</p>")
	buf.WriteString("<p>" + property.AddressLine1 + ", " + property.City + ", " + property.State + " " + property.ZipCode + "</p>")
	
	buf.WriteString("<h2>Services</h2>")
	buf.WriteString("<table border='1'>")
	buf.WriteString("<tr><th>Service</th><th>Quantity</th><th>Unit Price</th><th>Total</th></tr>")
	
	for _, svc := range services {
		buf.WriteString("<tr>")
		buf.WriteString("<td>Service " + svc.ServiceID.String() + "</td>")
		buf.WriteString("<td>" + fmt.Sprintf("%.2f", svc.Quantity) + "</td>")
		buf.WriteString("<td>$" + fmt.Sprintf("%.2f", svc.UnitPrice) + "</td>")
		buf.WriteString("<td>$" + fmt.Sprintf("%.2f", svc.TotalPrice) + "</td>")
		buf.WriteString("</tr>")
	}
	
	buf.WriteString("</table>")
	
	buf.WriteString("<h2>Totals</h2>")
	buf.WriteString("<p>Subtotal: $" + fmt.Sprintf("%.2f", quote.Subtotal) + "</p>")
	buf.WriteString("<p>Tax (" + strconv.FormatFloat(quote.TaxRate*100, 'f', 1, 64) + "%): $" + fmt.Sprintf("%.2f", quote.TaxAmount) + "</p>")
	buf.WriteString("<p><strong>Total: $" + fmt.Sprintf("%.2f", quote.TotalAmount) + "</strong></p>")
	
	if quote.ValidUntil != nil {
		buf.WriteString("<p>Valid until: " + quote.ValidUntil.Format("January 2, 2006") + "</p>")
	}
	
	if quote.TermsAndConditions != nil {
		buf.WriteString("<h2>Terms and Conditions</h2>")
		buf.WriteString("<p>" + *quote.TermsAndConditions + "</p>")
	}
	
	buf.WriteString("</body></html>")
	
	return buf.String()
}

// timePtr helper is defined in notification_service.go