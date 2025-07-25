import { test, expect } from '@playwright/test';
import { loginUser, testUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('User Profile Management', () => {
  test.beforeEach(async ({ page }) => {
    // Start fresh for each test
    await page.context().clearCookies();
    await page.context().clearPermissions();
    
    // Setup common API mocks
    await page.route('**/api/v1/auth/me', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: '1',
            name: testUser.name,
            username: testUser.username,
            email: testUser.email,
            bio: 'Test user bio',
            website: 'https://example.com',
            location: 'San Francisco, CA',
            company: 'Test Company',
            avatar_url: 'https://example.com/avatar.jpg',
            created_at: '2023-01-01T00:00:00Z'
          }
        })
      });
    });

    // Mock user update API
    await page.route('**/api/v1/user', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Profile updated successfully'
          })
        });
      }
    });

    // Mock repositories API for dashboard
    await page.route('**/api/v1/repositories**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: []
        })
      });
    });
  });

  test('should display user profile information correctly', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Verify we're on the settings page
    await expect(page.locator('h1')).toContainText('Settings');
    
    // Should default to profile tab
    await expect(page.locator('[data-testid="profile-tab"]').or(page.locator('button:has-text("Profile")'))).toHaveClass(/bg-primary/);
    
    // Verify profile information is loaded
    await expect(page.locator('input[value="' + testUser.name + '"]')).toBeVisible();
    await expect(page.locator('input[value="' + testUser.username + '"]')).toBeVisible();
    await expect(page.locator('input[value="' + testUser.email + '"]')).toBeVisible();
  });

  test('should edit and save profile information', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Update profile fields
    const newName = 'Updated Test User';
    const newBio = 'Updated bio description';
    const newWebsite = 'https://updated-website.com';
    const newLocation = 'New York, NY';
    const newCompany = 'Updated Company';
    
    // Fill in the form fields
    await page.fill('input[placeholder="Your full name"]', newName);
    await page.fill('textarea[placeholder="Tell us about yourself..."]', newBio);
    await page.fill('input[placeholder="https://yourwebsite.com"]', newWebsite);
    await page.fill('input[placeholder="City, Country"]', newLocation);
    await page.fill('input[placeholder="Your company or organization"]', newCompany);
    
    // Save changes
    await page.click('button:has-text("Save Changes")');
    
    // Verify saving state
    await expect(page.locator('button:has-text("Saving...")')).toBeVisible();
    
    // Wait for save to complete
    await waitForLoadingToComplete(page);
    await expect(page.locator('button:has-text("Save Changes")')).toBeVisible();
  });

  test('should handle avatar change interaction', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Check avatar is displayed
    await expect(page.locator('img[alt*="User"], img[alt*="' + testUser.username + '"]')).toBeVisible();
    
    // Click change avatar button
    await page.click('button:has-text("Change Avatar")');
    
    // This would typically open a file picker or modal
    // For now, we just verify the button is clickable
    await expect(page.locator('button:has-text("Change Avatar")')).toBeVisible();
  });

  test('should validate required fields', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Clear required fields
    await page.fill('input[placeholder="Your full name"]', '');
    await page.fill('input[placeholder="your.email@example.com"]', '');
    
    // Try to save
    await page.click('button:has-text("Save Changes")');
    
    // Verify form validation (implementation-dependent)
    // This test would need to be adapted based on actual validation behavior
  });

  test('should navigate between settings tabs', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Test navigation to Account tab
    await page.click('button:has-text("Account")');
    await expect(page.locator('h3:has-text("Account Settings")')).toBeVisible();
    
    // Test navigation to Security tab
    await page.click('button:has-text("Security")');
    await expect(page.locator('h3:has-text("Password & Authentication")')).toBeVisible();
    
    // Test navigation to Notifications tab
    await page.click('button:has-text("Notifications")');
    await expect(page.locator('h3:has-text("Email Notifications")')).toBeVisible();
    
    // Test navigation back to Profile tab
    await page.click('button:has-text("Profile")');
    await expect(page.locator('h3:has-text("Profile Information")')).toBeVisible();
  });

  test('should display account status and type', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Account tab
    await page.click('button:has-text("Account")');
    
    // Verify account information
    await expect(page.locator('text=Account Type')).toBeVisible();
    await expect(page.locator('text=Free account with basic features')).toBeVisible();
    await expect(page.locator('text=Account Status')).toBeVisible();
    await expect(page.locator('text=Active')).toBeVisible();
  });

  test('should display theme toggle', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Account tab
    await page.click('button:has-text("Account")');
    
    // Verify theme section exists
    await expect(page.locator('text=Theme')).toBeVisible();
    await expect(page.locator('text=Choose your preferred color scheme')).toBeVisible();
    
    // Theme toggle should be present
    await expect(page.locator('[data-testid="theme-toggle"]').or(page.locator('button[aria-label*="theme"], button[title*="theme"]'))).toBeVisible();
  });

  test('should show danger zone with delete account option', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Account tab
    await page.click('button:has-text("Account")');
    
    // Verify danger zone exists
    await expect(page.locator('h3:has-text("Danger Zone")')).toBeVisible();
    await expect(page.locator('text=Delete Account')).toBeVisible();
    await expect(page.locator('text=Permanently delete your account and all associated data')).toBeVisible();
    await expect(page.locator('button:has-text("Delete Account")')).toBeVisible();
  });

  test('should handle mobile responsive design', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    await loginUser(page);
    await page.goto('/settings');
    
    // Verify page renders correctly on mobile
    await expect(page.locator('h1:has-text("Settings")')).toBeVisible();
    
    // Check that navigation is accessible (may be collapsible on mobile)
    await expect(page.locator('button:has-text("Profile")')).toBeVisible();
    
    // Form fields should be responsive
    await expect(page.locator('input[placeholder="Your full name"]')).toBeVisible();
    
    // Save button should be accessible
    await expect(page.locator('button:has-text("Save Changes")')).toBeVisible();
  });

  test('should display proper loading states', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Fill in a field
    await page.fill('input[placeholder="Your full name"]', 'New Name');
    
    // Mock a delayed save response
    await page.route('**/api/v1/user', async route => {
      // Add delay before fulfilling
      await new Promise(resolve => setTimeout(resolve, 1000));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Profile updated successfully'
        })
      });
    });
    
    // Click save and verify loading state
    await page.click('button:has-text("Save Changes")');
    await expect(page.locator('button:has-text("Saving...")')).toBeVisible();
    
    // Wait for loading to complete
    await waitForLoadingToComplete(page);
    await expect(page.locator('button:has-text("Save Changes")')).toBeVisible();
  });

  test('should handle form validation errors gracefully', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock validation error response
    await page.route('**/api/v1/user', async route => {
      await route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'Email is already taken',
          validation_errors: {
            email: ['This email address is already in use']
          }
        })
      });
    });
    
    // Try to save with duplicate email
    await page.fill('input[placeholder="your.email@example.com"]', 'existing@example.com');
    await page.click('button:has-text("Save Changes")');
    
    // Should show error message (implementation-dependent)
    // This test would need to be adapted based on actual error handling
  });

  test('should maintain form state when switching tabs', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Make changes to profile form
    const newName = 'Modified Name';
    await page.fill('input[placeholder="Your full name"]', newName);
    
    // Switch to another tab
    await page.click('button:has-text("Account")');
    await expect(page.locator('h3:has-text("Account Settings")')).toBeVisible();
    
    // Switch back to Profile tab
    await page.click('button:has-text("Profile")');
    
    // Verify the form data is preserved
    await expect(page.locator('input[value="' + newName + '"]')).toBeVisible();
  });
});

test.describe('Public Profile Display', () => {
  test.beforeEach(async ({ page }) => {
    // Mock public profile API
    await page.route('**/api/v1/users/' + testUser.username, async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: '1',
            name: testUser.name,
            username: testUser.username,
            bio: 'Test user bio',
            website: 'https://example.com',
            location: 'San Francisco, CA',
            company: 'Test Company',
            avatar_url: 'https://example.com/avatar.jpg',
            created_at: '2023-01-01T00:00:00Z',
            public_repos: 5,
            followers: 10,
            following: 15
          }
        })
      });
    });

    // Mock user repositories for public profile
    await page.route('**/api/v1/users/' + testUser.username + '/repositories', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              id: '1',
              name: 'test-repo',
              description: 'A test repository',
              is_private: false,
              language: 'TypeScript',
              stars_count: 5,
              updated_at: '2023-12-01T00:00:00Z'
            }
          ]
        })
      });
    });
  });

  test('should display public profile page correctly', async ({ page }) => {
    await page.goto('/users/' + testUser.username);
    
    // Verify profile information is displayed
    await expect(page.locator('h1:has-text("' + testUser.name + '")')).toBeVisible();
    await expect(page.locator('text=@' + testUser.username)).toBeVisible();
    await expect(page.locator('text=Test user bio')).toBeVisible();
    await expect(page.locator('text=San Francisco, CA')).toBeVisible();
    await expect(page.locator('text=Test Company')).toBeVisible();
    
    // Verify stats are displayed
    await expect(page.locator('text=5').first()).toBeVisible(); // repos
    await expect(page.locator('text=10').first()).toBeVisible(); // followers
    await expect(page.locator('text=15').first()).toBeVisible(); // following
  });

  test('should display user repositories on public profile', async ({ page }) => {
    await page.goto('/users/' + testUser.username);
    
    // Verify repository is listed
    await expect(page.locator('text=test-repo')).toBeVisible();
    await expect(page.locator('text=A test repository')).toBeVisible();
    await expect(page.locator('text=TypeScript')).toBeVisible();
    await expect(page.locator('text=5').first()).toBeVisible(); // stars
  });

  test('should handle profile not found gracefully', async ({ page }) => {
    // Mock 404 response
    await page.route('**/api/v1/users/nonexistent', async route => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'User not found'
        })
      });
    });
    
    await page.goto('/users/nonexistent');
    
    // Should show 404 or error message
    await expect(page.locator('text=not found, text=404, text=User not found')).toBeVisible();
  });

  test('should allow navigation from public profile to repositories', async ({ page }) => {
    await page.goto('/users/' + testUser.username);
    
    // Click on repository link
    await page.click('text=test-repo');
    
    // Should navigate to repository page
    await expect(page).toHaveURL(/\/repositories\//);
  });
});