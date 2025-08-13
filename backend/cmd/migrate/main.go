package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pageza/landscaping-app/backend/internal/config"
)

func main() {
	var (
		migrationsPath = flag.String("path", "backend/migrations", "Path to migrations directory")
	)
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Println("Usage: migrate [up|down|create] [name]")
		fmt.Println("Commands:")
		fmt.Println("  up     - Apply all pending migrations")
		fmt.Println("  down   - Rollback the last migration")
		fmt.Println("  create - Create a new migration file")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	command := os.Args[1]

	switch command {
	case "up":
		runMigrationsUp(cfg.DatabaseURL, *migrationsPath)
	case "down":
		runMigrationsDown(cfg.DatabaseURL, *migrationsPath)
	case "create":
		if len(os.Args) < 3 {
			log.Fatal("Migration name is required for create command")
		}
		createMigration(*migrationsPath, os.Args[2])
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}

func runMigrationsUp(databaseURL, migrationsPath string) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

	log.Println("Migrations applied successfully")
}

func runMigrationsDown(databaseURL, migrationsPath string) {
	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		databaseURL,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer m.Close()

	if err := m.Steps(-1); err != nil {
		log.Fatalf("Failed to rollback migration: %v", err)
	}

	log.Println("Migration rolled back successfully")
}

func createMigration(migrationsPath, name string) {
	// This is a simple implementation. In production, you might want to use
	// a more sophisticated migration file generator
	fmt.Printf("Creating migration files for: %s\n", name)
	fmt.Printf("Please create the files manually in: %s\n", migrationsPath)
	fmt.Printf("Format: NNNNNN_%s.up.sql and NNNNNN_%s.down.sql\n", name, name)
}