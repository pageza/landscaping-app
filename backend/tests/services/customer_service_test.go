package services_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// MockCustomerRepository is a mock implementation of services.CustomerRepository
type MockCustomerRepository struct {
	mock.Mock
}

func (m *MockCustomerRepository) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) GetCustomerByID(ctx context.Context, id uuid.UUID) (*domain.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCustomerRepository) GetCustomerByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.Customer, error) {
	args := m.Called(ctx, tenantID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockCustomerRepository) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockCustomerRepository) DeleteCustomer(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockCustomerRepository) ListCustomers(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}, offset, limit int) ([]*domain.Customer, int, error) {
	args := m.Called(ctx, tenantID, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.Customer), args.Int(1), args.Error(2)
}

func (m *MockCustomerRepository) SearchCustomers(ctx context.Context, tenantID uuid.UUID, query string, limit int) ([]*domain.Customer, error) {
	args := m.Called(ctx, tenantID, query, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Customer), args.Error(1)
}

// MockAuditService is a mock implementation of services.AuditService
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) LogActivity(ctx context.Context, activity *domain.AuditLog) error {
	args := m.Called(ctx, activity)
	return args.Error(0)
}

func (m *MockAuditService) GetAuditLogs(ctx context.Context, tenantID uuid.UUID, filters map[string]interface{}, offset, limit int) ([]*domain.AuditLog, int, error) {
	args := m.Called(ctx, tenantID, filters, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*domain.AuditLog), args.Int(1), args.Error(2)
}

// MockNotificationService is a mock implementation of services.NotificationService
type MockNotificationService struct {
	mock.Mock
}

func (m *MockNotificationService) SendNotification(ctx context.Context, notification *domain.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationService) SendEmail(ctx context.Context, to, subject, body string, data map[string]interface{}) error {
	args := m.Called(ctx, to, subject, body, data)
	return args.Error(0)
}

func (m *MockNotificationService) SendSMS(ctx context.Context, to, message string) error {
	args := m.Called(ctx, to, message)
	return args.Error(0)
}

func (m *MockNotificationService) SendPushNotification(ctx context.Context, userID uuid.UUID, title, message string, data map[string]interface{}) error {
	args := m.Called(ctx, userID, title, message, data)
	return args.Error(0)
}

// Test data helpers
func createTestCustomer() *domain.Customer {
	return &domain.Customer{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Email:    "test@example.com",
		Phone:    "+1234567890",
		Name:     "John Doe",
		Company:  "Test Company",
		Address:  "123 Main St",
		City:     "TestCity",
		State:    "TS",
		ZipCode:  "12345",
		Country:  "USA",
		Status:   "active",
		Tags:     []string{"vip", "residential"},
		Metadata: map[string]interface{}{
			"source": "website",
			"notes":  "Important customer",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestCustomerService_CreateCustomer(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockCustomerRepository)
	mockAudit := new(MockAuditService)
	mockNotif := new(MockNotificationService)

	svc := services.NewCustomerService(mockRepo, mockAudit, mockNotif)

	t.Run("successful customer creation", func(t *testing.T) {
		customer := createTestCustomer()
		userID := uuid.New()
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		// Mock expectations
		mockRepo.On("GetCustomerByEmail", ctx, customer.TenantID, customer.Email).Return(nil, errors.New("not found")).Once()
		mockRepo.On("CreateCustomer", ctx, mock.AnythingOfType("*domain.Customer")).Return(nil).Once()
		mockAudit.On("LogActivity", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()
		mockNotif.On("SendEmail", ctx, customer.Email, mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()

		err := svc.CreateCustomer(ctx, customer)

		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, customer.ID)
		assert.Equal(t, "active", customer.Status)
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
		mockNotif.AssertExpectations(t)
	})

	t.Run("duplicate email error", func(t *testing.T) {
		customer := createTestCustomer()
		existingCustomer := createTestCustomer()
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		// Mock expectations
		mockRepo.On("GetCustomerByEmail", ctx, customer.TenantID, customer.Email).Return(existingCustomer, nil).Once()

		err := svc.CreateCustomer(ctx, customer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already exists")
		mockRepo.AssertExpectations(t)
	})

	t.Run("validation error - missing email", func(t *testing.T) {
		customer := createTestCustomer()
		customer.Email = ""
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		err := svc.CreateCustomer(ctx, customer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email is required")
	})

	t.Run("validation error - invalid email format", func(t *testing.T) {
		customer := createTestCustomer()
		customer.Email = "invalid-email"
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		err := svc.CreateCustomer(ctx, customer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})

	t.Run("repository error", func(t *testing.T) {
		customer := createTestCustomer()
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		// Mock expectations
		mockRepo.On("GetCustomerByEmail", ctx, customer.TenantID, customer.Email).Return(nil, errors.New("not found")).Once()
		mockRepo.On("CreateCustomer", ctx, mock.AnythingOfType("*domain.Customer")).Return(errors.New("database error")).Once()

		err := svc.CreateCustomer(ctx, customer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database error")
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerService_GetCustomer(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockCustomerRepository)
	mockAudit := new(MockAuditService)
	mockNotif := new(MockNotificationService)

	svc := services.NewCustomerService(mockRepo, mockAudit, mockNotif)

	t.Run("successful customer retrieval", func(t *testing.T) {
		customer := createTestCustomer()
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		// Mock expectations
		mockRepo.On("GetCustomerByID", ctx, customer.ID).Return(customer, nil).Once()

		result, err := svc.GetCustomer(ctx, customer.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, customer.ID, result.ID)
		assert.Equal(t, customer.Email, result.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("customer not found", func(t *testing.T) {
		customerID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", uuid.New())

		// Mock expectations
		mockRepo.On("GetCustomerByID", ctx, customerID).Return(nil, errors.New("not found")).Once()

		result, err := svc.GetCustomer(ctx, customerID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("tenant isolation check", func(t *testing.T) {
		customer := createTestCustomer()
		wrongTenantID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", wrongTenantID)

		// Mock expectations
		mockRepo.On("GetCustomerByID", ctx, customer.ID).Return(customer, nil).Once()

		result, err := svc.GetCustomer(ctx, customer.ID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "access denied")
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerService_UpdateCustomer(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockCustomerRepository)
	mockAudit := new(MockAuditService)
	mockNotif := new(MockNotificationService)

	svc := services.NewCustomerService(mockRepo, mockAudit, mockNotif)

	t.Run("successful customer update", func(t *testing.T) {
		customer := createTestCustomer()
		userID := uuid.New()
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		updatedCustomer := *customer
		updatedCustomer.Name = "Jane Doe"
		updatedCustomer.Phone = "+9876543210"

		// Mock expectations
		mockRepo.On("GetCustomerByID", ctx, customer.ID).Return(customer, nil).Once()
		mockRepo.On("UpdateCustomer", ctx, mock.AnythingOfType("*domain.Customer")).Return(nil).Once()
		mockAudit.On("LogActivity", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()

		err := svc.UpdateCustomer(ctx, &updatedCustomer)

		assert.NoError(t, err)
		assert.Equal(t, "Jane Doe", updatedCustomer.Name)
		assert.Equal(t, "+9876543210", updatedCustomer.Phone)
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})

	t.Run("customer not found", func(t *testing.T) {
		customer := createTestCustomer()
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		// Mock expectations
		mockRepo.On("GetCustomerByID", ctx, customer.ID).Return(nil, errors.New("not found")).Once()

		err := svc.UpdateCustomer(ctx, customer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("email uniqueness validation", func(t *testing.T) {
		customer := createTestCustomer()
		existingCustomer := createTestCustomer()
		existingCustomer.ID = uuid.New()
		existingCustomer.Email = "existing@example.com"

		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		updatedCustomer := *customer
		updatedCustomer.Email = "existing@example.com"

		// Mock expectations
		mockRepo.On("GetCustomerByID", ctx, customer.ID).Return(customer, nil).Once()
		mockRepo.On("GetCustomerByEmail", ctx, customer.TenantID, "existing@example.com").Return(existingCustomer, nil).Once()

		err := svc.UpdateCustomer(ctx, &updatedCustomer)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "email already in use")
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerService_DeleteCustomer(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockCustomerRepository)
	mockAudit := new(MockAuditService)
	mockNotif := new(MockNotificationService)

	svc := services.NewCustomerService(mockRepo, mockAudit, mockNotif)

	t.Run("successful customer deletion", func(t *testing.T) {
		customer := createTestCustomer()
		userID := uuid.New()
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		// Mock expectations
		mockRepo.On("GetCustomerByID", ctx, customer.ID).Return(customer, nil).Once()
		mockRepo.On("DeleteCustomer", ctx, customer.ID).Return(nil).Once()
		mockAudit.On("LogActivity", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()

		err := svc.DeleteCustomer(ctx, customer.ID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})

	t.Run("customer not found", func(t *testing.T) {
		customerID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", uuid.New())

		// Mock expectations
		mockRepo.On("GetCustomerByID", ctx, customerID).Return(nil, errors.New("not found")).Once()

		err := svc.DeleteCustomer(ctx, customerID)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("soft delete for customer with history", func(t *testing.T) {
		customer := createTestCustomer()
		customer.Metadata["has_jobs"] = true
		userID := uuid.New()
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", customer.TenantID)

		// Mock expectations
		mockRepo.On("GetCustomerByID", ctx, customer.ID).Return(customer, nil).Once()
		mockRepo.On("UpdateCustomer", ctx, mock.AnythingOfType("*domain.Customer")).Return(nil).Once()
		mockAudit.On("LogActivity", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()

		err := svc.DeleteCustomer(ctx, customer.ID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})
}

func TestCustomerService_ListCustomers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockCustomerRepository)
	mockAudit := new(MockAuditService)
	mockNotif := new(MockNotificationService)

	svc := services.NewCustomerService(mockRepo, mockAudit, mockNotif)

	t.Run("successful customer listing", func(t *testing.T) {
		tenantID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		customers := []*domain.Customer{
			createTestCustomer(),
			createTestCustomer(),
			createTestCustomer(),
		}

		filters := map[string]interface{}{
			"status": "active",
			"tags":   []string{"vip"},
		}

		// Mock expectations
		mockRepo.On("ListCustomers", ctx, tenantID, filters, 0, 10).Return(customers, 3, nil).Once()

		result, total, err := svc.ListCustomers(ctx, filters, 0, 10)

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Equal(t, 3, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty result set", func(t *testing.T) {
		tenantID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		filters := map[string]interface{}{
			"status": "inactive",
		}

		// Mock expectations
		mockRepo.On("ListCustomers", ctx, tenantID, filters, 0, 10).Return([]*domain.Customer{}, 0, nil).Once()

		result, total, err := svc.ListCustomers(ctx, filters, 0, 10)

		assert.NoError(t, err)
		assert.Len(t, result, 0)
		assert.Equal(t, 0, total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("pagination", func(t *testing.T) {
		tenantID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		customers := []*domain.Customer{
			createTestCustomer(),
			createTestCustomer(),
		}

		// Mock expectations
		mockRepo.On("ListCustomers", ctx, tenantID, map[string]interface{}{}, 10, 10).Return(customers, 25, nil).Once()

		result, total, err := svc.ListCustomers(ctx, map[string]interface{}{}, 10, 10)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Equal(t, 25, total)
		mockRepo.AssertExpectations(t)
	})
}

func TestCustomerService_SearchCustomers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockCustomerRepository)
	mockAudit := new(MockAuditService)
	mockNotif := new(MockNotificationService)

	svc := services.NewCustomerService(mockRepo, mockAudit, mockNotif)

	t.Run("successful search", func(t *testing.T) {
		tenantID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		customers := []*domain.Customer{
			createTestCustomer(),
			createTestCustomer(),
		}

		query := "john"

		// Mock expectations
		mockRepo.On("SearchCustomers", ctx, tenantID, query, 10).Return(customers, nil).Once()

		result, err := svc.SearchCustomers(ctx, query, 10)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		mockRepo.AssertExpectations(t)
	})

	t.Run("empty search query", func(t *testing.T) {
		tenantID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		result, err := svc.SearchCustomers(ctx, "", 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "search query cannot be empty")
	})

	t.Run("search query too short", func(t *testing.T) {
		tenantID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		result, err := svc.SearchCustomers(ctx, "a", 10)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "search query must be at least 2 characters")
	})
}

func TestCustomerService_ImportCustomers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockCustomerRepository)
	mockAudit := new(MockAuditService)
	mockNotif := new(MockNotificationService)

	svc := services.NewCustomerService(mockRepo, mockAudit, mockNotif)

	t.Run("successful bulk import", func(t *testing.T) {
		tenantID := uuid.New()
		userID := uuid.New()
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		customers := []*domain.Customer{
			createTestCustomer(),
			createTestCustomer(),
			createTestCustomer(),
		}

		for _, customer := range customers {
			customer.TenantID = tenantID
		}

		// Mock expectations
		for _, customer := range customers {
			mockRepo.On("GetCustomerByEmail", ctx, tenantID, customer.Email).Return(nil, errors.New("not found")).Once()
			mockRepo.On("CreateCustomer", ctx, mock.AnythingOfType("*domain.Customer")).Return(nil).Once()
		}
		mockAudit.On("LogActivity", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()

		results, err := svc.ImportCustomers(ctx, customers)

		assert.NoError(t, err)
		assert.Len(t, results, 3)
		for _, result := range results {
			assert.True(t, result.Success)
			assert.Empty(t, result.Error)
		}
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})

	t.Run("partial import with errors", func(t *testing.T) {
		tenantID := uuid.New()
		userID := uuid.New()
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		customers := []*domain.Customer{
			createTestCustomer(),
			createTestCustomer(),
			createTestCustomer(),
		}

		for _, customer := range customers {
			customer.TenantID = tenantID
		}

		// Mock expectations - first succeeds, second fails (duplicate), third succeeds
		mockRepo.On("GetCustomerByEmail", ctx, tenantID, customers[0].Email).Return(nil, errors.New("not found")).Once()
		mockRepo.On("CreateCustomer", ctx, mock.AnythingOfType("*domain.Customer")).Return(nil).Once()

		mockRepo.On("GetCustomerByEmail", ctx, tenantID, customers[1].Email).Return(customers[1], nil).Once()

		mockRepo.On("GetCustomerByEmail", ctx, tenantID, customers[2].Email).Return(nil, errors.New("not found")).Once()
		mockRepo.On("CreateCustomer", ctx, mock.AnythingOfType("*domain.Customer")).Return(nil).Once()

		mockAudit.On("LogActivity", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()

		results, err := svc.ImportCustomers(ctx, customers)

		assert.NoError(t, err)
		assert.Len(t, results, 3)
		assert.True(t, results[0].Success)
		assert.False(t, results[1].Success)
		assert.Contains(t, results[1].Error, "already exists")
		assert.True(t, results[2].Success)
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})
}

func TestCustomerService_ExportCustomers(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockCustomerRepository)
	mockAudit := new(MockAuditService)
	mockNotif := new(MockNotificationService)

	svc := services.NewCustomerService(mockRepo, mockAudit, mockNotif)

	t.Run("successful CSV export", func(t *testing.T) {
		tenantID := uuid.New()
		userID := uuid.New()
		ctx = context.WithValue(ctx, "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		customers := []*domain.Customer{
			createTestCustomer(),
			createTestCustomer(),
			createTestCustomer(),
		}

		filters := map[string]interface{}{
			"status": "active",
		}

		// Mock expectations
		mockRepo.On("ListCustomers", ctx, tenantID, filters, 0, 10000).Return(customers, 3, nil).Once()
		mockAudit.On("LogActivity", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()

		data, err := svc.ExportCustomers(ctx, "csv", filters)

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
		assert.Contains(t, string(data), "Name,Email,Phone")
		mockRepo.AssertExpectations(t)
		mockAudit.AssertExpectations(t)
	})

	t.Run("successful JSON export", func(t *testing.T) {
		tenantID := uuid.New()
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		customers := []*domain.Customer{
			createTestCustomer(),
			createTestCustomer(),
		}

		// Mock expectations
		mockRepo.On("ListCustomers", ctx, tenantID, map[string]interface{}{}, 0, 10000).Return(customers, 2, nil).Once()
		mockAudit.On("LogActivity", ctx, mock.AnythingOfType("*domain.AuditLog")).Return(nil).Once()

		data, err := svc.ExportCustomers(ctx, "json", map[string]interface{}{})

		assert.NoError(t, err)
		assert.NotEmpty(t, data)
		
		// Verify JSON structure
		var exported []map[string]interface{}
		require.NoError(t, json.Unmarshal(data, &exported))
		assert.Len(t, exported, 2)
		mockRepo.AssertExpectations(t)
	})
}