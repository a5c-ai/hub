import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('File Editor & Management Features', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication for all tests
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
            clone_url: 'https://hub.example.com/testuser/awesome-project.git',
            stargazers_count: 42,
            forks_count: 8,
            issues_count: 5,
            updated_at: '2024-07-20T10:00:00Z',
            owner: {
              username: 'testuser',
              avatar_url: 'https://avatars.githubusercontent.com/u/1?v=4'
            }
          }
        })
      });
    });
  });

  test('should load file content in editor correctly', async ({ page }) => {
    // Mock file content for editing
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/src/utils.js', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            name: 'utils.js',
            path: 'src/utils.js',
            size: 256,
            encoding: 'utf-8',
            content: `// Utility functions
export function formatDate(date) {
  return date.toISOString().split('T')[0];
}

export function capitalize(str) {
  return str.charAt(0).toUpperCase() + str.slice(1);
}`,
            sha: 'edit123'
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/edit/main/src/utils.js');
    
    // Should display edit page header
    await expect(page.locator('h1:has-text("Edit src/utils.js")')).toBeVisible();
    await expect(page.locator('text=Editing utils.js')).toBeVisible();
    await expect(page.locator('text=on branch main')).toBeVisible();
    
    // Should load file content in textarea
    const textarea = page.locator('textarea');
    await expect(textarea).toBeVisible();
    await expect(textarea).toHaveValue(/Utility functions/);
    await expect(textarea).toHaveValue(/formatDate/);
    await expect(textarea).toHaveValue(/capitalize/);
  });

  test('should allow editing file content', async ({ page }) => {
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/config.json', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            name: 'config.json',
            path: 'config.json',
            size: 128,
            encoding: 'utf-8',
            content: `{
  "name": "test-app",
  "version": "1.0.0"
}`,
            sha: 'config123'
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/edit/main/config.json');
    
    // Should load original content
    const textarea = page.locator('textarea');
    await expect(textarea).toHaveValue(/test-app/);
    
    // Should allow editing content
    await textarea.clear();
    await textarea.fill(`{
  "name": "updated-app",
  "version": "2.0.0",
  "description": "Updated configuration"
}`);
    
    // Should reflect changes
    await expect(textarea).toHaveValue(/updated-app/);
    await expect(textarea).toHaveValue(/2.0.0/);
    await expect(textarea).toHaveValue(/Updated configuration/);
  });

  test('should handle successful file commit', async ({ page }) => {
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/test.txt', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              name: 'test.txt',
              path: 'test.txt',
              size: 64,
              encoding: 'utf-8',
              content: 'Original content',
              sha: 'original123'
            }
          })
        });
      } else if (route.request().method() === 'PUT') {
        // Mock successful commit
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              commit: {
                sha: 'newcommit456',
                message: 'Update test.txt'
              }
            }
          })
        });
      }
    });

    await page.goto('/repositories/testuser/awesome-project/edit/main/test.txt');
    
    // Edit content
    const textarea = page.locator('textarea');
    await textarea.clear();
    await textarea.fill('Updated content with new changes');
    
    // Should have default commit message
    const commitMessageInput = page.locator('input[placeholder*="Commit message"]');
    await expect(commitMessageInput).toHaveValue('Update test.txt');
    
    // Update commit message
    await commitMessageInput.clear();
    await commitMessageInput.fill('Fix typos and update content');
    
    // Commit button should be enabled
    const commitButton = page.locator('button:has-text("Commit changes")');
    await expect(commitButton).toBeEnabled();
    
    // Click commit
    await commitButton.click();
    
    // Should redirect to blob view after successful commit
    await expect(page).toHaveURL('/repositories/testuser/awesome-project/blob/main/test.txt');
  });

  test('should validate commit message is required', async ({ page }) => {
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/validate.txt', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            name: 'validate.txt',
            path: 'validate.txt',
            size: 32,
            encoding: 'utf-8',
            content: 'Test content',
            sha: 'validate123'
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/edit/main/validate.txt');
    
    // Clear commit message
    const commitMessageInput = page.locator('input[placeholder*="Commit message"]');
    await commitMessageInput.clear();
    
    // Commit button should be disabled when message is empty
    const commitButton = page.locator('button:has-text("Commit changes")');
    await expect(commitButton).toBeDisabled();
    
    // Enter commit message
    await commitMessageInput.fill('Valid commit message');
    
    // Commit button should be enabled
    await expect(commitButton).toBeEnabled();
  });

  test('should provide cancel functionality', async ({ page }) => {
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/cancel-test.txt', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            name: 'cancel-test.txt',
            path: 'cancel-test.txt',
            size: 32,
            encoding: 'utf-8',
            content: 'Original content',
            sha: 'cancel123'
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/edit/main/cancel-test.txt');
    
    // Should show cancel buttons
    const headerCancelButton = page.locator('button:has-text("Cancel")').first();
    await expect(headerCancelButton).toBeVisible();
    
    // Click cancel should navigate back to blob view
    await headerCancelButton.click();
    await expect(page).toHaveURL('/repositories/testuser/awesome-project/blob/main/cancel-test.txt');
  });

  test('should handle file not found for editing', async ({ page }) => {
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/missing-edit-file.txt', async route => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          message: 'File not found'
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/edit/main/missing-edit-file.txt');
    
    // Should show error message
    await expect(page.locator('text=Error: File not found')).toBeVisible();
    
    // Should show try again button
    const retryButton = page.locator('button:has-text("Try Again")');
    await expect(retryButton).toBeVisible();
    await expect(retryButton).toBeEnabled();
  });
});