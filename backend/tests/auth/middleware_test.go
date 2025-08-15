package auth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pageza/landscaping-app/backend/internal/auth"
	"github.com/pageza/landscaping-app/backend/internal/config"
	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/middleware"
)

// MockAuthService is a mock implementation of auth.AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ComparePassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

func (m *MockAuthService) GenerateTokens(user *domain.EnhancedUser, sessionID uuid.UUID) (*auth.TokenPair, error) {
	args := m.Called(user, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string, tokenType auth.TokenType) (*auth.Claims, error) {
	args := m.Called(tokenString, tokenType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockAuthService) RefreshTokens(refreshToken string) (*auth.TokenPair, error) {
	args := m.Called(refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.TokenPair), args.Error(1)
}

func (m *MockAuthService) RevokeToken(sessionID uuid.UUID) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockAuthService) GenerateAPIKey(name string, permissions []string) (*auth.APIKeyPair, error) {
	args := m.Called(name, permissions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.APIKeyPair), args.Error(1)
}

func (m *MockAuthService) ValidateAPIKey(keyString string) (*auth.APIKeyClaims, error) {
	args := m.Called(keyString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.APIKeyClaims), args.Error(1)
}

func (m *MockAuthService) CreateSession(userID uuid.UUID, deviceInfo map[string]interface{}, ipAddress, userAgent string) (*domain.UserSession, error) {
	args := m.Called(userID, deviceInfo, ipAddress, userAgent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserSession), args.Error(1)
}

func (m *MockAuthService) ValidateSession(sessionID uuid.UUID) (*domain.UserSession, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserSession), args.Error(1)
}

func (m *MockAuthService) RevokeSession(sessionID uuid.UUID) error {
	args := m.Called(sessionID)
	return args.Error(0)
}

func (m *MockAuthService) RevokeAllUserSessions(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) GenerateTOTPSecret() (string, error) {
	args := m.Called()
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateTOTP(secret, token string) bool {
	args := m.Called(secret, token)
	return args.Bool(0)
}

func (m *MockAuthService) GenerateBackupCodes() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// Helper functions
func createTestConfig() *config.Config {
	return &config.Config{
		CORSAllowedOrigins:         []string{"http://localhost:3000"},
		RateLimitRequestsPerMinute: 100,
	}
}

func createTestRedisClient() *redis.Client {
	// Use Redis client that connects to a test Redis instance
	// In practice, you might use miniredis for testing
	return redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
		DB:   1, // Use a test database
	})
}

func createTestClaims() *auth.Claims {
	return &auth.Claims{
		UserID:      uuid.New(),
		TenantID:    uuid.New(),
		Role:        domain.RoleUser,
		TokenType:   auth.AccessToken,
		SessionID:   uuid.New(),
		Permissions: []string{domain.PermissionCustomerManage, domain.PermissionJobManage},
	}
}

// TestRequestID tests the RequestID middleware
func TestRequestID(t *testing.T) {
	cfg := createTestConfig()
	mockAuth := &MockAuthService{}
	redisClient := createTestRedisClient()
	
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, redisClient)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(middleware.RequestIDKey)
		assert.NotNil(t, requestID)
		assert.IsType(t, "", requestID)
		w.WriteHeader(http.StatusOK)
	})
	
	wrappedHandler := mw.RequestID(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.NotEmpty(t, rr.Header().Get("X-Request-ID"))
}

// TestJWTAuth tests the JWT authentication middleware
func TestJWTAuth(t *testing.T) {
	cfg := createTestConfig()
	mockAuth := &MockAuthService{}
	redisClient := createTestRedisClient()
	
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, redisClient)
	
	claims := createTestClaims()
	session := createTestSession(claims.UserID)
	
	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func()
		expectedStatus int
		checkContext   bool
	}{
		{
			name:       "missing authorization header",
			authHeader: "",
			mockSetup:  func() {},
			expectedStatus: http.StatusUnauthorized,
			checkContext: false,
		},
		{
			name:       "invalid authorization header format",
			authHeader: "InvalidFormat token123",
			mockSetup:  func() {},
			expectedStatus: http.StatusUnauthorized,
			checkContext: false,
		},
		{
			name:       "empty bearer token",
			authHeader: "Bearer ",
			mockSetup:  func() {},
			expectedStatus: http.StatusUnauthorized,
			checkContext: false,
		},
		{
			name:       "valid token and session",
			authHeader: "Bearer valid-token",
			mockSetup: func() {
				mockAuth.On("ValidateToken", "valid-token", auth.AccessToken).Return(claims, nil)
				mockAuth.On("ValidateSession", claims.SessionID).Return(session, nil)
			},
			expectedStatus: http.StatusOK,
			checkContext: true,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			mockSetup: func() {
				mockAuth.On("ValidateToken", "invalid-token", auth.AccessToken).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext: false,
		},
		{
			name:       "valid token but invalid session",
			authHeader: "Bearer valid-token",
			mockSetup: func() {
				mockAuth.On("ValidateToken", "valid-token", auth.AccessToken).Return(claims, nil)
				mockAuth.On("ValidateSession", claims.SessionID).Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockAuth.ExpectedCalls = nil
			tt.mockSetup()
			
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext {
					userID := r.Context().Value(middleware.UserIDKey)
					tenantID := r.Context().Value(middleware.TenantIDKey)
					role := r.Context().Value(middleware.UserRoleKey)
					permissions := r.Context().Value(middleware.PermissionsKey)
					
					assert.Equal(t, claims.UserID, userID)
					assert.Equal(t, claims.TenantID, tenantID)
					assert.Equal(t, claims.Role, role)
					assert.Equal(t, claims.Permissions, permissions)
				}
				w.WriteHeader(http.StatusOK)
			})
			
			wrappedHandler := mw.JWTAuth(handler)
			
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rr := httptest.NewRecorder()
			
			wrappedHandler.ServeHTTP(rr, req)
			
			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			if tt.expectedStatus != http.StatusOK {
				var errorResp domain.ErrorResponse
				err := json.Unmarshal(rr.Body.Bytes(), &errorResp)
				assert.NoError(t, err)
				assert.NotEmpty(t, errorResp.Message)
			}
			
			mockAuth.AssertExpectations(t)
		})
	}
}

// TestAPIKeyAuth tests the API key authentication middleware
func TestAPIKeyAuth(t *testing.T) {
	cfg := createTestConfig()
	mockAuth := &MockAuthService{}
	redisClient := createTestRedisClient()
	
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, redisClient)
	
	apiKeyClaims := &auth.APIKeyClaims{
		TenantID:    uuid.New(),
		KeyID:       uuid.New(),
		Permissions: []string{domain.PermissionCustomerManage},
	}
	
	tests := []struct {
		name           string
		apiKeyHeader   string
		mockSetup      func()
		expectedStatus int
		checkContext   bool
	}{
		{
			name:         "missing API key header",
			apiKeyHeader: "",
			mockSetup:    func() {},
			expectedStatus: http.StatusUnauthorized,
			checkContext: false,
		},
		{
			name:         "valid API key",
			apiKeyHeader: "valid-api-key",
			mockSetup: func() {
				mockAuth.On("ValidateAPIKey", "valid-api-key").Return(apiKeyClaims, nil)
			},
			expectedStatus: http.StatusOK,
			checkContext: true,
		},
		{
			name:         "invalid API key",
			apiKeyHeader: "invalid-api-key",
			mockSetup: func() {
				mockAuth.On("ValidateAPIKey", "invalid-api-key").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			checkContext: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockAuth.ExpectedCalls = nil
			tt.mockSetup()
			
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkContext {
					tenantID := r.Context().Value(middleware.TenantIDKey)
					permissions := r.Context().Value(middleware.PermissionsKey)
					role := r.Context().Value(middleware.UserRoleKey)
					
					assert.Equal(t, apiKeyClaims.TenantID, tenantID)
					assert.Equal(t, apiKeyClaims.Permissions, permissions)
					assert.Equal(t, "api_key", role)
				}
				w.WriteHeader(http.StatusOK)
			})
			
			wrappedHandler := mw.APIKeyAuth(handler)
			
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.apiKeyHeader != "" {
				req.Header.Set("X-API-Key", tt.apiKeyHeader)
			}
			rr := httptest.NewRecorder()
			
			wrappedHandler.ServeHTTP(rr, req)
			
			assert.Equal(t, tt.expectedStatus, rr.Code)
			
			mockAuth.AssertExpectations(t)
		})
	}
}

// TestRequirePermission tests the permission checking middleware
func TestRequirePermission(t *testing.T) {
	cfg := createTestConfig()
	mockAuth := &MockAuthService{}
	redisClient := createTestRedisClient()
	
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, redisClient)
	
	tests := []struct {
		name               string
		userPermissions    []string
		requiredPermission string
		expectedStatus     int
	}{
		{
			name:               "has required permission",
			userPermissions:    []string{domain.PermissionCustomerManage, domain.PermissionJobManage},
			requiredPermission: domain.PermissionCustomerManage,
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "has wildcard permission",
			userPermissions:    []string{"*"},
			requiredPermission: domain.PermissionCustomerManage,
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "missing required permission",
			userPermissions:    []string{domain.PermissionJobManage},
			requiredPermission: domain.PermissionCustomerManage,
			expectedStatus:     http.StatusForbidden,
		},
		{
			name:               "no permissions",
			userPermissions:    []string{},
			requiredPermission: domain.PermissionCustomerManage,
			expectedStatus:     http.StatusForbidden,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			
			wrappedHandler := mw.RequirePermission(tt.requiredPermission)(handler)
			
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), middleware.PermissionsKey, tt.userPermissions)
			req = req.WithContext(ctx)
			
			rr := httptest.NewRecorder()
			
			wrappedHandler.ServeHTTP(rr, req)
			
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

// TestRequireRole tests the role checking middleware
func TestRequireRole(t *testing.T) {
	cfg := createTestConfig()
	mockAuth := &MockAuthService{}
	redisClient := createTestRedisClient()
	
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, redisClient)
	
	tests := []struct {
		name           string
		userRole       string
		requiredRoles  []string
		expectedStatus int
	}{
		{
			name:           "has required role",
			userRole:       domain.RoleAdmin,
			requiredRoles:  []string{domain.RoleAdmin, domain.RoleOwner},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "has one of multiple required roles",
			userRole:       domain.RoleUser,
			requiredRoles:  []string{domain.RoleAdmin, domain.RoleUser},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing required role",
			userRole:       domain.RoleUser,
			requiredRoles:  []string{domain.RoleAdmin, domain.RoleOwner},
			expectedStatus: http.StatusForbidden,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			
			wrappedHandler := mw.RequireRole(tt.requiredRoles...)(handler)
			
			req := httptest.NewRequest("GET", "/test", nil)
			ctx := context.WithValue(req.Context(), middleware.UserRoleKey, tt.userRole)
			req = req.WithContext(ctx)
			
			rr := httptest.NewRecorder()
			
			wrappedHandler.ServeHTTP(rr, req)
			
			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}

// TestSecurityHeaders tests the security headers middleware
func TestSecurityHeaders(t *testing.T) {
	cfg := createTestConfig()
	cfg.Env = "production" // Set to production to test CSP header
	
	mockAuth := &MockAuthService{}
	redisClient := createTestRedisClient()
	
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, redisClient)
	
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	wrappedHandler := mw.SecurityHeaders(handler)
	
	req := httptest.NewRequest("GET", "/test", nil)
	rr := httptest.NewRecorder()
	
	wrappedHandler.ServeHTTP(rr, req)
	
	assert.Equal(t, http.StatusOK, rr.Code)
	
	// Check security headers
	assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", rr.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "max-age=31536000; includeSubDomains", rr.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "strict-origin-when-cross-origin", rr.Header().Get("Referrer-Policy"))
	assert.Equal(t, "default-src 'self'", rr.Header().Get("Content-Security-Policy"))
}

// TestPagination tests the pagination middleware
func TestPagination(t *testing.T) {
	cfg := createTestConfig()
	mockAuth := &MockAuthService{}
	redisClient := createTestRedisClient()
	
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, redisClient)
	
	tests := []struct {
		name            string
		queryParams     string
		expectedPage    int
		expectedPerPage int
		expectedOffset  int
	}{
		{
			name:            "default values",
			queryParams:     "",
			expectedPage:    1,
			expectedPerPage: 20,
			expectedOffset:  0,
		},
		{
			name:            "custom page and per_page",
			queryParams:     "?page=3&per_page=10",
			expectedPage:    3,
			expectedPerPage: 10,
			expectedOffset:  20,
		},
		{
			name:            "invalid page defaults to 1",
			queryParams:     "?page=0&per_page=15",
			expectedPage:    1,
			expectedPerPage: 15,
			expectedOffset:  0,
		},
		{
			name:            "per_page too large capped at default",
			queryParams:     "?page=2&per_page=200",
			expectedPage:    2,
			expectedPerPage: 20,
			expectedOffset:  20,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				page := r.Context().Value("page")
				perPage := r.Context().Value("per_page")
				offset := r.Context().Value("offset")
				
				assert.Equal(t, tt.expectedPage, page)
				assert.Equal(t, tt.expectedPerPage, perPage)
				assert.Equal(t, tt.expectedOffset, offset)
				
				w.WriteHeader(http.StatusOK)
			})
			
			wrappedHandler := mw.Pagination(handler)
			
			req := httptest.NewRequest("GET", "/test"+tt.queryParams, nil)
			rr := httptest.NewRecorder()
			
			wrappedHandler.ServeHTTP(rr, req)
			
			assert.Equal(t, http.StatusOK, rr.Code)
		})
	}
}