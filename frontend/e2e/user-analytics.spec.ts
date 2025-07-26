import { test, expect } from '@playwright/test';
import { loginUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('User Analytics Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await loginUser(page);
    
    // Mock analytics API endpoints
    await page.route('**/api/v1/analytics/user/**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            repositories: { total: 12, private: 8, public: 4 },
            contributions: { total: 156, thisMonth: 23, thisWeek: 5 },
            pullRequests: { total: 45, merged: 38, open: 4, closed: 3 },
            issues: { total: 32, closed: 28, open: 4 },
            codeActivity: [
              { date: '2024-01-01', commits: 5, linesAdded: 245, linesDeleted: 12 },
              { date: '2024-01-02', commits: 3, linesAdded: 156, linesDeleted: 8 }
            ]
          }
        })
      });
    });
  });

  test('should display personal analytics dashboard', async ({ page }) => {
    await page.goto('/analytics');
    await waitForLoadingToComplete(page);

    // Verify main analytics sections
    await expect(page.locator('h1')).toContainText('Your Analytics');
    await expect(page.locator('text=Repository Statistics')).toBeVisible();
    await expect(page.locator('text=Contribution Activity')).toBeVisible();
    await expect(page.locator('text=Code Activity')).toBeVisible();

    // Check repository stats
    await expect(page.locator('text=12 Total Repositories')).toBeVisible();
    await expect(page.locator('text=8 Private')).toBeVisible();
    await expect(page.locator('text=4 Public')).toBeVisible();

    // Check contribution stats
    await expect(page.locator('text=156 Total Contributions')).toBeVisible();
    await expect(page.locator('text=23 This Month')).toBeVisible();
    await expect(page.locator('text=5 This Week')).toBeVisible();
  });

  test('should display contribution activity heatmap', async ({ page }) => {
    await page.goto('/analytics');
    await waitForLoadingToComplete(page);

    // Check for contribution heatmap
    await expect(page.locator('[data-testid="contribution-heatmap"]')).toBeVisible();
    await expect(page.locator('text=Contribution Activity')).toBeVisible();
    
    // Verify heatmap has days of the week
    await expect(page.locator('text=Mon')).toBeVisible();
    await expect(page.locator('text=Wed')).toBeVisible();
    await expect(page.locator('text=Fri')).toBeVisible();
    
    // Check for month labels
    await expect(page.locator('text=Jan')).toBeVisible();
  });

  test('should show pull request and issue metrics', async ({ page }) => {
    await page.goto('/analytics');
    await waitForLoadingToComplete(page);

    // Navigate to detailed view
    await page.click('[data-testid="pr-issues-tab"]');
    await waitForLoadingToComplete(page);

    // Verify PR metrics
    await expect(page.locator('text=Pull Request Statistics')).toBeVisible();
    await expect(page.locator('text=45 Total PRs')).toBeVisible();
    await expect(page.locator('text=38 Merged')).toBeVisible();
    await expect(page.locator('text=4 Open')).toBeVisible();
    await expect(page.locator('text=3 Closed')).toBeVisible();

    // Verify issue metrics
    await expect(page.locator('text=Issue Statistics')).toBeVisible();
    await expect(page.locator('text=32 Total Issues')).toBeVisible();
    await expect(page.locator('text=28 Closed')).toBeVisible();
    await expect(page.locator('text=4 Open')).toBeVisible();
  });

  test('should display code activity trends', async ({ page }) => {
    await page.goto('/analytics');
    await waitForLoadingToComplete(page);

    // Navigate to code activity tab
    await page.click('[data-testid="code-activity-tab"]');
    await waitForLoadingToComplete(page);

    // Verify code activity chart
    await expect(page.locator('text=Code Activity Over Time')).toBeVisible();
    await expect(page.locator('[data-testid="code-activity-chart"]')).toBeVisible();
    
    // Check for activity metrics
    await expect(page.locator('text=Lines Added')).toBeVisible();
    await expect(page.locator('text=Lines Deleted')).toBeVisible();
    await expect(page.locator('text=Commits')).toBeVisible();
  });

  test('should allow filtering analytics by date range', async ({ page }) => {
    await page.goto('/analytics');
    await waitForLoadingToComplete(page);

    // Open date range picker
    await page.click('[data-testid="date-range-picker"]');
    
    // Select last 30 days
    await page.click('[data-testid="last-30-days"]');
    await waitForLoadingToComplete(page);
    
    // Verify date range is applied
    await expect(page.locator('text=Last 30 Days')).toBeVisible();
    
    // Test custom date range
    await page.click('[data-testid="date-range-picker"]');
    await page.click('[data-testid="custom-range"]');
    
    await page.fill('[data-testid="start-date"]', '2024-01-01');
    await page.fill('[data-testid="end-date"]', '2024-01-31');
    await page.click('[data-testid="apply-range"]');
    
    await waitForLoadingToComplete(page);
    await expect(page.locator('text=Jan 1 - Jan 31, 2024')).toBeVisible();
  });

  test('should export analytics data', async ({ page }) => {
    await page.goto('/analytics');
    await waitForLoadingToComplete(page);

    // Mock download functionality
    const downloadPromise = page.waitForEvent('download');

    // Click export button
    await page.click('[data-testid="export-analytics"]');
    
    // Select export format
    await page.click('[data-testid="export-csv"]');
    
    // Wait for download to start
    const download = await downloadPromise;
    expect(download.suggestedFilename()).toContain('analytics');
    expect(download.suggestedFilename()).toContain('.csv');
  });

  test('should display repository insights', async ({ page }) => {
    await page.goto('/analytics/repositories');
    await waitForLoadingToComplete(page);

    // Verify repository analytics
    await expect(page.locator('h1')).toContainText('Repository Insights');
    await expect(page.locator('text=Most Active Repositories')).toBeVisible();
    await expect(page.locator('text=Language Distribution')).toBeVisible();
    await expect(page.locator('text=Repository Growth')).toBeVisible();

    // Check for repository cards
    await expect(page.locator('[data-testid="repo-card"]').first()).toBeVisible();
    
    // Verify language chart
    await expect(page.locator('[data-testid="language-chart"]')).toBeVisible();
  });

  test('should handle mobile analytics view', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 });
    
    await page.goto('/analytics');
    await waitForLoadingToComplete(page);

    // Verify mobile-friendly layout
    await expect(page.locator('h1')).toContainText('Your Analytics');
    
    // Check that stats are stacked vertically
    const statsContainer = page.locator('[data-testid="stats-container"]');
    await expect(statsContainer).toBeVisible();
    
    // Verify charts are responsive
    await expect(page.locator('[data-testid="contribution-heatmap"]')).toBeVisible();
    
    // Test mobile navigation
    await page.click('[data-testid="mobile-menu-toggle"]');
    await expect(page.locator('[data-testid="mobile-menu"]')).toBeVisible();
  });

  test('should display analytics performance metrics', async ({ page }) => {
    const startTime = Date.now();
    
    await page.goto('/analytics');
    await waitForLoadingToComplete(page);
    
    const loadTime = Date.now() - startTime;
    
    // Analytics page should load within 5 seconds
    expect(loadTime).toBeLessThan(5000);
    
    // Check for performance optimizations
    const images = page.locator('img');
    const imageCount = await images.count();
    
    // Verify images have loading attributes
    for (let i = 0; i < imageCount; i++) {
      const img = images.nth(i);
      const loading = await img.getAttribute('loading');
      expect(loading).toBeTruthy();
    }
  });

  test('should handle analytics data loading states', async ({ page }) => {
    // Mock slow API response
    await page.route('**/api/v1/analytics/user/**', async route => {
      await new Promise(resolve => setTimeout(resolve, 2000));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ success: true, data: {} })
      });
    });

    await page.goto('/analytics');
    
    // Check for loading states
    await expect(page.locator('[data-testid="analytics-loading"]')).toBeVisible();
    await expect(page.locator('text=Loading your analytics...')).toBeVisible();
    
    // Wait for data to load
    await waitForLoadingToComplete(page);
    await expect(page.locator('[data-testid="analytics-loading"]')).not.toBeVisible();
  });

  test('should handle analytics API errors gracefully', async ({ page }) => {
    // Mock API error
    await page.route('**/api/v1/analytics/user/**', async route => {
      await route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({ success: false, error: 'Analytics service unavailable' })
      });
    });

    await page.goto('/analytics');
    await waitForLoadingToComplete(page);

    // Check error handling
    await expect(page.locator('[data-testid="analytics-error"]')).toBeVisible();
    await expect(page.locator('text=Unable to load analytics')).toBeVisible();
    
    // Check retry functionality
    await page.click('[data-testid="retry-analytics"]');
    await expect(page.locator('[data-testid="analytics-loading"]')).toBeVisible();
  });
});