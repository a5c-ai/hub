# Search System Implementation

This document describes the search system implementation with Elasticsearch integration.

## Overview

The search system provides advanced full-text search capabilities across all content types in the platform:

- **Users**: Search by username, full name, bio, company, location
- **Repositories**: Search by name, description, language, topics
- **Issues**: Search by title, body, labels, assignees
- **Commits**: Search by message, author, committer
- **Organizations**: Search by name, description, location
- **Code**: Search within repository file contents (requires Elasticsearch)

## Architecture

### Backend Components

1. **ElasticsearchService** (`internal/services/elasticsearch_service.go`)
   - Handles connection to Elasticsearch cluster
   - Manages index creation and mapping
   - Provides document indexing and search methods

2. **SearchService** (`internal/services/search_service.go`) 
   - Main search service with hybrid approach
   - Uses Elasticsearch when available, falls back to PostgreSQL
   - Provides automatic document indexing methods

3. **Search Handlers** (`internal/api/search_handlers.go`)
   - REST API endpoints for search operations
   - Handles query parameters and pagination
   - Returns structured search results

### Frontend Components

1. **Search Page** (`frontend/src/app/search/page.tsx`)
   - Global search interface with type filtering
   - Displays results for all content types
   - Supports pagination and sorting

2. **Search API** (`frontend/src/lib/api.ts`)
   - Client-side API methods for search operations
   - Handles query building and parameter encoding

## Configuration

### Elasticsearch Configuration

Add to your `config.yaml`:

```yaml
elasticsearch:
  enabled: true
  addresses:
    - http://localhost:9200
  username: ""           # Optional: for authenticated clusters
  password: ""           # Optional: for authenticated clusters
  cloud_id: ""           # Optional: for Elastic Cloud
  api_key: ""            # Optional: API key authentication
  index_prefix: "hub"    # Prefix for all indices
```

### Environment Variables

```bash
ELASTICSEARCH_ENABLED=true
ELASTICSEARCH_ADDRESSES=http://localhost:9200
ELASTICSEARCH_USERNAME=elastic
ELASTICSEARCH_PASSWORD=password
ELASTICSEARCH_INDEX_PREFIX=hub
```

## Docker Setup

Use the provided `docker-compose.yml` to run Elasticsearch locally:

```bash
docker-compose up -d elasticsearch
```

This starts Elasticsearch on `http://localhost:9200` with security disabled for development.

## API Endpoints

### Global Search
```
GET /api/v1/search?q=query&type=all&page=1&per_page=30
```

### Type-Specific Search
```
GET /api/v1/search/repositories?q=query&language=go&sort=stars
GET /api/v1/search/issues?q=query&state=open&labels=bug
GET /api/v1/search/users?q=query&sort=followers
GET /api/v1/search/commits?q=query&sort=author-date
GET /api/v1/search/code?q=query&language=go&repo=uuid
```

### Repository-Specific Search
```
GET /api/v1/repositories/:owner/:repo/search?q=query&type=issues
```

## Search Features

### Advanced Query Syntax

The search supports various query types:

- **Simple text**: `user authentication`
- **Quoted phrases**: `"user authentication"`
- **Wildcards**: `auth*`
- **Fuzzy search**: `authentiaction~` (with typos)
- **Field-specific**: `title:bug` (when implemented)

### Filters

- **Repository visibility**: `public`, `private`, `internal`
- **Issue state**: `open`, `closed`
- **Programming language**: `go`, `javascript`, `python`, etc.
- **Date ranges**: `created:>2023-01-01`
- **User/organization**: `user:username`, `org:orgname`

### Sorting Options

- **Relevance**: `_score` (default)
- **Date**: `created`, `updated`, `pushed`
- **Popularity**: `stars`, `forks`, `watchers`
- **Activity**: `comments`

## Indexing

### Automatic Indexing

Documents are automatically indexed when:

- Users are created or updated
- Repositories are created, updated, or deleted
- Issues are created, updated, or closed
- Commits are pushed to repositories
- Organizations are created or updated

### Manual Reindexing

Use the reindex command to rebuild all indices:

```bash
go run cmd/reindex/main.go
```

This command:
1. Connects to the database
2. Fetches all existing data
3. Bulk indexes everything to Elasticsearch
4. Reports progress and completion

### Code Indexing

Code files are indexed when:
- Repository content is pushed via Git
- Files are created/updated via the API
- Manual indexing is triggered

To index code files, call:
```go
searchService.IndexCodeFile(repoID, repoName, filePath, content, language, branch, sha)
```

## Performance

### Elasticsearch Performance

- **Indices**: Separate indices for each content type
- **Sharding**: Single shard for development, multiple for production
- **Replication**: No replicas for development, 1+ for production
- **Caching**: Built-in Elasticsearch caching for repeated queries

### PostgreSQL Fallback

When Elasticsearch is unavailable:
- Falls back to PostgreSQL full-text search
- Uses `tsvector` and `tsquery` for text search
- Limited functionality compared to Elasticsearch
- Automatic failover and recovery

### Query Optimization

- **Pagination**: Efficient offset-based pagination
- **Field selection**: Only indexes necessary fields
- **Highlighting**: Search term highlighting in results
- **Relevance scoring**: Elasticsearch BM25 algorithm

## Monitoring

### Health Checks

Check Elasticsearch health:
```bash
curl http://localhost:9200/_cluster/health
```

Check index status:
```bash
curl http://localhost:9200/hub_*/_stats
```

### Logging

The search service logs:
- Connection status
- Index creation/updates
- Search performance metrics
- Error conditions and fallbacks

### Metrics

Monitor these metrics:
- Search request rate
- Search latency (p95, p99)
- Index size and growth
- Elasticsearch cluster health
- Fallback usage rate

## Troubleshooting

### Common Issues

1. **Elasticsearch connection failed**
   - Check if Elasticsearch is running
   - Verify connection settings in config
   - Check network connectivity

2. **Search returns no results**
   - Verify data has been indexed
   - Check index mappings
   - Try simpler queries first

3. **Poor search relevance**
   - Review field mappings
   - Adjust boosting for important fields
   - Consider custom analyzers

4. **Indexing errors**
   - Check Elasticsearch logs
   - Verify index templates
   - Monitor disk space

### Debug Mode

Enable debug logging:
```yaml
log_level: 5  # Debug level
```

This provides detailed logs of:
- Search queries executed
- Elasticsearch responses
- Indexing operations
- Performance metrics

## Development

### Running Tests

```bash
go test ./internal/services -v
```

### Adding New Content Types

1. Define document struct in `elasticsearch_service.go`
2. Create index mapping method
3. Add indexing method to `SearchService`
4. Update conversion helpers
5. Add API endpoints if needed

### Customizing Search

- **Analyzers**: Modify mapping for language-specific analysis
- **Boosting**: Adjust field importance in queries  
- **Filters**: Add new filter types in handlers
- **Aggregations**: Add faceted search capabilities

## Production Deployment

### Elasticsearch Cluster

- Use managed Elasticsearch (AWS, GCP, Elastic Cloud)
- Configure appropriate instance sizes
- Set up monitoring and alerting
- Enable security and authentication

### Index Management

- Use index templates for consistent mappings
- Implement index rotation for large datasets
- Set up automated backups
- Monitor index growth and performance

### Security

- Enable Elasticsearch security features
- Use API keys or certificates for authentication
- Implement request rate limiting
- Log and monitor search queries