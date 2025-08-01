# Administrator Guide - Hub Git Hosting Service

This guide provides comprehensive information for administrators deploying, configuring, and managing Hub Git Hosting Service instances. Hub is designed for self-hosting with enterprise-grade features and complete data sovereignty.

## Table of Contents

- [System Requirements](#system-requirements)
- [Installation and Deployment](#installation-and-deployment)
- [Configuration Management](#configuration-management)
- [Security Configuration](#security-configuration)
- [User and Organization Management](#user-and-organization-management)
- [Storage and Backup](#storage-and-backup)
- [Monitoring and Maintenance](#monitoring-and-maintenance)
- [Scaling and Performance](#scaling-and-performance)
- [Troubleshooting](#troubleshooting)
- [Upgrade and Migration](#upgrade-and-migration)

## System Requirements

### Minimum Requirements (Small Teams < 100 users)

#### Hardware
- **CPU**: 4 vCPUs (x86_64 or ARM64)
- **Memory**: 8 GB RAM
- **Storage**: 100 GB SSD (system + initial data)
- **Network**: 1 Gbps network interface

#### Software
- **Operating System**: Linux (Ubuntu 20.04+, RHEL 8+, or equivalent)
- **Container Runtime**: Docker 20.10+ or containerd 1.6+
- **Orchestration**: Kubernetes 1.24+ (recommended) or Docker Compose
- **Database**: PostgreSQL 12+ (managed or self-hosted)
- **Cache**: Redis 6.0+ (managed or self-hosted)

### Enterprise Requirements (1000+ users)

#### Hardware
- **CPU**: 16+ vCPUs per node
- **Memory**: 64+ GB RAM per node
- **Storage**: 1+ TB NVMe SSD with high IOPS
- **Network**: 10+ Gbps network interface

#### Infrastructure
- **Load Balancer**: NGINX, HAProxy, or cloud load balancer
- **Database**: PostgreSQL cluster with read replicas
- **Cache**: Redis cluster with high availability
- **Storage**: Distributed storage (NFS, Azure Files, AWS EFS)
- **Monitoring**: Prometheus, Grafana, and log aggregation

### Azure-Specific Requirements

#### Azure Services
- **Compute**: Azure Kubernetes Service (AKS) or Virtual Machines
- **Database**: Azure Database for PostgreSQL
- **Storage**: Azure Blob Storage or Azure Files
- **Networking**: Virtual Network with subnets and security groups
- **Security**: Azure Key Vault for secrets management
- **Monitoring**: Azure Monitor and Log Analytics

#### Resource Recommendations
```yaml
# Azure VM SKUs
Small Deployment: Standard_D4s_v3 (4 vCPU, 16 GB RAM)
Medium Deployment: Standard_D8s_v3 (8 vCPU, 32 GB RAM)
Large Deployment: Standard_D16s_v3 (16 vCPU, 64 GB RAM)

# AKS Node Pool
Node Count: 3-10 nodes
Node Size: Standard_D4s_v3 or larger
Auto-scaling: Enabled (3-50 nodes)
```

## Installation and Deployment

### Docker Deployment

#### Quick Start with Docker Compose
1. **Clone the repository:**
   ```bash
   git clone https://github.com/a5c-ai/hub.git
   cd hub
   ```

2. **Configure environment:**
   ```bash
   cp config.example.yaml config.yaml
   # Edit config.yaml with your settings
   ```

3. **Start services:**
   ```bash
   docker-compose up -d
   ```

4. **Initialize database:**
   ```bash
   docker-compose exec backend ./cmd/migrate/migrate
   ```

#### Production Docker Setup
```yaml
# docker-compose.prod.yml
version: '3.8'
services:
  backend:
    image: hub/backend:latest
    environment:
      - DATABASE_URL=postgresql://user:pass@db:5432/hub
      - REDIS_URL=redis://redis:6379
      - JWT_SECRET=${JWT_SECRET}
    volumes:
      - ./repositories:/repositories
      - ./config.yaml:/config.yaml
    depends_on:
      - db
      - redis

  frontend:
    image: hub/frontend:latest
    environment:
      - NEXT_PUBLIC_API_URL=https://hub.yourdomain.com/api
    ports:
      - "3000:3000"

  db:
    image: postgres:14
    environment:
      - POSTGRES_DB=hub
      - POSTGRES_USER=hub
      - POSTGRES_PASSWORD=${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

### Kubernetes Deployment

#### Prerequisites
```bash
# Install required tools
kubectl version --client
helm version

# Verify cluster access
kubectl cluster-info
kubectl get nodes
```

#### Using Terraform (Recommended)
1. **Navigate to Terraform directory:**
   ```bash
   cd terraform/environments/production
   ```

2. **Configure variables:**
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your settings
   ```

3. **Deploy infrastructure:**
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

4. **Deploy application:**
   ```bash
   cd ../../../
   ./scripts/deploy-k8s.sh production --wait
   ```

#### Manual Kubernetes Deployment
```bash
# Create namespace
kubectl create namespace hub

# Apply configurations
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml  # Update secrets first!
kubectl apply -f k8s/storage.yaml
kubectl apply -f k8s/postgresql-deployment.yaml
kubectl apply -f k8s/redis-deployment.yaml
kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml
kubectl apply -f k8s/services.yaml
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/network-policy.yaml

# Verify deployment
kubectl get pods -n hub
kubectl get services -n hub
```

### Azure-Specific Deployment

#### Using Azure CLI and Terraform
```bash
# Login to Azure
az login
az account set --subscription "your-subscription-id"

# Deploy with Terraform
cd terraform/environments/production
terraform init \
  -backend-config="resource_group_name=hub-terraform-state" \
  -backend-config="storage_account_name=tfstateXXXXX" \
  -backend-config="container_name=tfstate" \
  -backend-config="key=production.terraform.tfstate"
terraform apply -var-file="azure.tfvars"
```

#### AKS Deployment
```bash
# Get AKS credentials
az aks get-credentials --resource-group hub-rg --name hub-aks

# Deploy with Helm
helm upgrade --install hub ./helm/hub \
  --namespace hub --create-namespace \
  --values values-azure.yaml \
  --wait
```

## Configuration Management

### Core Configuration

#### Main Configuration File (config.yaml)
```yaml
# Application settings
app:
  environment: production
  log_level: info
  port: 8080
  domain: hub.yourdomain.com

# Database configuration
database:
  host: postgresql.hub.svc.cluster.local
  port: 5432
  name: hub
  user: hub
  password: ${DATABASE_PASSWORD}
  ssl_mode: require
  max_connections: 100

# Redis configuration  
redis:
  host: redis.hub.svc.cluster.local
  port: 6379
  password: ${REDIS_PASSWORD}
  database: 0

# Authentication
auth:
  jwt_secret: ${JWT_SECRET}
  jwt_expiry: 24h
  session_timeout: 7d
  password_policy:
    min_length: 12
    require_uppercase: true
    require_lowercase: true
    require_numbers: true
    require_symbols: true

# OAuth providers
oauth:
  github:
    client_id: ${GITHUB_CLIENT_ID}
    client_secret: ${GITHUB_CLIENT_SECRET}
    enabled: true
  
  azure_ad:
    client_id: ${AZURE_AD_CLIENT_ID}
    client_secret: ${AZURE_AD_CLIENT_SECRET}
    tenant_id: ${AZURE_AD_TENANT_ID}
    enabled: false

# Storage configuration
storage:
  backend: local  # local, s3, azure_blob
  path: /repositories
  max_repo_size: 10GB
  cleanup_policy:
    enabled: true
    retention_days: 90

# Email settings
email:
  smtp_host: smtp.yourdomain.com
  smtp_port: 587
  smtp_user: noreply@yourdomain.com
  smtp_password: ${SMTP_PASSWORD}
  from_address: Hub <noreply@yourdomain.com>
  tls: true

# Monitoring and logging
monitoring:
  metrics_enabled: true
  prometheus_endpoint: /metrics
  health_check_endpoint: /health
  log_format: json
  log_file: /var/log/hub/hub.log
```

### Environment Variables

#### Backend Environment Variables
```bash
# Core application
APP_ENV=production
LOG_LEVEL=info
GIN_MODE=release
PORT=8080

# Database
DATABASE_URL=postgresql://user:pass@host:5432/hub
DATABASE_MAX_CONNECTIONS=100
DATABASE_SSL_MODE=require

# Redis
REDIS_URL=redis://password@host:6379/0
REDIS_MAX_CONNECTIONS=50

# Authentication
JWT_SECRET=your-jwt-secret-key
JWT_EXPIRY=24h
SESSION_TIMEOUT=168h

# OAuth
GITHUB_OAUTH_CLIENT_ID=your-github-client-id
GITHUB_OAUTH_CLIENT_SECRET=your-github-client-secret
AZURE_AD_CLIENT_ID=your-azure-ad-client-id
AZURE_AD_CLIENT_SECRET=your-azure-ad-client-secret
AZURE_AD_TENANT_ID=your-azure-ad-tenant-id

# Storage
GIT_DATA_PATH=/repositories
STORAGE_BACKEND=local
STORAGE_MAX_REPO_SIZE=10737418240

# Email
SMTP_HOST=smtp.yourdomain.com
SMTP_PORT=587
SMTP_USER=noreply@yourdomain.com
SMTP_PASSWORD=your-smtp-password
SMTP_FROM=Hub <noreply@yourdomain.com>
SMTP_TLS=true
```

#### Frontend Environment Variables
```bash
# Application
NODE_ENV=production
PORT=3000

# API endpoints
NEXT_PUBLIC_API_URL=https://hub.yourdomain.com/api
NEXT_PUBLIC_APP_URL=https://hub.yourdomain.com
NEXT_PUBLIC_WS_URL=wss://hub.yourdomain.com/ws

# Features
NEXT_PUBLIC_ENABLE_REGISTRATION=true
NEXT_PUBLIC_ENABLE_GITHUB_OAUTH=true
NEXT_PUBLIC_ENABLE_AZURE_AD=false
```

### Kubernetes Configuration

#### ConfigMap Example
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: hub-config
  namespace: hub
data:
  config.yaml: |
    app:
      environment: production
      log_level: info
      port: 8080
      domain: hub.yourdomain.com
    database:
      host: postgresql
      port: "5432"
      name: hub
      user: hub
      ssl_mode: require
      max_connections: 100
    redis:
      host: redis
      port: "6379"
      database: "0"
    # ... additional configuration
```

#### Secrets Management
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: hub-secrets
  namespace: hub
type: Opaque
data:
  # Base64 encoded values
  database-password: <base64-encoded-password>
  jwt-secret: <base64-encoded-secret>
  github-client-secret: <base64-encoded-secret>
  smtp-password: <base64-encoded-password>
```

## Security Configuration

### TLS/SSL Configuration

#### Certificate Management
```yaml
# Using cert-manager with Let's Encrypt
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: letsencrypt-prod
  namespace: hub
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@yourdomain.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

#### Ingress TLS Configuration
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: hub-ingress
  namespace: hub
  annotations:
    cert-manager.io/issuer: letsencrypt-prod
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/force-ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - hub.yourdomain.com
    secretName: hub-tls
  rules:
  - host: hub.yourdomain.com
    http:
      paths:
      - path: /api
        pathType: Prefix
        backend:
          service:
            name: hub-backend-service
            port:
              number: 8080
      - path: /
        pathType: Prefix
        backend:
          service:
            name: hub-frontend-service
            port:
              number: 3000
```

### Authentication Configuration

#### LDAP/Active Directory Integration
```yaml
# In config.yaml
auth:
  ldap:
    enabled: true
    host: ldap.yourdomain.com
    port: 636
    use_ssl: true
    bind_dn: cn=hub-service,ou=services,dc=yourdomain,dc=com
    bind_password: ${LDAP_BIND_PASSWORD}
    base_dn: ou=users,dc=yourdomain,dc=com
    user_filter: (uid=%s)
    attributes:
      username: uid
      email: mail
      first_name: givenName
      last_name: sn
    group_filter: (member=%s)
    group_base_dn: ou=groups,dc=yourdomain,dc=com
```

#### SAML Configuration
```yaml
auth:
  saml:
    enabled: true
    entity_id: https://hub.yourdomain.com
    sso_url: https://your-idp.com/sso
    slo_url: https://your-idp.com/slo
    certificate: ${SAML_CERTIFICATE}
    private_key: ${SAML_PRIVATE_KEY}
    attributes:
      username: NameID
      email: email
      first_name: firstName
      last_name: lastName
      groups: groups
```

### Network Security

#### Network Policies
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: hub-network-policy
  namespace: hub
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 3000
  - from:
    - podSelector:
        matchLabels:
          app: hub-backend
    - podSelector:
        matchLabels:
          app: hub-frontend
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 5432  # PostgreSQL
    - protocol: TCP
      port: 6379  # Redis
    - protocol: TCP
      port: 53    # DNS
    - protocol: UDP
      port: 53    # DNS
    - protocol: TCP
      port: 443   # HTTPS
    - protocol: TCP
      port: 80    # HTTP
```

### Security Policies

#### Pod Security Standards
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: hub
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/audit: restricted
    pod-security.kubernetes.io/warn: restricted
```

#### Security Context
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hub-backend
spec:
  template:
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
        runAsGroup: 1000
        fsGroup: 1000
        seccompProfile:
          type: RuntimeDefault
      containers:
      - name: backend
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
            - ALL
        # ... rest of container spec
```

## User and Organization Management

### Initial Setup

#### Create Administrator Account
```bash
# Using the backend CLI
kubectl exec -it deployment/hub-backend -n hub -- \
  ./cmd/server/server admin create-user \
  --username admin \
  --email admin@yourdomain.com \
  --password "secure-password" \
  --role admin
```

#### Database Seeding
```bash
# Run database seeds
kubectl exec -it deployment/hub-backend -n hub -- \
  ./cmd/migrate/migrate seed
```

### User Management

#### User Creation Methods
1. **Self-registration** (if enabled)
2. **Admin invitation**
3. **LDAP/AD synchronization**
4. **API-based creation**
5. **Bulk import from CSV**

#### User Roles and Permissions
```yaml
# Role hierarchy
roles:
  site_admin:
    - manage_users
    - manage_organizations
    - manage_system_settings
    - access_audit_logs
  
  organization_owner:
    - manage_organization
    - manage_teams
    - manage_billing
    - view_organization_audit_log
  
  organization_member:
    - create_repositories
    - join_teams
    - view_organization
  
  repository_admin:
    - manage_repository_settings
    - manage_collaborators
    - delete_repository
  
  repository_maintainer:
    - manage_issues_and_prs
    - manage_repository_content
    - manage_releases
```

### Organization Management

#### Creating Organizations
```bash
# Via API
curl -X POST https://hub.yourdomain.com/api/organizations \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-organization",
    "display_name": "My Organization",
    "description": "Organization description",
    "website": "https://myorg.com",
    "location": "San Francisco, CA"
  }'
```

#### Team Management
```yaml
# Team structure example
organization: my-organization
teams:
  - name: developers
    description: Development team
    permissions: write
    privacy: closed
    members:
      - alice
      - bob
    repositories:
      - backend
      - frontend
  
  - name: devops
    description: DevOps and infrastructure
    permissions: admin
    privacy: secret
    members:
      - charlie
      - dave
    repositories:
      - infrastructure
      - monitoring
```

## Storage and Backup

### Storage Configuration

#### Local Storage
```yaml
# For single-node deployments
storage:
  backend: local
  path: /var/lib/hub/repositories
  backup_path: /var/lib/hub/backups
```

#### Network Storage (NFS)
```yaml
# Kubernetes PV with NFS
apiVersion: v1
kind: PersistentVolume
metadata:
  name: hub-repositories-pv
spec:
  capacity:
    storage: 1Ti
  accessModes:
    - ReadWriteMany
  nfs:
    server: nfs.yourdomain.com
    path: /hub/repositories
  mountOptions:
    - nfsvers=4.1
    - hard
    - intr
```

#### Azure Blob Storage
```yaml
# In config.yaml
storage:
  backend: azure_blob
  azure:
    account_name: hubstorage
    account_key: ${AZURE_STORAGE_KEY}
    container_name: repositories
    endpoint: https://hubstorage.blob.core.windows.net
```

#### AWS S3 Storage
```yaml
storage:
  backend: s3
  s3:
    bucket: hub-repositories
    region: us-west-2
    access_key: ${AWS_ACCESS_KEY}
    secret_key: ${AWS_SECRET_KEY}
    endpoint: https://s3.us-west-2.amazonaws.com
```

### Backup and Recovery

#### Automated Backup Script
```bash
#!/bin/bash
# backup-hub.sh

BACKUP_DIR="/backups/hub"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_PATH="$BACKUP_DIR/hub_backup_$DATE"

# Create backup directory
mkdir -p "$BACKUP_PATH"

# Backup database
kubectl exec deployment/postgresql -n hub -- \
  pg_dump -U hub hub | gzip > "$BACKUP_PATH/database.sql.gz"

# Backup repositories
kubectl exec deployment/hub-backend -n hub -- \
  tar czf - /repositories | cat > "$BACKUP_PATH/repositories.tar.gz"

# Backup configuration
kubectl get configmap hub-config -n hub -o yaml > "$BACKUP_PATH/config.yaml"
kubectl get secret hub-secrets -n hub -o yaml > "$BACKUP_PATH/secrets.yaml"

# Upload to backup storage (optional)
az storage blob upload-batch \
  --destination backups \
  --source "$BACKUP_PATH" \
  --account-name "$AZURE_STORAGE_ACCOUNT"

echo "Backup completed: $BACKUP_PATH"
```

#### Cron Job for Automated Backups
```yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: hub-backup
  namespace: hub
spec:
  schedule: "0 2 * * *"  # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
          - name: backup
            image: hub/backup:latest
            command:
            - /bin/bash
            - -c
            - |
              # Backup script content here
              echo "Running backup..."
              /scripts/backup-hub.sh
          restartPolicy: OnFailure
```

#### Recovery Procedures
```bash
# Database recovery
kubectl exec -i deployment/postgresql -n hub -- \
  psql -U hub hub < backup/database.sql

# Repository recovery
kubectl exec -i deployment/hub-backend -n hub -- \
  tar xzf - -C / < backup/repositories.tar.gz

# Configuration recovery
kubectl apply -f backup/config.yaml
kubectl apply -f backup/secrets.yaml
kubectl rollout restart deployment/hub-backend -n hub
kubectl rollout restart deployment/hub-frontend -n hub
```

## Monitoring and Maintenance

### Health Monitoring

#### Health Check Endpoints
- **Backend**: `GET /health`
- **Frontend**: `GET /health`
- **Database**: Connection pool status
- **Redis**: Connection and memory status

#### Kubernetes Health Checks
```yaml
# Liveness and readiness probes
containers:
- name: backend
  livenessProbe:
    httpGet:
      path: /health
      port: 8080
    initialDelaySeconds: 30
    periodSeconds: 10
    timeoutSeconds: 5
    failureThreshold: 3
  
  readinessProbe:
    httpGet:
      path: /ready
      port: 8080
    initialDelaySeconds: 5
    periodSeconds: 5
    timeoutSeconds: 3
    failureThreshold: 3
```

### Metrics and Monitoring

#### Prometheus Configuration
```yaml
# ServiceMonitor for Prometheus
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: hub-metrics
  namespace: hub
spec:
  selector:
    matchLabels:
      app: hub-backend
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

#### Key Metrics to Monitor
```yaml
# Application metrics
- hub_http_requests_total
- hub_http_request_duration_seconds
- hub_active_users
- hub_repositories_total
- hub_git_operations_total

# Infrastructure metrics
- container_cpu_usage_seconds_total
- container_memory_usage_bytes
- postgresql_up
- redis_up

# Business metrics
- hub_user_registrations_total
- hub_repository_creates_total
- hub_pull_requests_total
- hub_builds_total
```

#### Grafana Dashboard
```json
{
  "dashboard": {
    "title": "Hub Git Hosting Service",
    "panels": [
      {
        "title": "HTTP Requests",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(hub_http_requests_total[5m])",
            "legendFormat": "{{method}} {{status}}"
          }
        ]
      },
      {
        "title": "Response Time",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(hub_http_request_duration_seconds_bucket[5m]))",
            "legendFormat": "95th percentile"
          }
        ]
      }
    ]
  }
}
```

### Log Management

#### Structured Logging Configuration
```yaml
# In config.yaml
logging:
  level: info
  format: json
  output: stdout
  fields:
    service: hub
    version: v1.0.0
```

#### Log Aggregation with Fluentd
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: fluentd-config
data:
  fluent.conf: |
    <source>
      @type tail
      path /var/log/containers/*hub*.log
      pos_file /var/log/fluentd/hub.log.pos
      format json
      tag kubernetes.hub
    </source>
    
    <match kubernetes.hub>
      @type elasticsearch
      host elasticsearch.logging.svc.cluster.local
      port 9200
      index_name hub-logs
    </match>
```

### Maintenance Tasks

#### Regular Maintenance Checklist
- [ ] **Weekly**:
  - Check system resource usage
  - Review error logs
  - Verify backup completion
  - Monitor certificate expiration

- [ ] **Monthly**:
  - Update system packages
  - Rotate log files
  - Clean up old artifacts
  - Review user access

- [ ] **Quarterly**:
  - Update Hub to latest version
  - Audit security configurations
  - Performance tuning
  - Disaster recovery testing

#### Cleanup Scripts
```bash
#!/bin/bash
# cleanup-hub.sh

# Clean up old build artifacts (older than 30 days)
find /var/lib/hub/artifacts -type f -mtime +30 -delete

# Clean up old log files (older than 90 days)
find /var/log/hub -type f -name "*.log" -mtime +90 -delete

# Clean up unused Docker images
docker system prune -f

# Clean up unused Git objects
kubectl exec deployment/hub-backend -n hub -- \
  find /repositories -name "*.git" -type d -exec git -C {} gc --aggressive \;

echo "Cleanup completed"
```

## Scaling and Performance

### Horizontal Scaling

#### Backend Scaling
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hub-backend
spec:
  replicas: 5  # Scale to 5 replicas
  # ... rest of deployment spec
```

#### Auto-scaling Configuration
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: hub-backend-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: hub-backend
  minReplicas: 3
  maxReplicas: 20
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

### Database Scaling

#### Read Replicas
```yaml
# PostgreSQL with read replicas
database:
  primary:
    host: postgresql-primary
    port: 5432
  replicas:
    - host: postgresql-replica-1
      port: 5432
    - host: postgresql-replica-2
      port: 5432
  read_preference: replica
```

#### Connection Pooling
```yaml
# PgBouncer configuration
pgbouncer:
  enabled: true
  pool_mode: transaction
  max_client_conn: 1000
  default_pool_size: 25
  reserve_pool_size: 5
```

### Performance Optimization

#### Caching Strategy
```yaml
# Redis caching configuration
cache:
  redis:
    enabled: true
    cluster: true
    nodes:
      - redis-1:6379
      - redis-2:6379
      - redis-3:6379
  strategies:
    repository_metadata: 1h
    user_sessions: 24h
    git_objects: 7d
```

#### CDN Configuration
```yaml
# CloudFlare or Azure CDN
cdn:
  enabled: true
  provider: azure
  cache_rules:
    - pattern: "/static/*"
      ttl: 1y
    - pattern: "/assets/*"
      ttl: 30d
    - pattern: "/api/*"
      ttl: 0
```

## Troubleshooting

### Common Issues

#### Application Won't Start
```bash
# Check pod status
kubectl get pods -n hub

# Check pod logs
kubectl logs deployment/hub-backend -n hub

# Check events
kubectl get events -n hub --sort-by=.metadata.creationTimestamp

# Check configuration
kubectl exec deployment/hub-backend -n hub -- cat /config.yaml
```

#### Database Connection Issues
```bash
# Test database connectivity
kubectl exec deployment/hub-backend -n hub -- \
  nc -zv postgresql 5432

# Check database status
kubectl exec deployment/postgresql -n hub -- \
  psql -U hub -c "SELECT version();"

# Check database logs
kubectl logs deployment/postgresql -n hub
```

#### Performance Issues
```bash
# Check resource usage
kubectl top pods -n hub
kubectl top nodes

# Check metrics
curl http://hub-backend:8080/metrics

# Analyze slow queries
kubectl exec deployment/postgresql -n hub -- \
  psql -U hub -c "SELECT query, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"
```

### Diagnostic Commands

#### System Information
```bash
# Cluster information
kubectl cluster-info
kubectl get nodes -o wide

# Hub deployment status
kubectl get all -n hub
kubectl describe deployment hub-backend -n hub

# Resource quotas and limits
kubectl describe resourcequota -n hub
kubectl describe limitrange -n hub
```

#### Log Analysis
```bash
# Application logs
kubectl logs -f deployment/hub-backend -n hub --tail=100

# System logs
journalctl -u kubelet -f

# Ingress logs
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller
```

## Upgrade and Migration

### Version Upgrades

#### Pre-upgrade Checklist
- [ ] **Backup**: Complete system backup
- [ ] **Test**: Verify upgrade in staging environment
- [ ] **Dependencies**: Check compatibility requirements
- [ ] **Downtime**: Plan maintenance window
- [ ] **Rollback**: Prepare rollback procedure

#### Rolling Upgrade Process
```bash
# 1. Update images
kubectl set image deployment/hub-backend backend=hub/backend:v1.1.0 -n hub
kubectl set image deployment/hub-frontend frontend=hub/frontend:v1.1.0 -n hub

# 2. Monitor rollout
kubectl rollout status deployment/hub-backend -n hub
kubectl rollout status deployment/hub-frontend -n hub

# 3. Verify deployment
kubectl get pods -n hub
curl -f https://hub.yourdomain.com/health
```

#### Database Migrations
```bash
# Run database migrations
kubectl exec deployment/hub-backend -n hub -- \
  ./cmd/migrate/migrate up

# Verify migration status
kubectl exec deployment/hub-backend -n hub -- \
  ./cmd/migrate/migrate status
```

### Platform Migration

#### From GitHub Enterprise
```bash
# Use migration tool
kubectl exec deployment/hub-backend -n hub -- \
  ./cmd/migrate/migrate import-github \
  --token $GITHUB_TOKEN \
  --org source-org \
  --target-org hub-org
```

#### From GitLab
```bash
# Import GitLab projects
kubectl exec deployment/hub-backend -n hub -- \
  ./cmd/migrate/migrate import-gitlab \
  --url https://gitlab.company.com \
  --token $GITLAB_TOKEN \
  --group source-group
```

### Disaster Recovery

#### Complete System Recovery
```bash
# 1. Restore infrastructure
terraform apply -var-file="disaster-recovery.tfvars"

# 2. Restore database
kubectl exec -i deployment/postgresql -n hub -- \
  psql -U hub hub < backup/database.sql

# 3. Restore repositories
kubectl exec -i deployment/hub-backend -n hub -- \
  tar xzf - -C / < backup/repositories.tar.gz

# 4. Restore configuration
kubectl apply -f backup/config.yaml
kubectl apply -f backup/secrets.yaml

# 5. Restart services
kubectl rollout restart deployment -n hub
```

#### Recovery Testing
```bash
# Regular disaster recovery tests
./scripts/dr-test.sh --scenario complete-failure
./scripts/dr-test.sh --scenario database-corruption
./scripts/dr-test.sh --scenario storage-failure
```

---

This administrator guide provides comprehensive information for deploying and managing Hub Git Hosting Service. For additional support, refer to the [Developer Guide](developer-guide.md) for API and integration details, or the [User Guide](user-guide.md) for end-user instructions.

## DNS Routing with ExternalDNS

ExternalDNS integrates with Kubernetes to automatically manage DNS records in Azure public DNS zones based on Kubernetes resources such as Ingress or Service.

### Prerequisites

- Azure public DNS zone created and configured for your domain.
- Service principal credentials with permissions to manage DNS records.
- Kubernetes secret or ConfigMap containing Azure credentials and DNS zone information.

### Deploy ExternalDNS

Apply the ExternalDNS manifest to your cluster:

```bash
kubectl apply -f k8s/external-dns.yaml
```

### Configure Ingress Resources

Annotate your Ingress resources to enable ExternalDNS to manage DNS records:

```yaml
metadata:
  annotations:
    external-dns.alpha.kubernetes.io/hostname: <your.hostname.example.com>
```

### Configuration Options

- **Azure Resource Group**: Set via `--azure-resource-group` argument or environment variable.
- **Public DNS Zone**: Set via `--domain-filter` or environment variable `PUBLIC_DNS_ZONE_NAME`.
- **Image Version**: ExternalDNS image version is currently hardcoded (`v0.14.2`), consider updating the manifest for newer versions.

For the latest updates and community support, visit the [project repository](https://github.com/a5c-ai/hub).
