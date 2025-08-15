package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// CustomerRepositoryImpl implements the CustomerRepository interface
type CustomerRepositoryImpl struct {
	db *Database
}

// NewCustomerRepositoryImpl creates a new customer repository instance
func NewCustomerRepositoryImpl(db *Database) services.CustomerRepository {
	return &CustomerRepositoryImpl{db: db}
}

// Create creates a new customer
func (r *CustomerRepositoryImpl) Create(ctx context.Context, customer *domain.EnhancedCustomer) error {
	query := `
		INSERT INTO customers (
			id, tenant_id, first_name, last_name, email, phone, company_name,
			address_line1, address_line2, city, state, zip_code, country,
			preferred_contact_method, lead_source, customer_type, credit_limit,
			payment_terms, notes, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22
		)`

	_, err := r.db.ExecContext(ctx, query,
		customer.ID,
		customer.TenantID,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.CompanyName,
		customer.AddressLine1,
		customer.AddressLine2,
		customer.City,
		customer.State,
		customer.ZipCode,
		customer.Country,
		customer.PreferredContactMethod,
		customer.LeadSource,
		customer.CustomerType,
		customer.CreditLimit,
		customer.PaymentTerms,
		customer.Notes,
		customer.Status,
		customer.CreatedAt,
		customer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	return nil
}

// GetByID retrieves a customer by ID
func (r *CustomerRepositoryImpl) GetByID(ctx context.Context, tenantID, customerID uuid.UUID) (*domain.EnhancedCustomer, error) {
	query := `
		SELECT 
			id, tenant_id, first_name, last_name, email, phone, company_name,
			address_line1, address_line2, city, state, zip_code, country,
			preferred_contact_method, lead_source, customer_type, credit_limit,
			payment_terms, notes, status, created_at, updated_at
		FROM customers
		WHERE id = $1 AND tenant_id = $2 AND status != 'deleted'`

	row := r.db.QueryRowContext(ctx, query, customerID, tenantID)

	customer := &domain.EnhancedCustomer{}
	err := row.Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.CompanyName,
		&customer.AddressLine1,
		&customer.AddressLine2,
		&customer.City,
		&customer.State,
		&customer.ZipCode,
		&customer.Country,
		&customer.PreferredContactMethod,
		&customer.LeadSource,
		&customer.CustomerType,
		&customer.CreditLimit,
		&customer.PaymentTerms,
		&customer.Notes,
		&customer.Status,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer: %w", err)
	}

	return customer, nil
}

// Update updates an existing customer
func (r *CustomerRepositoryImpl) Update(ctx context.Context, customer *domain.EnhancedCustomer) error {
	query := `
		UPDATE customers SET
			first_name = $3,
			last_name = $4,
			email = $5,
			phone = $6,
			company_name = $7,
			address_line1 = $8,
			address_line2 = $9,
			city = $10,
			state = $11,
			zip_code = $12,
			country = $13,
			preferred_contact_method = $14,
			lead_source = $15,
			customer_type = $16,
			credit_limit = $17,
			payment_terms = $18,
			notes = $19,
			status = $20,
			updated_at = $21
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		customer.ID,
		customer.TenantID,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.CompanyName,
		customer.AddressLine1,
		customer.AddressLine2,
		customer.City,
		customer.State,
		customer.ZipCode,
		customer.Country,
		customer.PreferredContactMethod,
		customer.LeadSource,
		customer.CustomerType,
		customer.CreditLimit,
		customer.PaymentTerms,
		customer.Notes,
		customer.Status,
		customer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer not found or not authorized")
	}

	return nil
}

// Delete deletes a customer (soft delete)
func (r *CustomerRepositoryImpl) Delete(ctx context.Context, tenantID, customerID uuid.UUID) error {
	query := `
		UPDATE customers 
		SET status = 'deleted', updated_at = $3
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, customerID, tenantID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete customer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("customer not found or not authorized")
	}

	return nil
}

// List retrieves customers with filtering and pagination
func (r *CustomerRepositoryImpl) List(ctx context.Context, tenantID uuid.UUID, filter *services.CustomerFilter) ([]*domain.EnhancedCustomer, int64, error) {
	// Build WHERE clause
	whereClause := "WHERE tenant_id = $1 AND status != 'deleted'"
	args := []interface{}{tenantID}
	argIndex := 2

	if filter.CustomerType != "" {
		whereClause += fmt.Sprintf(" AND customer_type = $%d", argIndex)
		args = append(args, filter.CustomerType)
		argIndex++
	}

	if filter.Status != "" {
		whereClause += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.LeadSource != "" {
		whereClause += fmt.Sprintf(" AND lead_source = $%d", argIndex)
		args = append(args, filter.LeadSource)
		argIndex++
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		whereClause += fmt.Sprintf(" AND (first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d OR company_name ILIKE $%d)", 
			argIndex, argIndex, argIndex, argIndex)
		args = append(args, searchPattern)
		argIndex++
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM customers " + whereClause
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count customers: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if filter.SortBy != "" {
		direction := "ASC"
		if filter.SortDesc {
			direction = "DESC"
		}
		// Validate sort field to prevent SQL injection
		validSortFields := map[string]bool{
			"first_name": true,
			"last_name":  true,
			"email":      true,
			"created_at": true,
			"updated_at": true,
			"company_name": true,
		}
		if validSortFields[filter.SortBy] {
			orderBy = fmt.Sprintf("ORDER BY %s %s", filter.SortBy, direction)
		}
	}

	// Add pagination
	limit := filter.PerPage
	offset := (filter.Page - 1) * filter.PerPage
	paginationClause := fmt.Sprintf(" %s LIMIT $%d OFFSET $%d", orderBy, argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute main query
	query := `
		SELECT 
			id, tenant_id, first_name, last_name, email, phone, company_name,
			address_line1, address_line2, city, state, zip_code, country,
			preferred_contact_method, lead_source, customer_type, credit_limit,
			payment_terms, notes, status, created_at, updated_at
		FROM customers ` + whereClause + paginationClause

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list customers: %w", err)
	}
	defer rows.Close()

	customers := make([]*domain.EnhancedCustomer, 0)
	for rows.Next() {
		customer := &domain.EnhancedCustomer{}
		err := rows.Scan(
			&customer.ID,
			&customer.TenantID,
			&customer.FirstName,
			&customer.LastName,
			&customer.Email,
			&customer.Phone,
			&customer.CompanyName,
			&customer.AddressLine1,
			&customer.AddressLine2,
			&customer.City,
			&customer.State,
			&customer.ZipCode,
			&customer.Country,
			&customer.PreferredContactMethod,
			&customer.LeadSource,
			&customer.CustomerType,
			&customer.CreditLimit,
			&customer.PaymentTerms,
			&customer.Notes,
			&customer.Status,
			&customer.CreatedAt,
			&customer.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan customer: %w", err)
		}
		customers = append(customers, customer)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating customers: %w", err)
	}

	return customers, total, nil
}

// Search searches customers by query string
func (r *CustomerRepositoryImpl) Search(ctx context.Context, tenantID uuid.UUID, query string, filter *services.CustomerFilter) ([]*domain.EnhancedCustomer, int64, error) {
	// Update filter to include search
	if filter == nil {
		filter = &services.CustomerFilter{}
	}
	filter.Search = query

	return r.List(ctx, tenantID, filter)
}

// GetByEmail retrieves a customer by email
func (r *CustomerRepositoryImpl) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.EnhancedCustomer, error) {
	query := `
		SELECT 
			id, tenant_id, first_name, last_name, email, phone, company_name,
			address_line1, address_line2, city, state, zip_code, country,
			preferred_contact_method, lead_source, customer_type, credit_limit,
			payment_terms, notes, status, created_at, updated_at
		FROM customers
		WHERE tenant_id = $1 AND email = $2 AND status != 'deleted'`

	row := r.db.QueryRowContext(ctx, query, tenantID, email)

	customer := &domain.EnhancedCustomer{}
	err := row.Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.CompanyName,
		&customer.AddressLine1,
		&customer.AddressLine2,
		&customer.City,
		&customer.State,
		&customer.ZipCode,
		&customer.Country,
		&customer.PreferredContactMethod,
		&customer.LeadSource,
		&customer.CustomerType,
		&customer.CreditLimit,
		&customer.PaymentTerms,
		&customer.Notes,
		&customer.Status,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer by email: %w", err)
	}

	return customer, nil
}

// GetByPhone retrieves a customer by phone
func (r *CustomerRepositoryImpl) GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*domain.EnhancedCustomer, error) {
	query := `
		SELECT 
			id, tenant_id, first_name, last_name, email, phone, company_name,
			address_line1, address_line2, city, state, zip_code, country,
			preferred_contact_method, lead_source, customer_type, credit_limit,
			payment_terms, notes, status, created_at, updated_at
		FROM customers
		WHERE tenant_id = $1 AND phone = $2 AND status != 'deleted'`

	row := r.db.QueryRowContext(ctx, query, tenantID, phone)

	customer := &domain.EnhancedCustomer{}
	err := row.Scan(
		&customer.ID,
		&customer.TenantID,
		&customer.FirstName,
		&customer.LastName,
		&customer.Email,
		&customer.Phone,
		&customer.CompanyName,
		&customer.AddressLine1,
		&customer.AddressLine2,
		&customer.City,
		&customer.State,
		&customer.ZipCode,
		&customer.Country,
		&customer.PreferredContactMethod,
		&customer.LeadSource,
		&customer.CustomerType,
		&customer.CreditLimit,
		&customer.PaymentTerms,
		&customer.Notes,
		&customer.Status,
		&customer.CreatedAt,
		&customer.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get customer by phone: %w", err)
	}

	return customer, nil
}

// GetCustomerSummary retrieves customer analytics and summary
func (r *CustomerRepositoryImpl) GetCustomerSummary(ctx context.Context, tenantID, customerID uuid.UUID) (*services.CustomerSummary, error) {
	// Get job statistics
	jobStatsQuery := `
		SELECT 
			COUNT(*) as total_jobs,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_jobs,
			COALESCE(SUM(CASE WHEN status = 'completed' THEN total_amount END), 0) as total_revenue,
			MAX(CASE WHEN status = 'completed' THEN scheduled_date END) as last_job_date
		FROM jobs
		WHERE tenant_id = $1 AND customer_id = $2`

	var totalJobs, completedJobs int
	var totalRevenue float64
	var lastJobDate *time.Time

	err := r.db.QueryRowContext(ctx, jobStatsQuery, tenantID, customerID).Scan(
		&totalJobs, &completedJobs, &totalRevenue, &lastJobDate,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get job statistics: %w", err)
	}

	// Calculate average job value
	averageJobValue := 0.0
	if completedJobs > 0 {
		averageJobValue = totalRevenue / float64(completedJobs)
	}

	// Get payment history (last 10 payments)
	paymentHistoryQuery := `
		SELECT p.processed_at, p.amount, p.status
		FROM payments p
		JOIN invoices i ON p.invoice_id = i.id
		WHERE i.tenant_id = $1 AND i.customer_id = $2 AND p.status = 'completed'
		ORDER BY p.processed_at DESC
		LIMIT 10`

	rows, err := r.db.QueryContext(ctx, paymentHistoryQuery, tenantID, customerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payment history: %w", err)
	}
	defer rows.Close()

	paymentHistory := make([]services.PaymentHistoryEntry, 0)
	for rows.Next() {
		var entry services.PaymentHistoryEntry
		var processedAt time.Time
		err := rows.Scan(&processedAt, &entry.Amount, &entry.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment history: %w", err)
		}
		entry.Date = processedAt
		paymentHistory = append(paymentHistory, entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating payment history: %w", err)
	}

	summary := &services.CustomerSummary{
		TotalJobs:      totalJobs,
		CompletedJobs:  completedJobs,
		TotalRevenue:   totalRevenue,
		LastJobDate:    lastJobDate,
		AverageJobValue: averageJobValue,
		PaymentHistory: paymentHistory,
	}

	return summary, nil
}

// GetCustomerCount retrieves the total count of customers for a tenant
func (r *CustomerRepositoryImpl) GetCustomerCount(ctx context.Context, tenantID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM customers WHERE tenant_id = $1 AND status != 'deleted'`
	
	var count int64
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get customer count: %w", err)
	}

	return count, nil
}

// Transaction support for complex operations
func (r *CustomerRepositoryImpl) CreateWithTransaction(ctx context.Context, customer *domain.EnhancedCustomer, tx *sql.Tx) error {
	query := `
		INSERT INTO customers (
			id, tenant_id, first_name, last_name, email, phone, company_name,
			address_line1, address_line2, city, state, zip_code, country,
			preferred_contact_method, lead_source, customer_type, credit_limit,
			payment_terms, notes, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22
		)`

	_, err := tx.ExecContext(ctx, query,
		customer.ID,
		customer.TenantID,
		customer.FirstName,
		customer.LastName,
		customer.Email,
		customer.Phone,
		customer.CompanyName,
		customer.AddressLine1,
		customer.AddressLine2,
		customer.City,
		customer.State,
		customer.ZipCode,
		customer.Country,
		customer.PreferredContactMethod,
		customer.LeadSource,
		customer.CustomerType,
		customer.CreditLimit,
		customer.PaymentTerms,
		customer.Notes,
		customer.Status,
		customer.CreatedAt,
		customer.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create customer with transaction: %w", err)
	}

	return nil
}

// Bulk operations for data import/export
func (r *CustomerRepositoryImpl) BulkCreate(ctx context.Context, customers []*domain.EnhancedCustomer) error {
	if len(customers) == 0 {
		return nil
	}

	// Begin transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Prepare statement for bulk insert
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO customers (
			id, tenant_id, first_name, last_name, email, phone, company_name,
			address_line1, address_line2, city, state, zip_code, country,
			preferred_contact_method, lead_source, customer_type, credit_limit,
			payment_terms, notes, status, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22
		)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	// Execute for each customer
	for _, customer := range customers {
		_, err := stmt.ExecContext(ctx,
			customer.ID,
			customer.TenantID,
			customer.FirstName,
			customer.LastName,
			customer.Email,
			customer.Phone,
			customer.CompanyName,
			customer.AddressLine1,
			customer.AddressLine2,
			customer.City,
			customer.State,
			customer.ZipCode,
			customer.Country,
			customer.PreferredContactMethod,
			customer.LeadSource,
			customer.CustomerType,
			customer.CreditLimit,
			customer.PaymentTerms,
			customer.Notes,
			customer.Status,
			customer.CreatedAt,
			customer.UpdatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to bulk create customer %s: %w", customer.ID, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit bulk create transaction: %w", err)
	}

	return nil
}