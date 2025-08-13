-- Rollback initial schema migration

-- Drop triggers
DROP TRIGGER IF EXISTS update_equipment_updated_at ON equipment;
DROP TRIGGER IF EXISTS update_payments_updated_at ON payments;
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
DROP TRIGGER IF EXISTS update_jobs_updated_at ON jobs;
DROP TRIGGER IF EXISTS update_services_updated_at ON services;
DROP TRIGGER IF EXISTS update_properties_updated_at ON properties;
DROP TRIGGER IF EXISTS update_customers_updated_at ON customers;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes (will be dropped automatically with tables, but listing for clarity)
DROP INDEX IF EXISTS idx_file_attachments_entity;
DROP INDEX IF EXISTS idx_file_attachments_tenant_id;
DROP INDEX IF EXISTS idx_equipment_tenant_id;
DROP INDEX IF EXISTS idx_payments_invoice_id;
DROP INDEX IF EXISTS idx_payments_tenant_id;
DROP INDEX IF EXISTS idx_invoices_job_id;
DROP INDEX IF EXISTS idx_invoices_customer_id;
DROP INDEX IF EXISTS idx_invoices_tenant_id;
DROP INDEX IF EXISTS idx_job_services_service_id;
DROP INDEX IF EXISTS idx_job_services_job_id;
DROP INDEX IF EXISTS idx_jobs_status;
DROP INDEX IF EXISTS idx_jobs_scheduled_date;
DROP INDEX IF EXISTS idx_jobs_assigned_user_id;
DROP INDEX IF EXISTS idx_jobs_property_id;
DROP INDEX IF EXISTS idx_jobs_customer_id;
DROP INDEX IF EXISTS idx_jobs_tenant_id;
DROP INDEX IF EXISTS idx_services_tenant_id;
DROP INDEX IF EXISTS idx_properties_customer_id;
DROP INDEX IF EXISTS idx_properties_tenant_id;
DROP INDEX IF EXISTS idx_customers_tenant_id;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_tenant_id;

-- Drop tables in reverse order of dependencies
DROP TABLE IF EXISTS file_attachments;
DROP TABLE IF EXISTS equipment;
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS invoices;
DROP TABLE IF EXISTS job_services;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS properties;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;

-- Drop extensions
DROP EXTENSION IF EXISTS "uuid-ossp";