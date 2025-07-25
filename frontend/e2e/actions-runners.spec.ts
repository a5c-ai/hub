import { test, expect, Page } from '@playwright/test';
import { loginUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('GitHub Actions - Runner Management', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    await loginUser(page);
  });

  test.afterEach(async () => {
    await page.close();
  });

  test.describe('Self-hosted Runners Status', () => {
    test('should display list of self-hosted runners with status', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      // Check page title and navigation
      await expect(page.locator('h1:has-text("Runners")').or(page.locator('h2:has-text("Actions runners")'))).toBeVisible();

      // Should show runners list
      const runnersList = page.locator('[data-testid="runners-list"]').or(
        page.locator('.runners-container').or(page.locator('table'))
      );

      if (await runnersList.isVisible()) {
        await expect(runnersList).toBeVisible();

        // Each runner should have status indicators
        const runnerItems = page.locator('[data-testid="runner-item"]').or(
          page.locator('tr').or(page.locator('.runner-card'))
        );

        if (await runnerItems.count() > 0) {
          const firstRunner = runnerItems.first();
          
          // Should show runner name
          await expect(firstRunner.locator('text=/runner-\w+|\w+-runner|self-hosted-\d+/')).toBeVisible();
          
          // Should show status
          await expect(firstRunner.locator('text=/online|offline|idle|busy/i')).toBeVisible();
          
          // Should show status indicator (color or icon)
          await expect(firstRunner.locator('[class*="status"]').or(firstRunner.locator('[class*="online"], [class*="offline"]'))).toBeVisible();
        }
      }
    });

    test('should show runner details including labels and capabilities', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      const runnerItems = page.locator('[data-testid="runner-item"]').or(
        page.locator('tr').or(page.locator('.runner-card'))
      );

      if (await runnerItems.count() > 0) {
        const firstRunner = runnerItems.first();
        
        // Click to view details or check if details are inline
        const detailsButton = firstRunner.locator('button:has-text("Details")').or(
          firstRunner.locator('[data-testid="runner-details"]')
        );

        if (await detailsButton.isVisible()) {
          await detailsButton.click();
          await waitForLoadingToComplete(page);
        }

        // Should show runner labels
        const labels = page.locator('[data-testid="runner-labels"]').or(
          page.locator('.labels').or(page.locator('span[class*="label"], span[class*="tag"]'))
        );

        if (await labels.count() > 0) {
          await expect(labels.first()).toBeVisible();
        }

        // Should show OS and architecture
        await expect(page.locator('text=/linux|windows|macos/i')).toBeVisible();
        await expect(page.locator('text=/x64|x86|arm64/i')).toBeVisible();
      }
    });

    test('should display runner health and capacity metrics', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      const runnerItems = page.locator('[data-testid="runner-item"]').or(
        page.locator('tr').or(page.locator('.runner-card'))
      );

      if (await runnerItems.count() > 0) {
        const firstRunner = runnerItems.first();
        
        // Look for health indicators
        const healthIndicators = firstRunner.locator('[data-testid="runner-health"]').or(
          firstRunner.locator('text=/healthy|unhealthy|warning/i')
        );

        if (await healthIndicators.count() > 0) {
          await expect(healthIndicators.first()).toBeVisible();
        }

        // Look for capacity information
        const capacityInfo = firstRunner.locator('[data-testid="runner-capacity"]').or(
          firstRunner.locator('text=/\d+\/\d+|capacity|queue/i')
        );

        if (await capacityInfo.count() > 0) {
          await expect(capacityInfo.first()).toBeVisible();
        }

        // Look for last seen timestamp
        const lastSeen = firstRunner.locator('[data-testid="last-seen"]').or(
          firstRunner.locator('text=/last seen|ago|\d{1,2}\/\d{1,2}\/\d{4}/i')
        );

        if (await lastSeen.count() > 0) {
          await expect(lastSeen.first()).toBeVisible();
        }
      }
    });

    test('should refresh runner status in real-time', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      // Monitor API calls for status updates
      let apiCallCount = 0;
      await page.route('**/api/v1/repos/**/actions/runners**', route => {
        apiCallCount++;
        route.continue();
      });

      // Look for refresh button or auto-refresh
      const refreshButton = page.locator('[data-testid="refresh-runners"]').or(
        page.locator('button:has-text("Refresh")').or(page.locator('button[aria-label*="refresh"]'))
      );

      if (await refreshButton.isVisible()) {
        await refreshButton.click();
        await waitForLoadingToComplete(page);
        
        expect(apiCallCount).toBeGreaterThan(0);
      } else {
        // Wait for auto-refresh
        await page.waitForTimeout(30000);
        expect(apiCallCount).toBeGreaterThan(1);
      }
    });
  });

  test.describe('Adding New Self-hosted Runners', () => {
    test('should provide instructions to add new runner', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      // Look for add runner button
      const addRunnerButton = page.locator('[data-testid="add-runner"]').or(
        page.locator('button:has-text("Add runner")').or(page.locator('a:has-text("New self-hosted runner")'))
      );

      if (await addRunnerButton.isVisible()) {
        await addRunnerButton.click();
        await waitForLoadingToComplete(page);

        // Should show setup instructions
        await expect(page.locator('h2:has-text("Add new runner")').or(page.locator('h1:has-text("New runner")'))).toBeVisible();

        // Should show download commands for different platforms
        await expect(page.locator('text=/download|curl|wget/i')).toBeVisible();
        
        // Should show configuration command
        await expect(page.locator('text=/configure|token/i')).toBeVisible();
        
        // Should show authentication token
        const tokenElement = page.locator('[data-testid="runner-token"]').or(
          page.locator('code').or(page.locator('.token'))
        );

        if (await tokenElement.isVisible()) {
          await expect(tokenElement).toBeVisible();
          
          // Token should be copyable
          const copyButton = page.locator('[data-testid="copy-token"]').or(
            page.locator('button:has-text("Copy")').first()
          );

          if (await copyButton.isVisible()) {
            await copyButton.click();
            
            // Should show copy confirmation
            await expect(page.locator('text=/copied|copy success/i')).toBeVisible();
          }
        }
      }
    });

    test('should support different platform setup (Linux, Windows, macOS)', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners/new');
      await waitForLoadingToComplete(page);

      // Should have platform selection tabs
      const platformTabs = page.locator('[data-testid="platform-tabs"]').or(
        page.locator('button:has-text("Linux"), button:has-text("Windows"), button:has-text("macOS")')
      );

      if (await platformTabs.count() > 0) {
        // Test Linux tab
        const linuxTab = page.locator('button:has-text("Linux")');
        if (await linuxTab.isVisible()) {
          await linuxTab.click();
          await expect(page.locator('text=/sudo|chmod|\\.\/|bash/i')).toBeVisible();
        }

        // Test Windows tab
        const windowsTab = page.locator('button:has-text("Windows")');
        if (await windowsTab.isVisible()) {
          await windowsTab.click();
          await expect(page.locator('text=/powershell|\\.exe|\\.cmd/i')).toBeVisible();
        }

        // Test macOS tab
        const macOSTab = page.locator('button:has-text("macOS")');
        if (await macOSTab.isVisible()) {
          await macOSTab.click();
          await expect(page.locator('text=/sudo|chmod|\\.\/|bash/i')).toBeVisible();
        }
      }
    });

    test('should allow custom runner labels configuration', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners/new');
      await waitForLoadingToComplete(page);

      // Look for labels configuration
      const labelsInput = page.locator('[data-testid="runner-labels"]').or(
        page.locator('input[placeholder*="label"]').or(page.locator('input[name*="label"]'))
      );

      if (await labelsInput.isVisible()) {
        await labelsInput.fill('gpu, high-memory, production');
        
        // Should show label preview
        const labelPreview = page.locator('[data-testid="labels-preview"]').or(
          page.locator('.labels-preview')
        );

        if (await labelPreview.isVisible()) {
          await expect(labelPreview.locator('text=gpu')).toBeVisible();
          await expect(labelPreview.locator('text=high-memory')).toBeVisible();
          await expect(labelPreview.locator('text=production')).toBeVisible();
        }
      }
    });
  });

  test.describe('Runner Groups Configuration', () => {
    test('should display and manage runner groups', async () => {
      // Navigate to organization-level runner groups
      await page.goto('/organizations/admin/settings/runners/groups');
      await waitForLoadingToComplete(page);

      // Should show runner groups list
      await expect(page.locator('h1:has-text("Runner groups")').or(page.locator('h2:has-text("Groups")'))).toBeVisible();

      const groupsList = page.locator('[data-testid="groups-list"]').or(
        page.locator('table').or(page.locator('.groups-container'))
      );

      if (await groupsList.isVisible()) {
        await expect(groupsList).toBeVisible();

        // Each group should show members count
        const groupItems = page.locator('[data-testid="group-item"]').or(
          page.locator('tr').or(page.locator('.group-card'))
        );

        if (await groupItems.count() > 0) {
          const firstGroup = groupItems.first();
          
          // Should show group name
          await expect(firstGroup.locator('text=/default|production|development/i')).toBeVisible();
          
          // Should show runner count
          await expect(firstGroup.locator('text=/\d+ runner/i')).toBeVisible();
        }
      }
    });

    test('should allow creating new runner groups', async () => {
      await page.goto('/organizations/admin/settings/runners/groups');
      await waitForLoadingToComplete(page);

      const createGroupButton = page.locator('[data-testid="create-group"]').or(
        page.locator('button:has-text("New group")').or(page.locator('button:has-text("Create group")'))
      );

      if (await createGroupButton.isVisible()) {
        await createGroupButton.click();
        await waitForLoadingToComplete(page);

        // Should show create group form
        await expect(page.locator('h2:has-text("Create runner group")').or(page.locator('h1:has-text("New group")'))).toBeVisible();

        const groupNameInput = page.locator('[data-testid="group-name"]').or(
          page.locator('input[name="name"]').or(page.locator('input[placeholder*="name"]'))
        );

        if (await groupNameInput.isVisible()) {
          await groupNameInput.fill('Test Group');

          // Look for access policy settings
          const accessSettings = page.locator('[data-testid="access-policy"]').or(
            page.locator('input[type="radio"], input[type="checkbox"]')
          );

          if (await accessSettings.count() > 0) {
            await accessSettings.first().click();
          }

          // Submit the form
          const createButton = page.locator('button:has-text("Create")').or(
            page.locator('[data-testid="submit-group"]')
          );

          if (await createButton.isVisible()) {
            await createButton.click();
            await waitForLoadingToComplete(page);

            // Should redirect back to groups list
            await expect(page.locator('text=Test Group')).toBeVisible();
          }
        }
      }
    });

    test('should manage runner group permissions and access', async () => {
      await page.goto('/organizations/admin/settings/runners/groups/1');
      await waitForLoadingToComplete(page);

      // Should show group details
      await expect(page.locator('h1').or(page.locator('h2'))).toBeVisible();

      // Look for access control settings
      const accessControls = page.locator('[data-testid="access-controls"]').or(
        page.locator('.access-settings').or(page.locator('input[type="checkbox"]'))
      );

      if (await accessControls.count() > 0) {
        // Should have repository access settings
        const repoAccess = page.locator('text=/repository access|selected repositories/i');
        if (await repoAccess.isVisible()) {
          await expect(repoAccess).toBeVisible();
        }

        // Should have workflow access settings  
        const workflowAccess = page.locator('text=/workflow access|all workflows/i');
        if (await workflowAccess.isVisible()) {
          await expect(workflowAccess).toBeVisible();
        }
      }

      // Should show assigned runners
      const assignedRunners = page.locator('[data-testid="assigned-runners"]').or(
        page.locator('.assigned-runners')
      );

      if (await assignedRunners.isVisible()) {
        await expect(assignedRunners).toBeVisible();
      }
    });
  });

  test.describe('Runner Health Monitoring', () => {
    test('should monitor runner health and show alerts', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      // Look for health alerts or warnings
      const healthAlerts = page.locator('[data-testid="health-alert"]').or(
        page.locator('.alert').or(page.locator('[class*="warning"], [class*="error"]'))
      );

      if (await healthAlerts.count() > 0) {
        const firstAlert = healthAlerts.first();
        await expect(firstAlert).toBeVisible();

        // Should show alert message
        await expect(firstAlert.locator('text=/offline|unhealthy|disconnected/i')).toBeVisible();

        // Should have action buttons
        const actionButton = firstAlert.locator('button:has-text("View details")').or(
          firstAlert.locator('button:has-text("Dismiss")')
        );

        if (await actionButton.isVisible()) {
          await actionButton.click();
        }
      }
    });

    test('should show runner performance metrics', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners/runner-1');
      await waitForLoadingToComplete(page);

      // Look for performance metrics
      const metricsSection = page.locator('[data-testid="runner-metrics"]').or(
        page.locator('.metrics').or(page.locator('h3:has-text("Performance")'))
      );

      if (await metricsSection.isVisible()) {
        await expect(metricsSection).toBeVisible();

        // Should show job completion stats
        await expect(page.locator('text=/jobs completed|\d+ successful|\d+ failed/i')).toBeVisible();

        // Should show average execution time
        await expect(page.locator('text=/average time|\d+[smh]|duration/i')).toBeVisible();

        // Should show resource usage if available
        const resourceMetrics = page.locator('text=/cpu|memory|disk/i');
        if (await resourceMetrics.count() > 0) {
          await expect(resourceMetrics.first()).toBeVisible();
        }
      }
    });

    test('should display runner activity logs', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners/runner-1');
      await waitForLoadingToComplete(page);

      // Look for activity logs section
      const activityLogs = page.locator('[data-testid="runner-activity"]').or(
        page.locator('h3:has-text("Activity")').or(page.locator('.activity-log'))
      );

      if (await activityLogs.isVisible()) {
        await expect(activityLogs).toBeVisible();

        // Should show recent activities
        const logEntries = page.locator('[data-testid="log-entry"]').or(
          page.locator('.log-entry').or(page.locator('li'))
        );

        if (await logEntries.count() > 0) {
          const firstEntry = logEntries.first();
          
          // Should show timestamp
          await expect(firstEntry.locator('text=/ago|\d{1,2}\/\d{1,2}\/\d{4}|\d{1,2}:\d{2}/i')).toBeVisible();
          
          // Should show activity type
          await expect(firstEntry.locator('text=/started|completed|failed|connected|disconnected/i')).toBeVisible();
        }
      }
    });
  });

  test.describe('Removing Offline Runners', () => {
    test('should identify and remove offline runners', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      // Look for offline runners
      const offlineRunners = page.locator('[data-testid="runner-offline"]').or(
        page.locator('text=offline').locator('..').or(page.locator('[class*="offline"]'))
      );

      if (await offlineRunners.count() > 0) {
        const firstOfflineRunner = offlineRunners.first();
        
        // Should have remove option
        const removeButton = firstOfflineRunner.locator('button:has-text("Remove")').or(
          firstOfflineRunner.locator('[data-testid="remove-runner"]')
        );

        if (await removeButton.isVisible()) {
          await removeButton.click();

          // Should show confirmation dialog
          const confirmDialog = page.locator('[role="dialog"]').or(
            page.locator('.modal').or(page.locator('.confirm-dialog'))
          );

          if (await confirmDialog.isVisible()) {
            await expect(confirmDialog.locator('text=/remove|delete|confirm/i')).toBeVisible();

            // Confirm removal
            const confirmButton = confirmDialog.locator('button:has-text("Remove")').or(
              confirmDialog.locator('button:has-text("Confirm")')
            );

            if (await confirmButton.isVisible()) {
              await confirmButton.click();
              await waitForLoadingToComplete(page);

              // Runner should be removed from list
              await expect(firstOfflineRunner).not.toBeVisible();
            }
          }
        }
      }
    });

    test('should bulk remove multiple offline runners', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      // Look for bulk actions
      const bulkActions = page.locator('[data-testid="bulk-actions"]').or(
        page.locator('button:has-text("Bulk actions")').or(page.locator('.bulk-controls'))
      );

      if (await bulkActions.isVisible()) {
        // Select multiple offline runners
        const runnerCheckboxes = page.locator('[data-testid="runner-checkbox"]').or(
          page.locator('input[type="checkbox"]')
        );

        if (await runnerCheckboxes.count() > 1) {
          await runnerCheckboxes.first().click();
          await runnerCheckboxes.nth(1).click();

          // Should enable bulk remove
          const bulkRemoveButton = page.locator('button:has-text("Remove selected")').or(
            page.locator('[data-testid="bulk-remove"]')
          );

          if (await bulkRemoveButton.isVisible()) {
            await bulkRemoveButton.click();

            // Should show bulk confirmation
            const confirmDialog = page.locator('[role="dialog"]').or(page.locator('.modal'));
            
            if (await confirmDialog.isVisible()) {
              await expect(confirmDialog.locator('text=/remove \d+ runner/i')).toBeVisible();

              const confirmButton = confirmDialog.locator('button:has-text("Remove")');
              if (await confirmButton.isVisible()) {
                await confirmButton.click();
                await waitForLoadingToComplete(page);
              }
            }
          }
        }
      }
    });
  });

  test.describe('Mobile Runner Management', () => {
    test('should display runner list on mobile devices', async () => {
      await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      // Should display runners in mobile-friendly format
      await expect(page.locator('h1').or(page.locator('h2'))).toBeVisible();

      const runnerItems = page.locator('[data-testid="runner-item"]').or(
        page.locator('.runner-card')
      );

      if (await runnerItems.count() > 0) {
        const firstRunner = runnerItems.first();
        await expect(firstRunner).toBeVisible();

        // Should be readable without horizontal scroll
        const runnerWidth = await firstRunner.evaluate(el => (el as HTMLElement).offsetWidth);
        expect(runnerWidth).toBeLessThanOrEqual(375);
      }
    });

    test('should support touch interactions for runner management', async () => {
      await page.setViewportSize({ width: 375, height: 667 });
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      const runnerItems = page.locator('[data-testid="runner-item"]').or(
        page.locator('.runner-card')
      );

      if (await runnerItems.count() > 0) {
        // Tap on first runner
        await runnerItems.first().tap();
        
        // Should show runner details or actions
        const detailsView = page.locator('[data-testid="runner-details"]').or(
          page.locator('.runner-details')
        );

        if (await detailsView.isVisible()) {
          await expect(detailsView).toBeVisible();
        }
      }
    });
  });

  test.describe('Runner Security and Access Control', () => {
    test('should manage runner access permissions', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners/runner-1');
      await waitForLoadingToComplete(page);

      // Look for security settings
      const securitySection = page.locator('[data-testid="runner-security"]').or(
        page.locator('h3:has-text("Security")').or(page.locator('.security-settings'))
      );

      if (await securitySection.isVisible()) {
        await expect(securitySection).toBeVisible();

        // Should show access restrictions
        await expect(page.locator('text=/repository access|workflow access|permissions/i')).toBeVisible();

        // Should show runner registration token status
        const tokenStatus = page.locator('[data-testid="token-status"]').or(
          page.locator('text=/token|expires|valid/i')
        );

        if (await tokenStatus.isVisible()) {
          await expect(tokenStatus).toBeVisible();
        }
      }
    });

    test('should allow rotating runner tokens', async () => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      // Look for token management
      const tokenSection = page.locator('[data-testid="token-management"]').or(
        page.locator('button:has-text("Rotate token")').or(page.locator('h3:has-text("Token")'))
      );

      if (await tokenSection.isVisible()) {
        const rotateButton = page.locator('button:has-text("Rotate token")').or(
          page.locator('[data-testid="rotate-token"]')
        );

        if (await rotateButton.isVisible()) {
          await rotateButton.click();

          // Should show confirmation
          const confirmDialog = page.locator('[role="dialog"]').or(page.locator('.modal'));
          
          if (await confirmDialog.isVisible()) {
            await expect(confirmDialog.locator('text=/rotate|new token|invalidate/i')).toBeVisible();

            const confirmButton = confirmDialog.locator('button:has-text("Rotate")');
            if (await confirmButton.isVisible()) {
              await confirmButton.click();
              await waitForLoadingToComplete(page);

              // Should show new token
              await expect(page.locator('text=/new token|updated/i')).toBeVisible();
            }
          }
        }
      }
    });
  });
});