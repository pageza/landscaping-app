const { test, expect } = require('@playwright/test');
const { execSync } = require('child_process');

/**
 * Comprehensive Customer Journey Test Suite
 * Tests the complete end-to-end customer experience from signup to service booking
 */

test.describe('Comprehensive Customer Journey Testing', () => {
  let customerData;
  
  test.beforeAll(async () => {
    // Load realistic test data into database
    try {
      console.log('Loading realistic test data...');
      execSync('psql -d landscaping_app_test -f backend/tests/realistic_test_data.sql', { 
        cwd: process.cwd(),
        stdio: 'pipe'
      });
      console.log('Test data loaded successfully');
    } catch (error) {
      console.warn('Could not load test data automatically, ensure database is populated');
    }
    
    // Test customer data for journey scenarios
    customerData = {
      newCustomer: {
        firstName: 'Jennifer',
        lastName: 'Martinez',
        email: 'jennifer.martinez.test@email.com',
        phone: '+1-555-9999',
        password: 'TestPass123!',
        address: '789 Test Lane',
        city: 'Mount Vernon',
        state: 'NY',
        zipCode: '10550',
        propertyType: 'residential',
        squareFootage: '2800'
      },
      existingCustomer: {
        email: 'john.smith@email.com',
        password: 'john123'
      }
    };
  });

  test.describe('Scenario A: New Customer Journey', () => {
    test('should complete full customer onboarding journey', async ({ page }) => {
      console.log('Starting new customer onboarding journey...');
      
      // Step 1: Navigate to landing page
      await page.goto('/');
      await expect(page).toHaveTitle(/LandscapePro/);
      
      // Step 2: Customer signup
      await page.click('text=Get Started');
      await expect(page).toHaveURL('**/signup');
      
      // Fill customer registration form
      await page.fill('input[name="firstName"]', customerData.newCustomer.firstName);
      await page.fill('input[name="lastName"]', customerData.newCustomer.lastName);
      await page.fill('input[name="email"]', customerData.newCustomer.email);
      await page.fill('input[name="phone"]', customerData.newCustomer.phone);
      await page.fill('input[name="password"]', customerData.newCustomer.password);
      await page.fill('input[name="confirmPassword"]', customerData.newCustomer.password);
      
      // Accept terms and submit
      await page.check('input[name="acceptTerms"]');
      await page.click('button[type="submit"]');
      
      // Should redirect to property details
      await expect(page).toHaveURL('**/property-setup');
      await expect(page.locator('[data-testid="welcome-message"]'))
        .toContainText('Welcome Jennifer');
      
      // Step 3: Property details entry
      await page.fill('input[name="address"]', customerData.newCustomer.address);
      await page.fill('input[name="city"]', customerData.newCustomer.city);
      await page.selectOption('select[name="state"]', customerData.newCustomer.state);
      await page.fill('input[name="zipCode"]', customerData.newCustomer.zipCode);
      await page.selectOption('select[name="propertyType"]', customerData.newCustomer.propertyType);
      await page.fill('input[name="squareFootage"]', customerData.newCustomer.squareFootage);
      
      // Add special instructions
      await page.fill('textarea[name="specialInstructions"]', 
        'New homeowner, please be gentle with newly planted shrubs');
      
      await page.click('[data-testid="save-property-button"]');
      
      // Should redirect to service selection
      await expect(page).toHaveURL('**/services');
      
      // Step 4: Service selection and pricing
      // Select lawn mowing service
      await page.check('input[data-service="lawn-mowing"]');
      await expect(page.locator('[data-testid="service-price-lawn-mowing"]'))
        .toContainText('$45.00');
      
      // Select hedge trimming
      await page.check('input[data-service="hedge-trimming"]');
      await expect(page.locator('[data-testid="service-price-hedge-trimming"]'))
        .toContainText('$35.00');
      
      // Verify total calculation (should be around $95-120 for medium property)
      const totalText = await page.locator('[data-testid="total-price"]').textContent();
      const total = parseFloat(totalText.replace(/[^0-9.]/g, ''));
      expect(total).toBeGreaterThan(90);
      expect(total).toBeLessThan(130);
      
      await page.click('[data-testid="continue-to-booking"]');
      
      // Step 5: Booking with discount scenarios
      await expect(page).toHaveURL('**/booking');
      
      // Select date (tomorrow for same-day discount test)
      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      const tomorrowStr = tomorrow.toISOString().split('T')[0];
      
      await page.fill('input[name="scheduledDate"]', tomorrowStr);
      await page.selectOption('select[name="scheduledTime"]', '09:00');
      
      // Check for same-day booking discount message
      await expect(page.locator('[data-testid="discount-message"]'))
        .toContainText(/discount|reduced rate/i);
      
      // Step 6: Schedule confirmation
      await page.fill('textarea[name="notes"]', 
        'First time customer - please call when arriving');
      
      await page.click('[data-testid="confirm-booking"]');
      
      // Should show confirmation
      await expect(page).toHaveURL('**/booking-confirmed');
      await expect(page.locator('[data-testid="confirmation-message"]'))
        .toContainText('Booking confirmed');
      
      // Should display booking number
      await expect(page.locator('[data-testid="booking-number"]')).toBeVisible();
      
      // Should show estimated arrival time
      await expect(page.locator('[data-testid="estimated-arrival"]')).toBeVisible();
      
      console.log('New customer journey completed successfully!');
    });
    
    test('should handle validation errors gracefully', async ({ page }) => {
      await page.goto('/signup');
      
      // Try to submit empty form
      await page.click('button[type="submit"]');
      
      // Should show validation errors
      await expect(page.locator('[data-testid="firstName-error"]'))
        .toContainText('First name is required');
      await expect(page.locator('[data-testid="email-error"]'))
        .toContainText('Email is required');
      
      // Test invalid email format
      await page.fill('input[name="email"]', 'invalid-email');
      await page.click('button[type="submit"]');
      await expect(page.locator('[data-testid="email-error"]'))
        .toContainText('valid email');
    });
  });

  test.describe('Scenario B: Pricing Variations Testing', () => {
    test('should show correct pricing for small property lawn care', async ({ page }) => {
      await page.goto('/quote-calculator');
      
      // Small property (800 sq ft - should be ~$90-120)
      await page.fill('input[name="squareFootage"]', '800');
      await page.selectOption('select[name="propertyType"]', 'residential');
      await page.check('input[data-service="lawn-mowing"]');
      
      // Trigger pricing calculation
      await page.click('[data-testid="calculate-price"]');
      
      const priceText = await page.locator('[data-testid="calculated-price"]').textContent();
      const price = parseFloat(priceText.replace(/[^0-9.]/g, ''));
      
      expect(price).toBeGreaterThan(90);
      expect(price).toBeLessThan(120);
      
      console.log(`Small property pricing: $${price}`);
    });
    
    test('should show premium pricing for large property full service', async ({ page }) => {
      await page.goto('/quote-calculator');
      
      // Large property (15000 sq ft - should be $300-500+)
      await page.fill('input[name="squareFootage"]', '15000');
      await page.selectOption('select[name="propertyType"]', 'luxury');
      
      // Select multiple services
      await page.check('input[data-service="lawn-mowing"]');
      await page.check('input[data-service="hedge-trimming"]');
      await page.check('input[data-service="fertilization"]');
      await page.check('input[data-service="landscaping-design"]');
      
      await page.click('[data-testid="calculate-price"]');
      
      const priceText = await page.locator('[data-testid="calculated-price"]').textContent();
      const price = parseFloat(priceText.replace(/[^0-9.]/g, ''));
      
      expect(price).toBeGreaterThan(300);
      expect(price).toBeLessThan(600);
      
      // Should show premium service badge
      await expect(page.locator('[data-testid="premium-service-badge"]')).toBeVisible();
      
      console.log(`Large property full service pricing: $${price}`);
    });
    
    test('should apply same-day booking discount', async ({ page }) => {
      // Login as existing customer
      await page.goto('/login');
      await page.fill('input[name="email"]', customerData.existingCustomer.email);
      await page.fill('input[name="password"]', customerData.existingCustomer.password);
      await page.click('button[type="submit"]');
      
      await page.goto('/book-service');
      
      // Get regular price first
      await page.check('input[data-service="lawn-mowing"]');
      const regularPrice = await page.locator('[data-testid="base-price"]').textContent();
      const regular = parseFloat(regularPrice.replace(/[^0-9.]/g, ''));
      
      // Select same-day service (today)
      const today = new Date().toISOString().split('T')[0];
      await page.fill('input[name="scheduledDate"]', today);
      
      // Should show discount
      await expect(page.locator('[data-testid="same-day-discount"]')).toBeVisible();
      
      const discountedText = await page.locator('[data-testid="final-price"]').textContent();
      const discounted = parseFloat(discountedText.replace(/[^0-9.]/g, ''));
      
      expect(discounted).toBeLessThan(regular);
      
      const discountAmount = regular - discounted;
      console.log(`Same-day discount applied: $${discountAmount.toFixed(2)}`);
    });
    
    test('should apply bulk service discount for multiple services', async ({ page }) => {
      await page.goto('/quote-calculator');
      
      await page.fill('input[name="squareFootage"]', '5000');
      
      // Single service price
      await page.check('input[data-service="lawn-mowing"]');
      await page.click('[data-testid="calculate-price"]');
      const singlePrice = parseFloat(
        (await page.locator('[data-testid="calculated-price"]').textContent())
        .replace(/[^0-9.]/g, '')
      );
      
      // Multiple services (should trigger bulk discount)
      await page.check('input[data-service="hedge-trimming"]');
      await page.check('input[data-service="leaf-cleanup"]');
      await page.check('input[data-service="fertilization"]');
      
      await page.click('[data-testid="calculate-price"]');
      
      // Should show bulk discount message
      await expect(page.locator('[data-testid="bulk-discount-message"]'))
        .toContainText(/package discount|bulk savings/i);
      
      const bundlePrice = parseFloat(
        (await page.locator('[data-testid="calculated-price"]').textContent())
        .replace(/[^0-9.]/g, '')
      );
      
      // Bundle should be less than 4x single service
      expect(bundlePrice).toBeLessThan(singlePrice * 4);
      
      console.log(`Bulk discount pricing: $${bundlePrice} vs individual: $${singlePrice * 4}`);
    });
    
    test('should show different pricing for geographic zones', async ({ page }) => {
      await page.goto('/quote-calculator');
      
      // Local zone (within 15 miles) - base pricing
      await page.fill('input[name="zipCode"]', '10601'); // Westchester, NY
      await page.fill('input[name="squareFootage"]', '3000');
      await page.check('input[data-service="lawn-mowing"]');
      await page.click('[data-testid="calculate-price"]');
      
      const localPrice = parseFloat(
        (await page.locator('[data-testid="calculated-price"]').textContent())
        .replace(/[^0-9.]/g, '')
      );
      
      // Extended zone (15-30 miles) - should have travel surcharge
      await page.fill('input[name="zipCode"]', '06830'); // Greenwich, CT
      await page.click('[data-testid="calculate-price"]');
      
      await expect(page.locator('[data-testid="travel-surcharge"]'))
        .toContainText(/travel fee|surcharge/i);
      
      const extendedPrice = parseFloat(
        (await page.locator('[data-testid="calculated-price"]').textContent())
        .replace(/[^0-9.]/g, '')
      );
      
      expect(extendedPrice).toBeGreaterThan(localPrice);
      
      // Outside service area (30+ miles) - should show premium pricing or unavailable
      await page.fill('input[name="zipCode"]', '06902'); // Stamford, CT (far)
      await page.click('[data-testid="calculate-price"]');
      
      // Should either show premium pricing or service unavailable message
      const isServiceAvailable = await page.locator('[data-testid="service-unavailable"]').isVisible();
      const isPremiumPricing = await page.locator('[data-testid="premium-zone-pricing"]').isVisible();
      
      expect(isServiceAvailable || isPremiumPricing).toBeTruthy();
      
      console.log(`Pricing zones - Local: $${localPrice}, Extended: $${extendedPrice}`);
    });
  });

  test.describe('Property Size Based Pricing Validation', () => {
    const propertySizes = [
      { size: 800, expectedRange: [90, 120], description: 'Small townhouse' },
      { size: 2500, expectedRange: [110, 150], description: 'Medium suburban home' },
      { size: 5000, expectedRange: [140, 200], description: 'Large residential' },
      { size: 15000, expectedRange: [300, 500], description: 'Estate property' }
    ];

    propertySizes.forEach(({ size, expectedRange, description }) => {
      test(`should price ${description} (${size} sq ft) correctly`, async ({ page }) => {
        await page.goto('/quote-calculator');
        
        await page.fill('input[name="squareFootage"]', size.toString());
        await page.selectOption('select[name="propertyType"]', 'residential');
        await page.check('input[data-service="lawn-mowing"]');
        await page.check('input[data-service="hedge-trimming"]');
        
        await page.click('[data-testid="calculate-price"]');
        
        const priceText = await page.locator('[data-testid="calculated-price"]').textContent();
        const price = parseFloat(priceText.replace(/[^0-9.]/g, ''));
        
        expect(price).toBeGreaterThan(expectedRange[0]);
        expect(price).toBeLessThan(expectedRange[1]);
        
        console.log(`${description}: ${size} sq ft = $${price} (expected: $${expectedRange[0]}-${expectedRange[1]})`);
      });
    });
  });

  test.describe('Customer Portal Functionality', () => {
    test('should allow customers to view service history', async ({ page }) => {
      // Login as existing customer
      await page.goto('/customer-login');
      await page.fill('input[name="email"]', 'john.smith@email.com');
      await page.fill('input[name="password"]', 'john123');
      await page.click('button[type="submit"]');
      
      await expect(page).toHaveURL('**/customer-dashboard');
      
      // Navigate to service history
      await page.click('[data-testid="service-history-tab"]');
      
      // Should show historical services
      await expect(page.locator('[data-testid="service-history-table"]')).toBeVisible();
      await expect(page.locator('[data-testid="service-entry"]').first()).toBeVisible();
      
      // Should show service details when clicked
      await page.click('[data-testid="service-entry"]');
      await expect(page.locator('[data-testid="service-details-modal"]')).toBeVisible();
    });
    
    test('should allow customers to schedule recurring services', async ({ page }) => {
      await page.goto('/customer-login');
      await page.fill('input[name="email"]', 'john.smith@email.com');
      await page.fill('input[name="password"]', 'john123');
      await page.click('button[type="submit"]');
      
      await page.click('[data-testid="schedule-service-button"]');
      
      // Set up recurring service
      await page.check('input[data-service="lawn-mowing"]');
      await page.selectOption('select[name="frequency"]', 'weekly');
      
      // Select start date
      const nextWeek = new Date();
      nextWeek.setDate(nextWeek.getDate() + 7);
      await page.fill('input[name="startDate"]', nextWeek.toISOString().split('T')[0]);
      
      await page.click('[data-testid="schedule-recurring-button"]');
      
      await expect(page.locator('[data-testid="recurring-confirmation"]'))
        .toContainText('Recurring service scheduled');
    });
    
    test('should display accurate pricing for returning customers', async ({ page }) => {
      // Login as customer with history
      await page.goto('/customer-login');
      await page.fill('input[name="email"]', 'lisa.johnson@email.com');
      await page.fill('input[name="password"]', 'lisa123');
      await page.click('button[type="submit"]');
      
      // Should show loyalty discount for repeat customers
      await page.click('[data-testid="book-service-button"]');
      
      await expect(page.locator('[data-testid="loyalty-discount"]'))
        .toContainText(/loyal customer|returning customer|discount/i);
      
      // Verify pricing shows loyalty discount
      const originalPrice = await page.locator('[data-testid="base-price"]').textContent();
      const loyaltyPrice = await page.locator('[data-testid="loyalty-price"]').textContent();
      
      const original = parseFloat(originalPrice.replace(/[^0-9.]/g, ''));
      const loyalty = parseFloat(loyaltyPrice.replace(/[^0-9.]/g, ''));
      
      expect(loyalty).toBeLessThan(original);
    });
  });

  test.describe('Error Handling and Edge Cases', () => {
    test('should handle invalid zip codes gracefully', async ({ page }) => {
      await page.goto('/quote-calculator');
      
      await page.fill('input[name="zipCode"]', '00000'); // Invalid zip
      await page.fill('input[name="squareFootage"]', '2000');
      await page.check('input[data-service="lawn-mowing"]');
      
      await page.click('[data-testid="calculate-price"]');
      
      await expect(page.locator('[data-testid="invalid-zip-message"]'))
        .toContainText(/invalid zip code|service area/i);
    });
    
    test('should handle extremely large property sizes', async ({ page }) => {
      await page.goto('/quote-calculator');
      
      await page.fill('input[name="squareFootage"]', '100000'); // 100k sq ft
      await page.check('input[data-service="lawn-mowing"]');
      
      await page.click('[data-testid="calculate-price"]');
      
      // Should either show custom quote message or handle gracefully
      const hasCustomQuoteMessage = await page.locator('[data-testid="custom-quote-required"]').isVisible();
      const hasPrice = await page.locator('[data-testid="calculated-price"]').isVisible();
      
      expect(hasCustomQuoteMessage || hasPrice).toBeTruthy();
    });
    
    test('should handle network connectivity issues', async ({ page }) => {
      // Simulate network failure
      await page.route('**/api/quotes/**', route => route.abort());
      
      await page.goto('/quote-calculator');
      await page.fill('input[name="squareFootage"]', '2000');
      await page.check('input[data-service="lawn-mowing"]');
      await page.click('[data-testid="calculate-price"]');
      
      // Should show error message
      await expect(page.locator('[data-testid="network-error"]'))
        .toContainText(/network error|connection problem/i);
      
      // Should show retry option
      await expect(page.locator('[data-testid="retry-button"]')).toBeVisible();
    });
  });
});

/**
 * Helper functions for test data validation
 */
function validatePriceRange(price, min, max, description) {
  if (price < min || price > max) {
    throw new Error(`${description} price $${price} is outside expected range $${min}-$${max}`);
  }
  return true;
}

function calculateExpectedDiscount(originalPrice, discountType) {
  const discounts = {
    'same-day': 0.10, // 10%
    'loyalty': 0.15,  // 15%
    'bulk': 0.20,     // 20%
    'first-time': 0.05 // 5%
  };
  
  return originalPrice * (1 - (discounts[discountType] || 0));
}