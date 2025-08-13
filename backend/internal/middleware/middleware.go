package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/pageza/landscaping-app/backend/internal/config"
)

// Middleware holds all middleware functions
type Middleware struct {
	config *config.Config
}

// NewMiddleware creates a new middleware instance
func NewMiddleware(config *config.Config) *Middleware {
	return &Middleware{
		config: config,
	}
}

// CORS handles Cross-Origin Resource Sharing
func (m *Middleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		origin := r.Header.Get("Origin")
		for _, allowedOrigin := range m.config.CORSAllowedOrigins {
			if origin == allowedOrigin || allowedOrigin == "*" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Tenant-ID")
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

// Logging logs HTTP requests
func (m *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create a response writer wrapper to capture status code
		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start)
		log.Printf("%s %s %d %s %s",
			r.Method,
			r.RequestURI,
			wrapped.statusCode,
			duration,
			r.UserAgent(),
		)
	})
}

// RateLimit implements basic rate limiting
func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	// TODO: Implement proper rate limiting with Redis or in-memory store
	// For now, this is a placeholder
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

// RequireAuth validates JWT tokens and sets user context
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		// Extract token
		token := strings.TrimPrefix(authHeader, "Bearer ")
		if token == "" {
			http.Error(w, "Token required", http.StatusUnauthorized)
			return
		}

		// TODO: Validate JWT token and extract user information
		// For now, this is a placeholder
		userID := "placeholder-user-id"
		tenantID := "placeholder-tenant-id"

		// Set user context
		ctx := context.WithValue(r.Context(), "user_id", userID)
		ctx = context.WithValue(ctx, "tenant_id", tenantID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// TenantIsolation ensures proper tenant isolation
func (m *Middleware) TenantIsolation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract tenant ID from context (set by RequireAuth)
		tenantID := r.Context().Value("tenant_id")
		if tenantID == nil {
			http.Error(w, "Tenant context not found", http.StatusInternalServerError)
			return
		}

		// TODO: Add additional tenant validation logic if needed

		next.ServeHTTP(w, r)
	})
}

// RequireRole checks if user has required role
func (m *Middleware) RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// TODO: Extract user role from context and validate
			// For now, this is a placeholder
			userRole := "admin" // This would come from the JWT token

			if userRole != role && role != "" {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}