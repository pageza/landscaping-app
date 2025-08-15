package tenant

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	
	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/repository"
)

// TenantService handles tenant business logic
type TenantService interface {
	CreateTenant(ctx context.Context, req *CreateTenantRequest) (*domain.EnhancedTenant, error)
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.EnhancedTenant, error)
	GetTenantBySubdomain(ctx context.Context, subdomain string) (*domain.EnhancedTenant, error)
	GetTenantByDomain(ctx context.Context, domain string) (*domain.EnhancedTenant, error)
	UpdateTenant(ctx context.Context, tenantID uuid.UUID, req *UpdateTenantRequest) (*domain.EnhancedTenant, error)
	ListTenants(ctx context.Context, page, perPage int) (*PaginatedTenantsResponse, error)
	SuspendTenant(ctx context.Context, tenantID uuid.UUID) error
	ReactivateTenant(ctx context.Context, tenantID uuid.UUID) error
	DeleteTenant(ctx context.Context, tenantID uuid.UUID) error
	GetTenantUsage(ctx context.Context, tenantID uuid.UUID) (*TenantUsageResponse, error)
	ValidateTenantLimits(ctx context.Context, tenantID uuid.UUID) error
}

// CreateTenantRequest represents a request to create a new tenant
type CreateTenantRequest struct {
	Name         string                 `json:"name" validate:"required,min=2,max=100"`
	Subdomain    string                 `json:"subdomain" validate:"required,alphanum,min=3,max=63"`
	Plan         string                 `json:"plan" validate:"required,oneof=basic premium enterprise"`
	Domain       *string                `json:"domain,omitempty" validate:"omitempty,fqdn"`
	LogoURL      *string                `json:"logo_url,omitempty" validate:"omitempty,url"`
	ThemeConfig  map[string]interface{} `json:"theme_config,omitempty"`
	FeatureFlags map[string]interface{} `json:"feature_flags,omitempty"`
	TrialDays    *int                   `json:"trial_days,omitempty" validate:"omitempty,min=0,max=365"`
}

// UpdateTenantRequest represents a request to update a tenant
type UpdateTenantRequest struct {
	Name         *string                `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Domain       *string                `json:"domain,omitempty" validate:"omitempty,fqdn"`
	LogoURL      *string                `json:"logo_url,omitempty" validate:"omitempty,url"`
	Plan         *string                `json:"plan,omitempty" validate:"omitempty,oneof=basic premium enterprise"`
	ThemeConfig  map[string]interface{} `json:"theme_config,omitempty"`
	FeatureFlags map[string]interface{} `json:"feature_flags,omitempty"`
	MaxUsers     *int                   `json:"max_users,omitempty" validate:"omitempty,min=1"`
	MaxCustomers *int                   `json:"max_customers,omitempty" validate:"omitempty,min=1"`
	StorageQuota *int                   `json:"storage_quota_gb,omitempty" validate:"omitempty,min=1"`
}

// PaginatedTenantsResponse represents a paginated list of tenants
type PaginatedTenantsResponse struct {
	Tenants    []*domain.EnhancedTenant `json:"tenants"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PerPage    int                      `json:"per_page"`
	TotalPages int                      `json:"total_pages"`
}

// TenantUsageResponse represents tenant usage statistics
type TenantUsageResponse struct {
	TenantID      uuid.UUID               `json:"tenant_id"`
	UserCount     int                     `json:"user_count"`
	CustomerCount int                     `json:"customer_count"`
	StorageUsed   int64                   `json:"storage_used_bytes"`
	Limits        TenantLimits            `json:"limits"`
	Usage         map[string]interface{}  `json:"usage"`
	LastUpdated   time.Time               `json:"last_updated"`
}

// TenantLimits represents the limits for a tenant based on their plan
type TenantLimits struct {
	MaxUsers     int   `json:"max_users"`
	MaxCustomers int   `json:"max_customers"`
	StorageGB    int   `json:"storage_gb"`
}

// tenantService implements TenantService
type tenantService struct {
	tenantRepo repository.TenantRepository
	db         *sql.DB
}

// NewTenantService creates a new tenant service
func NewTenantService(tenantRepo repository.TenantRepository, db *sql.DB) TenantService {
	return &tenantService{
		tenantRepo: tenantRepo,
		db:         db,
	}
}

// CreateTenant creates a new tenant
func (s *tenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*domain.EnhancedTenant, error) {
	// Validate subdomain uniqueness
	existing, err := s.tenantRepo.GetTenantBySubdomain(ctx, req.Subdomain)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("subdomain already exists")
	}

	// Check custom domain if provided
	if req.Domain != nil {
		existing, err := s.tenantRepo.GetTenantByDomain(ctx, *req.Domain)
		if err == nil && existing != nil {
			return nil, fmt.Errorf("domain already exists")
		}
	}

	// Set default limits based on plan
	limits := s.getDefaultLimitsForPlan(req.Plan)

	// Create tenant
	tenant := &domain.EnhancedTenant{
		Tenant: domain.Tenant{
			ID:        uuid.New(),
			Name:      req.Name,
			Subdomain: req.Subdomain,
			Plan:      req.Plan,
			Status:    domain.TenantStatusTrial,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Domain:           req.Domain,
		LogoURL:          req.LogoURL,
		ThemeConfig:      req.ThemeConfig,
		BillingSettings:  make(map[string]interface{}),
		FeatureFlags:     req.FeatureFlags,
		MaxUsers:         limits.MaxUsers,
		MaxCustomers:     limits.MaxCustomers,
		StorageQuotaGB:   limits.StorageGB,
	}

	// Set trial end date if specified
	if req.TrialDays != nil && *req.TrialDays > 0 {
		trialEnd := time.Now().AddDate(0, 0, *req.TrialDays)
		tenant.TrialEndsAt = &trialEnd
	}

	// Initialize theme config with defaults
	if tenant.ThemeConfig == nil {
		tenant.ThemeConfig = s.getDefaultThemeConfig()
	}

	// Initialize feature flags with defaults
	if tenant.FeatureFlags == nil {
		tenant.FeatureFlags = s.getDefaultFeatureFlags(req.Plan)
	}

	err = s.tenantRepo.CreateTenant(ctx, tenant)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	return tenant, nil
}

// GetTenant retrieves a tenant by ID
func (s *tenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.EnhancedTenant, error) {
	return s.tenantRepo.GetTenant(ctx, tenantID)
}

// GetTenantBySubdomain retrieves a tenant by subdomain
func (s *tenantService) GetTenantBySubdomain(ctx context.Context, subdomain string) (*domain.EnhancedTenant, error) {
	return s.tenantRepo.GetTenantBySubdomain(ctx, subdomain)
}

// GetTenantByDomain retrieves a tenant by custom domain
func (s *tenantService) GetTenantByDomain(ctx context.Context, domain string) (*domain.EnhancedTenant, error) {
	return s.tenantRepo.GetTenantByDomain(ctx, domain)
}

// UpdateTenant updates a tenant
func (s *tenantService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, req *UpdateTenantRequest) (*domain.EnhancedTenant, error) {
	// Check if tenant exists
	existing, err := s.tenantRepo.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	// Build updates map
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Domain != nil {
		// Check if domain is already in use by another tenant
		if *req.Domain != "" {
			domainTenant, err := s.tenantRepo.GetTenantByDomain(ctx, *req.Domain)
			if err == nil && domainTenant != nil && domainTenant.ID != tenantID {
				return nil, fmt.Errorf("domain already in use")
			}
		}
		updates["domain"] = *req.Domain
	}
	if req.LogoURL != nil {
		updates["logo_url"] = *req.LogoURL
	}
	if req.Plan != nil {
		updates["plan"] = *req.Plan
		// Update limits when plan changes
		limits := s.getDefaultLimitsForPlan(*req.Plan)
		updates["max_users"] = limits.MaxUsers
		updates["max_customers"] = limits.MaxCustomers
		updates["storage_quota_gb"] = limits.StorageGB
	}
	if req.ThemeConfig != nil {
		updates["theme_config"] = req.ThemeConfig
	}
	if req.FeatureFlags != nil {
		updates["feature_flags"] = req.FeatureFlags
	}
	if req.MaxUsers != nil {
		updates["max_users"] = *req.MaxUsers
	}
	if req.MaxCustomers != nil {
		updates["max_customers"] = *req.MaxCustomers
	}
	if req.StorageQuota != nil {
		updates["storage_quota_gb"] = *req.StorageQuota
	}

	if len(updates) == 0 {
		return existing, nil
	}

	err = s.tenantRepo.UpdateTenant(ctx, tenantID, updates)
	if err != nil {
		return nil, fmt.Errorf("failed to update tenant: %w", err)
	}

	// Return updated tenant
	return s.tenantRepo.GetTenant(ctx, tenantID)
}

// ListTenants returns a paginated list of tenants
func (s *tenantService) ListTenants(ctx context.Context, page, perPage int) (*PaginatedTenantsResponse, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	offset := (page - 1) * perPage
	tenants, total, err := s.tenantRepo.ListTenants(ctx, perPage, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants: %w", err)
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return &PaginatedTenantsResponse{
		Tenants:    tenants,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

// SuspendTenant suspends a tenant
func (s *tenantService) SuspendTenant(ctx context.Context, tenantID uuid.UUID) error {
	return s.tenantRepo.UpdateTenantStatus(ctx, tenantID, domain.TenantStatusSuspended)
}

// ReactivateTenant reactivates a suspended tenant
func (s *tenantService) ReactivateTenant(ctx context.Context, tenantID uuid.UUID) error {
	return s.tenantRepo.UpdateTenantStatus(ctx, tenantID, domain.TenantStatusActive)
}

// DeleteTenant soft deletes a tenant
func (s *tenantService) DeleteTenant(ctx context.Context, tenantID uuid.UUID) error {
	return s.tenantRepo.DeleteTenant(ctx, tenantID)
}

// GetTenantUsage returns usage statistics for a tenant
func (s *tenantService) GetTenantUsage(ctx context.Context, tenantID uuid.UUID) (*TenantUsageResponse, error) {
	tenant, err := s.tenantRepo.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("tenant not found: %w", err)
	}

	usage, err := s.tenantRepo.GetTenantUsage(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant usage: %w", err)
	}

	return &TenantUsageResponse{
		TenantID:      tenantID,
		UserCount:     usage.UserCount,
		CustomerCount: usage.CustomerCount,
		StorageUsed:   usage.StorageUsed,
		Limits: TenantLimits{
			MaxUsers:     tenant.MaxUsers,
			MaxCustomers: tenant.MaxCustomers,
			StorageGB:    tenant.StorageQuotaGB,
		},
		Usage: map[string]interface{}{
			"users_percentage":     float64(usage.UserCount) / float64(tenant.MaxUsers) * 100,
			"customers_percentage": float64(usage.CustomerCount) / float64(tenant.MaxCustomers) * 100,
			"storage_percentage":   float64(usage.StorageUsed) / float64(tenant.StorageQuotaGB*1024*1024*1024) * 100,
		},
		LastUpdated: usage.LastUpdated,
	}, nil
}

// ValidateTenantLimits checks if a tenant is within their usage limits
func (s *tenantService) ValidateTenantLimits(ctx context.Context, tenantID uuid.UUID) error {
	usageResp, err := s.GetTenantUsage(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to get tenant usage: %w", err)
	}

	var violations []string

	if usageResp.UserCount >= usageResp.Limits.MaxUsers {
		violations = append(violations, fmt.Sprintf("user limit exceeded (%d/%d)", usageResp.UserCount, usageResp.Limits.MaxUsers))
	}

	if usageResp.CustomerCount >= usageResp.Limits.MaxCustomers {
		violations = append(violations, fmt.Sprintf("customer limit exceeded (%d/%d)", usageResp.CustomerCount, usageResp.Limits.MaxCustomers))
	}

	storageGB := usageResp.StorageUsed / (1024 * 1024 * 1024)
	if storageGB >= int64(usageResp.Limits.StorageGB) {
		violations = append(violations, fmt.Sprintf("storage limit exceeded (%dGB/%dGB)", storageGB, usageResp.Limits.StorageGB))
	}

	if len(violations) > 0 {
		return fmt.Errorf("tenant limits exceeded: %v", violations)
	}

	return nil
}

// Helper methods

func (s *tenantService) getDefaultLimitsForPlan(plan string) TenantLimits {
	switch plan {
	case "basic":
		return TenantLimits{
			MaxUsers:     5,
			MaxCustomers: 100,
			StorageGB:    5,
		}
	case "premium":
		return TenantLimits{
			MaxUsers:     25,
			MaxCustomers: 1000,
			StorageGB:    50,
		}
	case "enterprise":
		return TenantLimits{
			MaxUsers:     100,
			MaxCustomers: 10000,
			StorageGB:    500,
		}
	default:
		return TenantLimits{
			MaxUsers:     5,
			MaxCustomers: 100,
			StorageGB:    5,
		}
	}
}

func (s *tenantService) getDefaultThemeConfig() map[string]interface{} {
	return map[string]interface{}{
		"primary_color":   "#3B82F6",
		"secondary_color": "#64748B",
		"accent_color":    "#10B981",
		"font_family":     "Inter",
		"logo_position":   "left",
		"dark_mode":       false,
	}
}

func (s *tenantService) getDefaultFeatureFlags(plan string) map[string]interface{} {
	baseFeatures := map[string]interface{}{
		"customer_portal":     true,
		"mobile_app":          true,
		"basic_reports":       true,
		"email_notifications": true,
	}

	switch plan {
	case "premium":
		baseFeatures["advanced_reports"] = true
		baseFeatures["api_access"] = true
		baseFeatures["webhook_integrations"] = true
	case "enterprise":
		baseFeatures["advanced_reports"] = true
		baseFeatures["api_access"] = true
		baseFeatures["webhook_integrations"] = true
		baseFeatures["custom_branding"] = true
		baseFeatures["sso_integration"] = true
		baseFeatures["priority_support"] = true
	}

	return baseFeatures
}