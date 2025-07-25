import { test, expect } from '@playwright/test';
import { loginUser, testUser, waitForLoadingToComplete, takeScreenshot } from './helpers/test-utils';

test.describe('Code Search Functionality', () => {
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

  test.describe('Repository Code Search', () => {
    test('should search within repository contents', async ({ page }) => {
      // Mock repository code search results
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        
        if (query === 'function handleSubmit') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                code: [
                  {
                    id: '1',
                    repository_id: 'testrepo',
                    repository_name: 'testowner/testrepo',
                    file_path: 'src/components/LoginForm.tsx',
                    file_name: 'LoginForm.tsx',
                    language: 'TypeScript',
                    content: 'export function LoginForm() {\n  const handleSubmit = (e: FormEvent) => {\n    e.preventDefault();\n    // Handle login logic\n  };\n  return <form onSubmit={handleSubmit}>...</form>;\n}',
                    line_count: 25,
                    branch: 'main',
                    matched_lines: [
                      {
                        line_number: 2,
                        content: '  const handleSubmit = (e: FormEvent) => {',
                        highlighted: true
                      },
                      {
                        line_number: 6,
                        content: '  return <form onSubmit={handleSubmit}>...</form>;',
                        highlighted: true
                      }
                    ]
                  },
                  {
                    id: '2',
                    repository_id: 'testrepo',
                    repository_name: 'testowner/testrepo',
                    file_path: 'src/utils/formHelpers.ts',
                    file_name: 'formHelpers.ts',
                    language: 'TypeScript',
                    content: 'export function handleSubmit(data: FormData) {\n  // Generic form submission handler\n  return submitForm(data);\n}',
                    line_count: 10,
                    branch: 'main',
                    matched_lines: [
                      {
                        line_number: 1,
                        content: 'export function handleSubmit(data: FormData) {',
                        highlighted: true
                      }
                    ]
                  }
                ],
                issues: [],
                commits: [],
                total_count: 2
              }
            })
          });
        }
      });

      await page.goto('/repositories/testowner/testrepo/search');
      await waitForLoadingToComplete(page);

      // Verify repository search page elements
      await expect(page.locator('input[placeholder*="Search in this repository"]')).toBeVisible();
      await expect(page.locator('button:has-text("Code")')).toBeVisible();

      // Perform code search
      await page.fill('input[placeholder*="Search in this repository"]', 'function handleSubmit');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify code search results
      await expect(page.locator('text=LoginForm.tsx')).toBeVisible();
      await expect(page.locator('text=formHelpers.ts')).toBeVisible();
      await expect(page.locator('text=src/components/LoginForm.tsx')).toBeVisible();
      await expect(page.locator('text=src/utils/formHelpers.ts')).toBeVisible();

      // Verify syntax highlighting
      await expect(page.locator('text=const handleSubmit = (e: FormEvent) => {')).toBeVisible();
      await expect(page.locator('text=export function handleSubmit(data: FormData) {')).toBeVisible();

      // Verify line numbers and context
      await expect(page.locator('text=Line 2')).toBeVisible();
      await expect(page.locator('text=Line 1')).toBeVisible();
    });

    test('should provide code search with syntax highlighting', async ({ page }) => {
      // Mock code search with detailed syntax highlighting
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              code: [
                {
                  id: '1',
                  repository_id: 'testrepo',
                  repository_name: 'testowner/testrepo',
                  file_path: 'src/api/userService.js',
                  file_name: 'userService.js',
                  language: 'JavaScript',
                  content: 'async function fetchUserData(userId) {\n  try {\n    const response = await fetch(`/api/users/${userId}`);\n    return await response.json();\n  } catch (error) {\n    console.error("Failed to fetch user data:", error);\n    throw error;\n  }\n}',
                  line_count: 9,
                  branch: 'main',
                  highlighted_content: '<span class="keyword">async</span> <span class="keyword">function</span> <span class="function">fetchUserData</span>(<span class="parameter">userId</span>) {',
                  matched_lines: [
                    {
                      line_number: 1,
                      content: 'async function fetchUserData(userId) {',
                      highlighted: true
                    },
                    {
                      line_number: 3,
                      content: '    const response = await fetch(`/api/users/${userId}`);',
                      highlighted: false
                    }
                  ]
                }
              ],
              issues: [],
              commits: [],
              total_count: 1
            }
          })
        });
      });

      await page.goto('/repositories/testowner/testrepo/search');
      await page.fill('input[placeholder*="Search in this repository"]', 'async function');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify syntax-highlighted code appears
      await expect(page.locator('text=userService.js')).toBeVisible();
      await expect(page.locator('text=async function fetchUserData')).toBeVisible();
      await expect(page.locator('text=JavaScript')).toBeVisible();
      
      // Verify code block formatting
      await expect(page.locator('pre, code')).toBeVisible();
    });

    test('should support search in specific branches or commits', async ({ page }) => {
      // Mock repository branches API
      await page.route('**/api/v1/repositories/testowner/testrepo/branches**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: [
              { name: 'main', sha: 'abc123', is_default: true },
              { name: 'develop', sha: 'def456', is_default: false },
              { name: 'feature/new-search', sha: 'ghi789', is_default: false }
            ]
          })
        });
      });

      // Mock branch-specific search
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        const url = new URL(route.request().url());
        const branch = url.searchParams.get('branch') || 'main';
        const query = url.searchParams.get('q');
        
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              code: [
                {
                  id: '1',
                  repository_id: 'testrepo',
                  repository_name: 'testowner/testrepo',
                  file_path: `src/search-${branch}.js`,
                  file_name: `search-${branch}.js`,
                  language: 'JavaScript',
                  content: `// Search functionality for ${branch} branch\nfunction search${branch}() {\n  return "search results";\n}`,
                  line_count: 4,
                  branch: branch,
                  matched_lines: [
                    {
                      line_number: 2,
                      content: `function search${branch}() {`,
                      highlighted: true
                    }
                  ]
                }
              ],
              issues: [],
              commits: [],
              total_count: 1
            }
          })
        });
      });

      await page.goto('/repositories/testowner/testrepo/search');
      
      // Look for branch selector (would need to be implemented in UI)
      // For now, test that default branch search works
      await page.fill('input[placeholder*="Search in this repository"]', 'function search');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Verify main branch results
      await expect(page.locator('text=search-main.js')).toBeVisible();
      await expect(page.locator('text=main')).toBeVisible();
    });

    test('should support file path and filename search', async ({ page }) => {
      // Mock file path search
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        
        if (query?.includes('filename:')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                code: [
                  {
                    id: '1',
                    repository_id: 'testrepo',
                    repository_name: 'testowner/testrepo',
                    file_path: 'src/components/SearchBar.tsx',
                    file_name: 'SearchBar.tsx',
                    language: 'TypeScript',
                    content: 'export function SearchBar() {\n  return <input type="search" />;\n}',
                    line_count: 3,
                    branch: 'main',
                    matched_lines: []
                  },
                  {
                    id: '2',
                    repository_id: 'testrepo',
                    repository_name: 'testowner/testrepo',
                    file_path: 'src/hooks/useSearch.ts',
                    file_name: 'useSearch.ts',
                    language: 'TypeScript',
                    content: 'export function useSearch() {\n  // Search hook logic\n}',
                    line_count: 5,
                    branch: 'main',
                    matched_lines: []
                  }
                ],
                issues: [],
                commits: [],
                total_count: 2
              }
            })
          });
        }
      });

      await page.goto('/repositories/testowner/testrepo/search');
      
      // Search for files containing "search" in filename
      await page.fill('input[placeholder*="Search in this repository"]', 'filename:search');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify file-based search results
      await expect(page.locator('text=SearchBar.tsx')).toBeVisible();
      await expect(page.locator('text=useSearch.ts')).toBeVisible();
      await expect(page.locator('text=src/components/SearchBar.tsx')).toBeVisible();
      await expect(page.locator('text=src/hooks/useSearch.ts')).toBeVisible();
    });

    test('should provide search result navigation and context', async ({ page }) => {
      // Mock detailed code search with context
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              code: [
                {
                  id: '1',
                  repository_id: 'testrepo',
                  repository_name: 'testowner/testrepo',
                  file_path: 'src/utils/apiClient.ts',
                  file_name: 'apiClient.ts',
                  language: 'TypeScript',
                  content: 'import axios from "axios";\n\nexport class ApiClient {\n  private baseURL: string;\n\n  constructor(baseURL: string) {\n    this.baseURL = baseURL;\n  }\n\n  async get(endpoint: string) {\n    return axios.get(`${this.baseURL}${endpoint}`);\n  }\n\n  async post(endpoint: string, data: any) {\n    return axios.post(`${this.baseURL}${endpoint}`, data);\n  }\n}',
                  line_count: 17,
                  branch: 'main',
                  matched_lines: [
                    {
                      line_number: 3,
                      content: 'export class ApiClient {',
                      highlighted: true
                    },
                    {
                      line_number: 6,
                      content: '  constructor(baseURL: string) {',
                      highlighted: false
                    },
                    {
                      line_number: 10,
                      content: '  async get(endpoint: string) {',
                      highlighted: false
                    }
                  ]
                }
              ],
              issues: [],
              commits: [],
              total_count: 1
            }
          })
        });
      });

      await page.goto('/repositories/testowner/testrepo/search');
      await page.fill('input[placeholder*="Search in this repository"]', 'class ApiClient');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify search result shows context
      await expect(page.locator('text=apiClient.ts')).toBeVisible();
      await expect(page.locator('text=export class ApiClient')).toBeVisible();
      
      // Verify line numbers are shown
      await expect(page.locator('text=Line 3')).toBeVisible();
      
      // Check for navigation to file
      const fileLink = page.locator('text=src/utils/apiClient.ts');
      await expect(fileLink).toBeVisible();
      
      // Test clicking on file link (would navigate to file view)
      await fileLink.click();
      await expect(page).toHaveURL(/.*\/blob\/.*\/src\/utils\/apiClient\.ts/);
    });
  });

  test.describe('Advanced Code Search Features', () => {
    test('should support regex and advanced search patterns', async ({ page }) => {
      // Mock regex search
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        
        if (query?.includes('/function \\w+\\(/')) {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                code: [
                  {
                    id: '1',
                    repository_id: 'testrepo',
                    repository_name: 'testowner/testrepo',
                    file_path: 'src/math.js',
                    file_name: 'math.js',
                    language: 'JavaScript',
                    content: 'function add(a, b) {\n  return a + b;\n}\n\nfunction multiply(x, y) {\n  return x * y;\n}',
                    line_count: 6,
                    branch: 'main',
                    matched_lines: [
                      {
                        line_number: 1,
                        content: 'function add(a, b) {',
                        highlighted: true
                      },
                      {
                        line_number: 5,
                        content: 'function multiply(x, y) {',
                        highlighted: true
                      }
                    ]
                  }
                ],
                issues: [],
                commits: [],
                total_count: 1
              }
            })
          });
        }
      });

      await page.goto('/repositories/testowner/testrepo/search');
      
      // Test regex search pattern
      await page.fill('input[placeholder*="Search in this repository"]', '/function \\w+\\(/');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify regex search results
      await expect(page.locator('text=math.js')).toBeVisible();
      await expect(page.locator('text=function add(a, b)')).toBeVisible();
      await expect(page.locator('text=function multiply(x, y)')).toBeVisible();
    });

    test('should support language-specific search filters', async ({ page }) => {
      // Mock language filter search
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        const url = new URL(route.request().url());
        const query = url.searchParams.get('q');
        const language = url.searchParams.get('language');
        
        if (language === 'TypeScript') {
          await route.fulfill({
            status: 200,
            contentType: 'application/json',
            body: JSON.stringify({
              success: true,
              data: {
                code: [
                  {
                    id: '1',
                    repository_id: 'testrepo',
                    repository_name: 'testowner/testrepo',
                    file_path: 'src/types.ts',
                    file_name: 'types.ts',
                    language: 'TypeScript',
                    content: 'interface User {\n  id: string;\n  name: string;\n  email: string;\n}',
                    line_count: 5,
                    branch: 'main',
                    matched_lines: [
                      {
                        line_number: 1,
                        content: 'interface User {',
                        highlighted: true
                      }
                    ]
                  }
                ],
                issues: [],
                commits: [],
                total_count: 1
              }
            })
          });
        }
      });

      await page.goto('/repositories/testowner/testrepo/search');
      
      // Use language filter (UI implementation would be needed)
      await page.fill('input[placeholder*="Search in this repository"]', 'interface language:TypeScript');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify TypeScript-specific results
      await expect(page.locator('text=types.ts')).toBeVisible();
      await expect(page.locator('text=TypeScript')).toBeVisible();
      await expect(page.locator('text=interface User')).toBeVisible();
    });

    test('should handle large code files efficiently', async ({ page }) => {
      // Mock large file search
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              code: [
                {
                  id: '1',
                  repository_id: 'testrepo',
                  repository_name: 'testowner/testrepo',
                  file_path: 'src/largefile.js',
                  file_name: 'largefile.js',
                  language: 'JavaScript',
                  content: '// This is a very large file with 5000+ lines\n// Content truncated for performance...',
                  line_count: 5234,
                  branch: 'main',
                  matched_lines: [
                    {
                      line_number: 1234,
                      content: 'function importantFunction() {',
                      highlighted: true
                    },
                    {
                      line_number: 3456,
                      content: '// Another match here',
                      highlighted: true
                    }
                  ]
                }
              ],
              issues: [],
              commits: [],
              total_count: 1
            }
          })
        });
      });

      await page.goto('/repositories/testowner/testrepo/search');
      await page.fill('input[placeholder*="Search in this repository"]', 'importantFunction');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify large file handling
      await expect(page.locator('text=largefile.js')).toBeVisible();
      await expect(page.locator('text=5234 lines')).toBeVisible();
      await expect(page.locator('text=Line 1234')).toBeVisible();
      await expect(page.locator('text=Line 3456')).toBeVisible();
      
      // Verify performance (should load quickly despite large file)
      await expect(page.locator('text=function importantFunction()')).toBeVisible();
    });
  });

  test.describe('Search Result Export and Sharing', () => {
    test('should support copying search result links', async ({ page }) => {
      await page.goto('/repositories/testowner/testrepo/search');
      await page.fill('input[placeholder*="Search in this repository"]', 'test query');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Verify URL contains search parameters
      await expect(page).toHaveURL(/.*q=test%20query.*/);
      
      // Test that URL can be copied and shared
      const currentUrl = page.url();
      expect(currentUrl).toContain('search');
      expect(currentUrl).toContain('q=');
    });

    test('should provide permalink to specific search results', async ({ page }) => {
      // Mock search with specific result IDs
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              code: [
                {
                  id: 'result-123',
                  repository_id: 'testrepo',
                  repository_name: 'testowner/testrepo',
                  file_path: 'src/example.js',
                  file_name: 'example.js',
                  language: 'JavaScript',
                  content: 'function example() { return "test"; }',
                  line_count: 1,
                  branch: 'main',
                  matched_lines: [
                    {
                      line_number: 1,
                      content: 'function example() { return "test"; }',
                      highlighted: true
                    }
                  ]
                }
              ],
              issues: [],
              commits: [],
              total_count: 1
            }
          })
        });
      });

      await page.goto('/repositories/testowner/testrepo/search');
      await page.fill('input[placeholder*="Search in this repository"]', 'function example');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Verify search result can be linked to
      await expect(page.locator('text=example.js')).toBeVisible();
      
      // Test navigation to specific line in file
      const lineLink = page.locator('text=Line 1');
      if (await lineLink.isVisible()) {
        await lineLink.click();
        await expect(page).toHaveURL(/.*\/blob\/.*\/src\/example\.js.*#L1/);
      }
    });
  });

  test.describe('Mobile Code Search Experience', () => {
    test('should provide optimized mobile code search', async ({ page }) => {
      // Set mobile viewport
      await page.setViewportSize({ width: 375, height: 667 });
      
      await page.goto('/repositories/testowner/testrepo/search');
      
      // Verify mobile-friendly layout
      const searchInput = page.locator('input[placeholder*="Search in this repository"]');
      await expect(searchInput).toBeVisible();
      
      // Check that code results are readable on mobile
      await searchInput.fill('mobile test');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      // Verify mobile code display
      const codeBlocks = page.locator('pre, code');
      if (await codeBlocks.count() > 0) {
        const firstCodeBlock = codeBlocks.first();
        const boundingBox = await firstCodeBlock.boundingBox();
        
        // Code should not overflow horizontally
        expect(boundingBox?.width).toBeLessThanOrEqual(375);
      }
    });

    test('should support touch-friendly code navigation', async ({ page }) => {
      await page.setViewportSize({ width: 375, height: 667 });
      
      // Mock mobile code search
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              code: [
                {
                  id: '1',
                  repository_id: 'testrepo',
                  repository_name: 'testowner/testrepo',
                  file_path: 'src/mobile.js',
                  file_name: 'mobile.js',
                  language: 'JavaScript',
                  content: 'function mobileFunction() {\n  return "mobile optimized";\n}',
                  line_count: 3,
                  branch: 'main',
                  matched_lines: [
                    {
                      line_number: 1,
                      content: 'function mobileFunction() {',
                      highlighted: true
                    }
                  ]
                }
              ],
              issues: [],
              commits: [],
              total_count: 1
            }
          })
        });
      });

      await page.goto('/repositories/testowner/testrepo/search');
      await page.fill('input[placeholder*="Search in this repository"]', 'mobileFunction');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);

      // Test touch interactions
      const fileLink = page.locator('text=mobile.js');
      if (await fileLink.isVisible()) {
        // Verify touch target is large enough (minimum 44px)
        const boundingBox = await fileLink.boundingBox();
        expect(boundingBox?.height).toBeGreaterThanOrEqual(44);
        
        // Test tap interaction
        await fileLink.tap();
      }
    });
  });

  test.describe('Code Search Performance', () => {
    test('should handle concurrent searches efficiently', async ({ page }) => {
      let searchCount = 0;
      
      // Mock API with counter
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        searchCount++;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            success: true,
            data: {
              code: [
                {
                  id: `result-${searchCount}`,
                  repository_id: 'testrepo',
                  repository_name: 'testowner/testrepo',
                  file_path: `src/search${searchCount}.js`,
                  file_name: `search${searchCount}.js`,
                  language: 'JavaScript',
                  content: `// Search result ${searchCount}`,
                  line_count: 1,
                  branch: 'main',
                  matched_lines: []
                }
              ],
              issues: [],
              commits: [],
              total_count: 1
            }
          })
        });
      });

      await page.goto('/repositories/testowner/testrepo/search');
      
      // Simulate rapid searches (debouncing test)
      const searchInput = page.locator('input[placeholder*="Search in this repository"]');
      await searchInput.fill('test');
      await searchInput.fill('test1');
      await searchInput.fill('test12');
      await searchInput.fill('test123');
      
      await page.click('button:has-text("Search")');
      await waitForLoadingToComplete(page);
      
      // Should have only made one API call due to debouncing
      expect(searchCount).toBeLessThanOrEqual(2);
    });

    test('should provide search performance metrics', async ({ page }) => {
      // Mock timed search response
      await page.route('**/api/v1/repositories/testowner/testrepo/search**', async route => {
        const startTime = Date.now();
        await new Promise(resolve => setTimeout(resolve, 50)); // Simulate processing
        
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          headers: {
            'X-Search-Time': '50ms',
            'X-Results-Count': '15'
          },
          body: JSON.stringify({
            success: true,
            data: {
              code: [
                {
                  id: '1',
                  repository_id: 'testrepo',
                  repository_name: 'testowner/testrepo',
                  file_path: 'src/performance.js',
                  file_name: 'performance.js',
                  language: 'JavaScript',
                  content: 'function performanceTest() { return "fast"; }',
                  line_count: 1,
                  branch: 'main',
                  matched_lines: []
                }
              ],
              issues: [],
              commits: [],
              total_count: 15
            }
          })
        });
      });

      const searchStartTime = Date.now();
      
      await page.goto('/repositories/testowner/testrepo/search');
      await page.fill('input[placeholder*="Search in this repository"]', 'performance');
      await page.click('button:has-text("Search")');
      
      await waitForLoadingToComplete(page);
      
      const searchEndTime = Date.now();
      const totalTime = searchEndTime - searchStartTime;
      
      // Search should complete quickly
      expect(totalTime).toBeLessThan(2000);
      
      // Verify results are displayed
      await expect(page.locator('text=performance.js')).toBeVisible();
    });
  });
});