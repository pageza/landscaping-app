package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// SaaS Service Interfaces - This file defines all SaaS-related service interfaces

// SupportService defines support ticket management operations
type SupportService interface {
	// Ticket management
	GetTickets(ctx context.Context, filter *TicketFilter) (*TicketList, error)
	CreateTicket(ctx context.Context, req *CreateTicketRequest) (*SupportTicket, error)
	GetTicket(ctx context.Context, ticketID, tenantID uuid.UUID) (*SupportTicket, error)
	UpdateTicket(ctx context.Context, ticketID, tenantID uuid.UUID, req *UpdateTicketRequest) (*SupportTicket, error)
	UpdateTicketStatus(ctx context.Context, ticketID, tenantID uuid.UUID, status, reason string, updatedBy uuid.UUID) error
	UpdateTicketPriority(ctx context.Context, ticketID, tenantID uuid.UUID, priority, reason string, updatedBy uuid.UUID) error
	AssignTicket(ctx context.Context, ticketID, tenantID, assignTo uuid.UUID, note string, assignedBy uuid.UUID) error
	
	// Ticket messaging
	GetTicketMessages(ctx context.Context, ticketID, tenantID uuid.UUID) ([]TicketMessage, error)
	AddTicketMessage(ctx context.Context, req *AddTicketMessageRequest) (*TicketMessage, error)
	
	// Ticket resolution
	ResolveTicket(ctx context.Context, ticketID, tenantID uuid.UUID, resolution, category string, satisfaction *int, metadata map[string]interface{}, resolvedBy uuid.UUID) error
	ReopenTicket(ctx context.Context, ticketID, tenantID uuid.UUID, reason string, reopenedBy uuid.UUID) error
	EscalateTicket(ctx context.Context, ticketID, tenantID uuid.UUID, level, reason string, escalatedBy uuid.UUID) error
	
	// Support information
	GetCategories(ctx context.Context) ([]TicketCategory, error)
	GetTicketTemplates(ctx context.Context, tenantID uuid.UUID) ([]TicketTemplate, error)
	
	// Knowledge base
	GetKnowledgeBaseArticles(ctx context.Context, category, search string, page, perPage int) (*KnowledgeBaseList, error)
	GetKnowledgeBaseArticle(ctx context.Context, articleID uuid.UUID) (*KnowledgeBaseArticle, error)
	SearchKnowledgeBase(ctx context.Context, query string, limit int) ([]KnowledgeBaseSearchResult, error)
	
	// Analytics and metrics
	GetSLAMetrics(ctx context.Context, tenantID uuid.UUID, period *AnalyticsPeriod) (*SLAMetrics, error)
	GetSupportMetrics(ctx context.Context, tenantID uuid.UUID, period *AnalyticsPeriod) (*SupportMetrics, error)
	
	// Notifications
	GetNotifications(ctx context.Context, tenantID, userID uuid.UUID, page, perPage int, unreadOnly bool) (*NotificationList, error)
	MarkNotificationRead(ctx context.Context, notificationID, tenantID, userID uuid.UUID) error
	MarkAllNotificationsRead(ctx context.Context, tenantID, userID uuid.UUID) error
}

// FeatureFlagService defines feature flag management operations
type FeatureFlagService interface {
	// Global feature flags
	GetGlobalFeatureFlags(ctx context.Context) (map[string]bool, error)
	UpdateGlobalFeatureFlag(ctx context.Context, flag string, enabled bool) error
	
	// Tenant-specific feature flags
	GetTenantFeatureFlags(ctx context.Context, tenantID uuid.UUID) (map[string]bool, error)
	UpdateTenantFeatureFlag(ctx context.Context, tenantID uuid.UUID, flag string, enabled bool) error
	
	// Feature flag evaluation
	IsFeatureEnabled(ctx context.Context, tenantID uuid.UUID, feature string) (bool, error)
	GetFeatureConfig(ctx context.Context, tenantID uuid.UUID, feature string) (map[string]interface{}, error)
	
	// Feature flag analytics
	GetFeatureFlagUsage(ctx context.Context) (*FeatureFlagUsage, error)
	GetFeatureFlagImpact(ctx context.Context, flag string, period *AnalyticsPeriod) (*FeatureFlagImpact, error)
}

// PlatformAnalyticsService defines platform-level analytics operations
type PlatformAnalyticsService interface {
	// Platform overview
	GetPlatformOverview(ctx context.Context) (*PlatformOverview, error)
	GetSystemHealth(ctx context.Context) (*SystemHealth, error)
	
	// Tenant analytics
	GetTenantGrowthMetrics(ctx context.Context, period *AnalyticsPeriod) (*TenantGrowthMetrics, error)
	GetTenantEngagementMetrics(ctx context.Context, period *AnalyticsPeriod) (*TenantEngagementMetrics, error)
	GetTenantRetentionMetrics(ctx context.Context, period *AnalyticsPeriod) (*TenantRetentionMetrics, error)
	
	// Performance analytics
	GetAPIPerformanceMetrics(ctx context.Context, period *AnalyticsPeriod) (*APIPerformanceMetrics, error)
	GetDatabasePerformanceMetrics(ctx context.Context, period *AnalyticsPeriod) (*DatabasePerformanceMetrics, error)
	
	// Security analytics
	GetSecurityIncidents(ctx context.Context, period *AnalyticsPeriod) (*SecurityIncidentMetrics, error)
	GetSecurityThreats(ctx context.Context, period *AnalyticsPeriod) (*SecurityThreatMetrics, error)
	
	// Resource utilization
	GetResourceUtilization(ctx context.Context, period *AnalyticsPeriod) (*ResourceUtilization, error)
	GetCostAnalysis(ctx context.Context, period *AnalyticsPeriod) (*CostAnalysis, error)
}

// Data structures for Support Service

type TicketFilter struct {
	TenantID    uuid.UUID  `json:"tenant_id"`
	Status      *string    `json:"status,omitempty"`
	Priority    *string    `json:"priority,omitempty"`
	Category    *string    `json:"category,omitempty"`
	AssignedTo  *uuid.UUID `json:"assigned_to,omitempty"`
	CreatedBy   *uuid.UUID `json:"created_by,omitempty"`
	Search      *string    `json:"search,omitempty"`
	DateFrom    *time.Time `json:"date_from,omitempty"`
	DateTo      *time.Time `json:"date_to,omitempty"`
	Page        int        `json:"page"`
	PerPage     int        `json:"per_page"`
	SortBy      string     `json:"sort_by"`
	SortOrder   string     `json:"sort_order"`
}

type CreateTicketRequest struct {
	TenantID     uuid.UUID              `json:"tenant_id"`
	Subject      string                 `json:"subject"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Priority     string                 `json:"priority"`
	Tags         []string               `json:"tags,omitempty"`
	Attachments  []string               `json:"attachments,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy    uuid.UUID              `json:"created_by"`
}

type UpdateTicketRequest struct {
	Subject     *string                `json:"subject,omitempty"`
	Description *string                `json:"description,omitempty"`
	Category    *string                `json:"category,omitempty"`
	Priority    *string                `json:"priority,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	UpdatedBy   uuid.UUID              `json:"updated_by"`
}

type SupportTicket struct {
	ID           uuid.UUID              `json:"id"`
	TenantID     uuid.UUID              `json:"tenant_id"`
	TicketNumber string                 `json:"ticket_number"`
	Subject      string                 `json:"subject"`
	Description  string                 `json:"description"`
	Category     string                 `json:"category"`
	Priority     string                 `json:"priority"`
	Status       string                 `json:"status"`
	Tags         []string               `json:"tags"`
	Attachments  []string               `json:"attachments"`
	Metadata     map[string]interface{} `json:"metadata"`
	CreatedBy    uuid.UUID              `json:"created_by"`
	AssignedTo   *uuid.UUID             `json:"assigned_to"`
	ResolvedBy   *uuid.UUID             `json:"resolved_by"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	ResolvedAt   *time.Time             `json:"resolved_at"`
	FirstResponseAt *time.Time          `json:"first_response_at"`
	SLATarget    *time.Time             `json:"sla_target"`
	Satisfaction *int                   `json:"satisfaction"`
}

type TicketList struct {
	Tickets      []SupportTicket `json:"tickets"`
	TotalCount   int             `json:"total_count"`
	Page         int             `json:"page"`
	PerPage      int             `json:"per_page"`
	TotalPages   int             `json:"total_pages"`
}

type AddTicketMessageRequest struct {
	TicketID    uuid.UUID              `json:"ticket_id"`
	Message     string                 `json:"message"`
	IsInternal  bool                   `json:"is_internal"`
	Attachments []string               `json:"attachments,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedBy   uuid.UUID              `json:"created_by"`
}

type TicketMessage struct {
	ID          uuid.UUID              `json:"id"`
	TicketID    uuid.UUID              `json:"ticket_id"`
	Message     string                 `json:"message"`
	IsInternal  bool                   `json:"is_internal"`
	Attachments []string               `json:"attachments"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedBy   uuid.UUID              `json:"created_by"`
	CreatedAt   time.Time              `json:"created_at"`
}

type TicketCategory struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SLAHours    int    `json:"sla_hours"`
}

type TicketTemplate struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Subject     string    `json:"subject"`
	Description string    `json:"description"`
	Fields      []TemplateField `json:"fields"`
}

type TemplateField struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Required    bool        `json:"required"`
	Options     []string    `json:"options,omitempty"`
	DefaultValue interface{} `json:"default_value,omitempty"`
}

type KnowledgeBaseArticle struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Category    string    `json:"category"`
	Tags        []string  `json:"tags"`
	Published   bool      `json:"published"`
	ViewCount   int       `json:"view_count"`
	Helpful     int       `json:"helpful"`
	NotHelpful  int       `json:"not_helpful"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type KnowledgeBaseList struct {
	Articles     []KnowledgeBaseArticle `json:"articles"`
	TotalCount   int                    `json:"total_count"`
	Page         int                    `json:"page"`
	PerPage      int                    `json:"per_page"`
	TotalPages   int                    `json:"total_pages"`
}

type KnowledgeBaseSearchResult struct {
	ID       uuid.UUID `json:"id"`
	Title    string    `json:"title"`
	Summary  string    `json:"summary"`
	Category string    `json:"category"`
	Score    float64   `json:"score"`
}

type SLAMetrics struct {
	TenantID              uuid.UUID          `json:"tenant_id"`
	Period                AnalyticsPeriod    `json:"period"`
	AverageResponseTime   time.Duration      `json:"average_response_time"`
	AverageResolutionTime time.Duration      `json:"average_resolution_time"`
	SLACompliance         float64            `json:"sla_compliance"`
	BreachedTickets       int                `json:"breached_tickets"`
	OnTimeTickets         int                `json:"on_time_tickets"`
	SLAByPriority         map[string]SLAData `json:"sla_by_priority"`
	SLAByCategory         map[string]SLAData `json:"sla_by_category"`
}

type SLAData struct {
	AverageResponseTime   time.Duration `json:"average_response_time"`
	AverageResolutionTime time.Duration `json:"average_resolution_time"`
	Compliance            float64       `json:"compliance"`
	BreachedCount         int           `json:"breached_count"`
	OnTimeCount           int           `json:"on_time_count"`
}

type SupportMetrics struct {
	TenantID             uuid.UUID               `json:"tenant_id"`
	Period               AnalyticsPeriod         `json:"period"`
	TotalTickets         int                     `json:"total_tickets"`
	OpenTickets          int                     `json:"open_tickets"`
	ResolvedTickets      int                     `json:"resolved_tickets"`
	ClosedTickets        int                     `json:"closed_tickets"`
	EscalatedTickets     int                     `json:"escalated_tickets"`
	TicketsByStatus      map[string]int          `json:"tickets_by_status"`
	TicketsByPriority    map[string]int          `json:"tickets_by_priority"`
	TicketsByCategory    map[string]int          `json:"tickets_by_category"`
	CustomerSatisfaction float64                 `json:"customer_satisfaction"`
	ResolutionTrends     []ResolutionTrendPoint  `json:"resolution_trends"`
	AgentPerformance     []AgentPerformanceData  `json:"agent_performance"`
}

type ResolutionTrendPoint struct {
	Date      time.Time `json:"date"`
	Resolved  int       `json:"resolved"`
	Created   int       `json:"created"`
	Backlog   int       `json:"backlog"`
}

type AgentPerformanceData struct {
	AgentID          uuid.UUID     `json:"agent_id"`
	AgentName        string        `json:"agent_name"`
	TicketsResolved  int           `json:"tickets_resolved"`
	AverageRating    float64       `json:"average_rating"`
	ResponseTime     time.Duration `json:"response_time"`
	ResolutionTime   time.Duration `json:"resolution_time"`
}

type NotificationList struct {
	Notifications []Notification `json:"notifications"`
	TotalCount    int            `json:"total_count"`
	UnreadCount   int            `json:"unread_count"`
	Page          int            `json:"page"`
	PerPage       int            `json:"per_page"`
	TotalPages    int            `json:"total_pages"`
}

type Notification struct {
	ID        uuid.UUID              `json:"id"`
	TenantID  uuid.UUID              `json:"tenant_id"`
	UserID    uuid.UUID              `json:"user_id"`
	Type      string                 `json:"type"`
	Title     string                 `json:"title"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data"`
	Read      bool                   `json:"read"`
	ReadAt    *time.Time             `json:"read_at"`
	CreatedAt time.Time              `json:"created_at"`
}

// Data structures for Feature Flag Service

type FeatureFlagUsage struct {
	TotalFlags         int                        `json:"total_flags"`
	EnabledFlags       int                        `json:"enabled_flags"`
	TenantOverrides    int                        `json:"tenant_overrides"`
	FlagsByCategory    map[string]int             `json:"flags_by_category"`
	PopularFlags       []FeatureFlagUsageStats    `json:"popular_flags"`
	TenantDistribution map[string]int             `json:"tenant_distribution"`
	GeneratedAt        time.Time                  `json:"generated_at"`
}

type FeatureFlagUsageStats struct {
	Flag         string  `json:"flag"`
	UsageCount   int     `json:"usage_count"`
	EnabledRate  float64 `json:"enabled_rate"`
	TenantCount  int     `json:"tenant_count"`
}

type FeatureFlagImpact struct {
	Flag            string                 `json:"flag"`
	Period          AnalyticsPeriod        `json:"period"`
	EnabledTenants  int                    `json:"enabled_tenants"`
	DisabledTenants int                    `json:"disabled_tenants"`
	ImpactMetrics   map[string]interface{} `json:"impact_metrics"`
	PerformanceData []PerformanceDataPoint `json:"performance_data"`
}

type PerformanceDataPoint struct {
	Date     time.Time              `json:"date"`
	Enabled  map[string]interface{} `json:"enabled"`
	Disabled map[string]interface{} `json:"disabled"`
}

// Data structures for Platform Analytics Service

type PlatformOverview struct {
	TotalTenants       int                    `json:"total_tenants"`
	ActiveTenants      int                    `json:"active_tenants"`
	NewTenantsToday    int                    `json:"new_tenants_today"`
	NewTenantsThisWeek int                    `json:"new_tenants_this_week"`
	TotalRevenue       float64                `json:"total_revenue"`
	MonthlyRevenue     float64                `json:"monthly_revenue"`
	RevenueGrowth      float64                `json:"revenue_growth"`
	SystemStatus       string                 `json:"system_status"`
	AlertsCount        int                    `json:"alerts_count"`
	UptimePercentage   float64                `json:"uptime_percentage"`
	TopMetrics         map[string]interface{} `json:"top_metrics"`
	RecentActivity     []ActivityItem         `json:"recent_activity"`
	GeneratedAt        time.Time              `json:"generated_at"`
}

type ActivityItem struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	TenantID    *uuid.UUID             `json:"tenant_id,omitempty"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
}

type SystemHealth struct {
	Status           string                    `json:"status"`
	Services         map[string]ServiceHealth  `json:"services"`
	DatabaseHealth   DatabaseHealth            `json:"database_health"`
	CacheHealth      CacheHealth               `json:"cache_health"`
	APIHealth        APIHealth                 `json:"api_health"`
	ThirdPartyHealth map[string]ServiceHealth  `json:"third_party_health"`
	Alerts           []SystemAlert             `json:"alerts"`
	LastChecked      time.Time                 `json:"last_checked"`
}

type ServiceHealth struct {
	Status      string        `json:"status"`
	ResponseTime time.Duration `json:"response_time"`
	Uptime      float64       `json:"uptime"`
	LastError   *string       `json:"last_error,omitempty"`
	LastChecked time.Time     `json:"last_checked"`
}

type DatabaseHealth struct {
	Status           string        `json:"status"`
	ConnectionCount  int           `json:"connection_count"`
	MaxConnections   int           `json:"max_connections"`
	QueryPerformance time.Duration `json:"query_performance"`
	SlowQueries      int           `json:"slow_queries"`
	Replication      ReplicationStatus `json:"replication"`
}

type ReplicationStatus struct {
	Status    string        `json:"status"`
	LagTime   time.Duration `json:"lag_time"`
	Replicas  int           `json:"replicas"`
	LastSync  time.Time     `json:"last_sync"`
}

type CacheHealth struct {
	Status    string  `json:"status"`
	HitRate   float64 `json:"hit_rate"`
	Memory    MemoryUsage `json:"memory"`
	KeyCount  int     `json:"key_count"`
}

type MemoryUsage struct {
	Used      int64   `json:"used"`
	Available int64   `json:"available"`
	Usage     float64 `json:"usage"`
}

type APIHealth struct {
	Status           string        `json:"status"`
	RequestsPerSecond float64      `json:"requests_per_second"`
	AverageLatency   time.Duration `json:"average_latency"`
	ErrorRate        float64       `json:"error_rate"`
	RateLimit        RateLimitStatus `json:"rate_limit"`
}

type RateLimitStatus struct {
	CurrentRequests int     `json:"current_requests"`
	Limit          int     `json:"limit"`
	WindowSize     string  `json:"window_size"`
	Remaining      int     `json:"remaining"`
	ResetTime      time.Time `json:"reset_time"`
}

type SystemAlert struct {
	ID          uuid.UUID              `json:"id"`
	Level       string                 `json:"level"`
	Service     string                 `json:"service"`
	Message     string                 `json:"message"`
	Details     map[string]interface{} `json:"details"`
	CreatedAt   time.Time              `json:"created_at"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
}

type TenantGrowthMetrics struct {
	Period              AnalyticsPeriod    `json:"period"`
	NewTenants          int                `json:"new_tenants"`
	ChurnedTenants      int                `json:"churned_tenants"`
	NetGrowth           int                `json:"net_growth"`
	GrowthRate          float64            `json:"growth_rate"`
	TenantsByPlan       map[string]int     `json:"tenants_by_plan"`
	GrowthTrend         []GrowthDataPoint  `json:"growth_trend"`
	ConversionFunnel    ConversionFunnel   `json:"conversion_funnel"`
	GeographicDistribution map[string]int `json:"geographic_distribution"`
}

type GrowthDataPoint struct {
	Date       time.Time `json:"date"`
	NewTenants int       `json:"new_tenants"`
	Churned    int       `json:"churned"`
	Total      int       `json:"total"`
}

type ConversionFunnel struct {
	Signups      int     `json:"signups"`
	Trials       int     `json:"trials"`
	Conversions  int     `json:"conversions"`
	SignupRate   float64 `json:"signup_rate"`
	TrialRate    float64 `json:"trial_rate"`
	ConversionRate float64 `json:"conversion_rate"`
}

type TenantEngagementMetrics struct {
	Period                AnalyticsPeriod         `json:"period"`
	DailyActiveUsers      int                     `json:"daily_active_users"`
	WeeklyActiveUsers     int                     `json:"weekly_active_users"`
	MonthlyActiveUsers    int                     `json:"monthly_active_users"`
	SessionDuration       time.Duration           `json:"session_duration"`
	PageViews             int                     `json:"page_views"`
	FeatureUsage          map[string]int          `json:"feature_usage"`
	EngagementTrend       []EngagementDataPoint   `json:"engagement_trend"`
	TenantSegmentation    TenantSegmentation      `json:"tenant_segmentation"`
}

type EngagementDataPoint struct {
	Date       time.Time `json:"date"`
	ActiveUsers int      `json:"active_users"`
	Sessions   int       `json:"sessions"`
	PageViews  int       `json:"page_views"`
}

type TenantSegmentation struct {
	HighEngagement    int `json:"high_engagement"`
	MediumEngagement  int `json:"medium_engagement"`
	LowEngagement     int `json:"low_engagement"`
	InactiveTenants   int `json:"inactive_tenants"`
}

type TenantRetentionMetrics struct {
	Period           AnalyticsPeriod      `json:"period"`
	RetentionRate    float64              `json:"retention_rate"`
	ChurnRate        float64              `json:"churn_rate"`
	CohortAnalysis   []CohortData         `json:"cohort_analysis"`
	RetentionByPlan  map[string]float64   `json:"retention_by_plan"`
	ChurnReasons     map[string]int       `json:"churn_reasons"`
	RetentionTrend   []RetentionDataPoint `json:"retention_trend"`
}

type CohortData struct {
	CohortMonth    string             `json:"cohort_month"`
	TenantCount    int                `json:"tenant_count"`
	RetentionRates map[string]float64 `json:"retention_rates"`
}

type RetentionDataPoint struct {
	Month         string  `json:"month"`
	RetentionRate float64 `json:"retention_rate"`
	ChurnRate     float64 `json:"churn_rate"`
	TenantCount   int     `json:"tenant_count"`
}

type APIPerformanceMetrics struct {
	Period               AnalyticsPeriod           `json:"period"`
	TotalRequests        int64                     `json:"total_requests"`
	AverageLatency       time.Duration             `json:"average_latency"`
	P95Latency           time.Duration             `json:"p95_latency"`
	P99Latency           time.Duration             `json:"p99_latency"`
	ErrorRate            float64                   `json:"error_rate"`
	RequestsPerSecond    float64                   `json:"requests_per_second"`
	EndpointPerformance  map[string]EndpointStats  `json:"endpoint_performance"`
	ErrorsByType         map[string]int            `json:"errors_by_type"`
	PerformanceTrend     []PerformanceDataPoint    `json:"performance_trend"`
}

type EndpointStats struct {
	RequestCount     int64         `json:"request_count"`
	AverageLatency   time.Duration `json:"average_latency"`
	ErrorRate        float64       `json:"error_rate"`
	ThroughputPerSec float64       `json:"throughput_per_sec"`
}

type DatabasePerformanceMetrics struct {
	Period              AnalyticsPeriod     `json:"period"`
	QueryCount          int64               `json:"query_count"`
	AverageQueryTime    time.Duration       `json:"average_query_time"`
	SlowQueries         int                 `json:"slow_queries"`
	DeadlockCount       int                 `json:"deadlock_count"`
	ConnectionPoolUsage float64             `json:"connection_pool_usage"`
	CacheHitRate        float64             `json:"cache_hit_rate"`
	TableSizes          map[string]int64    `json:"table_sizes"`
	IndexUsage          map[string]float64  `json:"index_usage"`
}

type SecurityIncidentMetrics struct {
	Period                  AnalyticsPeriod    `json:"period"`
	TotalIncidents          int                `json:"total_incidents"`
	CriticalIncidents       int                `json:"critical_incidents"`
	ResolvedIncidents       int                `json:"resolved_incidents"`
	AverageResolutionTime   time.Duration      `json:"average_resolution_time"`
	IncidentsByType         map[string]int     `json:"incidents_by_type"`
	IncidentsBySeverity     map[string]int     `json:"incidents_by_severity"`
	AttackVectors           map[string]int     `json:"attack_vectors"`
	AffectedTenants         int                `json:"affected_tenants"`
}

type SecurityThreatMetrics struct {
	Period              AnalyticsPeriod    `json:"period"`
	ThreatCount         int                `json:"threat_count"`
	BlockedRequests     int                `json:"blocked_requests"`
	SuspiciousActivities int               `json:"suspicious_activities"`
	ThreatsByType       map[string]int     `json:"threats_by_type"`
	ThreatsByCountry    map[string]int     `json:"threats_by_country"`
	ThreatTrend         []ThreatDataPoint  `json:"threat_trend"`
}

type ThreatDataPoint struct {
	Date    time.Time `json:"date"`
	Threats int       `json:"threats"`
	Blocked int       `json:"blocked"`
}

type ResourceUtilization struct {
	Period        AnalyticsPeriod       `json:"period"`
	CPUUsage      ResourceUsageStats    `json:"cpu_usage"`
	MemoryUsage   ResourceUsageStats    `json:"memory_usage"`
	DiskUsage     ResourceUsageStats    `json:"disk_usage"`
	NetworkUsage  NetworkUsageStats     `json:"network_usage"`
	ServiceUsage  map[string]ResourceUsageStats `json:"service_usage"`
}

type ResourceUsageStats struct {
	Average  float64 `json:"average"`
	Peak     float64 `json:"peak"`
	Minimum  float64 `json:"minimum"`
	Current  float64 `json:"current"`
}

type NetworkUsageStats struct {
	InboundTraffic  int64 `json:"inbound_traffic"`
	OutboundTraffic int64 `json:"outbound_traffic"`
	RequestCount    int64 `json:"request_count"`
	Bandwidth       ResourceUsageStats `json:"bandwidth"`
}

type CostAnalysis struct {
	Period              AnalyticsPeriod    `json:"period"`
	TotalCost           float64            `json:"total_cost"`
	CostPerTenant       float64            `json:"cost_per_tenant"`
	CostBreakdown       map[string]float64 `json:"cost_breakdown"`
	CostTrend           []CostDataPoint    `json:"cost_trend"`
	ResourceCosts       map[string]float64 `json:"resource_costs"`
	ServiceCosts        map[string]float64 `json:"service_costs"`
	ProjectedCosts      ProjectedCosts     `json:"projected_costs"`
}

type CostDataPoint struct {
	Date time.Time `json:"date"`
	Cost float64   `json:"cost"`
}

type ProjectedCosts struct {
	NextMonth   float64 `json:"next_month"`
	NextQuarter float64 `json:"next_quarter"`
	NextYear    float64 `json:"next_year"`
}

// RevenueDataPoint is defined in dto.go

// AnalyticsPeriod for all analytics (already defined in super_admin_service.go)
type AnalyticsPeriod struct {
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Interval  string    `json:"interval"` // day, week, month, quarter, year
}