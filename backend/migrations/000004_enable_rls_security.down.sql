-- Rollback script for Row Level Security implementation

-- Drop RLS violation monitoring view
DROP VIEW IF EXISTS rls_violation_summary;

-- Drop RLS audit indexes
DROP INDEX IF EXISTS idx_rls_audit_log_tenant_id;
DROP INDEX IF EXISTS idx_rls_audit_log_user_id;
DROP INDEX IF EXISTS idx_rls_audit_log_created_at;

-- Drop helper functions
DROP FUNCTION IF EXISTS switch_tenant_context(UUID);
DROP FUNCTION IF EXISTS clear_application_context();
DROP FUNCTION IF EXISTS set_application_context(UUID, UUID, TEXT, TEXT, TEXT);
DROP FUNCTION IF EXISTS log_rls_violation(TEXT, TEXT, UUID);
DROP FUNCTION IF EXISTS is_super_admin();
DROP FUNCTION IF EXISTS current_user_id();
DROP FUNCTION IF EXISTS current_tenant_id();

-- Drop audit table
DROP TABLE IF EXISTS rls_audit_log;

-- Drop all RLS policies
DROP POLICY IF EXISTS tenant_isolation_policy ON tenants;
DROP POLICY IF EXISTS user_tenant_isolation ON users;
DROP POLICY IF EXISTS user_self_modification ON users;
DROP POLICY IF EXISTS customer_tenant_isolation ON customers;
DROP POLICY IF EXISTS property_tenant_isolation ON properties;
DROP POLICY IF EXISTS service_tenant_isolation ON services;
DROP POLICY IF EXISTS job_tenant_isolation ON jobs;
DROP POLICY IF EXISTS job_crew_access ON jobs;
DROP POLICY IF EXISTS job_service_tenant_isolation ON job_services;
DROP POLICY IF EXISTS invoice_tenant_isolation ON invoices;
DROP POLICY IF EXISTS payment_tenant_isolation ON payments;
DROP POLICY IF EXISTS equipment_tenant_isolation ON equipment;
DROP POLICY IF EXISTS file_attachment_tenant_isolation ON file_attachments;

-- Disable RLS on all tables
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

-- Drop application roles
REVOKE ALL PRIVILEGES ON SCHEMA public FROM landscaping_app_user;
REVOKE ALL PRIVILEGES ON ALL TABLES IN SCHEMA public FROM landscaping_app_user;
REVOKE ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public FROM landscaping_app_user;
REVOKE ALL PRIVILEGES ON ALL FUNCTIONS IN SCHEMA public FROM landscaping_app_user;

REVOKE authenticated_users FROM landscaping_app_user;

DROP ROLE IF EXISTS landscaping_app_user;
DROP ROLE IF EXISTS authenticated_users;