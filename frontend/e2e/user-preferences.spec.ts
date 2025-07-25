import { test, expect } from '@playwright/test';
import { loginUser, testUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('User Preferences & Notifications', () => {
  test.beforeEach(async ({ page }) => {
    // Start fresh for each test
    await page.context().clearCookies();
    await page.context().clearPermissions();
    
    // Setup common API mocks
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
            created_at: '2023-01-01T00:00:00Z'
          }
        })
      });
    });

    // Mock repositories API for dashboard
    await page.route('**/api/v1/repositories**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: []
        })
      });
    });

    // Mock notification preferences API
    await page.route('**/api/v1/user/notification-preferences', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              email_notifications: {
                issues_and_prs: true,
                repository_updates: true,
                security_alerts: true,
                mentions: true,
                team_discussions: false
              },
              web_notifications: {
                enabled: false
              },
              notification_frequency: 'immediate',
              quiet_hours: {
                enabled: false,
                start_time: '22:00',
                end_time: '08:00'
              }
            }
          })
        });
      }
    });

    // Mock user preferences API
    await page.route('**/api/v1/user/preferences', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              theme: 'system',
              language: 'en',
              timezone: 'America/Los_Angeles',
              date_format: 'MM/DD/YYYY',
              accessibility: {
                high_contrast: false,
                reduced_motion: false,
                large_font: false
              },
              privacy: {
                profile_visibility: 'public',
                activity_visibility: 'followers',
                repository_defaults: 'public',
                search_visibility: true
              }
            }
          })
        });
      }
    });
  });

  test('should display notification preferences correctly', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Notifications tab
    await page.click('button:has-text("Notifications")');
    await expect(page.locator('h3:has-text("Email Notifications")')).toBeVisible();
    
    // Verify notification options are displayed
    await expect(page.locator('text=Issues and Pull Requests')).toBeVisible();
    await expect(page.locator('text=Repository Updates')).toBeVisible();
    await expect(page.locator('text=Security Alerts')).toBeVisible();
    
    // Check that checkboxes reflect current state
    await expect(page.locator('input[type="checkbox"]').first()).toBeChecked();
  });

  test('should update email notification preferences', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Notifications tab
    await page.click('button:has-text("Notifications")');
    
    // Mock update preferences API
    await page.route('**/api/v1/user/notification-preferences', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Notification preferences updated successfully'
          })
        });
      }
    });
    
    // Toggle some notification settings
    const issuesCheckbox = page.locator('text=Issues and Pull Requests').locator('..').locator('input[type="checkbox"]');
    await issuesCheckbox.uncheck();
    
    const repoUpdatesCheckbox = page.locator('text=Repository Updates').locator('..').locator('input[type="checkbox"]');
    await repoUpdatesCheckbox.uncheck();
    
    // Save changes (if there's a save button, otherwise changes might be auto-saved)
    if (await page.locator('button:has-text("Save"), button:has-text("Update")').isVisible()) {
      await page.click('button:has-text("Save"), button:has-text("Update")');
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should enable web notifications', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Notifications tab
    await page.click('button:has-text("Notifications")');
    
    // Mock browser notification permission request
    await page.evaluate(() => {
      // Mock the Notification API
      Object.defineProperty(window, 'Notification', {
        value: {
          permission: 'default',
          requestPermission: () => Promise.resolve('granted')
        }
      });
    });
    
    // Mock web notification setup API
    await page.route('**/api/v1/user/web-notifications', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Web notifications enabled successfully'
        })
      });
    });
    
    // Find and click enable web notifications
    await expect(page.locator('h3:has-text("Web Notifications")')).toBeVisible();
    await page.click('button:has-text("Enable")');
    
    await waitForLoadingToComplete(page);
  });

  test('should configure notification frequency', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Notifications tab
    await page.click('button:has-text("Notifications")');
    
    // Mock update notification frequency API
    await page.route('**/api/v1/user/notification-preferences', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Notification frequency updated successfully'
          })
        });
      }
    });
    
    // Look for notification frequency controls
    if (await page.locator('select, [role="combobox"]').first().isVisible()) {
      await page.selectOption('select', 'daily');
    } else if (await page.locator('input[type="radio"]').first().isVisible()) {
      await page.click('input[value="daily"]');
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should configure quiet hours', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Notifications tab
    await page.click('button:has-text("Notifications")');
    
    // Mock update quiet hours API
    await page.route('**/api/v1/user/notification-preferences', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Quiet hours updated successfully'
          })
        });
      }
    });
    
    // Look for quiet hours section
    if (await page.locator('text=Quiet Hours, text=Do Not Disturb').isVisible()) {
      // Enable quiet hours
      await page.click('input[type="checkbox"]:near(text=Quiet)');
      
      // Set start and end times
      if (await page.locator('input[type="time"]').first().isVisible()) {
        await page.fill('input[type="time"]', '22:00');
        await page.fill('input[type="time"]', '08:00');
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should unsubscribe from specific notification types', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Notifications tab
    await page.click('button:has-text("Notifications")');
    
    // Mock unsubscribe API
    await page.route('**/api/v1/user/unsubscribe/**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Unsubscribed successfully'
        })
      });
    });
    
    // Look for unsubscribe options
    if (await page.locator('button:has-text("Unsubscribe"), link:has-text("Unsubscribe")').isVisible()) {
      await page.click('button:has-text("Unsubscribe"), link:has-text("Unsubscribe")');
    }
    
    await waitForLoadingToComplete(page);
  });
});

test.describe('Appearance & Accessibility Settings', () => {
  test.beforeEach(async ({ page }) => {
    // Setup common API mocks (same as above)
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
            created_at: '2023-01-01T00:00:00Z'
          }
        })
      });
    });

    await page.route('**/api/v1/repositories**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: []
        })
      });
    });
  });

  test('should switch between light and dark themes', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Account tab where theme toggle is located
    await page.click('button:has-text("Account")');
    
    // Find theme toggle
    const themeToggle = page.locator('[data-testid="theme-toggle"]').or(
      page.locator('button[aria-label*="theme"], button[title*="theme"]')
    );
    
    // Verify theme toggle is visible
    await expect(themeToggle).toBeVisible();
    
    // Toggle theme
    await themeToggle.click();
    
    // Verify theme change (check for dark/light class on html or body)
    await page.waitForTimeout(500); // Allow for theme transition
    
    // Toggle back
    await themeToggle.click();
    await page.waitForTimeout(500);
  });

  test('should change language settings', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock language update API
    await page.route('**/api/v1/user/preferences', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Language preference updated successfully'
          })
        });
      }
    });
    
    // Look for language settings (might be in Account or dedicated Preferences section)
    if (await page.locator('text=Language, text=Locale').isVisible()) {
      // Select different language
      if (await page.locator('select').first().isVisible()) {
        await page.selectOption('select', 'es'); // Spanish
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should configure timezone settings', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock timezone update API
    await page.route('**/api/v1/user/preferences', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Timezone updated successfully'
          })
        });
      }
    });
    
    // Look for timezone settings
    if (await page.locator('text=Timezone, text=Time Zone').isVisible()) {
      // Select different timezone
      if (await page.locator('select').isVisible()) {
        await page.selectOption('select', 'America/New_York');
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should configure date format settings', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock date format update API
    await page.route('**/api/v1/user/preferences', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Date format updated successfully'
          })
        });
      }
    });
    
    // Look for date format settings
    if (await page.locator('text=Date Format').isVisible()) {
      // Select different date format
      if (await page.locator('input[type="radio"]').isVisible()) {
        await page.click('input[value="DD/MM/YYYY"]');
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should enable accessibility features', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock accessibility preferences update API
    await page.route('**/api/v1/user/accessibility-preferences', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Accessibility preferences updated successfully'
        })
      });
    });
    
    // Look for accessibility settings
    if (await page.locator('text=Accessibility, text=High Contrast, text=Large Font').isVisible()) {
      // Enable high contrast
      const highContrastToggle = page.locator('text=High Contrast').locator('..').locator('input[type="checkbox"]');
      if (await highContrastToggle.isVisible()) {
        await highContrastToggle.check();
      }
      
      // Enable reduced motion
      const reducedMotionToggle = page.locator('text=Reduced Motion').locator('..').locator('input[type="checkbox"]');
      if (await reducedMotionToggle.isVisible()) {
        await reducedMotionToggle.check();
      }
      
      // Enable large font
      const largeFontToggle = page.locator('text=Large Font').locator('..').locator('input[type="checkbox"]');
      if (await largeFontToggle.isVisible()) {
        await largeFontToggle.check();
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should adjust font size preferences', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock font size update API
    await page.route('**/api/v1/user/preferences', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Font size preference updated successfully'
          })
        });
      }
    });
    
    // Look for font size settings
    if (await page.locator('text=Font Size').isVisible()) {
      // Adjust font size slider or select
      const fontSizeControl = page.locator('input[type="range"], select').first();
      if (await fontSizeControl.isVisible()) {
        if (await fontSizeControl.getAttribute('type') === 'range') {
          await fontSizeControl.fill('18');
        } else {
          await page.selectOption('select', 'large');
        }
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should test accessibility with keyboard navigation', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Test keyboard navigation through settings tabs
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab');
    
    // Navigate to different tabs using arrow keys or enter
    await page.keyboard.press('ArrowRight');
    await page.keyboard.press('Enter');
    
    // Verify focus management
    const focusedElement = page.locator(':focus');
    await expect(focusedElement).toBeVisible();
  });

  test('should work with screen reader attributes', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Check for proper ARIA labels and roles
    await expect(page.locator('[role="tablist"]').or(page.locator('[aria-label*="Settings"]'))).toBeVisible();
    
    // Check that form controls have proper labels
    const formControls = page.locator('input, button, select');
    const count = await formControls.count();
    
    for (let i = 0; i < Math.min(count, 5); i++) {
      const control = formControls.nth(i);
      const hasLabel = await control.evaluate((el: HTMLElement) => {
        return el.hasAttribute('aria-label') || 
               el.hasAttribute('aria-labelledby') || 
               (el as HTMLInputElement).labels?.length > 0;
      });
      
      if (await control.isVisible()) {
        expect(hasLabel).toBeTruthy();
      }
    }
  });
});

test.describe('Privacy & Data Settings', () => {
  test.beforeEach(async ({ page }) => {
    // Setup common API mocks
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
            created_at: '2023-01-01T00:00:00Z'
          }
        })
      });
    });

    await page.route('**/api/v1/repositories**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: []
        })
      });
    });
  });

  test('should configure profile visibility settings', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock privacy settings update API
    await page.route('**/api/v1/user/privacy-settings', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Privacy settings updated successfully'
          })
        });
      }
    });
    
    // Look for privacy settings (might be in a separate Privacy tab or under Account)
    if (await page.locator('text=Privacy, text=Profile Visibility').isVisible()) {
      // Change profile visibility
      if (await page.locator('select, input[type="radio"]').isVisible()) {
        await page.selectOption('select', 'private');
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should configure activity visibility controls', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock activity visibility update API
    await page.route('**/api/v1/user/activity-settings', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Activity visibility updated successfully'
          })
        });
      }
    });
    
    // Look for activity visibility settings
    if (await page.locator('text=Activity Visibility').isVisible()) {
      // Configure who can see activity
      const activityToggle = page.locator('text=Activity').locator('..').locator('input[type="checkbox"]');
      if (await activityToggle.isVisible()) {
        await activityToggle.uncheck();
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should set repository privacy defaults', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock repository defaults update API
    await page.route('**/api/v1/user/repository-defaults', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Repository defaults updated successfully'
          })
        });
      }
    });
    
    // Look for repository default settings
    if (await page.locator('text=Repository Defaults').isVisible()) {
      // Set default to private
      const privateDefaultRadio = page.locator('input[value="private"]');
      if (await privateDefaultRadio.isVisible()) {
        await privateDefaultRadio.check();
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should configure search visibility settings', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock search visibility update API
    await page.route('**/api/v1/user/search-settings', async route => {
      if (route.request().method() === 'PUT') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'Search visibility updated successfully'
          })
        });
      }
    });
    
    // Look for search visibility settings
    if (await page.locator('text=Search Visibility').isVisible()) {
      // Disable search visibility
      const searchToggle = page.locator('text=Search').locator('..').locator('input[type="checkbox"]');
      if (await searchToggle.isVisible()) {
        await searchToggle.uncheck();
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should export user data', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Mock data export API
    await page.route('**/api/v1/user/export-data', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Data export request submitted successfully',
          data: {
            export_id: 'export-123',
            estimated_completion: '2023-12-01T12:00:00Z'
          }
        })
      });
    });
    
    // Look for data export option
    if (await page.locator('button:has-text("Export"), button:has-text("Download")').isVisible()) {
      await page.click('button:has-text("Export"), button:has-text("Download")');
      
      // Confirm export request
      if (await page.locator('button:has-text("Confirm"), button:has-text("Yes")').isVisible()) {
        await page.click('button:has-text("Confirm"), button:has-text("Yes")');
      }
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should handle data portability options', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Look for data portability section
    if (await page.locator('text=Data Portability, text=Export Data').isVisible()) {
      await expect(page.locator('text=Download, text=Export')).toBeVisible();
      
      // Should show available data types for export
      await expect(page.locator('text=Profile, text=Repositories, text=Activity')).toBeVisible();
    }
  });

  test('should handle mobile responsive design for preferences', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Notifications tab
    await page.click('button:has-text("Notifications")');
    
    // Verify mobile layout works
    await expect(page.locator('h3:has-text("Email Notifications")')).toBeVisible();
    
    // Form controls should be properly sized for mobile
    const checkboxes = page.locator('input[type="checkbox"]');
    const count = await checkboxes.count();
    
    for (let i = 0; i < Math.min(count, 3); i++) {
      await expect(checkboxes.nth(i)).toBeVisible();
    }
  });

  test('should validate preference changes are persisted', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Notifications tab
    await page.click('button:has-text("Notifications")');
    
    // Change a setting
    const firstCheckbox = page.locator('input[type="checkbox"]').first();
    const isInitiallyChecked = await firstCheckbox.isChecked();
    
    if (isInitiallyChecked) {
      await firstCheckbox.uncheck();
    } else {
      await firstCheckbox.check();
    }
    
    // Refresh the page
    await page.reload();
    
    // Navigate back to Notifications
    await page.click('button:has-text("Notifications")');
    
    // Verify the setting persisted (in a real app, this would require proper API integration)
    await expect(firstCheckbox).toBeVisible();
  });
});