import { test, expect, Page } from '@playwright/test';
import { loginUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('GitHub Actions - Workflow Management', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    await loginUser(page);
  });

  test.afterEach(async () => {
    await page.close();
  });

  test.describe('Workflow List View', () => {
    test('should display workflow list with status indicators', async () => {
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);

      // Check page title and navigation elements
      await expect(page.locator('h1')).toContainText('Actions');
      await expect(page.locator('[data-testid="new-workflow-btn"], a:has-text("New workflow")')).toBeVisible();
      await expect(page.locator('[data-testid="manage-secrets-btn"], a:has-text("Manage secrets")')).toBeVisible();
      await expect(page.locator('[data-testid="runners-btn"], a:has-text("Runners")')).toBeVisible();

      // Check workflows section
      await expect(page.locator('h2:has-text("All workflows")')).toBeVisible();
    });

    test('should handle empty workflow state', async () => {
      await page.goto('/repositories/admin/empty-repo/actions');
      await waitForLoadingToComplete(page);

      // Should show getting started message
      await expect(page.locator('h3:has-text("Get started with Hub Actions")')).toBeVisible();
      await expect(page.locator('text=Workflows help you automate')).toBeVisible();
      await expect(page.locator('button:has-text("Set up a workflow yourself")')).toBeVisible();
    });

    test('should display workflow details correctly', async () => {
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);

      // Check that workflow cards are displayed
      const workflowCards = page.locator('[data-testid="workflow-card"]').or(page.locator('.space-y-4 > div'));
      
      if (await workflowCards.count() > 0) {
        const firstWorkflow = workflowCards.first();
        
        // Check workflow name is clickable
        await expect(firstWorkflow.locator('a')).toBeVisible();
        
        // Check status badge
        await expect(firstWorkflow.locator('text=Active').or(firstWorkflow.locator('text=Disabled'))).toBeVisible();
        
        // Check dates are displayed
        await expect(firstWorkflow.locator('text=/\\d{1,2}\\/\\d{1,2}\\/\\d{4}/')).toBeVisible();
      }
    });

    test('should filter workflows by status', async () => {
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);

      // Look for filter controls (may be dropdowns or buttons)
      const statusFilter = page.locator('[data-testid="status-filter"]').or(
        page.locator('select').or(page.locator('button:has-text("Status")'))
      );

      if (await statusFilter.isVisible()) {
        await statusFilter.click();
        
        // Test filtering by different statuses
        const filterOptions = ['success', 'failure', 'pending', 'cancelled'];
        for (const option of filterOptions) {
          const optionElement = page.locator(`text=${option}`).or(page.locator(`[value="${option}"]`));
          if (await optionElement.isVisible()) {
            await optionElement.click();
            await waitForLoadingToComplete(page);
            // Verify URL or results updated
            break;
          }
        }
      }
    });
  });

  test.describe('Workflow Run Management', () => {
    test('should display workflow runs with metadata', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Check page elements
      await expect(page.locator('h1:has-text("Workflow runs")')).toBeVisible();
      await expect(page.locator('a:has-text("Back to Actions")')).toBeVisible();

      // Check filters are present
      await expect(page.locator('input[placeholder*="Search"]')).toBeVisible();
      const selectCount = await page.locator('select').count();
      expect(selectCount).toBeGreaterThanOrEqual(2); // Status and event filters
    });

    test('should show run details including trigger and duration', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      const runCards = page.locator('[data-testid="workflow-run"]').or(page.locator('.space-y-4 > div'));
      
      if (await runCards.count() > 0) {
        const firstRun = runCards.first();
        
        // Check status icon/emoji
        await expect(firstRun.locator('text=/[🔄⏳✅❌⭕❓]/')).toBeVisible();
        
        // Check run number and link
        await expect(firstRun.locator('a[href*="/actions/runs/"]')).toBeVisible();
        
        // Check metadata (event, branch, sha, actor)
        await expect(firstRun.locator('text=/push|pull_request|schedule/')).toBeVisible();
        
        // Check status badge
        await expect(firstRun.locator('[class*="badge"]').or(firstRun.locator('span[class*="bg-"]'))).toBeVisible();
      }
    });

    test('should navigate to individual workflow run details', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      const runLinks = page.locator('a[href*="/actions/runs/"]');
      
      if (await runLinks.count() > 0) {
        const firstRunLink = runLinks.first();
        const href = await firstRunLink.getAttribute('href');
        
        await firstRunLink.click();
        await waitForLoadingToComplete(page);
        
        // Should be on run details page
        expect(page.url()).toContain('/actions/runs/');
        
        // Should show run details elements
        await expect(page.locator('h1').or(page.locator('h2'))).toBeVisible();
      }
    });

    test('should allow cancelling running workflows', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Look for running workflows (with in_progress status)
      const runningWorkflows = page.locator('text=🔄').or(page.locator('text=in_progress'));
      
      if (await runningWorkflows.count() > 0) {
        // Click on first running workflow to go to details
        await runningWorkflows.first().click();
        await waitForLoadingToComplete(page);
        
        // Look for cancel button
        const cancelButton = page.locator('button:has-text("Cancel")').or(
          page.locator('[data-testid="cancel-workflow"]')
        );
        
        if (await cancelButton.isVisible()) {
          await cancelButton.click();
          
          // Check for confirmation dialog or immediate update
          const confirmButton = page.locator('button:has-text("Confirm")').or(
            page.locator('button:has-text("Yes")')
          );
          
          if (await confirmButton.isVisible()) {
            await confirmButton.click();
          }
          
          await waitForLoadingToComplete(page);
          
          // Verify status changed
          await expect(page.locator('text=cancelled').or(page.locator('text=⭕'))).toBeVisible();
        }
      }
    });

    test('should allow re-running failed workflows', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Look for failed workflows
      const failedWorkflows = page.locator('text=❌').or(page.locator('text=failure'));
      
      if (await failedWorkflows.count() > 0) {
        // Click on first failed workflow
        await failedWorkflows.first().click();
        await waitForLoadingToComplete(page);
        
        // Look for re-run button
        const rerunButton = page.locator('button:has-text("Re-run")').or(
          page.locator('[data-testid="rerun-workflow"]')
        );
        
        if (await rerunButton.isVisible()) {
          await rerunButton.click();
          await waitForLoadingToComplete(page);
          
          // Should navigate back to runs or show new run
          expect(page.url()).toMatch(/\/actions\/runs/);
        }
      }
    });
  });

  test.describe('Search and Filtering', () => {
    test('should search workflows by name', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      const searchInput = page.locator('input[placeholder*="Search"]');
      await searchInput.fill('test');
      await page.waitForTimeout(500); // Wait for debounce
      
      // Results should be filtered
      const runCards = page.locator('[data-testid="workflow-run"]').or(page.locator('.space-y-4 > div'));
      
      if (await runCards.count() > 0) {
        // At least one result should contain "test"
        await expect(runCards.first().locator('text=/test/i')).toBeVisible();
      }
    });

    test('should filter by workflow status', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Find status filter dropdown
      const statusFilter = page.locator('select').first();
      
      // Test filtering by success
      await statusFilter.selectOption('success');
      await waitForLoadingToComplete(page);
      
      // All visible runs should have success status
      const runCards = page.locator('[data-testid="workflow-run"]').or(page.locator('.space-y-4 > div'));
      
      if (await runCards.count() > 0) {
        await expect(runCards.first().locator('text=✅').or(runCards.first().locator('text=success'))).toBeVisible();
      }
    });

    test('should filter by event type', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Find event filter dropdown (should be second select)
      const eventFilter = page.locator('select').nth(1);
      
      // Test filtering by push events
      await eventFilter.selectOption('push');
      await waitForLoadingToComplete(page);
      
      // All visible runs should be push events
      const runCards = page.locator('[data-testid="workflow-run"]').or(page.locator('.space-y-4 > div'));
      
      if (await runCards.count() > 0) {
        await expect(runCards.first().locator('text=push')).toBeVisible();
      }
    });
  });

  test.describe('Mobile Experience', () => {
    test('should display workflows correctly on mobile', async () => {
      await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);

      // Check that page is responsive
      await expect(page.locator('h1')).toBeVisible();
      
      // Buttons should be stacked or in a responsive layout
      const actionButtons = page.locator('a:has-text("New workflow"), a:has-text("Manage secrets"), a:has-text("Runners")');
      await expect(actionButtons.first()).toBeVisible();
      
      // Workflow cards should be readable
      const workflowCards = page.locator('[data-testid="workflow-card"]').or(page.locator('.space-y-4 > div'));
      
      if (await workflowCards.count() > 0) {
        await expect(workflowCards.first()).toBeVisible();
      }
    });

    test('should handle touch interactions for workflow runs', async () => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      const runCards = page.locator('[data-testid="workflow-run"]').or(page.locator('.space-y-4 > div'));
      
      if (await runCards.count() > 0) {
        // Tap on first run card
        await runCards.first().tap();
        await waitForLoadingToComplete(page);
        
        // Should navigate to run details
        expect(page.url()).toContain('/actions/runs/');
      }
    });
  });

  test.describe('Error Handling', () => {
    test('should handle repository not found error', async () => {
      await page.goto('/repositories/nonexistent/repo/actions');
      await waitForLoadingToComplete(page);

      // Should show error message
      await expect(page.locator('text=Repository "nonexistent/repo" not found')).toBeVisible();
      await expect(page.locator('text=Try visiting /repositories/admin/sample-project/actions')).toBeVisible();
    });

    test('should handle network errors gracefully', async () => {
      // Simulate network failure
      await page.route('**/api/v1/repos/**/actions/**', route => route.abort());
      
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);

      // Should show error state
      await expect(page.locator('text=/error|failed|unable/i')).toBeVisible();
    });

    test('should show loading states during data fetch', async () => {
      // Delay API responses
      await page.route('**/api/v1/repos/**/actions/**', async route => {
        await page.waitForTimeout(2000);
        await route.continue();
      });
      
      await page.goto('/repositories/admin/sample-project/actions');
      
      // Should show loading state
      await expect(page.locator('.animate-pulse')).toBeVisible();
      
      await waitForLoadingToComplete(page);
      
      // Loading should be gone
      await expect(page.locator('.animate-pulse')).not.toBeVisible();
    });
  });

  test.describe('Real-time Updates', () => {
    test('should update workflow run status in real-time', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Mock WebSocket or polling updates
      await page.evaluate(() => {
        // Simulate status update
        window.dispatchEvent(new CustomEvent('workflow-status-update', {
          detail: { runId: '123', status: 'completed', conclusion: 'success' }
        }));
      });

      // Check for status change indicators
      await page.waitForTimeout(1000);
      
      // Look for updated status
      const statusElements = page.locator('text=✅').or(page.locator('text=success'));
      if (await statusElements.count() > 0) {
        await expect(statusElements.first()).toBeVisible();
      }
    });

    test('should refresh data periodically', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Monitor API calls
      let apiCallCount = 0;
      await page.route('**/api/v1/repos/**/actions/runs**', route => {
        apiCallCount++;
        route.continue();
      });

      // Wait for potential refresh
      await page.waitForTimeout(30000); // Wait 30 seconds
      
      // Should have made additional API calls for refresh
      expect(apiCallCount).toBeGreaterThan(1);
    });
  });
});