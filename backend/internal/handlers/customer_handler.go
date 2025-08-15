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

// CustomerHandler handles HTTP requests for customer operations
type CustomerHandler struct {
	customerService services.CustomerService
	logger          *log.Logger
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(customerService services.CustomerService, logger *log.Logger) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
		logger:          logger,
	}
}

// RegisterRoutes registers customer routes with the router
func (h *CustomerHandler) RegisterRoutes(router *mux.Router) {
	// Customer CRUD routes
	router.HandleFunc("/customers", h.CreateCustomer).Methods("POST")
	router.HandleFunc("/customers", h.ListCustomers).Methods("GET")
	router.HandleFunc("/customers/search", h.SearchCustomers).Methods("GET")
	router.HandleFunc("/customers/{id}", h.GetCustomer).Methods("GET")
	router.HandleFunc("/customers/{id}", h.UpdateCustomer).Methods("PUT")
	router.HandleFunc("/customers/{id}", h.DeleteCustomer).Methods("DELETE")
	
	// Customer-related data routes
	router.HandleFunc("/customers/{id}/properties", h.GetCustomerProperties).Methods("GET")
	router.HandleFunc("/customers/{id}/jobs", h.GetCustomerJobs).Methods("GET")
	router.HandleFunc("/customers/{id}/invoices", h.GetCustomerInvoices).Methods("GET")
	router.HandleFunc("/customers/{id}/quotes", h.GetCustomerQuotes).Methods("GET")
	router.HandleFunc("/customers/{id}/summary", h.GetCustomerSummary).Methods("GET")
}

// CreateCustomer creates a new customer
// @Summary Create a new customer
// @Description Create a new customer in the system
// @Tags customers
// @Accept json
// @Produce json
// @Param customer body domain.CreateCustomerRequest true "Customer creation request"
// @Success 201 {object} domain.EnhancedCustomer
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers [post]
func (h *CustomerHandler) CreateCustomer(w http.ResponseWriter, r *http.Request) {
	var req domain.CreateCustomerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	customer, err := h.customerService.CreateCustomer(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create customer", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create customer", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, customer)
}

// GetCustomer retrieves a customer by ID
// @Summary Get a customer by ID
// @Description Retrieve a customer by their unique ID
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} domain.EnhancedCustomer
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/{id} [get]
func (h *CustomerHandler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	customer, err := h.customerService.GetCustomer(r.Context(), customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			h.respondWithError(w, http.StatusNotFound, "Customer not found", nil)
			return
		}
		h.logger.Error("Failed to get customer", "error", err, "customer_id", customerID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get customer", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, customer)
}

// UpdateCustomer updates an existing customer
// @Summary Update a customer
// @Description Update an existing customer's information
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param customer body services.CustomerUpdateRequest true "Customer update request"
// @Success 200 {object} domain.EnhancedCustomer
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/{id} [put]
func (h *CustomerHandler) UpdateCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	var req services.CustomerUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	customer, err := h.customerService.UpdateCustomer(r.Context(), customerID, &req)
	if err != nil {
		if err.Error() == "customer not found" {
			h.respondWithError(w, http.StatusNotFound, "Customer not found", nil)
			return
		}
		h.logger.Error("Failed to update customer", "error", err, "customer_id", customerID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update customer", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, customer)
}

// DeleteCustomer deletes a customer
// @Summary Delete a customer
// @Description Delete a customer from the system
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/{id} [delete]
func (h *CustomerHandler) DeleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	err = h.customerService.DeleteCustomer(r.Context(), customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			h.respondWithError(w, http.StatusNotFound, "Customer not found", nil)
			return
		}
		h.logger.Error("Failed to delete customer", "error", err, "customer_id", customerID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete customer", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListCustomers lists customers with pagination and filtering
// @Summary List customers
// @Description Get a paginated list of customers with optional filtering
// @Tags customers
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param sort_by query string false "Sort field"
// @Param sort_desc query bool false "Sort descending"
// @Param search query string false "Search query"
// @Param customer_type query string false "Customer type filter"
// @Param status query string false "Status filter"
// @Param lead_source query string false "Lead source filter"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers [get]
func (h *CustomerHandler) ListCustomers(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseCustomerFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.customerService.ListCustomers(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list customers", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list customers", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// SearchCustomers searches customers by query string
// @Summary Search customers
// @Description Search customers by name, email, or company
// @Tags customers
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param customer_type query string false "Customer type filter"
// @Param status query string false "Status filter"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/search [get]
func (h *CustomerHandler) SearchCustomers(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		h.respondWithError(w, http.StatusBadRequest, "Search query is required", nil)
		return
	}

	filter, err := h.parseCustomerFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.customerService.SearchCustomers(r.Context(), query, filter)
	if err != nil {
		h.logger.Error("Failed to search customers", "error", err, "query", query)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to search customers", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetCustomerProperties gets properties for a customer
// @Summary Get customer properties
// @Description Retrieve all properties belonging to a customer
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {array} domain.EnhancedProperty
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/{id}/properties [get]
func (h *CustomerHandler) GetCustomerProperties(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	properties, err := h.customerService.GetCustomerProperties(r.Context(), customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			h.respondWithError(w, http.StatusNotFound, "Customer not found", nil)
			return
		}
		h.logger.Error("Failed to get customer properties", "error", err, "customer_id", customerID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get customer properties", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, properties)
}

// GetCustomerJobs gets jobs for a customer
// @Summary Get customer jobs
// @Description Retrieve jobs for a customer with pagination and filtering
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param status query string false "Job status filter"
// @Param priority query string false "Job priority filter"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/{id}/jobs [get]
func (h *CustomerHandler) GetCustomerJobs(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	jobFilter, err := h.parseJobFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.customerService.GetCustomerJobs(r.Context(), customerID, jobFilter)
	if err != nil {
		if err.Error() == "customer not found" {
			h.respondWithError(w, http.StatusNotFound, "Customer not found", nil)
			return
		}
		h.logger.Error("Failed to get customer jobs", "error", err, "customer_id", customerID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get customer jobs", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetCustomerInvoices gets invoices for a customer
// @Summary Get customer invoices
// @Description Retrieve invoices for a customer with pagination and filtering
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param status query string false "Invoice status filter"
// @Param overdue query bool false "Filter overdue invoices"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/{id}/invoices [get]
func (h *CustomerHandler) GetCustomerInvoices(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	invoiceFilter, err := h.parseInvoiceFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.customerService.GetCustomerInvoices(r.Context(), customerID, invoiceFilter)
	if err != nil {
		if err.Error() == "customer not found" {
			h.respondWithError(w, http.StatusNotFound, "Customer not found", nil)
			return
		}
		h.logger.Error("Failed to get customer invoices", "error", err, "customer_id", customerID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get customer invoices", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetCustomerQuotes gets quotes for a customer
// @Summary Get customer quotes
// @Description Retrieve quotes for a customer with pagination and filtering
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param status query string false "Quote status filter"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/{id}/quotes [get]
func (h *CustomerHandler) GetCustomerQuotes(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	quoteFilter, err := h.parseQuoteFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.customerService.GetCustomerQuotes(r.Context(), customerID, quoteFilter)
	if err != nil {
		if err.Error() == "customer not found" {
			h.respondWithError(w, http.StatusNotFound, "Customer not found", nil)
			return
		}
		h.logger.Error("Failed to get customer quotes", "error", err, "customer_id", customerID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get customer quotes", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GetCustomerSummary gets analytics summary for a customer
// @Summary Get customer summary
// @Description Retrieve analytics summary for a customer
// @Tags customers
// @Accept json
// @Produce json
// @Param id path string true "Customer ID"
// @Success 200 {object} services.CustomerSummary
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /customers/{id}/summary [get]
func (h *CustomerHandler) GetCustomerSummary(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	customerID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid customer ID", err)
		return
	}

	summary, err := h.customerService.GetCustomerSummary(r.Context(), customerID)
	if err != nil {
		if err.Error() == "customer not found" {
			h.respondWithError(w, http.StatusNotFound, "Customer not found", nil)
			return
		}
		h.logger.Error("Failed to get customer summary", "error", err, "customer_id", customerID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get customer summary", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, summary)
}

// Helper methods

func (h *CustomerHandler) parseCustomerFilter(r *http.Request) (*services.CustomerFilter, error) {
	filter := &services.CustomerFilter{}

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
	filter.CustomerType = r.URL.Query().Get("customer_type")
	filter.Status = r.URL.Query().Get("status")
	filter.LeadSource = r.URL.Query().Get("lead_source")

	return filter, nil
}

func (h *CustomerHandler) parseJobFilter(r *http.Request) (*services.JobFilter, error) {
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

func (h *CustomerHandler) parseInvoiceFilter(r *http.Request) (*services.InvoiceFilter, error) {
	filter := &services.InvoiceFilter{}

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
	if overdue := r.URL.Query().Get("overdue"); overdue != "" {
		if od, err := strconv.ParseBool(overdue); err == nil {
			filter.Overdue = od
		}
	}

	return filter, nil
}

func (h *CustomerHandler) parseQuoteFilter(r *http.Request) (*services.QuoteFilter, error) {
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

func (h *CustomerHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
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

func (h *CustomerHandler) respondWithError(w http.ResponseWriter, code int, message string, err error) {
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