-- Rollback Enhanced Multi-tenant Schema Migration

-- Drop RLS policies
DROP POLICY IF EXISTS tenant_isolation ON tenants;
DROP POLICY IF EXISTS user_tenant_isolation ON users;
DROP POLICY IF EXISTS customer_tenant_isolation ON customers;
DROP POLICY IF EXISTS property_tenant_isolation ON properties;
DROP POLICY IF EXISTS service_tenant_isolation ON services;
DROP POLICY IF EXISTS job_tenant_isolation ON jobs;
DROP POLICY IF EXISTS job_service_tenant_isolation ON job_services;
DROP POLICY IF EXISTS invoice_tenant_isolation ON invoices;
DROP POLICY IF EXISTS payment_tenant_isolation ON payments;
DROP POLICY IF EXISTS equipment_tenant_isolation ON equipment;
DROP POLICY IF EXISTS file_attachment_tenant_isolation ON file_attachments;
DROP POLICY IF EXISTS api_key_tenant_isolation ON api_keys;
DROP POLICY IF EXISTS session_user_isolation ON user_sessions;
DROP POLICY IF EXISTS audit_log_tenant_isolation ON audit_logs;
DROP POLICY IF EXISTS notification_user_isolation ON notifications;
DROP POLICY IF EXISTS webhook_tenant_isolation ON webhooks;
DROP POLICY IF EXISTS webhook_delivery_tenant_isolation ON webhook_deliveries;
DROP POLICY IF EXISTS crew_tenant_isolation ON crews;
DROP POLICY IF EXISTS crew_member_tenant_isolation ON crew_members;
DROP POLICY IF EXISTS quote_tenant_isolation ON quotes;
DROP POLICY IF EXISTS quote_service_tenant_isolation ON quote_services;
DROP POLICY IF EXISTS schedule_template_tenant_isolation ON schedule_templates;

-- Disable Row Level Security
ALTER TABLE tenants DISABLE ROW LEVEL SECURITY;
ALTER TABLE users DISABLE ROW LEVEL SECURITY;
ALTER TABLE customers DISABLE ROW LEVEL SECURITY;
ALTER TABLE properties DISABLE ROW LEVEL SECURITY;
ALTER TABLE services DISABLE ROW LEVEL SECURITY;
ALTER TABLE jobs DISABLE ROW LEVEL SECURITY;
ALTER TABLE job_services DISABLE ROW LEVEL SECURITY;
ALTER TABLE invoices DISABLE ROW LEVEL SECURITY;
ALTER TABLE payments DISABLE ROW LEVEL SECURITY;
ALTER TABLE equipment DISABLE ROW LEVEL SECURITY;
ALTER TABLE file_attachments DISABLE ROW LEVEL SECURITY;

-- Drop functions
DROP FUNCTION IF EXISTS set_tenant_context(UUID, UUID, VARCHAR);
DROP FUNCTION IF EXISTS clear_tenant_context();
DROP FUNCTION IF EXISTS generate_job_number(UUID);
DROP FUNCTION IF EXISTS generate_quote_number(UUID);
DROP FUNCTION IF EXISTS generate_invoice_number(UUID);

-- Drop triggers for new tables
DROP TRIGGER IF EXISTS update_api_keys_updated_at ON api_keys;
DROP TRIGGER IF EXISTS update_webhooks_updated_at ON webhooks;
DROP TRIGGER IF EXISTS update_crews_updated_at ON crews;
DROP TRIGGER IF EXISTS update_quotes_updated_at ON quotes;
DROP TRIGGER IF EXISTS update_schedule_templates_updated_at ON schedule_templates;

-- Drop new tables (in reverse order of dependencies)
DROP TABLE IF EXISTS webhook_deliveries;
DROP TABLE IF EXISTS webhooks;
DROP TABLE IF EXISTS quote_services;
DROP TABLE IF EXISTS quotes;
DROP TABLE IF EXISTS crew_members;
DROP TABLE IF EXISTS crews;
DROP TABLE IF EXISTS schedule_templates;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS audit_logs;
DROP TABLE IF EXISTS user_sessions;
DROP TABLE IF EXISTS api_keys;

-- Remove added columns from existing tables
ALTER TABLE tenants DROP COLUMN IF EXISTS domain;
ALTER TABLE tenants DROP COLUMN IF EXISTS logo_url;
ALTER TABLE tenants DROP COLUMN IF EXISTS theme_config;
ALTER TABLE tenants DROP COLUMN IF EXISTS billing_settings;
ALTER TABLE tenants DROP COLUMN IF EXISTS feature_flags;
ALTER TABLE tenants DROP COLUMN IF EXISTS max_users;
ALTER TABLE tenants DROP COLUMN IF EXISTS max_customers;
ALTER TABLE tenants DROP COLUMN IF EXISTS storage_quota_gb;
ALTER TABLE tenants DROP COLUMN IF EXISTS trial_ends_at;

ALTER TABLE users DROP COLUMN IF EXISTS phone;
ALTER TABLE users DROP COLUMN IF EXISTS avatar_url;
ALTER TABLE users DROP COLUMN IF EXISTS timezone;
ALTER TABLE users DROP COLUMN IF EXISTS language;
ALTER TABLE users DROP COLUMN IF EXISTS permissions;
ALTER TABLE users DROP COLUMN IF EXISTS two_factor_enabled;
ALTER TABLE users DROP COLUMN IF EXISTS two_factor_secret;
ALTER TABLE users DROP COLUMN IF EXISTS backup_codes;
ALTER TABLE users DROP COLUMN IF EXISTS failed_login_attempts;
ALTER TABLE users DROP COLUMN IF EXISTS locked_until;

ALTER TABLE customers DROP COLUMN IF EXISTS company_name;
ALTER TABLE customers DROP COLUMN IF EXISTS tax_id;
ALTER TABLE customers DROP COLUMN IF EXISTS preferred_contact_method;
ALTER TABLE customers DROP COLUMN IF EXISTS lead_source;
ALTER TABLE customers DROP COLUMN IF EXISTS customer_type;
ALTER TABLE customers DROP COLUMN IF EXISTS credit_limit;
ALTER TABLE customers DROP COLUMN IF EXISTS payment_terms;

ALTER TABLE properties DROP COLUMN IF EXISTS latitude;
ALTER TABLE properties DROP COLUMN IF EXISTS longitude;
ALTER TABLE properties DROP COLUMN IF EXISTS square_footage;
ALTER TABLE properties DROP COLUMN IF EXISTS access_instructions;
ALTER TABLE properties DROP COLUMN IF EXISTS gate_code;
ALTER TABLE properties DROP COLUMN IF EXISTS special_instructions;
ALTER TABLE properties DROP COLUMN IF EXISTS property_value;

ALTER TABLE jobs DROP COLUMN IF EXISTS job_number;
ALTER TABLE jobs DROP COLUMN IF EXISTS recurring_schedule;
ALTER TABLE jobs DROP COLUMN IF EXISTS parent_job_id;
ALTER TABLE jobs DROP COLUMN IF EXISTS weather_dependent;
ALTER TABLE jobs DROP COLUMN IF EXISTS requires_equipment;
ALTER TABLE jobs DROP COLUMN IF EXISTS crew_size;
ALTER TABLE jobs DROP COLUMN IF EXISTS completion_photos;
ALTER TABLE jobs DROP COLUMN IF EXISTS customer_signature;
ALTER TABLE jobs DROP COLUMN IF EXISTS gps_check_in;
ALTER TABLE jobs DROP COLUMN IF EXISTS gps_check_out;