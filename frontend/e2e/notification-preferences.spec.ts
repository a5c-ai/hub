import { test, expect } from '@playwright/test';
import { loginUser, testUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('Notification Preferences & Settings', () => {
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
            avatar_url: 'https://example.com/avatar.jpg',
            created_at: new Date().toISOString()
          }
        })
      });
    });
  });

  test.describe('Email Notification Settings', () => {
    test('should display and manage email notification preferences', async ({ page }) => {
      const mockPreferences = {
        email_notifications: {
          issues_and_prs: true,
          repository_updates: true,
          security_alerts: true,
          mentions: true,
          comments: false,
          team_notifications: true
        },
        web_notifications: {
          browser_notifications: false,
          desktop_notifications: true,
          mobile_push: false
        },
        delivery_settings: {
          frequency: 'immediate',
          quiet_hours: {
            enabled: true,
            start: '22:00',
            end: '08:00'
          },
          digest_enabled: false,
          digest_frequency: 'weekly'
        }
      };

      await page.route('**/api/v1/user/notification-preferences', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(mockPreferences)
          });
        } else if (route.request().method() === 'PUT') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await loginUser(page);
      await page.goto('/settings');
      await waitForLoadingToComplete(page);

      // Navigate to notifications tab
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Check email notification settings section
      await expect(page.locator('h3')).toContainText('Email Notifications');

      // Verify current settings are displayed correctly
      await expect(page.locator('[data-testid="email-issues-prs"]')).toBeChecked();
      await expect(page.locator('[data-testid="email-repository-updates"]')).toBeChecked();
      await expect(page.locator('[data-testid="email-security-alerts"]')).toBeChecked();

      // Test toggling a setting
      await page.uncheck('[data-testid="email-issues-prs"]');
      
      // Save settings
      await page.click('[data-testid="save-notification-settings"]');
      
      // Verify success message
      await expect(page.locator('[data-testid="settings-success-message"]')).toBeVisible();
    });

    test('should configure notification delivery timing', async ({ page }) => {
      await page.route('**/api/v1/user/notification-preferences', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            email_notifications: {},
            web_notifications: {},
            delivery_settings: {
              frequency: 'immediate',
              quiet_hours: { enabled: false },
              digest_enabled: false
            }
          })
        });
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Test delivery frequency settings
      await expect(page.locator('[data-testid="delivery-frequency"]')).toBeVisible();
      
      // Change to batched delivery
      await page.selectOption('[data-testid="delivery-frequency"]', 'hourly');
      
      // Enable quiet hours
      await page.check('[data-testid="quiet-hours-enabled"]');
      await page.fill('[data-testid="quiet-hours-start"]', '23:00');
      await page.fill('[data-testid="quiet-hours-end"]', '07:00');
      
      // Enable email digest
      await page.check('[data-testid="digest-enabled"]');
      await page.selectOption('[data-testid="digest-frequency"]', 'daily');
      
      await page.click('[data-testid="save-notification-settings"]');
      await expect(page.locator('[data-testid="settings-success-message"]')).toBeVisible();
    });

    test('should manage thread subscription settings', async ({ page }) => {
      const threadSubscriptions = [
        {
          id: '1',
          type: 'issue',
          title: 'Critical bug in authentication',
          repository: 'team/auth-service',
          url: '/repositories/team/auth-service/issues/123',
          subscribed: true,
          reason: 'mentioned'
        },
        {
          id: '2',
          type: 'pull_request',
          title: 'Add new user dashboard',
          repository: 'company/frontend-app',
          url: '/repositories/company/frontend-app/pulls/456',
          subscribed: true,
          reason: 'author'
        }
      ];

      await page.route('**/api/v1/user/thread-subscriptions', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(threadSubscriptions)
          });
        }
      });

      await page.route('**/api/v1/user/thread-subscriptions/*', async route => {
        if (route.request().method() === 'DELETE') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Check thread subscriptions section
      await expect(page.locator('h4')).toContainText('Thread Subscriptions');
      
      // Verify subscriptions are listed
      await expect(page.locator('[data-testid="thread-subscription"]')).toHaveCount(2);
      await expect(page.locator('text=Critical bug in authentication')).toBeVisible();
      await expect(page.locator('text=Add new user dashboard')).toBeVisible();
      
      // Test unsubscribing from a thread
      await page.click('[data-testid="unsubscribe-thread-1"]');
      await expect(page.locator('[data-testid="thread-subscription"]')).toHaveCount(1);
    });
  });

  test.describe('Web Notification Settings', () => {
    test('should manage browser notification permissions', async ({ page }) => {
      // Mock Notification API
      await page.addInitScript(() => {
        Object.defineProperty(window, 'Notification', {
          writable: true,
          value: class MockNotification {
            static permission = 'default';
            static requestPermission = async () => 'granted';
            title: string;
            body?: string;
            constructor(title: string, options?: any) {
              this.title = title;
              this.body = options?.body;
            }
          }
        });
      });

      await page.route('**/api/v1/user/notification-preferences', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            web_notifications: {
              browser_notifications: false,
              desktop_notifications: false
            }
          })
        });
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Check web notifications section
      await expect(page.locator('h3')).toContainText('Web Notifications');
      
      // Test enabling browser notifications
      const enableButton = page.locator('[data-testid="enable-browser-notifications"]');
      await expect(enableButton).toBeVisible();
      await enableButton.click();
      
      // Should request permission and update settings
      await expect(page.locator('[data-testid="browser-notifications-enabled"]')).toBeVisible();
    });

    test('should configure desktop notification settings', async ({ page }) => {
      await page.route('**/api/v1/user/notification-preferences', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            web_notifications: {
              browser_notifications: true,
              desktop_notifications: true,
              notification_types: {
                mentions: true,
                issues: false,
                pull_requests: true,
                security_alerts: true
              }
            }
          })
        });
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Check current desktop notification settings
      await expect(page.locator('[data-testid="desktop-mentions"]')).toBeChecked();
      await expect(page.locator('[data-testid="desktop-issues"]')).not.toBeChecked();
      await expect(page.locator('[data-testid="desktop-pull-requests"]')).toBeChecked();
      
      // Toggle issue notifications
      await page.check('[data-testid="desktop-issues"]');
      
      await page.click('[data-testid="save-notification-settings"]');
      await expect(page.locator('[data-testid="settings-success-message"]')).toBeVisible();
    });
  });

  test.describe('Notification Cleanup Settings', () => {
    test('should configure automatic notification cleanup', async ({ page }) => {
      await page.route('**/api/v1/user/notification-preferences', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            cleanup_settings: {
              auto_cleanup_enabled: true,
              cleanup_after_days: 30,
              cleanup_read_notifications: true,
              cleanup_archived_threads: false
            }
          })
        });
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Check cleanup settings
      await expect(page.locator('h4')).toContainText('Automatic Cleanup');
      
      await expect(page.locator('[data-testid="auto-cleanup-enabled"]')).toBeChecked();
      await expect(page.locator('[data-testid="cleanup-after-days"]')).toHaveValue('30');
      await expect(page.locator('[data-testid="cleanup-read-notifications"]')).toBeChecked();
      
      // Modify cleanup settings
      await page.fill('[data-testid="cleanup-after-days"]', '60');
      await page.check('[data-testid="cleanup-archived-threads"]');
      
      await page.click('[data-testid="save-notification-settings"]');
      await expect(page.locator('[data-testid="settings-success-message"]')).toBeVisible();
    });

    test('should manually trigger notification cleanup', async ({ page }) => {
      await page.route('**/api/v1/user/notifications/cleanup', async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              cleaned_count: 145,
              message: 'Cleaned up 145 old notifications'
            })
          });
        }
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Test manual cleanup
      await page.click('[data-testid="manual-cleanup-button"]');
      
      // Should show confirmation dialog
      await expect(page.locator('[data-testid="cleanup-confirmation"]')).toBeVisible();
      await page.click('[data-testid="confirm-cleanup"]');
      
      // Should show success message with count
      await expect(page.locator('text=Cleaned up 145 old notifications')).toBeVisible();
    });
  });

  test.describe('Repository-Specific Settings', () => {
    test('should manage repository watching and notification settings', async ({ page }) => {
      const watchedRepos = [
        {
          id: '1',
          name: 'critical-service',
          full_name: 'company/critical-service',
          watching: true,
          notification_types: {
            issues: true,
            pull_requests: true,
            releases: false,
            discussions: false
          }
        },
        {
          id: '2',
          name: 'documentation',
          full_name: 'team/documentation',
          watching: false,
          notification_types: {
            issues: false,
            pull_requests: false,
            releases: true,
            discussions: true
          }
        }
      ];

      await page.route('**/api/v1/user/watched-repositories', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(watchedRepos)
          });
        }
      });

      await page.route('**/api/v1/repositories/*/watch', async route => {
        if (route.request().method() === 'PUT') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Check repository watching section
      await expect(page.locator('h4')).toContainText('Repository Notifications');
      
      // Verify watched repositories
      await expect(page.locator('[data-testid="watched-repo"]')).toHaveCount(2);
      await expect(page.locator('text=company/critical-service')).toBeVisible();
      await expect(page.locator('text=team/documentation')).toBeVisible();
      
      // Test modifying notification types for a repo
      const criticalRepo = page.locator('[data-testid="watched-repo-1"]');
      await expect(criticalRepo.locator('[data-testid="repo-issues"]')).toBeChecked();
      await expect(criticalRepo.locator('[data-testid="repo-releases"]')).not.toBeChecked();
      
      // Enable release notifications
      await criticalRepo.locator('[data-testid="repo-releases"]').check();
      
      await page.click('[data-testid="save-repo-settings-1"]');
      await expect(page.locator('[data-testid="settings-success-message"]')).toBeVisible();
    });

    test('should manage organization notification preferences', async ({ page }) => {
      const organizations = [
        {
          id: '1',
          name: 'my-company',
          display_name: 'My Company',
          notification_settings: {
            team_mentions: true,
            repository_invitations: true,
            security_alerts: true,
            member_activity: false
          }
        },
        {
          id: '2',
          name: 'open-source-org',
          display_name: 'Open Source Organization',
          notification_settings: {
            team_mentions: false,
            repository_invitations: false,
            security_alerts: true,
            member_activity: false
          }
        }
      ];

      await page.route('**/api/v1/user/organization-notifications', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(organizations)
          });
        }
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Check organization notifications section
      await expect(page.locator('h4')).toContainText('Organization Notifications');
      
      // Verify organizations are listed
      await expect(page.locator('[data-testid="org-notification-settings"]')).toHaveCount(2);
      await expect(page.locator('text=My Company')).toBeVisible();
      await expect(page.locator('text=Open Source Organization')).toBeVisible();
      
      // Test modifying organization settings
      const myCompany = page.locator('[data-testid="org-settings-1"]');
      await expect(myCompany.locator('[data-testid="org-team-mentions"]')).toBeChecked();
      await expect(myCompany.locator('[data-testid="org-member-activity"]')).not.toBeChecked();
      
      // Enable member activity notifications
      await myCompany.locator('[data-testid="org-member-activity"]').check();
      
      await page.click('[data-testid="save-org-settings-1"]');
      await expect(page.locator('[data-testid="settings-success-message"]')).toBeVisible();
    });
  });

  test.describe('Import/Export Settings', () => {
    test('should export notification preferences', async ({ page }) => {
      await page.route('**/api/v1/user/notification-preferences/export', async route => {
        const exportData = {
          version: '1.0',
          exported_at: new Date().toISOString(),
          preferences: {
            email_notifications: { issues_and_prs: true },
            web_notifications: { browser_notifications: false },
            delivery_settings: { frequency: 'immediate' }
          }
        };
        
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(exportData)
        });
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Test export functionality
      const downloadPromise = page.waitForEvent('download');
      await page.click('[data-testid="export-preferences"]');
      
      const download = await downloadPromise;
      expect(download.suggestedFilename()).toMatch(/notification-preferences.*\.json/);
    });

    test('should import notification preferences', async ({ page }) => {
      await page.route('**/api/v1/user/notification-preferences/import', async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              imported_settings: 15,
              message: 'Successfully imported 15 notification settings'
            })
          });
        }
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Test import functionality
      const fileInput = page.locator('[data-testid="import-preferences-file"]');
      await fileInput.setInputFiles({
        name: 'preferences.json',
        mimeType: 'application/json',
        buffer: Buffer.from(JSON.stringify({
          version: '1.0',
          preferences: {
            email_notifications: { issues_and_prs: false }
          }
        }))
      });
      
      await page.click('[data-testid="import-preferences"]');
      
      // Should show import confirmation
      await expect(page.locator('[data-testid="import-confirmation"]')).toBeVisible();
      await page.click('[data-testid="confirm-import"]');
      
      // Should show success message
      await expect(page.locator('text=Successfully imported 15 notification settings')).toBeVisible();
    });
  });

  test.describe('Mobile Responsiveness', () => {
    test('should provide mobile-friendly notification settings interface', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });

      await page.route('**/api/v1/user/notification-preferences', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            email_notifications: {
              issues_and_prs: true,
              security_alerts: true
            }
          })
        });
      });

      await loginUser(page);
      await page.goto('/settings');
      await waitForLoadingToComplete(page);

      // Test mobile navigation to notifications
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Verify mobile layout
      await expect(page.locator('h3')).toContainText('Email Notifications');
      
      // Test mobile-friendly controls
      const checkbox = page.locator('[data-testid="email-issues-prs"]');
      await expect(checkbox).toBeVisible();
      
      // Verify touch targets are appropriately sized
      const checkboxBox = await checkbox.boundingBox();
      expect(checkboxBox!.height).toBeGreaterThan(40);
      
      // Test mobile save action
      await page.click('[data-testid="save-notification-settings"]');
      await expect(page.locator('[data-testid="settings-success-message"]')).toBeVisible();
    });

    test('should handle mobile swipe gestures for settings navigation', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });

      await loginUser(page);
      await page.goto('/settings');
      
      // Test swipe navigation between settings tabs (if implemented)
      // This assumes the settings interface supports swipe navigation
      const settingsContainer = page.locator('[data-testid="settings-container"]');
      
      if (await settingsContainer.isVisible()) {
        // Simulate swipe gesture
        await settingsContainer.dragTo(settingsContainer, {
          sourcePosition: { x: 200, y: 100 },
          targetPosition: { x: 50, y: 100 }
        });
        
        // Should navigate to next settings tab
        await waitForLoadingToComplete(page);
      }
    });
  });

  test.describe('Error Handling', () => {
    test('should handle API errors when saving preferences', async ({ page }) => {
      await page.route('**/api/v1/user/notification-preferences', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ email_notifications: {} })
          });
        } else if (route.request().method() === 'PUT') {
          await route.fulfill({
            status: 500,
            contentType: 'application/json',
            body: JSON.stringify({ error: 'Failed to save preferences' })
          });
        }
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Modify a setting and try to save
      await page.check('[data-testid="email-issues-prs"]');
      await page.click('[data-testid="save-notification-settings"]');
      
      // Should show error message
      await expect(page.locator('[data-testid="settings-error-message"]')).toBeVisible();
      await expect(page.locator('text=Failed to save preferences')).toBeVisible();
    });

    test('should validate notification preference values', async ({ page }) => {
      await page.route('**/api/v1/user/notification-preferences', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            delivery_settings: {
              cleanup_after_days: 30
            }
          })
        });
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Test invalid input validation
      await page.fill('[data-testid="cleanup-after-days"]', '-5');
      await page.click('[data-testid="save-notification-settings"]');
      
      // Should show validation error
      await expect(page.locator('[data-testid="validation-error"]')).toBeVisible();
      
      // Test valid input
      await page.fill('[data-testid="cleanup-after-days"]', '45');
      await page.click('[data-testid="save-notification-settings"]');
      
      // Should succeed
      await expect(page.locator('[data-testid="settings-success-message"]')).toBeVisible();
    });
  });

  test.describe('Performance & Accessibility', () => {
    test('should be accessible to screen readers', async ({ page }) => {
      await page.route('**/api/v1/user/notification-preferences', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ email_notifications: {} })
        });
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Check for proper ARIA labels and roles
      await expect(page.locator('[data-testid="email-issues-prs"]')).toHaveAttribute('aria-label');
      await expect(page.locator('h3')).toHaveRole('heading');
      
      // Test keyboard navigation
      await page.keyboard.press('Tab');
      await page.keyboard.press('Space'); // Should toggle checkbox
      
      // Verify focus management
      const focusedElement = page.locator(':focus');
      await expect(focusedElement).toBeVisible();
    });

    test('should load settings efficiently with minimal API calls', async ({ page }) => {
      let apiCallCount = 0;
      
      await page.route('**/api/v1/user/notification-preferences', async route => {
        apiCallCount++;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ email_notifications: {} })
        });
      });

      await loginUser(page);
      await page.goto('/settings');
      await page.click('[data-testid="settings-tab-notifications"]');
      await waitForLoadingToComplete(page);

      // Should only make one API call for initial load
      expect(apiCallCount).toBe(1);
      
      // Navigating between sections shouldn't trigger additional calls
      await page.click('[data-testid="settings-tab-profile"]');
      await page.click('[data-testid="settings-tab-notifications"]');
      
      expect(apiCallCount).toBe(1); // Still only one call
    });
  });
});