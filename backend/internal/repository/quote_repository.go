package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// QuoteRepositoryImpl implements the full quote repository interface
type QuoteRepositoryImpl struct {
	db *Database
}

// NewQuoteRepositoryFull creates a new full quote repository
func NewQuoteRepositoryFull(db *Database) services.QuoteRepositoryFull {
	return &QuoteRepositoryImpl{db: db}
}

// Create creates a new quote
func (r *QuoteRepositoryImpl) Create(ctx context.Context, quote *domain.Quote) error {
	query := `
		INSERT INTO quotes (
			id, tenant_id, customer_id, property_id, quote_number, title, description,
			subtotal, tax_rate, tax_amount, total_amount, status, valid_until,
			terms_and_conditions, notes, created_by, approved_by, approved_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20
		)`

	_, err := r.db.ExecContext(ctx, query,
		quote.ID,
		quote.TenantID,
		quote.CustomerID,
		quote.PropertyID,
		quote.QuoteNumber,
		quote.Title,
		quote.Description,
		quote.Subtotal,
		quote.TaxRate,
		quote.TaxAmount,
		quote.TotalAmount,
		quote.Status,
		quote.ValidUntil,
		quote.TermsAndConditions,
		quote.Notes,
		quote.CreatedBy,
		quote.ApprovedBy,
		quote.ApprovedAt,
		quote.CreatedAt,
		quote.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create quote: %w", err)
	}

	return nil
}

// GetByID retrieves a quote by ID
func (r *QuoteRepositoryImpl) GetByID(ctx context.Context, tenantID, quoteID uuid.UUID) (*domain.Quote, error) {
	query := `
		SELECT id, tenant_id, customer_id, property_id, quote_number, title, description,
			   subtotal, tax_rate, tax_amount, total_amount, status, valid_until,
			   terms_and_conditions, notes, created_by, approved_by, approved_at,
			   created_at, updated_at
		FROM quotes
		WHERE id = $1 AND tenant_id = $2`

	var quote domain.Quote
	err := r.db.QueryRowContext(ctx, query, quoteID, tenantID).Scan(
		&quote.ID,
		&quote.TenantID,
		&quote.CustomerID,
		&quote.PropertyID,
		&quote.QuoteNumber,
		&quote.Title,
		&quote.Description,
		&quote.Subtotal,
		&quote.TaxRate,
		&quote.TaxAmount,
		&quote.TotalAmount,
		&quote.Status,
		&quote.ValidUntil,
		&quote.TermsAndConditions,
		&quote.Notes,
		&quote.CreatedBy,
		&quote.ApprovedBy,
		&quote.ApprovedAt,
		&quote.CreatedAt,
		&quote.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}

	return &quote, nil
}

// Update updates an existing quote
func (r *QuoteRepositoryImpl) Update(ctx context.Context, quote *domain.Quote) error {
	query := `
		UPDATE quotes SET
			customer_id = $3, property_id = $4, quote_number = $5, title = $6, description = $7,
			subtotal = $8, tax_rate = $9, tax_amount = $10, total_amount = $11, status = $12,
			valid_until = $13, terms_and_conditions = $14, notes = $15, created_by = $16,
			approved_by = $17, approved_at = $18, updated_at = $19
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		quote.ID,
		quote.TenantID,
		quote.CustomerID,
		quote.PropertyID,
		quote.QuoteNumber,
		quote.Title,
		quote.Description,
		quote.Subtotal,
		quote.TaxRate,
		quote.TaxAmount,
		quote.TotalAmount,
		quote.Status,
		quote.ValidUntil,
		quote.TermsAndConditions,
		quote.Notes,
		quote.CreatedBy,
		quote.ApprovedBy,
		quote.ApprovedAt,
		quote.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update quote: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quote not found")
	}

	return nil
}

// Delete deletes a quote (soft delete by updating status)
func (r *QuoteRepositoryImpl) Delete(ctx context.Context, tenantID, quoteID uuid.UUID) error {
	query := `
		UPDATE quotes SET
			status = 'deleted',
			updated_at = $3
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, quoteID, tenantID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete quote: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quote not found")
	}

	return nil
}

// List lists quotes with filtering and pagination
func (r *QuoteRepositoryImpl) List(ctx context.Context, tenantID uuid.UUID, filter *services.QuoteFilter) ([]*domain.Quote, int64, error) {
	baseQuery := `
		FROM quotes
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

	if filter.PropertyID != nil {
		conditions = append(conditions, fmt.Sprintf("property_id = $%d", argIndex))
		args = append(args, *filter.PropertyID)
		argIndex++
	}

	if filter.ValidUntil != nil {
		conditions = append(conditions, fmt.Sprintf("valid_until <= $%d", argIndex))
		args = append(args, *filter.ValidUntil)
		argIndex++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(quote_number ILIKE $%d OR title ILIKE $%d OR notes ILIKE $%d)", argIndex, argIndex, argIndex))
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
		return nil, 0, fmt.Errorf("failed to count quotes: %w", err)
	}

	// Main query with pagination
	selectFields := `
		SELECT id, tenant_id, customer_id, property_id, quote_number, title, description,
			   subtotal, tax_rate, tax_amount, total_amount, status, valid_until,
			   terms_and_conditions, notes, created_by, approved_by, approved_at,
			   created_at, updated_at`

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
		return nil, 0, fmt.Errorf("failed to list quotes: %w", err)
	}
	defer rows.Close()

	var quotes []*domain.Quote
	for rows.Next() {
		var quote domain.Quote
		err := rows.Scan(
			&quote.ID,
			&quote.TenantID,
			&quote.CustomerID,
			&quote.PropertyID,
			&quote.QuoteNumber,
			&quote.Title,
			&quote.Description,
			&quote.Subtotal,
			&quote.TaxRate,
			&quote.TaxAmount,
			&quote.TotalAmount,
			&quote.Status,
			&quote.ValidUntil,
			&quote.TermsAndConditions,
			&quote.Notes,
			&quote.CreatedBy,
			&quote.ApprovedBy,
			&quote.ApprovedAt,
			&quote.CreatedAt,
			&quote.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan quote: %w", err)
		}
		quotes = append(quotes, &quote)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating quote rows: %w", err)
	}

	return quotes, total, nil
}

// GetByCustomerID retrieves quotes for a specific customer
func (r *QuoteRepositoryImpl) GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *services.QuoteFilter) ([]*domain.Quote, int64, error) {
	if filter == nil {
		filter = &services.QuoteFilter{}
	}
	filter.CustomerID = &customerID
	return r.List(ctx, tenantID, filter)
}

// GetByPropertyID retrieves quotes for a specific property
func (r *QuoteRepositoryImpl) GetByPropertyID(ctx context.Context, tenantID, propertyID uuid.UUID, filter *services.QuoteFilter) ([]*domain.Quote, int64, error) {
	if filter == nil {
		filter = &services.QuoteFilter{}
	}
	filter.PropertyID = &propertyID
	return r.List(ctx, tenantID, filter)
}

// GetByStatus retrieves quotes by status
func (r *QuoteRepositoryImpl) GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Quote, error) {
	query := `
		SELECT id, tenant_id, customer_id, property_id, quote_number, title, description,
			   subtotal, tax_rate, tax_amount, total_amount, status, valid_until,
			   terms_and_conditions, notes, created_by, approved_by, approved_at,
			   created_at, updated_at
		FROM quotes
		WHERE tenant_id = $1 AND status = $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotes by status: %w", err)
	}
	defer rows.Close()

	var quotes []*domain.Quote
	for rows.Next() {
		var quote domain.Quote
		err := rows.Scan(
			&quote.ID,
			&quote.TenantID,
			&quote.CustomerID,
			&quote.PropertyID,
			&quote.QuoteNumber,
			&quote.Title,
			&quote.Description,
			&quote.Subtotal,
			&quote.TaxRate,
			&quote.TaxAmount,
			&quote.TotalAmount,
			&quote.Status,
			&quote.ValidUntil,
			&quote.TermsAndConditions,
			&quote.Notes,
			&quote.CreatedBy,
			&quote.ApprovedBy,
			&quote.ApprovedAt,
			&quote.CreatedAt,
			&quote.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quote: %w", err)
		}
		quotes = append(quotes, &quote)
	}

	return quotes, nil
}

// CreateQuoteService creates a quote service line item
func (r *QuoteRepositoryImpl) CreateQuoteService(ctx context.Context, quoteService *domain.QuoteService) error {
	query := `
		INSERT INTO quote_services (
			id, quote_id, service_id, quantity, unit_price, total_price, description, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		quoteService.ID,
		quoteService.QuoteID,
		quoteService.ServiceID,
		quoteService.Quantity,
		quoteService.UnitPrice,
		quoteService.TotalPrice,
		quoteService.Description,
		quoteService.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create quote service: %w", err)
	}

	return nil
}

// UpdateQuoteService updates a quote service line item
func (r *QuoteRepositoryImpl) UpdateQuoteService(ctx context.Context, quoteService *domain.QuoteService) error {
	query := `
		UPDATE quote_services SET
			service_id = $3, quantity = $4, unit_price = $5, total_price = $6, description = $7
		WHERE id = $1 AND quote_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		quoteService.ID,
		quoteService.QuoteID,
		quoteService.ServiceID,
		quoteService.Quantity,
		quoteService.UnitPrice,
		quoteService.TotalPrice,
		quoteService.Description,
	)

	if err != nil {
		return fmt.Errorf("failed to update quote service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quote service not found")
	}

	return nil
}

// DeleteQuoteService deletes a quote service line item
func (r *QuoteRepositoryImpl) DeleteQuoteService(ctx context.Context, quoteServiceID uuid.UUID) error {
	query := `DELETE FROM quote_services WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, quoteServiceID)
	if err != nil {
		return fmt.Errorf("failed to delete quote service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("quote service not found")
	}

	return nil
}

// GetQuoteServices retrieves all services for a quote
func (r *QuoteRepositoryImpl) GetQuoteServices(ctx context.Context, quoteID uuid.UUID) ([]*domain.QuoteService, error) {
	query := `
		SELECT id, quote_id, service_id, quantity, unit_price, total_price, description, created_at
		FROM quote_services
		WHERE quote_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, quoteID)
	if err != nil {
		return nil, fmt.Errorf("failed to get quote services: %w", err)
	}
	defer rows.Close()

	var services []*domain.QuoteService
	for rows.Next() {
		var service domain.QuoteService
		err := rows.Scan(
			&service.ID,
			&service.QuoteID,
			&service.ServiceID,
			&service.Quantity,
			&service.UnitPrice,
			&service.TotalPrice,
			&service.Description,
			&service.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan quote service: %w", err)
		}
		services = append(services, &service)
	}

	return services, nil
}

// GetNextQuoteNumber generates the next quote number for a tenant
func (r *QuoteRepositoryImpl) GetNextQuoteNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	// Get the current year
	currentYear := time.Now().Year()

	// Find the highest quote number for this year
	query := `
		SELECT COALESCE(MAX(CAST(SUBSTRING(quote_number FROM '[0-9]+$') AS INTEGER)), 0)
		FROM quotes
		WHERE tenant_id = $1 
		  AND quote_number ~ '^Q-' || $2 || '-[0-9]+$'`

	var maxNumber int
	err := r.db.QueryRowContext(ctx, query, tenantID, currentYear).Scan(&maxNumber)
	if err != nil {
		return "", fmt.Errorf("failed to get next quote number: %w", err)
	}

	// Generate next number
	nextNumber := maxNumber + 1
	quoteNumber := fmt.Sprintf("Q-%d-%04d", currentYear, nextNumber)

	return quoteNumber, nil
}

