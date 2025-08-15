package repositories_test

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
	"github.com/pageza/landscaping-app/backend/tests/testutils"
)

func TestCustomerRepository_Integration(t *testing.T) {
	// Skip integration tests in unit test mode
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	// Setup test database connection
	db, cleanup := setupTestDB(t)
	defer cleanup()

	repo := repository.NewCustomerRepository(db)
	fixtures := testutils.NewTestFixtures(db)
	
	ctx := context.Background()

	t.Run("CreateCustomer", func(t *testing.T) {
		tenant := fixtures.CreateTestTenant()
		customer := fixtures.CreateTestCustomer(tenant.ID)
		
		err := repo.CreateCustomer(ctx, customer)
		assert.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, customer.ID)
		assert.NotZero(t, customer.CreatedAt)
		assert.NotZero(t, customer.UpdatedAt)

		// Verify customer was created
		retrieved, err := repo.GetCustomerByID(ctx, customer.ID)
		assert.NoError(t, err)
		assert.Equal(t, customer.Email, retrieved.Email)
		assert.Equal(t, customer.TenantID, retrieved.TenantID)
	})

	t.Run("GetCustomerByEmail", func(t *testing.T) {
		tenant := fixtures.CreateTestTenant()
		customer := fixtures.CreateTestCustomer(tenant.ID)
		
		err := repo.CreateCustomer(ctx, customer)
		require.NoError(t, err)

		retrieved, err := repo.GetCustomerByEmail(ctx, tenant.ID, customer.Email)
		assert.NoError(t, err)
		assert.Equal(t, customer.ID, retrieved.ID)
		assert.Equal(t, customer.Email, retrieved.Email)

		// Test email not found
		_, err = repo.GetCustomerByEmail(ctx, tenant.ID, "nonexistent@example.com")
		assert.Error(t, err)
	})

	t.Run("UpdateCustomer", func(t *testing.T) {
		tenant := fixtures.CreateTestTenant()
		customer := fixtures.CreateTestCustomer(tenant.ID)
		
		err := repo.CreateCustomer(ctx, customer)
		require.NoError(t, err)

		// Update customer details
		customer.Name = "Updated Name"
		customer.Phone = "+1999888777"
		customer.Tags = []string{"updated", "test"}
		customer.Metadata["updated"] = true

		err = repo.UpdateCustomer(ctx, customer)
		assert.NoError(t, err)

		// Verify updates
		retrieved, err := repo.GetCustomerByID(ctx, customer.ID)
		assert.NoError(t, err)
		assert.Equal(t, "Updated Name", retrieved.Name)
		assert.Equal(t, "+1999888777", retrieved.Phone)
		assert.Contains(t, retrieved.Tags, "updated")
		assert.Equal(t, true, retrieved.Metadata["updated"])
		assert.True(t, retrieved.UpdatedAt.After(customer.CreatedAt))
	})

	t.Run("DeleteCustomer", func(t *testing.T) {
		tenant := fixtures.CreateTestTenant()
		customer := fixtures.CreateTestCustomer(tenant.ID)
		
		err := repo.CreateCustomer(ctx, customer)
		require.NoError(t, err)

		err = repo.DeleteCustomer(ctx, customer.ID)
		assert.NoError(t, err)

		// Verify customer is deleted
		_, err = repo.GetCustomerByID(ctx, customer.ID)
		assert.Error(t, err)
	})

	t.Run("ListCustomers", func(t *testing.T) {
		tenant := fixtures.CreateTestTenant()
		
		// Create test customers with different attributes
		customers := []*domain.Customer{
			fixtures.CreateTestCustomer(tenant.ID, func(c *domain.Customer) {
				c.Status = "active"
				c.Tags = []string{"vip", "residential"}
			}),
			fixtures.CreateTestCustomer(tenant.ID, func(c *domain.Customer) {
				c.Status = "active"
				c.Tags = []string{"commercial"}
			}),
			fixtures.CreateTestCustomer(tenant.ID, func(c *domain.Customer) {
				c.Status = "inactive"
				c.Tags = []string{"residential"}
			}),
		}

		for _, customer := range customers {
			err := repo.CreateCustomer(ctx, customer)
			require.NoError(t, err)
		}

		// Test basic listing
		result, total, err := repo.ListCustomers(ctx, tenant.ID, map[string]interface{}{}, 0, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
		assert.GreaterOrEqual(t, total, 3)

		// Test filtering by status
		result, total, err = repo.ListCustomers(ctx, tenant.ID, map[string]interface{}{
			"status": "active",
		}, 0, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
		for _, customer := range result {
			assert.Equal(t, "active", customer.Status)
		}

		// Test filtering by tags
		result, total, err = repo.ListCustomers(ctx, tenant.ID, map[string]interface{}{
			"tags": []string{"vip"},
		}, 0, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)
		for _, customer := range result {
			assert.Contains(t, customer.Tags, "vip")
		}

		// Test pagination
		result, total, err = repo.ListCustomers(ctx, tenant.ID, map[string]interface{}{}, 0, 2)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(result), 2)
		assert.GreaterOrEqual(t, total, 3)

		// Test second page
		result2, total2, err := repo.ListCustomers(ctx, tenant.ID, map[string]interface{}{}, 2, 2)
		assert.NoError(t, err)
		assert.Equal(t, total, total2)
		if len(result) > 0 && len(result2) > 0 {
			assert.NotEqual(t, result[0].ID, result2[0].ID)
		}
	})

	t.Run("SearchCustomers", func(t *testing.T) {
		tenant := fixtures.CreateTestTenant()
		
		// Create customers with searchable data
		customers := []*domain.Customer{
			fixtures.CreateTestCustomer(tenant.ID, func(c *domain.Customer) {
				c.Name = "John Smith"
				c.Email = "john.smith@example.com"
				c.Company = "Smith Landscaping"
			}),
			fixtures.CreateTestCustomer(tenant.ID, func(c *domain.Customer) {
				c.Name = "Jane Johnson"
				c.Email = "jane.johnson@example.com"
				c.Company = "Johnson Properties"
			}),
			fixtures.CreateTestCustomer(tenant.ID, func(c *domain.Customer) {
				c.Name = "Bob Wilson"
				c.Email = "bob.wilson@example.com"
				c.Company = "Wilson Corp"
			}),
		}

		for _, customer := range customers {
			err := repo.CreateCustomer(ctx, customer)
			require.NoError(t, err)
		}

		// Search by name
		result, err := repo.SearchCustomers(ctx, tenant.ID, "john", 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2) // John Smith and Jane Johnson
		
		// Search by email
		result, err = repo.SearchCustomers(ctx, tenant.ID, "smith@example", 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)
		assert.Contains(t, result[0].Email, "smith@example")

		// Search by company
		result, err = repo.SearchCustomers(ctx, tenant.ID, "landscaping", 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 1)
		assert.Contains(t, result[0].Company, "Landscaping")

		// Search with no results
		result, err = repo.SearchCustomers(ctx, tenant.ID, "nonexistent", 10)
		assert.NoError(t, err)
		assert.Len(t, result, 0)

		// Test search limit
		result, err = repo.SearchCustomers(ctx, tenant.ID, "example", 2)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(result), 2)
	})

	t.Run("TenantIsolation", func(t *testing.T) {
		tenant1 := fixtures.CreateTestTenant()
		tenant2 := fixtures.CreateTestTenant()
		
		customer1 := fixtures.CreateTestCustomer(tenant1.ID)
		customer2 := fixtures.CreateTestCustomer(tenant2.ID)
		
		err := repo.CreateCustomer(ctx, customer1)
		require.NoError(t, err)
		err = repo.CreateCustomer(ctx, customer2)
		require.NoError(t, err)

		// Verify tenant1 can only see its customers
		result, total, err := repo.ListCustomers(ctx, tenant1.ID, map[string]interface{}{}, 0, 10)
		assert.NoError(t, err)
		for _, customer := range result {
			assert.Equal(t, tenant1.ID, customer.TenantID)
		}

		// Verify tenant2 can only see its customers
		result, total, err = repo.ListCustomers(ctx, tenant2.ID, map[string]interface{}{}, 0, 10)
		assert.NoError(t, err)
		for _, customer := range result {
			assert.Equal(t, tenant2.ID, customer.TenantID)
		}

		// Verify cross-tenant access is denied
		_, err = repo.GetCustomerByEmail(ctx, tenant2.ID, customer1.Email)
		assert.Error(t, err)

		// Search should also respect tenant isolation
		result, err = repo.SearchCustomers(ctx, tenant1.ID, customer2.Name, 10)
		assert.NoError(t, err)
		assert.Len(t, result, 0)
	})

	t.Run("ConcurrentOperations", func(t *testing.T) {
		tenant := fixtures.CreateTestTenant()
		
		// Test concurrent customer creation
		const numGoroutines = 10
		done := make(chan error, numGoroutines)
		
		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				customer := fixtures.CreateTestCustomer(tenant.ID, func(c *domain.Customer) {
					c.Email = fmt.Sprintf("concurrent%d@example.com", index)
				})
				done <- repo.CreateCustomer(ctx, customer)
			}(i)
		}

		// Wait for all goroutines to complete
		for i := 0; i < numGoroutines; i++ {
			err := <-done
			assert.NoError(t, err)
		}

		// Verify all customers were created
		result, total, err := repo.ListCustomers(ctx, tenant.ID, map[string]interface{}{}, 0, 50)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, numGoroutines)
		
		// Verify no duplicate emails
		emailMap := make(map[string]bool)
		for _, customer := range result {
			assert.False(t, emailMap[customer.Email], "Duplicate email found: %s", customer.Email)
			emailMap[customer.Email] = true
		}
	})

	t.Run("DataIntegrity", func(t *testing.T) {
		tenant := fixtures.CreateTestTenant()
		
		// Test email uniqueness constraint
		customer1 := fixtures.CreateTestCustomer(tenant.ID)
		customer2 := fixtures.CreateTestCustomer(tenant.ID, func(c *domain.Customer) {
			c.Email = customer1.Email // Same email
		})
		
		err := repo.CreateCustomer(ctx, customer1)
		require.NoError(t, err)
		
		err = repo.CreateCustomer(ctx, customer2)
		assert.Error(t, err, "Should not allow duplicate emails within same tenant")

		// Test that same email can exist in different tenants
		tenant2 := fixtures.CreateTestTenant()
		customer3 := fixtures.CreateTestCustomer(tenant2.ID, func(c *domain.Customer) {
			c.Email = customer1.Email // Same email, different tenant
		})
		
		err = repo.CreateCustomer(ctx, customer3)
		assert.NoError(t, err, "Should allow same email in different tenants")

		// Test foreign key constraints
		invalidCustomer := fixtures.CreateTestCustomer(uuid.New()) // Invalid tenant ID
		err = repo.CreateCustomer(ctx, invalidCustomer)
		assert.Error(t, err, "Should not allow invalid tenant ID")
	})

	t.Run("TransactionBehavior", func(t *testing.T) {
		// This would test transaction rollback scenarios
		// For brevity, showing concept with database error simulation
		tenant := fixtures.CreateTestTenant()
		customer := fixtures.CreateTestCustomer(tenant.ID)
		
		// Create customer successfully
		err := repo.CreateCustomer(ctx, customer)
		require.NoError(t, err)
		
		originalName := customer.Name
		customer.Name = "Updated in Transaction"
		
		// Simulate transaction rollback by causing constraint violation
		customer.Email = "" // This should cause a validation error
		
		err = repo.UpdateCustomer(ctx, customer)
		assert.Error(t, err, "Update should fail due to validation")
		
		// Verify original data is preserved
		retrieved, err := repo.GetCustomerByID(ctx, customer.ID)
		assert.NoError(t, err)
		assert.Equal(t, originalName, retrieved.Name, "Original name should be preserved on rollback")
	})
}

func TestCustomerRepository_Unit(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := repository.NewCustomerRepository(db)
	ctx := context.Background()

	t.Run("CreateCustomer_SQL", func(t *testing.T) {
		customer := &domain.Customer{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Email:    "test@example.com",
			Name:     "Test Customer",
			Phone:    "+1234567890",
			Status:   "active",
			Tags:     []string{"test"},
			Metadata: map[string]interface{}{"source": "test"},
		}

		mock.ExpectExec("INSERT INTO customers").
			WithArgs(
				customer.ID,
				customer.TenantID,
				customer.Email,
				customer.Phone,
				customer.Name,
				customer.Company,
				customer.Address,
				customer.City,
				customer.State,
				customer.ZipCode,
				customer.Country,
				customer.Status,
				sqlmock.AnyArg(), // tags JSON
				sqlmock.AnyArg(), // metadata JSON
				sqlmock.AnyArg(), // created_at
				sqlmock.AnyArg(), // updated_at
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.CreateCustomer(ctx, customer)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetCustomerByID_SQL", func(t *testing.T) {
		customerID := uuid.New()
		tenantID := uuid.New()
		
		rows := sqlmock.NewRows([]string{
			"id", "tenant_id", "email", "phone", "name", "company",
			"address", "city", "state", "zip_code", "country", "status",
			"tags", "metadata", "created_at", "updated_at",
		}).AddRow(
			customerID, tenantID, "test@example.com", "+1234567890", "Test Customer", "Test Co",
			"123 Main St", "Test City", "TS", "12345", "USA", "active",
			`["test"]`, `{"source": "test"}`, time.Now(), time.Now(),
		)

		mock.ExpectQuery("SELECT (.+) FROM customers WHERE id = ?").
			WithArgs(customerID).
			WillReturnRows(rows)

		customer, err := repo.GetCustomerByID(ctx, customerID)
		assert.NoError(t, err)
		assert.Equal(t, customerID, customer.ID)
		assert.Equal(t, "test@example.com", customer.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("GetCustomerByID_NotFound", func(t *testing.T) {
		customerID := uuid.New()
		
		mock.ExpectQuery("SELECT (.+) FROM customers WHERE id = ?").
			WithArgs(customerID).
			WillReturnError(sql.ErrNoRows)

		customer, err := repo.GetCustomerByID(ctx, customerID)
		assert.Error(t, err)
		assert.Nil(t, customer)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("UpdateCustomer_SQL", func(t *testing.T) {
		customer := &domain.Customer{
			ID:       uuid.New(),
			TenantID: uuid.New(),
			Email:    "updated@example.com",
			Name:     "Updated Customer",
			Status:   "active",
		}

		mock.ExpectExec("UPDATE customers SET").
			WithArgs(
				customer.Email,
				customer.Phone,
				customer.Name,
				customer.Company,
				customer.Address,
				customer.City,
				customer.State,
				customer.ZipCode,
				customer.Country,
				customer.Status,
				sqlmock.AnyArg(), // tags JSON
				sqlmock.AnyArg(), // metadata JSON
				sqlmock.AnyArg(), // updated_at
				customer.ID,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.UpdateCustomer(ctx, customer)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("DeleteCustomer_SQL", func(t *testing.T) {
		customerID := uuid.New()

		mock.ExpectExec("DELETE FROM customers WHERE id = ?").
			WithArgs(customerID).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.DeleteCustomer(ctx, customerID)
		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("ListCustomers_WithFilters_SQL", func(t *testing.T) {
		tenantID := uuid.New()
		
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(5)
		mock.ExpectQuery("SELECT COUNT\\(\\*\\) FROM customers").
			WithArgs(tenantID, "active").
			WillReturnRows(countRows)

		dataRows := sqlmock.NewRows([]string{
			"id", "tenant_id", "email", "phone", "name", "company",
			"address", "city", "state", "zip_code", "country", "status",
			"tags", "metadata", "created_at", "updated_at",
		}).AddRow(
			uuid.New(), tenantID, "test1@example.com", "+1234567890", "Customer 1", "Company 1",
			"123 Main St", "City 1", "ST", "12345", "USA", "active",
			`["tag1"]`, `{}`, time.Now(), time.Now(),
		).AddRow(
			uuid.New(), tenantID, "test2@example.com", "+9876543210", "Customer 2", "Company 2",
			"456 Oak Ave", "City 2", "ST", "54321", "USA", "active",
			`["tag2"]`, `{}`, time.Now(), time.Now(),
		)

		mock.ExpectQuery("SELECT (.+) FROM customers").
			WithArgs(tenantID, "active", 10, 0).
			WillReturnRows(dataRows)

		filters := map[string]interface{}{
			"status": "active",
		}
		
		customers, total, err := repo.ListCustomers(ctx, tenantID, filters, 0, 10)
		assert.NoError(t, err)
		assert.Len(t, customers, 2)
		assert.Equal(t, 5, total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

// setupTestDB creates a test database connection for integration tests
func setupTestDB(t *testing.T) (*sql.DB, func()) {
	// This would connect to a test database (PostgreSQL, SQLite, etc.)
	// For this example, we'll use a mock setup
	// In a real integration test, you'd use something like:
	// - dockertest to spin up a real PostgreSQL instance
	// - testcontainers to manage database lifecycle
	// - or connect to a dedicated test database
	
	db, _, err := sqlmock.New()
	require.NoError(t, err)
	
	cleanup := func() {
		db.Close()
	}
	
	return db, cleanup
}

// Alternative integration test setup using dockertest
func setupRealTestDB(t *testing.T) (*sql.DB, func()) {
	// Example using dockertest (commented out as it requires Docker)
	/*
	pool, err := dockertest.NewPool("")
	require.NoError(t, err)

	resource, err := pool.Run("postgres", "13", []string{
		"POSTGRES_PASSWORD=secret",
		"POSTGRES_DB=landscaping_test",
	})
	require.NoError(t, err)

	hostAndPort := resource.GetHostPort("5432/tcp")
	databaseUrl := fmt.Sprintf("postgres://postgres:secret@%s/landscaping_test?sslmode=disable", hostAndPort)

	var db *sql.DB
	pool.Retry(func() error {
		var err error
		db, err = sql.Open("postgres", databaseUrl)
		if err != nil {
			return err
		}
		return db.Ping()
	})
	require.NotNil(t, db)

	// Run migrations
	runMigrations(t, db)

	cleanup := func() {
		db.Close()
		pool.Purge(resource)
	}

	return db, cleanup
	*/
	
	// Placeholder for now
	return setupTestDB(t)
}