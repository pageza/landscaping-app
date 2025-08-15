package services

import (
	"io"
	"net/http"
	"time"

	"github.com/pageza/landscaping-app/web/internal/config"
)

type APIService struct {
	config     *config.Config
	httpClient *http.Client
	baseURL    string
}

func NewAPIService(cfg *config.Config) *APIService {
	return &APIService{
		config:  cfg,
		baseURL: cfg.BackendURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Get makes a GET request to the backend API
func (s *APIService) Get(path string) (*http.Response, error) {
	return s.makeRequest("GET", path, nil)
}

// Post makes a POST request to the backend API
func (s *APIService) Post(path string, body io.Reader) (*http.Response, error) {
	return s.makeRequest("POST", path, body)
}

// Put makes a PUT request to the backend API
func (s *APIService) Put(path string, body io.Reader) (*http.Response, error) {
	return s.makeRequest("PUT", path, body)
}

// Delete makes a DELETE request to the backend API
func (s *APIService) Delete(path string) (*http.Response, error) {
	return s.makeRequest("DELETE", path, nil)
}

// AuthenticatedGet makes a GET request with authentication
func (s *APIService) AuthenticatedGet(path, token string) (*http.Response, error) {
	req, err := http.NewRequest("GET", s.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return s.httpClient.Do(req)
}

// AuthenticatedPost makes a POST request with authentication
func (s *APIService) AuthenticatedPost(path, token string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", s.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return s.httpClient.Do(req)
}

// AuthenticatedPut makes a PUT request with authentication
func (s *APIService) AuthenticatedPut(path, token string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("PUT", s.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return s.httpClient.Do(req)
}

// AuthenticatedDelete makes a DELETE request with authentication
func (s *APIService) AuthenticatedDelete(path, token string) (*http.Response, error) {
	req, err := http.NewRequest("DELETE", s.baseURL+path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	return s.httpClient.Do(req)
}

// makeRequest is a helper method for making HTTP requests
func (s *APIService) makeRequest(method, path string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, s.baseURL+path, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return s.httpClient.Do(req)
}