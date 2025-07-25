# CI/CD Actions System Documentation

## Overview

The A5C Hub CI/CD Actions system provides a comprehensive GitHub Actions-compatible CI/CD platform with advanced features including real-time monitoring, artifact management, and multi-level runner support.

## Features

### Core Capabilities
- **GitHub Actions Compatibility**: Run existing workflows with compatible syntax
- **Real-time Job Execution**: Kubernetes-based job execution with live log streaming
- **Advanced Runner Management**: Repository, organization, and global runners
- **Artifact Management**: Complete artifact lifecycle with retention policies
- **Build Monitoring**: Real-time build status and log streaming
- **Webhook Integration**: Comprehensive webhook system with security verification

### Job Execution System
- **Kubernetes Execution**: Enhanced executor running jobs in ephemeral containers
- **Step-by-step Processing**: Individual step tracking with success/failure handling
- **Environment Variables**: Proper GitHub Actions environment variable support
- **Action Support**: Checkout actions, setup actions, and custom actions
- **Matrix Builds**: Infrastructure ready for matrix job expansion

## Architecture

### Components
```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Workflow      │    │   Job Executor  │    │   Runner Pool   │
│   Parser        │◄───│   (Kubernetes)  │◄───│   Management    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Webhook       │    │   Artifact      │    │   Log Streaming │
│   Processing    │    │   Storage       │    │   (SSE)         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Configuration

### Environment Variables

```bash
# CI/CD Configuration
ENABLE_ACTIONS=true
KUBERNETES_NAMESPACE=hub-runners
KUBERNETES_CONFIG_PATH=/config/kubeconfig

# Runner Configuration
RUNNER_REGISTRATION_TOKEN=your-runner-token
RUNNER_HEARTBEAT_INTERVAL=30
RUNNER_CLEANUP_INTERVAL=300

# Artifact Storage
ARTIFACT_STORAGE_BACKEND=azure_blob  # local, s3, azure_blob
ARTIFACT_RETENTION_DAYS=90
ARTIFACT_MAX_SIZE_MB=1024

# Webhook Configuration
WEBHOOK_SECRET=your-webhook-secret
WEBHOOK_TIMEOUT_SECONDS=30
```

### YAML Configuration

```yaml
actions:
  enabled: true
  kubernetes:
    namespace: "hub-runners"
    config_path: "/config/kubeconfig"
    image_pull_policy: "Always"
    resource_limits:
      cpu: "2"
      memory: "4Gi"
    resource_requests:
      cpu: "100m"
      memory: "256Mi"

runners:
  registration_token: "${RUNNER_TOKEN}"
  heartbeat_interval: "30s"
  cleanup_interval: "5m"
  max_concurrent_jobs: 10
  labels:
    - "self-hosted"
    - "kubernetes"

artifacts:
  storage:
    backend: "azure_blob"
    retention_days: 90
    max_size_mb: 1024
  azure:
    account_name: "${AZURE_STORAGE_ACCOUNT}"
    container_name: "artifacts"

webhooks:
  secret: "${WEBHOOK_SECRET}"
  timeout: "30s"
  verify_ssl: true
```

## Workflow Syntax

### Basic Workflow

```yaml
name: CI/CD Pipeline
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '18'
      - name: Install dependencies
        run: npm install
      - name: Run tests
        run: npm test

  build:
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Build application
        run: npm run build
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: build-artifacts
          path: dist/
```

### Advanced Features

```yaml
name: Advanced Pipeline
on:
  push:
    paths:
      - 'src/**'
      - '!docs/**'
  schedule:
    - cron: '0 2 * * *'
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deployment environment'
        required: true
        default: 'staging'
        type: choice
        options:
          - staging
          - production

jobs:
  build:
    runs-on: [self-hosted, kubernetes, gpu]
    environment: ${{ github.event.inputs.environment }}
    timeout-minutes: 30
    strategy:
      matrix:
        node-version: [16, 18, 20]
        os: [ubuntu-latest, windows-latest]
    
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      
      - name: Setup Node.js ${{ matrix.node-version }}
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run tests with coverage
        run: npm run test:coverage
        env:
          NODE_ENV: test
      
      - name: Upload coverage reports
        uses: actions/upload-artifact@v4
        with:
          name: coverage-${{ matrix.node-version }}-${{ matrix.os }}
          path: coverage/
          retention-days: 30
```

## Runner Management

### Runner Types

**Repository Runners**
```bash
# Register repository runner
curl -X POST /api/v1/repositories/owner/repo/runners \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "repo-runner-1",
    "labels": ["self-hosted", "linux", "x64"]
  }'
```

**Organization Runners**
```bash
# Register organization runner
curl -X POST /api/v1/organizations/org/runners \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "org-runner-1",
    "labels": ["self-hosted", "kubernetes", "gpu"]
  }'
```

**Global Runners**
```bash
# Register global runner (admin only)
curl -X POST /api/v1/admin/runners \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "global-runner-1",
    "labels": ["self-hosted", "macos", "arm64"]
  }'
```

### Runner Configuration

```bash
# Download runner
curl -o runner.tar.gz https://hub.example.com/runner/download/linux-x64

# Extract and configure
tar xzf runner.tar.gz
./config.sh \
  --url https://hub.example.com \
  --token $REGISTRATION_TOKEN \
  --name "my-runner" \
  --labels "self-hosted,linux,x64"

# Run as service
sudo ./svc.sh install
sudo ./svc.sh start
```

## Artifact Management

### Uploading Artifacts

```yaml
- name: Upload build artifacts
  uses: actions/upload-artifact@v4
  with:
    name: build-output
    path: |
      dist/
      build/
    retention-days: 30
    if-no-files-found: warn
    compression-level: 6
```

### Downloading Artifacts

```yaml
- name: Download artifacts
  uses: actions/download-artifact@v4
  with:
    name: build-output
    path: ./artifacts
```

### Artifact API

```bash
# List artifacts for a workflow run
GET /api/v1/repositories/owner/repo/actions/runs/123/artifacts

# Download specific artifact
GET /api/v1/repositories/owner/repo/actions/artifacts/456/zip

# Delete artifact
DELETE /api/v1/repositories/owner/repo/actions/artifacts/456
```

## Real-time Monitoring

### Live Log Streaming

```javascript
// Frontend: Subscribe to live logs
const eventSource = new EventSource(
  `/api/v1/repositories/owner/repo/actions/runs/123/logs/stream`
);

eventSource.onmessage = (event) => {
  const logData = JSON.parse(event.data);
  console.log(`[${logData.step}] ${logData.message}`);
};

eventSource.onerror = (error) => {
  console.error('Log stream error:', error);
  eventSource.close();
};
```

### Build Status Updates

```bash
# Get workflow run status
GET /api/v1/repositories/owner/repo/actions/runs/123

# Get job details
GET /api/v1/repositories/owner/repo/actions/runs/123/jobs

# Get step logs
GET /api/v1/repositories/owner/repo/actions/runs/123/jobs/456/logs
```

## Webhook Integration

### Webhook Configuration

```bash
# Create repository webhook
curl -X POST /api/v1/repositories/owner/repo/hooks \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "config": {
      "url": "https://external-service.com/webhook",
      "content_type": "json",
      "secret": "webhook-secret"
    },
    "events": ["push", "pull_request", "workflow_run"]
  }'
```

### Webhook Events

**Workflow Events**
- `workflow_run` - Workflow run started, completed, or failed
- `workflow_job` - Individual job started, completed, or failed
- `workflow_dispatch` - Manual workflow trigger

**Repository Events**
- `push` - Code pushed to repository
- `pull_request` - Pull request opened, updated, or merged
- `release` - Release created or published
- `issues` - Issue opened, closed, or updated

### Webhook Security

```go
// Verify webhook signature
func verifyWebhookSignature(payload []byte, signature string, secret string) bool {
    mac := hmac.New(sha256.New, []byte(secret))
    mac.Write(payload)
    expectedSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))
    return hmac.Equal([]byte(signature), []byte(expectedSignature))
}
```

## API Reference

### Workflow Runs

```bash
# List workflow runs
GET /api/v1/repositories/owner/repo/actions/runs?status=completed&branch=main

# Get specific run
GET /api/v1/repositories/owner/repo/actions/runs/123

# Cancel workflow run
POST /api/v1/repositories/owner/repo/actions/runs/123/cancel

# Re-run workflow
POST /api/v1/repositories/owner/repo/actions/runs/123/rerun
```

### Jobs and Steps

```bash
# List jobs for a run
GET /api/v1/repositories/owner/repo/actions/runs/123/jobs

# Get job details
GET /api/v1/repositories/owner/repo/actions/jobs/456

# Get job logs
GET /api/v1/repositories/owner/repo/actions/jobs/456/logs
```

### Manual Triggers

```bash
# Trigger workflow_dispatch
POST /api/v1/repositories/owner/repo/actions/workflows/ci.yml/dispatches
Content-Type: application/json

{
  "ref": "main",
  "inputs": {
    "environment": "staging",
    "debug": "true"
  }
}
```

## Performance Optimization

### Caching Strategies

```yaml
- name: Cache dependencies
  uses: actions/cache@v4
  with:
    path: ~/.npm
    key: ${{ runner.os }}-node-${{ hashFiles('**/package-lock.json') }}
    restore-keys: |
      ${{ runner.os }}-node-

- name: Cache build output
  uses: actions/cache@v4
  with:
    path: dist/
    key: build-${{ github.sha }}
```

### Resource Management

```yaml
jobs:
  build:
    runs-on: ubuntu-latest
    timeout-minutes: 30
    strategy:
      fail-fast: false
      max-parallel: 4
```

### Optimization Tips
- Use specific runner labels for job routing
- Implement caching for dependencies and build outputs
- Optimize Docker images for faster startup
- Use conditional job execution
- Implement job parallelization where possible

## Monitoring and Metrics

### Key Metrics
- **Job Success Rate**: Percentage of successful job executions
- **Average Job Duration**: Mean execution time across all jobs
- **Queue Time**: Time jobs spend waiting for available runners
- **Runner Utilization**: Percentage of time runners are actively working
- **Artifact Storage Usage**: Total storage consumed by artifacts

### Health Checks

```bash
# Check Actions system health
GET /api/v1/actions/health

# Check runner status
GET /api/v1/admin/runners/status

# Check artifact storage
GET /api/v1/actions/artifacts/stats
```

## Troubleshooting

### Common Issues

**Workflow not triggering**
- Check webhook configuration and delivery
- Verify trigger conditions (branches, paths, etc.)
- Review repository permissions
- Check event payload format

**Job stuck in queue**
- Verify runner availability and labels
- Check runner heartbeat status
- Review job requirements and constraints
- Monitor resource availability

**Artifact upload/download failures**
- Verify storage backend configuration
- Check network connectivity
- Review file size limits
- Validate authentication credentials

**Log streaming issues**
- Check WebSocket/SSE connectivity
- Verify authentication tokens
- Review network firewall rules
- Monitor server resources

### Debug Mode

```yaml
# Enable debug logging
steps:
  - name: Debug information
    run: |
      echo "Runner environment:"
      printenv | sort
      echo "Available disk space:"
      df -h
      echo "System resources:"
      free -h
    env:
      ACTIONS_STEP_DEBUG: true
```

## Security Considerations

### Access Control
- Repository-level permissions enforced
- Organization-level runner access control
- Secure secret management and injection
- Workflow approval requirements for protected branches

### Secret Management
- Encrypted secret storage
- Secure secret injection into job environments
- Audit logging for secret access
- Automatic secret masking in logs

### Runner Security
- Ephemeral job execution environments
- Network isolation between jobs
- Resource limits and quotas
- Regular security updates

## Migration Guide

### From GitHub Actions
1. Export existing workflow files
2. Update any GitHub-specific actions
3. Configure runners and webhooks
4. Test workflow execution
5. Update any external integrations

### From Other CI/CD Systems
1. Convert pipeline definitions to GitHub Actions format
2. Migrate secrets and environment variables
3. Set up equivalent runners
4. Test and validate workflows
5. Update deployment processes

## Best Practices

### Workflow Design
- Keep workflows focused and modular
- Use conditional execution to reduce unnecessary work
- Implement proper error handling and recovery
- Use secrets for sensitive configuration
- Document workflow purpose and requirements

### Performance
- Optimize Docker images for fast startup
- Use caching effectively
- Implement job parallelization
- Monitor resource usage
- Regular cleanup of old artifacts

### Security
- Use least privilege principle
- Regularly rotate secrets and tokens
- Monitor workflow execution logs
- Implement approval workflows for sensitive operations
- Keep runners and dependencies updated

## Support

For CI/CD Actions system issues:
- Check workflow logs and execution history
- Review runner status and availability
- Monitor system health endpoints
- Consult troubleshooting guides
- Contact system administrators

## References

- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [Actions API Reference](../api/actions.md)
- [Deployment Guide](../DEPLOYMENT.md)