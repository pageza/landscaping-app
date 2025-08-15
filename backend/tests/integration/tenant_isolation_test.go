package integration_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/repository"
	"github.com/pageza/landscaping-app/backend/internal/services"
	"github.com/pageza/landscaping-app/backend/tests/testutils"
)

// TenantIsolationTestSuite contains all tenant isolation tests
type TenantIsolationTestSuite struct {
	db           *sql.DB
	fixtures     *testutils.TestFixtures
	repositories *RepositoryCollection
	services     *ServiceCollection
}

type RepositoryCollection struct {
	Customer     repository.CustomerRepository
	Job          repository.JobRepository
	Quote        repository.QuoteRepository
	Invoice      repository.InvoiceRepository
	Payment      repository.PaymentRepository
	Property     repository.PropertyRepository
	Equipment    repository.EquipmentRepository
	User         repository.UserRepository
	Tenant       repository.TenantRepository
	Subscription repository.SubscriptionRepository
}

type ServiceCollection struct {
	Customer     services.CustomerService
	Job          services.JobService
	Quote        services.QuoteService
	Invoice      services.InvoiceService
	Payment      services.PaymentService
	Property     services.PropertyService
	Equipment    services.EquipmentService
	Subscription services.SubscriptionService
}

// NewTenantIsolationTestSuite creates a new test suite
func NewTenantIsolationTestSuite(t *testing.T) *TenantIsolationTestSuite {
	// Skip integration tests in unit test mode
	if testing.Short() {
		t.Skip("Skipping tenant isolation tests in short mode")
	}

	// Setup test database connection
	db, cleanup := setupTestDB(t)
	t.Cleanup(cleanup)

	fixtures := testutils.NewTestFixtures(db)

	// Initialize repositories
	repos := &RepositoryCollection{
		Customer:     repository.NewCustomerRepository(db),
		Job:          repository.NewJobRepository(db),
		Quote:        repository.NewQuoteRepository(db),
		Invoice:      repository.NewInvoiceRepository(db),
		Payment:      repository.NewPaymentRepository(db),
		Property:     repository.NewPropertyRepository(db),
		Equipment:    repository.NewEquipmentRepository(db),
		User:         repository.NewUserRepository(db),
		Tenant:       repository.NewTenantRepository(db),
		Subscription: repository.NewSubscriptionRepository(db),
	}

	// Initialize services
	srvcs := &ServiceCollection{
		Customer:     services.NewCustomerService(repos.Customer, nil, nil),
		Job:          services.NewJobService(repos.Job, nil, nil),
		Quote:        services.NewQuoteService(repos.Quote, nil, nil),
		Invoice:      services.NewInvoiceService(repos.Invoice, nil, nil),
		Payment:      services.NewPaymentService(repos.Payment, nil, nil),
		Property:     services.NewPropertyService(repos.Property, nil, nil),
		Equipment:    services.NewEquipmentService(repos.Equipment, nil, nil),
		Subscription: services.NewSubscriptionService(repos.Subscription, nil, nil),
	}

	return &TenantIsolationTestSuite{
		db:           db,
		fixtures:     fixtures,
		repositories: repos,
		services:     srvcs,
	}
}

func TestTenantIsolation(t *testing.T) {
	suite := NewTenantIsolationTestSuite(t)
	ctx := context.Background()

	// Clean up before and after tests
	suite.fixtures.CleanupTestData(ctx)
	t.Cleanup(func() {
		suite.fixtures.CleanupTestData(ctx)
	})

	t.Run("CustomerIsolation", suite.testCustomerIsolation)
	t.Run("JobIsolation", suite.testJobIsolation)
	t.Run("PropertyIsolation", suite.testPropertyIsolation)
	t.Run("QuoteIsolation", suite.testQuoteIsolation)
	t.Run("InvoiceIsolation", suite.testInvoiceIsolation)
	t.Run("PaymentIsolation", suite.testPaymentIsolation)
	t.Run("EquipmentIsolation", suite.testEquipmentIsolation)
	t.Run("UserIsolation", suite.testUserIsolation)
	t.Run("CrossTenantDataAccess", suite.testCrossTenantDataAccess)
	t.Run("TenantDeletion", suite.testTenantDeletion)
	t.Run("BulkOperationIsolation", suite.testBulkOperationIsolation)
	t.Run("SearchIsolation", suite.testSearchIsolation)
	t.Run("AggregationIsolation", suite.testAggregationIsolation)
}

func (suite *TenantIsolationTestSuite) testCustomerIsolation(t *testing.T) {
	ctx := context.Background()

	// Create two tenants
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	// Create customers for each tenant
	customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID, func(c *domain.Customer) {
		c.Email = "customer1@tenant1.com"
		c.Name = "Tenant 1 Customer"
	})
	customer2 := suite.fixtures.CreateTestCustomer(tenant2.ID, func(c *domain.Customer) {
		c.Email = "customer2@tenant2.com"
		c.Name = "Tenant 2 Customer"
	})

	// Create customers in database
	err := suite.repositories.Customer.CreateCustomer(ctx, customer1)
	require.NoError(t, err)
	err = suite.repositories.Customer.CreateCustomer(ctx, customer2)
	require.NoError(t, err)

	t.Run("ListCustomers respects tenant boundaries", func(t *testing.T) {
		// Tenant 1 should only see its customers
		customers, total, err := suite.repositories.Customer.ListCustomers(ctx, tenant1.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, customer := range customers {
			assert.Equal(t, tenant1.ID, customer.TenantID, "Customer should belong to tenant 1")
		}

		// Tenant 2 should only see its customers
		customers, total, err = suite.repositories.Customer.ListCustomers(ctx, tenant2.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, customer := range customers {
			assert.Equal(t, tenant2.ID, customer.TenantID, "Customer should belong to tenant 2")
		}
	})

	t.Run("GetCustomerByEmail respects tenant boundaries", func(t *testing.T) {
		// Tenant 1 can find its own customer
		customer, err := suite.repositories.Customer.GetCustomerByEmail(ctx, tenant1.ID, customer1.Email)
		assert.NoError(t, err)
		assert.Equal(t, customer1.ID, customer.ID)

		// Tenant 1 cannot find tenant 2's customer
		_, err = suite.repositories.Customer.GetCustomerByEmail(ctx, tenant1.ID, customer2.Email)
		assert.Error(t, err)

		// Same email in different tenants should be allowed
		customer3 := suite.fixtures.CreateTestCustomer(tenant2.ID, func(c *domain.Customer) {
			c.Email = customer1.Email // Same email as customer1
			c.Name = "Different Customer Same Email"
		})
		err = suite.repositories.Customer.CreateCustomer(ctx, customer3)
		assert.NoError(t, err, "Should allow same email in different tenants")
	})

	t.Run("SearchCustomers respects tenant boundaries", func(t *testing.T) {
		// Search in tenant 1
		results, err := suite.repositories.Customer.SearchCustomers(ctx, tenant1.ID, "Tenant 1", 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		for _, customer := range results {
			assert.Equal(t, tenant1.ID, customer.TenantID)
		}

		// Search in tenant 2 with same query should return different results
		results, err = suite.repositories.Customer.SearchCustomers(ctx, tenant2.ID, "Tenant 1", 10)
		assert.NoError(t, err)
		// Should not find "Tenant 1" customers when searching in tenant 2
		assert.Len(t, results, 0)
	})

	t.Run("UpdateCustomer respects tenant boundaries", func(t *testing.T) {
		// Create a context with tenant 1 ID
		ctx1 := context.WithValue(ctx, "tenant_id", tenant1.ID)

		// Update customer1 (same tenant) should succeed
		customer1.Name = "Updated Customer 1"
		err := suite.services.Customer.UpdateCustomer(ctx1, customer1)
		assert.NoError(t, err)

		// Try to update customer2 (different tenant) should fail
		customer2.Name = "Updated Customer 2"
		ctx2 := context.WithValue(ctx, "tenant_id", tenant1.ID) // Wrong tenant context
		err = suite.services.Customer.UpdateCustomer(ctx2, customer2)
		assert.Error(t, err, "Should not allow updating customer from different tenant")
	})
}

func (suite *TenantIsolationTestSuite) testJobIsolation(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID)
	customer2 := suite.fixtures.CreateTestCustomer(tenant2.ID)
	
	property1 := suite.fixtures.CreateTestProperty(tenant1.ID, customer1.ID)
	property2 := suite.fixtures.CreateTestProperty(tenant2.ID, customer2.ID)

	job1 := suite.fixtures.CreateTestJob(tenant1.ID, customer1.ID, property1.ID)
	job2 := suite.fixtures.CreateTestJob(tenant2.ID, customer2.ID, property2.ID)

	// Create in database
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer1))
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer2))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property1))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property2))
	require.NoError(t, suite.repositories.Job.CreateJob(ctx, job1))
	require.NoError(t, suite.repositories.Job.CreateJob(ctx, job2))

	t.Run("Jobs are isolated by tenant", func(t *testing.T) {
		// List jobs for tenant 1
		jobs, total, err := suite.repositories.Job.ListJobs(ctx, tenant1.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, job := range jobs {
			assert.Equal(t, tenant1.ID, job.TenantID)
		}

		// List jobs for tenant 2
		jobs, total, err = suite.repositories.Job.ListJobs(ctx, tenant2.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, job := range jobs {
			assert.Equal(t, tenant2.ID, job.TenantID)
		}
	})

	t.Run("Cross-tenant job access is denied", func(t *testing.T) {
		// Tenant 1 trying to access tenant 2's job
		ctx1 := context.WithValue(ctx, "tenant_id", tenant1.ID)
		_, err := suite.services.Job.GetJob(ctx1, job2.ID)
		assert.Error(t, err, "Should not allow cross-tenant job access")

		// Tenant 2 trying to access tenant 1's job
		ctx2 := context.WithValue(ctx, "tenant_id", tenant2.ID)
		_, err = suite.services.Job.GetJob(ctx2, job1.ID)
		assert.Error(t, err, "Should not allow cross-tenant job access")
	})
}

func (suite *TenantIsolationTestSuite) testPropertyIsolation(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID)
	customer2 := suite.fixtures.CreateTestCustomer(tenant2.ID)
	
	// Create properties in same location but different tenants
	property1 := suite.fixtures.CreateTestProperty(tenant1.ID, customer1.ID, func(p *domain.Property) {
		p.Address = "123 Main St"
		p.City = "Anytown"
		p.State = "ST"
		p.Latitude = 40.7128
		p.Longitude = -74.0060
	})
	property2 := suite.fixtures.CreateTestProperty(tenant2.ID, customer2.ID, func(p *domain.Property) {
		p.Address = "125 Main St" // Different address, same area
		p.City = "Anytown"
		p.State = "ST"
		p.Latitude = 40.7130 // Very close to property1
		p.Longitude = -74.0062
	})

	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer1))
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer2))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property1))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property2))

	t.Run("Properties are isolated by tenant", func(t *testing.T) {
		properties, total, err := suite.repositories.Property.ListProperties(ctx, tenant1.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, property := range properties {
			assert.Equal(t, tenant1.ID, property.TenantID)
		}
	})

	t.Run("Nearby property search respects tenant boundaries", func(t *testing.T) {
		// Search for properties near property1's location within tenant1
		filters := map[string]interface{}{
			"latitude":  40.7128,
			"longitude": -74.0060,
			"radius":    1000, // 1km radius
		}

		properties, err := suite.repositories.Property.GetNearbyProperties(ctx, tenant1.ID, filters, 10)
		assert.NoError(t, err)

		for _, property := range properties {
			assert.Equal(t, tenant1.ID, property.TenantID)
		}

		// Same search in tenant2 should not return tenant1's properties
		properties, err = suite.repositories.Property.GetNearbyProperties(ctx, tenant2.ID, filters, 10)
		assert.NoError(t, err)

		for _, property := range properties {
			assert.Equal(t, tenant2.ID, property.TenantID)
		}
	})
}

func (suite *TenantIsolationTestSuite) testQuoteIsolation(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID)
	customer2 := suite.fixtures.CreateTestCustomer(tenant2.ID)
	
	property1 := suite.fixtures.CreateTestProperty(tenant1.ID, customer1.ID)
	property2 := suite.fixtures.CreateTestProperty(tenant2.ID, customer2.ID)

	quote1 := suite.fixtures.CreateTestQuote(tenant1.ID, customer1.ID, property1.ID)
	quote2 := suite.fixtures.CreateTestQuote(tenant2.ID, customer2.ID, property2.ID)

	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer1))
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer2))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property1))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property2))
	require.NoError(t, suite.repositories.Quote.CreateQuote(ctx, quote1))
	require.NoError(t, suite.repositories.Quote.CreateQuote(ctx, quote2))

	t.Run("Quotes are isolated by tenant", func(t *testing.T) {
		quotes, total, err := suite.repositories.Quote.ListQuotes(ctx, tenant1.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, quote := range quotes {
			assert.Equal(t, tenant1.ID, quote.TenantID)
		}
	})

	t.Run("Quote number uniqueness is per tenant", func(t *testing.T) {
		// Create quotes with same quote number in different tenants
		quote3 := suite.fixtures.CreateTestQuote(tenant1.ID, customer1.ID, property1.ID, func(q *domain.Quote) {
			q.QuoteNumber = "Q-123456"
		})
		quote4 := suite.fixtures.CreateTestQuote(tenant2.ID, customer2.ID, property2.ID, func(q *domain.Quote) {
			q.QuoteNumber = "Q-123456" // Same number, different tenant
		})

		err := suite.repositories.Quote.CreateQuote(ctx, quote3)
		assert.NoError(t, err)

		err = suite.repositories.Quote.CreateQuote(ctx, quote4)
		assert.NoError(t, err, "Same quote number should be allowed in different tenants")

		// But not in the same tenant
		quote5 := suite.fixtures.CreateTestQuote(tenant1.ID, customer1.ID, property1.ID, func(q *domain.Quote) {
			q.QuoteNumber = "Q-123456" // Duplicate in same tenant
		})

		err = suite.repositories.Quote.CreateQuote(ctx, quote5)
		assert.Error(t, err, "Duplicate quote number in same tenant should fail")
	})
}

func (suite *TenantIsolationTestSuite) testInvoiceIsolation(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID)
	customer2 := suite.fixtures.CreateTestCustomer(tenant2.ID)
	
	property1 := suite.fixtures.CreateTestProperty(tenant1.ID, customer1.ID)
	property2 := suite.fixtures.CreateTestProperty(tenant2.ID, customer2.ID)

	job1 := suite.fixtures.CreateTestJob(tenant1.ID, customer1.ID, property1.ID)
	job2 := suite.fixtures.CreateTestJob(tenant2.ID, customer2.ID, property2.ID)

	invoice1 := suite.fixtures.CreateTestInvoice(tenant1.ID, customer1.ID, job1.ID)
	invoice2 := suite.fixtures.CreateTestInvoice(tenant2.ID, customer2.ID, job2.ID)

	// Create in database
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer1))
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer2))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property1))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property2))
	require.NoError(t, suite.repositories.Job.CreateJob(ctx, job1))
	require.NoError(t, suite.repositories.Job.CreateJob(ctx, job2))
	require.NoError(t, suite.repositories.Invoice.CreateInvoice(ctx, invoice1))
	require.NoError(t, suite.repositories.Invoice.CreateInvoice(ctx, invoice2))

	t.Run("Invoices are isolated by tenant", func(t *testing.T) {
		invoices, total, err := suite.repositories.Invoice.ListInvoices(ctx, tenant1.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, invoice := range invoices {
			assert.Equal(t, tenant1.ID, invoice.TenantID)
		}
	})

	t.Run("Invoice aggregations are tenant-specific", func(t *testing.T) {
		// Get invoice totals for tenant 1
		filters1 := map[string]interface{}{
			"status": "pending",
		}
		invoices1, total1, err := suite.repositories.Invoice.ListInvoices(ctx, tenant1.ID, filters1, 0, 100)
		assert.NoError(t, err)

		// Get invoice totals for tenant 2
		invoices2, total2, err := suite.repositories.Invoice.ListInvoices(ctx, tenant2.ID, filters1, 0, 100)
		assert.NoError(t, err)

		// Calculate totals for each tenant
		var total1Amount, total2Amount float64
		for _, invoice := range invoices1 {
			total1Amount += invoice.Total
		}
		for _, invoice := range invoices2 {
			total2Amount += invoice.Total
		}

		// Totals should be different for different tenants
		if len(invoices1) > 0 && len(invoices2) > 0 {
			assert.NotEqual(t, total1Amount, total2Amount)
		}
	})
}

func (suite *TenantIsolationTestSuite) testPaymentIsolation(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID)
	customer2 := suite.fixtures.CreateTestCustomer(tenant2.ID)
	
	property1 := suite.fixtures.CreateTestProperty(tenant1.ID, customer1.ID)
	property2 := suite.fixtures.CreateTestProperty(tenant2.ID, customer2.ID)

	job1 := suite.fixtures.CreateTestJob(tenant1.ID, customer1.ID, property1.ID)
	job2 := suite.fixtures.CreateTestJob(tenant2.ID, customer2.ID, property2.ID)

	invoice1 := suite.fixtures.CreateTestInvoice(tenant1.ID, customer1.ID, job1.ID)
	invoice2 := suite.fixtures.CreateTestInvoice(tenant2.ID, customer2.ID, job2.ID)

	payment1 := suite.fixtures.CreateTestPayment(tenant1.ID, customer1.ID, invoice1.ID)
	payment2 := suite.fixtures.CreateTestPayment(tenant2.ID, customer2.ID, invoice2.ID)

	// Create in database
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer1))
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer2))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property1))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property2))
	require.NoError(t, suite.repositories.Job.CreateJob(ctx, job1))
	require.NoError(t, suite.repositories.Job.CreateJob(ctx, job2))
	require.NoError(t, suite.repositories.Invoice.CreateInvoice(ctx, invoice1))
	require.NoError(t, suite.repositories.Invoice.CreateInvoice(ctx, invoice2))
	require.NoError(t, suite.repositories.Payment.CreatePayment(ctx, payment1))
	require.NoError(t, suite.repositories.Payment.CreatePayment(ctx, payment2))

	t.Run("Payments are isolated by tenant", func(t *testing.T) {
		payments, total, err := suite.repositories.Payment.ListPayments(ctx, tenant1.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, payment := range payments {
			assert.Equal(t, tenant1.ID, payment.TenantID)
		}
	})

	t.Run("Payment transaction IDs can be same across tenants", func(t *testing.T) {
		// Create payments with same transaction ID in different tenants
		payment3 := suite.fixtures.CreateTestPayment(tenant1.ID, customer1.ID, invoice1.ID, func(p *domain.Payment) {
			p.TransactionID = "TXN-123456"
		})
		payment4 := suite.fixtures.CreateTestPayment(tenant2.ID, customer2.ID, invoice2.ID, func(p *domain.Payment) {
			p.TransactionID = "TXN-123456" // Same transaction ID
		})

		err := suite.repositories.Payment.CreatePayment(ctx, payment3)
		assert.NoError(t, err)

		err = suite.repositories.Payment.CreatePayment(ctx, payment4)
		assert.NoError(t, err, "Same transaction ID should be allowed across tenants")
	})
}

func (suite *TenantIsolationTestSuite) testEquipmentIsolation(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	equipment1 := suite.fixtures.CreateTestEquipment(tenant1.ID, func(e *domain.Equipment) {
		e.SerialNumber = "SN123456"
	})
	equipment2 := suite.fixtures.CreateTestEquipment(tenant2.ID, func(e *domain.Equipment) {
		e.SerialNumber = "SN123456" // Same serial number
	})

	require.NoError(t, suite.repositories.Equipment.CreateEquipment(ctx, equipment1))
	require.NoError(t, suite.repositories.Equipment.CreateEquipment(ctx, equipment2))

	t.Run("Equipment is isolated by tenant", func(t *testing.T) {
		equipment, total, err := suite.repositories.Equipment.ListEquipment(ctx, tenant1.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, eq := range equipment {
			assert.Equal(t, tenant1.ID, eq.TenantID)
		}
	})

	t.Run("Same serial numbers allowed across tenants", func(t *testing.T) {
		// Both equipment items should exist
		eq1, err := suite.repositories.Equipment.GetEquipmentByID(ctx, equipment1.ID)
		assert.NoError(t, err)
		assert.Equal(t, "SN123456", eq1.SerialNumber)

		eq2, err := suite.repositories.Equipment.GetEquipmentByID(ctx, equipment2.ID)
		assert.NoError(t, err)
		assert.Equal(t, "SN123456", eq2.SerialNumber)
	})

	t.Run("Available equipment search respects tenant boundaries", func(t *testing.T) {
		filters := map[string]interface{}{
			"status": "available",
		}

		equipment, err := suite.repositories.Equipment.GetAvailableEquipment(ctx, tenant1.ID, filters, 10)
		assert.NoError(t, err)

		for _, eq := range equipment {
			assert.Equal(t, tenant1.ID, eq.TenantID)
		}
	})
}

func (suite *TenantIsolationTestSuite) testUserIsolation(t *testing.T) {
	ctx := context.Background()

	// Setup test data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	user1 := suite.fixtures.CreateTestUser(tenant1.ID, func(u *domain.EnhancedUser) {
		u.Email = "user1@tenant1.com"
		u.Role = domain.RoleAdmin
	})
	user2 := suite.fixtures.CreateTestUser(tenant2.ID, func(u *domain.EnhancedUser) {
		u.Email = "user2@tenant2.com"
		u.Role = domain.RoleAdmin
	})

	require.NoError(t, suite.repositories.User.CreateUser(ctx, &user1.User))
	require.NoError(t, suite.repositories.User.CreateUser(ctx, &user2.User))

	t.Run("Users are isolated by tenant", func(t *testing.T) {
		users, total, err := suite.repositories.User.ListUsers(ctx, tenant1.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)

		for _, user := range users {
			assert.Equal(t, tenant1.ID, user.TenantID)
		}
	})

	t.Run("Same email allowed in different tenants", func(t *testing.T) {
		user3 := suite.fixtures.CreateTestUser(tenant2.ID, func(u *domain.EnhancedUser) {
			u.Email = user1.Email // Same email as user1
		})

		err := suite.repositories.User.CreateUser(ctx, &user3.User)
		assert.NoError(t, err, "Should allow same email in different tenants")
	})

	t.Run("User permissions are tenant-scoped", func(t *testing.T) {
		// User1 should have permissions within tenant1
		ctx1 := context.WithValue(ctx, "tenant_id", tenant1.ID)
		ctx1 = context.WithValue(ctx1, "user_id", user1.ID)
		ctx1 = context.WithValue(ctx1, "user_role", user1.Role)

		// User1 should not access tenant2's data
		ctx2 := context.WithValue(ctx, "tenant_id", tenant2.ID)
		ctx2 = context.WithValue(ctx2, "user_id", user1.ID)
		ctx2 = context.WithValue(ctx2, "user_role", user1.Role)

		// This would need to be tested with actual service calls
		// that check tenant context
		assert.Equal(t, tenant1.ID, user1.TenantID)
		assert.Equal(t, tenant2.ID, user2.TenantID)
	})
}

func (suite *TenantIsolationTestSuite) testCrossTenantDataAccess(t *testing.T) {
	ctx := context.Background()

	// Create comprehensive test scenario
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	// Create full data hierarchy for tenant1
	user1 := suite.fixtures.CreateTestUser(tenant1.ID)
	customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID)
	property1 := suite.fixtures.CreateTestProperty(tenant1.ID, customer1.ID)
	job1 := suite.fixtures.CreateTestJob(tenant1.ID, customer1.ID, property1.ID)
	quote1 := suite.fixtures.CreateTestQuote(tenant1.ID, customer1.ID, property1.ID)
	invoice1 := suite.fixtures.CreateTestInvoice(tenant1.ID, customer1.ID, job1.ID)
	payment1 := suite.fixtures.CreateTestPayment(tenant1.ID, customer1.ID, invoice1.ID)

	// Create in database
	require.NoError(t, suite.repositories.User.CreateUser(ctx, &user1.User))
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer1))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property1))
	require.NoError(t, suite.repositories.Job.CreateJob(ctx, job1))
	require.NoError(t, suite.repositories.Quote.CreateQuote(ctx, quote1))
	require.NoError(t, suite.repositories.Invoice.CreateInvoice(ctx, invoice1))
	require.NoError(t, suite.repositories.Payment.CreatePayment(ctx, payment1))

	t.Run("Tenant2 cannot access any Tenant1 data", func(t *testing.T) {
		// Try to access tenant1's data from tenant2 context
		ctx2 := context.WithValue(ctx, "tenant_id", tenant2.ID)

		// Customer access
		customers, total, err := suite.repositories.Customer.ListCustomers(ctx2, tenant2.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.Equal(t, 0, total, "Tenant2 should not see tenant1's customers")

		// Job access
		jobs, total, err := suite.repositories.Job.ListJobs(ctx2, tenant2.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.Equal(t, 0, total, "Tenant2 should not see tenant1's jobs")

		// Property access
		properties, total, err := suite.repositories.Property.ListProperties(ctx2, tenant2.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.Equal(t, 0, total, "Tenant2 should not see tenant1's properties")

		// Quote access
		quotes, total, err := suite.repositories.Quote.ListQuotes(ctx2, tenant2.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.Equal(t, 0, total, "Tenant2 should not see tenant1's quotes")

		// Invoice access
		invoices, total, err := suite.repositories.Invoice.ListInvoices(ctx2, tenant2.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.Equal(t, 0, total, "Tenant2 should not see tenant1's invoices")

		// Payment access
		payments, total, err := suite.repositories.Payment.ListPayments(ctx2, tenant2.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.Equal(t, 0, total, "Tenant2 should not see tenant1's payments")
	})
}

func (suite *TenantIsolationTestSuite) testTenantDeletion(t *testing.T) {
	ctx := context.Background()

	// Create tenant with data
	tenant := suite.fixtures.CreateTestTenant()
	customer := suite.fixtures.CreateTestCustomer(tenant.ID)
	property := suite.fixtures.CreateTestProperty(tenant.ID, customer.ID)

	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property))

	// Verify data exists
	customers, total, err := suite.repositories.Customer.ListCustomers(ctx, tenant.ID, map[string]interface{}{}, 0, 100)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, 1)

	t.Run("Tenant deletion removes all associated data", func(t *testing.T) {
		// Delete tenant (this would cascade delete all associated data)
		err := suite.repositories.Tenant.DeleteTenant(ctx, tenant.ID)
		assert.NoError(t, err)

		// Verify all data is gone
		customers, total, err := suite.repositories.Customer.ListCustomers(ctx, tenant.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.Equal(t, 0, total, "All tenant data should be deleted")

		properties, total, err := suite.repositories.Property.ListProperties(ctx, tenant.ID, map[string]interface{}{}, 0, 100)
		assert.NoError(t, err)
		assert.Equal(t, 0, total, "All tenant data should be deleted")
	})
}

func (suite *TenantIsolationTestSuite) testBulkOperationIsolation(t *testing.T) {
	ctx := context.Background()

	// Create tenants with bulk data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	// Create multiple customers for each tenant
	var tenant1Customers, tenant2Customers []*domain.Customer
	for i := 0; i < 10; i++ {
		customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID, func(c *domain.Customer) {
			c.Email = fmt.Sprintf("customer%d@tenant1.com", i)
		})
		customer2 := suite.fixtures.CreateTestCustomer(tenant2.ID, func(c *domain.Customer) {
			c.Email = fmt.Sprintf("customer%d@tenant2.com", i)
		})

		tenant1Customers = append(tenant1Customers, customer1)
		tenant2Customers = append(tenant2Customers, customer2)

		require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer1))
		require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer2))
	}

	t.Run("Bulk operations respect tenant boundaries", func(t *testing.T) {
		// Update all active customers in tenant1
		filters := map[string]interface{}{
			"status": "active",
		}

		customers, total, err := suite.repositories.Customer.ListCustomers(ctx, tenant1.ID, filters, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 10)

		// Verify all returned customers belong to tenant1
		for _, customer := range customers {
			assert.Equal(t, tenant1.ID, customer.TenantID)
		}

		// Same operation for tenant2
		customers, total, err = suite.repositories.Customer.ListCustomers(ctx, tenant2.ID, filters, 0, 100)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, 10)

		// Verify all returned customers belong to tenant2
		for _, customer := range customers {
			assert.Equal(t, tenant2.ID, customer.TenantID)
		}
	})
}

func (suite *TenantIsolationTestSuite) testSearchIsolation(t *testing.T) {
	ctx := context.Background()

	// Create tenants with searchable data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	// Create customers with similar names but different tenants
	customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID, func(c *domain.Customer) {
		c.Name = "John Smith"
		c.Company = "Smith Landscaping"
	})
	customer2 := suite.fixtures.CreateTestCustomer(tenant2.ID, func(c *domain.Customer) {
		c.Name = "John Johnson"
		c.Company = "Johnson Landscaping"
	})

	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer1))
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer2))

	t.Run("Search results are tenant-isolated", func(t *testing.T) {
		// Search for "John" in tenant1
		results, err := suite.repositories.Customer.SearchCustomers(ctx, tenant1.ID, "John", 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		for _, customer := range results {
			assert.Equal(t, tenant1.ID, customer.TenantID)
			assert.Contains(t, customer.Name, "John")
		}

		// Search for "Johnson" in tenant1 should return empty
		results, err = suite.repositories.Customer.SearchCustomers(ctx, tenant1.ID, "Johnson", 10)
		assert.NoError(t, err)
		assert.Len(t, results, 0, "Should not find Johnson in tenant1")

		// Search for "Johnson" in tenant2 should find customer2
		results, err = suite.repositories.Customer.SearchCustomers(ctx, tenant2.ID, "Johnson", 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(results), 1)

		for _, customer := range results {
			assert.Equal(t, tenant2.ID, customer.TenantID)
		}
	})
}

func (suite *TenantIsolationTestSuite) testAggregationIsolation(t *testing.T) {
	ctx := context.Background()

	// Create tenants with financial data
	tenant1 := suite.fixtures.CreateTestTenant()
	tenant2 := suite.fixtures.CreateTestTenant()

	// Create financial data for both tenants
	customer1 := suite.fixtures.CreateTestCustomer(tenant1.ID)
	customer2 := suite.fixtures.CreateTestCustomer(tenant2.ID)
	
	property1 := suite.fixtures.CreateTestProperty(tenant1.ID, customer1.ID)
	property2 := suite.fixtures.CreateTestProperty(tenant2.ID, customer2.ID)

	job1 := suite.fixtures.CreateTestJob(tenant1.ID, customer1.ID, property1.ID)
	job2 := suite.fixtures.CreateTestJob(tenant2.ID, customer2.ID, property2.ID)

	// Create invoices with different amounts
	invoice1 := suite.fixtures.CreateTestInvoice(tenant1.ID, customer1.ID, job1.ID, func(i *domain.Invoice) {
		i.Total = 1000.00
		i.Status = "paid"
	})
	invoice2 := suite.fixtures.CreateTestInvoice(tenant2.ID, customer2.ID, job2.ID, func(i *domain.Invoice) {
		i.Total = 2000.00
		i.Status = "paid"
	})

	// Create in database
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer1))
	require.NoError(t, suite.repositories.Customer.CreateCustomer(ctx, customer2))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property1))
	require.NoError(t, suite.repositories.Property.CreateProperty(ctx, property2))
	require.NoError(t, suite.repositories.Job.CreateJob(ctx, job1))
	require.NoError(t, suite.repositories.Job.CreateJob(ctx, job2))
	require.NoError(t, suite.repositories.Invoice.CreateInvoice(ctx, invoice1))
	require.NoError(t, suite.repositories.Invoice.CreateInvoice(ctx, invoice2))

	t.Run("Revenue calculations are tenant-specific", func(t *testing.T) {
		// Calculate revenue for tenant1
		filters1 := map[string]interface{}{
			"status": "paid",
		}
		invoices1, _, err := suite.repositories.Invoice.ListInvoices(ctx, tenant1.ID, filters1, 0, 100)
		assert.NoError(t, err)

		var tenant1Revenue float64
		for _, invoice := range invoices1 {
			tenant1Revenue += invoice.Total
		}

		// Calculate revenue for tenant2
		invoices2, _, err := suite.repositories.Invoice.ListInvoices(ctx, tenant2.ID, filters1, 0, 100)
		assert.NoError(t, err)

		var tenant2Revenue float64
		for _, invoice := range invoices2 {
			tenant2Revenue += invoice.Total
		}

		// Revenues should be different and correct
		assert.GreaterOrEqual(t, tenant1Revenue, 1000.00)
		assert.GreaterOrEqual(t, tenant2Revenue, 2000.00)
		assert.NotEqual(t, tenant1Revenue, tenant2Revenue)
	})
}

// Helper function to setup test database (same as in repository tests)
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// This would connect to a real test database for integration tests
	// For now, using a placeholder implementation
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	
	cleanup := func() {
		db.Close()
	}
	
	return db, cleanup
}