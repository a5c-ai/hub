import { Page, expect } from '@playwright/test';

/**
 * Test utilities for common operations
 */

/**
 * Mock user credentials for testing
 */
export const testUser = {
  username: 'testuser',
  email: 'test@example.com',
  password: 'TestPassword123!',
  name: 'Test User'
};

/**
 * Setup authentication for tests - mocks all necessary auth endpoints and sets localStorage
 * @param page - Playwright page object
 * @param userData - Optional user data to use for authentication
 */
export async function setupAuthentication(page: Page, userData = testUser) {
  // Mock authentication endpoints
  await page.route('**/api/v1/profile', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        data: {
          id: '1',
          name: userData.name,
          username: userData.username,
          email: userData.email
        }
      })
    });
  });

  // Mock repositories endpoint
  await page.route('**/api/v1/repositories**', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        data: {
          repositories: [
            {
              id: '1',
              name: 'awesome-project',
              full_name: `${userData.username}/awesome-project`,
              description: 'An awesome test project',
              private: false,
              owner: {
                id: '1',
                username: userData.username,
                avatar_url: 'https://example.com/avatar.jpg'
              },
              default_branch: 'main',
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
              language: 'TypeScript',
              stars_count: 42,
              forks_count: 5,
              watchers_count: 10
            }
          ],
          pagination: {
            page: 1,
            per_page: 30,
            total: 1,
            total_pages: 1
          }
        }
      })
    });
  });

  // Mock activity endpoint
  await page.route('**/api/v1/activity**', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        success: true,
        data: {
          activities: [
            {
              id: '1',
              type: 'push',
              actor: {
                id: '1',
                username: userData.username,
                avatar_url: 'https://example.com/avatar.jpg'
              },
              repository: {
                id: '1',
                name: 'awesome-project',
                full_name: `${userData.username}/awesome-project`
              },
              payload: {
                ref: 'refs/heads/main',
                commits: [
                  {
                    sha: 'abc123',
                    message: 'Update README',
                    author: {
                      name: userData.name,
                      email: userData.email
                    }
                  }
                ]
              },
              public: true,
              created_at: new Date().toISOString()
            }
          ],
          pagination: {
            page: 1,
            per_page: 30,
            total: 1,
            total_pages: 1
          }
        }
      })
    });
  });

  // Set authentication state in localStorage
  await page.addInitScript((userData) => {
    // Set authentication token
    window.localStorage.setItem('auth_token', 'mock-jwt-token');
    
    // Set auth store state for zustand persist
    window.localStorage.setItem('auth-storage', JSON.stringify({
      state: {
        user: {
          id: '1',
          name: userData.name,
          username: userData.username,
          email: userData.email
        },
        token: 'mock-jwt-token',
        isAuthenticated: true
      },
      version: 0
    }));
  }, userData);
}

/**
 * Login helper - logs in a user with credentials
 * @param page - Playwright page object
 * @param email - User email
 * @param password - User password
 */
export async function loginUser(page: Page, email: string = testUser.email, password: string = testUser.password) {
  await page.goto('/login');
  await page.fill('[data-testid="email-input"]', email);
  await page.fill('[data-testid="password-input"]', password);
  await page.click('[data-testid="login-button"]');
  
  // Wait for navigation to dashboard
  await page.waitForURL('/dashboard');
}

/**
 * Register helper - registers a new user
 * @param page - Playwright page object  
 * @param userData - User registration data
 */
export async function registerUser(page: Page, userData = testUser) {
  await page.goto('/register');
  await page.fill('[data-testid="name-input"]', userData.name);
  await page.fill('[data-testid="username-input"]', userData.username);
  await page.fill('[data-testid="email-input"]', userData.email);
  await page.fill('[data-testid="password-input"]', userData.password);
  await page.fill('[data-testid="confirm-password-input"]', userData.password);
  await page.click('[data-testid="register-button"]');
}

/**
 * Wait for loading to complete
 * @param page - Playwright page object
 */
export async function waitForLoadingToComplete(page: Page) {
  // Wait for any loading spinners to disappear
  await page.waitForFunction(() => {
    const spinners = document.querySelectorAll('.animate-spin');
    return spinners.length === 0;
  }, { timeout: 10000 });
}

/**
 * Check if user is on the login page
 * @param page - Playwright page object
 */
export async function expectLoginPage(page: Page) {
  await expect(page).toHaveURL('/login');
  await expect(page.locator('h2')).toContainText('Sign in to Hub');
}

/**
 * Check if user is on the dashboard page
 * @param page - Playwright page object
 */
export async function expectDashboardPage(page: Page) {
  await expect(page).toHaveURL('/dashboard');
  await expect(page.locator('h1')).toContainText('Welcome back,');
}

/**
 * Take a screenshot with a descriptive name
 * @param page - Playwright page object
 * @param name - Screenshot name
 */
export async function takeScreenshot(page: Page, name: string) {
  await page.screenshot({ 
    path: `e2e/screenshots/${name}.png`, 
    fullPage: true 
  });
}
