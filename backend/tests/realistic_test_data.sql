-- Realistic Test Data for Comprehensive Landscaping App Testing
-- This script creates realistic landscaper companies and customer users with properties
-- across different geographic zones for comprehensive business logic testing

BEGIN;

-- Enable UUID extension if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create function to generate realistic property coordinates
CREATE OR REPLACE FUNCTION generate_coords(base_lat DECIMAL, base_lng DECIMAL, radius_miles INTEGER)
RETURNS TABLE(lat DECIMAL, lng DECIMAL) AS $$
DECLARE
    random_radius DECIMAL;
    random_angle DECIMAL;
    lat_offset DECIMAL;
    lng_offset DECIMAL;
BEGIN
    -- Generate random distance within radius
    random_radius := random() * radius_miles * 0.014492754; -- Convert miles to degrees (roughly)
    random_angle := random() * 2 * pi();
    
    lat_offset := random_radius * cos(random_angle);
    lng_offset := random_radius * sin(random_angle);
    
    lat := base_lat + lat_offset;
    lng := base_lng + lng_offset;
    
    RETURN NEXT;
END;
$$ LANGUAGE plpgsql;

-- Clear existing test data
TRUNCATE TABLE payments, invoices, job_services, jobs, quotes, quote_services, 
                properties, customers, equipment, crews, crew_members, 
                api_keys, user_sessions, users, tenants RESTART IDENTITY CASCADE;

-- ===== TENANT: LANDSCAPING COMPANY =====
INSERT INTO tenants (id, name, subdomain, plan, status, domain, logo_url, settings, feature_flags, max_users, max_customers, storage_quota_gb, trial_ends_at, created_at, updated_at)
VALUES (
    '11111111-1111-1111-1111-111111111111'::uuid,
    'LandscapePro Solutions',
    'landscapepro',
    'professional',
    'active',
    'landscapepro.com',
    'https://landscapepro.com/logo.png',
    '{"timezone": "America/New_York", "currency": "USD", "business_hours": "7AM-6PM", "service_radius_miles": 30, "base_location": {"lat": 40.7128, "lng": -74.0060, "city": "New York", "state": "NY"}}',
    '{"route_optimization": true, "weather_integration": true, "customer_portal": true, "mobile_app": true}',
    25,
    5000,
    100,
    NOW() + INTERVAL '1 year',
    NOW(),
    NOW()
);

-- ===== LANDSCAPER COMPANY USERS (Admins) =====

-- Sarah Williams (Owner)
INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, phone, avatar_url, timezone, permissions, email_verified, created_at, updated_at)
VALUES (
    '22222222-2222-2222-2222-222222222222'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'sarah@landscapepro.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY.MzkwLlYXqsIa', -- sarah2024
    'Sarah',
    'Williams',
    'owner',
    'active',
    '+1-555-0001',
    'https://landscapepro.com/avatars/sarah.jpg',
    'America/New_York',
    '["*"]',
    TRUE,
    NOW() - INTERVAL '2 years',
    NOW()
);

-- Mike Rodriguez (Operations Manager)
INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, phone, avatar_url, timezone, permissions, email_verified, created_at, updated_at)
VALUES (
    '33333333-3333-3333-3333-333333333333'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'mike@landscapepro.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY.MzkwLlYXqsIa', -- mike2024
    'Mike',
    'Rodriguez',
    'admin',
    'active',
    '+1-555-0002',
    'https://landscapepro.com/avatars/mike.jpg',
    'America/New_York',
    '["job_manage", "customer_manage", "crew_manage", "equipment_manage", "report_view"]',
    TRUE,
    NOW() - INTERVAL '18 months',
    NOW()
);

-- Jennifer Chen (Sales Manager)
INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, phone, avatar_url, timezone, permissions, email_verified, created_at, updated_at)
VALUES (
    '44444444-4444-4444-4444-444444444444'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'jen@landscapepro.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY.MzkwLlYXqsIa', -- jen2024
    'Jennifer',
    'Chen',
    'admin',
    'active',
    '+1-555-0003',
    'https://landscapepro.com/avatars/jen.jpg',
    'America/New_York',
    '["quote_manage", "customer_manage", "invoice_manage", "report_view"]',
    TRUE,
    NOW() - INTERVAL '1 year',
    NOW()
);

-- David Thompson (Crew Lead)
INSERT INTO users (id, tenant_id, email, password_hash, first_name, last_name, role, status, phone, avatar_url, timezone, permissions, email_verified, created_at, updated_at)
VALUES (
    '55555555-5555-5555-5555-555555555555'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'david@landscapepro.com',
    '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewY.MzkwLlYXqsIa', -- david2024
    'David',
    'Thompson',
    'crew_lead',
    'active',
    '+1-555-0004',
    'https://landscapepro.com/avatars/david.jpg',
    'America/New_York',
    '["job_manage", "crew_manage", "equipment_view"]',
    TRUE,
    NOW() - INTERVAL '8 months',
    NOW()
);

-- ===== SERVICES OFFERED =====
INSERT INTO services (id, tenant_id, name, description, category, base_price, unit, duration_minutes, status, created_at, updated_at)
VALUES 
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Lawn Mowing', 'Professional lawn mowing service', 'maintenance', 45.00, 'per_visit', 45, 'active', NOW(), NOW()),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Hedge Trimming', 'Precision hedge and shrub trimming', 'maintenance', 35.00, 'per_hour', 60, 'active', NOW(), NOW()),
    ('cccccccc-cccc-cccc-cccc-cccccccccccc'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Leaf Cleanup', 'Seasonal leaf removal and cleanup', 'seasonal', 65.00, 'per_visit', 90, 'active', NOW(), NOW()),
    ('dddddddd-dddd-dddd-dddd-dddddddddddd'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Fertilization', 'Professional lawn fertilization', 'treatment', 85.00, 'per_application', 30, 'active', NOW(), NOW()),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Landscaping Design', 'Custom landscaping design and installation', 'design', 150.00, 'per_hour', 180, 'active', NOW(), NOW()),
    ('ffffffff-ffff-ffff-ffff-ffffffffffff'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Snow Removal', 'Emergency snow removal service', 'seasonal', 120.00, 'per_visit', 60, 'active', NOW(), NOW());

-- ===== RESIDENTIAL CUSTOMERS =====

-- John Smith (Suburban homeowner)
INSERT INTO customers (id, tenant_id, first_name, last_name, email, phone, company_name, address_line1, city, state, zip_code, customer_type, preferred_contact_method, lead_source, payment_terms, status, created_at, updated_at)
VALUES (
    '66666666-6666-6666-6666-666666666666'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'John',
    'Smith',
    'john.smith@email.com',
    '+1-555-1001',
    NULL,
    '145 Maple Street',
    'Westchester',
    'NY',
    '10601',
    'residential',
    'email',
    'website',
    30,
    'active',
    NOW() - INTERVAL '8 months',
    NOW()
);

-- Property for John Smith (Medium suburban home)
INSERT INTO properties (id, tenant_id, customer_id, name, address_line1, city, state, zip_code, property_type, latitude, longitude, square_footage, lot_size, access_instructions, special_instructions, created_at, updated_at)
SELECT 
    'p6666666-6666-6666-6666-666666666666'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    '66666666-6666-6666-6666-666666666666'::uuid,
    'Smith Residence',
    '145 Maple Street',
    'Westchester',
    'NY',
    '10601',
    'residential',
    lat,
    lng,
    3500,
    0.25,
    'Ring doorbell, key under mat if no answer',
    'Dog in backyard - please close gate',
    NOW() - INTERVAL '8 months',
    NOW()
FROM generate_coords(40.7128, -74.0060, 15);

-- Lisa Johnson (Luxury home)
INSERT INTO customers (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip_code, customer_type, preferred_contact_method, lead_source, payment_terms, status, created_at, updated_at)
VALUES (
    '77777777-7777-7777-7777-777777777777'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'Lisa',
    'Johnson',
    'lisa.johnson@email.com',
    '+1-555-1002',
    '892 Estate Drive',
    'Greenwich',
    'CT',
    '06830',
    'residential',
    'phone',
    'referral',
    15,
    'active',
    NOW() - INTERVAL '1 year',
    NOW()
);

-- Property for Lisa Johnson (Large luxury property)
INSERT INTO properties (id, tenant_id, customer_id, name, address_line1, city, state, zip_code, property_type, latitude, longitude, square_footage, lot_size, access_instructions, special_instructions, created_at, updated_at)
SELECT 
    'p7777777-7777-7777-7777-777777777777'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    '77777777-7777-7777-7777-777777777777'::uuid,
    'Johnson Estate',
    '892 Estate Drive',
    'Greenwich',
    'CT',
    '06830',
    'residential',
    lat,
    lng,
    15000,
    2.5,
    'Gate code 4578, intercom to main house',
    'Formal gardens - premium care required. Prize-winning roses.',
    NOW() - INTERVAL '1 year',
    NOW()
FROM generate_coords(40.7128, -74.0060, 25);

-- Robert Davis (Townhouse)
INSERT INTO customers (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip_code, customer_type, preferred_contact_method, lead_source, payment_terms, status, created_at, updated_at)
VALUES (
    '88888888-8888-8888-8888-888888888888'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'Robert',
    'Davis',
    'robert.davis@email.com',
    '+1-555-1003',
    '67 Townhouse Lane, Unit 12',
    'Stamford',
    'CT',
    '06901',
    'residential',
    'email',
    'google_ads',
    30,
    'active',
    NOW() - INTERVAL '4 months',
    NOW()
);

-- Property for Robert Davis (Small townhouse yard)
INSERT INTO properties (id, tenant_id, customer_id, name, address_line1, city, state, zip_code, property_type, latitude, longitude, square_footage, lot_size, access_instructions, special_instructions, created_at, updated_at)
SELECT 
    'p8888888-8888-8888-8888-888888888888'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    '88888888-8888-8888-8888-888888888888'::uuid,
    'Davis Townhouse',
    '67 Townhouse Lane, Unit 12',
    'Stamford',
    'CT',
    '06901',
    'residential',
    lat,
    lng,
    800,
    0.05,
    'Park in designated visitor spots only',
    'Small front and back areas, quick service',
    NOW() - INTERVAL '4 months',
    NOW()
FROM generate_coords(40.7128, -74.0060, 20);

-- Maria Garcia (Small yard)
INSERT INTO customers (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip_code, customer_type, preferred_contact_method, lead_source, payment_terms, status, created_at, updated_at)
VALUES (
    '99999999-9999-9999-9999-999999999999'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'Maria',
    'Garcia',
    'maria.garcia@email.com',
    '+1-555-1004',
    '234 Pine Avenue',
    'Yonkers',
    'NY',
    '10701',
    'residential',
    'text',
    'neighbor_referral',
    30,
    'active',
    NOW() - INTERVAL '6 months',
    NOW()
);

-- Property for Maria Garcia (Small property)
INSERT INTO properties (id, tenant_id, customer_id, name, address_line1, city, state, zip_code, property_type, latitude, longitude, square_footage, lot_size, access_instructions, special_instructions, created_at, updated_at)
SELECT 
    'p9999999-9999-9999-9999-999999999999'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    '99999999-9999-9999-9999-999999999999'::uuid,
    'Garcia Home',
    '234 Pine Avenue',
    'Yonkers',
    'NY',
    '10701',
    'residential',
    lat,
    lng,
    1200,
    0.1,
    'Side gate access, spare key with neighbor at 236',
    'Elderly homeowner - gentle service',
    NOW() - INTERVAL '6 months',
    NOW()
FROM generate_coords(40.7128, -74.0060, 12);

-- Tom Wilson (Large property)
INSERT INTO customers (id, tenant_id, first_name, last_name, email, phone, address_line1, city, state, zip_code, customer_type, preferred_contact_method, lead_source, payment_terms, status, created_at, updated_at)
VALUES (
    'aaaaaaaa-1111-1111-1111-111111111111'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'Tom',
    'Wilson',
    'tom.wilson@email.com',
    '+1-555-1005',
    '1847 Country Club Road',
    'Rye',
    'NY',
    '10580',
    'residential',
    'phone',
    'country_club_referral',
    15,
    'active',
    NOW() - INTERVAL '10 months',
    NOW()
);

-- Property for Tom Wilson (Large estate)
INSERT INTO properties (id, tenant_id, customer_id, name, address_line1, city, state, zip_code, property_type, latitude, longitude, square_footage, lot_size, access_instructions, special_instructions, created_at, updated_at)
SELECT 
    'paaaaaaaa-1111-1111-1111-111111111111'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'aaaaaaaa-1111-1111-1111-111111111111'::uuid,
    'Wilson Estate',
    '1847 Country Club Road',
    'Rye',
    'NY',
    '10580',
    'residential',
    lat,
    lng,
    8500,
    1.8,
    'Main gate code 7890, service entrance on left',
    'Multiple zones: front lawn, back gardens, pool area, tennis court',
    NOW() - INTERVAL '10 months',
    NOW()
FROM generate_coords(40.7128, -74.0060, 18);

-- ===== COMMERCIAL CUSTOMERS =====

-- Amanda Foster (Office complex manager)
INSERT INTO customers (id, tenant_id, first_name, last_name, email, phone, company_name, address_line1, city, state, zip_code, customer_type, preferred_contact_method, lead_source, credit_limit, payment_terms, status, created_at, updated_at)
VALUES (
    'bbbbbbbb-2222-2222-2222-222222222222'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'Amanda',
    'Foster',
    'amanda@midtownoffice.com',
    '+1-555-2001',
    'Midtown Office Complex',
    '500 Business Plaza',
    'White Plains',
    'NY',
    '10601',
    'commercial',
    'email',
    'linkedin',
    10000.00,
    30,
    'active',
    NOW() - INTERVAL '2 years',
    NOW()
);

-- Property for Midtown Office Complex
INSERT INTO properties (id, tenant_id, customer_id, name, address_line1, city, state, zip_code, property_type, latitude, longitude, square_footage, lot_size, access_instructions, special_instructions, created_at, updated_at)
SELECT 
    'pbbbbbbbb-2222-2222-2222-222222222222'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'bbbbbbbb-2222-2222-2222-222222222222'::uuid,
    'Midtown Office Complex',
    '500 Business Plaza',
    'White Plains',
    'NY',
    '10601',
    'commercial',
    lat,
    lng,
    25000,
    5.0,
    'Check in with security desk, service after 6PM preferred',
    'Corporate campus with multiple buildings, parking areas, landscaped courtyards',
    NOW() - INTERVAL '2 years',
    NOW()
FROM generate_coords(40.7128, -74.0060, 22);

-- Steve Park (Restaurant owner)
INSERT INTO customers (id, tenant_id, first_name, last_name, email, phone, company_name, address_line1, city, state, zip_code, customer_type, preferred_contact_method, lead_source, credit_limit, payment_terms, status, created_at, updated_at)
VALUES (
    'cccccccc-3333-3333-3333-333333333333'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'Steve',
    'Park',
    'steve@greencafe.com',
    '+1-555-2002',
    'Green Cafe Restaurant',
    '78 Main Street',
    'New Rochelle',
    'NY',
    '10801',
    'commercial',
    'phone',
    'google_search',
    5000.00,
    15,
    'active',
    NOW() - INTERVAL '1 year',
    NOW()
);

-- Property for Green Cafe
INSERT INTO properties (id, tenant_id, customer_id, name, address_line1, city, state, zip_code, property_type, latitude, longitude, square_footage, lot_size, access_instructions, special_instructions, created_at, updated_at)
SELECT 
    'pcccccccc-3333-3333-3333-333333333333'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'cccccccc-3333-3333-3333-333333333333'::uuid,
    'Green Cafe Outdoor Dining',
    '78 Main Street',
    'New Rochelle',
    'NY',
    '10801',
    'commercial',
    lat,
    lng,
    2500,
    0.3,
    'Service entrance in rear alley, avoid lunch hours 11AM-2PM',
    'Outdoor dining patio with planters, herbs garden, seasonal decorations required',
    NOW() - INTERVAL '1 year',
    NOW()
FROM generate_coords(40.7128, -74.0060, 16);

-- Rachel Kim (HOA manager)
INSERT INTO customers (id, tenant_id, first_name, last_name, email, phone, company_name, address_line1, city, state, zip_code, customer_type, preferred_contact_method, lead_source, credit_limit, payment_terms, status, created_at, updated_at)
VALUES (
    'dddddddd-4444-4444-4444-444444444444'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'Rachel',
    'Kim',
    'rachel@willowcreek-hoa.com',
    '+1-555-2003',
    'Willow Creek HOA',
    '100 Community Circle',
    'Scarsdale',
    'NY',
    '10583',
    'commercial',
    'email',
    'property_manager_referral',
    25000.00,
    30,
    'active',
    NOW() - INTERVAL '3 years',
    NOW()
);

-- Property for Willow Creek HOA (Common areas)
INSERT INTO properties (id, tenant_id, customer_id, name, address_line1, city, state, zip_code, property_type, latitude, longitude, square_footage, lot_size, access_instructions, special_instructions, created_at, updated_at)
SELECT 
    'pdddddddd-4444-4444-4444-444444444444'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'dddddddd-4444-4444-4444-444444444444'::uuid,
    'Willow Creek Common Areas',
    '100 Community Circle',
    'Scarsdale',
    'NY',
    '10583',
    'commercial',
    lat,
    lng,
    50000,
    15.0,
    'Management office key required, coordinate with residents',
    'Multiple common areas: entrance landscaping, playground, walking paths, community garden, retention pond',
    NOW() - INTERVAL '3 years',
    NOW()
FROM generate_coords(40.7128, -74.0060, 19);

-- ===== SAMPLE QUOTES AND JOBS =====

-- Quote for John Smith (Basic lawn care)
INSERT INTO quotes (id, tenant_id, customer_id, property_id, quote_number, title, description, subtotal, tax_rate, tax_amount, total_amount, status, valid_until, terms_and_conditions, created_by, created_at, updated_at)
VALUES (
    'q6666666-6666-6666-6666-666666666666'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    '66666666-6666-6666-6666-666666666666'::uuid,
    'p6666666-6666-6666-6666-666666666666'::uuid,
    generate_quote_number('11111111-1111-1111-1111-111111111111'::uuid),
    'Weekly Lawn Maintenance - Smith Residence',
    'Weekly lawn mowing and basic maintenance for suburban property',
    95.00,
    0.0875,
    8.31,
    103.31,
    'accepted',
    NOW() + INTERVAL '30 days',
    'Payment due within 15 days of service completion. Weather delays may apply.',
    '44444444-4444-4444-4444-444444444444'::uuid,
    NOW() - INTERVAL '3 weeks',
    NOW() - INTERVAL '2 weeks'
);

-- Quote services for John Smith
INSERT INTO quote_services (id, quote_id, service_id, quantity, unit_price, total_price, description, created_at)
VALUES 
    ('qs666666-6666-6666-6666-666666666666'::uuid, 'q6666666-6666-6666-6666-666666666666'::uuid, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid, 1, 45.00, 45.00, 'Lawn mowing service', NOW()),
    ('qs666667-6666-6666-6666-666666666666'::uuid, 'q6666666-6666-6666-6666-666666666666'::uuid, 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid, 1, 50.00, 50.00, 'Hedge trimming and edging', NOW());

-- Recurring job for John Smith
INSERT INTO jobs (id, tenant_id, customer_id, property_id, assigned_user_id, title, description, status, priority, scheduled_date, scheduled_time, estimated_duration, total_amount, job_number, recurring_schedule, weather_dependent, crew_size, created_at, updated_at)
VALUES (
    'j6666666-6666-6666-6666-666666666666'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    '66666666-6666-6666-6666-666666666666'::uuid,
    'p6666666-6666-6666-6666-666666666666'::uuid,
    '55555555-5555-5555-5555-555555555555'::uuid,
    'Weekly Lawn Maintenance - Smith',
    'Standard weekly lawn mowing and maintenance',
    'scheduled',
    'medium',
    CURRENT_DATE + 2,
    '09:00:00',
    60,
    103.31,
    generate_job_number('11111111-1111-1111-1111-111111111111'::uuid),
    'weekly',
    TRUE,
    2,
    NOW() - INTERVAL '2 weeks',
    NOW()
);

-- Large property quote for Lisa Johnson (Premium service)
INSERT INTO quotes (id, tenant_id, customer_id, property_id, quote_number, title, description, subtotal, tax_rate, tax_amount, total_amount, status, valid_until, terms_and_conditions, created_by, created_at, updated_at)
VALUES (
    'q7777777-7777-7777-7777-777777777777'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    '77777777-7777-7777-7777-777777777777'::uuid,
    'p7777777-7777-7777-7777-777777777777'::uuid,
    generate_quote_number('11111111-1111-1111-1111-111111111111'::uuid),
    'Premium Estate Maintenance - Johnson Estate',
    'Comprehensive estate maintenance including formal gardens, lawn care, and seasonal services',
    425.00,
    0.0875,
    37.19,
    462.19,
    'pending',
    NOW() + INTERVAL '45 days',
    'Premium service package with guaranteed same-day response. Net 15 payment terms.',
    '44444444-4444-4444-4444-444444444444'::uuid,
    NOW() - INTERVAL '1 week',
    NOW() - INTERVAL '1 week'
);

-- Quote services for Lisa Johnson
INSERT INTO quote_services (id, quote_id, service_id, quantity, unit_price, total_price, description, created_at)
VALUES 
    ('qs777777-7777-7777-7777-777777777777'::uuid, 'q7777777-7777-7777-7777-777777777777'::uuid, 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa'::uuid, 3, 60.00, 180.00, 'Large property lawn mowing (premium rate)', NOW()),
    ('qs777778-7777-7777-7777-777777777777'::uuid, 'q7777777-7777-7777-7777-777777777777'::uuid, 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb'::uuid, 4, 35.00, 140.00, 'Formal garden hedge maintenance', NOW()),
    ('qs777779-7777-7777-7777-777777777777'::uuid, 'q7777777-7777-7777-7777-777777777777'::uuid, 'eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee'::uuid, 0.7, 150.00, 105.00, 'Landscape design consultation', NOW());

-- Commercial job for Amanda Foster
INSERT INTO jobs (id, tenant_id, customer_id, property_id, assigned_user_id, title, description, status, priority, scheduled_date, scheduled_time, estimated_duration, total_amount, job_number, weather_dependent, crew_size, created_at, updated_at)
VALUES (
    'jbbbbbb-2222-2222-2222-222222222222'::uuid,
    '11111111-1111-1111-1111-111111111111'::uuid,
    'bbbbbbbb-2222-2222-2222-222222222222'::uuid,
    'pbbbbbbbb-2222-2222-2222-222222222222'::uuid,
    '33333333-3333-3333-3333-333333333333'::uuid,
    'Weekly Commercial Grounds Maintenance',
    'Complete grounds maintenance for office complex including all landscaped areas',
    'in_progress',
    'high',
    CURRENT_DATE,
    '06:00:00',
    240,
    875.00,
    generate_job_number('11111111-1111-1111-1111-111111111111'::uuid),
    FALSE,
    4,
    NOW() - INTERVAL '1 day',
    NOW()
);

-- Equipment for the company
INSERT INTO equipment (id, tenant_id, name, type, model, serial_number, purchase_date, purchase_price, status, maintenance_schedule, last_maintenance, next_maintenance, notes, created_at, updated_at)
VALUES 
    ('e1111111-1111-1111-1111-111111111111'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'John Deere Z535M', 'Zero-Turn Mower', 'Z535M', 'JD2024001', '2024-03-15', 4200.00, 'active', '90 days', '2024-12-01', '2025-03-01', 'Primary mower for large properties', NOW(), NOW()),
    ('e2222222-2222-2222-2222-222222222222'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'STIHL FS 131', 'String Trimmer', 'FS 131', 'ST2024002', '2024-04-10', 450.00, 'active', '30 days', '2024-12-15', '2025-01-15', 'Heavy-duty trimmer for commercial work', NOW(), NOW()),
    ('e3333333-3333-3333-3333-333333333333'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Echo PB-580T', 'Leaf Blower', 'PB-580T', 'EC2024003', '2024-02-20', 380.00, 'active', '45 days', '2024-11-20', '2025-01-05', 'High-performance backpack blower', NOW(), NOW());

-- Create crews
INSERT INTO crews (id, tenant_id, name, description, capacity, specializations, equipment_ids, status, created_at, updated_at)
VALUES 
    ('c1111111-1111-1111-1111-111111111111'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Team Alpha', 'Primary residential maintenance crew', 3, '["lawn_care", "hedge_trimming", "cleanup"]', '["e1111111-1111-1111-1111-111111111111", "e2222222-2222-2222-2222-222222222222"]', 'active', NOW(), NOW()),
    ('c2222222-2222-2222-2222-222222222222'::uuid, '11111111-1111-1111-1111-111111111111'::uuid, 'Team Bravo', 'Commercial and large property specialists', 4, '["commercial_maintenance", "landscaping", "snow_removal"]', '["e3333333-3333-3333-3333-333333333333"]', 'active', NOW(), NOW());

-- Crew assignments
INSERT INTO crew_members (id, crew_id, user_id, role, joined_at)
VALUES 
    ('cm111111-1111-1111-1111-111111111111'::uuid, 'c1111111-1111-1111-1111-111111111111'::uuid, '55555555-5555-5555-5555-555555555555'::uuid, 'lead', NOW() - INTERVAL '6 months'),
    ('cm222222-2222-2222-2222-222222222222'::uuid, 'c2222222-2222-2222-2222-222222222222'::uuid, '33333333-3333-3333-3333-333333333333'::uuid, 'lead', NOW() - INTERVAL '1 year');

-- Drop the helper function
DROP FUNCTION IF EXISTS generate_coords(DECIMAL, DECIMAL, INTEGER);

COMMIT;

-- Verification queries (commented out for actual use)
/*
SELECT 'Tenants' as table_name, count(*) as count FROM tenants
UNION ALL
SELECT 'Users' as table_name, count(*) as count FROM users
UNION ALL
SELECT 'Customers' as table_name, count(*) as count FROM customers
UNION ALL
SELECT 'Properties' as table_name, count(*) as count FROM properties
UNION ALL
SELECT 'Services' as table_name, count(*) as count FROM services
UNION ALL
SELECT 'Quotes' as table_name, count(*) as count FROM quotes
UNION ALL
SELECT 'Jobs' as table_name, count(*) as count FROM jobs
UNION ALL
SELECT 'Equipment' as table_name, count(*) as count FROM equipment
UNION ALL
SELECT 'Crews' as table_name, count(*) as count FROM crews;

-- View customer distribution
SELECT 
    c.customer_type,
    count(*) as customer_count,
    avg(p.square_footage) as avg_sq_ft,
    min(p.square_footage) as min_sq_ft,
    max(p.square_footage) as max_sq_ft
FROM customers c 
JOIN properties p ON c.id = p.customer_id 
GROUP BY c.customer_type;
*/