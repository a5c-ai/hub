import { test, expect } from '@playwright/test';
import { loginUser, testUser } from './helpers/test-utils';

/**
 * E2E tests for Pull Request Files & Diff workflow
 * 
 * Tests cover:
 * - View changed files in pull request
 * - Display file diffs with syntax highlighting
 * - Navigate between changed files
 * - Add inline comments on code lines
 * - Resolve and unresolve conversations
 */

test.describe('Pull Request Files and Diff View', () => {
  test.beforeEach(async ({ page }) => {
    // Mock pull request files
    await page.route('**/api/repositories/*/pulls/*/files', async route => {
      await route.fulfill({
        json: [
          {
            filename: 'src/auth/jwt.js',
            status: 'modified',
            additions: 45,
            deletions: 12,
            changes: 57,
            blob_url: '/repositories/testuser/testrepo/blob/feature/auth/src/auth/jwt.js',
            patch: `@@ -1,10 +1,15 @@
 const jwt = require('jsonwebtoken');
+const bcrypt = require('bcrypt');
 
 class JWTService {
   constructor(secret) {
     this.secret = secret;
+    this.saltRounds = 10;
   }
 
-  sign(payload) {
+  async sign(payload) {
+    // Hash sensitive data before signing
+    if (payload.password) {
+      payload.password = await bcrypt.hash(payload.password, this.saltRounds);
+    }
     return jwt.sign(payload, this.secret);
   }
 
@@ -15,6 +20,12 @@
   verify(token) {
     return jwt.verify(token, this.secret);
   }
+
+  async validatePassword(plainPassword, hashedPassword) {
+    return await bcrypt.compare(plainPassword, hashedPassword);
+  }
+
+  generateRefreshToken() {
+    return jwt.sign({ type: 'refresh' }, this.secret, { expiresIn: '7d' });
+  }
 }
 
 module.exports = JWTService;`
          },
          {
            filename: 'src/auth/middleware.js',
            status: 'added',
            additions: 23,
            deletions: 0,
            changes: 23,
            blob_url: '/repositories/testuser/testrepo/blob/feature/auth/src/auth/middleware.js',
            patch: `@@ -0,0 +1,23 @@
+const JWTService = require('./jwt');
+
+class AuthMiddleware {
+  constructor(jwtService) {
+    this.jwtService = jwtService;
+  }
+
+  authenticate(req, res, next) {
+    const token = req.headers.authorization?.split(' ')[1];
+    
+    if (!token) {
+      return res.status(401).json({ error: 'No token provided' });
+    }
+
+    try {
+      const decoded = this.jwtService.verify(token);
+      req.user = decoded;
+      next();
+    } catch (error) {
+      return res.status(401).json({ error: 'Invalid token' });
+    }
+  }
+}
+
+module.exports = AuthMiddleware;`
          },
          {
            filename: 'tests/auth.test.js',
            status: 'added',
            additions: 67,
            deletions: 0,
            changes: 67,
            blob_url: '/repositories/testuser/testrepo/blob/feature/auth/tests/auth.test.js',
            patch: `@@ -0,0 +1,67 @@
+const JWTService = require('../src/auth/jwt');
+const AuthMiddleware = require('../src/auth/middleware');
+
+describe('JWT Service', () => {
+  let jwtService;
+
+  beforeEach(() => {
+    jwtService = new JWTService('test-secret');
+  });
+
+  test('should sign and verify tokens', async () => {
+    const payload = { userId: 1, username: 'testuser' };
+    const token = await jwtService.sign(payload);
+    const decoded = jwtService.verify(token);
+    
+    expect(decoded.userId).toBe(1);
+    expect(decoded.username).toBe('testuser');
+  });
+
+  test('should hash passwords before signing', async () => {
+    const payload = { userId: 1, password: 'plaintext' };
+    const token = await jwtService.sign(payload);
+    const decoded = jwtService.verify(token);
+    
+    expect(decoded.password).not.toBe('plaintext');
+    expect(decoded.password.startsWith('$2b$')).toBe(true);
+  });
+
+  test('should validate passwords correctly', async () => {
+    const plainPassword = 'testpassword123';
+    const hashedPassword = await bcrypt.hash(plainPassword, 10);
+    
+    const isValid = await jwtService.validatePassword(plainPassword, hashedPassword);
+    expect(isValid).toBe(true);
+    
+    const isInvalid = await jwtService.validatePassword('wrongpassword', hashedPassword);
+    expect(isInvalid).toBe(false);
+  });
+});
+
+describe('Auth Middleware', () => {
+  let authMiddleware;
+  let jwtService;
+  let req, res, next;
+
+  beforeEach(() => {
+    jwtService = new JWTService('test-secret');
+    authMiddleware = new AuthMiddleware(jwtService);
+    
+    req = {
+      headers: {}
+    };
+    res = {
+      status: jest.fn().mockReturnThis(),
+      json: jest.fn()
+    };
+    next = jest.fn();
+  });
+
+  test('should authenticate valid tokens', async () => {
+    const token = await jwtService.sign({ userId: 1 });
+    req.headers.authorization = \`Bearer \${token}\`;
+    
+    authMiddleware.authenticate(req, res, next);
+    
+    expect(req.user.userId).toBe(1);
+    expect(next).toHaveBeenCalled();
+  });
+});`
          },
          {
            filename: 'package.json',
            status: 'modified',
            additions: 2,
            deletions: 0,
            changes: 2,
            blob_url: '/repositories/testuser/testrepo/blob/feature/auth/package.json',
            patch: `@@ -10,6 +10,8 @@
   "dependencies": {
     "express": "^4.18.0",
     "jsonwebtoken": "^9.0.0",
+    "bcrypt": "^5.1.0",
+    "jest": "^29.0.0",
     "lodash": "^4.17.21"
   },
   "devDependencies": {`
          }
        ]
      });
    });

    // Mock pull request details
    await page.route('**/api/repositories/*/pulls/1', async route => {
      await route.fulfill({
        json: {
          id: 1,
          issue: {
            number: 1,
            title: 'Add JWT authentication system',
            state: 'open',
            user: { username: 'developer' }
          },
          head_ref: 'feature/auth',
          base_ref: 'main',
          merged: false,
          mergeable: true,
          draft: false,
          additions: 137,
          deletions: 12,
          changed_files: 4
        }
      });
    });

    // Mock review comments
    await page.route('**/api/repositories/*/pulls/*/comments', async route => {
      if (route.request().method() === 'GET') {
        await route.fulfill({
          json: [
            {
              id: 1,
              path: 'src/auth/jwt.js',
              line: 12,
              body: 'Consider adding input validation here',
              user: { 
                username: 'reviewer1', 
                avatar_url: 'https://example.com/reviewer1.jpg' 
              },
              created_at: '2024-01-15T10:30:00Z',
              updated_at: '2024-01-15T10:30:00Z',
              in_reply_to_id: null
            },
            {
              id: 2,
              path: 'src/auth/jwt.js',
              line: 12,
              body: 'Good point! I\'ll add validation in the next commit.',
              user: { 
                username: 'developer', 
                avatar_url: 'https://example.com/developer.jpg' 
              },
              created_at: '2024-01-15T11:00:00Z',
              updated_at: '2024-01-15T11:00:00Z',
              in_reply_to_id: 1
            },
            {
              id: 3,
              path: 'tests/auth.test.js',
              line: 25,
              body: 'This test looks comprehensive, great coverage!',
              user: { 
                username: 'reviewer2', 
                avatar_url: 'https://example.com/reviewer2.jpg' 
              },
              created_at: '2024-01-15T12:00:00Z',
              updated_at: '2024-01-15T12:00:00Z',
              in_reply_to_id: null
            }
          ]
        });
      } else if (route.request().method() === 'POST') {
        const body = await route.request().postDataJSON();
        await route.fulfill({
          json: {
            id: 4,
            path: body.path,
            line: body.line,
            body: body.body,
            user: { 
              username: 'testuser', 
              avatar_url: 'https://example.com/testuser.jpg' 
            },
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
            in_reply_to_id: body.in_reply_to_id || null
          }
        });
      }
    });

    await loginUser(page);
  });

  test('displays changed files list', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Check files tab is active
    await expect(page.locator('[data-testid="files-tab"]')).toHaveClass(/active/);
    
    // Check file list is displayed
    await expect(page.locator('[data-testid="changed-files-list"]')).toBeVisible();
    await expect(page.locator('[data-testid="file-item"]')).toHaveCount(4);
    
    // Check file statistics summary
    await expect(page.locator('[data-testid="files-summary"]')).toContainText('4 changed files');
    await expect(page.locator('[data-testid="additions-summary"]')).toContainText('137 additions');
    await expect(page.locator('[data-testid="deletions-summary"]')).toContainText('12 deletions');
    
    // Check individual file items
    const firstFile = page.locator('[data-testid="file-item"]').first();
    await expect(firstFile.locator('[data-testid="file-name"]')).toContainText('src/auth/jwt.js');
    await expect(firstFile.locator('[data-testid="file-status"]')).toContainText('modified');
    await expect(firstFile.locator('[data-testid="file-additions"]')).toContainText('+45');
    await expect(firstFile.locator('[data-testid="file-deletions"]')).toContainText('-12');
    
    // Check added file
    const addedFile = page.locator('[data-testid="file-item"]').nth(1);
    await expect(addedFile.locator('[data-testid="file-name"]')).toContainText('src/auth/middleware.js');
    await expect(addedFile.locator('[data-testid="file-status"]')).toContainText('added');
    await expect(addedFile.locator('[data-testid="file-additions"]')).toContainText('+23');
  });

  test('displays file diff with syntax highlighting', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Check diff viewer is displayed
    await expect(page.locator('[data-testid="diff-viewer"]')).toBeVisible();
    
    // Check first file diff
    const firstFileDiff = page.locator('[data-testid="file-diff"]').first();
    await expect(firstFileDiff.locator('[data-testid="file-header"]')).toContainText('src/auth/jwt.js');
    
    // Check diff content
    await expect(firstFileDiff.locator('[data-testid="diff-content"]')).toBeVisible();
    
    // Check added lines (green background)
    const addedLines = firstFileDiff.locator('[data-testid="diff-line-added"]');
    await expect(addedLines).toHaveCount(9); // Should have 9+ added lines
    
    // Check deleted lines (red background)
    const deletedLines = firstFileDiff.locator('[data-testid="diff-line-deleted"]');
    await expect(deletedLines).toHaveCount(1); // Should have 1 deleted line
    
    // Check context lines (no background)
    const contextLines = firstFileDiff.locator('[data-testid="diff-line-context"]');
    const contextCount = await contextLines.count();
    expect(contextCount).toBeGreaterThan(5);
    
    // Check line numbers are displayed
    const lineNumbers = firstFileDiff.locator('[data-testid="line-number"]');
    const lineNumberCount = await lineNumbers.count();
    expect(lineNumberCount).toBeGreaterThan(10);
    
    // Check syntax highlighting is applied (look for highlighted keywords)
    const keywords = firstFileDiff.locator('.hljs-keyword');
    const keywordCount = await keywords.count();
    expect(keywordCount).toBeGreaterThan(3);
  });

  test('navigates between changed files', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Check file navigation
    await expect(page.locator('[data-testid="file-nav"]')).toBeVisible();
    await expect(page.locator('[data-testid="file-nav-item"]')).toHaveCount(4);
    
    // Check first file is active by default
    const firstNavItem = page.locator('[data-testid="file-nav-item"]').first();
    await expect(firstNavItem).toHaveClass(/active/);
    await expect(firstNavItem).toContainText('src/auth/jwt.js');
    
    // Click on second file
    const secondNavItem = page.locator('[data-testid="file-nav-item"]').nth(1);
    await secondNavItem.click();
    
    // Should show second file's diff
    await expect(secondNavItem).toHaveClass(/active/);
    await expect(page.locator('[data-testid="file-diff"]').first().locator('[data-testid="file-header"]'))
      .toContainText('src/auth/middleware.js');
    
    // Check file status indicator for added file
    await expect(page.locator('[data-testid="file-status-indicator"]')).toContainText('A');
    await expect(page.locator('[data-testid="file-status-indicator"]')).toHaveClass(/added/);
    
    // Use keyboard navigation
    await page.keyboard.press('ArrowDown');
    const thirdNavItem = page.locator('[data-testid="file-nav-item"]').nth(2);
    await expect(thirdNavItem).toHaveClass(/active/);
    
    await page.keyboard.press('ArrowUp');
    await expect(secondNavItem).toHaveClass(/active/);
  });

  test('adds inline comments on code lines', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Hover over a line to show comment button
    const codeLine = page.locator('[data-testid="diff-line"]').nth(5);
    await codeLine.hover();
    
    // Should show add comment button
    await expect(page.locator('[data-testid="add-comment-button"]')).toBeVisible();
    
    // Click add comment button
    await page.click('[data-testid="add-comment-button"]');
    
    // Should show comment form
    await expect(page.locator('[data-testid="inline-comment-form"]')).toBeVisible();
    await expect(page.locator('[data-testid="comment-textarea"]')).toBeVisible();
    await expect(page.locator('[data-testid="submit-comment-button"]')).toBeVisible();
    await expect(page.locator('[data-testid="cancel-comment-button"]')).toBeVisible();
    
    // Fill and submit comment
    await page.fill('[data-testid="comment-textarea"]', 'This logic could be simplified using destructuring.');
    await page.click('[data-testid="submit-comment-button"]');
    
    // Should show success message and add comment to the line
    await expect(page.locator('[data-testid="comment-success"]')).toContainText('Comment added successfully');
    await expect(page.locator('[data-testid="inline-comment"]')).toContainText('This logic could be simplified');
    
    // Should show comment count indicator
    await expect(page.locator('[data-testid="comment-count-indicator"]')).toContainText('1');
  });

  test('displays existing inline comments', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Should show existing comments inline
    await expect(page.locator('[data-testid="inline-comment"]')).toHaveCount(2); // 2 comment threads
    
    // Check first comment thread
    const firstComment = page.locator('[data-testid="comment-thread"]').first();
    await expect(firstComment.locator('[data-testid="comment-author"]')).toContainText('reviewer1');
    await expect(firstComment.locator('[data-testid="comment-body"]')).toContainText('Consider adding input validation here');
    await expect(firstComment.locator('[data-testid="comment-timestamp"]')).toBeVisible();
    
    // Check reply in the thread
    await expect(firstComment.locator('[data-testid="comment-reply"]')).toHaveCount(1);
    const reply = firstComment.locator('[data-testid="comment-reply"]').first();
    await expect(reply.locator('[data-testid="comment-author"]')).toContainText('developer');
    await expect(reply.locator('[data-testid="comment-body"]')).toContainText('Good point! I\'ll add validation');
    
    // Check second comment thread
    const secondComment = page.locator('[data-testid="comment-thread"]').nth(1);
    await expect(secondComment.locator('[data-testid="comment-author"]')).toContainText('reviewer2');
    await expect(secondComment.locator('[data-testid="comment-body"]')).toContainText('This test looks comprehensive');
  });

  test('replies to existing comments', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Find first comment thread and click reply
    const firstCommentThread = page.locator('[data-testid="comment-thread"]').first();
    await firstCommentThread.locator('[data-testid="reply-button"]').click();
    
    // Should show reply form
    await expect(page.locator('[data-testid="reply-form"]')).toBeVisible();
    await expect(page.locator('[data-testid="reply-textarea"]')).toBeVisible();
    
    // Fill and submit reply
    await page.fill('[data-testid="reply-textarea"]', 'Thanks for the suggestion! Validation added in commit abc123.');
    await page.click('[data-testid="submit-reply-button"]');
    
    // Should show success message and add reply to thread
    await expect(page.locator('[data-testid="reply-success"]')).toContainText('Reply added successfully');
    await expect(firstCommentThread.locator('[data-testid="comment-reply"]')).toHaveCount(2);
    await expect(firstCommentThread.locator('[data-testid="comment-reply"]').last())
      .toContainText('Thanks for the suggestion!');
  });

  test('resolves and unresolves comment conversations', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Find first comment thread
    const firstCommentThread = page.locator('[data-testid="comment-thread"]').first();
    
    // Should show resolve button
    await expect(firstCommentThread.locator('[data-testid="resolve-button"]')).toBeVisible();
    await expect(firstCommentThread.locator('[data-testid="resolve-button"]')).toContainText('Resolve conversation');
    
    // Click resolve
    await firstCommentThread.locator('[data-testid="resolve-button"]').click();
    
    // Should show confirmation dialog
    await expect(page.locator('[data-testid="resolve-confirmation"]')).toBeVisible();
    await page.click('[data-testid="confirm-resolve-button"]');
    
    // Comment thread should be marked as resolved
    await expect(firstCommentThread).toHaveClass(/resolved/);
    await expect(firstCommentThread.locator('[data-testid="resolved-badge"]')).toBeVisible();
    await expect(firstCommentThread.locator('[data-testid="resolved-badge"]')).toContainText('Resolved');
    
    // Should show unresolve button
    await expect(firstCommentThread.locator('[data-testid="unresolve-button"]')).toBeVisible();
    
    // Click unresolve
    await firstCommentThread.locator('[data-testid="unresolve-button"]').click();
    
    // Comment thread should be unresolved again
    await expect(firstCommentThread).not.toHaveClass(/resolved/);
    await expect(firstCommentThread.locator('[data-testid="resolved-badge"]')).not.toBeVisible();
    await expect(firstCommentThread.locator('[data-testid="resolve-button"]')).toBeVisible();
  });

  test('filters files by status', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Check filter options
    await expect(page.locator('[data-testid="file-filter"]')).toBeVisible();
    await expect(page.locator('[data-testid="filter-all"]')).toBeVisible();
    await expect(page.locator('[data-testid="filter-modified"]')).toBeVisible();
    await expect(page.locator('[data-testid="filter-added"]')).toBeVisible();
    await expect(page.locator('[data-testid="filter-deleted"]')).toBeVisible();
    
    // All files shown by default
    await expect(page.locator('[data-testid="file-item"]')).toHaveCount(4);
    
    // Filter by modified files
    await page.click('[data-testid="filter-modified"]');
    await expect(page.locator('[data-testid="file-item"]')).toHaveCount(2); // jwt.js and package.json
    
    // Filter by added files
    await page.click('[data-testid="filter-added"]');
    await expect(page.locator('[data-testid="file-item"]')).toHaveCount(2); // middleware.js and test.js
    
    // Clear filter
    await page.click('[data-testid="filter-all"]');
    await expect(page.locator('[data-testid="file-item"]')).toHaveCount(4);
  });

  test('shows file preview for images and binary files', async ({ page }) => {
    // Mock binary file
    await page.route('**/api/repositories/*/pulls/*/files', async route => {
      await route.fulfill({
        json: [
          {
            filename: 'assets/logo.png',
            status: 'added',
            additions: 0,
            deletions: 0,
            changes: 0,
            blob_url: '/repositories/testuser/testrepo/blob/feature/auth/assets/logo.png',
            patch: null, // Binary files don't have patches
            binary: true
          },
          {
            filename: 'docs/screenshot.jpg',
            status: 'modified',
            additions: 0,
            deletions: 0,
            changes: 0,
            blob_url: '/repositories/testuser/testrepo/blob/feature/auth/docs/screenshot.jpg',
            patch: null,
            binary: true
          }
        ]
      });
    });

    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Should show binary file indicators
    const binaryFile = page.locator('[data-testid="file-item"]').first();
    await expect(binaryFile.locator('[data-testid="binary-file-indicator"]')).toBeVisible();
    await expect(binaryFile.locator('[data-testid="binary-file-indicator"]')).toContainText('Binary file');
    
    // Should show image preview for image files
    await expect(page.locator('[data-testid="image-preview"]')).toBeVisible();
    await expect(page.locator('[data-testid="image-preview"] img')).toHaveAttribute('src', /logo\.png/);
    
    // Should show "No preview available" for other binary files
    const otherBinaryFile = page.locator('[data-testid="file-item"]').nth(1);
    await otherBinaryFile.click();
    await expect(page.locator('[data-testid="no-preview-message"]')).toContainText('No preview available for binary files');
  });

  test('handles large files gracefully', async ({ page }) => {
    // Mock large file
    await page.route('**/api/repositories/*/pulls/*/files', async route => {
      await route.fulfill({
        json: [
          {
            filename: 'data/large-dataset.json',
            status: 'modified',
            additions: 5000,
            deletions: 200,
            changes: 5200,
            blob_url: '/repositories/testuser/testrepo/blob/feature/auth/data/large-dataset.json',
            patch: null, // Large files don't show patches
            too_large: true
          }
        ]
      });
    });

    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Should show large file warning
    await expect(page.locator('[data-testid="large-file-warning"]')).toBeVisible();
    await expect(page.locator('[data-testid="large-file-warning"]')).toContainText('File too large to display');
    
    // Should show view file button
    await expect(page.locator('[data-testid="view-file-button"]')).toBeVisible();
    await expect(page.locator('[data-testid="view-file-button"]')).toContainText('View file');
    
    // Should show download file button
    await expect(page.locator('[data-testid="download-file-button"]')).toBeVisible();
  });
});

test.describe('Mobile File Diff Experience', () => {
  test.use({ viewport: { width: 375, height: 667 } }); // iPhone SE size

  test.beforeEach(async ({ page }) => {
    // Mock pull request files
    await page.route('**/api/repositories/*/pulls/*/files', async route => {
      await route.fulfill({
        json: [
          {
            filename: 'src/components/Button.js',
            status: 'modified',
            additions: 15,
            deletions: 3,
            changes: 18,
            patch: `@@ -1,8 +1,12 @@
 import React from 'react';
+import PropTypes from 'prop-types';
 
-const Button = ({ children, onClick }) => {
+const Button = ({ children, onClick, variant = 'primary', disabled = false }) => {
+  const baseClasses = 'px-4 py-2 rounded font-medium';
+  const variantClasses = variant === 'primary' ? 'bg-blue-500 text-white' : 'bg-gray-200 text-gray-800';
+  
   return (
-    <button onClick={onClick}>
+    <button 
+      onClick={onClick} 
+      disabled={disabled}
+      className={\`\${baseClasses} \${variantClasses}\`}
+    >
       {children}
     </button>
   );`
          }
        ]
      });
    });

    await loginUser(page);
  });

  test('mobile file diff navigation', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Check mobile-optimized layout
    await expect(page.locator('[data-testid="mobile-diff-container"]')).toBeVisible();
    
    // Check file navigation is collapsible on mobile
    await expect(page.locator('[data-testid="mobile-file-nav-toggle"]')).toBeVisible();
    await page.click('[data-testid="mobile-file-nav-toggle"]');
    await expect(page.locator('[data-testid="file-nav"]')).toBeVisible();
    
    // Check diff content is scrollable horizontally
    const diffContent = page.locator('[data-testid="diff-content"]');
    await expect(diffContent).toHaveCSS('overflow-x', 'auto');
    
    // Check line numbers are sticky on mobile
    await expect(page.locator('[data-testid="line-numbers"]')).toHaveCSS('position', 'sticky');
  });

  test('mobile comment interaction', async ({ page }) => {
    await page.goto('/repositories/testuser/testrepo/pull/1/files');
    
    // Tap on a code line to show comment options
    const codeLine = page.locator('[data-testid="diff-line"]').first();
    await codeLine.tap();
    
    // Should show mobile-optimized comment interface
    await expect(page.locator('[data-testid="mobile-comment-modal"]')).toBeVisible();
    
    // Should have large touch targets
    const commentButton = page.locator('[data-testid="add-comment-button"]');
    const buttonBox = await commentButton.boundingBox();
    expect(buttonBox?.height).toBeGreaterThan(44); // iOS recommended minimum touch target
    
    // Should show mobile-friendly textarea
    await commentButton.tap();
    await expect(page.locator('[data-testid="mobile-comment-textarea"]')).toBeVisible();
    await expect(page.locator('[data-testid="mobile-comment-textarea"]')).toHaveAttribute('rows', '6'); // Larger on mobile
  });
});