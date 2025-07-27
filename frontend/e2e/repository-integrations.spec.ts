import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('Repository Webhooks & Integrations', () => {
  const mockRepository = {
    id: '1',
    name: 'test-repo',
    full_name: 'testuser/test-repo',
    description: 'Test repository for integrations tests',
    private: false,
    language: 'TypeScript',
    default_branch: 'main',
    stargazers_count: 10,
    forks_count: 2,
    updated_at: '2024-07-20T10:00:00Z'
  };

  const mockWebhooks = [
    {
      id: '1',
      name: 'CI/CD Webhook',
      config: {
        url: 'https://ci.example.com/webhook',
        content_type: 'json',
        secret: 'webhook-secret'
      },
      events: ['push', 'pull_request'],
      active: true,
      created_at: '2024-07-20T10:00:00Z',
      updated_at: '2024-07-20T10:00:00Z'
    },
    {
      id: '2',
      name: 'Issue Tracker',
      config: {
        url: 'https://tracker.example.com/api/webhook',
        content_type: 'form'
      },
      events: ['issues', 'issue_comment'],
      active: false,
      created_at: '2024-07-19T15:30:00Z',
      updated_at: '2024-07-19T15:30:00Z'
    }
  ];

  test.beforeEach(async ({ page }) => {
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
            email: testUser.email
          }
        })
      });
    });

    // Mock repository details
    await page.route('**/api/v1/repositories/testuser/test-repo', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: mockRepository
        })
      });
    });

    // Set authentication state
    await page.addInitScript(() => {
      window.localStorage.setItem('auth_token', 'mock-jwt-token');
    });
  });

  test.describe('Webhooks Management', () => {
    test.beforeEach(async ({ page }) => {
      // Mock webhooks API
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(mockWebhooks)
          });
        }
      });
    });

    test('should display webhooks management page', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Should display webhooks page
      await expect(page.locator('h1')).toContainText('Webhooks');
      await expect(page.locator('text=Webhooks allow external services to be notified')).toBeVisible();
      await expect(page.locator('button:has-text("Add Webhook")')).toBeVisible();
    });

    test('should list existing webhooks', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Should display webhook list
      await expect(page.locator('text=CI/CD Webhook')).toBeVisible();
      await expect(page.locator('text=Issue Tracker')).toBeVisible();
      
      // Should show webhook details
      await expect(page.locator('text=https://ci.example.com/webhook')).toBeVisible();
      await expect(page.locator('text=https://tracker.example.com/api/webhook')).toBeVisible();
      
      // Should show status badges
      await expect(page.locator('text=Active')).toBeVisible();
      await expect(page.locator('text=Inactive')).toBeVisible();
      
      // Should show event tags
      await expect(page.locator('text=push')).toBeVisible();
      await expect(page.locator('text=pull_request')).toBeVisible();
      await expect(page.locator('text=issues')).toBeVisible();
      await expect(page.locator('text=issue_comment')).toBeVisible();
    });

    test('should show empty state when no webhooks exist', async ({ page }) => {
      // Mock empty webhooks response
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Should show empty state
      await expect(page.locator('text=No webhooks')).toBeVisible();
      await expect(page.locator('text=Get started by creating your first webhook')).toBeVisible();
    });

    test('should display webhook action buttons', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Should show action buttons for each webhook
      await expect(page.locator('button:has-text("Ping")')).toHaveCount(2);
      await expect(page.locator('button:has-text("Disable")')).toBeVisible(); // For active webhook
      await expect(page.locator('button:has-text("Enable")')).toBeVisible(); // For inactive webhook
      await expect(page.locator('button:has-text("Edit")')).toHaveCount(2);
      await expect(page.locator('button:has-text("Delete")')).toHaveCount(2);
    });

    test('should show correct breadcrumb navigation', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Should show breadcrumb navigation
      await expect(page.locator('nav a[href="/repositories"]')).toContainText('Repositories');
      await expect(page.locator('nav a[href="/repositories/testuser/test-repo"]')).toContainText('testuser/test-repo');
      await expect(page.locator('nav a[href="/repositories/testuser/test-repo/settings"]')).toContainText('Settings');
      await expect(page.locator('nav text=Webhooks')).toBeVisible();
    });
  });

  test.describe('Webhook Creation', () => {
    test('should open create webhook modal', async ({ page }) => {
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Click add webhook button
      await page.click('button:has-text("Add Webhook")');
      
      // Should show modal
      await expect(page.locator('text=Add Webhook')).toBeVisible();
      await expect(page.locator('input[placeholder="Webhook name"]')).toBeVisible();
      await expect(page.locator('input[placeholder="https://example.com/webhook"]')).toBeVisible();
      await expect(page.locator('select')).toBeVisible(); // Content type select
      await expect(page.locator('input[placeholder="Secret for webhook validation"]')).toBeVisible();
    });

    test('should display webhook event options', async ({ page }) => {
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      await page.click('button:has-text("Add Webhook")');
      
      // Should show event checkboxes
      const expectedEvents = [
        'push', 'pull_request', 'issues', 'issue_comment', 'create', 'delete',
        'fork', 'star', 'watch', 'release', 'pull_request_review', 'pull_request_review_comment'
      ];
      
      for (const event of expectedEvents) {
        await expect(page.locator(`text=${event}`)).toBeVisible();
        await expect(page.locator(`input[type="checkbox"]`).locator(`text=${event}`).locator('..')).toBeVisible();
      }
      
      // Push should be checked by default
      const pushCheckbox = page.locator('text=push').locator('..').locator('input[type="checkbox"]');
      await expect(pushCheckbox).toBeChecked();
    });

    test('should validate webhook form fields', async ({ page }) => {
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      await page.click('button:has-text("Add Webhook")');
      
      // Create button should be disabled without required fields
      const createButton = page.locator('button:has-text("Create Webhook")');
      await expect(createButton).toBeDisabled();
      
      // Fill name only
      await page.fill('input[placeholder="Webhook name"]', 'Test Webhook');
      await expect(createButton).toBeDisabled();
      
      // Fill URL as well
      await page.fill('input[placeholder="https://example.com/webhook"]', 'https://example.com/webhook');
      await expect(createButton).not.toBeDisabled();
    });

    test('should create new webhook', async ({ page }) => {
      // Mock create webhook API
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        if (route.request().method() === 'POST') {
          const requestData = await route.request().postDataJSON();
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              id: '3',
              name: requestData.name,
              config: requestData.config,
              events: requestData.events,
              active: requestData.active,
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString()
            })
          });
        } else {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([])
          });
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      await page.click('button:has-text("Add Webhook")');
      
      // Fill webhook form
      await page.fill('input[placeholder="Webhook name"]', 'New Webhook');
      await page.fill('input[placeholder="https://example.com/webhook"]', 'https://test.example.com/hook');
      await page.fill('input[placeholder="Secret for webhook validation"]', 'my-secret');
      
      // Select content type
      await page.selectOption('select', 'form');
      
      // Select additional events
      await page.check('text=pull_request');
      await page.check('text=issues');
      
      // Ensure active is checked
      await page.check('input[type="checkbox"]#active');
      
      // Submit form
      await page.click('button:has-text("Create Webhook")');
      
      // Should show creating state
      await expect(page.locator('button:has-text("Creating...")')).toBeVisible();
    });

    test('should handle webhook creation errors', async ({ page }) => {
      // Mock error response
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 400,
            contentType: 'application/json',
            body: JSON.stringify({
              success: false,
              error: 'Invalid webhook URL'
            })
          });
        } else {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify([])
          });
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      await page.click('button:has-text("Add Webhook")');
      
      // Fill and submit form
      await page.fill('input[placeholder="Webhook name"]', 'Test Webhook');
      await page.fill('input[placeholder="https://example.com/webhook"]', 'invalid-url');
      await page.click('button:has-text("Create Webhook")');
      
      // Should show error message
      await expect(page.locator('text=Invalid webhook URL')).toBeVisible();
    });
  });

  test.describe('Webhook Operations', () => {
    test.beforeEach(async ({ page }) => {
      // Mock webhooks list
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(mockWebhooks)
          });
        }
      });
    });

    test('should ping webhook', async ({ page }) => {
      // Mock ping webhook API
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks/1/pings', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Mock alert
      page.on('dialog', async dialog => {
        expect(dialog.message()).toContain('Ping sent successfully');
        await dialog.accept();
      });
      
      // Click ping button on first webhook
      await page.locator('text=CI/CD Webhook').locator('..').locator('button:has-text("Ping")').click();
    });

    test('should toggle webhook status', async ({ page }) => {
      // Mock toggle webhook API
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks/1', async route => {
        if (route.request().method() === 'PATCH') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              ...mockWebhooks[0],
              active: false
            })
          });
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Click disable on active webhook
      await page.locator('text=CI/CD Webhook').locator('..').locator('button:has-text("Disable")').click();
    });

    test('should delete webhook with confirmation', async ({ page }) => {
      // Mock delete webhook API
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks/1', async route => {
        if (route.request().method() === 'DELETE') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Mock confirmation dialog
      page.on('dialog', async dialog => {
        expect(dialog.message()).toContain('Are you sure you want to delete this webhook');
        await dialog.accept();
      });
      
      // Click delete on first webhook
      await page.locator('text=CI/CD Webhook').locator('..').locator('button:has-text("Delete")').click();
    });

    test('should open edit webhook modal', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Click edit on first webhook
      await page.locator('text=CI/CD Webhook').locator('..').locator('button:has-text("Edit")').click();
      
      // Should open edit functionality (exact behavior depends on implementation)
      // This could be a modal or navigation to edit page
    });
  });

  test.describe('Webhook Configuration Options', () => {
    test('should configure content type options', async ({ page }) => {
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      await page.click('button:has-text("Add Webhook")');
      
      // Should show content type options
      const select = page.locator('select');
      await expect(select.locator('option[value="json"]')).toContainText('application/json');
      await expect(select.locator('option[value="form"]')).toContainText('application/x-www-form-urlencoded');
      
      // Test changing content type
      await page.selectOption('select', 'form');
      await expect(select).toHaveValue('form');
    });

    test('should configure webhook events', async ({ page }) => {
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      await page.click('button:has-text("Add Webhook")');
      
      // Test event selection
      const pushCheckbox = page.locator('text=push').locator('..').locator('input[type="checkbox"]');
      const issuesCheckbox = page.locator('text=issues').locator('..').locator('input[type="checkbox"]');
      
      // Push should be checked by default
      await expect(pushCheckbox).toBeChecked();
      await expect(issuesCheckbox).not.toBeChecked();
      
      // Check issues
      await issuesCheckbox.check();
      await expect(issuesCheckbox).toBeChecked();
      
      // Uncheck push
      await pushCheckbox.uncheck();
      await expect(pushCheckbox).not.toBeChecked();
    });

    test('should configure webhook secret', async ({ page }) => {
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      await page.click('button:has-text("Add Webhook")');
      
      // Should have secret field
      const secretInput = page.locator('input[placeholder="Secret for webhook validation"]');
      await expect(secretInput).toBeVisible();
      await expect(secretInput).toHaveAttribute('type', 'password');
      
      // Test filling secret
      await secretInput.fill('my-webhook-secret');
      await expect(secretInput).toHaveValue('my-webhook-secret');
    });
  });

  test.describe('Integration with Repository Settings', () => {
    test('should access webhooks from repository settings page', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Navigate to webhooks tab
      await page.click('button:has-text("Webhooks")');
      
      // Should show webhooks overview in settings
      await expect(page.locator('h3')).toContainText('Webhooks');
      await expect(page.locator('text=Webhooks allow external services to be notified')).toBeVisible();
      
      // Should have links to detailed webhook management
      await expect(page.locator('a[href="/repositories/testuser/test-repo/settings/webhooks"]')).toBeVisible();
      await expect(page.locator('button:has-text("Manage Webhooks")')).toBeVisible();
      await expect(page.locator('button:has-text("Go to Webhooks")')).toBeVisible();
    });

    test('should navigate to webhooks page from settings', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Webhooks")');
      
      // Click on webhook management link
      await page.click('a[href="/repositories/testuser/test-repo/settings/webhooks"]');
      
      // Should navigate to webhooks page
      await expect(page).toHaveURL('/repositories/testuser/test-repo/settings/webhooks');
    });
  });

  test.describe('Error Handling & Edge Cases', () => {
    test('should handle webhooks API errors gracefully', async ({ page }) => {
      // Mock error response
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            error: 'Internal server error'
          })
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Should show error and retry option
      await expect(page.locator('text=Error: Internal server error')).toBeVisible();
      await expect(page.locator('button:has-text("Try Again")')).toBeVisible();
    });

    test('should handle network failures', async ({ page }) => {
      // Simulate network failure
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.abort('internetdisconnected');
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Should show loading state or error handling
      await expect(page.locator('.animate-pulse')).toBeVisible();
    });

    test('should handle ping webhook errors', async ({ page }) => {
      // Mock successful webhook list
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockWebhooks)
        });
      });

      // Mock ping error
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks/1/pings', async route => {
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            error: 'Webhook URL is unreachable'
          })
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Click ping button
      await page.locator('text=CI/CD Webhook').locator('..').locator('button:has-text("Ping")').click();
      
      // Should handle error appropriately (exact behavior depends on implementation)
    });
  });

  test.describe('Mobile Webhook Management', () => {
    test('should be responsive on mobile devices', async ({ page }) => {
      // Set mobile viewport
      await page.setViewportSize({ width: 375, height: 667 });
      
      // Mock webhooks for mobile test
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockWebhooks)
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      
      // Should display mobile-friendly layout
      await expect(page.locator('h1')).toContainText('Webhooks');
      await expect(page.locator('button:has-text("Add Webhook")')).toBeVisible();
      
      // Test modal on mobile
      await page.click('button:has-text("Add Webhook")');
      await expect(page.locator('text=Add Webhook')).toBeVisible();
    });
  });

  test.describe('Webhook Security Validation', () => {
    test('should validate webhook URL format', async ({ page }) => {
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      await page.click('button:has-text("Add Webhook")');
      
      // Test valid URLs
      const validUrls = [
        'https://example.com/webhook',
        'http://localhost:3000/hook',
        'https://api.example.com/v1/webhooks/github'
      ];

      // Fill name first
      await page.fill('input[placeholder="Webhook name"]', 'Test Webhook');
      
      for (const url of validUrls) {
        await page.fill('input[placeholder="https://example.com/webhook"]', url);
        // Create button should not be disabled for valid URLs
        await expect(page.locator('button:has-text("Create Webhook")')).not.toBeDisabled();
      }
    });

    test('should provide security guidance for webhook secrets', async ({ page }) => {
      await page.route('**/api/v1/repositories/testuser/test-repo/hooks', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/webhooks');
      await page.click('button:has-text("Add Webhook")');
      
      // Secret field should be password type for security
      const secretInput = page.locator('input[placeholder="Secret for webhook validation"]');
      await expect(secretInput).toHaveAttribute('type', 'password');
      
      // Should be optional but recommended for security
      await expect(page.locator('text=Secret (optional)')).toBeVisible();
    });
  });
});