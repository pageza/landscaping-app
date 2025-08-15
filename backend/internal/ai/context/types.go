package context

import (
	"time"

	"github.com/google/uuid"
)

// AI types that are needed by the context package to avoid import cycles

// AssistantType represents the type of AI assistant
type AssistantType string

const (
	CustomerAssistant AssistantType = "customer"
	BusinessAssistant AssistantType = "business"
)

// MessageRole represents the role of a message in a conversation
type MessageRole string

const (
	RoleSystem    MessageRole = "system"
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleTool      MessageRole = "tool"
)

// ConversationContext represents the context of an AI conversation
type ConversationContext struct {
	ConversationID uuid.UUID              `json:"conversation_id"`
	TenantID       uuid.UUID              `json:"tenant_id"`
	UserID         *uuid.UUID             `json:"user_id,omitempty"`
	CustomerID     *uuid.UUID             `json:"customer_id,omitempty"`
	AssistantType  AssistantType          `json:"assistant_type"`
	SessionData    map[string]interface{} `json:"session_data"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
}

// Message represents a single message in a conversation
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

// ToolCall represents a function/tool call made by the assistant
type ToolCall struct {
	ID       string      `json:"id"`
	Type     string      `json:"type"`
	Function FunctionCall `json:"function"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FunctionCall represents a function call within a tool call
type FunctionCall struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResult represents the result of a tool call
type ToolResult struct {
	ToolCallID string      `json:"tool_call_id"`
	Result     interface{} `json:"result"`
	Error      string      `json:"error,omitempty"`
}

// ConversationSummary represents a summary of a conversation
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
	CustomerSentiment string                 `json:"customer_sentiment"`
	Summary           string                 `json:"summary"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// ConversationFilter represents filters for querying conversations
type ConversationFilter struct {
	TenantID      *uuid.UUID     `json:"tenant_id,omitempty"`
	UserID        *uuid.UUID     `json:"user_id,omitempty"`
	CustomerID    *uuid.UUID     `json:"customer_id,omitempty"`
	AssistantType *AssistantType `json:"assistant_type,omitempty"`
	StartDate     *time.Time     `json:"start_date,omitempty"`
	EndDate       *time.Time     `json:"end_date,omitempty"`
	Limit         int            `json:"limit,omitempty"`
	Offset        int            `json:"offset,omitempty"`
}

// AIMetrics represents AI usage metrics
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