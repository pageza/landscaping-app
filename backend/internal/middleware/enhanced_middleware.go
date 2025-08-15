package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"

	"github.com/pageza/landscaping-app/backend/internal/auth"
	"github.com/pageza/landscaping-app/backend/internal/config"
	"github.com/pageza/landscaping-app/backend/internal/domain"
	"github.com/pageza/landscaping-app/backend/pkg/security"
)

// Context keys for request context
type contextKey string

const (
	UserIDKey      contextKey = "user_id"
	TenantIDKey    contextKey = "tenant_id"
	UserRoleKey    contextKey = "user_role"
	SessionIDKey   contextKey = "session_id"
	PermissionsKey contextKey = "permissions"
	RequestIDKey   contextKey = "request_id"
)

// EnhancedMiddleware provides comprehensive middleware functionality
type EnhancedMiddleware struct {
	config         *config.Config
	authService    auth.AuthService
	redisClient    *redis.Client
	rateLimiter    security.RateLimiter
}

// NewEnhancedMiddleware creates a new enhanced middleware instance
func NewEnhancedMiddleware(config *config.Config, authService auth.AuthService, redisClient *redis.Client) *EnhancedMiddleware {
	// Create Redis-based rate limiter
	rateLimiter := security.NewRedisRateLimiter(
		redisClient,
		config.RateLimitRequestsPerMinute,
		time.Minute,
	)

	return &EnhancedMiddleware{
		config:      config,
		authService: authService,
		redisClient: redisClient,
		rateLimiter: rateLimiter,
	}
}

// RequestID adds a unique request ID to each request
func (m *EnhancedMiddleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New().String()
		ctx := context.WithValue(r.Context(), RequestIDKey, requestID)
		w.Header().Set("X-Request-ID", requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// EnhancedCORS handles Cross-Origin Resource Sharing with dynamic origin validation
func (m *EnhancedMiddleware) EnhancedCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Check if origin is in allowed list or if wildcard is allowed
		allowed := false
		for _, allowedOrigin := range m.config.CORSAllowedOrigins {
			if allowedOrigin == "*" || origin == allowedOrigin {
				allowed = true
				break
			}
			// Check for subdomain patterns (e.g., *.example.com)
			if strings.HasPrefix(allowedOrigin, "*.") {
				domain := strings.TrimPrefix(allowedOrigin, "*.")
				if strings.HasSuffix(origin, domain) {
					allowed = true
					break
				}
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID, X-API-Key, X-Request-ID")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EnhancedLogging provides comprehensive request logging
func (m *EnhancedMiddleware) EnhancedLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		requestID := r.Context().Value(RequestIDKey)

		// Create a response writer wrapper to capture status code and size
		wrapped := &enhancedResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		
		// Extract context values
		userID := r.Context().Value(UserIDKey)
		tenantID := r.Context().Value(TenantIDKey)

		// Create structured log entry
		logEntry := map[string]interface{}{
			"timestamp":   start.Format(time.RFC3339),
			"request_id":  requestID,
			"method":      r.Method,
			"path":        r.URL.Path,
			"query":       r.URL.RawQuery,
			"status_code": wrapped.statusCode,
			"duration_ms": duration.Milliseconds(),
			"size_bytes":  wrapped.size,
			"user_agent":  r.UserAgent(),
			"ip_address":  getClientIP(r),
			"user_id":     userID,
			"tenant_id":   tenantID,
		}

		// Convert to JSON for structured logging
		logJSON, _ := json.Marshal(logEntry)
		fmt.Printf("%s\n", logJSON)
	})
}

// JWTAuth validates JWT tokens and sets user context
func (m *EnhancedMiddleware) JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.writeErrorResponse(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			m.writeErrorResponse(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			m.writeErrorResponse(w, "Token required", http.StatusUnauthorized)
			return
		}

		// Validate JWT token
		claims, err := m.authService.ValidateToken(token, auth.AccessToken)
		if err != nil {
			m.writeErrorResponse(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Validate session
		session, err := m.authService.ValidateSession(claims.SessionID)
		if err != nil {
			m.writeErrorResponse(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		// Set context values
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, TenantIDKey, claims.TenantID)
		ctx = context.WithValue(ctx, UserRoleKey, claims.Role)
		ctx = context.WithValue(ctx, SessionIDKey, claims.SessionID)
		ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)

		// Update session activity
		go func() {
			// Non-blocking session update
			_, _ = m.authService.ValidateSession(session.ID)
		}()

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// APIKeyAuth validates API keys for machine-to-machine authentication
func (m *EnhancedMiddleware) APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract API key from header
		apiKey := r.Header.Get("X-API-Key")
		if apiKey == "" {
			m.writeErrorResponse(w, "API key required", http.StatusUnauthorized)
			return
		}

		// Validate API key
		keyClaims, err := m.authService.ValidateAPIKey(apiKey)
		if err != nil {
			m.writeErrorResponse(w, "Invalid API key", http.StatusUnauthorized)
			return
		}

		// Set context values for API key authentication
		ctx := context.WithValue(r.Context(), TenantIDKey, keyClaims.TenantID)
		ctx = context.WithValue(ctx, PermissionsKey, keyClaims.Permissions)
		ctx = context.WithValue(ctx, UserRoleKey, "api_key")

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TenantContext sets tenant context for Row Level Security
func (m *EnhancedMiddleware) TenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tenantID := r.Context().Value(TenantIDKey)
		userID := r.Context().Value(UserIDKey)
		userRole := r.Context().Value(UserRoleKey)
		
		// TODO: Use userID and userRole for RLS context
		_ = userID
		_ = userRole

		if tenantID == nil {
			m.writeErrorResponse(w, "Tenant context not found", http.StatusInternalServerError)
			return
		}

		// TODO: Set database context for RLS
		// This would typically involve setting session variables for PostgreSQL RLS
		// Example: SET app.current_tenant_id = 'tenant-uuid'

		next.ServeHTTP(w, r)
	})
}

// RateLimit implements advanced rate limiting with different limits per endpoint
func (m *EnhancedMiddleware) RateLimit(requestsPerMinute int) func(http.Handler) http.Handler {
	// Create a custom rate limiter for this specific limit
	limiter := security.NewRedisRateLimiter(m.redisClient, requestsPerMinute, time.Minute)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Generate rate limit key based on IP and optionally user ID
			key := m.generateRateLimitKey(r)
			
			// Check rate limit
			allowed, err := limiter.Allow(r.Context(), key)
			if err != nil {
				m.writeErrorResponse(w, "Rate limit check failed", http.StatusInternalServerError)
				return
			}
			
			if !allowed {
				// Get rate limit info for headers
				info, _ := limiter.GetInfo(r.Context(), key)
				if info != nil {
					w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
					w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
					w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))
				}
				
				m.writeErrorResponse(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			// Add rate limit info to response headers
			if info, err := limiter.GetInfo(r.Context(), key); err == nil && info != nil {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission checks if user has required permission
func (m *EnhancedMiddleware) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			permissions, ok := r.Context().Value(PermissionsKey).([]string)
			if !ok {
				m.writeErrorResponse(w, "Permissions not found in context", http.StatusInternalServerError)
				return
			}

			if !auth.HasPermission(permissions, permission) {
				m.writeErrorResponse(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireRole checks if user has required role
func (m *EnhancedMiddleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole, ok := r.Context().Value(UserRoleKey).(string)
			if !ok {
				m.writeErrorResponse(w, "User role not found in context", http.StatusInternalServerError)
				return
			}

			// Check if user has any of the required roles
			hasRole := false
			for _, role := range roles {
				if userRole == role {
					hasRole = true
					break
				}
			}

			if !hasRole {
				m.writeErrorResponse(w, "Insufficient role permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantValidation validates tenant access for the current user
func (m *EnhancedMiddleware) TenantValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userTenantID, ok := r.Context().Value(TenantIDKey).(uuid.UUID)
		if !ok {
			m.writeErrorResponse(w, "User tenant ID not found", http.StatusInternalServerError)
			return
		}

		userRole, ok := r.Context().Value(UserRoleKey).(string)
		if !ok {
			m.writeErrorResponse(w, "User role not found", http.StatusInternalServerError)
			return
		}

		// Extract tenant ID from URL path if present
		vars := mux.Vars(r)
		if tenantIDStr, exists := vars["tenantId"]; exists {
			requestedTenantID, err := uuid.Parse(tenantIDStr)
			if err != nil {
				m.writeErrorResponse(w, "Invalid tenant ID format", http.StatusBadRequest)
				return
			}

			// Check if user can access the requested tenant
			if !auth.CanAccessTenant(userTenantID, requestedTenantID, userRole) {
				m.writeErrorResponse(w, "Access denied to tenant", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// SecurityHeaders adds security headers to responses
func (m *EnhancedMiddleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		if m.config.IsProduction() {
			w.Header().Set("Content-Security-Policy", "default-src 'self'")
		}

		next.ServeHTTP(w, r)
	})
}

// AuditLog logs important actions for security auditing
func (m *EnhancedMiddleware) AuditLog(actions ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if this request should be audited
			shouldAudit := len(actions) == 0 // Audit all if no specific actions
			if !shouldAudit {
				for _, action := range actions {
					if strings.Contains(r.URL.Path, action) || r.Method == action {
						shouldAudit = true
						break
					}
				}
			}

			if shouldAudit {
				// Create audit log entry
				go m.createAuditLog(r)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper methods

// enhancedResponseWriter wraps http.ResponseWriter to capture additional metrics
type enhancedResponseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *enhancedResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *enhancedResponseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// writeErrorResponse writes a JSON error response
func (m *EnhancedMiddleware) writeErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	errorResponse := domain.ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
		Code:    statusCode,
	}
	
	json.NewEncoder(w).Encode(errorResponse)
}

// getClientIP extracts the real client IP address
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the list
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	if colon := strings.LastIndex(ip, ":"); colon != -1 {
		ip = ip[:colon]
	}
	return ip
}

// createAuditLog creates an audit log entry for the request
func (m *EnhancedMiddleware) createAuditLog(r *http.Request) {
	// Extract context values
	userID := r.Context().Value(UserIDKey)
	tenantID := r.Context().Value(TenantIDKey)
	requestID := r.Context().Value(RequestIDKey)

	// Create audit log entry
	auditEntry := map[string]interface{}{
		"timestamp":  time.Now().Format(time.RFC3339),
		"request_id": requestID,
		"user_id":    userID,
		"tenant_id":  tenantID,
		"action":     fmt.Sprintf("%s %s", r.Method, r.URL.Path),
		"ip_address": getClientIP(r),
		"user_agent": r.UserAgent(),
	}

	// Log audit entry (in production, this would go to a secure audit store)
	auditJSON, _ := json.Marshal(auditEntry)
	fmt.Printf("AUDIT: %s\n", auditJSON)
}

// Pagination middleware for handling pagination parameters
func (m *EnhancedMiddleware) Pagination(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse pagination parameters
		pageStr := r.URL.Query().Get("page")
		perPageStr := r.URL.Query().Get("per_page")

		page := 1
		perPage := 20 // default

		if pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		if perPageStr != "" {
			if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
				perPage = pp
			}
		}

		// Add pagination to context
		ctx := context.WithValue(r.Context(), "page", page)
		ctx = context.WithValue(ctx, "per_page", perPage)
		ctx = context.WithValue(ctx, "offset", (page-1)*perPage)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// generateRateLimitKey creates a unique key for rate limiting based on user and IP
func (m *EnhancedMiddleware) generateRateLimitKey(r *http.Request) string {
	// Priority order: User ID > API Key > IP Address
	if userID := r.Context().Value(UserIDKey); userID != nil {
		return fmt.Sprintf("user:%s", userID)
	}
	
	// Check for API key authentication
	if apiKey := r.Header.Get("X-API-Key"); apiKey != "" {
		// Use first 8 characters of API key as identifier
		if len(apiKey) >= 8 {
			return fmt.Sprintf("apikey:%s", apiKey[:8])
		}
	}
	
	// Fall back to IP address
	ip := getClientIP(r)
	return fmt.Sprintf("ip:%s", ip)
}

// HierarchicalRateLimit applies multiple rate limits (per second, minute, hour)
func (m *EnhancedMiddleware) HierarchicalRateLimit(secondLimit, minuteLimit, hourLimit int) func(http.Handler) http.Handler {
	hierarchical := security.NewHierarchicalRateLimiter()
	
	// Add different time window limiters
	hierarchical.AddLimiter("second", security.NewRedisRateLimiter(m.redisClient, secondLimit, time.Second))
	hierarchical.AddLimiter("minute", security.NewRedisRateLimiter(m.redisClient, minuteLimit, time.Minute))
	hierarchical.AddLimiter("hour", security.NewRedisRateLimiter(m.redisClient, hourLimit, time.Hour))
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := m.generateRateLimitKey(r)
			
			allowed, err := hierarchical.Allow(r.Context(), key)
			if err != nil {
				m.writeErrorResponse(w, "Rate limit check failed", http.StatusInternalServerError)
				return
			}
			
			if !allowed {
				m.writeErrorResponse(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}

// SlidingWindowRateLimit implements sliding window rate limiting for more precise control
func (m *EnhancedMiddleware) SlidingWindowRateLimit(limit int, window time.Duration) func(http.Handler) http.Handler {
	slidingLimiter := security.NewSlidingWindowRateLimiter(m.redisClient)
	
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := m.generateRateLimitKey(r)
			
			config := security.SlidingWindowConfig{
				Limit:  limit,
				Window: window,
			}
			
			allowed, info, err := slidingLimiter.AllowWithConfig(r.Context(), key, config)
			if err != nil {
				m.writeErrorResponse(w, "Rate limit check failed", http.StatusInternalServerError)
				return
			}
			
			// Add rate limit headers
			if info != nil {
				w.Header().Set("X-RateLimit-Limit", strconv.Itoa(info.Limit))
				w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(info.Remaining))
				w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(info.Reset.Unix(), 10))
				w.Header().Set("X-RateLimit-Window", info.Window.String())
			}
			
			if !allowed {
				m.writeErrorResponse(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
			
			next.ServeHTTP(w, r)
		})
	}
}