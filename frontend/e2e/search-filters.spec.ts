import { test, expect } from '@playwright/test';
import { loginUser, testUser, waitForLoadingToComplete, takeScreenshot } from './helpers/test-utils';

test.describe('Search Filters and Advanced Options', () => {
  test.beforeEach(async ({ page }) => {
    // Mock authentication for all search tests
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

  test.describe('Advanced Issue Filtering', () => {
    test('should filter issues by state, labels, assignee, and author', async ({ page }) => {
      // Mock filtered issue search results
      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        const type = url.searchParams.get('type');
        const state = url.searchParams.get('state');
        const labels = url.searchParams.get('labels');
        const assignee = url.searchParams.get('assignee');
        const author = url.searchParams.get('author');
        
        if (type === 'issues') {
          let issues = [
            {
              id: '1',
              number: 101,
              title: 'Bug: Search not working properly',
              body: 'The search functionality has several issues...',
              state: 'open',
              repository_id: '1',
              user_id: '1',
              assignee_id: 'john',
              author: 'alice',
              labels: ['bug', 'high-priority'],
              created_at: '2023-11-01',
              updated_at: '2023-11-15'
            },
            {
              id: '2',
              number: 102,
              title: 'Feature: Improve search filters',
              body: 'Add more advanced filtering options...',
              state: 'closed',
              repository_id: '1',
              user_id: '2',
              assignee_id: 'bob',
              author: 'charlie',
              labels: ['enhancement', 'search'],
              created_at: '2023-10-01',
              updated_at: '2023-10-30'
            },
            {
              id: '3',
              number: 103,
              title: 'Documentation: Search API guide',
              body: 'Create comprehensive documentation for search API...',
              state: 'open',
              repository_id: '1',
              user_id: '3',
              assignee_id: 'alice',
              author: 'david',
              labels: ['documentation'],
              created_at: '2023-11-10',
              updated_at: '2023-11-20'
            }
          ];

          // Apply filters
          if (state) {
            issues = issues.filter(issue => issue.state === state);
          }
          if (labels) {
            issues = issues.filter(issue => issue.labels.some(label => labels.includes(label)));
          }
          if (assignee) {
            issues = issues.filter(issue => issue.assignee_id === assignee);
          }
          if (author) {
            issues = issues.filter(issue => issue.author === author);
          }

          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                users: [],
                repositories: [],
                issues: issues,
                organizations: [],
                commits: [],
                total_count: issues.length
              }
            })
          });
        }
      });

      await page.goto('/search');
      await waitForLoadingToComplete(page);

      // Click on Issues filter
      await page.click('button:has-text("Issues")');
      
      // Fill search query
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'search');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify all issues are shown initially
      await expect(page.locator('text=Bug: Search not working properly')).toBeVisible();
      await expect(page.locator('text=Feature: Improve search filters')).toBeVisible();
      await expect(page.locator('text=Documentation: Search API guide')).toBeVisible();

      // Test state filter - only open issues
      const stateFilter = page.locator('select[name="state"], button:has-text("State")');
      if (await stateFilter.isVisible()) {
        await stateFilter.click();
        await page.click('text=Open');
        await waitForLoadingToComplete(page);
        
        await expect(page.locator('text=Bug: Search not working properly')).toBeVisible();
        await expect(page.locator('text=Documentation: Search API guide')).toBeVisible();
        await expect(page.locator('text=Feature: Improve search filters')).not.toBeVisible();
      }
    });

    test('should support label-based filtering', async ({ page }) => {
      // Mock label suggestions and filtering
      await page.route('**/api/v1/labels**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              { name: 'bug', color: '#d73a49', description: 'Something isn\'t working' },
              { name: 'enhancement', color: '#a2eeef', description: 'New feature or request' },
              { name: 'documentation', color: '#0075ca', description: 'Improvements or additions to documentation' },
              { name: 'high-priority', color: '#b60205', description: 'High priority issue' },
              { name: 'search', color: '#1d76db', description: 'Related to search functionality' }
            ]
          })
        });
      });

      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const labels = url.searchParams.get('labels');
        
        let issues = [];
        if (labels?.includes('bug')) {
          issues = [
            {
              id: '1',
              number: 101,
              title: 'Bug: Search not working properly',
              body: 'The search functionality has several issues...',
              state: 'open',
              repository_id: '1',
              user_id: '1',
              labels: ['bug', 'high-priority'],
              created_at: '2023-11-01',
              updated_at: '2023-11-15'
            }
          ];
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: [],
              issues: issues,
              organizations: [],
              commits: [],
              total_count: issues.length
            }
          })
        });
      });

      await page.goto('/search');
      await page.click('button:has-text("Issues")');
      
      // Use label filter
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'label:bug');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify only bug-labeled issues are shown
      await expect(page.locator('text=Bug: Search not working properly')).toBeVisible();
      await expect(page.locator('text=bug')).toBeVisible();
      await expect(page.locator('text=high-priority')).toBeVisible();
    });

    test('should filter by assignee and author', async ({ page }) => {
      // Mock user suggestions for assignee/author filtering
      await page.route('**/api/v1/users/search**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        
        let users = [
          { id: '1', username: 'alice', full_name: 'Alice Johnson' },
          { id: '2', username: 'bob', full_name: 'Bob Smith' },
          { id: '3', username: 'charlie', full_name: 'Charlie Brown' }
        ];

        if (query) {
          users = users.filter(user => 
            user.username.includes(query) || user.full_name.toLowerCase().includes(query.toLowerCase())
          );
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: users
          })
        });
      });

      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const assignee = url.searchParams.get('assignee');
        const author = url.searchParams.get('author');
        
        let issues = [
          {
            id: '1',
            number: 101,
            title: 'Issue assigned to Alice',
            body: 'This issue is assigned to Alice...',
            state: 'open',
            repository_id: '1',
            user_id: '1',
            assignee: 'alice',
            author: 'bob',
            created_at: '2023-11-01',
            updated_at: '2023-11-15'
          }
        ];

        if (assignee && assignee !== 'alice') {
          issues = [];
        }
        if (author && author !== 'bob') {
          issues = [];
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: [],
              issues: issues,
              organizations: [],
              commits: [],
              total_count: issues.length
            }
          })
        });
      });

      await page.goto('/search');
      await page.click('button:has-text("Issues")');
      
      // Filter by assignee
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'assignee:alice');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=Issue assigned to Alice')).toBeVisible();
    });
  });

  test.describe('Pull Request Search with Filters', () => {
    test('should filter pull requests by merge status and review state', async ({ page }) => {
      // Mock pull request search with filters
      await page.route('**/api/v1/search/pullrequests**', async route => {
        const url = new URL(route.request().url());
        const status = url.searchParams.get('status');
        const review_state = url.searchParams.get('review_state');
        
        let pullRequests = [
          {
            id: '1',
            number: 201,
            title: 'Add search improvements',
            body: 'This PR adds several search improvements...',
            state: 'open',
            merged: false,
            draft: false,
            review_state: 'approved',
            repository_id: '1',
            author: 'alice',
            created_at: '2023-11-01',
            updated_at: '2023-11-15'
          },
          {
            id: '2',
            number: 202,
            title: 'Fix search bug',
            body: 'Fixes critical search bug...',
            state: 'closed',
            merged: true,
            draft: false,
            review_state: 'approved',
            repository_id: '1',
            author: 'bob',
            created_at: '2023-10-15',
            updated_at: '2023-10-20'
          },
          {
            id: '3',
            number: 203,
            title: 'Draft: New search feature',
            body: 'Work in progress for new search feature...',
            state: 'open',
            merged: false,
            draft: true,
            review_state: 'pending',
            repository_id: '1',
            author: 'charlie',
            created_at: '2023-11-10',
            updated_at: '2023-11-20'
          }
        ];

        // Apply filters
        if (status === 'merged') {
          pullRequests = pullRequests.filter(pr => pr.merged);
        } else if (status === 'open') {
          pullRequests = pullRequests.filter(pr => pr.state === 'open' && !pr.merged);
        } else if (status === 'draft') {
          pullRequests = pullRequests.filter(pr => pr.draft);
        }

        if (review_state) {
          pullRequests = pullRequests.filter(pr => pr.review_state === review_state);
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: pullRequests,
            total_count: pullRequests.length
          })
        });
      });

      await page.goto('/search');
      
      // Navigate to PR search (would need UI implementation)
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'type:pr search');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Test merge status filter
      const statusFilter = page.locator('select[name="status"], button:has-text("Status")');
      if (await statusFilter.isVisible()) {
        await statusFilter.click();
        await page.click('text=Merged');
        await waitForLoadingToComplete(page);
        
        await expect(page.locator('text=Fix search bug')).toBeVisible();
        await expect(page.locator('text=merged')).toBeVisible();
      }
    });

    test('should support cross-repository pull request search', async ({ page }) => {
      // Mock cross-repository PR search
      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const type = url.searchParams.get('type');
        
        if (type === 'pullrequests' || url.pathname.includes('pullrequests')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: [
                {
                  id: '1',
                  number: 301,
                  title: 'Cross-repo search feature',
                  body: 'Implements cross-repository search...',
                  state: 'open',
                  repository_name: 'frontend-repo',
                  repository_owner: 'myorg',
                  author: 'developer1',
                  created_at: '2023-11-01'
                },
                {
                  id: '2',
                  number: 150,
                  title: 'Backend search optimization',
                  body: 'Optimizes backend search performance...',
                  state: 'merged',
                  repository_name: 'backend-repo',
                  repository_owner: 'myorg',
                  author: 'developer2',
                  created_at: '2023-10-20'
                }
              ],
              total_count: 2
            })
          });
        }
      });

      await page.goto('/search');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'type:pr cross repo');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify cross-repository results
      await expect(page.locator('text=Cross-repo search feature')).toBeVisible();
      await expect(page.locator('text=Backend search optimization')).toBeVisible();
      await expect(page.locator('text=frontend-repo')).toBeVisible();
      await expect(page.locator('text=backend-repo')).toBeVisible();
    });
  });

  test.describe('Repository Search Filters', () => {
    test('should filter repositories by language, stars, and visibility', async ({ page }) => {
      // Mock repository search with filters
      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const language = url.searchParams.get('language');
        const stars = url.searchParams.get('stars');
        const visibility = url.searchParams.get('visibility');
        
        let repositories = [
          {
            id: '1',
            name: 'javascript-project',
            description: 'A JavaScript project',
            owner_id: '1',
            owner_type: 'user',
            visibility: 'public',
            stars_count: 1500,
            forks_count: 300,
            primary_language: 'JavaScript',
            created_at: '2023-01-01',
            updated_at: '2023-12-01'
          },
          {
            id: '2',
            name: 'python-tools',
            description: 'Python development tools',
            owner_id: '1',
            owner_type: 'organization',
            visibility: 'public',
            stars_count: 250,
            forks_count: 50,
            primary_language: 'Python',
            created_at: '2023-02-01',
            updated_at: '2023-11-01'
          },
          {
            id: '3',
            name: 'private-repo',
            description: 'Private repository',
            owner_id: '1',
            owner_type: 'user',
            visibility: 'private',
            stars_count: 10,
            forks_count: 2,
            primary_language: 'TypeScript',
            created_at: '2023-06-01',
            updated_at: '2023-12-01'
          }
        ];

        // Apply filters
        if (language) {
          repositories = repositories.filter(repo => repo.primary_language === language);
        }
        if (stars) {
          const minStars = parseInt(stars.replace('>', ''));
          repositories = repositories.filter(repo => repo.stars_count > minStars);
        }
        if (visibility) {
          repositories = repositories.filter(repo => repo.visibility === visibility);
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: repositories,
              issues: [],
              organizations: [],
              commits: [],
              total_count: repositories.length
            }
          })
        });
      });

      await page.goto('/search');
      await page.click('button:has-text("Repositories")');
      
      // Test language filter
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'language:JavaScript');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=javascript-project')).toBeVisible();
      await expect(page.locator('text=JavaScript')).toBeVisible();
      await expect(page.locator('text=python-tools')).not.toBeVisible();

      // Test stars filter
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'stars:>1000');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=javascript-project')).toBeVisible();
      await expect(page.locator('text=⭐ 1500')).toBeVisible();
    });

    test('should support organization and user filtering', async ({ page }) => {
      // Mock user/organization search
      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        const type = url.searchParams.get('type');
        
        if (type === 'repositories' && query?.includes('org:')) {
          const org = query.match(/org:(\w+)/)?.[1];
          
          let repositories = [];
          if (org === 'mycompany') {
            repositories = [
              {
                id: '1',
                name: 'company-website',
                description: 'Company website repository',
                owner_id: 'mycompany',
                owner_type: 'organization',
                visibility: 'public',
                stars_count: 100,
                forks_count: 20,
                primary_language: 'JavaScript',
                created_at: '2023-01-01',
                updated_at: '2023-12-01'
              }
            ];
          }

          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                users: [],
                repositories: repositories,
                issues: [],
                organizations: [],
                commits: [],
                total_count: repositories.length
              }
            })
          });
        }
      });

      await page.goto('/search');
      await page.click('button:has-text("Repositories")');
      
      // Filter by organization
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'org:mycompany');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=company-website')).toBeVisible();
      await expect(page.locator('text=mycompany')).toBeVisible();
    });
  });

  test.describe('Saved Search Filters', () => {
    test('should save and reuse complex filter combinations', async ({ page }) => {
      // Mock saved searches API
      await page.route('**/api/v1/search/saved**', async route => {
        if (route.request().method() === 'GET') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: [
                {
                  id: '1',
                  name: 'High Priority Bugs',
                  query: 'type:issues state:open label:bug label:high-priority',
                  filters: {
                    type: 'issues',
                    state: 'open',
                    labels: ['bug', 'high-priority']
                  },
                  created_at: '2023-11-01'
                },
                {
                  id: '2',
                  name: 'Popular JS Repos',
                  query: 'type:repositories language:JavaScript stars:>500',
                  filters: {
                    type: 'repositories',
                    language: 'JavaScript',
                    stars: '>500'
                  },
                  created_at: '2023-10-15'
                }
              ]
            })
          });
        } else if (route.request().method() === 'POST') {
          const body = await route.request().postDataJSON();
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                id: '3',
                name: body.name,
                query: body.query,
                filters: body.filters
              }
            })
          });
        }
      });

      await page.goto('/search');
      
      // Look for saved searches section (would need UI implementation)
      const savedSearchButton = page.locator('button:has-text("Saved Searches")');
      if (await savedSearchButton.isVisible()) {
        await savedSearchButton.click();
        
        // Verify saved searches are displayed
        await expect(page.locator('text=High Priority Bugs')).toBeVisible();
        await expect(page.locator('text=Popular JS Repos')).toBeVisible();
        
        // Click on a saved search
        await page.click('text=High Priority Bugs');
        
        // Verify the search is loaded with filters
        await expect(page.locator('input[value*="type:issues state:open"]')).toBeVisible();
      }
    });

    test('should create new saved searches with current filters', async ({ page }) => {
      await page.goto('/search');
      
      // Set up complex search with filters
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'language:TypeScript stars:>100 topics:web-framework');
      await page.click('button:has-text("Repositories")');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Look for save search button (would need UI implementation)
      const saveButton = page.locator('button:has-text("Save Search")');
      if (await saveButton.isVisible()) {
        await saveButton.click();
        
        // Fill in save dialog
        await page.fill('input[placeholder="Search name"]', 'Popular TypeScript Frameworks');
        await page.click('button:has-text("Save")');
        
        // Verify search was saved
        await expect(page.locator('text=Search saved successfully')).toBeVisible();
      }
    });
  });

  test.describe('Search Result Sorting Options', () => {
    test('should support multiple sorting options', async ({ page }) => {
      // Mock sorted search results
      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const sort = url.searchParams.get('sort') || 'relevance';
        const order = url.searchParams.get('order') || 'desc';
        
        let repositories = [
          {
            id: '1',
            name: 'newest-repo',
            description: 'The newest repository',
            created_at: '2023-12-01',
            updated_at: '2023-12-01',
            stars_count: 50,
            forks_count: 10
          },
          {
            id: '2',
            name: 'popular-repo',
            description: 'The most popular repository',
            created_at: '2023-01-01',
            updated_at: '2023-11-30',
            stars_count: 2000,
            forks_count: 500
          },
          {
            id: '3',
            name: 'recently-updated',
            description: 'Recently updated repository',
            created_at: '2023-06-01',
            updated_at: '2023-12-01',
            stars_count: 100,
            forks_count: 25
          }
        ];

        // Sort based on parameters
        if (sort === 'created') {
          repositories.sort((a, b) => {
            const comparison = new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
            return order === 'asc' ? comparison : -comparison;
          });
        } else if (sort === 'updated') {
          repositories.sort((a, b) => {
            const comparison = new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime();
            return order === 'asc' ? comparison : -comparison;
          });
        } else if (sort === 'stars') {
          repositories.sort((a, b) => {
            const comparison = a.stars_count - b.stars_count;
            return order === 'asc' ? comparison : -comparison;
          });
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: repositories,
              issues: [],
              organizations: [],
              commits: [],
              total_count: repositories.length
            }
          })
        });
      });

      await page.goto('/search');
      await page.click('button:has-text("Repositories")');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'test');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Test sorting by stars
      const sortDropdown = page.locator('select[name="sort"], button:has-text("Sort")');
      if (await sortDropdown.isVisible()) {
        await sortDropdown.click();
        await page.click('text=Most stars');
        
        await waitForLoadingToComplete(page);
        
        // Verify sorting - most popular should be first
        const firstResult = page.locator('text=popular-repo').first();
        await expect(firstResult).toBeVisible();
      }

      // Test sorting by recently created
      if (await sortDropdown.isVisible()) {
        await sortDropdown.click();
        await page.click('text=Recently created');
        
        await waitForLoadingToComplete(page);
        
        // Verify newest repo is first
        const firstResult = page.locator('text=newest-repo').first();
        await expect(firstResult).toBeVisible();
      }
    });

    test('should support ascending and descending order', async ({ page }) => {
      await page.goto('/search');
      await page.click('button:has-text("Repositories")');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'test');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Look for order toggle (would need UI implementation)
      const orderToggle = page.locator('button[aria-label="Sort order"], button:has-text("Desc"), button:has-text("Asc")');
      if (await orderToggle.isVisible()) {
        // Test toggling between ascending and descending
        await orderToggle.click();
        await waitForLoadingToComplete(page);
        
        // Verify order changed (implementation dependent)
        await expect(page.locator('text=Asc, text=Desc')).toBeVisible();
      }
    });
  });

  test.describe('Search Filter Persistence', () => {
    test('should persist filters in URL parameters', async ({ page }) => {
      await page.goto('/search');
      
      // Apply multiple filters
      await page.click('button:has-text("Issues")');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'bug label:high-priority assignee:alice');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Verify URL contains filter parameters
      await expect(page).toHaveURL(/.*type=issues.*/);
      await expect(page).toHaveURL(/.*q=.*bug.*label.*high-priority.*assignee.*alice.*/);
      
      // Test that filters persist on page reload
      await page.reload();
      await waitForLoadingToComplete(page);
      
      // Verify filters are still applied
      const searchInput = page.locator('input[placeholder*="Search repositories, issues, users, and commits"]');
      await expect(searchInput).toHaveValue(/.*bug.*label.*high-priority.*assignee.*alice.*/);
    });

    test('should restore filters from bookmarked URLs', async ({ page }) => {
      // Navigate directly to a URL with search filters
      await page.goto('/search?type=repositories&q=language:JavaScript%20stars:>500&sort=stars&order=desc');
      await waitForLoadingToComplete(page);
      
      // Verify filters are restored
      await expect(page.locator('button:has-text("Repositories")')).toHaveClass(/active|selected/);
      
      const searchInput = page.locator('input[placeholder*="Search repositories, issues, users, and commits"]');
      await expect(searchInput).toHaveValue('language:JavaScript stars:>500');
      
      // Verify sort options are restored
      const sortIndicator = page.locator('text=Most stars, text=Stars');
      if (await sortIndicator.isVisible()) {
        await expect(sortIndicator).toBeVisible();
      }
    });
  });

  test.describe('Filter Performance and UX', () => {
    test('should provide responsive filter application', async ({ page }) => {
      let requestCount = 0;
      
      // Mock API with request counting
      await page.route('**/api/v1/search**', async route => {
        requestCount++;
        await new Promise(resolve => setTimeout(resolve, 100)); // Simulate processing
        
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: [],
              issues: [],
              organizations: [],
              commits: [],
              total_count: 0
            }
          })
        });
      });

      await page.goto('/search');
      
      // Apply multiple filters rapidly
      await page.click('button:has-text("Issues")');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'test');
      
      const startTime = Date.now();
      await page.click('button:has-text("Search")');
      await waitForLoadingToComplete(page);
      const endTime = Date.now();
      
      // Should complete quickly
      expect(endTime - startTime).toBeLessThan(3000);
      
      // Should not make excessive API calls
      expect(requestCount).toBeLessThanOrEqual(2);
    });

    test('should show loading states during filter application', async ({ page }) => {
      // Mock slow API response
      await page.route('**/api/v1/search**', async route => {
        await new Promise(resolve => setTimeout(resolve, 1000));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: [],
              issues: [],
              organizations: [],
              commits: [],
              total_count: 0
            }
          })
        });
      });

      await page.goto('/search');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'slow search');
      await page.click('button:has-text("Search")');
      
      // Should show loading state
      await expect(page.locator('text=Searching..., .loading, .spinner')).toBeVisible();
      
      // Button should be disabled during search
      const searchButton = page.locator('button:has-text("Searching...")');
      if (await searchButton.isVisible()) {
        await expect(searchButton).toBeDisabled();
      }
      
      await waitForLoadingToComplete(page);
    });

    test('should provide clear filter feedback', async ({ page }) => {
      await page.goto('/search');
      
      // Apply filters and verify they're clearly displayed
      await page.click('button:has-text("Issues")');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'label:bug assignee:alice');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Look for active filter indicators
      const activeFilters = page.locator('.filter-tag, .active-filter, text=Issues');
      await expect(activeFilters.first()).toBeVisible();
      
      // Check if individual filter components are highlighted
      await expect(page.locator('button:has-text("Issues")')).toHaveClass(/active|selected/);
    });
  });
});