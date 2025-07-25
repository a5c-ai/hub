import { test, expect } from '@playwright/test';
import { loginUser, testUser } from './helpers/test-utils';

/**
 * E2E tests for Pull Request Review and Collaboration workflow
 * 
 * Tests cover:
 * - Review workflow (approve, request changes, comment)
 * - Merge pull request (different merge types: merge, squash, rebase)
 * - Close pull request without merging
 * - Reopen closed pull request
 * - Convert draft to ready for review
 */

test.describe('Pull Request Review Workflow', () => {
  test.beforeEach(async ({ page }) => {
    // Mock pull request details
    await page.route('**/api/repositories/*/pulls/*', async route => {
      const url = route.request().url();
      const prNumber = url.split('/').pop() || '1';
      
      if (route.request().method() === 'GET') {
        await route.fulfill({
          json: {
            id: parseInt(prNumber),
            issue: {
              number: parseInt(prNumber),
              title: 'Feature: Add user authentication',
              state: 'open',
              body: 'This PR implements user authentication with JWT tokens.',
              user: { 
                username: 'developer', 
                avatar_url: 'https://example.com/dev.jpg',
                name: 'Developer User'
              },
              comments_count: 2,
              created_at: '2024-01-15T10:00:00Z',
              updated_at: '2024-01-15T15:30:00Z'
            },
            head_ref: 'feature/auth',
            base_ref: 'main',
            merged: false,
            mergeable: true,
            draft: prNumber === '2', // PR #2 is draft
            additions: 156,
            deletions: 23,
            changed_files: 7,
            commits: 5,
            created_at: '2024-01-15T10:00:00Z',
            updated_at: '2024-01-15T15:30:00Z'
          }
        });
      }
    });

    // Mock reviews list
    await page.route('**/api/repositories/*/pulls/*/reviews', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          json: [
            {
              id: 1,
              user: { 
                username: 'reviewer1', 
                avatar_url: 'https://example.com/reviewer1.jpg',
                name: 'First Reviewer'
              },
              state: 'APPROVED',
              submitted_at: '2024-01-15T12:00:00Z',
              body: 'Looks good to me! Great implementation.'
            },
            {
              id: 2,
              user: { 
                username: 'reviewer2', 
                avatar_url: 'https://example.com/reviewer2.jpg',
                name: 'Second Reviewer'
              },
              state: 'CHANGES_REQUESTED',
              submitted_at: '2024-01-15T14:00:00Z',
              body: 'Please add more unit tests before merging.'
            }
          ]
        });
      } else if (route.request().method() === 'POST') {
        const body = await route.request().postDataJSON();
        await route.fulfill({
          json: {
            id: 3,
            user: { 
              username: 'testuser', 
              avatar_url: 'https://example.com/testuser.jpg',
              name: 'Test User'
            },
            state: body.event,
            submitted_at: new Date().toISOString(),
            body: body.body
          }
        });
      }
    });

    // Mock merge endpoint
    await page.route('**/api/repositories/*/pulls/*/merge', async route => {
      const body = await route.request().postDataJSON();
      await route.fulfill({
        json: {
          sha: 'abc123def456',
          merged: true,
          message: `Successfully merged via ${body.merge_method}`
        }
      });
    });

    // Mock status checks
    await page.route('**/api/repositories/*/pulls/*/status', async route => {
      await route.fulfill({
        json: {
          state: 'success',
          total_count: 3,
          statuses: [
            {
              state: 'success',
              context: 'ci/tests',
              description: 'All tests passed',
              target_url: 'https://ci.example.com/builds/123'
            },
            {
              state: 'success',
              context: 'ci/lint',
              description: 'Linting passed',
              target_url: 'https://ci.example.com/builds/124'
            },
            {
              state: 'success',
              context: 'ci/security',
              description: 'Security scan passed',
              target_url: 'https://ci.example.com/builds/125'
            }
          ]
        }
      });
    });

    await loginUser(page);
  });

  test('displays pull request review interface', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Check PR details are displayed
    await expect(page.locator('[data-testid="pr-title"]')).toContainText('Feature: Add user authentication');
    await expect(page.locator('[data-testid="pr-author"]')).toContainText('developer');
    await expect(page.locator('[data-testid="pr-branch-info"]')).toContainText('feature/auth → main');
    await expect(page.locator('[data-testid="pr-state-badge"]')).toContainText('Open');
    
    // Check review section exists
    await expect(page.locator('[data-testid="review-section"]')).toBeVisible();
    await expect(page.locator('[data-testid="review-form"]')).toBeVisible();
    
    // Check existing reviews are displayed
    await expect(page.locator('[data-testid="review-item"]')).toHaveCount(2);
    
    const approvedReview = page.locator('[data-testid="review-item"]').first();
    await expect(approvedReview.locator('[data-testid="review-state"]')).toContainText('Approved');
    await expect(approvedReview.locator('[data-testid="reviewer-name"]')).toContainText('reviewer1');
    
    const changesRequestedReview = page.locator('[data-testid="review-item"]').nth(1);
    await expect(changesRequestedReview.locator('[data-testid="review-state"]')).toContainText('Changes Requested');
    await expect(changesRequestedReview.locator('[data-testid="reviewer-name"]')).toContainText('reviewer2');
  });

  test('submit approval review', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Fill in review form
    await page.fill('[data-testid="review-comment-textarea"]', 'Great work! The implementation looks solid.');
    await page.check('[data-testid="review-approve-radio"]');
    
    // Submit review
    await page.click('[data-testid="submit-review-button"]');
    
    // Should show success message
    await expect(page.locator('[data-testid="review-success-message"]')).toContainText('Review submitted successfully');
    
    // Should add new review to the list
    await expect(page.locator('[data-testid="review-item"]')).toHaveCount(3);
  });

  test('submit changes requested review', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Fill in review form
    await page.fill('[data-testid="review-comment-textarea"]', 'Please address the security concerns in the auth module.');
    await page.check('[data-testid="review-changes-radio"]');
    
    // Submit review
    await page.click('[data-testid="submit-review-button"]');
    
    // Should show success message
    await expect(page.locator('[data-testid="review-success-message"]')).toContainText('Review submitted successfully');
    
    // Should add new review to the list
    await expect(page.locator('[data-testid="review-item"]')).toHaveCount(3);
    const newReview = page.locator('[data-testid="review-item"]').nth(2);
    await expect(newReview.locator('[data-testid="review-state"]')).toContainText('Changes Requested');
  });

  test('submit comment-only review', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Fill in review form
    await page.fill('[data-testid="review-comment-textarea"]', 'Thanks for the PR! I have some questions about the implementation.');
    await page.check('[data-testid="review-comment-radio"]');
    
    // Submit review
    await page.click('[data-testid="submit-review-button"]');
    
    // Should show success message
    await expect(page.locator('[data-testid="review-success-message"]')).toContainText('Review submitted successfully');
  });

  test('validates review submission', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Try to submit without selecting review type
    await page.fill('[data-testid="review-comment-textarea"]', 'Some comment');
    await page.click('[data-testid="submit-review-button"]');
    
    // Should show validation error
    await expect(page.locator('[data-testid="review-type-error"]')).toContainText('Please select a review type');
    
    // Try to submit changes requested without comment
    await page.check('[data-testid="review-changes-radio"]');
    await page.fill('[data-testid="review-comment-textarea"]', '');
    await page.click('[data-testid="submit-review-button"]');
    
    // Should show validation error
    await expect(page.locator('[data-testid="review-comment-error"]')).toContainText('Comment is required when requesting changes');
  });

  test('displays status checks', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Check status checks section
    await expect(page.locator('[data-testid="status-checks-section"]')).toBeVisible();
    await expect(page.locator('[data-testid="status-check-item"]')).toHaveCount(3);
    
    // Check individual status checks
    const testCheck = page.locator('[data-testid="status-check-item"]').first();
    await expect(testCheck.locator('[data-testid="status-check-name"]')).toContainText('ci/tests');
    await expect(testCheck.locator('[data-testid="status-check-state"]')).toContainText('success');
    await expect(testCheck.locator('[data-testid="status-check-description"]')).toContainText('All tests passed');
    
    // Should show overall status as passing
    await expect(page.locator('[data-testid="overall-status"]')).toContainText('All checks have passed');
  });
});

test.describe('Pull Request Merge Workflow', () => {
  test.beforeEach(async ({ page }) => {
    // Mock approved PR
    await page.route('**/api/repositories/*/pulls/1', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          json: {
            id: 1,
            issue: {
              number: 1,
              title: 'Ready to merge PR',
              state: 'open',
              user: { username: 'developer' },
              comments_count: 0
            },
            head_ref: 'feature/ready',
            base_ref: 'main',
            merged: false,
            mergeable: true,
            draft: false,
            additions: 50,
            deletions: 10,
            changed_files: 3,
            commits: 2
          }
        });
      } else if (route.request().method() === 'PATCH') {
        const body = await route.request().postDataJSON();
        await route.fulfill({
          json: {
            ...body,
            issue: { ...body.issue, state: body.issue?.state || 'closed' }
          }
        });
      }
    });

    // Mock successful reviews
    await page.route('**/api/repositories/*/pulls/1/reviews', async route => {
      await route.fulfill({
        json: [
          {
            id: 1,
            user: { username: 'reviewer1' },
            state: 'APPROVED',
            submitted_at: '2024-01-15T12:00:00Z',
            body: 'LGTM!'
          }
        ]
      });
    });

    // Mock passing status checks
    await page.route('**/api/repositories/*/pulls/1/status', async route => {
      await route.fulfill({
        json: {
          state: 'success',
          total_count: 2,
          statuses: [
            { state: 'success', context: 'ci/tests' },
            { state: 'success', context: 'ci/lint' }
          ]
        }
      });
    });

    await loginUser(page);
  });

  test('merge pull request with merge commit', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Should show merge button when ready
    await expect(page.locator('[data-testid="merge-section"]')).toBeVisible();
    await expect(page.locator('[data-testid="merge-button"]')).toBeEnabled();
    
    // Select merge commit option
    await page.selectOption('[data-testid="merge-method-select"]', 'merge');
    
    // Click merge
    await page.click('[data-testid="merge-button"]');
    
    // Should show confirmation dialog
    await expect(page.locator('[data-testid="merge-confirmation-dialog"]')).toBeVisible();
    await expect(page.locator('[data-testid="merge-method-display"]')).toContainText('Create a merge commit');
    
    // Confirm merge
    await page.click('[data-testid="confirm-merge-button"]');
    
    // Should show success message
    await expect(page.locator('[data-testid="merge-success-message"]')).toContainText('Pull request merged successfully');
  });

  test('merge pull request with squash and merge', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Select squash and merge option
    await page.selectOption('[data-testid="merge-method-select"]', 'squash');
    
    // Click merge
    await page.click('[data-testid="merge-button"]');
    
    // Should show confirmation dialog with commit title/message fields
    await expect(page.locator('[data-testid="merge-confirmation-dialog"]')).toBeVisible();
    await expect(page.locator('[data-testid="commit-title-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="commit-message-textarea"]')).toBeVisible();
    
    // Customize commit message
    await page.fill('[data-testid="commit-title-input"]', 'feat: Add user authentication system');
    await page.fill('[data-testid="commit-message-textarea"]', 'Implements JWT-based authentication with proper security measures.');
    
    // Confirm merge
    await page.click('[data-testid="confirm-merge-button"]');
    
    // Should show success message
    await expect(page.locator('[data-testid="merge-success-message"]')).toContainText('Pull request merged successfully');
  });

  test('merge pull request with rebase and merge', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Select rebase and merge option
    await page.selectOption('[data-testid="merge-method-select"]', 'rebase');
    
    // Click merge
    await page.click('[data-testid="merge-button"]');
    
    // Should show confirmation dialog
    await expect(page.locator('[data-testid="merge-confirmation-dialog"]')).toBeVisible();
    await expect(page.locator('[data-testid="merge-method-display"]')).toContainText('Rebase and merge');
    
    // Confirm merge
    await page.click('[data-testid="confirm-merge-button"]');
    
    // Should show success message
    await expect(page.locator('[data-testid="merge-success-message"]')).toContainText('Pull request merged successfully');
  });

  test('close pull request without merging', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Click close button
    await page.click('[data-testid="close-pr-button"]');
    
    // Should show confirmation dialog
    await expect(page.locator('[data-testid="close-confirmation-dialog"]')).toBeVisible();
    await expect(page.locator('[data-testid="close-warning"]')).toContainText('This will close the pull request without merging');
    
    // Add close reason
    await page.fill('[data-testid="close-reason-textarea"]', 'No longer needed after discussion');
    
    // Confirm close
    await page.click('[data-testid="confirm-close-button"]');
    
    // Should show success message and update PR state
    await expect(page.locator('[data-testid="close-success-message"]')).toContainText('Pull request closed');
    await expect(page.locator('[data-testid="pr-state-badge"]')).toContainText('Closed');
  });

  test('reopen closed pull request', async ({ page }) => {
    // Mock closed PR
    await page.route('**/api/repositories/*/pulls/1', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          json: {
            id: 1,
            issue: {
              number: 1,
              title: 'Closed PR to reopen',
              state: 'closed',
              user: { username: 'developer' }
            },
            head_ref: 'feature/closed',
            base_ref: 'main',
            merged: false,
            mergeable: true,
            draft: false
          }
        });
      }
    });

    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Should show reopen button for closed PR
    await expect(page.locator('[data-testid="reopen-pr-button"]')).toBeVisible();
    
    // Click reopen
    await page.click('[data-testid="reopen-pr-button"]');
    
    // Should show success message and update PR state
    await expect(page.locator('[data-testid="reopen-success-message"]')).toContainText('Pull request reopened');
    await expect(page.locator('[data-testid="pr-state-badge"]')).toContainText('Open');
  });
});

test.describe('Draft Pull Request Workflow', () => {
  test.beforeEach(async ({ page }) => {
    // Mock draft PR
    await page.route('**/api/repositories/*/pulls/2', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          json: {
            id: 2,
            issue: {
              number: 2,
              title: 'Draft: Work in progress feature',
              state: 'open',
              user: { username: 'developer' }
            },
            head_ref: 'feature/wip',
            base_ref: 'main',
            merged: false,
            mergeable: true,
            draft: true,
            additions: 25,
            deletions: 5,
            changed_files: 2
          }
        });
      } else if (route.request().method() === 'PATCH') {
        const body = await route.request().postDataJSON();
        await route.fulfill({
          json: {
            ...body,
            draft: body.draft !== undefined ? body.draft : true
          }
        });
      }
    });

    await loginUser(page);
  });

  test('convert draft to ready for review', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/2');
    
    // Should show draft badge
    await expect(page.locator('[data-testid="pr-draft-badge"]')).toBeVisible();
    await expect(page.locator('[data-testid="pr-draft-badge"]')).toContainText('Draft');
    
    // Should show convert to ready button
    await expect(page.locator('[data-testid="ready-for-review-button"]')).toBeVisible();
    
    // Should not show merge button for draft
    await expect(page.locator('[data-testid="merge-button"]')).not.toBeVisible();
    
    // Click ready for review
    await page.click('[data-testid="ready-for-review-button"]');
    
    // Should show confirmation dialog
    await expect(page.locator('[data-testid="ready-confirmation-dialog"]')).toBeVisible();
    
    // Confirm conversion
    await page.click('[data-testid="confirm-ready-button"]');
    
    // Should show success message and update state
    await expect(page.locator('[data-testid="ready-success-message"]')).toContainText('Pull request is now ready for review');
    await expect(page.locator('[data-testid="pr-draft-badge"]')).not.toBeVisible();
    await expect(page.locator('[data-testid="merge-section"]')).toBeVisible();
  });

  test('convert ready PR back to draft', async ({ page }) => {
    // First mock as ready PR
    await page.route('**/api/repositories/*/pulls/2', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          json: {
            id: 2,
            issue: {
              number: 2,
              title: 'Ready PR to convert back',
              state: 'open',
              user: { username: 'developer' }
            },
            head_ref: 'feature/ready',
            base_ref: 'main',
            merged: false,
            mergeable: true,
            draft: false
          }
        });
      }
    });

    await page.goto('/repositories/testuser/testrepo/pull/2');
    
    // Should show convert to draft button
    await expect(page.locator('[data-testid="convert-to-draft-button"]')).toBeVisible();
    
    // Click convert to draft
    await page.click('[data-testid="convert-to-draft-button"]');
    
    // Should show confirmation dialog
    await expect(page.locator('[data-testid="draft-confirmation-dialog"]')).toBeVisible();
    
    // Confirm conversion
    await page.click('[data-testid="confirm-draft-button"]');
    
    // Should show success message and update state
    await expect(page.locator('[data-testid="draft-success-message"]')).toContainText('Pull request converted to draft');
    await expect(page.locator('[data-testid="pr-draft-badge"]')).toBeVisible();
  });
});

test.describe('Blocked Merge Scenarios', () => {
  test('shows merge blocked when conflicts exist', async ({ page }) => {
    // Mock PR with conflicts
    await page.route('**/api/repositories/*/pulls/3', async route => {
      await route.fulfill({
        json: {
          id: 3,
          issue: {
            number: 3,
            title: 'PR with conflicts',
            state: 'open',
            user: { username: 'developer' }
          },
          head_ref: 'feature/conflicts',
          base_ref: 'main',
          merged: false,
          mergeable: false,
          draft: false
        }
      });
    });

    await loginUser(page);
    await page.goto('/repositories/testuser/testrepo/pull/3');
    
    // Should show conflicts warning
    await expect(page.locator('[data-testid="merge-conflicts-warning"]')).toBeVisible();
    await expect(page.locator('[data-testid="merge-conflicts-warning"]')).toContainText('This pull request has conflicts');
    
    // Merge button should be disabled
    await expect(page.locator('[data-testid="merge-button"]')).toBeDisabled();
    
    // Should show resolve conflicts button
    await expect(page.locator('[data-testid="resolve-conflicts-button"]')).toBeVisible();
  });

  test('shows merge blocked when required status checks fail', async ({ page }) => {
    // Mock failing status checks
    await page.route('**/api/repositories/*/pulls/1/status', async route => {
      await route.fulfill({
        json: {
          state: 'failure',
          total_count: 2,
          statuses: [
            { 
              state: 'failure', 
              context: 'ci/tests',
              description: 'Tests failed',
              target_url: 'https://ci.example.com/builds/456'
            },
            { 
              state: 'success', 
              context: 'ci/lint',
              description: 'Linting passed'
            }
          ]
        }
      });
    });

    await loginUser(page);
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Should show failing status checks
    await expect(page.locator('[data-testid="status-checks-failed"]')).toBeVisible();
    await expect(page.locator('[data-testid="overall-status"]')).toContainText('Some checks have failed');
    
    // Merge button should be disabled
    await expect(page.locator('[data-testid="merge-button"]')).toBeDisabled();
  });

  test('shows merge blocked when required reviews missing', async ({ page }) => {
    // Mock PR without required reviews
    await page.route('**/api/repositories/*/pulls/1/reviews', async route => {
      await route.fulfill({
        json: [] // No reviews
      });
    });

    await loginUser(page);
    await page.goto('/repositories/testuser/testrepo/pull/1');
    
    // Should show review requirements
    await expect(page.locator('[data-testid="required-reviews-warning"]')).toBeVisible();
    await expect(page.locator('[data-testid="required-reviews-warning"]')).toContainText('This pull request requires at least 1 approving review');
    
    // Merge button should be disabled
    await expect(page.locator('[data-testid="merge-button"]')).toBeDisabled();
  });
});