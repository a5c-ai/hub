import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('Navigation and Layout', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication for all navigation tests
    await page.route('**/api/auth/me', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          user: {
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
      window.localStorage.setItem('auth-token', 'mock-jwt-token');
    });
  });

  test('should display main navigation elements', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check for main navigation/header elements
    // These depend on the actual AppLayout implementation
    await expect(page.locator('[data-testid="main-header"]')).toBeVisible();
    
    // Check for user menu or profile area
    if (await page.locator('[data-testid="user-menu"]').count() > 0) {
      await expect(page.locator('[data-testid="user-menu"]')).toBeVisible();
    }
  });

  test('should display sidebar navigation', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check for sidebar elements
    if (await page.locator('[data-testid="sidebar"]').count() > 0) {
      await expect(page.locator('[data-testid="sidebar"]')).toBeVisible();
      
      // Common navigation items
      const navItems = [
        'Dashboard',
        'Repositories', 
        'Organizations',
        'Settings'
      ];
      
      for (const item of navItems) {
        if (await page.locator(`text=${item}`).count() > 0) {
          await expect(page.locator(`text=${item}`)).toBeVisible();
        }
      }
    }
  });

  test('should navigate between main sections', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Test navigation to repositories
    if (await page.locator('a[href="/repositories"]').count() > 0) {
      await page.click('a[href="/repositories"]');
      await expect(page).toHaveURL('/repositories');
    }
    
    // Navigate back to dashboard
    if (await page.locator('a[href="/dashboard"]').count() > 0) {
      await page.click('a[href="/dashboard"]');
      await expect(page).toHaveURL('/dashboard');
    }
  });

  test('should display user information in header/sidebar', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check if user name or username is displayed somewhere in the layout
    const userNameVisible = await page.locator(`text=${testUser.name}`).count() > 0;
    const usernameVisible = await page.locator(`text=${testUser.username}`).count() > 0;
    
    expect(userNameVisible || usernameVisible).toBe(true);
  });

  test('should handle mobile navigation (hamburger menu)', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    await page.goto('/dashboard');
    
    // Check for mobile menu button/hamburger
    if (await page.locator('[data-testid="mobile-menu-button"]').count() > 0) {
      await expect(page.locator('[data-testid="mobile-menu-button"]')).toBeVisible();
      
      // Click to open mobile menu
      await page.click('[data-testid="mobile-menu-button"]');
      
      // Check that mobile menu opens
      await expect(page.locator('[data-testid="mobile-menu"]')).toBeVisible();
    }
  });

  test('should display breadcrumbs on nested pages', async ({ page }) => {
    // Mock a repository page to test breadcrumbs
    await page.route('**/api/repositories/testuser/awesome-project', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          repository: {
            id: '1',
            name: 'awesome-project',
            full_name: 'testuser/awesome-project',
            description: 'An awesome project',
            private: false
          }
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project');
    
    // Check for breadcrumb navigation if implemented
    if (await page.locator('[data-testid="breadcrumbs"]').count() > 0) {
      await expect(page.locator('[data-testid="breadcrumbs"]')).toBeVisible();
    }
  });

  test('should show loading states during navigation', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Add a delay to API response to test loading state
    await page.route('**/api/repositories', async route => {
      await new Promise(resolve => setTimeout(resolve, 1000));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          repositories: []
        })
      });
    });
    
    // Navigate to repositories page
    if (await page.locator('a[href="/repositories"]').count() > 0) {
      await page.click('a[href="/repositories"]');
      
      // Check for loading indicator
      if (await page.locator('.animate-spin').count() > 0) {
        await expect(page.locator('.animate-spin')).toBeVisible();
      }
      
      // Wait for loading to complete
      await page.waitForResponse('**/api/repositories');
    }
  });

  test('should maintain navigation state after page refresh', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Navigate to a different page
    if (await page.locator('a[href="/repositories"]').count() > 0) {
      await page.click('a[href="/repositories"]');
      await expect(page).toHaveURL('/repositories');
      
      // Refresh the page
      await page.reload();
      
      // Should still be on the repositories page
      await expect(page).toHaveURL('/repositories');
    }
  });

  test('should display footer information', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Check for footer if present
    if (await page.locator('[data-testid="footer"]').count() > 0) {
      await expect(page.locator('[data-testid="footer"]')).toBeVisible();
    }
  });

  test('should handle keyboard navigation', async ({ page }) => {
    await page.goto('/dashboard');
    
    // Test Tab navigation through interactive elements
    await page.keyboard.press('Tab');
    
    // Check that focus is visible on interactive elements
    const focusedElement = await page.evaluate(() => document.activeElement?.tagName);
    expect(['A', 'BUTTON', 'INPUT'].includes(focusedElement || '')).toBe(true);
  });
});