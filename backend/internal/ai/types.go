package ai

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AssistantType defines the type of AI assistant
type AssistantType string

const (
	CustomerAssistant AssistantType = "customer"
	BusinessAssistant AssistantType = "business"
)

// ConversationContext holds the context for an AI conversation
type ConversationContext struct {
	ConversationID uuid.UUID     `json:"conversation_id"`
	TenantID       uuid.UUID     `json:"tenant_id"`
	UserID         *uuid.UUID    `json:"user_id,omitempty"`
	CustomerID     *uuid.UUID    `json:"customer_id,omitempty"`
	AssistantType  AssistantType `json:"assistant_type"`
	SessionData    map[string]interface{} `json:"session_data"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
}

// Message represents a message in a conversation
type Message struct {
	ID             uuid.UUID              `json:"id"`
	ConversationID uuid.UUID              `json:"conversation_id"`
	Role           MessageRole            `json:"role"`
	Content        string                 `json:"content"`
	ToolCalls      []ToolCall             `json:"tool_calls,omitempty"`
	ToolResults    []ToolResult           `json:"tool_results,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// MessageRole defines the role of a message
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
	RoleTool      MessageRole = "tool"
)

// ToolCall represents a function/tool call made by the AI
type ToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function FunctionCall           `json:"function"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FunctionCall represents a function call
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	ToolCallID string                 `json:"tool_call_id"`
	Success    bool                   `json:"success"`
	Result     interface{}            `json:"result,omitempty"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// ChatRequest represents a request to the AI assistant
type ChatRequest struct {
	ConversationID uuid.UUID              `json:"conversation_id,omitempty"`
	Message        string                 `json:"message"`
	Context        map[string]interface{} `json:"context,omitempty"`
	Model          string                 `json:"model,omitempty"`
	Temperature    *float64               `json:"temperature,omitempty"`
	MaxTokens      *int                   `json:"max_tokens,omitempty"`
}

// ChatResponse represents a response from the AI assistant
type ChatResponse struct {
	ConversationID uuid.UUID              `json:"conversation_id"`
	MessageID      uuid.UUID              `json:"message_id"`
	Content        string                 `json:"content"`
	ToolCalls      []ToolCall             `json:"tool_calls,omitempty"`
	Metadata       map[string]interface{} `json:"metadata"`
	Usage          *TokenUsage            `json:"usage,omitempty"`
}

// TokenUsage represents token usage information
type TokenUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// AssistantConfig holds configuration for AI assistants
type AssistantConfig struct {
	CustomerAssistant CustomerAssistantConfig `json:"customer_assistant"`
	BusinessAssistant BusinessAssistantConfig `json:"business_assistant"`
	RateLimit         RateLimitConfig         `json:"rate_limit"`
	Security          SecurityConfig          `json:"security"`
}

// CustomerAssistantConfig configures the customer-facing AI assistant
type CustomerAssistantConfig struct {
	Enabled       bool                   `json:"enabled"`
	Model         string                 `json:"model"`
	Temperature   float64                `json:"temperature"`
	MaxTokens     int                    `json:"max_tokens"`
	SystemPrompt  string                 `json:"system_prompt"`
	Tools         []string               `json:"tools"`
	Capabilities  []string               `json:"capabilities"`
	Restrictions  []string               `json:"restrictions"`
	SessionTTL    time.Duration          `json:"session_ttl"`
	MaxMessages   int                    `json:"max_messages"`
	Context       map[string]interface{} `json:"context"`
}

// BusinessAssistantConfig configures the business-facing AI assistant
type BusinessAssistantConfig struct {
	Enabled       bool                   `json:"enabled"`
	Model         string                 `json:"model"`
	Temperature   float64                `json:"temperature"`
	MaxTokens     int                    `json:"max_tokens"`
	SystemPrompt  string                 `json:"system_prompt"`
	Tools         []string               `json:"tools"`
	Capabilities  []string               `json:"capabilities"`
	Permissions   []string               `json:"permissions"`
	SessionTTL    time.Duration          `json:"session_ttl"`
	MaxMessages   int                    `json:"max_messages"`
	Context       map[string]interface{} `json:"context"`
}

// RateLimitConfig configures rate limiting for AI assistants
type RateLimitConfig struct {
	Enabled              bool          `json:"enabled"`
	RequestsPerMinute    int           `json:"requests_per_minute"`
	RequestsPerHour      int           `json:"requests_per_hour"`
	RequestsPerDay       int           `json:"requests_per_day"`
	TokensPerMinute      int           `json:"tokens_per_minute"`
	TokensPerHour        int           `json:"tokens_per_hour"`
	TokensPerDay         int           `json:"tokens_per_day"`
	CostLimitPerDay      float64       `json:"cost_limit_per_day"`
	CooldownPeriod       time.Duration `json:"cooldown_period"`
	WhitelistedUsers     []uuid.UUID   `json:"whitelisted_users"`
}

// SecurityConfig configures security settings for AI assistants
type SecurityConfig struct {
	EnableModeration     bool     `json:"enable_moderation"`
	ContentFilters       []string `json:"content_filters"`
	BlockedKeywords      []string `json:"blocked_keywords"`
	RequiredPermissions  []string `json:"required_permissions"`
	AllowedDomains       []string `json:"allowed_domains"`
	LogConversations     bool     `json:"log_conversations"`
	RedactSensitiveData  bool     `json:"redact_sensitive_data"`
	MaxConversationAge   time.Duration `json:"max_conversation_age"`
	EncryptStorage       bool     `json:"encrypt_storage"`
}

// Function represents a tool/function available to the AI assistant
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
	Handler     FunctionHandler        `json:"-"`
	Permissions []string               `json:"permissions,omitempty"`
	RateLimit   *FunctionRateLimit     `json:"rate_limit,omitempty"`
}

// FunctionHandler is the signature for function handlers
type FunctionHandler func(ctx context.Context, params map[string]interface{}) (interface{}, error)

// FunctionRateLimit configures rate limiting for specific functions
type FunctionRateLimit struct {
	CallsPerMinute int           `json:"calls_per_minute"`
	CallsPerHour   int           `json:"calls_per_hour"`
	CooldownPeriod time.Duration `json:"cooldown_period"`
}

// ConversationSummary represents a summary of conversation history
type ConversationSummary struct {
	ConversationID    uuid.UUID              `json:"conversation_id"`
	TotalMessages     int                    `json:"total_messages"`
	TotalTokens       int                    `json:"total_tokens"`
	TotalCost         float64                `json:"total_cost"`
	StartTime         time.Time              `json:"start_time"`
	LastActivity      time.Time              `json:"last_activity"`
	KeyTopics         []string               `json:"key_topics"`
	ResolvedIssues    []string               `json:"resolved_issues"`
	PendingActions    []string               `json:"pending_actions"`
	CustomerSentiment string                 `json:"customer_sentiment,omitempty"`
	Summary           string                 `json:"summary"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// AIMetrics represents metrics for AI assistant usage
type AIMetrics struct {
	TenantID              uuid.UUID `json:"tenant_id"`
	Date                  time.Time `json:"date"`
	TotalConversations    int       `json:"total_conversations"`
	TotalMessages         int       `json:"total_messages"`
	TotalTokens           int       `json:"total_tokens"`
	TotalCost             float64   `json:"total_cost"`
	CustomerConversations int       `json:"customer_conversations"`
	BusinessConversations int       `json:"business_conversations"`
	ToolCallsCount        int       `json:"tool_calls_count"`
	ErrorCount            int       `json:"error_count"`
	AvgResponseTime       float64   `json:"avg_response_time"`
	AvgTokensPerMessage   float64   `json:"avg_tokens_per_message"`
	TopTools              []string  `json:"top_tools"`
	TopTopics             []string  `json:"top_topics"`
}

// ConversationFilter represents filters for conversation queries
type ConversationFilter struct {
	TenantID      *uuid.UUID     `json:"tenant_id"`
	UserID        *uuid.UUID     `json:"user_id"`
	CustomerID    *uuid.UUID     `json:"customer_id"`
	AssistantType *AssistantType `json:"assistant_type"`
	StartDate     *time.Time     `json:"start_date"`
	EndDate       *time.Time     `json:"end_date"`
	Keywords      []string       `json:"keywords"`
	HasToolCalls  *bool          `json:"has_tool_calls"`
	MinMessages   *int           `json:"min_messages"`
	MaxMessages   *int           `json:"max_messages"`
	Limit         int            `json:"limit"`
	Offset        int            `json:"offset"`
}

// WebSocketMessage represents a WebSocket message for real-time chat
type WebSocketMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// WebSocketMessageType defines types of WebSocket messages
type WebSocketMessageType string

const (
	MessageTypeChat              WebSocketMessageType = "chat"
	MessageTypeTyping            WebSocketMessageType = "typing"
	MessageTypeError             WebSocketMessageType = "error"
	MessageTypeConnected         WebSocketMessageType = "connected"
	MessageTypeDisconnected      WebSocketMessageType = "disconnected"
	MessageTypeConversationStart WebSocketMessageType = "conversation_start"
	MessageTypeConversationEnd   WebSocketMessageType = "conversation_end"
	MessageTypeToolCall          WebSocketMessageType = "tool_call"
	MessageTypeToolResult        WebSocketMessageType = "tool_result"
)