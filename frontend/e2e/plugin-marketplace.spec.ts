import { test, expect } from '@playwright/test';
import { setupAuthentication, testUser } from './helpers/test-utils';

test.describe('Plugin Marketplace UI', () => {
  test.beforeEach(async ({ page }) => {
    await setupAuthentication(page);
    // Mock plugin marketplace endpoint
    await page.route('**/api/v1/plugins', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              apiVersion: 'v1',
              kind: 'Plugin',
              metadata: {
                name: 'sample-plugin',
                version: '1.0.0',
                description: 'A sample plugin for testing',
                author: 'Test Author',
                website: 'https://example.com',
                license: 'MIT',
              },
              spec: {
                runtime: 'go',
                entry: '',
                permissions: [],
                hooks: [],
                webhooks: [],
                settings: [],
                dependencies: [],
              },
            },
          ],
        }),
      });
    });
  });

  test('should list plugins and navigate to detail page', async ({ page }) => {
    await page.goto('/plugins');
    await expect(page.locator('h1', { hasText: 'Plugin Marketplace' })).toBeVisible();
    await expect(page.locator('text=sample-plugin')).toBeVisible();
    await page.click('text=sample-plugin');
    await page.waitForURL('/plugins/sample-plugin');
    await expect(page.locator('h1', { hasText: 'sample-plugin' })).toBeVisible();
    await expect(page.locator('text=A sample plugin for testing')).toBeVisible();
  });
});
