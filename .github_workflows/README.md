# CI/CD Pipeline Documentation

This directory contains GitHub Actions workflows for automated testing, building, and deployment of the Hub application.

## Workflows Overview

### 1. Continuous Integration (`ci.yml`)
**Triggers:** Push to main/develop, Pull Requests
- **Backend Tests**: Go unit tests with PostgreSQL service
- **Frontend Tests**: Jest tests with coverage reporting
- **Security Scan**: Trivy vulnerability scanning
- **Code Quality**: ESLint, TypeScript checks, Go linting

### 2. Build and Package (`build.yml`)
**Triggers:** Push to main/develop, Git tags
- **Multi-platform Docker builds** (amd64, arm64)
- **Container Registry**: GitHub Container Registry (ghcr.io)
- **Image Tagging**: Branch-based and semantic versioning
- **Build Caching**: GitHub Actions cache for faster builds

### 3. Infrastructure Deployment (`infrastructure.yml`)
**Triggers:** Push to terraform/ directory, Manual dispatch
- **Terraform Validation**: Plan and validation
- **Environment Support**: development, staging, production
- **Azure Integration**: AKS, networking, security resources
- **State Management**: Remote state with Azure backend

### 4. Application Deployment (`deploy.yml`)
**Triggers:** Successful build completion, Manual dispatch
- **Kubernetes Deployment**: AKS cluster deployment
- **Environment Isolation**: Separate namespaces per environment
- **Health Checks**: Rollout status verification
- **Blue-Green Support**: Zero-downtime deployments

### 5. Pull Request Automation (`pr-automation.yml`)
**Triggers:** PR events (open, sync, reopen)
- **Auto-assignment**: Reviewers based on file patterns
- **Size Labeling**: Automatic PR size classification
- **Conventional Commits**: Commit message validation
- **Dependency Review**: Security vulnerability scanning
- **Conflict Detection**: Merge conflict identification

### 6. Release Automation (`release.yml`)
**Triggers:** Git tags (v*)
- **Changelog Generation**: Conventional commit based
- **GitHub Releases**: Automated release creation
- **Production Deployment**: Automatic production rollout
- **Versioning**: Semantic versioning support

### 7. Notifications (`notifications.yml`)
**Triggers:** Deployment completion
- **Slack Integration**: Deployment status notifications
- **Teams Integration**: Microsoft Teams notifications
- **Performance Monitoring**: Lighthouse CI integration
- **Security Scanning**: OWASP ZAP security testing

## Environment Configuration

### Required Secrets

#### Repository Secrets
- `AZURE_CREDENTIALS`: Azure service principal credentials
- `AZURE_RESOURCE_GROUP`: Resource group name
- `AKS_CLUSTER_NAME`: AKS cluster name
- `DOMAIN_NAME`: Application domain name
- `SLACK_WEBHOOK`: Slack webhook URL
- `TEAMS_WEBHOOK`: Microsoft Teams webhook URL

#### Environment-specific Secrets
Each environment (development, staging, production) should have:
- `AZURE_CREDENTIALS_<ENV>`: Environment-specific Azure credentials
- `DATABASE_URL_<ENV>`: Environment-specific database connection
- `REDIS_URL_<ENV>`: Environment-specific Redis connection

### Environment Protection Rules

#### Development
- No protection rules
- Automatic deployment from develop branch

#### Staging
- Manual approval required
- Deployment branch: develop, main

#### Production
- Required reviewers: platform-team
- Wait timer: 5 minutes
- Deployment branch: main only

## Monitoring and Quality Gates

### Code Coverage
- **Backend**: 80% minimum coverage
- **Frontend**: 80% minimum coverage
- **Reporting**: Codecov integration

### Security
- **Dependency Scanning**: GitHub Dependabot
- **Container Scanning**: Trivy vulnerability scanner
- **Runtime Scanning**: OWASP ZAP baseline scan

### Performance
- **Lighthouse CI**: Performance, accessibility, SEO audits
- **Bundle Analysis**: Frontend bundle size monitoring
- **Load Testing**: K6 performance testing (when configured)

## Branch Protection

### Main Branch
- Require PR before merging
- Require status checks to pass
- Require up-to-date branches
- Require review from code owners
- Restrict pushes to main branch

### Develop Branch
- Require PR before merging
- Require status checks to pass
- Allow force pushes for maintainers

## Usage Examples

### Manual Deployment
```bash
# Deploy to specific environment
gh workflow run deploy.yml -f environment=staging
```

### Infrastructure Updates
```bash
# Deploy infrastructure changes
gh workflow run infrastructure.yml -f environment=production
```

### Emergency Rollback
```bash
# Rollback to previous version
kubectl rollout undo deployment/hub-backend -n hub-production
kubectl rollout undo deployment/hub-frontend -n hub-production
```

## Troubleshooting

### Common Issues

1. **Build Failures**
   - Check Docker registry permissions
   - Verify secrets configuration
   - Review build logs for dependency issues

2. **Deployment Failures**
   - Verify AKS credentials
   - Check namespace permissions
   - Review Kubernetes resource limits

3. **Test Failures**
   - Check database service health
   - Verify test environment variables
   - Review test isolation issues

### Debugging Commands

```bash
# Check workflow status
gh run list --workflow=ci.yml

# View workflow logs
gh run view <run-id> --log

# Debug Kubernetes deployment
kubectl describe deployment hub-backend -n hub-production
kubectl logs -l app=hub-backend -n hub-production
```

## Best Practices

1. **Security**
   - Never commit secrets to repository
   - Use least privilege access for service accounts
   - Regularly rotate authentication credentials

2. **Performance**
   - Use build caching effectively
   - Optimize Docker layer ordering
   - Monitor resource usage

3. **Reliability**
   - Test deployments in staging first
   - Implement proper health checks
   - Monitor application metrics

4. **Maintenance**
   - Keep actions up to date
   - Review and update dependencies
   - Monitor workflow execution times

For more information, see the [GitHub Actions documentation](https://docs.github.com/en/actions).