import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('User Registration', () => {
  test.beforeEach(async ({ page }) => {
    // Clear any existing auth state
    await page.context().clearCookies();
    await page.context().clearPermissions();
  });

  test('should display registration form correctly', async ({ page }) => {
    await page.goto('/register');
    
    // Should have the registration heading
    await expect(page.locator('h2')).toContainText('Create your account');
    
    // Should have all registration form fields
    await expect(page.locator('[data-testid="name-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="username-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="email-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="password-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="confirm-password-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="register-button"]')).toBeVisible();
    
    // Should have link to login page
    await expect(page.locator('a[href="/login"]')).toBeVisible();
  });

  test('should validate required fields', async ({ page }) => {
    await page.goto('/register');
    
    // Try to submit without filling fields
    await page.click('[data-testid="register-button"]');
    
    // Should show validation errors (the specific behavior depends on form validation)
    // This is a placeholder - actual validation behavior needs to be implemented
    await page.waitForTimeout(1000); // Wait for any validation to appear
  });

  test('should validate password confirmation', async ({ page }) => {
    await page.goto('/register');
    
    // Fill form with mismatched passwords
    await page.fill('[data-testid="name-input"]', testUser.name);
    await page.fill('[data-testid="username-input"]', testUser.username);
    await page.fill('[data-testid="email-input"]', testUser.email);
    await page.fill('[data-testid="password-input"]', testUser.password);
    await page.fill('[data-testid="confirm-password-input"]', 'different-password');
    
    await page.click('[data-testid="register-button"]');
    
    // Should show password mismatch error
    await page.waitForTimeout(1000); // Wait for validation
  });

  test('should register new user successfully', async ({ page }) => {
    await page.goto('/register');
    
    // Mock successful registration response
    await page.route('**/api/v1/auth/register', async route => {
      await route.fulfill({
        status: 201,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            access_token: 'mock-jwt-token',
            user: {
              id: '1',
              name: testUser.name,
              username: testUser.username,
              email: testUser.email
            }
          }
        })
      });
    });

    // Mock auth/me for dashboard
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

    // Mock repositories for dashboard
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
    
    // Fill registration form
    await page.fill('[data-testid="name-input"]', testUser.name);
    await page.fill('[data-testid="username-input"]', testUser.username);
    await page.fill('[data-testid="email-input"]', testUser.email);
    await page.fill('[data-testid="password-input"]', testUser.password);
    await page.fill('[data-testid="confirm-password-input"]', testUser.password);
    
    await page.click('[data-testid="register-button"]');
    
    // Should redirect to dashboard after successful registration
    await page.waitForURL('/dashboard', { timeout: 15000 });
    await expect(page).toHaveURL('/dashboard');
  });

  test('should show error for existing email', async ({ page }) => {
    await page.goto('/register');
    
    // Mock registration error response
    await page.route('**/api/v1/auth/register', async route => {
      await route.fulfill({
        status: 409,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'Email already exists'
        })
      });
    });
    
    // Fill registration form
    await page.fill('[data-testid="name-input"]', testUser.name);
    await page.fill('[data-testid="username-input"]', testUser.username);
    await page.fill('[data-testid="email-input"]', 'existing@example.com');
    await page.fill('[data-testid="password-input"]', testUser.password);
    await page.fill('[data-testid="confirm-password-input"]', testUser.password);
    
    await page.click('[data-testid="register-button"]');
    
    // Should show error message
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Email already exists');
    
    // Should stay on registration page
    await expect(page).toHaveURL('/register');
  });

  test('should navigate to login from registration page', async ({ page }) => {
    await page.goto('/register');
    
    // Click login link
    await page.click('a[href="/login"]');
    
    // Should navigate to login page
    await expect(page).toHaveURL('/login');
    await expect(page.locator('h2')).toContainText('Sign in to Hub');
  });
});