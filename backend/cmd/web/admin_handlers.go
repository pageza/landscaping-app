package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// Initialize only teams and employees - customers and jobs will be created from real interactions
func initializeTeamsAndEmployees() {
	now := time.Now()
	
	// Initialize teams
	teams["1"] = Team{
		ID:          "1",
		Name:        "Lawn Care Team A",
		Members:     []string{"1", "2"},
		Specialties: []string{"lawn_care", "edging", "fertilization"},
		Active:      true,
	}
	
	teams["2"] = Team{
		ID:          "2",
		Name:        "Landscaping Crew",
		Members:     []string{"3", "4"},
		Specialties: []string{"garden_design", "planting", "hardscaping"},
		Active:      true,
	}
	
	// Initialize employees
	employees["1"] = Employee{
		ID:        "1",
		FirstName: "Tom",
		LastName:  "Wilson",
		Email:     "tom.wilson@landscapepro.com",
		Phone:     "(555) 111-2222",
		Role:      "Crew Leader",
		HireDate:  now.AddDate(-2, 0, 0),
		Active:    true,
		TeamID:    "1",
	}
	
	employees["2"] = Employee{
		ID:        "2",
		FirstName: "Lisa",
		LastName:  "Brown",
		Email:     "lisa.brown@landscapepro.com",
		Phone:     "(555) 222-3333",
		Role:      "Landscaper",
		HireDate:  now.AddDate(-1, -6, 0),
		Active:    true,
		TeamID:    "1",
	}
	
	employees["3"] = Employee{
		ID:        "3",
		FirstName: "David",
		LastName:  "Garcia",
		Email:     "david.garcia@landscapepro.com",
		Phone:     "(555) 333-4444",
		Role:      "Designer",
		HireDate:  now.AddDate(-3, 0, 0),
		Active:    true,
		TeamID:    "2",
	}
	
	employees["4"] = Employee{
		ID:        "4",
		FirstName: "Emily",
		LastName:  "Davis",
		Email:     "emily.davis@landscapepro.com",
		Phone:     "(555) 444-5555",
		Role:      "Hardscape Specialist",
		HireDate:  now.AddDate(-1, -2, 0),
		Active:    true,
		TeamID:    "2",
	}
	
	// Note: customers and jobs will be created from real user interactions
	// serviceRequests will be populated when customers submit booking forms
}

// Admin Dashboard Handlers
func getAdminDashboard(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Admin Dashboard - LandscapePro",
		Page:  "admin-dashboard",
		Data: map[string]interface{}{
			"Section": "overview",
		},
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing admin dashboard template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getAdminRequests(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Service Requests - LandscapePro",
		Page:  "admin-dashboard",
		Data: map[string]interface{}{
			"Section": "requests",
		},
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing admin requests template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getAdminCustomers(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Customer Management - LandscapePro",
		Page:  "admin-dashboard",
		Data: map[string]interface{}{
			"Section": "customers",
		},
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing admin customers template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getAdminJobs(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Job Management - LandscapePro",
		Page:  "admin-dashboard",
		Data: map[string]interface{}{
			"Section": "jobs",
		},
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing admin jobs template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getAdminServices(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Service Management - LandscapePro",
		Page:  "admin-dashboard",
		Data: map[string]interface{}{
			"Section": "services",
		},
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing admin services template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getAdminTeam(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Team Management - LandscapePro",
		Page:  "admin-dashboard",
		Data: map[string]interface{}{
			"Section": "team",
		},
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing admin team template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func getAdminReports(w http.ResponseWriter, r *http.Request) {
	data := PageData{
		Title: "Reports & Analytics - LandscapePro",
		Page:  "admin-dashboard",
		Data: map[string]interface{}{
			"Section": "reports",
		},
	}
	
	err := tmpl.ExecuteTemplate(w, "base.html", data)
	if err != nil {
		log.Printf("Error executing admin reports template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// HTMX Partial Handlers
func getOverviewPartial(w http.ResponseWriter, r *http.Request) {
	totalCustomers := len(customers)
	activeJobs := 0
	totalRevenue := 0.0
	pendingRequests := 0
	
	// Count pending service requests
	for _, request := range serviceRequests {
		if request.Status == "pending" {
			pendingRequests++
		}
	}
	
	// Count active jobs and calculate revenue from completed jobs only
	for _, job := range jobs {
		if job.Status == "in-progress" || job.Status == "scheduled" {
			activeJobs++
		}
		if job.Status == "completed" {
			totalRevenue += job.Price
		}
	}
	
	html := fmt.Sprintf(`
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex items-center">
				<div class="p-3 rounded-full bg-orange-100 text-orange-600">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-5 5-5-5h5v-5a7.5 7.5 0 0 1 7.5-7.5h0a7.5 7.5 0 0 1 7.5 7.5"></path>
					</svg>
				</div>
				<div class="ml-4">
					<h3 class="text-lg font-medium text-gray-900">Pending Requests</h3>
					<p class="text-3xl font-bold text-gray-900">%d</p>
				</div>
			</div>
		</div>
		
		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex items-center">
				<div class="p-3 rounded-full bg-blue-100 text-blue-600">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"></path>
					</svg>
				</div>
				<div class="ml-4">
					<h3 class="text-lg font-medium text-gray-900">Total Customers</h3>
					<p class="text-3xl font-bold text-gray-900">%d</p>
				</div>
			</div>
		</div>
		
		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex items-center">
				<div class="p-3 rounded-full bg-green-100 text-green-600">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5H7a2 2 0 00-2 2v2m4-4h8a2 2 0 012 2v12a2 2 0 01-2 2H7a2 2 0 01-2-2v-2m4-4h8m-8-4h8"></path>
					</svg>
				</div>
				<div class="ml-4">
					<h3 class="text-lg font-medium text-gray-900">Active Jobs</h3>
					<p class="text-3xl font-bold text-gray-900">%d</p>
				</div>
			</div>
		</div>
		
		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex items-center">
				<div class="p-3 rounded-full bg-yellow-100 text-yellow-600">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1"></path>
					</svg>
				</div>
				<div class="ml-4">
					<h3 class="text-lg font-medium text-gray-900">Revenue</h3>
					<p class="text-3xl font-bold text-gray-900">$%.0f</p>
				</div>
			</div>
		</div>
	</div>
	
	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
		<div class="bg-white rounded-lg shadow p-6">
			<h3 class="text-lg font-medium text-gray-900 mb-4">Recent Service Requests</h3>
			<div class="space-y-4">`, pendingRequests, totalCustomers, activeJobs, totalRevenue)

	// Add recent service requests
	requestCount := 0
	for _, request := range serviceRequests {
		if requestCount >= 5 { // Limit to 5 recent requests
			break
		}
		
		statusColor := "gray"
		switch request.Status {
		case "pending":
			statusColor = "orange"
		case "accepted":
			statusColor = "green"
		case "denied":
			statusColor = "red"
		case "scheduled":
			statusColor = "blue"
		}
		
		serviceTypeDisplay := strings.Replace(request.ServiceType, "_", " ", -1)
		serviceTypeDisplay = strings.Title(serviceTypeDisplay)
		
		html += fmt.Sprintf(`
				<div class="flex items-center justify-between border-b pb-2">
					<div>
						<p class="font-medium">%s</p>
						<p class="text-sm text-gray-600">%s - %s</p>
						<p class="text-xs text-gray-500">%s</p>
					</div>
					<span class="px-2 py-1 text-xs rounded-full bg-%s-100 text-%s-800">%s</span>
				</div>`, request.CustomerName, serviceTypeDisplay, request.CustomerEmail, request.CreatedAt.Format("Jan 2, 3:04 PM"), statusColor, statusColor, request.Status)
		requestCount++
	}
	
	// Show message if no service requests
	if len(serviceRequests) == 0 {
		html += `
				<div class="text-center text-gray-500 py-4">
					<p>No service requests yet.</p>
					<p class="text-sm">Requests will appear here when customers submit booking forms.</p>
				</div>`
	}
	
	html += `
			</div>
		</div>
		
		<div class="bg-white rounded-lg shadow p-6">
			<h3 class="text-lg font-medium text-gray-900 mb-4">Recent Customers</h3>
			<div class="space-y-4">`

	// Add recent customers
	customerCount := 0
	for _, customer := range customers {
		if customerCount >= 5 { // Limit to 5 recent customers
			break
		}
		html += fmt.Sprintf(`
				<div class="flex items-center justify-between border-b pb-2">
					<div>
						<p class="font-medium">%s %s</p>
						<p class="text-sm text-gray-600">%s</p>
						<p class="text-xs text-gray-500">Joined %s</p>
					</div>
					<span class="text-sm font-medium text-green-600">$%.0f spent</span>
				</div>`, customer.FirstName, customer.LastName, customer.Email, customer.CreatedAt.Format("Jan 2"), customer.TotalSpent)
		customerCount++
	}
	
	// Show message if no customers
	if len(customers) == 0 {
		html += `
				<div class="text-center text-gray-500 py-4">
					<p>No customers yet.</p>
					<p class="text-sm">Customers will appear here when they submit service requests.</p>
				</div>`
	}
	
	html += `
			</div>
		</div>
	</div>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getRequestsPartial(w http.ResponseWriter, r *http.Request) {
	html := `
	<div class="flex justify-between items-center mb-6">
		<h2 class="text-2xl font-bold text-gray-900">Service Requests</h2>
		<div class="text-sm text-gray-600">
			<span class="bg-orange-100 text-orange-800 px-2 py-1 rounded">Pending</span>
			<span class="bg-green-100 text-green-800 px-2 py-1 rounded ml-1">Accepted</span>
			<span class="bg-red-100 text-red-800 px-2 py-1 rounded ml-1">Denied</span>
		</div>
	</div>
	
	<div class="bg-white rounded-lg shadow overflow-hidden">
		<div class="p-4 border-b">
			<input type="text" placeholder="Search service requests..." 
			       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
		</div>
		
		<div id="request-list">`

	if len(serviceRequests) == 0 {
		html += `
			<div class="p-8 text-center text-gray-500">
				<svg class="w-16 h-16 mx-auto mb-4 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
					<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"></path>
				</svg>
				<h3 class="text-lg font-medium mb-2">No Service Requests Yet</h3>
				<p>When customers submit booking forms, their requests will appear here.</p>
				<p class="text-sm mt-2">You'll be able to accept, deny, or schedule these requests as jobs.</p>
			</div>`
	} else {
		html += `
			<div class="divide-y divide-gray-200">`
		
		for _, request := range serviceRequests {
			statusColor := "gray"
			statusText := request.Status
			actionButtons := ""
			
			switch request.Status {
			case "pending":
				statusColor = "orange"
				actionButtons = fmt.Sprintf(`
					<div class="flex space-x-2 mt-2">
						<button hx-post="/admin/api/requests/%s/accept" 
						        hx-target="closest .request-card"
						        hx-swap="outerHTML"
						        class="bg-green-600 text-white px-3 py-1 rounded text-sm hover:bg-green-700">
							Accept
						</button>
						<button hx-post="/admin/api/requests/%s/deny" 
						        hx-target="closest .request-card"
						        hx-swap="outerHTML"
						        class="bg-red-600 text-white px-3 py-1 rounded text-sm hover:bg-red-700">
							Deny
						</button>
					</div>`, request.ID, request.ID)
			case "accepted":
				statusColor = "green"
				actionButtons = `
					<div class="mt-2">
						<button hx-get="/admin/partials/job-form" 
						        hx-target="#modal-content"
						        class="bg-blue-600 text-white px-3 py-1 rounded text-sm hover:bg-blue-700">
							Schedule Job
						</button>
					</div>`
			case "denied":
				statusColor = "red"
			}
			
			serviceTypeDisplay := strings.Replace(request.ServiceType, "_", " ", -1)
			serviceTypeDisplay = strings.Title(serviceTypeDisplay)
			
			html += fmt.Sprintf(`
				<div class="request-card p-6">
					<div class="flex items-start justify-between">
						<div class="flex-1">
							<div class="flex items-center justify-between mb-2">
								<h3 class="text-lg font-medium text-gray-900">%s</h3>
								<span class="px-2 py-1 text-xs rounded-full bg-%s-100 text-%s-800">%s</span>
							</div>
							<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm text-gray-600 mb-2">
								<div>
									<p><strong>Email:</strong> %s</p>
									<p><strong>Phone:</strong> %s</p>
									<p><strong>Service:</strong> %s</p>
								</div>
								<div>
									<p><strong>Submitted:</strong> %s</p>
									<p><strong>Request ID:</strong> %s</p>
								</div>
							</div>
							<div class="mb-2">
								<p class="text-sm text-gray-600"><strong>Message:</strong></p>
								<p class="text-sm text-gray-800 bg-gray-50 p-2 rounded">%s</p>
							</div>
							%s
						</div>
					</div>
				</div>`, 
				request.CustomerName, statusColor, statusColor, statusText,
				request.CustomerEmail, request.CustomerPhone, serviceTypeDisplay,
				request.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"), request.ID,
				request.Message, actionButtons)
		}
		
		html += `
			</div>`
	}

	html += `
		</div>
	</div>
	
	<!-- Modal -->
	<div id="modal-content"></div>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getCustomersPartial(w http.ResponseWriter, r *http.Request) {
	html := `
	<div class="flex justify-between items-center mb-6">
		<h2 class="text-2xl font-bold text-gray-900">Customer Management</h2>
		<button hx-get="/admin/partials/customer-form" 
		        hx-target="#modal-content" 
		        hx-trigger="click"
		        class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-secondary transition">
			Add Customer
		</button>
	</div>
	
	<div class="bg-white rounded-lg shadow overflow-hidden">
		<div class="p-4 border-b">
			<input type="text" placeholder="Search customers..." 
			       hx-get="/admin/partials/customers" 
			       hx-trigger="keyup changed delay:300ms" 
			       hx-target="#customer-list"
			       name="search"
			       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
		</div>
		
		<div id="customer-list">
			<table class="min-w-full divide-y divide-gray-200">
				<thead class="bg-gray-50">
					<tr>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Customer</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Contact</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Location</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Total Spent</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Last Service</th>
						<th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>
					</tr>
				</thead>
				<tbody class="bg-white divide-y divide-gray-200">`

	for _, customer := range customers {
		html += fmt.Sprintf(`
					<tr class="hover:bg-gray-50">
						<td class="px-6 py-4 whitespace-nowrap">
							<div>
								<div class="text-sm font-medium text-gray-900">%s %s</div>
								<div class="text-sm text-gray-500">Customer #%s</div>
							</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm text-gray-900">%s</div>
							<div class="text-sm text-gray-500">%s</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap">
							<div class="text-sm text-gray-900">%s</div>
							<div class="text-sm text-gray-500">%s, %s %s</div>
						</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">$%.0f</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">%s</td>
						<td class="px-6 py-4 whitespace-nowrap text-sm font-medium">
							<button hx-get="/admin/partials/customer-form?id=%s" 
							        hx-target="#modal-content"
							        class="text-primary hover:text-secondary mr-3">Edit</button>
							<button hx-delete="/admin/api/customers/%s" 
							        hx-confirm="Are you sure you want to delete this customer?"
							        hx-target="closest tr"
							        hx-swap="outerHTML"
							        class="text-red-600 hover:text-red-900">Delete</button>
						</td>
					</tr>`, 
			customer.FirstName, customer.LastName, customer.ID,
			customer.Email, customer.Phone,
			customer.Address, customer.City, customer.State, customer.ZipCode,
			customer.TotalSpent,
			customer.LastService.Format("Jan 2, 2006"),
			customer.ID, customer.ID)
	}

	html += `
				</tbody>
			</table>
		</div>
	</div>
	
	<!-- Modal -->
	<div id="modal-content"></div>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getJobsPartial(w http.ResponseWriter, r *http.Request) {
	html := `
	<div class="flex justify-between items-center mb-6">
		<h2 class="text-2xl font-bold text-gray-900">Job Management</h2>
		<button hx-get="/admin/partials/job-form" 
		        hx-target="#modal-content"
		        class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-secondary transition">
			Schedule Job
		</button>
	</div>
	
	<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-6">
		<div class="bg-yellow-50 rounded-lg p-4">
			<h3 class="text-lg font-medium text-yellow-800 mb-2">Scheduled</h3>
			<div class="space-y-3">`

	// Scheduled jobs
	for _, job := range jobs {
		if job.Status == "scheduled" {
			customer := customers[job.CustomerID]
			html += fmt.Sprintf(`
				<div class="bg-white p-3 rounded border-l-4 border-yellow-400">
					<div class="flex justify-between items-start">
						<div>
							<p class="font-medium">%s %s</p>
							<p class="text-sm text-gray-600">%s</p>
							<p class="text-xs text-gray-500">%s</p>
						</div>
						<span class="text-sm font-medium text-green-600">$%.0f</span>
					</div>
				</div>`, customer.FirstName, customer.LastName, job.Notes, job.ScheduledAt.Format("Jan 2, 3:04 PM"), job.Price)
		}
	}

	html += `
			</div>
		</div>
		
		<div class="bg-blue-50 rounded-lg p-4">
			<h3 class="text-lg font-medium text-blue-800 mb-2">In Progress</h3>
			<div class="space-y-3">`

	// In-progress jobs
	for _, job := range jobs {
		if job.Status == "in-progress" {
			customer := customers[job.CustomerID]
			html += fmt.Sprintf(`
				<div class="bg-white p-3 rounded border-l-4 border-blue-400">
					<div class="flex justify-between items-start">
						<div>
							<p class="font-medium">%s %s</p>
							<p class="text-sm text-gray-600">%s</p>
							<p class="text-xs text-gray-500">Started: %s</p>
						</div>
						<span class="text-sm font-medium text-green-600">$%.0f</span>
					</div>
					<button hx-put="/admin/api/jobs/%s" 
					        hx-vals='{"status": "completed"}'
					        hx-target="closest div"
					        class="mt-2 text-xs bg-green-600 text-white px-2 py-1 rounded hover:bg-green-700">
						Mark Complete
					</button>
				</div>`, customer.FirstName, customer.LastName, job.Notes, job.ScheduledAt.Format("Jan 2"), job.Price, job.ID)
		}
	}

	html += `
			</div>
		</div>
		
		<div class="bg-green-50 rounded-lg p-4">
			<h3 class="text-lg font-medium text-green-800 mb-2">Completed</h3>
			<div class="space-y-3">`

	// Completed jobs
	for _, job := range jobs {
		if job.Status == "completed" {
			customer := customers[job.CustomerID]
			completedDate := ""
			if job.CompletedAt != nil {
				completedDate = job.CompletedAt.Format("Jan 2")
			}
			html += fmt.Sprintf(`
				<div class="bg-white p-3 rounded border-l-4 border-green-400">
					<div class="flex justify-between items-start">
						<div>
							<p class="font-medium">%s %s</p>
							<p class="text-sm text-gray-600">%s</p>
							<p class="text-xs text-gray-500">Completed: %s</p>
						</div>
						<span class="text-sm font-medium text-green-600">$%.0f</span>
					</div>
				</div>`, customer.FirstName, customer.LastName, job.Notes, completedDate, job.Price)
		}
	}

	html += `
			</div>
		</div>
	</div>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getServicesPartial(w http.ResponseWriter, r *http.Request) {
	html := `
	<div class="flex justify-between items-center mb-6">
		<h2 class="text-2xl font-bold text-gray-900">Service Management</h2>
		<button hx-get="/admin/partials/service-form" 
		        hx-target="#modal-content"
		        class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-secondary transition">
			Add Service
		</button>
	</div>
	
	<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">`

	// Get services from the existing service data
	services := []Service{
		{ID: "1", Name: "Lawn Care", Description: "Professional mowing, edging, and maintenance", Icon: "🌱", Price: "From $50/visit"},
		{ID: "2", Name: "Garden Design", Description: "Custom landscape design and installation", Icon: "🌺", Price: "Free consultation"},
		{ID: "3", Name: "Tree Service", Description: "Trimming, removal, and health assessment", Icon: "🌳", Price: "From $200"},
		{ID: "4", Name: "Irrigation", Description: "Sprinkler system design and repair", Icon: "💧", Price: "From $150"},
		{ID: "5", Name: "Hardscaping", Description: "Stone patios, walkways, and outdoor living", Icon: "🏗️", Price: "From $1,000"},
		{ID: "6", Name: "Snow Removal", Description: "Winter snow removal services", Icon: "❄️", Price: "From $75/visit"},
	}

	for _, service := range services {
		html += fmt.Sprintf(`
		<div class="bg-white rounded-lg shadow p-6">
			<div class="flex items-center justify-between mb-4">
				<div class="text-4xl">%s</div>
				<div class="flex space-x-2">
					<button hx-get="/admin/partials/service-form?id=%s" 
					        hx-target="#modal-content"
					        class="text-primary hover:text-secondary">
						<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"></path>
						</svg>
					</button>
				</div>
			</div>
			<h3 class="text-lg font-semibold mb-2">%s</h3>
			<p class="text-gray-600 mb-3">%s</p>
			<p class="text-primary font-semibold">%s</p>
		</div>`, service.Icon, service.ID, service.Name, service.Description, service.Price)
	}

	html += `
	</div>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getTeamPartial(w http.ResponseWriter, r *http.Request) {
	html := `
	<div class="flex justify-between items-center mb-6">
		<h2 class="text-2xl font-bold text-gray-900">Team Management</h2>
		<button hx-get="/admin/partials/employee-form" 
		        hx-target="#modal-content"
		        class="bg-primary text-white px-4 py-2 rounded-lg hover:bg-secondary transition">
			Add Employee
		</button>
	</div>
	
	<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
		<div class="bg-white rounded-lg shadow">
			<div class="p-6 border-b">
				<h3 class="text-lg font-medium text-gray-900">Teams</h3>
			</div>
			<div class="p-6 space-y-4">`

	for _, team := range teams {
		memberCount := len(team.Members)
		html += fmt.Sprintf(`
				<div class="border rounded-lg p-4">
					<div class="flex justify-between items-start mb-2">
						<h4 class="font-medium">%s</h4>
						<span class="px-2 py-1 text-xs rounded-full bg-green-100 text-green-800">Active</span>
					</div>
					<p class="text-sm text-gray-600 mb-2">%d members</p>
					<div class="flex flex-wrap gap-1">`, team.Name, memberCount)

		for _, specialty := range team.Specialties {
			html += fmt.Sprintf(`
						<span class="px-2 py-1 text-xs bg-gray-100 text-gray-700 rounded">%s</span>`, specialty)
		}

		html += `
					</div>
				</div>`
	}

	html += `
			</div>
		</div>
		
		<div class="bg-white rounded-lg shadow">
			<div class="p-6 border-b">
				<h3 class="text-lg font-medium text-gray-900">Employees</h3>
			</div>
			<div class="p-6 space-y-4">`

	for _, employee := range employees {
		team := teams[employee.TeamID]
		html += fmt.Sprintf(`
				<div class="flex items-center justify-between border-b pb-4">
					<div class="flex items-center">
						<div class="w-10 h-10 bg-primary text-white rounded-full flex items-center justify-center text-sm font-medium">
							%s%s
						</div>
						<div class="ml-3">
							<p class="text-sm font-medium text-gray-900">%s %s</p>
							<p class="text-sm text-gray-500">%s • %s</p>
						</div>
					</div>
					<div class="flex space-x-2">
						<button hx-get="/admin/partials/employee-form?id=%s" 
						        hx-target="#modal-content"
						        class="text-primary hover:text-secondary">Edit</button>
						<button hx-delete="/admin/api/employees/%s" 
						        hx-confirm="Are you sure?"
						        hx-target="closest div"
						        hx-swap="outerHTML"
						        class="text-red-600 hover:text-red-900">Remove</button>
					</div>
				</div>`, 
			string(employee.FirstName[0]), string(employee.LastName[0]),
			employee.FirstName, employee.LastName, 
			employee.Role, team.Name,
			employee.ID, employee.ID)
	}

	html += `
			</div>
		</div>
	</div>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getReportsPartial(w http.ResponseWriter, r *http.Request) {
	// Calculate analytics
	totalRevenue := 0.0
	completedJobs := 0
	avgJobValue := 0.0
	
	for _, job := range jobs {
		if job.Status == "completed" {
			totalRevenue += job.Price
			completedJobs++
		}
	}
	
	if completedJobs > 0 {
		avgJobValue = totalRevenue / float64(completedJobs)
	}

	html := fmt.Sprintf(`
	<div class="mb-6">
		<h2 class="text-2xl font-bold text-gray-900 mb-4">Reports & Analytics</h2>
		
		<div class="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
			<div class="bg-white rounded-lg shadow p-6">
				<h3 class="text-lg font-medium text-gray-900 mb-2">Total Revenue</h3>
				<p class="text-3xl font-bold text-green-600">$%.0f</p>
				<p class="text-sm text-gray-500">From completed jobs</p>
			</div>
			
			<div class="bg-white rounded-lg shadow p-6">
				<h3 class="text-lg font-medium text-gray-900 mb-2">Completed Jobs</h3>
				<p class="text-3xl font-bold text-blue-600">%d</p>
				<p class="text-sm text-gray-500">This period</p>
			</div>
			
			<div class="bg-white rounded-lg shadow p-6">
				<h3 class="text-lg font-medium text-gray-900 mb-2">Average Job Value</h3>
				<p class="text-3xl font-bold text-purple-600">$%.0f</p>
				<p class="text-sm text-gray-500">Per completed job</p>
			</div>
		</div>
		
		<div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
			<div class="bg-white rounded-lg shadow p-6">
				<h3 class="text-lg font-medium text-gray-900 mb-4">Customer Distribution</h3>
				<div class="space-y-3">`, totalRevenue, completedJobs, avgJobValue)

	// Customer analytics
	for _, customer := range customers {
		percentage := (customer.TotalSpent / 4200.0) * 100 // Total of all customer spending
		html += fmt.Sprintf(`
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium">%s %s</span>
						<div class="flex items-center">
							<div class="w-32 bg-gray-200 rounded-full h-2 mr-3">
								<div class="bg-primary h-2 rounded-full" style="width: %.1f%%"></div>
							</div>
							<span class="text-sm text-gray-600">$%.0f</span>
						</div>
					</div>`, customer.FirstName, customer.LastName, percentage, customer.TotalSpent)
	}

	html += `
				</div>
			</div>
			
			<div class="bg-white rounded-lg shadow p-6">
				<h3 class="text-lg font-medium text-gray-900 mb-4">Job Status Distribution</h3>
				<div class="space-y-3">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium">Completed</span>
						<div class="flex items-center">
							<div class="w-32 bg-gray-200 rounded-full h-2 mr-3">
								<div class="bg-green-500 h-2 rounded-full" style="width: 33%"></div>
							</div>
							<span class="text-sm text-gray-600">1</span>
						</div>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium">In Progress</span>
						<div class="flex items-center">
							<div class="w-32 bg-gray-200 rounded-full h-2 mr-3">
								<div class="bg-blue-500 h-2 rounded-full" style="width: 33%"></div>
							</div>
							<span class="text-sm text-gray-600">1</span>
						</div>
					</div>
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium">Scheduled</span>
						<div class="flex items-center">
							<div class="w-32 bg-gray-200 rounded-full h-2 mr-3">
								<div class="bg-yellow-500 h-2 rounded-full" style="width: 33%"></div>
							</div>
							<span class="text-sm text-gray-600">1</span>
						</div>
					</div>
				</div>
			</div>
		</div>
		
		<div class="mt-6 text-center">
			<button class="bg-primary text-white px-6 py-2 rounded-lg hover:bg-secondary transition">
				Export Reports
			</button>
		</div>
	</div>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

// CRUD Operations
func createCustomer(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("cust_%d", len(customers)+1)
	customer := Customer{
		ID:        id,
		FirstName: r.FormValue("first_name"),
		LastName:  r.FormValue("last_name"),
		Email:     r.FormValue("email"),
		Phone:     r.FormValue("phone"),
		Address:   r.FormValue("address"),
		City:      r.FormValue("city"),
		State:     r.FormValue("state"),
		ZipCode:   r.FormValue("zip_code"),
		CreatedAt: time.Now(),
	}

	customers[id] = customer

	// Return success response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">Customer added successfully!</div>`))
}

func updateCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	customer, exists := customers[id]
	if !exists {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Update customer fields
	customer.FirstName = r.FormValue("first_name")
	customer.LastName = r.FormValue("last_name")
	customer.Email = r.FormValue("email")
	customer.Phone = r.FormValue("phone")
	customer.Address = r.FormValue("address")
	customer.City = r.FormValue("city")
	customer.State = r.FormValue("state")
	customer.ZipCode = r.FormValue("zip_code")

	customers[id] = customer

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">Customer updated successfully!</div>`))
}

func deleteCustomer(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	delete(customers, id)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(""))
}

func createJob(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	scheduledAt, _ := time.Parse("2006-01-02T15:04", r.FormValue("scheduled_at"))
	price, _ := strconv.ParseFloat(r.FormValue("price"), 64)
	duration, _ := strconv.Atoi(r.FormValue("duration"))
	serviceRequestID := r.FormValue("service_request_id")

	id := fmt.Sprintf("job_%d", len(jobs)+1)
	job := Job{
		ID:               id,
		ServiceRequestID: serviceRequestID,
		CustomerID:       r.FormValue("customer_id"),
		ServiceID:        r.FormValue("service_id"),
		TeamID:           r.FormValue("team_id"),
		Status:           "scheduled",
		ScheduledAt:      scheduledAt,
		Notes:            r.FormValue("notes"),
		Price:            price,
		Duration:         duration,
	}

	jobs[id] = job

	// If this job was created from a service request, update the request status
	if serviceRequestID != "" {
		if request, exists := serviceRequests[serviceRequestID]; exists {
			request.Status = "scheduled"
			serviceRequests[serviceRequestID] = request
			log.Printf("Service request %s status updated to 'scheduled' - Job %s created", serviceRequestID, id)
		}
	}

	log.Printf("New job created: %s for customer %s", id, r.FormValue("customer_id"))

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">Job scheduled successfully!</div>`))
}

func updateJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	job, exists := jobs[id]
	if !exists {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	// Parse JSON body for status updates
	var updateData map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updateData)
	if err == nil {
		if status, ok := updateData["status"].(string); ok {
			oldStatus := job.Status
			job.Status = status
			if status == "completed" && oldStatus != "completed" {
				now := time.Now()
				job.CompletedAt = &now
				
				// Update customer's total spent and service count when job is completed
				if customer, exists := customers[job.CustomerID]; exists {
					customer.TotalSpent += job.Price
					customer.ServiceCount++
					customer.LastService = now
					customers[job.CustomerID] = customer
					log.Printf("Customer %s updated: Total spent $%.2f, Service count %d", customer.ID, customer.TotalSpent, customer.ServiceCount)
				}
				
				log.Printf("Job %s completed successfully. Revenue: $%.2f", id, job.Price)
			}
		}
	}

	jobs[id] = job

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">Job updated successfully!</div>`))
}

func deleteJob(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	delete(jobs, id)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(""))
}

func createEmployee(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	id := fmt.Sprintf("emp_%d", len(employees)+1)
	employee := Employee{
		ID:        id,
		FirstName: r.FormValue("first_name"),
		LastName:  r.FormValue("last_name"),
		Email:     r.FormValue("email"),
		Phone:     r.FormValue("phone"),
		Role:      r.FormValue("role"),
		HireDate:  time.Now(),
		Active:    true,
		TeamID:    r.FormValue("team_id"),
	}

	employees[id] = employee

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">Employee added successfully!</div>`))
}

func updateEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	employee, exists := employees[id]
	if !exists {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	employee.FirstName = r.FormValue("first_name")
	employee.LastName = r.FormValue("last_name")
	employee.Email = r.FormValue("email")
	employee.Phone = r.FormValue("phone")
	employee.Role = r.FormValue("role")
	employee.TeamID = r.FormValue("team_id")

	employees[id] = employee

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(`<div class="bg-green-100 border border-green-400 text-green-700 px-4 py-3 rounded">Employee updated successfully!</div>`))
}

func deleteEmployee(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	
	delete(employees, id)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(""))
}

// Form handlers for modals
func getCustomerForm(w http.ResponseWriter, r *http.Request) {
	customerID := r.URL.Query().Get("id")
	var customer Customer
	isEdit := false
	
	if customerID != "" {
		if c, exists := customers[customerID]; exists {
			customer = c
			isEdit = true
		}
	}
	
	formTitle := "Add Customer"
	if isEdit {
		formTitle = "Edit Customer"
	}
	
	html := fmt.Sprintf(`
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-white rounded-lg max-w-2xl w-full p-6 max-h-screen overflow-y-auto">
			<div class="flex justify-between items-center mb-6">
				<h2 class="text-2xl font-bold text-gray-900">%s</h2>
				<button onclick="document.getElementById('modal-content').innerHTML = ''" 
				        class="text-gray-500 hover:text-gray-700">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>
			
			<form hx-post="/admin/api/customers" hx-target="#modal-content" class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">First Name</label>
						<input type="text" name="first_name" value="%s" required 
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Last Name</label>
						<input type="text" name="last_name" value="%s" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
				</div>
				
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
						<input type="email" name="email" value="%s" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Phone</label>
						<input type="tel" name="phone" value="%s" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
				</div>
				
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Address</label>
					<input type="text" name="address" value="%s" required
					       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
				</div>
				
				<div class="grid grid-cols-3 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">City</label>
						<input type="text" name="city" value="%s" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">State</label>
						<input type="text" name="state" value="%s" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Zip Code</label>
						<input type="text" name="zip_code" value="%s" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
				</div>
				
				<div class="flex justify-end space-x-3 pt-4">
					<button type="button" 
					        onclick="document.getElementById('modal-content').innerHTML = ''"
					        class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300 transition">
						Cancel
					</button>
					<button type="submit" 
					        class="px-4 py-2 bg-primary text-white rounded-md hover:bg-secondary transition">
						%s
					</button>
				</div>
			</form>
		</div>
	</div>`, formTitle, customer.FirstName, customer.LastName, customer.Email, customer.Phone, customer.Address, customer.City, customer.State, customer.ZipCode, formTitle)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getJobForm(w http.ResponseWriter, r *http.Request) {
	requestID := r.URL.Query().Get("request_id")
	var serviceRequest ServiceRequest
	var preSelectedCustomerID string
	var preSelectedServiceType string
	
	// If this job is being created from a service request, pre-populate data
	if requestID != "" {
		if req, exists := serviceRequests[requestID]; exists {
			serviceRequest = req
			preSelectedCustomerID = req.CustomerID
			preSelectedServiceType = req.ServiceType
		}
	}
	
	html := fmt.Sprintf(`
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-white rounded-lg max-w-2xl w-full p-6 max-h-screen overflow-y-auto">
			<div class="flex justify-between items-center mb-6">
				<h2 class="text-2xl font-bold text-gray-900">%s</h2>
				<button onclick="document.getElementById('modal-content').innerHTML = ''" 
				        class="text-gray-500 hover:text-gray-700">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>
			
			<form hx-post="/admin/api/jobs" hx-target="#modal-content" class="space-y-4">
				<input type="hidden" name="service_request_id" value="%s">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Customer</label>
						<select name="customer_id" required class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
							<option value="">Select Customer</option>`,
		func() string {
			if requestID != "" {
				return fmt.Sprintf("Schedule Job for %s", serviceRequest.CustomerName)
			}
			return "Schedule Job"
		}(),
		requestID)

	for _, customer := range customers {
		selected := ""
		if customer.ID == preSelectedCustomerID {
			selected = "selected"
		}
		html += fmt.Sprintf(`<option value="%s" %s>%s %s</option>`, customer.ID, selected, customer.FirstName, customer.LastName)
	}

	html += `
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Service</label>
						<select name="service_id" required class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
							<option value="">Select Service</option>`
							
	// Add service options with pre-selection if from service request
	services := []struct{ID, Name string}{
		{"lawn_care", "Lawn Care"},
		{"garden_design", "Garden Design"},
		{"tree_service", "Tree Service"},
		{"irrigation", "Irrigation"},
		{"hardscaping", "Hardscaping"},
		{"snow_removal", "Snow Removal"},
	}
	
	for _, service := range services {
		selected := ""
		if service.ID == preSelectedServiceType {
			selected = "selected"
		}
		html += fmt.Sprintf(`<option value="%s" %s>%s</option>`, service.ID, selected, service.Name)
	}
	
	html += `
						</select>
					</div>
				</div>`
				
	// Show service request details if applicable
	if requestID != "" {
		html += fmt.Sprintf(`
				<div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
					<h4 class="text-sm font-medium text-blue-800 mb-2">Service Request Details:</h4>
					<div class="text-sm text-blue-700">
						<p><strong>Customer:</strong> %s</p>
						<p><strong>Requested Service:</strong> %s</p>
						<p><strong>Message:</strong> %s</p>
						<p><strong>Submitted:</strong> %s</p>
					</div>
				</div>`,
			serviceRequest.CustomerName,
			strings.Title(strings.Replace(serviceRequest.ServiceType, "_", " ", -1)),
			serviceRequest.Message,
			serviceRequest.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"))
	}
	
	html += `
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Team</label>
						<select name="team_id" required class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
							<option value="">Select Team</option>`

	for _, team := range teams {
		html += fmt.Sprintf(`<option value="%s">%s</option>`, team.ID, team.Name)
	}

	html += `
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Price ($)</label>
						<input type="number" name="price" step="0.01" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
				</div>
				
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Scheduled Date & Time</label>
						<input type="datetime-local" name="scheduled_at" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Duration (minutes)</label>
						<input type="number" name="duration" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
				</div>
				
				<div>
					<label class="block text-sm font-medium text-gray-700 mb-1">Notes</label>
					<textarea name="notes" rows="3"
					          class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary"></textarea>
				</div>
				
				<div class="flex justify-end space-x-3 pt-4">
					<button type="button" 
					        onclick="document.getElementById('modal-content').innerHTML = ''"
					        class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300 transition">
						Cancel
					</button>
					<button type="submit" 
					        class="px-4 py-2 bg-primary text-white rounded-md hover:bg-secondary transition">
						Schedule Job
					</button>
				</div>
			</form>
		</div>
	</div>`
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getEmployeeForm(w http.ResponseWriter, r *http.Request) {
	employeeID := r.URL.Query().Get("id")
	var employee Employee
	isEdit := false
	
	if employeeID != "" {
		if e, exists := employees[employeeID]; exists {
			employee = e
			isEdit = true
		}
	}
	
	formTitle := "Add Employee"
	if isEdit {
		formTitle = "Edit Employee"
	}
	
	html := fmt.Sprintf(`
	<div class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
		<div class="bg-white rounded-lg max-w-2xl w-full p-6 max-h-screen overflow-y-auto">
			<div class="flex justify-between items-center mb-6">
				<h2 class="text-2xl font-bold text-gray-900">%s</h2>
				<button onclick="document.getElementById('modal-content').innerHTML = ''" 
				        class="text-gray-500 hover:text-gray-700">
					<svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
						<path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path>
					</svg>
				</button>
			</div>
			
			<form hx-post="/admin/api/employees" hx-target="#modal-content" class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">First Name</label>
						<input type="text" name="first_name" value="%s" required 
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Last Name</label>
						<input type="text" name="last_name" value="%s" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
				</div>
				
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Email</label>
						<input type="email" name="email" value="%s" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Phone</label>
						<input type="tel" name="phone" value="%s" required
						       class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
					</div>
				</div>
				
				<div class="grid grid-cols-2 gap-4">
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Role</label>
						<select name="role" required class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
							<option value="">Select Role</option>
							<option value="Crew Leader" %s>Crew Leader</option>
							<option value="Landscaper" %s>Landscaper</option>
							<option value="Designer" %s>Designer</option>
							<option value="Hardscape Specialist" %s>Hardscape Specialist</option>
							<option value="Foreman" %s>Foreman</option>
						</select>
					</div>
					<div>
						<label class="block text-sm font-medium text-gray-700 mb-1">Team</label>
						<select name="team_id" required class="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary">
							<option value="">Select Team</option>`, 
		formTitle, employee.FirstName, employee.LastName, employee.Email, employee.Phone,
		getSelected(employee.Role, "Crew Leader"),
		getSelected(employee.Role, "Landscaper"),
		getSelected(employee.Role, "Designer"),
		getSelected(employee.Role, "Hardscape Specialist"),
		getSelected(employee.Role, "Foreman"))

	for _, team := range teams {
		selected := ""
		if team.ID == employee.TeamID {
			selected = "selected"
		}
		html += fmt.Sprintf(`<option value="%s" %s>%s</option>`, team.ID, selected, team.Name)
	}

	html += fmt.Sprintf(`
						</select>
					</div>
				</div>
				
				<div class="flex justify-end space-x-3 pt-4">
					<button type="button" 
					        onclick="document.getElementById('modal-content').innerHTML = ''"
					        class="px-4 py-2 text-gray-700 bg-gray-200 rounded-md hover:bg-gray-300 transition">
						Cancel
					</button>
					<button type="submit" 
					        class="px-4 py-2 bg-primary text-white rounded-md hover:bg-secondary transition">
						%s
					</button>
				</div>
			</form>
		</div>
	</div>`, formTitle)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func getSelected(current, option string) string {
	if current == option {
		return "selected"
	}
	return ""
}

// Service Request Management
func acceptServiceRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]
	
	request, exists := serviceRequests[requestID]
	if !exists {
		http.Error(w, "Service request not found", http.StatusNotFound)
		return
	}
	
	// Update status to accepted
	request.Status = "accepted"
	serviceRequests[requestID] = request
	
	log.Printf("Service request %s accepted for customer %s", requestID, request.CustomerName)
	
	// Return updated request card
	statusColor := "green"
	serviceTypeDisplay := strings.Replace(request.ServiceType, "_", " ", -1)
	serviceTypeDisplay = strings.Title(serviceTypeDisplay)
	
	html := fmt.Sprintf(`
		<div class="request-card p-6">
			<div class="flex items-start justify-between">
				<div class="flex-1">
					<div class="flex items-center justify-between mb-2">
						<h3 class="text-lg font-medium text-gray-900">%s</h3>
						<span class="px-2 py-1 text-xs rounded-full bg-%s-100 text-%s-800">accepted</span>
					</div>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm text-gray-600 mb-2">
						<div>
							<p><strong>Email:</strong> %s</p>
							<p><strong>Phone:</strong> %s</p>
							<p><strong>Service:</strong> %s</p>
						</div>
						<div>
							<p><strong>Submitted:</strong> %s</p>
							<p><strong>Request ID:</strong> %s</p>
						</div>
					</div>
					<div class="mb-2">
						<p class="text-sm text-gray-600"><strong>Message:</strong></p>
						<p class="text-sm text-gray-800 bg-gray-50 p-2 rounded">%s</p>
					</div>
					<div class="mt-2">
						<button hx-get="/admin/partials/job-form?request_id=%s" 
						        hx-target="#modal-content"
						        class="bg-blue-600 text-white px-3 py-1 rounded text-sm hover:bg-blue-700">
							Schedule Job
						</button>
					</div>
				</div>
			</div>
		</div>`, 
		request.CustomerName, statusColor, statusColor,
		request.CustomerEmail, request.CustomerPhone, serviceTypeDisplay,
		request.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"), request.ID,
		request.Message, request.ID)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}

func denyServiceRequest(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	requestID := vars["id"]
	
	request, exists := serviceRequests[requestID]
	if !exists {
		http.Error(w, "Service request not found", http.StatusNotFound)
		return
	}
	
	// Update status to denied
	request.Status = "denied"
	serviceRequests[requestID] = request
	
	log.Printf("Service request %s denied for customer %s", requestID, request.CustomerName)
	
	// Return updated request card
	statusColor := "red"
	serviceTypeDisplay := strings.Replace(request.ServiceType, "_", " ", -1)
	serviceTypeDisplay = strings.Title(serviceTypeDisplay)
	
	html := fmt.Sprintf(`
		<div class="request-card p-6">
			<div class="flex items-start justify-between">
				<div class="flex-1">
					<div class="flex items-center justify-between mb-2">
						<h3 class="text-lg font-medium text-gray-900">%s</h3>
						<span class="px-2 py-1 text-xs rounded-full bg-%s-100 text-%s-800">denied</span>
					</div>
					<div class="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm text-gray-600 mb-2">
						<div>
							<p><strong>Email:</strong> %s</p>
							<p><strong>Phone:</strong> %s</p>
							<p><strong>Service:</strong> %s</p>
						</div>
						<div>
							<p><strong>Submitted:</strong> %s</p>
							<p><strong>Request ID:</strong> %s</p>
						</div>
					</div>
					<div class="mb-2">
						<p class="text-sm text-gray-600"><strong>Message:</strong></p>
						<p class="text-sm text-gray-800 bg-gray-50 p-2 rounded">%s</p>
					</div>
					<div class="mt-2">
						<span class="text-sm text-gray-500 italic">Request has been denied.</span>
					</div>
				</div>
			</div>
		</div>`, 
		request.CustomerName, statusColor, statusColor,
		request.CustomerEmail, request.CustomerPhone, serviceTypeDisplay,
		request.CreatedAt.Format("Jan 2, 2006 at 3:04 PM"), request.ID,
		request.Message)
	
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}