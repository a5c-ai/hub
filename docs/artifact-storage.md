# Artifact Storage System Documentation

## Overview

The A5C Hub Artifact Storage system provides comprehensive artifact management for CI/CD workflows with support for multiple storage backends, automated retention policies, and secure access controls.

## Features

### Storage Backends
- **Filesystem Storage**: Local filesystem storage with configurable paths
- **Azure Blob Storage**: Enterprise Azure Blob Storage integration
- **S3-Compatible Storage**: Amazon S3 and S3-compatible storage providers
- **Pluggable Architecture**: Easy addition of new storage backends

### Artifact Management
- **Upload/Download**: Secure artifact upload and download with streaming support
- **Metadata Storage**: Comprehensive artifact information and statistics
- **Lifecycle Management**: Automated retention policies and cleanup
- **Size Validation**: Configurable size limits and validation
- **Access Control**: Authentication-protected operations with proper permissions

### Build Log Storage
- **Log Storage**: Store and retrieve job logs via storage backend
- **Search Integration**: Search capabilities (with Elasticsearch integration)
- **Retention Policies**: Automated cleanup based on configurable retention periods
- **Streaming Access**: Efficient log streaming for real-time viewing

## Architecture

### Storage Backend Interface
```go
type StorageBackend interface {
    Store(key string, data io.Reader, metadata map[string]string) error
    Retrieve(key string) (io.ReadCloser, error)
    Delete(key string) error
    List(prefix string) ([]string, error)
    GetMetadata(key string) (map[string]string, error)
    GetSize(key string) (int64, error)
}
```

### Backend Implementations
- **FilesystemBackend**: Complete implementation with local storage
- **AzureBlobBackend**: Interface ready (requires Azure SDK integration)
- **S3Backend**: Interface ready (requires AWS SDK integration)

## Configuration

### Environment Variables
```bash
# Storage Backend Selection
STORAGE_BACKEND=filesystem  # Options: filesystem, azure, s3

# Filesystem Configuration
STORAGE_BASE_PATH=/var/lib/hub/artifacts
STORAGE_MAX_SIZE_MB=1024
STORAGE_RETENTION_DAYS=90

# Azure Blob Configuration
AZURE_STORAGE_ACCOUNT_NAME=hubstorage
AZURE_STORAGE_ACCOUNT_KEY=your-account-key
AZURE_STORAGE_CONTAINER_NAME=artifacts

# S3 Configuration
S3_REGION=us-west-2
S3_BUCKET=hub-artifacts
S3_ACCESS_KEY_ID=your-access-key
S3_SECRET_ACCESS_KEY=your-secret-key
S3_USE_SSL=true
```

### Configuration File (config.yaml)
```yaml
storage:
  artifacts:
    backend: "filesystem"  # Options: filesystem, azure, s3
    base_path: "/var/lib/hub/artifacts"
    max_size_mb: 1024
    retention_days: 90
    
    # Azure Blob Storage
    azure:
      account_name: "hubstorage"
      account_key: "your-account-key"
      container_name: "artifacts"
    
    # S3-Compatible Storage
    s3:
      region: "us-west-2"
      bucket: "hub-artifacts"
      access_key_id: "your-access-key"
      secret_access_key: "your-secret-key"
      use_ssl: true
      endpoint: ""  # Optional: for S3-compatible services
```

## API Endpoints

### Artifact Operations
```bash
# List artifacts for a workflow run
GET /api/v1/repos/{owner}/{repo}/actions/runs/{runId}/artifacts

# Upload artifact (multipart form data)
POST /api/v1/repos/{owner}/{repo}/actions/runs/{runId}/artifacts
Content-Type: multipart/form-data

# Get artifact metadata
GET /api/v1/repos/{owner}/{repo}/actions/runs/{runId}/artifacts/{artifactId}

# Download artifact
GET /api/v1/repos/{owner}/{repo}/actions/artifacts/{artifactId}/download

# Delete artifact
DELETE /api/v1/repos/{owner}/{repo}/actions/artifacts/{artifactId}
```

### Build Log Operations
```bash
# Get build logs for a job
GET /api/v1/repos/{owner}/{repo}/actions/runs/{runId}/jobs/{jobId}/logs

# Stream live build logs (Server-Sent Events)
GET /api/v1/repos/{owner}/{repo}/actions/runs/{runId}/jobs/{jobId}/logs/stream

# Search build logs (requires Elasticsearch)
GET /api/v1/repos/{owner}/{repo}/actions/logs/search?q={query}
```

## Usage Examples

### Uploading Artifacts
```bash
# Upload a build artifact
curl -X POST \
  -H "Authorization: Bearer {token}" \
  -F "name=build-artifacts" \
  -F "artifact=@build.zip" \
  "https://hub.example.com/api/v1/repos/owner/repo/actions/runs/123/artifacts"
```

Response:
```json
{
  "id": "artifact_123",
  "name": "build-artifacts",
  "size": 2048576,
  "created_at": "2024-01-15T10:30:00Z",
  "expires_at": "2024-04-15T10:30:00Z",
  "download_url": "/api/v1/repos/owner/repo/actions/artifacts/artifact_123/download"
}
```

### Downloading Artifacts
```bash
# Download artifact by ID
curl -H "Authorization: Bearer {token}" \
  "https://hub.example.com/api/v1/repos/owner/repo/actions/artifacts/artifact_123/download" \
  -o artifact.zip
```

### Listing Artifacts
```bash
# List all artifacts for a workflow run
curl -H "Authorization: Bearer {token}" \
  "https://hub.example.com/api/v1/repos/owner/repo/actions/runs/123/artifacts"
```

Response:
```json
{
  "total_count": 3,
  "artifacts": [
    {
      "id": "artifact_123",
      "name": "build-artifacts",
      "size": 2048576,
      "created_at": "2024-01-15T10:30:00Z",
      "expires_at": "2024-04-15T10:30:00Z"
    },
    {
      "id": "artifact_124",
      "name": "test-results",
      "size": 512000,
      "created_at": "2024-01-15T10:32:00Z",
      "expires_at": "2024-04-15T10:32:00Z"
    }
  ]
}
```

## Storage Backend Details

### Filesystem Backend
- **Complete Implementation**: Fully implemented and tested
- **Path Management**: Organized storage with configurable base paths
- **Metadata Storage**: JSON metadata files alongside artifacts
- **Atomic Operations**: Safe concurrent access with proper locking
- **Cleanup**: Automated cleanup of expired artifacts

### Azure Blob Storage Backend
- **Implementation**: Fully implemented using Azure Storage SDK
- **SDK Integration**: `github.com/Azure/azure-storage-blob-go`
- **Container Management**: Automatic container creation and management
- **Metadata Support**: Azure Blob metadata for artifact information
- **Secure Access**: SAS tokens for secure artifact access

### S3-Compatible Backend
- **Interface Ready**: Complete interface implementation  
- **SDK Integration**: Requires `github.com/aws/aws-sdk-go-v2`
- **Bucket Management**: Automatic bucket creation and configuration
- **Versioning Support**: S3 versioning for artifact history
- **Cross-Region**: Support for multi-region deployments

## Security Features

### Access Control
- **Authentication Required**: All operations require valid authentication
- **Permission Checking**: Repository-level permissions enforced
- **Secure URLs**: Time-limited download URLs for enhanced security
- **Audit Logging**: All artifact operations logged for compliance

### Data Protection
- **Encryption at Rest**: Backend-specific encryption support
- **Encryption in Transit**: TLS for all data transfers
- **Integrity Checking**: Checksums for data integrity verification
- **Access Logging**: Comprehensive access logging and monitoring

## Retention and Cleanup

### Retention Policies
```yaml
retention:
  default_days: 90
  max_days: 365
  min_days: 1
  
  # Per-repository overrides
  repository_overrides:
    "owner/critical-repo": 180
    "owner/temp-repo": 30
```

### Cleanup Process
- **Scheduled Cleanup**: Automated cleanup based on retention policies
- **Graceful Deletion**: Proper cleanup of both storage and database records
- **Monitoring**: Cleanup metrics and reporting
- **Manual Cleanup**: Admin tools for manual cleanup operations

## Performance Considerations

### Optimization Strategies
- **Streaming Operations**: Efficient handling of large artifacts
- **Parallel Processing**: Concurrent upload/download operations
- **Caching**: Metadata caching for improved performance
- **Connection Pooling**: Optimized backend connections

### Scaling Considerations
- **Horizontal Scaling**: Multiple storage backend instances
- **Load Balancing**: Distribute storage operations across backends
- **CDN Integration**: Content delivery network for faster downloads
- **Monitoring**: Performance metrics and alerting

## Monitoring and Observability

### Metrics
```yaml
# Storage metrics
- hub_artifacts_total
- hub_artifacts_size_bytes
- hub_artifacts_upload_duration_seconds
- hub_artifacts_download_duration_seconds
- hub_storage_operations_total
- hub_storage_errors_total
```

### Health Checks
```bash
# Storage backend health
GET /api/v1/storage/health

# Artifact system status
GET /api/v1/artifacts/status
```

### Logging
- **Structured Logging**: JSON-formatted logs with correlation IDs
- **Error Tracking**: Detailed error logging and stack traces
- **Performance Logging**: Request duration and size metrics
- **Audit Logging**: Security-relevant operations logged

## Troubleshooting

### Common Issues

1. **Upload Failures**
   - Check storage backend connectivity
   - Verify available disk space (filesystem backend)
   - Check authentication credentials
   - Review file size limits

2. **Download Issues**
   - Verify artifact exists and hasn't expired
   - Check user permissions
   - Test storage backend connectivity
   - Review network connectivity

3. **Storage Backend Errors**
   - Validate backend configuration
   - Check service credentials
   - Test network connectivity
   - Review backend service status

### Diagnostic Commands
```bash
# Test storage backend connectivity
hub admin storage test

# Check artifact status
hub admin artifacts status

# Cleanup expired artifacts
hub admin artifacts cleanup --dry-run

# Storage backend diagnostics
hub admin storage diagnostics
```

## Migration and Backup

### Backend Migration
```bash
# Migrate from filesystem to Azure Blob
hub admin storage migrate --from filesystem --to azure --config migration.yaml

# Backup artifacts
hub admin storage backup --backend filesystem --destination /backup/artifacts
```

### Data Recovery
- **Point-in-time Recovery**: Restore artifacts from specific dates
- **Selective Recovery**: Restore specific repositories or runs
- **Cross-backend Recovery**: Migrate between storage backends
- **Integrity Verification**: Verify restored data integrity

## Future Enhancements

### Planned Features
- **Artifact Deduplication**: Reduce storage usage with content deduplication
- **Compression**: Automatic compression for space efficiency
- **Replication**: Multi-backend replication for high availability
- **Caching Layer**: Local caching for frequently accessed artifacts
- **Advanced Search**: Enhanced search capabilities across artifacts
- **Webhook Integration**: Artifact lifecycle event webhooks
