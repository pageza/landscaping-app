package types

import (
	"time"

	"github.com/google/uuid"
)

// Shared service-related types used across repository and services layers

// InvoiceService represents a service on an invoice
type InvoiceService struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	InvoiceID   uuid.UUID              `json:"invoice_id" db:"invoice_id"`
	ServiceID   *uuid.UUID             `json:"service_id,omitempty" db:"service_id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	Quantity    float64                `json:"quantity" db:"quantity"`
	UnitPrice   float64                `json:"unit_price" db:"unit_price"`
	Total       float64                `json:"total" db:"total"`
	TaxRate     float64                `json:"tax_rate" db:"tax_rate"`
	TaxAmount   float64                `json:"tax_amount" db:"tax_amount"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" db:"updated_at"`
}

// RecurringJobSeries represents a series of recurring jobs
type RecurringJobSeries struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	TenantID           uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	BaseJobID          uuid.UUID              `json:"base_job_id" db:"base_job_id"`
	CustomerID         uuid.UUID              `json:"customer_id" db:"customer_id"`
	PropertyID         uuid.UUID              `json:"property_id" db:"property_id"`
	Title              string                 `json:"title" db:"title"`
	Description        string                 `json:"description" db:"description"`
	RecurrencePattern  string                 `json:"recurrence_pattern" db:"recurrence_pattern"`
	RecurrenceInterval int                    `json:"recurrence_interval" db:"recurrence_interval"`
	RecurrenceUnit     string                 `json:"recurrence_unit" db:"recurrence_unit"`
	StartDate          time.Time              `json:"start_date" db:"start_date"`
	EndDate            *time.Time             `json:"end_date,omitempty" db:"end_date"`
	NextRun            time.Time              `json:"next_run" db:"next_run"`
	LastRun            *time.Time             `json:"last_run,omitempty" db:"last_run"`
	IsActive           bool                   `json:"is_active" db:"is_active"`
	MaxOccurrences     *int                   `json:"max_occurrences,omitempty" db:"max_occurrences"`
	CompletedRuns      int                    `json:"completed_runs" db:"completed_runs"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at" db:"updated_at"`
}

// CustomerSummary represents a summary of customer data
type CustomerSummary struct {
	CustomerID         uuid.UUID                 `json:"customer_id"`
	TotalJobs          int64                     `json:"total_jobs"`
	CompletedJobs      int64                     `json:"completed_jobs"`
	PendingJobs        int64                     `json:"pending_jobs"`
	TotalInvoices      int64                     `json:"total_invoices"`
	PaidInvoices       int64                     `json:"paid_invoices"`
	UnpaidInvoices     int64                     `json:"unpaid_invoices"`
	TotalRevenue       float64                   `json:"total_revenue"`
	OutstandingBalance float64                   `json:"outstanding_balance"`
	AverageJobValue    float64                   `json:"average_job_value"`
	PaymentHistory     []PaymentHistoryEntry     `json:"payment_history"`
	RecentJobs         []JobSummary              `json:"recent_jobs"`
	Metadata           map[string]interface{}    `json:"metadata"`
}

// PaymentHistoryEntry represents an entry in payment history
type PaymentHistoryEntry struct {
	PaymentID     uuid.UUID `json:"payment_id"`
	InvoiceID     uuid.UUID `json:"invoice_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	PaymentDate   time.Time `json:"payment_date"`
	PaymentMethod string    `json:"payment_method"`
	Status        string    `json:"status"`
}

// JobSummary represents a summary of job data
type JobSummary struct {
	ID           uuid.UUID `json:"id"`
	Title        string    `json:"title"`
	Status       string    `json:"status"`
	ScheduledDate *time.Time `json:"scheduled_date,omitempty"`
	CompletedDate *time.Time `json:"completed_date,omitempty"`
	TotalAmount   float64   `json:"total_amount"`
}

// PaymentSummary represents a summary of payment data
type PaymentSummary struct {
	TotalPayments      int64   `json:"total_payments"`
	TotalAmount        float64 `json:"total_amount"`
	SuccessfulPayments int64   `json:"successful_payments"`
	FailedPayments     int64   `json:"failed_payments"`
	RefundedPayments   int64   `json:"refunded_payments"`
	TotalRefunds       float64 `json:"total_refunds"`
	AveragePayment     float64 `json:"average_payment"`
	Currency           string  `json:"currency"`
}

// UsageRecord represents a usage record for subscriptions
type UsageRecord struct {
	ID             uuid.UUID              `json:"id" db:"id"`
	TenantID       uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	SubscriptionID uuid.UUID              `json:"subscription_id" db:"subscription_id"`
	MetricName     string                 `json:"metric_name" db:"metric_name"`
	Quantity       float64                `json:"quantity" db:"quantity"`
	Unit           string                 `json:"unit" db:"unit"`
	Timestamp      time.Time              `json:"timestamp" db:"timestamp"`
	BillingPeriod  string                 `json:"billing_period" db:"billing_period"`
	Cost           float64                `json:"cost" db:"cost"`
	Metadata       map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt      time.Time              `json:"created_at" db:"created_at"`
}

// UsageSummary represents a summary of usage data
type UsageSummary struct {
	SubscriptionID   uuid.UUID       `json:"subscription_id"`
	Period           string          `json:"period"`
	StartDate        time.Time       `json:"start_date"`
	EndDate          time.Time       `json:"end_date"`
	TotalUsage       float64         `json:"total_usage"`
	TotalCost        float64         `json:"total_cost"`
	MetricSummaries  []MetricSummary `json:"metric_summaries"`
	UsageTrends      []UsageTrend    `json:"usage_trends"`
}

// MetricSummary represents a summary for a specific metric
type MetricSummary struct {
	MetricName string  `json:"metric_name"`
	Unit       string  `json:"unit"`
	Total      float64 `json:"total"`
	Average    float64 `json:"average"`
	Peak       float64 `json:"peak"`
	Cost       float64 `json:"cost"`
}

// UsageTrend represents usage trend data
type UsageTrend struct {
	Date       time.Time `json:"date" db:"date"`
	MetricName string    `json:"metric_name" db:"metric_name"`
	Quantity   float64   `json:"quantity" db:"quantity"`
	Cost       float64   `json:"cost" db:"cost"`
}