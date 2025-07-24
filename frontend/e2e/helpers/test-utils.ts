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
  await expect(page.locator('h1')).toContainText('Sign in');
}

/**
 * Check if user is on the dashboard page
 * @param page - Playwright page object
 */
export async function expectDashboardPage(page: Page) {
  await expect(page).toHaveURL('/dashboard');
  await expect(page.locator('h1')).toContainText('Welcome back');
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