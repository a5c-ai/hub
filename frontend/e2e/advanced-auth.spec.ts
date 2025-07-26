import { test, expect } from '@playwright/test';
import { waitForLoadingToComplete } from './helpers/test-utils';

test.describe('Advanced Authentication Features', () => {
  test.beforeEach(async ({ page }) => {
    await page.context().clearCookies();
    await page.context().clearPermissions();
  });

  test.describe('Multi-Factor Authentication (MFA)', () => {
    test('should setup MFA with TOTP', async ({ page }) => {
      // Login first
      await page.goto('/login');
      await page.fill('[data-testid="email-input"]', 'user@example.com');
      await page.fill('[data-testid="password-input"]', 'password123');
      
      // Mock login response
      await page.route('**/api/v1/auth/login', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { access_token: 'mock-token', user: { id: '1', email: 'user@example.com' } }
          })
        });
      });
      
      await page.click('[data-testid="login-button"]');
      await page.waitForURL('/dashboard');

      // Navigate to security settings
      await page.goto('/settings/security');
      await waitForLoadingToComplete(page);

      // Mock MFA setup API
      await page.route('**/api/v1/auth/mfa/setup', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              secret: 'JBSWY3DPEHPK3PXP',
              qrCode: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8/5+hHgAHggJ/PchI7wAAAABJRU5ErkJggg==',
              backupCodes: ['123456', '789012', '345678']
            }
          })
        });
      });

      // Start MFA setup
      await expect(page.locator('text=Two-Factor Authentication')).toBeVisible();
      await page.click('[data-testid="setup-mfa-button"]');

      // Verify MFA setup modal
      await expect(page.locator('[data-testid="mfa-setup-modal"]')).toBeVisible();
      await expect(page.locator('text=Scan QR Code')).toBeVisible();
      await expect(page.locator('[data-testid="qr-code"]')).toBeVisible();

      // Enter TOTP code
      await page.fill('[data-testid="totp-code"]', '123456');
      
      // Mock MFA verification
      await page.route('**/api/v1/auth/mfa/verify', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true, data: { verified: true } })
        });
      });

      await page.click('[data-testid="verify-mfa-button"]');

      // Verify backup codes are shown
      await expect(page.locator('[data-testid="backup-codes"]')).toBeVisible();
      await expect(page.locator('text=123456')).toBeVisible();
      await expect(page.locator('text=789012')).toBeVisible();

      // Complete setup
      await page.click('[data-testid="complete-mfa-setup"]');
      await expect(page.locator('text=MFA Enabled')).toBeVisible();
    });

    test('should require MFA on login when enabled', async ({ page }) => {
      await page.goto('/login');
      await page.fill('[data-testid="email-input"]', 'mfa-user@example.com');
      await page.fill('[data-testid="password-input"]', 'password123');

      // Mock login response requiring MFA
      await page.route('**/api/v1/auth/login', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { requiresMFA: true, tempToken: 'temp-token-123' }
          })
        });
      });

      await page.click('[data-testid="login-button"]');

      // Should show MFA challenge
      await expect(page.locator('[data-testid="mfa-challenge"]')).toBeVisible();
      await expect(page.locator('text=Enter Authentication Code')).toBeVisible();
      await expect(page.locator('[data-testid="mfa-code-input"]')).toBeVisible();

      // Enter MFA code
      await page.fill('[data-testid="mfa-code-input"]', '123456');

      // Mock MFA verification
      await page.route('**/api/v1/auth/mfa/challenge', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { access_token: 'full-token', user: { id: '1', email: 'mfa-user@example.com' } }
          })
        });
      });

      await page.click('[data-testid="verify-mfa-code"]');

      // Should redirect to dashboard
      await expect(page).toHaveURL('/dashboard');
    });

    test('should allow using backup codes for MFA', async ({ page }) => {
      await page.goto('/login');
      await page.fill('[data-testid="email-input"]', 'mfa-user@example.com');
      await page.fill('[data-testid="password-input"]', 'password123');

      await page.route('**/api/v1/auth/login', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { requiresMFA: true, tempToken: 'temp-token-123' }
          })
        });
      });

      await page.click('[data-testid="login-button"]');
      await expect(page.locator('[data-testid="mfa-challenge"]')).toBeVisible();

      // Click "Use backup code" link
      await page.click('[data-testid="use-backup-code"]');
      await expect(page.locator('[data-testid="backup-code-input"]')).toBeVisible();

      // Enter backup code
      await page.fill('[data-testid="backup-code-input"]', '123456');

      // Mock backup code verification
      await page.route('**/api/v1/auth/mfa/backup-code', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { access_token: 'full-token', user: { id: '1' } }
          })
        });
      });

      await page.click('[data-testid="verify-backup-code"]');
      await expect(page).toHaveURL('/dashboard');
    });
  });

  test.describe('WebAuthn/FIDO2 Authentication', () => {
    test('should setup WebAuthn security key', async ({ page }) => {
      // Skip if WebAuthn not supported
      const webAuthnSupported = await page.evaluate(() => {
        return 'credentials' in navigator && 'create' in navigator.credentials;
      });
      
      if (!webAuthnSupported) {
        test.skip();
      }

      // Login and navigate to security settings
      await page.goto('/login');
      await page.fill('[data-testid="email-input"]', 'user@example.com');
      await page.fill('[data-testid="password-input"]', 'password123');
      await page.click('[data-testid="login-button"]');
      await page.waitForURL('/dashboard');

      await page.goto('/settings/security');
      await waitForLoadingToComplete(page);

      // Start WebAuthn setup
      await expect(page.locator('text=Security Keys')).toBeVisible();
      await page.click('[data-testid="add-security-key"]');

      // Mock WebAuthn registration
      await page.addInitScript(() => {
        // Mock navigator.credentials.create
        Object.defineProperty(navigator, 'credentials', {
          value: {
            create: () => Promise.resolve({
              id: 'mock-credential-id',
              rawId: new ArrayBuffer(32),
              response: {
                attestationObject: new ArrayBuffer(64),
                clientDataJSON: new ArrayBuffer(32)
              },
              type: 'public-key'
            })
          },
          writable: true
        });
      });

      await expect(page.locator('[data-testid="webauthn-setup-modal"]')).toBeVisible();
      await page.fill('[data-testid="key-name"]', 'My Security Key');

      // Mock WebAuthn registration API
      await page.route('**/api/v1/auth/webauthn/register', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true, data: { credentialId: 'mock-id' } })
        });
      });

      await page.click('[data-testid="register-key"]');

      // Verify key was added
      await expect(page.locator('text=My Security Key')).toBeVisible();
      await expect(page.locator('[data-testid="security-key-item"]')).toBeVisible();
    });

    test('should authenticate with WebAuthn', async ({ page }) => {
      const webAuthnSupported = await page.evaluate(() => {
        return 'credentials' in navigator && 'get' in navigator.credentials;
      });
      
      if (!webAuthnSupported) {
        test.skip();
      }

      await page.goto('/login');
      await page.fill('[data-testid="email-input"]', 'webauth-user@example.com');
      await page.fill('[data-testid="password-input"]', 'password123');

      // Mock WebAuthn authentication
      await page.addInitScript(() => {
        Object.defineProperty(navigator, 'credentials', {
          value: {
            get: () => Promise.resolve({
              id: 'mock-credential-id',
              rawId: new ArrayBuffer(32),
              response: {
                authenticatorData: new ArrayBuffer(64),
                clientDataJSON: new ArrayBuffer(32),
                signature: new ArrayBuffer(32)
              },
              type: 'public-key'
            })
          },
          writable: true
        });
      });

      // Mock login response requiring WebAuthn
      await page.route('**/api/v1/auth/login', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { requiresWebAuthn: true, challenge: 'mock-challenge' }
          })
        });
      });

      await page.click('[data-testid="login-button"]');

      // Should show WebAuthn prompt
      await expect(page.locator('[data-testid="webauthn-challenge"]')).toBeVisible();
      await expect(page.locator('text=Touch your security key')).toBeVisible();

      // Mock WebAuthn verification
      await page.route('**/api/v1/auth/webauthn/authenticate', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { access_token: 'token', user: { id: '1' } }
          })
        });
      });

      await page.click('[data-testid="authenticate-webauthn"]');
      await expect(page).toHaveURL('/dashboard');
    });
  });

  test.describe('SAML SSO Integration', () => {
    test('should initiate SAML SSO login', async ({ page }) => {
      await page.goto('/login');

      // Check for SSO options
      await expect(page.locator('[data-testid="sso-section"]')).toBeVisible();
      await expect(page.locator('text=Sign in with SSO')).toBeVisible();

      // Mock SAML initiation
      await page.route('**/api/v1/auth/saml/initiate', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { redirectUrl: 'https://idp.example.com/saml/sso?SAMLRequest=mock' }
          })
        });
      });

      await page.click('[data-testid="saml-sso-button"]');

      // Should redirect to SAML IdP (mocked)
      await expect(page).toHaveURL(/idp\.example\.com/);
    });

    test('should handle SAML SSO callback', async ({ page }) => {
      // Simulate SAML callback with response
      const samlResponse = 'mock-saml-response';
      await page.goto(`/auth/saml/callback?SAMLResponse=${encodeURIComponent(samlResponse)}`);

      // Mock SAML response processing
      await page.route('**/api/v1/auth/saml/callback', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              access_token: 'saml-token',
              user: {
                id: '1',
                email: 'saml-user@example.com',
                name: 'SAML User',
                ssoProvider: 'SAML'
              }
            }
          })
        });
      });

      await waitForLoadingToComplete(page);

      // Should redirect to dashboard after successful SAML auth
      await expect(page).toHaveURL('/dashboard');
    });
  });

  test.describe('OIDC Authentication', () => {
    test('should initiate OIDC login', async ({ page }) => {
      await page.goto('/login');

      await expect(page.locator('[data-testid="oidc-providers"]')).toBeVisible();
      await expect(page.locator('text=Sign in with Google')).toBeVisible();
      await expect(page.locator('text=Sign in with Microsoft')).toBeVisible();

      // Mock OIDC initiation
      await page.route('**/api/v1/auth/oidc/google/initiate', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { authUrl: 'https://accounts.google.com/oauth/authorize?client_id=mock' }
          })
        });
      });

      await page.click('[data-testid="google-oidc-button"]');

      // Should redirect to Google OAuth (mocked)
      await expect(page).toHaveURL(/accounts\.google\.com/);
    });

    test('should handle OIDC callback', async ({ page }) => {
      // Simulate OIDC callback with authorization code
      await page.goto('/auth/oidc/callback?code=mock-auth-code&state=mock-state');

      // Mock OIDC token exchange
      await page.route('**/api/v1/auth/oidc/callback', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              access_token: 'oidc-token',
              user: {
                id: '1',
                email: 'oidc-user@example.com',
                name: 'OIDC User',
                avatar: 'https://example.com/avatar.jpg',
                ssoProvider: 'Google'
              }
            }
          })
        });
      });

      await waitForLoadingToComplete(page);
      await expect(page).toHaveURL('/dashboard');
    });
  });

  test.describe('Session Management', () => {
    test('should display active sessions', async ({ page }) => {
      // Login first
      await page.goto('/login');
      await page.fill('[data-testid="email-input"]', 'user@example.com');
      await page.fill('[data-testid="password-input"]', 'password123');
      await page.click('[data-testid="login-button"]');
      await page.waitForURL('/dashboard');

      await page.goto('/settings/security/sessions');
      await waitForLoadingToComplete(page);

      // Mock sessions API
      await page.route('**/api/v1/auth/sessions', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'session-1',
                userAgent: 'Mozilla/5.0 (Windows NT 10.0; Win64; x64)',
                ip: '192.168.1.1',
                location: 'New York, US',
                lastActive: '2024-01-15T10:30:00Z',
                current: true
              },
              {
                id: 'session-2',
                userAgent: 'Mozilla/5.0 (iPhone; CPU iPhone OS 17_0)',
                ip: '192.168.1.100',
                location: 'New York, US',
                lastActive: '2024-01-14T08:15:00Z',
                current: false
              }
            ]
          })
        });
      });

      await expect(page.locator('h1')).toContainText('Active Sessions');
      await expect(page.locator('[data-testid="session-item"]')).toHaveCount(2);
      await expect(page.locator('text=Current Session')).toBeVisible();
      await expect(page.locator('text=New York, US')).toBeVisible();
    });

    test('should revoke individual sessions', async ({ page }) => {
      await page.goto('/settings/security/sessions');
      await waitForLoadingToComplete(page);

      // Mock revoke session API
      await page.route('**/api/v1/auth/sessions/session-2/revoke', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      // Click revoke on non-current session
      await page.click('[data-testid="revoke-session-session-2"]');
      
      // Confirm revocation
      await expect(page.locator('[data-testid="confirm-revoke-modal"]')).toBeVisible();
      await page.click('[data-testid="confirm-revoke"]');

      // Verify session is removed
      await expect(page.locator('[data-testid="session-item"]')).toHaveCount(1);
    });

    test('should revoke all other sessions', async ({ page }) => {
      await page.goto('/settings/security/sessions');
      await waitForLoadingToComplete(page);

      // Mock revoke all sessions API
      await page.route('**/api/v1/auth/sessions/revoke-all', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.click('[data-testid="revoke-all-sessions"]');
      
      // Confirm revocation
      await expect(page.locator('[data-testid="confirm-revoke-all-modal"]')).toBeVisible();
      await page.click('[data-testid="confirm-revoke-all"]');

      // Should only show current session
      await expect(page.locator('[data-testid="session-item"]')).toHaveCount(1);
      await expect(page.locator('text=Current Session')).toBeVisible();
    });
  });

  test.describe('Device Trust and Management', () => {
    test('should mark device as trusted', async ({ page }) => {
      await page.goto('/login');
      await page.fill('[data-testid="email-input"]', 'user@example.com');
      await page.fill('[data-testid="password-input"]', 'password123');

      // Check "Trust this device" option
      await expect(page.locator('[data-testid="trust-device-checkbox"]')).toBeVisible();
      await page.check('[data-testid="trust-device-checkbox"]');

      await page.route('**/api/v1/auth/login', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { 
              access_token: 'token', 
              user: { id: '1' },
              deviceTrusted: true
            }
          })
        });
      });

      await page.click('[data-testid="login-button"]');
      await expect(page).toHaveURL('/dashboard');
    });

    test('should manage trusted devices', async ({ page }) => {
      await page.goto('/settings/security/devices');
      await waitForLoadingToComplete(page);

      // Mock trusted devices API
      await page.route('**/api/v1/auth/trusted-devices', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'device-1',
                name: 'Chrome on Windows',
                fingerprint: 'abc123',
                lastUsed: '2024-01-15T10:30:00Z',
                current: true
              },
              {
                id: 'device-2',
                name: 'Safari on iPhone',
                fingerprint: 'def456',
                lastUsed: '2024-01-14T08:15:00Z',
                current: false
              }
            ]
          })
        });
      });

      await expect(page.locator('h1')).toContainText('Trusted Devices');
      await expect(page.locator('[data-testid="device-item"]')).toHaveCount(2);
      
      // Remove trusted device
      await page.route('**/api/v1/auth/trusted-devices/device-2', async route => {
        await route.fulfill({ status: 200, body: JSON.stringify({ success: true }) });
      });

      await page.click('[data-testid="remove-device-device-2"]');
      await page.click('[data-testid="confirm-remove-device"]');
      
      await expect(page.locator('[data-testid="device-item"]')).toHaveCount(1);
    });
  });
});