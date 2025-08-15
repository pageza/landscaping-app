const { test, expect } = require('@playwright/test');

test.describe('Customer Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin
    await page.goto('/');
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');
    
    // Navigate to customers page
    await page.click('text=Customers');
    await page.waitForURL('**/customers');
  });

  test('should display customers list', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Customers');
    await expect(page.locator('[data-testid="customers-table"]')).toBeVisible();
    await expect(page.locator('[data-testid="add-customer-button"]')).toBeVisible();
    await expect(page.locator('[data-testid="search-input"]')).toBeVisible();
  });

  test('should add new customer successfully', async ({ page }) => {
    // Click add customer button
    await page.click('[data-testid="add-customer-button"]');
    
    // Should open modal or navigate to form
    await expect(page.locator('[data-testid="customer-form"]')).toBeVisible();
    
    // Fill customer details
    await page.fill('input[name="name"]', 'John Smith');
    await page.fill('input[name="email"]', 'john.smith@example.com');
    await page.fill('input[name="phone"]', '+1-555-0123');
    await page.fill('input[name="company"]', 'Smith Landscaping');
    await page.fill('input[name="address"]', '123 Main Street');
    await page.fill('input[name="city"]', 'Springfield');
    await page.fill('input[name="state"]', 'IL');
    await page.fill('input[name="zipCode"]', '62701');
    
    // Submit form
    await page.click('[data-testid="save-customer-button"]');
    
    // Should show success message
    await expect(page.locator('[data-testid="success-toast"]')).toContainText('Customer added successfully');
    
    // Should see new customer in list
    await expect(page.locator('text=John Smith')).toBeVisible();
    await expect(page.locator('text=john.smith@example.com')).toBeVisible();
  });

  test('should validate required fields when adding customer', async ({ page }) => {
    await page.click('[data-testid="add-customer-button"]');
    
    // Try to submit empty form
    await page.click('[data-testid="save-customer-button"]');
    
    // Should show validation errors
    await expect(page.locator('[data-testid="name-error"]')).toContainText('Name is required');
    await expect(page.locator('[data-testid="email-error"]')).toContainText('Email is required');
    await expect(page.locator('[data-testid="phone-error"]')).toContainText('Phone is required');
  });

  test('should validate email format', async ({ page }) => {
    await page.click('[data-testid="add-customer-button"]');
    
    await page.fill('input[name="name"]', 'John Smith');
    await page.fill('input[name="email"]', 'invalid-email');
    await page.fill('input[name="phone"]', '+1-555-0123');
    
    await page.click('[data-testid="save-customer-button"]');
    
    await expect(page.locator('[data-testid="email-error"]')).toContainText('Please enter a valid email address');
  });

  test('should edit existing customer', async ({ page }) => {
    // Click edit button for first customer
    await page.click('[data-testid="edit-customer-0"]');
    
    await expect(page.locator('[data-testid="customer-form"]')).toBeVisible();
    
    // Update customer details
    await page.fill('input[name="name"]', 'John Updated Smith');
    await page.fill('input[name="phone"]', '+1-555-9999');
    
    await page.click('[data-testid="save-customer-button"]');
    
    await expect(page.locator('[data-testid="success-toast"]')).toContainText('Customer updated successfully');
    
    // Should see updated information
    await expect(page.locator('text=John Updated Smith')).toBeVisible();
  });

  test('should view customer details', async ({ page }) => {
    // Click view button for first customer
    await page.click('[data-testid="view-customer-0"]');
    
    await expect(page.locator('[data-testid="customer-details"]')).toBeVisible();
    await expect(page.locator('h2')).toContainText('Customer Details');
    
    // Should show customer information
    await expect(page.locator('[data-testid="customer-name"]')).toBeVisible();
    await expect(page.locator('[data-testid="customer-email"]')).toBeVisible();
    await expect(page.locator('[data-testid="customer-phone"]')).toBeVisible();
    
    // Should show related data tabs
    await expect(page.locator('[data-testid="properties-tab"]')).toBeVisible();
    await expect(page.locator('[data-testid="jobs-tab"]')).toBeVisible();
    await expect(page.locator('[data-testid="invoices-tab"]')).toBeVisible();
  });

  test('should delete customer with confirmation', async ({ page }) => {
    // Click delete button for first customer
    await page.click('[data-testid="delete-customer-0"]');
    
    // Should show confirmation dialog
    await expect(page.locator('[data-testid="confirmation-dialog"]')).toBeVisible();
    await expect(page.locator('text=Are you sure you want to delete this customer?')).toBeVisible();
    
    // Confirm deletion
    await page.click('[data-testid="confirm-delete-button"]');
    
    await expect(page.locator('[data-testid="success-toast"]')).toContainText('Customer deleted successfully');
  });

  test('should cancel customer deletion', async ({ page }) => {
    const initialCustomerCount = await page.locator('[data-testid^="customer-row-"]').count();
    
    await page.click('[data-testid="delete-customer-0"]');
    await expect(page.locator('[data-testid="confirmation-dialog"]')).toBeVisible();
    
    // Cancel deletion
    await page.click('[data-testid="cancel-delete-button"]');
    
    // Dialog should close and customer should still exist
    await expect(page.locator('[data-testid="confirmation-dialog"]')).not.toBeVisible();
    const currentCustomerCount = await page.locator('[data-testid^="customer-row-"]').count();
    expect(currentCustomerCount).toBe(initialCustomerCount);
  });

  test('should search customers by name', async ({ page }) => {
    // Search for specific customer
    await page.fill('[data-testid="search-input"]', 'John');
    await page.press('[data-testid="search-input"]', 'Enter');
    
    // Should filter results
    await page.waitForTimeout(1000); // Wait for search results
    
    const visibleCustomers = page.locator('[data-testid^="customer-row-"]');
    await expect(visibleCustomers.first()).toContainText('John');
  });

  test('should search customers by email', async ({ page }) => {
    await page.fill('[data-testid="search-input"]', 'example.com');
    await page.press('[data-testid="search-input"]', 'Enter');
    
    await page.waitForTimeout(1000);
    
    const visibleCustomers = page.locator('[data-testid^="customer-row-"]');
    await expect(visibleCustomers.first()).toContainText('example.com');
  });

  test('should filter customers by status', async ({ page }) => {
    // Click status filter dropdown
    await page.click('[data-testid="status-filter"]');
    
    // Select active customers only
    await page.click('[data-testid="status-active"]');
    
    // All visible customers should be active
    const customerStatuses = page.locator('[data-testid^="customer-status-"]');
    const count = await customerStatuses.count();
    
    for (let i = 0; i < count; i++) {
      await expect(customerStatuses.nth(i)).toContainText('Active');
    }
  });

  test('should sort customers by name', async ({ page }) => {
    // Click name column header to sort
    await page.click('[data-testid="sort-by-name"]');
    
    // Get all customer names
    const customerNames = await page.locator('[data-testid^="customer-name-"]').allTextContents();
    
    // Verify they are sorted alphabetically
    const sortedNames = [...customerNames].sort();
    expect(customerNames).toEqual(sortedNames);
  });

  test('should paginate customers list', async ({ page }) => {
    // Check if pagination is visible (only if there are enough customers)
    const paginationVisible = await page.locator('[data-testid="pagination"]').isVisible();
    
    if (paginationVisible) {
      // Click next page
      await page.click('[data-testid="next-page"]');
      
      // Should load different customers
      await page.waitForTimeout(1000);
      
      // Click previous page
      await page.click('[data-testid="prev-page"]');
      
      await page.waitForTimeout(1000);
    }
  });

  test('should export customers list', async ({ page }) => {
    // Click export button
    await page.click('[data-testid="export-customers-button"]');
    
    // Should show export options
    await expect(page.locator('[data-testid="export-options"]')).toBeVisible();
    
    // Download CSV
    const downloadPromise = page.waitForEvent('download');
    await page.click('[data-testid="export-csv"]');
    const download = await downloadPromise;
    
    expect(download.suggestedFilename()).toContain('customers');
    expect(download.suggestedFilename()).toContain('.csv');
  });

  test('should import customers from CSV', async ({ page }) => {
    // Click import button
    await page.click('[data-testid="import-customers-button"]');
    
    // Should show file upload dialog
    await expect(page.locator('[data-testid="file-upload-dialog"]')).toBeVisible();
    
    // Upload CSV file (would need actual test file)
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles('./test-data/customers.csv');
    
    // Submit import
    await page.click('[data-testid="submit-import-button"]');
    
    // Should show import results
    await expect(page.locator('[data-testid="import-results"]')).toBeVisible();
    await expect(page.locator('[data-testid="success-toast"]')).toContainText('imported successfully');
  });

  test('should handle bulk operations', async ({ page }) => {
    // Select multiple customers
    await page.check('[data-testid="select-customer-0"]');
    await page.check('[data-testid="select-customer-1"]');
    
    // Should show bulk action bar
    await expect(page.locator('[data-testid="bulk-actions"]')).toBeVisible();
    
    // Bulk status update
    await page.click('[data-testid="bulk-status-button"]');
    await page.click('[data-testid="status-inactive"]');
    
    await expect(page.locator('[data-testid="success-toast"]')).toContainText('customers updated');
  });

  test('should validate customer data integrity', async ({ page }) => {
    await page.click('[data-testid="add-customer-button"]');
    
    // Try to add customer with duplicate email
    await page.fill('input[name="name"]', 'Duplicate User');
    await page.fill('input[name="email"]', 'admin@example.com'); // Existing email
    await page.fill('input[name="phone"]', '+1-555-0000');
    
    await page.click('[data-testid="save-customer-button"]');
    
    await expect(page.locator('[data-testid="error-toast"]')).toContainText('Email already exists');
  });

  test('should display customer activity timeline', async ({ page }) => {
    await page.click('[data-testid="view-customer-0"]');
    
    // Navigate to activity tab
    await page.click('[data-testid="activity-tab"]');
    
    // Should show activity timeline
    await expect(page.locator('[data-testid="activity-timeline"]')).toBeVisible();
    await expect(page.locator('[data-testid="activity-item"]').first()).toBeVisible();
  });

  test('should manage customer tags', async ({ page }) => {
    await page.click('[data-testid="view-customer-0"]');
    
    // Add new tag
    await page.click('[data-testid="add-tag-button"]');
    await page.fill('[data-testid="tag-input"]', 'VIP');
    await page.press('[data-testid="tag-input"]', 'Enter');
    
    // Should see new tag
    await expect(page.locator('[data-testid="customer-tag-vip"]')).toBeVisible();
    
    // Remove tag
    await page.click('[data-testid="remove-tag-vip"]');
    await expect(page.locator('[data-testid="customer-tag-vip"]')).not.toBeVisible();
  });

  test('should handle network errors gracefully', async ({ page }) => {
    // Simulate network failure
    await page.route('**/api/v1/customers', route => route.abort());
    
    await page.reload();
    
    // Should show error message
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Failed to load customers');
    await expect(page.locator('[data-testid="retry-button"]')).toBeVisible();
    
    // Restore network and retry
    await page.unroute('**/api/v1/customers');
    await page.click('[data-testid="retry-button"]');
    
    // Should load successfully
    await expect(page.locator('[data-testid="customers-table"]')).toBeVisible();
  });
});

test.describe('Customer Properties Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login and navigate to customer details
    await page.goto('/');
    await page.fill('input[name="email"]', 'admin@example.com');
    await page.fill('input[name="password"]', 'password123');
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');
    
    await page.click('text=Customers');
    await page.click('[data-testid="view-customer-0"]');
    await page.click('[data-testid="properties-tab"]');
  });

  test('should add property to customer', async ({ page }) => {
    await page.click('[data-testid="add-property-button"]');
    
    // Fill property form
    await page.fill('input[name="address"]', '456 Oak Street');
    await page.fill('input[name="city"]', 'Springfield');
    await page.fill('input[name="state"]', 'IL');
    await page.fill('input[name="zipCode"]', '62702');
    await page.fill('input[name="propertyType"]', 'Residential');
    await page.fill('input[name="squareFootage"]', '5000');
    
    await page.click('[data-testid="save-property-button"]');
    
    await expect(page.locator('[data-testid="success-toast"]')).toContainText('Property added successfully');
    await expect(page.locator('text=456 Oak Street')).toBeVisible();
  });

  test('should edit customer property', async ({ page }) => {
    await page.click('[data-testid="edit-property-0"]');
    
    await page.fill('input[name="squareFootage"]', '6000');
    await page.click('[data-testid="save-property-button"]');
    
    await expect(page.locator('[data-testid="success-toast"]')).toContainText('Property updated successfully');
  });

  test('should delete customer property', async ({ page }) => {
    await page.click('[data-testid="delete-property-0"]');
    
    await expect(page.locator('[data-testid="confirmation-dialog"]')).toBeVisible();
    await page.click('[data-testid="confirm-delete-button"]');
    
    await expect(page.locator('[data-testid="success-toast"]')).toContainText('Property deleted successfully');
  });
});