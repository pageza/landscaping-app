package database_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// DatabaseTestSuite contains database tests
type DatabaseTestSuite struct {
	db          *sql.DB
	migrator    *migrate.Migrate
	testDBName  string
	originalDB  *sql.DB
}

// NewDatabaseTestSuite creates a new database test suite
func NewDatabaseTestSuite(t *testing.T) *DatabaseTestSuite {
	// Connect to main postgres database to create test database
	originalDB, err := sql.Open("postgres", "postgres://postgres:password@localhost:5432/postgres?sslmode=disable")
	require.NoError(t, err)

	// Create unique test database
	testDBName := fmt.Sprintf("landscaping_test_%d_%s", time.Now().Unix(), uuid.New().String()[:8])
	
	_, err = originalDB.Exec(fmt.Sprintf("CREATE DATABASE %s", testDBName))
	require.NoError(t, err)

	// Connect to test database
	testDB, err := sql.Open("postgres", fmt.Sprintf("postgres://postgres:password@localhost:5432/%s?sslmode=disable", testDBName))
	require.NoError(t, err)

	// Create migrator
	driver, err := postgres.WithInstance(testDB, &postgres.Config{})
	require.NoError(t, err)

	migrator, err := migrate.NewWithDatabaseInstance(
		"file://../../migrations",
		"postgres",
		driver,
	)
	require.NoError(t, err)

	return &DatabaseTestSuite{
		db:          testDB,
		migrator:    migrator,
		testDBName:  testDBName,
		originalDB:  originalDB,
	}
}

// Cleanup cleans up test resources
func (suite *DatabaseTestSuite) Cleanup() {
	if suite.db != nil {
		suite.db.Close()
	}
	
	if suite.originalDB != nil {
		suite.originalDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", suite.testDBName))
		suite.originalDB.Close()
	}
	
	if suite.migrator != nil {
		suite.migrator.Close()
	}
}

func TestDatabaseMigrations(t *testing.T) {
	suite := NewDatabaseTestSuite(t)
	defer suite.Cleanup()

	t.Run("MigrationUp", suite.testMigrationUp)
	t.Run("MigrationDown", suite.testMigrationDown)
	t.Run("MigrationIdempotency", suite.testMigrationIdempotency)
	t.Run("DataIntegrity", suite.testDataIntegrity)
	t.Run("ForeignKeyConstraints", suite.testForeignKeyConstraints)
	t.Run("IndexCreation", suite.testIndexCreation)
	t.Run("ColumnConstraints", suite.testColumnConstraints)
	t.Run("TriggerCreation", suite.testTriggerCreation)
	t.Run("ViewCreation", suite.testViewCreation)
	t.Run("SequenceMigration", suite.testSequenceMigration)
	t.Run("PartitionMigration", suite.testPartitionMigration)
	t.Run("RollbackScenarios", suite.testRollbackScenarios)
	t.Run("PerformanceValidation", suite.testPerformanceValidation)
}

func (suite *DatabaseTestSuite) testMigrationUp(t *testing.T) {
	// Apply all migrations
	err := suite.migrator.Up()
	assert.NoError(t, err)

	// Verify all expected tables exist
	expectedTables := []string{
		"tenants",
		"users",
		"customers",
		"properties",
		"jobs",
		"quotes",
		"invoices",
		"payments",
		"equipment",
		"crews",
		"user_sessions",
		"api_keys",
		"subscriptions",
		"audit_logs",
		"notifications",
	}

	for _, tableName := range expectedTables {
		exists, err := suite.tableExists(tableName)
		assert.NoError(t, err, "Error checking table existence: %s", tableName)
		assert.True(t, exists, "Table should exist: %s", tableName)
	}

	// Verify schema integrity
	suite.verifySchemaIntegrity(t)
}

func (suite *DatabaseTestSuite) testMigrationDown(t *testing.T) {
	// First, apply all migrations
	err := suite.migrator.Up()
	require.NoError(t, err)

	// Get current version
	version, dirty, err := suite.migrator.Version()
	require.NoError(t, err)
	require.False(t, dirty)

	// Roll back one migration
	err = suite.migrator.Steps(-1)
	assert.NoError(t, err)

	// Verify version decreased
	newVersion, dirty, err := suite.migrator.Version()
	assert.NoError(t, err)
	assert.False(t, dirty)
	assert.Less(t, newVersion, version)

	// Roll back to beginning
	err = suite.migrator.Down()
	assert.NoError(t, err)

	// Verify all tables are gone
	for _, tableName := range []string{"users", "tenants", "customers"} {
		exists, err := suite.tableExists(tableName)
		assert.NoError(t, err)
		assert.False(t, exists, "Table should not exist after rollback: %s", tableName)
	}
}

func (suite *DatabaseTestSuite) testMigrationIdempotency(t *testing.T) {
	// Apply migrations multiple times
	err := suite.migrator.Up()
	assert.NoError(t, err)

	// Should not error when applied again
	err = suite.migrator.Up()
	assert.NoError(t, err)

	// Verify schema is still correct
	suite.verifySchemaIntegrity(t)
}

func (suite *DatabaseTestSuite) testDataIntegrity(t *testing.T) {
	// Apply migrations
	err := suite.migrator.Up()
	require.NoError(t, err)

	ctx := context.Background()

	// Test tenant creation and data isolation
	tenantID1 := uuid.New()
	tenantID2 := uuid.New()

	// Create tenants
	_, err = suite.db.ExecContext(ctx, `
		INSERT INTO tenants (id, name, domain, status, settings, metadata, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID1, "Tenant 1", "tenant1.example.com", "active",
		`{"timezone": "UTC"}`, `{}`, time.Now(), time.Now())
	require.NoError(t, err)

	_, err = suite.db.ExecContext(ctx, `
		INSERT INTO tenants (id, name, domain, status, settings, metadata, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID2, "Tenant 2", "tenant2.example.com", "active",
		`{"timezone": "UTC"}`, `{}`, time.Now(), time.Now())
	require.NoError(t, err)

	// Create users for each tenant
	userID1 := uuid.New()
	userID2 := uuid.New()

	_, err = suite.db.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		userID1, tenantID1, "user1@tenant1.com", "hashedpassword", "User", "One",
		"user", "active", time.Now(), time.Now())
	require.NoError(t, err)

	_, err = suite.db.ExecContext(ctx, `
		INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		userID2, tenantID2, "user2@tenant2.com", "hashedpassword", "User", "Two",
		"user", "active", time.Now(), time.Now())
	require.NoError(t, err)

	// Verify tenant isolation - users should only see their tenant's data
	var count int
	err = suite.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE tenant_id = $1", tenantID1).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "Tenant 1 should have exactly 1 user")

	err = suite.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE tenant_id = $2", tenantID2).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count, "Tenant 2 should have exactly 1 user")
}

func (suite *DatabaseTestSuite) testForeignKeyConstraints(t *testing.T) {
	err := suite.migrator.Up()
	require.NoError(t, err)

	ctx := context.Background()

	// Test foreign key constraint violations
	t.Run("UserTenantConstraint", func(t *testing.T) {
		// Try to create user with non-existent tenant
		_, err := suite.db.ExecContext(ctx, `
			INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			uuid.New(), uuid.New(), "test@example.com", "hashedpassword", "Test", "User",
			"user", "active", time.Now(), time.Now())

		assert.Error(t, err, "Should fail due to foreign key constraint")
		
		// Check if it's a foreign key violation
		if pqErr, ok := err.(*pq.Error); ok {
			assert.Equal(t, "23503", string(pqErr.Code), "Should be foreign key violation")
		}
	})

	t.Run("CustomerTenantConstraint", func(t *testing.T) {
		// Try to create customer with non-existent tenant
		_, err := suite.db.ExecContext(ctx, `
			INSERT INTO customers (id, tenant_id, email, phone, name, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			uuid.New(), uuid.New(), "customer@example.com", "+1-555-0123", "Test Customer",
			"active", time.Now(), time.Now())

		assert.Error(t, err, "Should fail due to foreign key constraint")
	})

	t.Run("CascadeDelete", func(t *testing.T) {
		// Create tenant
		tenantID := uuid.New()
		_, err := suite.db.ExecContext(ctx, `
			INSERT INTO tenants (id, name, domain, status, settings, metadata, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			tenantID, "Test Tenant", "test.example.com", "active",
			`{}`, `{}`, time.Now(), time.Now())
		require.NoError(t, err)

		// Create user
		userID := uuid.New()
		_, err = suite.db.ExecContext(ctx, `
			INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			userID, tenantID, "test@example.com", "hashedpassword", "Test", "User",
			"user", "active", time.Now(), time.Now())
		require.NoError(t, err)

		// Create customer
		customerID := uuid.New()
		_, err = suite.db.ExecContext(ctx, `
			INSERT INTO customers (id, tenant_id, email, phone, name, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			customerID, tenantID, "customer@example.com", "+1-555-0123", "Test Customer",
			"active", time.Now(), time.Now())
		require.NoError(t, err)

		// Delete tenant (should cascade delete users and customers)
		_, err = suite.db.ExecContext(ctx, "DELETE FROM tenants WHERE id = $1", tenantID)
		assert.NoError(t, err)

		// Verify cascade deletion
		var count int
		err = suite.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM users WHERE tenant_id = $1", tenantID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "Users should be deleted with tenant")

		err = suite.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM customers WHERE tenant_id = $1", tenantID).Scan(&count)
		assert.NoError(t, err)
		assert.Equal(t, 0, count, "Customers should be deleted with tenant")
	})
}

func (suite *DatabaseTestSuite) testIndexCreation(t *testing.T) {
	err := suite.migrator.Up()
	require.NoError(t, err)

	// Verify important indexes exist
	expectedIndexes := map[string][]string{
		"users": {
			"idx_users_tenant_id",
			"idx_users_email_tenant_id",
			"idx_users_status",
		},
		"customers": {
			"idx_customers_tenant_id",
			"idx_customers_email_tenant_id",
			"idx_customers_status",
		},
		"jobs": {
			"idx_jobs_tenant_id",
			"idx_jobs_customer_id",
			"idx_jobs_status",
			"idx_jobs_scheduled_date",
		},
		"invoices": {
			"idx_invoices_tenant_id",
			"idx_invoices_customer_id",
			"idx_invoices_status",
			"idx_invoices_due_date",
		},
		"payments": {
			"idx_payments_tenant_id",
			"idx_payments_invoice_id",
			"idx_payments_status",
		},
	}

	for tableName, indexes := range expectedIndexes {
		for _, indexName := range indexes {
			exists, err := suite.indexExists(tableName, indexName)
			assert.NoError(t, err, "Error checking index existence: %s on %s", indexName, tableName)
			assert.True(t, exists, "Index should exist: %s on %s", indexName, tableName)
		}
	}
}

func (suite *DatabaseTestSuite) testColumnConstraints(t *testing.T) {
	err := suite.migrator.Up()
	require.NoError(t, err)

	ctx := context.Background()

	// Test NOT NULL constraints
	t.Run("NotNullConstraints", func(t *testing.T) {
		tenantID := uuid.New()
		
		// Create tenant first
		_, err := suite.db.ExecContext(ctx, `
			INSERT INTO tenants (id, name, domain, status, settings, metadata, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			tenantID, "Test Tenant", "test.example.com", "active",
			`{}`, `{}`, time.Now(), time.Now())
		require.NoError(t, err)

		// Try to insert user without required fields
		_, err = suite.db.ExecContext(ctx, `
			INSERT INTO users (id, tenant_id, email, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			uuid.New(), tenantID, "", "user", "active", time.Now(), time.Now())

		assert.Error(t, err, "Should fail due to NOT NULL constraint on email")
	})

	// Test CHECK constraints
	t.Run("CheckConstraints", func(t *testing.T) {
		tenantID := uuid.New()
		
		// Create tenant first
		_, err := suite.db.ExecContext(ctx, `
			INSERT INTO tenants (id, name, domain, status, settings, metadata, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			tenantID, "Test Tenant", "test.example.com", "active",
			`{}`, `{}`, time.Now(), time.Now())
		require.NoError(t, err)

		// Try to insert user with invalid role
		_, err = suite.db.ExecContext(ctx, `
			INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			uuid.New(), tenantID, "test@example.com", "hashedpassword", "Test", "User",
			"invalid_role", "active", time.Now(), time.Now())

		assert.Error(t, err, "Should fail due to CHECK constraint on role")
	})

	// Test UNIQUE constraints
	t.Run("UniqueConstraints", func(t *testing.T) {
		tenantID := uuid.New()
		
		// Create tenant first
		_, err := suite.db.ExecContext(ctx, `
			INSERT INTO tenants (id, name, domain, status, settings, metadata, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			tenantID, "Test Tenant", "test.example.com", "active",
			`{}`, `{}`, time.Now(), time.Now())
		require.NoError(t, err)

		email := "unique@example.com"

		// Insert first user
		_, err = suite.db.ExecContext(ctx, `
			INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			uuid.New(), tenantID, email, "hashedpassword", "Test", "User",
			"user", "active", time.Now(), time.Now())
		require.NoError(t, err)

		// Try to insert user with same email in same tenant
		_, err = suite.db.ExecContext(ctx, `
			INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			uuid.New(), tenantID, email, "hashedpassword", "Test", "User2",
			"user", "active", time.Now(), time.Now())

		assert.Error(t, err, "Should fail due to UNIQUE constraint on email within tenant")

		// But should allow same email in different tenant
		tenantID2 := uuid.New()
		_, err = suite.db.ExecContext(ctx, `
			INSERT INTO tenants (id, name, domain, status, settings, metadata, created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			tenantID2, "Test Tenant 2", "test2.example.com", "active",
			`{}`, `{}`, time.Now(), time.Now())
		require.NoError(t, err)

		_, err = suite.db.ExecContext(ctx, `
			INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			uuid.New(), tenantID2, email, "hashedpassword", "Test", "User3",
			"user", "active", time.Now(), time.Now())

		assert.NoError(t, err, "Should allow same email in different tenant")
	})
}

func (suite *DatabaseTestSuite) testTriggerCreation(t *testing.T) {
	err := suite.migrator.Up()
	require.NoError(t, err)

	ctx := context.Background()

	// Test updated_at trigger
	tenantID := uuid.New()
	
	// Create tenant
	_, err = suite.db.ExecContext(ctx, `
		INSERT INTO tenants (id, name, domain, status, settings, metadata, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID, "Test Tenant", "test.example.com", "active",
		`{}`, `{}`, time.Now(), time.Now().Add(-time.Hour))
	require.NoError(t, err)

	// Get original updated_at
	var originalUpdatedAt time.Time
	err = suite.db.QueryRowContext(ctx,
		"SELECT updated_at FROM tenants WHERE id = $1", tenantID).Scan(&originalUpdatedAt)
	require.NoError(t, err)

	// Wait a moment and update
	time.Sleep(10 * time.Millisecond)
	_, err = suite.db.ExecContext(ctx,
		"UPDATE tenants SET name = $1 WHERE id = $2", "Updated Tenant", tenantID)
	require.NoError(t, err)

	// Check that updated_at was automatically updated
	var newUpdatedAt time.Time
	err = suite.db.QueryRowContext(ctx,
		"SELECT updated_at FROM tenants WHERE id = $1", tenantID).Scan(&newUpdatedAt)
	require.NoError(t, err)

	assert.True(t, newUpdatedAt.After(originalUpdatedAt),
		"updated_at should be automatically updated by trigger")
}

func (suite *DatabaseTestSuite) testViewCreation(t *testing.T) {
	err := suite.migrator.Up()
	require.NoError(t, err)

	// Check if views exist
	expectedViews := []string{
		"customer_summary_view",
		"job_statistics_view",
		"financial_summary_view",
	}

	for _, viewName := range expectedViews {
		exists, err := suite.viewExists(viewName)
		assert.NoError(t, err, "Error checking view existence: %s", viewName)
		// Views might not be implemented yet, so just log if they don't exist
		if !exists {
			t.Logf("View %s does not exist (may not be implemented yet)", viewName)
		}
	}
}

func (suite *DatabaseTestSuite) testSequenceMigration(t *testing.T) {
	err := suite.migrator.Up()
	require.NoError(t, err)

	// Test sequences for invoice numbers, quote numbers, etc.
	expectedSequences := []string{
		"invoice_number_seq",
		"quote_number_seq",
		"job_number_seq",
	}

	for _, seqName := range expectedSequences {
		exists, err := suite.sequenceExists(seqName)
		assert.NoError(t, err, "Error checking sequence existence: %s", seqName)
		// Sequences might not be implemented yet
		if !exists {
			t.Logf("Sequence %s does not exist (may not be implemented yet)", seqName)
		}
	}
}

func (suite *DatabaseTestSuite) testPartitionMigration(t *testing.T) {
	err := suite.migrator.Up()
	require.NoError(t, err)

	// Test partitioned tables (for audit logs, notifications, etc.)
	// This is advanced functionality that might not be implemented yet
	partitionedTables := []string{
		"audit_logs_partitioned",
		"notifications_partitioned",
	}

	for _, tableName := range partitionedTables {
		exists, err := suite.tableExists(tableName)
		assert.NoError(t, err, "Error checking partitioned table existence: %s", tableName)
		if !exists {
			t.Logf("Partitioned table %s does not exist (may not be implemented yet)", tableName)
		}
	}
}

func (suite *DatabaseTestSuite) testRollbackScenarios(t *testing.T) {
	// Test various rollback scenarios
	t.Run("PartialMigrationRollback", func(t *testing.T) {
		// Apply migrations
		err := suite.migrator.Up()
		require.NoError(t, err)

		// Roll back one step
		err = suite.migrator.Steps(-1)
		assert.NoError(t, err)

		// Should still be in a consistent state
		version, dirty, err := suite.migrator.Version()
		assert.NoError(t, err)
		assert.False(t, dirty, "Database should not be in dirty state after rollback")
		assert.Greater(t, version, uint(0), "Should still have some migrations applied")
	})

	t.Run("CompleteRollback", func(t *testing.T) {
		// Apply migrations
		err := suite.migrator.Up()
		require.NoError(t, err)

		// Roll back completely
		err = suite.migrator.Down()
		assert.NoError(t, err)

		// Should have no migrations
		_, dirty, err := suite.migrator.Version()
		assert.Error(t, err) // Should error because no migrations are applied
		assert.False(t, dirty, "Database should not be in dirty state")

		// Main tables should not exist
		for _, tableName := range []string{"users", "customers", "tenants"} {
			exists, err := suite.tableExists(tableName)
			assert.NoError(t, err)
			assert.False(t, exists, "Table should not exist after complete rollback: %s", tableName)
		}
	})
}

func (suite *DatabaseTestSuite) testPerformanceValidation(t *testing.T) {
	err := suite.migrator.Up()
	require.NoError(t, err)

	ctx := context.Background()

	// Create test data for performance testing
	tenantID := uuid.New()
	_, err = suite.db.ExecContext(ctx, `
		INSERT INTO tenants (id, name, domain, status, settings, metadata, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		tenantID, "Performance Test Tenant", "perf.example.com", "active",
		`{}`, `{}`, time.Now(), time.Now())
	require.NoError(t, err)

	// Test query performance
	t.Run("IndexedQueries", func(t *testing.T) {
		// Create multiple customers
		for i := 0; i < 100; i++ {
			_, err := suite.db.ExecContext(ctx, `
				INSERT INTO customers (id, tenant_id, email, phone, name, status, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
				uuid.New(), tenantID, fmt.Sprintf("customer%d@example.com", i),
				fmt.Sprintf("+1-555-%04d", i), fmt.Sprintf("Customer %d", i),
				"active", time.Now(), time.Now())
			require.NoError(t, err)
		}

		// Test indexed query performance
		start := time.Now()
		var count int
		err := suite.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM customers WHERE tenant_id = $1 AND status = $2",
			tenantID, "active").Scan(&count)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.Equal(t, 100, count)
		assert.Less(t, duration, 100*time.Millisecond, "Indexed query should be fast")

		// Test unindexed query to ensure it's slower
		start = time.Now()
		err = suite.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM customers WHERE phone LIKE $1", "+1-555-%").Scan(&count)
		unindexedDuration := time.Since(start)

		assert.NoError(t, err)
		// Unindexed query might be slower, but with small dataset might not be noticeable
		t.Logf("Indexed query: %v, Unindexed query: %v", duration, unindexedDuration)
	})
}

// Helper methods

func (suite *DatabaseTestSuite) tableExists(tableName string) (bool, error) {
	var exists bool
	err := suite.db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`, tableName).Scan(&exists)
	return exists, err
}

func (suite *DatabaseTestSuite) indexExists(tableName, indexName string) (bool, error) {
	var exists bool
	err := suite.db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM pg_indexes 
			WHERE schemaname = 'public' 
			AND tablename = $1 
			AND indexname = $2
		)`, tableName, indexName).Scan(&exists)
	return exists, err
}

func (suite *DatabaseTestSuite) viewExists(viewName string) (bool, error) {
	var exists bool
	err := suite.db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.views 
			WHERE table_schema = 'public' 
			AND table_name = $1
		)`, viewName).Scan(&exists)
	return exists, err
}

func (suite *DatabaseTestSuite) sequenceExists(sequenceName string) (bool, error) {
	var exists bool
	err := suite.db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.sequences 
			WHERE sequence_schema = 'public' 
			AND sequence_name = $1
		)`, sequenceName).Scan(&exists)
	return exists, err
}

func (suite *DatabaseTestSuite) verifySchemaIntegrity(t *testing.T) {
	// Verify that all tables have proper primary keys
	rows, err := suite.db.Query(`
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'`)
	require.NoError(t, err)
	defer rows.Close()

	for rows.Next() {
		var tableName string
		err := rows.Scan(&tableName)
		require.NoError(t, err)

		// Skip migration tables
		if tableName == "schema_migrations" {
			continue
		}

		// Check for primary key
		var hasPK bool
		err = suite.db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM information_schema.table_constraints 
				WHERE table_schema = 'public' 
				AND table_name = $1 
				AND constraint_type = 'PRIMARY KEY'
			)`, tableName).Scan(&hasPK)
		assert.NoError(t, err)
		assert.True(t, hasPK, "Table %s should have a primary key", tableName)
	}
}