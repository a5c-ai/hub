# Analytics System Documentation

## Overview

The A5C Hub Analytics system provides comprehensive real-time analytics for users, organizations, and repositories with advanced data processing, performance metrics, and data export capabilities.

## Features

### Core Analytics
- **User Analytics**: Repository statistics, contribution metrics, and activity tracking
- **Organization Analytics**: Member statistics, repository metrics, team analytics, and security insights
- **Repository Analytics**: Code statistics, language detection, contribution patterns, and activity metrics
- **Performance Metrics**: Build times, success rates, resource usage with percentile calculations
- **Real-time Data**: Live analytics with database-driven queries and statistical processing

### Data Export
- **Multiple Formats**: Export analytics data in JSON, CSV, and XLSX formats
- **Flexible Queries**: Custom date ranges, filtering, and aggregation options
- **Batch Processing**: Efficient data processing for large datasets
- **Scheduled Reports**: Automated report generation and delivery

### Technical Implementation
- **Database-Driven**: Real-time queries using GORM with proper error handling
- **Statistical Processing**: Percentile calculations, trend analysis, and data aggregation
- **Time Series Analysis**: Historical data tracking with trend calculations
- **Helper Functions**: Comprehensive data processing and analytics utilities

## API Endpoints

### User Analytics
```bash
# Get user repository analytics
GET /api/v1/users/{username}/analytics
GET /api/v1/users/{username}/analytics/repositories

# Get public user analytics
GET /api/v1/users/{username}/analytics/public
```

### Organization Analytics
```bash
# Get organization overview
GET /api/v1/orgs/{org}/analytics
GET /api/v1/orgs/{org}/analytics/overview

# Get organization insights
GET /api/v1/orgs/{org}/analytics/insights

# Get member analytics
GET /api/v1/orgs/{org}/analytics/members

# Get repository analytics
GET /api/v1/orgs/{org}/analytics/repositories

# Get team analytics
GET /api/v1/orgs/{org}/analytics/teams

# Get security analytics
GET /api/v1/orgs/{org}/analytics/security
```

### Repository Analytics
```bash
# Get repository statistics
GET /api/v1/repos/{owner}/{repo}/analytics
GET /api/v1/repos/{owner}/{repo}/analytics/stats

# Get language statistics
GET /api/v1/repos/{owner}/{repo}/analytics/languages

# Get contributor analytics
GET /api/v1/repos/{owner}/{repo}/analytics/contributors
```

### Performance Metrics
```bash
# Get usage analytics (admin only)
GET /api/v1/analytics/usage

# Get cost analytics (admin only)
GET /api/v1/analytics/cost

# Get performance metrics
GET /api/v1/analytics/performance
```

### Data Export
```bash
# Export analytics data
GET /api/v1/analytics/export?format={json|csv|xlsx}
GET /api/v1/analytics/export?format=csv&start_date=2024-01-01&end_date=2024-12-31

# Export specific analytics
GET /api/v1/orgs/{org}/analytics/export?format=xlsx
GET /api/v1/users/{username}/analytics/export?format=json
```

## Configuration

### Database Configuration
The analytics system uses the existing database configuration with optimized queries for performance:

```yaml
database:
  host: postgresql
  port: 5432
  name: hub
  user: hub
  ssl_mode: require
  # Analytics-specific optimizations
  max_connections: 100
  query_timeout: 30s
```

### Analytics Settings
```yaml
analytics:
  enabled: true
  # Data retention (in days)
  retention_days: 365
  # Export limits
  max_export_rows: 100000
  # Cache settings
  cache_ttl: 300  # 5 minutes
  # Performance settings
  batch_size: 1000
  query_timeout: 30s
```

## Usage Examples

### Getting Organization Analytics
```bash
curl -H "Authorization: Bearer {token}" \
  "https://hub.example.com/api/v1/orgs/myorg/analytics"
```

Response:
```json
{
  "overview": {
    "total_members": 150,
    "total_repositories": 75,
    "total_teams": 12,
    "active_members_30d": 120
  },
  "repositories": {
    "public": 25,
    "private": 50,
    "total_commits_30d": 1250,
    "top_languages": ["TypeScript", "Go", "Python"]
  },
  "activity": {
    "pull_requests_30d": 89,
    "issues_30d": 156,
    "releases_30d": 12
  }
}
```

### Exporting Data
```bash
curl -H "Authorization: Bearer {token}" \
  "https://hub.example.com/api/v1/analytics/export?format=csv&start_date=2024-01-01" \
  -o analytics_export.csv
```

### Performance Metrics
```bash
curl -H "Authorization: Bearer {token}" \
  "https://hub.example.com/api/v1/analytics/performance"
```

Response:
```json
{
  "build_times": {
    "average_ms": 125000,
    "p50_ms": 95000,
    "p90_ms": 180000,
    "p99_ms": 350000
  },
  "success_rates": {
    "overall": 0.94,
    "last_30d": 0.96,
    "trend": "improving"
  },
  "resource_usage": {
    "cpu_average": 0.65,
    "memory_average_gb": 2.4,
    "storage_used_gb": 1250.5
  }
}
```

## Security and Access Control

### Authentication
All analytics endpoints require authentication via:
- JWT token in Authorization header
- Valid API token with analytics permissions
- Session-based authentication for web interface

### Authorization
- **User Analytics**: Users can view their own analytics
- **Organization Analytics**: Organization members can view org analytics
- **Admin Analytics**: System-wide analytics require admin permissions
- **Data Export**: Requires appropriate read permissions for the data being exported

### Rate Limiting
```yaml
rate_limits:
  analytics_api: 100/hour
  export_api: 10/hour
  admin_analytics: 1000/hour
```

## Performance Considerations

### Database Optimization
- Indexed queries for common analytics operations
- Query result caching for frequently accessed data
- Batch processing for large data exports
- Connection pooling for concurrent requests

### Scaling
- Read replicas for analytics queries
- Background job processing for heavy computations
- Horizontal scaling support for high-traffic scenarios
- Result caching with configurable TTL

### Monitoring
- Query performance metrics
- Export operation tracking
- Cache hit/miss ratios
- Error rate monitoring

## Integration with Other Systems

### CI/CD Integration
Analytics data includes build metrics, job success rates, and pipeline performance data from the CI/CD system.

### Search Integration
Analytics can be enhanced with search data when Elasticsearch is enabled, providing insights into search patterns and content discovery.

### Audit Integration
All analytics access is logged in the audit system for compliance and security monitoring.

## Troubleshooting

### Common Issues

1. **Slow Analytics Queries**
   - Check database indices
   - Review query complexity
   - Consider increasing query timeout
   - Enable query result caching

2. **Export Timeouts**
   - Reduce date range
   - Use smaller batch sizes
   - Export data in multiple requests
   - Check export row limits

3. **Missing Data**
   - Verify data retention settings
   - Check service permissions
   - Review database connectivity
   - Validate analytics configuration

### Performance Optimization
```yaml
# Example optimizations
analytics:
  query_timeout: 60s
  batch_size: 500
  cache_ttl: 600
  enable_query_cache: true
  max_concurrent_exports: 5
```

## Future Enhancements

### Planned Features
- Real-time dashboard updates with WebSocket support
- Custom analytics dashboards and visualizations
- Machine learning insights and predictions
- Advanced filtering and drill-down capabilities
- Automated alerting based on analytics thresholds