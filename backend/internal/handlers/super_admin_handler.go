package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// SuperAdminHandler handles super admin operations
type SuperAdminHandler struct {
	superAdminService services.SuperAdminService
}

// NewSuperAdminHandler creates a new super admin handler
func NewSuperAdminHandler(superAdminService services.SuperAdminService) *SuperAdminHandler {
	return &SuperAdminHandler{
		superAdminService: superAdminService,
	}
}

// SetupSuperAdminRoutes sets up super admin routes
func (h *SuperAdminHandler) SetupSuperAdminRoutes(router *mux.Router) {
	// Super admin routes - require super admin privileges
	admin := router.PathPrefix("/super-admin").Subrouter()
	
	// Platform overview and dashboard
	admin.HandleFunc("/overview", h.GetPlatformOverview).Methods("GET")
	admin.HandleFunc("/health", h.GetSystemHealth).Methods("GET")
	
	// Tenant management
	admin.HandleFunc("/tenants", h.ListTenants).Methods("GET")
	admin.HandleFunc("/tenants/{id}", h.GetTenantDetails).Methods("GET")
	admin.HandleFunc("/tenants/{id}/suspend", h.SuspendTenant).Methods("POST")
	admin.HandleFunc("/tenants/{id}/reactivate", h.ReactivateTenant).Methods("POST")
	admin.HandleFunc("/tenants/{id}/upgrade", h.UpgradeTenant).Methods("PUT")
	admin.HandleFunc("/tenants/{id}", h.DeleteTenant).Methods("DELETE")
	
	// Analytics and metrics
	admin.HandleFunc("/metrics/tenants", h.GetTenantMetrics).Methods("GET")
	admin.HandleFunc("/analytics/revenue", h.GetRevenueAnalytics).Methods("GET")
	admin.HandleFunc("/analytics/usage", h.GetUsageAnalytics).Methods("GET")
	admin.HandleFunc("/analytics/churn", h.GetChurnAnalysis).Methods("GET")
	admin.HandleFunc("/metrics/performance", h.GetPerformanceMetrics).Methods("GET")
	admin.HandleFunc("/metrics/security", h.GetSecurityMetrics).Methods("GET")
	
	// Support management
	admin.HandleFunc("/support/tickets", h.GetSupportTickets).Methods("GET")
	admin.HandleFunc("/support/tickets/{id}/resolve", h.ResolveSupportTicket).Methods("POST")
	
	// Feature flag management
	admin.HandleFunc("/feature-flags", h.GetFeatureFlagUsage).Methods("GET")
	admin.HandleFunc("/feature-flags/global", h.UpdateGlobalFeatureFlag).Methods("PUT")
	admin.HandleFunc("/feature-flags/tenants/{id}", h.UpdateTenantFeatureFlag).Methods("PUT")
	
	// Platform configuration
	admin.HandleFunc("/settings", h.GetPlatformSettings).Methods("GET")
	admin.HandleFunc("/settings", h.UpdatePlatformSettings).Methods("PUT")
	
	// Revenue operations
	admin.HandleFunc("/revenue/refund", h.ProcessRefund).Methods("POST")
	admin.HandleFunc("/revenue/credit", h.ApplyCredit).Methods("POST")
}

// GetPlatformOverview returns high-level platform statistics
func (h *SuperAdminHandler) GetPlatformOverview(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	overview, err := h.superAdminService.GetPlatformOverview(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get platform overview: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, overview)
}

// GetSystemHealth returns current system health status
func (h *SuperAdminHandler) GetSystemHealth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	health, err := h.superAdminService.GetSystemHealth(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get system health: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, health)
}

// ListTenants returns a paginated list of tenants with filtering
func (h *SuperAdminHandler) ListTenants(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Parse query parameters
	filter := &services.TenantFilter{
		Page:      getIntQueryParam(r, "page", 1),
		PerPage:   getIntQueryParam(r, "per_page", 25),
		SortBy:    getStringQueryParam(r, "sort_by", "created_at"),
		SortOrder: getStringQueryParam(r, "sort_order", "desc"),
	}
	
	if search := r.URL.Query().Get("search"); search != "" {
		filter.Search = &search
	}
	if status := r.URL.Query().Get("status"); status != "" {
		filter.Status = &status
	}
	if plan := r.URL.Query().Get("plan"); plan != "" {
		filter.Plan = &plan
	}
	if churnRisk := r.URL.Query().Get("churn_risk"); churnRisk != "" {
		filter.ChurnRisk = &churnRisk
	}
	
	tenants, err := h.superAdminService.ListTenants(ctx, filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list tenants: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, tenants)
}

// GetTenantDetails returns comprehensive details for a specific tenant
func (h *SuperAdminHandler) GetTenantDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	details, err := h.superAdminService.GetTenantDetails(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tenant details: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, details)
}

// SuspendTenant suspends a tenant with a reason
func (h *SuperAdminHandler) SuspendTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
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
		http.Error(w, "Suspension reason is required", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.SuspendTenant(ctx, tenantID, req.Reason); err != nil {
		http.Error(w, fmt.Sprintf("Failed to suspend tenant: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tenant suspended successfully"})
}

// ReactivateTenant reactivates a suspended tenant
func (h *SuperAdminHandler) ReactivateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.ReactivateTenant(ctx, tenantID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reactivate tenant: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tenant reactivated successfully"})
}

// UpgradeTenant upgrades a tenant to a new plan
func (h *SuperAdminHandler) UpgradeTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		NewPlan string `json:"new_plan"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.NewPlan == "" {
		http.Error(w, "New plan is required", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.UpgradeTenant(ctx, tenantID, req.NewPlan); err != nil {
		http.Error(w, fmt.Sprintf("Failed to upgrade tenant: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tenant upgraded successfully"})
}

// DeleteTenant deletes a tenant
func (h *SuperAdminHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
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
		http.Error(w, "Deletion reason is required", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.DeleteTenant(ctx, tenantID, req.Reason); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete tenant: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tenant deleted successfully"})
}

// GetTenantMetrics returns tenant metrics with filtering
func (h *SuperAdminHandler) GetTenantMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	filter := &services.TenantMetricsFilter{}
	
	// Parse query parameters for filtering
	if tenantIDs := r.URL.Query()["tenant_ids"]; len(tenantIDs) > 0 {
		for _, idStr := range tenantIDs {
			if id, err := uuid.Parse(idStr); err == nil {
				filter.TenantIDs = append(filter.TenantIDs, id)
			}
		}
	}
	
	if plans := r.URL.Query()["plans"]; len(plans) > 0 {
		filter.Plans = plans
	}
	
	if statuses := r.URL.Query()["statuses"]; len(statuses) > 0 {
		filter.Statuses = statuses
	}
	
	if minRevStr := r.URL.Query().Get("min_revenue"); minRevStr != "" {
		if minRev, err := strconv.ParseFloat(minRevStr, 64); err == nil {
			filter.MinRevenue = &minRev
		}
	}
	
	if maxRevStr := r.URL.Query().Get("max_revenue"); maxRevStr != "" {
		if maxRev, err := strconv.ParseFloat(maxRevStr, 64); err == nil {
			filter.MaxRevenue = &maxRev
		}
	}
	
	metrics, err := h.superAdminService.GetTenantMetrics(ctx, filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get tenant metrics: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, metrics)
}

// GetRevenueAnalytics returns revenue analytics for a period
func (h *SuperAdminHandler) GetRevenueAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	period, err := parseAnalyticsPeriod(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid period parameters: %v", err), http.StatusBadRequest)
		return
	}
	
	analytics, err := h.superAdminService.GetRevenueAnalytics(ctx, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get revenue analytics: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, analytics)
}

// GetUsageAnalytics returns usage analytics for a period
func (h *SuperAdminHandler) GetUsageAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	period, err := parseAnalyticsPeriod(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid period parameters: %v", err), http.StatusBadRequest)
		return
	}
	
	analytics, err := h.superAdminService.GetUsageAnalytics(ctx, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get usage analytics: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, analytics)
}

// GetChurnAnalysis returns churn analysis for a period
func (h *SuperAdminHandler) GetChurnAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	period, err := parseAnalyticsPeriod(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid period parameters: %v", err), http.StatusBadRequest)
		return
	}
	
	analysis, err := h.superAdminService.GetChurnAnalysis(ctx, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get churn analysis: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, analysis)
}

// GetPerformanceMetrics returns performance metrics for a period
func (h *SuperAdminHandler) GetPerformanceMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	period, err := parseAnalyticsPeriod(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid period parameters: %v", err), http.StatusBadRequest)
		return
	}
	
	metrics, err := h.superAdminService.GetPerformanceMetrics(ctx, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get performance metrics: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, metrics)
}

// GetSecurityMetrics returns security metrics for a period
func (h *SuperAdminHandler) GetSecurityMetrics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	period, err := parseAnalyticsPeriod(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid period parameters: %v", err), http.StatusBadRequest)
		return
	}
	
	metrics, err := h.superAdminService.GetSecurityMetrics(ctx, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get security metrics: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, metrics)
}

// GetSupportTickets returns support tickets with filtering
func (h *SuperAdminHandler) GetSupportTickets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	filter := &services.SupportTicketFilter{
		Page:    getIntQueryParam(r, "page", 1),
		PerPage: getIntQueryParam(r, "per_page", 25),
	}
	
	if tenantIDStr := r.URL.Query().Get("tenant_id"); tenantIDStr != "" {
		if tenantID, err := uuid.Parse(tenantIDStr); err == nil {
			filter.TenantID = &tenantID
		}
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
	
	tickets, err := h.superAdminService.GetSupportTickets(ctx, filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get support tickets: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, tickets)
}

// ResolveSupportTicket resolves a support ticket
func (h *SuperAdminHandler) ResolveSupportTicket(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	ticketID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid ticket ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Resolution string `json:"resolution"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Resolution == "" {
		http.Error(w, "Resolution is required", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.ResolveSupportTicket(ctx, ticketID, req.Resolution); err != nil {
		http.Error(w, fmt.Sprintf("Failed to resolve support ticket: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Support ticket resolved successfully"})
}

// UpdateGlobalFeatureFlag updates a global feature flag
func (h *SuperAdminHandler) UpdateGlobalFeatureFlag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		Flag    string `json:"flag"`
		Enabled bool   `json:"enabled"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Flag == "" {
		http.Error(w, "Flag name is required", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.UpdateGlobalFeatureFlag(ctx, req.Flag, req.Enabled); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update feature flag: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Feature flag updated successfully"})
}

// UpdateTenantFeatureFlag updates a tenant-specific feature flag
func (h *SuperAdminHandler) UpdateTenantFeatureFlag(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		Flag    string `json:"flag"`
		Enabled bool   `json:"enabled"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Flag == "" {
		http.Error(w, "Flag name is required", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.UpdateTenantFeatureFlag(ctx, tenantID, req.Flag, req.Enabled); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update tenant feature flag: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Tenant feature flag updated successfully"})
}

// GetFeatureFlagUsage returns feature flag usage statistics
func (h *SuperAdminHandler) GetFeatureFlagUsage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	usage, err := h.superAdminService.GetFeatureFlagUsage(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get feature flag usage: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, usage)
}

// GetPlatformSettings returns platform settings
func (h *SuperAdminHandler) GetPlatformSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	settings, err := h.superAdminService.GetPlatformSettings(ctx)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get platform settings: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, settings)
}

// UpdatePlatformSettings updates platform settings
func (h *SuperAdminHandler) UpdatePlatformSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var settings services.PlatformSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.UpdatePlatformSettings(ctx, &settings); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update platform settings: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Platform settings updated successfully"})
}

// ProcessRefund processes a refund
func (h *SuperAdminHandler) ProcessRefund(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		PaymentID uuid.UUID `json:"payment_id"`
		Amount    float64   `json:"amount"`
		Reason    string    `json:"reason"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.PaymentID == uuid.Nil {
		http.Error(w, "Payment ID is required", http.StatusBadRequest)
		return
	}
	
	if req.Amount <= 0 {
		http.Error(w, "Refund amount must be positive", http.StatusBadRequest)
		return
	}
	
	if req.Reason == "" {
		http.Error(w, "Refund reason is required", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.ProcessRefund(ctx, req.PaymentID, req.Amount, req.Reason); err != nil {
		http.Error(w, fmt.Sprintf("Failed to process refund: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Refund processed successfully"})
}

// ApplyCredit applies a credit to a tenant account
func (h *SuperAdminHandler) ApplyCredit(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req struct {
		TenantID uuid.UUID `json:"tenant_id"`
		Amount   float64   `json:"amount"`
		Reason   string    `json:"reason"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.TenantID == uuid.Nil {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}
	
	if req.Amount <= 0 {
		http.Error(w, "Credit amount must be positive", http.StatusBadRequest)
		return
	}
	
	if req.Reason == "" {
		http.Error(w, "Credit reason is required", http.StatusBadRequest)
		return
	}
	
	if err := h.superAdminService.ApplyCredit(ctx, req.TenantID, req.Amount, req.Reason); err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply credit: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Credit applied successfully"})
}

// Helper functions

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func getIntQueryParam(r *http.Request, key string, defaultValue int) int {
	if value := r.URL.Query().Get(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getStringQueryParam(r *http.Request, key, defaultValue string) string {
	if value := r.URL.Query().Get(key); value != "" {
		return value
	}
	return defaultValue
}

func parseAnalyticsPeriod(r *http.Request) (*services.AnalyticsPeriod, error) {
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	interval := getStringQueryParam(r, "interval", "day")
	
	if startDateStr == "" || endDateStr == "" {
		// Default to last 30 days
		endDate := time.Now()
		startDate := endDate.AddDate(0, 0, -30)
		return &services.AnalyticsPeriod{
			StartDate: startDate,
			EndDate:   endDate,
			Interval:  interval,
		}, nil
	}
	
	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format: %v", err)
	}
	
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end_date format: %v", err)
	}
	
	return &services.AnalyticsPeriod{
		StartDate: startDate,
		EndDate:   endDate,
		Interval:  interval,
	}, nil
}