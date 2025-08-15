package services

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// Mock implementations for testing

type MockInvoiceRepository struct {
	mock.Mock
}

func (m *MockInvoiceRepository) Create(ctx context.Context, invoice *domain.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockInvoiceRepository) GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error) {
	args := m.Called(ctx, tenantID, invoiceID)
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) Update(ctx context.Context, invoice *domain.Invoice) error {
	args := m.Called(ctx, invoice)
	return args.Error(0)
}

func (m *MockInvoiceRepository) Delete(ctx context.Context, tenantID, invoiceID uuid.UUID) error {
	args := m.Called(ctx, tenantID, invoiceID)
	return args.Error(0)
}

func (m *MockInvoiceRepository) List(ctx context.Context, tenantID uuid.UUID, filter *InvoiceFilter) ([]*domain.Invoice, int64, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Get(0).([]*domain.Invoice), args.Get(1).(int64), args.Error(2)
}

func (m *MockInvoiceRepository) GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *InvoiceFilter) ([]*domain.Invoice, int64, error) {
	args := m.Called(ctx, tenantID, customerID, filter)
	return args.Get(0).([]*domain.Invoice), args.Get(1).(int64), args.Error(2)
}

func (m *MockInvoiceRepository) GetByJobID(ctx context.Context, tenantID, jobID uuid.UUID) (*domain.Invoice, error) {
	args := m.Called(ctx, tenantID, jobID)
	return args.Get(0).(*domain.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Invoice, error) {
	args := m.Called(ctx, tenantID, status)
	return args.Get(0).([]*domain.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) GetOverdue(ctx context.Context, tenantID uuid.UUID) ([]*domain.Invoice, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*domain.Invoice), args.Error(1)
}

func (m *MockInvoiceRepository) CreateInvoiceService(ctx context.Context, invoiceService *InvoiceService) error {
	args := m.Called(ctx, invoiceService)
	return args.Error(0)
}

func (m *MockInvoiceRepository) UpdateInvoiceService(ctx context.Context, invoiceService *InvoiceService) error {
	args := m.Called(ctx, invoiceService)
	return args.Error(0)
}

func (m *MockInvoiceRepository) DeleteInvoiceService(ctx context.Context, invoiceServiceID uuid.UUID) error {
	args := m.Called(ctx, invoiceServiceID)
	return args.Error(0)
}

func (m *MockInvoiceRepository) GetInvoiceServices(ctx context.Context, invoiceID uuid.UUID) ([]*InvoiceService, error) {
	args := m.Called(ctx, invoiceID)
	return args.Get(0).([]*InvoiceService), args.Error(1)
}

func (m *MockInvoiceRepository) GetNextInvoiceNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	args := m.Called(ctx, tenantID)
	return args.String(0), args.Error(1)
}

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) GetByID(ctx context.Context, tenantID, paymentID uuid.UUID) (*domain.Payment, error) {
	args := m.Called(ctx, tenantID, paymentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) Update(ctx context.Context, payment *domain.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) Delete(ctx context.Context, tenantID, paymentID uuid.UUID) error {
	args := m.Called(ctx, tenantID, paymentID)
	return args.Error(0)
}

func (m *MockPaymentRepository) List(ctx context.Context, tenantID uuid.UUID, filter *PaymentFilter) ([]*domain.Payment, int64, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Get(0).([]*domain.Payment), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRepository) GetByInvoiceID(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]*domain.Payment, error) {
	args := m.Called(ctx, tenantID, invoiceID)
	return args.Get(0).([]*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *PaymentFilter) ([]*domain.Payment, int64, error) {
	args := m.Called(ctx, tenantID, customerID, filter)
	return args.Get(0).([]*domain.Payment), args.Get(1).(int64), args.Error(2)
}

func (m *MockPaymentRepository) GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Payment, error) {
	args := m.Called(ctx, tenantID, status)
	return args.Get(0).([]*domain.Payment), args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentSummary(ctx context.Context, tenantID uuid.UUID, filter *PaymentFilter) (*PaymentSummary, error) {
	args := m.Called(ctx, tenantID, filter)
	return args.Get(0).(*PaymentSummary), args.Error(1)
}

type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) GetByID(ctx context.Context, tenantID, customerID uuid.UUID) (*domain.EnhancedCustomer, error) {
	args := m.Called(ctx, tenantID, customerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.EnhancedCustomer), args.Error(1)
}

type MockPaymentsIntegration struct {
	mock.Mock
}

func (m *MockPaymentsIntegration) CreatePaymentIntent(ctx context.Context, req *PaymentIntentRequest) (*PaymentIntentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*PaymentIntentResponse), args.Error(1)
}

func (m *MockPaymentsIntegration) ConfirmPayment(ctx context.Context, req *ConfirmPaymentRequest) (*PaymentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*PaymentResponse), args.Error(1)
}

func (m *MockPaymentsIntegration) CapturePayment(ctx context.Context, req *CapturePaymentRequest) (*PaymentResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*PaymentResponse), args.Error(1)
}

func (m *MockPaymentsIntegration) RefundPayment(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*RefundResponse), args.Error(1)
}

func (m *MockPaymentsIntegration) CreateCustomer(ctx context.Context, req *CreateCustomerRequest) (*CustomerResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*CustomerResponse), args.Error(1)
}

func (m *MockPaymentsIntegration) GetPaymentStatus(ctx context.Context, paymentID string) (*PaymentStatusResponse, error) {
	args := m.Called(ctx, paymentID)
	return args.Get(0).(*PaymentStatusResponse), args.Error(1)
}

func (m *MockPaymentsIntegration) ProcessWebhook(ctx context.Context, payload []byte, signature string) (*WebhookEvent, error) {
	args := m.Called(ctx, payload, signature)
	return args.Get(0).(*WebhookEvent), args.Error(1)
}

func (m *MockPaymentsIntegration) CalculateFees(amount float64, currency string, paymentMethod string) (*FeeCalculation, error) {
	args := m.Called(amount, currency, paymentMethod)
	return args.Get(0).(*FeeCalculation), args.Error(1)
}

type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogAction(ctx context.Context, req *AuditLogRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

type MockCommunicationService struct {
	mock.Mock
}

func (m *MockCommunicationService) SendEmail(ctx context.Context, req *EmailRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) Store(ctx context.Context, key string, data []byte) error {
	args := m.Called(ctx, key, data)
	return args.Error(0)
}

// Integration tests

func TestInvoiceServiceIntegration(t *testing.T) {
	// Setup
	mockInvoiceRepo := new(MockInvoiceRepository)
	mockPaymentRepo := new(MockPaymentRepository)
	mockCustomerRepo := new(MockCustomerRepository)
	mockAuditService := new(MockAuditService)
	mockCommunicationService := new(MockCommunicationService)
	mockPaymentsIntegration := new(MockPaymentsIntegration)
	mockStorageService := new(MockStorageService)

	service := NewInvoiceService(
		mockInvoiceRepo,
		mockPaymentRepo,
		mockCustomerRepo,
		nil, // jobRepo
		nil, // quoteRepo
		mockAuditService,
		mockCommunicationService,
		mockPaymentsIntegration,
		mockStorageService,
		nil, // logger
	)

	// Test data
	tenantID := uuid.New()
	customerID := uuid.New()
	ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
	ctx = context.WithValue(ctx, "user_id", uuid.New())

	customer := &domain.EnhancedCustomer{
		Customer: domain.Customer{
			ID:        customerID,
			TenantID:  tenantID,
			FirstName: "John",
			LastName:  "Doe",
		},
		Email: stringPtr("john@example.com"),
	}

	t.Run("CreateInvoice", func(t *testing.T) {
		// Mock expectations
		mockCustomerRepo.On("GetByID", ctx, tenantID, customerID).Return(customer, nil)
		mockInvoiceRepo.On("GetNextInvoiceNumber", ctx, tenantID).Return("INV-2024-0001", nil)
		mockInvoiceRepo.On("Create", ctx, mock.AnythingOfType("*domain.Invoice")).Return(nil)
		mockInvoiceRepo.On("CreateInvoiceService", ctx, mock.AnythingOfType("*services.InvoiceService")).Return(nil)
		mockAuditService.On("LogAction", ctx, mock.AnythingOfType("*services.AuditLogRequest")).Return(nil)

		// Test
		req := &InvoiceCreateRequest{
			CustomerID: customerID,
			Services: []InvoiceServiceRequest{
				{
					ServiceID: uuid.New(),
					Quantity:  1.0,
					UnitPrice: 100.0,
				},
			},
			TaxRate: 0.08,
		}

		invoice, err := service.CreateInvoice(ctx, req)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, invoice)
		assert.Equal(t, customerID, invoice.CustomerID)
		assert.Equal(t, "INV-2024-0001", invoice.InvoiceNumber)
		assert.Equal(t, 100.0, invoice.Subtotal)
		assert.Equal(t, 8.0, invoice.TaxAmount)
		assert.Equal(t, 108.0, invoice.TotalAmount)

		// Verify mocks
		mockCustomerRepo.AssertExpectations(t)
		mockInvoiceRepo.AssertExpectations(t)
		mockAuditService.AssertExpectations(t)
	})

	t.Run("ProcessInvoicePayment", func(t *testing.T) {
		// Reset mocks
		mockInvoiceRepo = new(MockInvoiceRepository)
		mockPaymentRepo = new(MockPaymentRepository)
		mockCustomerRepo = new(MockCustomerRepository)
		mockPaymentsIntegration = new(MockPaymentsIntegration)
		mockAuditService = new(MockAuditService)

		service := NewInvoiceService(
			mockInvoiceRepo,
			mockPaymentRepo,
			mockCustomerRepo,
			nil, nil,
			mockAuditService,
			mockCommunicationService,
			mockPaymentsIntegration,
			mockStorageService,
			nil,
		)

		invoiceID := uuid.New()
		invoice := &domain.Invoice{
			ID:            invoiceID,
			TenantID:      tenantID,
			CustomerID:    customerID,
			InvoiceNumber: "INV-2024-0001",
			Status:        "sent",
			TotalAmount:   108.0,
		}

		paymentIntentResponse := &PaymentIntentResponse{
			ID:           "pi_test123",
			ClientSecret: "pi_test123_secret",
			Status:       "requires_payment_method",
			Amount:       10800, // $108.00 in cents
			Currency:     "usd",
			CreatedAt:    time.Now(),
		}

		// Mock expectations
		mockInvoiceRepo.On("GetByID", ctx, tenantID, invoiceID).Return(invoice, nil)
		mockCustomerRepo.On("GetByID", ctx, tenantID, customerID).Return(customer, nil)
		mockPaymentsIntegration.On("CreatePaymentIntent", ctx, mock.AnythingOfType("*services.PaymentIntentRequest")).Return(paymentIntentResponse, nil)
		mockPaymentRepo.On("Create", ctx, mock.AnythingOfType("*domain.Payment")).Return(nil)
		mockAuditService.On("LogAction", ctx, mock.AnythingOfType("*services.AuditLogRequest")).Return(nil)

		// Test
		req := &PaymentProcessRequest{
			InvoiceID:     invoiceID,
			Amount:        108.0,
			PaymentMethod: "card",
			CustomerID:    customerID.String(),
		}

		payment, err := service.ProcessInvoicePayment(ctx, invoiceID, req)

		// Assertions
		assert.NoError(t, err)
		assert.NotNil(t, payment)
		assert.Equal(t, invoiceID, payment.InvoiceID)
		assert.Equal(t, 108.0, payment.Amount)
		assert.Equal(t, "pending", payment.Status)

		// Verify mocks
		mockInvoiceRepo.AssertExpectations(t)
		mockCustomerRepo.AssertExpectations(t)
		mockPaymentsIntegration.AssertExpectations(t)
		mockPaymentRepo.AssertExpectations(t)
		mockAuditService.AssertExpectations(t)
	})
}

func TestEquipmentServiceIntegration(t *testing.T) {
	// This would test equipment service functionality
	// Similar structure to invoice service tests
	t.Skip("Equipment service integration tests to be implemented")
}

func TestQuoteServiceIntegration(t *testing.T) {
	// This would test quote service functionality including AI integration
	// Similar structure to invoice service tests
	t.Skip("Quote service integration tests to be implemented")
}

func TestBillingServiceIntegration(t *testing.T) {
	// This would test subscription billing functionality
	// Similar structure to invoice service tests
	t.Skip("Billing service integration tests to be implemented")
}

// Benchmark tests

func BenchmarkInvoiceCreation(b *testing.B) {
	// Setup service with mock dependencies
	mockInvoiceRepo := new(MockInvoiceRepository)
	mockCustomerRepo := new(MockCustomerRepository)
	mockAuditService := new(MockAuditService)

	// Setup mocks to return quickly
	mockCustomerRepo.On("GetByID", mock.Anything, mock.Anything, mock.Anything).Return(&domain.EnhancedCustomer{}, nil)
	mockInvoiceRepo.On("GetNextInvoiceNumber", mock.Anything, mock.Anything).Return("INV-2024-0001", nil)
	mockInvoiceRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
	mockInvoiceRepo.On("CreateInvoiceService", mock.Anything, mock.Anything).Return(nil)
	mockAuditService.On("LogAction", mock.Anything, mock.Anything).Return(nil)

	service := NewInvoiceService(
		mockInvoiceRepo,
		nil, // paymentRepo
		mockCustomerRepo,
		nil, nil, // jobRepo, quoteRepo
		mockAuditService,
		nil, nil, nil, nil, // other services
	)

	ctx := context.WithValue(context.Background(), "tenant_id", uuid.New())
	ctx = context.WithValue(ctx, "user_id", uuid.New())

	req := &InvoiceCreateRequest{
		CustomerID: uuid.New(),
		Services: []InvoiceServiceRequest{
			{
				ServiceID: uuid.New(),
				Quantity:  1.0,
				UnitPrice: 100.0,
			},
		},
		TaxRate: 0.08,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := service.CreateInvoice(ctx, req)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Error handling tests

func TestInvoiceServiceErrorHandling(t *testing.T) {
	mockInvoiceRepo := new(MockInvoiceRepository)
	mockCustomerRepo := new(MockCustomerRepository)

	service := NewInvoiceService(
		mockInvoiceRepo,
		nil, // paymentRepo
		mockCustomerRepo,
		nil, nil, // jobRepo, quoteRepo
		nil, nil, nil, nil, nil, // other services
	)

	t.Run("CreateInvoice_InvalidTenantID", func(t *testing.T) {
		ctx := context.Background() // No tenant ID in context

		req := &InvoiceCreateRequest{
			CustomerID: uuid.New(),
			Services: []InvoiceServiceRequest{
				{
					ServiceID: uuid.New(),
					Quantity:  1.0,
					UnitPrice: 100.0,
				},
			},
			TaxRate: 0.08,
		}

		invoice, err := service.CreateInvoice(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "tenant ID not found in context")
	})

	t.Run("CreateInvoice_CustomerNotFound", func(t *testing.T) {
		tenantID := uuid.New()
		customerID := uuid.New()
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)

		mockCustomerRepo.On("GetByID", ctx, tenantID, customerID).Return(nil, nil)

		req := &InvoiceCreateRequest{
			CustomerID: customerID,
			Services: []InvoiceServiceRequest{
				{
					ServiceID: uuid.New(),
					Quantity:  1.0,
					UnitPrice: 100.0,
				},
			},
			TaxRate: 0.08,
		}

		invoice, err := service.CreateInvoice(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, invoice)
		assert.Contains(t, err.Error(), "customer not found")
		mockCustomerRepo.AssertExpectations(t)
	})
}