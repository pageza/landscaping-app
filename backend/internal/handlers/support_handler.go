package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// SupportHandler handles support ticket operations
type SupportHandler struct {
	supportService services.SupportService
}

// NewSupportHandler creates a new support handler
func NewSupportHandler(supportService services.SupportService) *SupportHandler {
	return &SupportHandler{
		supportService: supportService,
	}
}

// SetupSupportRoutes sets up support ticket routes
func (h *SupportHandler) SetupSupportRoutes(router *mux.Router) {
	// Support ticket routes
	support := router.PathPrefix("/support").Subrouter()
	
	// Ticket management
	tickets := support.PathPrefix("/tickets").Subrouter()
	tickets.HandleFunc("", h.GetTickets).Methods("GET")
	tickets.HandleFunc("", h.CreateTicket).Methods("POST")
	tickets.HandleFunc("/{id}", h.GetTicket).Methods("GET")
	tickets.HandleFunc("/{id}", h.UpdateTicket).Methods("PUT")
	tickets.HandleFunc("/{id}/status", h.UpdateTicketStatus).Methods("PUT")
	tickets.HandleFunc("/{id}/priority", h.UpdateTicketPriority).Methods("PUT")
	tickets.HandleFunc("/{id}/assign", h.AssignTicket).Methods("PUT")
	tickets.HandleFunc("/{id}/messages", h.GetTicketMessages).Methods("GET")
	tickets.HandleFunc("/{id}/messages", h.AddTicketMessage).Methods("POST")
	tickets.HandleFunc("/{id}/resolve", h.ResolveTicket).Methods("POST")
	tickets.HandleFunc("/{id}/reopen", h.ReopenTicket).Methods("POST")
	tickets.HandleFunc("/{id}/escalate", h.EscalateTicket).Methods("POST")
	
	// Categories and templates
	support.HandleFunc("/categories", h.GetCategories).Methods("GET")
	support.HandleFunc("/templates", h.GetTicketTemplates).Methods("GET")
	
	// Knowledge base
	kb := support.PathPrefix("/knowledge-base").Subrouter()
	kb.HandleFunc("/articles", h.GetKnowledgeBaseArticles).Methods("GET")
	kb.HandleFunc("/articles/{id}", h.GetKnowledgeBaseArticle).Methods("GET")
	kb.HandleFunc("/search", h.SearchKnowledgeBase).Methods("GET")
	
	// SLA and metrics
	support.HandleFunc("/sla", h.GetSLAMetrics).Methods("GET")
	support.HandleFunc("/metrics", h.GetSupportMetrics).Methods("GET")
	
	// Internal notification management
	notifications := support.PathPrefix("/notifications").Subrouter()
	notifications.HandleFunc("", h.GetNotifications).Methods("GET")
	notifications.HandleFunc("/{id}/read", h.MarkNotificationRead).Methods("PUT")
	notifications.HandleFunc("/mark-all-read", h.MarkAllNotificationsRead).Methods("PUT")
}

// Ticket Management

func (h *SupportHandler) GetTickets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	filter := &services.TicketFilter{
		TenantID: tenantID,
		Page:     getIntQueryParam(r, "page", 1),
		PerPage:  getIntQueryParam(r, "per_page", 25),
		SortBy:   getStringQueryParam(r, "sort_by", "created_at"),
		SortOrder: getStringQueryParam(r, "sort_order", "desc"),
	}
	
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	
	if priority := r.URL.Query().Get("priority"); priority != "" {
		filter.Priority = &priority
	}
	
	if category := r.URL.Query().Get("category"); category != "" {
		filter.Category = &category
	}
	
	if assignedToStr := r.URL.Query().Get("assigned_to"); assignedToStr != "" {
		if assignedTo, err := uuid.Parse(assignedToStr); err == nil {
			filter.AssignedTo = &assignedTo
		}
	}
	
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}
	
	tickets, err := h.supportService.GetTickets(ctx, filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tickets: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, tickets)
}

func (h *SupportHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.CreateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if req.Subject == "" {
		http.Error(w, "Subject is required", http.StatusBadRequest)
		return
	}
	
	if req.Description == "" {
		http.Error(w, "Description is required", http.StatusBadRequest)
		return
	}
	
	if req.Category == "" {
		req.Category = "general"
	}
	
	if req.Priority == "" {
		req.Priority = "medium"
	}
	
	// Set tenant and user context
	req.TenantID = tenantID
	req.CreatedBy = userID
	
	ticket, err := h.supportService.CreateTicket(ctx, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create ticket: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, ticket)
}

func (h *SupportHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	ticket, err := h.supportService.GetTicket(ctx, ticketID, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get ticket: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, ticket)
}

func (h *SupportHandler) UpdateTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	var req services.UpdateTicketRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	req.UpdatedBy = userID
	
	ticket, err := h.supportService.UpdateTicket(ctx, ticketID, tenantID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update ticket: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, ticket)
}

func (h *SupportHandler) UpdateTicketStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Status string `json:"status"`
		Reason string `json:"reason,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Status == "" {
		http.Error(w, "Status is required", http.StatusBadRequest)
		return
	}
	
	if err := h.supportService.UpdateTicketStatus(ctx, ticketID, tenantID, req.Status, req.Reason, userID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update ticket status: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ticket status updated successfully"})
}

func (h *SupportHandler) UpdateTicketPriority(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Priority string `json:"priority"`
		Reason   string `json:"reason,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Priority == "" {
		http.Error(w, "Priority is required", http.StatusBadRequest)
		return
	}
	
	if err := h.supportService.UpdateTicketPriority(ctx, ticketID, tenantID, req.Priority, req.Reason, userID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update ticket priority: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ticket priority updated successfully"})
}

func (h *SupportHandler) AssignTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		AssignTo uuid.UUID `json:"assign_to"`
		Note     string    `json:"note,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.AssignTo == uuid.Nil {
		http.Error(w, "Assign to user ID is required", http.StatusBadRequest)
		return
	}
	
	if err := h.supportService.AssignTicket(ctx, ticketID, tenantID, req.AssignTo, req.Note, userID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to assign ticket: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ticket assigned successfully"})
}

func (h *SupportHandler) GetTicketMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	messages, err := h.supportService.GetTicketMessages(ctx, ticketID, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get ticket messages: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, messages)
}

func (h *SupportHandler) AddTicketMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	var req services.AddTicketMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Message == "" {
		http.Error(w, "Message is required", http.StatusBadRequest)
		return
	}
	
	req.TicketID = ticketID
	req.CreatedBy = userID
	
	message, err := h.supportService.AddTicketMessage(ctx, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to add ticket message: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, message)
}

func (h *SupportHandler) ResolveTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Resolution   string                 `json:"resolution"`
		Category     string                 `json:"category,omitempty"`
		Satisfaction *int                   `json:"satisfaction,omitempty"`
		Metadata     map[string]interface{} `json:"metadata,omitempty"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Resolution == "" {
		http.Error(w, "Resolution is required", http.StatusBadRequest)
		return
	}
	
	if err := h.supportService.ResolveTicket(ctx, ticketID, tenantID, req.Resolution, req.Category, req.Satisfaction, req.Metadata, userID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to resolve ticket: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ticket resolved successfully"})
}

func (h *SupportHandler) ReopenTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Reason string `json:"reason"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Reason == "" {
		http.Error(w, "Reason for reopening is required", http.StatusBadRequest)
		return
	}
	
	if err := h.supportService.ReopenTicket(ctx, ticketID, tenantID, req.Reason, userID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reopen ticket: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ticket reopened successfully"})
}

func (h *SupportHandler) EscalateTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Level  string `json:"level"`
		Reason string `json:"reason"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Level == "" {
		req.Level = "supervisor"
	}
	
	if req.Reason == "" {
		http.Error(w, "Escalation reason is required", http.StatusBadRequest)
		return
	}
	
	if err := h.supportService.EscalateTicket(ctx, ticketID, tenantID, req.Level, req.Reason, userID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to escalate ticket: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Ticket escalated successfully"})
}

// Support Information and Analytics

func (h *SupportHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	categories, err := h.supportService.GetCategories(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get categories: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, categories)
}

func (h *SupportHandler) GetTicketTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	templates, err := h.supportService.GetTicketTemplates(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get ticket templates: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, templates)
}

func (h *SupportHandler) GetKnowledgeBaseArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	category := r.URL.Query().Get("category")
	search := r.URL.Query().Get("search")
	page := getIntQueryParam(r, "page", 1)
	perPage := getIntQueryParam(r, "per_page", 20)
	
	articles, err := h.supportService.GetKnowledgeBaseArticles(ctx, category, search, page, perPage)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get knowledge base articles: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, articles)
}

func (h *SupportHandler) GetKnowledgeBaseArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	articleID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}
	
	article, err := h.supportService.GetKnowledgeBaseArticle(ctx, articleID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get knowledge base article: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, article)
}

func (h *SupportHandler) SearchKnowledgeBase(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Search query is required", http.StatusBadRequest)
		return
	}
	
	limit := getIntQueryParam(r, "limit", 10)
	
	results, err := h.supportService.SearchKnowledgeBase(ctx, query, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to search knowledge base: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, results)
}

func (h *SupportHandler) GetSLAMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	period, err := parseAnalyticsPeriod(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid period parameters: %v", err), http.StatusBadRequest)
		return
	}
	
	slaMetrics, err := h.supportService.GetSLAMetrics(ctx, tenantID, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get SLA metrics: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, slaMetrics)
}

func (h *SupportHandler) GetSupportMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	period, err := parseAnalyticsPeriod(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid period parameters: %v", err), http.StatusBadRequest)
		return
	}
	
	metrics, err := h.supportService.GetSupportMetrics(ctx, tenantID, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get support metrics: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, metrics)
}

// Notification Management

func (h *SupportHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	page := getIntQueryParam(r, "page", 1)
	perPage := getIntQueryParam(r, "per_page", 25)
	unreadOnly := r.URL.Query().Get("unread_only") == "true"
	
	notifications, err := h.supportService.GetNotifications(ctx, tenantID, userID, page, perPage, unreadOnly)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get notifications: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, notifications)
}

func (h *SupportHandler) MarkNotificationRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	vars := mux.Vars(r)
	notificationID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}
	
	if err := h.supportService.MarkNotificationRead(ctx, notificationID, tenantID, userID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to mark notification as read: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Notification marked as read"})
}

func (h *SupportHandler) MarkAllNotificationsRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	userID := getUserIDFromContext(ctx)
	
	if err := h.supportService.MarkAllNotificationsRead(ctx, tenantID, userID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to mark all notifications as read: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "All notifications marked as read"})
}

// Helper function to get user ID from context
func getUserIDFromContext(ctx context.Context) uuid.UUID {
	// TODO: Implement getting user ID from context
	// This would be set by authentication middleware
	return uuid.Nil
}