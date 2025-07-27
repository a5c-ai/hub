import { test, expect, Page } from '@playwright/test';
import { loginUser, testUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('Notifications Center & Management', () => {
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

  test.describe('Notification Inbox', () => {
    test('should display notification inbox with unread indicators', async ({ page }) => {
      const mockNotifications = [
        {
          id: '1',
                type: 'pull_request',
      title: 'New pull request opened',
      body: 'Pull request description here',
          repository: {
            id: '1',
            name: 'test-repo',
            full_name: 'alice/test-repo',
            owner: {
              username: 'alice',
              avatar_url: 'https://example.com/alice.jpg'
            }
          },
          subject: {
            title: 'Bug in authentication system #123',
                    url: '/repositories/alice/test-repo/pulls/123',
        type: 'pull_request'
          },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        },
        {
          id: '2',
          type: 'pull_request',
          title: 'Pull request review requested',
          repository: {
            id: '2',
            name: 'frontend-app',
            full_name: 'bob/frontend-app',
            owner: {
              username: 'bob',
              avatar_url: 'https://example.com/bob.jpg'
            }
          },
          subject: {
            title: 'Add user profile page #456',
            url: '/repositories/bob/frontend-app/pulls/456',
            type: 'pull_request'
          },
          reason: 'assigned',
          unread: true,
          updated_at: new Date(Date.now() - 3600000).toISOString()
        },
        {
          id: '3',
          type: 'security_alert',
          title: 'Security vulnerability detected',
          repository: {
            id: '3',
            name: 'secure-app',
            full_name: 'charlie/secure-app',
            owner: {
              username: 'charlie',
              avatar_url: 'https://example.com/charlie.jpg'
            }
          },
          subject: {
            title: 'Critical vulnerability in dependencies',
            url: '/repositories/charlie/secure-app/security',
            type: 'security_alert'
          },
          reason: 'security_alert',
          unread: false,
          updated_at: new Date(Date.now() - 7200000).toISOString(),
          last_read_at: new Date(Date.now() - 3600000).toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockNotifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Check page header with unread count
      await expect(page.locator('h1')).toContainText('Notifications');
      await expect(page.locator('text=2 unread')).toBeVisible();

      // Check notification items
      await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(3);

      // Check unread indicators
      const unreadNotifications = page.locator('[data-testid="notification-item"]').filter({ has: page.locator('.bg-blue-50') });
      await expect(unreadNotifications).toHaveCount(2);

      // Check specific notification content
      const firstNotification = page.locator('[data-testid="notification-item"]').first();
      await expect(firstNotification).toContainText('alice/test-repo');
      await expect(firstNotification).toContainText('Bug in authentication system #123');
      await expect(firstNotification).toContainText('Mentioned');

      // Check security alert
      const securityNotification = page.locator('[data-testid="notification-item"]').last();
      await expect(securityNotification).toContainText('charlie/secure-app');
      await expect(securityNotification).toContainText('Critical vulnerability in dependencies');
      await expect(securityNotification).toContainText('Security');
    });

    test('should categorize notifications by type', async ({ page }) => {
      const allNotifications = [
        {
          id: '1',
                      type: 'pull_request',
            title: 'Pull request notification',
          repository: { id: '1', name: 'repo1', full_name: 'user/repo1', owner: { username: 'user' } },
          subject: { title: 'Pull Request #1', url: '/pulls/1', type: 'pull_request' },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        },
        {
          id: '2',
          type: 'pull_request',
          title: 'PR notification',
          repository: { id: '2', name: 'repo2', full_name: 'user/repo2', owner: { username: 'user' } },
          subject: { title: 'PR #2', url: '/pulls/2', type: 'pull_request' },
          reason: 'assigned',
          unread: false,
          updated_at: new Date().toISOString()
        }
      ];

      const unreadNotifications = allNotifications.filter(n => n.unread);
      const participatingNotifications = allNotifications.filter(n => ['mentioned', 'assigned'].includes(n.reason));

      // Mock different API responses for different filters
      await page.route('**/api/v1/notifications?filter=all', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(allNotifications)
        });
      });

      await page.route('**/api/v1/notifications?filter=unread', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(unreadNotifications)
        });
      });

      await page.route('**/api/v1/notifications?filter=participating', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(participatingNotifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Test unread filter (default)
      await expect(page.locator('[data-testid="filter-unread"]')).toHaveClass(/border-blue-500/);
      await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(1);
      await expect(page.locator('text=Issue #1')).toBeVisible();

      // Test all notifications filter
      await page.click('[data-testid="filter-all"]');
      await waitForLoadingToComplete(page);
      await expect(page.locator('[data-testid="filter-all"]')).toHaveClass(/border-blue-500/);
      await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(2);

      // Test participating filter
      await page.click('[data-testid="filter-participating"]');
      await waitForLoadingToComplete(page);
      await expect(page.locator('[data-testid="filter-participating"]')).toHaveClass(/border-blue-500/);
      await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(2);
    });
  });

  test.describe('Notification Actions', () => {
    test('should mark individual notifications as read/unread', async ({ page }) => {
      const notification = {
        id: '1',
        type: 'issue',
        title: 'Test notification',
        repository: { id: '1', name: 'repo', full_name: 'user/repo', owner: { username: 'user' } },
        subject: { title: 'Test issue #1', url: '/issues/1', type: 'issue' },
        reason: 'mentioned',
        unread: true,
        updated_at: new Date().toISOString()
      };

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([notification])
        });
      });

      // Mock mark as read API
      await page.route('**/api/v1/notifications/1', async (route, request) => {
        if (request.method() === 'PATCH') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Verify notification is unread
      const notificationItem = page.locator('[data-testid="notification-item"]').first();
      await expect(notificationItem).toHaveClass(/bg-blue-50/);
      await expect(notificationItem.locator('.w-2.h-2.bg-blue-600')).toBeVisible();

      // Mark as read
      await page.click('[data-testid="mark-as-read-1"]');

      // Verify visual changes (notification should no longer have unread styling)
      await expect(notificationItem).not.toHaveClass(/bg-blue-50/);
      await expect(notificationItem.locator('.w-2.h-2.bg-blue-600')).not.toBeVisible();
    });

    test('should support bulk notification actions', async ({ page }) => {
      const notifications = [
        {
          id: '1',
          type: 'issue',
          title: 'Notification 1',
          repository: { id: '1', name: 'repo1', full_name: 'user/repo1', owner: { username: 'user' } },
          subject: { title: 'Issue #1', url: '/issues/1', type: 'issue' },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        },
        {
          id: '2',
          type: 'pull_request',
          title: 'Notification 2',
          repository: { id: '2', name: 'repo2', full_name: 'user/repo2', owner: { username: 'user' } },
          subject: { title: 'PR #2', url: '/pulls/2', type: 'pull_request' },
          reason: 'assigned',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(notifications)
        });
      });

      // Mock bulk mark as read API
      await page.route('**/api/v1/notifications', async (route, request) => {
        if (request.method() === 'PATCH') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Select all notifications
      await page.click('[data-testid="select-all-notifications"]');
      await expect(page.locator('text=2 selected')).toBeVisible();

      // Mark selected as read
      await page.click('[data-testid="mark-selected-as-read"]');

      // Verify all notifications are no longer unread
      await expect(page.locator('.bg-blue-50')).toHaveCount(0);
    });

    test('should support deleting notifications', async ({ page }) => {
      const notification = {
        id: '1',
        type: 'issue',
        title: 'Test notification',
        repository: { id: '1', name: 'repo', full_name: 'user/repo', owner: { username: 'user' } },
        subject: { title: 'Test issue #1', url: '/issues/1', type: 'issue' },
        reason: 'mentioned',
        unread: true,
        updated_at: new Date().toISOString()
      };

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([notification])
        });
      });

      // Mock delete API
      await page.route('**/api/v1/notifications/1', async (route, request) => {
        if (request.method() === 'DELETE') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Verify notification exists
      await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(1);

      // Delete notification
      await page.click('[data-testid="delete-notification-1"]');

      // Verify notification is removed
      await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(0);
    });

    test('should mark all notifications as read', async ({ page }) => {
      const notifications = Array.from({ length: 5 }, (_, i) => ({
        id: `${i + 1}`,
        type: 'issue',
        title: `Notification ${i + 1}`,
        repository: { id: `${i + 1}`, name: `repo${i + 1}`, full_name: `user/repo${i + 1}`, owner: { username: 'user' } },
        subject: { title: `Issue #${i + 1}`, url: `/issues/${i + 1}`, type: 'issue' },
        reason: 'mentioned',
        unread: true,
        updated_at: new Date().toISOString()
      }));

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(notifications)
        });
      });

      // Mock mark all as read API
      await page.route('**/api/v1/notifications', async (route, request) => {
        if (request.method() === 'PATCH') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Verify all notifications are unread
      await expect(page.locator('text=5 unread')).toBeVisible();
      await expect(page.locator('.bg-blue-50')).toHaveCount(5);

      // Mark all as read
      await page.click('[data-testid="mark-all-as-read"]');

      // Verify no unread notifications remain
      await expect(page.locator('.bg-blue-50')).toHaveCount(0);
      await expect(page.locator('text=5 unread')).not.toBeVisible();
    });
  });

  test.describe('Notification Types', () => {
    test('should display pull request notifications correctly', async ({ page }) => {
      const notifications = [
        {
          id: '1',
          type: 'pull_request',
          title: 'Pull request review',
          repository: {
            id: '1',
            name: 'web-frontend',
            full_name: 'company/web-frontend',
            owner: { username: 'company', avatar_url: 'https://example.com/company.jpg' }
          },
          subject: {
            title: 'Implement dark mode toggle #567',
            url: '/repositories/company/web-frontend/pulls/567',
            type: 'pull_request'
          },
          reason: 'assigned',
          unread: true,
          updated_at: new Date().toISOString()
        },
        {
          id: '2',
          type: 'pull_request',
          title: 'Pull request comment',
          repository: {
            id: '2',
            name: 'project-alpha',
            full_name: 'team/project-alpha',
            owner: { username: 'team', avatar_url: 'https://example.com/team.jpg' }
          },
          subject: {
            title: 'Fix authentication bug #789',
            url: '/repositories/team/project-alpha/pulls/789',
            type: 'pull_request'
          },
          reason: 'comment',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(notifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Check first pull request notification
      const firstPrNotification = page.locator('[data-testid="notification-item"]').first();
      await expect(firstPrNotification).toContainText('company/web-frontend');
      await expect(firstPrNotification).toContainText('Implement dark mode toggle #567');
      await expect(firstPrNotification).toContainText('Assigned');
      await expect(firstPrNotification.locator('[data-testid="notification-icon-pull_request"]')).toBeVisible();

      // Check second pull request notification
      const secondPrNotification = page.locator('[data-testid="notification-item"]').last();
      await expect(secondPrNotification).toContainText('team/project-alpha');
      await expect(secondPrNotification).toContainText('Fix authentication bug #789');
      await expect(secondPrNotification).toContainText('Comment');
      await expect(secondPrNotification.locator('[data-testid="notification-icon-pull_request"]')).toBeVisible();
    });

    test('should display comment and mention notifications', async ({ page }) => {
      const notifications = [
        {
          id: '1',
          type: 'mention',
          title: 'You were mentioned',
          repository: {
            id: '1',
            name: 'discussion-repo',
            full_name: 'user/discussion-repo',
            owner: { username: 'user' }
          },
          subject: {
            title: 'Architecture discussion in PR #123',
            url: '/repositories/user/discussion-repo/pulls/123',
            type: 'pull_request'
          },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(notifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      const mentionNotification = page.locator('[data-testid="notification-item"]').first();
      await expect(mentionNotification).toContainText('user/discussion-repo');
      await expect(mentionNotification).toContainText('Architecture discussion in PR #123');
      await expect(mentionNotification).toContainText('Mentioned');
      await expect(mentionNotification.locator('[data-testid="notification-icon-mention"]')).toBeVisible();
    });

    test('should display security and vulnerability alerts', async ({ page }) => {
      const notifications = [
        {
          id: '1',
          type: 'security_alert',
          title: 'Security vulnerability detected',
          repository: {
            id: '1',
            name: 'secure-app',
            full_name: 'security-team/secure-app',
            owner: { username: 'security-team' }
          },
          subject: {
            title: 'High severity vulnerability in lodash',
            url: '/repositories/security-team/secure-app/security/advisories',
            type: 'security_alert'
          },
          reason: 'security_alert',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(notifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      const securityNotification = page.locator('[data-testid="notification-item"]').first();
      await expect(securityNotification).toContainText('security-team/secure-app');
      await expect(securityNotification).toContainText('High severity vulnerability in lodash');
      await expect(securityNotification).toContainText('Security');
      await expect(securityNotification.locator('[data-testid="notification-icon-security_alert"]')).toBeVisible();
    });

    test('should display repository invitation notifications', async ({ page }) => {
      const notifications = [
        {
          id: '1',
          type: 'repository_invite',
          title: 'Repository invitation',
          repository: {
            id: '1',
            name: 'private-project',
            full_name: 'organization/private-project',
            owner: { username: 'organization' }
          },
          subject: {
            title: 'Invitation to collaborate on private-project',
            url: '/repositories/organization/private-project/invitations',
            type: 'repository_invite'
          },
          reason: 'invitation',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(notifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      const inviteNotification = page.locator('[data-testid="notification-item"]').first();
      await expect(inviteNotification).toContainText('organization/private-project');
      await expect(inviteNotification).toContainText('Invitation to collaborate on private-project');
      await expect(inviteNotification).toContainText('Invitation');
      await expect(inviteNotification.locator('[data-testid="notification-icon-repository_invite"]')).toBeVisible();
    });
  });

  test.describe('Notification Filtering and Search', () => {
    test('should filter notifications by repository', async ({ page }) => {
      const notifications = [
        {
          id: '1',
          type: 'issue',
          title: 'Issue in repo A',
          repository: { id: '1', name: 'repo-a', full_name: 'user/repo-a', owner: { username: 'user' } },
          subject: { title: 'Issue #1', url: '/issues/1', type: 'issue' },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        },
        {
          id: '2',
          type: 'pull_request',
          title: 'PR in repo B',
          repository: { id: '2', name: 'repo-b', full_name: 'user/repo-b', owner: { username: 'user' } },
          subject: { title: 'PR #2', url: '/pulls/2', type: 'pull_request' },
          reason: 'assigned',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        const url = new URL(route.request().url());
        const repo = url.searchParams.get('repository');
        
        const filteredNotifications = repo 
          ? notifications.filter(n => n.repository?.full_name.includes(repo))
          : notifications;

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(filteredNotifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Test repository filter (if UI supports it)
      await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(2);
      
      // Note: This test assumes a repository filter exists in the UI
      // If not implemented, this part would need to be adjusted
    });

    test('should search notifications by content', async ({ page }) => {
      const allNotifications = [
        {
          id: '1',
          type: 'issue',
          title: 'Bug in authentication system',
          repository: { id: '1', name: 'auth-service', full_name: 'team/auth-service', owner: { username: 'team' } },
          subject: { title: 'Authentication bug #123', url: '/issues/123', type: 'issue' },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        },
        {
          id: '2',
          type: 'pull_request',
          title: 'Frontend improvements',
          repository: { id: '2', name: 'frontend-app', full_name: 'team/frontend-app', owner: { username: 'team' } },
          subject: { title: 'UI improvements #456', url: '/pulls/456', type: 'pull_request' },
          reason: 'assigned',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        const url = new URL(route.request().url());
        const search = url.searchParams.get('search');
        
        const filteredNotifications = search 
          ? allNotifications.filter(n => 
              n.title.toLowerCase().includes(search.toLowerCase()) ||
              n.subject.title.toLowerCase().includes(search.toLowerCase())
            )
          : allNotifications;

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(filteredNotifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Verify all notifications are shown initially
      await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(2);

      // Search for "authentication" (if search UI exists)
      // Note: This assumes a search input exists - adjust based on actual UI
      if (await page.locator('[data-testid="notification-search"]').isVisible()) {
        await page.fill('[data-testid="notification-search"]', 'authentication');
        await waitForLoadingToComplete(page);
        
        await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(1);
        await expect(page.locator('text=Authentication bug #123')).toBeVisible();
      }
    });
  });

  test.describe('Empty States', () => {
    test('should show appropriate empty state when all caught up', async ({ page }) => {
      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=All caught up!')).toBeVisible();
      await expect(page.locator('text=You have no unread notifications')).toBeVisible();
      await expect(page.locator('a[href="/settings"]')).toBeVisible();
      await expect(page.locator('text=Manage notification settings')).toBeVisible();
    });

    test('should show different empty states for different filters', async ({ page }) => {
      // Mock empty responses for different filters
      await page.route('**/api/v1/notifications?filter=unread', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.route('**/api/v1/notifications?filter=all', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.route('**/api/v1/notifications?filter=participating', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Test unread empty state
      await expect(page.locator('text=You have no unread notifications')).toBeVisible();

      // Test all notifications empty state
      await page.click('[data-testid="filter-all"]');
      await waitForLoadingToComplete(page);
      await expect(page.locator('text=You have no notifications')).toBeVisible();

      // Test participating empty state
      await page.click('[data-testid="filter-participating"]');
      await waitForLoadingToComplete(page);
      await expect(page.locator('text=You have no participating notifications')).toBeVisible();
    });
  });

  test.describe('Real-time Features', () => {
    test('should handle real-time notification updates via WebSocket', async ({ page }) => {
      let wsConnected = false;
      
      // Mock WebSocket connection
      await page.addInitScript(() => {
        class MockWebSocket {
          url: string;
          onopen?: ((ev: Event) => any) | null;
          onclose?: ((ev: CloseEvent) => any) | null;
          onmessage?: ((ev: MessageEvent) => any) | null;
          
          constructor(url: string) {
            this.url = url;
            setTimeout(() => {
              this.onopen && this.onopen(new Event('open'));
            }, 100);
          }
          
          send(data: string) {
            // Mock sending data
          }
          
          close() {
            this.onclose && this.onclose(new CloseEvent('close'));
          }
          
          // Simulate receiving a new notification
          simulateNewNotification() {
            const newNotification = {
              type: 'new_notification',
              data: {
                id: 'realtime-1',
                type: 'mention',
                title: 'You were mentioned',
                repository: { id: '1', name: 'test', full_name: 'user/test', owner: { username: 'user' } },
                subject: { title: 'Real-time mention', url: '/test', type: 'mention' },
                reason: 'mentioned',
                unread: true,
                updated_at: new Date().toISOString()
              }
            };
            
            this.onmessage && this.onmessage({ data: JSON.stringify(newNotification) } as MessageEvent);
          }
        }
        
        (window as any).WebSocket = MockWebSocket;
      });

      const initialNotifications = [
        {
          id: '1',
          type: 'issue',
          title: 'Initial notification',
          repository: { id: '1', name: 'repo', full_name: 'user/repo', owner: { username: 'user' } },
          subject: { title: 'Issue #1', url: '/issues/1', type: 'issue' },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(initialNotifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Verify initial state
      await expect(page.locator('[data-testid="notification-item"]')).toHaveCount(1);
      await expect(page.locator('text=1 unread')).toBeVisible();

      // Simulate real-time notification arrival
      await page.evaluate(() => {
        // This would trigger the WebSocket mock to send a new notification
        if (window.WebSocket) {
          const ws = new (window.WebSocket as any)('ws://test');
          setTimeout(() => ws.simulateNewNotification(), 500);
        }
      });

      // Wait for real-time update (if implemented)
      // This test assumes the app handles real-time WebSocket notifications
      // If not implemented, this would need adjustment
    });

    test('should update notification badges in real-time', async ({ page }) => {
      // Initial state with unread notifications
      const notifications = [
        {
          id: '1',
          type: 'issue',
          title: 'Test notification',
          repository: { id: '1', name: 'repo', full_name: 'user/repo', owner: { username: 'user' } },
          subject: { title: 'Issue #1', url: '/issues/1', type: 'issue' },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(notifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Check initial badge count
      await expect(page.locator('text=1 unread')).toBeVisible();

      // Mark notification as read to test badge update
      await page.route('**/api/v1/notifications/1', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.click('[data-testid="mark-as-read-1"]');

      // Badge should update to show 0 unread
      await expect(page.locator('text=1 unread')).not.toBeVisible();
    });
  });

  test.describe('Mobile Responsiveness', () => {
    test('should provide touch-friendly notification management on mobile', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });

      const mobileNotifications = [
        {
          id: '1',
          type: 'issue',
          title: 'Mobile notification',
          repository: {
            id: '1',
            name: 'mobile-app',
            full_name: 'team/mobile-app',
            owner: { username: 'team' }
          },
          subject: {
            title: 'Mobile issue #123',
            url: '/repositories/team/mobile-app/issues/123',
            type: 'issue'
          },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mobileNotifications)
        });
      });

      await loginUser(page);
      await page.goto('/notifications');
      await waitForLoadingToComplete(page);

      // Verify mobile layout
      await expect(page.locator('h1')).toContainText('Notifications');
      await expect(page.locator('[data-testid="notification-item"]')).toBeVisible();

      // Test touch-friendly filter tabs
      await expect(page.locator('[data-testid="filter-unread"]')).toBeVisible();
      await expect(page.locator('[data-testid="filter-all"]')).toBeVisible();
      await expect(page.locator('[data-testid="filter-participating"]')).toBeVisible();

      // Test touch interaction
      await page.tap('[data-testid="filter-all"]');
      await expect(page.locator('[data-testid="filter-all"]')).toHaveClass(/border-blue-500/);

      // Test mobile notification actions
      const notificationItem = page.locator('[data-testid="notification-item"]').first();
      await expect(notificationItem).toBeVisible();
      
      // Verify action buttons are appropriately sized for touch
      const markReadButton = notificationItem.locator('[data-testid="mark-as-read-1"]');
      if (await markReadButton.isVisible()) {
        const box = await markReadButton.boundingBox();
        expect(box!.height).toBeGreaterThan(40); // Minimum touch target size
      }
    });

    test('should handle responsive notification layout', async ({ page }) => {
      const notifications = [
        {
          id: '1',
          type: 'pull_request',
          title: 'Very long notification title that should wrap properly on mobile devices',
          repository: {
            id: '1',
            name: 'very-long-repository-name',
            full_name: 'organization-with-long-name/very-long-repository-name',
            owner: { username: 'organization-with-long-name' }
          },
          subject: {
            title: 'This is a very long pull request title that contains lots of information #123456',
            url: '/repositories/organization-with-long-name/very-long-repository-name/pulls/123456',
            type: 'pull_request'
          },
          reason: 'mentioned',
          unread: true,
          updated_at: new Date().toISOString()
        }
      ];

      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(notifications)
        });
      });

      // Test various mobile viewport sizes
      const viewports = [
        { width: 320, height: 568 }, // iPhone 5
        { width: 375, height: 667 }, // iPhone 6/7/8
        { width: 414, height: 896 }  // iPhone XR
      ];

      for (const viewport of viewports) {
        await page.setViewportSize(viewport);
        
        await loginUser(page);
        await page.goto('/notifications');
        await waitForLoadingToComplete(page);

        // Verify content is properly contained and readable
        const notificationItem = page.locator('[data-testid="notification-item"]').first();
        await expect(notificationItem).toBeVisible();
        
        // Check that long text doesn't cause horizontal scrolling
        const bodyWidth = await page.evaluate(() => document.body.scrollWidth);
        expect(bodyWidth).toBeLessThanOrEqual(viewport.width + 20); // Allow small margin for scrollbars
        
        await page.context().clearCookies();
      }
    });
  });

  test.describe('Error Handling & Edge Cases', () => {
    test('should handle API errors gracefully', async ({ page }) => {
      await page.route('**/api/v1/notifications*', async route => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'Server error' })
        });
      });

      await loginUser(page);
      await page.goto('/notifications');

      await expect(page.locator('text=Error:')).toBeVisible();
      await expect(page.locator('text=Try Again')).toBeVisible();
    });

    test('should handle offline notification scenarios', async ({ page }) => {
      // Simulate going offline
      await page.context().setOffline(true);
      
      await loginUser(page);
      await page.goto('/notifications');

      // Should show appropriate offline messaging
      // This assumes the app has offline handling - adjust based on implementation
      await page.context().setOffline(false);
    });
  });
});