-- AI Assistant Schema Migration Rollback
-- This migration removes AI assistant functionality

-- Drop views
DROP VIEW IF EXISTS ai_function_analytics;
DROP VIEW IF EXISTS ai_conversation_analytics;

-- Drop functions
DROP FUNCTION IF EXISTS aggregate_ai_usage_metrics(DATE);
DROP FUNCTION IF EXISTS cleanup_old_ai_conversations(INTEGER);

-- Drop triggers
DROP TRIGGER IF EXISTS trigger_ai_usage_metrics_updated_at ON ai_usage_metrics;
DROP TRIGGER IF EXISTS trigger_ai_conversations_updated_at ON ai_conversations;

-- Drop trigger functions
DROP FUNCTION IF EXISTS update_ai_usage_metrics_updated_at();
DROP FUNCTION IF EXISTS update_ai_conversations_updated_at();

-- Drop indexes (will be automatically dropped with tables, but explicit for clarity)
DROP INDEX IF EXISTS idx_ai_function_usage_created_at;
DROP INDEX IF EXISTS idx_ai_function_usage_function_name;
DROP INDEX IF EXISTS idx_ai_function_usage_conversation_id;
DROP INDEX IF EXISTS idx_ai_function_usage_user_id;
DROP INDEX IF EXISTS idx_ai_function_usage_tenant_id;

DROP INDEX IF EXISTS idx_ai_usage_metrics_assistant_type;
DROP INDEX IF EXISTS idx_ai_usage_metrics_date;
DROP INDEX IF EXISTS idx_ai_usage_metrics_user_id;
DROP INDEX IF EXISTS idx_ai_usage_metrics_tenant_id;

DROP INDEX IF EXISTS idx_ai_messages_created_at;
DROP INDEX IF EXISTS idx_ai_messages_role;
DROP INDEX IF EXISTS idx_ai_messages_conversation_id;

DROP INDEX IF EXISTS idx_ai_conversations_created_at;
DROP INDEX IF EXISTS idx_ai_conversations_status;
DROP INDEX IF EXISTS idx_ai_conversations_assistant_type;
DROP INDEX IF EXISTS idx_ai_conversations_customer_id;
DROP INDEX IF EXISTS idx_ai_conversations_user_id;
DROP INDEX IF EXISTS idx_ai_conversations_tenant_id;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS ai_function_usage;
DROP TABLE IF EXISTS ai_usage_metrics;
DROP TABLE IF EXISTS ai_conversation_summaries;
DROP TABLE IF EXISTS ai_messages;
DROP TABLE IF EXISTS ai_conversations;