package types

import (
	"time"

	"github.com/google/uuid"
)

// Shared filter types used across repository and services layers

// CustomerFilter represents filters for querying customers
type CustomerFilter struct {
	Status         *string    `json:"status,omitempty"`
	Type           *string    `json:"type,omitempty"`
	Tags           []string   `json:"tags,omitempty"`
	CreatedAfter   *time.Time `json:"created_after,omitempty"`
	CreatedBefore  *time.Time `json:"created_before,omitempty"`
	UpdatedAfter   *time.Time `json:"updated_after,omitempty"`
	UpdatedBefore  *time.Time `json:"updated_before,omitempty"`
	Search         *string    `json:"search,omitempty"`
	Page           int        `json:"page"`
	PerPage        int        `json:"per_page"`
	SortBy         string     `json:"sort_by"`
	SortOrder      string     `json:"sort_order"`
}

// PropertyFilter represents filters for querying properties
type PropertyFilter struct {
	CustomerID     *uuid.UUID `json:"customer_id,omitempty"`
	PropertyType   *string    `json:"property_type,omitempty"`
	Size           *float64   `json:"size,omitempty"`
	MinSize        *float64   `json:"min_size,omitempty"`
	MaxSize        *float64   `json:"max_size,omitempty"`
	Features       []string   `json:"features,omitempty"`
	City           *string    `json:"city,omitempty"`
	State          *string    `json:"state,omitempty"`
	PostalCode     *string    `json:"postal_code,omitempty"`
	Search         *string    `json:"search,omitempty"`
	Page           int        `json:"page"`
	PerPage        int        `json:"per_page"`
	SortBy         string     `json:"sort_by"`
	SortOrder      string     `json:"sort_order"`
}

// JobFilter represents filters for querying jobs
type JobFilter struct {
	CustomerID       *uuid.UUID `json:"customer_id,omitempty"`
	PropertyID       *uuid.UUID `json:"property_id,omitempty"`
	AssignedUserID   *uuid.UUID `json:"assigned_user_id,omitempty"`
	Status           *string    `json:"status,omitempty"`
	Priority         *string    `json:"priority,omitempty"`
	ServiceType      *string    `json:"service_type,omitempty"`
	ScheduledAfter   *time.Time `json:"scheduled_after,omitempty"`
	ScheduledBefore  *time.Time `json:"scheduled_before,omitempty"`
	CreatedAfter     *time.Time `json:"created_after,omitempty"`
	CreatedBefore    *time.Time `json:"created_before,omitempty"`
	CompletedAfter   *time.Time `json:"completed_after,omitempty"`
	CompletedBefore  *time.Time `json:"completed_before,omitempty"`
	Tags             []string   `json:"tags,omitempty"`
	IsRecurring      *bool      `json:"is_recurring,omitempty"`
	Search           *string    `json:"search,omitempty"`
	Page             int        `json:"page"`
	PerPage          int        `json:"per_page"`
	SortBy           string     `json:"sort_by"`
	SortOrder        string     `json:"sort_order"`
}

// InvoiceFilter represents filters for querying invoices
type InvoiceFilter struct {
	CustomerID      *uuid.UUID `json:"customer_id,omitempty"`
	JobID           *uuid.UUID `json:"job_id,omitempty"`
	Status          *string    `json:"status,omitempty"`
	IssuedAfter     *time.Time `json:"issued_after,omitempty"`
	IssuedBefore    *time.Time `json:"issued_before,omitempty"`
	DueAfter        *time.Time `json:"due_after,omitempty"`
	DueBefore       *time.Time `json:"due_before,omitempty"`
	PaidAfter       *time.Time `json:"paid_after,omitempty"`
	PaidBefore      *time.Time `json:"paid_before,omitempty"`
	MinAmount       *float64   `json:"min_amount,omitempty"`
	MaxAmount       *float64   `json:"max_amount,omitempty"`
	InvoiceNumber   *string    `json:"invoice_number,omitempty"`
	Search          *string    `json:"search,omitempty"`
	Page            int        `json:"page"`
	PerPage         int        `json:"per_page"`
	SortBy          string     `json:"sort_by"`
	SortOrder       string     `json:"sort_order"`
}

// PaymentFilter represents filters for querying payments
type PaymentFilter struct {
	CustomerID       *uuid.UUID `json:"customer_id,omitempty"`
	InvoiceID        *uuid.UUID `json:"invoice_id,omitempty"`
	PaymentMethodID  *string    `json:"payment_method_id,omitempty"`
	Status           *string    `json:"status,omitempty"`
	ProcessedAfter   *time.Time `json:"processed_after,omitempty"`
	ProcessedBefore  *time.Time `json:"processed_before,omitempty"`
	MinAmount        *float64   `json:"min_amount,omitempty"`
	MaxAmount        *float64   `json:"max_amount,omitempty"`
	Currency         *string    `json:"currency,omitempty"`
	PaymentProvider  *string    `json:"payment_provider,omitempty"`
	Search           *string    `json:"search,omitempty"`
	Page             int        `json:"page"`
	PerPage          int        `json:"per_page"`
	SortBy           string     `json:"sort_by"`
	SortOrder        string     `json:"sort_order"`
}

// QuoteFilter represents filters for querying quotes
type QuoteFilter struct {
	CustomerID       *uuid.UUID `json:"customer_id,omitempty"`
	PropertyID       *uuid.UUID `json:"property_id,omitempty"`
	Status           *string    `json:"status,omitempty"`
	ValidAfter       *time.Time `json:"valid_after,omitempty"`
	ValidBefore      *time.Time `json:"valid_before,omitempty"`
	CreatedAfter     *time.Time `json:"created_after,omitempty"`
	CreatedBefore    *time.Time `json:"created_before,omitempty"`
	MinAmount        *float64   `json:"min_amount,omitempty"`
	MaxAmount        *float64   `json:"max_amount,omitempty"`
	ServiceType      *string    `json:"service_type,omitempty"`
	Tags             []string   `json:"tags,omitempty"`
	Search           *string    `json:"search,omitempty"`
	Page             int        `json:"page"`
	PerPage          int        `json:"per_page"`
	SortBy           string     `json:"sort_by"`
	SortOrder        string     `json:"sort_order"`
}

// SubscriptionFilter represents filters for querying subscriptions
type SubscriptionFilter struct {
	CustomerID       *uuid.UUID `json:"customer_id,omitempty"`
	Status           *string    `json:"status,omitempty"`
	PlanID           *string    `json:"plan_id,omitempty"`
	BillingInterval  *string    `json:"billing_interval,omitempty"`
	CreatedAfter     *time.Time `json:"created_after,omitempty"`
	CreatedBefore    *time.Time `json:"created_before,omitempty"`
	TrialEnd         *time.Time `json:"trial_end,omitempty"`
	MinAmount        *float64   `json:"min_amount,omitempty"`
	MaxAmount        *float64   `json:"max_amount,omitempty"`
	Search           *string    `json:"search,omitempty"`
	Page             int        `json:"page"`
	PerPage          int        `json:"per_page"`
	SortBy           string     `json:"sort_by"`
	SortOrder        string     `json:"sort_order"`
}