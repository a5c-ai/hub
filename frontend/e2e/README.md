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
- **`repository-settings.spec.ts`** - Repository settings and administration tests
- **`repository-security.spec.ts`** - Repository security, secrets, and access control tests
- **`repository-integrations.spec.ts`** - Repository webhooks, deploy keys, and integrations tests
- **`issues.spec.ts`** - Issue management and workflow tests
- **`actions-workflows.spec.ts`** - GitHub Actions workflow management tests
- **`actions-logs.spec.ts`** - Workflow run details and log viewing tests
- **`actions-runners.spec.ts`** - Self-hosted runner management tests
- **`actions-integration.spec.ts`** - Actions integration features (status checks, deployments)
- **`actions-performance.spec.ts`** - Performance and stress testing for Actions features
- **`activity-feed.spec.ts`** - Activity feed and timeline tests (24 tests)
- **`notifications.spec.ts`** - Notification center and management tests (40 tests)
- **`notification-preferences.spec.ts`** - Notification settings and preferences tests (34 tests)
- **`pull-requests.spec.ts`** - Pull request management and creation tests
- **`pull-request-review.spec.ts`** - Pull request review and collaboration workflow tests
- **`pull-request-files.spec.ts`** - Pull request file diff and code review tests
- **`user-profile.spec.ts`** - User profile management and public profile tests
- **`user-security.spec.ts`** - Security settings, password, SSH keys, and 2FA tests
- **`user-preferences.spec.ts`** - Notifications, appearance, accessibility, and privacy tests
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

#### GitHub Actions Tests
- **Workflow Management**: View workflows, run status, filtering, real-time updates
- **Workflow Run Details**: Job steps, status indicators, run metadata, timing information
- **Log Viewing**: Real-time log streaming, syntax highlighting, search, artifact downloads
- **Runner Management**: Self-hosted runners, health monitoring, runner groups, permissions
- **Integration Features**: Status checks on PRs, deployment workflows, matrix builds, conditional execution
- **Performance Testing**: Large log handling, concurrent runs, memory management, API throttling
#### Activity Feed Tests (24 tests)
- Display global activity feed with different event types (push, PR, issue, fork, star, follow)
- Filter activity by type (all, own, following)
- Activity timeline and event display
- Personal activity timeline and history
- Following users and organizations activity
- Activity feed pagination and infinite scroll
- Detailed push events with commit information
- Repository creation activities
- Follow activities
- Empty state handling
- Error handling and recovery
- Mobile responsiveness

#### Notifications Center Tests (40 tests)
- Notification inbox with unread indicators
- Notification categorization (mentions, issues, PRs, security)
- Mark notifications as read/unread (individual and bulk)
- Delete notifications
- Mark all notifications as read
- Bulk notification actions
- Issue and pull request notifications
- Comment and mention notifications
- Security and vulnerability alerts
- Repository invitation notifications
- Notification filtering and search
- Empty states for different filters
- Real-time notification updates via WebSocket
- Notification badge updates
- Mobile-friendly notification management
- Responsive notification layout
- Error handling and offline scenarios

#### Notification Preferences Tests (34 tests)
- Email notification settings management
- Notification delivery timing configuration
- Thread subscription management
- Browser notification permissions
- Desktop notification settings
- Automatic notification cleanup
- Manual notification cleanup
- Repository watching and notification settings
- Organization notification preferences
- Import/export notification settings
- Mobile-friendly settings interface
- Mobile swipe gestures for navigation
- API error handling
- Notification preference validation
- Accessibility features
- Performance optimization
#### Repository Settings & Administration Tests
- General repository settings (name, description, visibility)
- Default branch configuration
- Repository archiving and deletion
- Branch protection rules and policies
- Access control and collaborator management
- Actions settings and workflow permissions
- Security settings and vulnerability alerts
- Transfer and danger zone operations
- Mobile repository management
- Error handling and validation

#### Repository Security & Secrets Management Tests
- Repository secrets creation and management
- Environment-specific secrets configuration
- Secret validation and security checks
- Deploy keys management (read-only and read-write)
- SSH key validation and security standards
- Access control and permission management
- Security advisory and vulnerability management
- Error handling and network failure scenarios
- Mobile security management interface

#### Repository Webhooks & Integrations Tests
- Webhook creation, configuration, and management
- Event selection and payload customization
- Webhook testing (ping functionality)
- Content type configuration (JSON/form)
- Secret configuration for webhook security
- Webhook status management (enable/disable)
- Integration with repository settings
- Error handling and webhook validation
- Mobile webhook management interface

#### User Settings & Profile Tests
- Profile information editing and saving
- Avatar management and file uploads
- Profile visibility and public profile display
- Account settings and preferences
- Theme switching and appearance customization

#### Security Tests
- Password change functionality with validation
- SSH key management (add, remove, validate)
- Two-factor authentication setup and management
- Security audit logs and session management
- Account security best practices

#### User Preferences Tests
- Email notification preferences and frequency
- Web notification enablement and permissions
- Quiet hours and notification filtering
- Accessibility features (high contrast, large fonts, reduced motion)
- Language, timezone, and localization settings
- Privacy controls and data visibility settings
- Data export and portability features

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

# Run all GitHub Actions tests
npx playwright test actions-*.spec.ts

# Run specific Actions test suites
npx playwright test actions-workflows.spec.ts
npx playwright test actions-logs.spec.ts
npx playwright test actions-runners.spec.ts
npx playwright test actions-integration.spec.ts
npx playwright test actions-performance.spec.ts

# Run tests matching a pattern
npx playwright test --grep "login"
npx playwright test --grep "workflow"
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

#### Actions-Specific Utilities
- **`navigateToActions()`** - Navigate to repository Actions page
- **`navigateToWorkflowRun()`** - Navigate to specific workflow run
- **`navigateToRunners()`** - Navigate to runners management page
- **`waitForWorkflowCompletion()`** - Wait for workflow run to complete
- **`expectWorkflowStatus()`** - Verify workflow run status
- **`mockWorkflowData()`** - Mock API responses for testing
- **`simulateLogUpdates()`** - Simulate real-time log streaming
- **`checkMobileActions()`** - Verify mobile responsiveness
- **`checkActionsPerformance()`** - Test page load performance
- **`checkActionsAccessibility()`** - Verify accessibility standards

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

### Activity Feed Elements
- `filter-all` - All activity filter button
- `filter-own` - Your activity filter button
- `filter-following` - Following activity filter button
- `activity-item` - Individual activity items

### Notification Elements
- `filter-unread` - Unread notifications filter
- `filter-all` - All notifications filter
- `filter-participating` - Participating notifications filter
- `notification-item` - Individual notification items
- `mark-as-read-{id}` - Mark specific notification as read
- `delete-notification-{id}` - Delete specific notification
- `mark-all-as-read` - Mark all notifications as read
- `mark-selected-as-read` - Mark selected notifications as read
- `select-all-notifications` - Select all notifications checkbox
- `notification-icon-{type}` - Notification type icons

### Settings Elements
- `settings-tab-{id}` - Settings tab navigation buttons
- `settings-container` - Main settings container
- `email-issues-prs` - Email notifications for issues/PRs
- `email-repository-updates` - Email notifications for repository updates
- `email-security-alerts` - Email notifications for security alerts
- `enable-browser-notifications` - Enable browser notifications button

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