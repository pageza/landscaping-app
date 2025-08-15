package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pageza/landscaping-app/backend/internal/config"
	"github.com/pageza/landscaping-app/web/internal/handlers"
	"github.com/pageza/landscaping-app/web/internal/middleware"
	"github.com/pageza/landscaping-app/web/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize services
	svc, err := services.NewServices(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize services: %v", err)
	}

	// Initialize middleware
	mw := middleware.NewMiddleware(cfg, svc)

	// Initialize handlers
	h := handlers.NewHandlers(cfg, svc, mw)

	// Setup routes
	router := h.SetupRoutes()

	// Create HTTP server
	srv := &http.Server{
		Addr:         ":" + cfg.WebPort,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Web server starting on port %s", cfg.WebPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}