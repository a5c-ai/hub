import { test, expect } from '@playwright/test';
import { loginUser, testUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('Activity Feed & Timeline', () => {
  test.beforeEach(async ({ page }) => {
    await page.context().clearCookies();
    await page.context().clearPermissions();
    
    // Mock authentication
    await page.route('**/api/v1/auth/me', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: '1',
            name: testUser.name,
            username: testUser.username,
            email: testUser.email,
            avatar_url: 'https://example.com/avatar.jpg'
          }
        })
      });
    });
  });

  test.describe('Activity Feed Display', () => {
    test('should display global activity feed with different event types', async ({ page }) => {
      const mockActivities = [
        {
          id: '1',
          type: 'push',
          action: 'pushed',
          actor: {
            id: '1',
            username: 'alice',
            avatar_url: 'https://example.com/alice.jpg'
          },
          repository: {
            id: '1',
            name: 'test-repo',
            full_name: 'alice/test-repo',
            owner: { username: 'alice' }
          },
          payload: {
            commits: [
              { sha: 'abc123', message: 'Add new feature' },
              { sha: 'def456', message: 'Fix bug in authentication' }
            ]
          },
          created_at: new Date().toISOString()
        },
        {
          id: '2',
          type: 'pull_request',
          action: 'opened',
          actor: {
            id: '2',
            username: 'bob',
            avatar_url: 'https://example.com/bob.jpg'
          },
          repository: {
            id: '2',
            name: 'project-x',
            full_name: 'bob/project-x',
            owner: { username: 'bob' }
          },
          payload: {
            number: 123,
            title: 'Implement user authentication'
          },
          created_at: new Date(Date.now() - 3600000).toISOString()
        },
        {
          id: '3',
          type: 'issue',
          action: 'opened',
          actor: {
            id: '3',
            username: 'charlie',
            avatar_url: 'https://example.com/charlie.jpg'
          },
          repository: {
            id: '3',
            name: 'bug-tracker',
            full_name: 'charlie/bug-tracker',
            owner: { username: 'charlie' }
          },
          payload: {
            number: 456,
            title: 'Application crashes on startup'
          },
          created_at: new Date(Date.now() - 7200000).toISOString()
        },
        {
          id: '4',
          type: 'star',
          action: 'starred',
          actor: {
            id: '4',
            username: 'diana',
            avatar_url: 'https://example.com/diana.jpg'
          },
          repository: {
            id: '4',
            name: 'awesome-project',
            full_name: 'diana/awesome-project',
            owner: { username: 'diana' }
          },
          payload: {},
          created_at: new Date(Date.now() - 10800000).toISOString()
        },
        {
          id: '5',
          type: 'fork',
          action: 'forked',
          actor: {
            id: '5',
            username: 'eve',
            avatar_url: 'https://example.com/eve.jpg'
          },
          repository: {
            id: '5',
            name: 'popular-lib',
            full_name: 'eve/popular-lib',
            owner: { username: 'eve' }
          },
          payload: {},
          created_at: new Date(Date.now() - 14400000).toISOString()
        }
      ];

      await page.route('**/api/v1/activity*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockActivities)
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      await waitForLoadingToComplete(page);

      // Check page header
      await expect(page.locator('h1')).toContainText('Activity Feed');
      await expect(page.locator('text=Stay up to date with what\'s happening')).toBeVisible();

      // Check that all activity types are displayed
      await expect(page.locator('[data-testid="activity-item"]')).toHaveCount(5);

      // Verify push activity
      const pushActivity = page.locator('[data-testid="activity-item"]').first();
      await expect(pushActivity).toContainText('alice pushed 2 commits');
      await expect(pushActivity).toContainText('alice/test-repo');
      await expect(pushActivity).toContainText('Add new feature');
      await expect(pushActivity).toContainText('abc123');

      // Verify pull request activity
      const prActivity = page.locator('[data-testid="activity-item"]').nth(1);
      await expect(prActivity).toContainText('bob opened pull request');
      await expect(prActivity).toContainText('#123');
      await expect(prActivity).toContainText('bob/project-x');

      // Verify issue activity
      const issueActivity = page.locator('[data-testid="activity-item"]').nth(2);
      await expect(issueActivity).toContainText('charlie opened issue');
      await expect(issueActivity).toContainText('#456');
      await expect(issueActivity).toContainText('charlie/bug-tracker');

      // Verify star activity
      const starActivity = page.locator('[data-testid="activity-item"]').nth(3);
      await expect(starActivity).toContainText('diana starred');
      await expect(starActivity).toContainText('diana/awesome-project');

      // Verify fork activity
      const forkActivity = page.locator('[data-testid="activity-item"]').nth(4);
      await expect(forkActivity).toContainText('eve forked');
      await expect(forkActivity).toContainText('eve/popular-lib');
    });

    test('should filter activity by type', async ({ page }) => {
      const allActivities = [
        {
          id: '1',
          type: 'push',
          action: 'pushed',
          actor: { id: '1', username: 'alice' },
          repository: { id: '1', name: 'repo1', full_name: 'alice/repo1', owner: { username: 'alice' } },
          payload: { commits: [{ sha: 'abc123', message: 'Test commit' }] },
          created_at: new Date().toISOString()
        }
      ];

      const ownActivities = [
        {
          id: '2',
          type: 'pull_request',
          action: 'opened',
          actor: { id: '1', username: testUser.username },
          repository: { id: '2', name: 'my-repo', full_name: `${testUser.username}/my-repo`, owner: { username: testUser.username } },
          payload: { number: 1, title: 'My PR' },
          created_at: new Date().toISOString()
        }
      ];

      const followingActivities = [
        {
          id: '3',
          type: 'star',
          action: 'starred',
          actor: { id: '3', username: 'followed-user' },
          repository: { id: '3', name: 'followed-repo', full_name: 'followed-user/followed-repo', owner: { username: 'followed-user' } },
          payload: {},
          created_at: new Date().toISOString()
        }
      ];

      // Mock different responses for different filters
      await page.route('**/api/v1/activity?filter=all', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(allActivities)
        });
      });

      await page.route('**/api/v1/activity?filter=own', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(ownActivities)
        });
      });

      await page.route('**/api/v1/activity?filter=following', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(followingActivities)
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      await waitForLoadingToComplete(page);

      // Test 'All activity' filter (default)
      await expect(page.locator('[data-testid="filter-all"]')).toHaveClass(/border-primary/);
      await expect(page.locator('[data-testid="activity-item"]')).toHaveCount(1);
      await expect(page.locator('text=alice pushed')).toBeVisible();

      // Test 'Your activity' filter
      await page.click('[data-testid="filter-own"]');
      await waitForLoadingToComplete(page);
      await expect(page.locator('[data-testid="filter-own"]')).toHaveClass(/border-primary/);
      await expect(page.locator('[data-testid="activity-item"]')).toHaveCount(1);
      await expect(page.locator('text=testuser opened pull request')).toBeVisible();

      // Test 'Following' filter
      await page.click('[data-testid="filter-following"]');
      await waitForLoadingToComplete(page);
      await expect(page.locator('[data-testid="filter-following"]')).toHaveClass(/border-primary/);
      await expect(page.locator('[data-testid="activity-item"]')).toHaveCount(1);
      await expect(page.locator('text=followed-user starred')).toBeVisible();
    });

    test('should display empty state when no activity', async ({ page }) => {
      await page.route('**/api/v1/activity*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=No activity yet')).toBeVisible();
      await expect(page.locator('text=There\'s no activity to show yet')).toBeVisible();
    });
  });

  test.describe('Activity Event Types', () => {
    test('should display repository creation activities', async ({ page }) => {
      const createRepoActivity = [{
        id: '1',
        type: 'create_repository',
        action: 'created',
        actor: {
          id: '1',
          username: 'dev-user',
          avatar_url: 'https://example.com/dev.jpg'
        },
        repository: {
          id: '1',
          name: 'new-project',
          full_name: 'dev-user/new-project',
          owner: { username: 'dev-user' }
        },
        payload: {},
        created_at: new Date().toISOString()
      }];

      await page.route('**/api/v1/activity*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(createRepoActivity)
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=dev-user created repository')).toBeVisible();
      await expect(page.locator('text=dev-user/new-project')).toBeVisible();
    });

    test('should display detailed push events with commit information', async ({ page }) => {
      const pushActivity = [{
        id: '1',
        type: 'push',
        action: 'pushed',
        actor: {
          id: '1',
          username: 'developer',
          avatar_url: 'https://example.com/dev.jpg'
        },
        repository: {
          id: '1',
          name: 'my-app',
          full_name: 'developer/my-app',
          owner: { username: 'developer' }
        },
        payload: {
          commits: [
            { sha: '1a2b3c4d5e6f', message: 'feat: Add user authentication system' },
            { sha: '2b3c4d5e6f7g', message: 'fix: Resolve login redirect issue' },
            { sha: '3c4d5e6f7g8h', message: 'docs: Update API documentation' },
            { sha: '4d5e6f7g8h9i', message: 'test: Add integration tests for auth' },
            { sha: '5e6f7g8h9i0j', message: 'refactor: Clean up authentication code' }
          ]
        },
        created_at: new Date().toISOString()
      }];

      await page.route('**/api/v1/activity*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(pushActivity)
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      await waitForLoadingToComplete(page);

      // Check push activity summary
      await expect(page.locator('text=developer pushed 5 commits')).toBeVisible();
      await expect(page.locator('text=developer/my-app')).toBeVisible();

      // Check that first 3 commits are displayed
      await expect(page.locator('text=1a2b3c4')).toBeVisible();
      await expect(page.locator('text=feat: Add user authentication system')).toBeVisible();
      await expect(page.locator('text=2b3c4d5')).toBeVisible();
      await expect(page.locator('text=fix: Resolve login redirect issue')).toBeVisible();
      await expect(page.locator('text=3c4d5e6')).toBeVisible();
      await expect(page.locator('text=docs: Update API documentation')).toBeVisible();

      // Check that additional commits are summarized
      await expect(page.locator('text=...and 2 more commits')).toBeVisible();
    });

    test('should display follow activities', async ({ page }) => {
      const followActivity = [{
        id: '1',
        type: 'follow',
        action: 'followed',
        actor: {
          id: '1',
          username: 'follower',
          avatar_url: 'https://example.com/follower.jpg'
        },
        repository: null,
        payload: {
          target: { username: 'followee' }
        },
        created_at: new Date().toISOString()
      }];

      await page.route('**/api/v1/activity*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(followActivity)
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=follower followed')).toBeVisible();
      await expect(page.locator('a[href="/users/followee"]')).toBeVisible();
    });
  });

  test.describe('Activity Feed Pagination', () => {
    test('should handle activity feed pagination and infinite scroll', async ({ page }) => {
      const page1Activities = Array.from({ length: 10 }, (_, i) => ({
        id: `page1-${i}`,
        type: 'push',
        action: 'pushed',
        actor: { id: `${i}`, username: `user${i}` },
        repository: { id: `${i}`, name: `repo${i}`, full_name: `user${i}/repo${i}`, owner: { username: `user${i}` } },
        payload: { commits: [{ sha: `sha${i}`, message: `Commit ${i}` }] },
        created_at: new Date(Date.now() - i * 3600000).toISOString()
      }));

      const page2Activities = Array.from({ length: 10 }, (_, i) => ({
        id: `page2-${i}`,
        type: 'pull_request',
        action: 'opened',
        actor: { id: `${i + 10}`, username: `user${i + 10}` },
        repository: { id: `${i + 10}`, name: `repo${i + 10}`, full_name: `user${i + 10}/repo${i + 10}`, owner: { username: `user${i + 10}` } },
        payload: { number: i + 1, title: `PR ${i + 1}` },
        created_at: new Date(Date.now() - (i + 10) * 3600000).toISOString()
      }));

      // Mock first page
      await page.route('**/api/v1/activity?filter=all&page=1*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(page1Activities)
        });
      });

      // Mock second page
      await page.route('**/api/v1/activity?filter=all&page=2*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(page2Activities)
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      await waitForLoadingToComplete(page);

      // Verify first page is loaded
      await expect(page.locator('[data-testid="activity-item"]')).toHaveCount(10);
      await expect(page.locator('text=user0 pushed')).toBeVisible();
      await expect(page.locator('text=user9 pushed')).toBeVisible();

      // Simulate infinite scroll by scrolling to bottom
      await page.evaluate(() => window.scrollTo(0, document.body.scrollHeight));

      // Wait for second page to load
      await expect(page.locator('[data-testid="activity-item"]')).toHaveCount(20);
      await expect(page.locator('text=user10 opened pull request')).toBeVisible();
      await expect(page.locator('text=user19 opened pull request')).toBeVisible();
    });
  });

  test.describe('Personal Activity Timeline', () => {
    test('should display personal activity timeline and history', async ({ page }) => {
      const personalActivities = [
        {
          id: '1',
          type: 'push',
          action: 'pushed',
          actor: { id: '1', username: testUser.username },
          repository: { id: '1', name: 'my-project', full_name: `${testUser.username}/my-project`, owner: { username: testUser.username } },
          payload: { commits: [{ sha: 'abc123', message: 'Initial commit' }] },
          created_at: new Date().toISOString()
        },
        {
          id: '2',
          type: 'create_repository',
          action: 'created',
          actor: { id: '1', username: testUser.username },
          repository: { id: '2', name: 'new-repo', full_name: `${testUser.username}/new-repo`, owner: { username: testUser.username } },
          payload: {},
          created_at: new Date(Date.now() - 86400000).toISOString() // 1 day ago
        }
      ];

      await page.route('**/api/v1/activity?filter=own', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(personalActivities)
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      
      // Switch to personal activity
      await page.click('[data-testid="filter-own"]');
      await waitForLoadingToComplete(page);

      // Verify personal activities are shown
      await expect(page.locator('[data-testid="activity-item"]')).toHaveCount(2);
      await expect(page.locator(`text=${testUser.username} pushed`)).toBeVisible();
      await expect(page.locator(`text=${testUser.username} created repository`)).toBeVisible();
      
      // Verify chronological order (newest first)
      const activities = page.locator('[data-testid="activity-item"]');
      await expect(activities.first()).toContainText('pushed');
      await expect(activities.last()).toContainText('created repository');
    });
  });

  test.describe('Following Users Activity', () => {
    test('should display activity from followed users and organizations', async ({ page }) => {
      const followingActivities = [
        {
          id: '1',
          type: 'star',
          action: 'starred',
          actor: { id: '2', username: 'followed-user' },
          repository: { id: '1', name: 'cool-project', full_name: 'followed-user/cool-project', owner: { username: 'followed-user' } },
          payload: {},
          created_at: new Date().toISOString()
        },
        {
          id: '2',
          type: 'pull_request',
          action: 'merged',
          actor: { id: '3', username: 'org-member' },
          repository: { id: '2', name: 'org-repo', full_name: 'my-org/org-repo', owner: { username: 'my-org' } },
          payload: { number: 42, title: 'Feature implementation' },
          created_at: new Date(Date.now() - 3600000).toISOString()
        }
      ];

      await page.route('**/api/v1/activity?filter=following', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(followingActivities)
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      
      // Switch to following activity
      await page.click('[data-testid="filter-following"]');
      await waitForLoadingToComplete(page);

      // Verify following activities are shown
      await expect(page.locator('[data-testid="activity-item"]')).toHaveCount(2);
      await expect(page.locator('text=followed-user starred')).toBeVisible();
      await expect(page.locator('text=org-member merged pull request')).toBeVisible();
      await expect(page.locator('text=#42')).toBeVisible();
    });

    test('should show empty state for following when not following anyone', async ({ page }) => {
      await page.route('**/api/v1/activity?filter=following', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      
      // Switch to following activity
      await page.click('[data-testid="filter-following"]');
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=No activity yet')).toBeVisible();
      await expect(page.locator('text=Follow users to see their activity here')).toBeVisible();
      await expect(page.locator('a[href="/search"]')).toBeVisible();
      await expect(page.locator('text=Find users to follow')).toBeVisible();
    });
  });

  test.describe('Error Handling', () => {
    test('should handle API errors gracefully', async ({ page }) => {
      await page.route('**/api/v1/activity*', async route => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Internal server error' })
        });
      });

      await loginUser(page);
      await page.goto('/activity');

      await expect(page.locator('text=Error:')).toBeVisible();
      await expect(page.locator('text=Try Again')).toBeVisible();
      
      // Test retry functionality
      await page.route('**/api/v1/activity*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([{
            id: '1',
            type: 'push',
            action: 'pushed',
            actor: { id: '1', username: 'user' },
            repository: { id: '1', name: 'repo', full_name: 'user/repo', owner: { username: 'user' } },
            payload: { commits: [{ sha: 'abc', message: 'fix' }] },
            created_at: new Date().toISOString()
          }])
        });
      });

      await page.click('text=Try Again');
      await waitForLoadingToComplete(page);
      
      await expect(page.locator('text=user pushed')).toBeVisible();
    });
  });

  test.describe('Mobile Responsiveness', () => {
    test('should work correctly on mobile devices', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });

      const mobileActivities = [{
        id: '1',
        type: 'push',
        action: 'pushed',
        actor: { id: '1', username: 'mobile-user' },
        repository: { id: '1', name: 'mobile-app', full_name: 'mobile-user/mobile-app', owner: { username: 'mobile-user' } },
        payload: { commits: [{ sha: 'abc123', message: 'Mobile optimization' }] },
        created_at: new Date().toISOString()
      }];

      await page.route('**/api/v1/activity*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mobileActivities)
        });
      });

      await loginUser(page);
      await page.goto('/activity');
      await waitForLoadingToComplete(page);

      // Verify mobile layout
      await expect(page.locator('h1')).toContainText('Activity Feed');
      await expect(page.locator('[data-testid="activity-item"]')).toBeVisible();
      
      // Check that filter tabs are accessible on mobile
      await expect(page.locator('[data-testid="filter-all"]')).toBeVisible();
      await expect(page.locator('[data-testid="filter-own"]')).toBeVisible();
      await expect(page.locator('[data-testid="filter-following"]')).toBeVisible();

      // Test touch-friendly navigation
      await page.click('[data-testid="filter-own"]');
      await expect(page.locator('[data-testid="filter-own"]')).toHaveClass(/border-primary/);
    });
  });
});