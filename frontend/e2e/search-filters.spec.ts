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
            title: 'Fix authentication bug',
            body: 'This PR fixes the authentication bug...',
            state: 'closed',
            merged: true,
            draft: false,
            review_state: 'changes_requested',
            repository_id: '1',
            author: 'bob',
            created_at: '2023-10-15',
            updated_at: '2023-10-30'
          },
          {
            id: '3',
            number: 203,
            title: 'Update documentation',
            body: 'Update API documentation...',
            state: 'open',
            merged: false,
            draft: true,
            review_state: 'pending',
            repository_id: '2',
            author: 'charlie',
            created_at: '2023-09-01',
            updated_at: '2023-09-05'
          }
        ];

        // Apply filters
        if (status === 'merged') {
          pullRequests = pullRequests.filter(pr => pr.merged);
        } else if (status === 'open') {
          pullRequests = pullRequests.filter(pr => pr.state === 'open');
        }
        
        if (review_state) {
          pullRequests = pullRequests.filter(pr => pr.review_state === review_state);
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              results: {
                pull_requests: pullRequests,
                repositories: [],
                users: []
              },
              total_count: pullRequests.length
            }
          })
        });
      });

      await page.goto('/search');
      
      // Search for pull requests
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'type:pr search');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      // Verify all PRs are shown initially
      await expect(page.locator('text=Add search improvements')).toBeVisible();
      await expect(page.locator('text=Fix authentication bug')).toBeVisible();
      await expect(page.locator('text=Update documentation')).toBeVisible();

      // Test merge status filter
      await page.selectOption('select[name="status"]', 'merged');
      await page.click('button:has-text("Apply Filters")');
      
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=Fix authentication bug')).toBeVisible();
      await expect(page.locator('text=Add search improvements')).not.toBeVisible();
      await expect(page.locator('text=Update documentation')).not.toBeVisible();
    });

    test('should support cross-repository pull request search', async ({ page }) => {
      // Mock cross-repository search
      await page.route('**/api/v1/search**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              results: {
                pull_requests: [
                  {
                    id: '1',
                    number: 201,
                    title: 'Cross-repo feature implementation',
                    body: 'Implements cross-repository search functionality...',
                    state: 'open',
                    merged: false,
                    repository: {
                      id: '1',
                      name: 'main-repo',
                      full_name: 'org/main-repo'
                    },
                    author: 'alice',
                    created_at: '2023-11-01'
                  },
                  {
                    id: '2',
                    number: 105,
                    title: 'Add cross-repo support',
                    body: 'Adds support for searching across repositories...',
                    state: 'closed',
                    merged: true,
                    repository: {
                      id: '2',
                      name: 'helper-repo',
                      full_name: 'org/helper-repo'
                    },
                    author: 'bob',
                    created_at: '2023-10-15'
                  }
                ],
                repositories: [],
                users: []
              },
              total_count: 2
            }
          })
        });
      });

      await page.goto('/search');
      
      // Search across repositories
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'type:pr cross repo');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      // Verify cross-repo results
      await expect(page.locator('text=Cross-repo feature implementation')).toBeVisible();
      await expect(page.locator('text=Add cross-repo support')).toBeVisible();
      await expect(page.locator('text=org/main-repo')).toBeVisible();
      await expect(page.locator('text=org/helper-repo')).toBeVisible();
    });
  });

  test.describe('Repository Search Filters', () => {
    test('should filter repositories by language, stars, and visibility', async ({ page }) => {
      // Mock repository search with filters
      await page.route('**/api/v1/search/repositories**', async route => {
        const url = new URL(route.request().url());
        const language = url.searchParams.get('language');
        const stars = url.searchParams.get('stars');
        const visibility = url.searchParams.get('visibility');
        
        let repositories = [
          {
            id: '1',
            name: 'awesome-js-lib',
            full_name: 'org/awesome-js-lib',
            description: 'An awesome JavaScript library',
            language: 'JavaScript',
            stargazers_count: 1500,
            private: false,
            updated_at: '2023-11-01'
          },
          {
            id: '2',
            name: 'python-utils',
            full_name: 'org/python-utils',
            description: 'Python utility functions',
            language: 'Python',
            stargazers_count: 800,
            private: true,
            updated_at: '2023-10-15'
          },
          {
            id: '3',
            name: 'typescript-framework',
            full_name: 'org/typescript-framework',
            description: 'Modern TypeScript framework',
            language: 'TypeScript',
            stargazers_count: 2000,
            private: false,
            updated_at: '2023-09-20'
          }
        ];

        // Apply filters
        if (language) {
          repositories = repositories.filter(repo => repo.language === language);
        }
        
        if (stars) {
          const starsValue = parseInt(stars.replace('>', ''));
          repositories = repositories.filter(repo => repo.stargazers_count > starsValue);
        }
        
        if (visibility === 'public') {
          repositories = repositories.filter(repo => !repo.private);
        } else if (visibility === 'private') {
          repositories = repositories.filter(repo => repo.private);
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              results: {
                repositories: repositories,
                pull_requests: [],
                users: []
              },
              total_count: repositories.length
            }
          })
        });
      });

      await page.goto('/search');
      
      // Search for repositories with language filter
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'language:JavaScript');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=awesome-js-lib')).toBeVisible();
      await expect(page.locator('text=JavaScript')).toBeVisible();

      // Test stars filter
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'stars:>1000');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=awesome-js-lib')).toBeVisible();
      await expect(page.locator('text=typescript-framework')).toBeVisible();
      await expect(page.locator('text=python-utils')).not.toBeVisible();
    });

    test('should support organization and user filtering', async ({ page }) => {
      // Mock organization/user filtered search
      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        
        let repositories = [];
        let users = [];
        
        if (query && query.includes('org:mycompany')) {
          repositories = [
            {
              id: '1',
              name: 'company-frontend',
              full_name: 'mycompany/company-frontend',
              description: 'Company frontend application',
              language: 'React',
              stargazers_count: 150,
              private: false,
              organization: {
                id: '1',
                name: 'mycompany',
                avatar_url: '/avatars/mycompany.png'
              }
            }
          ];
        } else if (query && query.includes('user:johndoe')) {
          repositories = [
            {
              id: '2',
              name: 'personal-blog',
              full_name: 'johndoe/personal-blog',
              description: 'Personal blog built with Next.js',
              language: 'JavaScript',
              stargazers_count: 25,
              private: false,
              owner: {
                id: '2',
                username: 'johndoe',
                avatar_url: '/avatars/johndoe.png'
              }
            }
          ];
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              results: {
                repositories: repositories,
                pull_requests: [],
                users: users
              },
              total_count: repositories.length + users.length
            }
          })
        });
      });

      await page.goto('/search');
      
      // Search within organization
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'org:mycompany');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=company-frontend')).toBeVisible();
      await expect(page.locator('text=mycompany')).toBeVisible();

      // Search by user
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'user:johndoe');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      await expect(page.locator('text=personal-blog')).toBeVisible();
      await expect(page.locator('text=johndoe')).toBeVisible();
    });
  });

  test.describe('Saved Search Filters', () => {
    test('should create new saved searches with current filters', async ({ page }) => {
      // Mock saved search creation
      await page.route('**/api/v1/saved-searches', async route => {
        if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                id: '1',
                name: 'My Repository Search',
                query: 'language:TypeScript stars:>100 topics:web-framework',
                type: 'repositories',
                created_at: '2023-11-01',
                updated_at: '2023-11-01'
              }
            })
          });
        } else {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                saved_searches: []
              }
            })
          });
        }
      });

      await page.goto('/search');
      
      // Perform a search with filters
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'language:TypeScript stars:>100 topics:web-framework');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      // Save the search
      await page.click('button:has-text("Save Search")');
      await page.fill('input[placeholder="Enter search name"]', 'My Repository Search');
      await page.click('button:has-text("Save")');

      await expect(page.locator('text=Search saved successfully')).toBeVisible();
    });
  });

  test.describe('Search Result Sorting Options', () => {
    test('should support multiple sorting options', async ({ page }) => {
      // Mock search with sorting
      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const sort = url.searchParams.get('sort');
        
        let repositories = [
          {
            id: '1',
            name: 'repo-a',
            full_name: 'org/repo-a',
            description: 'Repository A',
            language: 'JavaScript',
            stargazers_count: 100,
            updated_at: '2023-11-01'
          },
          {
            id: '2',
            name: 'repo-b',
            full_name: 'org/repo-b',
            description: 'Repository B',
            language: 'TypeScript',
            stargazers_count: 500,
            updated_at: '2023-10-15'
          }
        ];

        // Apply sorting
        if (sort === 'stars') {
          repositories.sort((a, b) => b.stargazers_count - a.stargazers_count);
        } else if (sort === 'updated') {
          repositories.sort((a, b) => new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime());
        }

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              results: {
                repositories: repositories,
                pull_requests: [],
                users: []
              },
              total_count: repositories.length
            }
          })
        });
      });

      await page.goto('/search');
      
      // Perform search and test sorting
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'test');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      // Sort by stars
      await page.selectOption('select[name="sort"]', 'stars');
      
      await waitForLoadingToComplete(page);

      // Verify sorting (repo-b should come first due to higher stars)
      const firstResult = page.locator('.search-result').first();
      await expect(firstResult.locator('text=repo-b')).toBeVisible();
    });

    test('should support ascending and descending order', async ({ page }) => {
      await page.goto('/search');
      
      // Test order toggle
      await page.click('button:has-text("Sort")');
      await page.click('button:has-text("Ascending")');
      
      await expect(page.locator('button:has-text("Descending")')).toBeVisible();
    });
  });

  test.describe('Search Filter Persistence', () => {
    test('should persist filters in URL parameters', async ({ page }) => {
      await page.goto('/search');
      
      // Apply filters and check URL
      await page.selectOption('select[name="type"]', 'repositories');
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'language:JavaScript stars:>100');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      await expect(page).toHaveURL(/.*type=repositories.*/);
      await expect(page).toHaveURL(/.*q=language%3AJavaScript.*/);
    });

    test('should restore filters from bookmarked URLs', async ({ page }) => {
      // Navigate to URL with pre-set filters
      await page.goto('/search?type=repositories&q=language%3AJavaScript&sort=stars');
      
      const searchInput = page.locator('input[placeholder*="Search repositories, pull requests, users, and commits"]');
      await expect(searchInput).toHaveValue('language:JavaScript');
      await expect(page.locator('select[name="type"]')).toHaveValue('repositories');
      await expect(page.locator('select[name="sort"]')).toHaveValue('stars');
    });
  });

  test.describe('Filter Performance and UX', () => {
    test('should provide responsive filter application', async ({ page }) => {
      // Mock search with performance metrics
      await page.route('**/api/v1/search**', async route => {
        // Simulate slight delay
        await new Promise(resolve => setTimeout(resolve, 100));
        
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              results: {
                repositories: [],
                pull_requests: [],
                users: []
              },
              total_count: 0,
              query_time: 0.045
            }
          })
        });
      });

      await page.goto('/search');
      
      // Apply filters and verify quick response
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'test');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      // Should complete quickly
      await waitForLoadingToComplete(page);
      await expect(page.locator('text=Search completed')).toBeVisible({ timeout: 2000 });
    });

    test('should show loading states during filter application', async ({ page }) => {
      // Mock delayed search response
      await page.route('**/api/v1/search**', async route => {
        await new Promise(resolve => setTimeout(resolve, 1000));
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              results: {
                repositories: [],
                pull_requests: [],
                users: []
              },
              total_count: 0
            }
          })
        });
      });

      await page.goto('/search');
      
      // Start search and verify loading state
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'slow search');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      // Should show loading indicator
      await expect(page.locator('text=Searching...', { timeout: 500 })).toBeVisible();
    });

    test('should provide clear filter feedback', async ({ page }) => {
      await page.goto('/search');
      
      // Apply multiple filters
      await page.fill('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'language:JavaScript stars:>100');
      await page.press('input[placeholder*="Search repositories, pull requests, users, and commits"]', 'Enter');
      
      await waitForLoadingToComplete(page);

      // Verify filter feedback
      const activeFilters = page.locator('.filter-tag, .active-filter');
      await expect(activeFilters.first()).toBeVisible();
    });
  });
});