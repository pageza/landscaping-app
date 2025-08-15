-- AI Assistant Schema Migration
-- This migration adds tables for AI assistant functionality

-- AI conversations table
CREATE TABLE IF NOT EXISTS ai_conversations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    customer_id UUID REFERENCES customers(id) ON DELETE SET NULL,
    assistant_type VARCHAR(20) NOT NULL CHECK (assistant_type IN ('customer', 'business')),
    session_data JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'ended', 'archived')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE
);

-- AI messages table
CREATE TABLE IF NOT EXISTS ai_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    role VARCHAR(20) NOT NULL CHECK (role IN ('user', 'assistant', 'system', 'tool')),
    content TEXT NOT NULL,
    tool_calls JSONB DEFAULT '[]',
    tool_results JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- AI conversation summaries table
CREATE TABLE IF NOT EXISTS ai_conversation_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    conversation_id UUID NOT NULL REFERENCES ai_conversations(id) ON DELETE CASCADE,
    total_messages INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    total_cost DECIMAL(10,4) DEFAULT 0.0000,
    start_time TIMESTAMP WITH TIME ZONE,
    last_activity TIMESTAMP WITH TIME ZONE,
    key_topics JSONB DEFAULT '[]',
    resolved_issues JSONB DEFAULT '[]',
    pending_actions JSONB DEFAULT '[]',
    customer_sentiment VARCHAR(20),
    summary TEXT,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(conversation_id)
);

-- AI usage metrics table (for tracking and billing)
CREATE TABLE IF NOT EXISTS ai_usage_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    date DATE NOT NULL,
    assistant_type VARCHAR(20) NOT NULL,
    total_conversations INTEGER DEFAULT 0,
    total_messages INTEGER DEFAULT 0,
    total_tokens INTEGER DEFAULT 0,
    total_cost DECIMAL(10,4) DEFAULT 0.0000,
    tool_calls_count INTEGER DEFAULT 0,
    error_count INTEGER DEFAULT 0,
    avg_response_time_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, user_id, date, assistant_type)
);

-- AI function usage tracking
CREATE TABLE IF NOT EXISTS ai_function_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    conversation_id UUID REFERENCES ai_conversations(id) ON DELETE SET NULL,
    function_name VARCHAR(100) NOT NULL,
    execution_time_ms INTEGER,
    success BOOLEAN DEFAULT true,
    error_message TEXT,
    parameters JSONB DEFAULT '{}',
    result JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_ai_conversations_tenant_id ON ai_conversations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_user_id ON ai_conversations(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_customer_id ON ai_conversations(customer_id);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_assistant_type ON ai_conversations(assistant_type);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_status ON ai_conversations(status);
CREATE INDEX IF NOT EXISTS idx_ai_conversations_created_at ON ai_conversations(created_at);

CREATE INDEX IF NOT EXISTS idx_ai_messages_conversation_id ON ai_messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_ai_messages_role ON ai_messages(role);
CREATE INDEX IF NOT EXISTS idx_ai_messages_created_at ON ai_messages(created_at);

CREATE INDEX IF NOT EXISTS idx_ai_usage_metrics_tenant_id ON ai_usage_metrics(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_metrics_user_id ON ai_usage_metrics(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_usage_metrics_date ON ai_usage_metrics(date);
CREATE INDEX IF NOT EXISTS idx_ai_usage_metrics_assistant_type ON ai_usage_metrics(assistant_type);

CREATE INDEX IF NOT EXISTS idx_ai_function_usage_tenant_id ON ai_function_usage(tenant_id);
CREATE INDEX IF NOT EXISTS idx_ai_function_usage_user_id ON ai_function_usage(user_id);
CREATE INDEX IF NOT EXISTS idx_ai_function_usage_conversation_id ON ai_function_usage(conversation_id);
CREATE INDEX IF NOT EXISTS idx_ai_function_usage_function_name ON ai_function_usage(function_name);
CREATE INDEX IF NOT EXISTS idx_ai_function_usage_created_at ON ai_function_usage(created_at);

-- Updated at trigger for ai_conversations
CREATE OR REPLACE FUNCTION update_ai_conversations_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ai_conversations_updated_at
    BEFORE UPDATE ON ai_conversations
    FOR EACH ROW EXECUTE FUNCTION update_ai_conversations_updated_at();

-- Updated at trigger for ai_usage_metrics
CREATE OR REPLACE FUNCTION update_ai_usage_metrics_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_ai_usage_metrics_updated_at
    BEFORE UPDATE ON ai_usage_metrics
    FOR EACH ROW EXECUTE FUNCTION update_ai_usage_metrics_updated_at();

-- Function to clean up old AI conversations
CREATE OR REPLACE FUNCTION cleanup_old_ai_conversations(max_age_days INTEGER DEFAULT 90)
RETURNS INTEGER AS $$
DECLARE
    deleted_count INTEGER;
BEGIN
    -- Delete old conversations and related data
    WITH deleted_conversations AS (
        DELETE FROM ai_conversations 
        WHERE updated_at < CURRENT_TIMESTAMP - INTERVAL '%s days' 
        AND status = 'ended'
        RETURNING id
    )
    SELECT COUNT(*) INTO deleted_count FROM deleted_conversations;
    
    RETURN deleted_count;
END;
$$ LANGUAGE plpgsql;

-- Function to aggregate daily AI usage metrics
CREATE OR REPLACE FUNCTION aggregate_ai_usage_metrics(target_date DATE DEFAULT CURRENT_DATE - INTERVAL '1 day')
RETURNS VOID AS $$
BEGIN
    -- Aggregate conversation and message counts by tenant, user, and assistant type
    INSERT INTO ai_usage_metrics (
        tenant_id, user_id, date, assistant_type,
        total_conversations, total_messages, total_tokens, total_cost
    )
    SELECT 
        c.tenant_id,
        c.user_id,
        target_date,
        c.assistant_type,
        COUNT(DISTINCT c.id) as total_conversations,
        COUNT(m.id) as total_messages,
        COALESCE(SUM(CAST(m.metadata->>'tokens' AS INTEGER)), 0) as total_tokens,
        COALESCE(SUM(CAST(m.metadata->>'cost' AS DECIMAL)), 0.0) as total_cost
    FROM ai_conversations c
    LEFT JOIN ai_messages m ON c.id = m.conversation_id
    WHERE DATE(c.created_at) = target_date
    GROUP BY c.tenant_id, c.user_id, c.assistant_type
    ON CONFLICT (tenant_id, user_id, date, assistant_type) 
    DO UPDATE SET
        total_conversations = EXCLUDED.total_conversations,
        total_messages = EXCLUDED.total_messages,
        total_tokens = EXCLUDED.total_tokens,
        total_cost = EXCLUDED.total_cost,
        updated_at = CURRENT_TIMESTAMP;
END;
$$ LANGUAGE plpgsql;

-- Create a view for AI conversation analytics
CREATE OR REPLACE VIEW ai_conversation_analytics AS
SELECT 
    c.tenant_id,
    c.assistant_type,
    DATE(c.created_at) as conversation_date,
    COUNT(c.id) as conversation_count,
    COUNT(DISTINCT c.user_id) as unique_users,
    COUNT(DISTINCT c.customer_id) as unique_customers,
    AVG(EXTRACT(EPOCH FROM (c.ended_at - c.created_at))/60) as avg_duration_minutes,
    COUNT(CASE WHEN c.status = 'ended' THEN 1 END) as completed_conversations,
    COUNT(CASE WHEN c.status = 'active' THEN 1 END) as active_conversations
FROM ai_conversations c
GROUP BY c.tenant_id, c.assistant_type, DATE(c.created_at);

-- Create a view for AI function usage analytics
CREATE OR REPLACE VIEW ai_function_analytics AS
SELECT 
    f.tenant_id,
    f.function_name,
    DATE(f.created_at) as usage_date,
    COUNT(*) as total_calls,
    COUNT(CASE WHEN f.success THEN 1 END) as successful_calls,
    COUNT(CASE WHEN NOT f.success THEN 1 END) as failed_calls,
    AVG(f.execution_time_ms) as avg_execution_time_ms,
    COUNT(DISTINCT f.user_id) as unique_users,
    COUNT(DISTINCT f.conversation_id) as unique_conversations
FROM ai_function_usage f
GROUP BY f.tenant_id, f.function_name, DATE(f.created_at);

-- Add comments for documentation
COMMENT ON TABLE ai_conversations IS 'Stores AI assistant conversation contexts and metadata';
COMMENT ON TABLE ai_messages IS 'Stores individual messages within AI conversations';
COMMENT ON TABLE ai_conversation_summaries IS 'Stores generated summaries of completed conversations';
COMMENT ON TABLE ai_usage_metrics IS 'Tracks daily AI usage metrics for billing and monitoring';
COMMENT ON TABLE ai_function_usage IS 'Tracks individual AI function/tool calls for analytics';

COMMENT ON COLUMN ai_conversations.assistant_type IS 'Type of AI assistant: customer or business';
COMMENT ON COLUMN ai_conversations.session_data IS 'JSON data for conversation context and state';
COMMENT ON COLUMN ai_messages.tool_calls IS 'JSON array of AI function/tool calls made';
COMMENT ON COLUMN ai_messages.tool_results IS 'JSON array of function/tool execution results';
COMMENT ON COLUMN ai_usage_metrics.total_cost IS 'Total cost in dollars for AI API usage';