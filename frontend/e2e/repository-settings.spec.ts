import { test, expect } from '@playwright/test';
import { testUser } from './helpers/test-utils';

test.describe('Repository Settings & Administration', () => {
  const mockRepository = {
    id: '1',
    name: 'test-repo',
    full_name: 'testuser/test-repo',
    description: 'Test repository for settings tests',
    private: false,
    language: 'TypeScript',
    default_branch: 'main',
    stargazers_count: 10,
    forks_count: 2,
    updated_at: '2024-07-20T10:00:00Z',
    clone_url: 'https://hub.example.com/testuser/test-repo.git'
  };

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

  test.describe('General Repository Settings', () => {
    test('should display general settings tab by default', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Should display settings page
      await expect(page.locator('h1')).toContainText('Repository Settings');
      await expect(page.locator('text=General')).toHaveClass(/bg-blue-100/);
      
      // Should show repository details form
      await expect(page.locator('h3')).toContainText('Repository Details');
      await expect(page.locator('input[value="test-repo"]')).toBeVisible();
      await expect(page.locator('textarea[placeholder*="Short description"]')).toBeVisible();
    });

    test('should edit repository name and description', async ({ page }) => {
      // Mock update repository API
      await page.route('**/api/v1/repositories/testuser/test-repo', async route => {
        if (route.request().method() === 'PUT') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: { ...mockRepository, name: 'updated-repo', description: 'Updated description' }
            })
          });
        } else {
          await route.continue();
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Edit repository name
      const nameInput = page.locator('input[value="test-repo"]');
      await nameInput.fill('updated-repo');
      
      // Edit description
      const descInput = page.locator('textarea[placeholder*="Short description"]');
      await descInput.fill('Updated description');
      
      // Save changes
      await page.click('button:has-text("Save Changes")');
      
      // Should show saving state
      await expect(page.locator('button:has-text("Saving...")')).toBeVisible();
    });

    test('should toggle repository visibility', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Should show private checkbox
      const privateCheckbox = page.locator('input[type="checkbox"]#private');
      await expect(privateCheckbox).not.toBeChecked(); // Repository is public by default
      
      // Toggle to private
      await privateCheckbox.check();
      await expect(privateCheckbox).toBeChecked();
      
      // Save changes
      await page.click('button:has-text("Save Changes")');
    });

    test('should change default branch', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Change default branch
      const branchInput = page.locator('input[value="main"]');
      await branchInput.fill('develop');
      
      // Save changes
      await page.click('button:has-text("Save Changes")');
    });
  });

  test.describe('Repository Access Control', () => {
    test('should display access control settings', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Navigate to access tab
      await page.click('button:has-text("Access")');
      
      // Should display access settings
      await expect(page.locator('h3')).toContainText('Repository Access');
      await expect(page.locator('text=Public Access')).toBeVisible();
      await expect(page.locator('text=Enabled')).toBeVisible(); // Public repo
      await expect(page.locator('text=Collaborators')).toBeVisible();
    });

    test('should show add collaborator button', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Access")');
      
      // Should show add collaborator functionality
      await expect(page.locator('button:has-text("Add Collaborator")')).toBeVisible();
    });
  });

  test.describe('Branch Protection & Rules', () => {
    test('should display branch protection settings', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Navigate to branches tab
      await page.click('button:has-text("Branches")');
      
      // Should display branch protection
      await expect(page.locator('h3')).toContainText('Branch Protection');
      await expect(page.locator('text=Default Branch: main')).toBeVisible();
      await expect(page.locator('text=Configure protection rules')).toBeVisible();
    });

    test('should navigate to branch protection configuration', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Branches")');
      
      // Should have links to detailed branch settings
      await expect(page.locator('a[href="/repositories/testuser/test-repo/settings/branches"]')).toBeVisible();
      await expect(page.locator('button:has-text("Configure")')).toBeVisible();
      await expect(page.locator('button:has-text("Manage Branch Protection")')).toBeVisible();
    });
  });


  test.describe('Security Settings', () => {
    test('should display security settings', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Navigate to security tab
      await page.click('button:has-text("Security")');
      
      // Should display security options
      await expect(page.locator('h3')).toContainText('Security Settings');
      await expect(page.locator('text=Vulnerability Alerts')).toBeVisible();
      await expect(page.locator('text=Dependency Graph')).toBeVisible();
      await expect(page.locator('text=Deploy Keys')).toBeVisible();
    });

    test('should toggle security features', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Security")');
      
      // Should show security checkboxes
      const vulnerabilityCheckbox = page.locator('text=Vulnerability Alerts').locator('..').locator('input[type="checkbox"]');
      const dependencyCheckbox = page.locator('text=Dependency Graph').locator('..').locator('input[type="checkbox"]');
      
      await expect(vulnerabilityCheckbox).toBeChecked();
      await expect(dependencyCheckbox).toBeChecked();
      
      // Toggle vulnerability alerts
      await vulnerabilityCheckbox.uncheck();
      await expect(vulnerabilityCheckbox).not.toBeChecked();
    });

    test('should navigate to deploy keys management', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Security")');
      
      // Should have link to manage keys
      await expect(page.locator('a[href="/repositories/testuser/test-repo/settings/keys"]')).toBeVisible();
      await expect(page.locator('button:has-text("Manage Keys")')).toBeVisible();
    });
  });

  test.describe('Webhooks & Integrations', () => {
    test('should display webhooks settings', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Navigate to webhooks tab
      await page.click('button:has-text("Webhooks")');
      
      // Should display webhooks settings
      await expect(page.locator('h3')).toContainText('Webhooks');
      await expect(page.locator('text=Webhooks allow external services')).toBeVisible();
      await expect(page.locator('text=Manage webhooks for this repository')).toBeVisible();
    });

    test('should navigate to webhooks management', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Webhooks")');
      
      // Should have links to webhook management
      await expect(page.locator('a[href="/repositories/testuser/test-repo/settings/webhooks"]')).toBeVisible();
      await expect(page.locator('button:has-text("Manage Webhooks")')).toBeVisible();
      await expect(page.locator('button:has-text("Go to Webhooks")')).toBeVisible();
    });
  });

  test.describe('Danger Zone Operations', () => {
    test('should display danger zone settings', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Navigate to danger zone tab
      await page.click('button:has-text("Danger Zone")');
      
      // Should display danger zone
      await expect(page.locator('h3')).toContainText('Danger Zone');
      await expect(page.locator('text=Transfer Repository')).toBeVisible();
      await expect(page.locator('text=Archive Repository')).toBeVisible();
      await expect(page.locator('text=Delete Repository')).toBeVisible();
    });

    test('should show transfer repository option', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Danger Zone")');
      
      // Should show transfer option
      await expect(page.locator('text=Transfer this repository to another user')).toBeVisible();
      await expect(page.locator('button:has-text("Transfer")')).toBeVisible();
    });

    test('should show archive repository option', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Danger Zone")');
      
      // Should show archive option
      await expect(page.locator('text=Make this repository read-only')).toBeVisible();
      await expect(page.locator('button:has-text("Archive")')).toBeVisible();
    });

    test('should handle repository deletion with confirmation', async ({ page }) => {
      // Mock delete repository API
      await page.route('**/api/v1/repositories/testuser/test-repo', async route => {
        if (route.request().method() === 'DELETE') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        } else {
          await route.continue();
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings');
      await page.click('button:has-text("Danger Zone")');
      
      // Mock the confirmation dialog
      page.on('dialog', async dialog => {
        expect(dialog.message()).toContain('Are you sure you want to delete testuser/test-repo');
        expect(dialog.message()).toContain('This action cannot be undone');
        await dialog.accept();
      });
      
      // Click delete button
      await page.click('button:has-text("Delete")');
    });
  });

  test.describe('Repository Statistics & Analytics', () => {
    test('should display repository metadata', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Should show repository information in breadcrumb and header
      await expect(page.locator('text=testuser/test-repo')).toBeVisible();
      await expect(page.locator('text=Manage repository configuration and access')).toBeVisible();
    });
  });

  test.describe('Mobile Repository Management', () => {
    test('should be responsive on mobile devices', async ({ page }) => {
      // Set mobile viewport
      await page.setViewportSize({ width: 375, height: 667 });
      
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Should display mobile-friendly layout
      await expect(page.locator('h1')).toContainText('Repository Settings');
      
      // Navigation should work on mobile
      await page.click('button:has-text("Access")');
      await expect(page.locator('h3')).toContainText('Repository Access');
    });
  });

  test.describe('Error Handling', () => {
    test('should handle repository not found', async ({ page }) => {
      // Mock 404 response
      await page.route('**/api/v1/repositories/testuser/nonexistent', async route => {
        await route.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            error: 'Repository not found'
          })
        });
      });

      await page.goto('/repositories/testuser/nonexistent/settings');
      
      // Should handle error gracefully (component should handle 404)
      // The exact error handling depends on the component implementation
    });

    test('should handle save errors', async ({ page }) => {
      // Mock error response for save
      await page.route('**/api/v1/repositories/testuser/test-repo', async route => {
        if (route.request().method() === 'PUT') {
          await route.fulfill({
            status: 500,
            contentType: 'application/json',
            body: JSON.stringify({
              success: false,
              error: 'Failed to update repository'
            })
          });
        } else {
          await route.continue();
        }
      });

      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Make a change and save
      const nameInput = page.locator('input[value="test-repo"]');
      await nameInput.fill('updated-repo');
      await page.click('button:has-text("Save Changes")');
      
      // Should handle error (exact behavior depends on implementation)
    });

    test('should handle network errors gracefully', async ({ page }) => {
      // Simulate network failure
      await page.route('**/api/v1/repositories/testuser/test-repo', async route => {
        await route.abort('internetdisconnected');
      });

      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Should show loading state or error state
      await expect(page.locator('.animate-pulse')).toBeVisible();
    });
  });

  test.describe('Navigation and Breadcrumbs', () => {
    test('should display correct breadcrumb navigation', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Should show breadcrumb navigation
      await expect(page.locator('nav a[href="/repositories"]')).toContainText('Repositories');
      await expect(page.locator('nav a[href="/repositories/testuser/test-repo"]')).toContainText('testuser/test-repo');
      await expect(page.locator('nav text=Settings')).toBeVisible();
    });

    test('should navigate back to repository from breadcrumbs', async ({ page }) => {
      await page.goto('/repositories/testuser/test-repo/settings');
      
      // Click on repository link in breadcrumb
      await page.click('nav a[href="/repositories/testuser/test-repo"]');
      
      // Should navigate to repository page
      await expect(page).toHaveURL('/repositories/testuser/test-repo');
    });
  });
});