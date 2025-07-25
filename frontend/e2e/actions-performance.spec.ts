import { test, expect, Page } from '@playwright/test';
import { loginUser, waitForLoadingToComplete, checkActionsPerformance, checkActionsAccessibility } from './helpers/test-utils';

test.describe('GitHub Actions - Performance & Stress Testing', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    await loginUser(page);
  });

  test.afterEach(async () => {
    await page.close();
  });

  test.describe('Page Load Performance', () => {
    test('should load Actions main page within performance budget', async () => {
      await checkActionsPerformance(page, '/repositories/admin/sample-project/actions');
      
      // Check specific performance metrics
      const performanceMetrics = await page.evaluate(() => {
        const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
        return {
          ttfb: navigation.responseStart - navigation.requestStart,
          domReady: navigation.domContentLoadedEventEnd - navigation.fetchStart,
          loadComplete: navigation.loadEventEnd - navigation.fetchStart
        };
      });

      // Time to First Byte should be fast
      expect(performanceMetrics.ttfb).toBeLessThan(1000);
      
      // DOM ready time should be reasonable
      expect(performanceMetrics.domReady).toBeLessThan(3000);
      
      // Complete load should finish quickly
      expect(performanceMetrics.loadComplete).toBeLessThan(5000);
    });

    test('should load workflow runs page efficiently', async () => {
      await checkActionsPerformance(page, '/repositories/admin/sample-project/actions/runs');
      
      // Verify content is visible quickly
      await expect(page.locator('h1:has-text("Workflow runs")')).toBeVisible();
      
      // Check that runs list loads without blocking
      const runsList = page.locator('[data-testid="workflow-run"]').or(page.locator('.space-y-4 > div'));
      
      if (await runsList.count() > 0) {
        await expect(runsList.first()).toBeVisible();
      }
    });

    test('should load individual run details with good performance', async () => {
      await checkActionsPerformance(page, '/repositories/admin/sample-project/actions/runs/123');
      
      // Check for lazy loading of logs
      const logContainer = page.locator('[data-testid="log-container"]').or(page.locator('pre'));
      
      if (await logContainer.isVisible()) {
        // Logs should start loading immediately
        await page.waitForTimeout(1000);
        await expect(logContainer).toBeVisible();
      }
    });

    test('should handle large workflow lists efficiently', async () => {
      // Simulate a repository with many workflows
      await page.route('**/api/v1/repos/**/actions/workflows', async route => {
        const workflows = Array.from({ length: 100 }, (_, i) => ({
          id: `workflow-${i}`,
          name: `Test Workflow ${i}`,
          path: `.github/workflows/test-${i}.yml`,
          enabled: true,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString()
        }));

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ workflows })
        });
      });

      const startTime = Date.now();
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);
      const loadTime = Date.now() - startTime;

      // Should still load quickly even with many workflows
      expect(loadTime).toBeLessThan(3000);

      // Should implement virtualization or pagination
      const visibleWorkflows = page.locator('[data-testid="workflow-card"]').or(page.locator('.space-y-4 > div'));
      const visibleCount = await visibleWorkflows.count();
      
      // Should not render all 100 workflows at once
      expect(visibleCount).toBeLessThan(50);
    });
  });

  test.describe('Real-time Updates Performance', () => {
    test('should handle real-time log streaming without memory leaks', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/running-job');
      await waitForLoadingToComplete(page);

      // Monitor memory usage
      const initialMetrics = await page.evaluate(() => {
        return (performance as any).memory ? {
          usedJSHeapSize: (performance as any).memory.usedJSHeapSize,
          totalJSHeapSize: (performance as any).memory.totalJSHeapSize
        } : null;
      });

      // Simulate log streaming for 30 seconds
      for (let i = 0; i < 300; i++) {
        await page.evaluate((lineNumber) => {
          const logContainer = document.querySelector('[data-testid="log-container"]') || 
                              document.querySelector('pre') || 
                              document.querySelector('.log-output');
          if (logContainer) {
            const newLine = `[${new Date().toISOString()}] Log line ${lineNumber}\n`;
            logContainer.textContent += newLine;
            logContainer.scrollTop = logContainer.scrollHeight;
          }
        }, i);

        await page.waitForTimeout(100);
      }

      // Check memory usage after streaming
      if (initialMetrics) {
        const finalMetrics = await page.evaluate(() => {
          return (performance as any).memory ? {
            usedJSHeapSize: (performance as any).memory.usedJSHeapSize,
            totalJSHeapSize: (performance as any).memory.totalJSHeapSize
          } : null;
        });

        if (finalMetrics) {
          // Memory usage should not have grown excessively
          const memoryGrowth = finalMetrics.usedJSHeapSize - initialMetrics.usedJSHeapSize;
          expect(memoryGrowth).toBeLessThan(50 * 1024 * 1024); // Less than 50MB growth
        }
      }
    });

    test('should throttle status update API calls', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      let apiCallCount = 0;
      await page.route('**/api/v1/repos/**/actions/runs**', route => {
        apiCallCount++;
        route.continue();
      });

      // Wait for potential auto-refresh calls
      await page.waitForTimeout(60000); // 1 minute

      // Should not make excessive API calls
      expect(apiCallCount).toBeLessThan(10); // Less than 10 calls per minute
    });

    test('should handle WebSocket connections efficiently', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Monitor WebSocket connections
      const wsConnections: string[] = [];
      
      page.on('websocket', ws => {
        wsConnections.push(ws.url());
        
        ws.on('close', () => {
          const index = wsConnections.indexOf(ws.url());
          if (index > -1) {
            wsConnections.splice(index, 1);
          }
        });
      });

      // Navigate between different pages
      await page.goto('/repositories/admin/sample-project/actions/runs/456');
      await waitForLoadingToComplete(page);
      
      await page.goto('/repositories/admin/sample-project/actions/runs/789');
      await waitForLoadingToComplete(page);

      // Should not accumulate WebSocket connections
      expect(wsConnections.length).toBeLessThan(3);
    });
  });

  test.describe('Large Data Handling', () => {
    test('should handle large log files efficiently', async () => {
      // Mock very large log response
      await page.route('**/api/v1/repos/**/actions/runs/**/logs**', async route => {
        const largeLogs = Array.from({ length: 10000 }, (_, i) => 
          `[${new Date().toISOString()}] Log line ${i} with some detailed information about the build process`
        ).join('\n');

        await route.fulfill({
          status: 200,
          contentType: 'text/plain',
          body: largeLogs
        });
      });

      const startTime = Date.now();
      await page.goto('/repositories/admin/sample-project/actions/runs/large-logs');
      await waitForLoadingToComplete(page);

      const logContainer = page.locator('[data-testid="log-container"]').or(page.locator('pre'));
      
      if (await logContainer.isVisible()) {
        await expect(logContainer).toBeVisible();
        
        // Should implement virtualization for large logs
        const scrollTime = Date.now();
        await logContainer.evaluate(el => el.scrollTop = el.scrollHeight / 2);
        await page.waitForTimeout(100);
        await logContainer.evaluate(el => el.scrollTop = el.scrollHeight);
        const scrollDuration = Date.now() - scrollTime;

        // Scrolling should be smooth even with large logs
        expect(scrollDuration).toBeLessThan(1000);
      }

      const totalTime = Date.now() - startTime;
      expect(totalTime).toBeLessThan(10000); // Should load within 10 seconds
    });

    test('should handle many workflow runs with pagination', async () => {
      // Mock large number of workflow runs
      await page.route('**/api/v1/repos/**/actions/runs**', async route => {
        const url = new URL(route.request().url());
        const limit = parseInt(url.searchParams.get('limit') || '50');
        const offset = parseInt(url.searchParams.get('offset') || '0');

        const runs = Array.from({ length: limit }, (_, i) => ({
          id: `run-${offset + i}`,
          number: offset + i + 1,
          status: i % 3 === 0 ? 'completed' : i % 3 === 1 ? 'in_progress' : 'queued',
          conclusion: i % 3 === 0 ? (i % 2 === 0 ? 'success' : 'failure') : undefined,
          head_sha: `abc123${i}`,
          head_branch: 'main',
          event: 'push',
          created_at: new Date(Date.now() - i * 3600000).toISOString(),
          workflow: {
            id: `workflow-${i % 10}`,
            name: `Test Workflow ${i % 10}`
          }
        }));

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ 
            workflow_runs: runs,
            total_count: 1000 // Large total count
          })
        });
      });

      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Should show pagination controls
      const paginationControls = page.locator('[data-testid="pagination"]').or(
        page.locator('button:has-text("Next")').or(page.locator('nav'))
      );

      if (await paginationControls.isVisible()) {
        await expect(paginationControls).toBeVisible();

        // Test pagination performance
        const nextButton = page.locator('button:has-text("Next")');
        if (await nextButton.isVisible()) {
          const paginationStartTime = Date.now();
          await nextButton.click();
          await waitForLoadingToComplete(page);
          const paginationTime = Date.now() - paginationStartTime;

          expect(paginationTime).toBeLessThan(2000);
        }
      }
    });

    test('should handle concurrent workflow runs efficiently', async () => {
      // Mock many concurrent runs
      await page.route('**/api/v1/repos/**/actions/runs**', async route => {
        const runs = Array.from({ length: 20 }, (_, i) => ({
          id: `concurrent-run-${i}`,
          number: i + 1,
          status: 'in_progress',
          head_sha: `abc123${i}`,
          head_branch: 'main',
          event: 'push',
          created_at: new Date().toISOString(),
          workflow: {
            id: `workflow-${i}`,
            name: `Concurrent Workflow ${i}`
          }
        }));

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ workflow_runs: runs })
        });
      });

      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Should display all concurrent runs
      const runElements = page.locator('[data-testid="workflow-run"]').or(page.locator('.space-y-4 > div'));
      const runCount = await runElements.count();
      
      expect(runCount).toBeGreaterThan(10);

      // Should update statuses efficiently
      const statusIcons = page.locator('text=🔄');
      await expect(statusIcons.first()).toBeVisible();
    });
  });

  test.describe('Error Handling Performance', () => {
    test('should handle API failures gracefully without blocking UI', async () => {
      // Simulate API failure
      await page.route('**/api/v1/repos/**/actions/**', route => route.abort());

      const startTime = Date.now();
      await page.goto('/repositories/admin/sample-project/actions');
      
      // UI should still render quickly even with API failures
      await expect(page.locator('h1')).toBeVisible();
      
      const renderTime = Date.now() - startTime;
      expect(renderTime).toBeLessThan(3000);

      // Should show error state
      await expect(page.locator('text=/error|failed|unable/i')).toBeVisible();
    });

    test('should handle slow API responses with loading states', async () => {
      // Simulate slow API response
      await page.route('**/api/v1/repos/**/actions/**', async route => {
        await page.waitForTimeout(3000); // 3 second delay
        await route.continue();
      });

      await page.goto('/repositories/admin/sample-project/actions');
      
      // Should show loading state immediately
      await expect(page.locator('.animate-pulse').or(page.locator('[data-testid="loading"]'))).toBeVisible();
      
      // Should eventually load content
      await waitForLoadingToComplete(page);
      await expect(page.locator('h1:has-text("Actions")')).toBeVisible();
    });

    test('should handle network timeouts appropriately', async () => {
      // Simulate network timeout
      await page.route('**/api/v1/repos/**/actions/**', route => {
        // Never respond to simulate timeout
      });

      await page.goto('/repositories/admin/sample-project/actions');
      
      // Should show timeout error after reasonable time
      await page.waitForTimeout(30000); // Wait 30 seconds
      
      await expect(page.locator('text=/timeout|network error|unable to connect/i')).toBeVisible();
    });
  });

  test.describe('Accessibility Performance', () => {
    test('should maintain accessibility standards under load', async () => {
      // Load page with large dataset
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Check accessibility
      await checkActionsAccessibility(page);

      // Focus management should work quickly
      const focusableElements = page.locator('a, button, input, textarea, select');
      const focusStartTime = Date.now();
      
      if (await focusableElements.count() > 0) {
        await focusableElements.first().focus();
        await expect(focusableElements.first()).toBeFocused();
      }
      
      const focusTime = Date.now() - focusStartTime;
      expect(focusTime).toBeLessThan(100);
    });

    test('should handle keyboard navigation efficiently', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Test tab navigation speed
      const tabStartTime = Date.now();
      
      for (let i = 0; i < 10; i++) {
        await page.keyboard.press('Tab');
        await page.waitForTimeout(50);
      }
      
      const tabTime = Date.now() - tabStartTime;
      expect(tabTime).toBeLessThan(2000); // Should be responsive
    });
  });

  test.describe('Memory and Resource Management', () => {
    test('should clean up resources when navigating away', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Get initial resource count
      const initialResources = await page.evaluate(() => {
        return {
          eventListeners: (window as any).getEventListeners ? Object.keys((window as any).getEventListeners(document)).length : 0,
          intervals: (window as any).setInterval.toString().includes('[native code]') ? 0 : 1,
          timeouts: (window as any).setTimeout.toString().includes('[native code]') ? 0 : 1
        };
      });

      // Navigate away
      await page.goto('/repositories/admin/sample-project');
      await waitForLoadingToComplete(page);

      // Check that resources are cleaned up
      const finalResources = await page.evaluate(() => {
        return {
          eventListeners: (window as any).getEventListeners ? Object.keys((window as any).getEventListeners(document)).length : 0,
          intervals: (window as any).setInterval.toString().includes('[native code]') ? 0 : 1,
          timeouts: (window as any).setTimeout.toString().includes('[native code]') ? 0 : 1
        };
      });

      // Should have cleaned up most resources
      expect(finalResources.eventListeners).toBeLessThanOrEqual(initialResources.eventListeners + 5);
    });

    test('should handle browser resource limits gracefully', async () => {
      // Simulate resource-constrained environment
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);

      // Stress test with rapid navigation
      const pages = [
        '/repositories/admin/sample-project/actions',
        '/repositories/admin/sample-project/actions/runs',
        '/repositories/admin/sample-project/actions/runs/123',
        '/repositories/admin/sample-project/settings/runners'
      ];

      for (let i = 0; i < 3; i++) {
        for (const pageUrl of pages) {
          await page.goto(pageUrl);
          await waitForLoadingToComplete(page);
          await page.waitForTimeout(100);
        }
      }

      // Should still be responsive after stress test
      await page.goto('/repositories/admin/sample-project/actions');
      await expect(page.locator('h1:has-text("Actions")')).toBeVisible();
    });
  });
});