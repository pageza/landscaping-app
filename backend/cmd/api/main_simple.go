package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

// Simple health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC(),
		"version":   "1.0.0",
		"service":   "landscaping-api",
		"message":   "Landscaping SaaS API is running successfully!",
	}
	
	json.NewEncoder(w).Encode(response)
}

// Simple API info endpoint
func apiInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	response := map[string]interface{}{
		"name":        "Landscaping SaaS API",
		"version":     "1.0.0",
		"description": "Enterprise Landscaping Management Platform",
		"endpoints": map[string]string{
			"health": "/health",
			"info":   "/api/info",
			"docs":   "/api/docs",
		},
		"status": "Running in development mode",
	}
	
	json.NewEncoder(w).Encode(response)
}

// Simple placeholder endpoint
func notImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	
	response := map[string]string{
		"error":   "Not Implemented",
		"message": "This endpoint is not yet implemented",
		"path":    r.URL.Path,
	}
	
	json.NewEncoder(w).Encode(response)
}

func main() {
	// Simple HTTP server setup
	mux := http.NewServeMux()
	
	// Health and info endpoints
	mux.HandleFunc("/health", healthCheck)
	mux.HandleFunc("/api/info", apiInfo)
	
	// Placeholder endpoints for the main API routes
	mux.HandleFunc("/api/", notImplemented)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/api/info", http.StatusFound)
			return
		}
		notImplemented(w, r)
	})
	
	// Start server
	port := "8080"
	if envPort := os.Getenv("PORT"); envPort != "" {
		port = envPort
	}
	
	log.Printf("🚀 Landscaping SaaS API starting on port %s", port)
	log.Printf("📋 Health check: http://localhost:%s/health", port)
	log.Printf("📖 API info: http://localhost:%s/api/info", port)
	
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}