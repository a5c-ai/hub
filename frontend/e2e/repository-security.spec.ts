import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('Repository Security & Secrets Management', () => {
  const mockRepository = {
    id: '1',
    name: 'test-repo',
    full_name: 'testuser/test-repo',
    description: 'Test repository for security tests',
    private: false,
    language: 'TypeScript',
    default_branch: 'main',
    actions_enabled: true,
    stargazers_count: 10,
    forks_count: 2,
    updated_at: '2024-07-20T10:00:00Z'
  };

  const mockSecrets = [
    {
      id: '1',
      name: 'DATABASE_URL',
      created_at: '2024-07-20T10:00:00Z',
      updated_at: '2024-07-20T10:00:00Z',
      environment: 'production'
    },
    {
      id: '2',
      name: 'API_KEY',
      created_at: '2024-07-19T15:30:00Z',
      updated_at: '2024-07-19T15:30:00Z'
    }
  ];

  const mockDeployKeys = [
    {
      id: '1',
      title: 'Production Deploy Key',
      key: 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7...',
      read_only: true,
      verified: true,
      created_at: '2024-07-20T10:00:00Z',
      last_used: '2024-07-24T08:00:00Z'
    },
    {
      id: '2',
      title: 'CI/CD Key',
      key: 'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI...',
      read_only: false,
      verified: false,
      created_at: '2024-07-19T15:30:00Z'
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

    // Set authentication state
    await page.addInitScript(() => {
      window.localStorage.setItem('auth_token', 'mock-jwt-token');
    });
  });

  test.describe('Repository Secrets Management', () => {
    test.beforeEach(async ({ page }) => {
      // Mock secrets API
      await page.route('**/api/v1/repos/testuser/test-repo/actions/secrets', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              secrets: mockSecrets
            })
          });
        }
      });
    });

    test('should display secrets management page', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      
      // Should display secrets page
      await expect(page.locator('h1')).toContainText('Secrets');
      await expect(page.locator('text=Secrets are encrypted environment variables')).toBeVisible();
      await expect(page.locator('button:has-text("New repository secret")')).toBeVisible();
    });

    test('should list existing secrets', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      
      // Should display secret list
      await expect(page.locator('text=DATABASE_URL')).toBeVisible();
      await expect(page.locator('text=API_KEY')).toBeVisible();
      await expect(page.locator('text=production')).toBeVisible(); // Environment tag
      
      // Should show update/delete buttons
      await expect(page.locator('button:has-text("Update")')).toHaveCount(2);
      await expect(page.locator('button:has-text("Delete")')).toHaveCount(2);
    });

    test('should show empty state when no secrets exist', async ({ page }) => {
      // Mock empty secrets response
      await page.route('**/api/v1/repos/testuser/test-repo/actions/secrets', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ secrets: [] })
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      
      // Should show empty state
      await expect(page.locator('text=No secrets yet')).toBeVisible();
      await expect(page.locator('text=Secrets are environment variables that are encrypted')).toBeVisible();
      await expect(page.locator('button:has-text("Add your first secret")')).toBeVisible();
    });

    test('should open create secret modal', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      
      // Click new secret button
      await page.click('button:has-text("New repository secret")');
      
      // Should show modal
      await expect(page.locator('text=New repository secret')).toBeVisible();
      await expect(page.locator('input[placeholder="SECRET_NAME"]')).toBeVisible();
      await expect(page.locator('textarea[placeholder="Enter secret value..."]')).toBeVisible();
      await expect(page.locator('input[placeholder="production"]')).toBeVisible();
    });

    test('should validate secret name format', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      await page.click('button:has-text("New repository secret")');
      
      // Test invalid secret name
      await page.fill('input[placeholder="SECRET_NAME"]', 'invalid-name');
      await expect(page.locator('text=Invalid secret name format')).toBeVisible();
      
      // Test valid secret name
      await page.fill('input[placeholder="SECRET_NAME"]', 'VALID_SECRET_NAME');
      await expect(page.locator('text=Invalid secret name format')).not.toBeVisible();
    });

    test('should create new secret', async ({ page }) => {
      // Mock create secret API
      await page.route('**/api/v1/repos/testuser/test-repo/actions/secrets', async route => {
        if (route.request().method() === 'POST') {
          const requestData = await route.request().postDataJSON();
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                id: '3',
                name: requestData.name,
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
                environment: requestData.environment
              }
            })
          });
        } else {
          await route.continue();
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      await page.click('button:has-text("New repository secret")');
      
      // Fill form
      await page.fill('input[placeholder="SECRET_NAME"]', 'NEW_SECRET');
      await page.fill('textarea[placeholder="Enter secret value..."]', 'secret-value-123');
      await page.fill('input[placeholder="production"]', 'staging');
      
      // Submit form
      await page.click('button:has-text("Add secret")');
      
      // Should show saving state
      await expect(page.locator('button:has-text("Saving...")')).toBeVisible();
    });

    test('should update existing secret', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      
      // Click update on first secret
      await page.locator('text=DATABASE_URL').locator('..').locator('button:has-text("Update")').click();
      
      // Should show update modal with pre-filled name
      await expect(page.locator('text=Update secret')).toBeVisible();
      await expect(page.locator('input[value="DATABASE_URL"]')).toBeDisabled();
      await expect(page.locator('input[value="production"]')).toBeVisible();
    });

    test('should delete secret with confirmation', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      
      // Click delete on first secret
      await page.locator('text=DATABASE_URL').locator('..').locator('button:has-text("Delete")').click();
      
      // Should show delete confirmation modal
      await expect(page.locator('text=Delete secret')).toBeVisible();
      await expect(page.locator('text=Are you sure you want to delete this secret')).toBeVisible();
      await expect(page.locator('button:has-text("Delete secret")')).toBeVisible();
    });

    test('should handle secret creation errors', async ({ page }) => {
      // Mock error response
      await page.route('**/api/v1/repos/testuser/test-repo/actions/secrets', async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 400,
            contentType: 'application/json',
            body: JSON.stringify({
              success: false,
              error: 'Secret name already exists'
            })
          });
        } else {
          await route.continue();
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      await page.click('button:has-text("New repository secret")');
      
      // Fill and submit form
      await page.fill('input[placeholder="SECRET_NAME"]', 'EXISTING_SECRET');
      await page.fill('textarea[placeholder="Enter secret value..."]', 'secret-value');
      await page.click('button:has-text("Add secret")');
      
      // Should show error message
      await expect(page.locator('text=Secret name already exists')).toBeVisible();
    });
  });

  test.describe('Deploy Keys Management', () => {
    test.beforeEach(async ({ page }) => {
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

      // Mock deploy keys API
      await page.route('**/api/v1/repositories/testuser/test-repo/keys', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify(mockDeployKeys)
          });
        }
      });
    });

    test('should display deploy keys management page', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/keys');
      
      // Should display deploy keys page
      await expect(page.locator('h1')).toContainText('Deploy Keys');
      await expect(page.locator('text=Deploy keys allow read-only or read-write access')).toBeVisible();
      await expect(page.locator('button:has-text("Add Deploy Key")')).toBeVisible();
    });

    test('should list existing deploy keys', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/keys');
      
      // Should display deploy key list
      await expect(page.locator('text=Production Deploy Key')).toBeVisible();
      await expect(page.locator('text=CI/CD Key')).toBeVisible();
      
      // Should show key properties
      await expect(page.locator('text=Read-only')).toBeVisible();
      await expect(page.locator('text=Read-write')).toBeVisible();
      await expect(page.locator('text=Verified')).toBeVisible();
      
      // Should show key fingerprints
      await expect(page.locator('text=ssh-rsa AAAAB3NzaC1yc2EAAAA...')).toBeVisible();
      await expect(page.locator('text=ssh-ed25519 AAAAC3NzaC1lZDI1NTE5A...')).toBeVisible();
    });

    test('should show empty state when no deploy keys exist', async ({ page }) => {
      // Mock empty deploy keys response
      await page.route('**/api/v1/repositories/testuser/test-repo/keys', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify([])
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/keys');
      
      // Should show empty state
      await expect(page.locator('text=No deploy keys')).toBeVisible();
      await expect(page.locator('text=Deploy keys allow servers to access your repository')).toBeVisible();
    });

    test('should open create deploy key modal', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/keys');
      
      // Click add deploy key button
      await page.click('button:has-text("Add Deploy Key")');
      
      // Should show modal
      await expect(page.locator('text=Add Deploy Key')).toBeVisible();
      await expect(page.locator('input[placeholder="My Deploy Key"]')).toBeVisible();
      await expect(page.locator('textarea[placeholder="ssh-rsa AAAAB3NzaC1yc2EAAAA..."]')).toBeVisible();
      await expect(page.locator('text=Allow write access')).toBeVisible();
    });

    test('should validate SSH key format', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/keys');
      await page.click('button:has-text("Add Deploy Key")');
      
      // Test invalid SSH key
      await page.fill('textarea[placeholder*="ssh-rsa"]', 'invalid-ssh-key');
      await expect(page.locator('text=Please enter a valid SSH public key')).toBeVisible();
      
      // Test valid SSH key
      await page.fill('textarea[placeholder*="ssh-rsa"]', 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... user@example.com');
      await expect(page.locator('text=Please enter a valid SSH public key')).not.toBeVisible();
    });

    test('should create new deploy key', async ({ page }) => {
      // Mock create deploy key API
      await page.route('**/api/v1/repositories/testuser/test-repo/keys', async route => {
        if (route.request().method() === 'POST') {
          const requestData = await route.request().postDataJSON();
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              id: '3',
              title: requestData.title,
              key: requestData.key,
              read_only: requestData.read_only,
              verified: false,
              created_at: new Date().toISOString()
            })
          });
        } else {
          await route.continue();
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings/keys');
      await page.click('button:has-text("Add Deploy Key")');
      
      // Fill form
      await page.fill('input[placeholder="My Deploy Key"]', 'Test Deploy Key');
      await page.fill('textarea[placeholder*="ssh-rsa"]', 'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... test@example.com');
      
      // Toggle write access
      await page.uncheck('input[type="checkbox"]#read_only');
      
      // Submit form
      await page.click('button:has-text("Add Key")');
      
      // Should show loading state
      await expect(page.locator('button:has-text("Adding...")')).toBeVisible();
    });

    test('should delete deploy key with confirmation', async ({ page }) => {
      // Mock delete deploy key API
      await page.route('**/api/v1/repositories/testuser/test-repo/keys/1', async route => {
        if (route.request().method() === 'DELETE') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings/keys');
      
      // Mock confirmation dialog
      page.on('dialog', async dialog => {
        expect(dialog.message()).toContain('Are you sure you want to delete this deploy key');
        await dialog.accept();
      });
      
      // Click delete on first deploy key
      await page.locator('text=Production Deploy Key').locator('..').locator('button:has-text("Delete")').click();
    });

    test('should display deploy key help information', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/keys');
      
      // Should show help section
      await expect(page.locator('text=About Deploy Keys')).toBeVisible();
      await expect(page.locator('text=Deploy keys are SSH keys that grant access')).toBeVisible();
      await expect(page.locator('text=Read-only keys can pull')).toBeVisible();
      await expect(page.locator('text=Read-write keys can both pull from and push')).toBeVisible();
      await expect(page.locator('text=useful for CI/CD systems')).toBeVisible();
    });

    test('should show SSH key generation tip', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/keys');
      await page.click('button:has-text("Add Deploy Key")');
      
      // Should show tip for generating SSH key
      await expect(page.locator('text=Generate an SSH key with:')).toBeVisible();
      await expect(page.locator('code')).toContainText('ssh-keygen -t ed25519');
    });
  });

  test.describe('Security Advisory & Vulnerability Management', () => {
    test('should access security settings from repository settings', async ({ page }) => {
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

      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Navigate to security tab
      await page.click('button:has-text("Security")');
      
      // Should display security settings
      await expect(page.locator('h3')).toContainText('Security Settings');
      await expect(page.locator('text=Vulnerability Alerts')).toBeVisible();
      await expect(page.locator('text=Dependency Graph')).toBeVisible();
    });
  });

  test.describe('Access Control & Collaborators', () => {
    const mockCollaborators = [
      {
        id: '1',
        username: 'collaborator1',
        email: 'collab1@example.com',
        permission: 'write',
        added_at: '2024-07-20T10:00:00Z'
      },
      {
        id: '2',
        username: 'collaborator2',
        email: 'collab2@example.com',
        permission: 'read',
        added_at: '2024-07-19T15:30:00Z'
      }
    ];

    test('should display repository access control', async ({ page }) => {
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

      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Navigate to access tab
      await page.click('button:has-text("Access")');
      
      // Should display access control settings
      await expect(page.locator('h3')).toContainText('Repository Access');
      await expect(page.locator('text=Public Access')).toBeVisible();
      await expect(page.locator('text=Collaborators')).toBeVisible();
      await expect(page.locator('button:has-text("Add Collaborator")')).toBeVisible();
    });

    test('should show correct visibility status', async ({ page }) => {
      // Mock private repository
      await page.route('**/api/v1/repositories/testuser/test-repo', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { ...mockRepository, private: true }
          })
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Access")');
      
      // Should show disabled for private repository
      await expect(page.locator('text=Disabled')).toBeVisible();
    });
  });

  test.describe('Error Handling & Network Issues', () => {
    test('should handle secrets API errors gracefully', async ({ page }) => {
      // Mock error response
      await page.route('**/api/v1/repos/testuser/test-repo/actions/secrets', async route => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            error: 'Internal server error'
          })
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      
      // Should show error message
      await expect(page.locator('text=Internal server error')).toBeVisible();
    });

    test('should handle deploy keys API errors gracefully', async ({ page }) => {
      // Mock repository details first
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

      // Mock error response for deploy keys
      await page.route('**/api/v1/repositories/testuser/test-repo/keys', async route => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            error: 'Failed to fetch deploy keys'
          })
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/keys');
      
      // Should show error and retry option
      await expect(page.locator('text=Error: Failed to fetch deploy keys')).toBeVisible();
      await expect(page.locator('button:has-text("Try Again")')).toBeVisible();
    });

    test('should handle network failures', async ({ page }) => {
      // Simulate network failure
      await page.route('**/api/v1/repos/testuser/test-repo/actions/secrets', async route => {
        await route.abort('internetdisconnected');
      });

      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      
      // Should show loading state or error handling
      await expect(page.locator('.animate-pulse')).toBeVisible();
    });
  });

  test.describe('Mobile Security Management', () => {
    test('should be responsive on mobile devices', async ({ page }) => {
      // Set mobile viewport
      await page.setViewportSize({ width: 375, height: 667 });
      
      // Mock secrets for mobile test
      await page.route('**/api/v1/repos/testuser/test-repo/actions/secrets', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ secrets: mockSecrets })
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      
      // Should display mobile-friendly layout
      await expect(page.locator('h1')).toContainText('Secrets');
      await expect(page.locator('button:has-text("New repository secret")')).toBeVisible();
      
      // Test modal on mobile
      await page.click('button:has-text("New repository secret")');
      await expect(page.locator('text=New repository secret')).toBeVisible();
    });
  });

  test.describe('Security Best Practices Validation', () => {
    test('should enforce secure secret naming conventions', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings/secrets');
      await page.click('button:has-text("New repository secret")');
      
      // Test various invalid secret names
      const invalidNames = ['lowercase', 'with-dashes', 'with spaces', '123STARTS_WITH_NUMBER'];
      
      for (const name of invalidNames) {
        await page.fill('input[placeholder="SECRET_NAME"]', name);
        await expect(page.locator('text=Invalid secret name format')).toBeVisible();
      }
      
      // Test valid name
      await page.fill('input[placeholder="SECRET_NAME"]', 'VALID_SECRET_NAME');
      await expect(page.locator('text=Invalid secret name format')).not.toBeVisible();
    });

    test('should validate SSH key security standards', async ({ page }) => {
      // Mock repository and deploy keys
      await page.route('**/api/v1/repositories/testuser/test-repo', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true, data: mockRepository })
        });
      });

      await page.route('**/api/v1/repositories/testuser/test-repo/keys', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(mockDeployKeys)
        });
      });

      await page.goto('/repositories/testuser/test-repo/settings/keys');
      await page.click('button:has-text("Add Deploy Key")');
      
      // Test various SSH key formats
      const validKeys = [
        'ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7... user@example.com',
        'ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAI... user@example.com',
        'ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAy... user@example.com'
      ];

      const invalidKeys = [
        'not-an-ssh-key',
        'ssh-rsa invalid-key-content',
        'BEGIN RSA PRIVATE KEY'
      ];

      // Test valid keys
      for (const key of validKeys) {
        await page.fill('textarea[placeholder*="ssh-rsa"]', key);
        await expect(page.locator('text=Please enter a valid SSH public key')).not.toBeVisible();
      }

      // Test invalid keys
      for (const key of invalidKeys) {
        await page.fill('textarea[placeholder*="ssh-rsa"]', key);
        await expect(page.locator('text=Please enter a valid SSH public key')).toBeVisible();
      }
    });
  });
});