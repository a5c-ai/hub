import { test, expect } from '@playwright/test';

test.describe('Basic Application Tests', () => {
  test('should load the application', async ({ page }) => {
    await page.goto('/');
    
    // Should have a title containing Hub
    await expect(page).toHaveTitle(/Hub/);
    
    // Should redirect to login for unauthenticated users (with longer timeout)
    await expect(page).toHaveURL(/\/login/, { timeout: 15000 });
  });

  test('should display login page correctly', async ({ page }) => {
    await page.goto('/login');
    
    // Should have the sign in heading (using the actual h2 text from LoginForm)
    await expect(page.locator('h2')).toContainText('Sign in to Hub');
    
    // Should have the login form elements visible using test-ids
    await expect(page.locator('[data-testid="email-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="password-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="login-button"]')).toBeVisible();
  });

  test('should navigate to register page', async ({ page }) => {
    await page.goto('/login');
    
    // Click the register link
    await page.click('a[href="/register"]');
    
    // Should be on register page
    await expect(page).toHaveURL(/\/register/);
    // Check for the actual h2 heading on the register page
    await expect(page.locator('h2')).toContainText('Create your account');
  });
});