package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// TenantRepository defines tenant storage operations
type TenantRepository interface {
	CreateTenant(ctx context.Context, tenant *domain.EnhancedTenant) error
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.EnhancedTenant, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*domain.EnhancedTenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*domain.EnhancedTenant, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) error
	ListTenants(ctx context.Context, limit, offset int) ([]*domain.EnhancedTenant, int64, error)
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error
	GetTenantUsage(ctx context.Context, tenantID uuid.UUID) (*TenantUsage, error)
}

// TenantUsage represents resource usage for a tenant
type TenantUsage struct {
	TenantID     uuid.UUID `json:"tenant_id"`
	UserCount    int       `json:"user_count"`
	CustomerCount int       `json:"customer_count"`
	StorageUsed  int64     `json:"storage_used_bytes"`
	LastUpdated  time.Time `json:"last_updated"`
}

// tenantRepository implements TenantRepository using PostgreSQL
type tenantRepository struct {
	db *Database
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *Database) TenantRepository {
	return &tenantRepository{db: db}
}

// CreateTenant creates a new tenant
func (r *tenantRepository) CreateTenant(ctx context.Context, tenant *domain.EnhancedTenant) error {
	query := `
		INSERT INTO tenants (
			id, name, subdomain, plan, status, domain, logo_url, theme_config,
			billing_settings, feature_flags, max_users, max_customers, 
			storage_quota_gb, trial_ends_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
	`

	themeConfigJSON, err := json.Marshal(tenant.ThemeConfig)
	if err != nil {
		return fmt.Errorf("failed to marshal theme config: %w", err)
	}

	billingSettingsJSON, err := json.Marshal(tenant.BillingSettings)
	if err != nil {
		return fmt.Errorf("failed to marshal billing settings: %w", err)
	}

	featureFlagsJSON, err := json.Marshal(tenant.FeatureFlags)
	if err != nil {
		return fmt.Errorf("failed to marshal feature flags: %w", err)
	}

	_, err = r.db.ExecContext(ctx, query,
		tenant.ID,
		tenant.Name,
		tenant.Subdomain,
		tenant.Plan,
		tenant.Status,
		tenant.Domain,
		tenant.LogoURL,
		themeConfigJSON,
		billingSettingsJSON,
		featureFlagsJSON,
		tenant.MaxUsers,
		tenant.MaxCustomers,
		tenant.StorageQuotaGB,
		tenant.TrialEndsAt,
		tenant.CreatedAt,
		tenant.UpdatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23505": // unique_violation
				if pqErr.Constraint == "tenants_subdomain_key" {
					return fmt.Errorf("subdomain already exists")
				}
				if pqErr.Constraint == "tenants_domain_key" {
					return fmt.Errorf("domain already exists")
				}
			}
		}
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	return nil
}

// GetTenant retrieves a tenant by ID
func (r *tenantRepository) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.EnhancedTenant, error) {
	query := `
		SELECT id, name, subdomain, plan, status, domain, logo_url, theme_config,
		       billing_settings, feature_flags, max_users, max_customers,
		       storage_quota_gb, trial_ends_at, created_at, updated_at
		FROM tenants WHERE id = $1
	`

	return r.scanTenant(r.db.QueryRowContext(ctx, query, tenantID))
}

// GetTenantBySubdomain retrieves a tenant by subdomain
func (r *tenantRepository) GetTenantBySubdomain(ctx context.Context, subdomain string) (*domain.EnhancedTenant, error) {
	query := `
		SELECT id, name, subdomain, plan, status, domain, logo_url, theme_config,
		       billing_settings, feature_flags, max_users, max_customers,
		       storage_quota_gb, trial_ends_at, created_at, updated_at
		FROM tenants WHERE subdomain = $1 AND status != 'deleted'
	`

	return r.scanTenant(r.db.QueryRowContext(ctx, query, subdomain))
}

// GetTenantByDomain retrieves a tenant by custom domain
func (r *tenantRepository) GetTenantByDomain(ctx context.Context, domain string) (*domain.EnhancedTenant, error) {
	query := `
		SELECT id, name, subdomain, plan, status, domain, logo_url, theme_config,
		       billing_settings, feature_flags, max_users, max_customers,
		       storage_quota_gb, trial_ends_at, created_at, updated_at
		FROM tenants WHERE domain = $1 AND status != 'deleted'
	`

	return r.scanTenant(r.db.QueryRowContext(ctx, query, domain))
}

// UpdateTenant updates tenant fields
func (r *tenantRepository) UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	// Build dynamic query
	setParts := make([]string, 0, len(updates)+1)
	args := make([]interface{}, 0, len(updates)+2)
	argIndex := 1

	for field, value := range updates {
		switch field {
		case "name", "domain", "logo_url", "plan", "status":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		case "theme_config", "billing_settings", "feature_flags":
			if jsonValue, err := json.Marshal(value); err == nil {
				setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
				args = append(args, jsonValue)
				argIndex++
			}
		case "max_users", "max_customers", "storage_quota_gb":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		case "trial_ends_at":
			setParts = append(setParts, fmt.Sprintf("%s = $%d", field, argIndex))
			args = append(args, value)
			argIndex++
		}
	}

	// Always update the updated_at timestamp
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", argIndex))
	args = append(args, time.Now())
	argIndex++

	// Add WHERE clause
	args = append(args, tenantID)

	query := fmt.Sprintf("UPDATE tenants SET %s WHERE id = $%d", 
		fmt.Sprintf(setParts[0], setParts[1:]...), argIndex)

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// ListTenants returns a paginated list of tenants
func (r *tenantRepository) ListTenants(ctx context.Context, limit, offset int) ([]*domain.EnhancedTenant, int64, error) {
	// Get total count
	countQuery := "SELECT COUNT(*) FROM tenants WHERE status != 'deleted'"
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	// Get paginated results
	query := `
		SELECT id, name, subdomain, plan, status, domain, logo_url, theme_config,
		       billing_settings, feature_flags, max_users, max_customers,
		       storage_quota_gb, trial_ends_at, created_at, updated_at
		FROM tenants 
		WHERE status != 'deleted'
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query tenants: %w", err)
	}
	defer rows.Close()

	var tenants []*domain.EnhancedTenant
	for rows.Next() {
		tenant, err := r.scanTenantRow(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant: %w", err)
		}
		tenants = append(tenants, tenant)
	}

	return tenants, total, nil
}

// DeleteTenant soft deletes a tenant by setting status to 'deleted'
func (r *tenantRepository) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	query := "UPDATE tenants SET status = 'deleted', updated_at = $1 WHERE id = $2"
	
	result, err := r.db.ExecContext(ctx, query, time.Now(), tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete tenant: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// UpdateTenantStatus updates the status of a tenant
func (r *tenantRepository) UpdateTenantStatus(ctx context.Context, tenantID uuid.UUID, status string) error {
	query := "UPDATE tenants SET status = $1, updated_at = $2 WHERE id = $3"
	
	result, err := r.db.ExecContext(ctx, query, status, time.Now(), tenantID)
	if err != nil {
		return fmt.Errorf("failed to update tenant status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("tenant not found")
	}

	return nil
}

// GetTenantUsage returns resource usage statistics for a tenant
func (r *tenantRepository) GetTenantUsage(ctx context.Context, tenantID uuid.UUID) (*TenantUsage, error) {
	query := `
		SELECT 
			COUNT(DISTINCT u.id) as user_count,
			COUNT(DISTINCT c.id) as customer_count
		FROM tenants t
		LEFT JOIN users u ON t.id = u.tenant_id AND u.status != 'deleted'
		LEFT JOIN customers c ON t.id = c.tenant_id AND c.status != 'deleted'
		WHERE t.id = $1
		GROUP BY t.id
	`

	var userCount, customerCount int
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&userCount, &customerCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant usage: %w", err)
	}

	// TODO: Calculate storage usage from file_attachments table
	// For now, return 0 as placeholder
	storageUsed := int64(0)

	return &TenantUsage{
		TenantID:      tenantID,
		UserCount:     userCount,
		CustomerCount: customerCount,
		StorageUsed:   storageUsed,
		LastUpdated:   time.Now(),
	}, nil
}

// Helper methods

func (r *tenantRepository) scanTenant(scanner interface {
	Scan(dest ...interface{}) error
}) (*domain.EnhancedTenant, error) {
	var tenant domain.EnhancedTenant
	var themeConfigJSON, billingSettingsJSON, featureFlagsJSON []byte

	err := scanner.Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Subdomain,
		&tenant.Plan,
		&tenant.Status,
		&tenant.Domain,
		&tenant.LogoURL,
		&themeConfigJSON,
		&billingSettingsJSON,
		&featureFlagsJSON,
		&tenant.MaxUsers,
		&tenant.MaxCustomers,
		&tenant.StorageQuotaGB,
		&tenant.TrialEndsAt,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("tenant not found")
	}
	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if len(themeConfigJSON) > 0 {
		if err := json.Unmarshal(themeConfigJSON, &tenant.ThemeConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal theme config: %w", err)
		}
	}

	if len(billingSettingsJSON) > 0 {
		if err := json.Unmarshal(billingSettingsJSON, &tenant.BillingSettings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal billing settings: %w", err)
		}
	}

	if len(featureFlagsJSON) > 0 {
		if err := json.Unmarshal(featureFlagsJSON, &tenant.FeatureFlags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal feature flags: %w", err)
		}
	}

	return &tenant, nil
}

func (r *tenantRepository) scanTenantRow(rows *sql.Rows) (*domain.EnhancedTenant, error) {
	var tenant domain.EnhancedTenant
	var themeConfigJSON, billingSettingsJSON, featureFlagsJSON []byte

	err := rows.Scan(
		&tenant.ID,
		&tenant.Name,
		&tenant.Subdomain,
		&tenant.Plan,
		&tenant.Status,
		&tenant.Domain,
		&tenant.LogoURL,
		&themeConfigJSON,
		&billingSettingsJSON,
		&featureFlagsJSON,
		&tenant.MaxUsers,
		&tenant.MaxCustomers,
		&tenant.StorageQuotaGB,
		&tenant.TrialEndsAt,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	// Parse JSON fields
	if len(themeConfigJSON) > 0 {
		if err := json.Unmarshal(themeConfigJSON, &tenant.ThemeConfig); err != nil {
			return nil, fmt.Errorf("failed to unmarshal theme config: %w", err)
		}
	}

	if len(billingSettingsJSON) > 0 {
		if err := json.Unmarshal(billingSettingsJSON, &tenant.BillingSettings); err != nil {
			return nil, fmt.Errorf("failed to unmarshal billing settings: %w", err)
		}
	}

	if len(featureFlagsJSON) > 0 {
		if err := json.Unmarshal(featureFlagsJSON, &tenant.FeatureFlags); err != nil {
			return nil, fmt.Errorf("failed to unmarshal feature flags: %w", err)
		}
	}

	return &tenant, nil
}