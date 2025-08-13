package domain

import (
	"time"

	"github.com/google/uuid"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID        uuid.UUID              `json:"id" db:"id"`
	Name      string                 `json:"name" db:"name"`
	Subdomain string                 `json:"subdomain" db:"subdomain"`
	Plan      string                 `json:"plan" db:"plan"`
	Status    string                 `json:"status" db:"status"`
	Settings  map[string]interface{} `json:"settings" db:"settings"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt time.Time              `json:"updated_at" db:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID                       uuid.UUID  `json:"id" db:"id"`
	TenantID                 uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Email                    string     `json:"email" db:"email"`
	PasswordHash             string     `json:"-" db:"password_hash"`
	FirstName                string     `json:"first_name" db:"first_name"`
	LastName                 string     `json:"last_name" db:"last_name"`
	Role                     string     `json:"role" db:"role"`
	Status                   string     `json:"status" db:"status"`
	EmailVerified            bool       `json:"email_verified" db:"email_verified"`
	EmailVerificationToken   *string    `json:"-" db:"email_verification_token"`
	PasswordResetToken       *string    `json:"-" db:"password_reset_token"`
	PasswordResetExpires     *time.Time `json:"-" db:"password_reset_expires"`
	LastLogin                *time.Time `json:"last_login" db:"last_login"`
	CreatedAt                time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at" db:"updated_at"`
}

// Customer represents a customer
type Customer struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TenantID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	FirstName    string    `json:"first_name" db:"first_name"`
	LastName     string    `json:"last_name" db:"last_name"`
	Email        *string   `json:"email" db:"email"`
	Phone        *string   `json:"phone" db:"phone"`
	AddressLine1 *string   `json:"address_line1" db:"address_line1"`
	AddressLine2 *string   `json:"address_line2" db:"address_line2"`
	City         *string   `json:"city" db:"city"`
	State        *string   `json:"state" db:"state"`
	ZipCode      *string   `json:"zip_code" db:"zip_code"`
	Country      string    `json:"country" db:"country"`
	Notes        *string   `json:"notes" db:"notes"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Property represents a property
type Property struct {
	ID           uuid.UUID `json:"id" db:"id"`
	TenantID     uuid.UUID `json:"tenant_id" db:"tenant_id"`
	CustomerID   uuid.UUID `json:"customer_id" db:"customer_id"`
	Name         string    `json:"name" db:"name"`
	AddressLine1 string    `json:"address_line1" db:"address_line1"`
	AddressLine2 *string   `json:"address_line2" db:"address_line2"`
	City         string    `json:"city" db:"city"`
	State        string    `json:"state" db:"state"`
	ZipCode      string    `json:"zip_code" db:"zip_code"`
	Country      string    `json:"country" db:"country"`
	PropertyType string    `json:"property_type" db:"property_type"`
	LotSize      *float64  `json:"lot_size" db:"lot_size"`
	Notes        *string   `json:"notes" db:"notes"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// Service represents a service offered
type Service struct {
	ID              uuid.UUID `json:"id" db:"id"`
	TenantID        uuid.UUID `json:"tenant_id" db:"tenant_id"`
	Name            string    `json:"name" db:"name"`
	Description     *string   `json:"description" db:"description"`
	Category        string    `json:"category" db:"category"`
	BasePrice       *float64  `json:"base_price" db:"base_price"`
	Unit            *string   `json:"unit" db:"unit"`
	DurationMinutes *int      `json:"duration_minutes" db:"duration_minutes"`
	Status          string    `json:"status" db:"status"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// Job represents a job/work order
type Job struct {
	ID                  uuid.UUID  `json:"id" db:"id"`
	TenantID            uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	CustomerID          uuid.UUID  `json:"customer_id" db:"customer_id"`
	PropertyID          uuid.UUID  `json:"property_id" db:"property_id"`
	AssignedUserID      *uuid.UUID `json:"assigned_user_id" db:"assigned_user_id"`
	Title               string     `json:"title" db:"title"`
	Description         *string    `json:"description" db:"description"`
	Status              string     `json:"status" db:"status"`
	Priority            string     `json:"priority" db:"priority"`
	ScheduledDate       *time.Time `json:"scheduled_date" db:"scheduled_date"`
	ScheduledTime       *string    `json:"scheduled_time" db:"scheduled_time"`
	EstimatedDuration   *int       `json:"estimated_duration" db:"estimated_duration"`
	ActualStartTime     *time.Time `json:"actual_start_time" db:"actual_start_time"`
	ActualEndTime       *time.Time `json:"actual_end_time" db:"actual_end_time"`
	TotalAmount         *float64   `json:"total_amount" db:"total_amount"`
	Notes               *string    `json:"notes" db:"notes"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
}

// JobService represents the services associated with a job
type JobService struct {
	ID         uuid.UUID `json:"id" db:"id"`
	JobID      uuid.UUID `json:"job_id" db:"job_id"`
	ServiceID  uuid.UUID `json:"service_id" db:"service_id"`
	Quantity   float64   `json:"quantity" db:"quantity"`
	UnitPrice  float64   `json:"unit_price" db:"unit_price"`
	TotalPrice float64   `json:"total_price" db:"total_price"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}

// Invoice represents an invoice
type Invoice struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	CustomerID    uuid.UUID  `json:"customer_id" db:"customer_id"`
	JobID         *uuid.UUID `json:"job_id" db:"job_id"`
	InvoiceNumber string     `json:"invoice_number" db:"invoice_number"`
	Status        string     `json:"status" db:"status"`
	Subtotal      float64    `json:"subtotal" db:"subtotal"`
	TaxRate       float64    `json:"tax_rate" db:"tax_rate"`
	TaxAmount     float64    `json:"tax_amount" db:"tax_amount"`
	TotalAmount   float64    `json:"total_amount" db:"total_amount"`
	IssuedDate    *time.Time `json:"issued_date" db:"issued_date"`
	DueDate       *time.Time `json:"due_date" db:"due_date"`
	PaidDate      *time.Time `json:"paid_date" db:"paid_date"`
	Notes         *string    `json:"notes" db:"notes"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// Payment represents a payment
type Payment struct {
	ID                    uuid.UUID  `json:"id" db:"id"`
	TenantID              uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	InvoiceID             uuid.UUID  `json:"invoice_id" db:"invoice_id"`
	Amount                float64    `json:"amount" db:"amount"`
	PaymentMethod         string     `json:"payment_method" db:"payment_method"`
	PaymentGateway        *string    `json:"payment_gateway" db:"payment_gateway"`
	GatewayTransactionID  *string    `json:"gateway_transaction_id" db:"gateway_transaction_id"`
	Status                string     `json:"status" db:"status"`
	ProcessedAt           *time.Time `json:"processed_at" db:"processed_at"`
	Notes                 *string    `json:"notes" db:"notes"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// Equipment represents equipment/assets
type Equipment struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	TenantID             uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	Name                 string     `json:"name" db:"name"`
	Type                 string     `json:"type" db:"type"`
	Model                *string    `json:"model" db:"model"`
	SerialNumber         *string    `json:"serial_number" db:"serial_number"`
	PurchaseDate         *time.Time `json:"purchase_date" db:"purchase_date"`
	PurchasePrice        *float64   `json:"purchase_price" db:"purchase_price"`
	Status               string     `json:"status" db:"status"`
	MaintenanceSchedule  *string    `json:"maintenance_schedule" db:"maintenance_schedule"`
	LastMaintenance      *time.Time `json:"last_maintenance" db:"last_maintenance"`
	NextMaintenance      *time.Time `json:"next_maintenance" db:"next_maintenance"`
	Notes                *string    `json:"notes" db:"notes"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
}

// FileAttachment represents file attachments
type FileAttachment struct {
	ID               uuid.UUID  `json:"id" db:"id"`
	TenantID         uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	EntityType       string     `json:"entity_type" db:"entity_type"`
	EntityID         uuid.UUID  `json:"entity_id" db:"entity_id"`
	Filename         string     `json:"filename" db:"filename"`
	OriginalFilename string     `json:"original_filename" db:"original_filename"`
	FileSize         int64      `json:"file_size" db:"file_size"`
	ContentType      string     `json:"content_type" db:"content_type"`
	StoragePath      string     `json:"storage_path" db:"storage_path"`
	UploadedBy       *uuid.UUID `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}