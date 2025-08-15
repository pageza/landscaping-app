package middleware

import (
	"net/http"
	"strings"

	"github.com/pageza/landscaping-app/web/internal/config"
	"github.com/pageza/landscaping-app/web/internal/services"
)

type Middleware struct {
	config   *config.Config
	services *services.Services
}

func NewMiddleware(cfg *config.Config, svc *services.Services) *Middleware {
	return &Middleware{
		config:   cfg,
		services: svc,
	}
}

// CORS middleware for HTMX requests
func (m *Middleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, HX-Request, HX-Target, HX-Current-URL")
		w.Header().Set("Access-Control-Expose-Headers", "HX-Redirect, HX-Refresh, HX-Trigger")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Security headers middleware
func (m *Middleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		
		// CSP header - allow HTMX inline scripts
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' unpkg.com cdn.jsdelivr.net; " +
			"style-src 'self' 'unsafe-inline' cdn.jsdelivr.net; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' fonts.googleapis.com fonts.gstatic.com; " +
			"connect-src 'self' ws: wss:; " +
			"frame-ancestors 'none';"
		w.Header().Set("Content-Security-Policy", csp)

		next.ServeHTTP(w, r)
	})
}

// Logger middleware
func (m *Middleware) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simple logging - in production, use structured logging
		println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// HTMX middleware to set appropriate headers and context
func (m *Middleware) HTMX(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add HTMX context to request
		if r.Header.Get("HX-Request") == "true" {
			// This is an HTMX request
			w.Header().Set("Vary", "HX-Request")
		}

		next.ServeHTTP(w, r)
	})
}

// Auth middleware for protected routes
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for session or JWT token
		token := m.extractToken(r)
		if token == "" {
			// Redirect to login for HTML requests, return 401 for HTMX
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Validate token (placeholder - integrate with backend auth service)
		valid := m.services.Auth.ValidateToken(token)
		if !valid {
			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", "/login")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// RequireRole middleware for role-based access
func (m *Middleware) RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user role from token/session
			userRole := m.getUserRole(r)
			
			allowed := false
			for _, role := range roles {
				if userRole == role {
					allowed = true
					break
				}
			}

			if !allowed {
				if r.Header.Get("HX-Request") == "true" {
					w.WriteHeader(http.StatusForbidden)
					w.Write([]byte("Access denied"))
					return
				}
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Helper function to extract token from request
func (m *Middleware) extractToken(r *http.Request) string {
	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	// Check session cookie
	cookie, err := r.Cookie("session_token")
	if err == nil {
		return cookie.Value
	}

	return ""
}

// Helper function to get user role
func (m *Middleware) getUserRole(r *http.Request) string {
	// Placeholder - extract from validated token/session
	token := m.extractToken(r)
	if token == "" {
		return ""
	}
	
	// In real implementation, decode JWT or lookup session
	return "admin" // Placeholder
}