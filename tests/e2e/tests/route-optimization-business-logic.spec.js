const { test, expect } = require('@playwright/test');

/**
 * Route Optimization and Business Logic Test Suite
 * Tests advanced business logic including seasonal pricing, geographic optimization,
 * emergency services, and intelligent routing
 */

test.describe('Route Optimization and Business Logic Testing', () => {
  let testData;
  
  test.beforeAll(async () => {
    testData = {
      zones: {
        local: { zipCodes: ['10601', '10602', '10603'], multiplier: 1.0 },
        extended: { zipCodes: ['06830', '06831', '10580'], multiplier: 1.15 },
        premium: { zipCodes: ['06901', '06902', '10701'], multiplier: 1.30 }
      },
      seasons: {
        spring: { months: [3, 4, 5], multiplier: 1.20 },
        summer: { months: [6, 7, 8], multiplier: 1.25 },
        fall: { months: [9, 10, 11], multiplier: 1.15 },
        winter: { months: [12, 1, 2], multiplier: 0.90 }
      },
      services: {
        emergency: { multiplier: 1.50, availableHours: '24/7' },
        sameDay: { multiplier: 1.20, cutoffTime: '14:00' },
        recurring: { discount: 0.10, minFrequency: 'weekly' },
        bulk: { discount: 0.15, minServices: 3 }
      }
    };
  });

  test.describe('Scenario D: Route Optimization', () => {
    test('should apply area discount for multiple bookings in same neighborhood', async ({ page }) => {
      console.log('Testing area-based route optimization discounts...');
      
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', 'mike@landscapepro.com');
      await page.fill('input[name="password"]', 'mike2024');
      await page.click('button[type="submit"]');
      
      // Navigate to route optimization tool
      await page.click('[data-testid="nav-optimization"]');
      await expect(page).toHaveURL('**/admin/optimization');
      
      // Select date for route planning
      const planningDate = new Date();
      planningDate.setDate(planningDate.getDate() + 7);
      await page.fill('[data-testid="route-date"]', planningDate.toISOString().split('T')[0]);
      
      // Generate route optimization
      await page.click('[data-testid="generate-routes"]');
      
      // Should show optimized routes
      await expect(page.locator('[data-testid="route-results"]')).toBeVisible();
      await expect(page.locator('[data-testid="route-map"]')).toBeVisible();
      
      // Check for area clustering discounts
      const clusteredJobs = page.locator('[data-testid="clustered-jobs"]');
      const clusterCount = await clusteredJobs.count();
      
      if (clusterCount > 0) {
        // Verify area discount is applied
        await expect(page.locator('[data-testid="area-discount-badge"]')).toBeVisible();
        
        const discountText = await page.locator('[data-testid="area-discount-amount"]').textContent();
        const discountAmount = parseFloat(discountText.replace(/[^0-9.]/g, ''));
        expect(discountAmount).toBeGreaterThan(0);
        
        console.log(`Area clustering discount applied: $${discountAmount}`);
      }
      
      // Verify route efficiency metrics
      await expect(page.locator('[data-testid="route-efficiency"]')).toBeVisible();
      await expect(page.locator('[data-testid="estimated-drive-time"]')).toBeVisible();
      await expect(page.locator('[data-testid="fuel-cost-estimate"]')).toBeVisible();
      
      const efficiency = await page.locator('[data-testid="efficiency-percentage"]').textContent();
      const efficiencyValue = parseFloat(efficiency.replace(/[^0-9.]/g, ''));
      expect(efficiencyValue).toBeGreaterThan(70); // Should be at least 70% efficient
      
      console.log(`Route efficiency: ${efficiencyValue}%`);
    });
    
    test('should handle bookings outside service area with premium pricing', async ({ page }) => {
      await page.goto('/quote-calculator');
      
      // Test premium zone (30+ miles from base)
      await page.fill('input[name="address"]', '123 Far Street');
      await page.fill('input[name="zipCode"]', '06902'); // Stamford, CT (premium zone)
      await page.fill('input[name="squareFootage"]', '3000');
      await page.check('input[data-service="lawn-mowing"]');
      
      await page.click('[data-testid="calculate-price"]');
      
      // Should show premium zone pricing
      await expect(page.locator('[data-testid="premium-zone-notice"]')).toBeVisible();
      await expect(page.locator('[data-testid="travel-surcharge"]')).toBeVisible();
      
      const basePrice = await page.locator('[data-testid="base-service-price"]').textContent();
      const totalPrice = await page.locator('[data-testid="total-price"]').textContent();
      
      const base = parseFloat(basePrice.replace(/[^0-9.]/g, ''));
      const total = parseFloat(totalPrice.replace(/[^0-9.]/g, ''));
      
      // Premium zone should be 30% higher
      const expectedPremium = base * 1.30;
      expect(total).toBeGreaterThan(expectedPremium * 0.95); // Allow 5% variance
      expect(total).toBeLessThan(expectedPremium * 1.05);
      
      console.log(`Premium zone pricing - Base: $${base}, Total: $${total} (${((total/base - 1) * 100).toFixed(1)}% premium)`);
    });
    
    test('should optimize crew schedules for route efficiency', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', 'mike@landscapepro.com');
      await page.fill('input[name="password"]', 'mike2024');
      await page.click('button[type="submit"]');
      
      await page.click('[data-testid="nav-schedule"]');
      
      // View current schedule inefficiencies
      await expect(page.locator('[data-testid="schedule-overview"]')).toBeVisible();
      
      const beforeEfficiency = await page.locator('[data-testid="current-efficiency"]').textContent();
      const beforeValue = parseFloat(beforeEfficiency.replace(/[^0-9.]/g, ''));
      
      // Run intelligent schedule optimization
      await page.click('[data-testid="optimize-all-routes"]');
      
      // Should show optimization progress
      await expect(page.locator('[data-testid="optimization-progress"]')).toBeVisible();
      
      // Wait for optimization to complete
      await page.waitForSelector('[data-testid="optimization-complete"]', { timeout: 10000 });
      
      // Check improved efficiency
      const afterEfficiency = await page.locator('[data-testid="optimized-efficiency"]').textContent();
      const afterValue = parseFloat(afterEfficiency.replace(/[^0-9.]/g, ''));
      
      expect(afterValue).toBeGreaterThan(beforeValue);
      
      // Should show estimated savings
      await expect(page.locator('[data-testid="estimated-savings"]')).toBeVisible();
      
      const savings = await page.locator('[data-testid="daily-savings"]').textContent();
      const savingsValue = parseFloat(savings.replace(/[^0-9.]/g, ''));
      expect(savingsValue).toBeGreaterThan(0);
      
      console.log(`Schedule optimization improved efficiency from ${beforeValue}% to ${afterValue}%`);
      console.log(`Estimated daily savings: $${savingsValue}`);
      
      // Apply optimized schedule
      await page.click('[data-testid="apply-optimization"]');
      
      await expect(page.locator('[data-testid="success-toast"]'))
        .toContainText('Schedule optimized and applied');
    });
  });

  test.describe('Seasonal and Peak Pricing Logic', () => {
    test('should apply peak season pricing during spring/summer', async ({ page }) => {
      // Mock current date to be in peak season (June)
      await page.addInitScript(() => {
        Date.now = () => new Date('2024-06-15T10:00:00Z').getTime();
      });
      
      await page.goto('/quote-calculator');
      
      await page.fill('input[name="squareFootage"]', '3000');
      await page.check('input[data-service="lawn-mowing"]');
      await page.check('input[data-service="hedge-trimming"]');
      
      await page.click('[data-testid="calculate-price"]');
      
      // Should show peak season pricing notice
      await expect(page.locator('[data-testid="peak-season-notice"]')).toBeVisible();
      await expect(page.locator('[data-testid="peak-season-notice"]'))
        .toContainText(/peak season|summer pricing|high demand/i);
      
      const peakPrice = await page.locator('[data-testid="total-price"]').textContent();
      const peak = parseFloat(peakPrice.replace(/[^0-9.]/g, ''));
      
      // Compare with off-season pricing by mocking winter date
      await page.addInitScript(() => {
        Date.now = () => new Date('2024-01-15T10:00:00Z').getTime();
      });
      
      await page.reload();
      await page.fill('input[name="squareFootage"]', '3000');
      await page.check('input[data-service="lawn-mowing"]');
      await page.check('input[data-service="hedge-trimming"]');
      await page.click('[data-testid="calculate-price"]');
      
      const offSeasonPrice = await page.locator('[data-testid="total-price"]').textContent();
      const offSeason = parseFloat(offSeasonPrice.replace(/[^0-9.]/g, ''));
      
      // Peak season should be 20-25% higher
      const expectedIncrease = 1.20;
      expect(peak).toBeGreaterThan(offSeason * expectedIncrease * 0.95);
      
      console.log(`Seasonal pricing - Peak: $${peak}, Off-season: $${offSeason} (${((peak/offSeason - 1) * 100).toFixed(1)}% increase)`);
    });
    
    test('should apply fall cleanup premium pricing', async ({ page }) => {
      // Mock date to be in fall season
      await page.addInitScript(() => {
        Date.now = () => new Date('2024-10-15T10:00:00Z').getTime();
      });
      
      await page.goto('/quote-calculator');
      
      await page.fill('input[name="squareFootage"]', '5000');
      await page.check('input[data-service="leaf-cleanup"]');
      
      await page.click('[data-testid="calculate-price"]');
      
      // Should show fall season messaging
      await expect(page.locator('[data-testid="fall-season-notice"]')).toBeVisible();
      
      const fallPrice = await page.locator('[data-testid="total-price"]').textContent();
      const fall = parseFloat(fallPrice.replace(/[^0-9.]/g, ''));
      
      // Fall leaf cleanup should be premium priced
      expect(fall).toBeGreaterThan(120); // Minimum expected for large property
      
      // Should suggest additional services
      await expect(page.locator('[data-testid="suggested-services"]')).toBeVisible();
      await expect(page.locator('[data-testid="suggested-services"]'))
        .toContainText(/gutter cleaning|winter prep/i);
      
      console.log(`Fall cleanup pricing: $${fall} for 5000 sq ft property`);
    });
    
    test('should handle snow removal emergency pricing', async ({ page }) => {
      // Mock winter date
      await page.addInitScript(() => {
        Date.now = () => new Date('2024-01-15T10:00:00Z').getTime();
      });
      
      await page.goto('/emergency-service');
      
      // Emergency snow removal form
      await page.fill('input[name="address"]', '123 Emergency Lane');
      await page.fill('input[name="zipCode"]', '10601');
      await page.selectOption('select[name="emergency-type"]', 'snow-removal');
      await page.selectOption('select[name="urgency"]', 'immediate');
      
      await page.fill('textarea[name="emergency-details"]', 
        'Heavy snowfall blocking driveway and walkways. Need immediate clearing for medical appointment.');
      
      await page.click('[data-testid="get-emergency-quote"]');
      
      // Should show emergency pricing
      await expect(page.locator('[data-testid="emergency-pricing-notice"]')).toBeVisible();
      await expect(page.locator('[data-testid="emergency-pricing-notice"]'))
        .toContainText(/emergency rate|premium pricing|immediate service/i);
      
      const emergencyPrice = await page.locator('[data-testid="emergency-price"]').textContent();
      const price = parseFloat(emergencyPrice.replace(/[^0-9.]/g, ''));
      
      // Emergency should be 50% premium
      expect(price).toBeGreaterThan(120); // Minimum for emergency snow removal
      
      // Should show rapid response time
      await expect(page.locator('[data-testid="response-time"]')).toBeVisible();
      await expect(page.locator('[data-testid="response-time"]'))
        .toContainText(/within.*hour|immediate/i);
      
      console.log(`Emergency snow removal pricing: $${price}`);
    });
  });

  test.describe('Customer Loyalty and Repeat Discounts', () => {
    test('should apply loyalty discounts for repeat customers', async ({ page }) => {
      // Login as long-term customer (Lisa Johnson)
      await page.goto('/customer-login');
      await page.fill('input[name="email"]', 'lisa.johnson@email.com');
      await page.fill('input[name="password"]', 'lisa123');
      await page.click('button[type="submit"]');
      
      await expect(page).toHaveURL('**/customer-dashboard');
      
      // Check loyalty status
      await expect(page.locator('[data-testid="loyalty-status"]')).toBeVisible();
      await expect(page.locator('[data-testid="loyalty-tier"]')).toContainText(/gold|platinum|vip/i);
      
      // Book new service
      await page.click('[data-testid="book-service-button"]');
      
      await page.check('input[data-service="lawn-mowing"]');
      await page.check('input[data-service="hedge-trimming"]');
      
      // Should automatically show loyalty discount
      await expect(page.locator('[data-testid="loyalty-discount"]')).toBeVisible();
      
      const regularPrice = await page.locator('[data-testid="regular-price"]').textContent();
      const loyaltyPrice = await page.locator('[data-testid="loyalty-price"]').textContent();
      
      const regular = parseFloat(regularPrice.replace(/[^0-9.]/g, ''));
      const loyalty = parseFloat(loyaltyPrice.replace(/[^0-9.]/g, ''));
      
      expect(loyalty).toBeLessThan(regular);
      
      const discountPercent = ((regular - loyalty) / regular * 100);
      expect(discountPercent).toBeGreaterThan(10); // At least 10% loyalty discount
      
      console.log(`Loyalty discount: ${discountPercent.toFixed(1)}% off ($${(regular - loyalty).toFixed(2)} savings)`);
    });
    
    test('should offer contract pricing for commercial customers', async ({ page }) => {
      // Login as commercial customer (Amanda Foster)
      await page.goto('/customer-login');
      await page.fill('input[name="email"]', 'amanda@midtownoffice.com');
      await page.fill('input[name="password"]', 'amanda123');
      await page.click('button[type="submit"]');
      
      // Should see commercial dashboard
      await expect(page.locator('[data-testid="commercial-dashboard"]')).toBeVisible();
      await expect(page.locator('[data-testid="contract-status"]')).toBeVisible();
      
      // View contract pricing
      await page.click('[data-testid="view-contract-pricing"]');
      
      // Should show volume discounts
      await expect(page.locator('[data-testid="volume-discount-tiers"]')).toBeVisible();
      
      const tiers = page.locator('[data-testid^="tier-"]');
      const tierCount = await tiers.count();
      expect(tierCount).toBeGreaterThan(2); // Multiple discount tiers
      
      // Check annual contract option
      await page.click('[data-testid="annual-contract-tab"]');
      await expect(page.locator('[data-testid="annual-savings"]')).toBeVisible();
      
      const annualSavings = await page.locator('[data-testid="annual-savings-amount"]').textContent();
      const savings = parseFloat(annualSavings.replace(/[^0-9.]/g, ''));
      expect(savings).toBeGreaterThan(1000); // Significant annual savings
      
      console.log(`Commercial annual contract savings: $${savings}`);
    });
    
    test('should track and reward customer referrals', async ({ page }) => {
      await page.goto('/customer-login');
      await page.fill('input[name="email"]', 'john.smith@email.com');
      await page.fill('input[name="password"]', 'john123');
      await page.click('button[type="submit"]');
      
      // Access referral program
      await page.click('[data-testid="referral-program-tab"]');
      
      await expect(page.locator('[data-testid="referral-dashboard"]')).toBeVisible();
      
      // Should show referral history
      await expect(page.locator('[data-testid="referral-history"]')).toBeVisible();
      
      const referralCount = await page.locator('[data-testid="successful-referrals"]').textContent();
      const count = parseInt(referralCount.replace(/[^0-9]/g, ''));
      
      if (count > 0) {
        // Should show referral rewards
        await expect(page.locator('[data-testid="referral-rewards"]')).toBeVisible();
        
        const rewards = await page.locator('[data-testid="total-referral-credits"]').textContent();
        const creditAmount = parseFloat(rewards.replace(/[^0-9.]/g, ''));
        
        expect(creditAmount).toBeGreaterThan(0);
        console.log(`Referral rewards: $${creditAmount} for ${count} successful referrals`);
      }
      
      // Test new referral
      await page.click('[data-testid="send-referral-button"]');
      
      await page.fill('input[name="referral-name"]', 'Jane Neighbor');
      await page.fill('input[name="referral-email"]', 'jane.neighbor@email.com');
      await page.fill('input[name="referral-phone"]', '+1-555-7777');
      
      await page.click('[data-testid="send-referral"]');
      
      await expect(page.locator('[data-testid="referral-sent-confirmation"]'))
        .toContainText('Referral sent successfully');
    });
  });

  test.describe('Weather-Dependent Service Logic', () => {
    test('should handle weather-dependent scheduling', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', 'mike@landscapepro.com');
      await page.fill('input[name="password"]', 'mike2024');
      await page.click('button[type="submit"]');
      
      await page.click('[data-testid="nav-schedule"]');
      
      // View weather-dependent jobs
      await page.selectOption('[data-testid="job-filter"]', 'weather-dependent');
      await page.click('[data-testid="apply-filter"]');
      
      const weatherDependentJobs = page.locator('[data-testid^="weather-job-"]');
      const jobCount = await weatherDependentJobs.count();
      
      if (jobCount > 0) {
        // Should show weather forecast integration
        await expect(page.locator('[data-testid="weather-forecast"]')).toBeVisible();
        
        // Check for weather alerts
        const hasWeatherAlert = await page.locator('[data-testid="weather-alert"]').isVisible();
        
        if (hasWeatherAlert) {
          // Should suggest rescheduling for weather-dependent jobs
          await expect(page.locator('[data-testid="reschedule-suggestion"]')).toBeVisible();
          
          // Test automatic rescheduling
          await page.click('[data-testid="auto-reschedule-weather-jobs"]');
          
          await expect(page.locator('[data-testid="reschedule-confirmation"]'))
            .toContainText(/rescheduled.*weather/i);
        }
      }
    });
    
    test('should offer indoor alternatives during bad weather', async ({ page }) => {
      // Mock bad weather conditions
      await page.route('**/api/weather/**', route => {
        route.fulfill({
          status: 200,
          body: JSON.stringify({
            current: { condition: 'Rain', temperature: 45 },
            forecast: [
              { date: '2024-09-15', condition: 'Heavy Rain', precipitation: 85 },
              { date: '2024-09-16', condition: 'Thunderstorms', precipitation: 90 }
            ]
          })
        });
      });
      
      await page.goto('/customer-login');
      await page.fill('input[name="email"]', 'john.smith@email.com');
      await page.fill('input[name="password"]', 'john123');
      await page.click('button[type="submit"]');
      
      await page.click('[data-testid="book-service-button"]');
      
      // Should show weather warning
      await expect(page.locator('[data-testid="weather-warning"]')).toBeVisible();
      
      // Should suggest indoor alternatives or rescheduling
      await expect(page.locator('[data-testid="indoor-alternatives"]')).toBeVisible();
      await expect(page.locator('[data-testid="suggested-reschedule-dates"]')).toBeVisible();
      
      const indoorServices = page.locator('[data-testid^="indoor-service-"]');
      const indoorCount = await indoorServices.count();
      expect(indoorCount).toBeGreaterThan(0);
      
      console.log(`${indoorCount} indoor service alternatives offered during bad weather`);
    });
  });

  test.describe('Dynamic Pricing Algorithm Validation', () => {
    test('should adjust pricing based on demand and capacity', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', 'sarah@landscapepro.com');
      await page.fill('input[name="password']', 'sarah2024');
      await page.click('button[type="submit"]');
      
      // Access dynamic pricing dashboard
      await page.click('[data-testid="nav-pricing"]');
      
      await expect(page.locator('[data-testid="dynamic-pricing-dashboard"]')).toBeVisible();
      
      // View current demand metrics
      await expect(page.locator('[data-testid="demand-meter"]')).toBeVisible();
      await expect(page.locator('[data-testid="capacity-utilization"]')).toBeVisible();
      
      const demandLevel = await page.locator('[data-testid="current-demand-level"]').textContent();
      const capacityLevel = await page.locator('[data-testid="capacity-level"]').textContent();
      
      // High demand should increase pricing
      if (demandLevel.includes('High')) {
        await expect(page.locator('[data-testid="surge-pricing-active"]')).toBeVisible();
        
        const surgeMultiplier = await page.locator('[data-testid="surge-multiplier"]').textContent();
        const multiplier = parseFloat(surgeMultiplier.replace(/[^0-9.]/g, ''));
        expect(multiplier).toBeGreaterThan(1.0);
        
        console.log(`Surge pricing active: ${multiplier}x multiplier`);
      }
      
      // Test pricing adjustments
      await page.click('[data-testid="test-pricing-scenario"]');
      
      // Simulate high demand scenario
      await page.selectOption('[data-testid="demand-scenario"]', 'high');
      await page.selectOption('[data-testid="capacity-scenario"]', 'low');
      
      await page.click('[data-testid="calculate-dynamic-pricing"]');
      
      const dynamicPrice = await page.locator('[data-testid="calculated-dynamic-price"]').textContent();
      const basePrice = await page.locator('[data-testid="base-price"]').textContent();
      
      const dynamic = parseFloat(dynamicPrice.replace(/[^0-9.]/g, ''));
      const base = parseFloat(basePrice.replace(/[^0-9.]/g, ''));
      
      expect(dynamic).toBeGreaterThan(base);
      
      console.log(`Dynamic pricing - Base: $${base}, High demand: $${dynamic} (${((dynamic/base - 1) * 100).toFixed(1)}% increase)`);
    });
  });
});

/**
 * Utility functions for business logic testing
 */
function calculateSeasonalMultiplier(date) {
  const month = date.getMonth() + 1; // JavaScript months are 0-based
  
  if ([3, 4, 5].includes(month)) return 1.20; // Spring
  if ([6, 7, 8].includes(month)) return 1.25; // Summer
  if ([9, 10, 11].includes(month)) return 1.15; // Fall
  return 0.90; // Winter
}

function calculateGeographicMultiplier(zipCode) {
  const localZones = ['10601', '10602', '10603'];
  const extendedZones = ['06830', '06831', '10580'];
  const premiumZones = ['06901', '06902', '10701'];
  
  if (localZones.includes(zipCode)) return 1.0;
  if (extendedZones.includes(zipCode)) return 1.15;
  if (premiumZones.includes(zipCode)) return 1.30;
  
  return 1.50; // Outside service area
}

function calculateRouteEfficiency(jobs) {
  // Simplified efficiency calculation for testing
  const totalDistance = jobs.reduce((sum, job, index) => {
    if (index === 0) return sum;
    return sum + calculateDistance(jobs[index - 1].coordinates, job.coordinates);
  }, 0);
  
  const optimalDistance = jobs.length * 5; // Assume 5 miles average between jobs
  return Math.max(0, (optimalDistance / totalDistance) * 100);
}

function calculateDistance(coord1, coord2) {
  // Simplified distance calculation (Haversine formula would be more accurate)
  const latDiff = Math.abs(coord1.lat - coord2.lat);
  const lngDiff = Math.abs(coord1.lng - coord2.lng);
  return Math.sqrt(latDiff * latDiff + lngDiff * lngDiff) * 69; // Rough miles conversion
}

async function mockWeatherAPI(page, weatherCondition) {
  await page.route('**/api/weather/**', route => {
    const weatherData = {
      'sunny': { condition: 'Sunny', temperature: 75, precipitation: 0 },
      'rainy': { condition: 'Rain', temperature: 55, precipitation: 80 },
      'snow': { condition: 'Snow', temperature: 30, precipitation: 90 }
    };
    
    route.fulfill({
      status: 200,
      body: JSON.stringify({ current: weatherData[weatherCondition] || weatherData['sunny'] })
    });
  });
}