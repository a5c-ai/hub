import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('File Viewer & Display Features', () => {
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

  test('should display text file with syntax highlighting', async ({ page }) => {
    // Mock file content response
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/src/index.ts', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            name: 'index.ts',
            path: 'src/index.ts',
            size: 512,
            encoding: 'utf-8',
            content: `// TypeScript example file
import { App } from './App';

const app = new App();

function main() {
  console.log('Hello, World!');
  app.start();
}

export { main };`,
            sha: 'abc123'
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/blob/main/src/index.ts');
    
    // Should display file metadata
    await expect(page.locator('h2:has-text("index.ts")')).toBeVisible();
    await expect(page.locator('text=512 Bytes')).toBeVisible();
    
    // Should display file content with syntax highlighting
    await expect(page.locator('text=// TypeScript example file')).toBeVisible();
    await expect(page.locator('text=import { App } from')).toBeVisible();
    await expect(page.locator('text=console.log')).toBeVisible();
  });

  test('should display markdown file with rendered preview', async ({ page }) => {
    // Mock markdown file content
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/README.md', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            name: 'README.md',
            path: 'README.md',
            size: 2048,
            encoding: 'utf-8',
            content: `# Awesome Project

This is an **awesome** project built with modern technologies.

## Features

- Feature 1
- Feature 2
- Feature 3`,
            sha: 'def456'
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/blob/main/README.md');
    
    // Should display file metadata
    await expect(page.locator('h2:has-text("README.md")')).toBeVisible();
    await expect(page.locator('text=2.00 KB')).toBeVisible();
    
    // Should render markdown content
    await expect(page.locator('h1:has-text("Awesome Project")')).toBeVisible();
    await expect(page.locator('strong:has-text("awesome")')).toBeVisible();
    await expect(page.locator('h2:has-text("Features")')).toBeVisible();
  });

  test('should handle binary files with appropriate message', async ({ page }) => {
    // Mock binary file content
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/assets/logo.png', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            name: 'logo.png',
            path: 'assets/logo.png',
            size: 8192,
            encoding: 'base64',
            content: 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
            sha: 'ghi789'
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/blob/main/assets/logo.png');
    
    // Should display file metadata
    await expect(page.locator('h2:has-text("logo.png")')).toBeVisible();
    await expect(page.locator('text=8.00 KB')).toBeVisible();
    
    // Should show binary file message
    await expect(page.locator('text=Binary file cannot be displayed')).toBeVisible();
    await expect(page.locator('text=Use the download button to save the file')).toBeVisible();
  });

  test('should provide edit file functionality', async ({ page }) => {
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/src/config.js', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            name: 'config.js',
            path: 'src/config.js',
            size: 256,
            encoding: 'utf-8',
            content: `const config = {
  apiUrl: 'https://api.example.com',
  timeout: 5000
};

module.exports = config;`,
            sha: 'mno345'
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/blob/main/src/config.js');
    
    // Should show edit button
    const editButton = page.locator('button:has-text("Edit")');
    await expect(editButton).toBeVisible();
    await expect(editButton).toBeEnabled();
    
    // Click should navigate to edit page
    await editButton.click();
    await expect(page).toHaveURL('/repositories/testuser/awesome-project/edit/main/src/config.js');
  });

  test('should provide file download functionality', async ({ page }) => {
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/download-test.txt', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            name: 'download-test.txt',
            path: 'download-test.txt',
            size: 64,
            encoding: 'utf-8',
            content: 'This is test content for download',
            sha: 'stu901'
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/blob/main/download-test.txt');
    
    // Should show download button
    const downloadButton = page.locator('button:has-text("Download")');
    await expect(downloadButton).toBeVisible();
    await expect(downloadButton).toBeEnabled();
  });

  test('should handle file not found errors', async ({ page }) => {
    await page.route('**/api/v1/repositories/testuser/awesome-project/contents/missing-file.txt', async route => {
      await route.fulfill({
        status: 404,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          message: 'File not found'
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/blob/main/missing-file.txt');
    
    // Should show error message
    await expect(page.locator('text=Error: File not found')).toBeVisible();
    
    // Should show try again button
    const retryButton = page.locator('button:has-text("Try Again")');
    await expect(retryButton).toBeVisible();
    await expect(retryButton).toBeEnabled();
  });
});