import { test, expect } from '@playwright/test';
import { loginUser, waitForLoadingToComplete, navigateToActions } from './helpers/test-utils';

test.describe('Advanced CI/CD Workflow Features', () => {
  test.beforeEach(async ({ page }) => {
    await loginUser(page);
    
    // Mock workflow and environment APIs
    await page.route('**/api/v1/repos/**/actions/workflows/**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 'workflow-1',
            name: 'CI/CD Pipeline',
            path: '.github/workflows/ci-cd.yml',
            state: 'active',
            badges: ['badge1.svg'],
            runs: []
          }
        })
      });
    });
  });

  test.describe('Matrix Build Configurations', () => {
    test('should display matrix build strategy', async ({ page }) => {
      await navigateToActions(page);
      
      // Navigate to specific workflow with matrix builds
      await page.click('[data-testid="workflow-ci-cd-pipeline"]');
      await waitForLoadingToComplete(page);

      // Mock matrix workflow run
      await page.route('**/api/v1/repos/**/actions/runs/**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: 'run-1',
              name: 'CI/CD Pipeline',
              status: 'completed',
              conclusion: 'success',
              workflowId: 'workflow-1',
              headSha: 'abc123',
              event: 'push',
              jobs: [
                {
                  id: 'job-1',
                  name: 'test (ubuntu-latest, node-18)',
                  status: 'completed',
                  conclusion: 'success',
                  matrix: { os: 'ubuntu-latest', node: '18' }
                },
                {
                  id: 'job-2',
                  name: 'test (ubuntu-latest, node-20)',
                  status: 'completed',
                  conclusion: 'success',
                  matrix: { os: 'ubuntu-latest', node: '20' }
                },
                {
                  id: 'job-3',
                  name: 'test (windows-latest, node-18)',
                  status: 'completed',
                  conclusion: 'failure',
                  matrix: { os: 'windows-latest', node: '18' }
                },
                {
                  id: 'job-4',
                  name: 'test (macos-latest, node-20)',
                  status: 'completed',
                  conclusion: 'success',
                  matrix: { os: 'macos-latest', node: '20' }
                }
              ]
            }
          })
        });
      });

      await page.click('[data-testid="workflow-run-1"]');
      await waitForLoadingToComplete(page);

      // Verify matrix build visualization
      await expect(page.locator('[data-testid="matrix-visualization"]')).toBeVisible();
      await expect(page.locator('text=Matrix Strategy')).toBeVisible();
      
      // Check matrix job statuses
      await expect(page.locator('[data-testid="matrix-job-1"]')).toHaveClass(/success/);
      await expect(page.locator('[data-testid="matrix-job-2"]')).toHaveClass(/success/);
      await expect(page.locator('[data-testid="matrix-job-3"]')).toHaveClass(/failure/);
      await expect(page.locator('[data-testid="matrix-job-4"]')).toHaveClass(/success/);

      // Verify matrix combinations are displayed
      await expect(page.locator('text=ubuntu-latest, node-18')).toBeVisible();
      await expect(page.locator('text=windows-latest, node-18')).toBeVisible();
      await expect(page.locator('text=macos-latest, node-20')).toBeVisible();
    });

    test('should show matrix job failure details', async ({ page }) => {
      await navigateToActions(page);
      await page.click('[data-testid="workflow-ci-cd-pipeline"]');
      await page.click('[data-testid="workflow-run-1"]');
      await waitForLoadingToComplete(page);

      // Click on failed matrix job
      await page.click('[data-testid="matrix-job-3"]');
      await waitForLoadingToComplete(page);

      // Verify job details
      await expect(page.locator('h2')).toContainText('test (windows-latest, node-18)');
      await expect(page.locator('[data-testid="job-status"]')).toContainText('Failed');
      
      // Check matrix context display
      await expect(page.locator('[data-testid="matrix-context"]')).toBeVisible();
      await expect(page.locator('text=OS: windows-latest')).toBeVisible();
      await expect(page.locator('text=Node: 18')).toBeVisible();

      // Verify failure logs
      await expect(page.locator('[data-testid="job-logs"]')).toBeVisible();
      await expect(page.locator('text=Error:')).toBeVisible();
    });

    test('should filter matrix builds', async ({ page }) => {
      await navigateToActions(page);
      await page.click('[data-testid="workflow-ci-cd-pipeline"]');
      await page.click('[data-testid="workflow-run-1"]');
      await waitForLoadingToComplete(page);

      // Filter by OS
      await page.selectOption('[data-testid="matrix-os-filter"]', 'ubuntu-latest');
      await waitForLoadingToComplete(page);

      // Should only show Ubuntu jobs
      await expect(page.locator('[data-testid="matrix-job-1"]')).toBeVisible();
      await expect(page.locator('[data-testid="matrix-job-2"]')).toBeVisible();
      await expect(page.locator('[data-testid="matrix-job-3"]')).not.toBeVisible();

      // Filter by status
      await page.selectOption('[data-testid="matrix-status-filter"]', 'failure');
      await waitForLoadingToComplete(page);

      // Reset filters
      await page.click('[data-testid="reset-matrix-filters"]');
      await expect(page.locator('[data-testid="matrix-job"]')).toHaveCount(4);
    });
  });

  test.describe('Conditional Workflows', () => {
    test('should display conditional workflow execution', async ({ page }) => {
      await navigateToActions(page);
      
      // Mock conditional workflow
      await page.route('**/api/v1/repos/**/actions/workflows/conditional.yml', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: 'workflow-2',
              name: 'Conditional Deployment',
              conditions: [
                { if: "github.ref == 'refs/heads/main'", description: 'Deploy to production' },
                { if: "startsWith(github.ref, 'refs/heads/feature/')", description: 'Deploy to staging' },
                { if: "github.event_name == 'pull_request'", description: 'Run tests only' }
              ]
            }
          })
        });
      });

      await page.click('[data-testid="workflow-conditional-deployment"]');
      await waitForLoadingToComplete(page);

      // Verify conditional workflow display
      await expect(page.locator('text=Conditional Deployment')).toBeVisible();
      await expect(page.locator('[data-testid="workflow-conditions"]')).toBeVisible();
      
      // Check condition descriptions
      await expect(page.locator('text=Deploy to production')).toBeVisible();
      await expect(page.locator('text=Deploy to staging')).toBeVisible();
      await expect(page.locator('text=Run tests only')).toBeVisible();

      // Verify condition syntax
      await expect(page.locator("text=github.ref == 'refs/heads/main'")).toBeVisible();
    });

    test('should show workflow execution based on conditions', async ({ page }) => {
      await navigateToActions(page);
      await page.click('[data-testid="workflow-conditional-deployment"]');
      
      // Mock conditional run based on main branch
      await page.route('**/api/v1/repos/**/actions/runs/conditional-1', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: 'conditional-1',
              event: 'push',
              headBranch: 'main',
              jobs: [
                {
                  id: 'job-test',
                  name: 'Test',
                  status: 'completed',
                  conclusion: 'success',
                  if: 'always()'
                },
                {
                  id: 'job-deploy-prod',
                  name: 'Deploy to Production',
                  status: 'completed',
                  conclusion: 'success',
                  if: "github.ref == 'refs/heads/main'"
                },
                {
                  id: 'job-deploy-staging',
                  name: 'Deploy to Staging',
                  status: 'skipped',
                  conclusion: 'skipped',
                  if: "startsWith(github.ref, 'refs/heads/feature/')"
                }
              ]
            }
          })
        });
      });

      await page.click('[data-testid="run-conditional-1"]');
      await waitForLoadingToComplete(page);

      // Verify conditional execution
      await expect(page.locator('[data-testid="job-test"]')).toHaveClass(/success/);
      await expect(page.locator('[data-testid="job-deploy-prod"]')).toHaveClass(/success/);
      await expect(page.locator('[data-testid="job-deploy-staging"]')).toHaveClass(/skipped/);

      // Check condition evaluation display
      await expect(page.locator('[data-testid="condition-evaluation"]')).toBeVisible();
      await expect(page.locator('text=Branch: main → Deploy to Production')).toBeVisible();
    });
  });

  test.describe('Environment Management', () => {
    test('should display deployment environments', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/environments');
      await waitForLoadingToComplete(page);

      // Mock environments API
      await page.route('**/api/v1/repos/**/environments', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'env-1',
                name: 'production',
                url: 'https://app.example.com',
                description: 'Production environment',
                protectionRules: {
                  requiredReviewers: ['admin', 'devops-team'],
                  waitTimer: 0,
                  branches: ['main']
                },
                deployments: [
                  {
                    id: 'deploy-1',
                    sha: 'abc123',
                    ref: 'main',
                    status: 'success',
                    createdAt: '2024-01-15T10:30:00Z'
                  }
                ]
              },
              {
                id: 'env-2',
                name: 'staging',
                url: 'https://staging.example.com',
                description: 'Staging environment',
                protectionRules: {
                  requiredReviewers: [],
                  waitTimer: 5,
                  branches: ['main', 'develop']
                },
                deployments: []
              }
            ]
          })
        });
      });

      // Verify environments display
      await expect(page.locator('h1')).toContainText('Environments');
      await expect(page.locator('[data-testid="environment-item"]')).toHaveCount(2);
      
      // Check environment details
      await expect(page.locator('text=production')).toBeVisible();
      await expect(page.locator('text=https://app.example.com')).toBeVisible();
      await expect(page.locator('text=staging')).toBeVisible();
      
      // Verify protection rules indicators
      await expect(page.locator('[data-testid="protected-env-production"]')).toBeVisible();
      await expect(page.locator('text=Required reviewers')).toBeVisible();
    });

    test('should create new environment', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/environments');
      await waitForLoadingToComplete(page);

      // Click create environment
      await page.click('[data-testid="create-environment-button"]');
      await expect(page.locator('[data-testid="create-env-modal"]')).toBeVisible();

      // Fill environment details
      await page.fill('[data-testid="env-name"]', 'development');
      await page.fill('[data-testid="env-url"]', 'https://dev.example.com');
      await page.fill('[data-testid="env-description"]', 'Development environment');

      // Configure protection rules
      await page.check('[data-testid="enable-protection"]');
      await page.fill('[data-testid="required-reviewers"]', 'dev-team');
      await page.fill('[data-testid="wait-timer"]', '2');
      await page.fill('[data-testid="deployment-branches"]', 'develop,feature/*');

      // Mock environment creation
      await page.route('**/api/v1/repos/**/environments', async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: { id: 'env-3', name: 'development' }
            })
          });
        }
      });

      await page.click('[data-testid="create-environment"]');
      
      // Verify environment was created
      await expect(page.locator('text=Environment created successfully')).toBeVisible();
      await expect(page.locator('text=development')).toBeVisible();
    });

    test('should manage environment secrets', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/environments/production');
      await waitForLoadingToComplete(page);

      // Navigate to secrets tab
      await page.click('[data-testid="environment-secrets-tab"]');
      await waitForLoadingToComplete(page);

      // Mock secrets API
      await page.route('**/api/v1/repos/**/environments/production/secrets', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              { name: 'DATABASE_URL', updatedAt: '2024-01-15T10:30:00Z' },
              { name: 'API_KEY', updatedAt: '2024-01-14T09:15:00Z' }
            ]
          })
        });
      });

      // Verify secrets display
      await expect(page.locator('text=Environment Secrets')).toBeVisible();
      await expect(page.locator('[data-testid="secret-item"]')).toHaveCount(2);
      await expect(page.locator('text=DATABASE_URL')).toBeVisible();
      await expect(page.locator('text=API_KEY')).toBeVisible();

      // Add new secret
      await page.click('[data-testid="add-secret-button"]');
      await expect(page.locator('[data-testid="secret-modal"]')).toBeVisible();

      await page.fill('[data-testid="secret-name"]', 'NEW_SECRET');
      await page.fill('[data-testid="secret-value"]', 'secret-value-123');

      await page.route('**/api/v1/repos/**/environments/production/secrets/NEW_SECRET', async route => {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.click('[data-testid="save-secret"]');
      await expect(page.locator('text=Secret added successfully')).toBeVisible();
    });
  });

  test.describe('Deployment Workflows', () => {
    test('should display deployment history', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/deployments');
      await waitForLoadingToComplete(page);

      // Mock deployments API
      await page.route('**/api/v1/repos/**/deployments', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'deploy-1',
                environment: 'production',
                sha: 'abc123',
                ref: 'main',
                status: 'success',
                description: 'Deploy v2.1.0',
                creator: 'john-doe',
                createdAt: '2024-01-15T10:30:00Z',
                duration: 180
              },
              {
                id: 'deploy-2',
                environment: 'staging',
                sha: 'def456',
                ref: 'develop',
                status: 'in_progress',
                description: 'Deploy feature branch',
                creator: 'jane-smith',
                createdAt: '2024-01-15T11:00:00Z',
                duration: null
              },
              {
                id: 'deploy-3',
                environment: 'production',
                sha: 'ghi789',
                ref: 'main',
                status: 'failure',
                description: 'Deploy v2.0.9 (rollback)',
                creator: 'admin',
                createdAt: '2024-01-14T16:45:00Z',
                duration: 45
              }
            ]
          })
        });
      });

      // Verify deployment history
      await expect(page.locator('h1')).toContainText('Deployments');
      await expect(page.locator('[data-testid="deployment-item"]')).toHaveCount(3);
      
      // Check deployment statuses
      await expect(page.locator('[data-testid="deploy-1"]')).toHaveClass(/success/);
      await expect(page.locator('[data-testid="deploy-2"]')).toHaveClass(/in-progress/);
      await expect(page.locator('[data-testid="deploy-3"]')).toHaveClass(/failure/);

      // Verify deployment details
      await expect(page.locator('text=Deploy v2.1.0')).toBeVisible();
      await expect(page.locator('text=john-doe')).toBeVisible();
      await expect(page.locator('text=3m 0s')).toBeVisible(); // Duration
    });

    test('should trigger manual deployment', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/deployments');
      await waitForLoadingToComplete(page);

      // Click manual deploy button
      await page.click('[data-testid="manual-deploy-button"]');
      await expect(page.locator('[data-testid="deploy-modal"]')).toBeVisible();

      // Select deployment options
      await page.selectOption('[data-testid="deploy-environment"]', 'staging');
      await page.selectOption('[data-testid="deploy-ref"]', 'develop');
      await page.fill('[data-testid="deploy-description"]', 'Manual staging deployment');

      // Mock deployment trigger
      await page.route('**/api/v1/repos/**/deployments', async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: { id: 'deploy-4', status: 'pending' }
            })
          });
        }
      });

      await page.click('[data-testid="trigger-deployment"]');
      
      // Verify deployment started
      await expect(page.locator('text=Deployment triggered successfully')).toBeVisible();
      await expect(page.locator('[data-testid="deploy-4"]')).toBeVisible();
    });

    test('should handle deployment approvals', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/deployments/deploy-pending');
      await waitForLoadingToComplete(page);

      // Mock pending deployment requiring approval
      await page.route('**/api/v1/repos/**/deployments/deploy-pending', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: 'deploy-pending',
              environment: 'production',
              status: 'waiting',
              approvals: {
                required: 2,
                received: 1,
                reviewers: [
                  { user: 'admin', approved: true, approvedAt: '2024-01-15T10:30:00Z' },
                  { user: 'devops-team', approved: false, approvedAt: null }
                ]
              }
            }
          })
        });
      });

      // Verify approval status
      await expect(page.locator('text=Waiting for Approval')).toBeVisible();
      await expect(page.locator('[data-testid="approval-status"]')).toContainText('1 of 2 approvals');
      
      // Check reviewer statuses
      await expect(page.locator('[data-testid="reviewer-admin"]')).toHaveClass(/approved/);
      await expect(page.locator('[data-testid="reviewer-devops-team"]')).toHaveClass(/pending/);

      // Approve deployment (if user has permission)
      await page.route('**/api/v1/repos/**/deployments/deploy-pending/approve', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.click('[data-testid="approve-deployment"]');
      await expect(page.locator('text=Deployment approved')).toBeVisible();
    });
  });

  test.describe('Self-hosted Runner Scaling', () => {
    test('should display runner pools', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/settings/runners');
      await waitForLoadingToComplete(page);

      // Mock runner pools API
      await page.route('**/api/v1/repos/**/actions/runner-pools', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'pool-1',
                name: 'linux-pool',
                os: 'linux',
                minRunners: 2,
                maxRunners: 10,
                currentRunners: 4,
                activeJobs: 2,
                queuedJobs: 1,
                autoscaling: true
              },
              {
                id: 'pool-2',
                name: 'windows-pool',
                os: 'windows',
                minRunners: 1,
                maxRunners: 5,
                currentRunners: 2,
                activeJobs: 0,
                queuedJobs: 0,
                autoscaling: false
              }
            ]
          })
        });
      });

      // Verify runner pools display
      await expect(page.locator('text=Runner Pools')).toBeVisible();
      await expect(page.locator('[data-testid="runner-pool"]')).toHaveCount(2);
      
      // Check pool details
      await expect(page.locator('text=linux-pool')).toBeVisible();
      await expect(page.locator('text=4 / 10 runners')).toBeVisible();
      await expect(page.locator('text=2 active jobs')).toBeVisible();
      
      // Verify autoscaling indicators
      await expect(page.locator('[data-testid="autoscaling-enabled-pool-1"]')).toBeVisible();
      await expect(page.locator('[data-testid="autoscaling-disabled-pool-2"]')).toBeVisible();
    });

    test('should configure runner pool autoscaling', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/settings/runners/linux-pool');
      await waitForLoadingToComplete(page);

      // Click configure autoscaling
      await page.click('[data-testid="configure-autoscaling"]');
      await expect(page.locator('[data-testid="autoscaling-modal"]')).toBeVisible();

      // Configure scaling settings
      await page.fill('[data-testid="min-runners"]', '3');
      await page.fill('[data-testid="max-runners"]', '15');
      await page.fill('[data-testid="scale-up-threshold"]', '80');
      await page.fill('[data-testid="scale-down-threshold"]', '20');
      await page.fill('[data-testid="scale-up-factor"]', '2');
      await page.fill('[data-testid="scale-down-delay"]', '300');

      // Mock autoscaling configuration
      await page.route('**/api/v1/repos/**/actions/runner-pools/pool-1/autoscaling', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.click('[data-testid="save-autoscaling"]');
      await expect(page.locator('text=Autoscaling configured successfully')).toBeVisible();
    });

    test('should display runner metrics and performance', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/analytics/runners');
      await waitForLoadingToComplete(page);

      // Mock runner analytics API
      await page.route('**/api/v1/repos/**/actions/runner-analytics', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              utilization: {
                average: 65,
                peak: 95,
                idle: 35
              },
              performance: {
                averageJobTime: 180,
                queueTime: 15,
                throughput: 120
              },
              costs: {
                hourly: 5.40,
                daily: 129.60,
                monthly: 3888.00
              },
              trends: [
                { date: '2024-01-01', utilization: 60, jobs: 45 },
                { date: '2024-01-02', utilization: 70, jobs: 52 }
              ]
            }
          })
        });
      });

      // Verify runner analytics
      await expect(page.locator('text=Runner Analytics')).toBeVisible();
      await expect(page.locator('text=65% Average Utilization')).toBeVisible();
      await expect(page.locator('text=3m 0s Average Job Time')).toBeVisible();
      await expect(page.locator('text=$5.40/hour')).toBeVisible();

      // Check analytics charts
      await expect(page.locator('[data-testid="utilization-chart"]')).toBeVisible();
      await expect(page.locator('[data-testid="performance-chart"]')).toBeVisible();
      await expect(page.locator('[data-testid="cost-chart"]')).toBeVisible();
    });
  });

  test.describe('Artifact Lifecycle Management', () => {
    test('should display build artifacts', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/actions/artifacts');
      await waitForLoadingToComplete(page);

      // Mock artifacts API
      await page.route('**/api/v1/repos/**/actions/artifacts', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'artifact-1',
                name: 'build-output',
                size: 12485760,
                createdAt: '2024-01-15T10:30:00Z',
                expiresAt: '2024-04-15T10:30:00Z',
                workflowRun: { id: 'run-1', name: 'CI Pipeline' },
                downloadCount: 15
              },
              {
                id: 'artifact-2',
                name: 'test-results',
                size: 2048000,
                createdAt: '2024-01-15T09:15:00Z',
                expiresAt: '2024-04-15T09:15:00Z',
                workflowRun: { id: 'run-2', name: 'Test Suite' },
                downloadCount: 8
              },
              {
                id: 'artifact-3',
                name: 'docker-image',
                size: 524288000,
                createdAt: '2024-01-14T16:45:00Z',
                expiresAt: '2024-02-14T16:45:00Z',
                workflowRun: { id: 'run-3', name: 'Build Docker' },
                downloadCount: 3
              }
            ]
          })
        });
      });

      // Verify artifacts display
      await expect(page.locator('h1')).toContainText('Artifacts');
      await expect(page.locator('[data-testid="artifact-item"]')).toHaveCount(3);
      
      // Check artifact details
      await expect(page.locator('text=build-output')).toBeVisible();
      await expect(page.locator('text=11.9 MB')).toBeVisible(); // Formatted size
      await expect(page.locator('text=15 downloads')).toBeVisible();
      
      // Verify expiration dates
      await expect(page.locator('text=Expires in')).toHaveCount(3);
    });

    test('should configure artifact retention policies', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/settings/artifacts');
      await waitForLoadingToComplete(page);

      // Mock current retention settings
      await page.route('**/api/v1/repos/**/actions/artifact-settings', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              defaultRetention: 90,
              maxRetention: 400,
              retentionPolicies: [
                { pattern: 'build-*', retention: 180 },
                { pattern: 'test-*', retention: 30 },
                { pattern: 'release-*', retention: 365 }
              ]
            }
          })
        });
      });

      // Verify retention settings
      await expect(page.locator('text=Artifact Retention')).toBeVisible();
      await expect(page.locator('[data-testid="default-retention"]')).toHaveValue('90');
      
      // Check existing policies
      await expect(page.locator('[data-testid="policy-item"]')).toHaveCount(3);
      await expect(page.locator('text=build-*')).toBeVisible();
      await expect(page.locator('text=180 days')).toBeVisible();

      // Add new retention policy
      await page.click('[data-testid="add-policy-button"]');
      await expect(page.locator('[data-testid="policy-modal"]')).toBeVisible();

      await page.fill('[data-testid="policy-pattern"]', 'deploy-*');
      await page.fill('[data-testid="policy-retention"]', '30');

      await page.route('**/api/v1/repos/**/actions/artifact-settings', async route => {
        if (route.request().method() === 'PUT') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await page.click('[data-testid="save-policy"]');
      await expect(page.locator('text=Policy added successfully')).toBeVisible();
    });

    test('should manage artifact cleanup', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/actions/artifacts');
      await waitForLoadingToComplete(page);

      // Bulk select artifacts
      await page.check('[data-testid="select-artifact-2"]');
      await page.check('[data-testid="select-artifact-3"]');

      // Click bulk delete
      await page.click('[data-testid="bulk-delete-artifacts"]');
      await expect(page.locator('[data-testid="confirm-delete-modal"]')).toBeVisible();

      // Verify deletion summary
      await expect(page.locator('text=Delete 2 artifacts')).toBeVisible();
      await expect(page.locator('text=Total size: 500.2 MB')).toBeVisible();

      // Mock bulk deletion
      await page.route('**/api/v1/repos/**/actions/artifacts/bulk-delete', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true, deleted: 2 })
        });
      });

      await page.click('[data-testid="confirm-bulk-delete"]');
      await expect(page.locator('text=2 artifacts deleted successfully')).toBeVisible();
      
      // Verify artifacts are removed from list
      await expect(page.locator('[data-testid="artifact-item"]')).toHaveCount(1);
    });
  });

  test.describe('Performance Monitoring', () => {
    test('should display workflow performance metrics', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/analytics/workflows');
      await waitForLoadingToComplete(page);

      // Mock performance analytics
      await page.route('**/api/v1/repos/**/actions/performance', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              overview: {
                averageDuration: 450,
                successRate: 92.5,
                totalRuns: 1250,
                failureRate: 7.5
              },
              slowestWorkflows: [
                { name: 'Full Test Suite', averageDuration: 1200, runs: 45 },
                { name: 'Build & Deploy', averageDuration: 900, runs: 78 }
              ],
              bottlenecks: [
                { step: 'Install Dependencies', averageDuration: 180, percentage: 40 },
                { step: 'Run Tests', averageDuration: 120, percentage: 27 }
              ],
              trends: [
                { date: '2024-01-01', duration: 420, success: 95 },
                { date: '2024-01-02', duration: 480, success: 90 }
              ]
            }
          })
        });
      });

      // Verify performance overview
      await expect(page.locator('text=Workflow Performance')).toBeVisible();
      await expect(page.locator('text=7m 30s Average Duration')).toBeVisible();
      await expect(page.locator('text=92.5% Success Rate')).toBeVisible();
      await expect(page.locator('text=1,250 Total Runs')).toBeVisible();

      // Check performance breakdowns
      await expect(page.locator('text=Slowest Workflows')).toBeVisible();
      await expect(page.locator('text=Full Test Suite')).toBeVisible();
      await expect(page.locator('text=20m 0s')).toBeVisible();

      // Verify bottleneck analysis
      await expect(page.locator('text=Performance Bottlenecks')).toBeVisible();
      await expect(page.locator('text=Install Dependencies')).toBeVisible();
      await expect(page.locator('text=40%')).toBeVisible();

      // Check performance charts
      await expect(page.locator('[data-testid="duration-trend-chart"]')).toBeVisible();
      await expect(page.locator('[data-testid="success-rate-chart"]')).toBeVisible();
    });

    test('should identify performance optimization opportunities', async ({ page }) => {
      await page.goto('/repositories/admin/sample-project/analytics/optimization');
      await waitForLoadingToComplete(page);

      // Mock optimization recommendations
      await page.route('**/api/v1/repos/**/actions/optimization', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              recommendations: [
                {
                  type: 'caching',
                  title: 'Enable dependency caching',
                  description: 'Cache node_modules to reduce install time',
                  potentialSaving: 120,
                  impact: 'high',
                  workflow: 'CI Pipeline'
                },
                {
                  type: 'parallelization',
                  title: 'Parallelize test execution',
                  description: 'Run tests in parallel to reduce execution time',
                  potentialSaving: 240,
                  impact: 'high',
                  workflow: 'Test Suite'
                },
                {
                  type: 'resource',
                  title: 'Use larger runner',
                  description: 'Upgrade to 4-core runner for faster builds',
                  potentialSaving: 180,
                  impact: 'medium',
                  workflow: 'Build & Deploy'
                }
              ]
            }
          })
        });
      });

      // Verify optimization recommendations
      await expect(page.locator('text=Performance Optimization')).toBeVisible();
      await expect(page.locator('[data-testid="recommendation-item"]')).toHaveCount(3);
      
      // Check recommendation details
      await expect(page.locator('text=Enable dependency caching')).toBeVisible();
      await expect(page.locator('text=Save ~2m 0s')).toBeVisible();
      await expect(page.locator('[data-testid="impact-high"]')).toHaveCount(2);
      
      // Verify implementation suggestions
      await page.click('[data-testid="recommendation-caching"]');
      await expect(page.locator('[data-testid="implementation-guide"]')).toBeVisible();
      await expect(page.locator('text=Implementation Steps')).toBeVisible();
    });
  });
});