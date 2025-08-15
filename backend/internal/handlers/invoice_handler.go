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

// InvoiceHandler handles HTTP requests for invoice operations
type InvoiceHandler struct {
	invoiceService services.InvoiceService
	logger         *log.Logger
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(invoiceService services.InvoiceService, logger *log.Logger) *InvoiceHandler {
	return &InvoiceHandler{
		invoiceService: invoiceService,
		logger:         logger,
	}
}

// RegisterRoutes registers invoice routes with the router
func (h *InvoiceHandler) RegisterRoutes(router *mux.Router) {
	// Invoice CRUD routes
	router.HandleFunc("/invoices", h.CreateInvoice).Methods("POST")
	router.HandleFunc("/invoices", h.ListInvoices).Methods("GET")
	router.HandleFunc("/invoices/{id}", h.GetInvoice).Methods("GET")
	router.HandleFunc("/invoices/{id}", h.UpdateInvoice).Methods("PUT")
	router.HandleFunc("/invoices/{id}", h.DeleteInvoice).Methods("DELETE")
	
	// Invoice document and communication routes
	router.HandleFunc("/invoices/{id}/pdf", h.GenerateInvoicePDF).Methods("GET")
	router.HandleFunc("/invoices/{id}/send", h.SendInvoice).Methods("POST")
	
	// Payment tracking routes
	router.HandleFunc("/invoices/{id}/payments", h.GetInvoicePayments).Methods("GET")
	router.HandleFunc("/invoices/{id}/mark-paid", h.MarkInvoiceAsPaid).Methods("POST")
	
	// Overdue management routes
	router.HandleFunc("/invoices/overdue", h.GetOverdueInvoices).Methods("GET")
	router.HandleFunc("/invoices/send-reminders", h.SendOverdueReminders).Methods("POST")
	
	// Automation routes
	router.HandleFunc("/invoices/from-job/{job_id}", h.CreateInvoiceFromJob).Methods("POST")
}

// CreateInvoice creates a new invoice
// @Summary Create a new invoice
// @Description Create a new invoice in the system
// @Tags invoices
// @Accept json
// @Produce json
// @Param invoice body services.InvoiceCreateRequest true "Invoice creation request"
// @Success 201 {object} domain.Invoice
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices [post]
func (h *InvoiceHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req services.InvoiceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	invoice, err := h.invoiceService.CreateInvoice(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to create invoice", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create invoice", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, invoice)
}

// GetInvoice retrieves an invoice by ID
// @Summary Get an invoice by ID
// @Description Retrieve an invoice by its unique ID
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {object} domain.Invoice
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/{id} [get]
func (h *InvoiceHandler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid invoice ID", err)
		return
	}

	invoice, err := h.invoiceService.GetInvoice(r.Context(), invoiceID)
	if err != nil {
		if err.Error() == "invoice not found" {
			h.respondWithError(w, http.StatusNotFound, "Invoice not found", nil)
			return
		}
		h.logger.Error("Failed to get invoice", "error", err, "invoice_id", invoiceID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get invoice", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, invoice)
}

// UpdateInvoice updates an existing invoice
// @Summary Update an invoice
// @Description Update an existing invoice's information
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param invoice body services.InvoiceUpdateRequest true "Invoice update request"
// @Success 200 {object} domain.Invoice
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/{id} [put]
func (h *InvoiceHandler) UpdateInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid invoice ID", err)
		return
	}

	var req services.InvoiceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	invoice, err := h.invoiceService.UpdateInvoice(r.Context(), invoiceID, &req)
	if err != nil {
		if err.Error() == "invoice not found" {
			h.respondWithError(w, http.StatusNotFound, "Invoice not found", nil)
			return
		}
		h.logger.Error("Failed to update invoice", "error", err, "invoice_id", invoiceID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to update invoice", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, invoice)
}

// DeleteInvoice deletes an invoice
// @Summary Delete an invoice
// @Description Delete an invoice from the system
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/{id} [delete]
func (h *InvoiceHandler) DeleteInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid invoice ID", err)
		return
	}

	err = h.invoiceService.DeleteInvoice(r.Context(), invoiceID)
	if err != nil {
		if err.Error() == "invoice not found" {
			h.respondWithError(w, http.StatusNotFound, "Invoice not found", nil)
			return
		}
		h.logger.Error("Failed to delete invoice", "error", err, "invoice_id", invoiceID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to delete invoice", err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListInvoices lists invoices with pagination and filtering
// @Summary List invoices
// @Description Get a paginated list of invoices with optional filtering
// @Tags invoices
// @Accept json
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(50)
// @Param sort_by query string false "Sort field"
// @Param sort_desc query bool false "Sort descending"
// @Param status query string false "Invoice status filter"
// @Param customer_id query string false "Customer ID filter"
// @Param overdue query bool false "Filter overdue invoices"
// @Param search query string false "Search query"
// @Success 200 {object} domain.PaginatedResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices [get]
func (h *InvoiceHandler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	filter, err := h.parseInvoiceFilter(r)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid filter parameters", err)
		return
	}

	response, err := h.invoiceService.ListInvoices(r.Context(), filter)
	if err != nil {
		h.logger.Error("Failed to list invoices", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to list invoices", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, response)
}

// GenerateInvoicePDF generates a PDF for an invoice
// @Summary Generate invoice PDF
// @Description Generate a PDF document for an invoice
// @Tags invoices
// @Accept json
// @Produce application/pdf
// @Param id path string true "Invoice ID"
// @Success 200 {file} application/pdf
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/{id}/pdf [get]
func (h *InvoiceHandler) GenerateInvoicePDF(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid invoice ID", err)
		return
	}

	pdfData, err := h.invoiceService.GenerateInvoicePDF(r.Context(), invoiceID)
	if err != nil {
		if err.Error() == "invoice not found" {
			h.respondWithError(w, http.StatusNotFound, "Invoice not found", nil)
			return
		}
		h.logger.Error("Failed to generate invoice PDF", "error", err, "invoice_id", invoiceID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to generate invoice PDF", err)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=invoice.pdf")
	w.WriteHeader(http.StatusOK)
	w.Write(pdfData)
}

// SendInvoice sends an invoice to the customer
// @Summary Send invoice
// @Description Send an invoice to the customer via email
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param options body services.InvoiceSendOptions true "Invoice send options"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/{id}/send [post]
func (h *InvoiceHandler) SendInvoice(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid invoice ID", err)
		return
	}

	var options services.InvoiceSendOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	err = h.invoiceService.SendInvoice(r.Context(), invoiceID, &options)
	if err != nil {
		if err.Error() == "invoice not found" {
			h.respondWithError(w, http.StatusNotFound, "Invoice not found", nil)
			return
		}
		h.logger.Error("Failed to send invoice", "error", err, "invoice_id", invoiceID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to send invoice", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Invoice sent successfully"})
}

// GetInvoicePayments gets payments for an invoice
// @Summary Get invoice payments
// @Description Retrieve payments associated with an invoice
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Success 200 {array} domain.Payment
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/{id}/payments [get]
func (h *InvoiceHandler) GetInvoicePayments(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid invoice ID", err)
		return
	}

	payments, err := h.invoiceService.GetInvoicePayments(r.Context(), invoiceID)
	if err != nil {
		if err.Error() == "invoice not found" {
			h.respondWithError(w, http.StatusNotFound, "Invoice not found", nil)
			return
		}
		h.logger.Error("Failed to get invoice payments", "error", err, "invoice_id", invoiceID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get invoice payments", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, payments)
}

// MarkInvoiceAsPaid marks an invoice as paid
// @Summary Mark invoice as paid
// @Description Mark an invoice as paid with a payment reference
// @Tags invoices
// @Accept json
// @Produce json
// @Param id path string true "Invoice ID"
// @Param payment body map[string]string true "Payment reference"
// @Success 200 {object} map[string]string
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/{id}/mark-paid [post]
func (h *InvoiceHandler) MarkInvoiceAsPaid(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	invoiceID, err := uuid.Parse(vars["id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid invoice ID", err)
		return
	}

	var payment map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payment); err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	paymentIDStr := payment["payment_id"]
	if paymentIDStr == "" {
		h.respondWithError(w, http.StatusBadRequest, "Payment ID is required", nil)
		return
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid payment ID", err)
		return
	}

	err = h.invoiceService.MarkInvoiceAsPaid(r.Context(), invoiceID, paymentID)
	if err != nil {
		if err.Error() == "invoice not found" {
			h.respondWithError(w, http.StatusNotFound, "Invoice not found", nil)
			return
		}
		h.logger.Error("Failed to mark invoice as paid", "error", err, "invoice_id", invoiceID, "payment_id", paymentID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to mark invoice as paid", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Invoice marked as paid successfully"})
}

// GetOverdueInvoices gets overdue invoices
// @Summary Get overdue invoices
// @Description Retrieve all overdue invoices
// @Tags invoices
// @Accept json
// @Produce json
// @Success 200 {array} domain.Invoice
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/overdue [get]
func (h *InvoiceHandler) GetOverdueInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.invoiceService.GetOverdueInvoices(r.Context())
	if err != nil {
		h.logger.Error("Failed to get overdue invoices", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to get overdue invoices", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, invoices)
}

// SendOverdueReminders sends reminders for overdue invoices
// @Summary Send overdue reminders
// @Description Send reminder emails for all overdue invoices
// @Tags invoices
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/send-reminders [post]
func (h *InvoiceHandler) SendOverdueReminders(w http.ResponseWriter, r *http.Request) {
	err := h.invoiceService.SendOverdueReminders(r.Context())
	if err != nil {
		h.logger.Error("Failed to send overdue reminders", "error", err)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to send overdue reminders", err)
		return
	}

	h.respondWithJSON(w, http.StatusOK, map[string]string{"message": "Overdue reminders sent successfully"})
}

// CreateInvoiceFromJob creates an invoice from a job
// @Summary Create invoice from job
// @Description Create an invoice automatically from a completed job
// @Tags invoices
// @Accept json
// @Produce json
// @Param job_id path string true "Job ID"
// @Success 201 {object} domain.Invoice
// @Failure 400 {object} domain.ErrorResponse
// @Failure 404 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /invoices/from-job/{job_id} [post]
func (h *InvoiceHandler) CreateInvoiceFromJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	jobID, err := uuid.Parse(vars["job_id"])
	if err != nil {
		h.respondWithError(w, http.StatusBadRequest, "Invalid job ID", err)
		return
	}

	invoice, err := h.invoiceService.CreateInvoiceFromJob(r.Context(), jobID)
	if err != nil {
		if err.Error() == "job not found" {
			h.respondWithError(w, http.StatusNotFound, "Job not found", nil)
			return
		}
		h.logger.Error("Failed to create invoice from job", "error", err, "job_id", jobID)
		h.respondWithError(w, http.StatusInternalServerError, "Failed to create invoice from job", err)
		return
	}

	h.respondWithJSON(w, http.StatusCreated, invoice)
}

// Helper methods

func (h *InvoiceHandler) parseInvoiceFilter(r *http.Request) (*services.InvoiceFilter, error) {
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

	// Parse boolean filter
	if overdueStr := r.URL.Query().Get("overdue"); overdueStr != "" {
		if overdue, err := strconv.ParseBool(overdueStr); err == nil {
			filter.Overdue = overdue
		}
	}

	// Parse UUID filters
	if customerIDStr := r.URL.Query().Get("customer_id"); customerIDStr != "" {
		if customerID, err := uuid.Parse(customerIDStr); err == nil {
			filter.CustomerID = &customerID
		}
	}

	return filter, nil
}

func (h *InvoiceHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
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

func (h *InvoiceHandler) respondWithError(w http.ResponseWriter, code int, message string, err error) {
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