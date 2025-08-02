import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('Repository Code Browser & File Management', () => {
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

    // Mock repository data
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
            description: 'An awesome project built with modern technologies',
            private: false,
            language: 'TypeScript',
            default_branch: 'main',
            clone_url: 'https://hub.a5c.ai/testuser/awesome-project.git',
            stargazers_count: 42,
            forks_count: 8,
    
            updated_at: '2024-07-20T10:00:00Z',
            owner: {
              username: 'testuser',
              avatar_url: 'https://avatars.githubusercontent.com/u/1?v=4'
            }
          }
        })
      });
    });

    // Mock branches response
    await page.route('**/api/v1/repositories/testuser/awesome-project/branches', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            { name: 'main', sha: 'abc123' },
            { name: 'develop', sha: 'def456' },
            { name: 'feature/new-ui', sha: 'ghi789' }
          ]
        })
      });
    });
  });

  test('should display repository file tree on main branch', async ({ page }) => {
    // Mock tree response for root directory
    await page.route('**/api/v1/repositories/testuser/awesome-project/tree**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            entries: [
              {
                name: 'src',
                path: 'src',
                type: 'tree',
                size: 0
              },
              {
                name: 'package.json',
                path: 'package.json',
                type: 'blob',
                size: 1024
              },
              {
                name: 'README.md',
                path: 'README.md',
                type: 'blob',
                size: 2048
              }
            ]
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project');
    
    // Should display repository name and metadata
    await expect(page.locator('h1')).toContainText('awesome-project');
    await expect(page.locator('text=TypeScript')).toBeVisible();
    
    // Should display file tree
    await expect(page.locator('text=src')).toBeVisible();
    await expect(page.locator('text=package.json')).toBeVisible();
    await expect(page.locator('text=README.md')).toBeVisible();
  });

  test('should navigate through directory structure', async ({ page }) => {
    // Mock tree responses for navigation
    await page.route('**/api/v1/repositories/testuser/awesome-project/tree**', async route => {
      const url = new URL(route.request().url());
      const path = url.searchParams.get('path') || '';
      
      if (path === '') {
        // Root directory
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              entries: [
                { name: 'src', path: 'src', type: 'tree', size: 0 }
              ]
            }
          })
        });
      } else if (path === 'src') {
        // src directory
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              entries: [
                { name: 'index.ts', path: 'src/index.ts', type: 'blob', size: 512 }
              ]
            }
          })
        });
      }
    });

    await page.goto('/repositories/testuser/awesome-project');
    
    // Click on src directory
    await page.click('text=src');
    
    // Should navigate to src directory
    await expect(page).toHaveURL('/repositories/testuser/awesome-project/tree/main/src');
    
    // Should show src directory contents
    await expect(page.locator('text=index.ts')).toBeVisible();
  });

  test('should handle empty directories gracefully', async ({ page }) => {
    await page.route('**/api/v1/repositories/testuser/awesome-project/tree**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: { entries: [] }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project');
    
    // Should show empty state message
    await expect(page.locator('text=This directory is empty')).toBeVisible();
  });

  test('should handle file tree errors with retry functionality', async ({ page }) => {
    let attempts = 0;
    await page.route('**/api/v1/repositories/testuser/awesome-project/tree**', async route => {
      attempts++;
      if (attempts === 1) {
        // First attempt fails
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            message: 'Internal server error'
          })
        });
      } else {
        // Second attempt succeeds
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              entries: [
                { name: 'file.txt', path: 'file.txt', type: 'blob', size: 100 }
              ]
            }
          })
        });
      }
    });

    await page.goto('/repositories/testuser/awesome-project');
    
    // Should show error message
    await expect(page.locator('text=Internal server error')).toBeVisible();
    
    // Should show retry button
    const retryButton = page.locator('button:has-text("Try Again")');
    await expect(retryButton).toBeVisible();
    
    // Click retry should reload successfully
    await retryButton.click();
    await expect(page.locator('text=file.txt')).toBeVisible();
  });
});