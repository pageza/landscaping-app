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

// InvoiceRepositoryImpl implements the full invoice repository interface
type InvoiceRepositoryImpl struct {
	db *Database
}

// NewInvoiceRepositoryFull creates a new full invoice repository
func NewInvoiceRepositoryFull(db *Database) services.InvoiceRepositoryFull {
	return &InvoiceRepositoryImpl{db: db}
}

// Create creates a new invoice
func (r *InvoiceRepositoryImpl) Create(ctx context.Context, invoice *domain.Invoice) error {
	query := `
		INSERT INTO invoices (
			id, tenant_id, customer_id, job_id, invoice_number, status,
			subtotal, tax_rate, tax_amount, total_amount, issued_date, due_date,
			paid_date, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16
		)`

	_, err := r.db.ExecContext(ctx, query,
		invoice.ID,
		invoice.TenantID,
		invoice.CustomerID,
		invoice.JobID,
		invoice.InvoiceNumber,
		invoice.Status,
		invoice.Subtotal,
		invoice.TaxRate,
		invoice.TaxAmount,
		invoice.TotalAmount,
		invoice.IssuedDate,
		invoice.DueDate,
		invoice.PaidDate,
		invoice.Notes,
		invoice.CreatedAt,
		invoice.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}

	return nil
}

// GetByID retrieves an invoice by ID
func (r *InvoiceRepositoryImpl) GetByID(ctx context.Context, tenantID, invoiceID uuid.UUID) (*domain.Invoice, error) {
	query := `
		SELECT id, tenant_id, customer_id, job_id, invoice_number, status,
			   subtotal, tax_rate, tax_amount, total_amount, issued_date, due_date,
			   paid_date, notes, created_at, updated_at
		FROM invoices
		WHERE id = $1 AND tenant_id = $2`

	var invoice domain.Invoice
	err := r.db.QueryRowContext(ctx, query, invoiceID, tenantID).Scan(
		&invoice.ID,
		&invoice.TenantID,
		&invoice.CustomerID,
		&invoice.JobID,
		&invoice.InvoiceNumber,
		&invoice.Status,
		&invoice.Subtotal,
		&invoice.TaxRate,
		&invoice.TaxAmount,
		&invoice.TotalAmount,
		&invoice.IssuedDate,
		&invoice.DueDate,
		&invoice.PaidDate,
		&invoice.Notes,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invoice: %w", err)
	}

	return &invoice, nil
}

// Update updates an existing invoice
func (r *InvoiceRepositoryImpl) Update(ctx context.Context, invoice *domain.Invoice) error {
	query := `
		UPDATE invoices SET
			customer_id = $3, job_id = $4, invoice_number = $5, status = $6,
			subtotal = $7, tax_rate = $8, tax_amount = $9, total_amount = $10,
			issued_date = $11, due_date = $12, paid_date = $13, notes = $14,
			updated_at = $15
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		invoice.ID,
		invoice.TenantID,
		invoice.CustomerID,
		invoice.JobID,
		invoice.InvoiceNumber,
		invoice.Status,
		invoice.Subtotal,
		invoice.TaxRate,
		invoice.TaxAmount,
		invoice.TotalAmount,
		invoice.IssuedDate,
		invoice.DueDate,
		invoice.PaidDate,
		invoice.Notes,
		invoice.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// Delete deletes an invoice (soft delete by updating status)
func (r *InvoiceRepositoryImpl) Delete(ctx context.Context, tenantID, invoiceID uuid.UUID) error {
	query := `
		UPDATE invoices SET
			status = 'deleted',
			updated_at = $3
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, invoiceID, tenantID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete invoice: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice not found")
	}

	return nil
}

// List lists invoices with filtering and pagination
func (r *InvoiceRepositoryImpl) List(ctx context.Context, tenantID uuid.UUID, filter *services.InvoiceFilter) ([]*domain.Invoice, int64, error) {
	baseQuery := `
		FROM invoices
		WHERE tenant_id = $1 AND status != 'deleted'`

	var conditions []string
	var args []interface{}
	args = append(args, tenantID)
	argIndex := 2

	// Apply filters
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, filter.Status)
		argIndex++
	}

	if filter.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("customer_id = $%d", argIndex))
		args = append(args, *filter.CustomerID)
		argIndex++
	}

	if filter.JobID != nil {
		conditions = append(conditions, fmt.Sprintf("job_id = $%d", argIndex))
		args = append(args, *filter.JobID)
		argIndex++
	}

	if filter.DueDate != nil {
		conditions = append(conditions, fmt.Sprintf("due_date <= $%d", argIndex))
		args = append(args, *filter.DueDate)
		argIndex++
	}

	if filter.Overdue {
		conditions = append(conditions, "due_date < NOW() AND status NOT IN ('paid', 'cancelled')")
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(invoice_number ILIKE $%d OR notes ILIKE $%d)", argIndex, argIndex))
		args = append(args, "%"+filter.Search+"%")
		argIndex++
	}

	whereClause := baseQuery
	if len(conditions) > 0 {
		whereClause += " AND " + strings.Join(conditions, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) " + whereClause
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count invoices: %w", err)
	}

	// Main query with pagination
	selectFields := `
		SELECT id, tenant_id, customer_id, job_id, invoice_number, status,
			   subtotal, tax_rate, tax_amount, total_amount, issued_date, due_date,
			   paid_date, notes, created_at, updated_at`

	orderBy := " ORDER BY created_at DESC"
	if filter.SortBy != "" {
		direction := "ASC"
		if filter.SortDesc {
			direction = "DESC"
		}
		orderBy = fmt.Sprintf(" ORDER BY %s %s", filter.SortBy, direction)
	}

	limit := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PerPage, (filter.Page-1)*filter.PerPage)

	query := selectFields + whereClause + orderBy + limit

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err := rows.Scan(
			&invoice.ID,
			&invoice.TenantID,
			&invoice.CustomerID,
			&invoice.JobID,
			&invoice.InvoiceNumber,
			&invoice.Status,
			&invoice.Subtotal,
			&invoice.TaxRate,
			&invoice.TaxAmount,
			&invoice.TotalAmount,
			&invoice.IssuedDate,
			&invoice.DueDate,
			&invoice.PaidDate,
			&invoice.Notes,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, &invoice)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating invoice rows: %w", err)
	}

	return invoices, total, nil
}

// GetByCustomerID retrieves invoices for a specific customer
func (r *InvoiceRepositoryImpl) GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *services.InvoiceFilter) ([]*domain.Invoice, int64, error) {
	if filter == nil {
		filter = &services.InvoiceFilter{}
	}
	filter.CustomerID = &customerID
	return r.List(ctx, tenantID, filter)
}

// GetByJobID retrieves an invoice for a specific job
func (r *InvoiceRepositoryImpl) GetByJobID(ctx context.Context, tenantID, jobID uuid.UUID) (*domain.Invoice, error) {
	query := `
		SELECT id, tenant_id, customer_id, job_id, invoice_number, status,
			   subtotal, tax_rate, tax_amount, total_amount, issued_date, due_date,
			   paid_date, notes, created_at, updated_at
		FROM invoices
		WHERE job_id = $1 AND tenant_id = $2 AND status != 'deleted'`

	var invoice domain.Invoice
	err := r.db.QueryRowContext(ctx, query, jobID, tenantID).Scan(
		&invoice.ID,
		&invoice.TenantID,
		&invoice.CustomerID,
		&invoice.JobID,
		&invoice.InvoiceNumber,
		&invoice.Status,
		&invoice.Subtotal,
		&invoice.TaxRate,
		&invoice.TaxAmount,
		&invoice.TotalAmount,
		&invoice.IssuedDate,
		&invoice.DueDate,
		&invoice.PaidDate,
		&invoice.Notes,
		&invoice.CreatedAt,
		&invoice.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get invoice by job ID: %w", err)
	}

	return &invoice, nil
}

// GetByStatus retrieves invoices by status
func (r *InvoiceRepositoryImpl) GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Invoice, error) {
	query := `
		SELECT id, tenant_id, customer_id, job_id, invoice_number, status,
			   subtotal, tax_rate, tax_amount, total_amount, issued_date, due_date,
			   paid_date, notes, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1 AND status = $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoices by status: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err := rows.Scan(
			&invoice.ID,
			&invoice.TenantID,
			&invoice.CustomerID,
			&invoice.JobID,
			&invoice.InvoiceNumber,
			&invoice.Status,
			&invoice.Subtotal,
			&invoice.TaxRate,
			&invoice.TaxAmount,
			&invoice.TotalAmount,
			&invoice.IssuedDate,
			&invoice.DueDate,
			&invoice.PaidDate,
			&invoice.Notes,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, &invoice)
	}

	return invoices, nil
}

// GetOverdue retrieves overdue invoices
func (r *InvoiceRepositoryImpl) GetOverdue(ctx context.Context, tenantID uuid.UUID) ([]*domain.Invoice, error) {
	query := `
		SELECT id, tenant_id, customer_id, job_id, invoice_number, status,
			   subtotal, tax_rate, tax_amount, total_amount, issued_date, due_date,
			   paid_date, notes, created_at, updated_at
		FROM invoices
		WHERE tenant_id = $1 
		  AND due_date < NOW() 
		  AND status NOT IN ('paid', 'cancelled', 'deleted')
		ORDER BY due_date ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*domain.Invoice
	for rows.Next() {
		var invoice domain.Invoice
		err := rows.Scan(
			&invoice.ID,
			&invoice.TenantID,
			&invoice.CustomerID,
			&invoice.JobID,
			&invoice.InvoiceNumber,
			&invoice.Status,
			&invoice.Subtotal,
			&invoice.TaxRate,
			&invoice.TaxAmount,
			&invoice.TotalAmount,
			&invoice.IssuedDate,
			&invoice.DueDate,
			&invoice.PaidDate,
			&invoice.Notes,
			&invoice.CreatedAt,
			&invoice.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan overdue invoice: %w", err)
		}
		invoices = append(invoices, &invoice)
	}

	return invoices, nil
}

// CreateInvoiceService creates an invoice service line item
func (r *InvoiceRepositoryImpl) CreateInvoiceService(ctx context.Context, invoiceService *services.InvoiceLineItem) error {
	query := `
		INSERT INTO invoice_services (
			id, invoice_id, service_id, quantity, unit_price, total_price, description, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		invoiceService.ID,
		invoiceService.InvoiceID,
		invoiceService.ServiceID,
		invoiceService.Quantity,
		invoiceService.UnitPrice,
		invoiceService.TotalPrice,
		invoiceService.Description,
		invoiceService.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create invoice service: %w", err)
	}

	return nil
}

// UpdateInvoiceService updates an invoice service line item
func (r *InvoiceRepositoryImpl) UpdateInvoiceService(ctx context.Context, invoiceService *services.InvoiceLineItem) error {
	query := `
		UPDATE invoice_services SET
			service_id = $3, quantity = $4, unit_price = $5, total_price = $6, description = $7
		WHERE id = $1 AND invoice_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		invoiceService.ID,
		invoiceService.InvoiceID,
		invoiceService.ServiceID,
		invoiceService.Quantity,
		invoiceService.UnitPrice,
		invoiceService.TotalPrice,
		invoiceService.Description,
	)

	if err != nil {
		return fmt.Errorf("failed to update invoice service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice service not found")
	}

	return nil
}

// DeleteInvoiceService deletes an invoice service line item
func (r *InvoiceRepositoryImpl) DeleteInvoiceService(ctx context.Context, invoiceServiceID uuid.UUID) error {
	query := `DELETE FROM invoice_services WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, invoiceServiceID)
	if err != nil {
		return fmt.Errorf("failed to delete invoice service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("invoice service not found")
	}

	return nil
}

// GetInvoiceServices retrieves all services for an invoice
func (r *InvoiceRepositoryImpl) GetInvoiceServices(ctx context.Context, invoiceID uuid.UUID) ([]*services.InvoiceLineItem, error) {
	query := `
		SELECT id, invoice_id, service_id, quantity, unit_price, total_price, description, created_at
		FROM invoice_services
		WHERE invoice_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invoice services: %w", err)
	}
	defer rows.Close()

	var services []*services.InvoiceLineItem
	for rows.Next() {
		var service services.InvoiceLineItem
		err := rows.Scan(
			&service.ID,
			&service.InvoiceID,
			&service.ServiceID,
			&service.Quantity,
			&service.UnitPrice,
			&service.TotalPrice,
			&service.Description,
			&service.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invoice service: %w", err)
		}
		services = append(services, &service)
	}

	return services, nil
}

// GetNextInvoiceNumber generates the next invoice number for a tenant
func (r *InvoiceRepositoryImpl) GetNextInvoiceNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	// Get the current year
	currentYear := time.Now().Year()

	// Find the highest invoice number for this year
	query := `
		SELECT COALESCE(MAX(CAST(SUBSTRING(invoice_number FROM '[0-9]+$') AS INTEGER)), 0)
		FROM invoices
		WHERE tenant_id = $1 
		  AND invoice_number ~ '^INV-' || $2 || '-[0-9]+$'`

	var maxNumber int
	err := r.db.QueryRowContext(ctx, query, tenantID, currentYear).Scan(&maxNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get next invoice number: %w", err)
	}

	// Generate next number
	nextNumber := maxNumber + 1
	invoiceNumber := fmt.Sprintf("INV-%d-%04d", currentYear, nextNumber)

	return invoiceNumber, nil
}