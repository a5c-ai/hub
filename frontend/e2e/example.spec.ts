import { test, expect } from '@playwright/test';

test.describe('Basic Application Tests', () => {
  test('should load the application', async ({ page }) => {
    await page.goto('/');
    
    // Should have a title
    await expect(page).toHaveTitle(/Hub/);
    
    // Should redirect to login for unauthenticated users
    await expect(page).toHaveURL(/\/login/);
  });

  test('should display login page correctly', async ({ page }) => {
    await page.goto('/login');
    
    // Should have the sign in heading
    await expect(page.locator('h2')).toContainText('Sign in to Hub');
    
    // Should have the login form elements visible
    await expect(page.locator('input[type="email"]')).toBeVisible();
    await expect(page.locator('input[type="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('should navigate to register page', async ({ page }) => {
    await page.goto('/login');
    
    // Click the register link
    await page.click('a[href="/register"]');
    
    // Should be on register page
    await expect(page).toHaveURL(/\/register/);
    await expect(page.locator('h2')).toContainText('Create your account');
  });
});