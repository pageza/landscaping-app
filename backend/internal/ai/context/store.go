package context

import (
	"context"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// ConversationStore interface defines methods for storing conversation context
type ConversationStore interface {
	// Conversation management
	SaveConversation(ctx context.Context, conversation *ConversationContext) error
	GetConversation(ctx context.Context, conversationID uuid.UUID) (*ConversationContext, error)
	UpdateConversation(ctx context.Context, conversation *ConversationContext) error
	EndConversation(ctx context.Context, conversationID uuid.UUID) error
	DeleteConversation(ctx context.Context, conversationID uuid.UUID) error
	
	// Message management
	SaveMessage(ctx context.Context, message *Message) error
	GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*Message, error)
	GetMessage(ctx context.Context, messageID uuid.UUID) (*Message, error)
	DeleteMessage(ctx context.Context, messageID uuid.UUID) error
	
	// Conversation summaries
	SaveConversationSummary(ctx context.Context, summary *ConversationSummary) error
	GetConversationSummary(ctx context.Context, conversationID uuid.UUID) (*ConversationSummary, error)
	
	// Queries and analytics
	GetConversations(ctx context.Context, filter *ConversationFilter) ([]*ConversationContext, error)
	GetMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*AIMetrics, error)
	
	// Cleanup
	CleanupExpiredConversations(ctx context.Context, maxAge time.Duration) error
}

// PostgreSQLStore implements ConversationStore using PostgreSQL
type PostgreSQLStore struct {
	db *sqlx.DB
}

// NewPostgreSQLStore creates a new PostgreSQL conversation store
func NewPostgreSQLStore(db *sqlx.DB) *PostgreSQLStore {
	return &PostgreSQLStore{db: db}
}

// Database models for conversation storage

type dbConversation struct {
	ID            uuid.UUID              `db:"id"`
	TenantID      uuid.UUID              `db:"tenant_id"`
	UserID        *uuid.UUID             `db:"user_id"`
	CustomerID    *uuid.UUID             `db:"customer_id"`
	AssistantType AssistantType          `db:"assistant_type"`
	SessionData   SessionDataJSON        `db:"session_data"`
	Status        string                 `db:"status"`
	CreatedAt     time.Time              `db:"created_at"`
	UpdatedAt     time.Time              `db:"updated_at"`
	EndedAt       *time.Time             `db:"ended_at"`
}

type dbMessage struct {
	ID             uuid.UUID              `db:"id"`
	ConversationID uuid.UUID              `db:"conversation_id"`
	Role           MessageRole            `db:"role"`
	Content        string                 `db:"content"`
	ToolCalls      ToolCallsJSON          `db:"tool_calls"`
	ToolResults    ToolResultsJSON        `db:"tool_results"`
	Metadata       MetadataJSON           `db:"metadata"`
	CreatedAt      time.Time              `db:"created_at"`
}

type dbConversationSummary struct {
	ID                uuid.UUID              `db:"id"`
	ConversationID    uuid.UUID              `db:"conversation_id"`
	TotalMessages     int                    `db:"total_messages"`
	TotalTokens       int                    `db:"total_tokens"`
	TotalCost         float64                `db:"total_cost"`
	StartTime         time.Time              `db:"start_time"`
	LastActivity      time.Time              `db:"last_activity"`
	KeyTopics         StringArrayJSON        `db:"key_topics"`
	ResolvedIssues    StringArrayJSON        `db:"resolved_issues"`
	PendingActions    StringArrayJSON        `db:"pending_actions"`
	CustomerSentiment string                 `db:"customer_sentiment"`
	Summary           string                 `db:"summary"`
	Metadata          MetadataJSON           `db:"metadata"`
	CreatedAt         time.Time              `db:"created_at"`
}

// JSON field types for PostgreSQL
type SessionDataJSON map[string]interface{}
type ToolCallsJSON []ToolCall
type ToolResultsJSON []ToolResult
type MetadataJSON map[string]interface{}
type StringArrayJSON []string

// Implement SQL driver interfaces
func (s SessionDataJSON) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *SessionDataJSON) Scan(value interface{}) error {
	if value == nil {
		*s = make(map[string]interface{})
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into SessionDataJSON", value)
	}
	
	return json.Unmarshal(bytes, s)
}

func (t ToolCallsJSON) Value() (driver.Value, error) {
	if len(t) == 0 {
		return "[]", nil
	}
	return json.Marshal(t)
}

func (t *ToolCallsJSON) Scan(value interface{}) error {
	if value == nil {
		*t = []ToolCall{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into ToolCallsJSON", value)
	}
	
	return json.Unmarshal(bytes, t)
}

func (t ToolResultsJSON) Value() (driver.Value, error) {
	if len(t) == 0 {
		return "[]", nil
	}
	return json.Marshal(t)
}

func (t *ToolResultsJSON) Scan(value interface{}) error {
	if value == nil {
		*t = []ToolResult{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into ToolResultsJSON", value)
	}
	
	return json.Unmarshal(bytes, t)
}

func (m MetadataJSON) Value() (driver.Value, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	return json.Marshal(m)
}

func (m *MetadataJSON) Scan(value interface{}) error {
	if value == nil {
		*m = make(map[string]interface{})
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into MetadataJSON", value)
	}
	
	return json.Unmarshal(bytes, m)
}

func (s StringArrayJSON) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

func (s *StringArrayJSON) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into StringArrayJSON", value)
	}
	
	return json.Unmarshal(bytes, s)
}

// ConversationStore implementation

func (s *PostgreSQLStore) SaveConversation(ctx context.Context, conversation *ConversationContext) error {
	query := `
		INSERT INTO ai_conversations (
			id, tenant_id, user_id, customer_id, assistant_type, 
			session_data, status, created_at, updated_at
		) VALUES (
			:id, :tenant_id, :user_id, :customer_id, :assistant_type,
			:session_data, :status, :created_at, :updated_at
		)`
	
	dbConv := &dbConversation{
		ID:            conversation.ConversationID,
		TenantID:      conversation.TenantID,
		UserID:        conversation.UserID,
		CustomerID:    conversation.CustomerID,
		AssistantType: conversation.AssistantType,
		SessionData:   SessionDataJSON(conversation.SessionData),
		Status:        "active",
		CreatedAt:     conversation.CreatedAt,
		UpdatedAt:     conversation.UpdatedAt,
	}
	
	_, err := s.db.NamedExecContext(ctx, query, dbConv)
	return err
}

func (s *PostgreSQLStore) GetConversation(ctx context.Context, conversationID uuid.UUID) (*ConversationContext, error) {
	query := `
		SELECT id, tenant_id, user_id, customer_id, assistant_type,
			   session_data, status, created_at, updated_at, ended_at
		FROM ai_conversations 
		WHERE id = $1`
	
	var dbConv dbConversation
	err := s.db.GetContext(ctx, &dbConv, query, conversationID)
	if err != nil {
		return nil, err
	}
	
	return &ConversationContext{
		ConversationID: dbConv.ID,
		TenantID:       dbConv.TenantID,
		UserID:         dbConv.UserID,
		CustomerID:     dbConv.CustomerID,
		AssistantType:  dbConv.AssistantType,
		SessionData:    map[string]interface{}(dbConv.SessionData),
		CreatedAt:      dbConv.CreatedAt,
		UpdatedAt:      dbConv.UpdatedAt,
	}, nil
}

func (s *PostgreSQLStore) UpdateConversation(ctx context.Context, conversation *ConversationContext) error {
	query := `
		UPDATE ai_conversations 
		SET session_data = :session_data, updated_at = :updated_at
		WHERE id = :id`
	
	dbConv := &dbConversation{
		ID:          conversation.ConversationID,
		SessionData: SessionDataJSON(conversation.SessionData),
		UpdatedAt:   time.Now(),
	}
	
	_, err := s.db.NamedExecContext(ctx, query, dbConv)
	return err
}

func (s *PostgreSQLStore) EndConversation(ctx context.Context, conversationID uuid.UUID) error {
	query := `
		UPDATE ai_conversations 
		SET status = 'ended', ended_at = $2, updated_at = $2
		WHERE id = $1`
	
	_, err := s.db.ExecContext(ctx, query, conversationID, time.Now())
	return err
}

func (s *PostgreSQLStore) DeleteConversation(ctx context.Context, conversationID uuid.UUID) error {
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Delete messages first (foreign key constraint)
	_, err = tx.ExecContext(ctx, "DELETE FROM ai_messages WHERE conversation_id = $1", conversationID)
	if err != nil {
		return err
	}
	
	// Delete conversation summary
	_, err = tx.ExecContext(ctx, "DELETE FROM ai_conversation_summaries WHERE conversation_id = $1", conversationID)
	if err != nil {
		return err
	}
	
	// Delete conversation
	_, err = tx.ExecContext(ctx, "DELETE FROM ai_conversations WHERE id = $1", conversationID)
	if err != nil {
		return err
	}
	
	return tx.Commit()
}

func (s *PostgreSQLStore) SaveMessage(ctx context.Context, message *Message) error {
	query := `
		INSERT INTO ai_messages (
			id, conversation_id, role, content, tool_calls, 
			tool_results, metadata, created_at
		) VALUES (
			:id, :conversation_id, :role, :content, :tool_calls,
			:tool_results, :metadata, :created_at
		)`
	
	dbMsg := &dbMessage{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Role:           message.Role,
		Content:        message.Content,
		ToolCalls:      ToolCallsJSON(message.ToolCalls),
		ToolResults:    ToolResultsJSON(message.ToolResults),
		Metadata:       MetadataJSON(message.Metadata),
		CreatedAt:      message.CreatedAt,
	}
	
	_, err := s.db.NamedExecContext(ctx, query, dbMsg)
	return err
}

func (s *PostgreSQLStore) GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*Message, error) {
	query := `
		SELECT id, conversation_id, role, content, tool_calls,
			   tool_results, metadata, created_at
		FROM ai_messages 
		WHERE conversation_id = $1 
		ORDER BY created_at ASC`
	
	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", offset)
	}
	
	var dbMessages []dbMessage
	err := s.db.SelectContext(ctx, &dbMessages, query, conversationID)
	if err != nil {
		return nil, err
	}
	
	messages := make([]*Message, len(dbMessages))
	for i, dbMsg := range dbMessages {
		messages[i] = &Message{
			ID:             dbMsg.ID,
			ConversationID: dbMsg.ConversationID,
			Role:           dbMsg.Role,
			Content:        dbMsg.Content,
			ToolCalls:      []ToolCall(dbMsg.ToolCalls),
			ToolResults:    []ToolResult(dbMsg.ToolResults),
			Metadata:       map[string]interface{}(dbMsg.Metadata),
			CreatedAt:      dbMsg.CreatedAt,
		}
	}
	
	return messages, nil
}

func (s *PostgreSQLStore) GetMessage(ctx context.Context, messageID uuid.UUID) (*Message, error) {
	query := `
		SELECT id, conversation_id, role, content, tool_calls,
			   tool_results, metadata, created_at
		FROM ai_messages 
		WHERE id = $1`
	
	var dbMsg dbMessage
	err := s.db.GetContext(ctx, &dbMsg, query, messageID)
	if err != nil {
		return nil, err
	}
	
	return &Message{
		ID:             dbMsg.ID,
		ConversationID: dbMsg.ConversationID,
		Role:           dbMsg.Role,
		Content:        dbMsg.Content,
		ToolCalls:      []ToolCall(dbMsg.ToolCalls),
		ToolResults:    []ToolResult(dbMsg.ToolResults),
		Metadata:       map[string]interface{}(dbMsg.Metadata),
		CreatedAt:      dbMsg.CreatedAt,
	}, nil
}

func (s *PostgreSQLStore) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	query := "DELETE FROM ai_messages WHERE id = $1"
	_, err := s.db.ExecContext(ctx, query, messageID)
	return err
}

func (s *PostgreSQLStore) SaveConversationSummary(ctx context.Context, summary *ConversationSummary) error {
	query := `
		INSERT INTO ai_conversation_summaries (
			id, conversation_id, total_messages, total_tokens, total_cost,
			start_time, last_activity, key_topics, resolved_issues,
			pending_actions, customer_sentiment, summary, metadata, created_at
		) VALUES (
			:id, :conversation_id, :total_messages, :total_tokens, :total_cost,
			:start_time, :last_activity, :key_topics, :resolved_issues,
			:pending_actions, :customer_sentiment, :summary, :metadata, :created_at
		) ON CONFLICT (conversation_id) DO UPDATE SET
			total_messages = EXCLUDED.total_messages,
			total_tokens = EXCLUDED.total_tokens,
			total_cost = EXCLUDED.total_cost,
			last_activity = EXCLUDED.last_activity,
			key_topics = EXCLUDED.key_topics,
			resolved_issues = EXCLUDED.resolved_issues,
			pending_actions = EXCLUDED.pending_actions,
			customer_sentiment = EXCLUDED.customer_sentiment,
			summary = EXCLUDED.summary,
			metadata = EXCLUDED.metadata`
	
	dbSummary := &dbConversationSummary{
		ID:                uuid.New(),
		ConversationID:    summary.ConversationID,
		TotalMessages:     summary.TotalMessages,
		TotalTokens:       summary.TotalTokens,
		TotalCost:         summary.TotalCost,
		StartTime:         summary.StartTime,
		LastActivity:      summary.LastActivity,
		KeyTopics:         StringArrayJSON(summary.KeyTopics),
		ResolvedIssues:    StringArrayJSON(summary.ResolvedIssues),
		PendingActions:    StringArrayJSON(summary.PendingActions),
		CustomerSentiment: summary.CustomerSentiment,
		Summary:           summary.Summary,
		Metadata:          MetadataJSON(summary.Metadata),
		CreatedAt:         time.Now(),
	}
	
	_, err := s.db.NamedExecContext(ctx, query, dbSummary)
	return err
}

func (s *PostgreSQLStore) GetConversationSummary(ctx context.Context, conversationID uuid.UUID) (*ConversationSummary, error) {
	query := `
		SELECT id, conversation_id, total_messages, total_tokens, total_cost,
			   start_time, last_activity, key_topics, resolved_issues,
			   pending_actions, customer_sentiment, summary, metadata, created_at
		FROM ai_conversation_summaries 
		WHERE conversation_id = $1`
	
	var dbSummary dbConversationSummary
	err := s.db.GetContext(ctx, &dbSummary, query, conversationID)
	if err != nil {
		return nil, err
	}
	
	return &ConversationSummary{
		ConversationID:    dbSummary.ConversationID,
		TotalMessages:     dbSummary.TotalMessages,
		TotalTokens:       dbSummary.TotalTokens,
		TotalCost:         dbSummary.TotalCost,
		StartTime:         dbSummary.StartTime,
		LastActivity:      dbSummary.LastActivity,
		KeyTopics:         []string(dbSummary.KeyTopics),
		ResolvedIssues:    []string(dbSummary.ResolvedIssues),
		PendingActions:    []string(dbSummary.PendingActions),
		CustomerSentiment: dbSummary.CustomerSentiment,
		Summary:           dbSummary.Summary,
		Metadata:          map[string]interface{}(dbSummary.Metadata),
	}, nil
}

func (s *PostgreSQLStore) GetConversations(ctx context.Context, filter *ConversationFilter) ([]*ConversationContext, error) {
	query := `
		SELECT id, tenant_id, user_id, customer_id, assistant_type,
			   session_data, status, created_at, updated_at, ended_at
		FROM ai_conversations 
		WHERE 1=1`
	
	args := []interface{}{}
	argIndex := 1
	
	if filter.TenantID != nil {
		query += fmt.Sprintf(" AND tenant_id = $%d", argIndex)
		args = append(args, *filter.TenantID)
		argIndex++
	}
	
	if filter.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *filter.UserID)
		argIndex++
	}
	
	if filter.CustomerID != nil {
		query += fmt.Sprintf(" AND customer_id = $%d", argIndex)
		args = append(args, *filter.CustomerID)
		argIndex++
	}
	
	if filter.AssistantType != nil {
		query += fmt.Sprintf(" AND assistant_type = $%d", argIndex)
		args = append(args, *filter.AssistantType)
		argIndex++
	}
	
	if filter.StartDate != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIndex)
		args = append(args, *filter.StartDate)
		argIndex++
	}
	
	if filter.EndDate != nil {
		query += fmt.Sprintf(" AND created_at <= $%d", argIndex)
		args = append(args, *filter.EndDate)
		argIndex++
	}
	
	query += " ORDER BY created_at DESC"
	
	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}
	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}
	
	var dbConversations []dbConversation
	err := s.db.SelectContext(ctx, &dbConversations, query, args...)
	if err != nil {
		return nil, err
	}
	
	conversations := make([]*ConversationContext, len(dbConversations))
	for i, dbConv := range dbConversations {
		conversations[i] = &ConversationContext{
			ConversationID: dbConv.ID,
			TenantID:       dbConv.TenantID,
			UserID:         dbConv.UserID,
			CustomerID:     dbConv.CustomerID,
			AssistantType:  dbConv.AssistantType,
			SessionData:    map[string]interface{}(dbConv.SessionData),
			CreatedAt:      dbConv.CreatedAt,
			UpdatedAt:      dbConv.UpdatedAt,
		}
	}
	
	return conversations, nil
}

func (s *PostgreSQLStore) GetMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*AIMetrics, error) {
	query := `
		SELECT 
			COUNT(DISTINCT c.id) as total_conversations,
			COUNT(m.id) as total_messages,
			COALESCE(SUM(s.total_tokens), 0) as total_tokens,
			COALESCE(SUM(s.total_cost), 0) as total_cost,
			COUNT(DISTINCT CASE WHEN c.assistant_type = 'customer' THEN c.id END) as customer_conversations,
			COUNT(DISTINCT CASE WHEN c.assistant_type = 'business' THEN c.id END) as business_conversations,
			COUNT(CASE WHEN jsonb_array_length(m.tool_calls) > 0 THEN 1 END) as tool_calls_count
		FROM ai_conversations c
		LEFT JOIN ai_messages m ON c.id = m.conversation_id
		LEFT JOIN ai_conversation_summaries s ON c.id = s.conversation_id
		WHERE c.tenant_id = $1 
		  AND c.created_at BETWEEN $2 AND $3`
	
	var metrics struct {
		TotalConversations    int     `db:"total_conversations"`
		TotalMessages         int     `db:"total_messages"`
		TotalTokens           int     `db:"total_tokens"`
		TotalCost             float64 `db:"total_cost"`
		CustomerConversations int     `db:"customer_conversations"`
		BusinessConversations int     `db:"business_conversations"`
		ToolCallsCount        int     `db:"tool_calls_count"`
	}
	
	err := s.db.GetContext(ctx, &metrics, query, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}
	
	return &AIMetrics{
		TenantID:              tenantID,
		Date:                  time.Now(),
		TotalConversations:    metrics.TotalConversations,
		TotalMessages:         metrics.TotalMessages,
		TotalTokens:           metrics.TotalTokens,
		TotalCost:             metrics.TotalCost,
		CustomerConversations: metrics.CustomerConversations,
		BusinessConversations: metrics.BusinessConversations,
		ToolCallsCount:        metrics.ToolCallsCount,
		ErrorCount:            0, // Would need additional tracking
		AvgResponseTime:       0, // Would need additional tracking
		AvgTokensPerMessage:   float64(metrics.TotalTokens) / float64(metrics.TotalMessages),
		TopTools:              []string{}, // Would need additional analysis
		TopTopics:             []string{}, // Would need additional analysis
	}, nil
}

func (s *PostgreSQLStore) CleanupExpiredConversations(ctx context.Context, maxAge time.Duration) error {
	cutoffTime := time.Now().Add(-maxAge)
	
	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Delete old messages
	_, err = tx.ExecContext(ctx, `
		DELETE FROM ai_messages 
		WHERE conversation_id IN (
			SELECT id FROM ai_conversations 
			WHERE updated_at < $1
		)`, cutoffTime)
	if err != nil {
		return err
	}
	
	// Delete old conversation summaries
	_, err = tx.ExecContext(ctx, `
		DELETE FROM ai_conversation_summaries 
		WHERE conversation_id IN (
			SELECT id FROM ai_conversations 
			WHERE updated_at < $1
		)`, cutoffTime)
	if err != nil {
		return err
	}
	
	// Delete old conversations
	_, err = tx.ExecContext(ctx, "DELETE FROM ai_conversations WHERE updated_at < $1", cutoffTime)
	if err != nil {
		return err
	}
	
	return tx.Commit()
}