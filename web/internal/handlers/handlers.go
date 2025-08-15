package handlers

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/pageza/landscaping-app/web/internal/config"
	"github.com/pageza/landscaping-app/web/internal/middleware"
	"github.com/pageza/landscaping-app/web/internal/services"
)

// Handlers contains all HTTP handlers for the web application
type Handlers struct {
	config     *config.Config
	services   *services.Services
	middleware *middleware.Middleware
}

// NewHandlers creates a new handlers instance
func NewHandlers(cfg *config.Config, svc *services.Services, mw *middleware.Middleware) *Handlers {
	return &Handlers{
		config:     cfg,
		services:   svc,
		middleware: mw,
	}
}

// SetupRoutes configures all routes for the web application
func (h *Handlers) SetupRoutes() *mux.Router {
	r := mux.NewRouter()

	// Apply global middleware
	r.Use(h.middleware.Logger)
	r.Use(h.middleware.SecurityHeaders)
	r.Use(h.middleware.CORS)
	r.Use(h.middleware.HTMX)

	// Static files
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(h.config.StaticPath))))

	// Public routes (no authentication required)
	h.setupPublicRoutes(r)

	// Protected routes (authentication required)
	protected := r.PathPrefix("").Subrouter()
	protected.Use(h.middleware.RequireAuth)
	h.setupProtectedRoutes(protected)

	// Admin routes (admin role required)
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(h.middleware.RequireAuth)
	admin.Use(h.middleware.RequireRole("admin", "owner"))
	h.setupAdminRoutes(admin)

	// API proxy routes for HTMX requests
	api := r.PathPrefix("/api").Subrouter()
	api.Use(h.middleware.RequireAuth)
	h.setupAPIProxyRoutes(api)

	// WebSocket routes
	h.setupWebSocketRoutes(r)

	return r
}

// setupPublicRoutes configures public routes
func (h *Handlers) setupPublicRoutes(r *mux.Router) {
	// Authentication pages
	r.HandleFunc("/", h.redirectToLogin).Methods("GET")
	r.HandleFunc("/login", h.showLogin).Methods("GET")
	r.HandleFunc("/login", h.handleLogin).Methods("POST")
	r.HandleFunc("/register", h.showRegister).Methods("GET")
	r.HandleFunc("/register", h.handleRegister).Methods("POST")
	r.HandleFunc("/forgot-password", h.showForgotPassword).Methods("GET")
	r.HandleFunc("/forgot-password", h.handleForgotPassword).Methods("POST")
	r.HandleFunc("/reset-password", h.showResetPassword).Methods("GET")
	r.HandleFunc("/reset-password", h.handleResetPassword).Methods("POST")
	r.HandleFunc("/logout", h.handleLogout).Methods("GET", "POST")
}

// setupProtectedRoutes configures authenticated routes
func (h *Handlers) setupProtectedRoutes(r *mux.Router) {
	// Dashboard
	r.HandleFunc("/dashboard", h.showDashboard).Methods("GET")
	
	// Customer portal routes
	r.HandleFunc("/portal", h.showCustomerPortal).Methods("GET")
	r.HandleFunc("/portal/services", h.showCustomerServices).Methods("GET")
	r.HandleFunc("/portal/billing", h.showCustomerBilling).Methods("GET")
	r.HandleFunc("/portal/quotes", h.showCustomerQuotes).Methods("GET")
	
	// Profile management
	r.HandleFunc("/profile", h.showProfile).Methods("GET")
	r.HandleFunc("/profile", h.updateProfile).Methods("POST")
}

// setupAdminRoutes configures admin routes
func (h *Handlers) setupAdminRoutes(r *mux.Router) {
	// Admin dashboard
	r.HandleFunc("", h.showAdminDashboard).Methods("GET")
	r.HandleFunc("/", h.showAdminDashboard).Methods("GET")
	
	// Customer management
	r.HandleFunc("/customers", h.listCustomers).Methods("GET")
	r.HandleFunc("/customers/new", h.showCreateCustomer).Methods("GET")
	r.HandleFunc("/customers", h.createCustomer).Methods("POST")
	r.HandleFunc("/customers/{id}", h.showCustomer).Methods("GET")
	r.HandleFunc("/customers/{id}", h.updateCustomer).Methods("POST")
	r.HandleFunc("/customers/{id}/delete", h.deleteCustomer).Methods("POST")
	
	// Property management
	r.HandleFunc("/properties", h.listProperties).Methods("GET")
	r.HandleFunc("/properties/new", h.showCreateProperty).Methods("GET")
	r.HandleFunc("/properties", h.createProperty).Methods("POST")
	r.HandleFunc("/properties/{id}", h.showProperty).Methods("GET")
	r.HandleFunc("/properties/{id}", h.updateProperty).Methods("POST")
	
	// Job management
	r.HandleFunc("/jobs", h.listJobs).Methods("GET")
	r.HandleFunc("/jobs/new", h.showCreateJob).Methods("GET")
	r.HandleFunc("/jobs", h.createJob).Methods("POST")
	r.HandleFunc("/jobs/{id}", h.showJob).Methods("GET")
	r.HandleFunc("/jobs/{id}", h.updateJob).Methods("POST")
	r.HandleFunc("/jobs/calendar", h.showJobCalendar).Methods("GET")
	
	// Quote management
	r.HandleFunc("/quotes", h.listQuotes).Methods("GET")
	r.HandleFunc("/quotes/new", h.showCreateQuote).Methods("GET")
	r.HandleFunc("/quotes", h.createQuote).Methods("POST")
	r.HandleFunc("/quotes/{id}", h.showQuote).Methods("GET")
	r.HandleFunc("/quotes/{id}", h.updateQuote).Methods("POST")
	
	// Invoice management
	r.HandleFunc("/invoices", h.listInvoices).Methods("GET")
	r.HandleFunc("/invoices/new", h.showCreateInvoice).Methods("GET")
	r.HandleFunc("/invoices", h.createInvoice).Methods("POST")
	r.HandleFunc("/invoices/{id}", h.showInvoice).Methods("GET")
	r.HandleFunc("/invoices/{id}", h.updateInvoice).Methods("POST")
	
	// Equipment management
	r.HandleFunc("/equipment", h.listEquipment).Methods("GET")
	r.HandleFunc("/equipment/new", h.showCreateEquipment).Methods("GET")
	r.HandleFunc("/equipment", h.createEquipment).Methods("POST")
	r.HandleFunc("/equipment/{id}", h.showEquipment).Methods("GET")
	r.HandleFunc("/equipment/{id}", h.updateEquipment).Methods("POST")
	
	// Reports
	r.HandleFunc("/reports", h.showReports).Methods("GET")
	r.HandleFunc("/reports/revenue", h.showRevenueReport).Methods("GET")
	r.HandleFunc("/reports/jobs", h.showJobsReport).Methods("GET")
	
	// Settings
	r.HandleFunc("/settings", h.showSettings).Methods("GET")
	r.HandleFunc("/settings", h.updateSettings).Methods("POST")
}

// setupAPIProxyRoutes configures API proxy routes for HTMX
func (h *Handlers) setupAPIProxyRoutes(r *mux.Router) {
	// These routes proxy HTMX requests to the backend API
	// and return HTML fragments instead of JSON
	
	// Data tables with HTMX
	r.HandleFunc("/v1/customers/table", h.customersTable).Methods("GET")
	r.HandleFunc("/v1/properties/table", h.propertiesTable).Methods("GET")
	r.HandleFunc("/v1/jobs/table", h.jobsTable).Methods("GET")
	r.HandleFunc("/v1/quotes/table", h.quotesTable).Methods("GET")
	r.HandleFunc("/v1/invoices/table", h.invoicesTable).Methods("GET")
	r.HandleFunc("/v1/equipment/table", h.equipmentTable).Methods("GET")
	
	// Real-time updates
	r.HandleFunc("/v1/notifications", h.getNotifications).Methods("GET")
	r.HandleFunc("/v1/dashboard/stats", h.getDashboardStats).Methods("GET")
}

// setupWebSocketRoutes configures WebSocket routes
func (h *Handlers) setupWebSocketRoutes(r *mux.Router) {
	r.HandleFunc("/ws", h.handleWebSocket).Methods("GET")
	r.HandleFunc("/ws/chat", h.handleChatWebSocket).Methods("GET")
}

// Placeholder for authentication redirect
func (h *Handlers) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}