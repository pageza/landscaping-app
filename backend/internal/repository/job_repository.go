package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// JobRepositoryImpl implements the JobRepositoryComplete interface
type JobRepositoryImpl struct {
	db *Database
}

// NewJobRepositoryImpl creates a new job repository instance
func NewJobRepositoryImpl(db *Database) services.JobRepositoryComplete {
	return &JobRepositoryImpl{db: db}
}

// Create creates a new job
func (r *JobRepositoryImpl) Create(ctx context.Context, job *domain.EnhancedJob) error {
	query := `
		INSERT INTO jobs (
			id, tenant_id, customer_id, property_id, assigned_user_id, title, description,
			status, priority, scheduled_date, scheduled_time, estimated_duration,
			actual_start_time, actual_end_time, total_amount, notes, job_number,
			recurring_schedule, parent_job_id, weather_dependent, requires_equipment,
			crew_size, completion_photos, customer_signature, gps_check_in, gps_check_out,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
			$17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28
		)`

	_, err := r.db.ExecContext(ctx, query,
		job.ID,
		job.TenantID,
		job.CustomerID,
		job.PropertyID,
		job.AssignedUserID,
		job.Title,
		job.Description,
		job.Status,
		job.Priority,
		job.ScheduledDate,
		job.ScheduledTime,
		job.EstimatedDuration,
		job.ActualStartTime,
		job.ActualEndTime,
		job.TotalAmount,
		job.Notes,
		job.JobNumber,
		job.RecurringSchedule,
		job.ParentJobID,
		job.WeatherDependent,
		pq.Array(job.RequiresEquipment),
		job.CrewSize,
		pq.Array(job.CompletionPhotos),
		job.CustomerSignature,
		job.GPSCheckIn,
		job.GPSCheckOut,
		job.CreatedAt,
		job.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}

	return nil
}

// GetByID retrieves a job by ID
func (r *JobRepositoryImpl) GetByID(ctx context.Context, tenantID, jobID uuid.UUID) (*domain.EnhancedJob, error) {
	query := `
		SELECT 
			id, tenant_id, customer_id, property_id, assigned_user_id, title, description,
			status, priority, scheduled_date, scheduled_time, estimated_duration,
			actual_start_time, actual_end_time, total_amount, notes, job_number,
			recurring_schedule, parent_job_id, weather_dependent, requires_equipment,
			crew_size, completion_photos, customer_signature, gps_check_in, gps_check_out,
			created_at, updated_at
		FROM jobs
		WHERE id = $1 AND tenant_id = $2`

	row := r.db.QueryRowContext(ctx, query, jobID, tenantID)

	job := &domain.EnhancedJob{}
	var requiresEquipment pq.UuidArray
	var completionPhotos pq.StringArray

	err := row.Scan(
		&job.ID,
		&job.TenantID,
		&job.CustomerID,
		&job.PropertyID,
		&job.AssignedUserID,
		&job.Title,
		&job.Description,
		&job.Status,
		&job.Priority,
		&job.ScheduledDate,
		&job.ScheduledTime,
		&job.EstimatedDuration,
		&job.ActualStartTime,
		&job.ActualEndTime,
		&job.TotalAmount,
		&job.Notes,
		&job.JobNumber,
		&job.RecurringSchedule,
		&job.ParentJobID,
		&job.WeatherDependent,
		&requiresEquipment,
		&job.CrewSize,
		&completionPhotos,
		&job.CustomerSignature,
		&job.GPSCheckIn,
		&job.GPSCheckOut,
		&job.CreatedAt,
		&job.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	// Convert pq arrays to Go slices
	job.RequiresEquipment = []uuid.UUID(requiresEquipment)
	job.CompletionPhotos = []string(completionPhotos)

	return job, nil
}

// Update updates an existing job
func (r *JobRepositoryImpl) Update(ctx context.Context, job *domain.EnhancedJob) error {
	query := `
		UPDATE jobs SET
			customer_id = $3,
			property_id = $4,
			assigned_user_id = $5,
			title = $6,
			description = $7,
			status = $8,
			priority = $9,
			scheduled_date = $10,
			scheduled_time = $11,
			estimated_duration = $12,
			actual_start_time = $13,
			actual_end_time = $14,
			total_amount = $15,
			notes = $16,
			job_number = $17,
			recurring_schedule = $18,
			parent_job_id = $19,
			weather_dependent = $20,
			requires_equipment = $21,
			crew_size = $22,
			completion_photos = $23,
			customer_signature = $24,
			gps_check_in = $25,
			gps_check_out = $26,
			updated_at = $27
		WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		job.ID,
		job.TenantID,
		job.CustomerID,
		job.PropertyID,
		job.AssignedUserID,
		job.Title,
		job.Description,
		job.Status,
		job.Priority,
		job.ScheduledDate,
		job.ScheduledTime,
		job.EstimatedDuration,
		job.ActualStartTime,
		job.ActualEndTime,
		job.TotalAmount,
		job.Notes,
		job.JobNumber,
		job.RecurringSchedule,
		job.ParentJobID,
		job.WeatherDependent,
		pq.Array(job.RequiresEquipment),
		job.CrewSize,
		pq.Array(job.CompletionPhotos),
		job.CustomerSignature,
		job.GPSCheckIn,
		job.GPSCheckOut,
		job.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job not found or not authorized")
	}

	return nil
}

// Delete deletes a job (hard delete for this case, or you could implement soft delete)
func (r *JobRepositoryImpl) Delete(ctx context.Context, tenantID, jobID uuid.UUID) error {
	query := `DELETE FROM jobs WHERE id = $1 AND tenant_id = $2`

	result, err := r.db.ExecContext(ctx, query, jobID, tenantID)
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job not found or not authorized")
	}

	return nil
}

// List retrieves jobs with filtering and pagination
func (r *JobRepositoryImpl) List(ctx context.Context, tenantID uuid.UUID, filter *services.JobFilter) ([]*domain.EnhancedJob, int64, error) {
	// Build WHERE clause
	whereClause := "WHERE tenant_id = $1"
	args := []interface{}{tenantID}
	argIndex := 2

	if filter.Status != "" {
		// Handle multiple statuses separated by comma
		if len(filter.Status) > 0 {
			whereClause += fmt.Sprintf(" AND status = ANY($%d)", argIndex)
			statuses := pq.Array([]string{filter.Status}) // Simplified for single status
			args = append(args, statuses)
			argIndex++
		}
	}

	if filter.Priority != "" {
		whereClause += fmt.Sprintf(" AND priority = $%d", argIndex)
		args = append(args, filter.Priority)
		argIndex++
	}

	if filter.AssignedUserID != nil {
		whereClause += fmt.Sprintf(" AND assigned_user_id = $%d", argIndex)
		args = append(args, *filter.AssignedUserID)
		argIndex++
	}

	if filter.CustomerID != nil {
		whereClause += fmt.Sprintf(" AND customer_id = $%d", argIndex)
		args = append(args, *filter.CustomerID)
		argIndex++
	}

	if filter.PropertyID != nil {
		whereClause += fmt.Sprintf(" AND property_id = $%d", argIndex)
		args = append(args, *filter.PropertyID)
		argIndex++
	}

	if filter.ScheduledStart != nil && filter.ScheduledEnd != nil {
		whereClause += fmt.Sprintf(" AND scheduled_date BETWEEN $%d AND $%d", argIndex, argIndex+1)
		args = append(args, *filter.ScheduledStart, *filter.ScheduledEnd)
		argIndex += 2
	}

	if filter.Search != "" {
		searchPattern := "%" + filter.Search + "%"
		whereClause += fmt.Sprintf(" AND (title ILIKE $%d OR description ILIKE $%d OR notes ILIKE $%d)", 
			argIndex, argIndex, argIndex)
		args = append(args, searchPattern)
		argIndex++
	}

	// Count total records
	countQuery := "SELECT COUNT(*) FROM jobs " + whereClause
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count jobs: %w", err)
	}

	// Build ORDER BY clause
	orderBy := "ORDER BY created_at DESC"
	if filter.SortBy != "" {
		direction := "ASC"
		if filter.SortDesc {
			direction = "DESC"
		}
		// Validate sort field to prevent SQL injection
		validSortFields := map[string]bool{
			"title":          true,
			"status":         true,
			"priority":       true,
			"scheduled_date": true,
			"created_at":     true,
			"updated_at":     true,
		}
		if validSortFields[filter.SortBy] {
			orderBy = fmt.Sprintf("ORDER BY %s %s", filter.SortBy, direction)
		}
	}

	// Add pagination
	limit := filter.PerPage
	offset := (filter.Page - 1) * filter.PerPage
	paginationClause := fmt.Sprintf(" %s LIMIT $%d OFFSET $%d", orderBy, argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Execute main query
	query := `
		SELECT 
			id, tenant_id, customer_id, property_id, assigned_user_id, title, description,
			status, priority, scheduled_date, scheduled_time, estimated_duration,
			actual_start_time, actual_end_time, total_amount, notes, job_number,
			recurring_schedule, parent_job_id, weather_dependent, requires_equipment,
			crew_size, completion_photos, customer_signature, gps_check_in, gps_check_out,
			created_at, updated_at
		FROM jobs ` + whereClause + paginationClause

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list jobs: %w", err)
	}
	defer rows.Close()

	jobs := make([]*domain.EnhancedJob, 0)
	for rows.Next() {
		job := &domain.EnhancedJob{}
		var requiresEquipment pq.UuidArray
		var completionPhotos pq.StringArray

		err := rows.Scan(
			&job.ID,
			&job.TenantID,
			&job.CustomerID,
			&job.PropertyID,
			&job.AssignedUserID,
			&job.Title,
			&job.Description,
			&job.Status,
			&job.Priority,
			&job.ScheduledDate,
			&job.ScheduledTime,
			&job.EstimatedDuration,
			&job.ActualStartTime,
			&job.ActualEndTime,
			&job.TotalAmount,
			&job.Notes,
			&job.JobNumber,
			&job.RecurringSchedule,
			&job.ParentJobID,
			&job.WeatherDependent,
			&requiresEquipment,
			&job.CrewSize,
			&completionPhotos,
			&job.CustomerSignature,
			&job.GPSCheckIn,
			&job.GPSCheckOut,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan job: %w", err)
		}

		// Convert pq arrays to Go slices
		job.RequiresEquipment = []uuid.UUID(requiresEquipment)
		job.CompletionPhotos = []string(completionPhotos)

		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating jobs: %w", err)
	}

	return jobs, total, nil
}

// GetByCustomerID retrieves jobs for a specific customer
func (r *JobRepositoryImpl) GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *services.JobFilter) ([]*domain.EnhancedJob, int64, error) {
	// Update filter to include customer ID
	if filter == nil {
		filter = &services.JobFilter{}
	}
	filter.CustomerID = &customerID

	return r.List(ctx, tenantID, filter)
}

// GetByPropertyID retrieves jobs for a specific property
func (r *JobRepositoryImpl) GetByPropertyID(ctx context.Context, tenantID, propertyID uuid.UUID, filter *services.JobFilter) ([]*domain.EnhancedJob, int64, error) {
	// Update filter to include property ID
	if filter == nil {
		filter = &services.JobFilter{}
	}
	filter.PropertyID = &propertyID

	return r.List(ctx, tenantID, filter)
}

// GetByAssignedUserID retrieves jobs for a specific user
func (r *JobRepositoryImpl) GetByAssignedUserID(ctx context.Context, tenantID, userID uuid.UUID, filter *services.JobFilter) ([]*domain.EnhancedJob, int64, error) {
	// Update filter to include assigned user ID
	if filter == nil {
		filter = &services.JobFilter{}
	}
	filter.AssignedUserID = &userID

	return r.List(ctx, tenantID, filter)
}

// GetByStatus retrieves jobs by status
func (r *JobRepositoryImpl) GetByStatus(ctx context.Context, tenantID uuid.UUID, status string, filter *services.JobFilter) ([]*domain.EnhancedJob, int64, error) {
	// Update filter to include status
	if filter == nil {
		filter = &services.JobFilter{}
	}
	filter.Status = status

	return r.List(ctx, tenantID, filter)
}

// GetByDateRange retrieves jobs within a date range
func (r *JobRepositoryImpl) GetByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*domain.EnhancedJob, error) {
	query := `
		SELECT 
			id, tenant_id, customer_id, property_id, assigned_user_id, title, description,
			status, priority, scheduled_date, scheduled_time, estimated_duration,
			actual_start_time, actual_end_time, total_amount, notes, job_number,
			recurring_schedule, parent_job_id, weather_dependent, requires_equipment,
			crew_size, completion_photos, customer_signature, gps_check_in, gps_check_out,
			created_at, updated_at
		FROM jobs
		WHERE tenant_id = $1 AND scheduled_date BETWEEN $2 AND $3
		ORDER BY scheduled_date ASC`

	rows, err := r.db.QueryContext(ctx, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by date range: %w", err)
	}
	defer rows.Close()

	jobs := make([]*domain.EnhancedJob, 0)
	for rows.Next() {
		job := &domain.EnhancedJob{}
		var requiresEquipment pq.UuidArray
		var completionPhotos pq.StringArray

		err := rows.Scan(
			&job.ID,
			&job.TenantID,
			&job.CustomerID,
			&job.PropertyID,
			&job.AssignedUserID,
			&job.Title,
			&job.Description,
			&job.Status,
			&job.Priority,
			&job.ScheduledDate,
			&job.ScheduledTime,
			&job.EstimatedDuration,
			&job.ActualStartTime,
			&job.ActualEndTime,
			&job.TotalAmount,
			&job.Notes,
			&job.JobNumber,
			&job.RecurringSchedule,
			&job.ParentJobID,
			&job.WeatherDependent,
			&requiresEquipment,
			&job.CrewSize,
			&completionPhotos,
			&job.CustomerSignature,
			&job.GPSCheckIn,
			&job.GPSCheckOut,
			&job.CreatedAt,
			&job.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job: %w", err)
		}

		// Convert pq arrays to Go slices
		job.RequiresEquipment = []uuid.UUID(requiresEquipment)
		job.CompletionPhotos = []string(completionPhotos)

		jobs = append(jobs, job)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating jobs by date range: %w", err)
	}

	return jobs, nil
}

// Job Services management

// CreateJobService creates a job service association
func (r *JobRepositoryImpl) CreateJobService(ctx context.Context, jobService *domain.JobService) error {
	query := `
		INSERT INTO job_services (id, job_id, service_id, quantity, unit_price, total_price, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		jobService.ID,
		jobService.JobID,
		jobService.ServiceID,
		jobService.Quantity,
		jobService.UnitPrice,
		jobService.TotalPrice,
		jobService.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create job service: %w", err)
	}

	return nil
}

// UpdateJobService updates a job service
func (r *JobRepositoryImpl) UpdateJobService(ctx context.Context, jobService *domain.JobService) error {
	query := `
		UPDATE job_services SET
			quantity = $3,
			unit_price = $4,
			total_price = $5
		WHERE id = $1 AND job_id = $2`

	result, err := r.db.ExecContext(ctx, query,
		jobService.ID,
		jobService.JobID,
		jobService.Quantity,
		jobService.UnitPrice,
		jobService.TotalPrice,
	)

	if err != nil {
		return fmt.Errorf("failed to update job service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job service not found")
	}

	return nil
}

// DeleteJobService deletes a job service
func (r *JobRepositoryImpl) DeleteJobService(ctx context.Context, jobServiceID uuid.UUID) error {
	query := `DELETE FROM job_services WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, jobServiceID)
	if err != nil {
		return fmt.Errorf("failed to delete job service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("job service not found")
	}

	return nil
}

// GetJobServices retrieves services for a job
func (r *JobRepositoryImpl) GetJobServices(ctx context.Context, jobID uuid.UUID) ([]*domain.JobService, error) {
	query := `
		SELECT id, job_id, service_id, quantity, unit_price, total_price, created_at
		FROM job_services
		WHERE job_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.QueryContext(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job services: %w", err)
	}
	defer rows.Close()

	services := make([]*domain.JobService, 0)
	for rows.Next() {
		service := &domain.JobService{}
		err := rows.Scan(
			&service.ID,
			&service.JobID,
			&service.ServiceID,
			&service.Quantity,
			&service.UnitPrice,
			&service.TotalPrice,
			&service.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan job service: %w", err)
		}
		services = append(services, service)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating job services: %w", err)
	}

	return services, nil
}

// GetNextJobNumber generates the next job number for a tenant
func (r *JobRepositoryImpl) GetNextJobNumber(ctx context.Context, tenantID uuid.UUID) (string, error) {
	// Get the count of jobs for this tenant to generate a sequence number
	query := `SELECT COUNT(*) FROM jobs WHERE tenant_id = $1`
	
	var count int64
	err := r.db.QueryRowContext(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return "", fmt.Errorf("failed to get job count: %w", err)
	}

	// Generate job number in format: JOB-YYYY-NNNN
	year := time.Now().Year()
	jobNumber := fmt.Sprintf("JOB-%d-%04d", year, count+1)

	return jobNumber, nil
}

// Recurring job management

// CreateRecurringJobSeries creates a recurring job series
func (r *JobRepositoryImpl) CreateRecurringJobSeries(ctx context.Context, series *services.RecurringJobSeries) error {
	query := `
		INSERT INTO recurring_job_series (id, base_job_id, frequency, next_occurrence, jobs_created, upcoming_jobs)
		VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		series.ID,
		series.BaseJobID,
		series.Frequency,
		series.NextOccurrence,
		series.JobsCreated,
		pq.Array(series.UpcomingJobs),
	)

	if err != nil {
		return fmt.Errorf("failed to create recurring job series: %w", err)
	}

	return nil
}

// GetRecurringJobSeries retrieves a recurring job series
func (r *JobRepositoryImpl) GetRecurringJobSeries(ctx context.Context, tenantID uuid.UUID, baseJobID uuid.UUID) (*services.RecurringJobSeries, error) {
	query := `
		SELECT rjs.id, rjs.base_job_id, rjs.frequency, rjs.next_occurrence, rjs.jobs_created, rjs.upcoming_jobs
		FROM recurring_job_series rjs
		JOIN jobs j ON rjs.base_job_id = j.id
		WHERE j.tenant_id = $1 AND rjs.base_job_id = $2`

	row := r.db.QueryRowContext(ctx, query, tenantID, baseJobID)

	series := &services.RecurringJobSeries{}
	var upcomingJobs pq.UuidArray

	err := row.Scan(
		&series.ID,
		&series.BaseJobID,
		&series.Frequency,
		&series.NextOccurrence,
		&series.JobsCreated,
		&upcomingJobs,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get recurring job series: %w", err)
	}

	series.UpcomingJobs = []uuid.UUID(upcomingJobs)

	return series, nil
}

// Analytics and reporting methods

// GetJobStatistics retrieves job statistics for a tenant
func (r *JobRepositoryImpl) GetJobStatistics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*JobStatistics, error) {
	query := `
		SELECT 
			COUNT(*) as total_jobs,
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_jobs,
			COUNT(CASE WHEN status = 'scheduled' THEN 1 END) as scheduled_jobs,
			COUNT(CASE WHEN status = 'in_progress' THEN 1 END) as in_progress_jobs,
			COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_jobs,
			COUNT(CASE WHEN status = 'cancelled' THEN 1 END) as cancelled_jobs,
			AVG(CASE WHEN status = 'completed' AND total_amount IS NOT NULL THEN total_amount END) as avg_job_value,
			SUM(CASE WHEN status = 'completed' AND total_amount IS NOT NULL THEN total_amount ELSE 0 END) as total_revenue
		FROM jobs
		WHERE tenant_id = $1 AND created_at BETWEEN $2 AND $3`

	var stats JobStatistics
	err := r.db.QueryRowContext(ctx, query, tenantID, startDate, endDate).Scan(
		&stats.TotalJobs,
		&stats.PendingJobs,
		&stats.ScheduledJobs,
		&stats.InProgressJobs,
		&stats.CompletedJobs,
		&stats.CancelledJobs,
		&stats.AvgJobValue,
		&stats.TotalRevenue,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get job statistics: %w", err)
	}

	return &stats, nil
}

// GetByEquipmentID gets jobs that used specific equipment within date range
func (r *JobRepositoryImpl) GetByEquipmentID(ctx context.Context, tenantID uuid.UUID, equipmentID uuid.UUID, startDate, endDate time.Time) ([]*domain.EnhancedJob, error) {
	// TODO: Implement proper equipment-job relationship query
	// For now, return empty list to prevent compilation errors
	// This would typically join with a job_equipment table or similar
	return []*domain.EnhancedJob{}, nil
}


// Helper types for statistics
type JobStatistics struct {
	TotalJobs      int      `json:"total_jobs"`
	PendingJobs    int      `json:"pending_jobs"`
	ScheduledJobs  int      `json:"scheduled_jobs"`
	InProgressJobs int      `json:"in_progress_jobs"`
	CompletedJobs  int      `json:"completed_jobs"`
	CancelledJobs  int      `json:"cancelled_jobs"`
	AvgJobValue    *float64 `json:"avg_job_value"`
	TotalRevenue   float64  `json:"total_revenue"`
}