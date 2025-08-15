package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/pageza/landscaping-app/web/internal/services"
)

// Customer represents a customer for the frontend
type Customer struct {
	ID           string  `json:"id"`
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Email        *string `json:"email"`
	Phone        *string `json:"phone"`
	AddressLine1 *string `json:"address_line1"`
	AddressLine2 *string `json:"address_line2"`
	City         *string `json:"city"`
	State        *string `json:"state"`
	ZipCode      *string `json:"zip_code"`
	Status       string  `json:"status"`
	CreatedAt    string  `json:"created_at"`
}

// listCustomers displays the customer list page
func (h *Handlers) listCustomers(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := services.TemplateData{
		Title:   "Customers",
		User:    user,
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
	}

	content, err := h.services.Template.Render("customers_list.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// customersTable returns customer data as HTML table for HTMX
func (h *Handlers) customersTable(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit < 1 || limit > 100 {
		limit = 20
	}
	search := r.URL.Query().Get("search")
	sortBy := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")

	// Get customers from API
	customers, totalCount, err := h.getCustomers(user, page, limit, search, sortBy, sortOrder)
	if err != nil {
		http.Error(w, "Failed to load customers", http.StatusInternalServerError)
		return
	}

	data := services.TemplateData{
		User:   user,
		IsHTMX: true,
		Data: map[string]interface{}{
			"customers":   customers,
			"totalCount":  totalCount,
			"currentPage": page,
			"limit":       limit,
			"search":      search,
			"sortBy":      sortBy,
			"sortOrder":   sortOrder,
		},
	}

	content, err := h.services.Template.Render("customers_table.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// showCreateCustomer displays the create customer form
func (h *Handlers) showCreateCustomer(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := services.TemplateData{
		Title:   "Add New Customer",
		User:    user,
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
		Flash:   make(map[string]string),
	}

	content, err := h.services.Template.Render("customer_form.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// createCustomer handles customer creation
func (h *Handlers) createCustomer(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.renderCustomerFormError(w, r, nil, "Invalid form data")
		return
	}

	customer := &Customer{
		FirstName:    strings.TrimSpace(r.FormValue("first_name")),
		LastName:     strings.TrimSpace(r.FormValue("last_name")),
		Status:       "active",
	}

	// Optional fields
	if email := strings.TrimSpace(r.FormValue("email")); email != "" {
		customer.Email = &email
	}
	if phone := strings.TrimSpace(r.FormValue("phone")); phone != "" {
		customer.Phone = &phone
	}
	if addr1 := strings.TrimSpace(r.FormValue("address_line1")); addr1 != "" {
		customer.AddressLine1 = &addr1
	}
	if addr2 := strings.TrimSpace(r.FormValue("address_line2")); addr2 != "" {
		customer.AddressLine2 = &addr2
	}
	if city := strings.TrimSpace(r.FormValue("city")); city != "" {
		customer.City = &city
	}
	if state := strings.TrimSpace(r.FormValue("state")); state != "" {
		customer.State = &state
	}
	if zipCode := strings.TrimSpace(r.FormValue("zip_code")); zipCode != "" {
		customer.ZipCode = &zipCode
	}

	// Validate required fields
	if customer.FirstName == "" || customer.LastName == "" {
		h.renderCustomerFormError(w, r, customer, "First name and last name are required")
		return
	}

	// Create customer via API
	createdCustomer, err := h.createCustomerAPI(user, customer)
	if err != nil {
		h.renderCustomerFormError(w, r, customer, "Failed to create customer: "+err.Error())
		return
	}

	// Redirect to customer detail page
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", fmt.Sprintf("/admin/customers/%s", createdCustomer.ID))
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/customers/%s", createdCustomer.ID), http.StatusSeeOther)
}

// showCustomer displays customer details
func (h *Handlers) showCustomer(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	customerID := vars["id"]

	customer, err := h.getCustomerByID(user, customerID)
	if err != nil {
		http.Error(w, "Customer not found", http.StatusNotFound)
		return
	}

	data := services.TemplateData{
		Title:   fmt.Sprintf("%s %s", customer.FirstName, customer.LastName),
		User:    user,
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
		Data:    customer,
	}

	content, err := h.services.Template.Render("customer_detail.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// updateCustomer handles customer updates
func (h *Handlers) updateCustomer(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	customerID := vars["id"]

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	customer := &Customer{
		ID:        customerID,
		FirstName: strings.TrimSpace(r.FormValue("first_name")),
		LastName:  strings.TrimSpace(r.FormValue("last_name")),
		Status:    r.FormValue("status"),
	}

	// Optional fields
	if email := strings.TrimSpace(r.FormValue("email")); email != "" {
		customer.Email = &email
	}
	if phone := strings.TrimSpace(r.FormValue("phone")); phone != "" {
		customer.Phone = &phone
	}
	if addr1 := strings.TrimSpace(r.FormValue("address_line1")); addr1 != "" {
		customer.AddressLine1 = &addr1
	}
	if addr2 := strings.TrimSpace(r.FormValue("address_line2")); addr2 != "" {
		customer.AddressLine2 = &addr2
	}
	if city := strings.TrimSpace(r.FormValue("city")); city != "" {
		customer.City = &city
	}
	if state := strings.TrimSpace(r.FormValue("state")); state != "" {
		customer.State = &state
	}
	if zipCode := strings.TrimSpace(r.FormValue("zip_code")); zipCode != "" {
		customer.ZipCode = &zipCode
	}

	// Update customer via API
	if err := h.updateCustomerAPI(user, customer); err != nil {
		http.Error(w, "Failed to update customer", http.StatusInternalServerError)
		return
	}

	// Return success response for HTMX
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Trigger", "customer-updated")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Customer updated successfully"))
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/admin/customers/%s", customerID), http.StatusSeeOther)
}

// deleteCustomer handles customer deletion
func (h *Handlers) deleteCustomer(w http.ResponseWriter, r *http.Request) {
	user, err := h.getCurrentUser(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	vars := mux.Vars(r)
	customerID := vars["id"]

	if err := h.deleteCustomerAPI(user, customerID); err != nil {
		http.Error(w, "Failed to delete customer", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/admin/customers")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/admin/customers", http.StatusSeeOther)
}

// Helper functions

func (h *Handlers) renderCustomerFormError(w http.ResponseWriter, r *http.Request, customer *Customer, errorMsg string) {
	data := services.TemplateData{
		Title:  "Add New Customer",
		IsHTMX: r.Header.Get("HX-Request") == "true",
		Flash:  map[string]string{"error": errorMsg},
		Data:   customer,
	}

	content, err := h.services.Template.Render("customer_form.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(content))
}

func (h *Handlers) getCustomers(user *services.User, page, limit int, search, sortBy, sortOrder string) ([]Customer, int, error) {
	token := h.extractToken(&http.Request{})
	
	// Build query parameters
	params := fmt.Sprintf("?page=%d&limit=%d", page, limit)
	if search != "" {
		params += "&search=" + search
	}
	if sortBy != "" {
		params += "&sort=" + sortBy
	}
	if sortOrder != "" {
		params += "&order=" + sortOrder
	}

	resp, err := h.services.API.AuthenticatedGet("/api/v1/customers"+params, token)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, 0, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var result struct {
		Data       []Customer `json:"data"`
		TotalCount int        `json:"total_count"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, 0, err
	}

	return result.Data, result.TotalCount, nil
}

func (h *Handlers) getCustomerByID(user *services.User, customerID string) (*Customer, error) {
	token := h.extractToken(&http.Request{})
	
	resp, err := h.services.API.AuthenticatedGet("/api/v1/customers/"+customerID, token)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("customer not found")
	}

	var customer Customer
	if err := json.NewDecoder(resp.Body).Decode(&customer); err != nil {
		return nil, err
	}

	return &customer, nil
}

func (h *Handlers) createCustomerAPI(user *services.User, customer *Customer) (*Customer, error) {
	token := h.extractToken(&http.Request{})
	
	jsonData, err := json.Marshal(customer)
	if err != nil {
		return nil, err
	}

	resp, err := h.services.API.AuthenticatedPost("/api/v1/customers", token, strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to create customer")
	}

	var created Customer
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, err
	}

	return &created, nil
}

func (h *Handlers) updateCustomerAPI(user *services.User, customer *Customer) error {
	token := h.extractToken(&http.Request{})
	
	jsonData, err := json.Marshal(customer)
	if err != nil {
		return err
	}

	resp, err := h.services.API.AuthenticatedPut("/api/v1/customers/"+customer.ID, token, strings.NewReader(string(jsonData)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to update customer")
	}

	return nil
}

func (h *Handlers) deleteCustomerAPI(user *services.User, customerID string) error {
	token := h.extractToken(&http.Request{})
	
	resp, err := h.services.API.AuthenticatedDelete("/api/v1/customers/"+customerID, token)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to delete customer")
	}

	return nil
}