const { test, expect } = require('@playwright/test');

test.describe('Authentication Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the application
    await page.goto('/');
  });

  test('should display login form', async ({ page }) => {
    await expect(page).toHaveTitle(/Landscaping App/);
    await expect(page.locator('h1')).toContainText('Sign In');
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('should login successfully with valid credentials', async ({ page }) => {
    // Fill in login form
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'password123');
    
    // Click login button
    await page.click('button[type="submit"]');
    
    // Wait for navigation to dashboard
    await page.waitForURL('**/dashboard');
    
    // Verify successful login
    await expect(page.locator('h1')).toContainText('Dashboard');
    await expect(page.locator('[data-testid="user-menu"]')).toBeVisible();
    await expect(page.locator('[data-testid="user-name"]')).toContainText('Admin');
  });

  test('should show error message with invalid credentials', async ({ page }) => {
    // Fill in incorrect credentials
    await page.fill('input[name="email"]', 'invalid@example.com');
    await page.fill('input[name="password"]', 'wrongpassword');
    
    // Click login button
    await page.click('button[type="submit"]');
    
    // Wait for error message
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Invalid email or password');
    
    // Should stay on login page
    await expect(page.locator('h1')).toContainText('Sign In');
  });

  test('should validate required fields', async ({ page }) => {
    // Try to submit empty form
    await page.click('button[type="submit"]');
    
    // Check for validation errors
    await expect(page.locator('[data-testid="email-error"]')).toContainText('Email is required');
    await expect(page.locator('[data-testid="password-error"]')).toContainText('Password is required');
  });

  test('should validate email format', async ({ page }) => {
    // Fill invalid email
    await page.fill('input[name="email"]', 'invalid-email');
    await page.fill('input[name="password"]', 'password123');
    
    await page.click('button[type="submit"]');
    
    await expect(page.locator('[data-testid="email-error"]')).toContainText('Please enter a valid email address');
  });

  test('should navigate to registration page', async ({ page }) => {
    await page.click('text=Create an account');
    
    await expect(page).toHaveURL('**/register');
    await expect(page.locator('h1')).toContainText('Create Account');
  });

  test('should handle forgot password flow', async ({ page }) => {
    await page.click('text=Forgot password?');
    
    await expect(page).toHaveURL('**/forgot-password');
    await expect(page.locator('h1')).toContainText('Reset Password');
    
    // Fill email and submit
    await page.fill('input[name="email"]', 'user@example.com');
    await page.click('button[type="submit"]');
    
    await expect(page.locator('[data-testid="success-message"]')).toContainText('Reset link sent');
  });

  test('should logout successfully', async ({ page }) => {
    // Login first
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    await page.waitForURL('**/dashboard');
    
    // Click user menu
    await page.click('[data-testid="user-menu"]');
    
    // Click logout
    await page.click('text=Logout');
    
    // Should redirect to login page
    await expect(page).toHaveURL('**/login');
    await expect(page.locator('h1')).toContainText('Sign In');
  });

  test('should remember login state', async ({ page, context }) => {
    // Login
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    await page.waitForURL('**/dashboard');
    
    // Close current page and open new one
    await page.close();
    const newPage = await context.newPage();
    await newPage.goto('/');
    
    // Should go directly to dashboard (already authenticated)
    await expect(newPage).toHaveURL('**/dashboard');
  });

  test('should handle session expiration', async ({ page }) => {
    // Login
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    await page.waitForURL('**/dashboard');
    
    // Simulate expired session by clearing localStorage
    await page.evaluate(() => {
      localStorage.removeItem('auth_token');
      sessionStorage.clear();
    });
    
    // Navigate to protected route
    await page.goto('/customers');
    
    // Should redirect to login
    await expect(page).toHaveURL('**/login');
    await expect(page.locator('[data-testid="info-message"]')).toContainText('Session expired');
  });

  test('should prevent access to protected routes when not authenticated', async ({ page }) => {
    // Try to access dashboard directly
    await page.goto('/dashboard');
    
    // Should redirect to login
    await expect(page).toHaveURL('**/login');
    
    // Try other protected routes
    const protectedRoutes = ['/customers', '/jobs', '/invoices', '/settings'];
    
    for (const route of protectedRoutes) {
      await page.goto(route);
      await expect(page).toHaveURL('**/login');
    }
  });

  test('should support different user roles', async ({ page }) => {
    // Login as customer
    await page.fill('input[name="email"]', 'customer@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    await page.waitForURL('**/customer-dashboard');
    
    // Verify customer-specific UI
    await expect(page.locator('h1')).toContainText('Your Properties');
    await expect(page.locator('[data-testid="customer-menu"]')).toBeVisible();
    
    // Should not see admin features
    await expect(page.locator('[data-testid="admin-panel"]')).not.toBeVisible();
  });
});

test.describe('Registration Flow', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/register');
  });

  test('should display registration form', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Create Account');
    await expect(page.locator('input[name="firstName"]')).toBeVisible();
    await expect(page.locator('input[name="lastName"]')).toBeVisible();
    await expect(page.locator('input[name="email"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('input[name="confirmPassword"]')).toBeVisible();
    await expect(page.locator('input[name="company"]')).toBeVisible();
  });

  test('should register successfully with valid data', async ({ page }) => {
    // Fill registration form
    await page.fill('input[name="firstName"]', 'John');
    await page.fill('input[name="lastName"]', 'Doe');
    await page.fill('input[name="email"]', 'john.doe.new@example.com');
    await page.fill('input[name="password"]', 'SecurePass123!');
    await page.fill('input[name="confirmPassword"]', 'SecurePass123!');
    await page.fill('input[name="company"]', 'Doe Landscaping');
    
    // Accept terms
    await page.check('input[name="acceptTerms"]');
    
    // Submit form
    await page.click('button[type="submit"]');
    
    // Should show success message
    await expect(page.locator('[data-testid="success-message"]')).toContainText('Account created successfully');
    
    // Should redirect to verification page
    await expect(page).toHaveURL('**/verify-email');
  });

  test('should validate password requirements', async ({ page }) => {
    // Fill form with weak password
    await page.fill('input[name="firstName"]', 'John');
    await page.fill('input[name="lastName"]', 'Doe');
    await page.fill('input[name="email"]', 'john.doe@example.com');
    await page.fill('input[name="password"]', '123');
    await page.fill('input[name="confirmPassword"]', '123');
    
    await page.click('button[type="submit"]');
    
    // Should show password validation errors
    await expect(page.locator('[data-testid="password-error"]')).toContainText('Password must be at least 8 characters');
    await expect(page.locator('[data-testid="password-error"]')).toContainText('Password must contain uppercase letter');
    await expect(page.locator('[data-testid="password-error"]')).toContainText('Password must contain special character');
  });

  test('should validate password confirmation', async ({ page }) => {
    await page.fill('input[name="password"]', 'SecurePass123!');
    await page.fill('input[name="confirmPassword"]', 'DifferentPass123!');
    
    await page.click('button[type="submit"]');
    
    await expect(page.locator('[data-testid="confirm-password-error"]')).toContainText('Passwords do not match');
  });

  test('should validate email uniqueness', async ({ page }) => {
    // Try to register with existing email
    await page.fill('input[name="firstName"]', 'Jane');
    await page.fill('input[name="lastName"]', 'Smith');
    await page.fill('input[name="email"]', 'admin@example.com'); // Existing email
    await page.fill('input[name="password"]', 'SecurePass123!');
    await page.fill('input[name="confirmPassword"]', 'SecurePass123!');
    await page.fill('input[name="company"]', 'Smith Landscaping');
    await page.check('input[name="acceptTerms"]');
    
    await page.click('button[type="submit"]');
    
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Email already exists');
  });

  test('should require terms acceptance', async ({ page }) => {
    // Fill valid form but don't accept terms
    await page.fill('input[name="firstName"]', 'John');
    await page.fill('input[name="lastName"]', 'Doe');
    await page.fill('input[name="email"]', 'john.doe@example.com');
    await page.fill('input[name="password"]', 'SecurePass123!');
    await page.fill('input[name="confirmPassword"]', 'SecurePass123!');
    await page.fill('input[name="company"]', 'Doe Landscaping');
    
    // Don't check terms
    await page.click('button[type="submit"]');
    
    await expect(page.locator('[data-testid="terms-error"]')).toContainText('You must accept the terms and conditions');
  });
});

test.describe('Two-Factor Authentication', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
  });

  test('should prompt for 2FA when enabled', async ({ page }) => {
    // Login with user who has 2FA enabled
    await page.fill('input[name="email"]', '2fa@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    // Should show 2FA verification page
    await expect(page).toHaveURL('**/verify-2fa');
    await expect(page.locator('h1')).toContainText('Two-Factor Authentication');
    await expect(page.locator('input[name="code"]')).toBeVisible();
  });

  test('should verify 2FA code successfully', async ({ page }) => {
    // Go through login to 2FA page
    await page.fill('input[name="email"]', '2fa@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    // Enter valid 2FA code
    await page.fill('input[name="code"]', '123456');
    await page.click('button[type="submit"]');
    
    // Should redirect to dashboard
    await page.waitForURL('**/dashboard');
    await expect(page.locator('h1')).toContainText('Dashboard');
  });

  test('should handle invalid 2FA code', async ({ page }) => {
    // Go through login to 2FA page
    await page.fill('input[name="email"]', '2fa@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    // Enter invalid 2FA code
    await page.fill('input[name="code"]', '999999');
    await page.click('button[type="submit"]');
    
    // Should show error message
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Invalid verification code');
    
    // Should stay on 2FA page
    await expect(page).toHaveURL('**/verify-2fa');
  });

  test('should allow 2FA setup for new users', async ({ page }) => {
    // Login
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    
    await page.waitForURL('**/dashboard');
    
    // Navigate to security settings
    await page.click('[data-testid="user-menu"]');
    await page.click('text=Settings');
    await page.click('text=Security');
    
    // Enable 2FA
    await page.click('[data-testid="enable-2fa-button"]');
    
    // Should show QR code
    await expect(page.locator('[data-testid="qr-code"]')).toBeVisible();
    await expect(page.locator('[data-testid="backup-codes"]')).toBeVisible();
    
    // Enter verification code
    await page.fill('input[name="verificationCode"]', '123456');
    await page.click('button[type="submit"]');
    
    await expect(page.locator('[data-testid="success-message"]')).toContainText('2FA enabled successfully');
  });
});