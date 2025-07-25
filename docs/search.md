# Advanced Search System Documentation

## Overview

The A5C Hub advanced search system provides powerful search capabilities across all content types with Elasticsearch integration and PostgreSQL fallback. The system supports global search, code search, and advanced filtering options.

## Features

### Search Capabilities
- **Global Search**: Search across users, repositories, issues, commits, and organizations
- **Code Search**: Search within repository files with syntax highlighting
- **Advanced Filtering**: Filter by language, visibility, state, labels, dates, and more
- **Fuzzy Matching**: Find content even with typos or partial matches
- **Real-time Results**: Fast search with pagination and result highlighting
- **Relevance Ranking**: BM25 algorithm for optimal result ordering

### Hybrid Architecture
- **Primary**: Elasticsearch for advanced search features
- **Fallback**: PostgreSQL full-text search when Elasticsearch unavailable
- **Graceful Degradation**: Automatic failover with no service interruption
- **No Breaking Changes**: Existing functionality preserved

## Configuration

### Environment Variables

```bash
# Elasticsearch Configuration
ELASTICSEARCH_ENABLED=true
ELASTICSEARCH_ADDRESSES=http://localhost:9200
ELASTICSEARCH_USERNAME=
ELASTICSEARCH_PASSWORD=
ELASTICSEARCH_CLOUD_ID=
ELASTICSEARCH_API_KEY=
ELASTICSEARCH_INDEX_PREFIX=hub
```

### YAML Configuration

```yaml
elasticsearch:
  enabled: true
  addresses: ["http://localhost:9200"]
  username: ""
  password: ""
  cloud_id: ""
  api_key: ""
  index_prefix: "hub"
  settings:
    number_of_shards: 1
    number_of_replicas: 0
```

### Docker Compose Setup

```yaml
version: '3.8'
services:
  elasticsearch:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.15.0
    container_name: hub-elasticsearch
    environment:
      - discovery.type=single-node
      - xpack.security.enabled=false
      - "ES_JAVA_OPTS=-Xms512m -Xmx512m"
    ports:
      - "9200:9200"
    volumes:
      - elasticsearch_data:/usr/share/elasticsearch/data

volumes:
  elasticsearch_data:
    driver: local
```

## API Endpoints

### Global Search
```bash
# Search across all content types
GET /api/v1/search?q=authentication&type=repositories&page=1&limit=20

# Search with filters
GET /api/v1/search?q=api&type=repositories&language=typescript&visibility=public
```

### Repository Search
```bash
# Repository-specific search
GET /api/v1/search/repositories?q=docker&language=go&sort=updated&order=desc

# Search with ownership filter
GET /api/v1/search/repositories?q=backend&user=a5c-ai&visibility=private
```

### Issue Search
```bash
# Search issues and PRs
GET /api/v1/search/issues?q=bug&state=open&labels=priority-high

# Search by assignee
GET /api/v1/search/issues?q=login&assignee=developer&type=issue
```

### User Search
```bash
# Search users
GET /api/v1/search/users?q=john&location=Seattle&sort=followers

# Search by company
GET /api/v1/search/users?q=developer&company=a5c-ai
```

### Code Search
```bash
# Search code (requires Elasticsearch)
GET /api/v1/search/code?q=func main&language=go&repo=a5c-ai/hub

# Search with file extension filter
GET /api/v1/search/code?q=interface&extension=ts&path=src/
```

### Repository-Scoped Search
```bash
# Search within specific repository
GET /api/v1/repositories/a5c-ai/hub/search?q=authentication&type=issues

# Search commits in repository
GET /api/v1/repositories/a5c-ai/hub/search?q=fix&type=commits&author=developer
```

## Search Indices

### Users Index
Fields indexed for user search:
- Username (boosted)
- Full name (boosted)
- Email
- Bio
- Company
- Location

### Repositories Index
Fields indexed for repository search:
- Name (boosted)
- Description (boosted)
- Language
- Topics/tags
- Visibility level
- Owner information

### Issues Index
Fields indexed for issue/PR search:
- Title (boosted)
- Body content
- Labels
- State (open/closed)
- Assignees
- Author

### Commits Index
Fields indexed for commit search:
- Commit message (boosted)
- Author name
- Author email
- Committer information
- File paths modified

### Code Index
Fields indexed for code search:
- File content
- File path
- File extension
- Programming language
- Repository information

### Organizations Index
Fields indexed for organization search:
- Name (boosted)
- Description
- Location
- Website

## Query Syntax

### Basic Search
```bash
# Simple text search
q=authentication

# Phrase search
q="multi factor authentication"

# Wildcard search
q=auth*
```

### Advanced Syntax
```bash
# Boolean operators
q=authentication AND (oauth OR saml)
q=bug NOT fixed

# Field-specific search
q=author:developer
q=language:typescript
q=extension:ts

# Date range search
q=created:>2024-01-01
q=updated:2024-01-01..2024-12-31

# Numeric range search
q=size:>1000
q=stars:10..100
```

### Code Search Syntax
```bash
# Function search
q="func main" language:go

# Class search
q="class User" extension:ts

# Import search
q="import React" path:src/components/

# Comment search
q="TODO:" language:typescript
```

## Performance Optimization

### Indexing Strategy
- **Bulk Operations**: Efficient bulk indexing for large datasets
- **Incremental Updates**: Real-time indexing on content changes
- **Index Templates**: Optimized mappings for each content type
- **Field Boosting**: Relevance tuning for better results

### Caching
- **Query Result Caching**: Cache frequent search results
- **Connection Pooling**: Efficient Elasticsearch connections
- **Request Deduplication**: Avoid duplicate search requests

### Performance Metrics
- Search response times < 1 second for 95% of queries
- Index update latency < 5 seconds
- Memory usage optimization with configurable analyzers
- Concurrent request handling with connection pooling

## Data Management

### Index Management
```bash
# Create indices
go run cmd/reindex/main.go --create-indices

# Reindex all data
go run cmd/reindex/main.go --reindex-all

# Reindex specific type
go run cmd/reindex/main.go --type=repositories

# Check index health
go run cmd/reindex/main.go --health-check
```

### Backup and Recovery
```bash
# Backup Elasticsearch indices
curl -X PUT "localhost:9200/_snapshot/backup/snapshot_1"

# Restore from backup
curl -X POST "localhost:9200/_snapshot/backup/snapshot_1/_restore"

# Monitor backup status
curl -X GET "localhost:9200/_snapshot/backup/_all"
```

## Monitoring and Maintenance

### Health Checks
- `/api/v1/search/health` - Search system health status
- `/api/v1/search/stats` - Search performance statistics
- `/api/v1/search/indices` - Index information and status

### Metrics
Monitor these key metrics:
- **Search Latency**: Average and P95 response times
- **Index Size**: Disk usage per index
- **Query Rate**: Searches per second
- **Error Rate**: Failed search percentage
- **Index Rate**: Documents indexed per second

### Maintenance Tasks
- **Index Optimization**: Regular index optimization for performance
- **Data Cleanup**: Remove stale or deleted content from indices
- **Mapping Updates**: Update index mappings for new features
- **Performance Tuning**: Adjust analyzer and search settings

## Troubleshooting

### Common Issues

**Elasticsearch not available**
- Check Elasticsearch service status
- Verify network connectivity
- Review authentication credentials
- Check system automatically falls back to PostgreSQL

**Search results missing**
- Verify data has been indexed
- Check index health and mapping
- Run reindex command if needed
- Review search query syntax

**Slow search performance**
- Monitor Elasticsearch cluster health
- Check system resources (CPU, memory)
- Review query complexity
- Consider index optimization

**Index synchronization issues**
- Check real-time indexing logs
- Verify webhook processing
- Run incremental reindex
- Monitor index update latency

### Debug Mode
Enable detailed logging for troubleshooting:

```yaml
elasticsearch:
  debug: true
  log_requests: true
  log_responses: true
```

### Performance Diagnostics
```bash
# Check Elasticsearch cluster stats
curl "localhost:9200/_cluster/stats?pretty"

# Monitor search performance
curl "localhost:9200/_search?explain=true"

# Check index statistics
curl "localhost:9200/hub_*/_stats?pretty"
```

## Migration Guide

### Enabling Search
1. Configure Elasticsearch connection
2. Run initial data indexing
3. Test search functionality
4. Enable search features in UI
5. Monitor performance and usage

### Upgrading Elasticsearch
1. Backup existing indices
2. Update Elasticsearch version
3. Reindex data if needed
4. Test search functionality
5. Monitor for compatibility issues

### Disabling Elasticsearch
1. Set `elasticsearch.enabled: false`
2. Search automatically falls back to PostgreSQL
3. No data loss or service interruption
4. Can re-enable at any time

## Best Practices

### Search Query Design
- Use specific search terms when possible
- Leverage field-specific searches for precision
- Combine filters to narrow results
- Use pagination for large result sets

### Index Management
- Regular index optimization
- Monitor index size and growth
- Use appropriate retention policies
- Keep mappings up to date

### Performance Optimization
- Cache frequent search queries
- Use connection pooling
- Monitor and tune Elasticsearch settings
- Regular performance testing

## Security Considerations

### Access Control
- Search respects existing permission models
- Users only see content they have access to
- Organization and repository visibility enforced
- Private content properly filtered

### Data Protection
- Search indices contain only necessary data
- Sensitive information excluded from indexing
- Regular audit of indexed content
- Compliance with data protection regulations

## Support

For search system issues:
- Check Elasticsearch logs
- Review search API responses
- Monitor system health endpoints
- Consult troubleshooting guide
- Contact system administrators

## References

- [Elasticsearch Documentation](https://www.elastic.co/guide/en/elasticsearch/reference/current/index.html)
- [Search API Reference](../api/search.md)
- [Performance Tuning Guide](performance.md)
- [Deployment Guide](../DEPLOYMENT.md)