import { test, expect } from '@playwright/test';
import { loginUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('Admin User Management', () => {
  // Mock admin credentials
  const adminUser = {
    email: 'admin@example.com',
    password: 'AdminPassword123!'
  };

  // Mock user data for testing
  const mockUsers = [
    { id: 1, username: 'john_doe', email: 'john@example.com', name: 'John Doe', status: 'active', role: 'user' },
    { id: 2, username: 'jane_smith', email: 'jane@example.com', name: 'Jane Smith', status: 'active', role: 'admin' },
    { id: 3, username: 'bob_wilson', email: 'bob@example.com', name: 'Bob Wilson', status: 'inactive', role: 'user' }
  ];

  test.beforeEach(async ({ page }) => {
    // Login as admin user
    await loginUser(page, adminUser.email, adminUser.password);
    
    // Mock API responses for user management endpoints
    await page.route('**/api/v1/admin/users/stats', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          total_users: mockUsers.length,
          active_users: mockUsers.filter(u => u.status === 'active').length,
          inactive_users: mockUsers.filter(u => u.status === 'inactive').length,
          admin_users: mockUsers.filter(u => u.role === 'admin').length,
          verified_users: mockUsers.length,
          mfa_enabled_users: 0,
          users_this_month: mockUsers.length,
          users_last_month: mockUsers.length,
          logins_this_week: 100
        })
      });
    });
    await page.route('**/api/v1/admin/users*', route => {
      const { method } = route.request();
      if (method() === 'GET') {
        route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ users: mockUsers, pagination: { total: mockUsers.length, total_pages: 1 } })
        });
      } else if (['POST', 'PATCH', 'DELETE'].includes(method())) {
        route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
      } else {
        route.continue();
      }
    });
  });

  test.describe('User Account Administration', () => {
    test('should display user management page with user list', async ({ page }) => {
      // Navigate to user management page
      // Note: This route may need to be created if it doesn't exist
      await page.goto('/admin/users');
      
      await waitForLoadingToComplete(page);

      // Verify page header and title
      await expect(page.locator('h1')).toContainText('User Management');
      
      // Check for search functionality
      await expect(page.locator('input[placeholder*="Search users"]')).toBeVisible();
      
      // Verify user table headers
      await expect(page.locator('text=Username')).toBeVisible();
      await expect(page.locator('text=Email')).toBeVisible();
      await expect(page.locator('text=Status')).toBeVisible();
      await expect(page.locator('text=Role')).toBeVisible();
      await expect(page.locator('text=Actions')).toBeVisible();
    });

    test('should allow activating and deactivating user accounts', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Find an active user and deactivate them
      const userRow = page.locator('[data-testid="user-row-1"]');
      await expect(userRow.locator('[data-testid="user-status"]')).toContainText('Active');
      
      // Click deactivate button
      await userRow.locator('[data-testid="deactivate-user-btn"]').click();
      
      // Confirm deactivation in modal/dialog
      await page.locator('[data-testid="confirm-deactivate"]').click();
      
      // Verify status changed
      await expect(userRow.locator('[data-testid="user-status"]')).toContainText('Inactive');
      
      // Test reactivation
      await userRow.locator('[data-testid="activate-user-btn"]').click();
      await page.locator('[data-testid="confirm-activate"]').click();
      await expect(userRow.locator('[data-testid="user-status"]')).toContainText('Active');
    });

    test('should allow deleting user accounts with confirmation', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Count initial users
      const initialUserCount = await page.locator('[data-testid^="user-row-"]').count();
      
      // Delete a user
      const userToDelete = page.locator('[data-testid="user-row-3"]');
      await userToDelete.locator('[data-testid="delete-user-btn"]').click();
      
      // Confirm deletion with additional verification steps
      await expect(page.locator('[data-testid="delete-confirmation-modal"]')).toBeVisible();
      await expect(page.locator('text=Are you sure you want to delete this user?')).toBeVisible();
      
      // Type username to confirm
      await page.fill('[data-testid="confirm-username-input"]', 'bob_wilson');
      await page.click('[data-testid="confirm-delete-btn"]');
      
      // Wait for deletion to complete
      await waitForLoadingToComplete(page);
      
      // Verify user was removed
      const finalUserCount = await page.locator('[data-testid^="user-row-"]').count();
      expect(finalUserCount).toBe(initialUserCount - 1);
    });

    test('should display user search and filtering functionality', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Test search functionality
      await page.fill('input[placeholder*="Search users"]', 'john');
      await page.press('input[placeholder*="Search users"]', 'Enter');
      
      // Verify filtered results
      await expect(page.locator('[data-testid="user-row-1"]')).toBeVisible();
      await expect(page.locator('text=John Doe')).toBeVisible();
      
      // Clear search
      await page.fill('input[placeholder*="Search users"]', '');
      await page.press('input[placeholder*="Search users"]', 'Enter');
      
      // Test status filter
      await page.selectOption('[data-testid="status-filter"]', 'active');
      await expect(page.locator('[data-testid="user-row-1"]')).toBeVisible();
      await expect(page.locator('[data-testid="user-row-2"]')).toBeVisible();
      
      // Test role filter
      await page.selectOption('[data-testid="role-filter"]', 'admin');
      await expect(page.locator('[data-testid="user-row-2"]')).toBeVisible();
      await expect(page.locator('text=Jane Smith')).toBeVisible();
    });
  });

  test.describe('User Role and Permission Management', () => {
    test('should allow changing user roles', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Select a user and change their role
      const userRow = page.locator('[data-testid="user-row-1"]');
      await userRow.locator('[data-testid="edit-user-btn"]').click();
      
      // Verify edit modal opens
      await expect(page.locator('[data-testid="edit-user-modal"]')).toBeVisible();
      
      // Change role from user to admin
      await page.selectOption('[data-testid="user-role-select"]', 'admin');
      
      // Save changes
      await page.click('[data-testid="save-user-changes"]');
      
      // Verify role changed in the table
      await expect(userRow.locator('[data-testid="user-role"]')).toContainText('Admin');
    });

    test('should display user permission matrix', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Open user details/permissions view
      const userRow = page.locator('[data-testid="user-row-1"]');
      await userRow.locator('[data-testid="view-permissions-btn"]').click();
      
      // Verify permissions modal/page
      await expect(page.locator('[data-testid="user-permissions-modal"]')).toBeVisible();
      
      // Check for common permissions
      await expect(page.locator('text=Repository Access')).toBeVisible();
      await expect(page.locator('text=Issue Management')).toBeVisible();
      await expect(page.locator('text=User Management')).toBeVisible();
      await expect(page.locator('text=System Administration')).toBeVisible();
    });

    test('should handle bulk user operations', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Select multiple users using checkboxes
      await page.check('[data-testid="user-checkbox-1"]');
      await page.check('[data-testid="user-checkbox-3"]');
      
      // Verify bulk actions become available
      await expect(page.locator('[data-testid="bulk-actions-bar"]')).toBeVisible();
      await expect(page.locator('[data-testid="bulk-deactivate-btn"]')).toBeVisible();
      await expect(page.locator('[data-testid="bulk-role-change-btn"]')).toBeVisible();
      
      // Perform bulk deactivation
      await page.click('[data-testid="bulk-deactivate-btn"]');
      await page.click('[data-testid="confirm-bulk-action"]');
      
      // Verify users were deactivated
      await expect(page.locator('[data-testid="user-row-1"] [data-testid="user-status"]')).toContainText('Inactive');
      await expect(page.locator('[data-testid="user-row-3"] [data-testid="user-status"]')).toContainText('Inactive');
    });
  });

  test.describe('User Activity Monitoring and Audit Logs', () => {
    test('should display user activity logs', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Open user activity view
      const userRow = page.locator('[data-testid="user-row-1"]');
      await userRow.locator('[data-testid="view-activity-btn"]').click();
      
      // Verify activity log modal/page
      await expect(page.locator('[data-testid="user-activity-modal"]')).toBeVisible();
      await expect(page.locator('text=Recent Activity')).toBeVisible();
      
      // Check for activity entries
      await expect(page.locator('[data-testid="activity-entry"]').first()).toBeVisible();
      
      // Verify activity details
      await expect(page.locator('text=Login')).toBeVisible();
      await expect(page.locator('text=Repository Access')).toBeVisible();
      await expect(page.locator('text=Profile Update')).toBeVisible();
    });

    test('should provide audit trail for user changes', async ({ page }) => {
      await page.goto('/admin/audit-logs');
      await waitForLoadingToComplete(page);

      // Verify audit log page
      await expect(page.locator('h1')).toContainText('Audit Logs');
      
      // Check for audit entries
      await expect(page.locator('[data-testid="audit-entry"]').first()).toBeVisible();
      
      // Filter by user management actions
      await page.selectOption('[data-testid="action-filter"]', 'user_management');
      
      // Verify filtered results show user-related actions
      await expect(page.locator('text=User Created')).toBeVisible();
      await expect(page.locator('text=Role Changed')).toBeVisible();
      await expect(page.locator('text=User Deactivated')).toBeVisible();
    });

    test('should track failed login attempts', async ({ page }) => {
      await page.goto('/admin/security');
      await waitForLoadingToComplete(page);

      // Navigate to security monitoring section
      await expect(page.locator('text=Failed Login Attempts')).toBeVisible();
      
      // Check for failed login entries
      await expect(page.locator('[data-testid="failed-login-entry"]').first()).toBeVisible();
      
      // Verify details include IP, timestamp, and user
      await expect(page.locator('text=IP Address')).toBeVisible();
      await expect(page.locator('text=Timestamp')).toBeVisible();
      await expect(page.locator('text=Attempted Username')).toBeVisible();
    });
  });

  test.describe('Impersonation and Support Functionality', () => {
    test('should allow admin to impersonate users for support', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Impersonate a user
      const userRow = page.locator('[data-testid="user-row-1"]');
      await userRow.locator('[data-testid="impersonate-user-btn"]').click();
      
      // Confirm impersonation
      await expect(page.locator('[data-testid="impersonate-modal"]')).toBeVisible();
      await page.click('[data-testid="confirm-impersonate"]');
      
      // Verify impersonation mode is active
      await expect(page.locator('[data-testid="impersonation-banner"]')).toBeVisible();
      await expect(page.locator('text=You are impersonating John Doe')).toBeVisible();
      
      // Test exit impersonation
      await page.click('[data-testid="exit-impersonation"]');
      await expect(page.locator('[data-testid="impersonation-banner"]')).not.toBeVisible();
    });

    test('should provide support tools for user assistance', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Open user support tools
      const userRow = page.locator('[data-testid="user-row-1"]');
      await userRow.locator('[data-testid="support-tools-btn"]').click();
      
      // Verify support tools modal
      await expect(page.locator('[data-testid="support-tools-modal"]')).toBeVisible();
      
      // Check available support actions
      await expect(page.locator('[data-testid="reset-password-btn"]')).toBeVisible();
      await expect(page.locator('[data-testid="unlock-account-btn"]')).toBeVisible();
      await expect(page.locator('[data-testid="reset-mfa-btn"]')).toBeVisible();
      await expect(page.locator('[data-testid="send-verification-email-btn"]')).toBeVisible();
    });
  });

  test.describe('User Management Error Handling', () => {
    test('should handle API errors gracefully', async ({ page }) => {
      // Navigate to user management page
      await page.goto('/admin/users');
      
      // In a real implementation, you would mock API failures here
      // For now, we'll test that error states can be displayed
      
      // Verify error handling UI exists
      const errorContainer = page.locator('[data-testid="error-container"]');
      const retryButton = page.locator('[data-testid="retry-button"]');
      
      // These elements might not be visible unless there's an actual error
      // In a complete implementation, you would mock network failures
      // and verify appropriate error messages are shown
    });

    test('should validate user input forms', async ({ page }) => {
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Try to create a user with invalid data
      await page.click('[data-testid="add-user-btn"]');
      await expect(page.locator('[data-testid="add-user-modal"]')).toBeVisible();
      
      // Submit without required fields
      await page.click('[data-testid="save-new-user"]');
      
      // Verify validation errors
      await expect(page.locator('text=Email is required')).toBeVisible();
      await expect(page.locator('text=Username is required')).toBeVisible();
      
      // Test email format validation
      await page.fill('[data-testid="email-input"]', 'invalid-email');
      await page.click('[data-testid="save-new-user"]');
      await expect(page.locator('text=Please enter a valid email')).toBeVisible();
    });
  });

  test.describe('Mobile User Management', () => {
    test('should be responsive on mobile devices', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Verify key elements are accessible on mobile
      await expect(page.locator('h1')).toContainText('User Management');
      
      // Check that user list adapts to mobile view
      await expect(page.locator('[data-testid="user-row-1"]')).toBeVisible();
      
      // Verify search functionality works on mobile
      await expect(page.locator('input[placeholder*="Search users"]')).toBeVisible();
    });

    test('should handle mobile interactions for user management', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      await page.goto('/admin/users');
      await waitForLoadingToComplete(page);

      // Test mobile menu/actions
      const userRow = page.locator('[data-testid="user-row-1"]');
      await userRow.locator('[data-testid="mobile-actions-menu"]').click();
      
      // Verify mobile action menu
      await expect(page.locator('[data-testid="mobile-action-edit"]')).toBeVisible();
      await expect(page.locator('[data-testid="mobile-action-deactivate"]')).toBeVisible();
    });
  });
});
