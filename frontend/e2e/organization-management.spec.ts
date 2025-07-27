import { test, expect } from '@playwright/test';
import { loginUser, waitForLoadingToComplete } from './helpers/test-utils';

test.describe('Organization Management Features', () => {
  const orgAdmin = {
    email: 'org-admin@example.com',
    password: 'AdminPassword123!'
  };

  test.beforeEach(async ({ page }) => {
    await loginUser(page, orgAdmin.email, orgAdmin.password);
    
    // Mock organization data
    await page.route('**/api/v1/organizations/**', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            id: 'org-1',
            name: 'Acme Corporation',
            slug: 'acme-corp',
            description: 'Software development company',
            website: 'https://acme.com',
            location: 'San Francisco, CA',
            memberCount: 45,
            repositoryCount: 28,
            settings: {
              visibility: 'private',
              memberVisibility: 'members',
              allowMemberRepositories: true,
              requireTwoFactor: true
            }
          }
        })
      });
    });
  });

  test.describe('Organization Settings Management', () => {
    test('should display organization profile settings', async ({ page }) => {
      await page.goto('/organizations/acme-corp/settings');
      await waitForLoadingToComplete(page);

      // Verify organization profile section
      await expect(page.locator('h1')).toContainText('Organization Settings');
      await expect(page.locator('[data-testid="org-profile-section"]')).toBeVisible();
      
      // Check profile fields
      await expect(page.locator('[data-testid="org-name"]')).toHaveValue('Acme Corporation');
      await expect(page.locator('[data-testid="org-description"]')).toHaveValue('Software development company');
      await expect(page.locator('[data-testid="org-website"]')).toHaveValue('https://acme.com');
      await expect(page.locator('[data-testid="org-location"]')).toHaveValue('San Francisco, CA');
    });

    test('should update organization profile', async ({ page }) => {
      await page.goto('/organizations/acme-corp/settings');
      await waitForLoadingToComplete(page);

      // Update organization info
      await page.fill('[data-testid="org-description"]', 'Leading software development company');
      await page.fill('[data-testid="org-location"]', 'San Francisco, California');

      // Mock update API
      await page.route('**/api/v1/organizations/org-1', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.click('[data-testid="save-org-profile"]');
      
      // Verify success message
      await expect(page.locator('[data-testid="success-message"]')).toBeVisible();
      await expect(page.locator('text=Organization profile updated')).toBeVisible();
    });

    test('should manage organization visibility settings', async ({ page }) => {
      await page.goto('/organizations/acme-corp/settings/visibility');
      await waitForLoadingToComplete(page);

      // Check visibility options
      await expect(page.locator('[data-testid="org-visibility-section"]')).toBeVisible();
      await expect(page.locator('[data-testid="visibility-private"]')).toBeChecked();
      
      // Change to public visibility
      await page.check('[data-testid="visibility-public"]');
      
      // Check member visibility options
      await expect(page.locator('[data-testid="member-visibility-section"]')).toBeVisible();
      await page.check('[data-testid="member-visibility-public"]');

      // Mock settings update
      await page.route('**/api/v1/organizations/org-1/settings', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.click('[data-testid="save-visibility-settings"]');
      await expect(page.locator('text=Visibility settings updated')).toBeVisible();
    });

    test('should configure organization policies', async ({ page }) => {
      await page.goto('/organizations/acme-corp/settings/policies');
      await waitForLoadingToComplete(page);

      // Verify policy sections
      await expect(page.locator('text=Organization Policies')).toBeVisible();
      await expect(page.locator('[data-testid="security-policies"]')).toBeVisible();
      await expect(page.locator('[data-testid="repository-policies"]')).toBeVisible();

      // Check two-factor requirement
      await expect(page.locator('[data-testid="require-2fa"]')).toBeChecked();
      
      // Enable additional policies
      await page.check('[data-testid="require-signed-commits"]');
      await page.check('[data-testid="restrict-repository-creation"]');
      await page.check('[data-testid="require-branch-protection"]');

      // Set repository defaults
      await page.selectOption('[data-testid="default-repo-visibility"]', 'private');
      await page.check('[data-testid="auto-delete-head-branches"]');

      await page.route('**/api/v1/organizations/org-1/policies', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.click('[data-testid="save-policies"]');
      await expect(page.locator('text=Policies updated successfully')).toBeVisible();
    });
  });

  test.describe('Team Hierarchy Management', () => {
    test('should display organization teams', async ({ page }) => {
      // Mock teams API
      await page.route('**/api/v1/organizations/org-1/teams', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'team-1',
                name: 'Engineering',
                slug: 'engineering',
                description: 'Software development teams',
                memberCount: 25,
                repositoryCount: 18,
                parent: null,
                privacy: 'closed'
              },
              {
                id: 'team-2',
                name: 'Frontend',
                slug: 'frontend',
                description: 'Frontend development team',
                memberCount: 8,
                repositoryCount: 6,
                parent: 'team-1',
                privacy: 'closed'
              },
              {
                id: 'team-3',
                name: 'Backend',
                slug: 'backend',
                description: 'Backend development team',
                memberCount: 12,
                repositoryCount: 10,
                parent: 'team-1',
                privacy: 'closed'
              }
            ]
          })
        });
      });

      await page.goto('/organizations/acme-corp/teams');
      await waitForLoadingToComplete(page);

      // Verify teams list
      await expect(page.locator('h1')).toContainText('Teams');
      await expect(page.locator('[data-testid="team-item"]')).toHaveCount(3);
      
      // Check team hierarchy display
      await expect(page.locator('text=Engineering')).toBeVisible();
      await expect(page.locator('text=Frontend')).toBeVisible();
      await expect(page.locator('text=Backend')).toBeVisible();
      
      // Verify parent-child relationships are shown
      await expect(page.locator('[data-testid="team-hierarchy"]')).toBeVisible();
    });

    test('should create new team', async ({ page }) => {
      await page.goto('/organizations/acme-corp/teams');
      await waitForLoadingToComplete(page);

      // Click create team button
      await page.click('[data-testid="create-team-button"]');
      await expect(page.locator('[data-testid="create-team-modal"]')).toBeVisible();

      // Fill team details
      await page.fill('[data-testid="team-name"]', 'DevOps');
      await page.fill('[data-testid="team-description"]', 'DevOps and Infrastructure team');
      await page.selectOption('[data-testid="parent-team"]', 'team-1'); // Engineering
      await page.selectOption('[data-testid="team-privacy"]', 'closed');

      // Mock team creation
      await page.route('**/api/v1/organizations/org-1/teams', async route => {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              id: 'team-4',
              name: 'DevOps',
              slug: 'devops',
              description: 'DevOps and Infrastructure team',
              parent: 'team-1'
            }
          })
        });
      });

      await page.click('[data-testid="create-team-submit"]');
      
      // Verify team was created
      await expect(page.locator('text=Team created successfully')).toBeVisible();
      await expect(page.locator('text=DevOps')).toBeVisible();
    });

    test('should manage team permissions', async ({ page }) => {
      await page.goto('/organizations/acme-corp/teams/engineering/settings');
      await waitForLoadingToComplete(page);

      // Mock team permissions API
      await page.route('**/api/v1/teams/team-1/permissions', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              repositories: 'read',
              teams: 'none',
              members: 'read',
              billing: 'none',
              settings: 'none'
            }
          })
        });
      });

      // Verify permissions section
      await expect(page.locator('text=Team Permissions')).toBeVisible();
      await expect(page.locator('[data-testid="permission-repositories"]')).toHaveValue('read');
      
      // Update permissions
      await page.selectOption('[data-testid="permission-repositories"]', 'write');
      await page.selectOption('[data-testid="permission-members"]', 'write');

      await page.route('**/api/v1/teams/team-1/permissions', async route => {
        if (route.request().method() === 'PUT') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await page.click('[data-testid="save-permissions"]');
      await expect(page.locator('text=Permissions updated')).toBeVisible();
    });
  });

  test.describe('Member Management', () => {
    test('should display organization members', async ({ page }) => {
      // Mock members API
      await page.route('**/api/v1/organizations/org-1/members', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'user-1',
                username: 'john-doe',
                name: 'John Doe',
                email: 'john@acme.com',
                role: 'owner',
                joinedAt: '2024-01-01T00:00:00Z',
                twoFactorEnabled: true,
                teams: ['Engineering', 'Backend']
              },
              {
                id: 'user-2',
                username: 'jane-smith',
                name: 'Jane Smith',
                email: 'jane@acme.com',
                role: 'admin',
                joinedAt: '2024-01-15T00:00:00Z',
                twoFactorEnabled: true,
                teams: ['Engineering', 'Frontend']
              },
              {
                id: 'user-3',
                username: 'bob-wilson',
                name: 'Bob Wilson',
                email: 'bob@acme.com',
                role: 'member',
                joinedAt: '2024-02-01T00:00:00Z',
                twoFactorEnabled: false,
                teams: ['Engineering']
              }
            ]
          })
        });
      });

      await page.goto('/organizations/acme-corp/members');
      await waitForLoadingToComplete(page);

      // Verify members list
      await expect(page.locator('h1')).toContainText('Members');
      await expect(page.locator('[data-testid="member-item"]')).toHaveCount(3);
      
      // Check member details
      await expect(page.locator('text=John Doe')).toBeVisible();
      await expect(page.locator('text=Owner')).toBeVisible();
      await expect(page.locator('text=jane-smith')).toBeVisible();
      await expect(page.locator('text=Admin')).toBeVisible();
      
      // Check 2FA status indicators
      await expect(page.locator('[data-testid="2fa-enabled"]')).toHaveCount(2);
      await expect(page.locator('[data-testid="2fa-disabled"]')).toHaveCount(1);
    });

    test('should invite new members', async ({ page }) => {
      await page.goto('/organizations/acme-corp/members');
      await waitForLoadingToComplete(page);

      // Click invite button
      await page.click('[data-testid="invite-member-button"]');
      await expect(page.locator('[data-testid="invite-modal"]')).toBeVisible();

      // Fill invitation details
      await page.fill('[data-testid="invite-email"]', 'new-member@example.com');
      await page.selectOption('[data-testid="invite-role"]', 'member');
      
      // Select teams
      await page.check('[data-testid="team-engineering"]');
      await page.check('[data-testid="team-frontend"]');

      // Mock invitation API
      await page.route('**/api/v1/organizations/org-1/invitations', async route => {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { id: 'invitation-1', email: 'new-member@example.com' }
          })
        });
      });

      await page.click('[data-testid="send-invitation"]');
      
      // Verify invitation sent
      await expect(page.locator('text=Invitation sent successfully')).toBeVisible();
    });

    test('should manage member roles', async ({ page }) => {
      await page.goto('/organizations/acme-corp/members');
      await waitForLoadingToComplete(page);

      // Click on member settings
      await page.click('[data-testid="member-settings-user-3"]');
      await expect(page.locator('[data-testid="member-settings-modal"]')).toBeVisible();

      // Change role
      await page.selectOption('[data-testid="member-role"]', 'admin');
      
      // Add to teams
      await page.check('[data-testid="team-frontend"]');

      // Mock role update
      await page.route('**/api/v1/organizations/org-1/members/user-3', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ success: true })
        });
      });

      await page.click('[data-testid="save-member-changes"]');
      await expect(page.locator('text=Member updated successfully')).toBeVisible();
    });

    test('should remove organization members', async ({ page }) => {
      await page.goto('/organizations/acme-corp/members');
      await waitForLoadingToComplete(page);

      // Click remove member
      await page.click('[data-testid="remove-member-user-3"]');
      await expect(page.locator('[data-testid="confirm-remove-modal"]')).toBeVisible();

      // Confirm removal
      await page.route('**/api/v1/organizations/org-1/members/user-3', async route => {
        if (route.request().method() === 'DELETE') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await page.click('[data-testid="confirm-remove-member"]');
      await expect(page.locator('text=Member removed successfully')).toBeVisible();
      
      // Verify member is removed from list
      await expect(page.locator('[data-testid="member-item"]')).toHaveCount(2);
    });
  });

  test.describe('Cross-Organization Collaboration', () => {
    test('should display organization collaborations', async ({ page }) => {
      // Mock collaborations API
      await page.route('**/api/v1/organizations/org-1/collaborations', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'collab-1',
                organization: { name: 'Tech Partners Inc', slug: 'tech-partners' },
                type: 'repository-sharing',
                repositories: ['project-alpha', 'shared-utils'],
                status: 'active',
                createdAt: '2024-01-01T00:00:00Z'
              },
              {
                id: 'collab-2',
                organization: { name: 'Design Agency', slug: 'design-agency' },
                type: 'team-collaboration',
                teams: ['Frontend', 'UX'],
                status: 'pending',
                createdAt: '2024-02-01T00:00:00Z'
              }
            ]
          })
        });
      });

      await page.goto('/organizations/acme-corp/collaborations');
      await waitForLoadingToComplete(page);

      // Verify collaborations list
      await expect(page.locator('h1')).toContainText('Collaborations');
      await expect(page.locator('[data-testid="collaboration-item"]')).toHaveCount(2);
      
      // Check collaboration details
      await expect(page.locator('text=Tech Partners Inc')).toBeVisible();
      await expect(page.locator('text=Repository Sharing')).toBeVisible();
      await expect(page.locator('text=Active')).toBeVisible();
      
      await expect(page.locator('text=Design Agency')).toBeVisible();
      await expect(page.locator('text=Team Collaboration')).toBeVisible();
      await expect(page.locator('text=Pending')).toBeVisible();
    });

    test('should create new collaboration', async ({ page }) => {
      await page.goto('/organizations/acme-corp/collaborations');
      await waitForLoadingToComplete(page);

      // Click create collaboration
      await page.click('[data-testid="create-collaboration-button"]');
      await expect(page.locator('[data-testid="collaboration-modal"]')).toBeVisible();

      // Fill collaboration details
      await page.fill('[data-testid="partner-organization"]', 'external-org');
      await page.selectOption('[data-testid="collaboration-type"]', 'repository-sharing');
      
      // Select repositories to share
      await page.check('[data-testid="repo-project-beta"]');
      await page.check('[data-testid="repo-common-libs"]');

      // Mock collaboration creation
      await page.route('**/api/v1/organizations/org-1/collaborations', async route => {
        await route.fulfill({
          status: 201,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: { id: 'collab-3', status: 'pending' }
          })
        });
      });

      await page.click('[data-testid="create-collaboration"]');
      await expect(page.locator('text=Collaboration request sent')).toBeVisible();
    });
  });

  test.describe('Organization Analytics', () => {
    test('should display organization-wide analytics', async ({ page }) => {
      // Mock organization analytics API
      await page.route('**/api/v1/organizations/org-1/analytics', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              overview: {
                totalRepositories: 28,
                totalMembers: 45,
                totalTeams: 8,
                totalCommits: 1234,
                                  totalPullRequests: 567
              },
              activity: {
                commitsThisMonth: 156,
                                  pullRequestsThisMonth: 45
              },
              topContributors: [
                { name: 'John Doe', contributions: 89 },
                { name: 'Jane Smith', contributions: 76 }
              ],
              languageDistribution: [
                { language: 'TypeScript', percentage: 45 },
                { language: 'Python', percentage: 30 },
                { language: 'Go', percentage: 25 }
              ]
            }
          })
        });
      });

      await page.goto('/organizations/acme-corp/analytics');
      await waitForLoadingToComplete(page);

      // Verify analytics dashboard
      await expect(page.locator('h1')).toContainText('Organization Analytics');
      
      // Check overview metrics
      await expect(page.locator('text=28 Repositories')).toBeVisible();
      await expect(page.locator('text=45 Members')).toBeVisible();
      await expect(page.locator('text=8 Teams')).toBeVisible();
      
      // Check activity metrics
      await expect(page.locator('text=156 Commits This Month')).toBeVisible();
      await expect(page.locator('text=45 PRs This Month')).toBeVisible();
      
      // Verify charts are present
      await expect(page.locator('[data-testid="activity-chart"]')).toBeVisible();
      await expect(page.locator('[data-testid="language-chart"]')).toBeVisible();
      
      // Check top contributors
      await expect(page.locator('text=Top Contributors')).toBeVisible();
      await expect(page.locator('text=John Doe')).toBeVisible();
      await expect(page.locator('text=Jane Smith')).toBeVisible();
    });

    test('should filter analytics by time period', async ({ page }) => {
      await page.goto('/organizations/acme-corp/analytics');
      await waitForLoadingToComplete(page);

      // Test time period filters
      const periods = ['7d', '30d', '90d', '1y'];
      
      for (const period of periods) {
        await page.click(`[data-testid="period-${period}"]`);
        await waitForLoadingToComplete(page);
        
        // Verify active state
        await expect(page.locator(`[data-testid="period-${period}"]`)).toHaveClass(/active/);
      }
    });

    test('should export organization analytics', async ({ page }) => {
      await page.goto('/organizations/acme-corp/analytics');
      await waitForLoadingToComplete(page);

      // Mock download
      const downloadPromise = page.waitForEvent('download');

      await page.click('[data-testid="export-analytics-button"]');
      await page.click('[data-testid="export-pdf"]');

      const download = await downloadPromise;
      expect(download.suggestedFilename()).toContain('organization-analytics');
      expect(download.suggestedFilename()).toContain('.pdf');
    });
  });

  test.describe('Audit Logs and Compliance', () => {
    test('should display organization audit logs', async ({ page }) => {
      // Mock audit logs API
      await page.route('**/api/v1/organizations/org-1/audit-logs', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              {
                id: 'log-1',
                event: 'member.invited',
                actor: 'john-doe',
                target: 'new-member@example.com',
                timestamp: '2024-01-15T10:30:00Z',
                details: { role: 'member', teams: ['Engineering'] }
              },
              {
                id: 'log-2',
                event: 'repository.created',
                actor: 'jane-smith',
                target: 'new-project',
                timestamp: '2024-01-15T09:15:00Z',
                details: { visibility: 'private' }
              },
              {
                id: 'log-3',
                event: 'team.created',
                actor: 'john-doe',
                target: 'DevOps',
                timestamp: '2024-01-14T16:45:00Z',
                details: { parent: 'Engineering' }
              }
            ]
          })
        });
      });

      await page.goto('/organizations/acme-corp/audit-logs');
      await waitForLoadingToComplete(page);

      // Verify audit logs display
      await expect(page.locator('h1')).toContainText('Audit Logs');
      await expect(page.locator('[data-testid="audit-log-item"]')).toHaveCount(3);
      
      // Check log details
      await expect(page.locator('text=Member invited')).toBeVisible();
      await expect(page.locator('text=Repository created')).toBeVisible();
      await expect(page.locator('text=Team created')).toBeVisible();
      
      // Verify timestamps and actors
      await expect(page.locator('text=john-doe')).toHaveCount(2);
      await expect(page.locator('text=jane-smith')).toBeVisible();
    });

    test('should filter audit logs', async ({ page }) => {
      await page.goto('/organizations/acme-corp/audit-logs');
      await waitForLoadingToComplete(page);

      // Filter by event type
      await page.selectOption('[data-testid="event-filter"]', 'member');
      await waitForLoadingToComplete(page);
      
      // Should only show member-related events
      await expect(page.locator('text=Member invited')).toBeVisible();
      await expect(page.locator('text=Repository created')).not.toBeVisible();

      // Filter by actor
      await page.fill('[data-testid="actor-filter"]', 'john-doe');
      await page.click('[data-testid="apply-filters"]');
      await waitForLoadingToComplete(page);

      // Filter by date range
      await page.fill('[data-testid="start-date"]', '2024-01-14');
      await page.fill('[data-testid="end-date"]', '2024-01-15');
      await page.click('[data-testid="apply-filters"]');
      await waitForLoadingToComplete(page);
    });

    test('should export audit logs', async ({ page }) => {
      await page.goto('/organizations/acme-corp/audit-logs');
      await waitForLoadingToComplete(page);

      // Mock download
      const downloadPromise = page.waitForEvent('download');

      await page.click('[data-testid="export-logs-button"]');
      await page.click('[data-testid="export-csv"]');

      const download = await downloadPromise;
      expect(download.suggestedFilename()).toContain('audit-logs');
      expect(download.suggestedFilename()).toContain('.csv');
    });
  });

  test.describe('Mobile Organization Management', () => {
    test('should handle mobile organization interface', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      await page.goto('/organizations/acme-corp');
      await waitForLoadingToComplete(page);

      // Verify mobile-friendly navigation
      await expect(page.locator('[data-testid="mobile-org-menu"]')).toBeVisible();
      
      // Test collapsible sections
      await page.click('[data-testid="teams-section-toggle"]');
      await expect(page.locator('[data-testid="teams-list"]')).toBeVisible();
      
      await page.click('[data-testid="members-section-toggle"]');
      await expect(page.locator('[data-testid="members-list"]')).toBeVisible();

      // Verify responsive member cards
      await expect(page.locator('[data-testid="member-card"]').first()).toBeVisible();
      
      // Test mobile settings access
      await page.click('[data-testid="mobile-settings-button"]');
      await expect(page.locator('[data-testid="mobile-settings-menu"]')).toBeVisible();
    });
  });
});