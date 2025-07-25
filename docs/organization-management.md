# Organization Management Documentation

## Overview

The A5C Hub organization management system provides enterprise-grade features for managing large teams and organizations with advanced permissions, policies, analytics, and compliance capabilities.

## Features

### Core Capabilities
- **Custom Roles**: Granular permission system with custom role creation
- **Policy Enforcement**: Repository creation, member invitation, and compliance policies
- **Team Hierarchies**: Parent/child team relationships with inheritance
- **Advanced Analytics**: Comprehensive organization metrics and insights
- **Audit Logging**: Enhanced activity tracking and compliance reporting
- **Template System**: Organization and team templates for standardized setups

### Advanced Permission System
- **Role-Based Access Control (RBAC)**: Fine-grained permissions with inheritance
- **Custom Roles**: Create roles tailored to organizational needs
- **Permission Templates**: Pre-configured role templates for common scenarios
- **Color-Coded Roles**: Visual distinction for role identification
- **Repository-Level Permissions**: Granular access control per repository

## Custom Roles and Permissions

### Creating Custom Roles

```bash
# Create a custom role via API
curl -X POST /api/v1/organizations/acme/roles \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "DevOps Engineer",
    "description": "Full access to repositories and CI/CD",
    "color": "#FF6B35",
    "permissions": {
      "repositories": "admin",
      "actions": "write",
      "secrets": "read",
      "packages": "write",
      "team_management": "read"
    }
  }'
```

### Permission Levels

**Repository Permissions**
- `read` - Read access to repositories
- `write` - Write access and pull request creation
- `admin` - Full repository administration

**Organization Permissions**
- `member` - Basic organization membership
- `moderator` - Team and member management
- `admin` - Full organization administration
- `owner` - Complete organizational control

**Specialized Permissions**
- `actions` - CI/CD workflow management
- `secrets` - Secret and environment variable access
- `packages` - Package registry access
- `security` - Security policy and audit access

### Role Templates

```yaml
# DevOps Engineer Role
name: "DevOps Engineer"
permissions:
  repositories: admin
  actions: write
  secrets: read
  packages: write
  deployments: write

# Senior Developer Role
name: "Senior Developer"
permissions:
  repositories: write
  actions: read
  secrets: none
  packages: read
  code_review: write

# Project Manager Role
name: "Project Manager"
permissions:
  repositories: read
  issues: admin
  projects: admin
  team_management: read
  analytics: read
```

## Organization Policies

### Repository Creation Policies

```bash
# Set repository naming convention policy
curl -X POST /api/v1/organizations/acme/policies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "policy_type": "repository_creation",
    "name": "Naming Convention",
    "enforcement": "block",
    "configuration": {
      "required_prefix": "acme-",
      "allowed_visibility": ["private", "internal"],
      "required_topics": ["project", "team"],
      "forbidden_patterns": ["test-", "temp-"]
    }
  }'
```

### Member Invitation Policies

```bash
# Set domain restriction policy
curl -X POST /api/v1/organizations/acme/policies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "policy_type": "member_invitation",
    "name": "Domain Restriction",
    "enforcement": "warn",
    "configuration": {
      "allowed_domains": ["acme.com", "partner.com"],
      "require_approval": true,
      "max_pending_invitations": 50
    }
  }'
```

### Compliance Policies

```bash
# Set up GDPR compliance policy
curl -X POST /api/v1/organizations/acme/policies \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "policy_type": "compliance",
    "name": "GDPR Compliance",
    "enforcement": "block",
    "configuration": {
      "data_retention_days": 365,
      "require_data_classification": true,
      "anonymize_deleted_users": true,
      "audit_log_retention": 2555
    }
  }'
```

### Policy Enforcement Levels

- **block** - Prevent action if policy violation occurs
- **warn** - Allow action but log warning and notify administrators
- **audit** - Log policy evaluation results for review

## Team Management

### Team Hierarchies

```bash
# Create parent team
curl -X POST /api/v1/organizations/acme/teams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Engineering",
    "description": "All engineering teams",
    "privacy": "closed"
  }'

# Create child team
curl -X POST /api/v1/organizations/acme/teams \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Backend Team",
    "description": "Backend development team",
    "privacy": "closed",
    "parent_team_id": 123
  }'
```

### Team Templates

```yaml
# Development Team Template
name: "Development Team Template"
description: "Standard template for development teams"
default_permissions:
  repositories: write
  actions: read
  secrets: none
default_repositories:
  - "shared-libraries"
  - "documentation"
settings:
  code_review_required: true
  branch_protection_enabled: true
  two_reviewers_required: false
```

### Team Performance Metrics

```bash
# Get team performance metrics
GET /api/v1/organizations/acme/teams/backend/metrics?period=30d

Response:
{
  "commits": 156,
  "pull_requests": 23,
  "code_reviews": 45,
  "issues_resolved": 18,
  "productivity_score": 85,
  "collaboration_index": 92
}
```

## Organization Analytics

### Dashboard Metrics

```bash
# Get organization dashboard data
GET /api/v1/organizations/acme/analytics/dashboard

Response:
{
  "overview": {
    "total_members": 150,
    "active_repositories": 75,
    "monthly_commits": 1250,
    "open_pull_requests": 25
  },
  "growth": {
    "new_members_30d": 12,
    "new_repositories_30d": 8,
    "member_growth_rate": 8.7
  },
  "activity": {
    "most_active_repositories": [...],
    "top_contributors": [...],
    "recent_activities": [...]
  }
}
```

### Repository Analytics

```bash
# Get repository usage statistics
GET /api/v1/organizations/acme/analytics/repositories?period=90d

Response:
{
  "language_distribution": {
    "TypeScript": 45.2,
    "Go": 28.6,
    "Python": 15.1,
    "JavaScript": 11.1
  },
  "size_distribution": {
    "small": 35,
    "medium": 28,
    "large": 12
  },
  "activity_metrics": {
    "most_active": [...],
    "least_active": [...],
    "archived": 5
  }
}
```

### Security and Compliance Analytics

```bash
# Get security score and compliance status
GET /api/v1/organizations/acme/analytics/security

Response:
{
  "security_score": 87,
  "compliance_status": {
    "gdpr": "compliant",
    "soc2": "pending",
    "iso27001": "compliant"
  },
  "security_events": {
    "last_30_days": 12,
    "high_severity": 2,
    "resolved": 10
  },
  "policy_violations": {
    "total": 8,
    "by_type": {
      "repository_naming": 5,
      "member_invitation": 3
    }
  }
}
```

## Advanced Activity Logging

### Enhanced Search and Filtering

```bash
# Search activity with multiple criteria
GET /api/v1/organizations/acme/activities?
  actor=john.doe&
  action=repository.create&
  date_from=2024-01-01&
  date_to=2024-01-31&
  risk_level=high&
  page=1&limit=50
```

### Activity Export

```bash
# Export activity logs as CSV
GET /api/v1/organizations/acme/activities/export?
  format=csv&
  date_from=2024-01-01&
  date_to=2024-12-31

# Export as JSON
GET /api/v1/organizations/acme/activities/export?
  format=json&
  include_details=true
```

### Real-time Activity Notifications

```javascript
// Subscribe to real-time activity updates
const ws = new WebSocket('wss://hub.example.com/api/v1/organizations/acme/activities/stream');

ws.onmessage = (event) => {
  const activity = JSON.parse(event.data);
  if (activity.risk_level === 'high') {
    showSecurityAlert(activity);
  }
};
```

### Audit Summaries

```bash
# Generate compliance audit summary
GET /api/v1/organizations/acme/audit/summary?
  period=quarterly&
  compliance_standard=soc2

Response:
{
  "period": "2024-Q1",
  "total_events": 2547,
  "security_events": 45,
  "policy_violations": 12,
  "compliance_score": 94,
  "recommendations": [
    "Enable 2FA for all members",
    "Review repository access permissions"
  ]
}
```

## Organization Settings

### Security Settings

```yaml
# Organization security configuration
security:
  two_factor_required: true
  ip_restrictions:
    enabled: true
    allowed_ranges:
      - "192.168.1.0/24"
      - "10.0.0.0/8"
  session_timeout: "8h"
  concurrent_sessions: 3
  
sso:
  saml_enabled: true
  ldap_enabled: true
  oauth_providers:
    - github
    - google
    - microsoft
```

### Repository Settings

```yaml
# Default repository settings
repository_defaults:
  visibility: "private"
  auto_init: true
  gitignore_template: "Node"
  license_template: "mit"
  allow_merge_commits: true
  allow_squash_merging: true
  allow_rebase_merging: false
  delete_branch_on_merge: true
```

### Billing and Usage Tracking

```bash
# Get organization usage metrics
GET /api/v1/organizations/acme/usage?period=monthly

Response:
{
  "storage": {
    "repositories_gb": 125.6,
    "artifacts_gb": 45.2,
    "total_gb": 170.8
  },
  "bandwidth": {
    "total_gb": 89.3,
    "git_operations_gb": 56.7,
    "api_calls_gb": 32.6
  },
  "compute": {
    "action_minutes": 1250,
    "runner_hours": 78
  }
}
```

## API Reference

### Organizations

```bash
# Create organization
POST /api/v1/organizations

# Get organization details
GET /api/v1/organizations/acme

# Update organization
PATCH /api/v1/organizations/acme

# Delete organization
DELETE /api/v1/organizations/acme
```

### Roles and Permissions

```bash
# List organization roles
GET /api/v1/organizations/acme/roles

# Create custom role
POST /api/v1/organizations/acme/roles

# Update role
PUT /api/v1/organizations/acme/roles/role-id

# Delete role
DELETE /api/v1/organizations/acme/roles/role-id

# Assign role to member
POST /api/v1/organizations/acme/members/user-id/roles
```

### Policies

```bash
# List organization policies
GET /api/v1/organizations/acme/policies

# Create policy
POST /api/v1/organizations/acme/policies

# Update policy
PUT /api/v1/organizations/acme/policies/policy-id

# Delete policy
DELETE /api/v1/organizations/acme/policies/policy-id

# Get policy violations
GET /api/v1/organizations/acme/policies/violations
```

### Teams

```bash
# List teams
GET /api/v1/organizations/acme/teams

# Create team
POST /api/v1/organizations/acme/teams

# Get team details
GET /api/v1/organizations/acme/teams/team-id

# Add team member
PUT /api/v1/organizations/acme/teams/team-id/members/user-id

# Remove team member
DELETE /api/v1/organizations/acme/teams/team-id/members/user-id
```

## Best Practices

### Role Management
- Create specific roles for different job functions
- Use inheritance to simplify permission management
- Regular audit of role assignments and permissions
- Document role purposes and responsibilities

### Policy Implementation
- Start with warning enforcement before blocking
- Regular review and update of policies
- Clear communication of policy changes to members
- Monitor policy violation trends

### Team Organization
- Use hierarchical teams to reflect organizational structure
- Implement consistent naming conventions
- Regular review of team memberships
- Clear documentation of team responsibilities

### Analytics and Monitoring
- Regular review of organization metrics
- Set up alerts for unusual activity patterns
- Use analytics to identify optimization opportunities
- Track compliance metrics consistently

## Security Considerations

### Access Control
- Principle of least privilege
- Regular access reviews and audits
- Multi-factor authentication enforcement
- IP restriction for sensitive operations

### Data Protection
- Encryption of sensitive configuration data
- Secure audit log storage
- Regular backup of organization data
- Compliance with data protection regulations

### Monitoring
- Real-time security event monitoring
- Automated threat detection
- Regular security assessments
- Incident response procedures

## Migration Guide

### From Basic Organizations
1. Assess current organizational structure
2. Design custom roles and permissions
3. Implement policies gradually
4. Migrate teams to new hierarchy
5. Train administrators on new features

### Data Migration
- Export existing organization data
- Map current permissions to new role system
- Preserve historical activity logs
- Validate data integrity after migration

## Troubleshooting

### Common Issues

**Permission denied errors**
- Verify user role assignments
- Check custom role permissions
- Review policy restrictions
- Validate organizational membership

**Policy violations not enforcing**
- Check policy configuration
- Verify enforcement level settings
- Review policy conditions
- Monitor policy evaluation logs

**Analytics data missing**
- Verify data collection is enabled
- Check background job status
- Review data retention settings
- Validate database connectivity

### Debug and Monitoring

```bash
# Check organization health
GET /api/v1/organizations/acme/health

# Review policy evaluation logs
GET /api/v1/organizations/acme/policies/logs

# Monitor role assignment changes
GET /api/v1/organizations/acme/roles/audit
```

## Support

For organization management issues:
- Review organization settings and policies
- Check user roles and permissions
- Monitor activity logs for issues
- Consult troubleshooting guides
- Contact system administrators

## References

- [Authentication Documentation](authentication.md)
- [API Reference](../api/organizations.md)
- [Security Guide](security.md)
- [Deployment Guide](../DEPLOYMENT.md)