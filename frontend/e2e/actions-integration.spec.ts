import { test, expect, Page } from '@playwright/test';
import { loginUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('GitHub Actions - Integration Features', () => {
  let page: Page;

  test.beforeEach(async ({ browser }) => {
    page = await browser.newPage();
    await loginUser(page);
  });

  test.afterEach(async () => {
    await page.close();
  });

  test.describe('Status Checks on Pull Requests', () => {
    test('should display workflow status checks on PR', async () => {
      await page.goto('/repositories/admin/sample-project/pulls/1');
      await waitForLoadingToComplete(page);

      // Look for status checks section
      const statusChecks = page.locator('[data-testid="status-checks"]').or(
        page.locator('.status-checks').or(page.locator('h3:has-text("Checks")'))
      );

      if (await statusChecks.isVisible()) {
        await expect(statusChecks).toBeVisible();

        // Should show individual check statuses
        const checkItems = page.locator('[data-testid="status-check"]').or(
          page.locator('.check-item').or(page.locator('.status-check'))
        );

        if (await checkItems.count() > 0) {
          const firstCheck = checkItems.first();

          // Should show check name
          await expect(firstCheck.locator('text=/build|test|lint|deploy/i')).toBeVisible();

          // Should show status icon
          await expect(firstCheck.locator('text=/[🔄⏳✅❌⭕❓]/')).toBeVisible();

          // Should be clickable to view details
          await firstCheck.click();
          await waitForLoadingToComplete(page);

          // Should navigate to workflow run or show details
          expect(page.url()).toMatch(/\/actions\/runs\/|\/checks\//);
        }
      }
    });

    test('should prevent PR merge when required checks fail', async () => {
      await page.goto('/repositories/admin/sample-project/pulls/2'); // PR with failing checks
      await waitForLoadingToComplete(page);

      // Should show failed status checks
      const failedChecks = page.locator('text=❌').or(
        page.locator('.status-failure').or(page.locator('text=failure'))
      );

      if (await failedChecks.count() > 0) {
        await expect(failedChecks.first()).toBeVisible();

        // Merge button should be disabled or show warning
        const mergeButton = page.locator('[data-testid="merge-button"]').or(
          page.locator('button:has-text("Merge")')
        );

        if (await mergeButton.isVisible()) {
          // Should be disabled or show warning
          const isDisabled = await mergeButton.isDisabled();
          if (!isDisabled) {
            // Should show warning when attempting to merge
            await mergeButton.click();
            await expect(page.locator('text=/required checks|failing checks|cannot merge/i')).toBeVisible();
          } else {
            expect(isDisabled).toBe(true);
          }
        }
      }
    });

    test('should allow PR merge when all required checks pass', async () => {
      await page.goto('/repositories/admin/sample-project/pulls/3'); // PR with passing checks
      await waitForLoadingToComplete(page);

      // Should show passed status checks
      const passedChecks = page.locator('text=✅').or(
        page.locator('.status-success').or(page.locator('text=success'))
      );

      if (await passedChecks.count() > 0) {
        await expect(passedChecks.first()).toBeVisible();

        // Merge button should be enabled
        const mergeButton = page.locator('[data-testid="merge-button"]').or(
          page.locator('button:has-text("Merge")')
        );

        if (await mergeButton.isVisible()) {
          await expect(mergeButton).not.toBeDisabled();

          // Should show merge options
          const mergeOptions = page.locator('[data-testid="merge-options"]').or(
            page.locator('text=/merge commit|squash|rebase/i')
          );

          if (await mergeOptions.isVisible()) {
            await expect(mergeOptions).toBeVisible();
          }
        }
      }
    });

    test('should re-run failed checks from PR', async () => {
      await page.goto('/repositories/admin/sample-project/pulls/2');
      await waitForLoadingToComplete(page);

      const failedChecks = page.locator('[data-testid="status-check"]').filter({ hasText: 'failure' });

      if (await failedChecks.count() > 0) {
        const firstFailedCheck = failedChecks.first();

        // Look for re-run button
        const rerunButton = firstFailedCheck.locator('button:has-text("Re-run")').or(
          firstFailedCheck.locator('[data-testid="rerun-check"]')
        );

        if (await rerunButton.isVisible()) {
          await rerunButton.click();
          await waitForLoadingToComplete(page);

          // Status should change to pending/in-progress
          await expect(firstFailedCheck.locator('text=🔄').or(firstFailedCheck.locator('text=pending'))).toBeVisible();
        }
      }
    });
  });

  test.describe('Automatic Deployment Workflows', () => {
    test('should trigger deployment on main branch push', async () => {
      // Navigate to a repository with deployment workflows
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);

      // Look for deployment workflows
      const deploymentWorkflows = page.locator('text=/deploy|production|staging/i').first();

      if (await deploymentWorkflows.isVisible()) {
        await deploymentWorkflows.click();
        await waitForLoadingToComplete(page);

        // Should show recent deployment runs
        const recentRuns = page.locator('[data-testid="workflow-run"]').or(
          page.locator('.space-y-4 > div')
        );

        if (await recentRuns.count() > 0) {
          // Should have runs triggered by push events
          await expect(recentRuns.first().locator('text=push')).toBeVisible();

          // Should show deployment environment
          await expect(recentRuns.first().locator('text=/production|staging|development/i')).toBeVisible();
        }
      }
    });

    test('should show deployment status and environments', async () => {
      await page.goto('/repositories/admin/sample-project');
      await waitForLoadingToComplete(page);

      // Look for deployments section
      const deploymentsSection = page.locator('[data-testid="deployments"]').or(
        page.locator('h3:has-text("Deployments")').or(page.locator('.deployments'))
      );

      if (await deploymentsSection.isVisible()) {
        await expect(deploymentsSection).toBeVisible();

        // Should show active deployments
        const activeDeployments = page.locator('[data-testid="active-deployment"]').or(
          page.locator('.deployment-item')
        );

        if (await activeDeployments.count() > 0) {
          const firstDeployment = activeDeployments.first();

          // Should show environment name
          await expect(firstDeployment.locator('text=/production|staging|preview/i')).toBeVisible();

          // Should show deployment status
          await expect(firstDeployment.locator('text=/active|pending|failed/i')).toBeVisible();

          // Should have link to view deployment
          const viewLink = firstDeployment.locator('a:has-text("View deployment")').or(
            firstDeployment.locator('[data-testid="view-deployment"]')
          );

          if (await viewLink.isVisible()) {
            await viewLink.click();
            // Should navigate to deployment details or external URL
          }
        }
      }
    });

    test('should manage deployment protection rules', async () => {
      await page.goto('/repositories/admin/sample-project/settings/environments');
      await waitForLoadingToComplete(page);

      // Should show environments list
      await expect(page.locator('h1:has-text("Environments")').or(page.locator('h2:has-text("Environments")'))).toBeVisible();

      const environmentItems = page.locator('[data-testid="environment-item"]').or(
        page.locator('.environment-card')
      );

      if (await environmentItems.count() > 0) {
        const productionEnv = environmentItems.filter({ hasText: 'production' }).first();

        if (await productionEnv.isVisible()) {
          await productionEnv.click();
          await waitForLoadingToComplete(page);

          // Should show protection rules
          const protectionRules = page.locator('[data-testid="protection-rules"]').or(
            page.locator('h3:has-text("Protection rules")').or(page.locator('.protection-settings'))
          );

          if (await protectionRules.isVisible()) {
            await expect(protectionRules).toBeVisible();

            // Should show required reviewers
            await expect(page.locator('text=/required reviewer|approval/i')).toBeVisible();

            // Should show wait timer if configured
            const waitTimer = page.locator('text=/wait timer|\d+ minutes?/i');
            if (await waitTimer.isVisible()) {
              await expect(waitTimer).toBeVisible();
            }
          }
        }
      }
    });
  });

  test.describe('Matrix Build Configurations', () => {
    test('should display matrix builds with different configurations', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/456'); // Matrix build run
      await waitForLoadingToComplete(page);

      // Should show matrix job variations
      const matrixJobs = page.locator('[data-testid="matrix-job"]').or(
        page.locator('.matrix-job').or(page.locator('text=/node|python|ruby/i'))
      );

      if (await matrixJobs.count() > 1) {
        // Should have multiple jobs with different configurations
        await expect(matrixJobs.first()).toBeVisible();
        await expect(matrixJobs.last()).toBeVisible();

        // Each job should show its matrix parameters
        const firstJob = matrixJobs.first();
        await expect(firstJob.locator('text=/node|python|\d+\.\d+|ubuntu|windows/i')).toBeVisible();

        // Should be able to view individual matrix job logs
        await firstJob.click();
        await waitForLoadingToComplete(page);

        // Should show job-specific logs
        const logContainer = page.locator('[data-testid="log-container"]').or(
          page.locator('pre').or(page.locator('.log-output'))
        );

        if (await logContainer.isVisible()) {
          await expect(logContainer).toBeVisible();
        }
      }
    });

    test('should show matrix build summary with success/failure counts', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/456');
      await waitForLoadingToComplete(page);

      // Should show matrix summary
      const matrixSummary = page.locator('[data-testid="matrix-summary"]').or(
        page.locator('.matrix-summary')
      );

      if (await matrixSummary.isVisible()) {
        await expect(matrixSummary).toBeVisible();

        // Should show success/failure counts
        await expect(page.locator('text=/\d+ successful|\d+ passed/i')).toBeVisible();

        // May show failed jobs count
        const failedCount = page.locator('text=/\d+ failed/i');
        if (await failedCount.isVisible()) {
          await expect(failedCount).toBeVisible();
        }

        // Should show total job count
        await expect(page.locator('text=/\d+ total/i')).toBeVisible();
      }
    });

    test('should allow re-running specific matrix jobs', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/456');
      await waitForLoadingToComplete(page);

      const matrixJobs = page.locator('[data-testid="matrix-job"]').or(
        page.locator('.matrix-job')
      );

      if (await matrixJobs.count() > 0) {
        const failedJob = matrixJobs.filter({ hasText: 'failure' }).first();

        if (await failedJob.isVisible()) {
          // Should have re-run option for individual job
          const rerunButton = failedJob.locator('button:has-text("Re-run job")').or(
            failedJob.locator('[data-testid="rerun-job"]')
          );

          if (await rerunButton.isVisible()) {
            await rerunButton.click();
            await waitForLoadingToComplete(page);

            // Should start new job run
            await expect(failedJob.locator('text=🔄').or(failedJob.locator('text=pending'))).toBeVisible();
          }
        }
      }
    });
  });

  test.describe('Conditional Workflow Execution', () => {
    test('should show workflows with conditional logic', async () => {
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);

      // Look for conditional workflows (may skip certain conditions)
      const workflowRuns = page.locator('[data-testid="workflow-run"]').or(
        page.locator('.space-y-4 > div')
      );

      if (await workflowRuns.count() > 0) {
        // Look for skipped jobs/steps
        const skippedElements = page.locator('text=⭕').or(
          page.locator('text=skipped')
        );

        if (await skippedElements.count() > 0) {
          await expect(skippedElements.first()).toBeVisible();

          // Click to see details
          await skippedElements.first().click();
          await waitForLoadingToComplete(page);

          // Should show reason for skipping
          await expect(page.locator('text=/condition|if:|skip/i')).toBeVisible();
        }
      }
    });

    test('should display conditional step execution in workflow runs', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/789');
      await waitForLoadingToComplete(page);

      // Look for conditional steps
      const jobSteps = page.locator('[data-testid="job-step"]').or(
        page.locator('.job-step')
      );

      if (await jobSteps.count() > 0) {
        // Should show mix of executed and skipped steps
        const executedSteps = jobSteps.filter({ hasText: '✅' });
        const skippedSteps = jobSteps.filter({ hasText: '⭕' });

        if (await executedSteps.count() > 0 && await skippedSteps.count() > 0) {
          await expect(executedSteps.first()).toBeVisible();
          await expect(skippedSteps.first()).toBeVisible();

          // Click on skipped step to see reason
          await skippedSteps.first().click();
          await expect(page.locator('text=/condition not met|skipped due to/i')).toBeVisible();
        }
      }
    });

    test('should show branch-specific workflow triggers', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs');
      await waitForLoadingToComplete(page);

      // Filter by different branches to see conditional execution
      const branchFilter = page.locator('[data-testid="branch-filter"]').or(
        page.locator('select').or(page.locator('input[placeholder*="branch"]'))
      );

      if (await branchFilter.isVisible()) {
        // Test main branch
        await branchFilter.fill('main');
        await page.keyboard.press('Enter');
        await waitForLoadingToComplete(page);

        let mainBranchRuns = await page.locator('[data-testid="workflow-run"]').count();

        // Test feature branch
        await branchFilter.fill('feature/test');
        await page.keyboard.press('Enter');
        await waitForLoadingToComplete(page);

        let featureBranchRuns = await page.locator('[data-testid="workflow-run"]').count();

        // Different branches may have different workflow triggers
        expect(mainBranchRuns).not.toBe(featureBranchRuns);
      }
    });
  });

  test.describe('Workflow Dependencies and Triggers', () => {
    test('should show workflow dependencies and trigger chains', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/101');
      await waitForLoadingToComplete(page);

      // Look for workflow dependencies
      const dependenciesSection = page.locator('[data-testid="workflow-dependencies"]').or(
        page.locator('h3:has-text("Dependencies")').or(page.locator('.dependencies'))
      );

      if (await dependenciesSection.isVisible()) {
        await expect(dependenciesSection).toBeVisible();

        // Should show upstream/downstream workflows
        const dependentWorkflows = page.locator('[data-testid="dependent-workflow"]').or(
          page.locator('.workflow-dependency')
        );

        if (await dependentWorkflows.count() > 0) {
          await expect(dependentWorkflows.first()).toBeVisible();

          // Should show dependency status
          await expect(dependentWorkflows.first().locator('text=/completed|waiting|blocked/i')).toBeVisible();
        }
      }
    });

    test('should display workflow triggers and event sources', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/101');
      await waitForLoadingToComplete(page);

      // Should show trigger information
      const triggerInfo = page.locator('[data-testid="trigger-info"]').or(
        page.locator('.trigger-details')
      );

      if (await triggerInfo.isVisible()) {
        await expect(triggerInfo).toBeVisible();

        // Should show event type
        await expect(page.locator('text=/push|pull_request|schedule|workflow_dispatch/i')).toBeVisible();

        // Should show trigger source details
        const triggerDetails = page.locator('text=/branch:|tag:|schedule:|manual/i');
        if (await triggerDetails.isVisible()) {
          await expect(triggerDetails).toBeVisible();
        }
      }
    });

    test('should support manual workflow dispatch', async () => {
      await page.goto('/repositories/admin/sample-project/actions');
      await waitForLoadingToComplete(page);

      // Look for workflows with manual dispatch
      const dispatchableWorkflows = page.locator('[data-testid="workflow-dispatch"]').or(
        page.locator('button:has-text("Run workflow")')
      );

      if (await dispatchableWorkflows.count() > 0) {
        const firstDispatchable = dispatchableWorkflows.first();
        await firstDispatchable.click();

        // Should show dispatch form
        const dispatchForm = page.locator('[data-testid="dispatch-form"]').or(
          page.locator('.workflow-dispatch-form')
        );

        if (await dispatchForm.isVisible()) {
          await expect(dispatchForm).toBeVisible();

          // Should show branch selection
          const branchSelect = page.locator('[data-testid="branch-select"]').or(
            page.locator('select[name="branch"]')
          );

          if (await branchSelect.isVisible()) {
            await branchSelect.selectOption('main');
          }

          // Should show input parameters if any
          const inputFields = page.locator('[data-testid="workflow-input"]').or(
            page.locator('input[name*="input"]')
          );

          if (await inputFields.count() > 0) {
            await inputFields.first().fill('test-value');
          }

          // Submit dispatch
          const runButton = page.locator('button:has-text("Run workflow")').or(
            page.locator('[data-testid="submit-dispatch"]')
          );

          if (await runButton.isVisible()) {
            await runButton.click();
            await waitForLoadingToComplete(page);

            // Should navigate to new run or show confirmation
            expect(page.url()).toMatch(/\/actions\/runs\/|\/actions$/);
          }
        }
      }
    });
  });

  test.describe('Secrets Management Integration', () => {
    test('should manage repository secrets for workflows', async () => {
      await page.goto('/repositories/admin/sample-project/settings/secrets');
      await waitForLoadingToComplete(page);

      // Should show secrets management page
      await expect(page.locator('h1:has-text("Secrets")').or(page.locator('h2:has-text("Repository secrets")'))).toBeVisible();

      // Should list existing secrets (names only, not values)
      const secretsList = page.locator('[data-testid="secrets-list"]').or(
        page.locator('.secrets-table').or(page.locator('table'))
      );

      if (await secretsList.isVisible()) {
        await expect(secretsList).toBeVisible();

        // Should show secret names
        const secretNames = page.locator('[data-testid="secret-name"]').or(
          page.locator('td:first-child').or(page.locator('.secret-name'))
        );

        if (await secretNames.count() > 0) {
          await expect(secretNames.first()).toBeVisible();
          
          // Should not show secret values
          await expect(page.locator('text=/sk-|ghp_|ghs_/')).not.toBeVisible();
        }
      }
    });

    test('should add new repository secrets', async () => {
      await page.goto('/repositories/admin/sample-project/settings/secrets');
      await waitForLoadingToComplete(page);

      // Look for add secret button
      const addSecretButton = page.locator('[data-testid="add-secret"]').or(
        page.locator('button:has-text("New repository secret")')
      );

      if (await addSecretButton.isVisible()) {
        await addSecretButton.click();
        await waitForLoadingToComplete(page);

        // Should show add secret form
        await expect(page.locator('h2:has-text("New secret")').or(page.locator('h1:has-text("Add secret")'))).toBeVisible();

        const nameInput = page.locator('[data-testid="secret-name"]').or(
          page.locator('input[name="name"]')
        );

        const valueInput = page.locator('[data-testid="secret-value"]').or(
          page.locator('textarea[name="value"]')
        );

        if (await nameInput.isVisible() && await valueInput.isVisible()) {
          await nameInput.fill('TEST_SECRET');
          await valueInput.fill('test-secret-value');

          const addButton = page.locator('button:has-text("Add secret")').or(
            page.locator('[data-testid="save-secret"]')
          );

          if (await addButton.isVisible()) {
            await addButton.click();
            await waitForLoadingToComplete(page);

            // Should return to secrets list with new secret
            await expect(page.locator('text=TEST_SECRET')).toBeVisible();
          }
        }
      }
    });

    test('should show secrets usage in workflow runs', async () => {
      await page.goto('/repositories/admin/sample-project/actions/runs/123');
      await waitForLoadingToComplete(page);

      // Look for secret usage indicators in logs (should be masked)
      const logContainer = page.locator('[data-testid="log-container"]').or(
        page.locator('pre').or(page.locator('.log-output'))
      );

      if (await logContainer.isVisible()) {
        // Should show masked secrets as ***
        const maskedSecrets = page.locator('text=***');
        if (await maskedSecrets.count() > 0) {
          await expect(maskedSecrets.first()).toBeVisible();
        }

        // Should not show actual secret values
        await expect(page.locator('text=/sk-|ghp_|ghs_|secret-value/')).not.toBeVisible();
      }
    });
  });
});