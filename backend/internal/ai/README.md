# AI Assistant System

This package implements a comprehensive AI assistant system for the landscaping SaaS application, providing both customer-facing and business-facing AI capabilities.

## Features

- **Dual Assistant Types**:
  - Customer Assistant: Helps customers with scheduling, billing, quotes, and service inquiries
  - Business Assistant: Provides analytics, reporting, optimization, and business insights

- **Provider-Agnostic LLM Integration**: Works with OpenAI, Anthropic, Google Gemini
- **Function/Tool Calling**: Extensible system for AI to interact with business services
- **Conversation Management**: Persistent conversation history and context
- **Rate Limiting**: Configurable rate limits to control usage and costs
- **Real-time Chat**: WebSocket support for real-time AI interactions
- **Security**: Content moderation, access controls, and data protection

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   AI Handler    │    │  AI Assistant   │    │ LLM Integration │
│  (HTTP/WS API)  │────│   (Core Logic)  │────│  (OpenAI/etc)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       
         │              ┌─────────────────┐              
         │              │ Conversation    │              
         └──────────────│     Store       │              
                        │  (PostgreSQL)   │              
                        └─────────────────┘              
                                 │
                        ┌─────────────────┐
                        │   Rate Limiter  │
                        │    (Redis)      │
                        └─────────────────┘
```

## Usage

### 1. Setting Up the AI Assistant

```go
package main

import (
    "log/slog"
    
    "github.com/pageza/landscaping-app/backend/internal/ai"
    "github.com/pageza/landscaping-app/backend/internal/config"
    "github.com/pageza/landscaping-app/backend/internal/services"
)

func main() {
    // Initialize dependencies
    config := &config.Config{...}
    db := setupDatabase()
    redis := setupRedis()
    services := setupServices()
    logger := slog.Default()

    // Create AI service factory
    aiFactory := ai.NewAIServiceFactory(config, db, redis, services, logger)

    // Validate configuration
    if err := aiFactory.ValidateAIConfiguration(); err != nil {
        log.Fatal("AI configuration invalid:", err)
    }

    // Create AI assistant
    assistant, err := aiFactory.CreateAIAssistant()
    if err != nil {
        log.Fatal("Failed to create AI assistant:", err)
    }

    // Create AI handler for HTTP/WebSocket endpoints
    aiHandler := aiFactory.CreateAIHandler(assistant)

    // Setup API routes
    router := setupAPIRouter()
    router.SetAIHandler(aiHandler)
}
```

### 2. Customer Assistant Usage

```go
// Start a customer conversation
ctx := context.Background()
userID := uuid.New()
customerID := uuid.New()

conversation, err := assistant.StartConversation(
    ctx, 
    ai.CustomerAssistant, 
    &userID, 
    &customerID,
)

// Send a chat message
chatReq := &ai.ChatRequest{
    ConversationID: conversation.ConversationID,
    Message:        "I'd like to schedule a lawn mowing appointment for next Tuesday",
}

response, err := assistant.Chat(ctx, chatReq)
if err != nil {
    log.Error("Chat failed:", err)
    return
}

fmt.Printf("AI Response: %s\n", response.Content)
```

### 3. Business Assistant Usage

```go
// Start a business conversation
conversation, err := assistant.StartConversation(
    ctx, 
    ai.BusinessAssistant, 
    &userID, 
    nil,
)

// Request business metrics
chatReq := &ai.ChatRequest{
    ConversationID: conversation.ConversationID,
    Message:        "Show me revenue metrics for the last quarter",
}

response, err := assistant.Chat(ctx, chatReq)
// Response will include structured business data and insights
```

### 4. WebSocket Real-time Chat

```javascript
// Frontend WebSocket usage
const ws = new WebSocket('wss://api.landscaping.com/api/v1/ai/chat/ws');

ws.onopen = function(event) {
    console.log('Connected to AI assistant');
};

ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    if (message.type === 'chat') {
        displayAIResponse(message.payload);
    } else if (message.type === 'error') {
        showError(message.payload.error);
    }
};

// Send a chat message
function sendMessage(text) {
    const message = {
        type: 'chat',
        payload: {
            message: text,
            conversation_id: currentConversationId
        }
    };
    ws.send(JSON.stringify(message));
}
```

## Configuration

### Environment Variables

```bash
# LLM Provider Configuration
AI_CUSTOMER_MODEL=gpt-3.5-turbo
AI_BUSINESS_MODEL=gpt-4
AI_CUSTOMER_TEMPERATURE=0.7
AI_BUSINESS_TEMPERATURE=0.3

# Rate Limiting
AI_RATE_LIMIT_RPM=20
AI_RATE_LIMIT_RPH=100
AI_RATE_LIMIT_RPD=500
AI_COST_LIMIT_PER_DAY=50.00

# Security
AI_ENABLE_MODERATION=true
AI_LOG_CONVERSATIONS=true
AI_ENCRYPT_STORAGE=true
```

### JSON Configuration File

```json
{
  "customer_assistant": {
    "enabled": true,
    "model": "gpt-3.5-turbo",
    "temperature": 0.7,
    "max_tokens": 1000,
    "tools": [
      "schedule_appointment",
      "check_service_history",
      "request_quote",
      "check_billing"
    ],
    "capabilities": [
      "appointment_scheduling",
      "service_inquiries",
      "billing_support"
    ]
  },
  "business_assistant": {
    "enabled": true,
    "model": "gpt-4",
    "temperature": 0.3,
    "max_tokens": 2000,
    "tools": [
      "get_business_metrics",
      "analyze_revenue",
      "optimize_schedule"
    ],
    "permissions": [
      "business:view_metrics",
      "admin"
    ]
  },
  "rate_limit": {
    "enabled": true,
    "requests_per_minute": 20,
    "requests_per_day": 500,
    "cost_limit_per_day": 50.00
  },
  "security": {
    "enable_moderation": true,
    "content_filters": ["violence", "sexual", "hate"],
    "log_conversations": true,
    "encrypt_storage": true
  }
}
```

## API Endpoints

### Chat Endpoints

- `POST /api/v1/ai/chat` - Send a chat message
- `GET /api/v1/ai/chat/ws` - WebSocket endpoint for real-time chat

### Conversation Management

- `GET /api/v1/ai/conversations` - List conversations
- `POST /api/v1/ai/conversations` - Start new conversation
- `GET /api/v1/ai/conversations/{id}` - Get conversation details
- `DELETE /api/v1/ai/conversations/{id}` - End conversation
- `GET /api/v1/ai/conversations/{id}/messages` - Get conversation messages
- `GET /api/v1/ai/conversations/{id}/summary` - Get conversation summary

### Admin Endpoints

- `GET /api/v1/ai/admin/config` - Get AI configuration
- `PUT /api/v1/ai/admin/config` - Update AI configuration
- `GET /api/v1/ai/admin/metrics` - Get usage metrics
- `GET /api/v1/ai/admin/functions` - List available functions

### Usage Monitoring

- `GET /api/v1/ai/usage/stats` - Get usage statistics
- `GET /api/v1/ai/usage/limits` - Get rate limit configuration

## Available Tools

### Customer Assistant Tools

1. **schedule_appointment** - Schedule new landscaping services
2. **check_service_history** - View past service history
3. **request_quote** - Request quotes for new work
4. **check_billing** - Check invoices and billing status
5. **get_job_status** - Check current job status
6. **modify_appointment** - Reschedule or modify appointments
7. **add_special_instructions** - Add notes to jobs or properties
8. **get_property_info** - View property information
9. **get_service_catalog** - Browse available services

### Business Assistant Tools

1. **get_business_metrics** - Comprehensive business KPIs
2. **analyze_revenue** - Revenue analysis and trends
3. **analyze_customers** - Customer analytics and insights
4. **analyze_job_performance** - Job efficiency metrics
5. **optimize_schedule** - Schedule optimization recommendations
6. **get_overdue_invoices** - Overdue payment tracking
7. **check_crew_availability** - Crew scheduling information
8. **check_equipment_status** - Equipment availability and maintenance
9. **analyze_quote_conversion** - Quote-to-job conversion rates
10. **optimize_routes** - Route optimization for efficiency

## Database Schema

The AI assistant uses the following PostgreSQL tables:

- `ai_conversations` - Conversation contexts and metadata
- `ai_messages` - Individual messages in conversations
- `ai_conversation_summaries` - Generated conversation summaries
- `ai_usage_metrics` - Daily usage tracking for billing
- `ai_function_usage` - Function call analytics

## Rate Limiting

Rate limiting is implemented using Redis and supports:

- Requests per minute/hour/day limits
- Token consumption tracking
- Cost-based limiting
- Per-user and per-tenant limits
- Automatic cooldown periods
- Whitelisted users

## Security Features

- **Content Moderation**: Automatic filtering of inappropriate content
- **Access Controls**: Role-based permissions for different tools
- **Data Protection**: Encryption of stored conversations
- **Audit Logging**: Complete audit trail of AI interactions
- **Rate Limiting**: Protection against abuse and cost overruns

## Monitoring and Analytics

The system provides comprehensive monitoring:

- Usage statistics and trends
- Function call analytics
- Error tracking and alerting
- Performance metrics
- Cost tracking and optimization

## Extending the System

### Adding New Tools

```go
// Create a new tool
func createCustomTool() *ai.Function {
    return &ai.Function{
        Name:        "my_custom_tool",
        Description: "Description of what this tool does",
        Parameters: map[string]interface{}{
            "type": "object",
            "properties": map[string]interface{}{
                "param1": map[string]interface{}{
                    "type":        "string",
                    "description": "Parameter description",
                },
            },
            "required": []string{"param1"},
        },
        Handler: func(ctx context.Context, params map[string]interface{}) (interface{}, error) {
            // Tool implementation
            return map[string]interface{}{"result": "success"}, nil
        },
        Permissions: []string{"custom:permission"},
    }
}

// Register the tool
assistant.RegisterFunction("my_custom_tool", createCustomTool())
```

### Custom Conversation Store

```go
type CustomConversationStore struct {
    // Your implementation
}

func (c *CustomConversationStore) SaveConversation(ctx context.Context, conv *ai.ConversationContext) error {
    // Custom storage logic
    return nil
}

// Implement all ConversationStore interface methods...
```

## Performance Considerations

- Conversation history is paginated to limit memory usage
- LLM responses are cached when appropriate
- Database queries are optimized with proper indexing
- Rate limiting prevents resource exhaustion
- Old conversations are automatically cleaned up

## Cost Management

- Token usage tracking for accurate billing
- Configurable daily/monthly cost limits
- Model selection based on task complexity
- Conversation length limits to control costs
- Usage analytics for optimization opportunities

## Troubleshooting

### Common Issues

1. **AI not responding**: Check LLM provider API keys and rate limits
2. **Tool calls failing**: Verify function permissions and service availability
3. **High costs**: Review rate limits and model selection
4. **WebSocket issues**: Check network connectivity and authentication

### Monitoring

Monitor these key metrics:
- Response times and error rates
- Token usage and costs
- Conversation success rates
- Function call performance
- User satisfaction scores

For detailed troubleshooting, check the application logs and AI usage metrics in the admin dashboard.