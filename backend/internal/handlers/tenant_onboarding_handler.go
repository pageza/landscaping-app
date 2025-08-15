package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// TenantOnboardingHandler handles tenant onboarding operations
type TenantOnboardingHandler struct {
	onboardingService services.TenantOnboardingService
}

// NewTenantOnboardingHandler creates a new tenant onboarding handler
func NewTenantOnboardingHandler(onboardingService services.TenantOnboardingService) *TenantOnboardingHandler {
	return &TenantOnboardingHandler{
		onboardingService: onboardingService,
	}
}

// SetupTenantOnboardingRoutes sets up tenant onboarding routes
func (h *TenantOnboardingHandler) SetupTenantOnboardingRoutes(router *mux.Router) {
	// Public onboarding endpoints - no authentication required
	onboarding := router.PathPrefix("/onboarding").Subrouter()
	
	// Start tenant onboarding
	onboarding.HandleFunc("/start", h.StartOnboarding).Methods("POST")
	
	// Get onboarding status (with token authentication)
	onboarding.HandleFunc("/status/{tenant_id}", h.GetOnboardingStatus).Methods("GET")
	
	// Complete onboarding
	onboarding.HandleFunc("/complete/{tenant_id}", h.CompleteOnboarding).Methods("POST")
	
	// Domain configuration
	onboarding.HandleFunc("/domain/{tenant_id}", h.ConfigureDomain).Methods("POST")
	
	// Protected onboarding endpoints - require authentication
	protected := router.PathPrefix("/tenant").Subrouter()
	
	// Provision tenant resources (internal use)
	protected.HandleFunc("/{tenant_id}/provision", h.ProvisionTenantResources).Methods("POST")
	
	// Setup default data
	protected.HandleFunc("/{tenant_id}/setup-defaults", h.SetupDefaultData).Methods("POST")
	
	// Send welcome sequence
	protected.HandleFunc("/{tenant_id}/welcome", h.SendWelcomeSequence).Methods("POST")
}

// StartOnboarding initiates the onboarding process for a new tenant
func (h *TenantOnboardingHandler) StartOnboarding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	var req services.TenantOnboardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if err := h.validateOnboardingRequest(&req); err != nil {
		http.Error(w, fmt.Sprintf("Validation error: %v", err), http.StatusBadRequest)
		return
	}
	
	response, err := h.onboardingService.StartOnboarding(ctx, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to start onboarding: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, response)
}

// GetOnboardingStatus returns the current onboarding status
func (h *TenantOnboardingHandler) GetOnboardingStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["tenant_id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	// TODO: Validate onboarding token from query parameters or headers
	
	status, err := h.onboardingService.GetOnboardingStatus(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get onboarding status: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, status)
}

// CompleteOnboarding marks onboarding as complete and activates the tenant
func (h *TenantOnboardingHandler) CompleteOnboarding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["tenant_id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	var completionData services.OnboardingCompletionData
	if err := json.NewDecoder(r.Body).Decode(&completionData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.onboardingService.CompleteOnboarding(ctx, tenantID, &completionData); err != nil {
		http.Error(w, fmt.Sprintf("Failed to complete onboarding: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Onboarding completed successfully",
		"tenant_id":  tenantID,
		"status":     "completed",
		"next_steps": []string{
			"Access your dashboard",
			"Start adding customers and jobs",
			"Configure your team settings",
			"Explore advanced features",
		},
	})
}

// ConfigureDomain sets up custom domain configuration
func (h *TenantOnboardingHandler) ConfigureDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["tenant_id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	var domainConfig services.DomainConfiguration
	if err := json.NewDecoder(r.Body).Decode(&domainConfig); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Validate domain configuration
	if domainConfig.Domain == "" {
		http.Error(w, "Domain is required", http.StatusBadRequest)
		return
	}
	
	if err := h.onboardingService.ConfigureDomain(ctx, tenantID, &domainConfig); err != nil {
		http.Error(w, fmt.Sprintf("Failed to configure domain: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Domain configuration started",
		"domain":  domainConfig.Domain,
		"status":  "pending_verification",
		"instructions": "Please add the required DNS records to verify domain ownership",
	})
}

// ProvisionTenantResources provisions all necessary resources for a new tenant
func (h *TenantOnboardingHandler) ProvisionTenantResources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["tenant_id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	if err := h.onboardingService.ProvisionTenantResources(ctx, tenantID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to provision tenant resources: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Tenant resources provisioned successfully",
	})
}

// SetupDefaultData creates default data for a new tenant
func (h *TenantOnboardingHandler) SetupDefaultData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["tenant_id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	if err := h.onboardingService.SetupDefaultData(ctx, tenantID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to setup default data: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Default data setup completed successfully",
	})
}

// SendWelcomeSequence sends welcome emails and setup instructions
func (h *TenantOnboardingHandler) SendWelcomeSequence(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	vars := mux.Vars(r)
	tenantID, err := uuid.Parse(vars["tenant_id"])
	if err != nil {
		http.Error(w, "Invalid tenant ID", http.StatusBadRequest)
		return
	}
	
	var req struct {
		OwnerEmail string `json:"owner_email"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.OwnerEmail == "" {
		http.Error(w, "Owner email is required", http.StatusBadRequest)
		return
	}
	
	if err := h.onboardingService.SendWelcomeSequence(ctx, tenantID, req.OwnerEmail); err != nil {
		http.Error(w, fmt.Sprintf("Failed to send welcome sequence: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Welcome sequence sent successfully",
	})
}

// validateOnboardingRequest validates the onboarding request
func (h *TenantOnboardingHandler) validateOnboardingRequest(req *services.TenantOnboardingRequest) error {
	if req.CompanyName == "" {
		return fmt.Errorf("company name is required")
	}
	
	if req.Subdomain == "" {
		return fmt.Errorf("subdomain is required")
	}
	
	if req.OwnerEmail == "" {
		return fmt.Errorf("owner email is required")
	}
	
	if req.OwnerFirstName == "" {
		return fmt.Errorf("owner first name is required")
	}
	
	if req.OwnerLastName == "" {
		return fmt.Errorf("owner last name is required")
	}
	
	if req.Plan == "" {
		return fmt.Errorf("plan is required")
	}
	
	if req.CompanyType == "" {
		return fmt.Errorf("company type is required")
	}
	
	if req.EmployeeCount == "" {
		return fmt.Errorf("employee count is required")
	}
	
	if req.Timezone == "" {
		return fmt.Errorf("timezone is required")
	}
	
	if req.Currency == "" {
		return fmt.Errorf("currency is required")
	}
	
	// Validate plan is one of allowed values
	validPlans := map[string]bool{
		"basic":      true,
		"premium":    true,
		"enterprise": true,
	}
	
	if !validPlans[req.Plan] {
		return fmt.Errorf("invalid plan: must be one of basic, premium, enterprise")
	}
	
	// Validate company type
	validCompanyTypes := map[string]bool{
		"landscaping":  true,
		"maintenance":  true,
		"construction": true,
	}
	
	if !validCompanyTypes[req.CompanyType] {
		return fmt.Errorf("invalid company type: must be one of landscaping, maintenance, construction")
	}
	
	// Validate employee count
	validEmployeeCounts := map[string]bool{
		"1-10":   true,
		"11-50":  true,
		"51-200": true,
		"200+":   true,
	}
	
	if !validEmployeeCounts[req.EmployeeCount] {
		return fmt.Errorf("invalid employee count: must be one of 1-10, 11-50, 51-200, 200+")
	}
	
	// Validate currency is 3 letters
	if len(req.Currency) != 3 {
		return fmt.Errorf("currency must be a 3-letter code (e.g., USD, EUR)")
	}
	
	// Validate subdomain format (basic validation)
	if len(req.Subdomain) < 3 || len(req.Subdomain) > 63 {
		return fmt.Errorf("subdomain must be between 3 and 63 characters")
	}
	
	// Validate trial days if provided
	if req.TrialDays != nil {
		if *req.TrialDays < 0 || *req.TrialDays > 90 {
			return fmt.Errorf("trial days must be between 0 and 90")
		}
	}
	
	return nil
}