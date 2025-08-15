package services

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
)

// EquipmentServiceImpl implements the EquipmentService interface
type EquipmentServiceImpl struct {
	equipmentRepo       EquipmentRepositoryFull
	jobRepo             JobRepositoryComplete
	maintenanceRepo     MaintenanceRepository
	auditService        AuditService
	notificationService NotificationService
	logger              *log.Logger
}

// EquipmentRepositoryFull defines the complete interface for equipment data access
type EquipmentRepositoryFull interface {
	// CRUD operations
	Create(ctx context.Context, equipment *domain.Equipment) error
	GetByID(ctx context.Context, tenantID, equipmentID uuid.UUID) (*domain.Equipment, error)
	Update(ctx context.Context, equipment *domain.Equipment) error
	Delete(ctx context.Context, tenantID, equipmentID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *EquipmentFilter) ([]*domain.Equipment, int64, error)
	
	// Filtering operations
	GetByType(ctx context.Context, tenantID uuid.UUID, equipmentType string) ([]*domain.Equipment, error)
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Equipment, error)
	GetByIDs(ctx context.Context, tenantID uuid.UUID, equipmentIDs []uuid.UUID) ([]*domain.Equipment, error)
	
	// Availability operations
	GetAvailable(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*domain.Equipment, error)
	CheckAvailability(ctx context.Context, equipmentIDs []uuid.UUID, startTime, endTime time.Time) (map[uuid.UUID]bool, error)
	
	// Maintenance operations
	GetMaintenanceDue(ctx context.Context, tenantID uuid.UUID) ([]*domain.Equipment, error)
	UpdateMaintenanceDate(ctx context.Context, equipmentID uuid.UUID, lastMaintenance, nextMaintenance time.Time) error
}

// MaintenanceRepository defines the interface for maintenance data access
type MaintenanceRepository interface {
	// CRUD operations
	CreateMaintenanceRecord(ctx context.Context, record *MaintenanceRecord) error
	GetMaintenanceRecord(ctx context.Context, recordID uuid.UUID) (*MaintenanceRecord, error)
	UpdateMaintenanceRecord(ctx context.Context, record *MaintenanceRecord) error
	DeleteMaintenanceRecord(ctx context.Context, recordID uuid.UUID) error
	
	// Equipment maintenance history
	GetMaintenanceHistory(ctx context.Context, equipmentID uuid.UUID) ([]*MaintenanceRecord, error)
	GetMaintenanceHistoryByDate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*MaintenanceRecord, error)
	
	// Maintenance scheduling
	CreateMaintenanceSchedule(ctx context.Context, schedule *MaintenanceSchedule) error
	GetMaintenanceSchedule(ctx context.Context, scheduleID uuid.UUID) (*MaintenanceSchedule, error)
	UpdateMaintenanceSchedule(ctx context.Context, schedule *MaintenanceSchedule) error
	DeleteMaintenanceSchedule(ctx context.Context, scheduleID uuid.UUID) error
	GetUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]*MaintenanceSchedule, error)
	GetMaintenanceSchedulesByEquipment(ctx context.Context, equipmentID uuid.UUID) ([]*MaintenanceSchedule, error)
}

// NewEquipmentService creates a new equipment service instance
func NewEquipmentService(
	equipmentRepo EquipmentRepositoryFull,
	jobRepo JobRepositoryComplete,
	maintenanceRepo MaintenanceRepository,
	auditService AuditService,
	notificationService NotificationService,
	logger *log.Logger,
) EquipmentService {
	return &EquipmentServiceImpl{
		equipmentRepo:       equipmentRepo,
		jobRepo:             jobRepo,
		maintenanceRepo:     maintenanceRepo,
		auditService:        auditService,
		notificationService: notificationService,
		logger:              logger,
	}
}

// CreateEquipment creates a new equipment item
func (s *EquipmentServiceImpl) CreateEquipment(ctx context.Context, req *EquipmentCreateRequest) (*domain.Equipment, error) {
	// Get tenant ID from context
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Validate the request
	if err := s.validateCreateEquipmentRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Create equipment entity
	equipment := &domain.Equipment{
		ID:                  uuid.New(),
		TenantID:            tenantID,
		Name:                req.Name,
		Type:                req.Type,
		Model:               req.Model,
		SerialNumber:        req.SerialNumber,
		PurchaseDate:        req.PurchaseDate,
		PurchasePrice:       req.PurchasePrice,
		Status:              "available",
		MaintenanceSchedule: req.MaintenanceSchedule,
		Notes:               req.Notes,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	// Set next maintenance date if maintenance schedule is provided
	if req.MaintenanceSchedule != nil {
		nextMaintenance := s.calculateNextMaintenanceDate(*req.MaintenanceSchedule, time.Now())
		equipment.NextMaintenance = &nextMaintenance
	}

	// Save to database
	if err := s.equipmentRepo.Create(ctx, equipment); err != nil {
		s.logger.Printf("Failed to create equipment", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to create equipment: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "equipment.create",
		ResourceType: "equipment",
		ResourceID:   &equipment.ID,
		NewValues: map[string]interface{}{
			"name":          equipment.Name,
			"type":          equipment.Type,
			"status":        equipment.Status,
			"serial_number": equipment.SerialNumber,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Equipment created successfully", "equipment_id", equipment.ID, "name", equipment.Name, "tenant_id", tenantID)
	return equipment, nil
}

// GetEquipment retrieves equipment by ID
func (s *EquipmentServiceImpl) GetEquipment(ctx context.Context, equipmentID uuid.UUID) (*domain.Equipment, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		s.logger.Printf("Failed to get equipment", "error", err, "equipment_id", equipmentID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	return equipment, nil
}

// UpdateEquipment updates an existing equipment item
func (s *EquipmentServiceImpl) UpdateEquipment(ctx context.Context, equipmentID uuid.UUID, req *EquipmentUpdateRequest) (*domain.Equipment, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get existing equipment
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	// Store old values for audit
	oldValues := map[string]interface{}{
		"name":   equipment.Name,
		"status": equipment.Status,
		"type":   equipment.Type,
	}

	// Update fields
	if req.Name != nil {
		equipment.Name = *req.Name
	}
	if req.Type != nil {
		equipment.Type = *req.Type
	}
	if req.Model != nil {
		equipment.Model = req.Model
	}
	if req.SerialNumber != nil {
		equipment.SerialNumber = req.SerialNumber
	}
	if req.PurchaseDate != nil {
		equipment.PurchaseDate = req.PurchaseDate
	}
	if req.PurchasePrice != nil {
		equipment.PurchasePrice = req.PurchasePrice
	}
	if req.Status != nil {
		equipment.Status = *req.Status
	}
	if req.MaintenanceSchedule != nil {
		equipment.MaintenanceSchedule = req.MaintenanceSchedule
		// Recalculate next maintenance date
		if equipment.LastMaintenance != nil {
			nextMaintenance := s.calculateNextMaintenanceDate(*req.MaintenanceSchedule, *equipment.LastMaintenance)
			equipment.NextMaintenance = &nextMaintenance
		} else {
			nextMaintenance := s.calculateNextMaintenanceDate(*req.MaintenanceSchedule, time.Now())
			equipment.NextMaintenance = &nextMaintenance
		}
	}
	if req.Notes != nil {
		equipment.Notes = req.Notes
	}

	equipment.UpdatedAt = time.Now()

	// Validate the updated equipment
	if err := s.validateEquipment(equipment); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to database
	if err := s.equipmentRepo.Update(ctx, equipment); err != nil {
		s.logger.Printf("Failed to update equipment", "error", err, "equipment_id", equipmentID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update equipment: %w", err)
	}

	// Log audit event
	newValues := map[string]interface{}{
		"name":   equipment.Name,
		"status": equipment.Status,
		"type":   equipment.Type,
	}

	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "equipment.update",
		ResourceType: "equipment",
		ResourceID:   &equipment.ID,
		OldValues:    oldValues,
		NewValues:    newValues,
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Equipment updated successfully", "equipment_id", equipmentID, "tenant_id", tenantID)
	return equipment, nil
}

// DeleteEquipment deletes equipment (soft delete)
func (s *EquipmentServiceImpl) DeleteEquipment(ctx context.Context, equipmentID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get equipment before deletion for audit log
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return fmt.Errorf("equipment not found")
	}

	// Check if equipment is currently in use
	if equipment.Status == "in_use" {
		return fmt.Errorf("cannot delete equipment that is currently in use")
	}

	// Soft delete by updating status
	equipment.Status = "retired"
	equipment.UpdatedAt = time.Now()

	if err := s.equipmentRepo.Update(ctx, equipment); err != nil {
		s.logger.Printf("Failed to delete equipment", "error", err, "equipment_id", equipmentID, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete equipment: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "equipment.delete",
		ResourceType: "equipment",
		ResourceID:   &equipment.ID,
		OldValues: map[string]interface{}{
			"status": "available",
		},
		NewValues: map[string]interface{}{
			"status": "retired",
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Equipment deleted successfully", "equipment_id", equipmentID, "tenant_id", tenantID)
	return nil
}

// ListEquipment lists equipment with filtering and pagination
func (s *EquipmentServiceImpl) ListEquipment(ctx context.Context, filter *EquipmentFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Set defaults
	if filter == nil {
		filter = &EquipmentFilter{}
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PerPage <= 0 {
		filter.PerPage = 50
	}
	if filter.PerPage > 100 {
		filter.PerPage = 100
	}

	equipment, total, err := s.equipmentRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Printf("Failed to list equipment", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to list equipment: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       equipment,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// GetAvailableEquipment retrieves available equipment for a time period
func (s *EquipmentServiceImpl) GetAvailableEquipment(ctx context.Context, startDate, endDate time.Time) ([]*domain.Equipment, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	equipment, err := s.equipmentRepo.GetAvailable(ctx, tenantID, startDate, endDate)
	if err != nil {
		s.logger.Printf("Failed to get available equipment", "error", err, "start_date", startDate, "end_date", endDate, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get available equipment: %w", err)
	}

	return equipment, nil
}

// CheckEquipmentAvailability checks if specific equipment is available
func (s *EquipmentServiceImpl) CheckEquipmentAvailability(ctx context.Context, equipmentID uuid.UUID, startDate, endDate time.Time) (bool, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return false, fmt.Errorf("tenant ID not found in context")
	}

	// Check if equipment exists and is not retired
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return false, fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return false, fmt.Errorf("equipment not found")
	}
	if equipment.Status == "retired" || equipment.Status == "maintenance" {
		return false, nil
	}

	// Check availability using repository
	availability, err := s.equipmentRepo.CheckAvailability(ctx, []uuid.UUID{equipmentID}, startDate, endDate)
	if err != nil {
		s.logger.Printf("Failed to check equipment availability", "error", err, "equipment_id", equipmentID)
		return false, fmt.Errorf("failed to check equipment availability: %w", err)
	}

	available, exists := availability[equipmentID]
	return exists && available, nil
}

// ScheduleMaintenance schedules maintenance for equipment
func (s *EquipmentServiceImpl) ScheduleMaintenance(ctx context.Context, equipmentID uuid.UUID, req *MaintenanceScheduleRequest) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Validate request
	if err := s.validateMaintenanceScheduleRequest(req); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Verify equipment exists
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return fmt.Errorf("equipment not found")
	}

	// Create maintenance schedule
	schedule := &MaintenanceSchedule{
		ID:            uuid.New(),
		EquipmentID:   equipmentID,
		EquipmentName: equipment.Name,
		Type:          req.Type,
		ScheduledDate: req.ScheduledDate,
		Description:   req.Description,
		Priority:      req.Priority,
		Status:        "scheduled",
	}

	if err := s.maintenanceRepo.CreateMaintenanceSchedule(ctx, schedule); err != nil {
		s.logger.Printf("Failed to schedule maintenance", "error", err, "equipment_id", equipmentID)
		return fmt.Errorf("failed to schedule maintenance: %w", err)
	}

	// Send notification to maintenance team
	if err := s.notificationService.SendNotification(ctx, &NotificationRequest{
		Type:    "maintenance.scheduled",
		Title:   "Maintenance Scheduled",
		Message: fmt.Sprintf("Maintenance scheduled for %s on %s", equipment.Name, req.ScheduledDate.Format("January 2, 2006")),
		Data: map[string]interface{}{
			"equipment_id":   equipmentID,
			"equipment_name": equipment.Name,
			"scheduled_date": req.ScheduledDate,
			"type":           req.Type,
		},
	}); err != nil {
		s.logger.Printf("Failed to send maintenance notification", "error", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "equipment.schedule_maintenance",
		ResourceType: "equipment",
		ResourceID:   &equipmentID,
		NewValues: map[string]interface{}{
			"maintenance_type": req.Type,
			"scheduled_date":   req.ScheduledDate,
			"priority":         req.Priority,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Maintenance scheduled successfully", "equipment_id", equipmentID, "scheduled_date", req.ScheduledDate)
	return nil
}

// GetMaintenanceHistory retrieves maintenance history for equipment
func (s *EquipmentServiceImpl) GetMaintenanceHistory(ctx context.Context, equipmentID uuid.UUID) ([]*MaintenanceRecord, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify equipment exists
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	history, err := s.maintenanceRepo.GetMaintenanceHistory(ctx, equipmentID)
	if err != nil {
		s.logger.Printf("Failed to get maintenance history", "error", err, "equipment_id", equipmentID)
		return nil, fmt.Errorf("failed to get maintenance history: %w", err)
	}

	return history, nil
}

// GetUpcomingMaintenance retrieves upcoming maintenance schedules
func (s *EquipmentServiceImpl) GetUpcomingMaintenance(ctx context.Context) ([]*MaintenanceSchedule, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	schedules, err := s.maintenanceRepo.GetUpcomingMaintenance(ctx, tenantID)
	if err != nil {
		s.logger.Printf("Failed to get upcoming maintenance", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get upcoming maintenance: %w", err)
	}

	return schedules, nil
}

// PerformMaintenance records completed maintenance
func (s *EquipmentServiceImpl) PerformMaintenance(ctx context.Context, equipmentID uuid.UUID, maintenanceType string, cost *float64, notes *string) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get equipment
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return fmt.Errorf("equipment not found")
	}

	now := time.Now()

	// Create maintenance record
	record := &MaintenanceRecord{
		ID:            uuid.New(),
		EquipmentID:   equipmentID,
		Type:          maintenanceType,
		PerformedDate: now,
		Description:   fmt.Sprintf("%s maintenance performed", maintenanceType),
		Cost:          cost,
		PerformedBy:   GetUserIDFromContext(ctx),
		Notes:         notes,
	}

	if err := s.maintenanceRepo.CreateMaintenanceRecord(ctx, record); err != nil {
		s.logger.Printf("Failed to create maintenance record", "error", err, "equipment_id", equipmentID)
		return fmt.Errorf("failed to create maintenance record: %w", err)
	}

	// Update equipment maintenance dates
	var nextMaintenance time.Time
	if equipment.MaintenanceSchedule != nil {
		nextMaintenance = s.calculateNextMaintenanceDate(*equipment.MaintenanceSchedule, now)
	} else {
		// Default to 3 months if no schedule
		nextMaintenance = now.AddDate(0, 3, 0)
	}

	if err := s.equipmentRepo.UpdateMaintenanceDate(ctx, equipmentID, now, nextMaintenance); err != nil {
		s.logger.Printf("Failed to update equipment maintenance dates", "error", err, "equipment_id", equipmentID)
	}

	// Update equipment status back to available if it was in maintenance
	if equipment.Status == "maintenance" {
		equipment.Status = "available"
		equipment.UpdatedAt = time.Now()
		if err := s.equipmentRepo.Update(ctx, equipment); err != nil {
			s.logger.Printf("Failed to update equipment status after maintenance", "error", err)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "equipment.maintenance_performed",
		ResourceType: "equipment",
		ResourceID:   &equipmentID,
		NewValues: map[string]interface{}{
			"maintenance_type":   maintenanceType,
			"performed_date":     now,
			"cost":              cost,
			"next_maintenance":  nextMaintenance,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Maintenance performed successfully", "equipment_id", equipmentID, "type", maintenanceType)
	return nil
}

// CheckMaintenanceDue checks for equipment with maintenance due
func (s *EquipmentServiceImpl) CheckMaintenanceDue(ctx context.Context) ([]*domain.Equipment, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	equipment, err := s.equipmentRepo.GetMaintenanceDue(ctx, tenantID)
	if err != nil {
		s.logger.Printf("Failed to get equipment with maintenance due", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get equipment with maintenance due: %w", err)
	}

	// Send notifications for overdue maintenance
	for _, eq := range equipment {
		if eq.NextMaintenance != nil && eq.NextMaintenance.Before(time.Now()) {
			if err := s.notificationService.SendNotification(ctx, &NotificationRequest{
				Type:    "maintenance.overdue",
				Title:   "Maintenance Overdue",
				Message: fmt.Sprintf("Maintenance is overdue for %s", eq.Name),
				Data: map[string]interface{}{
					"equipment_id":   eq.ID,
					"equipment_name": eq.Name,
					"due_date":      eq.NextMaintenance,
				},
			}); err != nil {
				s.logger.Printf("Failed to send overdue maintenance notification", "error", err, "equipment_id", eq.ID)
			}
		}
	}

	return equipment, nil
}

// GetEquipmentUtilization gets utilization statistics for equipment
func (s *EquipmentServiceImpl) GetEquipmentUtilization(ctx context.Context, equipmentID uuid.UUID, startDate, endDate time.Time) (*EquipmentUtilization, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify equipment exists
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	// Get jobs that used this equipment
	jobs, err := s.jobRepo.GetByEquipmentID(ctx, tenantID, equipmentID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment jobs: %w", err)
	}

	totalDays := int(endDate.Sub(startDate).Hours() / 24)
	utilizationDays := len(jobs)
	utilizationPercentage := float64(utilizationDays) / float64(totalDays) * 100

	var totalRevenue float64
	var totalHours int
	for _, job := range jobs {
		if job.TotalAmount != nil {
			totalRevenue += *job.TotalAmount
		}
		if job.ActualStartTime != nil && job.ActualEndTime != nil {
			duration := job.ActualEndTime.Sub(*job.ActualStartTime)
			totalHours += int(duration.Hours())
		}
	}

	utilization := &EquipmentUtilization{
		EquipmentID:           equipmentID,
		EquipmentName:         equipment.Name,
		Period:                TimeRange{Start: startDate, End: endDate},
		TotalJobs:             len(jobs),
		TotalHours:            totalHours,
		TotalRevenue:          totalRevenue,
		UtilizationPercentage: utilizationPercentage,
		AverageJobDuration:    0,
		RevenuePerHour:        0,
	}

	if totalHours > 0 {
		utilization.AverageJobDuration = totalHours / len(jobs)
		utilization.RevenuePerHour = totalRevenue / float64(totalHours)
	}

	return utilization, nil
}

// GetMaintenanceCosts gets maintenance cost analytics for equipment
func (s *EquipmentServiceImpl) GetMaintenanceCosts(ctx context.Context, equipmentID uuid.UUID, startDate, endDate time.Time) (*MaintenanceCostAnalysis, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify equipment exists
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	// Get maintenance records for the period
	maintenanceRecords, err := s.maintenanceRepo.GetMaintenanceHistoryByDate(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance history: %w", err)
	}

	// Filter for this equipment
	var equipmentRecords []*MaintenanceRecord
	var totalCost float64
	maintenanceTypes := make(map[string]float64)

	for _, record := range maintenanceRecords {
		if record.EquipmentID == equipmentID {
			equipmentRecords = append(equipmentRecords, record)
			if record.Cost != nil {
				totalCost += *record.Cost
				maintenanceTypes[record.Type] += *record.Cost
			}
		}
	}

	var averageCost float64
	if len(equipmentRecords) > 0 {
		averageCost = totalCost / float64(len(equipmentRecords))
	}

	analysis := &MaintenanceCostAnalysis{
		EquipmentID:       equipmentID,
		EquipmentName:     equipment.Name,
		Period:            TimeRange{Start: startDate, End: endDate},
		TotalCost:         totalCost,
		AverageCost:       averageCost,
		MaintenanceCount:  len(equipmentRecords),
		CostByType:        maintenanceTypes,
		Records:           equipmentRecords,
	}

	return analysis, nil
}

// PredictMaintenanceNeeds uses analytics to predict future maintenance needs
func (s *EquipmentServiceImpl) PredictMaintenanceNeeds(ctx context.Context, equipmentID uuid.UUID) (*MaintenancePrediction, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify equipment exists
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	// Get historical maintenance data
	maintenanceHistory, err := s.maintenanceRepo.GetMaintenanceHistory(ctx, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance history: %w", err)
	}

	// Get utilization data for the last year
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	utilization, err := s.GetEquipmentUtilization(ctx, equipmentID, oneYearAgo, time.Now())
	if err != nil {
		s.logger.Printf("Failed to get utilization data for prediction", "error", err)
	}

	// Simple prediction algorithm based on historical data
	prediction := &MaintenancePrediction{
		EquipmentID:   equipmentID,
		EquipmentName: equipment.Name,
		GeneratedAt:   time.Now(),
		Confidence:    0.7, // Default confidence
		Recommendations: []MaintenanceRecommendation{},
	}

	// Analyze maintenance frequency
	if len(maintenanceHistory) >= 2 {
		// Calculate average time between maintenance
		var totalDays int
		for i := 1; i < len(maintenanceHistory); i++ {
			days := int(maintenanceHistory[i].PerformedDate.Sub(maintenanceHistory[i-1].PerformedDate).Hours() / 24)
			totalDays += days
		}
		avgDaysBetweenMaintenance := totalDays / (len(maintenanceHistory) - 1)

		// Predict next maintenance based on last maintenance + average interval
		lastMaintenance := maintenanceHistory[len(maintenanceHistory)-1].PerformedDate
		predictedNext := lastMaintenance.AddDate(0, 0, avgDaysBetweenMaintenance)
		prediction.NextMaintenanceDate = &predictedNext

		// Add recommendations based on patterns
		if utilization != nil && utilization.UtilizationPercentage > 80 {
			prediction.Recommendations = append(prediction.Recommendations, MaintenanceRecommendation{
				Type:        "preventive",
				Priority:    "high",
				Description: "High utilization detected - consider more frequent preventive maintenance",
				EstimatedCost: 200.0,
			})
		}

		// Check for overdue maintenance
		if equipment.NextMaintenance != nil && equipment.NextMaintenance.Before(time.Now()) {
			prediction.Recommendations = append(prediction.Recommendations, MaintenanceRecommendation{
				Type:        "urgent",
				Priority:    "urgent",
				Description: "Maintenance is overdue - schedule immediately",
				EstimatedCost: 300.0,
			})
		}
	}

	return prediction, nil
}

// GetEquipmentPerformanceMetrics gets performance metrics for equipment
func (s *EquipmentServiceImpl) GetEquipmentPerformanceMetrics(ctx context.Context, equipmentID uuid.UUID) (*EquipmentPerformanceMetrics, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Verify equipment exists
	equipment, err := s.equipmentRepo.GetByID(ctx, tenantID, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}
	if equipment == nil {
		return nil, fmt.Errorf("equipment not found")
	}

	// Get various metrics
	oneYearAgo := time.Now().AddDate(-1, 0, 0)
	utilization, err := s.GetEquipmentUtilization(ctx, equipmentID, oneYearAgo, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get utilization: %w", err)
	}

	costAnalysis, err := s.GetMaintenanceCosts(ctx, equipmentID, oneYearAgo, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance costs: %w", err)
	}

	// Calculate ROI
	var roi float64
	if equipment.PurchasePrice != nil && *equipment.PurchasePrice > 0 {
		roi = (utilization.TotalRevenue - costAnalysis.TotalCost) / *equipment.PurchasePrice * 100
	}

	// Calculate efficiency score
	efficiencyScore := s.calculateEfficiencyScore(utilization, costAnalysis, equipment)

	metrics := &EquipmentPerformanceMetrics{
		EquipmentID:       equipmentID,
		EquipmentName:     equipment.Name,
		Period:            TimeRange{Start: oneYearAgo, End: time.Now()},
		UtilizationRate:   utilization.UtilizationPercentage,
		Revenue:           utilization.TotalRevenue,
		MaintenanceCost:   costAnalysis.TotalCost,
		ROI:              roi,
		EfficiencyScore:  efficiencyScore,
		DowntimeHours:    0, // Would calculate from maintenance records
		ReliabilityScore: s.calculateReliabilityScore(costAnalysis.Records),
	}

	return metrics, nil
}

// Helper methods

func (s *EquipmentServiceImpl) validateCreateEquipmentRequest(req *EquipmentCreateRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("equipment name is required")
	}
	if strings.TrimSpace(req.Type) == "" {
		return fmt.Errorf("equipment type is required")
	}
	if req.PurchasePrice != nil && *req.PurchasePrice < 0 {
		return fmt.Errorf("purchase price cannot be negative")
	}
	return nil
}

func (s *EquipmentServiceImpl) validateEquipment(equipment *domain.Equipment) error {
	if strings.TrimSpace(equipment.Name) == "" {
		return fmt.Errorf("equipment name is required")
	}
	if strings.TrimSpace(equipment.Type) == "" {
		return fmt.Errorf("equipment type is required")
	}
	validStatuses := []string{"available", "in_use", "maintenance", "retired"}
	isValidStatus := false
	for _, status := range validStatuses {
		if equipment.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return fmt.Errorf("invalid equipment status: %s", equipment.Status)
	}
	return nil
}

func (s *EquipmentServiceImpl) validateMaintenanceScheduleRequest(req *MaintenanceScheduleRequest) error {
	if strings.TrimSpace(req.Type) == "" {
		return fmt.Errorf("maintenance type is required")
	}
	if req.ScheduledDate.Before(time.Now()) {
		return fmt.Errorf("scheduled date cannot be in the past")
	}
	validPriorities := []string{"low", "medium", "high", "urgent"}
	isValidPriority := false
	for _, priority := range validPriorities {
		if req.Priority == priority {
			isValidPriority = true
			break
		}
	}
	if !isValidPriority {
		return fmt.Errorf("invalid priority: %s", req.Priority)
	}
	return nil
}

func (s *EquipmentServiceImpl) calculateNextMaintenanceDate(schedule string, lastMaintenance time.Time) time.Time {
	switch strings.ToLower(schedule) {
	case "weekly":
		return lastMaintenance.AddDate(0, 0, 7)
	case "monthly":
		return lastMaintenance.AddDate(0, 1, 0)
	case "quarterly":
		return lastMaintenance.AddDate(0, 3, 0)
	case "semi-annually":
		return lastMaintenance.AddDate(0, 6, 0)
	case "annually":
		return lastMaintenance.AddDate(1, 0, 0)
	default:
		// Default to quarterly maintenance
		return lastMaintenance.AddDate(0, 3, 0)
	}
}

func (s *EquipmentServiceImpl) calculateEfficiencyScore(utilization *EquipmentUtilization, costAnalysis *MaintenanceCostAnalysis, equipment *domain.Equipment) float64 {
	// Simple efficiency calculation based on utilization and costs
	baseScore := utilization.UtilizationPercentage

	// Adjust for maintenance costs
	if equipment.PurchasePrice != nil && *equipment.PurchasePrice > 0 {
		maintenanceCostRatio := costAnalysis.TotalCost / *equipment.PurchasePrice
		if maintenanceCostRatio > 0.1 { // More than 10% of purchase price in maintenance
			baseScore *= 0.8 // Reduce score
		}
	}

	// Cap at 100
	if baseScore > 100 {
		baseScore = 100
	}

	return baseScore
}

func (s *EquipmentServiceImpl) calculateReliabilityScore(maintenanceRecords []*MaintenanceRecord) float64 {
	if len(maintenanceRecords) == 0 {
		return 100.0 // No maintenance needed = reliable
	}

	// Simple calculation: fewer maintenance incidents = higher reliability
	// In a real system, this would be more sophisticated
	
	urgentMaintenanceCount := 0
	for _, record := range maintenanceRecords {
		if strings.Contains(strings.ToLower(record.Type), "urgent") || 
		   strings.Contains(strings.ToLower(record.Type), "emergency") {
			urgentMaintenanceCount++
		}
	}

	// Base score of 100, reduce by 10 for each urgent maintenance
	score := 100.0 - float64(urgentMaintenanceCount*10)
	if score < 0 {
		score = 0
	}

	return score
}