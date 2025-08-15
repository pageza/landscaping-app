package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pageza/landscaping-app/backend/internal/config"
	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/internal/middleware"
)

// MockAuthService for testing authentication middleware
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(token string, tokenType interface{}) (*domain.JWTClaims, error) {
	args := m.Called(token, tokenType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.JWTClaims), args.Error(1)
}

func (m *MockAuthService) ValidateSession(sessionID uuid.UUID) (*domain.UserSession, error) {
	args := m.Called(sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserSession), args.Error(1)
}

// MockTenantService for testing tenant context middleware
type MockTenantService struct {
	mock.Mock
}

func (m *MockTenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tenant), args.Error(1)
}

// MockRateLimitService for testing rate limiting middleware
type MockRateLimitService struct {
	mock.Mock
}

func (m *MockRateLimitService) CheckRateLimit(key string, requests int, duration time.Duration) (bool, error) {
	args := m.Called(key, requests, duration)
	return args.Bool(0), args.Error(1)
}

// Test helper to create a simple HTTP handler
func createTestHandler(message string, statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		w.Write([]byte(message))
	}
}

// Test helper to create middleware configuration
func createTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret:          "test-secret",
			AccessTokenTTL:  time.Hour,
			RefreshTokenTTL: 24 * time.Hour,
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerMinute: 60,
			BurstSize:        10,
		},
		Security: config.SecurityConfig{
			EnableCSRF:    true,
			TrustedOrigins: []string{"https://app.example.com"},
		},
	}
}

func TestRequestIDMiddleware(t *testing.T) {
	cfg := createTestConfig()
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)

	handler := mw.RequestID(createTestHandler("OK", http.StatusOK))

	t.Run("generates request ID when not present", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		requestID := w.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)
		assert.Len(t, requestID, 36) // UUID length
	})

	t.Run("preserves existing request ID", func(t *testing.T) {
		existingID := uuid.New().String()
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", existingID)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		requestID := w.Header().Get("X-Request-ID")
		assert.Equal(t, existingID, requestID)
	})

	t.Run("adds request ID to context", func(t *testing.T) {
		var contextRequestID string
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextRequestID = r.Context().Value("request_id").(string)
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.RequestID(testHandler)
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, contextRequestID)
		assert.Equal(t, contextRequestID, w.Header().Get("X-Request-ID"))
	})
}

func TestCORSMiddleware(t *testing.T) {
	cfg := createTestConfig()
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)

	handler := mw.EnhancedCORS(createTestHandler("OK", http.StatusOK))

	t.Run("handles preflight request", func(t *testing.T) {
		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "https://app.example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type, Authorization")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
		assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	})

	t.Run("sets CORS headers for actual request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://app.example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("rejects unauthorized origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://malicious.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("allows requests without origin", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	cfg := createTestConfig()
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)

	handler := mw.SecurityHeaders(createTestHandler("OK", http.StatusOK))

	t.Run("sets all required security headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=")
		assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src")
		assert.Equal(t, "noopen, nosniff", w.Header().Get("X-Download-Options"))
	})

	t.Run("sets referrer policy", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	})
}

func TestJWTAuthMiddleware(t *testing.T) {
	cfg := createTestConfig()
	mockAuth := new(MockAuthService)
	mockTenant := new(MockTenantService)
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, mockTenant)

	handler := mw.JWTAuth(createTestHandler("OK", http.StatusOK))

	t.Run("successful authentication with valid token", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()
		sessionID := uuid.New()

		claims := &domain.JWTClaims{
			UserID:    userID,
			TenantID:  tenantID,
			SessionID: sessionID,
			Role:      "admin",
			TokenType: "access",
		}

		session := &domain.UserSession{
			ID:       sessionID,
			UserID:   userID,
			Status:   "active",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockAuth.On("ValidateToken", "valid.jwt.token", mock.Anything).Return(claims, nil).Once()
		mockAuth.On("ValidateSession", sessionID).Return(session, nil).Once()

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid.jwt.token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockAuth.AssertExpectations(t)
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "missing authorization header")
	})

	t.Run("invalid authorization format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid authorization format")
	})

	t.Run("invalid token", func(t *testing.T) {
		mockAuth.On("ValidateToken", "invalid.jwt.token", mock.Anything).Return(nil, fmt.Errorf("invalid token")).Once()

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid.jwt.token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid token")
		mockAuth.AssertExpectations(t)
	})

	t.Run("expired session", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()
		sessionID := uuid.New()

		claims := &domain.JWTClaims{
			UserID:    userID,
			TenantID:  tenantID,
			SessionID: sessionID,
			Role:      "admin",
			TokenType: "access",
		}

		mockAuth.On("ValidateToken", "valid.jwt.token", mock.Anything).Return(claims, nil).Once()
		mockAuth.On("ValidateSession", sessionID).Return(nil, fmt.Errorf("session expired")).Once()

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid.jwt.token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "session expired")
		mockAuth.AssertExpectations(t)
	})

	t.Run("adds user context", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()
		sessionID := uuid.New()
		var contextUserID, contextTenantID uuid.UUID
		var contextRole string

		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contextUserID = r.Context().Value("user_id").(uuid.UUID)
			contextTenantID = r.Context().Value("tenant_id").(uuid.UUID)
			contextRole = r.Context().Value("user_role").(string)
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.JWTAuth(testHandler)

		claims := &domain.JWTClaims{
			UserID:    userID,
			TenantID:  tenantID,
			SessionID: sessionID,
			Role:      "admin",
			TokenType: "access",
		}

		session := &domain.UserSession{
			ID:       sessionID,
			UserID:   userID,
			Status:   "active",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		mockAuth.On("ValidateToken", "valid.jwt.token", mock.Anything).Return(claims, nil).Once()
		mockAuth.On("ValidateSession", sessionID).Return(session, nil).Once()

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer valid.jwt.token")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, userID, contextUserID)
		assert.Equal(t, tenantID, contextTenantID)
		assert.Equal(t, "admin", contextRole)
		mockAuth.AssertExpectations(t)
	})
}

func TestTenantContextMiddleware(t *testing.T) {
	cfg := createTestConfig()
	mockAuth := new(MockAuthService)
	mockTenant := new(MockTenantService)
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, mockTenant)

	handler := mw.TenantContext(createTestHandler("OK", http.StatusOK))

	t.Run("successful tenant context setup", func(t *testing.T) {
		tenantID := uuid.New()
		tenant := &domain.Tenant{
			ID:     tenantID,
			Name:   "Test Tenant",
			Status: "active",
		}

		mockTenant.On("GetTenant", mock.Anything, tenantID).Return(tenant, nil).Once()

		// Create a context with tenant_id (typically set by JWT middleware)
		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockTenant.AssertExpectations(t)
	})

	t.Run("tenant not found", func(t *testing.T) {
		tenantID := uuid.New()
		mockTenant.On("GetTenant", mock.Anything, tenantID).Return(nil, fmt.Errorf("tenant not found")).Once()

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "tenant not found")
		mockTenant.AssertExpectations(t)
	})

	t.Run("inactive tenant", func(t *testing.T) {
		tenantID := uuid.New()
		tenant := &domain.Tenant{
			ID:     tenantID,
			Name:   "Suspended Tenant",
			Status: "suspended",
		}

		mockTenant.On("GetTenant", mock.Anything, tenantID).Return(tenant, nil).Once()

		ctx := context.WithValue(context.Background(), "tenant_id", tenantID)
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "tenant is not active")
		mockTenant.AssertExpectations(t)
	})

	t.Run("missing tenant context", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "missing tenant context")
	})
}

func TestRolePermissionMiddleware(t *testing.T) {
	cfg := createTestConfig()
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)

	t.Run("require role - success", func(t *testing.T) {
		handler := mw.RequireRole("admin")(createTestHandler("OK", http.StatusOK))

		ctx := context.WithValue(context.Background(), "user_role", "admin")
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("require role - multiple roles success", func(t *testing.T) {
		handler := mw.RequireRole("admin", "owner")(createTestHandler("OK", http.StatusOK))

		ctx := context.WithValue(context.Background(), "user_role", "owner")
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("require role - insufficient role", func(t *testing.T) {
		handler := mw.RequireRole("admin")(createTestHandler("OK", http.StatusOK))

		ctx := context.WithValue(context.Background(), "user_role", "user")
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "insufficient role")
	})

	t.Run("require permission - success", func(t *testing.T) {
		handler := mw.RequirePermission("customer:manage")(createTestHandler("OK", http.StatusOK))

		permissions := []string{"customer:manage", "job:manage"}
		ctx := context.WithValue(context.Background(), "user_permissions", permissions)
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("require permission - wildcard permission", func(t *testing.T) {
		handler := mw.RequirePermission("customer:manage")(createTestHandler("OK", http.StatusOK))

		permissions := []string{"*"}
		ctx := context.WithValue(context.Background(), "user_permissions", permissions)
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("require permission - missing permission", func(t *testing.T) {
		handler := mw.RequirePermission("customer:manage")(createTestHandler("OK", http.StatusOK))

		permissions := []string{"job:manage"}
		ctx := context.WithValue(context.Background(), "user_permissions", permissions)
		req := httptest.NewRequest("GET", "/test", nil).WithContext(ctx)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "insufficient permissions")
	})

	t.Run("missing role context", func(t *testing.T) {
		handler := mw.RequireRole("admin")(createTestHandler("OK", http.StatusOK))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		assert.Contains(t, w.Body.String(), "missing role context")
	})
}

func TestRateLimitMiddleware(t *testing.T) {
	cfg := createTestConfig()
	mockRateLimit := new(MockRateLimitService)
	
	// Note: In real implementation, rate limit service would be injected
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)
	handler := mw.RateLimit(createTestHandler("OK", http.StatusOK))

	t.Run("allows requests within limit", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("rate limit headers", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	})

	// Note: Testing actual rate limiting would require either:
	// 1. Mocking the rate limiter
	// 2. Making many rapid requests
	// 3. Using a test rate limiter with very low limits
}

func TestLoggingMiddleware(t *testing.T) {
	cfg := createTestConfig()
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)

	handler := mw.EnhancedLogging(createTestHandler("OK", http.StatusOK))

	t.Run("logs request and response", func(t *testing.T) {
		// In a real implementation, you'd capture log output
		// For this test, we just verify the middleware doesn't break the flow
		req := httptest.NewRequest("GET", "/test?param=value", nil)
		req.Header.Set("User-Agent", "Test Agent")
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "OK", w.Body.String())
	})

	t.Run("logs different HTTP methods", func(t *testing.T) {
		methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

		for _, method := range methods {
			req := httptest.NewRequest(method, "/test", strings.NewReader(`{"test": "data"}`))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
		}
	})

	t.Run("logs error responses", func(t *testing.T) {
		errorHandler := mw.EnhancedLogging(createTestHandler("Error", http.StatusInternalServerError))

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		errorHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestPaginationMiddleware(t *testing.T) {
	cfg := createTestConfig()
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)

	handler := mw.Pagination(createTestHandler("OK", http.StatusOK))

	t.Run("default pagination parameters", func(t *testing.T) {
		var offset, limit int
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			offset = r.Context().Value("offset").(int)
			limit = r.Context().Value("limit").(int)
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.Pagination(testHandler)
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 0, offset)
		assert.Equal(t, 50, limit) // Default limit
	})

	t.Run("custom pagination parameters", func(t *testing.T) {
		var offset, limit int
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			offset = r.Context().Value("offset").(int)
			limit = r.Context().Value("limit").(int)
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.Pagination(testHandler)
		req := httptest.NewRequest("GET", "/test?offset=20&limit=10", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 20, offset)
		assert.Equal(t, 10, limit)
	})

	t.Run("validates pagination parameters", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test?offset=-1&limit=0", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid pagination parameters")
	})

	t.Run("limits maximum page size", func(t *testing.T) {
		var limit int
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limit = r.Context().Value("limit").(int)
			w.WriteHeader(http.StatusOK)
		})

		handler := mw.Pagination(testHandler)
		req := httptest.NewRequest("GET", "/test?limit=1000", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 100, limit) // Max limit should be enforced
	})
}

func TestMiddlewareChaining(t *testing.T) {
	cfg := createTestConfig()
	mockAuth := new(MockAuthService)
	mockTenant := new(MockTenantService)
	mw := middleware.NewEnhancedMiddleware(cfg, mockAuth, mockTenant)

	// Test chaining multiple middleware
	t.Run("full middleware chain", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()
		sessionID := uuid.New()

		claims := &domain.JWTClaims{
			UserID:    userID,
			TenantID:  tenantID,
			SessionID: sessionID,
			Role:      "admin",
			TokenType: "access",
		}

		session := &domain.UserSession{
			ID:       sessionID,
			UserID:   userID,
			Status:   "active",
			ExpiresAt: time.Now().Add(time.Hour),
		}

		tenant := &domain.Tenant{
			ID:     tenantID,
			Name:   "Test Tenant",
			Status: "active",
		}

		mockAuth.On("ValidateToken", "valid.jwt.token", mock.Anything).Return(claims, nil).Once()
		mockAuth.On("ValidateSession", sessionID).Return(session, nil).Once()
		mockTenant.On("GetTenant", mock.Anything, tenantID).Return(tenant, nil).Once()

		// Create a full middleware chain
		handler := mw.RequestID(
			mw.EnhancedCORS(
				mw.SecurityHeaders(
					mw.EnhancedLogging(
						mw.RateLimit(
							mw.JWTAuth(
								mw.TenantContext(
									mw.RequireRole("admin")(
										mw.Pagination(
											createTestHandler("OK", http.StatusOK),
										),
									),
								),
							),
						),
					),
				),
			),
		)

		req := httptest.NewRequest("GET", "/test?offset=10&limit=20", nil)
		req.Header.Set("Authorization", "Bearer valid.jwt.token")
		req.Header.Set("Origin", "https://app.example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
		assert.Equal(t, "https://app.example.com", w.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))

		mockAuth.AssertExpectations(t)
		mockTenant.AssertExpectations(t)
	})
}

// Benchmark tests for middleware performance
func BenchmarkRequestIDMiddleware(b *testing.B) {
	cfg := createTestConfig()
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)
	handler := mw.RequestID(createTestHandler("OK", http.StatusOK))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkSecurityHeadersMiddleware(b *testing.B) {
	cfg := createTestConfig()
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)
	handler := mw.SecurityHeaders(createTestHandler("OK", http.StatusOK))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

func BenchmarkFullMiddlewareChain(b *testing.B) {
	cfg := createTestConfig()
	mw := middleware.NewEnhancedMiddleware(cfg, nil, nil)
	
	handler := mw.RequestID(
		mw.SecurityHeaders(
			mw.EnhancedLogging(
				mw.Pagination(
					createTestHandler("OK", http.StatusOK),
				),
			),
		),
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}