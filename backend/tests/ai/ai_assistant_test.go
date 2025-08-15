package ai_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/pageza/landscaping-app/backend/internal/ai"
	"github.com/pageza/landscaping-app/backend/internal/services"
)

// Mock implementations for testing

type MockConversationStore struct {
	mock.Mock
}

func (m *MockConversationStore) SaveConversation(ctx context.Context, conversation *ai.ConversationContext) error {
	args := m.Called(ctx, conversation)
	return args.Error(0)
}

func (m *MockConversationStore) GetConversation(ctx context.Context, conversationID uuid.UUID) (*ai.ConversationContext, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).(*ai.ConversationContext), args.Error(1)
}

func (m *MockConversationStore) UpdateConversation(ctx context.Context, conversation *ai.ConversationContext) error {
	args := m.Called(ctx, conversation)
	return args.Error(0)
}

func (m *MockConversationStore) EndConversation(ctx context.Context, conversationID uuid.UUID) error {
	args := m.Called(ctx, conversationID)
	return args.Error(0)
}

func (m *MockConversationStore) DeleteConversation(ctx context.Context, conversationID uuid.UUID) error {
	args := m.Called(ctx, conversationID)
	return args.Error(0)
}

func (m *MockConversationStore) SaveMessage(ctx context.Context, message *ai.Message) error {
	args := m.Called(ctx, message)
	return args.Error(0)
}

func (m *MockConversationStore) GetMessages(ctx context.Context, conversationID uuid.UUID, limit, offset int) ([]*ai.Message, error) {
	args := m.Called(ctx, conversationID, limit, offset)
	return args.Get(0).([]*ai.Message), args.Error(1)
}

func (m *MockConversationStore) GetMessage(ctx context.Context, messageID uuid.UUID) (*ai.Message, error) {
	args := m.Called(ctx, messageID)
	return args.Get(0).(*ai.Message), args.Error(1)
}

func (m *MockConversationStore) DeleteMessage(ctx context.Context, messageID uuid.UUID) error {
	args := m.Called(ctx, messageID)
	return args.Error(0)
}

func (m *MockConversationStore) SaveConversationSummary(ctx context.Context, summary *ai.ConversationSummary) error {
	args := m.Called(ctx, summary)
	return args.Error(0)
}

func (m *MockConversationStore) GetConversationSummary(ctx context.Context, conversationID uuid.UUID) (*ai.ConversationSummary, error) {
	args := m.Called(ctx, conversationID)
	return args.Get(0).(*ai.ConversationSummary), args.Error(1)
}

func (m *MockConversationStore) GetConversations(ctx context.Context, filter *ai.ConversationFilter) ([]*ai.ConversationContext, error) {
	args := m.Called(ctx, filter)
	return args.Get(0).([]*ai.ConversationContext), args.Error(1)
}

func (m *MockConversationStore) GetMetrics(ctx context.Context, tenantID uuid.UUID, startDate, endDate time.Time) (*ai.AIMetrics, error) {
	args := m.Called(ctx, tenantID, startDate, endDate)
	return args.Get(0).(*ai.AIMetrics), args.Error(1)
}

func (m *MockConversationStore) CleanupExpiredConversations(ctx context.Context, maxAge time.Duration) error {
	args := m.Called(ctx, maxAge)
	return args.Error(0)
}

type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) CheckLimit(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	args := m.Called(ctx, tenantID, userID)
	return args.Error(0)
}

func (m *MockRateLimiter) RecordUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID, tokens int) error {
	args := m.Called(ctx, tenantID, userID, tokens)
	return args.Error(0)
}

func (m *MockRateLimiter) GetUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) (*ai.UsageStats, error) {
	args := m.Called(ctx, tenantID, userID)
	return args.Get(0).(*ai.UsageStats), args.Error(1)
}

func (m *MockRateLimiter) ResetUsage(ctx context.Context, tenantID uuid.UUID, userID *uuid.UUID) error {
	args := m.Called(ctx, tenantID, userID)
	return args.Error(0)
}

// Test functions

func TestAIAssistant_StartConversation(t *testing.T) {
	// Setup
	mockStore := &MockConversationStore{}
	mockRateLimiter := &MockRateLimiter{}
	
	config := ai.DefaultConfig()
	assistant := ai.NewAIAssistant(
		config,
		nil, // LLM client not needed for this test
		&services.Services{}, // Empty services for this test
		mockStore,
		mockRateLimiter,
		nil, // Logger not needed for this test
	)

	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()
	
	// Setup mocks
	mockStore.On("SaveConversation", ctx, mock.AnythingOfType("*ai.ConversationContext")).Return(nil)

	// Execute
	conversation, err := assistant.StartConversation(ctx, ai.CustomerAssistant, &userID, nil)

	// Assert
	require.NoError(t, err)
	assert.NotNil(t, conversation)
	assert.Equal(t, ai.CustomerAssistant, conversation.AssistantType)
	assert.Equal(t, userID, *conversation.UserID)
	assert.NotEqual(t, uuid.Nil, conversation.ConversationID)
	
	mockStore.AssertExpectations(t)
}

func TestAIAssistant_RegisterFunction(t *testing.T) {
	// Setup
	mockStore := &MockConversationStore{}
	mockRateLimiter := &MockRateLimiter{}
	
	config := ai.DefaultConfig()
	assistant := ai.NewAIAssistant(
		config,
		nil,
		&services.Services{},
		mockStore,
		mockRateLimiter,
		nil,
	)

	// Test function
	testFunction := &ai.Function{
		Name:        "test_function",
		Description: "A test function",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type":        "string",
					"description": "A test parameter",
				},
			},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"success": true}, nil
		},
	}

	// Execute
	err := assistant.RegisterFunction("test_function", testFunction)

	// Assert
	require.NoError(t, err)

	// Verify function is registered
	customerFunctions := assistant.GetAvailableFunctions(ai.CustomerAssistant)
	businessFunctions := assistant.GetAvailableFunctions(ai.BusinessAssistant)
	
	// Function should not appear in default configs since it's not in the tools list
	assert.Len(t, customerFunctions, len(config.CustomerAssistant.Tools))
	assert.Len(t, businessFunctions, len(config.BusinessAssistant.Tools))
}

func TestAIAssistant_GetConversation(t *testing.T) {
	// Setup
	mockStore := &MockConversationStore{}
	mockRateLimiter := &MockRateLimiter{}
	
	assistant := ai.NewAIAssistant(
		ai.DefaultConfig(),
		nil,
		&services.Services{},
		mockStore,
		mockRateLimiter,
		nil,
	)

	ctx := context.Background()
	conversationID := uuid.New()
	
	expectedConversation := &ai.ConversationContext{
		ConversationID: conversationID,
		TenantID:       uuid.New(),
		AssistantType:  ai.CustomerAssistant,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Setup mocks
	mockStore.On("GetConversation", ctx, conversationID).Return(expectedConversation, nil)

	// Execute
	conversation, err := assistant.GetConversation(ctx, conversationID)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, expectedConversation, conversation)
	
	mockStore.AssertExpectations(t)
}

func TestDefaultConfig(t *testing.T) {
	config := ai.DefaultConfig()

	// Test that default config is valid
	assert.True(t, config.CustomerAssistant.Enabled)
	assert.True(t, config.BusinessAssistant.Enabled)
	assert.NotEmpty(t, config.CustomerAssistant.Model)
	assert.NotEmpty(t, config.BusinessAssistant.Model)
	assert.NotEmpty(t, config.CustomerAssistant.SystemPrompt)
	assert.NotEmpty(t, config.BusinessAssistant.SystemPrompt)
	
	// Test customer assistant tools
	assert.Contains(t, config.CustomerAssistant.Tools, "schedule_appointment")
	assert.Contains(t, config.CustomerAssistant.Tools, "check_service_history")
	assert.Contains(t, config.CustomerAssistant.Tools, "request_quote")
	
	// Test business assistant tools
	assert.Contains(t, config.BusinessAssistant.Tools, "get_business_metrics")
	assert.Contains(t, config.BusinessAssistant.Tools, "analyze_revenue")
	assert.Contains(t, config.BusinessAssistant.Tools, "optimize_schedule")
	
	// Test rate limiting
	assert.True(t, config.RateLimit.Enabled)
	assert.Greater(t, config.RateLimit.RequestsPerMinute, 0)
	assert.Greater(t, config.RateLimit.RequestsPerDay, 0)
	
	// Test security
	assert.True(t, config.Security.EnableModeration)
	assert.NotEmpty(t, config.Security.ContentFilters)
	assert.True(t, config.Security.LogConversations)
}

func TestRateLimiter_CheckLimit(t *testing.T) {
	config := &ai.RateLimitConfig{
		Enabled:           true,
		RequestsPerMinute: 10,
		RequestsPerHour:   100,
		RequestsPerDay:    1000,
	}

	rateLimiter := ai.NewNoOpRateLimiter() // Use no-op for testing

	ctx := context.Background()
	tenantID := uuid.New()
	userID := uuid.New()

	// Test that no-op rate limiter allows all requests
	err := rateLimiter.CheckLimit(ctx, tenantID, &userID)
	assert.NoError(t, err)

	// Test recording usage
	err = rateLimiter.RecordUsage(ctx, tenantID, &userID, 100)
	assert.NoError(t, err)

	// Test getting usage stats
	stats, err := rateLimiter.GetUsage(ctx, tenantID, &userID)
	assert.NoError(t, err)
	assert.NotNil(t, stats)
}

func TestConversationTypes(t *testing.T) {
	// Test assistant types
	assert.Equal(t, "customer", string(ai.CustomerAssistant))
	assert.Equal(t, "business", string(ai.BusinessAssistant))

	// Test message roles
	assert.Equal(t, "user", string(ai.RoleUser))
	assert.Equal(t, "assistant", string(ai.RoleAssistant))
	assert.Equal(t, "system", string(ai.RoleSystem))
	assert.Equal(t, "tool", string(ai.RoleTool))

	// Test WebSocket message types
	assert.Equal(t, "chat", string(ai.MessageTypeChat))
	assert.Equal(t, "typing", string(ai.MessageTypeTyping))
	assert.Equal(t, "error", string(ai.MessageTypeError))
	assert.Equal(t, "connected", string(ai.MessageTypeConnected))
}

func TestFunctionDefinition(t *testing.T) {
	function := &ai.Function{
		Name:        "test_function",
		Description: "A test function for unit testing",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"param1": map[string]interface{}{
					"type":        "string",
					"description": "First parameter",
				},
				"param2": map[string]interface{}{
					"type":        "integer",
					"description": "Second parameter",
					"default":     42,
				},
			},
			"required": []string{"param1"},
		},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			param1 := params["param1"].(string)
			param2 := 42
			if p2, ok := params["param2"]; ok {
				param2 = int(p2.(float64))
			}
			
			return map[string]interface{}{
				"result": param1,
				"number": param2,
			}, nil
		},
		Permissions: []string{"test:function"},
	}

	// Test function execution
	ctx := context.Background()
	params := map[string]interface{}{
		"param1": "test_value",
		"param2": float64(123),
	}

	result, err := function.Handler(ctx, params)
	require.NoError(t, err)
	
	resultMap := result.(map[string]interface{})
	assert.Equal(t, "test_value", resultMap["result"])
	assert.Equal(t, 123, resultMap["number"])
}

// Benchmark tests

func BenchmarkStartConversation(b *testing.B) {
	mockStore := &MockConversationStore{}
	mockRateLimiter := &MockRateLimiter{}
	
	assistant := ai.NewAIAssistant(
		ai.DefaultConfig(),
		nil,
		&services.Services{},
		mockStore,
		mockRateLimiter,
		nil,
	)

	ctx := context.Background()
	userID := uuid.New()

	// Setup mocks for all benchmark iterations
	mockStore.On("SaveConversation", ctx, mock.AnythingOfType("*ai.ConversationContext")).Return(nil).Times(b.N)

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		_, err := assistant.StartConversation(ctx, ai.CustomerAssistant, &userID, nil)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkRegisterFunction(b *testing.B) {
	assistant := ai.NewAIAssistant(
		ai.DefaultConfig(),
		nil,
		&services.Services{},
		&MockConversationStore{},
		&MockRateLimiter{},
		nil,
	)

	testFunction := &ai.Function{
		Name:        "benchmark_function",
		Description: "A benchmark function",
		Parameters:  map[string]interface{}{},
		Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
			return nil, nil
		},
	}

	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		functionName := fmt.Sprintf("benchmark_function_%d", i)
		err := assistant.RegisterFunction(functionName, testFunction)
		if err != nil {
			b.Fatal(err)
		}
	}
}