import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('Repository Management', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication for all repository tests
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
            email: testUser.email
          }
        })
      });
    });

    // Set authentication state
    await page.addInitScript(() => {
      window.localStorage.setItem('auth_token', 'mock-jwt-token');
    });
  });

  test('should display repositories page with empty state', async ({ page }) => {
    // Mock empty repositories response
    await page.route('**/api/v1/repositories**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [],
          pagination: {
            page: 1,
            per_page: 30,
            total: 0
          }
        })
      });
    });

    await page.goto('/repositories');
    
    // Should show empty state
    await expect(page.locator('text=No repositories yet')).toBeVisible();
    await expect(page.locator('a[href="/repositories/new"]')).toBeVisible();
  });

  test('should display list of repositories', async ({ page }) => {
    // Mock repositories response with data
    await page.route('**/api/v1/repositories**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
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
          ],
          pagination: {
            page: 1,
            per_page: 30,
            total: 2
          }
        })
      });
    });

    await page.goto('/repositories');
    
    // Should display repository list
    await expect(page.locator('text=awesome-project')).toBeVisible();
    await expect(page.locator('text=api-service')).toBeVisible();
    await expect(page.locator('text=TypeScript')).toBeVisible();
    await expect(page.locator('text=Go')).toBeVisible();
    await expect(page.locator('text=Private')).toBeVisible();
  });

  test('should navigate to create new repository', async ({ page }) => {
    await page.goto('/repositories');
    
    // Click new repository button
    await page.click('a[href="/repositories/new"]');
    
    // Should navigate to create repository page
    await expect(page).toHaveURL('/repositories/new');
    await expect(page.locator('h1')).toContainText('Create a new repository');
  });

  test('should navigate to repository details', async ({ page }) => {
    // Mock repositories response
    await page.route('**/api/v1/repositories**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              id: '1',
              name: 'awesome-project',
              full_name: 'testuser/awesome-project',
              description: 'An awesome project',
              private: false,
              language: 'TypeScript',
              stargazers_count: 42,
              forks_count: 8,
              updated_at: '2024-07-20T10:00:00Z',
            }
          ]
        })
      });
    });

    // Mock individual repository response
    await page.route('**/api/v1/repositories/testuser/awesome-project', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: '1',
            name: 'awesome-project',
            full_name: 'testuser/awesome-project',
            description: 'An awesome project',
            private: false,
            language: 'TypeScript',
            default_branch: 'main',
            clone_url: 'https://hub.a5c.ai/testuser/awesome-project.git',
            stargazers_count: 42,
            forks_count: 8,
  
            updated_at: '2024-07-20T10:00:00Z',
          }
        })
      });
    });

    await page.goto('/repositories');
    
    // Click on repository link
    await page.click('text=awesome-project');
    
    // Should navigate to repository details
    await expect(page).toHaveURL('/repositories/testuser/awesome-project');
    await expect(page.locator('h1')).toContainText('awesome-project');
  });

  test('should search repositories', async ({ page }) => {
    // Mock search results
    await page.route('**/api/v1/repositories?*search=*', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              id: '1',
              name: 'api-service',
              full_name: 'testuser/api-service',
              description: 'RESTful API service',
              private: false,
              language: 'Go',
              stargazers_count: 15,
              forks_count: 3,
              updated_at: '2024-07-19T15:30:00Z',
            }
          ]
        })
      });
    });

    await page.goto('/repositories');
    
    // Search for repositories
    if (await page.locator('input[type="search"]').count() > 0) {
      await page.fill('input[type="search"]', 'api');
      await page.press('input[type="search"]', 'Enter');
      
      // Should show search results
      await expect(page.locator('text=api-service')).toBeVisible();
    }
  });
});