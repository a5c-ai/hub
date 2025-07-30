import { test, expect } from '@playwright/test';

test.describe('PWA Support', () => {
  test('should serve manifest.json', async ({ request }) => {
    const response = await request.get('/manifest.json');
    expect(response.ok()).toBeTruthy();
    const manifest = await response.json();
    expect(manifest.name).toBe('Hub - Git Platform');
    expect(manifest.short_name).toBe('Hub');
  });

  test('should show offline fallback page when offline', async ({ page, context }) => {
    await context.setOffline(true);
    await page.goto('/offline');
    await expect(page.locator('h1')).toHaveText("You're offline");
  });

  test('should include manifest link tag in HTML', async ({ page }) => {
    await page.goto('/');
    const manifestLink = page.locator('link[rel="manifest"]');
    await expect(manifestLink).toHaveCount(1);
  });
});
