package types

import (
	"time"

	"github.com/google/uuid"
)

// Equipment filters and types shared between repository and services layers

// EquipmentFilter represents filters for querying equipment
type EquipmentFilter struct {
	Type             *string  `json:"type,omitempty"`
	Status           *string  `json:"status,omitempty"`
	Location         *string  `json:"location,omitempty"`
	AssignedUserID   *uuid.UUID `json:"assigned_user_id,omitempty"`
	MinPurchaseDate  *time.Time `json:"min_purchase_date,omitempty"`
	MaxPurchaseDate  *time.Time `json:"max_purchase_date,omitempty"`
	MinWarrantyEnd   *time.Time `json:"min_warranty_end,omitempty"`
	MaxWarrantyEnd   *time.Time `json:"max_warranty_end,omitempty"`
	SerialNumber     *string    `json:"serial_number,omitempty"`
	Manufacturer     *string    `json:"manufacturer,omitempty"`
	Model            *string    `json:"model,omitempty"`
	Tags             []string   `json:"tags,omitempty"`
	Search           *string    `json:"search,omitempty"`
	Page             int        `json:"page"`
	PerPage          int        `json:"per_page"`
	SortBy           string     `json:"sort_by"`
	SortOrder        string     `json:"sort_order"`
}

// MaintenanceRecord represents a maintenance record
type MaintenanceRecord struct {
	ID               uuid.UUID              `json:"id" db:"id"`
	TenantID         uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	EquipmentID      uuid.UUID              `json:"equipment_id" db:"equipment_id"`
	PerformedBy      uuid.UUID              `json:"performed_by" db:"performed_by"`
	MaintenanceType  string                 `json:"maintenance_type" db:"maintenance_type"`
	Description      string                 `json:"description" db:"description"`
	MaintenanceDate  time.Time              `json:"maintenance_date" db:"maintenance_date"`
	Labor            float64                `json:"labor" db:"labor"`
	Parts            float64                `json:"parts" db:"parts"`
	TotalCost        float64                `json:"total_cost" db:"total_cost"`
	Notes            string                 `json:"notes" db:"notes"`
	PartsUsed        []string               `json:"parts_used" db:"parts_used"`
	Attachments      []string               `json:"attachments" db:"attachments"`
	NextDueDate      *time.Time             `json:"next_due_date,omitempty" db:"next_due_date"`
	Metadata         map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt        time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at" db:"updated_at"`
}

// MaintenanceSchedule represents a maintenance schedule
type MaintenanceSchedule struct {
	ID                uuid.UUID              `json:"id" db:"id"`
	TenantID          uuid.UUID              `json:"tenant_id" db:"tenant_id"`
	EquipmentID       uuid.UUID              `json:"equipment_id" db:"equipment_id"`
	MaintenanceType   string                 `json:"maintenance_type" db:"maintenance_type"`
	Title             string                 `json:"title" db:"title"`
	Description       string                 `json:"description" db:"description"`
	Frequency         string                 `json:"frequency" db:"frequency"`
	FrequencyValue    int                    `json:"frequency_value" db:"frequency_value"`
	FrequencyUnit     string                 `json:"frequency_unit" db:"frequency_unit"`
	LastPerformed     *time.Time             `json:"last_performed,omitempty" db:"last_performed"`
	NextDue           time.Time              `json:"next_due" db:"next_due"`
	AssignedTo        *uuid.UUID             `json:"assigned_to,omitempty" db:"assigned_to"`
	EstimatedDuration int                    `json:"estimated_duration" db:"estimated_duration"`
	EstimatedCost     float64                `json:"estimated_cost" db:"estimated_cost"`
	Priority          string                 `json:"priority" db:"priority"`
	Instructions      string                 `json:"instructions" db:"instructions"`
	RequiredParts     []string               `json:"required_parts" db:"required_parts"`
	RequiredTools     []string               `json:"required_tools" db:"required_tools"`
	Active            bool                   `json:"active" db:"active"`
	Metadata          map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt         time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time              `json:"updated_at" db:"updated_at"`
}