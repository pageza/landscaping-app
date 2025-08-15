package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/pageza/landscaping-app/web/internal/services"
)

// showDashboard displays the main dashboard
func (h *Handlers) showDashboard(w http.ResponseWriter, r *http.Request) {
	// Get current user
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Route based on user role
	switch user.Role {
	case "admin", "owner":
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	case "customer":
		http.Redirect(w, r, "/portal", http.StatusSeeOther)
	default:
		h.showGeneralDashboard(w, r, user)
	}
}

// showGeneralDashboard shows a general purpose dashboard
func (h *Handlers) showGeneralDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := services.TemplateData{
		Title:   "Dashboard",
		User:    user,
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
	}

	content, err := h.services.Template.Render("dashboard.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// showAdminDashboard displays the admin dashboard
func (h *Handlers) showAdminDashboard(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get dashboard statistics
	stats, err := h.getDashboardStatistics(user)
	if err != nil {
		http.Error(w, "Failed to load dashboard data", http.StatusInternalServerError)
		return
	}

	data := services.TemplateData{
		Title:   "Admin Dashboard",
		User:    user,
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
		Data:    stats,
	}

	content, err := h.services.Template.Render("admin_dashboard.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// showCustomerPortal displays the customer portal
func (h *Handlers) showCustomerPortal(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get customer-specific data
	customerData, err := h.getCustomerData(user)
	if err != nil {
		http.Error(w, "Failed to load customer data", http.StatusInternalServerError)
		return
	}

	data := services.TemplateData{
		Title:   "Customer Portal",
		User:    user,
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
		Data:    customerData,
	}

	content, err := h.services.Template.Render("customer_portal.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// showCustomerServices displays customer services
func (h *Handlers) showCustomerServices(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := services.TemplateData{
		Title:   "My Services",
		User:    user,
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
	}

	content, err := h.services.Template.Render("customer_services.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// showCustomerBilling displays customer billing
func (h *Handlers) showCustomerBilling(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := services.TemplateData{
		Title:   "Billing",
		User:    user,
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
	}

	content, err := h.services.Template.Render("customer_billing.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// showCustomerQuotes displays customer quotes
func (h *Handlers) showCustomerQuotes(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := services.TemplateData{
		Title:   "My Quotes",
		User:    user,
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
	}

	content, err := h.services.Template.Render("customer_quotes.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// getDashboardStats returns dashboard statistics as HTML fragment
func (h *Handlers) getDashboardStats(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	stats, err := h.getDashboardStatistics(user)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := services.TemplateData{
		User:   user,
		IsHTMX: true,
		Data:   stats,
	}

	content, err := h.services.Template.Render("dashboard_stats.html", data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// Helper functions

type DashboardStats struct {
	ActiveJobs      int     `json:"active_jobs"`
	PendingQuotes   int     `json:"pending_quotes"`
	Revenue         float64 `json:"revenue"`
	Customers       int     `json:"customers"`
	OverdueInvoices int     `json:"overdue_invoices"`
	Equipment       int     `json:"equipment"`
}

type CustomerData struct {
	UpcomingJobs   []interface{} `json:"upcoming_jobs"`
	RecentInvoices []interface{} `json:"recent_invoices"`
	ServiceHistory []interface{} `json:"service_history"`
	Properties     []interface{} `json:"properties"`
}

func (h *Handlers) getCurrentUser(r *http.Request) (*services.User, error) {
	// Extract token from cookie or header
	token := h.extractToken(r)
	if token == "" {
		return nil, NewValidationError("No authentication token")
	}

	// Get user info from backend
	return h.services.Auth.GetCurrentUser(token)
}

func (h *Handlers) extractToken(r *http.Request) string {
	// Check Authorization header
	if auth := r.Header.Get("Authorization"); auth != "" {
		return auth[7:] // Remove "Bearer " prefix
	}

	// Check session cookie
	if cookie, err := r.Cookie("session_token"); err == nil {
		return cookie.Value
	}

	return ""
}

func (h *Handlers) getDashboardStatistics(user *services.User) (*DashboardStats, error) {
	// Make API calls to get dashboard data
	token := "" // Extract from current context
	
	// In a real implementation, these would be actual API calls
	stats := &DashboardStats{
		ActiveJobs:      12,
		PendingQuotes:   5,
		Revenue:         25000.50,
		Customers:       45,
		OverdueInvoices: 3,
		Equipment:       8,
	}

	// Make authenticated API calls to get real data
	resp, err := h.services.API.AuthenticatedGet("/api/v1/reports/dashboard", token)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		json.NewDecoder(resp.Body).Decode(stats)
	}

	return stats, nil
}

func (h *Handlers) getCustomerData(user *services.User) (*CustomerData, error) {
	// Make API calls to get customer-specific data
	data := &CustomerData{
		UpcomingJobs:   []interface{}{},
		RecentInvoices: []interface{}{},
		ServiceHistory: []interface{}{},
		Properties:     []interface{}{},
	}

	// In real implementation, make API calls here
	
	return data, nil
}