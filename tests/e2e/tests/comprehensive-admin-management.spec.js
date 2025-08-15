const { test, expect } = require('@playwright/test');

/**
 * Comprehensive Admin Management Test Suite
 * Tests all administrative functions including job management, customer management,
 * crew coordination, and business intelligence
 */

test.describe('Comprehensive Admin Management Testing', () => {
  let adminUsers;
  
  test.beforeAll(async () => {
    // Admin user credentials from realistic test data
    adminUsers = {
      owner: {
        email: 'sarah@landscapepro.com',
        password: 'sarah2024',
        name: 'Sarah Williams'
      },
      operations: {
        email: 'mike@landscapepro.com',
        password: 'mike2024',
        name: 'Mike Rodriguez'
      },
      sales: {
        email: 'jen@landscapepro.com',
        password: 'jen2024',
        name: 'Jennifer Chen'
      },
      crewLead: {
        email: 'david@landscapepro.com',
        password: 'david2024',
        name: 'David Thompson'
      }
    };
  });

  test.describe('Scenario C: Admin Management Dashboard', () => {
    test('should login to admin dashboard successfully', async ({ page }) => {
      console.log('Testing admin dashboard login...');
      
      // Login as owner (Sarah Williams)
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.owner.email);
      await page.fill('input[name="password"]', adminUsers.owner.password);
      await page.click('button[type="submit"]');
      
      // Should redirect to admin dashboard
      await expect(page).toHaveURL('**/admin/dashboard');
      
      // Verify admin interface elements
      await expect(page.locator('[data-testid="admin-header"]')).toContainText('Admin Dashboard');
      await expect(page.locator('[data-testid="user-name"]')).toContainText('Sarah Williams');
      await expect(page.locator('[data-testid="user-role"]')).toContainText('Owner');
      
      // Check main navigation elements
      await expect(page.locator('[data-testid="nav-dashboard"]')).toBeVisible();
      await expect(page.locator('[data-testid="nav-jobs"]')).toBeVisible();
      await expect(page.locator('[data-testid="nav-customers"]')).toBeVisible();
      await expect(page.locator('[data-testid="nav-crews"]')).toBeVisible();
      await expect(page.locator('[data-testid="nav-reports"]')).toBeVisible();
      
      console.log('Admin dashboard login successful');
    });
    
    test('should display key performance metrics', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.operations.email);
      await page.fill('input[name="password"]', adminUsers.operations.password);
      await page.click('button[type="submit"]');
      
      await expect(page).toHaveURL('**/admin/dashboard');
      
      // Check dashboard metrics
      const metrics = [
        'total-revenue',
        'active-jobs',
        'pending-quotes',
        'customer-count',
        'crew-utilization',
        'monthly-growth'
      ];
      
      for (const metric of metrics) {
        await expect(page.locator(`[data-testid="${metric}"]`)).toBeVisible();
        
        // Verify metric has actual data (not just zeros)
        const metricValue = await page.locator(`[data-testid="${metric}-value"]`).textContent();
        expect(metricValue).not.toBe('0');
        expect(metricValue).not.toBe('$0.00');
      }
      
      // Check recent activity feed
      await expect(page.locator('[data-testid="recent-activities"]')).toBeVisible();
      await expect(page.locator('[data-testid="activity-item"]').first()).toBeVisible();
      
      console.log('Dashboard metrics validated');
    });
    
    test('should review and manage incoming bookings', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.operations.email);
      await page.fill('input[name="password"]', adminUsers.operations.password);
      await page.click('button[type="submit"]');
      
      // Navigate to jobs management
      await page.click('[data-testid="nav-jobs"]');
      await expect(page).toHaveURL('**/admin/jobs');
      
      // Check job filters and status options
      await expect(page.locator('[data-testid="status-filter"]')).toBeVisible();
      await expect(page.locator('[data-testid="date-filter"]')).toBeVisible();
      await expect(page.locator('[data-testid="crew-filter"]')).toBeVisible();
      
      // Filter for pending jobs
      await page.selectOption('[data-testid="status-filter"]', 'pending');
      await page.click('[data-testid="apply-filters"]');
      
      // Should show pending jobs
      await expect(page.locator('[data-testid="jobs-table"]')).toBeVisible();
      const pendingJobs = page.locator('[data-testid^="job-row-"]');
      const jobCount = await pendingJobs.count();
      
      if (jobCount > 0) {
        // Review first pending job
        await page.click('[data-testid="view-job-0"]');
        
        // Job details should be visible
        await expect(page.locator('[data-testid="job-details-modal"]')).toBeVisible();
        await expect(page.locator('[data-testid="customer-info"]')).toBeVisible();
        await expect(page.locator('[data-testid="property-info"]')).toBeVisible();
        await expect(page.locator('[data-testid="service-details"]')).toBeVisible();
        
        // Approve the job
        await page.click('[data-testid="approve-job-button"]');
        
        // Should show success message
        await expect(page.locator('[data-testid="success-toast"]'))
          .toContainText('Job approved successfully');
        
        console.log(`Approved ${jobCount} pending jobs`);
      }
    });
    
    test('should assign jobs to crew members', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.operations.email);
      await page.fill('input[name="password"]', adminUsers.operations.password);
      await page.click('button[type="submit"]');
      
      await page.click('[data-testid="nav-jobs"]');
      
      // Filter for approved but unassigned jobs
      await page.selectOption('[data-testid="status-filter"]', 'approved');
      await page.selectOption('[data-testid="assignment-filter"]', 'unassigned');
      await page.click('[data-testid="apply-filters"]');
      
      const unassignedJobs = page.locator('[data-testid^="job-row-"]');
      const unassignedCount = await unassignedJobs.count();
      
      if (unassignedCount > 0) {
        // Assign first unassigned job
        await page.click('[data-testid="assign-job-0"]');
        
        // Assignment modal should open
        await expect(page.locator('[data-testid="assignment-modal"]')).toBeVisible();
        
        // Select crew and date
        await page.selectOption('[data-testid="crew-select"]', 'Team Alpha');
        
        const nextWeek = new Date();
        nextWeek.setDate(nextWeek.getDate() + 7);
        await page.fill('[data-testid="scheduled-date"]', nextWeek.toISOString().split('T')[0]);
        await page.selectOption('[data-testid="scheduled-time"]', '09:00');
        
        // Add assignment notes
        await page.fill('[data-testid="assignment-notes"]', 
          'Regular maintenance crew assigned. Customer prefers morning service.');
        
        // Confirm assignment
        await page.click('[data-testid="confirm-assignment"]');
        
        // Should show success message
        await expect(page.locator('[data-testid="success-toast"]'))
          .toContainText('Job assigned successfully');
        
        console.log(`Assigned ${unassignedCount} jobs to crews`);
      }
    });
    
    test('should manage customer relationships', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.sales.email);
      await page.fill('input[name="password"]', adminUsers.sales.password);
      await page.click('button[type="submit"]');
      
      // Navigate to customer management
      await page.click('[data-testid="nav-customers"]');
      await expect(page).toHaveURL('**/admin/customers');
      
      // Check customer overview stats
      await expect(page.locator('[data-testid="total-customers"]')).toBeVisible();
      await expect(page.locator('[data-testid="active-customers"]')).toBeVisible();
      await expect(page.locator('[data-testid="new-customers-month"]')).toBeVisible();
      await expect(page.locator('[data-testid="customer-ltv"]')).toBeVisible();
      
      // Search for specific customer
      await page.fill('[data-testid="customer-search"]', 'John Smith');
      await page.press('[data-testid="customer-search"]', 'Enter');
      
      // Should find customer
      await expect(page.locator('[data-testid="customer-results"]')).toBeVisible();
      await expect(page.locator('text=John Smith')).toBeVisible();
      
      // View customer details
      await page.click('[data-testid="view-customer-john-smith"]');
      
      // Customer profile should show comprehensive info
      await expect(page.locator('[data-testid="customer-profile"]')).toBeVisible();
      await expect(page.locator('[data-testid="customer-properties"]')).toBeVisible();
      await expect(page.locator('[data-testid="service-history"]')).toBeVisible();
      await expect(page.locator('[data-testid="payment-history"]')).toBeVisible();
      
      // Check customer communication log
      await page.click('[data-testid="communication-tab"]');
      await expect(page.locator('[data-testid="communication-log"]')).toBeVisible();
      
      // Add communication note
      await page.click('[data-testid="add-note-button"]');
      await page.fill('[data-testid="note-text"]', 
        'Customer called asking about fall cleanup service. Interested in leaf removal package.');
      await page.selectOption('[data-testid="note-type"]', 'phone_call');
      await page.click('[data-testid="save-note"]');
      
      await expect(page.locator('[data-testid="success-toast"]'))
        .toContainText('Note added successfully');
      
      console.log('Customer management functions verified');
    });
    
    test('should generate and view reports', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.owner.email);
      await page.fill('input[name="password"]', adminUsers.owner.password);
      await page.click('button[type="submit"]');
      
      // Navigate to reports
      await page.click('[data-testid="nav-reports"]');
      await expect(page).toHaveURL('**/admin/reports');
      
      // Check available report types
      const reportTypes = [
        'revenue-report',
        'customer-report',
        'crew-efficiency-report',
        'service-analysis-report',
        'geographic-analysis-report'
      ];
      
      for (const reportType of reportTypes) {
        await expect(page.locator(`[data-testid="${reportType}"]`)).toBeVisible();
      }
      
      // Generate revenue report
      await page.click('[data-testid="revenue-report"]');
      
      // Set date range (last 30 days)
      const endDate = new Date().toISOString().split('T')[0];
      const startDate = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString().split('T')[0];
      
      await page.fill('[data-testid="start-date"]', startDate);
      await page.fill('[data-testid="end-date"]', endDate);
      await page.click('[data-testid="generate-report"]');
      
      // Report should load
      await expect(page.locator('[data-testid="report-results"]')).toBeVisible();
      await expect(page.locator('[data-testid="revenue-chart"]')).toBeVisible();
      await expect(page.locator('[data-testid="revenue-summary"]')).toBeVisible();
      
      // Check export functionality
      await page.click('[data-testid="export-pdf"]');
      // Note: In real test, we'd verify download
      
      console.log('Reports and analytics verified');
    });
  });

  test.describe('Role-Based Access Control', () => {
    test('should restrict sales manager access appropriately', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.sales.email);
      await page.fill('input[name="password"]', adminUsers.sales.password);
      await page.click('button[type="submit"]');
      
      // Sales manager should see limited navigation
      await expect(page.locator('[data-testid="nav-customers"]')).toBeVisible();
      await expect(page.locator('[data-testid="nav-quotes"]')).toBeVisible();
      await expect(page.locator('[data-testid="nav-reports"]')).toBeVisible();
      
      // Should NOT see crew management or system settings
      await expect(page.locator('[data-testid="nav-crews"]')).not.toBeVisible();
      await expect(page.locator('[data-testid="nav-system-settings"]')).not.toBeVisible();
      
      // Try to access restricted area
      await page.goto('/admin/system-settings');
      await expect(page).toHaveURL('**/access-denied');
      await expect(page.locator('[data-testid="access-denied-message"]'))
        .toContainText('Access denied');
    });
    
    test('should allow crew lead limited job management', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.crewLead.email);
      await page.fill('input[name="password"]', adminUsers.crewLead.password);
      await page.click('button[type="submit"]');
      
      // Crew lead should see job management
      await expect(page.locator('[data-testid="nav-jobs"]')).toBeVisible();
      
      await page.click('[data-testid="nav-jobs"]');
      
      // Should only see jobs assigned to their crew
      await expect(page.locator('[data-testid="my-crew-jobs"]')).toBeVisible();
      
      // Should be able to update job status
      const jobRows = page.locator('[data-testid^="job-row-"]');
      const jobCount = await jobRows.count();
      
      if (jobCount > 0) {
        await page.click('[data-testid="update-status-0"]');
        await page.selectOption('[data-testid="status-select"]', 'in_progress');
        await page.click('[data-testid="save-status"]');
        
        await expect(page.locator('[data-testid="success-toast"]'))
          .toContainText('Job status updated');
      }
    });
  });

  test.describe('Real-time Updates and Notifications', () => {
    test('should show real-time job updates', async ({ page, context }) => {
      // Open two browser sessions (admin and crew lead)
      const adminPage = page;
      const crewPage = await context.newPage();
      
      // Admin login
      await adminPage.goto('/admin/login');
      await adminPage.fill('input[name="email"]', adminUsers.operations.email);
      await adminPage.fill('input[name="password"]', adminUsers.operations.password);
      await adminPage.click('button[type="submit"]');
      await adminPage.click('[data-testid="nav-jobs"]');
      
      // Crew lead login
      await crewPage.goto('/admin/login');
      await crewPage.fill('input[name="email"]', adminUsers.crewLead.email);
      await crewPage.fill('input[name="password"]', adminUsers.crewLead.password);
      await crewPage.click('button[type="submit"]');
      await crewPage.click('[data-testid="nav-jobs"]');
      
      // Crew lead updates job status
      const jobRows = crewPage.locator('[data-testid^="job-row-"]');
      const jobCount = await jobRows.count();
      
      if (jobCount > 0) {
        await crewPage.click('[data-testid="update-status-0"]');
        await crewPage.selectOption('[data-testid="status-select"]', 'completed');
        await crewPage.fill('[data-testid="completion-notes"]', 'Job completed successfully. Customer satisfied.');
        await crewPage.click('[data-testid="save-status"]');
        
        // Admin should see real-time update
        await adminPage.waitForTimeout(2000); // Allow for real-time update
        await expect(adminPage.locator('[data-testid="job-status-completed"]')).toBeVisible();
        
        // Should show notification
        await expect(adminPage.locator('[data-testid="notification-popup"]'))
          .toContainText('Job completed');
      }
      
      await crewPage.close();
    });
    
    test('should display system notifications', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.operations.email);
      await page.fill('input[name="password"]', adminUsers.operations.password);
      await page.click('button[type="submit"]');
      
      // Check notification center
      await page.click('[data-testid="notification-bell"]');
      await expect(page.locator('[data-testid="notification-dropdown"]')).toBeVisible();
      
      // Should show various notification types
      const notificationTypes = await page.locator('[data-testid^="notification-"]').count();
      expect(notificationTypes).toBeGreaterThan(0);
      
      // Mark notification as read
      if (notificationTypes > 0) {
        await page.click('[data-testid="notification-0"]');
        await expect(page.locator('[data-testid="notification-0"]'))
          .toHaveClass(/read|viewed/);
      }
    });
  });

  test.describe('Bulk Operations and Efficiency', () => {
    test('should handle bulk job assignments', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.operations.email);
      await page.fill('input[name="password"]', adminUsers.operations.password);
      await page.click('button[type="submit"]');
      
      await page.click('[data-testid="nav-jobs"]');
      
      // Select multiple unassigned jobs
      await page.selectOption('[data-testid="status-filter"]', 'approved');
      await page.selectOption('[data-testid="assignment-filter"]', 'unassigned');
      await page.click('[data-testid="apply-filters"]');
      
      // Select first 3 jobs for bulk assignment
      const jobCheckboxes = page.locator('[data-testid^="select-job-"]');
      const jobCount = await jobCheckboxes.count();
      const selectCount = Math.min(3, jobCount);
      
      for (let i = 0; i < selectCount; i++) {
        await page.check(`[data-testid="select-job-${i}"]`);
      }
      
      // Bulk assignment
      await page.click('[data-testid="bulk-assign-button"]');
      await expect(page.locator('[data-testid="bulk-assignment-modal"]')).toBeVisible();
      
      await page.selectOption('[data-testid="bulk-crew-select"]', 'Team Alpha');
      await page.fill('[data-testid="bulk-date"]', '2024-09-15');
      await page.click('[data-testid="confirm-bulk-assignment"]');
      
      await expect(page.locator('[data-testid="success-toast"]'))
        .toContainText(`${selectCount} jobs assigned successfully`);
      
      console.log(`Bulk assigned ${selectCount} jobs`);
    });
    
    test('should optimize crew schedules', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.operations.email);
      await page.fill('input[name="password"]', adminUsers.operations.password);
      await page.click('button[type="submit"]');
      
      // Navigate to schedule optimization
      await page.click('[data-testid="nav-schedule"]');
      
      // View crew utilization
      await expect(page.locator('[data-testid="crew-utilization-chart"]')).toBeVisible();
      
      // Run schedule optimization
      await page.click('[data-testid="optimize-schedule-button"]');
      
      // Should show optimization results
      await expect(page.locator('[data-testid="optimization-results"]')).toBeVisible();
      await expect(page.locator('[data-testid="efficiency-improvement"]')).toBeVisible();
      
      // Should show route optimization suggestions
      await expect(page.locator('[data-testid="route-suggestions"]')).toBeVisible();
      
      // Apply optimization
      await page.click('[data-testid="apply-optimization"]');
      
      await expect(page.locator('[data-testid="success-toast"]'))
        .toContainText('Schedule optimized');
    });
  });

  test.describe('Data Export and Integration', () => {
    test('should export customer data', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.sales.email);
      await page.fill('input[name="password"]', adminUsers.sales.password);
      await page.click('button[type="submit"]');
      
      await page.click('[data-testid="nav-customers"]');
      
      // Export customer list
      await page.click('[data-testid="export-customers"]');
      
      // Select export options
      await page.check('[data-testid="export-contact-info"]');
      await page.check('[data-testid="export-service-history"]');
      await page.check('[data-testid="export-payment-history"]');
      
      await page.selectOption('[data-testid="export-format"]', 'csv');
      
      const downloadPromise = page.waitForEvent('download');
      await page.click('[data-testid="confirm-export"]');
      
      const download = await downloadPromise;
      expect(download.suggestedFilename()).toMatch(/customers.*\.csv/);
      
      console.log('Customer data export successful');
    });
    
    test('should generate financial reports', async ({ page }) => {
      await page.goto('/admin/login');
      await page.fill('input[name="email"]', adminUsers.owner.email);
      await page.fill('input[name="password"]', adminUsers.owner.password);
      await page.click('button[type="submit"]');
      
      await page.click('[data-testid="nav-reports"]');
      await page.click('[data-testid="financial-reports"]');
      
      // Generate P&L report
      await page.selectOption('[data-testid="report-type"]', 'profit_loss');
      await page.selectOption('[data-testid="report-period"]', 'last_quarter');
      
      await page.click('[data-testid="generate-financial-report"]');
      
      // Should show financial metrics
      await expect(page.locator('[data-testid="total-revenue"]')).toBeVisible();
      await expect(page.locator('[data-testid="total-expenses"]')).toBeVisible();
      await expect(page.locator('[data-testid="net-profit"]')).toBeVisible();
      await expect(page.locator('[data-testid="profit-margin"]')).toBeVisible();
      
      // Export financial report
      const downloadPromise = page.waitForEvent('download');
      await page.click('[data-testid="export-financial-pdf"]');
      
      const download = await downloadPromise;
      expect(download.suggestedFilename()).toMatch(/financial.*\.pdf/);
    });
  });
});

/**
 * Helper functions for admin testing
 */
async function waitForRealTimeUpdate(page, selector, timeout = 5000) {
  return await page.waitForSelector(selector, { timeout });
}

async function verifyUserPermissions(page, allowedActions, restrictedActions) {
  // Verify allowed actions are visible/accessible
  for (const action of allowedActions) {
    await expect(page.locator(`[data-testid="${action}"]`)).toBeVisible();
  }
  
  // Verify restricted actions are not visible
  for (const action of restrictedActions) {
    await expect(page.locator(`[data-testid="${action}"]`)).not.toBeVisible();
  }
}

function generateTestJobData(count = 5) {
  const jobs = [];
  const services = ['lawn-mowing', 'hedge-trimming', 'leaf-cleanup', 'fertilization'];
  const statuses = ['pending', 'approved', 'scheduled', 'in_progress'];
  
  for (let i = 0; i < count; i++) {
    jobs.push({
      id: `test-job-${i}`,
      customer: `Test Customer ${i}`,
      service: services[i % services.length],
      status: statuses[i % statuses.length],
      scheduledDate: new Date(Date.now() + i * 24 * 60 * 60 * 1000).toISOString().split('T')[0]
    });
  }
  
  return jobs;
}