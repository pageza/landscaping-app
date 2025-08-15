package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/pageza/landscaping-app/backend/internal/config"
)

// Connection holds database and Redis connections
type Connection struct {
	DB          *sql.DB
	RedisClient *redis.Client
}

// NewConnection creates new database and Redis connections
func NewConnection(cfg *config.Config) (*Connection, error) {
	// Create PostgreSQL connection
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.DatabaseMaxConnections)
	db.SetMaxIdleConns(cfg.DatabaseMaxIdle)
	db.SetConnMaxLifetime(cfg.DatabaseConnMaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Create Redis connection
	redisOpts := &redis.Options{
		Addr:     cfg.RedisURL,
		DB:       cfg.RedisDB,
		Password: cfg.RedisPassword,
	}

	redisClient := redis.NewClient(redisOpts)

	// Test Redis connection
	_, err = redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &Connection{
		DB:          db,
		RedisClient: redisClient,
	}, nil
}

// Close closes database and Redis connections
func (c *Connection) Close() error {
	var err error
	
	if c.DB != nil {
		if dbErr := c.DB.Close(); dbErr != nil {
			err = fmt.Errorf("failed to close database: %w", dbErr)
		}
	}
	
	if c.RedisClient != nil {
		if redisErr := c.RedisClient.Close(); redisErr != nil {
			if err != nil {
				err = fmt.Errorf("%v; failed to close Redis: %w", err, redisErr)
			} else {
				err = fmt.Errorf("failed to close Redis: %w", redisErr)
			}
		}
	}
	
	return err
}

// SetTenantContext sets the tenant context for Row Level Security
func (c *Connection) SetTenantContext(ctx context.Context, tenantID, userID uuid.UUID, userRole string) error {
	query := "SELECT set_tenant_context($1, $2, $3)"
	_, err := c.DB.ExecContext(ctx, query, tenantID, userID, userRole)
	if err != nil {
		return fmt.Errorf("failed to set tenant context: %w", err)
	}
	return nil
}

// ClearTenantContext clears the tenant context
func (c *Connection) ClearTenantContext(ctx context.Context) error {
	query := "SELECT clear_tenant_context()"
	_, err := c.DB.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to clear tenant context: %w", err)
	}
	return nil
}

// BeginTx starts a database transaction with tenant context
func (c *Connection) BeginTx(ctx context.Context, tenantID, userID uuid.UUID, userRole string) (*sql.Tx, error) {
	tx, err := c.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Set tenant context within the transaction
	query := "SELECT set_tenant_context($1, $2, $3)"
	_, err = tx.ExecContext(ctx, query, tenantID, userID, userRole)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to set tenant context in transaction: %w", err)
	}

	return tx, nil
}

// TenantAwareQuery executes a query with tenant context
func (c *Connection) TenantAwareQuery(ctx context.Context, tenantID, userID uuid.UUID, userRole, query string, args ...interface{}) (*sql.Rows, error) {
	// Set tenant context
	if err := c.SetTenantContext(ctx, tenantID, userID, userRole); err != nil {
		return nil, err
	}
	defer c.ClearTenantContext(ctx)

	// Execute query
	return c.DB.QueryContext(ctx, query, args...)
}

// TenantAwareExec executes a command with tenant context
func (c *Connection) TenantAwareExec(ctx context.Context, tenantID, userID uuid.UUID, userRole, query string, args ...interface{}) (sql.Result, error) {
	// Set tenant context
	if err := c.SetTenantContext(ctx, tenantID, userID, userRole); err != nil {
		return nil, err
	}
	defer c.ClearTenantContext(ctx)

	// Execute command
	return c.DB.ExecContext(ctx, query, args...)
}

// TenantAwareQueryRow executes a query returning at most one row with tenant context
func (c *Connection) TenantAwareQueryRow(ctx context.Context, tenantID, userID uuid.UUID, userRole, query string, args ...interface{}) *sql.Row {
	// Set tenant context (ignore error for single row query)
	c.SetTenantContext(ctx, tenantID, userID, userRole)
	defer c.ClearTenantContext(ctx)

	// Execute query
	return c.DB.QueryRowContext(ctx, query, args...)
}

// HealthCheck performs a health check on both database and Redis
func (c *Connection) HealthCheck(ctx context.Context) error {
	// Check database
	if err := c.DB.PingContext(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	// Check Redis
	if _, err := c.RedisClient.Ping(ctx).Result(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// GetStats returns connection statistics
func (c *Connection) GetStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	if c.DB != nil {
		dbStats := c.DB.Stats()
		stats["database"] = map[string]interface{}{
			"max_open_connections": dbStats.MaxOpenConnections,
			"open_connections":     dbStats.OpenConnections,
			"in_use":               dbStats.InUse,
			"idle":                 dbStats.Idle,
			"wait_count":           dbStats.WaitCount,
			"wait_duration":        dbStats.WaitDuration.String(),
			"max_idle_closed":      dbStats.MaxIdleClosed,
			"max_idle_time_closed": dbStats.MaxIdleTimeClosed,
			"max_lifetime_closed":  dbStats.MaxLifetimeClosed,
		}
	}

	if c.RedisClient != nil {
		poolStats := c.RedisClient.PoolStats()
		stats["redis"] = map[string]interface{}{
			"hits":         poolStats.Hits,
			"misses":       poolStats.Misses,
			"timeouts":     poolStats.Timeouts,
			"total_conns":  poolStats.TotalConns,
			"idle_conns":   poolStats.IdleConns,
			"stale_conns":  poolStats.StaleConns,
		}
	}

	return stats
}