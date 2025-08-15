package config

import (
	"os"

	backendConfig "github.com/pageza/landscaping-app/backend/internal/config"
)

// Config extends the backend config with web-specific settings
type Config struct {
	*backendConfig.Config
	WebPort        string
	StaticPath     string
	TemplatePath   string
	SessionSecret  string
	CSRFSecret     string
	TLSCertPath    string
	TLSKeyPath     string
	EnableTLS      bool
}

// LoadConfig loads configuration from environment variables and backend config
func LoadConfig() (*Config, error) {
	// Load backend config first
	backendCfg, err := backendConfig.Load()
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		Config:        backendCfg,
		WebPort:       getEnv("WEB_PORT", "8080"),
		StaticPath:    getEnv("STATIC_PATH", "./web/static"),
		TemplatePath:  getEnv("TEMPLATE_PATH", "./web/templates"),
		SessionSecret: getEnv("SESSION_SECRET", "your-secret-key-change-this"),
		CSRFSecret:    getEnv("CSRF_SECRET", "your-csrf-secret-change-this"),
		TLSCertPath:   getEnv("TLS_CERT_PATH", ""),
		TLSKeyPath:    getEnv("TLS_KEY_PATH", ""),
		EnableTLS:     getEnv("ENABLE_TLS", "false") == "true",
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}