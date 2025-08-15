package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pageza/landscaping-app/web/internal/config"
)

type AuthService struct {
	config *config.Config
	api    *APIService
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	ExpiresAt string `json:"expires_at"`
}

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
	TenantID  string `json:"tenant_id"`
}

type RegisterRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	CompanyName     string `json:"company_name,omitempty"`
}

func NewAuthService(cfg *config.Config, api *APIService) *AuthService {
	return &AuthService{
		config: cfg,
		api:    api,
	}
}

// Login authenticates a user
func (s *AuthService) Login(email, password string) (*LoginResponse, error) {
	loginReq := LoginRequest{
		Email:    email,
		Password: password,
	}

	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return nil, err
	}

	resp, err := s.api.Post("/api/v1/auth/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return nil, err
	}

	return &loginResp, nil
}

// Register creates a new user account
func (s *AuthService) Register(req RegisterRequest) error {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := s.api.Post("/api/v1/auth/register", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("registration failed with status: %d", resp.StatusCode)
	}

	return nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(token string) bool {
	// Make request to backend API to validate token
	req, err := http.NewRequest("GET", s.config.BackendURL+"/api/v1/auth/me", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetCurrentUser gets the current user info from token
func (s *AuthService) GetCurrentUser(token string) (*User, error) {
	req, err := http.NewRequest("GET", s.config.BackendURL+"/api/v1/auth/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get user info")
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}

	return &user, nil
}

// ForgotPassword initiates password reset
func (s *AuthService) ForgotPassword(email string) error {
	data := map[string]string{"email": email}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := s.api.Post("/api/v1/auth/forgot-password", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send password reset email")
	}

	return nil
}

// ResetPassword resets password with token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	data := map[string]string{
		"token":    token,
		"password": newPassword,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	resp, err := s.api.Post("/api/v1/auth/reset-password", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to reset password")
	}

	return nil
}