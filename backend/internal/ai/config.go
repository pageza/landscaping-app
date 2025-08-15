package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
)

// DefaultConfig returns a default AI assistant configuration
func DefaultConfig() *AssistantConfig {
	return &AssistantConfig{
		CustomerAssistant: CustomerAssistantConfig{
			Enabled:      true,
			Model:        "gpt-3.5-turbo",
			Temperature:  0.7,
			MaxTokens:    1000,
			SystemPrompt: getDefaultCustomerSystemPrompt(),
			Tools: []string{
				"schedule_appointment",
				"check_service_history",
				"request_quote",
				"check_billing",
				"get_job_status",
				"modify_appointment",
				"add_special_instructions",
				"get_upcoming_jobs",
				"get_property_info",
				"get_service_catalog",
			},
			Capabilities: []string{
				"appointment_scheduling",
				"service_inquiries",
				"billing_support",
				"quote_requests",
			},
			Restrictions: []string{
				"no_sensitive_data_sharing",
				"no_competitive_information",
				"customer_data_only",
			},
			SessionTTL:  4 * time.Hour,
			MaxMessages: 50,
			Context: map[string]interface{}{
				"business_name":    "Landscaping Pro",
				"support_hours":    "Monday-Friday 8AM-6PM",
				"emergency_contact": "+1-555-LANDSCAPE",
			},
		},
		BusinessAssistant: BusinessAssistantConfig{
			Enabled:      true,
			Model:        "gpt-4",
			Temperature:  0.3,
			MaxTokens:    2000,
			SystemPrompt: getDefaultBusinessSystemPrompt(),
			Tools: []string{
				"get_business_metrics",
				"analyze_revenue",
				"analyze_customers",
				"analyze_job_performance",
				"optimize_schedule",
				"get_overdue_invoices",
				"check_crew_availability",
				"check_equipment_status",
				"analyze_quote_conversion",
				"analyze_customer_retention",
				"optimize_routes",
				"analyze_profitability",
				"analyze_seasonal_trends",
				"analyze_operational_efficiency",
			},
			Capabilities: []string{
				"business_analytics",
				"performance_optimization",
				"financial_reporting",
				"operational_insights",
				"strategic_planning",
			},
			Permissions: []string{
				"business:view_metrics",
				"business:view_revenue",
				"business:view_customers",
				"business:view_performance",
				"business:manage_schedule",
				"admin",
			},
			SessionTTL:  8 * time.Hour,
			MaxMessages: 100,
			Context: map[string]interface{}{
				"business_type":     "landscaping_service",
				"reporting_period":  "monthly",
				"key_metrics":       []string{"revenue", "customer_satisfaction", "efficiency"},
			},
		},
		RateLimit: RateLimitConfig{
			Enabled:              true,
			RequestsPerMinute:    20,
			RequestsPerHour:      100,
			RequestsPerDay:       500,
			TokensPerMinute:      5000,
			TokensPerHour:        25000,
			TokensPerDay:         100000,
			CostLimitPerDay:      50.00,
			CooldownPeriod:       15 * time.Minute,
			WhitelistedUsers:     []uuid.UUID{},
		},
		Security: SecurityConfig{
			EnableModeration:     true,
			ContentFilters:       []string{"violence", "sexual", "hate", "harassment"},
			BlockedKeywords:      []string{"password", "ssn", "credit card"},
			RequiredPermissions:  []string{},
			AllowedDomains:       []string{},
			LogConversations:     true,
			RedactSensitiveData:  true,
			MaxConversationAge:   30 * 24 * time.Hour, // 30 days
			EncryptStorage:       true,
		},
	}
}

// LoadConfigFromFile loads AI configuration from a JSON file
func LoadConfigFromFile(filename string) (*AssistantConfig, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config AssistantConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// LoadConfigFromEnv loads AI configuration from environment variables
func LoadConfigFromEnv() *AssistantConfig {
	config := DefaultConfig()

	// Customer Assistant Configuration
	if model := os.Getenv("AI_CUSTOMER_MODEL"); model != "" {
		config.CustomerAssistant.Model = model
	}

	if tempStr := os.Getenv("AI_CUSTOMER_TEMPERATURE"); tempStr != "" {
		if temp, err := parseFloat(tempStr); err == nil {
			config.CustomerAssistant.Temperature = temp
		}
	}

	if tokensStr := os.Getenv("AI_CUSTOMER_MAX_TOKENS"); tokensStr != "" {
		if tokens, err := parseInt(tokensStr); err == nil {
			config.CustomerAssistant.MaxTokens = tokens
		}
	}

	if prompt := os.Getenv("AI_CUSTOMER_SYSTEM_PROMPT"); prompt != "" {
		config.CustomerAssistant.SystemPrompt = prompt
	}

	// Business Assistant Configuration
	if model := os.Getenv("AI_BUSINESS_MODEL"); model != "" {
		config.BusinessAssistant.Model = model
	}

	if tempStr := os.Getenv("AI_BUSINESS_TEMPERATURE"); tempStr != "" {
		if temp, err := parseFloat(tempStr); err == nil {
			config.BusinessAssistant.Temperature = temp
		}
	}

	if tokensStr := os.Getenv("AI_BUSINESS_MAX_TOKENS"); tokensStr != "" {
		if tokens, err := parseInt(tokensStr); err == nil {
			config.BusinessAssistant.MaxTokens = tokens
		}
	}

	if prompt := os.Getenv("AI_BUSINESS_SYSTEM_PROMPT"); prompt != "" {
		config.BusinessAssistant.SystemPrompt = prompt
	}

	// Rate Limiting Configuration
	if rpmStr := os.Getenv("AI_RATE_LIMIT_RPM"); rpmStr != "" {
		if rpm, err := parseInt(rpmStr); err == nil {
			config.RateLimit.RequestsPerMinute = rpm
		}
	}

	if rphStr := os.Getenv("AI_RATE_LIMIT_RPH"); rphStr != "" {
		if rph, err := parseInt(rphStr); err == nil {
			config.RateLimit.RequestsPerHour = rph
		}
	}

	if rpdStr := os.Getenv("AI_RATE_LIMIT_RPD"); rpdStr != "" {
		if rpd, err := parseInt(rpdStr); err == nil {
			config.RateLimit.RequestsPerDay = rpd
		}
	}

	if costLimitStr := os.Getenv("AI_COST_LIMIT_PER_DAY"); costLimitStr != "" {
		if costLimit, err := parseFloat(costLimitStr); err == nil {
			config.RateLimit.CostLimitPerDay = costLimit
		}
	}

	// Security Configuration
	if moderationStr := os.Getenv("AI_ENABLE_MODERATION"); moderationStr != "" {
		config.Security.EnableModeration = moderationStr == "true"
	}

	if loggingStr := os.Getenv("AI_LOG_CONVERSATIONS"); loggingStr != "" {
		config.Security.LogConversations = loggingStr == "true"
	}

	if encryptionStr := os.Getenv("AI_ENCRYPT_STORAGE"); encryptionStr != "" {
		config.Security.EncryptStorage = encryptionStr == "true"
	}

	return config
}

// SaveConfigToFile saves AI configuration to a JSON file
func SaveConfigToFile(config *AssistantConfig, filename string) error {
	// Validate configuration before saving
	if err := validateConfig(config); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// MergeConfigs merges two configurations, with the second one taking precedence
func MergeConfigs(base, override *AssistantConfig) *AssistantConfig {
	result := *base // Copy base config

	// Merge customer assistant config
	if override.CustomerAssistant.Model != "" {
		result.CustomerAssistant.Model = override.CustomerAssistant.Model
	}
	if override.CustomerAssistant.Temperature != 0 {
		result.CustomerAssistant.Temperature = override.CustomerAssistant.Temperature
	}
	if override.CustomerAssistant.MaxTokens != 0 {
		result.CustomerAssistant.MaxTokens = override.CustomerAssistant.MaxTokens
	}
	if override.CustomerAssistant.SystemPrompt != "" {
		result.CustomerAssistant.SystemPrompt = override.CustomerAssistant.SystemPrompt
	}
	if len(override.CustomerAssistant.Tools) > 0 {
		result.CustomerAssistant.Tools = override.CustomerAssistant.Tools
	}

	// Merge business assistant config
	if override.BusinessAssistant.Model != "" {
		result.BusinessAssistant.Model = override.BusinessAssistant.Model
	}
	if override.BusinessAssistant.Temperature != 0 {
		result.BusinessAssistant.Temperature = override.BusinessAssistant.Temperature
	}
	if override.BusinessAssistant.MaxTokens != 0 {
		result.BusinessAssistant.MaxTokens = override.BusinessAssistant.MaxTokens
	}
	if override.BusinessAssistant.SystemPrompt != "" {
		result.BusinessAssistant.SystemPrompt = override.BusinessAssistant.SystemPrompt
	}
	if len(override.BusinessAssistant.Tools) > 0 {
		result.BusinessAssistant.Tools = override.BusinessAssistant.Tools
	}

	// Merge rate limit config
	if override.RateLimit.RequestsPerMinute != 0 {
		result.RateLimit.RequestsPerMinute = override.RateLimit.RequestsPerMinute
	}
	if override.RateLimit.RequestsPerHour != 0 {
		result.RateLimit.RequestsPerHour = override.RateLimit.RequestsPerHour
	}
	if override.RateLimit.RequestsPerDay != 0 {
		result.RateLimit.RequestsPerDay = override.RateLimit.RequestsPerDay
	}
	if override.RateLimit.CostLimitPerDay != 0 {
		result.RateLimit.CostLimitPerDay = override.RateLimit.CostLimitPerDay
	}

	// Merge security config
	if len(override.Security.ContentFilters) > 0 {
		result.Security.ContentFilters = override.Security.ContentFilters
	}
	if len(override.Security.BlockedKeywords) > 0 {
		result.Security.BlockedKeywords = override.Security.BlockedKeywords
	}

	return &result
}

// ConfigurationManager manages AI assistant configurations
type ConfigurationManager struct {
	config       *AssistantConfig
	configFile   string
	lastModified time.Time
}

// NewConfigurationManager creates a new configuration manager
func NewConfigurationManager(configFile string) (*ConfigurationManager, error) {
	var config *AssistantConfig
	var err error

	// Try to load from file first
	if configFile != "" {
		if _, err := os.Stat(configFile); err == nil {
			config, err = LoadConfigFromFile(configFile)
			if err != nil {
				return nil, fmt.Errorf("failed to load config from file: %w", err)
			}
		}
	}

	// Fall back to environment variables or defaults
	if config == nil {
		config = LoadConfigFromEnv()
	}

	return &ConfigurationManager{
		config:       config,
		configFile:   configFile,
		lastModified: time.Now(),
	}, nil
}

// GetConfig returns the current configuration
func (cm *ConfigurationManager) GetConfig() *AssistantConfig {
	return cm.config
}

// UpdateConfig updates the configuration
func (cm *ConfigurationManager) UpdateConfig(newConfig *AssistantConfig) error {
	if err := validateConfig(newConfig); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	cm.config = newConfig
	cm.lastModified = time.Now()

	// Save to file if configured
	if cm.configFile != "" {
		if err := SaveConfigToFile(newConfig, cm.configFile); err != nil {
			return fmt.Errorf("failed to save config to file: %w", err)
		}
	}

	return nil
}

// ReloadConfig reloads configuration from file
func (cm *ConfigurationManager) ReloadConfig() error {
	if cm.configFile == "" {
		return fmt.Errorf("no config file specified")
	}

	config, err := LoadConfigFromFile(cm.configFile)
	if err != nil {
		return fmt.Errorf("failed to reload config: %w", err)
	}

	cm.config = config
	cm.lastModified = time.Now()

	return nil
}

// GetLastModified returns when the configuration was last modified
func (cm *ConfigurationManager) GetLastModified() time.Time {
	return cm.lastModified
}

// validateConfig validates the AI assistant configuration
func validateConfig(config *AssistantConfig) error {
	// Validate customer assistant config
	if config.CustomerAssistant.Enabled {
		if config.CustomerAssistant.Model == "" {
			return fmt.Errorf("customer assistant model cannot be empty")
		}
		if config.CustomerAssistant.Temperature < 0 || config.CustomerAssistant.Temperature > 2 {
			return fmt.Errorf("customer assistant temperature must be between 0 and 2")
		}
		if config.CustomerAssistant.MaxTokens <= 0 {
			return fmt.Errorf("customer assistant max tokens must be positive")
		}
		if config.CustomerAssistant.SystemPrompt == "" {
			return fmt.Errorf("customer assistant system prompt cannot be empty")
		}
	}

	// Validate business assistant config
	if config.BusinessAssistant.Enabled {
		if config.BusinessAssistant.Model == "" {
			return fmt.Errorf("business assistant model cannot be empty")
		}
		if config.BusinessAssistant.Temperature < 0 || config.BusinessAssistant.Temperature > 2 {
			return fmt.Errorf("business assistant temperature must be between 0 and 2")
		}
		if config.BusinessAssistant.MaxTokens <= 0 {
			return fmt.Errorf("business assistant max tokens must be positive")
		}
		if config.BusinessAssistant.SystemPrompt == "" {
			return fmt.Errorf("business assistant system prompt cannot be empty")
		}
	}

	// Validate rate limiting config
	if config.RateLimit.Enabled {
		if config.RateLimit.RequestsPerMinute < 0 {
			return fmt.Errorf("requests per minute cannot be negative")
		}
		if config.RateLimit.RequestsPerHour < 0 {
			return fmt.Errorf("requests per hour cannot be negative")
		}
		if config.RateLimit.RequestsPerDay < 0 {
			return fmt.Errorf("requests per day cannot be negative")
		}
		if config.RateLimit.CostLimitPerDay < 0 {
			return fmt.Errorf("cost limit per day cannot be negative")
		}
	}

	return nil
}

// getDefaultCustomerSystemPrompt returns the default system prompt for customer assistant
func getDefaultCustomerSystemPrompt() string {
	return `You are a helpful AI assistant for a professional landscaping company. Your primary role is to assist customers with:

1. Scheduling appointments and services
2. Checking service history and job status
3. Requesting quotes for new work
4. Billing and invoice inquiries
5. Property information and service recommendations

Guidelines:
- Always be professional, friendly, and helpful
- Provide accurate information based on the customer's account data
- If you cannot help with a request, politely explain and offer alternatives
- Protect customer privacy and never share sensitive information
- Focus on landscaping-related services and avoid unrelated topics
- When scheduling, confirm all details with the customer
- For billing issues, provide clear explanations and next steps

Remember: You represent the company's brand, so maintain high standards of customer service at all times.`
}

// getDefaultBusinessSystemPrompt returns the default system prompt for business assistant
func getDefaultBusinessSystemPrompt() string {
	return `You are an AI business intelligence assistant for a landscaping company. Your role is to help business owners and managers with:

1. Business performance analysis and reporting
2. Revenue and profitability insights
3. Customer analytics and retention strategies
4. Operational efficiency optimization
5. Schedule and resource management
6. Financial reporting and metrics

Capabilities:
- Analyze business data and identify trends
- Generate actionable insights and recommendations
- Create reports and visualizations
- Optimize operations for efficiency and profitability
- Monitor key performance indicators (KPIs)
- Provide strategic guidance based on data

Guidelines:
- Focus on data-driven insights and recommendations
- Present information clearly and concisely
- Highlight both opportunities and potential issues
- Provide context for metrics and comparisons
- Suggest specific actions when possible
- Maintain confidentiality of business information
- Use professional business language

Your goal is to help the business succeed through intelligent analysis and strategic recommendations.`
}

// Helper functions for parsing environment variables
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

func parseFloat(s string) (float64, error) {
	var result float64
	_, err := fmt.Sscanf(s, "%f", &result)
	return result, err
}