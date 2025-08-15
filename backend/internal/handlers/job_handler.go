package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"log"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// JobHandler handles HTTP requests for job operations
type JobHandler struct {
	jobService services.JobService
	logger     *log.Logger
}

// NewJobHandler creates a new job handler
func NewJobHandler(jobService services.JobService, logger *log.Logger) *JobHandler {
	return &JobHandler{
		jobService: jobService,
		logger:     logger,
	}
}

// RegisterRoutes registers job routes with the router
func (h *JobHandler) RegisterRoutes(router *mux.Router) {
	// Job CRUD routes
	router.HandleFunc("/jobs", h.CreateJob).Methods("POST")
	router.HandleFunc("/jobs", h.ListJobs).Methods("GET")
	router.HandleFunc("/jobs/{id}", h.GetJob).Methods("GET")
	router.HandleFunc("/jobs/{id}", h.UpdateJob).Methods("PUT")
	router.HandleFunc("/jobs/{id}", h.DeleteJob).Methods("DELETE")
	
	// Job lifecycle routes
	router.HandleFunc("/jobs/{id}/start", h.StartJob).Methods("POST")
	router.HandleFunc("/jobs/{id}/complete", h.CompleteJob).Methods("POST")
	router.HandleFunc("/jobs/{id}/cancel", h.CancelJob).Methods("POST")
	
	// Job assignment routes
	router.HandleFunc("/jobs/{id}/assign", h.AssignJob).Methods("POST")
	router.HandleFunc("/jobs/{id}/unassign", h.UnassignJob).Methods("POST")
	router.HandleFunc("/jobs/{id}/assign-crew", h.AssignJobToCrew).Methods("POST")
	
	// Job services and media routes
	router.HandleFunc("/jobs/{id}/services", h.GetJobServices).Methods("GET")
	router.HandleFunc("/jobs/{id}/services", h.UpdateJobServices).Methods("PUT")
	router.HandleFunc("/jobs/{id}/photos", h.UploadJobPhotos).Methods("POST")
	router.HandleFunc("/jobs/{id}/signature", h.AddJobSignature).Methods("POST")
	
	// Scheduling and calendar routes
	router.HandleFunc("/jobs/schedule", h.GetJobSchedule).Methods("GET")
	router.HandleFunc("/jobs/calendar", h.GetJobCalendar).Methods("GET")
	router.HandleFunc("/jobs/optimize-route", h.OptimizeJobRoute).Methods("POST")
	
	// Recurring jobs
	router.HandleFunc("/jobs/recurring", h.CreateRecurringJob).Methods("POST")
}

// CreateJob creates a new job
// @Summary Create a new job
// @Description Create a new job/work order in the system
// @Tags jobs
// @Accept json
// @Produce json
// @Param job body domain.CreateJobRequest true "Job creation request"
// @Success 201 {object} domain.EnhancedJob
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs [post]
func (h *JobHandler) CreateJob(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	job, err := h.jobService.CreateJob(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create job", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create job", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, job)
}

// GetJob retrieves a job by ID
// @Summary Get a job by ID
// @Description Retrieve a job by its unique ID
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} domain.EnhancedJob
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id} [get]
func (h *JobHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	job, err := h.jobService.GetJob(r.Context(), jobID)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to get job", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get job", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, job)
}

// UpdateJob updates an existing job
// @Summary Update a job
// @Description Update an existing job's information
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param job body domain.UpdateJobRequest true "Job update request"
// @Success 200 {object} domain.EnhancedJob
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id} [put]
func (h *JobHandler) UpdateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var req domain.UpdateJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	job, err := h.jobService.UpdateJob(r.Context(), jobID, &req)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to update job", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update job", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, job)
}

// DeleteJob deletes a job
// @Summary Delete a job
// @Description Delete a job from the system
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id} [delete]
func (h *JobHandler) DeleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	err = h.jobService.DeleteJob(r.Context(), jobID)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to delete job", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete job", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListJobs lists jobs with pagination and filtering
// @Summary List jobs
// @Description Get a paginated list of jobs with optional filtering
// @Tags jobs
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param sort_by query string false "Sort field"
// @Param sort_desc query bool false "Sort descending"
// @Param status query string false "Job status filter"
// @Param priority query string false "Job priority filter"
// @Param assigned_user_id query string false "Assigned user ID filter"
// @Param customer_id query string false "Customer ID filter"
// @Param property_id query string false "Property ID filter"
// @Param scheduled_start query string false "Scheduled start date (YYYY-MM-DD)"
// @Param scheduled_end query string false "Scheduled end date (YYYY-MM-DD)"
// @Param search query string false "Search query"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs [get]
func (h *JobHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseJobFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.jobService.ListJobs(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list jobs", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list jobs", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// StartJob starts a job
// @Summary Start a job
// @Description Start a job with GPS check-in and other details
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param details body services.JobStartDetails true "Job start details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/start [post]
func (h *JobHandler) StartJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var details services.JobStartDetails
	if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err = h.jobService.StartJob(r.Context(), jobID, &details)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to start job", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to start job", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Job started successfully"})
}

// CompleteJob completes a job
// @Summary Complete a job
// @Description Complete a job with GPS check-out and completion details
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param details body services.JobCompletionDetails true "Job completion details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/complete [post]
func (h *JobHandler) CompleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var details services.JobCompletionDetails
	if err := json.NewDecoder(r.Body).Decode(&details); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err = h.jobService.CompleteJob(r.Context(), jobID, &details)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to complete job", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to complete job", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Job completed successfully"})
}

// CancelJob cancels a job
// @Summary Cancel a job
// @Description Cancel a job with a reason
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param request body map[string]string true "Cancellation request with reason"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/cancel [post]
func (h *JobHandler) CancelJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var request map[string]string
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	reason := request["reason"]
	if reason == "" {
		h.respondWithError(w, http.StatusBadRequest, "Cancellation reason is required", nil)
		return
	}

	err = h.jobService.CancelJob(r.Context(), jobID, reason)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to cancel job", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to cancel job", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Job cancelled successfully"})
}

// AssignJob assigns a job to a user
// @Summary Assign job to user
// @Description Assign a job to a specific user
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param assignment body map[string]string true "Assignment request with user_id"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/assign [post]
func (h *JobHandler) AssignJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var assignment map[string]string
	if err := json.NewDecoder(r.Body).Decode(&assignment); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	userIDStr := assignment["user_id"]
	if userIDStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid user ID", err)
		return
	}

	err = h.jobService.AssignJob(r.Context(), jobID, userID)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to assign job", "error", err, "job_id", jobID, "user_id", userID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to assign job", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Job assigned successfully"})
}

// UnassignJob unassigns a job
// @Summary Unassign job
// @Description Remove assignment from a job
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/unassign [post]
func (h *JobHandler) UnassignJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	err = h.jobService.UnassignJob(r.Context(), jobID)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to unassign job", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to unassign job", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Job unassigned successfully"})
}

// AssignJobToCrew assigns a job to a crew
// @Summary Assign job to crew
// @Description Assign a job to a specific crew
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param assignment body map[string]string true "Crew assignment request with crew_id"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/assign-crew [post]
func (h *JobHandler) AssignJobToCrew(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var assignment map[string]string
	if err := json.NewDecoder(r.Body).Decode(&assignment); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	crewIDStr := assignment["crew_id"]
	if crewIDStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Crew ID is required", nil)
		return
	}

	crewID, err := uuid.Parse(crewIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid crew ID", err)
		return
	}

	err = h.jobService.AssignJobToCrew(r.Context(), jobID, crewID)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to assign job to crew", "error", err, "job_id", jobID, "crew_id", crewID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to assign job to crew", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Job assigned to crew successfully"})
}

// GetJobServices gets services for a job
// @Summary Get job services
// @Description Retrieve services associated with a job
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Success 200 {array} domain.JobService
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/services [get]
func (h *JobHandler) GetJobServices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	services, err := h.jobService.GetJobServices(r.Context(), jobID)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to get job services", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get job services", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, services)
}

// UpdateJobServices updates services for a job
// @Summary Update job services
// @Description Update the services associated with a job
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param services body []services.JobServiceUpdate true "Job service updates"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/services [put]
func (h *JobHandler) UpdateJobServices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var services []*services.JobServiceUpdate
	if err := json.NewDecoder(r.Body).Decode(&services); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err = h.jobService.UpdateJobServices(r.Context(), jobID, services)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to update job services", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update job services", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Job services updated successfully"})
}

// UploadJobPhotos uploads photos for a job
// @Summary Upload job photos
// @Description Upload photos associated with a job
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param photos body []services.JobPhoto true "Job photos"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/photos [post]
func (h *JobHandler) UploadJobPhotos(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var photos []*services.JobPhoto
	if err := json.NewDecoder(r.Body).Decode(&photos); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err = h.jobService.UploadJobPhotos(r.Context(), jobID, photos)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to upload job photos", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to upload job photos", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Job photos uploaded successfully"})
}

// AddJobSignature adds customer signature to a job
// @Summary Add job signature
// @Description Add customer signature to a job
// @Tags jobs
// @Accept json
// @Produce json
// @Param id path string true "Job ID"
// @Param signature body map[string]string true "Signature data"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/{id}/signature [post]
func (h *JobHandler) AddJobSignature(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	var signatureData map[string]string
	if err := json.NewDecoder(r.Body).Decode(&signatureData); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	signature := signatureData["signature"]
	if signature == "" {
		h.respondWithError(w, http.StatusBadRequest, "Signature data is required", nil)
		return
	}

	err = h.jobService.AddJobSignature(r.Context(), jobID, signature)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to add job signature", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to add job signature", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Job signature added successfully"})
}

// GetJobSchedule gets job schedule
// @Summary Get job schedule
// @Description Retrieve job schedule with filtering
// @Tags jobs
// @Accept json
// @Produce json
// @Param user_id query string false "User ID filter"
// @Param crew_id query string false "Crew ID filter"
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Success 200 {array} services.ScheduledJob
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/schedule [get]
func (h *JobHandler) GetJobSchedule(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseScheduleFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	schedule, err := h.jobService.GetJobSchedule(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to get job schedule", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get job schedule", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, schedule)
}

// GetJobCalendar gets job calendar events
// @Summary Get job calendar
// @Description Retrieve job calendar events for a date range
// @Tags jobs
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {array} services.CalendarEvent
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/calendar [get]
func (h *JobHandler) GetJobCalendar(w http.ResponseWriter, r *http.Request) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	if startDateStr == "" || endDateStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Start date and end date are required", nil)
		return
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid start date format", err)
		return
	}

	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid end date format", err)
		return
	}

	events, err := h.jobService.GetJobCalendar(r.Context(), startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get job calendar", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get job calendar", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, events)
}

// OptimizeJobRoute optimizes route for jobs
// @Summary Optimize job route
// @Description Optimize route for a set of jobs on a specific date
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Route optimization request"
// @Success 200 {object} services.RouteOptimization
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/optimize-route [post]
func (h *JobHandler) OptimizeJobRoute(w http.ResponseWriter, r *http.Request) {
	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Parse job IDs
	jobIDsInterface, ok := request["job_ids"]
	if !ok {
		h.respondWithError(w, http.StatusBadRequest, "Job IDs are required", nil)
		return
	}

	jobIDsArray, ok := jobIDsInterface.([]interface{})
	if !ok {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job IDs format", nil)
		return
	}

	jobIDs := make([]uuid.UUID, 0, len(jobIDsArray))
	for _, id := range jobIDsArray {
		idStr, ok := id.(string)
		if !ok {
			h.respondWithError(w, http.StatusBadRequest, "Invalid job ID format", nil)
			return
		}
		jobID, err := uuid.Parse(idStr)
		if err != nil {
			h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
			return
		}
		jobIDs = append(jobIDs, jobID)
	}

	// Parse date
	dateStr, ok := request["date"].(string)
	if !ok {
		h.respondWithError(w, http.StatusBadRequest, "Date is required", nil)
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid date format", err)
		return
	}

	optimization, err := h.jobService.OptimizeJobRoute(r.Context(), jobIDs, date)
	if err != nil {
		h.logger.Error("Failed to optimize job route", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to optimize job route", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, optimization)
}

// CreateRecurringJob creates a recurring job series
// @Summary Create recurring job
// @Description Create a recurring job series
// @Tags jobs
// @Accept json
// @Produce json
// @Param request body services.RecurringJobRequest true "Recurring job request"
// @Success 201 {object} services.RecurringJobSeries
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /jobs/recurring [post]
func (h *JobHandler) CreateRecurringJob(w http.ResponseWriter, r *http.Request) {
	var req services.RecurringJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	series, err := h.jobService.CreateRecurringJob(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create recurring job", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create recurring job", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, series)
}

// Helper methods

func (h *JobHandler) parseJobFilter(r *http.Request) (*services.JobFilter, error) {
	filter := &services.JobFilter{}

	// Parse pagination
	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			filter.Page = p
		}
	}
	if perPage := r.URL.Query().Get("per_page"); perPage != "" {
		if pp, err := strconv.Atoi(perPage); err == nil && pp > 0 && pp <= 100 {
			filter.PerPage = pp
		}
	}

	// Parse sorting
	filter.SortBy = r.URL.Query().Get("sort_by")
	if sortDesc := r.URL.Query().Get("sort_desc"); sortDesc != "" {
		if sd, err := strconv.ParseBool(sortDesc); err == nil {
			filter.SortDesc = sd
		}
	}

	// Parse filters
	filter.Status = r.URL.Query().Get("status")
	filter.Priority = r.URL.Query().Get("priority")
	filter.Search = r.URL.Query().Get("search")

	// Parse UUID filters
	if assignedUserIDStr := r.URL.Query().Get("assigned_user_id"); assignedUserIDStr != "" {
		if assignedUserID, err := uuid.Parse(assignedUserIDStr); err == nil {
			filter.AssignedUserID = &assignedUserID
		}
	}
	if customerIDStr := r.URL.Query().Get("customer_id"); customerIDStr != "" {
		if customerID, err := uuid.Parse(customerIDStr); err == nil {
			filter.CustomerID = &customerID
		}
	}
	if propertyIDStr := r.URL.Query().Get("property_id"); propertyIDStr != "" {
		if propertyID, err := uuid.Parse(propertyIDStr); err == nil {
			filter.PropertyID = &propertyID
		}
	}

	// Parse date filters
	if scheduledStartStr := r.URL.Query().Get("scheduled_start"); scheduledStartStr != "" {
		if scheduledStart, err := time.Parse("2006-01-02", scheduledStartStr); err == nil {
			filter.ScheduledStart = &scheduledStart
		}
	}
	if scheduledEndStr := r.URL.Query().Get("scheduled_end"); scheduledEndStr != "" {
		if scheduledEnd, err := time.Parse("2006-01-02", scheduledEndStr); err == nil {
			filter.ScheduledEnd = &scheduledEnd
		}
	}

	return filter, nil
}

func (h *JobHandler) parseScheduleFilter(r *http.Request) (*services.ScheduleFilter, error) {
	filter := &services.ScheduleFilter{}

	if userIDStr := r.URL.Query().Get("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			filter.UserID = &userID
		}
	}
	if crewIDStr := r.URL.Query().Get("crew_id"); crewIDStr != "" {
		if crewID, err := uuid.Parse(crewIDStr); err == nil {
			filter.CrewID = &crewID
		}
	}

	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filter.StartDate = &startDate
		}
	}
	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			filter.EndDate = &endDate
		}
	}

	return filter, nil
}

func (h *JobHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		h.logger.Error("Failed to marshal response", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

func (h *JobHandler) respondWithError(w http.ResponseWriter, code int, message string, err error) {
	errorResponse := domain.ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	}

	if err != nil {
		errorResponse.Details = map[string]interface{}{
			"error": err.Error(),
		}
	}

	h.respondWithJSON(w, code, errorResponse)
}