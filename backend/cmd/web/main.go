package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
	"strings"
	"strconv"
	"math"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
)

type PageData struct {
	Title       string
	Page        string
	User        interface{}
	Services    []Service
	Testimonials []Testimonial
	Data        map[string]interface{}
}

type User struct {
	ID           string `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Phone        string `json:"phone"`
	IsAdmin      bool   `json:"is_admin"`
	PasswordHash string `json:"-"`
}

// Admin data structures
type Customer struct {
	ID           string    `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	Phone        string    `json:"phone"`
	Address      string    `json:"address"`
	City         string    `json:"city"`
	State        string    `json:"state"`
	ZipCode      string    `json:"zip_code"`
	CreatedAt    time.Time `json:"created_at"`
	LastService  time.Time `json:"last_service"`
	TotalSpent   float64   `json:"total_spent"`
	ServiceCount int       `json:"service_count"`
}

type ServiceRequest struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	CustomerName string    `json:"customer_name"`
	CustomerEmail string   `json:"customer_email"`
	CustomerPhone string   `json:"customer_phone"`
	ServiceType  string    `json:"service_type"`
	Message      string    `json:"message"`
	Status       string    `json:"status"` // pending, accepted, denied, scheduled
	CreatedAt    time.Time `json:"created_at"`
	PropertyInfo map[string]interface{} `json:"property_info"` // Store property size, address, etc.
	EstimatedPrice float64 `json:"estimated_price"`
}

type Job struct {
	ID          string    `json:"id"`
	ServiceRequestID string `json:"service_request_id"`
	CustomerID  string    `json:"customer_id"`
	ServiceID   string    `json:"service_id"`
	TeamID      string    `json:"team_id"`
	Status      string    `json:"status"` // scheduled, in-progress, completed, cancelled
	ScheduledAt time.Time `json:"scheduled_at"`
	CompletedAt *time.Time `json:"completed_at"`
	Notes       string    `json:"notes"`
	Price       float64   `json:"price"`
	Duration    int       `json:"duration"` // minutes
}

type Team struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Members     []string `json:"members"`
	Specialties []string `json:"specialties"`
	Active      bool     `json:"active"`
}

type Employee struct {
	ID        string    `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Role      string    `json:"role"`
	HireDate  time.Time `json:"hire_date"`
	Active    bool      `json:"active"`
	TeamID    string    `json:"team_id"`
}

type Service struct {
	ID          string
	Name        string
	Description string
	Price       string
	Image       string
	Icon        string
}

type Testimonial struct {
	ID         string
	Name       string
	Role       string
	Content    string
	Rating     int
	Avatar     string
}

// Pricing structures for intelligent pricing algorithm
type PricingRequest struct {
	PropertySize     int      `json:"property_size"`     // Square footage
	ServiceTypes     []string `json:"service_types"`     // Array of service IDs
	Frequency        string   `json:"frequency"`         // weekly, bi-weekly, monthly, one-time
	ZipCode          string   `json:"zip_code"`          // For geographic zone pricing
	SameDayBooking   bool     `json:"same_day_booking"`  // Booking for today
	MultipleServices bool     `json:"multiple_services"` // Bulk discount
	Distance         float64  `json:"distance"`          // Miles from service area
}

type PricingResponse struct {
	BaseRate         float64            `json:"base_rate"`
	ServiceBreakdown map[string]float64 `json:"service_breakdown"`
	Subtotal         float64            `json:"subtotal"`
	Discounts        map[string]float64 `json:"discounts"`
	Surcharges       map[string]float64 `json:"surcharges"`
	TotalAmount      float64            `json:"total_amount"`
	EstimatedHours   float64            `json:"estimated_hours"`
}

// Service pricing configuration
type ServiceConfig struct {
	ID         string
	Name       string
	Multiplier float64 // Complexity multiplier
	BaseHours  float64 // Base hours per visit
}

// Global template variable
var tmpl *template.Template

// Session store
var store = sessions.NewCookieStore([]byte("landscaping-secret-key-change-in-production"))

// Real user accounts storage with predefined test accounts
var users = map[string]User{
	"admin@landscapepro.com": {
		ID:        "admin-001",
		Email:     "admin@landscapepro.com",
		FirstName: "Admin",
		LastName:  "User",
		Phone:     "(555) 123-4567",
		IsAdmin:   true,
		PasswordHash: "$2a$10$EF9YBmwhZSBfrGSTMMPUV.1Hz2Uubo9MUD0XIIQXyv4QO5HClR7Di", // password123
	},
	"manager@landscapepro.com": {
		ID:        "admin-002",
		Email:     "manager@landscapepro.com",
		FirstName: "Manager",
		LastName:  "User",
		Phone:     "(555) 234-5678",
		IsAdmin:   true,
		PasswordHash: "$2a$10$B/38SGvAvioxOScsmM4fCOzToEamHyp5oB.sJ.RpKZe6mOY1YM30S", // manager456
	},
	"customer1@email.com": {
		ID:        "customer-001",
		Email:     "customer1@email.com",
		FirstName: "John",
		LastName:  "Customer",
		Phone:     "(555) 345-6789",
		IsAdmin:   false,
		PasswordHash: "$2a$10$sWJkhJbgF55BUjjuG0JbcOFYz74gF8R9wb6J3jXy67i8GJv/14l4i", // customer123
	},
	"customer2@email.com": {
		ID:        "customer-002",
		Email:     "customer2@email.com",
		FirstName: "Jane",
		LastName:  "Customer",
		Phone:     "(555) 456-7890",
		IsAdmin:   false,
		PasswordHash: "$2a$10$NUbIbDo9K5OPHlcuBwRz7OyjbcehQNoYWelyb3oEU36I3U/vSXO0O", // customer456
	},
}

// Real data storage (in-memory for demo)
var serviceRequests = make(map[string]ServiceRequest)
var customers = make(map[string]Customer) // Now populated from actual registrations
var jobs = make(map[string]Job)
var teams = make(map[string]Team)
var employees = make(map[string]Employee)

// Pricing configuration
var serviceConfigs = map[string]ServiceConfig{
	"lawn_care": {
		ID:         "lawn_care",
		Name:       "Lawn Care",
		Multiplier: 1.0,
		BaseHours:  2.0,
	},
	"garden_design": {
		ID:         "garden_design", 
		Name:       "Garden Design",
		Multiplier: 1.8,
		BaseHours:  4.0,
	},
	"tree_service": {
		ID:         "tree_service",
		Name:       "Tree Service", 
		Multiplier: 2.2,
		BaseHours:  3.0,
	},
	"irrigation": {
		ID:         "irrigation",
		Name:       "Irrigation",
		Multiplier: 1.5,
		BaseHours:  3.5,
	},
	"hardscaping": {
		ID:         "hardscaping",
		Name:       "Hardscaping",
		Multiplier: 2.5,
		BaseHours:  6.0,
	},
}

// Geographic zones for pricing
var geographicZones = map[string]float64{
	// Zone A (local) - base rate
	"12345": 1.0, "12346": 1.0, "12347": 1.0,
	// Zone B (15+ miles) - 25% surcharge
	"12400": 1.25, "12401": 1.25, "12402": 1.25,
	// Zone C (30+ miles) - 50% surcharge  
	"12500": 1.5, "12501": 1.5, "12502": 1.5,
}

// Intelligent pricing algorithm
func calculateIntelligentPricing(req PricingRequest) PricingResponse {
	response := PricingResponse{
		ServiceBreakdown: make(map[string]float64),
		Discounts:        make(map[string]float64),
		Surcharges:       make(map[string]float64),
	}
	
	// 1. Determine base rate by property size
	baseRate := getBaseRateByPropertySize(req.PropertySize)
	response.BaseRate = baseRate
	
	var totalHours float64
	var subtotal float64
	
	// 2. Calculate service costs with complexity multipliers
	for _, serviceType := range req.ServiceTypes {
		if config, exists := serviceConfigs[serviceType]; exists {
			// Calculate service cost: base_rate * multiplier * estimated_hours
			serviceCost := baseRate * config.Multiplier * config.BaseHours
			response.ServiceBreakdown[config.Name] = serviceCost
			subtotal += serviceCost
			totalHours += config.BaseHours
		}
	}
	
	response.Subtotal = subtotal
	response.EstimatedHours = totalHours
	
	// 3. Apply frequency discounts
	if discount := getFrequencyDiscount(req.Frequency); discount > 0 {
		discountAmount := subtotal * discount
		response.Discounts["Frequency Discount"] = discountAmount
		subtotal -= discountAmount
	}
	
	// 4. Apply geographic zone pricing
	if zoneMultiplier, exists := geographicZones[req.ZipCode]; exists && zoneMultiplier > 1.0 {
		surchargeAmount := subtotal * (zoneMultiplier - 1.0)
		response.Surcharges["Geographic Zone"] = surchargeAmount
		subtotal += surchargeAmount
	}
	
	// 5. Apply smart discounting
	
	// Same-day discount: 15% off if booking within same geographic area
	if req.SameDayBooking {
		discountAmount := subtotal * 0.15
		response.Discounts["Same-Day Booking"] = discountAmount
		subtotal -= discountAmount
	}
	
	// Route premium: 25% surcharge for jobs outside efficient routes
	if req.Distance > 15 {
		surchargeAmount := subtotal * 0.25
		response.Surcharges["Route Premium"] = surchargeAmount
		subtotal += surchargeAmount
	}
	
	// Bulk discount: 10% off for multiple services
	if req.MultipleServices && len(req.ServiceTypes) > 1 {
		discountAmount := subtotal * 0.10
		response.Discounts["Multiple Services"] = discountAmount
		subtotal -= discountAmount
	}
	
	response.TotalAmount = math.Round(subtotal*100) / 100 // Round to 2 decimal places
	
	return response
}

// Get base hourly rate by property size
func getBaseRateByPropertySize(squareFootage int) float64 {
	if squareFootage < 2000 {
		return 45.0 // Small properties
	} else if squareFootage <= 5000 {
		return 55.0 // Medium properties  
	} else {
		return 65.0 // Large properties
	}
}

// Get frequency discount percentage
func getFrequencyDiscount(frequency string) float64 {
	switch strings.ToLower(frequency) {
	case "weekly":
		return 0.15 // 15% off
	case "bi-weekly", "biweekly":
		return 0.10 // 10% off
	case "monthly":
		return 0.05 // 5% off
	default:
		return 0.0 // No discount for one-time
	}
}

// Admin middleware
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")
		userEmail, exists := session.Values["user_email"]
		
		if !exists || userEmail == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		
		// Verify user still exists
		if _, userExists := users[userEmail.(string)]; !userExists {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		
		next(w, r)
	}
}

func requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, _ := store.Get(r, "session-name")
		userEmail, exists := session.Values["user_email"]
		
		if !exists || userEmail == nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		
		user, userExists := users[userEmail.(string)]
		if !userExists {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		
		// Real role-based access control
		if !user.IsAdmin {
			http.Error(w, "Access denied: Admin privileges required", http.StatusForbidden)
			return
		}
		
		next(w, r)
	}
}

func main() {
	// Initialize basic teams and employees only
	initializeTeamsAndEmployees()
	
	r := mux.NewRouter()

	// Parse templates from filesystem with explicit parsing
	var err error
	tmpl = template.New("")
	
	// Get the directory of the executable and construct absolute paths
	templateDir := filepath.Join("backend", "web", "templates")
	
	// Parse all templates
	tmpl, err = tmpl.ParseFiles(
		filepath.Join(templateDir, "base.html"),
		filepath.Join(templateDir, "index.html"),
		filepath.Join(templateDir, "service-detail.html"),
		filepath.Join(templateDir, "services.html"),
		filepath.Join(templateDir, "booking.html"),
		filepath.Join(templateDir, "login.html"),
		filepath.Join(templateDir, "signup.html"),
		filepath.Join(templateDir, "admin-dashboard.html"),
	)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
	
	// Debug: Log available templates
	for _, t := range tmpl.Templates() {
		log.Printf("Loaded template: %s", t.Name())
	}

	// Static routes
	staticPath := filepath.Join("backend", "web", "static")
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// Page routes
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title: "LandscapePro - Professional Landscaping Services",
			Page:  "index",
		}
		
		err := tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			log.Printf("Error executing index template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	r.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title: "Our Services - LandscapePro",
			Page:  "services",
		}
		err := tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			log.Printf("Error executing services template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})
	
	r.HandleFunc("/booking", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title: "Book Service - LandscapePro",
			Page:  "booking",
		}
		err := tmpl.ExecuteTemplate(w, "base.html", data)
		if err != nil {
			log.Printf("Error executing booking template: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	})

	// HTMX API endpoints
	r.HandleFunc("/api/services/featured", getFeaturedServices).Methods("GET")
	r.HandleFunc("/api/services/all", getAllServices).Methods("GET")
	r.HandleFunc("/api/quote/calculate", calculateQuote).Methods("POST")
	r.HandleFunc("/api/quote/json", calculateQuoteJSON).Methods("POST") // JSON API for advanced use
	r.HandleFunc("/api/testimonials/featured", getFeaturedTestimonials).Methods("GET")
	r.HandleFunc("/api/auth/status", getAuthStatus).Methods("GET")
	r.HandleFunc("/api/booking/submit", submitBooking).Methods("POST")
	r.HandleFunc("/booking/form", getBookingForm).Methods("GET")
	r.HandleFunc("/booking/consultation", getConsultationForm).Methods("GET")
	
	// Service detail pages
	r.HandleFunc("/services/{id}", getServiceDetail).Methods("GET")

	// Authentication routes
	r.HandleFunc("/login", getLoginPage).Methods("GET")
	r.HandleFunc("/login", handleLogin).Methods("POST")
	r.HandleFunc("/signup", getSignupPage).Methods("GET")
	r.HandleFunc("/signup", handleSignup).Methods("POST")
	r.HandleFunc("/logout", handleLogout).Methods("POST")

	// Admin routes
	r.HandleFunc("/admin", requireAdmin(getAdminDashboard)).Methods("GET")
	r.HandleFunc("/admin/requests", requireAdmin(getAdminRequests)).Methods("GET")
	r.HandleFunc("/admin/customers", requireAdmin(getAdminCustomers)).Methods("GET")
	r.HandleFunc("/admin/jobs", requireAdmin(getAdminJobs)).Methods("GET")
	r.HandleFunc("/admin/services", requireAdmin(getAdminServices)).Methods("GET")
	r.HandleFunc("/admin/team", requireAdmin(getAdminTeam)).Methods("GET")
	r.HandleFunc("/admin/reports", requireAdmin(getAdminReports)).Methods("GET")
	
	// Admin API routes
	r.HandleFunc("/admin/api/requests/{id}/accept", requireAdmin(acceptServiceRequest)).Methods("POST")
	r.HandleFunc("/admin/api/requests/{id}/deny", requireAdmin(denyServiceRequest)).Methods("POST")
	r.HandleFunc("/admin/api/customers", requireAdmin(createCustomer)).Methods("POST")
	r.HandleFunc("/admin/api/customers/{id}", requireAdmin(updateCustomer)).Methods("PUT")
	r.HandleFunc("/admin/api/customers/{id}", requireAdmin(deleteCustomer)).Methods("DELETE")
	r.HandleFunc("/admin/api/jobs", requireAdmin(createJob)).Methods("POST")
	r.HandleFunc("/admin/api/jobs/{id}", requireAdmin(updateJob)).Methods("PUT")
	r.HandleFunc("/admin/api/jobs/{id}", requireAdmin(deleteJob)).Methods("DELETE")
	r.HandleFunc("/admin/api/employees", requireAdmin(createEmployee)).Methods("POST")
	r.HandleFunc("/admin/api/employees/{id}", requireAdmin(updateEmployee)).Methods("PUT")
	r.HandleFunc("/admin/api/employees/{id}", requireAdmin(deleteEmployee)).Methods("DELETE")
	
	// Admin HTMX partials
	r.HandleFunc("/admin/partials/overview", requireAdmin(getOverviewPartial)).Methods("GET")
	r.HandleFunc("/admin/partials/requests", requireAdmin(getRequestsPartial)).Methods("GET")
	r.HandleFunc("/admin/partials/customers", requireAdmin(getCustomersPartial)).Methods("GET")
	r.HandleFunc("/admin/partials/jobs", requireAdmin(getJobsPartial)).Methods("GET")
	r.HandleFunc("/admin/partials/services", requireAdmin(getServicesPartial)).Methods("GET")
	r.HandleFunc("/admin/partials/team", requireAdmin(getTeamPartial)).Methods("GET")
	r.HandleFunc("/admin/partials/reports", requireAdmin(getReportsPartial)).Methods("GET")
	
	// Admin form partials
	r.HandleFunc("/admin/partials/customer-form", requireAdmin(getCustomerForm)).Methods("GET")
	r.HandleFunc("/admin/partials/job-form", requireAdmin(getJobForm)).Methods("GET")
	r.HandleFunc("/admin/partials/employee-form", requireAdmin(getEmployeeForm)).Methods("GET")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	log.Printf("🌿 LandscapePro web server starting on port %s", port)
	log.Printf("🌐 Visit http://localhost:%s to view the site", port)

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// HTMX handlers
func getFeaturedServices(w http.ResponseWriter, r *http.Request) {
	services := []Service{
		{
			ID:          "1",
			Name:        "Lawn Care",
			Description: "Professional mowing, edging, and maintenance",
			Icon:        "🌱",
			Price:       "From $50/visit",
		},
		{
			ID:          "2", 
			Name:        "Garden Design",
			Description: "Custom landscape design and installation",
			Icon:        "🌺",
			Price:       "Free consultation",
		},
		{
			ID:          "3",
			Name:        "Tree Service",
			Description: "Trimming, removal, and health assessment",
			Icon:        "🌳",
			Price:       "From $200",
		},
		{
			ID:          "4",
			Name:        "Irrigation",
			Description: "Sprinkler system design and repair",
			Icon:        "💧",
			Price:       "From $150",
		},
	}

	html := ""
	for _, service := range services {
		html += `
		<div class="bg-white rounded-lg shadow-md p-6 hover:shadow-xl transition cursor-pointer"
		     hx-get="/services/` + service.ID + `"
		     hx-push-url="true">
			<div class="text-4xl mb-4">` + service.Icon + `</div>
			<h3 class="text-xl font-semibold mb-2">` + service.Name + `</h3>
			<p class="text-gray-600 mb-3">` + service.Description + `</p>
			<p class="text-primary font-semibold">` + service.Price + `</p>
		</div>`
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func calculateQuote(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Extract pricing parameters from form
	serviceName := r.FormValue("service")
	frequencyStr := r.FormValue("frequency")
	zipCode := r.FormValue("zip_code")
	propertySizeStr := r.FormValue("property_size")
	sameDayStr := r.FormValue("same_day")
	
	// Set defaults if not provided
	if serviceName == "" {
		serviceName = "lawn_care"
	}
	if frequencyStr == "" {
		frequencyStr = "one-time"
	}
	if zipCode == "" {
		zipCode = "12345" // Default to Zone A
	}
	
	// Parse property size
	propertySize := 3000 // Default medium property
	if propertySizeStr != "" {
		if size, err := strconv.Atoi(propertySizeStr); err == nil {
			propertySize = size
		}
	}
	
	// Build pricing request
	pricingReq := PricingRequest{
		PropertySize:     propertySize,
		ServiceTypes:     []string{serviceName},
		Frequency:        frequencyStr,
		ZipCode:          zipCode,
		SameDayBooking:   sameDayStr == "true",
		MultipleServices: false,
		Distance:         0, // Default local
	}
	
	// Calculate intelligent pricing
	pricing := calculateIntelligentPricing(pricingReq)
	
	// Generate breakdown HTML
	var breakdownHTML strings.Builder
	
	// Service breakdown
	for serviceName, cost := range pricing.ServiceBreakdown {
		breakdownHTML.WriteString(fmt.Sprintf(
			`<div class="flex justify-between"><span>%s</span><span>$%.2f</span></div>`,
			serviceName, cost))
	}
	
	// Discounts
	for discountName, amount := range pricing.Discounts {
		if amount > 0 {
			breakdownHTML.WriteString(fmt.Sprintf(
				`<div class="flex justify-between text-green-600"><span>%s</span><span>-$%.2f</span></div>`,
				discountName, amount))
		}
	}
	
	// Surcharges
	for surchargeName, amount := range pricing.Surcharges {
		if amount > 0 {
			breakdownHTML.WriteString(fmt.Sprintf(
				`<div class="flex justify-between text-orange-600"><span>%s</span><span>+$%.2f</span></div>`,
				surchargeName, amount))
		}
	}
	
	// Generate pricing display HTML
	html := fmt.Sprintf(`
	<div class="bg-green-50 border border-green-200 rounded-lg p-6">
		<h3 class="text-2xl font-bold text-green-800 mb-4 text-center">Intelligent Quote</h3>
		
		<div class="bg-white rounded-lg p-4 mb-4">
			<div class="space-y-2 text-sm">
				<div class="flex justify-between font-semibold border-b pb-2">
					<span>Base Rate (Property: %d sq ft)</span>
					<span>$%.2f/hour</span>
				</div>
				%s
				<div class="flex justify-between font-bold text-lg border-t pt-2">
					<span>Total Amount</span>
					<span class="text-primary">$%.2f</span>
				</div>
				<div class="text-center text-sm text-gray-600">
					Estimated Duration: %.1f hours
				</div>
			</div>
		</div>
		
		<div class="text-center">
			<button hx-get="/booking/form" 
			        hx-target="#booking-modal"
			        class="bg-primary text-white px-6 py-2 rounded-lg hover:bg-secondary transition mr-2">
				Book This Service
			</button>
			<button onclick="this.parentElement.parentElement.innerHTML = ''" 
			        class="bg-gray-500 text-white px-4 py-2 rounded-lg hover:bg-gray-600 transition">
				Close
			</button>
		</div>
	</div>`, 
		propertySize, 
		pricing.BaseRate,
		breakdownHTML.String(),
		pricing.TotalAmount,
		pricing.EstimatedHours)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// JSON API for quote calculation
func calculateQuoteJSON(w http.ResponseWriter, r *http.Request) {
	var req PricingRequest
	
	// Parse JSON request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON request", http.StatusBadRequest)
		return
	}
	
	// Validate required fields
	if len(req.ServiceTypes) == 0 {
		req.ServiceTypes = []string{"lawn_care"} // Default service
	}
	if req.PropertySize == 0 {
		req.PropertySize = 3000 // Default medium property
	}
	if req.Frequency == "" {
		req.Frequency = "one-time"
	}
	if req.ZipCode == "" {
		req.ZipCode = "12345" // Default Zone A
	}
	
	// Calculate intelligent pricing
	pricing := calculateIntelligentPricing(req)
	
	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(pricing)
}

func getFeaturedTestimonials(w http.ResponseWriter, r *http.Request) {
	testimonials := []Testimonial{
		{
			Name:    "Sarah Johnson",
			Role:    "Homeowner",
			Content: "LandscapePro transformed our backyard into a beautiful oasis. Professional, timely, and great value!",
			Rating:  5,
		},
		{
			Name:    "Mike Chen",
			Role:    "Business Owner",
			Content: "We've been using their commercial services for 3 years. Consistently excellent work and responsive team.",
			Rating:  5,
		},
		{
			Name:    "Emily Rodriguez",
			Role:    "Property Manager",
			Content: "They handle maintenance for all our properties. Reliable, efficient, and always professional.",
			Rating:  5,
		},
	}

	html := ""
	for _, t := range testimonials {
		stars := ""
		for i := 0; i < t.Rating; i++ {
			stars += "⭐"
		}
		html += `
		<div class="bg-white rounded-lg shadow-md p-6">
			<div class="mb-4">` + stars + `</div>
			<p class="text-gray-700 mb-4 italic">"` + t.Content + `"</p>
			<div>
				<p class="font-semibold">` + t.Name + `</p>
				<p class="text-sm text-gray-600">` + t.Role + `</p>
			</div>
		</div>`
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getAuthStatus(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	userEmail, exists := session.Values["user_email"]
	
	if exists && userEmail != nil {
		// User is logged in
		user, userExists := users[userEmail.(string)]
		if userExists {
			adminLink := ""
			roleText := "Customer"
			if user.IsAdmin {
				adminLink = `<a href="/admin" class="text-primary hover:text-secondary px-3 py-2 rounded-md text-sm font-medium">Admin Dashboard</a>`
				roleText = "Admin"
			}
			html := fmt.Sprintf(`
			<div class="flex items-center space-x-4">
				<span class="text-gray-700 text-sm">Welcome, %s (%s)</span>
				%s
				<form hx-post="/logout" hx-swap="outerHTML" hx-target="#auth-buttons">
					<button type="submit" class="bg-gray-500 text-white hover:bg-gray-600 px-4 py-2 rounded-md text-sm font-medium">
						Logout
					</button>
				</form>
			</div>`, user.FirstName, roleText, adminLink)
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(html))
			return
		}
	}
	
	// User is not logged in
	html := `
	<a href="/login" class="text-gray-700 hover:text-primary px-3 py-2 rounded-md text-sm font-medium">
		Login
	</a>
	<a href="/signup" class="bg-primary text-white hover:bg-secondary px-4 py-2 rounded-md text-sm font-medium">
		Sign Up
	</a>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getBookingForm(w http.ResponseWriter, r *http.Request) {
	html := `
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50"
	     x-data="{ open: true }"
	     x-show="open"
	     @click.away="open = false">
		<div class="bg-white rounded-lg max-w-md w-full p-6" @click.stop>
			<div class="flex justify-between items-center mb-4">
				<h2 class="text-2xl font-bold">Request a Quote</h2>
				<button @click="open = false" class="text-gray-500 hover:text-gray-700">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>
			
			<form hx-post="/api/booking/submit" 
			      hx-swap="outerHTML"
			      class="space-y-4">
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Name</label>
					<input type="text" name="name" required 
					       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
				</div>
				
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
					<input type="email" name="email" required
					       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
				</div>
				
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Phone</label>
					<input type="tel" name="phone" required
					       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
				</div>
				
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Service Needed</label>
					<select name="service" required
					        class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
						<option value="">Select a service</option>
						<option value="lawn_care">Lawn Care</option>
						<option value="garden_design">Garden Design</option>
						<option value="tree_service">Tree Service</option>
						<option value="hardscaping">Hardscaping</option>
						<option value="other">Other</option>
					</select>
				</div>
				
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Message</label>
					<textarea name="message" rows="3"
					          class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"></textarea>
				</div>
				
				<button type="submit" 
				        class="w-full bg-primary text-white py-2 rounded-md hover:bg-secondary transition">
					Submit Request
				</button>
			</form>
		</div>
	</div>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getConsultationForm(w http.ResponseWriter, r *http.Request) {
	// Similar to booking form but for consultation
	getBookingForm(w, r)
}

func submitBooking(w http.ResponseWriter, r *http.Request) {
	// Parse form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")
	email := r.FormValue("email")
	phone := r.FormValue("phone")
	service := r.FormValue("service")
	message := r.FormValue("message")

	// Create real service request
	requestID := fmt.Sprintf("req_%d", len(serviceRequests)+1)
	serviceRequest := ServiceRequest{
		ID:           requestID,
		CustomerName: name,
		CustomerEmail: email,
		CustomerPhone: phone,
		ServiceType:  service,
		Message:      message,
		Status:       "pending",
		CreatedAt:    time.Now(),
		PropertyInfo: map[string]interface{}{
			"contact_method": "web_form",
		},
	}

	// Store the service request
	serviceRequests[requestID] = serviceRequest

	// Check if customer already exists or create new one
	var customerID string
	customerExists := false
	
	// Look for existing customer by email
	for id, customer := range customers {
		if customer.Email == email {
			customerID = id
			customerExists = true
			break
		}
	}

	// Create new customer if doesn't exist
	if !customerExists {
		customerID = fmt.Sprintf("cust_%d", len(customers)+1)
		customer := Customer{
			ID:           customerID,
			FirstName:    name, // We'll split this later if needed
			LastName:     "",   // Single name field for now
			Email:        email,
			Phone:        phone,
			Address:      "",   // Will be collected later
			City:         "",
			State:        "",
			ZipCode:      "",
			CreatedAt:    time.Now(),
			LastService:  time.Time{}, // Empty initially
			TotalSpent:   0.0,
			ServiceCount: 0,
		}
		customers[customerID] = customer
	}

	// Update service request with customer ID
	serviceRequest.CustomerID = customerID
	serviceRequests[requestID] = serviceRequest

	log.Printf("New service request created: %s from %s (%s) for %s", requestID, name, email, service)

	// Return success response that closes the modal
	html := `
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-white rounded-lg max-w-md w-full p-6">
			<div class="text-center">
				<div class="text-green-600 text-4xl mb-4">✅</div>
				<h3 class="text-xl font-bold text-green-800 mb-2">Request Submitted!</h3>
				<p class="text-green-700 mb-4">Thank you ` + name + `! We'll contact you within 24 hours at ` + email + ` to schedule your consultation. Your request ID is: ` + requestID + `</p>
				<button onclick="document.getElementById('booking-modal').innerHTML = ''" 
				        class="bg-green-600 text-white px-6 py-2 rounded-lg hover:bg-green-700 transition">
					Close
				</button>
			</div>
		</div>
	</div>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getAllServices(w http.ResponseWriter, r *http.Request) {
	services := []Service{
		{
			ID:          "1",
			Name:        "Lawn Care & Maintenance",
			Description: "Complete lawn care including mowing, edging, fertilization, and seasonal cleanup. Keep your lawn healthy and beautiful year-round.",
			Icon:        "🌱",
			Price:       "From $50/visit",
			Image:       "https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=400",
		},
		{
			ID:          "2",
			Name:        "Garden Design & Installation",
			Description: "Custom landscape design services from concept to completion. We create beautiful, sustainable gardens.",
			Icon:        "🌺",
			Price:       "Free consultation",
			Image:       "https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=400",
		},
		{
			ID:          "3",
			Name:        "Tree Service & Removal",
			Description: "Professional tree care including pruning, removal, health assessment, and emergency services.",
			Icon:        "🌳",
			Price:       "From $200",
			Image:       "https://images.unsplash.com/photo-1441974231531-c6227db76b6e?w=400",
		},
		{
			ID:          "4",
			Name:        "Irrigation Systems",
			Description: "Smart irrigation system design, installation, and maintenance. Water-efficient solutions.",
			Icon:        "💧",
			Price:       "From $150",
			Image:       "https://images.unsplash.com/photo-1585320806297-9794b3e4eeae?w=400",
		},
		{
			ID:          "5",
			Name:        "Hardscaping & Patios",
			Description: "Stone patios, walkways, retaining walls, and outdoor living spaces that enhance your property.",
			Icon:        "🏗️",
			Price:       "From $1,000",
			Image:       "https://images.unsplash.com/photo-1580587771525-78b9dba3b914?w=400",
		},
		{
			ID:          "6",
			Name:        "Snow Removal",
			Description: "Reliable winter snow removal services for driveways, walkways, and commercial properties.",
			Icon:        "❄️",
			Price:       "From $75/visit",
			Image:       "https://images.unsplash.com/photo-1578662996442-48f60103fc96?w=400",
		},
	}

	html := ""
	for _, service := range services {
		html += `
		<div class="bg-white rounded-lg shadow-md overflow-hidden hover:shadow-xl transition service-card"
		     hx-get="/services/` + service.ID + `"
		     hx-target="body"
		     hx-push-url="true">
			<img src="` + service.Image + `" alt="` + service.Name + `" class="w-full h-48 object-cover">
			<div class="p-6">
				<div class="flex items-center mb-3">
					<span class="text-3xl mr-3">` + service.Icon + `</span>
					<h3 class="text-xl font-semibold">` + service.Name + `</h3>
				</div>
				<p class="text-gray-600 mb-4">` + service.Description + `</p>
				<div class="flex items-center justify-between">
					<span class="text-lg font-bold text-primary">` + service.Price + `</span>
					<button class="text-primary hover:text-secondary font-medium">
						Learn More →
					</button>
				</div>
			</div>
		</div>`
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getServiceDetail(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["id"]

	// Mock service data
	services := map[string]Service{
		"1": {
			ID:          "1",
			Name:        "Lawn Care",
			Description: "Complete lawn maintenance including mowing, edging, fertilization, and seasonal cleanup. Our certified professionals use commercial-grade equipment to keep your lawn healthy and beautiful year-round.",
			Icon:        "🌱",
			Price:       "From $50/visit",
			Image:       "https://images.unsplash.com/photo-1558618666-fcd25c85cd64?w=800",
		},
		"2": {
			ID:          "2",
			Name:        "Garden Design",
			Description: "Custom landscape design services from concept to completion. We create beautiful, sustainable gardens that reflect your style and complement your property's natural features.",
			Icon:        "🌺",
			Price:       "Free consultation",
			Image:       "https://images.unsplash.com/photo-1416879595882-3373a0480b5b?w=800",
		},
		"3": {
			ID:          "3",
			Name:        "Tree Service",
			Description: "Professional tree care including pruning, removal, health assessment, and emergency services. Our certified arborists ensure your trees remain healthy and safe.",
			Icon:        "🌳",
			Price:       "From $200",
			Image:       "https://images.unsplash.com/photo-1441974231531-c6227db76b6e?w=800",
		},
		"4": {
			ID:          "4",
			Name:        "Irrigation",
			Description: "Smart irrigation system design, installation, and maintenance. Water-efficient solutions that keep your landscape healthy while reducing water usage and costs.",
			Icon:        "💧",
			Price:       "From $150",
			Image:       "https://images.unsplash.com/photo-1585320806297-9794b3e4eeae?w=800",
		},
	}

	service, exists := services[serviceID]
	if !exists {
		http.NotFound(w, r)
		return
	}

	// Use proper template with base layout
	data := PageData{
		Title: service.Name + " - LandscapePro",
		Page:  "service-detail",
		Data: map[string]interface{}{
			"Service": service,
		},
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing service detail template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// Authentication handlers
func getLoginPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Login - LandscapePro",
		Page:  "login",
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing login template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getSignupPage(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Sign Up - LandscapePro",
		Page:  "signup",
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing signup template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		showError(w, "Invalid form data")
		return
	}

	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	password := r.FormValue("password")
	
	// Validate input
	if email == "" || password == "" {
		showError(w, "Please fill in all fields")
		return
	}

	// Find user by email
	user, exists := users[email]
	if !exists {
		showError(w, "Invalid email or password")
		return
	}
	
	// Verify password using bcrypt
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		showError(w, "Invalid email or password")
		return
	}

	// Create session
	session, _ := store.Get(r, "session-name")
	session.Values["user_email"] = user.Email
	session.Values["user_id"] = user.ID
	session.Values["authenticated"] = true
	session.Save(r, w)

	// Success response with redirect
	html := fmt.Sprintf(`
	<div id="login-form" class="rounded-md shadow-sm space-y-6">
		<div class="bg-green-50 border border-green-200 rounded-lg p-4 text-center">
			<div class="text-green-600 text-2xl mb-2">✅</div>
			<h3 class="text-lg font-bold text-green-800 mb-2">Login Successful!</h3>
			<p class="text-green-700 mb-4">Welcome back, %s! Redirecting to homepage...</p>
		</div>
	</div>
	<script>
		setTimeout(function() {
			window.location.href = "/";
		}, 2000);
	</script>`, user.FirstName)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleSignup(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		showSignupError(w, "Invalid form data")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")
	firstName := r.FormValue("first_name")
	lastName := r.FormValue("last_name")
	phone := r.FormValue("phone")
	terms := r.FormValue("terms")
	
	// Validation
	if email == "" || password == "" || firstName == "" || lastName == "" {
		showSignupError(w, "Please fill in all required fields")
		return
	}
	
	if password != confirmPassword {
		showSignupError(w, "Passwords do not match")
		return
	}
	
	if terms != "on" {
		showSignupError(w, "Please accept the Terms of Service")
		return
	}
	
	// Check if user already exists
	for _, u := range users {
		if u.Email == email {
			showSignupError(w, "An account with this email already exists")
			return
		}
	}

	// Create new user
	user := User{
		ID:        fmt.Sprintf("user_%d", len(users)+1),
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Phone:     phone,
		IsAdmin:   true, // For demo: all users are admin
	}
	users[user.ID] = user

	// Create session
	session, _ := store.Get(r, "session-name")
	session.Values["user_id"] = user.ID
	session.Values["authenticated"] = true
	session.Save(r, w)

	// Success response
	html := `
	<div id="signup-form" class="space-y-6">
		<div class="bg-green-50 border border-green-200 rounded-lg p-4 text-center">
			<div class="text-green-600 text-2xl mb-2">🎉</div>
			<h3 class="text-lg font-bold text-green-800 mb-2">Account Created Successfully!</h3>
			<p class="text-green-700 mb-4">Welcome to LandscapePro, ` + firstName + `! Redirecting to homepage...</p>
		</div>
	</div>
	<script>
		setTimeout(function() {
			window.location.href = "/";
		}, 2000);
	</script>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "session-name")
	delete(session.Values, "user_email")
	delete(session.Values, "user_id")
	delete(session.Values, "authenticated")
	session.Save(r, w)

	// Return logged-out auth buttons
	html := `
	<a href="/login" class="text-gray-700 hover:text-primary px-3 py-2 rounded-md text-sm font-medium">
		Login
	</a>
	<a href="/signup" class="bg-primary text-white hover:bg-secondary px-4 py-2 rounded-md text-sm font-medium">
		Sign Up
	</a>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func showError(w http.ResponseWriter, message string) {
	html := fmt.Sprintf(`
	<div id="login-form" class="rounded-md shadow-sm -space-y-px">
		<div id="login-messages" class="mb-4">
			<div class="bg-red-50 border border-red-200 rounded-lg p-4">
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
						</svg>
					</div>
					<div class="ml-3">
						<p class="text-sm text-red-700">%s</p>
					</div>
				</div>
			</div>
		</div>
		<div>
			<label for="email" class="sr-only">Email address</label>
			<input id="email" name="email" type="email" autocomplete="email" required 
				class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-t-md focus:outline-none focus:ring-primary focus:border-primary focus:z-10 sm:text-sm" 
				placeholder="Email address">
		</div>
		<div>
			<label for="password" class="sr-only">Password</label>
			<input id="password" name="password" type="password" autocomplete="current-password" required 
				class="relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-b-md focus:outline-none focus:ring-primary focus:border-primary focus:z-10 sm:text-sm" 
				placeholder="Password">
		</div>
	</div>`, message)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func showSignupError(w http.ResponseWriter, message string) {
	html := fmt.Sprintf(`
	<div id="signup-form" class="space-y-4">
		<div id="signup-messages" class="mb-4">
			<div class="bg-red-50 border border-red-200 rounded-lg p-4">
				<div class="flex">
					<div class="flex-shrink-0">
						<svg class="h-5 w-5 text-red-400" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clip-rule="evenodd" />
						</svg>
					</div>
					<div class="ml-3">
						<p class="text-sm text-red-700">%s</p>
					</div>
				</div>
			</div>
		</div>
		<div class="grid grid-cols-2 gap-4">
			<div>
				<label for="first_name" class="block text-sm font-medium text-gray-700">First name</label>
				<input id="first_name" name="first_name" type="text" autocomplete="given-name" required 
					class="mt-1 block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-primary focus:border-primary sm:text-sm" 
					placeholder="John">
			</div>
			<div>
				<label for="last_name" class="block text-sm font-medium text-gray-700">Last name</label>
				<input id="last_name" name="last_name" type="text" autocomplete="family-name" required 
					class="mt-1 block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-primary focus:border-primary sm:text-sm" 
					placeholder="Doe">
			</div>
		</div>
		<div>
			<label for="email" class="block text-sm font-medium text-gray-700">Email address</label>
			<input id="email" name="email" type="email" autocomplete="email" required 
				class="mt-1 block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-primary focus:border-primary sm:text-sm" 
				placeholder="john@example.com">
		</div>
		<div>
			<label for="phone" class="block text-sm font-medium text-gray-700">Phone number</label>
			<input id="phone" name="phone" type="tel" autocomplete="tel" 
				class="mt-1 block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-primary focus:border-primary sm:text-sm" 
				placeholder="(555) 123-4567">
		</div>
		<div>
			<label for="password" class="block text-sm font-medium text-gray-700">Password</label>
			<input id="password" name="password" type="password" autocomplete="new-password" required 
				class="mt-1 block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-primary focus:border-primary sm:text-sm" 
				placeholder="Create a secure password">
		</div>
		<div>
			<label for="confirm_password" class="block text-sm font-medium text-gray-700">Confirm password</label>
			<input id="confirm_password" name="confirm_password" type="password" autocomplete="new-password" required 
				class="mt-1 block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 rounded-md focus:outline-none focus:ring-primary focus:border-primary sm:text-sm" 
				placeholder="Confirm your password">
		</div>
		<div class="flex items-start">
			<div class="flex items-center h-5">
				<input id="terms" name="terms" type="checkbox" required
					class="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded">
			</div>
			<div class="ml-3 text-sm">
				<label for="terms" class="text-gray-600">
					I agree to the <a href="#" class="text-primary hover:text-secondary font-medium">Terms of Service</a> and <a href="#" class="text-primary hover:text-secondary font-medium">Privacy Policy</a>
				</label>
			</div>
		</div>
		<div class="flex items-start">
			<div class="flex items-center h-5">
				<input id="marketing" name="marketing" type="checkbox" 
					class="h-4 w-4 text-primary focus:ring-primary border-gray-300 rounded">
			</div>
			<div class="ml-3 text-sm">
				<label for="marketing" class="text-gray-600">
					I'd like to receive updates about services and special offers
				</label>
			</div>
		</div>
	</div>`, message)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}