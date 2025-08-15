package services

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// InvoiceServiceImpl implements the InvoiceService interface
type InvoiceServiceImpl struct {
	invoiceRepo         InvoiceRepositoryFull
	paymentRepo         PaymentRepositoryFull
	customerRepo        CustomerRepository
	jobRepo             JobRepositoryComplete
	quoteRepo           QuoteRepositoryFull
	auditService        AuditService
	communicationService CommunicationService
	paymentsIntegration PaymentsIntegration
	storageService      StorageService
	logger              *log.Logger
}

// InvoiceRepositoryFull defines the complete interface for invoice data access
type InvoiceRepositoryFull interface {
	// CRUD operations
	Create(ctx context.Context, invoice *domain.Invoice) error
	GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error)
	Update(ctx context.Context, invoice *domain.Invoice) error
	Delete(ctx context.Context, tenantID, invoiceID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *InvoiceFilter) ([]*domain.Invoice, int64, error)
	
	// Filtering operations
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *InvoiceFilter) ([]*domain.Invoice, int64, error)
	GetByJobID(ctx context.Context, tenantID, jobID uuid.UUID) (*domain.Invoice, error)
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Invoice, error)
	GetOverdue(ctx context.Context, tenantID uuid.UUID) ([]*domain.Invoice, error)
	
	// Invoice services
	CreateInvoiceService(ctx context.Context, invoiceService *InvoiceLineItem) error
	UpdateInvoiceService(ctx context.Context, invoiceService *InvoiceLineItem) error
	DeleteInvoiceService(ctx context.Context, invoiceServiceID uuid.UUID) error
	GetInvoiceServices(ctx context.Context, invoiceID uuid.UUID) ([]*InvoiceLineItem, error)
	
	// Invoice numbering
	GetNextInvoiceNumber(ctx context.Context, tenantID uuid.UUID) (string, error)
}

// PaymentRepositoryFull defines the complete interface for payment data access
type PaymentRepositoryFull interface {
	// CRUD operations
	Create(ctx context.Context, payment *domain.Payment) error
	GetByID(ctx context.Context, tenantID, paymentID uuid.UUID) (*domain.Payment, error)
	Update(ctx context.Context, payment *domain.Payment) error
	Delete(ctx context.Context, tenantID, paymentID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *PaymentFilter) ([]*domain.Payment, int64, error)
	
	// Filtering operations
	GetByInvoiceID(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]*domain.Payment, error)
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *PaymentFilter) ([]*domain.Payment, int64, error)
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Payment, error)
	
	// Payment analytics
	GetPaymentSummary(ctx context.Context, tenantID uuid.UUID, filter *PaymentFilter) (*PaymentSummary, error)
}

// InvoiceLineItem represents a service line item on an invoice
type InvoiceLineItem struct {
	ID          uuid.UUID `json:"id" db:"id"`
	InvoiceID   uuid.UUID `json:"invoice_id" db:"invoice_id"`
	ServiceID   uuid.UUID `json:"service_id" db:"service_id"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price"`
	TotalPrice  float64   `json:"total_price" db:"total_price"`
	Description *string   `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// PaymentsIntegration wraps the payments integration
type PaymentsIntegration interface {
	CreatePaymentIntent(ctx context.Context, req *PaymentIntentRequest) (*PaymentIntentResponse, error)
	ConfirmPayment(ctx context.Context, req *ConfirmPaymentRequest) (*PaymentResponse, error)
	CapturePayment(ctx context.Context, req *CapturePaymentRequest) (*PaymentResponse, error)
	RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error)
	CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CustomerResponse, error)
	GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentStatusResponse, error)
	ProcessWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error)
	CalculateFees(amount float64, currency string, paymentMethod string) (*FeeCalculation, error)
}

// NewInvoiceService creates a new invoice service instance
func NewInvoiceService(
	invoiceRepo InvoiceRepositoryFull,
	paymentRepo PaymentRepositoryFull,
	customerRepo CustomerRepository,
	jobRepo JobRepositoryComplete,
	quoteRepo QuoteRepositoryFull,
	auditService AuditService,
	communicationService CommunicationService,
	paymentsIntegration PaymentsIntegration,
	storageService StorageService,
	logger *log.Logger,
) InvoiceService {
	return &InvoiceServiceImpl{
		invoiceRepo:          invoiceRepo,
		paymentRepo:          paymentRepo,
		customerRepo:         customerRepo,
		jobRepo:              jobRepo,
		quoteRepo:            quoteRepo,
		auditService:         auditService,
		communicationService: communicationService,
		paymentsIntegration:  paymentsIntegration,
		storageService:       storageService,
		logger:               logger,
	}
}

// CreateInvoice creates a new invoice
func (s *InvoiceServiceImpl) CreateInvoice(ctx context.Context, req *InvoiceCreateRequest) (*domain.Invoice, error) {
	// Get tenant ID from context
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Validate the request
	if err := s.validateCreateInvoiceRequest(req); err != nil {
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

	// Verify job exists if specified
	if req.JobID != nil {
		job, err := s.jobRepo.GetByID(ctx, tenantID, *req.JobID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify job: %w", err)
		}
		if job == nil {
			return nil, fmt.Errorf("job not found")
		}
		if job.CustomerID != req.CustomerID {
			return nil, fmt.Errorf("job does not belong to the specified customer")
		}
	}

	// Generate invoice number
	invoiceNumber, err := s.invoiceRepo.GetNextInvoiceNumber(ctx, tenantID)
	if err != nil {
		s.logger.Printf("Failed to generate invoice number", "error", err)
		invoiceNumber = fmt.Sprintf("INV-%d", time.Now().Unix())
	}

	// Calculate totals
	subtotal := 0.0
	for _, svc := range req.Services {
		subtotal += svc.Quantity * svc.UnitPrice
	}

	taxAmount := subtotal * req.TaxRate
	totalAmount := subtotal + taxAmount

	// Set due date (30 days from now if not specified)
	dueDate := req.DueDate
	if dueDate == nil {
		futureDate := time.Now().AddDate(0, 0, 30)
		dueDate = &futureDate
	}

	// Create invoice entity
	invoice := &domain.Invoice{
		ID:            uuid.New(),
		TenantID:      tenantID,
		CustomerID:    req.CustomerID,
		JobID:         req.JobID,
		InvoiceNumber: invoiceNumber,
		Status:        "draft",
		Subtotal:      subtotal,
		TaxRate:       req.TaxRate,
		TaxAmount:     taxAmount,
		TotalAmount:   totalAmount,
		IssuedDate:    nil, // Will be set when invoice is sent
		DueDate:       dueDate,
		Notes:         req.Notes,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save invoice to database
	if err := s.invoiceRepo.Create(ctx, invoice); err != nil {
		s.logger.Printf("Failed to create invoice", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to create invoice: %w", err)
	}

	// Create invoice services
	for _, svcReq := range req.Services {
		invoiceService := &InvoiceLineItem{
			ID:          uuid.New(),
			InvoiceID:   invoice.ID,
			ServiceID:   svcReq.ServiceID,
			Quantity:    svcReq.Quantity,
			UnitPrice:   svcReq.UnitPrice,
			TotalPrice:  svcReq.Quantity * svcReq.UnitPrice,
			Description: svcReq.Description,
			CreatedAt:   time.Now(),
		}

		if err := s.invoiceRepo.CreateInvoiceService(ctx, invoiceService); err != nil {
			s.logger.Printf("Failed to create invoice service", "error", err, "invoice_id", invoice.ID, "service_id", svcReq.ServiceID)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "invoice.create",
		ResourceType: "invoice",
		ResourceID:   &invoice.ID,
		NewValues: map[string]interface{}{
			"invoice_number": invoice.InvoiceNumber,
			"customer_id":    invoice.CustomerID,
			"job_id":         invoice.JobID,
			"total_amount":   invoice.TotalAmount,
			"status":         invoice.Status,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Invoice created successfully", "invoice_id", invoice.ID, "invoice_number", invoiceNumber, "tenant_id", tenantID)
	return invoice, nil
}

// GetInvoice retrieves an invoice by ID
func (s *InvoiceServiceImpl) GetInvoice(ctx context.Context, invoiceID uuid.UUID) (*domain.Invoice, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		s.logger.Printf("Failed to get invoice", "error", err, "invoice_id", invoiceID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	return invoice, nil
}

// UpdateInvoice updates an existing invoice
func (s *InvoiceServiceImpl) UpdateInvoice(ctx context.Context, invoiceID uuid.UUID, req *InvoiceUpdateRequest) (*domain.Invoice, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get existing invoice
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Check if invoice can be updated (only draft invoices)
	if invoice.Status != "draft" {
		return nil, fmt.Errorf("only draft invoices can be updated")
	}

	// Store old values for audit
	oldValues := map[string]interface{}{
		"total_amount": invoice.TotalAmount,
		"tax_rate":     invoice.TaxRate,
		"due_date":     invoice.DueDate,
	}

	// Update services if provided
	if len(req.Services) > 0 {
		// Delete existing invoice services
		existingServices, err := s.invoiceRepo.GetInvoiceServices(ctx, invoiceID)
		if err != nil {
			s.logger.Printf("Failed to get existing invoice services", "error", err)
		} else {
			for _, svc := range existingServices {
				if err := s.invoiceRepo.DeleteInvoiceService(ctx, svc.ID); err != nil {
					s.logger.Printf("Failed to delete invoice service", "error", err, "service_id", svc.ID)
				}
			}
		}

		// Create new invoice services and recalculate totals
		subtotal := 0.0
		for _, svcReq := range req.Services {
			totalPrice := svcReq.Quantity * svcReq.UnitPrice
			subtotal += totalPrice

			invoiceService := &InvoiceLineItem{
				ID:          uuid.New(),
				InvoiceID:   invoice.ID,
				ServiceID:   svcReq.ServiceID,
				Quantity:    svcReq.Quantity,
				UnitPrice:   svcReq.UnitPrice,
				TotalPrice:  totalPrice,
				Description: svcReq.Description,
				CreatedAt:   time.Now(),
			}

			if err := s.invoiceRepo.CreateInvoiceService(ctx, invoiceService); err != nil {
				s.logger.Printf("Failed to create updated invoice service", "error", err, "invoice_id", invoice.ID, "service_id", svcReq.ServiceID)
			}
		}

		// Update invoice totals
		invoice.Subtotal = subtotal
	}

	// Update other fields
	if req.TaxRate != nil {
		invoice.TaxRate = *req.TaxRate
	}
	if req.DueDate != nil {
		invoice.DueDate = req.DueDate
	}
	if req.Notes != nil {
		invoice.Notes = req.Notes
	}

	// Recalculate amounts
	invoice.TaxAmount = invoice.Subtotal * invoice.TaxRate
	invoice.TotalAmount = invoice.Subtotal + invoice.TaxAmount
	invoice.UpdatedAt = time.Now()

	// Save to database
	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		s.logger.Printf("Failed to update invoice", "error", err, "invoice_id", invoiceID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update invoice: %w", err)
	}

	// Log audit event
	newValues := map[string]interface{}{
		"total_amount": invoice.TotalAmount,
		"tax_rate":     invoice.TaxRate,
		"due_date":     invoice.DueDate,
	}

	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "invoice.update",
		ResourceType: "invoice",
		ResourceID:   &invoice.ID,
		OldValues:    oldValues,
		NewValues:    newValues,
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Invoice updated successfully", "invoice_id", invoiceID, "tenant_id", tenantID)
	return invoice, nil
}

// DeleteInvoice deletes an invoice (soft delete)
func (s *InvoiceServiceImpl) DeleteInvoice(ctx context.Context, invoiceID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get invoice before deletion for audit log
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return fmt.Errorf("invoice not found")
	}

	// Check if invoice can be deleted (only draft invoices can be deleted)
	if invoice.Status != "draft" {
		return fmt.Errorf("only draft invoices can be deleted")
	}

	// Soft delete by updating status
	invoice.Status = "cancelled"
	invoice.UpdatedAt = time.Now()

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		s.logger.Printf("Failed to delete invoice", "error", err, "invoice_id", invoiceID, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "invoice.delete",
		ResourceType: "invoice",
		ResourceID:   &invoice.ID,
		OldValues: map[string]interface{}{
			"status": "draft",
		},
		NewValues: map[string]interface{}{
			"status": "cancelled",
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Invoice deleted successfully", "invoice_id", invoiceID, "tenant_id", tenantID)
	return nil
}

// ListInvoices lists invoices with filtering and pagination
func (s *InvoiceServiceImpl) ListInvoices(ctx context.Context, filter *InvoiceFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
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
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	invoices, total, err := s.invoiceRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Printf("Failed to list invoices", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to list invoices: %w", err)
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

// SendInvoice sends an invoice to the customer
func (s *InvoiceServiceImpl) SendInvoice(ctx context.Context, invoiceID uuid.UUID, sendOptions *InvoiceSendOptions) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return fmt.Errorf("invoice not found")
	}

	// Get customer details
	customer, err := s.customerRepo.GetByID(ctx, tenantID, invoice.CustomerID)
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
	subject := fmt.Sprintf("Invoice %s", invoice.InvoiceNumber)
	if sendOptions.Subject != nil {
		subject = *sendOptions.Subject
	}

	body := fmt.Sprintf("Please find attached your invoice %s for $%.2f", invoice.InvoiceNumber, invoice.TotalAmount)
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

	// Generate and attach PDF
	pdfData, err := s.GenerateInvoicePDF(ctx, invoiceID)
	if err != nil {
		s.logger.Printf("Failed to generate PDF for invoice email", "error", err, "invoice_id", invoiceID)
	} else {
		attachment := Attachment{
			Name:        fmt.Sprintf("invoice_%s.pdf", invoice.InvoiceNumber),
			ContentType: "application/pdf",
			Data:        pdfData,
		}
		emailReq.Attachments = []Attachment{attachment}
	}

	// Send email
	if err := s.communicationService.SendEmail(ctx, emailReq); err != nil {
		s.logger.Printf("Failed to send invoice email", "error", err, "invoice_id", invoiceID)
		return fmt.Errorf("failed to send invoice email: %w", err)
	}

	// Update invoice status to sent and set issued date
	if invoice.Status == "draft" {
		invoice.Status = "sent"
		now := time.Now()
		invoice.IssuedDate = &now
		invoice.UpdatedAt = time.Now()
		if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
			s.logger.Printf("Failed to update invoice status after sending", "error", err)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "invoice.send",
		ResourceType: "invoice",
		ResourceID:   &invoice.ID,
		NewValues: map[string]interface{}{
			"sent_to": sendOptions.Email,
			"status":  "sent",
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Invoice sent successfully", "invoice_id", invoiceID, "email", sendOptions.Email)
	return nil
}

// GenerateInvoicePDF generates a PDF for the invoice
func (s *InvoiceServiceImpl) GenerateInvoicePDF(ctx context.Context, invoiceID uuid.UUID) ([]byte, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Get customer details
	customer, err := s.customerRepo.GetByID(ctx, tenantID, invoice.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	// Get invoice services
	invoiceServices, err := s.invoiceRepo.GetInvoiceServices(ctx, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice services: %w", err)
	}

	// Generate PDF content (simplified HTML to PDF conversion)
	pdfContent := s.generateInvoicePDFContent(invoice, customer, invoiceServices)

	// In a real implementation, you would use a PDF generation library
	// For now, return the HTML content as bytes
	return []byte(pdfContent), nil
}

// GetInvoicePayments retrieves payments for an invoice
func (s *InvoiceServiceImpl) GetInvoicePayments(ctx context.Context, invoiceID uuid.UUID) ([]*domain.Payment, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify invoice exists
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	payments, err := s.paymentRepo.GetByInvoiceID(ctx, tenantID, invoiceID)
	if err != nil {
		s.logger.Printf("Failed to get invoice payments", "error", err, "invoice_id", invoiceID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get invoice payments: %w", err)
	}

	return payments, nil
}

// MarkInvoiceAsPaid marks an invoice as paid
func (s *InvoiceServiceImpl) MarkInvoiceAsPaid(ctx context.Context, invoiceID uuid.UUID, paymentID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return fmt.Errorf("invoice not found")
	}

	// Verify payment exists and belongs to this invoice
	payment, err := s.paymentRepo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		return fmt.Errorf("failed to get payment: %w", err)
	}
	if payment == nil {
		return fmt.Errorf("payment not found")
	}
	if payment.InvoiceID != invoiceID {
		return fmt.Errorf("payment does not belong to this invoice")
	}

	// Update invoice status
	oldStatus := invoice.Status
	invoice.Status = "paid"
	now := time.Now()
	invoice.PaidDate = &now
	invoice.UpdatedAt = time.Now()

	if err := s.invoiceRepo.Update(ctx, invoice); err != nil {
		s.logger.Printf("Failed to mark invoice as paid", "error", err, "invoice_id", invoiceID)
		return fmt.Errorf("failed to mark invoice as paid: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "invoice.mark_paid",
		ResourceType: "invoice",
		ResourceID:   &invoice.ID,
		OldValues: map[string]interface{}{
			"status": oldStatus,
		},
		NewValues: map[string]interface{}{
			"status":     "paid",
			"paid_date":  invoice.PaidDate,
			"payment_id": paymentID,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Invoice marked as paid", "invoice_id", invoiceID, "payment_id", paymentID)
	return nil
}

// GetOverdueInvoices retrieves all overdue invoices
func (s *InvoiceServiceImpl) GetOverdueInvoices(ctx context.Context) ([]*domain.Invoice, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	invoices, err := s.invoiceRepo.GetOverdue(ctx, tenantID)
	if err != nil {
		s.logger.Printf("Failed to get overdue invoices", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}

	return invoices, nil
}

// SendOverdueReminders sends reminders for overdue invoices
func (s *InvoiceServiceImpl) SendOverdueReminders(ctx context.Context) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get all overdue invoices
	overdueInvoices, err := s.invoiceRepo.GetOverdue(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get overdue invoices: %w", err)
	}

	remindersSent := 0
	for _, invoice := range overdueInvoices {
		// Get customer
		customer, err := s.customerRepo.GetByID(ctx, tenantID, invoice.CustomerID)
		if err != nil {
			s.logger.Printf("Failed to get customer for overdue reminder", "error", err, "invoice_id", invoice.ID)
			continue
		}

		if customer.Email == nil || *customer.Email == "" {
			s.logger.Printf("Customer has no email for overdue reminder", "customer_id", customer.ID, "invoice_id", invoice.ID)
			continue
		}

		// Calculate days overdue
		daysOverdue := int(time.Since(*invoice.DueDate).Hours() / 24)

		// Send reminder email
		subject := fmt.Sprintf("Payment Reminder - Invoice %s", invoice.InvoiceNumber)
		body := fmt.Sprintf("Dear %s %s,\n\nYour invoice %s for $%.2f is %d days overdue. Please submit payment as soon as possible.\n\nThank you.",
			customer.FirstName, customer.LastName, invoice.InvoiceNumber, invoice.TotalAmount, daysOverdue)

		emailReq := &EmailRequest{
			To:      []string{*customer.Email},
			Subject: subject,
			Body:    body,
			IsHTML:  false,
		}

		if err := s.communicationService.SendEmail(ctx, emailReq); err != nil {
			s.logger.Printf("Failed to send overdue reminder", "error", err, "invoice_id", invoice.ID, "customer_email", *customer.Email)
			continue
		}

		remindersSent++
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "invoice.send_overdue_reminders",
		ResourceType: "invoice",
		NewValues: map[string]interface{}{
			"reminders_sent":  remindersSent,
			"total_overdue":   len(overdueInvoices),
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Overdue reminders sent successfully", "reminders_sent", remindersSent, "total_overdue", len(overdueInvoices))
	return nil
}

// CreateInvoiceFromJob creates an invoice from a completed job
func (s *InvoiceServiceImpl) CreateInvoiceFromJob(ctx context.Context, jobID uuid.UUID) (*domain.Invoice, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return nil, fmt.Errorf("job not found")
	}

	// Check if job is completed
	if job.Status != domain.JobStatusCompleted {
		return nil, fmt.Errorf("only completed jobs can be invoiced")
	}

	// Check if invoice already exists for this job
	existingInvoice, err := s.invoiceRepo.GetByJobID(ctx, tenantID, jobID)
	if err == nil && existingInvoice != nil {
		return nil, fmt.Errorf("invoice already exists for this job")
	}

	// Get job services
	jobServices, err := s.jobRepo.GetJobServices(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job services: %w", err)
	}

	if len(jobServices) == 0 {
		return nil, fmt.Errorf("job has no services to invoice")
	}

	// Convert job services to invoice services
	invoiceServiceReqs := make([]InvoiceServiceRequest, len(jobServices))
	for i, jobSvc := range jobServices {
		invoiceServiceReqs[i] = InvoiceServiceRequest{
			ServiceID: jobSvc.ServiceID,
			Quantity:  jobSvc.Quantity,
			UnitPrice: jobSvc.UnitPrice,
		}
	}

	// Create invoice request
	invoiceReq := &InvoiceCreateRequest{
		CustomerID: job.CustomerID,
		JobID:      &job.ID,
		Services:   invoiceServiceReqs,
		TaxRate:    0.08, // Default 8% tax rate
		Notes:      stringPtr(fmt.Sprintf("Invoice for completed job: %s", job.Title)),
	}

	// Create the invoice
	invoice, err := s.CreateInvoice(ctx, invoiceReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create invoice from job: %w", err)
	}

	s.logger.Printf("Invoice created from job successfully", "job_id", jobID, "invoice_id", invoice.ID)
	return invoice, nil
}

// ProcessInvoicePayment processes a payment for an invoice
func (s *InvoiceServiceImpl) ProcessInvoicePayment(ctx context.Context, invoiceID uuid.UUID, req *PaymentProcessRequest) (*domain.Payment, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get invoice
	invoice, err := s.invoiceRepo.GetByID(ctx, tenantID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}
	if invoice == nil {
		return nil, fmt.Errorf("invoice not found")
	}

	// Verify invoice can be paid
	if invoice.Status == "paid" {
		return nil, fmt.Errorf("invoice is already paid")
	}
	if invoice.Status == "cancelled" {
		return nil, fmt.Errorf("cannot pay cancelled invoice")
	}

	// Convert amount to cents for payment processor
	amountCents := int64(req.Amount * 100)

	// Create payment intent
	paymentIntentReq := &PaymentIntentRequest{
		Amount:      amountCents,
		Currency:    "usd", // Default currency
		Description: fmt.Sprintf("Payment for invoice %s", invoice.InvoiceNumber),
		CustomerID:  req.CustomerID.String(),
		Metadata: map[string]string{
			"invoice_id":     invoiceID.String(),
			"tenant_id":      tenantID.String(),
			"invoice_number": invoice.InvoiceNumber,
		},
		PaymentMethodTypes: []string{"card"},
		CaptureMethod:     "automatic",
	}

	// Add customer email if available
	customer, err := s.customerRepo.GetByID(ctx, tenantID, invoice.CustomerID)
	if err == nil && customer != nil && customer.Email != nil {
		paymentIntentReq.CustomerEmail = *customer.Email
	}

	// Create payment intent
	paymentIntent, err := s.paymentsIntegration.CreatePaymentIntent(ctx, paymentIntentReq)
	if err != nil {
		s.logger.Printf("Failed to create payment intent", "error", err, "invoice_id", invoiceID)
		return nil, fmt.Errorf("failed to create payment intent: %w", err)
	}

	// Create payment record
	payment := &domain.Payment{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		InvoiceID:            invoiceID,
		Amount:               req.Amount,
		PaymentMethod:        req.PaymentMethod,
		PaymentGateway:       stringPtr("stripe"), // Default gateway
		GatewayTransactionID: &paymentIntent.ID,
		Status:               "pending",
		Notes:                req.Description,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	// Save payment record
	if err := s.paymentRepo.Create(ctx, payment); err != nil {
		s.logger.Printf("Failed to create payment record", "error", err, "invoice_id", invoiceID)
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "payment.process",
		ResourceType: "payment",
		ResourceID:   &payment.ID,
		NewValues: map[string]interface{}{
			"invoice_id":    invoiceID,
			"amount":        req.Amount,
			"method":        req.PaymentMethod,
			"intent_id":     paymentIntent.ID,
			"status":        "pending",
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Payment processing initiated", "payment_id", payment.ID, "invoice_id", invoiceID, "amount", req.Amount)
	return payment, nil
}

// HandlePaymentWebhook handles webhook events from payment processors
func (s *InvoiceServiceImpl) HandlePaymentWebhook(ctx context.Context, payload []byte, signature string) error {
	event, err := s.paymentsIntegration.ProcessWebhook(ctx, payload, signature)
	if err != nil {
		s.logger.Printf("Failed to process webhook", "error", err)
		return fmt.Errorf("failed to process webhook: %w", err)
	}

	s.logger.Printf("Processing webhook event", "event_id", event.ID, "type", event.Type)

	switch event.Type {
	case "payment_intent.succeeded":
		return s.handlePaymentSucceeded(ctx, event)
	case "payment_intent.payment_failed":
		return s.handlePaymentFailed(ctx, event)
	case "invoice.payment_succeeded":
		return s.handleInvoicePaymentSucceeded(ctx, event)
	default:
		s.logger.Printf("Unhandled webhook event type", "type", event.Type)
	}

	return nil
}

// RefundInvoicePayment refunds a payment for an invoice
func (s *InvoiceServiceImpl) RefundInvoicePayment(ctx context.Context, paymentID uuid.UUID, amount *float64, reason string) (*RefundResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get payment
	payment, err := s.paymentRepo.GetByID(ctx, tenantID, paymentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}
	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Verify payment can be refunded
	if payment.Status != "completed" {
		return nil, fmt.Errorf("only completed payments can be refunded")
	}

	// Convert amount to cents
	var refundAmountCents *int64
	if amount != nil {
		cents := int64(*amount * 100)
		refundAmountCents = &cents
	}

	// Create refund request
	refundReq := &RefundRequest{
		PaymentID: *payment.GatewayTransactionID,
		Amount:    refundAmountCents,
		Reason:    reason,
		Metadata: map[string]string{
			"payment_id":  paymentID.String(),
			"invoice_id":  payment.InvoiceID.String(),
			"tenant_id":   tenantID.String(),
		},
	}

	// Process refund
	refund, err := s.paymentsIntegration.RefundPayment(ctx, refundReq)
	if err != nil {
		s.logger.Printf("Failed to process refund", "error", err, "payment_id", paymentID)
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// Update payment status
	payment.Status = "refunded"
	payment.UpdatedAt = time.Now()
	if err := s.paymentRepo.Update(ctx, payment); err != nil {
		s.logger.Printf("Failed to update payment status after refund", "error", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "payment.refund",
		ResourceType: "payment",
		ResourceID:   &payment.ID,
		NewValues: map[string]interface{}{
			"refund_id": refund.ID,
			"amount":    refund.Amount,
			"reason":    reason,
			"status":    "refunded",
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Payment refunded successfully", "payment_id", paymentID, "refund_id", refund.ID, "amount", refund.Amount)
	return refund, nil
}

// CalculatePaymentFees calculates payment processing fees for an invoice
func (s *InvoiceServiceImpl) CalculatePaymentFees(ctx context.Context, amount float64, paymentMethod string) (*FeeCalculation, error) {
	fees, err := s.paymentsIntegration.CalculateFees(amount, "usd", paymentMethod)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate fees: %w", err)
	}

	return fees, nil
}

// GetPaymentMethods returns available payment methods for a customer
func (s *InvoiceServiceImpl) GetPaymentMethods(ctx context.Context, customerID uuid.UUID) ([]PaymentMethod, error) {
	// This would integrate with the payment provider to get saved payment methods
	// For now, return a basic set
	return []PaymentMethod{
		{
			ID:        "card",
			Type:      "card",
			IsDefault: true,
		},
		{
			ID:   "bank_transfer",
			Type: "bank_transfer",
		},
	}, nil
}

// Private webhook handlers

func (s *InvoiceServiceImpl) handlePaymentSucceeded(ctx context.Context, event *WebhookEvent) error {
	// Extract payment intent ID from event data
	paymentIntentID, ok := event.Data["id"].(string)
	if !ok {
		return fmt.Errorf("invalid payment intent ID in webhook data")
	}

	// Find payment by gateway transaction ID
	// This would require a new repository method to find by gateway transaction ID
	s.logger.Printf("Payment succeeded", "payment_intent_id", paymentIntentID)

	// Update payment status and mark invoice as paid
	// Implementation would depend on your specific payment flow

	return nil
}

func (s *InvoiceServiceImpl) handlePaymentFailed(ctx context.Context, event *WebhookEvent) error {
	// Extract payment intent ID from event data
	paymentIntentID, ok := event.Data["id"].(string)
	if !ok {
		return fmt.Errorf("invalid payment intent ID in webhook data")
	}

	s.logger.Printf("Payment failed", "payment_intent_id", paymentIntentID)

	// Update payment status to failed
	// Send notification to customer
	// Implementation would depend on your specific payment flow

	return nil
}

func (s *InvoiceServiceImpl) handleInvoicePaymentSucceeded(ctx context.Context, event *WebhookEvent) error {
	// Handle successful invoice payment
	s.logger.Printf("Invoice payment succeeded", "event_id", event.ID)
	return nil
}

// Helper methods

func (s *InvoiceServiceImpl) validateCreateInvoiceRequest(req *InvoiceCreateRequest) error {
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
	if req.TaxRate < 0 || req.TaxRate > 1 {
		return fmt.Errorf("tax rate must be between 0 and 1")
	}
	if req.DueDate != nil && req.DueDate.Before(time.Now()) {
		return fmt.Errorf("due date cannot be in the past")
	}
	return nil
}

func (s *InvoiceServiceImpl) generateInvoicePDFContent(invoice *domain.Invoice, customer *domain.EnhancedCustomer, services []*InvoiceLineItem) string {
	// Generate a simple HTML template for the invoice
	var content strings.Builder
	
	content.WriteString("<html><head><title>Invoice " + invoice.InvoiceNumber + "</title></head><body>")
	content.WriteString("<h1>Invoice " + invoice.InvoiceNumber + "</h1>")
	
	if invoice.IssuedDate != nil {
		content.WriteString("<p>Issued: " + invoice.IssuedDate.Format("January 2, 2006") + "</p>")
	}
	if invoice.DueDate != nil {
		content.WriteString("<p>Due: " + invoice.DueDate.Format("January 2, 2006") + "</p>")
	}
	
	content.WriteString("<h2>Bill To</h2>")
	content.WriteString("<p>" + customer.FirstName + " " + customer.LastName + "</p>")
	if customer.Email != nil {
		content.WriteString("<p>Email: " + *customer.Email + "</p>")
	}
	if customer.Phone != nil {
		content.WriteString("<p>Phone: " + *customer.Phone + "</p>")
	}
	
	content.WriteString("<h2>Services</h2>")
	content.WriteString("<table border='1'>")
	content.WriteString("<tr><th>Description</th><th>Quantity</th><th>Unit Price</th><th>Total</th></tr>")
	
	for _, svc := range services {
		content.WriteString("<tr>")
		if svc.Description != nil {
			content.WriteString("<td>" + *svc.Description + "</td>")
		} else {
			content.WriteString("<td>Service " + svc.ServiceID.String() + "</td>")
		}
		content.WriteString("<td>" + fmt.Sprintf("%.2f", svc.Quantity) + "</td>")
		content.WriteString("<td>$" + fmt.Sprintf("%.2f", svc.UnitPrice) + "</td>")
		content.WriteString("<td>$" + fmt.Sprintf("%.2f", svc.TotalPrice) + "</td>")
		content.WriteString("</tr>")
	}
	
	content.WriteString("</table>")
	
	content.WriteString("<h2>Totals</h2>")
	content.WriteString("<p>Subtotal: $" + fmt.Sprintf("%.2f", invoice.Subtotal) + "</p>")
	content.WriteString("<p>Tax (" + strconv.FormatFloat(invoice.TaxRate*100, 'f', 1, 64) + "%): $" + fmt.Sprintf("%.2f", invoice.TaxAmount) + "</p>")
	content.WriteString("<p><strong>Total: $" + fmt.Sprintf("%.2f", invoice.TotalAmount) + "</strong></p>")
	
	content.WriteString("<p>Status: " + strings.Title(invoice.Status) + "</p>")
	
	if invoice.Notes != nil {
		content.WriteString("<h2>Notes</h2>")
		content.WriteString("<p>" + *invoice.Notes + "</p>")
	}
	
	content.WriteString("</body></html>")
	
	return content.String()
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// Helper function to get user ID from context with fallback
func GetUserIDFromContext(ctx context.Context) *uuid.UUID {
	if userID := ctx.Value("user_id"); userID != nil {
		if uid, ok := userID.(uuid.UUID); ok {
			return &uid
		}
		if uid, ok := userID.(*uuid.UUID); ok {
			return uid
		}
	}
	return nil
}

// Helper function to get tenant ID from context
func GetTenantIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	if tenantID := ctx.Value("tenant_id"); tenantID != nil {
		if tid, ok := tenantID.(uuid.UUID); ok {
			return tid, true
		}
		if tid, ok := tenantID.(*uuid.UUID); ok && tid != nil {
			return *tid, true
		}
	}
	return uuid.Nil, false
}