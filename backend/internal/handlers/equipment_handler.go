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

// EquipmentHandler handles HTTP requests for equipment operations
type EquipmentHandler struct {
	equipmentService services.EquipmentService
	logger           *log.Logger
}

// NewEquipmentHandler creates a new equipment handler
func NewEquipmentHandler(equipmentService services.EquipmentService, logger *log.Logger) *EquipmentHandler {
	return &EquipmentHandler{
		equipmentService: equipmentService,
		logger:           logger,
	}
}

// RegisterRoutes registers equipment routes with the router
func (h *EquipmentHandler) RegisterRoutes(router *mux.Router) {
	// Equipment CRUD routes
	router.HandleFunc("/equipment", h.CreateEquipment).Methods("POST")
	router.HandleFunc("/equipment", h.ListEquipment).Methods("GET")
	router.HandleFunc("/equipment/{id}", h.GetEquipment).Methods("GET")
	router.HandleFunc("/equipment/{id}", h.UpdateEquipment).Methods("PUT")
	router.HandleFunc("/equipment/{id}", h.DeleteEquipment).Methods("DELETE")
	
	// Equipment availability routes
	router.HandleFunc("/equipment/available", h.GetAvailableEquipment).Methods("GET")
	router.HandleFunc("/equipment/{id}/availability", h.CheckEquipmentAvailability).Methods("GET")
	
	// Equipment maintenance routes
	router.HandleFunc("/equipment/{id}/schedule-maintenance", h.ScheduleMaintenance).Methods("POST")
	router.HandleFunc("/equipment/{id}/maintenance-history", h.GetMaintenanceHistory).Methods("GET")
	router.HandleFunc("/equipment/maintenance/upcoming", h.GetUpcomingMaintenance).Methods("GET")
	router.HandleFunc("/equipment/maintenance/due", h.CheckMaintenanceDue).Methods("GET")
	router.HandleFunc("/equipment/{id}/perform-maintenance", h.PerformMaintenance).Methods("POST")
}

// CreateEquipment creates a new equipment item
// @Summary Create a new equipment item
// @Description Create a new equipment item in the system
// @Tags equipment
// @Accept json
// @Produce json
// @Param equipment body services.EquipmentCreateRequest true "Equipment creation request"
// @Success 201 {object} domain.Equipment
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment [post]
func (h *EquipmentHandler) CreateEquipment(w http.ResponseWriter, r *http.Request) {
	var req services.EquipmentCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	equipment, err := h.equipmentService.CreateEquipment(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create equipment", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create equipment", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, equipment)
}

// GetEquipment retrieves equipment by ID
// @Summary Get equipment by ID
// @Description Retrieve equipment by its unique ID
// @Tags equipment
// @Accept json
// @Produce json
// @Param id path string true "Equipment ID"
// @Success 200 {object} domain.Equipment
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/{id} [get]
func (h *EquipmentHandler) GetEquipment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	equipmentID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid equipment ID", err)
		return
	}

	equipment, err := h.equipmentService.GetEquipment(r.Context(), equipmentID)
	if err != nil {
		if err.Error() == "equipment not found" {
			h.respondWithError(w, http.StatusNotFound, "Equipment not found", nil)
			return
		}
		h.logger.Error("Failed to get equipment", "error", err, "equipment_id", equipmentID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get equipment", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, equipment)
}

// UpdateEquipment updates an existing equipment item
// @Summary Update equipment
// @Description Update an existing equipment item's information
// @Tags equipment
// @Accept json
// @Produce json
// @Param id path string true "Equipment ID"
// @Param equipment body services.EquipmentUpdateRequest true "Equipment update request"
// @Success 200 {object} domain.Equipment
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/{id} [put]
func (h *EquipmentHandler) UpdateEquipment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	equipmentID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid equipment ID", err)
		return
	}

	var req services.EquipmentUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	equipment, err := h.equipmentService.UpdateEquipment(r.Context(), equipmentID, &req)
	if err != nil {
		if err.Error() == "equipment not found" {
			h.respondWithError(w, http.StatusNotFound, "Equipment not found", nil)
			return
		}
		h.logger.Error("Failed to update equipment", "error", err, "equipment_id", equipmentID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update equipment", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, equipment)
}

// DeleteEquipment deletes equipment
// @Summary Delete equipment
// @Description Delete equipment from the system (soft delete)
// @Tags equipment
// @Accept json
// @Produce json
// @Param id path string true "Equipment ID"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/{id} [delete]
func (h *EquipmentHandler) DeleteEquipment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	equipmentID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid equipment ID", err)
		return
	}

	err = h.equipmentService.DeleteEquipment(r.Context(), equipmentID)
	if err != nil {
		if err.Error() == "equipment not found" {
			h.respondWithError(w, http.StatusNotFound, "Equipment not found", nil)
			return
		}
		h.logger.Error("Failed to delete equipment", "error", err, "equipment_id", equipmentID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete equipment", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListEquipment lists equipment with pagination and filtering
// @Summary List equipment
// @Description Get a paginated list of equipment with optional filtering
// @Tags equipment
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param sort_by query string false "Sort field"
// @Param sort_desc query bool false "Sort descending"
// @Param status query string false "Equipment status filter"
// @Param type query string false "Equipment type filter"
// @Param search query string false "Search query"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment [get]
func (h *EquipmentHandler) ListEquipment(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseEquipmentFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.equipmentService.ListEquipment(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list equipment", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list equipment", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetAvailableEquipment gets available equipment for a time period
// @Summary Get available equipment
// @Description Retrieve equipment available for a specific time period
// @Tags equipment
// @Accept json
// @Produce json
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {array} domain.Equipment
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/available [get]
func (h *EquipmentHandler) GetAvailableEquipment(w http.ResponseWriter, r *http.Request) {
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

	equipment, err := h.equipmentService.GetAvailableEquipment(r.Context(), startDate, endDate)
	if err != nil {
		h.logger.Error("Failed to get available equipment", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get available equipment", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, equipment)
}

// CheckEquipmentAvailability checks if specific equipment is available
// @Summary Check equipment availability
// @Description Check if specific equipment is available for a time period
// @Tags equipment
// @Accept json
// @Produce json
// @Param id path string true "Equipment ID"
// @Param start_date query string true "Start date (YYYY-MM-DD)"
// @Param end_date query string true "End date (YYYY-MM-DD)"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/{id}/availability [get]
func (h *EquipmentHandler) CheckEquipmentAvailability(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	equipmentID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid equipment ID", err)
		return
	}

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

	available, err := h.equipmentService.CheckEquipmentAvailability(r.Context(), equipmentID, startDate, endDate)
	if err != nil {
		if err.Error() == "equipment not found" {
			h.respondWithError(w, http.StatusNotFound, "Equipment not found", nil)
			return
		}
		h.logger.Error("Failed to check equipment availability", "error", err, "equipment_id", equipmentID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to check equipment availability", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]bool{"available": available})
}

// ScheduleMaintenance schedules maintenance for equipment
// @Summary Schedule equipment maintenance
// @Description Schedule maintenance for a specific equipment item
// @Tags equipment
// @Accept json
// @Produce json
// @Param id path string true "Equipment ID"
// @Param request body services.MaintenanceScheduleRequest true "Maintenance schedule request"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/{id}/schedule-maintenance [post]
func (h *EquipmentHandler) ScheduleMaintenance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	equipmentID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid equipment ID", err)
		return
	}

	var req services.MaintenanceScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err = h.equipmentService.ScheduleMaintenance(r.Context(), equipmentID, &req)
	if err != nil {
		if err.Error() == "equipment not found" {
			h.respondWithError(w, http.StatusNotFound, "Equipment not found", nil)
			return
		}
		h.logger.Error("Failed to schedule maintenance", "error", err, "equipment_id", equipmentID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to schedule maintenance", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Maintenance scheduled successfully"})
}

// GetMaintenanceHistory gets maintenance history for equipment
// @Summary Get equipment maintenance history
// @Description Retrieve maintenance history for a specific equipment item
// @Tags equipment
// @Accept json
// @Produce json
// @Param id path string true "Equipment ID"
// @Success 200 {array} services.MaintenanceRecord
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/{id}/maintenance-history [get]
func (h *EquipmentHandler) GetMaintenanceHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	equipmentID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid equipment ID", err)
		return
	}

	history, err := h.equipmentService.GetMaintenanceHistory(r.Context(), equipmentID)
	if err != nil {
		if err.Error() == "equipment not found" {
			h.respondWithError(w, http.StatusNotFound, "Equipment not found", nil)
			return
		}
		h.logger.Error("Failed to get maintenance history", "error", err, "equipment_id", equipmentID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get maintenance history", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, history)
}

// GetUpcomingMaintenance gets upcoming maintenance schedules
// @Summary Get upcoming maintenance schedules
// @Description Retrieve upcoming maintenance schedules for all equipment
// @Tags equipment
// @Accept json
// @Produce json
// @Success 200 {array} services.MaintenanceSchedule
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/maintenance/upcoming [get]
func (h *EquipmentHandler) GetUpcomingMaintenance(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.equipmentService.GetUpcomingMaintenance(r.Context())
	if err != nil {
		h.logger.Error("Failed to get upcoming maintenance", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get upcoming maintenance", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, schedules)
}

// CheckMaintenanceDue checks for equipment with maintenance due
// @Summary Check equipment maintenance due
// @Description Check for equipment that has maintenance due or overdue
// @Tags equipment
// @Accept json
// @Produce json
// @Success 200 {array} domain.Equipment
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/maintenance/due [get]
func (h *EquipmentHandler) CheckMaintenanceDue(w http.ResponseWriter, r *http.Request) {
	equipment, err := h.equipmentService.CheckMaintenanceDue(r.Context())
	if err != nil {
		h.logger.Error("Failed to check maintenance due", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to check maintenance due", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, equipment)
}

// PerformMaintenance records completed maintenance
// @Summary Perform equipment maintenance
// @Description Record completion of maintenance for equipment
// @Tags equipment
// @Accept json
// @Produce json
// @Param id path string true "Equipment ID"
// @Param request body map[string]interface{} true "Maintenance completion details"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /equipment/{id}/perform-maintenance [post]
func (h *EquipmentHandler) PerformMaintenance(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	equipmentID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid equipment ID", err)
		return
	}

	var request map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	maintenanceType, ok := request["maintenance_type"].(string)
	if !ok || maintenanceType == "" {
		h.respondWithError(w, http.StatusBadRequest, "Maintenance type is required", nil)
		return
	}

	var cost *float64
	if costValue, exists := request["cost"]; exists {
		if costFloat, ok := costValue.(float64); ok {
			cost = &costFloat
		}
	}

	var notes *string
	if notesValue, exists := request["notes"]; exists {
		if notesStr, ok := notesValue.(string); ok {
			notes = &notesStr
		}
	}

	err = h.equipmentService.PerformMaintenance(r.Context(), equipmentID, maintenanceType, cost, notes)
	if err != nil {
		if err.Error() == "equipment not found" {
			h.respondWithError(w, http.StatusNotFound, "Equipment not found", nil)
			return
		}
		h.logger.Error("Failed to perform maintenance", "error", err, "equipment_id", equipmentID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to perform maintenance", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Maintenance performed successfully"})
}

// Helper methods

func (h *EquipmentHandler) parseEquipmentFilter(r *http.Request) (*services.EquipmentFilter, error) {
	filter := &services.EquipmentFilter{}

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
	filter.Type = r.URL.Query().Get("type")
	filter.Search = r.URL.Query().Get("search")

	return filter, nil
}

func (h *EquipmentHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
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

func (h *EquipmentHandler) respondWithError(w http.ResponseWriter, code int, message string, err error) {
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