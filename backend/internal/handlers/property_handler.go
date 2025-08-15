package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/google/uuid"
	"log"

	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// PropertyHandler handles HTTP requests for property operations
type PropertyHandler struct {
	propertyService services.PropertyService
	logger          *log.Logger
}

// NewPropertyHandler creates a new property handler
func NewPropertyHandler(propertyService services.PropertyService, logger *log.Logger) *PropertyHandler {
	return &PropertyHandler{
		propertyService: propertyService,
		logger:          logger,
	}
}

// RegisterRoutes registers property routes with the router
func (h *PropertyHandler) RegisterRoutes(router *mux.Router) {
	// Property CRUD routes
	router.HandleFunc("/properties", h.CreateProperty).Methods("POST")
	router.HandleFunc("/properties", h.ListProperties).Methods("GET")
	router.HandleFunc("/properties/search", h.SearchProperties).Methods("GET")
	router.HandleFunc("/properties/nearby", h.GetNearbyProperties).Methods("GET")
	router.HandleFunc("/properties/{id}", h.GetProperty).Methods("GET")
	router.HandleFunc("/properties/{id}", h.UpdateProperty).Methods("PUT")
	router.HandleFunc("/properties/{id}", h.DeleteProperty).Methods("DELETE")
	
	// Property-related data routes
	router.HandleFunc("/properties/{id}/jobs", h.GetPropertyJobs).Methods("GET")
	router.HandleFunc("/properties/{id}/quotes", h.GetPropertyQuotes).Methods("GET")
	router.HandleFunc("/properties/{id}/value", h.GetPropertyValue).Methods("GET")
}

// CreateProperty creates a new property
// @Summary Create a new property
// @Description Create a new property in the system
// @Tags properties
// @Accept json
// @Produce json
// @Param property body domain.CreatePropertyRequest true "Property creation request"
// @Success 201 {object} domain.EnhancedProperty
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties [post]
func (h *PropertyHandler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	var req domain.CreatePropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	property, err := h.propertyService.CreateProperty(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create property", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create property", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, property)
}

// GetProperty retrieves a property by ID
// @Summary Get a property by ID
// @Description Retrieve a property by its unique ID
// @Tags properties
// @Accept json
// @Produce json
// @Param id path string true "Property ID"
// @Success 200 {object} domain.EnhancedProperty
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties/{id} [get]
func (h *PropertyHandler) GetProperty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	propertyID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid property ID", err)
		return
	}

	property, err := h.propertyService.GetProperty(r.Context(), propertyID)
	if err != nil {
		if err.Error() == "property not found" {
			h.respondWithError(w, http.StatusNotFound, "Property not found", nil)
			return
		}
		h.logger.Error("Failed to get property", "error", err, "property_id", propertyID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get property", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, property)
}

// UpdateProperty updates an existing property
// @Summary Update a property
// @Description Update an existing property's information
// @Tags properties
// @Accept json
// @Produce json
// @Param id path string true "Property ID"
// @Param property body services.PropertyUpdateRequest true "Property update request"
// @Success 200 {object} domain.EnhancedProperty
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties/{id} [put]
func (h *PropertyHandler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	propertyID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid property ID", err)
		return
	}

	var req services.PropertyUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	property, err := h.propertyService.UpdateProperty(r.Context(), propertyID, &req)
	if err != nil {
		if err.Error() == "property not found" {
			h.respondWithError(w, http.StatusNotFound, "Property not found", nil)
			return
		}
		h.logger.Error("Failed to update property", "error", err, "property_id", propertyID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update property", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, property)
}

// DeleteProperty deletes a property
// @Summary Delete a property
// @Description Delete a property from the system
// @Tags properties
// @Accept json
// @Produce json
// @Param id path string true "Property ID"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties/{id} [delete]
func (h *PropertyHandler) DeleteProperty(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	propertyID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid property ID", err)
		return
	}

	err = h.propertyService.DeleteProperty(r.Context(), propertyID)
	if err != nil {
		if err.Error() == "property not found" {
			h.respondWithError(w, http.StatusNotFound, "Property not found", nil)
			return
		}
		h.logger.Error("Failed to delete property", "error", err, "property_id", propertyID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete property", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListProperties lists properties with pagination and filtering
// @Summary List properties
// @Description Get a paginated list of properties with optional filtering
// @Tags properties
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param sort_by query string false "Sort field"
// @Param sort_desc query bool false "Sort descending"
// @Param search query string false "Search query"
// @Param property_type query string false "Property type filter"
// @Param customer_id query string false "Customer ID filter"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties [get]
func (h *PropertyHandler) ListProperties(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parsePropertyFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.propertyService.ListProperties(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list properties", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list properties", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// SearchProperties searches properties by query string
// @Summary Search properties
// @Description Search properties by address, description, or notes
// @Tags properties
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param property_type query string false "Property type filter"
// @Param customer_id query string false "Customer ID filter"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties/search [get]
func (h *PropertyHandler) SearchProperties(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.respondWithError(w, http.StatusBadRequest, "Search query is required", nil)
		return
	}

	filter, err := h.parsePropertyFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.propertyService.SearchProperties(r.Context(), query, filter)
	if err != nil {
		h.logger.Error("Failed to search properties", "error", err, "query", query)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search properties", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetNearbyProperties gets nearby properties
// @Summary Get nearby properties
// @Description Retrieve properties near a specific location
// @Tags properties
// @Accept json
// @Produce json
// @Param lat query float64 true "Latitude"
// @Param lng query float64 true "Longitude"
// @Param radius query float64 false "Search radius in miles" default(5)
// @Success 200 {array} domain.EnhancedProperty
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties/nearby [get]
func (h *PropertyHandler) GetNearbyProperties(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	radiusStr := r.URL.Query().Get("radius")

	if latStr == "" || lngStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Latitude and longitude are required", nil)
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid latitude", err)
		return
	}

	lng, err := strconv.ParseFloat(lngStr, 64)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid longitude", err)
		return
	}

	radius := 5.0 // Default radius
	if radiusStr != "" {
		if r, err := strconv.ParseFloat(radiusStr, 64); err == nil && r > 0 {
			radius = r
		}
	}

	properties, err := h.propertyService.GetNearbyProperties(r.Context(), lat, lng, radius)
	if err != nil {
		h.logger.Error("Failed to get nearby properties", "error", err, "lat", lat, "lng", lng, "radius", radius)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get nearby properties", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, properties)
}

// GetPropertyJobs gets jobs for a property
// @Summary Get property jobs
// @Description Retrieve jobs for a property with pagination and filtering
// @Tags properties
// @Accept json
// @Produce json
// @Param id path string true "Property ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param status query string false "Job status filter"
// @Param priority query string false "Job priority filter"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties/{id}/jobs [get]
func (h *PropertyHandler) GetPropertyJobs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	propertyID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid property ID", err)
		return
	}

	jobFilter, err := h.parseJobFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.propertyService.GetPropertyJobs(r.Context(), propertyID, jobFilter)
	if err != nil {
		if err.Error() == "property not found" {
			h.respondWithError(w, http.StatusNotFound, "Property not found", nil)
			return
		}
		h.logger.Error("Failed to get property jobs", "error", err, "property_id", propertyID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get property jobs", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetPropertyQuotes gets quotes for a property
// @Summary Get property quotes
// @Description Retrieve quotes for a property with pagination and filtering
// @Tags properties
// @Accept json
// @Produce json
// @Param id path string true "Property ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param status query string false "Quote status filter"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties/{id}/quotes [get]
func (h *PropertyHandler) GetPropertyQuotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	propertyID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid property ID", err)
		return
	}

	quoteFilter, err := h.parseQuoteFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.propertyService.GetPropertyQuotes(r.Context(), propertyID, quoteFilter)
	if err != nil {
		if err.Error() == "property not found" {
			h.respondWithError(w, http.StatusNotFound, "Property not found", nil)
			return
		}
		h.logger.Error("Failed to get property quotes", "error", err, "property_id", propertyID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get property quotes", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetPropertyValue gets property valuation
// @Summary Get property value estimation
// @Description Retrieve estimated value for a property
// @Tags properties
// @Accept json
// @Produce json
// @Param id path string true "Property ID"
// @Success 200 {object} services.PropertyValuation
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /properties/{id}/value [get]
func (h *PropertyHandler) GetPropertyValue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	propertyID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid property ID", err)
		return
	}

	valuation, err := h.propertyService.GetPropertyValue(r.Context(), propertyID)
	if err != nil {
		if err.Error() == "property not found" {
			h.respondWithError(w, http.StatusNotFound, "Property not found", nil)
			return
		}
		h.logger.Error("Failed to get property value", "error", err, "property_id", propertyID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get property value", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, valuation)
}

// Helper methods

func (h *PropertyHandler) parsePropertyFilter(r *http.Request) (*services.PropertyFilter, error) {
	filter := &services.PropertyFilter{}

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
	filter.Search = r.URL.Query().Get("search")
	filter.PropertyType = r.URL.Query().Get("property_type")
	
	if customerIDStr := r.URL.Query().Get("customer_id"); customerIDStr != "" {
		if customerID, err := uuid.Parse(customerIDStr); err == nil {
			filter.CustomerID = &customerID
		}
	}

	return filter, nil
}

func (h *PropertyHandler) parseJobFilter(r *http.Request) (*services.JobFilter, error) {
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

	// Parse filters
	filter.Status = r.URL.Query().Get("status")
	filter.Priority = r.URL.Query().Get("priority")

	return filter, nil
}

func (h *PropertyHandler) parseQuoteFilter(r *http.Request) (*services.QuoteFilter, error) {
	filter := &services.QuoteFilter{}

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

	// Parse filters
	filter.Status = r.URL.Query().Get("status")

	return filter, nil
}

func (h *PropertyHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
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

func (h *PropertyHandler) respondWithError(w http.ResponseWriter, code int, message string, err error) {
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