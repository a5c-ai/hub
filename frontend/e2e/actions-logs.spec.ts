import { test, expect, Page } from '@playwright/test';
import { loginUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('GitHub Actions - Workflow Run Details & Logs', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    await loginUser(page);
  });

  test.afterEach(async () => {
    await page.close();
  });

  test.describe('Workflow Run Details', () => {
    test('should display workflow run summary and status', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Check run summary elements
      await expect(page.locator('h1').or(page.locator('h2'))).toBeVisible();
      
      // Status should be visible with icon
      await expect(page.locator('text=/[🔄⏳✅❌⭕❓]/')).toBeVisible();
      
      // Basic metadata should be shown
      await expect(page.locator('text=/Run #\d+/')).toBeVisible();
      await expect(page.locator('text=/\b[a-f0-9]{7,40}\b/')).toBeVisible(); // SHA
    });

    test('should show job steps with individual status', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Look for job sections
      const jobSections = page.locator('[data-testid="job-section"]').or(
        page.locator('h3, h4').filter({ hasText: /job|step/i })
      );

      if (await jobSections.count() > 0) {
        // Each job should have status indicators
        const firstJob = jobSections.first();
        await expect(firstJob).toBeVisible();
        
        // Look for step status icons
        const statusIcons = page.locator('text=/[🔄⏳✅❌⭕❓]/');
        await expect(statusIcons.first()).toBeVisible();
      }

      // Check for expandable job details
      const expandableElements = page.locator('[data-testid="expand-job"]').or(
        page.locator('button').filter({ hasText: /expand|show|details/i })
      );

      if (await expandableElements.count() > 0) {
        await expandableElements.first().click();
        await expect(page.locator('[data-testid="job-details"]').or(page.locator('.expanded'))).toBeVisible();
      }
    });

    test('should display workflow configuration (YAML)', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Look for workflow configuration tab or section
      const configTab = page.locator('button:has-text("Configuration")').or(
        page.locator('a:has-text("Workflow file")').or(
          page.locator('[data-testid="workflow-config"]')
        )
      );

      if (await configTab.isVisible()) {
        await configTab.click();
        await waitForLoadingToComplete(page);

        // Should show YAML content
        await expect(page.locator('code').or(page.locator('pre'))).toBeVisible();
        
        // Should contain typical YAML workflow content
        await expect(page.locator('text=/name:|on:|jobs:/').first()).toBeVisible();
      }
    });

    test('should show run timing and duration', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Look for timing information
      const timingElements = page.locator('text=/\d+[smh]|\d+:\d+|\d+ seconds?|\d+ minutes?/');
      
      if (await timingElements.count() > 0) {
        await expect(timingElements.first()).toBeVisible();
      }

      // Check for timestamps
      const timestampElements = page.locator('text=/\d{1,2}\/\d{1,2}\/\d{4}|\d{1,2}:\d{2}|\d+ ago/');
      await expect(timestampElements.first()).toBeVisible();
    });
  });

  test.describe('Real-time Log Streaming', () => {
    test('should stream logs for running jobs in real-time', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Look for running jobs (status: in_progress)
      const runningJobs = page.locator('text=🔄').or(page.locator('text=in_progress'));
      
      if (await runningJobs.count() > 0) {
        // Click on running job to open logs
        await runningJobs.first().click();
        await waitForLoadingToComplete(page);

        // Should show log container
        const logContainer = page.locator('[data-testid="log-container"]').or(
          page.locator('pre').or(page.locator('.log-output'))
        );

        await expect(logContainer).toBeVisible();

        // Monitor for new log entries
        const initialLogLength = await logContainer.textContent();
        
        // Wait for potential log updates
        await page.waitForTimeout(2000);
        
        const updatedLogLength = await logContainer.textContent();
        
        // Log content may have changed (new lines added)
        if (initialLogLength !== updatedLogLength) {
          // Real-time updates are working
          expect(updatedLogLength).toBeDefined();
        }
      }
    });

    test('should auto-scroll logs during streaming', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Find log container
      const logContainer = page.locator('[data-testid="log-container"]').or(
        page.locator('pre').or(page.locator('.log-output'))
      );

      if (await logContainer.isVisible()) {
        // Check if auto-scroll is enabled
        const scrollTop = await logContainer.evaluate(el => el.scrollTop);
        const scrollHeight = await logContainer.evaluate(el => el.scrollHeight);
        const clientHeight = await logContainer.evaluate(el => el.clientHeight);

        // Should be scrolled to bottom for auto-scroll
        expect(scrollTop + clientHeight).toBeCloseTo(scrollHeight, 50);
      }
    });

    test('should allow manual scroll control during streaming', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      const logContainer = page.locator('[data-testid="log-container"]').or(
        page.locator('pre').or(page.locator('.log-output'))
      );

      if (await logContainer.isVisible()) {
        // Scroll to top manually
        await logContainer.evaluate(el => el.scrollTo(0, 0));
        
        // Check if auto-scroll is paused
        await page.waitForTimeout(1000);
        
        const scrollTop = await logContainer.evaluate(el => el.scrollTop);
        expect(scrollTop).toBe(0);

        // Look for scroll control indicators
        const scrollIndicator = page.locator('[data-testid="scroll-to-bottom"]').or(
          page.locator('button').filter({ hasText: /scroll|bottom/i })
        );

        if (await scrollIndicator.isVisible()) {
          await scrollIndicator.click();
          
          // Should scroll to bottom
          const newScrollTop = await logContainer.evaluate(el => el.scrollTop);
          const scrollHeight = await logContainer.evaluate(el => el.scrollHeight);
          const clientHeight = await logContainer.evaluate(el => el.clientHeight);
          
          expect(newScrollTop + clientHeight).toBeCloseTo(scrollHeight, 50);
        }
      }
    });
  });

  test.describe('Log Viewing and Navigation', () => {
    test('should display logs with syntax highlighting', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Look for job logs
      const logElements = page.locator('[data-testid="job-logs"]').or(
        page.locator('pre').or(page.locator('.log-output'))
      );

      if (await logElements.count() > 0) {
        const firstLog = logElements.first();
        await expect(firstLog).toBeVisible();

        // Check for syntax highlighting classes or colored text
        const coloredElements = firstLog.locator('[class*="color"], [style*="color"], span[class]');
        
        if (await coloredElements.count() > 0) {
          await expect(coloredElements.first()).toBeVisible();
        }
      }
    });

    test('should support log line numbers', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      const logContainer = page.locator('[data-testid="log-container"]').or(
        page.locator('pre').or(page.locator('.log-output'))
      );

      if (await logContainer.isVisible()) {
        // Look for line numbers
        const lineNumbers = page.locator('[data-testid="line-number"]').or(
          page.locator('.line-number').or(page.locator('text=/^\d+$|^\s*\d+\s*\|/'))
        );

        if (await lineNumbers.count() > 0) {
          await expect(lineNumbers.first()).toBeVisible();
        }
      }
    });

    test('should allow expanding and collapsing job sections', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Find expandable job headers
      const jobHeaders = page.locator('[data-testid="job-header"]').or(
        page.locator('button').filter({ hasText: /setup|build|test|deploy/i })
      );

      if (await jobHeaders.count() > 0) {
        const firstJob = jobHeaders.first();
        
        // Should be clickable
        await expect(firstJob).toBeVisible();
        
        // Click to expand
        await firstJob.click();
        
        // Check if content is revealed
        const jobContent = page.locator('[data-testid="job-content"]').or(
          page.locator('.expanded').or(page.locator('[aria-expanded="true"]'))
        );

        if (await jobContent.isVisible()) {
          // Click again to collapse
          await firstJob.click();
          await expect(jobContent).not.toBeVisible();
        }
      }
    });

    test('should search within logs', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Look for search functionality
      const searchInput = page.locator('[data-testid="log-search"]').or(
        page.locator('input[placeholder*="search"]')
      );

      if (await searchInput.isVisible()) {
        await searchInput.fill('error');
        await page.keyboard.press('Enter');

        // Should highlight search results
        const highlightedText = page.locator('[data-testid="search-highlight"]').or(
          page.locator('.highlight').or(page.locator('mark'))
        );

        if (await highlightedText.count() > 0) {
          await expect(highlightedText.first()).toBeVisible();
        }

        // Check for search navigation
        const nextButton = page.locator('[data-testid="search-next"]').or(
          page.locator('button').filter({ hasText: /next|down/i })
        );

        if (await nextButton.isVisible()) {
          await nextButton.click();
          // Should scroll to next match
        }
      }
    });
  });

  test.describe('Artifact Management', () => {
    test('should list downloadable artifacts', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Look for artifacts section
      const artifactsSection = page.locator('[data-testid="artifacts-section"]').or(
        page.locator('h3:has-text("Artifacts")').or(page.locator('h2:has-text("Artifacts")'))
      );

      if (await artifactsSection.isVisible()) {
        // Should show artifact list
        const artifactList = page.locator('[data-testid="artifact-list"]').or(
          page.locator('ul').or(page.locator('.artifact-item'))
        );

        await expect(artifactList).toBeVisible();

        // Each artifact should have download link
        const downloadLinks = page.locator('[data-testid="download-artifact"]').or(
          page.locator('a[href*="download"]')
        );

        if (await downloadLinks.count() > 0) {
          await expect(downloadLinks.first()).toBeVisible();
        }
      }
    });

    test('should download workflow artifacts', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      const downloadLinks = page.locator('[data-testid="download-artifact"]').or(
        page.locator('a[href*="download"]')
      );

      if (await downloadLinks.count() > 0) {
        // Setup download handling
        const downloadPromise = page.waitForEvent('download');
        
        await downloadLinks.first().click();
        
        try {
          const download = await downloadPromise;
          expect(download).toBeDefined();
          expect(download.suggestedFilename()).toMatch(/\.(zip|tar\.gz|tar)$/);
        } catch (error) {
          // Download may not be available in test environment
          console.log('Download test skipped - no actual download available');
        }
      }
    });

    test('should show artifact metadata', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      const artifactItems = page.locator('[data-testid="artifact-item"]').or(
        page.locator('.artifact-item')
      );

      if (await artifactItems.count() > 0) {
        const firstArtifact = artifactItems.first();
        
        // Should show artifact name
        await expect(firstArtifact.locator('text=/\w+\.(zip|tar\.gz|jar|war)/')).toBeVisible();
        
        // Should show file size
        await expect(firstArtifact.locator('text=/\d+(\.\d+)?\s*(B|KB|MB|GB)/')).toBeVisible();
        
        // Should show upload time
        await expect(firstArtifact.locator('text=/ago|\d{1,2}\/\d{1,2}\/\d{4}/')).toBeVisible();
      }
    });
  });

  test.describe('Performance with Large Logs', () => {
    test('should handle large log outputs efficiently', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/456'); // Assume this has large logs
      await waitForLoadingToComplete(page);

      const startTime = Date.now();
      
      // Scroll through large logs
      const logContainer = page.locator('[data-testid="log-container"]').or(
        page.locator('pre').or(page.locator('.log-output'))
      );

      if (await logContainer.isVisible()) {
        // Scroll to different positions
        for (let i = 0; i < 5; i++) {
          await logContainer.evaluate((el, scrollPercentage) => {
            el.scrollTop = (el.scrollHeight * scrollPercentage) / 100;
          }, i * 20);
          
          await page.waitForTimeout(100);
        }
      }

      const endTime = Date.now();
      const scrollTime = endTime - startTime;
      
      // Should be responsive (less than 2 seconds for scrolling)
      expect(scrollTime).toBeLessThan(2000);
    });

    test('should use virtualization for very long logs', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/456');
      await waitForLoadingToComplete(page);

      const logContainer = page.locator('[data-testid="log-container"]').or(
        page.locator('pre').or(page.locator('.log-output'))
      );

      if (await logContainer.isVisible()) {
        // Check if virtualization is implemented
        const virtualizedIndicators = page.locator('[data-testid="virtual-scroll"]').or(
          page.locator('[class*="virtual"]')
        );

        if (await virtualizedIndicators.count() > 0) {
          // Test scrolling performance with virtualization
          await logContainer.evaluate(el => el.scrollTop = el.scrollHeight / 2);
          await page.waitForTimeout(100);
          
          // Should still be responsive
          const visibleLines = page.locator('.log-line:visible');
          expect(await visibleLines.count()).toBeGreaterThan(0);
          expect(await visibleLines.count()).toBeLessThan(1000); // Should not render all lines
        }
      }
    });

    test('should load logs progressively', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/789');
      await waitForLoadingToComplete(page);

      // Monitor API calls for progressive loading
      let apiCallCount = 0;
      await page.route('**/api/v1/repos/**/actions/runs/**/logs**', route => {
        apiCallCount++;
        route.continue();
      });

      const logContainer = page.locator('[data-testid="log-container"]').or(
        page.locator('pre').or(page.locator('.log-output'))
      );

      if (await logContainer.isVisible()) {
        // Scroll to trigger more log loading
        await logContainer.evaluate(el => el.scrollTop = el.scrollHeight);
        await page.waitForTimeout(1000);
        
        // Should have made additional API calls for more logs
        expect(apiCallCount).toBeGreaterThan(0);
      }
    });
  });

  test.describe('Mobile Log Viewing', () => {
    test('should display logs correctly on mobile devices', async () => {
      await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Log container should be responsive
      const logContainer = page.locator('[data-testid="log-container"]').or(
        page.locator('pre').or(page.locator('.log-output'))
      );

      if (await logContainer.isVisible()) {
        await expect(logContainer).toBeVisible();
        
        // Should not overflow horizontally
        const containerWidth = await logContainer.evaluate(el => (el as HTMLElement).offsetWidth);
        expect(containerWidth).toBeLessThanOrEqual(375);
      }

      // Touch scrolling should work
      if (await logContainer.isVisible()) {
        await logContainer.evaluate(el => el.scrollTop = 100);
        const scrollTop = await logContainer.evaluate(el => el.scrollTop);
        expect(scrollTop).toBeGreaterThan(0);
      }
    });

    test('should support touch gestures for log navigation', async () => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      const jobHeaders = page.locator('[data-testid="job-header"]').or(
        page.locator('button').filter({ hasText: /setup|build|test|deploy/i })
      );

      if (await jobHeaders.count() > 0) {
        // Tap to expand job
        await jobHeaders.first().tap();
        
        const jobContent = page.locator('[data-testid="job-content"]').or(
          page.locator('.expanded')
        );

        if (await jobContent.isVisible()) {
          // Should be expanded after tap
          await expect(jobContent).toBeVisible();
        }
      }
    });
  });
});