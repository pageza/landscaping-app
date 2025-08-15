package types

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// Repository interfaces that were previously in services but belong here to avoid import cycles

// EquipmentRepositoryFull extends the basic equipment repository with additional functionality
type EquipmentRepositoryFull interface {
	// Basic CRUD
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Equipment, error)
	Create(ctx context.Context, equipment *domain.Equipment) error
	Update(ctx context.Context, equipment *domain.Equipment) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	
	// Extended functionality
	List(ctx context.Context, tenantID uuid.UUID, filter *EquipmentFilter) ([]*domain.Equipment, int64, error)
	GetByType(ctx context.Context, tenantID uuid.UUID, equipmentType string) ([]*domain.Equipment, error)
	GetByLocation(ctx context.Context, tenantID uuid.UUID, location string) ([]*domain.Equipment, error)
	GetByAssignedUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Equipment, error)
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Equipment, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string) ([]*domain.Equipment, error)
	UpdateStatus(ctx context.Context, tenantID, equipmentID uuid.UUID, status string) error
	BulkUpdateStatus(ctx context.Context, tenantID uuid.UUID, equipmentIDs []uuid.UUID, status string) error
	GetMaintenanceHistory(ctx context.Context, equipmentID uuid.UUID) ([]*MaintenanceRecord, error)
	GetUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]*MaintenanceSchedule, error)
	GetMaintenanceSchedule(ctx context.Context, equipmentID uuid.UUID) ([]*MaintenanceSchedule, error)
}

// MaintenanceRepository handles equipment maintenance operations
type MaintenanceRepository interface {
	// Maintenance Records
	CreateMaintenanceRecord(ctx context.Context, record *MaintenanceRecord) error
	GetMaintenanceRecord(ctx context.Context, recordID uuid.UUID) (*MaintenanceRecord, error)
	UpdateMaintenanceRecord(ctx context.Context, record *MaintenanceRecord) error
	DeleteMaintenanceRecord(ctx context.Context, recordID uuid.UUID) error
	GetMaintenanceHistory(ctx context.Context, equipmentID uuid.UUID) ([]*MaintenanceRecord, error)
	GetMaintenanceHistoryByDate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*MaintenanceRecord, error)
	
	// Maintenance Schedules
	CreateMaintenanceSchedule(ctx context.Context, schedule *MaintenanceSchedule) error
	GetMaintenanceSchedule(ctx context.Context, scheduleID uuid.UUID) (*MaintenanceSchedule, error)
	UpdateMaintenanceSchedule(ctx context.Context, schedule *MaintenanceSchedule) error
	DeleteMaintenanceSchedule(ctx context.Context, scheduleID uuid.UUID) error
	GetMaintenanceSchedulesByEquipment(ctx context.Context, equipmentID uuid.UUID) ([]*MaintenanceSchedule, error)
	GetUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]*MaintenanceSchedule, error)
	GetOverdueMaintenance(ctx context.Context, tenantID uuid.UUID) ([]*MaintenanceSchedule, error)
}

// CustomerRepository extends the basic customer repository
type CustomerRepository interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.EnhancedCustomer, error)
	Create(ctx context.Context, customer *domain.EnhancedCustomer) error
	Update(ctx context.Context, customer *domain.EnhancedCustomer) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *CustomerFilter) ([]*domain.EnhancedCustomer, int64, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string, filter *CustomerFilter) ([]*domain.EnhancedCustomer, int64, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*domain.EnhancedCustomer, error)
	GetByPhone(ctx context.Context, tenantID uuid.UUID, phone string) (*domain.EnhancedCustomer, error)
	GetCustomerSummary(ctx context.Context, tenantID, customerID uuid.UUID) (*CustomerSummary, error)
}

// PropertyRepositoryExtended extends the basic property repository
type PropertyRepositoryExtended interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.EnhancedProperty, error)
	Create(ctx context.Context, property *domain.EnhancedProperty) error
	Update(ctx context.Context, property *domain.EnhancedProperty) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *PropertyFilter) ([]*domain.EnhancedProperty, int64, error)
	Search(ctx context.Context, tenantID uuid.UUID, query string, filter *PropertyFilter) ([]*domain.EnhancedProperty, int64, error)
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID) ([]*domain.EnhancedProperty, error)
	GetByLocation(ctx context.Context, tenantID uuid.UUID, city, state string) ([]*domain.EnhancedProperty, error)
}

// JobRepositoryComplete extends the basic job repository
type JobRepositoryComplete interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.EnhancedJob, error)
	Create(ctx context.Context, job *domain.EnhancedJob) error
	Update(ctx context.Context, job *domain.EnhancedJob) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	GetByPropertyID(ctx context.Context, tenantID, propertyID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	GetByAssignedUserID(ctx context.Context, tenantID, userID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status string, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	CreateRecurringJobSeries(ctx context.Context, series *RecurringJobSeries) error
	GetRecurringJobSeries(ctx context.Context, tenantID uuid.UUID, baseJobID uuid.UUID) (*RecurringJobSeries, error)
	UpdateRecurringJobSeries(ctx context.Context, series *RecurringJobSeries) error
}

// InvoiceRepositoryFull extends the basic invoice repository
type InvoiceRepositoryFull interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Invoice, error)
	Create(ctx context.Context, invoice *domain.Invoice) error
	Update(ctx context.Context, invoice *domain.Invoice) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *InvoiceFilter) ([]*domain.Invoice, int64, error)
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *InvoiceFilter) ([]*domain.Invoice, int64, error)
	GetByJobID(ctx context.Context, tenantID, jobID uuid.UUID) (*domain.Invoice, error)
	CreateInvoiceService(ctx context.Context, invoiceService *InvoiceService) error
	UpdateInvoiceService(ctx context.Context, invoiceService *InvoiceService) error
	DeleteInvoiceService(ctx context.Context, invoiceServiceID uuid.UUID) error
	GetInvoiceServices(ctx context.Context, invoiceID uuid.UUID) ([]*InvoiceService, error)
}

// PaymentRepositoryFull extends the basic payment repository
type PaymentRepositoryFull interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Payment, error)
	Create(ctx context.Context, payment *domain.Payment) error
	Update(ctx context.Context, payment *domain.Payment) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *PaymentFilter) ([]*domain.Payment, int64, error)
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *PaymentFilter) ([]*domain.Payment, int64, error)
	GetByInvoiceID(ctx context.Context, tenantID, invoiceID uuid.UUID) ([]*domain.Payment, error)
	GetPaymentSummary(ctx context.Context, tenantID uuid.UUID, filter *PaymentFilter) (*PaymentSummary, error)
}

// QuoteRepositoryFull extends the basic quote repository
type QuoteRepositoryFull interface {
	GetByID(ctx context.Context, tenantID, id uuid.UUID) (*domain.Quote, error)
	Create(ctx context.Context, quote *domain.Quote) error
	Update(ctx context.Context, quote *domain.Quote) error
	Delete(ctx context.Context, tenantID, id uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *QuoteFilter) ([]*domain.Quote, int64, error)
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *QuoteFilter) ([]*domain.Quote, int64, error)
	GetByPropertyID(ctx context.Context, tenantID, propertyID uuid.UUID, filter *QuoteFilter) ([]*domain.Quote, int64, error)
}

// TODO: Re-enable SubscriptionRepository after creating domain.Subscription type
/*
type SubscriptionRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Subscription, error)
	Create(ctx context.Context, subscription *domain.Subscription) error
	Update(ctx context.Context, subscription *domain.Subscription) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, filter *SubscriptionFilter) ([]*domain.Subscription, int64, error)
	GetByCustomerID(ctx context.Context, customerID uuid.UUID) ([]*domain.Subscription, error)
	GetByStatus(ctx context.Context, status string) ([]*domain.Subscription, error)
	CreateUsageRecord(ctx context.Context, usage *UsageRecord) error
	GetUsageBySubscription(ctx context.Context, subscriptionID uuid.UUID, startDate, endDate time.Time) ([]*UsageRecord, error)
	GetUsageSummary(ctx context.Context, subscriptionID uuid.UUID, period string) (*UsageSummary, error)
}
*/