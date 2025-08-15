package domain

import (
	"time"

	"github.com/google/uuid"
)

// Enhanced Tenant model with SaaS features
type EnhancedTenant struct {
	Tenant
	Domain            *string                `json:"domain" db:"domain"`
	LogoURL           *string                `json:"logo_url" db:"logo_url"`
	ThemeConfig       map[string]interface{} `json:"theme_config" db:"theme_config"`
	BillingSettings   map[string]interface{} `json:"billing_settings" db:"billing_settings"`
	FeatureFlags      map[string]interface{} `json:"feature_flags" db:"feature_flags"`
	MaxUsers          int                    `json:"max_users" db:"max_users"`
	MaxCustomers      int                    `json:"max_customers" db:"max_customers"`
	StorageQuotaGB    int                    `json:"storage_quota_gb" db:"storage_quota_gb"`
	TrialEndsAt       *time.Time             `json:"trial_ends_at" db:"trial_ends_at"`
}

// Enhanced User model with security features
type EnhancedUser struct {
	User
	Phone                 *string                `json:"phone" db:"phone"`
	AvatarURL             *string                `json:"avatar_url" db:"avatar_url"`
	Timezone              string                 `json:"timezone" db:"timezone"`
	Language              string                 `json:"language" db:"language"`
	Permissions           []string               `json:"permissions" db:"permissions"`
	TwoFactorEnabled      bool                   `json:"two_factor_enabled" db:"two_factor_enabled"`
	TwoFactorSecret       *string                `json:"-" db:"two_factor_secret"`
	BackupCodes           []string               `json:"-" db:"backup_codes"`
	FailedLoginAttempts   int                    `json:"-" db:"failed_login_attempts"`
	LockedUntil           *time.Time             `json:"-" db:"locked_until"`
}

// Enhanced Customer model with business features
type EnhancedCustomer struct {
	Customer
	CompanyName             *string `json:"company_name" db:"company_name"`
	TaxID                   *string `json:"tax_id" db:"tax_id"`
	PreferredContactMethod  string  `json:"preferred_contact_method" db:"preferred_contact_method"`
	LeadSource              *string `json:"lead_source" db:"lead_source"`
	CustomerType            string  `json:"customer_type" db:"customer_type"`
	CreditLimit             *float64 `json:"credit_limit" db:"credit_limit"`
	PaymentTerms            int     `json:"payment_terms" db:"payment_terms"`
}

// Enhanced Property model with location and details
type EnhancedProperty struct {
	Property
	Latitude             *float64 `json:"latitude" db:"latitude"`
	Longitude            *float64 `json:"longitude" db:"longitude"`
	SquareFootage        *int     `json:"square_footage" db:"square_footage"`
	AccessInstructions   *string  `json:"access_instructions" db:"access_instructions"`
	GateCode             *string  `json:"gate_code" db:"gate_code"`
	SpecialInstructions  *string  `json:"special_instructions" db:"special_instructions"`
	PropertyValue        *float64 `json:"property_value" db:"property_value"`
}

// Enhanced Job model with operational features
type EnhancedJob struct {
	Job
	JobNumber          *string                `json:"job_number" db:"job_number"`
	RecurringSchedule  *string                `json:"recurring_schedule" db:"recurring_schedule"`
	ParentJobID        *uuid.UUID             `json:"parent_job_id" db:"parent_job_id"`
	WeatherDependent   bool                   `json:"weather_dependent" db:"weather_dependent"`
	RequiresEquipment  []uuid.UUID            `json:"requires_equipment" db:"requires_equipment"`
	CrewSize           int                    `json:"crew_size" db:"crew_size"`
	CompletionPhotos   []string               `json:"completion_photos" db:"completion_photos"`
	CustomerSignature  *string                `json:"customer_signature" db:"customer_signature"`
	GPSCheckIn         map[string]interface{} `json:"gps_check_in" db:"gps_check_in"`
	GPSCheckOut        map[string]interface{} `json:"gps_check_out" db:"gps_check_out"`
}

// API Key for external integrations
type APIKey struct {
	ID          uuid.UUID   `json:"id" db:"id"`
	TenantID    uuid.UUID   `json:"tenant_id" db:"tenant_id"`
	Name        string      `json:"name" db:"name"`
	KeyHash     string      `json:"-" db:"key_hash"`
	KeyPrefix   string      `json:"key_prefix" db:"key_prefix"`
	Permissions []string    `json:"permissions" db:"permissions"`
	LastUsedAt  *time.Time  `json:"last_used_at" db:"last_used_at"`
	ExpiresAt   *time.Time  `json:"expires_at" db:"expires_at"`
	Status      string      `json:"status" db:"status"`
	CreatedBy   *uuid.UUID  `json:"created_by" db:"created_by"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// User Session for authentication management
type UserSession struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	UserID       uuid.UUID              `json:"user_id" db:"user_id"`
	SessionToken string                 `json:"-" db:"session_token"`
	RefreshToken string                 `json:"-" db:"refresh_token"`
	DeviceInfo   map[string]interface{} `json:"device_info" db:"device_info"`
	IPAddress    *string                `json:"ip_address" db:"ip_address"`
	UserAgent    *string                `json:"user_agent" db:"user_agent"`
	ExpiresAt    time.Time              `json:"expires_at" db:"expires_at"`
	LastActivity time.Time              `json:"last_activity" db:"last_activity"`
	Status       string                 `json:"status" db:"status"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// Audit Log for security and compliance
type AuditLog struct {
	ID           uuid.UUID              `json:"id" db:"id"`
	TenantID     *uuid.UUID             `json:"tenant_id" db:"tenant_id"`
	UserID       *uuid.UUID             `json:"user_id" db:"user_id"`
	Action       string                 `json:"action" db:"action"`
	ResourceType string                 `json:"resource_type" db:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id" db:"resource_id"`
	OldValues    map[string]interface{} `json:"old_values" db:"old_values"`
	NewValues    map[string]interface{} `json:"new_values" db:"new_values"`
	IPAddress    *string                `json:"ip_address" db:"ip_address"`
	UserAgent    *string                `json:"user_agent" db:"user_agent"`
	CreatedAt    time.Time              `json:"created_at" db:"created_at"`
}

// Notification for user alerts
type Notification struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	TenantID  uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	UserID    *uuid.UUID             `json:"user_id" db:"user_id"`
	Type      string                 `json:"type" db:"type"`
	Title     string                 `json:"title" db:"title"`
	Message   string                 `json:"message" db:"message"`
	Data      map[string]interface{} `json:"data" db:"data"`
	ReadAt    *time.Time             `json:"read_at" db:"read_at"`
	ExpiresAt *time.Time             `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// Webhook for external integrations
type Webhook struct {
	ID              uuid.UUID              `json:"id" db:"id"`
	TenantID        uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name            string                 `json:"name" db:"name"`
	URL             string                 `json:"url" db:"url"`
	Secret          string                 `json:"-" db:"secret"`
	Events          []string               `json:"events" db:"events"`
	Headers         map[string]interface{} `json:"headers" db:"headers"`
	Status          string                 `json:"status" db:"status"`
	RetryCount      int                    `json:"retry_count" db:"retry_count"`
	LastSuccessAt   *time.Time             `json:"last_success_at" db:"last_success_at"`
	LastFailureAt   *time.Time             `json:"last_failure_at" db:"last_failure_at"`
	CreatedAt       time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at" db:"updated_at"`
}

// Webhook Delivery for tracking webhook calls
type WebhookDelivery struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	WebhookID      uuid.UUID              `json:"webhook_id" db:"webhook_id"`
	EventType      string                 `json:"event_type" db:"event_type"`
	Payload        map[string]interface{} `json:"payload" db:"payload"`
	ResponseStatus *int                   `json:"response_status" db:"response_status"`
	ResponseBody   *string                `json:"response_body" db:"response_body"`
	DeliveredAt    *time.Time             `json:"delivered_at" db:"delivered_at"`
	FailedAt       *time.Time             `json:"failed_at" db:"failed_at"`
	RetryCount     int                    `json:"retry_count" db:"retry_count"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// Crew for team management
type Crew struct {
	ID             uuid.UUID   `json:"id" db:"id"`
	TenantID       uuid.UUID   `json:"tenant_id" db:"tenant_id"`
	Name           string      `json:"name" db:"name"`
	Description    *string     `json:"description" db:"description"`
	Capacity       int         `json:"capacity" db:"capacity"`
	Specializations []string   `json:"specializations" db:"specializations"`
	EquipmentIDs   []uuid.UUID `json:"equipment_ids" db:"equipment_ids"`
	Status         string      `json:"status" db:"status"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
}

// Crew Member for crew composition
type CrewMember struct {
	ID       uuid.UUID  `json:"id" db:"id"`
	CrewID   uuid.UUID  `json:"crew_id" db:"crew_id"`
	UserID   uuid.UUID  `json:"user_id" db:"user_id"`
	Role     string     `json:"role" db:"role"`
	JoinedAt time.Time  `json:"joined_at" db:"joined_at"`
	LeftAt   *time.Time `json:"left_at" db:"left_at"`
}

// Quote for pricing estimates
type Quote struct {
	ID                 uuid.UUID  `json:"id" db:"id"`
	TenantID           uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	CustomerID         uuid.UUID  `json:"customer_id" db:"customer_id"`
	PropertyID         uuid.UUID  `json:"property_id" db:"property_id"`
	QuoteNumber        string     `json:"quote_number" db:"quote_number"`
	Title              string     `json:"title" db:"title"`
	Description        *string    `json:"description" db:"description"`
	Subtotal           float64    `json:"subtotal" db:"subtotal"`
	TaxRate            float64    `json:"tax_rate" db:"tax_rate"`
	TaxAmount          float64    `json:"tax_amount" db:"tax_amount"`
	TotalAmount        float64    `json:"total_amount" db:"total_amount"`
	Status             string     `json:"status" db:"status"`
	ValidUntil         *time.Time `json:"valid_until" db:"valid_until"`
	TermsAndConditions *string    `json:"terms_and_conditions" db:"terms_and_conditions"`
	Notes              *string    `json:"notes" db:"notes"`
	CreatedBy          *uuid.UUID `json:"created_by" db:"created_by"`
	ApprovedAt         *time.Time `json:"approved_at" db:"approved_at"`
	ApprovedBy         *uuid.UUID `json:"approved_by" db:"approved_by"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// Quote Service for quote line items
type QuoteService struct {
	ID          uuid.UUID `json:"id" db:"id"`
	QuoteID     uuid.UUID `json:"quote_id" db:"quote_id"`
	ServiceID   uuid.UUID `json:"service_id" db:"service_id"`
	Quantity    float64   `json:"quantity" db:"quantity"`
	UnitPrice   float64   `json:"unit_price" db:"unit_price"`
	TotalPrice  float64   `json:"total_price" db:"total_price"`
	Description *string   `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Schedule Template for recurring job patterns
type ScheduleTemplate struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	TenantID         uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	Name             string                 `json:"name" db:"name"`
	Description      *string                `json:"description" db:"description"`
	Frequency        string                 `json:"frequency" db:"frequency"`
	FrequencyConfig  map[string]interface{} `json:"frequency_config" db:"frequency_config"`
	ServiceIDs       []uuid.UUID            `json:"service_ids" db:"service_ids"`
	DefaultDuration  *int                   `json:"default_duration" db:"default_duration"`
	DefaultCrewSize  int                    `json:"default_crew_size" db:"default_crew_size"`
	Status           string                 `json:"status" db:"status"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// DTOs for API communication

// Auth DTOs
type LoginRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	TenantID  *uuid.UUID `json:"tenant_id,omitempty"`
	Subdomain *string `json:"subdomain,omitempty"`
}

type LoginResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	User         EnhancedUser `json:"user"`
	Tenant       EnhancedTenant `json:"tenant"`
	ExpiresIn    int       `json:"expires_in"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required"`
	LastName  string `json:"last_name" validate:"required"`
	Phone     *string `json:"phone,omitempty"`
	TenantName *string `json:"tenant_name,omitempty"`
	Subdomain  *string `json:"subdomain,omitempty"`
}

// Tenant DTOs
type CreateTenantRequest struct {
	Name      string                 `json:"name" validate:"required"`
	Subdomain string                 `json:"subdomain" validate:"required,alphanum,min=3,max=63"`
	Plan      string                 `json:"plan" validate:"required,oneof=basic premium enterprise"`
	Settings  map[string]interface{} `json:"settings,omitempty"`
}

type UpdateTenantRequest struct {
	Name         *string                `json:"name,omitempty"`
	Domain       *string                `json:"domain,omitempty"`
	LogoURL      *string                `json:"logo_url,omitempty"`
	ThemeConfig  map[string]interface{} `json:"theme_config,omitempty"`
	FeatureFlags map[string]interface{} `json:"feature_flags,omitempty"`
}

// User DTOs
type CreateUserRequest struct {
	Email       string   `json:"email" validate:"required,email"`
	FirstName   string   `json:"first_name" validate:"required"`
	LastName    string   `json:"last_name" validate:"required"`
	Role        string   `json:"role" validate:"required,oneof=admin user crew customer"`
	Phone       *string  `json:"phone,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

type UpdateUserRequest struct {
	FirstName   *string  `json:"first_name,omitempty"`
	LastName    *string  `json:"last_name,omitempty"`
	Phone       *string  `json:"phone,omitempty"`
	Role        *string  `json:"role,omitempty"`
	Status      *string  `json:"status,omitempty"`
	Timezone    *string  `json:"timezone,omitempty"`
	Language    *string  `json:"language,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// Customer DTOs
type CreateCustomerRequest struct {
	FirstName              string   `json:"first_name" validate:"required"`
	LastName               string   `json:"last_name" validate:"required"`
	Email                  *string  `json:"email,omitempty" validate:"omitempty,email"`
	Phone                  *string  `json:"phone,omitempty"`
	CompanyName            *string  `json:"company_name,omitempty"`
	AddressLine1           *string  `json:"address_line1,omitempty"`
	AddressLine2           *string  `json:"address_line2,omitempty"`
	City                   *string  `json:"city,omitempty"`
	State                  *string  `json:"state,omitempty"`
	ZipCode                *string  `json:"zip_code,omitempty"`
	PreferredContactMethod string   `json:"preferred_contact_method" validate:"oneof=email phone"`
	LeadSource             *string  `json:"lead_source,omitempty"`
	CustomerType           string   `json:"customer_type" validate:"oneof=residential commercial"`
}

// Property DTOs
type CreatePropertyRequest struct {
	CustomerID          uuid.UUID `json:"customer_id" validate:"required"`
	Name                string    `json:"name" validate:"required"`
	AddressLine1        string    `json:"address_line1" validate:"required"`
	AddressLine2        *string   `json:"address_line2,omitempty"`
	City                string    `json:"city" validate:"required"`
	State               string    `json:"state" validate:"required"`
	ZipCode             string    `json:"zip_code" validate:"required"`
	PropertyType        string    `json:"property_type" validate:"required,oneof=residential commercial"`
	LotSize             *float64  `json:"lot_size,omitempty"`
	SquareFootage       *int      `json:"square_footage,omitempty"`
	AccessInstructions  *string   `json:"access_instructions,omitempty"`
	GateCode            *string   `json:"gate_code,omitempty"`
	SpecialInstructions *string   `json:"special_instructions,omitempty"`
}

// Job DTOs
type CreateJobRequest struct {
	CustomerID         uuid.UUID   `json:"customer_id" validate:"required"`
	PropertyID         uuid.UUID   `json:"property_id" validate:"required"`
	Title              string      `json:"title" validate:"required"`
	Description        *string     `json:"description,omitempty"`
	Priority           string      `json:"priority" validate:"oneof=low medium high urgent"`
	ScheduledDate      *time.Time  `json:"scheduled_date,omitempty"`
	ScheduledTime      *string     `json:"scheduled_time,omitempty"`
	EstimatedDuration  *int        `json:"estimated_duration,omitempty"`
	ServiceIDs         []uuid.UUID `json:"service_ids,omitempty"`
	AssignedUserID     *uuid.UUID  `json:"assigned_user_id,omitempty"`
	CrewSize           int         `json:"crew_size" validate:"min=1"`
	WeatherDependent   bool        `json:"weather_dependent"`
	RequiresEquipment  []uuid.UUID `json:"requires_equipment,omitempty"`
}

type UpdateJobRequest struct {
	Title             *string     `json:"title,omitempty"`
	Description       *string     `json:"description,omitempty"`
	Status            *string     `json:"status,omitempty"`
	Priority          *string     `json:"priority,omitempty"`
	ScheduledDate     *time.Time  `json:"scheduled_date,omitempty"`
	ScheduledTime     *string     `json:"scheduled_time,omitempty"`
	EstimatedDuration *int        `json:"estimated_duration,omitempty"`
	AssignedUserID    *uuid.UUID  `json:"assigned_user_id,omitempty"`
	CrewSize          *int        `json:"crew_size,omitempty"`
	Notes             *string     `json:"notes,omitempty"`
}

// Common response structures
type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Validation and permissions constants
const (
	// User roles
	RoleSuperAdmin = "super_admin"
	RoleOwner      = "owner"
	RoleAdmin      = "admin"
	RoleUser       = "user"
	RoleCrew       = "crew"
	RoleCustomer   = "customer"

	// Permissions
	PermissionTenantManage    = "tenant:manage"
	PermissionUserManage      = "user:manage"
	PermissionCustomerManage  = "customer:manage"
	PermissionPropertyManage  = "property:manage"
	PermissionJobManage       = "job:manage"
	PermissionJobAssign       = "job:assign"
	PermissionInvoiceManage   = "invoice:manage"
	PermissionPaymentManage   = "payment:manage"
	PermissionEquipmentManage = "equipment:manage"
	PermissionReportView      = "report:view"
	PermissionWebhookManage   = "webhook:manage"
	PermissionAuditView       = "audit:view"

	// Job statuses
	JobStatusPending     = "pending"
	JobStatusScheduled   = "scheduled"
	JobStatusInProgress  = "in_progress"
	JobStatusCompleted   = "completed"
	JobStatusCancelled   = "cancelled"
	JobStatusOnHold      = "on_hold"

	// Tenant statuses
	TenantStatusActive    = "active"
	TenantStatusSuspended = "suspended"
	TenantStatusCancelled = "cancelled"
	TenantStatusTrial     = "trial"

	// User statuses
	UserStatusActive    = "active"
	UserStatusInactive  = "inactive"
	UserStatusSuspended = "suspended"
	UserStatusPending   = "pending"
)