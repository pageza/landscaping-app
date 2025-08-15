package handlers

import (
	"net/http"
	"strings"

	"github.com/pageza/landscaping-app/web/internal/services"
)

// showLogin displays the login page
func (h *Handlers) showLogin(w http.ResponseWriter, r *http.Request) {
	data := services.TemplateData{
		Title:   "Login",
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
		Flash:   make(map[string]string),
	}

	// Check for flash messages in query params
	if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
		data.Flash["error"] = errorMsg
	}

	if r.Header.Get("HX-Request") == "true" {
		// Return just the form for HTMX requests
		content, err := h.services.Template.Render("login_form.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(content))
		return
	}

	// Return full page
	content, err := h.services.Template.Render("login.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// handleLogin processes login form submission
func (h *Handlers) handleLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderLoginError(w, r, "Invalid form data")
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if email == "" || password == "" {
		h.renderLoginError(w, r, "Email and password are required")
		return
	}

	// Authenticate with backend API
	loginResp, err := h.services.Auth.Login(email, password)
	if err != nil {
		h.renderLoginError(w, r, "Invalid email or password")
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    loginResp.Token,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.config.EnableTLS,
		SameSite: http.SameSiteLaxMode,
	})

	// Redirect based on user role
	redirectURL := "/dashboard"
	if loginResp.User.Role == "customer" {
		redirectURL = "/portal"
	} else if loginResp.User.Role == "admin" || loginResp.User.Role == "owner" {
		redirectURL = "/admin"
	}

	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", redirectURL)
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *Handlers) renderLoginError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	data := services.TemplateData{
		Title:  "Login",
		IsHTMX: r.Header.Get("HX-Request") == "true",
		Flash:  map[string]string{"error": errorMsg},
	}

	if r.Header.Get("HX-Request") == "true" {
		content, err := h.services.Template.Render("login_form.html", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(content))
		return
	}

	content, err := h.services.Template.Render("login.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(content))
}

// showRegister displays the registration page
func (h *Handlers) showRegister(w http.ResponseWriter, r *http.Request) {
	data := services.TemplateData{
		Title:   "Register",
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
		Flash:   make(map[string]string),
	}

	if errorMsg := r.URL.Query().Get("error"); errorMsg != "" {
		data.Flash["error"] = errorMsg
	}

	content, err := h.services.Template.Render("register.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// handleRegister processes registration form submission
func (h *Handlers) handleRegister(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderRegisterError(w, r, "Invalid form data")
		return
	}

	req := services.RegisterRequest{
		Email:           strings.TrimSpace(r.FormValue("email")),
		Password:        r.FormValue("password"),
		ConfirmPassword: r.FormValue("confirm_password"),
		FirstName:       strings.TrimSpace(r.FormValue("first_name")),
		LastName:        strings.TrimSpace(r.FormValue("last_name")),
		CompanyName:     strings.TrimSpace(r.FormValue("company_name")),
	}

	// Validate form data
	if err := h.validateRegistration(req); err != nil {
		h.renderRegisterError(w, r, err.Error())
		return
	}

	// Register with backend API
	if err := h.services.Auth.Register(req); err != nil {
		h.renderRegisterError(w, r, "Registration failed. Please try again.")
		return
	}

	// Redirect to login with success message
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login?success=Registration successful. Please login.")
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/login?success=Registration successful. Please login.", http.StatusSeeOther)
}

func (h *Handlers) validateRegistration(req services.RegisterRequest) error {
	if req.Email == "" {
		return NewValidationError("Email is required")
	}
	if req.Password == "" {
		return NewValidationError("Password is required")
	}
	if req.Password != req.ConfirmPassword {
		return NewValidationError("Passwords do not match")
	}
	if len(req.Password) < 8 {
		return NewValidationError("Password must be at least 8 characters")
	}
	if req.FirstName == "" {
		return NewValidationError("First name is required")
	}
	if req.LastName == "" {
		return NewValidationError("Last name is required")
	}
	return nil
}

func (h *Handlers) renderRegisterError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	data := services.TemplateData{
		Title:  "Register",
		IsHTMX: r.Header.Get("HX-Request") == "true",
		Flash:  map[string]string{"error": errorMsg},
	}

	content, err := h.services.Template.Render("register.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(content))
}

// showForgotPassword displays the forgot password page
func (h *Handlers) showForgotPassword(w http.ResponseWriter, r *http.Request) {
	data := services.TemplateData{
		Title:   "Forgot Password",
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
		Flash:   make(map[string]string),
	}

	content, err := h.services.Template.Render("forgot_password.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// handleForgotPassword processes forgot password form submission
func (h *Handlers) handleForgotPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	email := strings.TrimSpace(r.FormValue("email"))
	if email == "" {
		h.renderForgotPasswordError(w, r, "Email is required")
		return
	}

	// Send password reset email
	if err := h.services.Auth.ForgotPassword(email); err != nil {
		// Don't reveal whether email exists - always show success
	}

	data := services.TemplateData{
		Title:   "Forgot Password",
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Flash:   map[string]string{"success": "If an account with that email exists, we've sent a password reset link."},
		Request: r,
	}

	content, err := h.services.Template.Render("forgot_password.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

func (h *Handlers) renderForgotPasswordError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	data := services.TemplateData{
		Title:  "Forgot Password",
		IsHTMX: r.Header.Get("HX-Request") == "true",
		Flash:  map[string]string{"error": errorMsg},
	}

	content, err := h.services.Template.Render("forgot_password.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(content))
}

// showResetPassword displays the reset password page
func (h *Handlers) showResetPassword(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Redirect(w, r, "/login?error=Invalid reset token", http.StatusSeeOther)
		return
	}

	data := services.TemplateData{
		Title:   "Reset Password",
		IsHTMX:  r.Header.Get("HX-Request") == "true",
		Request: r,
		Flash:   make(map[string]string),
		Data:    map[string]string{"token": token},
	}

	content, err := h.services.Template.Render("reset_password.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(content))
}

// handleResetPassword processes reset password form submission
func (h *Handlers) handleResetPassword(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	if token == "" {
		http.Redirect(w, r, "/login?error=Invalid reset token", http.StatusSeeOther)
		return
	}

	if password == "" || confirmPassword == "" {
		h.renderResetPasswordError(w, r, token, "Password is required")
		return
	}

	if password != confirmPassword {
		h.renderResetPasswordError(w, r, token, "Passwords do not match")
		return
	}

	if len(password) < 8 {
		h.renderResetPasswordError(w, r, token, "Password must be at least 8 characters")
		return
	}

	// Reset password
	if err := h.services.Auth.ResetPassword(token, password); err != nil {
		h.renderResetPasswordError(w, r, token, "Failed to reset password. The token may be expired.")
		return
	}

	// Redirect to login with success message
	http.Redirect(w, r, "/login?success=Password reset successful. Please login with your new password.", http.StatusSeeOther)
}

func (h *Handlers) renderResetPasswordError(w http.ResponseWriter, r *http.Request, token, errorMsg string) {
	data := services.TemplateData{
		Title:  "Reset Password",
		IsHTMX: r.Header.Get("HX-Request") == "true",
		Flash:  map[string]string{"error": errorMsg},
		Data:   map[string]string{"token": token},
	}

	content, err := h.services.Template.Render("reset_password.html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusBadRequest)
	w.Write([]byte(content))
}

// handleLogout processes logout requests
func (h *Handlers) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.config.EnableTLS,
		MaxAge:   -1,
	})

	// Redirect to login
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

// ValidationError represents a form validation error
type ValidationError struct {
	message string
}

func NewValidationError(message string) error {
	return &ValidationError{message: message}
}

func (e *ValidationError) Error() string {
	return e.message
}