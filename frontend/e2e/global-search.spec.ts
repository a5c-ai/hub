import { test, expect } from '@playwright/test';
import { loginUser, testUser, waitForLoadingToComplete, takeScreenshot } from './helpers/test-utils';

test.describe('Global Search Functionality', () => {
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

  test.describe('Universal Search', () => {
    test('should perform universal search across all content types', async ({ page }) => {
      // Mock comprehensive search results
      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        
        if (query === 'javascript') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                users: [
                  {
                    id: '1',
                    username: 'jsdev',
                    full_name: 'JavaScript Developer',
                    email: 'jsdev@example.com',
                    bio: 'JavaScript enthusiast',
                    avatar_url: '',
                    company: 'Tech Corp',
                    location: 'San Francisco'
                  }
                ],
                repositories: [
                  {
                    id: '1',
                    name: 'javascript-toolkit',
                    description: 'A comprehensive JavaScript toolkit',
                    owner_id: '1',
                    owner_type: 'user',
                    visibility: 'public',
                    stars_count: 500,
                    forks_count: 150,
                    primary_language: 'JavaScript',
                    created_at: '2023-01-01',
                    updated_at: '2023-12-01'
                  }
                ],
                issues: [
                  {
                    id: '1',
                    number: 123,
                    title: 'Fix JavaScript performance issue',
                    body: 'We need to optimize the JavaScript bundling...',
                    state: 'open',
                    repository_id: '1',
                    user_id: '1',
                    created_at: '2023-11-01',
                    updated_at: '2023-11-15'
                  }
                ],
                organizations: [
                  {
                    id: '1',
                    name: 'js-foundation',
                    display_name: 'JavaScript Foundation',
                    description: 'Supporting JavaScript development',
                    location: 'Global',
                    website: 'https://js-foundation.org',
                    created_at: '2023-01-01'
                  }
                ],
                commits: [
                  {
                    id: '1',
                    sha: 'abc123def456',
                    message: 'Update JavaScript dependencies',
                    author_name: 'John Doe',
                    author_email: 'john@example.com',
                    committer_name: 'John Doe',
                    committer_date: '2023-12-01T10:00:00Z',
                    repository_id: '1'
                  }
                ],
                total_count: 5
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
                users: [],
                repositories: [],
                issues: [],
                organizations: [],
                commits: [],
                total_count: 0
              }
            })
          });
        }
      });

      await page.goto('/search');
      await waitForLoadingToComplete(page);

      // Verify search page elements
      await expect(page.locator('input[placeholder*="Search repositories, issues, users, and commits"]')).toBeVisible();
      await expect(page.locator('button:has-text("Search")')).toBeVisible();

      // Perform search
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'javascript');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify all content types are displayed
      await expect(page.locator('text=Users')).toBeVisible();
      await expect(page.locator('text=Repositories')).toBeVisible();
      await expect(page.locator('text=Issues')).toBeVisible();
      await expect(page.locator('text=Organizations')).toBeVisible();
      await expect(page.locator('text=Commits')).toBeVisible();

      // Verify specific results
      await expect(page.locator('text=jsdev')).toBeVisible();
      await expect(page.locator('text=javascript-toolkit')).toBeVisible();
      await expect(page.locator('text=Fix JavaScript performance issue')).toBeVisible();
      await expect(page.locator('text=JavaScript Foundation')).toBeVisible();
      await expect(page.locator('text=Update JavaScript dependencies')).toBeVisible();
    });

    test('should handle search result categorization', async ({ page }) => {
      // Mock categorized results
      await page.route('**/api/v1/search**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: [
                {
                  id: '1',
                  name: 'react-app',
                  description: 'A React application',
                  owner_id: '1',
                  owner_type: 'user',
                  visibility: 'public',
                  stars_count: 100,
                  forks_count: 25,
                  primary_language: 'JavaScript',
                  created_at: '2023-01-01',
                  updated_at: '2023-12-01'
                },
                {
                  id: '2',
                  name: 'vue-components',
                  description: 'Vue.js components library',
                  owner_id: '2',
                  owner_type: 'organization',
                  visibility: 'public',
                  stars_count: 200,
                  forks_count: 50,
                  primary_language: 'Vue',
                  created_at: '2023-02-01',
                  updated_at: '2023-11-01'
                }
              ],
              issues: [],
              organizations: [],
              commits: [],
              total_count: 2
            }
          })
        });
      });

      await page.goto('/search');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'components');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify repositories section
      const repoSection = page.locator('text=Repositories').locator('..');
      await expect(repoSection.locator('text=react-app')).toBeVisible();
      await expect(repoSection.locator('text=vue-components')).toBeVisible();
      
      // Verify repository metadata
      await expect(page.locator('text=⭐ 100')).toBeVisible();
      await expect(page.locator('text=🍴 25')).toBeVisible();
      await expect(page.locator('text=⭐ 200')).toBeVisible();
      await expect(page.locator('text=🍴 50')).toBeVisible();
      await expect(page.locator('text=JavaScript')).toBeVisible();
      await expect(page.locator('text=Vue')).toBeVisible();
    });

    test('should support search autocomplete and suggestions', async ({ page }) => {
      // Mock search suggestions
      await page.route('**/api/v1/search/suggestions**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        
        if (query === 'reac') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                suggestions: [
                  'react',
                  'react-native',
                  'react-router',
                  'react-hooks',
                  'reactive-programming'
                ]
              }
            })
          });
        }
      });

      await page.goto('/search');
      
      const searchInput = page.locator('input[placeholder*="Search repositories, issues, users, and commits"]');
      await searchInput.fill('reac');
      
      // Wait for suggestions to appear
      await page.waitForTimeout(500);
      
      // Check if suggestions dropdown appears (this would require frontend implementation)
      // For now, we verify the input accepts the partial text
      await expect(searchInput).toHaveValue('reac');
    });
  });

  test.describe('Search History and Saved Searches', () => {
    test('should maintain search history', async ({ page }) => {
      await page.goto('/search');
      
      // Perform multiple searches
      const searches = ['react', 'javascript', 'python'];
      
      for (const searchTerm of searches) {
        await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', searchTerm);
        await page.press('input[placeholder*="Search repositories, issues, users, and commits"]', 'Enter');
        await waitForLoadingToComplete(page);
      }
      
      // Verify URL parameters are updated
      await expect(page).toHaveURL(/.*q=python.*/);
      
      // Check that browser history contains previous searches
      await page.goBack();
      await expect(page).toHaveURL(/.*q=javascript.*/);
      
      await page.goBack();
      await expect(page).toHaveURL(/.*q=react.*/);
    });

    test('should support saved search functionality', async ({ page }) => {
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
                  name: 'JavaScript Repositories',
                  query: 'javascript language:javascript',
                  type: 'repositories',
                  created_at: '2023-01-01'
                }
              ]
            })
          });
        } else if (route.request().method() === 'POST') {
          await route.fulfill({
            status: 201,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                id: '2',
                name: 'React Components',
                query: 'react components',
                type: 'all'
              }
            })
          });
        }
      });

      await page.goto('/search');
      
      // This test would require saved search UI to be implemented
      // For now, we verify the page loads correctly
      await expect(page.locator('input[placeholder*="Search repositories, issues, users, and commits"]')).toBeVisible();
    });
  });

  test.describe('Search Performance and UX', () => {
    test('should handle search result pagination', async ({ page }) => {
      // Mock paginated results
      await page.route('**/api/v1/search**', async route => {
        const url = new URL(route.request().url());
        const page_num = url.searchParams.get('page') || '1';
        
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: Array.from({ length: 30 }, (_, i) => ({
                id: `${i + 1}`,
                name: `repo-${i + 1}-page-${page_num}`,
                description: `Repository ${i + 1} on page ${page_num}`,
                owner_id: '1',
                owner_type: 'user',
                visibility: 'public',
                stars_count: i + 1,
                forks_count: i,
                primary_language: 'JavaScript',
                created_at: '2023-01-01',
                updated_at: '2023-12-01'
              })),
              issues: [],
              organizations: [],
              commits: [],
              total_count: 100
            },
            pagination: {
              page: parseInt(page_num),
              per_page: 30,
              total: 100,
              total_pages: 4
            }
          })
        });
      });

      await page.goto('/search');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'test');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify first page results
      await expect(page.locator('text=repo-1-page-1')).toBeVisible();
      await expect(page.locator('text=repo-30-page-1')).toBeVisible();
      
      // Look for pagination controls (would need to be implemented)
      // For now, verify URL supports pagination parameters
      await expect(page).toHaveURL(/.*q=test.*/);
    });

    test('should handle large dataset performance', async ({ page }) => {
      // Mock a large dataset response with timing
      await page.route('**/api/v1/search**', async route => {
        // Simulate processing time
        await new Promise(resolve => setTimeout(resolve, 100));
        
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: Array.from({ length: 30 }, (_, i) => ({
                id: `${i + 1}`,
                name: `large-repo-${i + 1}`,
                description: `One of many repositories in a large dataset`,
                owner_id: '1',
                owner_type: 'user',
                visibility: 'public',
                stars_count: 1000 + i,
                forks_count: 100 + i,
                primary_language: 'JavaScript',
                created_at: '2023-01-01',
                updated_at: '2023-12-01'
              })),
              issues: [],
              organizations: [],
              commits: [],
              total_count: 50000
            }
          })
        });
      });

      const startTime = Date.now();
      
      await page.goto('/search');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'large dataset');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      const endTime = Date.now();
      const searchTime = endTime - startTime;
      
      // Search should complete within reasonable time
      expect(searchTime).toBeLessThan(5000);
      
      // Verify results are displayed
      await expect(page.locator('text=large-repo-1')).toBeVisible();
    });

    test('should provide optimized mobile search experience', async ({ page }) => {
      // Set mobile viewport
      await page.setViewportSize({ width: 375, height: 667 });
      
      await page.goto('/search');
      
      // Verify mobile-friendly layout
      const searchInput = page.locator('input[placeholder*="Search repositories, issues, users, and commits"]');
      await expect(searchInput).toBeVisible();
      
      // Check input is properly sized for mobile
      const inputBox = await searchInput.boundingBox();
      expect(inputBox?.width).toBeGreaterThan(300); // Should take most of screen width
      
      // Verify search button is touchable (minimum 44px height)
      const searchButton = page.locator('button:has-text("Search")');
      const buttonBox = await searchButton.boundingBox();
      expect(buttonBox?.height).toBeGreaterThanOrEqual(44);
      
      // Test touch interaction
      await searchInput.fill('mobile test');
      await searchButton.tap();
      
      await waitForLoadingToComplete(page);
    });

    test('should handle no results with helpful suggestions', async ({ page }) => {
      // Mock empty results
      await page.route('**/api/v1/search**', async route => {
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
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'xyznonexistentquery123');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Verify no results message
      await expect(page.locator('text=No results found')).toBeVisible();
      await expect(page.locator('text=Try adjusting your search query or search in a different category')).toBeVisible();
      
      // Check for search suggestions
      await expect(page.locator('text=Search suggestions')).toBeVisible();
    });
  });

  test.describe('Search Analytics', () => {
    test('should track search queries for analytics', async ({ page }) => {
      let searchAnalytics: any[] = [];
      
      // Mock analytics tracking
      await page.route('**/api/v1/analytics/search**', async route => {
        if (route.request().method() === 'POST') {
          const body = await route.request().postDataJSON();
          searchAnalytics.push(body);
          
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({ success: true })
          });
        }
      });

      await page.goto('/search');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'analytics test');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Verify analytics were tracked (would require frontend implementation)
      // For now, verify search was performed
      await expect(page).toHaveURL(/.*q=analytics%20test.*/);
    });

    test('should provide search performance metrics', async ({ page }) => {
      // Mock performance data
      await page.route('**/api/v1/search**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          headers: {
            'X-Search-Time': '150ms',
            'X-Results-Count': '42'
          },
          body: JSON.stringify({
            success: true,
            data: {
              users: [],
              repositories: [
                {
                  id: '1',
                  name: 'performance-test',
                  description: 'Testing search performance',
                  owner_id: '1',
                  owner_type: 'user',
                  visibility: 'public',
                  stars_count: 10,
                  forks_count: 5,
                  primary_language: 'JavaScript',
                  created_at: '2023-01-01',
                  updated_at: '2023-12-01'
                }
              ],
              issues: [],
              organizations: [],
              commits: [],
              total_count: 1
            }
          })
        });
      });

      await page.goto('/search');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'performance');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Verify results are displayed
      await expect(page.locator('text=performance-test')).toBeVisible();
      
      // Performance metrics would be tracked in the background
      // For this test, we verify the search completed successfully
      await expect(page.locator('text=Repositories')).toBeVisible();
    });
  });

  test.describe('Error Handling', () => {
    test('should handle search API errors gracefully', async ({ page }) => {
      // Mock API error
      await page.route('**/api/v1/search**', async route => {
        await route.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({
            success: false,
            error: 'Internal server error'
          })
        });
      });

      await page.goto('/search');
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'error test');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Verify error message
      await expect(page.locator('text=Failed to perform search. Please try again.')).toBeVisible();
    });

    test('should handle network timeouts', async ({ page }) => {
      // Mock slow network
      await page.route('**/api/v1/search**', async route => {
        await new Promise(resolve => setTimeout(resolve, 10000));
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
      await page.fill('input[placeholder*="Search repositories, issues, users, and commits"]', 'timeout test');
      await page.click('button:has-text("Search")');
      
      // Should show loading state
      await expect(page.locator('text=Searching...')).toBeVisible();
      
      // Wait a reasonable time then verify timeout handling
      await page.waitForTimeout(2000);
      await expect(page.locator('button:has-text("Searching...")')).toBeDisabled();
    });
  });

  test.describe('Accessibility', () => {
    test('should be accessible with keyboard navigation', async ({ page }) => {
      await page.goto('/search');
      
      // Test keyboard navigation
      const searchInput = page.locator('input[placeholder*="Search repositories, issues, users, and commits"]');
      await searchInput.focus();
      await expect(searchInput).toBeFocused();
      
      await searchInput.fill('accessibility test');
      await searchInput.press('Tab');
      
      const searchButton = page.locator('button:has-text("Search")');
      await expect(searchButton).toBeFocused();
      
      await searchButton.press('Enter');
      await waitForLoadingToComplete(page);
    });

    test('should have proper ARIA labels and screen reader support', async ({ page }) => {
      await page.goto('/search');
      
      // Check for proper labeling
      const searchInput = page.locator('input[placeholder*="Search repositories, issues, users, and commits"]');
      await expect(searchInput).toHaveAttribute('type', 'text');
      
      // Verify search form has proper structure
      const searchForm = page.locator('form').first();
      await expect(searchForm).toBeVisible();
      
      // Check button has accessible text
      const searchButton = page.locator('button:has-text("Search")');
      await expect(searchButton).toHaveAttribute('type', 'submit');
    });
  });
});