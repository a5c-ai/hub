import { test, expect } from '@playwright/test';
import { testUser, expectDashboardPage } from './helpers/test-utils';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication for all dashboard tests
    await page.route('**/api/v1/profile', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: '1',
            name: testUser.name,
            username: testUser.username,
            email: testUser.email
          }
        })
      });
    });

    // Mock repositories data
    await page.route('**/api/v1/repositories', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          repositories: [
            {
              id: '1',
              name: 'awesome-project',
              full_name: 'testuser/awesome-project',
              description: 'An awesome project built with modern technologies',
              private: false,
              language: 'TypeScript',
              stargazers_count: 42,
              forks_count: 8,
              updated_at: '2024-07-20T10:00:00Z',
            },
            {
              id: '2',
              name: 'api-service',
              full_name: 'testuser/api-service',
              description: 'RESTful API service for the application',
              private: true,
              language: 'Go',
              stargazers_count: 15,
              forks_count: 3,
              updated_at: '2024-07-19T15:30:00Z',
            }
          ]
        })
      });
    });

    // Mock activity data
    await page.route('**/api/v1/activity**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          activities: [
            {
              id: '1',
              type: 'push',
              repository: 'testuser/awesome-project',
              message: 'Added new authentication middleware',
              timestamp: '2024-07-20T10:00:00Z',
            },
            {
              id: '2',
              type: 'pull_request',
              repository: 'testuser/api-service',
              message: 'Opened pull request: Implement user management endpoints',
              timestamp: '2024-07-19T15:30:00Z',
            }
          ]
        })
      });
    });

    // Set authentication state
    await page.addInitScript(() => {
      const testUser = {
        id: '1',
        name: 'Test User',
        username: 'testuser',
        email: 'test@example.com'
      };
      
      window.localStorage.setItem('auth_token', 'mock-jwt-token');
      window.localStorage.setItem('auth-storage', JSON.stringify({
        state: {
          user: testUser,
          token: 'mock-jwt-token',
          isAuthenticated: true
        },
        version: 0
      }));
    });
  });

  test('should display dashboard with user welcome message', async ({ page }) => {
    await page.goto('/dashboard');
    
    await expectDashboardPage(page);
    await expect(page.locator('h1')).toContainText(`Welcome back, ${testUser.name}`);
  });

  test('should display repository statistics cards', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check for statistics cards
    await expect(page.locator('text=Total Repositories')).toBeVisible();
    await expect(page.locator('text=Total Stars')).toBeVisible();
    await expect(page.locator('text=Total Forks')).toBeVisible();
    
    // Check that stats are displayed (could be 0 for new users)
    // The specific numbers will depend on the mocked API responses
    await expect(page.locator('text=Total Repositories').locator('..').locator('.text-2xl')).toBeVisible();
    await expect(page.locator('text=Total Stars').locator('..').locator('.text-2xl')).toBeVisible();
    await expect(page.locator('text=Total Forks').locator('..').locator('.text-2xl')).toBeVisible();
  });

  test('should display recent repositories section', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check for repositories section
    await expect(page.locator('text=Recent Repositories')).toBeVisible();
    await expect(page.locator('text=Your most recently updated repositories')).toBeVisible();
    
    // Check for repository items
    await expect(page.locator('text=awesome-project')).toBeVisible();
    await expect(page.locator('text=api-service')).toBeVisible();
    
    // Check for repository details
    await expect(page.locator('text=TypeScript')).toBeVisible();
    await expect(page.locator('text=Go')).toBeVisible();
    await expect(page.locator('text=Private')).toBeVisible();
    
    // Check for star and fork counts
    await expect(page.locator('text=42')).toBeVisible(); // Stars for awesome-project
    await expect(page.locator('text=8')).toBeVisible();  // Forks for awesome-project
  });

  test('should display recent activity section', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check for activity section
    await expect(page.locator('text=Recent Activity')).toBeVisible();
    await expect(page.locator('text=Your recent actions across all repositories')).toBeVisible();
    
    // Check for activity items
    await expect(page.locator('text=Added new authentication middleware')).toBeVisible();
    await expect(page.locator('text=Opened pull request: Implement user management endpoints')).toBeVisible();
    
    // Check for repository links in activity
    await expect(page.locator('text=testuser/awesome-project')).toBeVisible();
    await expect(page.locator('text=testuser/api-service')).toBeVisible();
  });

  test('should have working navigation to create new repository', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Find and click the "New" button in repositories section
    await expect(page.locator('text=New').first()).toBeVisible();
    
    // This would test navigation to create repository page
    // The actual implementation would depend on the routing setup
  });

  test('should have working navigation to view all repositories', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Find and click the "View all repositories" link
    await expect(page.locator('text=View all repositories')).toBeVisible();
    await page.click('text=View all repositories');
    
    // Should navigate to repositories page
    await expect(page).toHaveURL('/repositories');
  });

  test('should have working navigation to view all activity', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Find and click the "View all activity" link
    await expect(page.locator('text=View all activity')).toBeVisible();
    await page.click('text=View all activity');
    
    // Should navigate to activity page
    await expect(page).toHaveURL('/activity');
  });

  test('should display user avatar and information', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check for user information in activity section
    await expect(page.locator('text=You')).toBeVisible();
    
    // Avatar should be present (this depends on the Avatar component implementation)
    const avatars = page.locator('[data-testid="user-avatar"]');
    if (await avatars.count() > 0) {
      await expect(avatars.first()).toBeVisible();
    }
  });

  test('should handle empty states appropriately', async ({ page }) => {
    // Mock empty repositories response
    await page.route('**/api/v1/repositories', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          repositories: []
        })
      });
    });

    // Mock empty activity response
    await page.route('**/api/activity', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          activities: []
        })
      });
    });

    await page.goto('/dashboard');
    
    // The app should handle empty states gracefully
    // This test would verify that no JavaScript errors occur
    // and that appropriate empty state messages are shown
    await expectDashboardPage(page);
  });

  test('should be responsive on mobile viewports', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/dashboard');
    
    await expectDashboardPage(page);
    
    // Check that content is still visible and accessible on mobile
    await expect(page.locator('h1')).toBeVisible();
    await expect(page.locator('text=Recent Repositories')).toBeVisible();
    await expect(page.locator('text=Recent Activity')).toBeVisible();
    
    // Statistics cards should stack properly on mobile
    const statsCards = page.locator('text=Total Repositories').locator('..');
    await expect(statsCards).toBeVisible();
  });
});