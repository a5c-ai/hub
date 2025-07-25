import { Page, expect } from '@playwright/test';

/**
 * Test utilities for common operations
 */

/**
 * Mock user credentials for testing
 */
export const testUser = {
  username: 'testuser',
  email: 'test@example.com',
  password: 'TestPassword123!',
  name: 'Test User'
};

/**
 * Login helper - logs in a user with credentials
 * @param page - Playwright page object
 * @param email - User email
 * @param password - User password
 */
export async function loginUser(page: Page, email: string = testUser.email, password: string = testUser.password) {
  await page.goto('/login');
  await page.fill('[data-testid="email-input"]', email);
  await page.fill('[data-testid="password-input"]', password);
  await page.click('[data-testid="login-button"]');
  
  // Wait for navigation to dashboard
  await page.waitForURL('/dashboard');
}

/**
 * Register helper - registers a new user
 * @param page - Playwright page object  
 * @param userData - User registration data
 */
export async function registerUser(page: Page, userData = testUser) {
  await page.goto('/register');
  await page.fill('[data-testid="name-input"]', userData.name);
  await page.fill('[data-testid="username-input"]', userData.username);
  await page.fill('[data-testid="email-input"]', userData.email);
  await page.fill('[data-testid="password-input"]', userData.password);
  await page.fill('[data-testid="confirm-password-input"]', userData.password);
  await page.click('[data-testid="register-button"]');
}

/**
 * Wait for loading to complete
 * @param page - Playwright page object
 */
export async function waitForLoadingToComplete(page: Page) {
  // Wait for any loading spinners to disappear
  await page.waitForFunction(() => {
    const spinners = document.querySelectorAll('.animate-spin');
    return spinners.length === 0;
  }, { timeout: 10000 });
}

/**
 * Check if user is on the login page
 * @param page - Playwright page object
 */
export async function expectLoginPage(page: Page) {
  await expect(page).toHaveURL('/login');
  await expect(page.locator('h2')).toContainText('Sign in to Hub');
}

/**
 * Check if user is on the dashboard page
 * @param page - Playwright page object
 */
export async function expectDashboardPage(page: Page) {
  await expect(page).toHaveURL('/dashboard');
  await expect(page.locator('h1')).toContainText('Welcome back');
}

/**
 * Take a screenshot with a descriptive name
 * @param page - Playwright page object
 * @param name - Screenshot name
 */
export async function takeScreenshot(page: Page, name: string) {
  await page.screenshot({ 
    path: `e2e/screenshots/${name}.png`, 
    fullPage: true 
  });
}

/**
 * Actions-specific test utilities
 */

/**
 * Navigate to Actions page for a repository
 * @param page - Playwright page object
 * @param owner - Repository owner
 * @param repo - Repository name
 */
export async function navigateToActions(page: Page, owner: string = 'admin', repo: string = 'sample-project') {
  await page.goto(`/repositories/${owner}/${repo}/actions`);
  await waitForLoadingToComplete(page);
}

/**
 * Navigate to specific workflow run
 * @param page - Playwright page object
 * @param owner - Repository owner
 * @param repo - Repository name
 * @param runId - Workflow run ID
 */
export async function navigateToWorkflowRun(page: Page, runId: string, owner: string = 'admin', repo: string = 'sample-project') {
  await page.goto(`/repositories/${owner}/${repo}/actions/runs/${runId}`);
  await waitForLoadingToComplete(page);
}

/**
 * Navigate to runners management page
 * @param page - Playwright page object
 * @param owner - Repository owner
 * @param repo - Repository name
 */
export async function navigateToRunners(page: Page, owner: string = 'admin', repo: string = 'sample-project') {
  await page.goto(`/repositories/${owner}/${repo}/settings/runners`);
  await waitForLoadingToComplete(page);
}

/**
 * Wait for workflow run to complete
 * @param page - Playwright page object
 * @param timeout - Timeout in milliseconds
 */
export async function waitForWorkflowCompletion(page: Page, timeout: number = 300000) {
  await page.waitForFunction(() => {
    const statusElements = document.querySelectorAll('text=/[🔄⏳]/');
    return statusElements.length === 0;
  }, { timeout });
}

/**
 * Check if workflow run has specific status
 * @param page - Playwright page object
 * @param status - Expected status (success, failure, cancelled, etc.)
 */
export async function expectWorkflowStatus(page: Page, status: 'success' | 'failure' | 'cancelled' | 'pending' | 'in_progress') {
  const statusIcons = {
    success: '✅',
    failure: '❌', 
    cancelled: '⭕',
    pending: '⏳',
    in_progress: '🔄'
  };

  await expect(page.locator(`text=${statusIcons[status]}`)).toBeVisible();
}

/**
 * Mock API responses for workflow data
 * @param page - Playwright page object
 * @param mockData - Mock workflow data
 */
export async function mockWorkflowData(page: Page, mockData: any) {
  await page.route('**/api/v1/repos/**/actions/**', async route => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(mockData)
    });
  });
}

/**
 * Simulate real-time log updates
 * @param page - Playwright page object
 * @param logLines - Array of log lines to append
 */
export async function simulateLogUpdates(page: Page, logLines: string[]) {
  for (const line of logLines) {
    await page.evaluate((logLine) => {
      const logContainer = document.querySelector('[data-testid="log-container"]') || 
                          document.querySelector('pre') || 
                          document.querySelector('.log-output');
      if (logContainer) {
        logContainer.textContent += '\n' + logLine;
        logContainer.scrollTop = logContainer.scrollHeight;
      }
    }, line);
    
    await page.waitForTimeout(100); // Simulate real-time delay
  }
}

/**
 * Check mobile responsiveness for Actions pages
 * @param page - Playwright page object
 */
export async function checkMobileActions(page: Page) {
  await page.setViewportSize({ width: 375, height: 667 }); // iPhone SE
  
  // Check that essential elements are visible and usable
  await expect(page.locator('h1')).toBeVisible();
  
  // Check that buttons are touchable (minimum 44px)
  const buttons = page.locator('button, a[role="button"]');
  const buttonCount = await buttons.count();
  
  for (let i = 0; i < Math.min(buttonCount, 5); i++) {
    const button = buttons.nth(i);
    if (await button.isVisible()) {
      const boundingBox = await button.boundingBox();
      if (boundingBox) {
        expect(boundingBox.height).toBeGreaterThanOrEqual(44);
      }
    }
  }
}

/**
 * Test Actions page performance
 * @param page - Playwright page object
 * @param url - URL to test
 */
export async function checkActionsPerformance(page: Page, url: string) {
  const startTime = Date.now();
  
  await page.goto(url);
  await waitForLoadingToComplete(page);
  
  const loadTime = Date.now() - startTime;
  
  // Actions pages should load within 5 seconds
  expect(loadTime).toBeLessThan(5000);
  
  // Check for performance issues
  const performanceMetrics = await page.evaluate(() => {
    const navigation = performance.getEntriesByType('navigation')[0] as PerformanceNavigationTiming;
    return {
      domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
      loadComplete: navigation.loadEventEnd - navigation.loadEventStart,
      firstPaint: performance.getEntriesByType('paint').find(entry => entry.name === 'first-paint')?.startTime || 0
    };
  });
  
  // DOM should be interactive quickly
  expect(performanceMetrics.domContentLoaded).toBeLessThan(2000);
}

/**
 * Verify Actions accessibility
 * @param page - Playwright page object
 */
export async function checkActionsAccessibility(page: Page) {
  // Check for proper heading hierarchy
  const headings = page.locator('h1, h2, h3, h4, h5, h6');
  const headingCount = await headings.count();
  
  if (headingCount > 0) {
    // Should have at least one h1
    const h1Count = await page.locator('h1').count();
    expect(h1Count).toBeGreaterThanOrEqual(1);
  }
  
  // Check for alt text on images
  const images = page.locator('img');
  const imageCount = await images.count();
  
  for (let i = 0; i < imageCount; i++) {
    const img = images.nth(i);
    await expect(img).toHaveAttribute('alt');
  }
  
  // Check for proper button labels
  const buttons = page.locator('button:not([aria-label])');
  const buttonCount = await buttons.count();
  
  for (let i = 0; i < buttonCount; i++) {
    const button = buttons.nth(i);
    const text = await button.textContent();
    expect(text?.trim()).toBeTruthy();
  }
  
  // Check for focus management
  const focusableElements = page.locator('a, button, input, textarea, select, [tabindex]:not([tabindex="-1"])');
  const focusableCount = await focusableElements.count();
  
  if (focusableCount > 0) {
    // First focusable element should be reachable
    await focusableElements.first().focus();
    await expect(focusableElements.first()).toBeFocused();
  }
}