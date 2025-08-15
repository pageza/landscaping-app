package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/types"
)

// EquipmentRepositoryImpl implements the full equipment repository interface
type EquipmentRepositoryImpl struct {
	db *Database
}

// NewEquipmentRepositoryFull creates a new full equipment repository
func NewEquipmentRepositoryFull(db *Database) types.EquipmentRepositoryFull {
	return &EquipmentRepositoryImpl{db: db}
}

// Create creates a new equipment item
func (r *EquipmentRepositoryImpl) Create(ctx context.Context, equipment *domain.Equipment) error {
	query := `
		INSERT INTO equipment (
			id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price,
			status, maintenance_schedule, last_maintenance, next_maintenance, notes,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)`

	_, err := r.db.ExecContext(ctx, query,
		equipment.ID,
		equipment.TenantID,
		equipment.Name,
		equipment.Type,
		equipment.Model,
		equipment.SerialNumber,
		equipment.PurchaseDate,
		equipment.PurchasePrice,
		equipment.Status,
		equipment.MaintenanceSchedule,
		equipment.LastMaintenance,
		equipment.NextMaintenance,
		equipment.Notes,
		equipment.CreatedAt,
		equipment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create equipment: %w", err)
	}

	return nil
}

// GetByID retrieves equipment by ID
func (r *EquipmentRepositoryImpl) GetByID(ctx context.Context, tenantID, equipmentID uuid.UUID) (*domain.Equipment, error) {
	query := `
		SELECT id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price,
			   status, maintenance_schedule, last_maintenance, next_maintenance, notes,
			   created_at, updated_at
		FROM equipment
		WHERE id = $1 AND tenant_id = $2 AND status != 'deleted'`

	var equipment domain.Equipment
	err := r.db.QueryRowContext(ctx, query, equipmentID, tenantID).Scan(
		&equipment.ID,
		&equipment.TenantID,
		&equipment.Name,
		&equipment.Type,
		&equipment.Model,
		&equipment.SerialNumber,
		&equipment.PurchaseDate,
		&equipment.PurchasePrice,
		&equipment.Status,
		&equipment.MaintenanceSchedule,
		&equipment.LastMaintenance,
		&equipment.NextMaintenance,
		&equipment.Notes,
		&equipment.CreatedAt,
		&equipment.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get equipment: %w", err)
	}

	return &equipment, nil
}

// Update updates an existing equipment item
func (r *EquipmentRepositoryImpl) Update(ctx context.Context, equipment *domain.Equipment) error {
	query := `
		UPDATE equipment SET
			name = $3, type = $4, model = $5, serial_number = $6, purchase_date = $7,
			purchase_price = $8, status = $9, maintenance_schedule = $10,
			last_maintenance = $11, next_maintenance = $12, notes = $13, updated_at = $14
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		equipment.ID,
		equipment.TenantID,
		equipment.Name,
		equipment.Type,
		equipment.Model,
		equipment.SerialNumber,
		equipment.PurchaseDate,
		equipment.PurchasePrice,
		equipment.Status,
		equipment.MaintenanceSchedule,
		equipment.LastMaintenance,
		equipment.NextMaintenance,
		equipment.Notes,
		equipment.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update equipment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("equipment not found")
	}

	return nil
}

// Delete deletes equipment (soft delete by updating status)
func (r *EquipmentRepositoryImpl) Delete(ctx context.Context, tenantID, equipmentID uuid.UUID) error {
	query := `
		UPDATE equipment SET
			status = 'deleted',
			updated_at = $3
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, equipmentID, tenantID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete equipment: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("equipment not found")
	}

	return nil
}

// List lists equipment with filtering and pagination
func (r *EquipmentRepositoryImpl) List(ctx context.Context, tenantID uuid.UUID, filter *types.EquipmentFilter) ([]*domain.Equipment, int64, error) {
	baseQuery := `
		FROM equipment
		WHERE tenant_id = $1 AND status != 'deleted'`

	var conditions []string
	var args []interface{}
	args = append(args, tenantID)
	argIndex := 2

	// Apply filters
	if filter.Type != nil && *filter.Type != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, *filter.Type)
		argIndex++
	}

	if filter.Status != nil && *filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
		args = append(args, *filter.Status)
		argIndex++
	}

	// Note: Available and Maintenance fields don't exist in EquipmentFilter
	// These filters have been removed

	if filter.Search != nil && *filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR type ILIKE $%d OR model ILIKE $%d)", argIndex, argIndex, argIndex))
		args = append(args, "%"+*filter.Search+"%")
		argIndex++
	}

	whereClause := baseQuery
	if len(conditions) > 0 {
		whereClause += " AND " + strings.Join(conditions, " AND ")
	}

	// Count query
	countQuery := "SELECT COUNT(*) " + whereClause
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count equipment: %w", err)
	}

	// Main query with pagination
	selectFields := `
		SELECT id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price,
			   status, maintenance_schedule, last_maintenance, next_maintenance, notes,
			   created_at, updated_at`

	orderBy := " ORDER BY name ASC"
	// Note: SortBy and SortDesc fields don't exist in EquipmentFilter
	// Default ordering by name
	orderBy = " ORDER BY name ASC"

	limit := fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, filter.PerPage, (filter.Page-1)*filter.PerPage)

	query := selectFields + whereClause + orderBy + limit

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list equipment: %w", err)
	}
	defer rows.Close()

	var equipment []*domain.Equipment
	for rows.Next() {
		var eq domain.Equipment
		err := rows.Scan(
			&eq.ID,
			&eq.TenantID,
			&eq.Name,
			&eq.Type,
			&eq.Model,
			&eq.SerialNumber,
			&eq.PurchaseDate,
			&eq.PurchasePrice,
			&eq.Status,
			&eq.MaintenanceSchedule,
			&eq.LastMaintenance,
			&eq.NextMaintenance,
			&eq.Notes,
			&eq.CreatedAt,
			&eq.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, &eq)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating equipment rows: %w", err)
	}

	return equipment, total, nil
}

// GetByType retrieves equipment by type
func (r *EquipmentRepositoryImpl) GetByType(ctx context.Context, tenantID uuid.UUID, equipmentType string) ([]*domain.Equipment, error) {
	query := `
		SELECT id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price,
			   status, maintenance_schedule, last_maintenance, next_maintenance, notes,
			   created_at, updated_at
		FROM equipment
		WHERE tenant_id = $1 AND type = $2 AND status != 'deleted'
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, equipmentType)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment by type: %w", err)
	}
	defer rows.Close()

	var equipment []*domain.Equipment
	for rows.Next() {
		var eq domain.Equipment
		err := rows.Scan(
			&eq.ID,
			&eq.TenantID,
			&eq.Name,
			&eq.Type,
			&eq.Model,
			&eq.SerialNumber,
			&eq.PurchaseDate,
			&eq.PurchasePrice,
			&eq.Status,
			&eq.MaintenanceSchedule,
			&eq.LastMaintenance,
			&eq.NextMaintenance,
			&eq.Notes,
			&eq.CreatedAt,
			&eq.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, &eq)
	}

	return equipment, nil
}

// GetByStatus retrieves equipment by status
func (r *EquipmentRepositoryImpl) GetByStatus(ctx context.Context, tenantID uuid.UUID, status string) ([]*domain.Equipment, error) {
	query := `
		SELECT id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price,
			   status, maintenance_schedule, last_maintenance, next_maintenance, notes,
			   created_at, updated_at
		FROM equipment
		WHERE tenant_id = $1 AND status = $2
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, status)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment by status: %w", err)
	}
	defer rows.Close()

	var equipment []*domain.Equipment
	for rows.Next() {
		var eq domain.Equipment
		err := rows.Scan(
			&eq.ID,
			&eq.TenantID,
			&eq.Name,
			&eq.Type,
			&eq.Model,
			&eq.SerialNumber,
			&eq.PurchaseDate,
			&eq.PurchasePrice,
			&eq.Status,
			&eq.MaintenanceSchedule,
			&eq.LastMaintenance,
			&eq.NextMaintenance,
			&eq.Notes,
			&eq.CreatedAt,
			&eq.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, &eq)
	}

	return equipment, nil
}

// GetByAssignedUser retrieves equipment assigned to a specific user
func (r *EquipmentRepositoryImpl) GetByAssignedUser(ctx context.Context, tenantID, userID uuid.UUID) ([]*domain.Equipment, error) {
	query := `
		SELECT id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price,
			   status, maintenance_schedule, last_maintenance, next_maintenance, notes,
			   created_at, updated_at
		FROM equipment
		WHERE tenant_id = $1 AND assigned_user_id = $2 AND status != 'deleted'
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment by assigned user: %w", err)
	}
	defer rows.Close()

	var equipment []*domain.Equipment
	for rows.Next() {
		eq := &domain.Equipment{}
		err := rows.Scan(
			&eq.ID,
			&eq.TenantID,
			&eq.Name,
			&eq.Type,
			&eq.Model,
			&eq.SerialNumber,
			&eq.PurchaseDate,
			&eq.PurchasePrice,
			&eq.Status,
			&eq.MaintenanceSchedule,
			&eq.LastMaintenance,
			&eq.NextMaintenance,
			&eq.Notes,
			&eq.CreatedAt,
			&eq.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, eq)
	}

	return equipment, nil
}

// GetByLocation retrieves equipment by location
func (r *EquipmentRepositoryImpl) GetByLocation(ctx context.Context, tenantID uuid.UUID, location string) ([]*domain.Equipment, error) {
	query := `
		SELECT id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price,
			   status, maintenance_schedule, last_maintenance, next_maintenance, notes,
			   created_at, updated_at
		FROM equipment
		WHERE tenant_id = $1 AND location = $2 AND status != 'deleted'
		ORDER BY name ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment by location: %w", err)
	}
	defer rows.Close()

	var equipment []*domain.Equipment
	for rows.Next() {
		eq := &domain.Equipment{}
		err := rows.Scan(
			&eq.ID,
			&eq.TenantID,
			&eq.Name,
			&eq.Type,
			&eq.Model,
			&eq.SerialNumber,
			&eq.PurchaseDate,
			&eq.PurchasePrice,
			&eq.Status,
			&eq.MaintenanceSchedule,
			&eq.LastMaintenance,
			&eq.NextMaintenance,
			&eq.Notes,
			&eq.CreatedAt,
			&eq.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, eq)
	}

	return equipment, nil
}

// GetMaintenanceHistory retrieves maintenance history for equipment
func (r *EquipmentRepositoryImpl) GetMaintenanceHistory(ctx context.Context, equipmentID uuid.UUID) ([]*types.MaintenanceRecord, error) {
	query := `
		SELECT id, equipment_id, maintenance_type, maintenance_date, description, total_cost, performed_by, notes
		FROM maintenance_records
		WHERE equipment_id = $1
		ORDER BY maintenance_date DESC`

	rows, err := r.db.QueryContext(ctx, query, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance history: %w", err)
	}
	defer rows.Close()

	var records []*types.MaintenanceRecord
	for rows.Next() {
		var record types.MaintenanceRecord
		err := rows.Scan(
			&record.ID,
			&record.EquipmentID,
			&record.MaintenanceType,
			&record.MaintenanceDate,
			&record.Description,
			&record.TotalCost,
			&record.PerformedBy,
			&record.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance record: %w", err)
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetMaintenanceSchedule retrieves maintenance schedules for equipment (Equipment repository method)
func (r *EquipmentRepositoryImpl) GetMaintenanceSchedule(ctx context.Context, equipmentID uuid.UUID) ([]*types.MaintenanceSchedule, error) {
	query := `
		SELECT id, equipment_id, maintenance_type, next_due, description, priority, active
		FROM maintenance_schedules
		WHERE equipment_id = $1
		ORDER BY next_due ASC`

	rows, err := r.db.QueryContext(ctx, query, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance schedules for equipment: %w", err)
	}
	defer rows.Close()

	var schedules []*types.MaintenanceSchedule
	for rows.Next() {
		var schedule types.MaintenanceSchedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.EquipmentID,
			&schedule.MaintenanceType,
			&schedule.NextDue,
			&schedule.Description,
			&schedule.Priority,
			&schedule.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// GetUpcomingMaintenance retrieves upcoming maintenance for equipment (Equipment repository method)
func (r *EquipmentRepositoryImpl) GetUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]*types.MaintenanceSchedule, error) {
	query := `
		SELECT ms.id, ms.equipment_id, ms.maintenance_type, ms.next_due, ms.description, ms.priority, ms.active
		FROM maintenance_schedules ms
		JOIN equipment e ON ms.equipment_id = e.id
		WHERE e.tenant_id = $1 
		  AND ms.active = true
		  AND ms.next_due >= NOW()
		ORDER BY ms.next_due ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming maintenance: %w", err)
	}
	defer rows.Close()

	var schedules []*types.MaintenanceSchedule
	for rows.Next() {
		var schedule types.MaintenanceSchedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.EquipmentID,
			&schedule.MaintenanceType,
			&schedule.NextDue,
			&schedule.Description,
			&schedule.Priority,
			&schedule.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// GetByIDs retrieves multiple equipment items by their IDs
func (r *EquipmentRepositoryImpl) GetByIDs(ctx context.Context, tenantID uuid.UUID, equipmentIDs []uuid.UUID) ([]*domain.Equipment, error) {
	if len(equipmentIDs) == 0 {
		return []*domain.Equipment{}, nil
	}

	query := `
		SELECT id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price,
			   status, maintenance_schedule, last_maintenance, next_maintenance, notes,
			   created_at, updated_at
		FROM equipment
		WHERE tenant_id = $1 AND id = ANY($2) AND status != 'deleted'
		ORDER BY name ASC`

	// Convert UUIDs to string array for PostgreSQL
	idStrings := make([]string, len(equipmentIDs))
	for i, id := range equipmentIDs {
		idStrings[i] = id.String()
	}

	rows, err := r.db.QueryContext(ctx, query, tenantID, "{"+strings.Join(idStrings, ",")+"}") 
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment by IDs: %w", err)
	}
	defer rows.Close()

	var equipment []*domain.Equipment
	for rows.Next() {
		var eq domain.Equipment
		err := rows.Scan(
			&eq.ID,
			&eq.TenantID,
			&eq.Name,
			&eq.Type,
			&eq.Model,
			&eq.SerialNumber,
			&eq.PurchaseDate,
			&eq.PurchasePrice,
			&eq.Status,
			&eq.MaintenanceSchedule,
			&eq.LastMaintenance,
			&eq.NextMaintenance,
			&eq.Notes,
			&eq.CreatedAt,
			&eq.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment: %w", err)
		}
		equipment = append(equipment, &eq)
	}

	return equipment, nil
}

// GetAvailable retrieves available equipment for a time period
func (r *EquipmentRepositoryImpl) GetAvailable(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*domain.Equipment, error) {
	// This is a simplified version - in reality, you'd need to check job assignments and maintenance schedules
	query := `
		SELECT e.id, e.tenant_id, e.name, e.type, e.model, e.serial_number, e.purchase_date, e.purchase_price,
			   e.status, e.maintenance_schedule, e.last_maintenance, e.next_maintenance, e.notes,
			   e.created_at, e.updated_at
		FROM equipment e
		WHERE e.tenant_id = $1 
		  AND e.status = 'available'
		  AND (e.next_maintenance IS NULL OR e.next_maintenance > $3)
		  AND NOT EXISTS (
			  SELECT 1 FROM job_equipment je
			  JOIN jobs j ON je.job_id = j.id
			  WHERE je.equipment_id = e.id
			    AND j.status IN ('in_progress', 'scheduled')
			    AND (
			        (j.scheduled_start_time <= $3 AND j.scheduled_end_time >= $2) OR
			        (j.actual_start_time <= $3 AND (j.actual_end_time IS NULL OR j.actual_end_time >= $2))
			    )
		  )
		ORDER BY e.name ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get available equipment: %w", err)
	}
	defer rows.Close()

	var equipment []*domain.Equipment
	for rows.Next() {
		var eq domain.Equipment
		err := rows.Scan(
			&eq.ID,
			&eq.TenantID,
			&eq.Name,
			&eq.Type,
			&eq.Model,
			&eq.SerialNumber,
			&eq.PurchaseDate,
			&eq.PurchasePrice,
			&eq.Status,
			&eq.MaintenanceSchedule,
			&eq.LastMaintenance,
			&eq.NextMaintenance,
			&eq.Notes,
			&eq.CreatedAt,
			&eq.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan available equipment: %w", err)
		}
		equipment = append(equipment, &eq)
	}

	return equipment, nil
}

// CheckAvailability checks if specific equipment is available during a time period
func (r *EquipmentRepositoryImpl) CheckAvailability(ctx context.Context, equipmentIDs []uuid.UUID, startTime, endTime time.Time) (map[uuid.UUID]bool, error) {
	if len(equipmentIDs) == 0 {
		return map[uuid.UUID]bool{}, nil
	}

	// Convert UUIDs to string array for PostgreSQL
	idStrings := make([]string, len(equipmentIDs))
	for i, id := range equipmentIDs {
		idStrings[i] = id.String()
	}

	query := `
		SELECT e.id,
		       CASE 
		           WHEN e.status != 'available' THEN false
		           WHEN e.next_maintenance IS NOT NULL AND e.next_maintenance < $2 THEN false
		           WHEN EXISTS (
		               SELECT 1 FROM job_equipment je
		               JOIN jobs j ON je.job_id = j.id
		               WHERE je.equipment_id = e.id
		                 AND j.status IN ('in_progress', 'scheduled')
		                 AND (
		                     (j.scheduled_start_time <= $3 AND j.scheduled_end_time >= $2) OR
		                     (j.actual_start_time <= $3 AND (j.actual_end_time IS NULL OR j.actual_end_time >= $2))
		                 )
		           ) THEN false
		           ELSE true
		       END as available
		FROM equipment e
		WHERE e.id = ANY($1) AND e.status != 'deleted'`

	rows, err := r.db.QueryContext(ctx, query, "{"+strings.Join(idStrings, ",")+"}", startTime, endTime)
	if err != nil {
		return nil, fmt.Errorf("failed to check equipment availability: %w", err)
	}
	defer rows.Close()

	availability := make(map[uuid.UUID]bool)
	for rows.Next() {
		var equipmentID uuid.UUID
		var available bool
		err := rows.Scan(&equipmentID, &available)
		if err != nil {
			return nil, fmt.Errorf("failed to scan availability: %w", err)
		}
		availability[equipmentID] = available
	}

	// Set false for any equipment not found
	for _, id := range equipmentIDs {
		if _, exists := availability[id]; !exists {
			availability[id] = false
		}
	}

	return availability, nil
}

// GetMaintenanceDue retrieves equipment with maintenance due
func (r *EquipmentRepositoryImpl) GetMaintenanceDue(ctx context.Context, tenantID uuid.UUID) ([]*domain.Equipment, error) {
	query := `
		SELECT id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price,
			   status, maintenance_schedule, last_maintenance, next_maintenance, notes,
			   created_at, updated_at
		FROM equipment
		WHERE tenant_id = $1 
		  AND status != 'deleted'
		  AND (next_maintenance IS NOT NULL AND next_maintenance <= NOW())
		ORDER BY next_maintenance ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get equipment with maintenance due: %w", err)
	}
	defer rows.Close()

	var equipment []*domain.Equipment
	for rows.Next() {
		var eq domain.Equipment
		err := rows.Scan(
			&eq.ID,
			&eq.TenantID,
			&eq.Name,
			&eq.Type,
			&eq.Model,
			&eq.SerialNumber,
			&eq.PurchaseDate,
			&eq.PurchasePrice,
			&eq.Status,
			&eq.MaintenanceSchedule,
			&eq.LastMaintenance,
			&eq.NextMaintenance,
			&eq.Notes,
			&eq.CreatedAt,
			&eq.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan equipment with maintenance due: %w", err)
		}
		equipment = append(equipment, &eq)
	}

	return equipment, nil
}

// UpdateMaintenanceDate updates the maintenance dates for equipment
func (r *EquipmentRepositoryImpl) UpdateMaintenanceDate(ctx context.Context, equipmentID uuid.UUID, lastMaintenance, nextMaintenance time.Time) error {
	query := `
		UPDATE equipment SET
			last_maintenance = $2,
			next_maintenance = $3,
			updated_at = $4
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, equipmentID, lastMaintenance, nextMaintenance, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update maintenance date: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("equipment not found")
	}

	return nil
}

// MaintenanceRepositoryImpl implements maintenance data access
type MaintenanceRepositoryImpl struct {
	db *Database
}

// NewMaintenanceRepository creates a new maintenance repository
func NewMaintenanceRepository(db *Database) types.MaintenanceRepository {
	return &MaintenanceRepositoryImpl{db: db}
}

// CreateMaintenanceRecord creates a new maintenance record
func (r *MaintenanceRepositoryImpl) CreateMaintenanceRecord(ctx context.Context, record *types.MaintenanceRecord) error {
	query := `
		INSERT INTO maintenance_records (
			id, equipment_id, type, performed_date, description, cost, performed_by, notes
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.db.ExecContext(ctx, query,
		record.ID,
		record.EquipmentID,
		record.MaintenanceType,
		record.MaintenanceDate,
		record.Description,
		record.TotalCost,
		record.PerformedBy,
		record.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to create maintenance record: %w", err)
	}

	return nil
}

// GetMaintenanceRecord retrieves a maintenance record by ID
func (r *MaintenanceRepositoryImpl) GetMaintenanceRecord(ctx context.Context, recordID uuid.UUID) (*types.MaintenanceRecord, error) {
	query := `
		SELECT id, equipment_id, type, performed_date, description, cost, performed_by, notes
		FROM maintenance_records
		WHERE id = $1`

	var record types.MaintenanceRecord
	err := r.db.QueryRowContext(ctx, query, recordID).Scan(
		&record.ID,
		&record.EquipmentID,
		&record.MaintenanceType,
		&record.MaintenanceDate,
		&record.Description,
		&record.TotalCost,
		&record.PerformedBy,
		&record.Notes,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get maintenance record: %w", err)
	}

	return &record, nil
}

// UpdateMaintenanceRecord updates a maintenance record
func (r *MaintenanceRepositoryImpl) UpdateMaintenanceRecord(ctx context.Context, record *types.MaintenanceRecord) error {
	query := `
		UPDATE maintenance_records SET
			type = $2, performed_date = $3, description = $4, cost = $5, performed_by = $6, notes = $7
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		record.ID,
		record.MaintenanceType,
		record.MaintenanceDate,
		record.Description,
		record.TotalCost,
		record.PerformedBy,
		record.Notes,
	)

	if err != nil {
		return fmt.Errorf("failed to update maintenance record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("maintenance record not found")
	}

	return nil
}

// DeleteMaintenanceRecord deletes a maintenance record
func (r *MaintenanceRepositoryImpl) DeleteMaintenanceRecord(ctx context.Context, recordID uuid.UUID) error {
	query := `DELETE FROM maintenance_records WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, recordID)
	if err != nil {
		return fmt.Errorf("failed to delete maintenance record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("maintenance record not found")
	}

	return nil
}

// GetMaintenanceHistory retrieves maintenance history for equipment
func (r *MaintenanceRepositoryImpl) GetMaintenanceHistory(ctx context.Context, equipmentID uuid.UUID) ([]*types.MaintenanceRecord, error) {
	query := `
		SELECT id, equipment_id, type, performed_date, description, cost, performed_by, notes
		FROM maintenance_records
		WHERE equipment_id = $1
		ORDER BY performed_date DESC`

	rows, err := r.db.QueryContext(ctx, query, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance history: %w", err)
	}
	defer rows.Close()

	var records []*types.MaintenanceRecord
	for rows.Next() {
		var record types.MaintenanceRecord
		err := rows.Scan(
			&record.ID,
			&record.EquipmentID,
			&record.MaintenanceType,
			&record.MaintenanceDate,
			&record.Description,
			&record.TotalCost,
			&record.PerformedBy,
			&record.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance record: %w", err)
		}
		records = append(records, &record)
	}

	return records, nil
}

// GetMaintenanceHistoryByDate retrieves maintenance records by date range
func (r *MaintenanceRepositoryImpl) GetMaintenanceHistoryByDate(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*types.MaintenanceRecord, error) {
	query := `
		SELECT mr.id, mr.equipment_id, mr.type, mr.performed_date, mr.description, mr.cost, mr.performed_by, mr.notes
		FROM maintenance_records mr
		JOIN equipment e ON mr.equipment_id = e.id
		WHERE e.tenant_id = $1 
		  AND mr.performed_date >= $2 
		  AND mr.performed_date <= $3
		ORDER BY mr.performed_date DESC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance history by date: %w", err)
	}
	defer rows.Close()

	var records []*types.MaintenanceRecord
	for rows.Next() {
		var record types.MaintenanceRecord
		err := rows.Scan(
			&record.ID,
			&record.EquipmentID,
			&record.MaintenanceType,
			&record.MaintenanceDate,
			&record.Description,
			&record.TotalCost,
			&record.PerformedBy,
			&record.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance record: %w", err)
		}
		records = append(records, &record)
	}

	return records, nil
}

// Maintenance Schedule methods

// CreateMaintenanceSchedule creates a new maintenance schedule
func (r *MaintenanceRepositoryImpl) CreateMaintenanceSchedule(ctx context.Context, schedule *types.MaintenanceSchedule) error {
	query := `
		INSERT INTO maintenance_schedules (
			id, equipment_id, maintenance_type, next_due, description, priority, active
		) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		schedule.ID,
		schedule.EquipmentID,
		schedule.MaintenanceType,
		schedule.NextDue,
		schedule.Description,
		schedule.Priority,
		schedule.Active,
	)

	if err != nil {
		return fmt.Errorf("failed to create maintenance schedule: %w", err)
	}

	return nil
}

// GetMaintenanceSchedule retrieves a maintenance schedule by ID
func (r *MaintenanceRepositoryImpl) GetMaintenanceSchedule(ctx context.Context, scheduleID uuid.UUID) (*types.MaintenanceSchedule, error) {
	query := `
		SELECT id, equipment_id, maintenance_type, next_due, description, priority, active
		FROM maintenance_schedules
		WHERE id = $1`

	var schedule types.MaintenanceSchedule
	err := r.db.QueryRowContext(ctx, query, scheduleID).Scan(
		&schedule.ID,
		&schedule.EquipmentID,
		&schedule.MaintenanceType,
		&schedule.NextDue,
		&schedule.Description,
		&schedule.Priority,
		&schedule.Active,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get maintenance schedule: %w", err)
	}

	return &schedule, nil
}

// UpdateMaintenanceSchedule updates a maintenance schedule
func (r *MaintenanceRepositoryImpl) UpdateMaintenanceSchedule(ctx context.Context, schedule *types.MaintenanceSchedule) error {
	query := `
		UPDATE maintenance_schedules SET
			maintenance_type = $2, next_due = $3, description = $4, priority = $5, active = $6
		WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query,
		schedule.ID,
		schedule.MaintenanceType,
		schedule.NextDue,
		schedule.Description,
		schedule.Priority,
		schedule.Active,
	)

	if err != nil {
		return fmt.Errorf("failed to update maintenance schedule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("maintenance schedule not found")
	}

	return nil
}

// DeleteMaintenanceSchedule deletes a maintenance schedule
func (r *MaintenanceRepositoryImpl) DeleteMaintenanceSchedule(ctx context.Context, scheduleID uuid.UUID) error {
	query := `DELETE FROM maintenance_schedules WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, scheduleID)
	if err != nil {
		return fmt.Errorf("failed to delete maintenance schedule: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("maintenance schedule not found")
	}

	return nil
}

// GetUpcomingMaintenance retrieves upcoming maintenance schedules
func (r *MaintenanceRepositoryImpl) GetUpcomingMaintenance(ctx context.Context, tenantID uuid.UUID) ([]*types.MaintenanceSchedule, error) {
	query := `
		SELECT ms.id, ms.equipment_id, ms.maintenance_type, ms.next_due, ms.description, ms.priority, ms.active
		FROM maintenance_schedules ms
		JOIN equipment e ON ms.equipment_id = e.id
		WHERE e.tenant_id = $1 
		  AND ms.active = true
		  AND ms.next_due >= NOW()
		ORDER BY ms.next_due ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming maintenance: %w", err)
	}
	defer rows.Close()

	var schedules []*types.MaintenanceSchedule
	for rows.Next() {
		var schedule types.MaintenanceSchedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.EquipmentID,
			&schedule.MaintenanceType,
			&schedule.NextDue,
			&schedule.Description,
			&schedule.Priority,
			&schedule.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// GetOverdueMaintenance retrieves overdue maintenance schedules for a tenant
func (r *MaintenanceRepositoryImpl) GetOverdueMaintenance(ctx context.Context, tenantID uuid.UUID) ([]*types.MaintenanceSchedule, error) {
	query := `
		SELECT ms.id, ms.equipment_id, ms.maintenance_type, ms.next_due, ms.description, ms.priority, ms.active
		FROM maintenance_schedules ms
		JOIN equipment e ON ms.equipment_id = e.id
		WHERE e.tenant_id = $1 
		  AND ms.active = true
		  AND ms.next_due < NOW()
		ORDER BY ms.next_due ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get overdue maintenance: %w", err)
	}
	defer rows.Close()

	var schedules []*types.MaintenanceSchedule
	for rows.Next() {
		var schedule types.MaintenanceSchedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.EquipmentID,
			&schedule.MaintenanceType,
			&schedule.NextDue,
			&schedule.Description,
			&schedule.Priority,
			&schedule.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

// BulkUpdateStatus updates the status of multiple equipment items
func (r *EquipmentRepositoryImpl) BulkUpdateStatus(ctx context.Context, tenantID uuid.UUID, equipmentIDs []uuid.UUID, status string) error {
	if len(equipmentIDs) == 0 {
		return nil
	}

	// Build placeholders for the IN clause
	placeholders := make([]string, len(equipmentIDs))
	args := make([]interface{}, len(equipmentIDs)+2)
	args[0] = status
	args[1] = tenantID
	
	for i, id := range equipmentIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+3)
		args[i+2] = id
	}

	query := fmt.Sprintf(`
		UPDATE equipment 
		SET status = $1, updated_at = NOW()
		WHERE tenant_id = $2 
		AND id IN (%s)
	`, strings.Join(placeholders, ","))

	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to bulk update equipment status: %w", err)
	}

	return nil
}

// GetMaintenanceSchedulesByEquipment retrieves maintenance schedules for specific equipment
func (r *MaintenanceRepositoryImpl) GetMaintenanceSchedulesByEquipment(ctx context.Context, equipmentID uuid.UUID) ([]*types.MaintenanceSchedule, error) {
	query := `
		SELECT id, equipment_id, maintenance_type, next_due, description, priority, active
		FROM maintenance_schedules
		WHERE equipment_id = $1
		ORDER BY next_due ASC`

	rows, err := r.db.QueryContext(ctx, query, equipmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get maintenance schedules by equipment: %w", err)
	}
	defer rows.Close()

	var schedules []*types.MaintenanceSchedule
	for rows.Next() {
		var schedule types.MaintenanceSchedule
		err := rows.Scan(
			&schedule.ID,
			&schedule.EquipmentID,
			&schedule.MaintenanceType,
			&schedule.NextDue,
			&schedule.Description,
			&schedule.Priority,
			&schedule.Active,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan maintenance schedule: %w", err)
		}
		schedules = append(schedules, &schedule)
	}

	return schedules, nil
}

