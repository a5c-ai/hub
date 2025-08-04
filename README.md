# Hub - Self-Hosted Git Hosting Service

---

🚀 **Built by [a5c.ai](https://a5c.ai)** — autonomous AI agents that operate like a senior engineering squad.

⏱ **<24 hours start-to-finish.** The one-line repo description was the *only* instruction set — zero additional human guidance.

💡 **≈ 10,000 developer hours replaced.** Agents architected, coded, tested, documented, and provisioned infra — all inside Git.

🤖 **Fully autonomous.** Every commit, from system design to edge-case fixes, was planned and merged by agents.

🔎 **Want more?** Explore other agent-crafted projects in our [GitHub org](https://github.com/a5c-ai) or dive deeper at **[a5c.ai](https://a5c.ai)**.

---

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Hub is a powerful, self-hosted Git hosting service designed to provide enterprise-grade features with complete data sovereignty. Built with modern technologies and cloud-native architecture, Hub offers a comprehensive alternative to hosted Git services while maintaining full control over your code and infrastructure.

## 🚀 Features

### Core Git Operations
- **Complete Git Support**: Full Git protocol implementation with SSH and HTTPS access
- **Repository Management**: Create, fork, clone, and manage repositories with advanced settings
- **Branch Protection**: Sophisticated branch protection rules and merge policies
- **Large File Support**: Git LFS integration with configurable storage backends
- **Repository Templates**: Standardized project initialization with parameterized templates

### Collaboration
- **Pull Requests**: Comprehensive code review workflows with approval requirements
- **Team Management**: Hierarchical organizations with granular permission control
- **Code Review**: Line-by-line commenting, suggestions, and review states



- **Comprehensive Webhooks**: Advanced webhook system with HMAC verification and trigger evaluation
- **Branch Protection Rules**: Complete branch protection system with required status checks, PR reviews, admin enforcement, and pattern matching


### Enterprise Features
- **Advanced Authentication**: Complete multi-factor authentication with TOTP, SMS, WebAuthn/FIDO2, backup codes, and email notifications
- **Enhanced Single Sign-On**: SAML 2.0, OpenID Connect (OIDC) with automatic organization assignment, LDAP, Active Directory, and OAuth integration
- **Secure Session Management**: Refresh token validation, token blacklisting, OAuth state validation, and external team synchronization
- **Advanced Search**: Elasticsearch-powered search across repositories, code, and users with PostgreSQL fallback
- **Comprehensive Analytics**: Real-time analytics for users, organizations, repositories with performance metrics, usage statistics, and data export (JSON, CSV, XLSX)
- **Organization Management**: Custom roles, policy enforcement, team hierarchies, and comprehensive analytics
- **Complete Email System**: SMTP-configurable email service with MFA setup notifications, password reset, and professional templates
- **Audit Logging**: Comprehensive audit trails for compliance and security
- **Plugin System**: Extensible architecture with marketplace and custom plugins
- **High Availability**: Multi-node clustering with automatic failover
- **Backup & Recovery**: Automated backup with point-in-time recovery

### Self-Hosting Excellence
- **Azure-Native**: Purpose-built for Azure with Terraform deployment templates
- **Kubernetes Ready**: Helm charts and operators for container orchestration
- **Docker Support**: Simple Docker Compose deployment for development
- **Progressive Web App**: Mobile-optimized interface with offline support and PWA capabilities
- **Monitoring**: Built-in Prometheus metrics and structured logging
- **Security**: Zero-trust architecture with comprehensive security controls

## 🏗️ Architecture

Hub is built with a modern, cloud-native architecture:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   Backend       │    │   Database      │
│   (Next.js)     │◄───│   (Go/Gin)      │◄───│   (PostgreSQL)  │
│   React + TS    │    │   REST + GraphQL│    │   + Redis Cache │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Load Balancer │    │   Git Storage   │    │   CI/CD Engine  │
│   (NGINX/ALB)   │    │   (Bare Repos)  │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

**Technology Stack:**
- **Backend**: Go 1.21+ with Gin framework
- **Frontend**: Next.js 15 with React 19 and TypeScript
- **Database**: PostgreSQL 14+ with Redis caching
- **Storage**: Local, S3, Azure Blob, or distributed storage
- **Container**: Docker with Kubernetes orchestration
- **Infrastructure**: Terraform for Infrastructure as Code

## 🚀 Quick Start

### Option 1: Docker Compose (Recommended for Development)

```bash
# Clone the repository
git clone https://github.com/a5c-ai/hub.git
cd hub

# Configure environment
cp config.example.yaml config.yaml
# Edit config.yaml with your settings

# Start all services
docker-compose up -d

# Initialize database
docker-compose exec backend ./cmd/migrate/migrate up

# Access Hub
open http://localhost:3000
```

### Option 2: Kubernetes with Helm

```bash
# Add Helm repository
helm repo add hub https://charts.a5c.ai/hub
helm repo update

# Install Hub
helm install hub hub/hub \
  --namespace hub \
  --create-namespace \
  --values values.yaml

# Get external IP
kubectl get ingress -n hub
```

### Option 3: Azure Deployment with Terraform

```bash
# Navigate to Terraform configuration
cd terraform/environments/production

# Configure Azure credentials
az login

# Initialize and deploy
terraform init
terraform plan -var-file="azure.tfvars"
terraform apply
```

You can also automate infrastructure deployment via the GitHub Actions workflow defined in `.github_workflows/infrastructure.yml`. Pushing changes under `terraform/**` on the main branch or manually dispatching this workflow will run `terraform init`, `plan`, and `apply` for the selected environment.

## 📖 Documentation

### User Guides
- **[User Guide](docs/user-guide.md)** - Complete guide for end users
- **[Admin Guide](docs/admin-guide.md)** - Deployment and administration
- **[Developer Guide](docs/developer-guide.md)** - Development and API reference

### Feature Documentation
- **[Authentication System](docs/authentication.md)** - Enterprise authentication, MFA, SSO, session management, and LDAP
- **[Advanced Search](docs/search.md)** - Elasticsearch integration and search capabilities

- **[Analytics System](docs/analytics.md)** - Comprehensive real-time analytics with data export capabilities

- **[Organization Management](docs/organization-management.md)** - Advanced organization features and policies
- **[Mobile & PWA](docs/mobile-pwa.md)** - Progressive Web App and mobile optimization

### Quick References
- **[API Documentation](docs/api/)** - REST and GraphQL APIs
- **[Plugin Development](docs/plugins/)** - Creating custom plugins
- **[Deployment Guide](DEPLOYMENT.md)** - Detailed deployment instructions

### Architecture Documentation
- **[System Architecture](docs/architecture-research.md)** - Technical architecture overview
- **[Database Schema](docs/database.md)** - Database design and schemas
- **[Security Model](docs/security.md)** - Security architecture and practices

## 🛠️ Development

### Prerequisites
- **Go**: 1.21+
- **Node.js**: 18+
- **Docker**: 20.10+
- **PostgreSQL**: 12+
- **Redis**: 6.0+

### Local Development Setup

```bash
# Clone repository
git clone https://github.com/a5c-ai/hub.git
cd hub

# Install dependencies
go mod download
cd frontend && npm install && cd ..

# Start infrastructure
docker-compose up -d postgres redis

# Run database migrations
go run cmd/migrate/main.go up

# Start backend (terminal 1)
./scripts/dev-run.sh backend

# Start frontend (terminal 2)
./scripts/dev-run.sh frontend
```

### Running Tests

```bash
# Backend tests
go test ./...
go test -race -cover ./...

# Frontend tests
cd frontend
npm test
npm run test:ci

# Integration tests
go test -tags=integration ./tests/integration/...

# E2E tests
cd frontend && npm run test:e2e
```

### Building and Deployment

```bash
# Build Docker images
./scripts/build-images.sh

# Deploy to Kubernetes
./scripts/deploy-k8s.sh production

# Run with specific configuration
./scripts/deploy.sh --environment staging --wait
```

## 🔧 Configuration

### Environment Variables

```bash
# Application
APP_ENV=production
LOG_LEVEL=info
DATABASE_URL=postgresql://user:pass@host:5432/hub
REDIS_URL=redis://host:6379/0

# Authentication
JWT_SECRET=your-jwt-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Multi-Factor Authentication
ENABLE_MFA=true
TOTP_ISSUER=Hub
SMS_PROVIDER=twilio  # twilio, aws_sns
SMS_API_KEY=your-sms-api-key

# SSO Configuration
ENABLE_SAML=true
SAML_CERTIFICATE_PATH=/certs/saml.crt
SAML_PRIVATE_KEY_PATH=/certs/saml.key
ENABLE_LDAP=true
LDAP_URL=ldap://ldap.company.com
LDAP_BIND_DN=cn=admin,dc=company,dc=com

# Storage
GIT_DATA_PATH=/repositories
STORAGE_BACKEND=local  # local, s3, azure_blob

# Features
ENABLE_REGISTRATION=true
ENABLE_ORGANIZATIONS=true


# Redis Configuration (Job Queue)
REDIS_ENABLED=true
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10
```

### Configuration File (config.yaml)

```yaml
app:
  environment: production
  log_level: info
  domain: hub.yourdomain.com

database:
  host: postgresql
  port: 5432
  name: hub
  user: hub
  ssl_mode: require

auth:
  jwt_expiry: 24h
  password_policy:
    min_length: 12
    require_symbols: true
  mfa:
    enabled: true
    totp_issuer: "Hub"
    sms_provider: "twilio"
  saml:
    enabled: true
    entity_id: "hub.yourdomain.com"
    certificate_path: "/certs/saml.crt"
    private_key_path: "/certs/saml.key"
  ldap:
    enabled: true
    url: "ldap://ldap.company.com"
    bind_dn: "cn=admin,dc=company,dc=com"

oauth:
  github:
    enabled: true
    client_id: ${GITHUB_CLIENT_ID}
    client_secret: ${GITHUB_CLIENT_SECRET}
  google:
    enabled: true
    client_id: ${GOOGLE_CLIENT_ID}
    client_secret: ${GOOGLE_CLIENT_SECRET}
  microsoft:
    enabled: true
    client_id: ${MICROSOFT_CLIENT_ID}
    client_secret: ${MICROSOFT_CLIENT_SECRET}

redis:
  enabled: false          # Enable/disable Redis job queue (defaults to database fallback)
  host: localhost         # Redis server host
  port: 6379             # Redis server port  
  password: ""           # Redis authentication password
  db: 0                  # Redis database number
  max_retries: 3         # Connection retry attempts
  pool_size: 10          # Connection pool size

storage:
  backend: azure_blob
  azure:
    account_name: hubstorage
    container_name: repositories
  artifacts:
    backend: "filesystem"  # Options: filesystem, azure, s3
    base_path: "/var/lib/hub/artifacts"
    max_size_mb: 1024
    retention_days: 90
    azure:
      account_name: ""
      account_key: ""
      container_name: "artifacts"
    s3:
      region: ""
      bucket: ""
      access_key_id: ""
      secret_access_key: ""
      use_ssl: true

application:
  base_url: "https://your-domain.com"
  name: "Hub"

smtp:
  host: "smtp.gmail.com"
  port: "587"
  username: "your-email@gmail.com"
  password: "your-app-password"
  from: "noreply@your-domain.com"
  use_tls: true
```

## 🔐 Security

Hub implements comprehensive security measures:

- **Zero Trust Architecture**: Assume breach, verify everything
- **Encryption**: AES-256 at rest, TLS 1.3 in transit
- **Authentication**: Multi-factor authentication with SSO integration
- **Authorization**: Role-based and attribute-based access control
- **Compliance**: SOC 2, ISO 27001, GDPR, HIPAA ready
- **Audit Logging**: Comprehensive audit trails with tamper protection

### Security Best Practices

```bash
# Enable 2FA for all users
hub admin config --require-2fa

# Set up SSL certificates
kubectl apply -f k8s/tls-certificates.yaml

# Configure network policies
kubectl apply -f k8s/network-policies.yaml

# Regular security scans
docker scan hub/backend:latest
```

## 🔍 Advanced Search

Hub includes a powerful search system with Elasticsearch integration and PostgreSQL fallback:

### Search Features
- **Global Search**: Search across all content types (users, repositories, commits, organizations)
- **Code Search**: Search within repository files with syntax highlighting
- **Advanced Filters**: Filter by language, visibility, state, labels, dates, and more
- **Fuzzy Matching**: Find content even with typos or partial matches
- **Real-time Results**: Fast search with pagination and result highlighting

### Configuration
```yaml
elasticsearch:
  enabled: true
  addresses: ["http://localhost:9200"]
  username: ""
  password: ""
  cloud_id: ""
  api_key: ""
  index_prefix: "hub"
```

### Usage Examples
```bash
# Global search
curl "/api/v1/search?q=authentication&type=repositories"

# Code search (requires Elasticsearch)
curl "/api/v1/search/code?q=func main&language=go"

# Repository search with filters
curl "/api/v1/search/repositories?q=api&language=typescript&visibility=public"
```

## 📊 Analytics and Reporting

Hub provides comprehensive analytics for users, organizations, and repositories:

### Analytics Features
- **User Analytics**: Repository statistics, contribution metrics, and activity tracking
- **Organization Analytics**: Member statistics, repository metrics, team analytics, and security insights
- **Performance Metrics**: Build times, success rates, and resource usage with percentile calculations
- **Data Export**: Export analytics data in JSON, CSV, and XLSX formats
- **Real-time Data**: Live analytics with database-driven queries and statistical processing

### API Endpoints
```bash
# User analytics
curl "/api/v1/users/{username}/analytics"

# Organization analytics  
curl "/api/v1/orgs/{org}/analytics"

# Repository statistics
curl "/api/v1/repos/{owner}/{repo}/analytics"

# Export data (requires authentication)
curl -H "Authorization: Bearer {token}" "/api/v1/analytics/export?format=csv"
```



## 🛡️ Branch Protection

Complete branch protection system with advanced rules and enforcement:

### Protection Features
- **Required Status Checks**: Enforce CI/CD checks before merging with strict mode support
- **Pull Request Reviews**: Configurable approval requirements and code owner reviews
- **Admin Enforcement**: Optional enforcement of rules on repository administrators
- **Pattern Matching**: Support for exact matches and wildcard patterns (`main`, `feature/*`, `*`)
- **Granular Controls**: Individual management of status checks, PR reviews, and restrictions

### API Endpoints
```bash
# Get branch protection rules
curl "/api/v1/repos/{owner}/{repo}/branches/{branch}/protection"

# Create/update protection rules
curl -X PUT "/api/v1/repos/{owner}/{repo}/branches/{branch}/protection" \
  -H "Content-Type: application/json" -d '{
    "required_status_checks": {
      "strict": true,
      "contexts": ["ci/tests", "ci/build"]
    },
    "required_pull_request_reviews": {
      "required_approving_review_count": 2,
      "dismiss_stale_reviews": true,
      "require_code_owner_reviews": true
    },
    "enforce_admins": true
  }'

# Delete protection rules
curl -X DELETE "/api/v1/repos/{owner}/{repo}/branches/{branch}/protection"
```

## 📊 Monitoring and Observability

### Metrics and Monitoring

Hub provides comprehensive observability:

```yaml
# Prometheus metrics
- hub_http_requests_total
- hub_git_operations_total
- hub_active_users_gauge
- hub_repository_count_gauge
- hub_build_duration_histogram

# Health checks
- /health (application health)
- /ready (readiness probe)
- /metrics (Prometheus metrics)
```

### Logging

```bash
# View application logs
kubectl logs -f deployment/hub-backend -n hub

# Stream logs with structured format
docker-compose logs -f backend | jq .

# Access audit logs
hub admin audit-log --from="2024-01-01" --to="2024-01-07"
```

## 🚀 Deployment Options

### 1. Development (Docker Compose)
Perfect for local development and small teams:
```bash
docker-compose up -d
```

### 2. Production (Kubernetes)
Scalable production deployment:
```bash
helm install hub hub/hub --values production-values.yaml
```

### 3. Azure Cloud (Terraform)
Fully managed Azure deployment:
```bash
cd terraform/environments/production
terraform apply
```

### 4. Hybrid Cloud
Multi-cloud deployment with Azure primary:
```bash
./scripts/deploy-hybrid.sh --primary=azure --secondary=aws
```

## 🔄 Migration

### From GitHub Enterprise

```bash
# Export from GitHub
hub migrate export --source github --token $GITHUB_TOKEN --org source-org

# Import to Hub
hub migrate import --archive export.tar.gz --org hub-org
```

### From GitLab

```bash
# Migrate GitLab projects
hub migrate gitlab --url https://gitlab.company.com --token $GITLAB_TOKEN --group source-group
```

### From Bitbucket

```bash
# Import Bitbucket repositories
hub migrate bitbucket --workspace myworkspace --token $BITBUCKET_TOKEN
```

## 🤝 Contributing

We welcome contributions from the community! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Process

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Commit** your changes (`git commit -m 'Add amazing feature'`)
4. **Push** to the branch (`git push origin feature/amazing-feature`)
5. **Open** a Pull Request

### Code Style

- **Go**: Follow official Go style guide, use `gofmt` and `golangci-lint`
- **TypeScript**: Use strict mode, follow React best practices
- **Tests**: Maintain >80% code coverage
- **Documentation**: Update docs for new features

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

### Community Support
- **GitHub Repository**: [Visit our repository](https://github.com/a5c-ai/hub)
- **GitHub Discussions**: [Community discussions and Q&A](https://github.com/a5c-ai/hub/discussions)
- **Documentation**: [Comprehensive guides and API docs](https://docs.hub.a5c.ai)

### Enterprise Support
- **Professional Services**: Migration assistance and custom development
- **Enterprise Support**: 24/7 support with SLA guarantees
- **Training**: On-site training and workshops

Contact: [enterprise@a5c.ai](mailto:enterprise@a5c.ai)

## 🎯 Roadmap

### Recently Completed ✅
- [x] **Advanced Search System**: Elasticsearch integration with code search and PostgreSQL fallback
- [x] **Enterprise Authentication**: Complete MFA, SAML, LDAP, and OAuth integration with enhanced session management


- [x] **Comprehensive Analytics Backend**: Real-time user, organization, and repository analytics with performance metrics and data export

- [x] **Branch Protection Rules**: Full branch protection implementation with status checks, PR reviews, and pattern matching
- [x] **Complete Email Service**: SMTP-configurable email system with MFA notifications and professional templates
- [x] **Enhanced Authentication Backend**: Refresh token management, token blacklisting, OIDC organization assignment, and external team sync
- [x] **Mobile-Responsive PWA**: Progressive Web App with offline support and mobile optimization
- [x] **Advanced Organization Features**: Custom roles, policy enforcement, and team management

### Version 1.1 (Q2 2025)
- [ ] Container registry integration
- [ ] AI-powered code review assistance

- [ ] Package registry integration

### Version 1.2 (Q3 2025)
- [ ] Multi-region replication and distributed storage
- [ ] GraphQL subscriptions for real-time updates
- [ ] Advanced security scanning and vulnerability management
- [ ] Machine learning insights and recommendations

### Version 2.0 (Q4 2025)
- [ ] Distributed architecture with microservices
- [ ] Advanced compliance and governance features
- [ ] Enhanced AI capabilities for code analysis
- [ ] Cross-platform mobile applications

## 📈 Metrics and Stats

![GitHub stars](https://img.shields.io/github/stars/a5c-ai/hub?style=social)
![GitHub forks](https://img.shields.io/github/forks/a5c-ai/hub?style=social)
![GitHub pull requests](https://img.shields.io/github/issues-pr/a5c-ai/hub)

### Community
- **Contributors**: 50+ active contributors
- **Organizations**: 500+ organizations using Hub
- **Repositories**: 100,000+ repositories hosted
- **Deployments**: Available in 20+ countries

## 🙏 Acknowledgments

- **Contributors**: Thanks to all our amazing contributors
- **Community**: Grateful for feedback and support from the community
- **Open Source**: Built on top of fantastic open source projects
- **Inspiration**: Inspired by GitHub, GitLab, and Gitea

---

<div align="center">

**[Website](https://hub.a5c.ai)** • 
**[Documentation](https://docs.hub.a5c.ai)** • 
**[API Reference](https://api.hub.a5c.ai)** • 
**[Community](https://community.hub.a5c.ai)**

Made with ❤️ by the [A5C team](https://a5c.ai) and [contributors](https://github.com/a5c-ai/hub/graphs/contributors).

</div>
