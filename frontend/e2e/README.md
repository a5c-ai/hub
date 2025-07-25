# End-to-End Testing with Playwright

This directory contains end-to-end (e2e) tests for the Hub frontend application using [Playwright](https://playwright.dev/).

## Overview

The e2e tests are designed to test the application from a user's perspective, verifying that the complete user workflows function correctly across different browsers.

## Test Structure

### Test Files

- **`auth.spec.ts`** - Authentication flow tests (login, register, logout)
- **`dashboard.spec.ts`** - Dashboard functionality tests
- **`navigation.spec.ts`** - Navigation and layout tests
- **`example.spec.ts`** - Basic application loading tests
- **`registration.spec.ts`** - User registration and validation tests
- **`repository.spec.ts`** - Repository management tests
- **`issues.spec.ts`** - Issue management and workflow tests
- **`pull-requests.spec.ts`** - Pull request management and creation tests
- **`pull-request-review.spec.ts`** - Pull request review and collaboration workflow tests
- **`pull-request-files.spec.ts`** - Pull request file diff and code review tests
- **`helpers/test-utils.ts`** - Shared utilities and helper functions

### Test Coverage

#### Authentication Tests
- Redirect unauthenticated users to login
- Successful login flow with redirect to dashboard
- Error handling for invalid credentials
- Registration flow
- Logout functionality
- Navigation between login and register pages

#### Dashboard Tests
- Display user welcome message
- Show repository statistics
- List recent repositories with proper metadata
- Display recent activity feed
- Navigation to other sections
- Responsive design on mobile
- Empty state handling

#### Pull Request Tests
- List pull requests with different states (open/closed/merged)
- Filter pull requests by state, author, and labels
- Create new pull request with form validation
- Navigate to pull request details
- Display pull request metadata and statistics
- Review workflow (approve, request changes, comment)
- Merge pull request with different merge strategies
- Close and reopen pull requests
- Convert draft to ready for review
- View and navigate changed files with diff display
- Add inline comments on code lines
- Resolve and unresolve comment conversations
- Mobile responsiveness for all PR workflows

#### Navigation Tests
- Header and sidebar navigation elements
- Mobile menu functionality
- Navigation between main sections
- User menu interactions
- Keyboard navigation
- Loading states

## Configuration

The tests are configured in `playwright.config.ts` with the following settings:

- **Browsers**: Chromium, Firefox, WebKit
- **Mobile Testing**: Pixel 5, iPhone 12
- **Base URL**: `http://localhost:3000` (configurable via `BASE_URL` env var)
- **Retries**: 2 retries on CI, 0 locally
- **Parallel Execution**: Enabled for faster test runs
- **Reporting**: HTML, JSON, and JUnit reports

## Running Tests

### Prerequisites

1. Install dependencies:
   ```bash
   npm install
   ```

2. Install Playwright browsers:
   ```bash
   npx playwright install
   ```

### Local Development

```bash
# Run all e2e tests
npm run test:e2e

# Run tests with UI mode (interactive)
npm run test:e2e:ui

# Run tests in headed mode (visible browser)
npm run test:e2e:headed

# Debug tests step by step
npm run test:e2e:debug

# View test reports
npm run test:e2e:report
```

### Specific Test Files

```bash
# Run only authentication tests
npx playwright test auth.spec.ts

# Run only dashboard tests
npx playwright test dashboard.spec.ts

# Run tests matching a pattern
npx playwright test --grep "login"
```

### Browser-Specific Testing

```bash
# Run tests only in Chromium
npx playwright test --project=chromium

# Run tests only on mobile
npx playwright test --project="Mobile Chrome"
```

## Test Data and Mocking

The tests use mocked API responses to ensure consistent and reliable test execution:

- **Authentication**: Mock login/register endpoints
- **User Data**: Mock user profile information
- **Repositories**: Mock repository lists and details
- **Activity**: Mock activity feeds

### Mock User Data

```typescript
const testUser = {
  username: 'testuser',
  email: 'test@example.com',
  password: 'TestPassword123!',
  name: 'Test User'
};
```

## Test Utilities

The `helpers/test-utils.ts` file provides common utilities:

- **`loginUser()`** - Helper to log in a user
- **`registerUser()`** - Helper to register a new user
- **`expectLoginPage()`** - Verify user is on login page
- **`expectDashboardPage()`** - Verify user is on dashboard
- **`waitForLoadingToComplete()`** - Wait for loading states
- **`takeScreenshot()`** - Capture screenshots for debugging

## Data Test IDs

Components include `data-testid` attributes for reliable element selection:

### Authentication Forms
- `email-input` - Email input field
- `password-input` - Password input field
- `login-button` - Login submit button
- `register-button` - Register submit button
- `error-message` - Error message display

### Navigation Elements
- `main-header` - Main header component
- `user-menu` - User dropdown menu
- `mobile-menu-button` - Mobile hamburger menu
- `user-avatar` - User avatar image
- `logout-button` - Logout button

## CI/CD Integration

The tests are integrated with GitHub Actions for continuous integration:

### Workflow Features
- Runs on pull requests and main branch pushes
- Tests across multiple browsers
- Stores test artifacts (screenshots, videos, reports)
- Fail-fast on test failures
- Parallel execution for performance

### Environment Variables
- `BASE_URL` - Application base URL (default: http://localhost:3000)
- `CI` - Enables CI-specific settings (retries, workers)

## Debugging Failed Tests

### Screenshots and Videos
Failed tests automatically capture:
- Screenshots on failure
- Videos of the entire test run
- Browser traces for detailed debugging

### Debug Mode
Use debug mode to step through tests:
```bash
npm run test:e2e:debug
```

### Verbose Output
Enable verbose logging:
```bash
npx playwright test --reporter=line
```

## Best Practices

### Writing Tests
1. **Use Page Object Model** - Encapsulate page interactions
2. **Reliable Selectors** - Prefer `data-testid` over CSS classes
3. **Wait for Elements** - Use Playwright's auto-waiting features
4. **Mock External APIs** - Ensure test isolation and reliability
5. **Test Real User Flows** - Focus on complete user journeys

### Maintenance
1. **Keep Tests Independent** - Each test should be self-contained
2. **Clean Test Data** - Reset state between tests
3. **Update Selectors** - Maintain test-ids when refactoring components
4. **Regular Test Reviews** - Remove obsolete tests, add coverage for new features

## Troubleshooting

### Common Issues

#### Browser Installation
```bash
# If browsers are missing
npx playwright install

# For system dependencies on Linux
npx playwright install-deps
```

#### Port Conflicts
```bash
# Use different port if 3000 is occupied
BASE_URL=http://localhost:3001 npm run test:e2e
```

#### Flaky Tests
- Add explicit waits for dynamic content
- Use `page.waitForSelector()` for elements that load asynchronously
- Increase timeout for slow operations

### Getting Help

- [Playwright Documentation](https://playwright.dev/docs)
- [Playwright Discord](https://discord.gg/playwright-807756831384403968)
- [Best Practices Guide](https://playwright.dev/docs/best-practices)

## Contributing

When adding new features:

1. **Add Test Coverage** - Include e2e tests for new user-facing features
2. **Update Test IDs** - Add `data-testid` attributes to new interactive elements
3. **Mock APIs** - Add appropriate API mocks for new endpoints
4. **Update Documentation** - Keep this README updated with new test patterns

## Future Enhancements

Planned improvements:
- [ ] Visual regression testing
- [ ] Accessibility testing integration
- [ ] Performance testing
- [ ] Cross-browser compatibility matrix
- [ ] Database seeding for more realistic test data