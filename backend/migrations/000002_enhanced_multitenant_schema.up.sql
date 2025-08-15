-- Enhanced Multi-tenant Schema Migration for Landscaping SaaS
-- This migration adds Row Level Security, advanced indexing, and additional tables

-- Enable Row Level Security on existing tables
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE customers ENABLE ROW LEVEL SECURITY;
ALTER TABLE properties ENABLE ROW LEVEL SECURITY;
ALTER TABLE services ENABLE ROW LEVEL SECURITY;
ALTER TABLE jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE job_services ENABLE ROW LEVEL SECURITY;
ALTER TABLE invoices ENABLE ROW LEVEL SECURITY;
ALTER TABLE payments ENABLE ROW LEVEL SECURITY;
ALTER TABLE equipment ENABLE ROW LEVEL SECURITY;
ALTER TABLE file_attachments ENABLE ROW LEVEL SECURITY;

-- Add enhanced tenant fields
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS domain VARCHAR(255);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS logo_url VARCHAR(500);
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS theme_config JSONB DEFAULT '{}';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS billing_settings JSONB DEFAULT '{}';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS feature_flags JSONB DEFAULT '{}';
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS max_users INTEGER DEFAULT 10;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS max_customers INTEGER DEFAULT 1000;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS storage_quota_gb INTEGER DEFAULT 10;
ALTER TABLE tenants ADD COLUMN IF NOT EXISTS trial_ends_at TIMESTAMP WITH TIME ZONE;

-- Add user enhancements
ALTER TABLE users ADD COLUMN IF NOT EXISTS phone VARCHAR(20);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500);
ALTER TABLE users ADD COLUMN IF NOT EXISTS timezone VARCHAR(100) DEFAULT 'UTC';
ALTER TABLE users ADD COLUMN IF NOT EXISTS language VARCHAR(5) DEFAULT 'en';
ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions JSONB DEFAULT '[]';
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS two_factor_secret VARCHAR(32);
ALTER TABLE users ADD COLUMN IF NOT EXISTS backup_codes JSONB;
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INTEGER DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until TIMESTAMP WITH TIME ZONE;

-- Enhanced customer fields
ALTER TABLE customers ADD COLUMN IF NOT EXISTS company_name VARCHAR(255);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS tax_id VARCHAR(50);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS preferred_contact_method VARCHAR(20) DEFAULT 'email';
ALTER TABLE customers ADD COLUMN IF NOT EXISTS lead_source VARCHAR(100);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS customer_type VARCHAR(50) DEFAULT 'residential';
ALTER TABLE customers ADD COLUMN IF NOT EXISTS credit_limit DECIMAL(10,2);
ALTER TABLE customers ADD COLUMN IF NOT EXISTS payment_terms INTEGER DEFAULT 30;

-- Enhanced property fields
ALTER TABLE properties ADD COLUMN IF NOT EXISTS latitude DECIMAL(10, 8);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS longitude DECIMAL(11, 8);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS square_footage INTEGER;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS access_instructions TEXT;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS gate_code VARCHAR(50);
ALTER TABLE properties ADD COLUMN IF NOT EXISTS special_instructions TEXT;
ALTER TABLE properties ADD COLUMN IF NOT EXISTS property_value DECIMAL(12,2);

-- Enhanced job fields
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS job_number VARCHAR(50);
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS recurring_schedule VARCHAR(50);
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS parent_job_id UUID REFERENCES jobs(id);
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS weather_dependent BOOLEAN DEFAULT FALSE;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS requires_equipment JSONB DEFAULT '[]';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS crew_size INTEGER DEFAULT 1;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS completion_photos JSONB DEFAULT '[]';
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS customer_signature TEXT;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS gps_check_in JSONB;
ALTER TABLE jobs ADD COLUMN IF NOT EXISTS gps_check_out JSONB;

-- Create API Keys table for integrations
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(20) NOT NULL,
    permissions JSONB DEFAULT '[]',
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create User Sessions table for session management
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) NOT NULL UNIQUE,
    refresh_token VARCHAR(255) NOT NULL UNIQUE,
    device_info JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create Audit Log table for security and compliance
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50) NOT NULL,
    resource_id UUID,
    old_values JSONB,
    new_values JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create Notifications table
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    data JSONB DEFAULT '{}',
    read_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create Webhooks table for external integrations
CREATE TABLE IF NOT EXISTS webhooks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(500) NOT NULL,
    secret VARCHAR(255) NOT NULL,
    events JSONB NOT NULL DEFAULT '[]',
    headers JSONB DEFAULT '{}',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    retry_count INTEGER DEFAULT 0,
    last_success_at TIMESTAMP WITH TIME ZONE,
    last_failure_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create Webhook Deliveries table for tracking
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    webhook_id UUID NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    response_status INTEGER,
    response_body TEXT,
    delivered_at TIMESTAMP WITH TIME ZONE,
    failed_at TIMESTAMP WITH TIME ZONE,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create Crew Management tables
CREATE TABLE IF NOT EXISTS crews (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    capacity INTEGER NOT NULL DEFAULT 1,
    specializations JSONB DEFAULT '[]',
    equipment_ids JSONB DEFAULT '[]',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS crew_members (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    crew_id UUID NOT NULL REFERENCES crews(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'member',
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    left_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(crew_id, user_id, left_at)
);

-- Create Quotes table
CREATE TABLE IF NOT EXISTS quotes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
    property_id UUID NOT NULL REFERENCES properties(id) ON DELETE CASCADE,
    quote_number VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    subtotal DECIMAL(10,2) NOT NULL DEFAULT 0,
    tax_rate DECIMAL(5,4) DEFAULT 0,
    tax_amount DECIMAL(10,2) DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'draft',
    valid_until DATE,
    terms_and_conditions TEXT,
    notes TEXT,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMP WITH TIME ZONE,
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(tenant_id, quote_number)
);

-- Create Quote Services junction table
CREATE TABLE IF NOT EXISTS quote_services (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    quote_id UUID NOT NULL REFERENCES quotes(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    quantity DECIMAL(10,2) NOT NULL DEFAULT 1,
    unit_price DECIMAL(10,2) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create Schedule Templates for recurring jobs
CREATE TABLE IF NOT EXISTS schedule_templates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    frequency VARCHAR(50) NOT NULL, -- weekly, monthly, seasonal, etc.
    frequency_config JSONB NOT NULL DEFAULT '{}', -- days of week, etc.
    service_ids JSONB NOT NULL DEFAULT '[]',
    default_duration INTEGER, -- minutes
    default_crew_size INTEGER DEFAULT 1,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Enhanced indexes for performance
CREATE INDEX IF NOT EXISTS idx_tenants_subdomain ON tenants(subdomain);
CREATE INDEX IF NOT EXISTS idx_tenants_domain ON tenants(domain);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);

CREATE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_email_verified ON users(email_verified);

CREATE INDEX IF NOT EXISTS idx_customers_tenant_name ON customers(tenant_id, first_name, last_name);
CREATE INDEX IF NOT EXISTS idx_customers_email ON customers(email);
CREATE INDEX IF NOT EXISTS idx_customers_phone ON customers(phone);

CREATE INDEX IF NOT EXISTS idx_properties_coordinates ON properties(latitude, longitude);
CREATE INDEX IF NOT EXISTS idx_properties_type ON properties(property_type);

CREATE INDEX IF NOT EXISTS idx_jobs_number ON jobs(tenant_id, job_number);
CREATE INDEX IF NOT EXISTS idx_jobs_schedule ON jobs(scheduled_date, scheduled_time);
CREATE INDEX IF NOT EXISTS idx_jobs_priority_status ON jobs(priority, status);
CREATE INDEX IF NOT EXISTS idx_jobs_recurring ON jobs(recurring_schedule);

CREATE INDEX IF NOT EXISTS idx_api_keys_tenant_status ON api_keys(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(key_prefix);

CREATE INDEX IF NOT EXISTS idx_sessions_user_status ON user_sessions(user_id, status);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON user_sessions(session_token);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON user_sessions(expires_at);

CREATE INDEX IF NOT EXISTS idx_audit_tenant_resource ON audit_logs(tenant_id, resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_user_action ON audit_logs(user_id, action);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at);

CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, read_at);
CREATE INDEX IF NOT EXISTS idx_notifications_type ON notifications(type);

CREATE INDEX IF NOT EXISTS idx_webhooks_tenant_status ON webhooks(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook_status ON webhook_deliveries(webhook_id, response_status);

CREATE INDEX IF NOT EXISTS idx_crews_tenant_status ON crews(tenant_id, status);
CREATE INDEX IF NOT EXISTS idx_crew_members_crew_active ON crew_members(crew_id) WHERE left_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_quotes_tenant_number ON quotes(tenant_id, quote_number);
CREATE INDEX IF NOT EXISTS idx_quotes_customer_status ON quotes(customer_id, status);

-- Row Level Security Policies

-- Tenants: Super admins can see all, others only their own
CREATE POLICY tenant_isolation ON tenants
    USING (
        current_setting('app.current_user_role', true) = 'super_admin' OR
        id = current_setting('app.current_tenant_id', true)::uuid
    );

-- Users: Can only access users in their tenant
CREATE POLICY user_tenant_isolation ON users
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Customers: Tenant isolation
CREATE POLICY customer_tenant_isolation ON customers
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Properties: Tenant isolation
CREATE POLICY property_tenant_isolation ON properties
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Services: Tenant isolation
CREATE POLICY service_tenant_isolation ON services
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Jobs: Tenant isolation with user assignment restrictions
CREATE POLICY job_tenant_isolation ON jobs
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Job Services: Inherited from job tenant
CREATE POLICY job_service_tenant_isolation ON job_services
    USING (
        EXISTS (
            SELECT 1 FROM jobs 
            WHERE jobs.id = job_services.job_id 
            AND jobs.tenant_id = current_setting('app.current_tenant_id', true)::uuid
        )
    );

-- Invoices: Tenant isolation
CREATE POLICY invoice_tenant_isolation ON invoices
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Payments: Tenant isolation
CREATE POLICY payment_tenant_isolation ON payments
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Equipment: Tenant isolation
CREATE POLICY equipment_tenant_isolation ON equipment
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- File Attachments: Tenant isolation
CREATE POLICY file_attachment_tenant_isolation ON file_attachments
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- API Keys: Tenant isolation
CREATE POLICY api_key_tenant_isolation ON api_keys
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- User Sessions: User can only see their own sessions
CREATE POLICY session_user_isolation ON user_sessions
    USING (user_id = current_setting('app.current_user_id', true)::uuid);

-- Audit Logs: Tenant isolation (read-only for most users)
CREATE POLICY audit_log_tenant_isolation ON audit_logs
    USING (
        tenant_id = current_setting('app.current_tenant_id', true)::uuid AND
        current_setting('app.current_user_role', true) IN ('admin', 'owner', 'super_admin')
    );

-- Notifications: User can only see their own
CREATE POLICY notification_user_isolation ON notifications
    USING (
        user_id = current_setting('app.current_user_id', true)::uuid OR
        user_id IS NULL AND tenant_id = current_setting('app.current_tenant_id', true)::uuid
    );

-- Webhooks: Tenant isolation
CREATE POLICY webhook_tenant_isolation ON webhooks
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Webhook Deliveries: Inherited from webhook tenant
CREATE POLICY webhook_delivery_tenant_isolation ON webhook_deliveries
    USING (
        EXISTS (
            SELECT 1 FROM webhooks 
            WHERE webhooks.id = webhook_deliveries.webhook_id 
            AND webhooks.tenant_id = current_setting('app.current_tenant_id', true)::uuid
        )
    );

-- Crews: Tenant isolation
CREATE POLICY crew_tenant_isolation ON crews
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Crew Members: Inherited from crew tenant
CREATE POLICY crew_member_tenant_isolation ON crew_members
    USING (
        EXISTS (
            SELECT 1 FROM crews 
            WHERE crews.id = crew_members.crew_id 
            AND crews.tenant_id = current_setting('app.current_tenant_id', true)::uuid
        )
    );

-- Quotes: Tenant isolation
CREATE POLICY quote_tenant_isolation ON quotes
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Quote Services: Inherited from quote tenant
CREATE POLICY quote_service_tenant_isolation ON quote_services
    USING (
        EXISTS (
            SELECT 1 FROM quotes 
            WHERE quotes.id = quote_services.quote_id 
            AND quotes.tenant_id = current_setting('app.current_tenant_id', true)::uuid
        )
    );

-- Schedule Templates: Tenant isolation
CREATE POLICY schedule_template_tenant_isolation ON schedule_templates
    USING (tenant_id = current_setting('app.current_tenant_id', true)::uuid);

-- Create updated_at triggers for new tables
CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_webhooks_updated_at BEFORE UPDATE ON webhooks FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_crews_updated_at BEFORE UPDATE ON crews FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_quotes_updated_at BEFORE UPDATE ON quotes FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_schedule_templates_updated_at BEFORE UPDATE ON schedule_templates FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create function to generate unique job numbers
CREATE OR REPLACE FUNCTION generate_job_number(tenant_uuid UUID)
RETURNS VARCHAR AS $$
DECLARE
    year_suffix VARCHAR(2);
    sequence_num INTEGER;
    job_number VARCHAR(50);
BEGIN
    year_suffix := TO_CHAR(NOW(), 'YY');
    
    SELECT COALESCE(MAX(CAST(SUBSTRING(job_number FROM 'JOB-' || year_suffix || '-(\d+)') AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM jobs 
    WHERE tenant_id = tenant_uuid 
    AND job_number LIKE 'JOB-' || year_suffix || '-%';
    
    job_number := 'JOB-' || year_suffix || '-' || LPAD(sequence_num::TEXT, 5, '0');
    
    RETURN job_number;
END;
$$ LANGUAGE plpgsql;

-- Create function to generate unique quote numbers
CREATE OR REPLACE FUNCTION generate_quote_number(tenant_uuid UUID)
RETURNS VARCHAR AS $$
DECLARE
    year_suffix VARCHAR(2);
    sequence_num INTEGER;
    quote_number VARCHAR(50);
BEGIN
    year_suffix := TO_CHAR(NOW(), 'YY');
    
    SELECT COALESCE(MAX(CAST(SUBSTRING(quote_number FROM 'QUO-' || year_suffix || '-(\d+)') AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM quotes 
    WHERE tenant_id = tenant_uuid 
    AND quote_number LIKE 'QUO-' || year_suffix || '-%';
    
    quote_number := 'QUO-' || year_suffix || '-' || LPAD(sequence_num::TEXT, 5, '0');
    
    RETURN quote_number;
END;
$$ LANGUAGE plpgsql;

-- Create function to generate unique invoice numbers
CREATE OR REPLACE FUNCTION generate_invoice_number(tenant_uuid UUID)
RETURNS VARCHAR AS $$
DECLARE
    year_suffix VARCHAR(2);
    sequence_num INTEGER;
    invoice_number VARCHAR(50);
BEGIN
    year_suffix := TO_CHAR(NOW(), 'YY');
    
    SELECT COALESCE(MAX(CAST(SUBSTRING(invoice_number FROM 'INV-' || year_suffix || '-(\d+)') AS INTEGER)), 0) + 1
    INTO sequence_num
    FROM invoices 
    WHERE tenant_id = tenant_uuid 
    AND invoice_number LIKE 'INV-' || year_suffix || '-%';
    
    invoice_number := 'INV-' || year_suffix || '-' || LPAD(sequence_num::TEXT, 5, '0');
    
    RETURN invoice_number;
END;
$$ LANGUAGE plpgsql;

-- Security Functions for RLS context
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_uuid UUID, user_uuid UUID DEFAULT NULL, user_role VARCHAR DEFAULT NULL)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', tenant_uuid::text, true);
    IF user_uuid IS NOT NULL THEN
        PERFORM set_config('app.current_user_id', user_uuid::text, true);
    END IF;
    IF user_role IS NOT NULL THEN
        PERFORM set_config('app.current_user_role', user_role, true);
    END IF;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to clear tenant context
CREATE OR REPLACE FUNCTION clear_tenant_context()
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', '', true);
    PERFORM set_config('app.current_user_id', '', true);
    PERFORM set_config('app.current_user_role', '', true);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;