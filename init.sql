-- Simple database initialization for demo
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Basic tenant table
CREATE TABLE tenants (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(100) UNIQUE NOT NULL,
    plan VARCHAR(50) NOT NULL DEFAULT 'basic',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Basic users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id),
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Basic customers table
CREATE TABLE customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id),
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255),
    phone VARCHAR(50),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Basic jobs table
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id UUID REFERENCES tenants(id),
    customer_id UUID REFERENCES customers(id),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'pending',
    scheduled_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insert demo data
INSERT INTO tenants (id, name, domain, plan) VALUES 
    ('00000000-0000-0000-0000-000000000001', 'Demo Landscaping Co', 'demo', 'professional');

INSERT INTO users (tenant_id, email, name, role) VALUES 
    ('00000000-0000-0000-0000-000000000001', 'admin@demo.com', 'Admin User', 'admin');

INSERT INTO customers (tenant_id, name, email, phone, address) VALUES 
    ('00000000-0000-0000-0000-000000000001', 'John Smith', 'john@example.com', '555-0123', '123 Oak Street'),
    ('00000000-0000-0000-0000-000000000001', 'Jane Doe', 'jane@example.com', '555-0456', '456 Maple Avenue');

INSERT INTO jobs (tenant_id, customer_id, title, description, status, scheduled_date) VALUES 
    ('00000000-0000-0000-0000-000000000001', 
     (SELECT id FROM customers WHERE email = 'john@example.com'), 
     'Lawn Mowing', 'Weekly lawn maintenance service', 'scheduled', CURRENT_DATE + 1),
    ('00000000-0000-0000-0000-000000000001', 
     (SELECT id FROM customers WHERE email = 'jane@example.com'), 
     'Garden Cleanup', 'Spring garden cleanup and mulching', 'pending', CURRENT_DATE + 3);