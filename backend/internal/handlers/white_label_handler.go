package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// WhiteLabelHandler handles white-label customization operations
type WhiteLabelHandler struct {
	whiteLabelService services.WhiteLabelService
}

// NewWhiteLabelHandler creates a new white-label handler
func NewWhiteLabelHandler(whiteLabelService services.WhiteLabelService) *WhiteLabelHandler {
	return &WhiteLabelHandler{
		whiteLabelService: whiteLabelService,
	}
}

// SetupWhiteLabelRoutes sets up white-label customization routes
func (h *WhiteLabelHandler) SetupWhiteLabelRoutes(router *mux.Router) {
	// White-label customization routes
	whiteLabel := router.PathPrefix("/white-label").Subrouter()
	
	// Branding
	branding := whiteLabel.PathPrefix("/branding").Subrouter()
	branding.HandleFunc("", h.GetBranding).Methods("GET")
	branding.HandleFunc("", h.UpdateBranding).Methods("PUT")
	branding.HandleFunc("/logo", h.UploadLogo).Methods("POST")
	branding.HandleFunc("/logo", h.DeleteLogo).Methods("DELETE")
	
	// Themes
	themes := whiteLabel.PathPrefix("/themes").Subrouter()
	themes.HandleFunc("", h.GetTheme).Methods("GET")
	themes.HandleFunc("", h.UpdateTheme).Methods("PUT")
	themes.HandleFunc("/preview", h.PreviewTheme).Methods("POST")
	themes.HandleFunc("/templates/{template_id}", h.ApplyThemeTemplate).Methods("POST")
	themes.HandleFunc("/custom", h.CreateCustomTheme).Methods("POST")
	
	// Domain management
	domains := whiteLabel.PathPrefix("/domains").Subrouter()
	domains.HandleFunc("", h.ConfigureCustomDomain).Methods("POST")
	domains.HandleFunc("/verify", h.VerifyDomainOwnership).Methods("POST")
	domains.HandleFunc("/status", h.GetDomainStatus).Methods("GET")
	domains.HandleFunc("", h.RemoveCustomDomain).Methods("DELETE")
	
	// Email templates
	emailTemplates := whiteLabel.PathPrefix("/email-templates").Subrouter()
	emailTemplates.HandleFunc("", h.GetEmailTemplates).Methods("GET")
	emailTemplates.HandleFunc("", h.UpdateEmailTemplates).Methods("PUT")
	emailTemplates.HandleFunc("/preview/{template_type}", h.PreviewEmailTemplate).Methods("POST")
	emailTemplates.HandleFunc("/reset/{template_type}", h.ResetEmailTemplate).Methods("POST")
	
	// Mobile app customization
	mobileApp := whiteLabel.PathPrefix("/mobile-app").Subrouter()
	mobileApp.HandleFunc("", h.GetMobileAppConfig).Methods("GET")
	mobileApp.HandleFunc("", h.UpdateMobileAppConfig).Methods("PUT")
	mobileApp.HandleFunc("/icons", h.GenerateAppIcons).Methods("POST")
	mobileApp.HandleFunc("/store-settings", h.UpdateAppStoreSettings).Methods("PUT")
	
	// Feature configuration
	features := whiteLabel.PathPrefix("/features").Subrouter()
	features.HandleFunc("", h.GetFeatureConfig).Methods("GET")
	features.HandleFunc("", h.UpdateFeatureConfig).Methods("PUT")
	features.HandleFunc("/{feature_key}/enable", h.EnableCustomFeature).Methods("POST")
	features.HandleFunc("/{feature_key}/disable", h.DisableCustomFeature).Methods("POST")
	
	// Portal configuration
	portal := whiteLabel.PathPrefix("/portal").Subrouter()
	portal.HandleFunc("", h.GetPortalConfig).Methods("GET")
	portal.HandleFunc("", h.UpdatePortalConfig).Methods("PUT")
	
	// Asset management
	assets := whiteLabel.PathPrefix("/assets").Subrouter()
	assets.HandleFunc("", h.ListAssets).Methods("GET")
	assets.HandleFunc("", h.UploadAsset).Methods("POST")
	assets.HandleFunc("/{asset_id}", h.DeleteAsset).Methods("DELETE")
	
	// Configuration export/import
	config := whiteLabel.PathPrefix("/configuration").Subrouter()
	config.HandleFunc("/export", h.ExportConfiguration).Methods("GET")
	config.HandleFunc("/import", h.ImportConfiguration).Methods("POST")
	
	// Analytics
	analytics := whiteLabel.PathPrefix("/analytics").Subrouter()
	analytics.HandleFunc("/customization", h.GetCustomizationAnalytics).Methods("GET")
}

// Branding Management

func (h *WhiteLabelHandler) GetBranding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	branding, err := h.whiteLabelService.GetBranding(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get branding: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, branding)
}

func (h *WhiteLabelHandler) UpdateBranding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.BrandingUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	branding, err := h.whiteLabelService.UpdateBranding(ctx, tenantID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update branding: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, branding)
}

func (h *WhiteLabelHandler) UploadLogo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	// Parse multipart form
	if err := r.ParseMultipartForm(5 << 20); err != nil { // 5MB limit
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	
	file, header, err := r.FormFile("logo")
	if err != nil {
		http.Error(w, "Logo file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Read file data
	logoData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read logo file", http.StatusInternalServerError)
		return
	}
	
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg" // Default
	}
	
	response, err := h.whiteLabelService.UploadLogo(ctx, tenantID, logoData, contentType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload logo: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, response)
}

func (h *WhiteLabelHandler) DeleteLogo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	if err := h.whiteLabelService.DeleteLogo(ctx, tenantID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete logo: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Logo deleted successfully"})
}

// Theme Management

func (h *WhiteLabelHandler) GetTheme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	theme, err := h.whiteLabelService.GetTheme(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get theme: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, theme)
}

func (h *WhiteLabelHandler) UpdateTheme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.ThemeUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	theme, err := h.whiteLabelService.UpdateTheme(ctx, tenantID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update theme: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, theme)
}

func (h *WhiteLabelHandler) PreviewTheme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var theme services.ThemeConfig
	if err := json.NewDecoder(r.Body).Decode(&theme); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	preview, err := h.whiteLabelService.PreviewTheme(ctx, tenantID, &theme)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to preview theme: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, preview)
}

func (h *WhiteLabelHandler) ApplyThemeTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	templateID := vars["template_id"]
	
	theme, err := h.whiteLabelService.ApplyThemeTemplate(ctx, tenantID, templateID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to apply theme template: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, theme)
}

func (h *WhiteLabelHandler) CreateCustomTheme(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.CustomThemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Name == "" {
		http.Error(w, "Theme name is required", http.StatusBadRequest)
		return
	}
	
	theme, err := h.whiteLabelService.CreateCustomTheme(ctx, tenantID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create custom theme: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, theme)
}

// Domain Management

func (h *WhiteLabelHandler) ConfigureCustomDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.CustomDomainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.Domain == "" {
		http.Error(w, "Domain is required", http.StatusBadRequest)
		return
	}
	
	domainConfig, err := h.whiteLabelService.ConfigureCustomDomain(ctx, tenantID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to configure custom domain: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, domainConfig)
}

func (h *WhiteLabelHandler) VerifyDomainOwnership(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req struct {
		VerificationToken string `json:"verification_token"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if req.VerificationToken == "" {
		http.Error(w, "Verification token is required", http.StatusBadRequest)
		return
	}
	
	verification, err := h.whiteLabelService.VerifyDomainOwnership(ctx, tenantID, req.VerificationToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to verify domain ownership: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, verification)
}

func (h *WhiteLabelHandler) GetDomainStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	status, err := h.whiteLabelService.GetDomainStatus(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get domain status: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, status)
}

func (h *WhiteLabelHandler) RemoveCustomDomain(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	if err := h.whiteLabelService.RemoveCustomDomain(ctx, tenantID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to remove custom domain: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Custom domain removed successfully"})
}

// Email Template Management

func (h *WhiteLabelHandler) GetEmailTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	templates, err := h.whiteLabelService.GetEmailTemplates(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get email templates: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, templates)
}

func (h *WhiteLabelHandler) UpdateEmailTemplates(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.EmailTemplateUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.whiteLabelService.UpdateEmailTemplates(ctx, tenantID, &req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update email templates: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email templates updated successfully"})
}

func (h *WhiteLabelHandler) PreviewEmailTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	templateType := vars["template_type"]
	
	var data map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	preview, err := h.whiteLabelService.PreviewEmailTemplate(ctx, tenantID, templateType, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to preview email template: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, preview)
}

func (h *WhiteLabelHandler) ResetEmailTemplate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	templateType := vars["template_type"]
	
	if err := h.whiteLabelService.ResetEmailTemplate(ctx, tenantID, templateType); err != nil {
		http.Error(w, fmt.Sprintf("Failed to reset email template: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Email template reset successfully"})
}

// Mobile App Configuration

func (h *WhiteLabelHandler) GetMobileAppConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	config, err := h.whiteLabelService.GetMobileAppConfig(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get mobile app config: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, config)
}

func (h *WhiteLabelHandler) UpdateMobileAppConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.MobileAppConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	config, err := h.whiteLabelService.UpdateMobileAppConfig(ctx, tenantID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update mobile app config: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, config)
}

func (h *WhiteLabelHandler) GenerateAppIcons(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	// Parse multipart form
	if err := r.ParseMultipartForm(5 << 20); err != nil { // 5MB limit
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	
	file, _, err := r.FormFile("base_icon")
	if err != nil {
		http.Error(w, "Base icon file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Read file data
	baseIcon, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read base icon file", http.StatusInternalServerError)
		return
	}
	
	iconSet, err := h.whiteLabelService.GenerateAppIcons(ctx, tenantID, baseIcon)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to generate app icons: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, iconSet)
}

func (h *WhiteLabelHandler) UpdateAppStoreSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.AppStoreSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.whiteLabelService.UpdateAppStoreSettings(ctx, tenantID, &req); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update app store settings: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "App store settings updated successfully"})
}

// Feature Configuration

func (h *WhiteLabelHandler) GetFeatureConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	config, err := h.whiteLabelService.GetFeatureConfig(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get feature config: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, config)
}

func (h *WhiteLabelHandler) UpdateFeatureConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.FeatureConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	config, err := h.whiteLabelService.UpdateFeatureConfig(ctx, tenantID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update feature config: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, config)
}

func (h *WhiteLabelHandler) EnableCustomFeature(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	featureKey := vars["feature_key"]
	
	var config map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		config = make(map[string]interface{})
	}
	
	if err := h.whiteLabelService.EnableCustomFeature(ctx, tenantID, featureKey, config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to enable custom feature: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Custom feature enabled successfully"})
}

func (h *WhiteLabelHandler) DisableCustomFeature(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	featureKey := vars["feature_key"]
	
	if err := h.whiteLabelService.DisableCustomFeature(ctx, tenantID, featureKey); err != nil {
		http.Error(w, fmt.Sprintf("Failed to disable custom feature: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Custom feature disabled successfully"})
}

// Portal Configuration

func (h *WhiteLabelHandler) GetPortalConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	config, err := h.whiteLabelService.GetPortalConfig(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get portal config: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, config)
}

func (h *WhiteLabelHandler) UpdatePortalConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var req services.PortalConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	config, err := h.whiteLabelService.UpdatePortalConfig(ctx, tenantID, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update portal config: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, config)
}

// Asset Management

func (h *WhiteLabelHandler) ListAssets(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	assetType := r.URL.Query().Get("type")
	
	assets, err := h.whiteLabelService.ListAssets(ctx, tenantID, assetType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list assets: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, assets)
}

func (h *WhiteLabelHandler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	// Parse multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB limit
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	
	file, header, err := r.FormFile("asset")
	if err != nil {
		http.Error(w, "Asset file is required", http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	// Read file data
	assetData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Failed to read asset file", http.StatusInternalServerError)
		return
	}
	
	req := &services.AssetUploadRequest{
		Name:        r.FormValue("name"),
		Type:        r.FormValue("type"),
		ContentType: header.Header.Get("Content-Type"),
		Data:        assetData,
		Public:      r.FormValue("public") == "true",
	}
	
	if req.Name == "" {
		req.Name = header.Filename
	}
	
	if req.Type == "" {
		req.Type = "asset"
	}
	
	// Parse tags
	if tagsStr := r.FormValue("tags"); tagsStr != "" {
		var tags []string
		if err := json.Unmarshal([]byte(tagsStr), &tags); err == nil {
			req.Tags = tags
		}
	}
	
	asset, err := h.whiteLabelService.UploadAsset(ctx, tenantID, req)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to upload asset: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusCreated, asset)
}

func (h *WhiteLabelHandler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	vars := mux.Vars(r)
	assetID, err := uuid.Parse(vars["asset_id"])
	if err != nil {
		http.Error(w, "Invalid asset ID", http.StatusBadRequest)
		return
	}
	
	if err := h.whiteLabelService.DeleteAsset(ctx, tenantID, assetID); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete asset: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Asset deleted successfully"})
}

// Configuration Export/Import

func (h *WhiteLabelHandler) ExportConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	export, err := h.whiteLabelService.ExportConfiguration(ctx, tenantID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to export configuration: %v", err), http.StatusInternalServerError)
		return
	}
	
	// Set headers for file download
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\"white-label-config.json\"")
	
	respondWithJSON(w, http.StatusOK, export)
}

func (h *WhiteLabelHandler) ImportConfiguration(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	var config services.ConfigurationImport
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	if err := h.whiteLabelService.ImportConfiguration(ctx, tenantID, &config); err != nil {
		http.Error(w, fmt.Sprintf("Failed to import configuration: %v", err), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Configuration imported successfully"})
}

// Analytics

func (h *WhiteLabelHandler) GetCustomizationAnalytics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tenantID := getTenantIDFromContext(ctx)
	
	if tenantID == uuid.Nil {
		http.Error(w, "Tenant ID not found", http.StatusBadRequest)
		return
	}
	
	period, err := parseAnalyticsPeriod(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid period parameters: %v", err), http.StatusBadRequest)
		return
	}
	
	analytics, err := h.whiteLabelService.GetCustomizationAnalytics(ctx, tenantID, period)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get customization analytics: %v", err), http.StatusInternalServerError)
		return
	}
	
	respondWithJSON(w, http.StatusOK, analytics)
}