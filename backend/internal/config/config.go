package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	// Environment
	Env string

	// Server
	APIHost string
	APIPort string

	// Database
	DatabaseURL             string
	DatabaseMaxConnections  int
	DatabaseMaxIdle         int
	DatabaseConnMaxLifetime time.Duration

	// Redis
	RedisURL      string
	RedisDB       int
	RedisPassword string

	// JWT
	JWTSecret        string
	JWTExpiry        time.Duration
	JWTRefreshExpiry time.Duration

	// Authentication
	BcryptCost int

	// Email
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFromEmail string
	SMTPFromName  string

	// Storage
	StorageProvider   string
	StorageBucket     string
	StorageRegion     string
	StorageAccessKey  string
	StorageSecretKey  string
	StorageEndpoint   string
	StoragePublicURL  string

	// Payments
	StripePublicKey    string
	StripeSecretKey    string
	StripeWebhookSecret string

	// LLM
	OpenAIAPIKey      string
	AnthropicAPIKey   string
	DefaultLLMProvider string

	// Logging
	LogLevel  string
	LogFormat string

	// Security
	CORSAllowedOrigins         []string
	RateLimitRequestsPerMinute int
	SessionSecret              string

	// Monitoring
	SentryDSN         string
	PrometheusEnabled bool
	PrometheusPort    string

	// Background Jobs
	QueueName       string
	QueueMaxRetries int
	QueueRetryDelay time.Duration
	WorkerConcurrency int

	// Multi-tenancy
	DefaultTenant         string
	TenantIsolationLevel  string

	// Feature Flags
	EnableRegistration     bool
	EnableEmailVerification bool
	EnablePasswordReset    bool
	EnableTwoFactorAuth    bool

	// Development
	DebugSQL              bool
	EnableProfiler        bool
	MockExternalServices  bool
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		// Environment
		Env: getEnv("ENV", "development"),

		// Server
		APIHost: getEnv("API_HOST", "0.0.0.0"),
		APIPort: getEnv("API_PORT", "8080"),

		// Database
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/landscaping_dev?sslmode=disable"),
		DatabaseMaxConnections:  getEnvAsInt("DATABASE_MAX_CONNECTIONS", 25),
		DatabaseMaxIdle:         getEnvAsInt("DATABASE_MAX_IDLE_CONNECTIONS", 5),
		DatabaseConnMaxLifetime: getEnvAsDuration("DATABASE_CONNECTION_MAX_LIFETIME", 5*time.Minute),

		// Redis
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),

		// JWT
		JWTSecret:        getEnv("JWT_SECRET", "your-super-secret-jwt-key-change-this-in-production"),
		JWTExpiry:        getEnvAsDuration("JWT_EXPIRY", 24*time.Hour),
		JWTRefreshExpiry: getEnvAsDuration("JWT_REFRESH_EXPIRY", 168*time.Hour),

		// Authentication
		BcryptCost: getEnvAsInt("BCRYPT_COST", 12),

		// Email
		SMTPHost:      getEnv("SMTP_HOST", "localhost"),
		SMTPPort:      getEnvAsInt("SMTP_PORT", 1025),
		SMTPUsername:  getEnv("SMTP_USERNAME", ""),
		SMTPPassword:  getEnv("SMTP_PASSWORD", ""),
		SMTPFromEmail: getEnv("SMTP_FROM_EMAIL", "noreply@landscaping-app.com"),
		SMTPFromName:  getEnv("SMTP_FROM_NAME", "Landscaping App"),

		// Storage
		StorageProvider:   getEnv("STORAGE_PROVIDER", "s3"),
		StorageBucket:     getEnv("STORAGE_BUCKET", "landscaping-app-dev"),
		StorageRegion:     getEnv("STORAGE_REGION", "us-east-1"),
		StorageAccessKey:  getEnv("STORAGE_ACCESS_KEY", ""),
		StorageSecretKey:  getEnv("STORAGE_SECRET_KEY", ""),
		StorageEndpoint:   getEnv("STORAGE_ENDPOINT", ""),
		StoragePublicURL:  getEnv("STORAGE_PUBLIC_URL", ""),

		// Payments
		StripePublicKey:     getEnv("STRIPE_PUBLIC_KEY", ""),
		StripeSecretKey:     getEnv("STRIPE_SECRET_KEY", ""),
		StripeWebhookSecret: getEnv("STRIPE_WEBHOOK_SECRET", ""),

		// LLM
		OpenAIAPIKey:       getEnv("OPENAI_API_KEY", ""),
		AnthropicAPIKey:    getEnv("ANTHROPIC_API_KEY", ""),
		DefaultLLMProvider: getEnv("DEFAULT_LLM_PROVIDER", "openai"),

		// Logging
		LogLevel:  getEnv("LOG_LEVEL", "info"),
		LogFormat: getEnv("LOG_FORMAT", "json"),

		// Security
		CORSAllowedOrigins:         getEnvAsSlice("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:8080"}),
		RateLimitRequestsPerMinute: getEnvAsInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 100),
		SessionSecret:              getEnv("SESSION_SECRET", "your-session-secret-key"),

		// Monitoring
		SentryDSN:         getEnv("SENTRY_DSN", ""),
		PrometheusEnabled: getEnvAsBool("PROMETHEUS_ENABLED", true),
		PrometheusPort:    getEnv("PROMETHEUS_PORT", "9090"),

		// Background Jobs
		QueueName:         getEnv("QUEUE_NAME", "default"),
		QueueMaxRetries:   getEnvAsInt("QUEUE_MAX_RETRIES", 3),
		QueueRetryDelay:   getEnvAsDuration("QUEUE_RETRY_DELAY", 30*time.Second),
		WorkerConcurrency: getEnvAsInt("WORKER_CONCURRENCY", 10),

		// Multi-tenancy
		DefaultTenant:        getEnv("DEFAULT_TENANT", "default"),
		TenantIsolationLevel: getEnv("TENANT_ISOLATION_LEVEL", "database"),

		// Feature Flags
		EnableRegistration:      getEnvAsBool("ENABLE_REGISTRATION", true),
		EnableEmailVerification: getEnvAsBool("ENABLE_EMAIL_VERIFICATION", true),
		EnablePasswordReset:     getEnvAsBool("ENABLE_PASSWORD_RESET", true),
		EnableTwoFactorAuth:     getEnvAsBool("ENABLE_TWO_FACTOR_AUTH", false),

		// Development
		DebugSQL:             getEnvAsBool("DEBUG_SQL", false),
		EnableProfiler:       getEnvAsBool("ENABLE_PROFILER", false),
		MockExternalServices: getEnvAsBool("MOCK_EXTERNAL_SERVICES", false),
	}

	return cfg, cfg.validate()
}

// validate checks if the configuration is valid
func (c *Config) validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.JWTSecret == "" || c.JWTSecret == "your-super-secret-jwt-key-change-this-in-production" {
		if c.Env == "production" {
			return fmt.Errorf("JWT_SECRET must be set in production")
		}
	}

	return nil
}

// IsProduction returns true if the environment is production
func (c *Config) IsProduction() bool {
	return c.Env == "production"
}

// IsDevelopment returns true if the environment is development
func (c *Config) IsDevelopment() bool {
	return c.Env == "development"
}

// IsTest returns true if the environment is test
func (c *Config) IsTest() bool {
	return c.Env == "test"
}

// Helper functions for environment variable parsing

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getEnvAsSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		// Simple comma-separated parsing
		// In production, you might want more sophisticated parsing
		result := []string{}
		for _, item := range splitAndTrim(value, ",") {
			if item != "" {
				result = append(result, item)
			}
		}
		if len(result) > 0 {
			return result
		}
	}
	return defaultValue
}

func splitAndTrim(s, sep string) []string {
	var result []string
	for _, item := range split(s, sep) {
		if trimmed := trim(item); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func split(s, sep string) []string {
	// Simple string split implementation
	if s == "" {
		return []string{}
	}
	
	var result []string
	var current string
	
	for i, char := range s {
		if string(char) == sep {
			result = append(result, current)
			current = ""
		} else {
			current += string(char)
		}
		
		if i == len(s)-1 {
			result = append(result, current)
		}
	}
	
	return result
}

func trim(s string) string {
	// Simple trim implementation for spaces
	start := 0
	end := len(s)
	
	for start < end && s[start] == ' ' {
		start++
	}
	
	for end > start && s[end-1] == ' ' {
		end--
	}
	
	return s[start:end]
}