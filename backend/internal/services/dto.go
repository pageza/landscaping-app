package services

import (
	"time"

	"github.com/google/uuid"
)

// Common filter and pagination types
type BaseFilter struct {
	Page     int
	PerPage  int
	SortBy   string
	SortDesc bool
	Search   string
}

type TimeRange struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address,omitempty"`
}

// Authentication DTOs
type TwoFactorSetup struct {
	Secret    string   `json:"secret"`
	QRCode    string   `json:"qr_code"`
	BackupCodes []string `json:"backup_codes"`
}

// Tenant DTOs
type TenantFilter struct {
	BaseFilter
	Status string    `json:"status,omitempty"`
	Plan   string    `json:"plan,omitempty"`
	Since  *time.Time `json:"since,omitempty"`
}

type TenantUsage struct {
	Users         int     `json:"users"`
	MaxUsers      int     `json:"max_users"`
	Customers     int     `json:"customers"`
	MaxCustomers  int     `json:"max_customers"`
	StorageUsedGB float64 `json:"storage_used_gb"`
	StorageQuotaGB int    `json:"storage_quota_gb"`
	JobsThisMonth int     `json:"jobs_this_month"`
	RevenueThisMonth float64 `json:"revenue_this_month"`
}

type TenantBilling struct {
	Plan               string     `json:"plan"`
	BillingCycle       string     `json:"billing_cycle"`
	NextBillingDate    time.Time  `json:"next_billing_date"`
	CurrentPeriodStart time.Time  `json:"current_period_start"`
	CurrentPeriodEnd   time.Time  `json:"current_period_end"`
	Amount             float64    `json:"amount"`
	Currency           string     `json:"currency"`
	Status             string     `json:"status"`
	TrialEndsAt        *time.Time `json:"trial_ends_at,omitempty"`
}

type TenantBranding struct {
	LogoURL      string                 `json:"logo_url"`
	PrimaryColor string                 `json:"primary_color"`
	SecondaryColor string               `json:"secondary_color"`
	CompanyName  string                 `json:"company_name"`
	Domain       string                 `json:"domain"`
	ThemeConfig  map[string]interface{} `json:"theme_config"`
}

// User DTOs
type UserFilter struct {
	BaseFilter
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

type UserProfile struct {
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Phone     *string `json:"phone,omitempty"`
	Timezone  string  `json:"timezone"`
	Language  string  `json:"language"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}

// Customer DTOs
type CustomerFilter struct {
	BaseFilter
	CustomerType string `json:"customer_type,omitempty"`
	Status       string `json:"status,omitempty"`
	LeadSource   string `json:"lead_source,omitempty"`
}

type CustomerUpdateRequest struct {
	FirstName              *string  `json:"first_name,omitempty"`
	LastName               *string  `json:"last_name,omitempty"`
	Email                  *string  `json:"email,omitempty"`
	Phone                  *string  `json:"phone,omitempty"`
	CompanyName            *string  `json:"company_name,omitempty"`
	AddressLine1           *string  `json:"address_line1,omitempty"`
	AddressLine2           *string  `json:"address_line2,omitempty"`
	City                   *string  `json:"city,omitempty"`
	State                  *string  `json:"state,omitempty"`
	ZipCode                *string  `json:"zip_code,omitempty"`
	PreferredContactMethod *string  `json:"preferred_contact_method,omitempty"`
	CustomerType           *string  `json:"customer_type,omitempty"`
	CreditLimit            *float64 `json:"credit_limit,omitempty"`
	PaymentTerms           *int     `json:"payment_terms,omitempty"`
	Notes                  *string  `json:"notes,omitempty"`
}

type CustomerSummary struct {
	TotalJobs     int     `json:"total_jobs"`
	CompletedJobs int     `json:"completed_jobs"`
	TotalRevenue  float64 `json:"total_revenue"`
	LastJobDate   *time.Time `json:"last_job_date"`
	AverageJobValue float64 `json:"average_job_value"`
	PaymentHistory []PaymentHistoryEntry `json:"payment_history"`
}

type PaymentHistoryEntry struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
	Status string    `json:"status"`
}

// Property DTOs
type PropertyFilter struct {
	BaseFilter
	PropertyType string    `json:"property_type,omitempty"`
	CustomerID   *uuid.UUID `json:"customer_id,omitempty"`
	City         string    `json:"city,omitempty"`
	State        string    `json:"state,omitempty"`
}

type PropertyUpdateRequest struct {
	Name                *string  `json:"name,omitempty"`
	AddressLine1        *string  `json:"address_line1,omitempty"`
	AddressLine2        *string  `json:"address_line2,omitempty"`
	City                *string  `json:"city,omitempty"`
	State               *string  `json:"state,omitempty"`
	ZipCode             *string  `json:"zip_code,omitempty"`
	PropertyType        *string  `json:"property_type,omitempty"`
	LotSize             *float64 `json:"lot_size,omitempty"`
	SquareFootage       *int     `json:"square_footage,omitempty"`
	AccessInstructions  *string  `json:"access_instructions,omitempty"`
	GateCode            *string  `json:"gate_code,omitempty"`
	SpecialInstructions *string  `json:"special_instructions,omitempty"`
	PropertyValue       *float64 `json:"property_value,omitempty"`
	Notes               *string  `json:"notes,omitempty"`
}

type PropertyDetails struct {
	LotSize       *float64 `json:"lot_size,omitempty"`
	SquareFootage *int     `json:"square_footage,omitempty"`
	PropertyType  string   `json:"property_type"`
	Accessibility string   `json:"accessibility,omitempty"`
}

type PropertyValuation struct {
	EstimatedValue float64   `json:"estimated_value"`
	LastUpdated    time.Time `json:"last_updated"`
	Confidence     float64   `json:"confidence"`
	Source         string    `json:"source"`
}

type PropertyRouteInfo struct {
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	AddressLine1       string    `json:"address_line1"`
	City               string    `json:"city"`
	State              string    `json:"state"`
	ZipCode            string    `json:"zip_code"`
	Latitude           *float64  `json:"latitude"`
	Longitude          *float64  `json:"longitude"`
	AccessInstructions *string   `json:"access_instructions"`
	GateCode           *string   `json:"gate_code"`
}

type PropertyValueAnalytics struct {
	TotalProperties     int      `json:"total_properties"`
	PropertiesWithValue int      `json:"properties_with_value"`
	AvgValue           *float64  `json:"avg_value"`
	MinValue           *float64  `json:"min_value"`
	MaxValue           *float64  `json:"max_value"`
	MedianValue        *float64  `json:"median_value"`
}

// Service DTOs
type ServiceFilter struct {
	BaseFilter
	Category string `json:"category,omitempty"`
	Status   string `json:"status,omitempty"`
}

type ServiceCreateRequest struct {
	Name            string   `json:"name" validate:"required"`
	Description     *string  `json:"description,omitempty"`
	Category        string   `json:"category" validate:"required"`
	BasePrice       *float64 `json:"base_price,omitempty"`
	Unit            *string  `json:"unit,omitempty"`
	DurationMinutes *int     `json:"duration_minutes,omitempty"`
}

type ServiceUpdateRequest struct {
	Name            *string  `json:"name,omitempty"`
	Description     *string  `json:"description,omitempty"`
	Category        *string  `json:"category,omitempty"`
	BasePrice       *float64 `json:"base_price,omitempty"`
	Unit            *string  `json:"unit,omitempty"`
	DurationMinutes *int     `json:"duration_minutes,omitempty"`
	Status          *string  `json:"status,omitempty"`
}

type ServicePricing struct {
	BasePrice    float64 `json:"base_price"`
	Quantity     float64 `json:"quantity"`
	Adjustments  []PriceAdjustment `json:"adjustments"`
	TotalPrice   float64 `json:"total_price"`
}

type PriceAdjustment struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	IsPercentage bool   `json:"is_percentage"`
}

// Job DTOs
type JobFilter struct {
	BaseFilter
	Status         string     `json:"status,omitempty"`
	Priority       string     `json:"priority,omitempty"`
	AssignedUserID *uuid.UUID `json:"assigned_user_id,omitempty"`
	CustomerID     *uuid.UUID `json:"customer_id,omitempty"`
	PropertyID     *uuid.UUID `json:"property_id,omitempty"`
	ScheduledStart *time.Time `json:"scheduled_start,omitempty"`
	ScheduledEnd   *time.Time `json:"scheduled_end,omitempty"`
}

type JobStartDetails struct {
	StartTime time.Time              `json:"start_time"`
	GPSLocation *Location            `json:"gps_location,omitempty"`
	Notes     *string                `json:"notes,omitempty"`
	Photos    []string               `json:"photos,omitempty"`
	Weather   map[string]interface{} `json:"weather,omitempty"`
}

type JobCompletionDetails struct {
	EndTime         time.Time `json:"end_time"`
	GPSLocation     *Location `json:"gps_location,omitempty"`
	CompletionNotes *string   `json:"completion_notes,omitempty"`
	Photos          []string  `json:"photos,omitempty"`
	CustomerSatisfaction *int `json:"customer_satisfaction,omitempty"`
	RequiresFollowUp bool     `json:"requires_follow_up"`
}

type JobServiceUpdate struct {
	ServiceID   uuid.UUID `json:"service_id"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Description *string   `json:"description,omitempty"`
}

type JobPhoto struct {
	Data        []byte `json:"data"`
	ContentType string `json:"content_type"`
	Description string `json:"description,omitempty"`
}

type ScheduledJob struct {
	JobID         uuid.UUID  `json:"job_id"`
	Title         string     `json:"title"`
	CustomerName  string     `json:"customer_name"`
	PropertyAddress string   `json:"property_address"`
	ScheduledDate time.Time  `json:"scheduled_date"`
	ScheduledTime *string    `json:"scheduled_time,omitempty"`
	Duration      *int       `json:"duration,omitempty"`
	Status        string     `json:"status"`
	AssignedUser  *string    `json:"assigned_user,omitempty"`
	Priority      string     `json:"priority"`
}

type CalendarEvent struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Location    *Location `json:"location,omitempty"`
}

type RecurringJobRequest struct {
	BaseJobID    uuid.UUID              `json:"base_job_id"`
	Frequency    string                 `json:"frequency"`
	FrequencyConfig map[string]interface{} `json:"frequency_config"`
	StartDate    time.Time              `json:"start_date"`
	EndDate      *time.Time             `json:"end_date,omitempty"`
	MaxOccurrences *int                 `json:"max_occurrences,omitempty"`
}

type RecurringJobSeries struct {
	ID              uuid.UUID   `json:"id"`
	BaseJobID       uuid.UUID   `json:"base_job_id"`
	Frequency       string      `json:"frequency"`
	NextOccurrence  time.Time   `json:"next_occurrence"`
	JobsCreated     int         `json:"jobs_created"`
	UpcomingJobs    []uuid.UUID `json:"upcoming_jobs"`
}

type RouteOptimization struct {
	OptimizedRoute []RouteStop `json:"optimized_route"`
	TotalDistance  float64     `json:"total_distance"`
	TotalDuration  int         `json:"total_duration_minutes"`
	Savings        float64     `json:"savings_percent"`
}

type RouteStop struct {
	JobID       uuid.UUID `json:"job_id"`
	Address     string    `json:"address"`
	Sequence    int       `json:"sequence"`
	ArrivalTime time.Time `json:"arrival_time"`
	Duration    int       `json:"duration_minutes"`
	Distance    float64   `json:"distance_from_previous"`
}

// Quote DTOs
type QuoteFilter struct {
	BaseFilter
	Status     string     `json:"status,omitempty"`
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	PropertyID *uuid.UUID `json:"property_id,omitempty"`
	ValidUntil *time.Time `json:"valid_until,omitempty"`
}

type QuoteCreateRequest struct {
	CustomerID         uuid.UUID `json:"customer_id" validate:"required"`
	PropertyID         uuid.UUID `json:"property_id" validate:"required"`
	Title              string    `json:"title" validate:"required"`
	Description        *string   `json:"description,omitempty"`
	Services           []QuoteServiceRequest `json:"services" validate:"required"`
	ValidUntil         *time.Time `json:"valid_until,omitempty"`
	TermsAndConditions *string   `json:"terms_and_conditions,omitempty"`
	Notes              *string   `json:"notes,omitempty"`
}

type QuoteUpdateRequest struct {
	Title              *string   `json:"title,omitempty"`
	Description        *string   `json:"description,omitempty"`
	Services           []QuoteServiceRequest `json:"services,omitempty"`
	ValidUntil         *time.Time `json:"valid_until,omitempty"`
	TermsAndConditions *string   `json:"terms_and_conditions,omitempty"`
	Notes              *string   `json:"notes,omitempty"`
}

type QuoteServiceRequest struct {
	ServiceID   uuid.UUID `json:"service_id"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Description *string   `json:"description,omitempty"`
}

type QuoteSendOptions struct {
	Email       string  `json:"email"`
	Subject     *string `json:"subject,omitempty"`
	Message     *string `json:"message,omitempty"`
	CCEmails    []string `json:"cc_emails,omitempty"`
	IncludePDF  bool    `json:"include_pdf"`
}

type QuoteGenerationRequest struct {
	CustomerID    uuid.UUID `json:"customer_id"`
	PropertyID    uuid.UUID `json:"property_id"`
	Description   string    `json:"description"`
	Requirements  []string  `json:"requirements,omitempty"`
	Budget        *float64  `json:"budget,omitempty"`
	Timeline      *string   `json:"timeline,omitempty"`
	Images        []string  `json:"images,omitempty"`
}

type QuoteGenerationResponse struct {
	RecommendedServices []ServiceRecommendation `json:"recommended_services"`
	EstimatedTotal      float64                 `json:"estimated_total"`
	TimelineEstimate    string                  `json:"timeline_estimate"`
	Notes               string                  `json:"notes"`
	Confidence          float64                 `json:"confidence"`
}

type ServiceRecommendation struct {
	ServiceID   uuid.UUID `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	TotalPrice  float64   `json:"total_price"`
	Reasoning   string    `json:"reasoning"`
}

// Invoice DTOs
type InvoiceFilter struct {
	BaseFilter
	Status     string     `json:"status,omitempty"`
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	JobID      *uuid.UUID `json:"job_id,omitempty"`
	DueDate    *time.Time `json:"due_date,omitempty"`
	Overdue    bool       `json:"overdue,omitempty"`
}

type InvoiceCreateRequest struct {
	CustomerID  uuid.UUID `json:"customer_id" validate:"required"`
	JobID       *uuid.UUID `json:"job_id,omitempty"`
	Services    []InvoiceServiceRequest `json:"services" validate:"required"`
	TaxRate     float64   `json:"tax_rate"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	Notes       *string   `json:"notes,omitempty"`
}

type InvoiceUpdateRequest struct {
	Services []InvoiceServiceRequest `json:"services,omitempty"`
	TaxRate  *float64  `json:"tax_rate,omitempty"`
	DueDate  *time.Time `json:"due_date,omitempty"`
	Notes    *string   `json:"notes,omitempty"`
}

type InvoiceServiceRequest struct {
	ServiceID   uuid.UUID `json:"service_id"`
	Quantity    float64   `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	Description *string   `json:"description,omitempty"`
}

type InvoiceSendOptions struct {
	Email    string   `json:"email"`
	Subject  *string  `json:"subject,omitempty"`
	Message  *string  `json:"message,omitempty"`
	CCEmails []string `json:"cc_emails,omitempty"`
}

// Payment DTOs
type PaymentFilter struct {
	BaseFilter
	Status     string     `json:"status,omitempty"`
	Method     string     `json:"method,omitempty"`
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	InvoiceID  *uuid.UUID `json:"invoice_id,omitempty"`
	DateRange  *TimeRange `json:"date_range,omitempty"`
}

type PaymentProcessRequest struct {
	InvoiceID     uuid.UUID `json:"invoice_id"`
	Amount        float64   `json:"amount"`
	PaymentMethod string    `json:"payment_method"`
	PaymentToken  string    `json:"payment_token,omitempty"`
	CustomerID    uuid.UUID `json:"customer_id"`
	Description   *string   `json:"description,omitempty"`
}

type PaymentRefund struct {
	RefundID      string    `json:"refund_id"`
	Amount        float64   `json:"amount"`
	Status        string    `json:"status"`
	ProcessedAt   time.Time `json:"processed_at"`
	RefundMethod  string    `json:"refund_method"`
	EstimatedArrival *time.Time `json:"estimated_arrival,omitempty"`
}

// PaymentMethod is defined in enhanced_billing_service.go

type PaymentSummary struct {
	TotalAmount      float64 `json:"total_amount"`
	SuccessfulCount  int     `json:"successful_count"`
	FailedCount      int     `json:"failed_count"`
	RefundedAmount   float64 `json:"refunded_amount"`
	PendingAmount    float64 `json:"pending_amount"`
	AverageTransaction float64 `json:"average_transaction"`
}

// Equipment DTOs
type EquipmentFilter struct {
	BaseFilter
	Type         string `json:"type,omitempty"`
	Status       string `json:"status,omitempty"`
	Available    bool   `json:"available,omitempty"`
	Maintenance  bool   `json:"maintenance,omitempty"`
}

type EquipmentCreateRequest struct {
	Name                string     `json:"name" validate:"required"`
	Type                string     `json:"type" validate:"required"`
	Model               *string    `json:"model,omitempty"`
	SerialNumber        *string    `json:"serial_number,omitempty"`
	PurchaseDate        *time.Time `json:"purchase_date,omitempty"`
	PurchasePrice       *float64   `json:"purchase_price,omitempty"`
	MaintenanceSchedule *string    `json:"maintenance_schedule,omitempty"`
	Notes               *string    `json:"notes,omitempty"`
}

type EquipmentUpdateRequest struct {
	Name                *string    `json:"name,omitempty"`
	Type                *string    `json:"type,omitempty"`
	Model               *string    `json:"model,omitempty"`
	SerialNumber        *string    `json:"serial_number,omitempty"`
	PurchaseDate        *time.Time `json:"purchase_date,omitempty"`
	PurchasePrice       *float64   `json:"purchase_price,omitempty"`
	Status              *string    `json:"status,omitempty"`
	MaintenanceSchedule *string    `json:"maintenance_schedule,omitempty"`
	Notes               *string    `json:"notes,omitempty"`
}

type MaintenanceScheduleRequest struct {
	Type        string    `json:"type" validate:"required"`
	ScheduledDate time.Time `json:"scheduled_date" validate:"required"`
	Description string    `json:"description"`
	EstimatedCost *float64 `json:"estimated_cost,omitempty"`
	Priority    string    `json:"priority"`
}

type MaintenanceRecord struct {
	ID            uuid.UUID `json:"id"`
	EquipmentID   uuid.UUID `json:"equipment_id"`
	Type          string    `json:"type"`
	PerformedDate time.Time `json:"performed_date"`
	Description   string    `json:"description"`
	Cost          *float64  `json:"cost,omitempty"`
	PerformedBy   *uuid.UUID `json:"performed_by,omitempty"`
	Notes         *string   `json:"notes,omitempty"`
}

type MaintenanceSchedule struct {
	ID            uuid.UUID `json:"id"`
	EquipmentID   uuid.UUID `json:"equipment_id"`
	EquipmentName string    `json:"equipment_name"`
	Type          string    `json:"type"`
	ScheduledDate time.Time `json:"scheduled_date"`
	Description   string    `json:"description"`
	Priority      string    `json:"priority"`
	Status        string    `json:"status"`
}

// Crew DTOs
type CrewFilter struct {
	BaseFilter
	Status         string `json:"status,omitempty"`
	Specialization string `json:"specialization,omitempty"`
	Available      bool   `json:"available,omitempty"`
}

type CrewCreateRequest struct {
	Name            string      `json:"name" validate:"required"`
	Description     *string     `json:"description,omitempty"`
	Capacity        int         `json:"capacity" validate:"required,min=1"`
	Specializations []string    `json:"specializations,omitempty"`
	EquipmentIDs    []uuid.UUID `json:"equipment_ids,omitempty"`
}

type CrewUpdateRequest struct {
	Name            *string     `json:"name,omitempty"`
	Description     *string     `json:"description,omitempty"`
	Capacity        *int        `json:"capacity,omitempty"`
	Specializations []string    `json:"specializations,omitempty"`
	EquipmentIDs    []uuid.UUID `json:"equipment_ids,omitempty"`
	Status          *string     `json:"status,omitempty"`
}

type CrewMemberDetails struct {
	UserID    uuid.UUID  `json:"user_id"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Role      string     `json:"role"`
	JoinedAt  time.Time  `json:"joined_at"`
	LeftAt    *time.Time `json:"left_at,omitempty"`
}

type CrewScheduleEntry struct {
	Date        time.Time   `json:"date"`
	JobID       *uuid.UUID  `json:"job_id,omitempty"`
	JobTitle    *string     `json:"job_title,omitempty"`
	StartTime   *time.Time  `json:"start_time,omitempty"`
	EndTime     *time.Time  `json:"end_time,omitempty"`
	Status      string      `json:"status"`
	Availability string     `json:"availability"`
}

type CrewPerformance struct {
	JobsCompleted    int     `json:"jobs_completed"`
	AverageRating    float64 `json:"average_rating"`
	OnTimePercentage float64 `json:"on_time_percentage"`
	RevenueGenerated float64 `json:"revenue_generated"`
	EfficiencyScore  float64 `json:"efficiency_score"`
}

// Notification DTOs - NotificationFilter is defined in notification_service.go

type NotificationRequest struct {
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	CustomerID *uuid.UUID             `json:"customer_id,omitempty"`
	Email      string                 `json:"email,omitempty"`
	Phone      string                 `json:"phone,omitempty"`
	Type       string                 `json:"type"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Template   string                 `json:"template,omitempty"`
	Data       map[string]interface{} `json:"data,omitempty"`
	ExpiresAt  *time.Time             `json:"expires_at,omitempty"`
	Channels   []string               `json:"channels,omitempty"`
	Priority   string                 `json:"priority,omitempty"`
	ScheduleAt *time.Time             `json:"schedule_at,omitempty"`
}

// BulkNotificationRequest is defined in notification_service.go

// NotificationTemplate is defined in notification_service.go

// Webhook DTOs
type WebhookFilter struct {
	BaseFilter
	Status string `json:"status,omitempty"`
	Event  string `json:"event,omitempty"`
}

type WebhookCreateRequest struct {
	Name    string                 `json:"name" validate:"required"`
	URL     string                 `json:"url" validate:"required,url"`
	Events  []string               `json:"events" validate:"required"`
	Secret  string                 `json:"secret,omitempty"`
	Headers map[string]interface{} `json:"headers,omitempty"`
}

type WebhookUpdateRequest struct {
	Name    *string                `json:"name,omitempty"`
	URL     *string                `json:"url,omitempty"`
	Events  []string               `json:"events,omitempty"`
	Headers map[string]interface{} `json:"headers,omitempty"`
	Status  *string                `json:"status,omitempty"`
}

type DeliveryFilter struct {
	BaseFilter
	Status    string     `json:"status,omitempty"`
	EventType string     `json:"event_type,omitempty"`
	DateRange *TimeRange `json:"date_range,omitempty"`
}

// Audit DTOs - AuditFilter is defined in audit_service.go

type AuditLogRequest struct {
	UserID       *uuid.UUID             `json:"user_id,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type"`
	ResourceID   *uuid.UUID             `json:"resource_id,omitempty"`
	OldValues    map[string]interface{} `json:"old_values,omitempty"`
	NewValues    map[string]interface{} `json:"new_values,omitempty"`
	IPAddress    *string                `json:"ip_address,omitempty"`
	UserAgent    *string                `json:"user_agent,omitempty"`
	ErrorMessage *string                `json:"error_message,omitempty"`
	Duration     *time.Duration         `json:"duration,omitempty"`
}

type ActivityFilter struct {
	BaseFilter
	Action    string     `json:"action,omitempty"`
	DateRange *TimeRange `json:"date_range,omitempty"`
}

// ComplianceReport is defined later in this file (comprehensive version)

// Report DTOs
type DashboardData struct {
	TotalRevenue       float64              `json:"total_revenue"`
	RevenueGrowth      float64              `json:"revenue_growth"`
	TotalJobs          int                  `json:"total_jobs"`
	CompletedJobs      int                  `json:"completed_jobs"`
	PendingJobs        int                  `json:"pending_jobs"`
	TotalCustomers     int                  `json:"total_customers"`
	NewCustomers       int                  `json:"new_customers"`
	OverdueInvoices    int                  `json:"overdue_invoices"`
	RecentJobs         []ScheduledJob       `json:"recent_jobs"`
	RevenueChart       []RevenueDataPoint   `json:"revenue_chart"`
	JobStatusChart     []JobStatusCount     `json:"job_status_chart"`
	TopServices        []ServiceUsage       `json:"top_services"`
}

type RevenueDataPoint struct {
	Date   time.Time `json:"date"`
	Amount float64   `json:"amount"`
}

type JobStatusCount struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type ServiceUsage struct {
	ServiceID   uuid.UUID `json:"service_id"`
	ServiceName string    `json:"service_name"`
	JobCount    int       `json:"job_count"`
	Revenue     float64   `json:"revenue"`
}

type RevenueFilter struct {
	TimeRange
	CustomerID *uuid.UUID `json:"customer_id,omitempty"`
	ServiceID  *uuid.UUID `json:"service_id,omitempty"`
	Grouping   string     `json:"grouping"` // daily, weekly, monthly, yearly
}

// RevenueReport is defined in enhanced_billing_service.go

type RevenueBreakdown struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type ProfitLossFilter struct {
	TimeRange
	IncludeCosts bool `json:"include_costs"`
}

type ProfitLossReport struct {
	Period   TimeRange `json:"period"`
	Revenue  float64   `json:"revenue"`
	Costs    float64   `json:"costs"`
	Profit   float64   `json:"profit"`
	Margin   float64   `json:"margin"`
	Details  []ProfitLossLine `json:"details"`
}

type ProfitLossLine struct {
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Type     string  `json:"type"` // revenue or cost
}

type JobReportFilter struct {
	TimeRange
	Status       string     `json:"status,omitempty"`
	CustomerID   *uuid.UUID `json:"customer_id,omitempty"`
	AssignedUserID *uuid.UUID `json:"assigned_user_id,omitempty"`
}

type JobsReport struct {
	Period       TimeRange        `json:"period"`
	TotalJobs    int              `json:"total_jobs"`
	Completed    int              `json:"completed"`
	Cancelled    int              `json:"cancelled"`
	InProgress   int              `json:"in_progress"`
	Pending      int              `json:"pending"`
	StatusTrend  []JobStatusTrend `json:"status_trend"`
	AverageDuration int           `json:"average_duration"`
	CompletionRate float64       `json:"completion_rate"`
}

type JobStatusTrend struct {
	Date     time.Time `json:"date"`
	Status   string    `json:"status"`
	Count    int       `json:"count"`
}

type CustomerReportFilter struct {
	TimeRange
	CustomerType string `json:"customer_type,omitempty"`
	Status       string `json:"status,omitempty"`
}

type CustomersReport struct {
	Period        TimeRange           `json:"period"`
	TotalCustomers int                `json:"total_customers"`
	NewCustomers  int                 `json:"new_customers"`
	RetainedCustomers int             `json:"retained_customers"`
	ChurnRate     float64             `json:"churn_rate"`
	AverageValue  float64             `json:"average_value"`
	GrowthTrend   []CustomerGrowth    `json:"growth_trend"`
	TypeBreakdown []CustomerTypeCount `json:"type_breakdown"`
}

type CustomerGrowth struct {
	Date  time.Time `json:"date"`
	New   int       `json:"new"`
	Total int       `json:"total"`
}

type CustomerTypeCount struct {
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type PerformanceReportFilter struct {
	TimeRange
	UserID *uuid.UUID `json:"user_id,omitempty"`
	CrewID *uuid.UUID `json:"crew_id,omitempty"`
}

type PerformanceReport struct {
	Period           TimeRange            `json:"period"`
	ProductivityScore float64             `json:"productivity_score"`
	QualityScore     float64             `json:"quality_score"`
	EfficiencyScore  float64             `json:"efficiency_score"`
	UserPerformance  []UserPerformance   `json:"user_performance"`
	CrewPerformance  []CrewPerformance   `json:"crew_performance"`
	Trends           []PerformanceTrend  `json:"trends"`
}

type UserPerformance struct {
	UserID       uuid.UUID `json:"user_id"`
	Name         string    `json:"name"`
	JobsCompleted int      `json:"jobs_completed"`
	AverageRating float64  `json:"average_rating"`
	Efficiency   float64   `json:"efficiency"`
}

type PerformanceTrend struct {
	Date  time.Time `json:"date"`
	Score float64   `json:"score"`
	Type  string    `json:"type"`
}

// Communication DTOs
type EmailRequest struct {
	To          []string `json:"to"`
	CC          []string `json:"cc,omitempty"`
	BCC         []string `json:"bcc,omitempty"`
	Subject     string   `json:"subject"`
	Body        string   `json:"body"`
	IsHTML      bool     `json:"is_html"`
	Attachments []Attachment `json:"attachments,omitempty"`
}

type SMSRequest struct {
	To      []string `json:"to"`
	Message string   `json:"message"`
}

type PushNotificationRequest struct {
	UserIDs []uuid.UUID            `json:"user_ids"`
	Title   string                 `json:"title"`
	Body    string                 `json:"body"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type Attachment struct {
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

// LLM DTOs
type LLMOptions struct {
	Model       string  `json:"model,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
}

type JobAnalysis struct {
	EstimatedDuration int      `json:"estimated_duration"`
	RequiredServices  []string `json:"required_services"`
	Complexity        string   `json:"complexity"`
	RequiredSkills    []string `json:"required_skills"`
	RequiredEquipment []string `json:"required_equipment"`
	Seasonality       string   `json:"seasonality,omitempty"`
	WeatherDependency bool     `json:"weather_dependency"`
}

// Schedule DTOs
type ScheduleOptimizationRequest struct {
	Jobs        []uuid.UUID `json:"jobs"`
	TimeRange   TimeRange   `json:"time_range"`
	Constraints []ScheduleConstraint `json:"constraints,omitempty"`
	Objectives  []string    `json:"objectives"`
}

type ScheduleConstraint struct {
	Type       string      `json:"type"`
	Parameters interface{} `json:"parameters"`
}

type ScheduleOptimizationResult struct {
	Schedule     []ScheduleSlot `json:"schedule"`
	Metrics      ScheduleMetrics `json:"metrics"`
	Improvements []string       `json:"improvements"`
}

type ScheduleSlot struct {
	JobID     uuid.UUID `json:"job_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	UserID    uuid.UUID `json:"user_id"`
	CrewID    *uuid.UUID `json:"crew_id,omitempty"`
}

type ScheduleMetrics struct {
	Utilization      float64 `json:"utilization"`
	TravelTime       int     `json:"travel_time_minutes"`
	OverTimeHours    int     `json:"overtime_hours"`
	CustomerSatisfaction float64 `json:"customer_satisfaction"`
}

type AvailabilityRequest struct {
	UserIDs   []uuid.UUID `json:"user_ids,omitempty"`
	CrewIDs   []uuid.UUID `json:"crew_ids,omitempty"`
	TimeRange TimeRange   `json:"time_range"`
	JobType   string      `json:"job_type,omitempty"`
}

type AvailabilityResponse struct {
	AvailableSlots []AvailabilitySlot `json:"available_slots"`
	Conflicts      []AvailabilityConflict `json:"conflicts"`
}

type AvailabilitySlot struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	CrewID    *uuid.UUID `json:"crew_id,omitempty"`
	StartTime time.Time  `json:"start_time"`
	EndTime   time.Time  `json:"end_time"`
	Capacity  int        `json:"capacity"`
}

type AvailabilityConflict struct {
	ResourceID   uuid.UUID `json:"resource_id"`
	ResourceType string    `json:"resource_type"`
	ConflictTime TimeRange `json:"conflict_time"`
	Reason       string    `json:"reason"`
}

// Performance filter
type PerformanceFilter struct {
	TimeRange
	MetricType string `json:"metric_type,omitempty"`
}

// Payment Integration DTOs
type PaymentIntentRequest struct {
	Amount              int64                  `json:"amount"`
	Currency            string                 `json:"currency"`
	Description         string                 `json:"description"`
	CustomerID          string                 `json:"customer_id"`
	CustomerEmail       string                 `json:"customer_email"`
	PaymentMethodTypes  []string               `json:"payment_method_types"`
	CaptureMethod       string                 `json:"capture_method"`
	SetupFutureUsage    string                 `json:"setup_future_usage"`
	Metadata            map[string]string      `json:"metadata"`
	ShippingAddress     *PaymentAddress        `json:"shipping_address"`
}

type PaymentIntentResponse struct {
	ID           string    `json:"id"`
	ClientSecret string    `json:"client_secret"`
	Status       string    `json:"status"`
	Amount       int64     `json:"amount"`
	Currency     string    `json:"currency"`
	CreatedAt    time.Time `json:"created_at"`
}

type ConfirmPaymentRequest struct {
	PaymentIntentID string `json:"payment_intent_id"`
	PaymentMethodID string `json:"payment_method_id"`
}

type CapturePaymentRequest struct {
	PaymentIntentID string `json:"payment_intent_id"`
	Amount          *int64 `json:"amount"`
}

type PaymentResponse struct {
	ID              string            `json:"id"`
	Status          string            `json:"status"`
	Amount          int64             `json:"amount"`
	Currency        string            `json:"currency"`
	ProcessedAt     time.Time         `json:"processed_at"`
	PaymentMethodID string            `json:"payment_method_id"`
	ReceiptURL      string            `json:"receipt_url"`
	Metadata        map[string]string `json:"metadata"`
}

type RefundRequest struct {
	PaymentID string            `json:"payment_id"`
	Amount    *int64            `json:"amount"`
	Reason    string            `json:"reason"`
	Metadata  map[string]string `json:"metadata"`
}

type RefundResponse struct {
	ID          string    `json:"id"`
	Status      string    `json:"status"`
	Amount      int64     `json:"amount"`
	Currency    string    `json:"currency"`
	PaymentID   string    `json:"payment_id"`
	ProcessedAt time.Time `json:"processed_at"`
	Reason      string    `json:"reason"`
}

type CreateCustomerRequest struct {
	Email       string            `json:"email"`
	Name        string            `json:"name"`
	Phone       string            `json:"phone"`
	Description string            `json:"description"`
	Address     *PaymentAddress   `json:"address"`
	Metadata    map[string]string `json:"metadata"`
}

type CustomerResponse struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	Name        string    `json:"name"`
	Phone       string    `json:"phone"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type PaymentStatusResponse struct {
	ID              string    `json:"id"`
	Status          string    `json:"status"`
	Amount          int64     `json:"amount"`
	Currency        string    `json:"currency"`
	ProcessedAt     time.Time `json:"processed_at"`
	PaymentMethodID string    `json:"payment_method_id"`
	FailureReason   string    `json:"failure_reason"`
	ReceiptURL      string    `json:"receipt_url"`
}

type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt time.Time              `json:"created_at"`
}

type FeeCalculation struct {
	GrossAmount   float64 `json:"gross_amount"`
	ProcessingFee float64 `json:"processing_fee"`
	NetAmount     float64 `json:"net_amount"`
	FeePercentage float64 `json:"fee_percentage"`
	FixedFee      float64 `json:"fixed_fee"`
	Currency      string  `json:"currency"`
}

type PaymentAddress struct {
	Line1      string `json:"line1"`
	Line2      string `json:"line2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// Equipment Analytics DTOs
type EquipmentUtilization struct {
	EquipmentID           uuid.UUID `json:"equipment_id"`
	EquipmentName         string    `json:"equipment_name"`
	Period                TimeRange `json:"period"`
	TotalJobs             int       `json:"total_jobs"`
	TotalHours            int       `json:"total_hours"`
	TotalRevenue          float64   `json:"total_revenue"`
	UtilizationPercentage float64   `json:"utilization_percentage"`
	AverageJobDuration    int       `json:"average_job_duration"`
	RevenuePerHour        float64   `json:"revenue_per_hour"`
}

type MaintenanceCostAnalysis struct {
	EquipmentID      uuid.UUID              `json:"equipment_id"`
	EquipmentName    string                 `json:"equipment_name"`
	Period           TimeRange              `json:"period"`
	TotalCost        float64                `json:"total_cost"`
	AverageCost      float64                `json:"average_cost"`
	MaintenanceCount int                    `json:"maintenance_count"`
	CostByType       map[string]float64     `json:"cost_by_type"`
	Records          []*MaintenanceRecord   `json:"records"`
}

type MaintenancePrediction struct {
	EquipmentID         uuid.UUID                    `json:"equipment_id"`
	EquipmentName       string                       `json:"equipment_name"`
	NextMaintenanceDate *time.Time                   `json:"next_maintenance_date"`
	Confidence          float64                      `json:"confidence"`
	Recommendations     []MaintenanceRecommendation  `json:"recommendations"`
	GeneratedAt         time.Time                    `json:"generated_at"`
}

type MaintenanceRecommendation struct {
	Type          string  `json:"type"`
	Priority      string  `json:"priority"`
	Description   string  `json:"description"`
	EstimatedCost float64 `json:"estimated_cost"`
}

type EquipmentPerformanceMetrics struct {
	EquipmentID      uuid.UUID `json:"equipment_id"`
	EquipmentName    string    `json:"equipment_name"`
	Period           TimeRange `json:"period"`
	UtilizationRate  float64   `json:"utilization_rate"`
	Revenue          float64   `json:"revenue"`
	MaintenanceCost  float64   `json:"maintenance_cost"`
	ROI              float64   `json:"roi"`
	EfficiencyScore  float64   `json:"efficiency_score"`
	DowntimeHours    int       `json:"downtime_hours"`
	ReliabilityScore float64   `json:"reliability_score"`
}

// Subscription and Billing DTOs
type SubscriptionCreateRequest struct {
	TenantID        uuid.UUID `json:"tenant_id" validate:"required"`
	PlanID          string    `json:"plan_id" validate:"required"`
	BillingCycle    string    `json:"billing_cycle" validate:"required"`
	Amount          float64   `json:"amount" validate:"required,min=0"`
	Currency        string    `json:"currency"`
	PaymentMethodID string    `json:"payment_method_id"`
	TrialDays       int       `json:"trial_days"`
}

type SubscriptionUpdateRequest struct {
	PlanID       *string  `json:"plan_id,omitempty"`
	BillingCycle *string  `json:"billing_cycle,omitempty"`
	Amount       *float64 `json:"amount,omitempty"`
	Status       *string  `json:"status,omitempty"`
}

type SubscriptionFilter struct {
	BaseFilter
	TenantID        *uuid.UUID `json:"tenant_id,omitempty"`
	Status          string     `json:"status,omitempty"`
	PlanID          string     `json:"plan_id,omitempty"`
	NextBillingDate *time.Time `json:"next_billing_date,omitempty"`
}

// CreateSubscriptionRequest is defined later in this file (simpler version)

type SubscriptionResponse struct {
	ID                 string    `json:"id"`
	Status             string    `json:"status"`
	CustomerID         string    `json:"customer_id"`
	PriceID           string    `json:"price_id"`
	CurrentPeriodStart time.Time `json:"current_period_start"`
	CurrentPeriodEnd   time.Time `json:"current_period_end"`
	CreatedAt         time.Time `json:"created_at"`
}

type UsageRecord struct {
	ID             uuid.UUID              `json:"id"`
	SubscriptionID uuid.UUID              `json:"subscription_id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	UsageType      string                 `json:"usage_type"`
	Quantity       int                    `json:"quantity"`
	Timestamp      time.Time              `json:"timestamp"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type UsageSummary struct {
	SubscriptionID uuid.UUID              `json:"subscription_id"`
	Period         string                 `json:"period"`
	TotalUsage     map[string]int         `json:"total_usage"`
	Trends         []UsageTrend           `json:"trends"`
	Overage        map[string]int         `json:"overage"`
	Metadata       map[string]interface{} `json:"metadata"`
}

type UsageTrend struct {
	Date     time.Time `json:"date"`
	UsageType string   `json:"usage_type"`
	Quantity int       `json:"quantity"`
}

// Missing types needed by service interfaces

type PaymentMethod struct {
	ID            string                 `json:"id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	Type          string                 `json:"type"`
	IsDefault     bool                   `json:"is_default"`
	Status        string                 `json:"status"`
	Metadata      map[string]interface{} `json:"metadata"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

type RevenueReport struct {
	TotalRevenue    float64            `json:"total_revenue"`
	Period          string             `json:"period"`
	RevenueByMonth  []RevenueDataPoint `json:"revenue_by_month"`
	TopServices     []ServiceRevenue   `json:"top_services"`
	GeneratedAt     time.Time          `json:"generated_at"`
}

type ServiceRevenue struct {
	ServiceID   uuid.UUID `json:"service_id"`
	ServiceName string    `json:"service_name"`
	Revenue     float64   `json:"revenue"`
	JobCount    int       `json:"job_count"`
}

type PaginatedTenantsResponse struct {
	Tenants    []TenantSummary `json:"tenants"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

type TenantSummary struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Plan      string    `json:"plan"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Service interface support types

type BillingSubscription struct {
	ID       uuid.UUID `json:"id"`
	TenantID uuid.UUID `json:"tenant_id"`
	Plan     string    `json:"plan"`
	Status   string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateSubscriptionRequest struct {
	TenantID uuid.UUID `json:"tenant_id"`
	Plan     string    `json:"plan"`
}

type UpdateSubscriptionRequest struct {
	Plan   *string `json:"plan,omitempty"`
	Status *string `json:"status,omitempty"`
}

type MetricsFilter struct {
	StartTime time.Time         `json:"start_time"`
	EndTime   time.Time         `json:"end_time"`
	Tags      map[string]string `json:"tags,omitempty"`
}

type MetricsResponse struct {
	Metrics []MetricPoint `json:"metrics"`
	Count   int          `json:"count"`
}

type MetricPoint struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Tags      map[string]string `json:"tags"`
	Timestamp time.Time         `json:"timestamp"`
}

type ExternalServiceHealth struct {
	Name         string        `json:"name"`
	Status       string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	LastCheck    time.Time     `json:"last_check"`
	ErrorRate    float64       `json:"error_rate"`
}

type PerformanceOverview struct {
	ResponseTime    float64 `json:"response_time_ms"`
	Throughput      float64 `json:"throughput_rps"`
	ErrorRate       float64 `json:"error_rate_percentage"`
	MemoryUsage     float64 `json:"memory_usage_percentage"`
	CPUUsage        float64 `json:"cpu_usage_percentage"`
	DiskUsage       float64 `json:"disk_usage_percentage"`
}

type SecurityOverview struct {
	ThreatLevel       string    `json:"threat_level"`
	IncidentsToday    int       `json:"incidents_today"`
	VulnerabilitiesOpen int     `json:"vulnerabilities_open"`
	LastSecurityScan  time.Time `json:"last_security_scan"`
	ComplianceStatus  string    `json:"compliance_status"`
}

type DatabaseConnection interface {
	// Basic interface for database connections
	Close() error
}

type TenantLimits struct {
	MaxUsers       int `json:"max_users"`
	MaxCustomers   int `json:"max_customers"`
	MaxJobs        int `json:"max_jobs"`
	StorageGB      int `json:"storage_gb"`
}

type EfficiencyMetrics struct {
	Overall    float64 `json:"overall"`
	ByService  map[string]float64 `json:"by_service"`
	Trends     []EfficiencyTrend `json:"trends"`
}

type EfficiencyTrend struct {
	Date       time.Time `json:"date"`
	Efficiency float64   `json:"efficiency"`
}

type BillingPeriod struct {
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}

type RevenueReportFilter struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	ServiceID *string   `json:"service_id,omitempty"`
}

type ComplianceReport struct {
	Period        TimeRange                `json:"period"`
	OverallScore  int                      `json:"overall_score"`
	Violations    []ComplianceViolation    `json:"violations"`
	Recommendations []string               `json:"recommendations"`
	GeneratedAt   time.Time                `json:"generated_at"`
}

type ComplianceViolation struct {
	Type        string    `json:"type"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	Count       int       `json:"count"`
	FirstSeen   time.Time `json:"first_seen"`
	LastSeen    time.Time `json:"last_seen"`
}