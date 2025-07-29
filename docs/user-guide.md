# User Guide - Hub Git Hosting Service

Welcome to Hub, a powerful self-hosted git hosting service designed to provide enterprise-grade features with complete data sovereignty. This guide will help you get started with using Hub for your development projects.

## Table of Contents

- [Getting Started](#getting-started)
- [Repository Management](#repository-management)
- [Working with Git](#working-with-git)
- [Collaboration Features](#collaboration-features)

- [Pull Requests and Code Review](#pull-requests-and-code-review)
- [CI/CD and Automation](#cicd-and-automation)
- [Team and Organization Management](#team-and-organization-management)
- [Security and Access Control](#security-and-access-control)
- [Advanced Features](#advanced-features)

## Getting Started

### Accessing Hub

1. **Web Interface**: Navigate to your Hub instance URL (e.g., `https://hub.yourcompany.com`)
2. **Authentication**: Log in using your credentials or SSO provider
3. **Dashboard**: View your repositories, organizations, and recent activity

### Account Setup

#### Profile Configuration
1. Click your avatar in the top-right corner
2. Select "Profile Settings"
3. Update your profile information:
   - Display name and bio
   - Profile picture
   - Email preferences
   - Notification settings

#### SSH Key Management
1. Go to "Settings" → "SSH Keys"
2. Click "Add SSH Key"
3. Paste your public key and give it a descriptive name
4. Test your SSH connection:
   ```bash
   ssh -T git@hub.yourcompany.com
   ```

#### Personal Access Tokens
1. Navigate to "Settings" → "Access Tokens"
2. Click "Generate New Token"
3. Select appropriate scopes (read, write, admin)
4. Copy and securely store your token
5. Use for API access or Git operations over HTTPS

## Repository Management

### Creating Repositories

#### New Repository
1. Click the "+" button in the top navigation
2. Select "New Repository"
3. Fill in repository details:
   - **Name**: Repository name (required)
   - **Description**: Brief description of the project
   - **Visibility**: Public, Private, or Internal
   - **Initialize**: Add README, .gitignore, or license
   - **Template**: Use a repository template if available

#### From Template
1. Navigate to a template repository
2. Click "Use this template"
3. Configure your new repository settings
4. Template variables will be populated automatically

#### Import Existing Repository
1. Click "+" → "Import Repository"
2. Provide the source repository URL
3. Configure authentication if required
4. Set visibility and other options
5. Start the import process

### Repository Settings

#### General Settings
- **Repository name and description**
- **Visibility settings** (Public/Private/Internal)
- **Default branch** configuration
- **Repository features** (Issues, Wiki, Projects)
- **Danger zone** (Transfer ownership, Delete repository)

#### Branch Protection
1. Go to "Settings" → "Branches"
2. Add branch protection rules:
   - Require pull request reviews
   - Require status checks
   - Require branches to be up to date
   - Include administrators
   - Allow force pushes (not recommended)

#### Collaborators and Access
1. Navigate to "Settings" → "Collaborators"
2. Add individual users or teams
3. Assign permission levels:
   - **Read**: Clone and pull
   - **Write**: Push to non-protected branches
   - **Admin**: Full repository access

## Working with Git

### Cloning Repositories

#### HTTPS Clone
```bash
git clone https://hub.yourcompany.com/username/repository.git
cd repository
```

#### SSH Clone
```bash
git clone git@hub.yourcompany.com:username/repository.git
cd repository
```

### Basic Git Operations

#### Making Changes
```bash
# Make your changes
git add .
git commit -m "Descriptive commit message"
git push origin main
```

#### Working with Branches
```bash
# Create and switch to new branch
git checkout -b feature/new-feature

# Push new branch
git push -u origin feature/new-feature

# Switch between branches
git checkout main
git checkout feature/new-feature

# Delete local branch
git branch -d feature/new-feature

# Delete remote branch
git push origin --delete feature/new-feature
```

### Large File Support (Git LFS)
#### Server Configuration

Before using Git LFS, configure the storage backend in your `config.yaml`:
```yaml
lfs:
  backend: azure_blob
  azure:
    account_name: <your-account-name>
    account_key: <your-account-key>
    container_name: <your-container-name>
```

#### Setup
```bash
# Install Git LFS
git lfs install

# Track large files
git lfs track "*.psd"
git lfs track "*.zip"
git lfs track "assets/*"

# Add and commit
git add .gitattributes
git commit -m "Add Git LFS tracking"
```

#### Usage
```bash
# Add large files normally
git add large-file.zip
git commit -m "Add large asset file"
git push origin main
```

## Collaboration Features

### Organizations

#### Joining an Organization
1. Accept invitation via email or direct link
2. Set organization visibility in your profile
3. Access organization repositories and teams

#### Organization Dashboard
- View organization activity
- Browse repositories and teams
- Access organization settings (if you have permissions)

### Teams

#### Team Membership
- Teams provide grouped access to repositories
- Nested team structure with inheritance
- Team discussions and mentions

#### Team Permissions
- Organization-level permissions
- Repository-specific access
- Permission inheritance from parent teams

## Pull Requests and Code Review

### Creating Pull Requests

#### Basic Pull Request
1. Push your feature branch to the repository
2. Navigate to the repository on Hub
3. Click "New Pull Request"
4. Select base and compare branches
5. Fill in PR details:
   - **Title**: Clear description of changes
   - **Description**: Detailed explanation
   - **Reviewers**: Request code reviews
   - **Labels**: Categorize the PR


#### Draft Pull Requests
- Create work-in-progress PRs
- Share early feedback
- Convert to regular PR when ready

### Code Review Process

#### Reviewing Code
1. Navigate to the "Files changed" tab
2. Review line-by-line changes
3. Add comments and suggestions:
   - **Single comments**: Quick feedback
   - **Review comments**: Part of formal review
   - **Suggestions**: Propose specific code changes

#### Review States
- **Comment**: General feedback without approval
- **Approve**: Code looks good to merge
- **Request Changes**: Issues that must be addressed

#### Responding to Reviews
1. Address reviewer comments
2. Make requested changes
3. Push new commits or use suggested changes
4. Resolve conversations when addressed
5. Re-request review if needed

### Merging Pull Requests

#### Merge Options
- **Merge commit**: Preserve commit history
- **Squash and merge**: Combine commits into one
- **Rebase and merge**: Linear history without merge commit

#### Pre-merge Requirements
- Required status checks must pass
- Required reviews must be completed
- Branch must be up to date (if configured)
- No merge conflicts

## CI/CD and Automation

### GitHub-Compatible Actions

#### Workflow Files
Create `.github/workflows/ci.yml`:
```yaml
name: CI
on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - name: Setup Node.js
      uses: actions/setup-node@v3
      with:
        node-version: '18'
    - name: Install dependencies
      run: npm install
    - name: Run tests
      run: npm test
```

#### Status Checks
- Automatic status updates on commits
- Required checks before merge
- Build logs and artifact storage

### Webhooks

#### Setting Up Webhooks
1. Go to repository "Settings" → "Webhooks"
2. Click "Add webhook"
3. Configure webhook:
   - **URL**: External service endpoint
   - **Content Type**: JSON or form data
   - **Events**: Select triggering events
   - **Secret**: Optional security token

#### Common Webhook Events
- Push events
- Pull request events
- Issue events
- Release events
- Repository events

## Team and Organization Management

### User Roles

#### Repository Roles
- **Read**: Clone, pull, and view repository
- **Triage**: Manage issues and pull requests
- **Write**: Push to repository and manage some settings
- **Maintain**: Repository maintenance without destructive actions
- **Admin**: Full access including repository deletion

#### Organization Roles
- **Owner**: Full organization access
- **Billing Manager**: Manage billing and payments
- **Member**: Basic organization membership
- **Moderator**: Manage interactions and content

### Permissions Management

#### Team Permissions
- Assign teams to repositories
- Set base permissions for team members
- Override permissions for specific users
- Manage team hierarchies and inheritance

#### Access Reviews
- Regular review of user access
- Remove inactive users
- Audit permission changes
- Compliance reporting

## Security and Access Control

### Two-Factor Authentication

#### Enabling 2FA
1. Go to "Settings" → "Security"
2. Click "Enable two-factor authentication"
3. Choose method:
   - **TOTP App**: Authenticator app (recommended)
   - **SMS**: Text message codes
   - **Hardware Token**: FIDO2/WebAuthn devices
4. Save recovery codes securely

#### Recovery Options
- Store recovery codes in a secure location
- Generate new recovery codes if needed
- Disable 2FA only with recovery codes

### Security Policies

#### Password Requirements
- Minimum length and complexity
- Regular password rotation
- No password reuse

#### Access Restrictions
- IP address allowlists
- Time-based access controls
- Device registration requirements

### Audit Logs

#### Viewing Audit Logs
1. Organization owners can access audit logs
2. Navigate to organization "Settings" → "Audit Log"
3. Filter by:
   - Date range
   - User actions
   - Event types
   - Repository activity

#### Common Audit Events
- Login and authentication events
- Repository access and changes
- Permission modifications
- Organization membership changes

## Advanced Features

### Repository Templates

#### Using Templates
1. Browse available templates
2. Click "Use this template"
3. Configure template variables
4. Create repository from template

#### Creating Templates
1. Mark repository as template in settings
2. Add template configuration files
3. Define template variables and prompts
4. Set up template-specific documentation

### Advanced Git Features

#### Git Hooks
- **Pre-receive hooks**: Executed before any refs are updated, used for validation and access control.
- **Post-receive hooks**: Executed after refs update, used for notifications, webhooks, and repository synchronization.
- **Custom hook scripts**: Register scripts via API or CLI; scripts are placed in `hooks/<type>.d/` and executed in order.

#### Repository Mirroring
- Mirror repositories from external sources
- Automatic synchronization
- Bidirectional mirroring options

### API Access

#### REST API
- Full-featured REST API
- Comprehensive endpoint coverage
- Rate limiting and authentication
- Interactive API documentation

#### GraphQL API
- Flexible query capabilities
- Real-time subscriptions
- Efficient data fetching
- Schema introspection

### Integrations

#### Third-Party Integrations
- Slack and Microsoft Teams notifications
- Jira and Azure DevOps integration
- IDE plugins and extensions
- Monitoring and analytics tools

#### Custom Integrations
- Webhook-based integrations
- API-driven custom applications
- Plugin development framework
- Community-contributed plugins

## Troubleshooting

### Common Issues

#### Authentication Problems
- **SSH Key Issues**: Verify key format and permissions
- **Token Expiry**: Generate new personal access tokens
- **2FA Problems**: Use recovery codes or contact admin

#### Git Operation Issues
- **Push Rejected**: Check branch protection rules
- **Large File Issues**: Ensure Git LFS is properly configured
- **Permission Denied**: Verify repository access permissions

#### Performance Issues
- **Slow Clones**: Check network connectivity and repository size
- **Timeout Errors**: Increase Git timeout settings
- **Large Repository**: Consider repository cleanup or LFS migration

### Getting Help

#### Documentation Resources
- Administrator guide for deployment and configuration
- Developer guide for API and integration development
- Community forums and discussions

#### Support Channels
- Built-in help and documentation
- Community support forums
- Professional support options (if available)
- Issue tracking for bug reports and feature requests

## Best Practices

### Repository Organization
- Use clear, descriptive repository names
- Maintain comprehensive README files
- Implement consistent branching strategies
- Tag releases with semantic versioning

### Collaboration Workflow
- Create descriptive commit messages
- Use pull requests for all changes
- Implement code review processes
- Maintain project documentation

### Security Practices
- Enable two-factor authentication
- Regularly rotate access tokens
- Review and audit permissions
- Use branch protection rules

### CI/CD Best Practices
- Implement comprehensive testing
- Use staging environments
- Automate deployment processes
- Monitor build and deployment metrics

---

This user guide provides a comprehensive overview of Hub's features and capabilities. For more detailed technical information, see the [Administrator Guide](admin-guide.md) and [Developer Guide](developer-guide.md).

For the latest updates and announcements, check the [project repository](https://github.com/a5c-ai/hub) and community resources.
