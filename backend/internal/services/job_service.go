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

// JobServiceImpl implements the JobService interface
type JobServiceImpl struct {
	jobRepo            JobRepositoryComplete
	customerRepo       CustomerRepository
	propertyRepo       PropertyRepositoryExtended
	serviceRepo        ServiceRepository
	userRepo           UserRepository
	crewRepo           CrewRepository
	equipmentRepo      EquipmentRepository
	auditService       AuditService
	notificationService NotificationService
	storageService     StorageService
	scheduleService    ScheduleService
	logger             *log.Logger
}

// JobRepositoryComplete defines the complete interface for job data access
type JobRepositoryComplete interface {
	// CRUD operations
	Create(ctx context.Context, job *domain.EnhancedJob) error
	GetByID(ctx context.Context, tenantID, jobID uuid.UUID) (*domain.EnhancedJob, error)
	Update(ctx context.Context, job *domain.EnhancedJob) error
	Delete(ctx context.Context, tenantID, jobID uuid.UUID) error
	List(ctx context.Context, tenantID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	
	// Filtering operations
	GetByCustomerID(ctx context.Context, tenantID, customerID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	GetByPropertyID(ctx context.Context, tenantID, propertyID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	GetByAssignedUserID(ctx context.Context, tenantID, userID uuid.UUID, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	GetByStatus(ctx context.Context, tenantID uuid.UUID, status string, filter *JobFilter) ([]*domain.EnhancedJob, int64, error)
	GetByDateRange(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) ([]*domain.EnhancedJob, error)
	GetByEquipmentID(ctx context.Context, tenantID uuid.UUID, equipmentID uuid.UUID, startDate, endDate time.Time) ([]*domain.EnhancedJob, error)
	
	// Job services
	CreateJobService(ctx context.Context, jobService *domain.JobService) error
	UpdateJobService(ctx context.Context, jobService *domain.JobService) error
	DeleteJobService(ctx context.Context, jobServiceID uuid.UUID) error
	GetJobServices(ctx context.Context, jobID uuid.UUID) ([]*domain.JobService, error)
	
	// Job numbering
	GetNextJobNumber(ctx context.Context, tenantID uuid.UUID) (string, error)
	
	// Recurring jobs
	CreateRecurringJobSeries(ctx context.Context, series *RecurringJobSeries) error
	GetRecurringJobSeries(ctx context.Context, tenantID uuid.UUID, baseJobID uuid.UUID) (*RecurringJobSeries, error)
}

// Additional repository interfaces needed
type ServiceRepository interface {
	GetByID(ctx context.Context, tenantID, serviceID uuid.UUID) (*domain.Service, error)
	GetByIDs(ctx context.Context, tenantID uuid.UUID, serviceIDs []uuid.UUID) ([]*domain.Service, error)
}

type UserRepository interface {
	GetByID(ctx context.Context, tenantID, userID uuid.UUID) (*domain.EnhancedUser, error)
}

type CrewRepository interface {
	GetByID(ctx context.Context, tenantID, crewID uuid.UUID) (*domain.Crew, error)
	CheckAvailability(ctx context.Context, crewID uuid.UUID, startTime, endTime time.Time) (bool, error)
}

type EquipmentRepository interface {
	GetByIDs(ctx context.Context, tenantID uuid.UUID, equipmentIDs []uuid.UUID) ([]*domain.Equipment, error)
	CheckAvailability(ctx context.Context, equipmentIDs []uuid.UUID, startTime, endTime time.Time) (map[uuid.UUID]bool, error)
}

// NewJobService creates a new job service instance
func NewJobService(
	jobRepo JobRepositoryComplete,
	customerRepo CustomerRepository,
	propertyRepo PropertyRepositoryExtended,
	serviceRepo ServiceRepository,
	userRepo UserRepository,
	crewRepo CrewRepository,
	equipmentRepo EquipmentRepository,
	auditService AuditService,
	notificationService NotificationService,
	storageService StorageService,
	scheduleService ScheduleService,
	logger *log.Logger,
) JobService {
	return &JobServiceImpl{
		jobRepo:             jobRepo,
		customerRepo:        customerRepo,
		propertyRepo:        propertyRepo,
		serviceRepo:         serviceRepo,
		userRepo:            userRepo,
		crewRepo:            crewRepo,
		equipmentRepo:       equipmentRepo,
		auditService:        auditService,
		notificationService: notificationService,
		storageService:      storageService,
		scheduleService:     scheduleService,
		logger:              logger,
	}
}

// CreateJob creates a new job
func (s *JobServiceImpl) CreateJob(ctx context.Context, req *domain.CreateJobRequest) (*domain.EnhancedJob, error) {
	// Get tenant ID from context
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Validate the request
	if err := s.validateCreateJobRequest(req); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify customer exists
	customer, err := s.customerRepo.GetByID(ctx, tenantID, req.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify customer: %w", err)
	}
	if customer == nil {
		return nil, fmt.Errorf("customer not found")
	}

	// Verify property exists and belongs to customer
	property, err := s.propertyRepo.GetByID(ctx, tenantID, req.PropertyID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify property: %w", err)
	}
	if property == nil {
		return nil, fmt.Errorf("property not found")
	}
	if property.CustomerID != req.CustomerID {
		return nil, fmt.Errorf("property does not belong to the specified customer")
	}

	// Verify assigned user if specified
	if req.AssignedUserID != nil {
		user, err := s.userRepo.GetByID(ctx, tenantID, *req.AssignedUserID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify assigned user: %w", err)
		}
		if user == nil {
			return nil, fmt.Errorf("assigned user not found")
		}
	}

	// Verify services exist if specified
	if len(req.ServiceIDs) > 0 {
		services, err := s.serviceRepo.GetByIDs(ctx, tenantID, req.ServiceIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to verify services: %w", err)
		}
		if len(services) != len(req.ServiceIDs) {
			return nil, fmt.Errorf("one or more services not found")
		}
	}

	// Verify equipment availability if specified
	if len(req.RequiresEquipment) > 0 && req.ScheduledDate != nil {
		equipment, err := s.equipmentRepo.GetByIDs(ctx, tenantID, req.RequiresEquipment)
		if err != nil {
			return nil, fmt.Errorf("failed to verify equipment: %w", err)
		}
		if len(equipment) != len(req.RequiresEquipment) {
			return nil, fmt.Errorf("one or more equipment not found")
		}

		// Check equipment availability
		if req.EstimatedDuration != nil {
			endTime := req.ScheduledDate.Add(time.Duration(*req.EstimatedDuration) * time.Minute)
			availability, err := s.equipmentRepo.CheckAvailability(ctx, req.RequiresEquipment, *req.ScheduledDate, endTime)
			if err != nil {
				s.logger.Printf("Failed to check equipment availability", "error", err)
			} else {
				for equipmentID, available := range availability {
					if !available {
						return nil, fmt.Errorf("equipment %s is not available at the scheduled time", equipmentID)
					}
				}
			}
		}
	}

	// Generate job number
	jobNumber, err := s.jobRepo.GetNextJobNumber(ctx, tenantID)
	if err != nil {
		s.logger.Printf("Failed to generate job number", "error", err)
		jobNumber = fmt.Sprintf("JOB-%d", time.Now().Unix())
	}

	// Create job entity
	job := &domain.EnhancedJob{
		Job: domain.Job{
			ID:                uuid.New(),
			TenantID:          tenantID,
			CustomerID:        req.CustomerID,
			PropertyID:        req.PropertyID,
			AssignedUserID:    req.AssignedUserID,
			Title:             req.Title,
			Description:       req.Description,
			Status:            domain.JobStatusPending,
			Priority:          req.Priority,
			ScheduledDate:     req.ScheduledDate,
			ScheduledTime:     req.ScheduledTime,
			EstimatedDuration: req.EstimatedDuration,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		JobNumber:         &jobNumber,
		CrewSize:          req.CrewSize,
		WeatherDependent:  req.WeatherDependent,
		RequiresEquipment: req.RequiresEquipment,
	}

	// Save to database
	if err := s.jobRepo.Create(ctx, job); err != nil {
		s.logger.Printf("Failed to create job", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Create job services if specified
	if len(req.ServiceIDs) > 0 {
		services, err := s.serviceRepo.GetByIDs(ctx, tenantID, req.ServiceIDs)
		if err != nil {
			s.logger.Printf("Failed to get services for job creation", "error", err)
		} else {
			for _, service := range services {
				jobService := &domain.JobService{
					ID:         uuid.New(),
					JobID:      job.ID,
					ServiceID:  service.ID,
					Quantity:   1.0, // Default quantity
					UnitPrice:  getBasePriceOrZero(service.BasePrice),
					TotalPrice: getBasePriceOrZero(service.BasePrice),
					CreatedAt:  time.Now(),
				}

				if err := s.jobRepo.CreateJobService(ctx, jobService); err != nil {
					s.logger.Printf("Failed to create job service", "error", err, "job_id", job.ID, "service_id", service.ID)
				}
			}
		}
	}

	// Send notification to assigned user
	if job.AssignedUserID != nil {
		if err := s.notificationService.SendNotification(ctx, &NotificationRequest{
			UserID:  job.AssignedUserID,
			Type:    "job.assigned",
			Title:   "New Job Assigned",
			Message: fmt.Sprintf("You have been assigned to job: %s", job.Title),
			Data: map[string]interface{}{
				"job_id":    job.ID,
				"job_title": job.Title,
			},
		}); err != nil {
			s.logger.Printf("Failed to send job assignment notification", "error", err)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.create",
		ResourceType: "job",
		ResourceID:   &job.ID,
		NewValues: map[string]interface{}{
			"title":           job.Title,
			"customer_id":     job.CustomerID,
			"property_id":     job.PropertyID,
			"status":          job.Status,
			"priority":        job.Priority,
			"assigned_user_id": job.AssignedUserID,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job created successfully", "job_id", job.ID, "job_number", jobNumber, "tenant_id", tenantID)
	return job, nil
}

// GetJob retrieves a job by ID
func (s *JobServiceImpl) GetJob(ctx context.Context, jobID uuid.UUID) (*domain.EnhancedJob, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		s.logger.Printf("Failed to get job", "error", err, "job_id", jobID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	if job == nil {
		return nil, fmt.Errorf("job not found")
	}

	return job, nil
}

// UpdateJob updates an existing job
func (s *JobServiceImpl) UpdateJob(ctx context.Context, jobID uuid.UUID, req *domain.UpdateJobRequest) (*domain.EnhancedJob, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get existing job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return nil, fmt.Errorf("job not found")
	}

	// Store old values for audit
	oldValues := map[string]interface{}{
		"title":            job.Title,
		"status":           job.Status,
		"priority":         job.Priority,
		"assigned_user_id": job.AssignedUserID,
		"scheduled_date":   job.ScheduledDate,
	}

	// Update fields
	if req.Title != nil {
		job.Title = *req.Title
	}
	if req.Description != nil {
		job.Description = req.Description
	}
	if req.Status != nil {
		// Validate status transition
		if err := s.validateStatusTransition(job.Status, *req.Status); err != nil {
			return nil, fmt.Errorf("invalid status transition: %w", err)
		}
		job.Status = *req.Status
	}
	if req.Priority != nil {
		job.Priority = *req.Priority
	}
	if req.ScheduledDate != nil {
		job.ScheduledDate = req.ScheduledDate
	}
	if req.ScheduledTime != nil {
		job.ScheduledTime = req.ScheduledTime
	}
	if req.EstimatedDuration != nil {
		job.EstimatedDuration = req.EstimatedDuration
	}
	if req.AssignedUserID != nil {
		// Verify user exists
		if *req.AssignedUserID != uuid.Nil {
			user, err := s.userRepo.GetByID(ctx, tenantID, *req.AssignedUserID)
			if err != nil {
				return nil, fmt.Errorf("failed to verify assigned user: %w", err)
			}
			if user == nil {
				return nil, fmt.Errorf("assigned user not found")
			}
		}
		
		// Send notification if assignment changed
		if job.AssignedUserID == nil || *job.AssignedUserID != *req.AssignedUserID {
			if *req.AssignedUserID != uuid.Nil {
				if err := s.notificationService.SendNotification(ctx, &NotificationRequest{
					UserID:  req.AssignedUserID,
					Type:    "job.assigned",
					Title:   "Job Assigned",
					Message: fmt.Sprintf("You have been assigned to job: %s", job.Title),
					Data: map[string]interface{}{
						"job_id":    job.ID,
						"job_title": job.Title,
					},
				}); err != nil {
					s.logger.Printf("Failed to send job assignment notification", "error", err)
				}
			}
		}
		
		job.AssignedUserID = req.AssignedUserID
	}
	if req.CrewSize != nil {
		job.CrewSize = *req.CrewSize
	}
	if req.Notes != nil {
		job.Notes = req.Notes
	}

	job.UpdatedAt = time.Now()

	// Validate the updated job
	if err := s.validateJob(job); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Save to database
	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Printf("Failed to update job", "error", err, "job_id", jobID, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	// Log audit event
	newValues := map[string]interface{}{
		"title":            job.Title,
		"status":           job.Status,
		"priority":         job.Priority,
		"assigned_user_id": job.AssignedUserID,
		"scheduled_date":   job.ScheduledDate,
	}

	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.update",
		ResourceType: "job",
		ResourceID:   &job.ID,
		OldValues:    oldValues,
		NewValues:    newValues,
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job updated successfully", "job_id", jobID, "tenant_id", tenantID)
	return job, nil
}

// DeleteJob deletes a job
func (s *JobServiceImpl) DeleteJob(ctx context.Context, jobID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job before deletion for audit log
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	// Check if job can be deleted (only pending jobs can be deleted)
	if job.Status != domain.JobStatusPending {
		return fmt.Errorf("only pending jobs can be deleted")
	}

	// Soft delete by updating status
	job.Status = domain.JobStatusCancelled
	job.UpdatedAt = time.Now()
	job.Notes = stringPtr("Job deleted by user")

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Printf("Failed to delete job", "error", err, "job_id", jobID, "tenant_id", tenantID)
		return fmt.Errorf("failed to delete job: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.delete",
		ResourceType: "job",
		ResourceID:   &job.ID,
		OldValues: map[string]interface{}{
			"status": domain.JobStatusPending,
		},
		NewValues: map[string]interface{}{
			"status": domain.JobStatusCancelled,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job deleted successfully", "job_id", jobID, "tenant_id", tenantID)
	return nil
}

// ListJobs lists jobs with filtering and pagination
func (s *JobServiceImpl) ListJobs(ctx context.Context, filter *JobFilter) (*domain.PaginatedResponse, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Set defaults
	if filter == nil {
		filter = &JobFilter{}
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

	jobs, total, err := s.jobRepo.List(ctx, tenantID, filter)
	if err != nil {
		s.logger.Printf("Failed to list jobs", "error", err, "tenant_id", tenantID)
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	totalPages := int((total + int64(filter.PerPage) - 1) / int64(filter.PerPage))

	return &domain.PaginatedResponse{
		Data:       jobs,
		Total:      total,
		Page:       filter.Page,
		PerPage:    filter.PerPage,
		TotalPages: totalPages,
	}, nil
}

// StartJob starts a job
func (s *JobServiceImpl) StartJob(ctx context.Context, jobID uuid.UUID, startDetails *JobStartDetails) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	// Validate current status
	if job.Status != domain.JobStatusScheduled && job.Status != domain.JobStatusPending {
		return fmt.Errorf("job must be in scheduled or pending status to start")
	}

	// Update job status and start time
	job.Status = domain.JobStatusInProgress
	job.ActualStartTime = &startDetails.StartTime
	job.UpdatedAt = time.Now()

	// Store GPS check-in data
	if startDetails.GPSLocation != nil {
		job.GPSCheckIn = map[string]interface{}{
			"latitude":  startDetails.GPSLocation.Latitude,
			"longitude": startDetails.GPSLocation.Longitude,
			"address":   startDetails.GPSLocation.Address,
			"timestamp": startDetails.StartTime,
		}
	}

	// Add start notes
	if startDetails.Notes != nil {
		if job.Notes != nil {
			job.Notes = stringPtr(*job.Notes + "\n\nJob started: " + *startDetails.Notes)
		} else {
			job.Notes = stringPtr("Job started: " + *startDetails.Notes)
		}
	}

	// Save job
	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Printf("Failed to start job", "error", err, "job_id", jobID, "tenant_id", tenantID)
		return fmt.Errorf("failed to start job: %w", err)
	}

	// Upload photos if provided
	if len(startDetails.Photos) > 0 {
		for i, photoData := range startDetails.Photos {
			fileName := fmt.Sprintf("jobs/%s/start_photos/%d_%d.jpg", jobID, time.Now().Unix(), i)
			if _, err := s.storageService.Upload(ctx, fileName, []byte(photoData), "image/jpeg"); err != nil {
				s.logger.Printf("Failed to upload start photo", "error", err, "job_id", jobID)
			}
		}
	}

	// Send notifications
	if job.AssignedUserID != nil {
		if err := s.notificationService.SendNotification(ctx, &NotificationRequest{
			UserID:  job.AssignedUserID,
			Type:    "job.started",
			Title:   "Job Started",
			Message: fmt.Sprintf("Job '%s' has been started", job.Title),
			Data: map[string]interface{}{
				"job_id":    job.ID,
				"job_title": job.Title,
			},
		}); err != nil {
			s.logger.Printf("Failed to send job started notification", "error", err)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.start",
		ResourceType: "job",
		ResourceID:   &job.ID,
		NewValues: map[string]interface{}{
			"status":             domain.JobStatusInProgress,
			"actual_start_time":  startDetails.StartTime,
			"gps_check_in":      job.GPSCheckIn,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job started successfully", "job_id", jobID, "tenant_id", tenantID)
	return nil
}

// CompleteJob completes a job
func (s *JobServiceImpl) CompleteJob(ctx context.Context, jobID uuid.UUID, completionDetails *JobCompletionDetails) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	// Validate current status
	if job.Status != domain.JobStatusInProgress {
		return fmt.Errorf("job must be in progress to complete")
	}

	// Update job status and end time
	job.Status = domain.JobStatusCompleted
	job.ActualEndTime = &completionDetails.EndTime
	job.UpdatedAt = time.Now()

	// Store GPS check-out data
	if completionDetails.GPSLocation != nil {
		job.GPSCheckOut = map[string]interface{}{
			"latitude":  completionDetails.GPSLocation.Latitude,
			"longitude": completionDetails.GPSLocation.Longitude,
			"address":   completionDetails.GPSLocation.Address,
			"timestamp": completionDetails.EndTime,
		}
	}

	// Add completion notes
	if completionDetails.CompletionNotes != nil {
		if job.Notes != nil {
			job.Notes = stringPtr(*job.Notes + "\n\nJob completed: " + *completionDetails.CompletionNotes)
		} else {
			job.Notes = stringPtr("Job completed: " + *completionDetails.CompletionNotes)
		}
	}

	// Save job
	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Printf("Failed to complete job", "error", err, "job_id", jobID, "tenant_id", tenantID)
		return fmt.Errorf("failed to complete job: %w", err)
	}

	// Upload completion photos if provided
	if len(completionDetails.Photos) > 0 {
		photos := make([]string, 0, len(completionDetails.Photos))
		for i, photoData := range completionDetails.Photos {
			fileName := fmt.Sprintf("jobs/%s/completion_photos/%d_%d.jpg", jobID, time.Now().Unix(), i)
			url, err := s.storageService.Upload(ctx, fileName, []byte(photoData), "image/jpeg")
			if err != nil {
				s.logger.Printf("Failed to upload completion photo", "error", err, "job_id", jobID)
			} else {
				photos = append(photos, url)
			}
		}
		job.CompletionPhotos = photos
		if err := s.jobRepo.Update(ctx, job); err != nil {
			s.logger.Printf("Failed to update job with completion photos", "error", err)
		}
	}

	// Send notifications
	if job.AssignedUserID != nil {
		if err := s.notificationService.SendNotification(ctx, &NotificationRequest{
			UserID:  job.AssignedUserID,
			Type:    "job.completed",
			Title:   "Job Completed",
			Message: fmt.Sprintf("Job '%s' has been completed", job.Title),
			Data: map[string]interface{}{
				"job_id":    job.ID,
				"job_title": job.Title,
			},
		}); err != nil {
			s.logger.Printf("Failed to send job completed notification", "error", err)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.complete",
		ResourceType: "job",
		ResourceID:   &job.ID,
		NewValues: map[string]interface{}{
			"status":           domain.JobStatusCompleted,
			"actual_end_time":  completionDetails.EndTime,
			"gps_check_out":   job.GPSCheckOut,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job completed successfully", "job_id", jobID, "tenant_id", tenantID)
	return nil
}

// CancelJob cancels a job
func (s *JobServiceImpl) CancelJob(ctx context.Context, jobID uuid.UUID, reason string) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	// Validate current status
	if job.Status == domain.JobStatusCompleted || job.Status == domain.JobStatusCancelled {
		return fmt.Errorf("cannot cancel a %s job", job.Status)
	}

	// Store old status for audit
	oldStatus := job.Status

	// Update job status
	job.Status = domain.JobStatusCancelled
	job.UpdatedAt = time.Now()

	// Add cancellation reason to notes
	cancellationNote := fmt.Sprintf("Job cancelled: %s", reason)
	if job.Notes != nil {
		job.Notes = stringPtr(*job.Notes + "\n\n" + cancellationNote)
	} else {
		job.Notes = &cancellationNote
	}

	// Save job
	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Printf("Failed to cancel job", "error", err, "job_id", jobID, "tenant_id", tenantID)
		return fmt.Errorf("failed to cancel job: %w", err)
	}

	// Send notifications
	if job.AssignedUserID != nil {
		if err := s.notificationService.SendNotification(ctx, &NotificationRequest{
			UserID:  job.AssignedUserID,
			Type:    "job.cancelled",
			Title:   "Job Cancelled",
			Message: fmt.Sprintf("Job '%s' has been cancelled: %s", job.Title, reason),
			Data: map[string]interface{}{
				"job_id":    job.ID,
				"job_title": job.Title,
				"reason":    reason,
			},
		}); err != nil {
			s.logger.Printf("Failed to send job cancelled notification", "error", err)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.cancel",
		ResourceType: "job",
		ResourceID:   &job.ID,
		OldValues: map[string]interface{}{
			"status": oldStatus,
		},
		NewValues: map[string]interface{}{
			"status": domain.JobStatusCancelled,
			"reason": reason,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job cancelled successfully", "job_id", jobID, "reason", reason, "tenant_id", tenantID)
	return nil
}

// Helper methods

func (s *JobServiceImpl) validateCreateJobRequest(req *domain.CreateJobRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("job title is required")
	}
	if req.Priority != "low" && req.Priority != "medium" && req.Priority != "high" && req.Priority != "urgent" {
		return fmt.Errorf("priority must be one of: low, medium, high, urgent")
	}
	if req.CrewSize <= 0 {
		return fmt.Errorf("crew size must be greater than 0")
	}
	if req.EstimatedDuration != nil && *req.EstimatedDuration <= 0 {
		return fmt.Errorf("estimated duration must be greater than 0")
	}
	return nil
}

func (s *JobServiceImpl) validateJob(job *domain.EnhancedJob) error {
	if strings.TrimSpace(job.Title) == "" {
		return fmt.Errorf("job title is required")
	}
	if job.CrewSize <= 0 {
		return fmt.Errorf("crew size must be greater than 0")
	}
	return nil
}

func (s *JobServiceImpl) validateStatusTransition(currentStatus, newStatus string) error {
	validTransitions := map[string][]string{
		domain.JobStatusPending:    {domain.JobStatusScheduled, domain.JobStatusInProgress, domain.JobStatusCancelled},
		domain.JobStatusScheduled:  {domain.JobStatusInProgress, domain.JobStatusPending, domain.JobStatusCancelled},
		domain.JobStatusInProgress: {domain.JobStatusCompleted, domain.JobStatusOnHold, domain.JobStatusCancelled},
		domain.JobStatusOnHold:     {domain.JobStatusInProgress, domain.JobStatusCancelled},
		domain.JobStatusCompleted:  {}, // No transitions from completed
		domain.JobStatusCancelled:  {}, // No transitions from cancelled
	}

	allowedStatuses, exists := validTransitions[currentStatus]
	if !exists {
		return fmt.Errorf("invalid current status: %s", currentStatus)
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return nil
		}
	}

	return fmt.Errorf("cannot transition from %s to %s", currentStatus, newStatus)
}

// stringPtr helper is defined in billing_service.go

// AssignJob assigns a job to a user
func (s *JobServiceImpl) AssignJob(ctx context.Context, jobID, userID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	// Verify user exists
	user, err := s.userRepo.GetByID(ctx, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to verify user: %w", err)
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	// Update job assignment
	oldUserID := job.AssignedUserID
	job.AssignedUserID = &userID
	job.UpdatedAt = time.Now()

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Printf("Failed to assign job", "error", err, "job_id", jobID, "user_id", userID)
		return fmt.Errorf("failed to assign job: %w", err)
	}

	// Send notification
	if err := s.notificationService.SendNotification(ctx, &NotificationRequest{
		UserID:  &userID,
		Type:    "job.assigned",
		Title:   "Job Assigned",
		Message: fmt.Sprintf("You have been assigned to job: %s", job.Title),
		Data: map[string]interface{}{
			"job_id":    job.ID,
			"job_title": job.Title,
		},
	}); err != nil {
		s.logger.Printf("Failed to send job assignment notification", "error", err)
	}

	// Log audit event
	contextUserID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       contextUserID,
		Action:       "job.assign",
		ResourceType: "job",
		ResourceID:   &job.ID,
		OldValues: map[string]interface{}{
			"assigned_user_id": oldUserID,
		},
		NewValues: map[string]interface{}{
			"assigned_user_id": userID,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job assigned successfully", "job_id", jobID, "user_id", userID)
	return nil
}

// UnassignJob removes user assignment from a job
func (s *JobServiceImpl) UnassignJob(ctx context.Context, jobID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	oldUserID := job.AssignedUserID
	job.AssignedUserID = nil
	job.UpdatedAt = time.Now()

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Printf("Failed to unassign job", "error", err, "job_id", jobID)
		return fmt.Errorf("failed to unassign job: %w", err)
	}

	// Send notification to previously assigned user
	if oldUserID != nil {
		if err := s.notificationService.SendNotification(ctx, &NotificationRequest{
			UserID:  oldUserID,
			Type:    "job.unassigned",
			Title:   "Job Unassigned",
			Message: fmt.Sprintf("You have been unassigned from job: %s", job.Title),
			Data: map[string]interface{}{
				"job_id":    job.ID,
				"job_title": job.Title,
			},
		}); err != nil {
			s.logger.Printf("Failed to send job unassignment notification", "error", err)
		}
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.unassign",
		ResourceType: "job",
		ResourceID:   &job.ID,
		OldValues: map[string]interface{}{
			"assigned_user_id": oldUserID,
		},
		NewValues: map[string]interface{}{
			"assigned_user_id": nil,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job unassigned successfully", "job_id", jobID)
	return nil
}

// AssignJobToCrew assigns a job to a crew
func (s *JobServiceImpl) AssignJobToCrew(ctx context.Context, jobID, crewID uuid.UUID) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	// Verify crew exists
	crew, err := s.crewRepo.GetByID(ctx, tenantID, crewID)
	if err != nil {
		return fmt.Errorf("failed to verify crew: %w", err)
	}
	if crew == nil {
		return fmt.Errorf("crew not found")
	}

	// Check crew availability if job is scheduled
	if job.ScheduledDate != nil && job.EstimatedDuration != nil {
		endTime := job.ScheduledDate.Add(time.Duration(*job.EstimatedDuration) * time.Minute)
		available, err := s.crewRepo.CheckAvailability(ctx, crewID, *job.ScheduledDate, endTime)
		if err != nil {
			s.logger.Printf("Failed to check crew availability", "error", err)
		} else if !available {
			return fmt.Errorf("crew is not available at the scheduled time")
		}
	}

	// Update job with crew assignment
	// Note: This would require adding CrewID to the job model
	job.UpdatedAt = time.Now()

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Printf("Failed to assign job to crew", "error", err, "job_id", jobID, "crew_id", crewID)
		return fmt.Errorf("failed to assign job to crew: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.assign_crew",
		ResourceType: "job",
		ResourceID:   &job.ID,
		NewValues: map[string]interface{}{
			"crew_id": crewID,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job assigned to crew successfully", "job_id", jobID, "crew_id", crewID)
	return nil
}

// UpdateJobServices updates the services for a job
func (s *JobServiceImpl) UpdateJobServices(ctx context.Context, jobID uuid.UUID, services []*JobServiceUpdate) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	// Get current job services
	currentServices, err := s.jobRepo.GetJobServices(ctx, jobID)
	if err != nil {
		return fmt.Errorf("failed to get current job services: %w", err)
	}

	// Create map of current services for easy lookup
	currentServicesMap := make(map[uuid.UUID]*domain.JobService)
	for _, service := range currentServices {
		currentServicesMap[service.ServiceID] = service
	}

	var totalAmount float64

	// Process each service update
	for _, serviceUpdate := range services {
		// Verify service exists
		service, err := s.serviceRepo.GetByID(ctx, tenantID, serviceUpdate.ServiceID)
		if err != nil {
			return fmt.Errorf("failed to verify service %s: %w", serviceUpdate.ServiceID, err)
		}
		if service == nil {
			return fmt.Errorf("service %s not found", serviceUpdate.ServiceID)
		}

		totalPrice := serviceUpdate.Quantity * serviceUpdate.UnitPrice
		totalAmount += totalPrice

		if existingService, exists := currentServicesMap[serviceUpdate.ServiceID]; exists {
			// Update existing service
			existingService.Quantity = serviceUpdate.Quantity
			existingService.UnitPrice = serviceUpdate.UnitPrice
			existingService.TotalPrice = totalPrice

			if err := s.jobRepo.UpdateJobService(ctx, existingService); err != nil {
				return fmt.Errorf("failed to update job service: %w", err)
			}

			delete(currentServicesMap, serviceUpdate.ServiceID)
		} else {
			// Create new service
			jobService := &domain.JobService{
				ID:         uuid.New(),
				JobID:      jobID,
				ServiceID:  serviceUpdate.ServiceID,
				Quantity:   serviceUpdate.Quantity,
				UnitPrice:  serviceUpdate.UnitPrice,
				TotalPrice: totalPrice,
				CreatedAt:  time.Now(),
			}

			if err := s.jobRepo.CreateJobService(ctx, jobService); err != nil {
				return fmt.Errorf("failed to create job service: %w", err)
			}
		}
	}

	// Delete any remaining services that weren't in the update
	for _, service := range currentServicesMap {
		if err := s.jobRepo.DeleteJobService(ctx, service.ID); err != nil {
			s.logger.Printf("Failed to delete job service", "error", err, "service_id", service.ID)
		}
	}

	// Update job total amount
	job.TotalAmount = &totalAmount
	job.UpdatedAt = time.Now()

	if err := s.jobRepo.Update(ctx, job); err != nil {
		return fmt.Errorf("failed to update job total amount: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.update_services",
		ResourceType: "job",
		ResourceID:   &job.ID,
		NewValues: map[string]interface{}{
			"services_count": len(services),
			"total_amount":   totalAmount,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job services updated successfully", "job_id", jobID, "services_count", len(services))
	return nil
}

// GetJobServices retrieves services for a job
func (s *JobServiceImpl) GetJobServices(ctx context.Context, jobID uuid.UUID) ([]*domain.JobService, error) {
	services, err := s.jobRepo.GetJobServices(ctx, jobID)
	if err != nil {
		s.logger.Printf("Failed to get job services", "error", err, "job_id", jobID)
		return nil, fmt.Errorf("failed to get job services: %w", err)
	}

	return services, nil
}

// UploadJobPhotos uploads photos for a job
func (s *JobServiceImpl) UploadJobPhotos(ctx context.Context, jobID uuid.UUID, photos []*JobPhoto) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Verify job exists
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	uploadedPhotos := make([]string, 0, len(photos))

	// Upload each photo
	for i, photo := range photos {
		fileName := fmt.Sprintf("jobs/%s/photos/%d_%d.jpg", jobID, time.Now().Unix(), i)
		url, err := s.storageService.Upload(ctx, fileName, photo.Data, photo.ContentType)
		if err != nil {
			s.logger.Printf("Failed to upload job photo", "error", err, "job_id", jobID, "photo_index", i)
			continue
		}
		uploadedPhotos = append(uploadedPhotos, url)
	}

	// Update job with photo URLs
	if len(uploadedPhotos) > 0 {
		if job.CompletionPhotos == nil {
			job.CompletionPhotos = uploadedPhotos
		} else {
			job.CompletionPhotos = append(job.CompletionPhotos, uploadedPhotos...)
		}
		
		job.UpdatedAt = time.Now()

		if err := s.jobRepo.Update(ctx, job); err != nil {
			s.logger.Printf("Failed to update job with photos", "error", err, "job_id", jobID)
			return fmt.Errorf("failed to update job with photos: %w", err)
		}
	}

	s.logger.Printf("Job photos uploaded successfully", "job_id", jobID, "photos_count", len(uploadedPhotos))
	return nil
}

// AddJobSignature adds a digital signature to a job
func (s *JobServiceImpl) AddJobSignature(ctx context.Context, jobID uuid.UUID, signature string) error {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return fmt.Errorf("tenant ID not found in context")
	}

	// Get job
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return fmt.Errorf("job not found")
	}

	// Add signature
	job.CustomerSignature = &signature
	job.UpdatedAt = time.Now()

	if err := s.jobRepo.Update(ctx, job); err != nil {
		s.logger.Printf("Failed to add job signature", "error", err, "job_id", jobID)
		return fmt.Errorf("failed to add job signature: %w", err)
	}

	// Log audit event
	userID := GetUserIDFromContext(ctx)
	if err := s.auditService.LogAction(ctx, &AuditLogRequest{
		UserID:       userID,
		Action:       "job.add_signature",
		ResourceType: "job",
		ResourceID:   &job.ID,
		NewValues: map[string]interface{}{
			"signature_added": true,
		},
	}); err != nil {
		s.logger.Printf("Failed to log audit event", "error", err)
	}

	s.logger.Printf("Job signature added successfully", "job_id", jobID)
	return nil
}

// GetJobSchedule retrieves scheduled jobs based on filter
func (s *JobServiceImpl) GetJobSchedule(ctx context.Context, filter *ScheduleFilter) ([]*ScheduledJob, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Convert schedule filter to job filter
	jobFilter := &JobFilter{
		Status: "scheduled,in_progress",
	}

	if filter != nil {
		if filter.StartDate != nil && filter.EndDate != nil {
			jobFilter.ScheduledStart = filter.StartDate
			jobFilter.ScheduledEnd = filter.EndDate
		}
		if filter.AssignedUserID != nil {
			jobFilter.AssignedUserID = filter.AssignedUserID
		}
	}

	jobs, _, err := s.jobRepo.List(ctx, tenantID, jobFilter)
	if err != nil {
		return nil, fmt.Errorf("failed to get scheduled jobs: %w", err)
	}

	// Convert to scheduled job format
	scheduledJobs := make([]*ScheduledJob, 0, len(jobs))
	for _, job := range jobs {
		if job.ScheduledDate == nil {
			continue
		}

		// Get customer name
		customer, err := s.customerRepo.GetByID(ctx, tenantID, job.CustomerID)
		if err != nil {
			s.logger.Printf("Failed to get customer for scheduled job", "error", err, "job_id", job.ID)
			continue
		}

		// Get property address
		property, err := s.propertyRepo.GetByID(ctx, tenantID, job.PropertyID)
		if err != nil {
			s.logger.Printf("Failed to get property for scheduled job", "error", err, "job_id", job.ID)
			continue
		}

		var assignedUser *string
		if job.AssignedUserID != nil {
			user, err := s.userRepo.GetByID(ctx, tenantID, *job.AssignedUserID)
			if err == nil && user != nil {
				name := fmt.Sprintf("%s %s", user.FirstName, user.LastName)
				assignedUser = &name
			}
		}

		scheduledJob := &ScheduledJob{
			JobID:           job.ID,
			Title:           job.Title,
			CustomerName:    fmt.Sprintf("%s %s", customer.FirstName, customer.LastName),
			PropertyAddress: fmt.Sprintf("%s, %s, %s", property.AddressLine1, property.City, property.State),
			ScheduledDate:   *job.ScheduledDate,
			ScheduledTime:   job.ScheduledTime,
			Duration:        job.EstimatedDuration,
			Status:          job.Status,
			AssignedUser:    assignedUser,
			Priority:        job.Priority,
		}

		scheduledJobs = append(scheduledJobs, scheduledJob)
	}

	return scheduledJobs, nil
}

// GetJobCalendar retrieves calendar events for jobs
func (s *JobServiceImpl) GetJobCalendar(ctx context.Context, startDate, endDate time.Time) ([]*CalendarEvent, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	jobs, err := s.jobRepo.GetByDateRange(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs for calendar: %w", err)
	}

	events := make([]*CalendarEvent, 0, len(jobs))
	for _, job := range jobs {
		if job.ScheduledDate == nil {
			continue
		}

		startTime := *job.ScheduledDate
		endTime := startTime
		if job.EstimatedDuration != nil {
			endTime = startTime.Add(time.Duration(*job.EstimatedDuration) * time.Minute)
		} else {
			endTime = startTime.Add(2 * time.Hour) // Default 2 hours
		}

		// Get property location
		var location *Location
		property, err := s.propertyRepo.GetByID(ctx, tenantID, job.PropertyID)
		if err == nil && property != nil && property.Latitude != nil && property.Longitude != nil {
			location = &Location{
				Latitude:  *property.Latitude,
				Longitude: *property.Longitude,
				Address:   fmt.Sprintf("%s, %s, %s", property.AddressLine1, property.City, property.State),
			}
		}

		event := &CalendarEvent{
			ID:          job.ID,
			Title:       job.Title,
			Description: getStringOrEmpty(job.Description),
			StartTime:   startTime,
			EndTime:     endTime,
			Type:        "job",
			Status:      job.Status,
			Location:    location,
		}

		events = append(events, event)
	}

	return events, nil
}

// CreateRecurringJob creates a recurring job series
func (s *JobServiceImpl) CreateRecurringJob(ctx context.Context, req *RecurringJobRequest) (*RecurringJobSeries, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get base job
	baseJob, err := s.jobRepo.GetByID(ctx, tenantID, req.BaseJobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get base job: %w", err)
	}
	if baseJob == nil {
		return nil, fmt.Errorf("base job not found")
	}

	// Create recurring job series
	series := &RecurringJobSeries{
		ID:             uuid.New(),
		BaseJobID:      req.BaseJobID,
		Frequency:      req.Frequency,
		NextOccurrence: req.StartDate,
		JobsCreated:    0,
		UpcomingJobs:   []uuid.UUID{},
	}

	if err := s.jobRepo.CreateRecurringJobSeries(ctx, series); err != nil {
		return nil, fmt.Errorf("failed to create recurring job series: %w", err)
	}

	// Create initial jobs based on frequency
	// This is a simplified implementation - in practice, you'd have more sophisticated scheduling
	var nextDate time.Time = req.StartDate
	maxJobs := 12 // Limit initial creation to 12 jobs

	if req.MaxOccurrences != nil && *req.MaxOccurrences < maxJobs {
		maxJobs = *req.MaxOccurrences
	}

	for i := 0; i < maxJobs; i++ {
		if req.EndDate != nil && nextDate.After(*req.EndDate) {
			break
		}

		// Create new job based on base job
		newJob := &domain.EnhancedJob{
			Job: domain.Job{
				ID:                uuid.New(),
				TenantID:          baseJob.TenantID,
				CustomerID:        baseJob.CustomerID,
				PropertyID:        baseJob.PropertyID,
				AssignedUserID:    baseJob.AssignedUserID,
				Title:             baseJob.Title,
				Description:       baseJob.Description,
				Status:            domain.JobStatusPending,
				Priority:          baseJob.Priority,
				ScheduledDate:     &nextDate,
				ScheduledTime:     baseJob.ScheduledTime,
				EstimatedDuration: baseJob.EstimatedDuration,
				CreatedAt:         time.Now(),
				UpdatedAt:         time.Now(),
			},
			ParentJobID:       &baseJob.ID,
			RecurringSchedule: &req.Frequency,
			CrewSize:          baseJob.CrewSize,
			WeatherDependent:  baseJob.WeatherDependent,
			RequiresEquipment: baseJob.RequiresEquipment,
		}

		if err := s.jobRepo.Create(ctx, newJob); err != nil {
			s.logger.Printf("Failed to create recurring job instance", "error", err, "iteration", i)
			continue
		}

		series.UpcomingJobs = append(series.UpcomingJobs, newJob.ID)
		series.JobsCreated++

		// Calculate next date based on frequency
		switch req.Frequency {
		case "weekly":
			nextDate = nextDate.AddDate(0, 0, 7)
		case "biweekly":
			nextDate = nextDate.AddDate(0, 0, 14)
		case "monthly":
			nextDate = nextDate.AddDate(0, 1, 0)
		case "quarterly":
			nextDate = nextDate.AddDate(0, 3, 0)
		default:
			return nil, fmt.Errorf("unsupported frequency: %s", req.Frequency)
		}
	}

	// Update series with next occurrence
	series.NextOccurrence = nextDate

	s.logger.Printf("Recurring job series created", "series_id", series.ID, "jobs_created", series.JobsCreated)
	return series, nil
}

// OptimizeJobRoute optimizes the route for multiple jobs
func (s *JobServiceImpl) OptimizeJobRoute(ctx context.Context, jobIDs []uuid.UUID, date time.Time) (*RouteOptimization, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get all jobs
	jobs := make([]*domain.EnhancedJob, 0, len(jobIDs))
	for _, jobID := range jobIDs {
		job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
		if err != nil {
			s.logger.Printf("Failed to get job for route optimization", "error", err, "job_id", jobID)
			continue
		}
		if job != nil {
			jobs = append(jobs, job)
		}
	}

	if len(jobs) == 0 {
		return nil, fmt.Errorf("no valid jobs found for route optimization")
	}

	// Use the schedule service for route optimization
	optimization, err := s.scheduleService.OptimizeRoute(ctx, jobs, &Location{
		Latitude:  40.7128, // Default starting location (could be configurable)
		Longitude: -74.0060,
		Address:   "Office",
	})

	if err != nil {
		return nil, fmt.Errorf("failed to optimize route: %w", err)
	}

	s.logger.Printf("Route optimized successfully", "jobs_count", len(jobs), "total_distance", optimization.TotalDistance)
	return optimization, nil
}

// CheckSchedulingConflicts checks for scheduling conflicts
func (s *JobServiceImpl) CheckSchedulingConflicts(ctx context.Context, jobID uuid.UUID, scheduledDate time.Time, estimatedDuration int, assignedUserID *uuid.UUID) ([]SchedulingConflict, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	conflicts := make([]SchedulingConflict, 0)
	endTime := scheduledDate.Add(time.Duration(estimatedDuration) * time.Minute)

	// Check for user conflicts if user is assigned
	if assignedUserID != nil {
		userJobs, _, err := s.jobRepo.GetByAssignedUserID(ctx, tenantID, *assignedUserID, &JobFilter{
			Status: "scheduled,in_progress",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to check user scheduling conflicts: %w", err)
		}

		for _, job := range userJobs {
			if job.ID == jobID { // Skip self
				continue
			}
			if job.ScheduledDate != nil && job.EstimatedDuration != nil {
				jobEndTime := job.ScheduledDate.Add(time.Duration(*job.EstimatedDuration) * time.Minute)
				
				// Check for overlap
				if scheduledDate.Before(jobEndTime) && endTime.After(*job.ScheduledDate) {
					conflicts = append(conflicts, SchedulingConflict{
						Type:        "user_conflict",
						ConflictingJobID: job.ID,
						ConflictingJobTitle: job.Title,
						ConflictTime: TimeRange{
							Start: *job.ScheduledDate,
							End:   jobEndTime,
						},
						Severity: "high",
						Message:  fmt.Sprintf("User is already assigned to job '%s' during this time", job.Title),
					})
				}
			}
		}
	}

	return conflicts, nil
}

// SuggestOptimalSchedule suggests optimal scheduling for a job
func (s *JobServiceImpl) SuggestOptimalSchedule(ctx context.Context, jobID uuid.UUID, preferredDate *time.Time, constraints SchedulingConstraints) (*SchedulingSuggestion, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get job details
	job, err := s.jobRepo.GetByID(ctx, tenantID, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}
	if job == nil {
		return nil, fmt.Errorf("job not found")
	}

	// Determine search window
	searchStart := time.Now()
	if preferredDate != nil {
		searchStart = *preferredDate
	}
	searchEnd := searchStart.AddDate(0, 0, 30) // Search within 30 days

	suggestions := make([]TimeSlot, 0)
	
	// Generate potential time slots (simplified algorithm)
	for d := 0; d < 30; d++ {
		currentDate := searchStart.AddDate(0, 0, d)
		
		// Stop if we've exceeded the search window
		if currentDate.After(searchEnd) {
			break
		}
		
		// Skip weekends if constraints specify weekdays only
		if constraints.WeekdaysOnly && (currentDate.Weekday() == time.Saturday || currentDate.Weekday() == time.Sunday) {
			continue
		}

		// Check available time slots throughout the day
		for hour := constraints.EarliestStart; hour <= constraints.LatestStart; hour++ {
			slotStart := time.Date(currentDate.Year(), currentDate.Month(), currentDate.Day(), hour, 0, 0, 0, currentDate.Location())
			
			if job.EstimatedDuration == nil {
				continue
			}

			// Check for conflicts
			conflicts, err := s.CheckSchedulingConflicts(ctx, jobID, slotStart, *job.EstimatedDuration, job.AssignedUserID)
			if err != nil {
				continue
			}

			if len(conflicts) == 0 {
				// Calculate score based on various factors
				score := s.calculateSchedulingScore(slotStart, preferredDate, constraints)
				
				suggestions = append(suggestions, TimeSlot{
					StartTime: slotStart,
					EndTime:   slotStart.Add(time.Duration(*job.EstimatedDuration) * time.Minute),
					Score:     score,
					Available: true,
				})
			}
		}
	}

	// Sort suggestions by score (highest first)
	for i := 0; i < len(suggestions)-1; i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[j].Score > suggestions[i].Score {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	// Return top 5 suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return &SchedulingSuggestion{
		JobID:       jobID,
		Suggestions: suggestions,
		Constraints: constraints,
	}, nil
}

// GetJobsByLocation gets jobs within a geographic area for route planning
func (s *JobServiceImpl) GetJobsByLocation(ctx context.Context, centerLat, centerLng, radiusMiles float64, date time.Time) ([]*domain.EnhancedJob, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get jobs scheduled for the date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	jobs, err := s.jobRepo.GetByDateRange(ctx, tenantID, startOfDay, endOfDay)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs by date: %w", err)
	}

	// Filter jobs by location
	nearbyJobs := make([]*domain.EnhancedJob, 0)
	for _, job := range jobs {
		// Get property coordinates
		property, err := s.propertyRepo.GetByID(ctx, tenantID, job.PropertyID)
		if err != nil {
			continue
		}
		if property == nil || property.Latitude == nil || property.Longitude == nil {
			continue
		}

		// Calculate distance
		distance := haversineDistance(centerLat, centerLng, *property.Latitude, *property.Longitude)
		if distance <= radiusMiles {
			nearbyJobs = append(nearbyJobs, job)
		}
	}

	return nearbyJobs, nil
}

// OptimizeRouteForUser optimizes the route for a specific user's jobs
func (s *JobServiceImpl) OptimizeRouteForUser(ctx context.Context, userID uuid.UUID, date time.Time) (*RouteOptimization, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	// Get user's jobs for the date
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := &JobFilter{
		AssignedUserID:  &userID,
		Status:         "scheduled,in_progress",
		ScheduledStart: &startOfDay,
		ScheduledEnd:   &endOfDay,
	}

	jobs, _, err := s.jobRepo.List(ctx, tenantID, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get user jobs: %w", err)
	}

	if len(jobs) <= 1 {
		return &RouteOptimization{
			OptimizedRoute: []RouteStop{},
			TotalDistance:  0,
			TotalDuration:  0,
			Savings:        0,
		}, nil
	}

	// Use schedule service for optimization
	return s.scheduleService.OptimizeRoute(ctx, jobs, &Location{
		Latitude:  40.7128, // Default starting location
		Longitude: -74.0060,
		Address:   "Office",
	})
}

// GetJobUtilizationMetrics gets utilization metrics for jobs
func (s *JobServiceImpl) GetJobUtilizationMetrics(ctx context.Context, startDate, endDate time.Time) (*JobUtilizationMetrics, error) {
	tenantID, ok := GetTenantIDFromContext(ctx)
	if !ok {
		return nil, fmt.Errorf("tenant ID not found in context")
	}

	jobs, err := s.jobRepo.GetByDateRange(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get jobs for metrics: %w", err)
	}

	metrics := &JobUtilizationMetrics{
		Period:           TimeRange{Start: startDate, End: endDate},
		TotalJobs:        len(jobs),
		CompletedJobs:    0,
		CancelledJobs:    0,
		InProgressJobs:   0,
		TotalRevenue:     0,
		AverageJobValue:  0,
		TotalDuration:    0,
		AverageDuration:  0,
	}

	var totalJobValue float64
	var totalDurationMinutes int
	var completedJobsWithDuration int

	for _, job := range jobs {
		switch job.Status {
		case domain.JobStatusCompleted:
			metrics.CompletedJobs++
			if job.TotalAmount != nil {
				totalJobValue += *job.TotalAmount
				metrics.TotalRevenue += *job.TotalAmount
			}
			
			// Calculate actual duration if available
			if job.ActualStartTime != nil && job.ActualEndTime != nil {
				duration := int(job.ActualEndTime.Sub(*job.ActualStartTime).Minutes())
				totalDurationMinutes += duration
				completedJobsWithDuration++
			}
			
		case domain.JobStatusCancelled:
			metrics.CancelledJobs++
		case domain.JobStatusInProgress:
			metrics.InProgressJobs++
		}
	}

	if metrics.CompletedJobs > 0 {
		metrics.AverageJobValue = totalJobValue / float64(metrics.CompletedJobs)
	}

	if completedJobsWithDuration > 0 {
		metrics.AverageDuration = totalDurationMinutes / completedJobsWithDuration
	}

	metrics.TotalDuration = totalDurationMinutes
	metrics.CompletionRate = float64(metrics.CompletedJobs) / float64(metrics.TotalJobs) * 100

	return metrics, nil
}

// Helper methods for scheduling

func (s *JobServiceImpl) calculateSchedulingScore(slotStart time.Time, preferredDate *time.Time, constraints SchedulingConstraints) float64 {
	score := 100.0

	// Prefer preferred date
	if preferredDate != nil {
		daysDiff := slotStart.Sub(*preferredDate).Hours() / 24
		if daysDiff < 0 {
			daysDiff = -daysDiff
		}
		score -= daysDiff * 5 // Penalize each day away from preferred date
	}

	// Prefer certain hours
	hour := slotStart.Hour()
	if hour >= 9 && hour <= 15 {
		score += 20 // Prefer business hours
	} else if hour >= 8 && hour <= 17 {
		score += 10 // Acceptable hours
	}

	// Prefer weekdays if needed
	if constraints.WeekdaysOnly && (slotStart.Weekday() != time.Saturday && slotStart.Weekday() != time.Sunday) {
		score += 15
	}

	return score
}

// Additional helper structures
type ScheduleFilter struct {
	StartDate      *time.Time
	EndDate        *time.Time
	AssignedUserID *uuid.UUID
	Status         string
}

type SchedulingConflict struct {
	Type                string    `json:"type"`
	ConflictingJobID    uuid.UUID `json:"conflicting_job_id"`
	ConflictingJobTitle string    `json:"conflicting_job_title"`
	ConflictTime        TimeRange `json:"conflict_time"`
	Severity           string    `json:"severity"`
	Message            string    `json:"message"`
}

type SchedulingConstraints struct {
	WeekdaysOnly   bool `json:"weekdays_only"`
	EarliestStart  int  `json:"earliest_start"`  // Hour (0-23)
	LatestStart    int  `json:"latest_start"`    // Hour (0-23)
	MinDuration    int  `json:"min_duration"`    // Minutes
	MaxDuration    int  `json:"max_duration"`    // Minutes
}

type SchedulingSuggestion struct {
	JobID       uuid.UUID             `json:"job_id"`
	Suggestions []TimeSlot            `json:"suggestions"`
	Constraints SchedulingConstraints `json:"constraints"`
}

type TimeSlot struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Score     float64   `json:"score"`
	Available bool      `json:"available"`
}

type JobUtilizationMetrics struct {
	Period           TimeRange `json:"period"`
	TotalJobs        int       `json:"total_jobs"`
	CompletedJobs    int       `json:"completed_jobs"`
	CancelledJobs    int       `json:"cancelled_jobs"`
	InProgressJobs   int       `json:"in_progress_jobs"`
	TotalRevenue     float64   `json:"total_revenue"`
	AverageJobValue  float64   `json:"average_job_value"`
	TotalDuration    int       `json:"total_duration_minutes"`
	AverageDuration  int       `json:"average_duration_minutes"`
	CompletionRate   float64   `json:"completion_rate"`
}

// Helper function to handle base price safely
func getBasePriceOrZero(basePrice *float64) float64 {
	if basePrice != nil {
		return *basePrice
	}
	return 0.0
}

// Helper function to handle string pointer safely
func getStringOrEmpty(str *string) string {
	if str != nil {
		return *str
	}
	return ""
}