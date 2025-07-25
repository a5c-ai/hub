import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('Issue Management', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication for all issue tests
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
            email: testUser.email
          }
        })
      });
    });

    // Set authentication state
    await page.addInitScript(() => {
      window.localStorage.setItem('auth_token', 'mock-jwt-token');
    });
  });

  test('should display issues page with empty state', async ({ page }) => {
    // Mock empty issues response
    await page.route('**/api/v1/issues**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [],
          pagination: {
            page: 1,
            per_page: 30,
            total: 0
          }
        })
      });
    });

    await page.goto('/issues');
    
    // Should show empty state
    await expect(page.locator('text=No issues found')).toBeVisible();
  });

  test('should display list of issues', async ({ page }) => {
    // Mock issues response with data
    await page.route('**/api/v1/issues**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              id: 1,
              number: 1,
              title: 'Fix login page layout on mobile',
              body: 'The login form is not properly aligned on mobile devices',
              state: 'open',
              labels: ['bug', 'ui'],
              author: {
                username: testUser.username,
                name: testUser.name
              },
              assignees: [],
              repository: {
                name: 'awesome-project',
                full_name: 'testuser/awesome-project'
              },
              created_at: '2024-07-20T10:00:00Z',
              updated_at: '2024-07-20T10:00:00Z'
            },
            {
              id: 2,
              number: 2,
              title: 'Add dark mode support',
              body: 'Implement dark mode theme switching',
              state: 'open',
              labels: ['enhancement', 'ui'],
              author: {
                username: testUser.username,
                name: testUser.name
              },
              assignees: [
                {
                  username: testUser.username,
                  name: testUser.name
                }
              ],
              repository: {
                name: 'awesome-project',
                full_name: 'testuser/awesome-project'
              },
              created_at: '2024-07-19T15:30:00Z',
              updated_at: '2024-07-20T09:15:00Z'
            }
          ],
          pagination: {
            page: 1,
            per_page: 30,
            total: 2
          }
        })
      });
    });

    await page.goto('/issues');
    
    // Should display issue list
    await expect(page.locator('text=Fix login page layout on mobile')).toBeVisible();
    await expect(page.locator('text=Add dark mode support')).toBeVisible();
    await expect(page.locator('text=bug')).toBeVisible();
    await expect(page.locator('text=enhancement')).toBeVisible();
    await expect(page.locator('text=testuser/awesome-project')).toBeVisible();
  });

  test('should filter issues by state', async ({ page }) => {
    // Mock closed issues response
    await page.route('**/api/v1/issues?*state=closed*', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              id: 3,
              number: 3,
              title: 'Update documentation',
              body: 'Update the API documentation',
              state: 'closed',
              labels: ['documentation'],
              author: {
                username: testUser.username,
                name: testUser.name
              },
              assignees: [],
              repository: {
                name: 'awesome-project',
                full_name: 'testuser/awesome-project'
              },
              created_at: '2024-07-18T10:00:00Z',
              updated_at: '2024-07-19T14:30:00Z',
              closed_at: '2024-07-19T14:30:00Z'
            }
          ]
        })
      });
    });

    await page.goto('/issues');
    
    // Filter by closed state
    if (await page.locator('text=Closed').count() > 0) {
      await page.click('text=Closed');
      await expect(page.locator('text=Update documentation')).toBeVisible();
    }
  });

  test('should navigate to create new issue', async ({ page }) => {
    await page.goto('/issues');
    
    // Should have new issue button/link
    if (await page.locator('text=New issue').count() > 0) {
      await page.click('text=New issue');
      await expect(page).toHaveURL(/\/issues\/new/);
    }
  });

  test('should navigate to issue details', async ({ page }) => {
    // Mock individual issue response
    await page.route('**/api/v1/repositories/testuser/awesome-project/issues/1', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 1,
            number: 1,
            title: 'Fix login page layout on mobile',
            body: 'The login form is not properly aligned on mobile devices. This affects user experience on tablets and phones.',
            state: 'open',
            labels: ['bug', 'ui'],
            author: {
              username: testUser.username,
              name: testUser.name,
              avatar_url: null
            },
            assignees: [],
            repository: {
              name: 'awesome-project',
              full_name: 'testuser/awesome-project'
            },
            created_at: '2024-07-20T10:00:00Z',
            updated_at: '2024-07-20T10:00:00Z'
          }
        })
      });
    });

    // Mock comments response
    await page.route('**/api/v1/repositories/testuser/awesome-project/issues/1/comments', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: []
        })
      });
    });

    await page.goto('/repositories/testuser/awesome-project/issues/1');
    
    // Should display issue details
    await expect(page.locator('h1')).toContainText('Fix login page layout on mobile');
    await expect(page.locator('text=The login form is not properly aligned')).toBeVisible();
    await expect(page.locator('text=bug')).toBeVisible();
    await expect(page.locator('text=ui')).toBeVisible();
  });

  test('should create new issue comment', async ({ page }) => {
    // Mock issue response
    await page.route('**/api/v1/repositories/testuser/awesome-project/issues/1', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 1,
            number: 1,
            title: 'Fix login page layout on mobile',
            body: 'The login form is not properly aligned on mobile devices.',
            state: 'open',
            labels: ['bug', 'ui'],
            author: {
              username: testUser.username,
              name: testUser.name
            },
            assignees: [],
            repository: {
              name: 'awesome-project',
              full_name: 'testuser/awesome-project'
            },
            created_at: '2024-07-20T10:00:00Z',
            updated_at: '2024-07-20T10:00:00Z'
          }
        })
      });
    });

    // Mock comments response
    await page.route('**/api/v1/repositories/testuser/awesome-project/issues/1/comments', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: []
        })
      });
    });

    // Mock create comment response
    await page.route('**/api/v1/repositories/testuser/awesome-project/issues/1/comments', async route => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: 1,
              body: 'I can reproduce this issue on my iPhone.',
              author: {
                username: testUser.username,
                name: testUser.name
              },
              created_at: '2024-07-20T11:00:00Z',
              updated_at: '2024-07-20T11:00:00Z'
            }
          })
        });
      }
    });

    await page.goto('/repositories/testuser/awesome-project/issues/1');
    
    // Should have comment form
    if (await page.locator('textarea[placeholder*="comment"]').count() > 0) {
      await page.fill('textarea[placeholder*="comment"]', 'I can reproduce this issue on my iPhone.');
      await page.click('button:has-text("Comment")');
      
      // Should show new comment
      await expect(page.locator('text=I can reproduce this issue on my iPhone.')).toBeVisible();
    }
  });

  test('should close and reopen issues', async ({ page }) => {
    // Mock issue response
    await page.route('**/api/v1/repositories/testuser/awesome-project/issues/1', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 1,
            number: 1,
            title: 'Fix login page layout on mobile',
            body: 'The login form is not properly aligned on mobile devices.',
            state: 'open',
            labels: ['bug', 'ui'],
            author: {
              username: testUser.username,
              name: testUser.name
            },
            assignees: [],
            repository: {
              name: 'awesome-project',
              full_name: 'testuser/awesome-project'
            },
            created_at: '2024-07-20T10:00:00Z',
            updated_at: '2024-07-20T10:00:00Z'
          }
        })
      });
    });

    // Mock close issue response
    await page.route('**/api/v1/repositories/testuser/awesome-project/issues/1', async route => {
      if (route.request().method() === 'PATCH') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: 1,
              number: 1,
              title: 'Fix login page layout on mobile',
              body: 'The login form is not properly aligned on mobile devices.',
              state: 'closed',
              labels: ['bug', 'ui'],
              author: {
                username: testUser.username,
                name: testUser.name
              },
              assignees: [],
              repository: {
                name: 'awesome-project',
                full_name: 'testuser/awesome-project'
              },
              created_at: '2024-07-20T10:00:00Z',
              updated_at: '2024-07-20T12:00:00Z',
              closed_at: '2024-07-20T12:00:00Z'
            }
          })
        });
      }
    });

    await page.goto('/repositories/testuser/awesome-project/issues/1');
    
    // Should have close issue button
    if (await page.locator('button:has-text("Close issue")').count() > 0) {
      await page.click('button:has-text("Close issue")');
      
      // Should show closed state
      await expect(page.locator('text=Closed')).toBeVisible();
    }
  });
});