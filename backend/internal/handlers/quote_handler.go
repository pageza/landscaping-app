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

// QuoteHandler handles HTTP requests for quote operations
type QuoteHandler struct {
	quoteService services.QuoteService
	logger       *log.Logger
}

// NewQuoteHandler creates a new quote handler
func NewQuoteHandler(quoteService services.QuoteService, logger *log.Logger) *QuoteHandler {
	return &QuoteHandler{
		quoteService: quoteService,
		logger:       logger,
	}
}

// RegisterRoutes registers quote routes with the router
func (h *QuoteHandler) RegisterRoutes(router *mux.Router) {
	// Quote CRUD routes
	router.HandleFunc("/quotes", h.CreateQuote).Methods("POST")
	router.HandleFunc("/quotes", h.ListQuotes).Methods("GET")
	router.HandleFunc("/quotes/{id}", h.GetQuote).Methods("GET")
	router.HandleFunc("/quotes/{id}", h.UpdateQuote).Methods("PUT")
	router.HandleFunc("/quotes/{id}", h.DeleteQuote).Methods("DELETE")
	
	// Quote lifecycle routes
	router.HandleFunc("/quotes/{id}/approve", h.ApproveQuote).Methods("POST")
	router.HandleFunc("/quotes/{id}/reject", h.RejectQuote).Methods("POST")
	router.HandleFunc("/quotes/{id}/convert", h.ConvertQuoteToJob).Methods("POST")
	
	// Quote document and communication routes
	router.HandleFunc("/quotes/{id}/pdf", h.GenerateQuotePDF).Methods("GET")
	router.HandleFunc("/quotes/{id}/send", h.SendQuote).Methods("POST")
	
	// AI-powered quote generation
	router.HandleFunc("/quotes/generate", h.GenerateQuoteFromDescription).Methods("POST")
}

// CreateQuote creates a new quote
// @Summary Create a new quote
// @Description Create a new quote in the system
// @Tags quotes
// @Accept json
// @Produce json
// @Param quote body services.QuoteCreateRequest true "Quote creation request"
// @Success 201 {object} domain.Quote
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes [post]
func (h *QuoteHandler) CreateQuote(w http.ResponseWriter, r *http.Request) {
	var req services.QuoteCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	quote, err := h.quoteService.CreateQuote(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create quote", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create quote", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, quote)
}

// GetQuote retrieves a quote by ID
// @Summary Get a quote by ID
// @Description Retrieve a quote by its unique ID
// @Tags quotes
// @Accept json
// @Produce json
// @Param id path string true "Quote ID"
// @Success 200 {object} domain.Quote
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes/{id} [get]
func (h *QuoteHandler) GetQuote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quoteID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid quote ID", err)
		return
	}

	quote, err := h.quoteService.GetQuote(r.Context(), quoteID)
	if err != nil {
		if err.Error() == "quote not found" {
			h.respondWithError(w, http.StatusNotFound, "Quote not found", nil)
			return
		}
		h.logger.Error("Failed to get quote", "error", err, "quote_id", quoteID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get quote", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, quote)
}

// UpdateQuote updates an existing quote
// @Summary Update a quote
// @Description Update an existing quote's information
// @Tags quotes
// @Accept json
// @Produce json
// @Param id path string true "Quote ID"
// @Param quote body services.QuoteUpdateRequest true "Quote update request"
// @Success 200 {object} domain.Quote
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes/{id} [put]
func (h *QuoteHandler) UpdateQuote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quoteID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid quote ID", err)
		return
	}

	var req services.QuoteUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	quote, err := h.quoteService.UpdateQuote(r.Context(), quoteID, &req)
	if err != nil {
		if err.Error() == "quote not found" {
			h.respondWithError(w, http.StatusNotFound, "Quote not found", nil)
			return
		}
		h.logger.Error("Failed to update quote", "error", err, "quote_id", quoteID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update quote", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, quote)
}

// DeleteQuote deletes a quote
// @Summary Delete a quote
// @Description Delete a quote from the system
// @Tags quotes
// @Accept json
// @Produce json
// @Param id path string true "Quote ID"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes/{id} [delete]
func (h *QuoteHandler) DeleteQuote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quoteID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid quote ID", err)
		return
	}

	err = h.quoteService.DeleteQuote(r.Context(), quoteID)
	if err != nil {
		if err.Error() == "quote not found" {
			h.respondWithError(w, http.StatusNotFound, "Quote not found", nil)
			return
		}
		h.logger.Error("Failed to delete quote", "error", err, "quote_id", quoteID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete quote", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListQuotes lists quotes with pagination and filtering
// @Summary List quotes
// @Description Get a paginated list of quotes with optional filtering
// @Tags quotes
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param sort_by query string false "Sort field"
// @Param sort_desc query bool false "Sort descending"
// @Param status query string false "Quote status filter"
// @Param customer_id query string false "Customer ID filter"
// @Param property_id query string false "Property ID filter"
// @Param search query string false "Search query"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes [get]
func (h *QuoteHandler) ListQuotes(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseQuoteFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.quoteService.ListQuotes(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list quotes", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list quotes", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// ApproveQuote approves a quote
// @Summary Approve a quote
// @Description Approve a quote for conversion to a job
// @Tags quotes
// @Accept json
// @Produce json
// @Param id path string true "Quote ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes/{id}/approve [post]
func (h *QuoteHandler) ApproveQuote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quoteID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid quote ID", err)
		return
	}

	err = h.quoteService.ApproveQuote(r.Context(), quoteID)
	if err != nil {
		if err.Error() == "quote not found" {
			h.respondWithError(w, http.StatusNotFound, "Quote not found", nil)
			return
		}
		h.logger.Error("Failed to approve quote", "error", err, "quote_id", quoteID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to approve quote", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Quote approved successfully"})
}

// RejectQuote rejects a quote
// @Summary Reject a quote
// @Description Reject a quote with a reason
// @Tags quotes
// @Accept json
// @Produce json
// @Param id path string true "Quote ID"
// @Param request body map[string]string true "Rejection request with reason"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes/{id}/reject [post]
func (h *QuoteHandler) RejectQuote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quoteID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid quote ID", err)
		return
	}

	var request map[string]string
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	reason := request["reason"]
	if reason == "" {
		h.respondWithError(w, http.StatusBadRequest, "Rejection reason is required", nil)
		return
	}

	err = h.quoteService.RejectQuote(r.Context(), quoteID, reason)
	if err != nil {
		if err.Error() == "quote not found" {
			h.respondWithError(w, http.StatusNotFound, "Quote not found", nil)
			return
		}
		h.logger.Error("Failed to reject quote", "error", err, "quote_id", quoteID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to reject quote", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Quote rejected successfully"})
}

// ConvertQuoteToJob converts a quote to a job
// @Summary Convert quote to job
// @Description Convert an approved quote to a job
// @Tags quotes
// @Accept json
// @Produce json
// @Param id path string true "Quote ID"
// @Success 201 {object} domain.EnhancedJob
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes/{id}/convert [post]
func (h *QuoteHandler) ConvertQuoteToJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quoteID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid quote ID", err)
		return
	}

	job, err := h.quoteService.ConvertQuoteToJob(r.Context(), quoteID)
	if err != nil {
		if err.Error() == "quote not found" {
			h.respondWithError(w, http.StatusNotFound, "Quote not found", nil)
			return
		}
		h.logger.Error("Failed to convert quote to job", "error", err, "quote_id", quoteID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to convert quote to job", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, job)
}

// GenerateQuotePDF generates a PDF for a quote
// @Summary Generate quote PDF
// @Description Generate a PDF document for a quote
// @Tags quotes
// @Accept json
// @Produce application/pdf
// @Param id path string true "Quote ID"
// @Success 200 {file} application/pdf
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes/{id}/pdf [get]
func (h *QuoteHandler) GenerateQuotePDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quoteID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid quote ID", err)
		return
	}

	pdfData, err := h.quoteService.GenerateQuotePDF(r.Context(), quoteID)
	if err != nil {
		if err.Error() == "quote not found" {
			h.respondWithError(w, http.StatusNotFound, "Quote not found", nil)
			return
		}
		h.logger.Error("Failed to generate quote PDF", "error", err, "quote_id", quoteID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to generate quote PDF", err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=quote.pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfData)
}

// SendQuote sends a quote to the customer
// @Summary Send quote
// @Description Send a quote to the customer via email
// @Tags quotes
// @Accept json
// @Produce json
// @Param id path string true "Quote ID"
// @Param options body services.QuoteSendOptions true "Quote send options"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes/{id}/send [post]
func (h *QuoteHandler) SendQuote(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	quoteID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid quote ID", err)
		return
	}

	var options services.QuoteSendOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err = h.quoteService.SendQuote(r.Context(), quoteID, &options)
	if err != nil {
		if err.Error() == "quote not found" {
			h.respondWithError(w, http.StatusNotFound, "Quote not found", nil)
			return
		}
		h.logger.Error("Failed to send quote", "error", err, "quote_id", quoteID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to send quote", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Quote sent successfully"})
}

// GenerateQuoteFromDescription generates a quote from AI description
// @Summary Generate quote from description
// @Description Generate a quote using AI from a description
// @Tags quotes
// @Accept json
// @Produce json
// @Param request body services.QuoteGenerationRequest true "Quote generation request"
// @Success 201 {object} domain.Quote
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /quotes/generate [post]
func (h *QuoteHandler) GenerateQuoteFromDescription(w http.ResponseWriter, r *http.Request) {
	var req services.QuoteGenerationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	quote, err := h.quoteService.GenerateQuoteFromDescription(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to generate quote from description", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to generate quote from description", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, quote)
}

// Helper methods

func (h *QuoteHandler) parseQuoteFilter(r *http.Request) (*services.QuoteFilter, error) {
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

	// Parse sorting
	filter.SortBy = r.URL.Query().Get("sort_by")
	if sortDesc := r.URL.Query().Get("sort_desc"); sortDesc != "" {
		if sd, err := strconv.ParseBool(sortDesc); err == nil {
			filter.SortDesc = sd
		}
	}

	// Parse filters
	filter.Status = r.URL.Query().Get("status")
	filter.Search = r.URL.Query().Get("search")

	// Parse UUID filters
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

	return filter, nil
}

func (h *QuoteHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
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

func (h *QuoteHandler) respondWithError(w http.ResponseWriter, code int, message string, err error) {
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