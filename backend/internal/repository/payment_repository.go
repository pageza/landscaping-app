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

// PaymentRepositoryImpl implements the full payment repository interface
type PaymentRepositoryImpl struct {
	db *Database
}

// NewPaymentRepositoryFull creates a new full payment repository
func NewPaymentRepositoryFull(db *Database) services.PaymentRepositoryFull {
	return &PaymentRepositoryImpl{db: db}
}

// Create creates a new payment
func (r *PaymentRepositoryImpl) Create(ctx context.Context, payment *domain.Payment) error {
	query := `
		INSERT INTO payments (
			id, tenant_id, invoice_id, amount, payment_method, payment_gateway,
			gateway_transaction_id, status, processed_at, notes, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12
		)`

	_, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.TenantID,
		payment.InvoiceID,
		payment.Amount,
		payment.PaymentMethod,
		payment.PaymentGateway,
		payment.GatewayTransactionID,
		payment.Status,
		payment.ProcessedAt,
		payment.Notes,
		payment.CreatedAt,
		payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

// GetByID retrieves a payment by ID
func (r *PaymentRepositoryImpl) GetByID(ctx context.Context, tenantID, paymentID uuid.UUID) (*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, invoice_id, amount, payment_method, payment_gateway,
			   gateway_transaction_id, status, processed_at, notes, created_at, updated_at
		FROM payments
		WHERE id = $1 AND tenant_id = $2`

	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, query, paymentID, tenantID).Scan(
		&payment.ID,
		&payment.TenantID,
		&payment.InvoiceID,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.PaymentGateway,
		&payment.GatewayTransactionID,
		&payment.Status,
		&payment.ProcessedAt,
		&payment.Notes,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment: %w", err)
	}

	return &payment, nil
}

// Update updates an existing payment
func (r *PaymentRepositoryImpl) Update(ctx context.Context, payment *domain.Payment) error {
	query := `
		UPDATE payments SET
			invoice_id = $3, amount = $4, payment_method = $5, payment_gateway = $6,
			gateway_transaction_id = $7, status = $8, processed_at = $9, notes = $10, updated_at = $11
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		payment.ID,
		payment.TenantID,
		payment.InvoiceID,
		payment.Amount,
		payment.PaymentMethod,
		payment.PaymentGateway,
		payment.GatewayTransactionID,
		payment.Status,
		payment.ProcessedAt,
		payment.Notes,
		payment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}

// Delete deletes a payment (soft delete by updating status)
func (r *PaymentRepositoryImpl) Delete(ctx context.Context, tenantID, paymentID uuid.UUID) error {
	query := `
		UPDATE payments SET
			status = 'deleted',
			updated_at = $3
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, paymentID, tenantID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete payment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("payment not found")
	}

	return nil
}

// List lists payments with filtering and pagination
func (r *PaymentRepositoryImpl) List(ctx context.Context, tenantID uuid.UUID, filter *services.PaymentFilter) ([]*domain.Payment, int64, error) {
	baseQuery := `
		FROM payments
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

	if filter.Method != "" {
		conditions = append(conditions, fmt.Sprintf("payment_method = $%d", argIndex))
		args = append(args, filter.Method)
		argIndex++
	}

	if filter.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("invoice_id IN (SELECT id FROM invoices WHERE customer_id = $%d)", argIndex))
		args = append(args, *filter.CustomerID)
		argIndex++
	}

	if filter.InvoiceID != nil {
		conditions = append(conditions, fmt.Sprintf("invoice_id = $%d", argIndex))
		args = append(args, *filter.InvoiceID)
		argIndex++
	}

	if filter.DateRange != nil {
		conditions = append(conditions, fmt.Sprintf("processed_at >= $%d AND processed_at <= $%d", argIndex, argIndex+1))
		args = append(args, filter.DateRange.Start, filter.DateRange.End)
		argIndex += 2
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(gateway_transaction_id ILIKE $%d OR notes ILIKE $%d)", argIndex, argIndex))
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
		return nil, 0, fmt.Errorf("failed to count payments: %w", err)
	}

	// Main query with pagination
	selectFields := `
		SELECT id, tenant_id, invoice_id, amount, payment_method, payment_gateway,
			   gateway_transaction_id, status, processed_at, notes, created_at, updated_at`

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
		return nil, 0, fmt.Errorf("failed to list payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.TenantID,
			&payment.InvoiceID,
			&payment.Amount,
			&payment.PaymentMethod,
			&payment.PaymentGateway,
			&payment.GatewayTransactionID,
			&payment.Status,
			&payment.ProcessedAt,
			&payment.Notes,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, &payment)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating payment rows: %w", err)
	}

	return payments, total, nil
}

// GetByInvoiceID retrieves payments for a specific invoice
func (r *PaymentRepositoryImpl) GetByInvoiceID(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, invoice_id, amount, payment_method, payment_gateway,
			   gateway_transaction_id, status, processed_at, notes, created_at, updated_at
		FROM payments
		WHERE tenant_id = $1 AND invoice_id = $2 AND status != 'deleted'
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, invoiceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by invoice ID: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.TenantID,
			&payment.InvoiceID,
			&payment.Amount,
			&payment.PaymentMethod,
			&payment.PaymentGateway,
			&payment.GatewayTransactionID,
			&payment.Status,
			&payment.ProcessedAt,
			&payment.Notes,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}

// GetByCustomerID retrieves payments for a specific customer
func (r *PaymentRepositoryImpl) GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *services.PaymentFilter) ([]*domain.Payment, int64, error) {
	if filter == nil {
		filter = &services.PaymentFilter{}
	}
	filter.CustomerID = &customerID
	return r.List(ctx, tenantID, filter)
}

// GetByStatus retrieves payments by status
func (r *PaymentRepositoryImpl) GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, invoice_id, amount, payment_method, payment_gateway,
			   gateway_transaction_id, status, processed_at, notes, created_at, updated_at
		FROM payments
		WHERE tenant_id = $1 AND status = $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get payments by status: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.TenantID,
			&payment.InvoiceID,
			&payment.Amount,
			&payment.PaymentMethod,
			&payment.PaymentGateway,
			&payment.GatewayTransactionID,
			&payment.Status,
			&payment.ProcessedAt,
			&payment.Notes,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan payment: %w", err)
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}

// GetPaymentSummary returns payment analytics summary
func (r *PaymentRepositoryImpl) GetPaymentSummary(ctx context.Context, tenantID uuid.UUID, filter *services.PaymentFilter) (*services.PaymentSummary, error) {
	baseQuery := `
		FROM payments
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

	if filter.Method != "" {
		conditions = append(conditions, fmt.Sprintf("payment_method = $%d", argIndex))
		args = append(args, filter.Method)
		argIndex++
	}

	if filter.CustomerID != nil {
		conditions = append(conditions, fmt.Sprintf("invoice_id IN (SELECT id FROM invoices WHERE customer_id = $%d)", argIndex))
		args = append(args, *filter.CustomerID)
		argIndex++
	}

	if filter.InvoiceID != nil {
		conditions = append(conditions, fmt.Sprintf("invoice_id = $%d", argIndex))
		args = append(args, *filter.InvoiceID)
		argIndex++
	}

	if filter.DateRange != nil {
		conditions = append(conditions, fmt.Sprintf("processed_at >= $%d AND processed_at <= $%d", argIndex, argIndex+1))
		args = append(args, filter.DateRange.Start, filter.DateRange.End)
		argIndex += 2
	}

	whereClause := baseQuery
	if len(conditions) > 0 {
		whereClause += " AND " + strings.Join(conditions, " AND ")
	}

	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END), 0) as total_amount,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as successful_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_count,
			COALESCE(SUM(CASE WHEN status = 'refunded' THEN amount ELSE 0 END), 0) as refunded_amount,
			COALESCE(SUM(CASE WHEN status = 'pending' THEN amount ELSE 0 END), 0) as pending_amount,
			CASE 
				WHEN COUNT(CASE WHEN status = 'completed' THEN 1 END) > 0 
				THEN SUM(CASE WHEN status = 'completed' THEN amount ELSE 0 END) / COUNT(CASE WHEN status = 'completed' THEN 1 END)
				ELSE 0 
			END as average_transaction
		` + whereClause

	var summary services.PaymentSummary
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&summary.TotalAmount,
		&summary.SuccessfulCount,
		&summary.FailedCount,
		&summary.RefundedAmount,
		&summary.PendingAmount,
		&summary.AverageTransaction,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get payment summary: %w", err)
	}

	return &summary, nil
}

// GetByGatewayTransactionID retrieves a payment by gateway transaction ID
func (r *PaymentRepositoryImpl) GetByGatewayTransactionID(ctx context.Context, tenantID uuid.UUID, gatewayTransactionID string) (*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, invoice_id, amount, payment_method, payment_gateway,
			   gateway_transaction_id, status, processed_at, notes, created_at, updated_at
		FROM payments
		WHERE tenant_id = $1 AND gateway_transaction_id = $2 AND status != 'deleted'`

	var payment domain.Payment
	err := r.db.QueryRowContext(ctx, query, tenantID, gatewayTransactionID).Scan(
		&payment.ID,
		&payment.TenantID,
		&payment.InvoiceID,
		&payment.Amount,
		&payment.PaymentMethod,
		&payment.PaymentGateway,
		&payment.GatewayTransactionID,
		&payment.Status,
		&payment.ProcessedAt,
		&payment.Notes,
		&payment.CreatedAt,
		&payment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get payment by gateway transaction ID: %w", err)
	}

	return &payment, nil
}

// GetPendingPayments retrieves all pending payments for processing
func (r *PaymentRepositoryImpl) GetPendingPayments(ctx context.Context, tenantID uuid.UUID) ([]*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, invoice_id, amount, payment_method, payment_gateway,
			   gateway_transaction_id, status, processed_at, notes, created_at, updated_at
		FROM payments
		WHERE tenant_id = $1 AND status = 'pending'
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get pending payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.TenantID,
			&payment.InvoiceID,
			&payment.Amount,
			&payment.PaymentMethod,
			&payment.PaymentGateway,
			&payment.GatewayTransactionID,
			&payment.Status,
			&payment.ProcessedAt,
			&payment.Notes,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan pending payment: %w", err)
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}

// GetFailedPayments retrieves failed payments that might need retry
func (r *PaymentRepositoryImpl) GetFailedPayments(ctx context.Context, tenantID uuid.UUID, since time.Time) ([]*domain.Payment, error) {
	query := `
		SELECT id, tenant_id, invoice_id, amount, payment_method, payment_gateway,
			   gateway_transaction_id, status, processed_at, notes, created_at, updated_at
		FROM payments
		WHERE tenant_id = $1 AND status = 'failed' AND created_at >= $2
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, since)
	if err != nil {
		return nil, fmt.Errorf("failed to get failed payments: %w", err)
	}
	defer rows.Close()

	var payments []*domain.Payment
	for rows.Next() {
		var payment domain.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.TenantID,
			&payment.InvoiceID,
			&payment.Amount,
			&payment.PaymentMethod,
			&payment.PaymentGateway,
			&payment.GatewayTransactionID,
			&payment.Status,
			&payment.ProcessedAt,
			&payment.Notes,
			&payment.CreatedAt,
			&payment.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan failed payment: %w", err)
		}
		payments = append(payments, &payment)
	}

	return payments, nil
}

// GetRevenueByPeriod calculates revenue for a specific time period
func (r *PaymentRepositoryImpl) GetRevenueByPeriod(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (float64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM payments
		WHERE tenant_id = $1 
		  AND status = 'completed'
		  AND processed_at >= $2 
		  AND processed_at <= $3`

	var revenue float64
	err := r.db.QueryRowContext(ctx, query, tenantID, startDate, endDate).Scan(&revenue)
	if err != nil {
		return 0, fmt.Errorf("failed to get revenue by period: %w", err)
	}

	return revenue, nil
}

// GetMonthlyRevenue calculates monthly revenue for the current year
func (r *PaymentRepositoryImpl) GetMonthlyRevenue(ctx context.Context, tenantID uuid.UUID, year int) ([]float64, error) {
	query := `
		SELECT 
			EXTRACT(MONTH FROM processed_at) as month,
			COALESCE(SUM(amount), 0) as revenue
		FROM payments
		WHERE tenant_id = $1 
		  AND status = 'completed'
		  AND EXTRACT(YEAR FROM processed_at) = $2
		GROUP BY EXTRACT(MONTH FROM processed_at)
		ORDER BY month`

	rows, err := r.db.QueryContext(ctx, query, tenantID, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly revenue: %w", err)
	}
	defer rows.Close()

	// Initialize with 12 months of zero revenue
	monthlyRevenue := make([]float64, 12)
	
	for rows.Next() {
		var month int
		var revenue float64
		err := rows.Scan(&month, &revenue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan monthly revenue: %w", err)
		}
		
		// Month is 1-indexed, array is 0-indexed
		if month >= 1 && month <= 12 {
			monthlyRevenue[month-1] = revenue
		}
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating monthly revenue rows: %w", err)
	}

	return monthlyRevenue, nil
}