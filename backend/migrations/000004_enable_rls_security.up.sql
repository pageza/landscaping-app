-- Row Level Security (RLS) Implementation for Multi-Tenant Data Isolation
-- This migration enables comprehensive RLS policies to ensure tenant data isolation

-- Enable RLS on all tenant-aware tables
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

-- Create a function to get the current tenant ID from application context
CREATE OR REPLACE FUNCTION current_tenant_id() RETURNS UUID AS $$
DECLARE
    tenant_id UUID;
BEGIN
    -- Get tenant ID from session variable set by application
    SELECT current_setting('app.current_tenant_id', true)::UUID INTO tenant_id;
    
    -- If no tenant ID is set, return NULL (which will deny access)
    IF tenant_id IS NULL THEN
        RAISE EXCEPTION 'No tenant context set. Access denied.';
    END IF;
    
    RETURN tenant_id;
EXCEPTION
    WHEN OTHERS THEN
        -- If any error occurs (including invalid UUID), deny access
        RAISE EXCEPTION 'Invalid tenant context. Access denied.';
END;
$$ LANGUAGE plpgsql SECURITY DEFINER STABLE;

-- Create a function to get the current user ID from application context
CREATE OR REPLACE FUNCTION current_user_id() RETURNS UUID AS $$
DECLARE
    user_id UUID;
BEGIN
    SELECT current_setting('app.current_user_id', true)::UUID INTO user_id;
    RETURN user_id;
EXCEPTION
    WHEN OTHERS THEN
        RETURN NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER STABLE;

-- Create a function to check if the current user is a super admin
CREATE OR REPLACE FUNCTION is_super_admin() RETURNS BOOLEAN AS $$
DECLARE
    user_role TEXT;
BEGIN
    SELECT current_setting('app.current_user_role', true) INTO user_role;
    RETURN user_role = 'super_admin';
EXCEPTION
    WHEN OTHERS THEN
        RETURN FALSE;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER STABLE;

-- Tenants table policies
-- Super admins can access all tenants, regular users can only access their own tenant
CREATE POLICY tenant_isolation_policy ON tenants
    FOR ALL
    USING (
        is_super_admin() OR 
        id = current_tenant_id()
    );

-- Users table policies
-- Users can only access users within their tenant (except super admins)
CREATE POLICY user_tenant_isolation ON users
    FOR ALL
    USING (
        is_super_admin() OR 
        tenant_id = current_tenant_id()
    );

-- Additional policy for users to only modify their own record (except admins)
CREATE POLICY user_self_modification ON users
    FOR UPDATE
    USING (
        is_super_admin() OR
        (tenant_id = current_tenant_id() AND 
         (current_setting('app.current_user_role', true) IN ('owner', 'admin') OR 
          id = current_user_id())
        )
    );

-- Customers table policies
CREATE POLICY customer_tenant_isolation ON customers
    FOR ALL
    USING (
        is_super_admin() OR 
        tenant_id = current_tenant_id()
    );

-- Properties table policies
CREATE POLICY property_tenant_isolation ON properties
    FOR ALL
    USING (
        is_super_admin() OR 
        tenant_id = current_tenant_id()
    );

-- Services table policies
CREATE POLICY service_tenant_isolation ON services
    FOR ALL
    USING (
        is_super_admin() OR 
        tenant_id = current_tenant_id()
    );

-- Jobs table policies
CREATE POLICY job_tenant_isolation ON jobs
    FOR ALL
    USING (
        is_super_admin() OR 
        tenant_id = current_tenant_id()
    );

-- Additional policy for crew members to only see their assigned jobs
CREATE POLICY job_crew_access ON jobs
    FOR SELECT
    USING (
        is_super_admin() OR
        tenant_id = current_tenant_id() AND
        (current_setting('app.current_user_role', true) != 'crew' OR 
         assigned_user_id = current_user_id())
    );

-- Job services table policies (inherits from jobs)
CREATE POLICY job_service_tenant_isolation ON job_services
    FOR ALL
    USING (
        is_super_admin() OR 
        EXISTS (
            SELECT 1 FROM jobs 
            WHERE jobs.id = job_services.job_id 
            AND jobs.tenant_id = current_tenant_id()
        )
    );

-- Invoices table policies
CREATE POLICY invoice_tenant_isolation ON invoices
    FOR ALL
    USING (
        is_super_admin() OR 
        tenant_id = current_tenant_id()
    );

-- Payments table policies
CREATE POLICY payment_tenant_isolation ON payments
    FOR ALL
    USING (
        is_super_admin() OR 
        tenant_id = current_tenant_id()
    );

-- Equipment table policies
CREATE POLICY equipment_tenant_isolation ON equipment
    FOR ALL
    USING (
        is_super_admin() OR 
        tenant_id = current_tenant_id()
    );

-- File attachments table policies
CREATE POLICY file_attachment_tenant_isolation ON file_attachments
    FOR ALL
    USING (
        is_super_admin() OR 
        tenant_id = current_tenant_id()
    );

-- Create audit table for RLS policy violations
CREATE TABLE IF NOT EXISTS rls_audit_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    table_name TEXT NOT NULL,
    operation TEXT NOT NULL,
    attempted_tenant_id UUID,
    current_tenant_id UUID,
    user_id UUID,
    user_role TEXT,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create function to log RLS violations
CREATE OR REPLACE FUNCTION log_rls_violation(
    p_table_name TEXT,
    p_operation TEXT,
    p_attempted_tenant_id UUID DEFAULT NULL
) RETURNS VOID AS $$
BEGIN
    INSERT INTO rls_audit_log (
        table_name,
        operation,
        attempted_tenant_id,
        current_tenant_id,
        user_id,
        user_role,
        ip_address,
        user_agent
    ) VALUES (
        p_table_name,
        p_operation,
        p_attempted_tenant_id,
        current_setting('app.current_tenant_id', true)::UUID,
        current_setting('app.current_user_id', true)::UUID,
        current_setting('app.current_user_role', true),
        current_setting('app.client_ip', true)::INET,
        current_setting('app.user_agent', true)
    );
EXCEPTION
    WHEN OTHERS THEN
        -- Don't fail the original operation if audit logging fails
        NULL;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create helper function to set application context securely
CREATE OR REPLACE FUNCTION set_application_context(
    p_tenant_id UUID,
    p_user_id UUID,
    p_user_role TEXT,
    p_client_ip TEXT DEFAULT NULL,
    p_user_agent TEXT DEFAULT NULL
) RETURNS VOID AS $$
BEGIN
    -- Validate tenant_id and user_id are not null
    IF p_tenant_id IS NULL OR p_user_id IS NULL THEN
        RAISE EXCEPTION 'Tenant ID and User ID cannot be null';
    END IF;
    
    -- Validate user_role
    IF p_user_role NOT IN ('super_admin', 'owner', 'admin', 'user', 'crew', 'customer') THEN
        RAISE EXCEPTION 'Invalid user role: %', p_user_role;
    END IF;
    
    -- Set session variables
    PERFORM set_config('app.current_tenant_id', p_tenant_id::TEXT, false);
    PERFORM set_config('app.current_user_id', p_user_id::TEXT, false);
    PERFORM set_config('app.current_user_role', p_user_role, false);
    
    IF p_client_ip IS NOT NULL THEN
        PERFORM set_config('app.client_ip', p_client_ip, false);
    END IF;
    
    IF p_user_agent IS NOT NULL THEN
        PERFORM set_config('app.user_agent', p_user_agent, false);
    END IF;
    
    -- Log the context setting for audit purposes
    INSERT INTO rls_audit_log (
        table_name,
        operation,
        current_tenant_id,
        user_id,
        user_role,
        ip_address,
        user_agent
    ) VALUES (
        'CONTEXT',
        'SET_CONTEXT',
        p_tenant_id,
        p_user_id,
        p_user_role,
        p_client_ip::INET,
        p_user_agent
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function to clear application context
CREATE OR REPLACE FUNCTION clear_application_context() RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', '', false);
    PERFORM set_config('app.current_user_id', '', false);
    PERFORM set_config('app.current_user_role', '', false);
    PERFORM set_config('app.client_ip', '', false);
    PERFORM set_config('app.user_agent', '', false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create function for safe tenant switching (only for super admins)
CREATE OR REPLACE FUNCTION switch_tenant_context(p_new_tenant_id UUID) RETURNS VOID AS $$
BEGIN
    -- Only super admins can switch tenant context
    IF NOT is_super_admin() THEN
        RAISE EXCEPTION 'Insufficient privileges to switch tenant context';
    END IF;
    
    -- Validate the new tenant exists
    IF NOT EXISTS (SELECT 1 FROM tenants WHERE id = p_new_tenant_id) THEN
        RAISE EXCEPTION 'Tenant does not exist: %', p_new_tenant_id;
    END IF;
    
    -- Set new tenant context
    PERFORM set_config('app.current_tenant_id', p_new_tenant_id::TEXT, false);
    
    -- Log the tenant switch
    PERFORM log_rls_violation('CONTEXT', 'SWITCH_TENANT', p_new_tenant_id);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Create indexes for better RLS performance
CREATE INDEX IF NOT EXISTS idx_rls_audit_log_tenant_id ON rls_audit_log(current_tenant_id);
CREATE INDEX IF NOT EXISTS idx_rls_audit_log_user_id ON rls_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_rls_audit_log_created_at ON rls_audit_log(created_at);

-- Create a view for monitoring RLS violations
CREATE OR REPLACE VIEW rls_violation_summary AS
SELECT 
    table_name,
    operation,
    COUNT(*) as violation_count,
    COUNT(DISTINCT user_id) as unique_users,
    MIN(created_at) as first_violation,
    MAX(created_at) as last_violation
FROM rls_audit_log
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY table_name, operation
ORDER BY violation_count DESC;

-- Grant appropriate permissions
GRANT EXECUTE ON FUNCTION current_tenant_id() TO authenticated_users;
GRANT EXECUTE ON FUNCTION current_user_id() TO authenticated_users;
GRANT EXECUTE ON FUNCTION is_super_admin() TO authenticated_users;
GRANT EXECUTE ON FUNCTION set_application_context(UUID, UUID, TEXT, TEXT, TEXT) TO authenticated_users;
GRANT EXECUTE ON FUNCTION clear_application_context() TO authenticated_users;
GRANT EXECUTE ON FUNCTION switch_tenant_context(UUID) TO authenticated_users;
GRANT SELECT ON rls_violation_summary TO authenticated_users;

-- Create a role for the application with limited privileges
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'landscaping_app_user') THEN
        CREATE ROLE landscaping_app_user;
    END IF;
END
$$;

-- Grant necessary permissions to the application role
GRANT USAGE ON SCHEMA public TO landscaping_app_user;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO landscaping_app_user;
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO landscaping_app_user;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA public TO landscaping_app_user;

-- Create a role for authenticated users (used in RLS policies)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'authenticated_users') THEN
        CREATE ROLE authenticated_users;
    END IF;
END
$$;

-- Add application role to authenticated users
GRANT authenticated_users TO landscaping_app_user;