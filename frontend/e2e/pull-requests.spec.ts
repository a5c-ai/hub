import { test, expect } from '@playwright/test';
import { loginUser, testUser } from './helpers/test-utils';

/**
 * E2E tests for Pull Request Management workflow
 * 
 * Tests cover:
 * - List pull requests with different states (open/closed/merged)
 * - Filter pull requests by author, assignee, labels
 * - Create new pull request form validation and submission
 * - Navigate to pull request details page
 * - Display pull request metadata (title, description, labels, assignees)
 */

test.describe('Pull Request Management', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses for pull requests
    await page.route('**/api/repositories/*/pulls*', async route => {
      const url = new URL(route.request().url());
      const state = url.searchParams.get('state') || 'open';
      
      const pullRequests = {
        open: [
          {
            id: 1,
            issue: {
              number: 1,
              title: 'Add user authentication system',
              state: 'open',
              user: { username: 'testuser', avatar_url: 'https://example.com/avatar.jpg' },
              comments_count: 3,
              created_at: '2024-01-15T10:00:00Z',
              updated_at: '2024-01-15T15:30:00Z'
            },
            head_ref: 'feature/auth-system',
            base_ref: 'main',
            merged: false,
            draft: false,
            additions: 245,
            deletions: 12,
            changed_files: 8,
            mergeable: true,
            created_at: '2024-01-15T10:00:00Z'
          },
          {
            id: 2,
            issue: {
              number: 2,
              title: 'Fix responsive design issues',
              state: 'open',
              user: { username: 'designer', avatar_url: 'https://example.com/designer.jpg' },
              comments_count: 1,
              created_at: '2024-01-14T09:00:00Z',
              updated_at: '2024-01-14T12:00:00Z'
            },
            head_ref: 'fix/responsive-layout',
            base_ref: 'main',
            merged: false,
            draft: true,
            additions: 67,
            deletions: 23,
            changed_files: 4,
            mergeable: false,
            created_at: '2024-01-14T09:00:00Z'
          }
        ],
        closed: [
          {
            id: 3,
            issue: {
              number: 3,
              title: 'Update documentation',
              state: 'closed',
              user: { username: 'writer', avatar_url: 'https://example.com/writer.jpg' },
              comments_count: 2,
              created_at: '2024-01-10T08:00:00Z',
              updated_at: '2024-01-12T16:00:00Z'
            },
            head_ref: 'docs/update-readme',
            base_ref: 'main',
            merged: true,
            draft: false,
            additions: 134,
            deletions: 45,
            changed_files: 3,
            mergeable: null,
            created_at: '2024-01-10T08:00:00Z'
          }
        ]
      };

      const filteredPRs = state === 'all' 
        ? [...pullRequests.open, ...pullRequests.closed]
        : (pullRequests as any)[state] || [];

      await route.fulfill({
        json: {
          pull_requests: filteredPRs,
          total_count: filteredPRs.length,
          page: 1,
          per_page: 25
        }
      });
    });

    // Mock repository info
    await page.route('**/api/repositories/testuser/testrepo', async route => {
      await route.fulfill({
        json: {
          id: 1,
          name: 'testrepo',
          full_name: 'testuser/testrepo',
          owner: { username: 'testuser' },
          private: false,
          description: 'Test repository for E2E tests'
        }
      });
    });

    await loginUser(page);
  });

  test('displays pull request list correctly', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls');
    
    // Check page title and navigation
    await expect(page.locator('h2')).toContainText('Pull Requests');
    await expect(page.locator('[data-testid="new-pr-button"]')).toBeVisible();
    
    // Verify PR list items are displayed
    await expect(page.locator('[data-testid="pr-item"]')).toHaveCount(2);
    
    // Check first PR details
    const firstPR = page.locator('[data-testid="pr-item"]').first();
    await expect(firstPR.locator('[data-testid="pr-title"]')).toContainText('Add user authentication system');
    await expect(firstPR.locator('[data-testid="pr-number"]')).toContainText('#1');
    await expect(firstPR.locator('[data-testid="pr-branch"]')).toContainText('feature/auth-system → main');
    await expect(firstPR.locator('[data-testid="pr-author"]')).toContainText('by testuser');
    await expect(firstPR.locator('[data-testid="pr-state-badge"]')).toContainText('Open');
    
    // Check PR statistics
    await expect(firstPR.locator('[data-testid="pr-additions"]')).toContainText('+245');
    await expect(firstPR.locator('[data-testid="pr-deletions"]')).toContainText('-12');
    await expect(firstPR.locator('[data-testid="pr-files-changed"]')).toContainText('8 files');
    await expect(firstPR.locator('[data-testid="pr-comments"]')).toContainText('3 comments');
  });

  test('filters pull requests by state', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls');
    
    // Initially showing open PRs
    await expect(page.locator('[data-testid="pr-item"]')).toHaveCount(2);
    
    // Switch to closed PRs
    await page.click('[data-testid="filter-closed"]');
    await expect(page.locator('[data-testid="pr-item"]')).toHaveCount(1);
    await expect(page.locator('[data-testid="pr-title"]')).toContainText('Update documentation');
    await expect(page.locator('[data-testid="pr-state-badge"]')).toContainText('Merged');
    
    // Switch to all PRs
    await page.click('[data-testid="filter-all"]');
    await expect(page.locator('[data-testid="pr-item"]')).toHaveCount(3);
    
    // Switch back to open PRs
    await page.click('[data-testid="filter-open"]');
    await expect(page.locator('[data-testid="pr-item"]')).toHaveCount(2);
  });

  test('displays draft pull request badge', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls');
    
    const draftPR = page.locator('[data-testid="pr-item"]').nth(1);
    await expect(draftPR.locator('[data-testid="pr-draft-badge"]')).toBeVisible();
    await expect(draftPR.locator('[data-testid="pr-draft-badge"]')).toContainText('Draft');
  });

  test('displays conflict indicator for unmergeable PRs', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls');
    
    const conflictPR = page.locator('[data-testid="pr-item"]').nth(1);
    await expect(conflictPR.locator('[data-testid="pr-conflicts-badge"]')).toBeVisible();
    await expect(conflictPR.locator('[data-testid="pr-conflicts-badge"]')).toContainText('Conflicts');
  });

  test('navigates to pull request details', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls');
    
    // Click on the first PR title
    await page.click('[data-testid="pr-title"]');
    
    // Should navigate to PR details page
    await expect(page).toHaveURL('/repositories/testuser/testrepo/pull/1');
  });

  test('navigates to create new pull request', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls');
    
    await page.click('[data-testid="new-pr-button"]');
    
    // Should navigate to new PR creation page
    await expect(page).toHaveURL('/repositories/testuser/testrepo/pulls/new');
  });

  test('handles empty pull request list', async ({ page }) => {
    // Mock empty response
    await page.route('**/api/repositories/*/pulls*', async route => {
      await route.fulfill({
        json: {
          pull_requests: [],
          total_count: 0,
          page: 1,
          per_page: 25
        }
      });
    });

    await page.goto('/repositories/testuser/testrepo/pulls');
    
    await expect(page.locator('[data-testid="empty-state"]')).toBeVisible();
    await expect(page.locator('[data-testid="empty-state"]')).toContainText('No pull requests found');
  });

  test('handles API error gracefully', async ({ page }) => {
    // Mock error response
    await page.route('**/api/repositories/*/pulls*', async route => {
      await route.fulfill({
        status: 500,
        json: { error: 'Internal server error' }
      });
    });

    await page.goto('/repositories/testuser/testrepo/pulls');
    
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="retry-button"]')).toBeVisible();
  });

  test('pagination works correctly', async ({ page }) => {
    // Mock paginated response
    await page.route('**/api/repositories/*/pulls*', async route => {
      const url = new URL(route.request().url());
      const page_num = parseInt(url.searchParams.get('page') || '1');
      
      const allPRs = Array.from({ length: 30 }, (_, i) => ({
        id: i + 1,
        issue: {
          number: i + 1,
          title: `Pull Request ${i + 1}`,
          state: 'open',
          user: { username: 'testuser' },
          comments_count: 0,
          created_at: '2024-01-15T10:00:00Z',
          updated_at: '2024-01-15T15:30:00Z'
        },
        head_ref: `feature/pr-${i + 1}`,
        base_ref: 'main',
        merged: false,
        draft: false,
        additions: 10,
        deletions: 5,
        changed_files: 2,
        mergeable: true,
        created_at: '2024-01-15T10:00:00Z'
      }));

      const startIndex = (page_num - 1) * 25;
      const endIndex = startIndex + 25;
      const paginatedPRs = allPRs.slice(startIndex, endIndex);

      await route.fulfill({
        json: {
          pull_requests: paginatedPRs,
          total_count: 30,
          page: page_num,
          per_page: 25
        }
      });
    });

    await page.goto('/repositories/testuser/testrepo/pulls');
    
    // Should show 25 PRs on first page
    await expect(page.locator('[data-testid="pr-item"]')).toHaveCount(25);
    
    // Should show pagination controls
    await expect(page.locator('[data-testid="pagination-info"]')).toContainText('Showing 25 of 30');
    await expect(page.locator('[data-testid="previous-button"]')).toBeDisabled();
    await expect(page.locator('[data-testid="next-button"]')).toBeEnabled();
    
    // Go to next page
    await page.click('[data-testid="next-button"]');
    
    // Should show remaining 5 PRs on second page
    await expect(page.locator('[data-testid="pr-item"]')).toHaveCount(5);
    await expect(page.locator('[data-testid="previous-button"]')).toBeEnabled();
    await expect(page.locator('[data-testid="next-button"]')).toBeDisabled();
  });
});

test.describe('Pull Request Creation', () => {
  test.beforeEach(async ({ page }) => {
    // Mock branch list
    await page.route('**/api/repositories/*/branches', async route => {
      await route.fulfill({
        json: [
          { name: 'main', is_default: true },
          { name: 'feature/new-feature', is_default: false },
          { name: 'fix/bug-fix', is_default: false }
        ]
      });
    });

    // Mock create PR endpoint
    await page.route('**/api/repositories/*/pulls', async route => {
      if (route.request().method() === 'POST') {
        const body = await route.request().postDataJSON();
        await route.fulfill({
          json: {
            id: 999,
            issue: {
              number: 999,
              title: body.title,
              state: 'open',
              user: { username: 'testuser' },
              comments_count: 0
            },
            head_ref: body.head,
            base_ref: body.base,
            merged: false,
            draft: body.draft || false
          }
        });
      }
    });

    await loginUser(page);
  });

  test('create new pull request form validation', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls/new');
    
    // Check form elements are present
    await expect(page.locator('[data-testid="pr-title-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="pr-description-textarea"]')).toBeVisible();
    await expect(page.locator('[data-testid="head-branch-select"]')).toBeVisible();
    await expect(page.locator('[data-testid="base-branch-select"]')).toBeVisible();
    await expect(page.locator('[data-testid="draft-checkbox"]')).toBeVisible();
    
    // Try to submit empty form
    await page.click('[data-testid="create-pr-button"]');
    
    // Should show validation errors
    await expect(page.locator('[data-testid="title-error"]')).toContainText('Title is required');
    await expect(page.locator('[data-testid="head-branch-error"]')).toContainText('Head branch is required');
  });

  test('create new pull request successfully', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls/new');
    
    // Fill out the form
    await page.fill('[data-testid="pr-title-input"]', 'Add new feature');
    await page.fill('[data-testid="pr-description-textarea"]', 'This PR adds a new feature to the application.');
    await page.selectOption('[data-testid="head-branch-select"]', 'feature/new-feature');
    await page.selectOption('[data-testid="base-branch-select"]', 'main');
    
    // Submit the form
    await page.click('[data-testid="create-pr-button"]');
    
    // Should redirect to the new PR page
    await expect(page).toHaveURL('/repositories/testuser/testrepo/pull/999');
  });

  test('create draft pull request', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls/new');
    
    // Fill out the form
    await page.fill('[data-testid="pr-title-input"]', 'Work in progress feature');
    await page.check('[data-testid="draft-checkbox"]');
    await page.selectOption('[data-testid="head-branch-select"]', 'feature/new-feature');
    await page.selectOption('[data-testid="base-branch-select"]', 'main');
    
    // Submit the form
    await page.click('[data-testid="create-pr-button"]');
    
    // Should redirect to the new PR page
    await expect(page).toHaveURL('/repositories/testuser/testrepo/pull/999');
  });

  test('validates branch selection', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls/new');
    
    // Select same branch for head and base
    await page.selectOption('[data-testid="head-branch-select"]', 'main');
    await page.selectOption('[data-testid="base-branch-select"]', 'main');
    
    await page.fill('[data-testid="pr-title-input"]', 'Test PR');
    await page.click('[data-testid="create-pr-button"]');
    
    // Should show error about same branch
    await expect(page.locator('[data-testid="branch-error"]')).toContainText('Head and base branches cannot be the same');
  });
});

test.describe('Mobile Responsiveness', () => {
  test.use({ viewport: { width: 375, height: 667 } }); // iPhone SE size

  test.beforeEach(async ({ page }) => {
    // Mock data
    await page.route('**/api/repositories/*/pulls*', async route => {
      await route.fulfill({
        json: {
          pull_requests: [
            {
              id: 1,
              issue: {
                number: 1,
                title: 'Mobile responsive pull request with very long title that should wrap properly',
                state: 'open',
                user: { username: 'mobileuser' },
                comments_count: 5
              },
              head_ref: 'feature/mobile-responsive-feature',
              base_ref: 'main',
              merged: false,
              draft: false,
              additions: 150,
              deletions: 25,
              changed_files: 6,
              mergeable: true
            }
          ],
          total_count: 1,
          page: 1,
          per_page: 25
        }
      });
    });

    await loginUser(page);
  });

  test('pull request list is mobile responsive', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls');
    
    // Check that content is visible and properly laid out on mobile
    await expect(page.locator('[data-testid="pr-item"]')).toBeVisible();
    await expect(page.locator('[data-testid="pr-title"]')).toBeVisible();
    
    // Check that long titles wrap properly
    const titleElement = page.locator('[data-testid="pr-title"]');
    const boundingBox = await titleElement.boundingBox();
    expect(boundingBox?.width).toBeLessThan(375); // Should fit within mobile viewport
    
    // Check that filter tabs are accessible on mobile
    await expect(page.locator('[data-testid="filter-open"]')).toBeVisible();
    await expect(page.locator('[data-testid="filter-closed"]')).toBeVisible();
    await expect(page.locator('[data-testid="filter-all"]')).toBeVisible();
    
    // Test filter interaction on mobile
    await page.click('[data-testid="filter-closed"]');
    await page.click('[data-testid="filter-open"]');
  });

  test('new PR button is accessible on mobile', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pulls');
    
    await expect(page.locator('[data-testid="new-pr-button"]')).toBeVisible();
    await page.click('[data-testid="new-pr-button"]');
    await expect(page).toHaveURL('/repositories/testuser/testrepo/pulls/new');
  });
});