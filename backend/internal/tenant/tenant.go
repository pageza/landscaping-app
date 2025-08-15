package tenant

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// TenantContext holds tenant-specific information
type TenantContext struct {
	ID        string
	Subdomain string
	Plan      string
	Settings  map[string]interface{}
}

// GetTenantFromContext extracts tenant information from context
func GetTenantFromContext(ctx context.Context) (*TenantContext, error) {
	tenantID, ok := ctx.Value("tenant_id").(string)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// In a real implementation, you would fetch full tenant details from database
	// For now, return a basic tenant context
	return &TenantContext{
		ID:        tenantID,
		Subdomain: "placeholder",
		Plan:      "basic",
		Settings:  make(map[string]interface{}),
	}, nil
}

// SetTenantContext adds tenant information to context
func SetTenantContext(ctx context.Context, tenant *domain.Tenant) context.Context {
	ctx = context.WithValue(ctx, "tenant_id", tenant.ID.String())
	ctx = context.WithValue(ctx, "tenant_subdomain", tenant.Subdomain)
	ctx = context.WithValue(ctx, "tenant_plan", tenant.Plan)
	return ctx
}

// IsolationLevel represents the level of tenant isolation
type IsolationLevel string

const (
	// DatabaseIsolation - Each tenant has their own database
	DatabaseIsolation IsolationLevel = "database"
	
	// SchemaIsolation - Each tenant has their own schema within a shared database
	SchemaIsolation IsolationLevel = "schema"
	
	// RowIsolation - All tenants share the same database and schema, isolated by tenant_id
	RowIsolation IsolationLevel = "row"
)

// TenantManager handles tenant-related operations
type TenantManager struct {
	isolationLevel IsolationLevel
}

// NewTenantManager creates a new tenant manager
func NewTenantManager(isolationLevel string) *TenantManager {
	return &TenantManager{
		isolationLevel: IsolationLevel(isolationLevel),
	}
}

// GetDatabaseName returns the database name for a tenant based on isolation level
func (tm *TenantManager) GetDatabaseName(tenantID string) string {
	switch tm.isolationLevel {
	case DatabaseIsolation:
		return fmt.Sprintf("tenant_%s", tenantID)
	case SchemaIsolation, RowIsolation:
		return "landscaping_app" // Shared database
	default:
		return "landscaping_app"
	}
}

// GetSchemaName returns the schema name for a tenant based on isolation level
func (tm *TenantManager) GetSchemaName(tenantID string) string {
	switch tm.isolationLevel {
	case SchemaIsolation:
		return fmt.Sprintf("tenant_%s", tenantID)
	case DatabaseIsolation, RowIsolation:
		return "public" // Default schema
	default:
		return "public"
	}
}

// GetTablePrefix returns the table prefix for a tenant based on isolation level
func (tm *TenantManager) GetTablePrefix(tenantID string) string {
	switch tm.isolationLevel {
	case RowIsolation:
		// No prefix needed for row-level isolation as we filter by tenant_id
		return ""
	case DatabaseIsolation, SchemaIsolation:
		// No prefix needed as tables are isolated by database/schema
		return ""
	default:
		return ""
	}
}

// ValidateTenantAccess checks if a user has access to a specific tenant
func (tm *TenantManager) ValidateTenantAccess(userTenantID, requestedTenantID string) error {
	if userTenantID != requestedTenantID {
		return fmt.Errorf("access denied: user does not belong to requested tenant")
	}
	return nil
}

// AddTenantFilter adds tenant filtering to SQL queries for row-level isolation
func (tm *TenantManager) AddTenantFilter(query string, tenantID string) string {
	if tm.isolationLevel == RowIsolation {
		// This is a simple implementation. In practice, you'd use a proper SQL builder
		// or ORM that handles tenant filtering automatically
		return fmt.Sprintf("%s AND tenant_id = '%s'", query, tenantID)
	}
	return query
}

// TenantResolver resolves tenant information from various sources
type TenantResolver struct {
	tenantService TenantService
}

// NewTenantResolver creates a new tenant resolver
func NewTenantResolver(tenantService TenantService) *TenantResolver {
	return &TenantResolver{
		tenantService: tenantService,
	}
}

// ResolveFromSubdomain resolves tenant from subdomain
func (tr *TenantResolver) ResolveFromSubdomain(subdomain string) (*domain.Tenant, error) {
	enhancedTenant, err := tr.tenantService.GetTenantBySubdomain(context.Background(), subdomain)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant from subdomain: %w", err)
	}
	return &enhancedTenant.Tenant, nil
}

// ResolveFromHeader resolves tenant from HTTP header
func (tr *TenantResolver) ResolveFromHeader(tenantHeader string) (*domain.Tenant, error) {
	// Parse UUID from header
	tenantID, err := uuid.Parse(tenantHeader)
	if err != nil {
		return nil, fmt.Errorf("invalid tenant ID in header: %w", err)
	}

	enhancedTenant, err := tr.tenantService.GetTenant(context.Background(), tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant from header: %w", err)
	}
	return &enhancedTenant.Tenant, nil
}

// ResolveFromDomain resolves tenant from custom domain
func (tr *TenantResolver) ResolveFromDomain(domain string) (*domain.Tenant, error) {
	enhancedTenant, err := tr.tenantService.GetTenantByDomain(context.Background(), domain)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tenant from domain: %w", err)
	}
	return &enhancedTenant.Tenant, nil
}