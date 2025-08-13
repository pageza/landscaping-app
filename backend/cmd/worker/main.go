package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pageza/landscaping-app/backend/internal/config"
	"github.com/pageza/landscaping-app/backend/internal/repository"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database connection
	db, err := repository.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	svc := services.NewServices(repos, cfg)

	// Initialize worker service
	workerService := services.NewWorkerService(svc, cfg)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start worker in a goroutine
	go func() {
		log.Printf("Starting worker with concurrency %d", cfg.WorkerConcurrency)
		if err := workerService.Start(ctx); err != nil {
			log.Fatalf("Worker failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the worker
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down worker...")

	// Cancel context to stop worker
	cancel()

	// Give worker time to finish current jobs
	time.Sleep(10 * time.Second)

	log.Println("Worker exited")
}