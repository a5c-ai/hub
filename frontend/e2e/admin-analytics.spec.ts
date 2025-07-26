import { test, expect } from '@playwright/test';
import { loginUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('Admin System Analytics', () => {
  // Mock admin credentials
  const adminUser = {
    email: 'admin@example.com',
    password: 'AdminPassword123!'
  };

  test.beforeEach(async ({ page }) => {
    // Login as admin user
    await loginUser(page, adminUser.email, adminUser.password);
    
    // Mock API responses for analytics endpoints
    await page.route('**/api/v1/admin/analytics/platform*', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          total_views: 1000,
          total_users: 200,
          avg_response_time: 123,
          total_repositories: 50,
          trends: {
            views: [{ date: '2025-01-01', value: 100 }],
            users: [{ date: '2025-01-01', value: 20 }],
            response_time: [{ date: '2025-01-01', value: 120 }],
            repositories: [{ date: '2025-01-01', value: 5 }]
          }
        })
      });
    });
    await page.route('**/api/v1/admin/analytics/performance', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          cpu_usage_percent: 75,
          memory_usage_percent: 65,
          disk_usage_percent: 80,
          uptime_percent: 99,
          error_rate_percent: 2,
          active_connections: 150
        })
      });
    });
  });

  test.describe('Platform Usage Trends and Growth Metrics', () => {
    test('should display comprehensive platform usage analytics', async ({ page }) => {
      await page.goto('/admin/analytics');
      await waitForLoadingToComplete(page);

      // Verify main analytics dashboard is loaded
      await expect(page.locator('text=Platform Analytics')).toBeVisible();

      // Check for key growth metrics
      await expect(page.locator('text=Total Views')).toBeVisible();
      await expect(page.locator('text=Active Users')).toBeVisible();
      await expect(page.locator('text=Repositories')).toBeVisible();

      // Verify trend charts are present
      await expect(page.locator('svg')).toHaveCount({ min: 3 }); // Should have multiple charts
    });

    test('should allow time range selection for analytics', async ({ page }) => {
      await page.goto('/admin/analytics');
      await waitForLoadingToComplete(page);

      // Test different time ranges
      const timeRanges = ['Daily', 'Weekly', 'Monthly', 'Yearly'];
      
      for (const range of timeRanges) {
        await page.click(`button:has-text("${range}")`);
        
        // Verify the button becomes active
        await expect(page.locator(`button:has-text("${range}")`)).toHaveClass(/default/);
        
        // In a real implementation, you would verify that the data updates
        await waitForLoadingToComplete(page);
      }
    });

    test('should display user growth and activity trends', async ({ page }) => {
      await page.goto('/admin/analytics');
      await waitForLoadingToComplete(page);

      // Navigate to detailed analytics view
      await page.goto('/admin/analytics/users');
      await waitForLoadingToComplete(page);

      // Verify user analytics sections
      await expect(page.locator('text=User Growth')).toBeVisible();
      await expect(page.locator('text=Active Users')).toBeVisible();
      await expect(page.locator('text=User Retention')).toBeVisible();
      await expect(page.locator('text=Engagement Metrics')).toBeVisible();

      // Check for user activity breakdown
      await expect(page.locator('text=Daily Active Users')).toBeVisible();
      await expect(page.locator('text=Weekly Active Users')).toBeVisible();
      await expect(page.locator('text=Monthly Active Users')).toBeVisible();
    });

    test('should show platform feature adoption metrics', async ({ page }) => {
      await page.goto('/admin/analytics/features');
      await waitForLoadingToComplete(page);

      // Verify feature usage analytics
      await expect(page.locator('text=Feature Adoption')).toBeVisible();
      await expect(page.locator('text=Repository Creation')).toBeVisible();
      await expect(page.locator('text=Issue Tracking Usage')).toBeVisible();
      await expect(page.locator('text=Pull Request Activity')).toBeVisible();
      await expect(page.locator('text=Actions Usage')).toBeVisible();

      // Check adoption rates
      await expect(page.locator('text=Adoption Rate')).toBeVisible();
      await expect(page.locator('text=%').first()).toBeVisible();
    });
  });

  test.describe('Repository and Storage Analytics', () => {
    test('should display repository statistics and trends', async ({ page }) => {
      await page.goto('/admin/analytics/repositories');
      await waitForLoadingToComplete(page);

      // Verify repository analytics sections
      await expect(page.locator('text=Repository Overview')).toBeVisible();
      await expect(page.locator('text=Total Repositories')).toBeVisible();
      await expect(page.locator('text=Private Repositories')).toBeVisible();
      await expect(page.locator('text=Public Repositories')).toBeVisible();

      // Check repository activity metrics
      await expect(page.locator('text=Repository Activity')).toBeVisible();
      await expect(page.locator('text=New Repositories')).toBeVisible();
      await expect(page.locator('text=Repository Updates')).toBeVisible();
      await expect(page.locator('text=Repository Deletions')).toBeVisible();
    });

    test('should show storage usage and capacity planning', async ({ page }) => {
      await page.goto('/admin/analytics/storage');
      await waitForLoadingToComplete(page);

      // Verify storage analytics
      await expect(page.locator('text=Storage Overview')).toBeVisible();
      await expect(page.locator('text=Total Storage Used')).toBeVisible();
      await expect(page.locator('text=Available Storage')).toBeVisible();
      await expect(page.locator('text=Storage Growth Rate')).toBeVisible();

      // Check storage breakdown
      await expect(page.locator('text=Repository Data')).toBeVisible();
      await expect(page.locator('text=Build Artifacts')).toBeVisible();
      await expect(page.locator('text=Docker Images')).toBeVisible();
      await expect(page.locator('text=Log Files')).toBeVisible();

      // Verify capacity planning metrics
      await expect(page.locator('text=Projected Usage')).toBeVisible();
      await expect(page.locator('text=Days Until Full')).toBeVisible();
    });

    test('should display repository size distribution', async ({ page }) => {
      await page.goto('/admin/analytics/repositories');
      await waitForLoadingToComplete(page);

      // Scroll to repository size section
      await page.locator('text=Repository Size Distribution').scrollIntoViewIfNeeded();

      // Verify size distribution chart
      await expect(page.locator('text=Repository Size Distribution')).toBeVisible();
      await expect(page.locator('text=< 1MB')).toBeVisible();
      await expect(page.locator('text=1-10MB')).toBeVisible();
      await expect(page.locator('text=10-100MB')).toBeVisible();
      await expect(page.locator('text=> 100MB')).toBeVisible();
    });
  });

  test.describe('API Usage and Rate Limiting Statistics', () => {
    test('should display API usage metrics', async ({ page }) => {
      await page.goto('/admin/analytics/api');
      await waitForLoadingToComplete(page);

      // Verify API analytics sections
      await expect(page.locator('text=API Usage Overview')).toBeVisible();
      await expect(page.locator('text=Total API Requests')).toBeVisible();
      await expect(page.locator('text=Requests per Minute')).toBeVisible();
      await expect(page.locator('text=Response Time')).toBeVisible();

      // Check endpoint usage breakdown
      await expect(page.locator('text=Most Used Endpoints')).toBeVisible();
      await expect(page.locator('text=/api/v1/repositories')).toBeVisible();
      await expect(page.locator('text=/api/v1/users')).toBeVisible();
      await expect(page.locator('text=/api/v1/issues')).toBeVisible();
    });

    test('should show rate limiting statistics', async ({ page }) => {
      await page.goto('/admin/analytics/api');
      await waitForLoadingToComplete(page);

      // Scroll to rate limiting section
      await page.locator('text=Rate Limiting').scrollIntoViewIfNeeded();

      // Verify rate limiting metrics
      await expect(page.locator('text=Rate Limiting')).toBeVisible();
      await expect(page.locator('text=Rate Limited Requests')).toBeVisible();
      await expect(page.locator('text=Top Rate Limited Users')).toBeVisible();
      await expect(page.locator('text=Rate Limit Violations')).toBeVisible();

      // Check rate limit policies
      await expect(page.locator('text=Current Rate Limits')).toBeVisible();
      await expect(page.locator('text=requests/hour')).toBeVisible();
    });

    test('should display API error rates and status codes', async ({ page }) => {
      await page.goto('/admin/analytics/api');
      await waitForLoadingToComplete(page);

      // Verify error rate section
      await expect(page.locator('text=Error Rates')).toBeVisible();
      await expect(page.locator('text=4xx Errors')).toBeVisible();
      await expect(page.locator('text=5xx Errors')).toBeVisible();
      await expect(page.locator('text=Success Rate')).toBeVisible();

      // Check status code breakdown
      await expect(page.locator('text=Status Code Distribution')).toBeVisible();
      await expect(page.locator('text=200 OK')).toBeVisible();
      await expect(page.locator('text=404 Not Found')).toBeVisible();
      await expect(page.locator('text=500 Internal Server Error')).toBeVisible();
    });
  });

  test.describe('Performance Metrics and Bottlenecks', () => {
    test('should display system performance overview', async ({ page }) => {
      await page.goto('/admin/analytics/performance');
      await waitForLoadingToComplete(page);

      // Verify performance metrics
      await expect(page.locator('text=Performance Overview')).toBeVisible();
      await expect(page.locator('text=Average Response Time')).toBeVisible();
      await expect(page.locator('text=99th Percentile')).toBeVisible();
      await expect(page.locator('text=95th Percentile')).toBeVisible();
      await expect(page.locator('text=Throughput')).toBeVisible();

      // Check system resource usage
      await expect(page.locator('text=CPU Utilization')).toBeVisible();
      await expect(page.locator('text=Memory Usage')).toBeVisible();
      await expect(page.locator('text=Disk I/O')).toBeVisible();
      await expect(page.locator('text=Network I/O')).toBeVisible();
    });

    test('should identify performance bottlenecks', async ({ page }) => {
      await page.goto('/admin/analytics/performance');
      await waitForLoadingToComplete(page);

      // Scroll to bottlenecks section
      await page.locator('text=Performance Bottlenecks').scrollIntoViewIfNeeded();

      // Verify bottleneck identification
      await expect(page.locator('text=Performance Bottlenecks')).toBeVisible();
      await expect(page.locator('text=Slow Queries')).toBeVisible();
      await expect(page.locator('text=High CPU Operations')).toBeVisible();
      await expect(page.locator('text=Memory Intensive Tasks')).toBeVisible();

      // Check recommendations
      await expect(page.locator('text=Optimization Recommendations')).toBeVisible();
    });

    test('should display database performance metrics', async ({ page }) => {
      await page.goto('/admin/analytics/database');
      await waitForLoadingToComplete(page);

      // Verify database analytics
      await expect(page.locator('text=Database Performance')).toBeVisible();
      await expect(page.locator('text=Query Response Time')).toBeVisible();
      await expect(page.locator('text=Connection Pool Usage')).toBeVisible();
      await expect(page.locator('text=Lock Wait Time')).toBeVisible();

      // Check slow query analysis
      await expect(page.locator('text=Slow Queries')).toBeVisible();
      await expect(page.locator('text=Query Execution Time')).toBeVisible();
      await expect(page.locator('text=Query Frequency')).toBeVisible();
    });
  });

  test.describe('Data Export and Reporting Features', () => {
    test('should allow exporting analytics data in multiple formats', async ({ page }) => {
      await page.goto('/admin/analytics');
      await waitForLoadingToComplete(page);

      // Test CSV export
      page.on('dialog', async dialog => {
        expect(dialog.message()).toContain('Analytics data exported as CSV');
        await dialog.accept();
      });
      
      await page.click('button:has-text("Export CSV")');
      
      // Verify loading state
      await expect(page.locator('button:has-text("Exporting...")')).toBeVisible();

      // Test JSON export
      page.on('dialog', async dialog => {
        expect(dialog.message()).toContain('Analytics data exported as JSON');
        await dialog.accept();
      });
      
      await page.click('button:has-text("Export JSON")');
    });

    test('should support scheduled report generation', async ({ page }) => {
      await page.goto('/admin/analytics/reports');
      await waitForLoadingToComplete(page);

      // Verify reports section
      await expect(page.locator('text=Scheduled Reports')).toBeVisible();
      await expect(page.locator('button:has-text("Create Report")')).toBeVisible();

      // Create a new scheduled report
      await page.click('button:has-text("Create Report")');
      await expect(page.locator('[data-testid="report-modal"]')).toBeVisible();

      // Fill report details
      await page.fill('[data-testid="report-name"]', 'Weekly Platform Report');
      await page.selectOption('[data-testid="report-frequency"]', 'weekly');
      await page.selectOption('[data-testid="report-format"]', 'pdf');
      
      // Select report sections
      await page.check('[data-testid="include-users"]');
      await page.check('[data-testid="include-repositories"]');
      await page.check('[data-testid="include-performance"]');

      // Save report
      await page.click('[data-testid="save-report"]');
      
      // Verify report was created
      await expect(page.locator('text=Weekly Platform Report')).toBeVisible();
    });

    test('should display historical report archive', async ({ page }) => {
      await page.goto('/admin/analytics/reports');
      await waitForLoadingToComplete(page);

      // Verify report history
      await expect(page.locator('text=Report History')).toBeVisible();
      
      // Check for past reports
      await expect(page.locator('[data-testid="report-entry"]').first()).toBeVisible();
      
      // Test report download
      await page.locator('[data-testid="download-report"]').first().click();
      
      // Verify download initiated (in real implementation, you'd check the download)
    });

    test('should support custom date range for reports', async ({ page }) => {
      await page.goto('/admin/analytics');
      await waitForLoadingToComplete(page);

      // Open custom date range picker
      await page.click('[data-testid="custom-date-range"]');
      
      // Set custom date range
      await page.fill('[data-testid="start-date"]', '2024-01-01');
      await page.fill('[data-testid="end-date"]', '2024-01-31');
      await page.click('[data-testid="apply-date-range"]');

      // Verify data updates for custom range
      await waitForLoadingToComplete(page);
      await expect(page.locator('text=Jan 1, 2024 - Jan 31, 2024')).toBeVisible();
    });
  });

  test.describe('Historical Trend Analysis', () => {
    test('should display long-term trend analysis', async ({ page }) => {
      await page.goto('/admin/analytics/trends');
      await waitForLoadingToComplete(page);

      // Verify trend analysis section
      await expect(page.locator('text=Historical Trends')).toBeVisible();
      await expect(page.locator('text=Year over Year Growth')).toBeVisible();
      await expect(page.locator('text=Seasonal Patterns')).toBeVisible();
      await expect(page.locator('text=Growth Predictions')).toBeVisible();

      // Check trend charts
      await expect(page.locator('canvas, svg')).toHaveCount({ min: 2 });
    });

    test('should support trend comparison between periods', async ({ page }) => {
      await page.goto('/admin/analytics/trends');
      await waitForLoadingToComplete(page);

      // Enable comparison mode
      await page.click('[data-testid="enable-comparison"]');
      
      // Select comparison periods
      await page.selectOption('[data-testid="period-1"]', 'last-month');
      await page.selectOption('[data-testid="period-2"]', 'same-month-last-year');
      
      // Apply comparison
      await page.click('[data-testid="apply-comparison"]');
      
      // Verify comparison data is displayed
      await expect(page.locator('text=Period Comparison')).toBeVisible();
      await expect(page.locator('text=% Change')).toBeVisible();
    });
  });

  test.describe('Capacity Planning and Forecasting', () => {
    test('should display capacity planning metrics', async ({ page }) => {
      await page.goto('/admin/analytics/capacity');
      await waitForLoadingToComplete(page);

      // Verify capacity planning section
      await expect(page.locator('text=Capacity Planning')).toBeVisible();
      await expect(page.locator('text=Current Utilization')).toBeVisible();
      await expect(page.locator('text=Projected Growth')).toBeVisible();
      await expect(page.locator('text=Recommended Actions')).toBeVisible();

      // Check resource forecasting
      await expect(page.locator('text=CPU Forecast')).toBeVisible();
      await expect(page.locator('text=Memory Forecast')).toBeVisible();
      await expect(page.locator('text=Storage Forecast')).toBeVisible();
    });

    test('should provide scaling recommendations', async ({ page }) => {
      await page.goto('/admin/analytics/capacity');
      await waitForLoadingToComplete(page);

      // Scroll to recommendations
      await page.locator('text=Scaling Recommendations').scrollIntoViewIfNeeded();

      // Verify recommendations
      await expect(page.locator('text=Scaling Recommendations')).toBeVisible();
      await expect(page.locator('text=Infrastructure Scaling')).toBeVisible();
      await expect(page.locator('text=Performance Optimization')).toBeVisible();
      await expect(page.locator('text=Cost Optimization')).toBeVisible();
    });
  });

  test.describe('Mobile Analytics Experience', () => {
    test('should be responsive on mobile devices', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      await page.goto('/admin/analytics');
      await waitForLoadingToComplete(page);

      // Verify key elements are accessible on mobile
      await expect(page.locator('h1')).toContainText('Admin Analytics');
      await expect(page.locator('text=Total Views')).toBeVisible();
      
      // Check that charts are responsive
      await expect(page.locator('svg, canvas')).toHaveCount({ min: 1 });
    });

    test('should handle mobile interactions for analytics', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      await page.goto('/admin/analytics');
      await waitForLoadingToComplete(page);

      // Test mobile-friendly controls
      await page.click('button:has-text("Weekly")');
      await expect(page.locator('button:has-text("Weekly")')).toHaveClass(/default/);

      // Test mobile export functionality
      await page.click('button:has-text("Export CSV")');
    });
  });
});
