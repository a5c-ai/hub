# Hub Database Documentation

## Overview

The Hub git hosting service uses PostgreSQL 15+ as its primary database with a comprehensive schema designed to support all core features including user management, organizations, repositories, issues, pull requests, and collaboration features.

## Architecture

### Database Design Principles

- **ACID Compliance**: All transactions are atomic, consistent, isolated, and durable
- **UUID Primary Keys**: All tables use UUID primary keys for security and distributed scaling
- **Soft Deletes**: Most entities support soft deletion through `deleted_at` timestamps
- **Comprehensive Indexing**: Optimized indexes for query performance
- **Foreign Key Constraints**: Referential integrity enforced at the database level
- **JSON Support**: PostgreSQL JSON columns for flexible metadata storage

### Connection Management

- **Connection Pooling**: Configured with optimal pool settings
  - Max Idle Connections: 10
  - Max Open Connections: 100
  - Connection Max Lifetime: 1 hour
  - Connection Max Idle Time: 10 minutes

## Schema Overview

### Core Entities

#### 1. User Management

**users**
- Primary user accounts with authentication and profile information
- Supports email verification and two-factor authentication
- Soft deletion with audit trail

**ssh_keys**
- SSH public keys associated with user accounts
- Unique fingerprint validation
- Last used tracking for security auditing

#### 2. Organization Management

**organizations**
- Multi-user organizations with billing and settings
- Support for display names, avatars, and contact information
- Hierarchical permission model

**organization_members**
- User membership in organizations
- Role-based permissions (owner, admin, member, billing)
- Unique constraint prevents duplicate memberships

**teams**
- Sub-groups within organizations
- Privacy controls (closed, secret)
- Team-based repository access

**team_members**
- User membership in teams
- Role hierarchy (maintainer, member)

#### 3. Repository Management

**repositories**
- Git repositories with comprehensive metadata
- Multi-owner support (users and organizations)
- Visibility controls (public, private, internal)
- Fork relationships and template support
- Feature flags (issues, projects, wiki, downloads)
- Merge strategy configuration
- Statistics tracking (stars, forks, watchers)

**repository_collaborators**
- Individual repository access permissions
- Granular permission levels (read, triage, write, maintain, admin)

#### 4. Git Operations

**branches**
- Branch metadata and protection status
- Default branch designation
- SHA tracking for git operations

**branch_protection_rules**
- Configurable branch protection policies
- JSON-based rule storage for flexibility
- Pattern matching for branch names

**releases**
- Tagged releases and version management
- Draft and prerelease support
- Rich release notes with markdown

#### 5. Issue Tracking

**issues**
- Issue tracking with comprehensive metadata
- Sequential numbering per repository
- State management (open, closed) with reasons
- Assignment and milestone support
- Comment count tracking

**comments**
- Comments on issues and pull requests
- Markdown content support
- User attribution and timestamps

**pull_requests**
- Code review and merge requests
- Relationship to issues for unified tracking
- Merge statistics and status
- Draft support and merge conflict detection

**milestones**
- Project milestone management
- Due date tracking and completion status

**labels**
- Customizable issue and PR labeling
- Color-coded organization
- Many-to-many relationship with issues

## Migration System

### Structure

The database uses a custom migration system built on GORM with the following features:

- **Sequential Migrations**: Numbered migration files ensure proper ordering
- **Up/Down Support**: Full rollback capability for all migrations
- **Migration Tracking**: Dedicated migrations table tracks applied changes
- **Atomic Migrations**: Each migration runs in a transaction

### Migration Files

```
internal/db/migrations/
├── migrations.go          # Migration engine
├── registry.go           # Migration registration
├── 001_initial_schema.go # Core table creation
└── 002_indexes_and_constraints.go # Performance optimization
```

### Running Migrations

```bash
# Run all pending migrations
go run cmd/migrate/main.go

# Rollback last migration
go run cmd/migrate/main.go -rollback

# Run migrations and seed database
go run cmd/migrate/main.go -seed
```

## Performance Optimization

### Indexing Strategy

The database includes comprehensive indexing for optimal query performance:

#### Primary Indexes
- All UUID primary keys are automatically indexed
- Unique constraints on natural keys (usernames, emails, etc.)

#### Query Optimization Indexes
- Foreign key relationships
- Frequently filtered columns (state, visibility, dates)
- Composite indexes for complex queries
- Partial indexes for specific use cases

#### Examples
```sql
-- Repository queries by owner
CREATE INDEX idx_repositories_owner ON repositories(owner_id, owner_type);

-- Issue filtering by state and repository
CREATE INDEX idx_issues_repo_id ON issues(repository_id);
CREATE INDEX idx_issues_state ON issues(state);

-- Time-based queries
CREATE INDEX idx_issues_created_at ON issues(created_at);
CREATE INDEX idx_repositories_pushed_at ON repositories(pushed_at);
```

### Query Patterns

#### Efficient Queries
```sql
-- Repository listing for user
SELECT * FROM repositories 
WHERE owner_id = $1 AND owner_type = 'user' 
ORDER BY updated_at DESC;

-- Issues for repository with pagination
SELECT * FROM issues 
WHERE repository_id = $1 AND state = 'open'
ORDER BY created_at DESC 
LIMIT 25 OFFSET $2;
```

## Backup and Recovery

### Automated Backups

The system includes comprehensive backup scripts:

**scripts/backup-db.sh**
- Automated PostgreSQL dumps
- Compression and retention management
- Cloud storage integration support
- Configurable retention policies

**scripts/restore-db.sh**
- Safe database restoration
- Connection termination handling
- Verification prompts

### Backup Strategy

```bash
# Environment configuration
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=hub
export DB_USER=hub
export BACKUP_DIR=./backups
export RETENTION_DAYS=7

# Create backup
./scripts/backup-db.sh

# Restore from backup
./scripts/restore-db.sh ./backups/hub_backup_20240124_143000.sql.gz
```

## Development Setup

### Initial Setup

1. **Install PostgreSQL 15+**
   ```bash
   # Ubuntu/Debian
   sudo apt install postgresql-15 postgresql-client-15
   
   # macOS
   brew install postgresql@15
   ```

2. **Create Database**
   ```bash
   sudo -u postgres createuser hub
   sudo -u postgres createdb hub -O hub
   sudo -u postgres psql -c "ALTER USER hub PASSWORD 'password';"
   ```

3. **Configure Environment**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_NAME=hub
   export DB_USER=hub
   export DB_PASSWORD=password
   ```

4. **Run Migrations**
   ```bash
   go run cmd/migrate/main.go -seed
   ```

### Development Data

The seed system creates:
- Admin user (admin/admin123)
- Test user (testuser/test123)
- Test organization with team structure
- Sample repository with branches and protection rules
- Example issues, labels, and comments

## Security Considerations

### Access Control
- Row-level security can be implemented for multi-tenant scenarios
- All foreign key relationships enforce referential integrity
- Soft deletes preserve audit trails

### Data Protection
- Password hashes use bcrypt with appropriate cost factors
- SSH key fingerprints prevent duplicate key storage
- UUID primary keys prevent enumeration attacks

### Audit Trail
- Created/updated timestamps on all entities
- Soft delete support maintains historical data
- Comment and activity logging for security analysis

## Monitoring and Maintenance

### Database Health Checks
```sql
-- Check connection count
SELECT count(*) FROM pg_stat_activity;

-- Monitor slow queries
SELECT query, mean_time, calls 
FROM pg_stat_statements 
ORDER BY mean_time DESC 
LIMIT 10;

-- Check index usage
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

### Maintenance Tasks
- Regular VACUUM and ANALYZE operations
- Index maintenance and rebuild as needed
- Statistics updates for query optimization
- Backup verification and retention management

## Future Considerations

### Scalability
- Connection pooling with PgBouncer for high load
- Read replicas for query distribution
- Partitioning strategies for large datasets

### Features
- Full-text search integration with PostgreSQL
- JSON-based configuration storage
- Event sourcing for audit requirements
- Integration with external authentication systems

This documentation should be updated as the schema evolves and new features are added.