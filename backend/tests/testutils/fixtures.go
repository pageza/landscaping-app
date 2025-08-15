package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/google/uuid"
	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// TestFixtures provides test data generation utilities
type TestFixtures struct {
	db *sql.DB
}

// NewTestFixtures creates a new test fixtures instance
func NewTestFixtures(db *sql.DB) *TestFixtures {
	return &TestFixtures{db: db}
}

// CreateTestTenant creates a test tenant with optional customization
func (tf *TestFixtures) CreateTestTenant(opts ...func(*domain.Tenant)) *domain.Tenant {
	tenant := &domain.Tenant{
		ID:             uuid.New(),
		Name:           faker.Company(),
		Domain:         faker.DomainName(),
		Status:         "active",
		SubscriptionID: uuid.New(),
		Settings: map[string]interface{}{
			"timezone":       "America/New_York",
			"currency":       "USD",
			"business_hours": "9AM-5PM",
		},
		Metadata: map[string]interface{}{
			"industry": "landscaping",
			"size":     "medium",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(tenant)
	}

	return tenant
}

// CreateTestUser creates a test user with optional customization
func (tf *TestFixtures) CreateTestUser(tenantID uuid.UUID, opts ...func(*domain.EnhancedUser)) *domain.EnhancedUser {
	user := &domain.EnhancedUser{
		User: domain.User{
			ID:           uuid.New(),
			TenantID:     tenantID,
			Email:        faker.Email(),
			PasswordHash: "$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY.MzkwLlYXqsIa", // password123
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			Role:         domain.RoleUser,
			Status:       "active",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		},
		Permissions: []string{
			domain.PermissionCustomerManage,
			domain.PermissionJobManage,
			domain.PermissionQuoteManage,
		},
		TOTPSecret: nil,
	}

	for _, opt := range opts {
		opt(user)
	}

	return user
}

// CreateTestCustomer creates a test customer with optional customization
func (tf *TestFixtures) CreateTestCustomer(tenantID uuid.UUID, opts ...func(*domain.Customer)) *domain.Customer {
	customer := &domain.Customer{
		ID:       uuid.New(),
		TenantID: tenantID,
		Email:    faker.Email(),
		Phone:    faker.Phonenumber(),
		Name:     faker.Name(),
		Company:  faker.Company(),
		Address:  faker.Address().Address,
		City:     faker.Address().City,
		State:    faker.Address().State,
		ZipCode:  faker.Address().ZipCode,
		Country:  faker.Address().Country,
		Status:   "active",
		Tags:     []string{"residential", "premium"},
		Metadata: map[string]interface{}{
			"source":         "website",
			"preferred_day":  "weekends",
			"payment_method": "credit_card",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(customer)
	}

	return customer
}

// CreateTestProperty creates a test property with optional customization
func (tf *TestFixtures) CreateTestProperty(tenantID, customerID uuid.UUID, opts ...func(*domain.Property)) *domain.Property {
	property := &domain.Property{
		ID:         uuid.New(),
		TenantID:   tenantID,
		CustomerID: customerID,
		Address:    faker.Address().Address,
		City:       faker.Address().City,
		State:      faker.Address().State,
		ZipCode:    faker.Address().ZipCode,
		Country:    faker.Address().Country,
		Latitude:   faker.Address().Latitude,
		Longitude:  faker.Address().Longitude,
		Type:       "residential",
		Size:       faker.RandomInt(1000, 10000),
		SizeUnit:   "sqft",
		Notes:      faker.Sentence(),
		Metadata: map[string]interface{}{
			"lawn_type":      "bermuda",
			"irrigation":     true,
			"access_code":    "1234",
			"special_notes":  "Beware of dog",
			"last_service":   time.Now().AddDate(0, -1, 0),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(property)
	}

	return property
}

// CreateTestJob creates a test job with optional customization
func (tf *TestFixtures) CreateTestJob(tenantID, customerID, propertyID uuid.UUID, opts ...func(*domain.Job)) *domain.Job {
	scheduledDate := time.Now().AddDate(0, 0, 7)
	job := &domain.Job{
		ID:           uuid.New(),
		TenantID:     tenantID,
		CustomerID:   customerID,
		PropertyID:   propertyID,
		Title:        "Lawn Maintenance",
		Description:  faker.Paragraph(),
		Status:       domain.JobStatusScheduled,
		Priority:     "medium",
		ScheduledDate: &scheduledDate,
		EstimatedDuration: 120, // minutes
		ActualDuration:    0,
		AssignedTo:   []uuid.UUID{uuid.New()},
		Services:     []string{"mowing", "edging", "blowing"},
		Notes:        faker.Sentence(),
		Metadata: map[string]interface{}{
			"weather_dependent": true,
			"equipment_needed":  []string{"mower", "edger", "blower"},
			"crew_size":         2,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(job)
	}

	return job
}

// CreateTestQuote creates a test quote with optional customization
func (tf *TestFixtures) CreateTestQuote(tenantID, customerID, propertyID uuid.UUID, opts ...func(*domain.Quote)) *domain.Quote {
	validUntil := time.Now().AddDate(0, 0, 30)
	quote := &domain.Quote{
		ID:         uuid.New(),
		TenantID:   tenantID,
		CustomerID: customerID,
		PropertyID: propertyID,
		QuoteNumber: fmt.Sprintf("Q-%06d", faker.RandomInt(1, 999999)),
		Status:     "draft",
		ValidUntil: validUntil,
		Items: []domain.QuoteItem{
			{
				ID:          uuid.New(),
				Description: "Lawn Mowing Service",
				Quantity:    1,
				UnitPrice:   75.00,
				TotalPrice:  75.00,
			},
			{
				ID:          uuid.New(),
				Description: "Hedge Trimming",
				Quantity:    1,
				UnitPrice:   50.00,
				TotalPrice:  50.00,
			},
		},
		Subtotal:  125.00,
		Tax:       12.50,
		Discount:  0,
		Total:     137.50,
		Notes:     faker.Sentence(),
		Terms:     "Payment due within 30 days",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(quote)
	}

	return quote
}

// CreateTestInvoice creates a test invoice with optional customization
func (tf *TestFixtures) CreateTestInvoice(tenantID, customerID, jobID uuid.UUID, opts ...func(*domain.Invoice)) *domain.Invoice {
	dueDate := time.Now().AddDate(0, 0, 30)
	invoice := &domain.Invoice{
		ID:            uuid.New(),
		TenantID:      tenantID,
		CustomerID:    customerID,
		JobID:         &jobID,
		InvoiceNumber: fmt.Sprintf("INV-%06d", faker.RandomInt(1, 999999)),
		Status:        "pending",
		DueDate:       dueDate,
		Items: []domain.InvoiceItem{
			{
				ID:          uuid.New(),
				Description: "Lawn Maintenance Service",
				Quantity:    1,
				UnitPrice:   75.00,
				TotalPrice:  75.00,
			},
			{
				ID:          uuid.New(),
				Description: "Fertilizer Application",
				Quantity:    1,
				UnitPrice:   45.00,
				TotalPrice:  45.00,
			},
		},
		Subtotal:  120.00,
		Tax:       12.00,
		Discount:  0,
		Total:     132.00,
		PaidAmount: 0,
		Balance:   132.00,
		Notes:     faker.Sentence(),
		Terms:     "Net 30",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(invoice)
	}

	return invoice
}

// CreateTestEquipment creates test equipment with optional customization
func (tf *TestFixtures) CreateTestEquipment(tenantID uuid.UUID, opts ...func(*domain.Equipment)) *domain.Equipment {
	purchaseDate := time.Now().AddDate(-2, 0, 0)
	lastMaintenanceDate := time.Now().AddDate(0, -1, 0)
	nextMaintenanceDate := time.Now().AddDate(0, 1, 0)
	
	equipment := &domain.Equipment{
		ID:                   uuid.New(),
		TenantID:             tenantID,
		Name:                 "John Deere Z335E",
		Type:                 "Zero-Turn Mower",
		SerialNumber:         faker.UUIDDigit(),
		PurchaseDate:         &purchaseDate,
		PurchasePrice:        3500.00,
		CurrentValue:         2800.00,
		Status:               "available",
		LastMaintenanceDate:  &lastMaintenanceDate,
		NextMaintenanceDate:  &nextMaintenanceDate,
		MaintenanceInterval:  90, // days
		UsageHours:           450,
		Notes:                faker.Sentence(),
		Metadata: map[string]interface{}{
			"fuel_type":     "gasoline",
			"engine_hours":  450,
			"deck_size":     "42 inches",
			"warranty_expires": time.Now().AddDate(1, 0, 0),
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(equipment)
	}

	return equipment
}

// CreateTestCrew creates a test crew with optional customization
func (tf *TestFixtures) CreateTestCrew(tenantID uuid.UUID, memberIDs []uuid.UUID, opts ...func(*domain.Crew)) *domain.Crew {
	crew := &domain.Crew{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     fmt.Sprintf("Crew %s", faker.Word()),
		LeaderID: memberIDs[0],
		Members:  memberIDs,
		Status:   "active",
		Skills:   []string{"mowing", "trimming", "landscaping", "irrigation"},
		Metadata: map[string]interface{}{
			"vehicle":     "Ford F-150",
			"equipment":   []string{"mower", "trimmer", "blower"},
			"service_area": []string{"North Zone", "East Zone"},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(crew)
	}

	return crew
}

// CreateTestPayment creates a test payment with optional customization
func (tf *TestFixtures) CreateTestPayment(tenantID, customerID, invoiceID uuid.UUID, opts ...func(*domain.Payment)) *domain.Payment {
	payment := &domain.Payment{
		ID:                uuid.New(),
		TenantID:          tenantID,
		CustomerID:        customerID,
		InvoiceID:         &invoiceID,
		Amount:            132.00,
		Currency:          "USD",
		PaymentMethod:     "credit_card",
		Status:            "completed",
		TransactionID:     faker.UUIDDigit(),
		ProcessorResponse: map[string]interface{}{
			"authorization_code": faker.UUIDDigit(),
			"last_four":          "4242",
			"card_brand":         "visa",
		},
		Notes:     "Payment received via online portal",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(payment)
	}

	return payment
}

// CreateTestSubscription creates a test subscription with optional customization
func (tf *TestFixtures) CreateTestSubscription(tenantID uuid.UUID, opts ...func(*domain.Subscription)) *domain.Subscription {
	startDate := time.Now().AddDate(0, -6, 0)
	endDate := time.Now().AddDate(0, 6, 0)
	nextBillingDate := time.Now().AddDate(0, 1, 0)
	
	subscription := &domain.Subscription{
		ID:              uuid.New(),
		TenantID:        tenantID,
		PlanID:          "pro_monthly",
		Status:          "active",
		StartDate:       startDate,
		EndDate:         &endDate,
		NextBillingDate: &nextBillingDate,
		Amount:          199.00,
		Currency:        "USD",
		BillingInterval: "monthly",
		Features: map[string]interface{}{
			"max_users":      25,
			"max_customers":  1000,
			"api_access":     true,
			"custom_reports": true,
			"integrations":   []string{"quickbooks", "stripe", "twilio"},
		},
		Metadata: map[string]interface{}{
			"stripe_subscription_id": faker.UUIDDigit(),
			"discount_code":          "LAUNCH20",
			"referral_source":        "partner",
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	for _, opt := range opts {
		opt(subscription)
	}

	return subscription
}

// SeedTestData populates the database with test data
func (tf *TestFixtures) SeedTestData(ctx context.Context, numTenants int) error {
	for i := 0; i < numTenants; i++ {
		tenant := tf.CreateTestTenant()
		
		// Create users for the tenant
		owner := tf.CreateTestUser(tenant.ID, func(u *domain.EnhancedUser) {
			u.Role = domain.RoleOwner
			u.Permissions = []string{"*"}
		})
		
		admin := tf.CreateTestUser(tenant.ID, func(u *domain.EnhancedUser) {
			u.Role = domain.RoleAdmin
		})
		
		users := []*domain.EnhancedUser{owner, admin}
		for j := 0; j < 3; j++ {
			users = append(users, tf.CreateTestUser(tenant.ID))
		}
		
		// Create customers
		for j := 0; j < 10; j++ {
			customer := tf.CreateTestCustomer(tenant.ID)
			
			// Create properties for each customer
			for k := 0; k < 2; k++ {
				property := tf.CreateTestProperty(tenant.ID, customer.ID)
				
				// Create jobs for each property
				for l := 0; l < 3; l++ {
					job := tf.CreateTestJob(tenant.ID, customer.ID, property.ID)
					
					// Create invoice for completed jobs
					if job.Status == domain.JobStatusCompleted {
						invoice := tf.CreateTestInvoice(tenant.ID, customer.ID, job.ID)
						
						// Create payment for some invoices
						if faker.RandomInt(0, 1) == 1 {
							tf.CreateTestPayment(tenant.ID, customer.ID, invoice.ID)
						}
					}
				}
				
				// Create quotes
				for l := 0; l < 2; l++ {
					tf.CreateTestQuote(tenant.ID, customer.ID, property.ID)
				}
			}
		}
		
		// Create equipment
		for j := 0; j < 5; j++ {
			tf.CreateTestEquipment(tenant.ID)
		}
		
		// Create crews
		for j := 0; j < 2; j++ {
			memberIDs := []uuid.UUID{users[2].ID, users[3].ID}
			tf.CreateTestCrew(tenant.ID, memberIDs)
		}
		
		// Create subscription
		tf.CreateTestSubscription(tenant.ID)
	}
	
	return nil
}

// CleanupTestData removes all test data from the database
func (tf *TestFixtures) CleanupTestData(ctx context.Context) error {
	tables := []string{
		"payments",
		"invoices",
		"quotes",
		"jobs",
		"properties",
		"customers",
		"crews",
		"equipment",
		"api_keys",
		"user_sessions",
		"users",
		"subscriptions",
		"tenants",
	}
	
	for _, table := range tables {
		_, err := tf.db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			return fmt.Errorf("failed to truncate %s: %w", table, err)
		}
	}
	
	return nil
}