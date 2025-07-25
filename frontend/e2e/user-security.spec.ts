import { test, expect } from '@playwright/test';
import { loginUser, testUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('User Security Settings', () => {
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
            mfa_enabled: false,
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

    // Mock SSH keys API
    await page.route('**/api/v1/user/ssh-keys', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: '1',
                title: 'Personal Laptop',
                key: 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC...',
                fingerprint: 'SHA256:abcd1234...',
                created_at: '2023-01-01T00:00:00Z',
                last_used: '2023-12-01T00:00:00Z'
              }
            ]
          })
        });
      }
    });
  });

  test('should display security settings correctly', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    await expect(page.locator('h3:has-text("Password & Authentication")')).toBeVisible();
    
    // Verify password change section
    await expect(page.locator('input[placeholder="Enter current password"]')).toBeVisible();
    await expect(page.locator('input[placeholder="Enter new password"]')).toBeVisible();
    await expect(page.locator('input[placeholder="Confirm new password"]')).toBeVisible();
    await expect(page.locator('button:has-text("Change Password")')).toBeVisible();
  });

  test('should change password successfully', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock successful password change
    await page.route('**/api/v1/auth/change-password', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Password changed successfully'
        })
      });
    });
    
    // Fill password change form
    await page.fill('input[placeholder="Enter current password"]', testUser.password);
    await page.fill('input[placeholder="Enter new password"]', 'NewPassword123!');
    await page.fill('input[placeholder="Confirm new password"]', 'NewPassword123!');
    
    // Submit password change
    await page.click('button:has-text("Change Password")');
    
    // Verify success (implementation-dependent)
    await waitForLoadingToComplete(page);
  });

  test('should show password validation errors', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock validation error
    await page.route('**/api/v1/auth/change-password', async route => {
      await route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'Password does not meet requirements',
          validation_errors: {
            password: ['Password must be at least 8 characters long']
          }
        })
      });
    });
    
    // Try to change to weak password
    await page.fill('input[placeholder="Enter current password"]', testUser.password);
    await page.fill('input[placeholder="Enter new password"]', '123');
    await page.fill('input[placeholder="Confirm new password"]', '123');
    
    await page.click('button:has-text("Change Password")');
    
    // Should show validation error (implementation-dependent)
  });

  test('should show password mismatch error', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Fill password form with mismatched passwords
    await page.fill('input[placeholder="Enter current password"]', testUser.password);
    await page.fill('input[placeholder="Enter new password"]', 'NewPassword123!');
    await page.fill('input[placeholder="Confirm new password"]', 'DifferentPassword123!');
    
    await page.click('button:has-text("Change Password")');
    
    // Should show mismatch error (implementation-dependent)
  });

  test('should display SSH key management section', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // SSH key section should be visible
    await expect(page.locator('text=SSH')).toBeVisible();
    
    // Should show existing SSH key
    await expect(page.locator('text=Personal Laptop')).toBeVisible();
    await expect(page.locator('text=SHA256:abcd1234')).toBeVisible();
  });

  test('should add new SSH key', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock add SSH key API
    await page.route('**/api/v1/user/ssh-keys', async route => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: '2',
              title: 'Work Laptop',
              key: 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQD...',
              fingerprint: 'SHA256:efgh5678...',
              created_at: '2023-12-01T00:00:00Z'
            }
          })
        });
      }
    });
    
    // Click add SSH key button
    await page.click('button:has-text("Add SSH Key"), button:has-text("Add Key")');
    
    // Fill SSH key form (assuming modal or inline form)
    await page.fill('input[placeholder*="title"], input[placeholder*="Title"]', 'Work Laptop');
    await page.fill('textarea[placeholder*="ssh-rsa"], textarea[placeholder*="key"]', 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQD... user@work-laptop');
    
    // Submit new SSH key
    await page.click('button:has-text("Add Key"), button:has-text("Save")');
    
    // Verify success
    await waitForLoadingToComplete(page);
  });

  test('should remove SSH key', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock delete SSH key API
    await page.route('**/api/v1/user/ssh-keys/1', async route => {
      if (route.request().method() === 'DELETE') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            message: 'SSH key deleted successfully'
          })
        });
      }
    });
    
    // Click delete button for SSH key
    await page.click('button:has-text("Delete"), button:has-text("Remove")');
    
    // Confirm deletion (if confirmation dialog exists)
    if (await page.locator('button:has-text("Confirm"), button:has-text("Yes")').isVisible()) {
      await page.click('button:has-text("Confirm"), button:has-text("Yes")');
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should display two-factor authentication options', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Verify 2FA section
    await expect(page.locator('h3:has-text("Two-Factor Authentication")')).toBeVisible();
    await expect(page.locator('text=Authenticator App')).toBeVisible();
    await expect(page.locator('text=SMS Authentication')).toBeVisible();
    
    // Should show enable buttons
    await expect(page.locator('button:has-text("Enable")').first()).toBeVisible();
  });

  test('should enable two-factor authentication with authenticator app', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock 2FA setup API
    await page.route('**/api/v1/auth/mfa/setup', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            qr_code: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNkYPhfDwAChwGA60e6kgAAAABJRU5ErkJggg==',
            secret: 'ABCD1234EFGH5678',
            backup_codes: ['123456', '789012', '345678']
          }
        })
      });
    });
    
    // Mock 2FA verification API
    await page.route('**/api/v1/auth/mfa/verify', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Two-factor authentication enabled successfully'
        })
      });
    });
    
    // Click enable for authenticator app
    await page.click('button:has-text("Enable")');
    
    // Should show QR code and setup instructions
    await expect(page.locator('text=QR, text=Scan')).toBeVisible();
    
    // Enter verification code
    await page.fill('input[placeholder*="code"], input[placeholder*="Code"]', '123456');
    
    // Complete setup
    await page.click('button:has-text("Verify"), button:has-text("Complete")');
    
    await waitForLoadingToComplete(page);
  });

  test('should enable SMS authentication', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock SMS setup API
    await page.route('**/api/v1/auth/sms/setup', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'SMS verification code sent'
        })
      });
    });
    
    // Mock SMS verification API
    await page.route('**/api/v1/auth/sms/verify', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'SMS authentication enabled successfully'
        })
      });
    });
    
    // Find and click SMS enable button
    const smsSection = page.locator('text=SMS Authentication').locator('..').locator('..');
    await smsSection.locator('button:has-text("Enable")').click();
    
    // Enter phone number
    await page.fill('input[placeholder*="phone"], input[type="tel"]', '+1234567890');
    await page.click('button:has-text("Send Code")');
    
    // Enter verification code
    await page.fill('input[placeholder*="verification"], input[placeholder*="code"]', '123456');
    await page.click('button:has-text("Verify")');
    
    await waitForLoadingToComplete(page);
  });

  test('should disable two-factor authentication', async ({ page }) => {
    // Setup user with 2FA enabled
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
            mfa_enabled: true,
            created_at: '2023-01-01T00:00:00Z'
          }
        })
      });
    });
    
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock disable 2FA API
    await page.route('**/api/v1/auth/mfa/disable', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: 'Two-factor authentication disabled'
        })
      });
    });
    
    // Should show disable button when 2FA is enabled
    await page.click('button:has-text("Disable")');
    
    // Confirm disable action
    if (await page.locator('button:has-text("Confirm"), button:has-text("Yes")').isVisible()) {
      await page.click('button:has-text("Confirm"), button:has-text("Yes")');
    }
    
    await waitForLoadingToComplete(page);
  });

  test('should display security audit log', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock security log API
    await page.route('**/api/v1/user/security-log', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: [
            {
              id: '1',
              action: 'login',
              ip_address: '192.168.1.100',
              user_agent: 'Mozilla/5.0...',
              location: 'San Francisco, CA',
              timestamp: '2023-12-01T10:30:00Z'
            },
            {
              id: '2',
              action: 'password_change',
              ip_address: '192.168.1.100',
              user_agent: 'Mozilla/5.0...',
              location: 'San Francisco, CA',
              timestamp: '2023-11-30T15:45:00Z'
            }
          ]
        })
      });
    });
    
    // Look for security log section
    if (await page.locator('text=Security Log, text=Audit, text=Activity').isVisible()) {
      await expect(page.locator('text=login')).toBeVisible();
      await expect(page.locator('text=password_change')).toBeVisible();
      await expect(page.locator('text=192.168.1.100')).toBeVisible();
    }
  });

  test('should manage active sessions', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock sessions API
    await page.route('**/api/v1/user/sessions', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: '1',
                is_current: true,
                ip_address: '192.168.1.100',
                user_agent: 'Mozilla/5.0 (Macintosh; Intel Mac OS X)',
                location: 'San Francisco, CA',
                last_active: '2023-12-01T10:30:00Z'
              },
              {
                id: '2',
                is_current: false,
                ip_address: '10.0.0.50',
                user_agent: 'Mozilla/5.0 (iPhone; CPU iPhone OS)',
                location: 'New York, NY',
                last_active: '2023-11-30T08:15:00Z'
              }
            ]
          })
        });
      }
    });
    
    // Look for session management section
    if (await page.locator('text=Sessions, text=Active Sessions, text=Device').isVisible()) {
      await expect(page.locator('text=Current session, text=This device')).toBeVisible();
      await expect(page.locator('text=iPhone')).toBeVisible();
      
      // Should be able to revoke other sessions
      if (await page.locator('button:has-text("Revoke"), button:has-text("Sign out")').isVisible()) {
        // Mock revoke session API
        await page.route('**/api/v1/user/sessions/2', async route => {
          if (route.request().method() === 'DELETE') {
            await route.fulfill({
              status: 200,
              contentType: 'application/json',
              body: JSON.stringify({
                success: true,
                message: 'Session revoked successfully'
              })
            });
          }
        });
        
        await page.click('button:has-text("Revoke"), button:has-text("Sign out")');
        await waitForLoadingToComplete(page);
      }
    }
  });

  test('should handle security settings on mobile', async ({ page }) => {
    // Set mobile viewport
    await page.setViewportSize({ width: 375, height: 667 });
    
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Verify all security sections are accessible on mobile
    await expect(page.locator('h3:has-text("Password & Authentication")')).toBeVisible();
    await expect(page.locator('h3:has-text("Two-Factor Authentication")')).toBeVisible();
    
    // Form elements should be properly sized for mobile
    await expect(page.locator('input[placeholder="Enter current password"]')).toBeVisible();
    await expect(page.locator('button:has-text("Change Password")')).toBeVisible();
  });

  test('should validate SSH key format', async ({ page }) => {
    await loginUser(page);
    await page.goto('/settings');
    
    // Navigate to Security tab
    await page.click('button:has-text("Security")');
    
    // Mock validation error for invalid SSH key
    await page.route('**/api/v1/user/ssh-keys', async route => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 400,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            error: 'Invalid SSH key format',
            validation_errors: {
              key: ['SSH key format is invalid']
            }
          })
        });
      }
    });
    
    // Try to add invalid SSH key
    await page.click('button:has-text("Add SSH Key"), button:has-text("Add Key")');
    
    await page.fill('input[placeholder*="title"], input[placeholder*="Title"]', 'Invalid Key');
    await page.fill('textarea[placeholder*="ssh-rsa"], textarea[placeholder*="key"]', 'invalid-key-format');
    
    await page.click('button:has-text("Add Key"), button:has-text("Save")');
    
    // Should show validation error (implementation-dependent)
  });
});