# Hub Deployment Guide

This guide covers Docker containerization and Kubernetes deployment for the Hub Git Hosting Service.

## Overview

The Hub application consists of:
- **Backend**: Go-based REST API server
- **Frontend**: Next.js React application  
- **Database**: PostgreSQL
- **Cache**: Redis
- **Storage**: Persistent volumes for Git repositories

## Prerequisites

### Required Tools
- Docker 20.10+
- Kubernetes 1.24+
- kubectl configured for your cluster
- Helm 3.8+ (optional, for Helm deployments)

### Azure-specific Requirements
- Azure CLI logged in (for Azure Container Registry)
- AKS cluster with:
  - Ingress controller (nginx-ingress)
  - Cert-manager for TLS certificates
  - Storage classes: `managed-premium`, `azure-files`

## Quick Start

### 1. Build Docker Images

```bash
# Build locally
./scripts/build-images.sh

# Build and push to registry
REGISTRY=myregistry.azurecr.io VERSION=v1.0.0 ./scripts/build-images.sh
```

### 2. Configure Secrets

Before deploying, update the secrets in `k8s/secrets.yaml` with base64-encoded values:

```bash
# Generate base64 values
echo -n "your-database-password" | base64
echo -n "your-jwt-secret" | base64
echo -n "your-github-client-id" | base64
echo -n "your-github-client-secret" | base64
```

### 3. Deploy to Kubernetes

```bash
# Deploy to development environment
./scripts/deploy-k8s.sh development --wait

# Deploy to production with Helm
./scripts/deploy-k8s.sh production --helm --values prod-values.yaml --wait
```

## Docker Configuration

### Backend Dockerfile
- **Base Image**: golang:1.24-alpine (build) → alpine:latest (runtime)
- **Features**: Multi-stage build, non-root user, health checks
- **Security**: Minimal attack surface, capability dropping
- **Size**: ~20MB final image

### Frontend Dockerfile  
- **Base Image**: node:18-alpine (build) → node:18-alpine (runtime)
- **Features**: Production-optimized, non-root user, health checks
- **Build**: Static asset optimization via Next.js
- **Size**: ~150MB final image

### Build Arguments
Both Dockerfiles support these build arguments:
- `VERSION`: Application version
- `BUILD_DATE`: Build timestamp
- `VCS_REF`: Git commit hash

## Kubernetes Architecture

### Namespace Structure
```
hub/
├── ConfigMaps (configuration)
├── Secrets (sensitive data)
├── Deployments (app workloads)
├── Services (networking)
├── Ingress (external access)
├── PVCs (persistent storage)
├── HPAs (auto-scaling)
└── NetworkPolicies (security)
```

### Resource Requirements

| Component | CPU Request | Memory Request | CPU Limit | Memory Limit |
|-----------|-------------|----------------|----------|--------------|
| Backend | 250m | 256Mi | 500m | 512Mi |
| Frontend | 100m | 128Mi | 200m | 256Mi |
| PostgreSQL | 250m | 256Mi | 500m | 1Gi |
| Redis | 100m | 128Mi | 200m | 256Mi |

### Storage Requirements

| Volume | Size | Access Mode | Storage Class |
|--------|------|-------------|---------------|
| Repositories | 1Ti | ReadWriteMany | azure-files |
| Database | 100Gi | ReadWriteOnce | managed-premium |
| Redis | 10Gi | ReadWriteOnce | managed-premium |

## Configuration Management

### Environment Variables

#### Backend Configuration
```yaml
APP_ENV: production
LOG_LEVEL: info
GIN_MODE: release
PORT: 8080
DATABASE_URL: postgresql://...
REDIS_URL: redis://...
JWT_SECRET: <secret>
GITHUB_OAUTH_CLIENT_ID: <secret>
GITHUB_OAUTH_CLIENT_SECRET: <secret>
GIT_DATA_PATH: /repositories
```

#### Frontend Configuration  
```yaml
NODE_ENV: production
PORT: 3000
NEXT_PUBLIC_API_URL: https://hub.example.com/api
NEXT_PUBLIC_APP_URL: https://hub.example.com
```

### Secrets Management

All sensitive data is stored in Kubernetes Secrets:
- Database credentials
- JWT signing keys
- OAuth client secrets
- Admin user credentials

**Security Notes:**
- Use external secret management (Azure Key Vault, HashiCorp Vault) in production
- Rotate secrets regularly
- Enable secret encryption at rest

## Deployment Methods

### Method 1: Direct kubectl

```bash
# Apply manifests in order
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/secrets.yaml  # Update with real secrets first!
kubectl apply -f k8s/storage.yaml
kubectl apply -f k8s/postgresql-deployment.yaml
kubectl apply -f k8s/redis-deployment.yaml
kubectl apply -f k8s/backend-deployment.yaml
kubectl apply -f k8s/frontend-deployment.yaml
kubectl apply -f k8s/services.yaml
kubectl apply -f k8s/ingress.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/network-policy.yaml
```

### Method 2: Deployment Script

```bash
# Basic deployment
./scripts/deploy-k8s.sh production

# With options
./scripts/deploy-k8s.sh production --wait --skip-dependencies
```

### Method 3: Helm Chart

```bash
# Install dependencies
helm dependency update helm/hub

# Deploy with custom values
helm upgrade --install hub helm/hub \
  --namespace hub --create-namespace \
  --values production-values.yaml \
  --wait
```

## Helm Configuration

### Custom Values Example

```yaml
# production-values.yaml
replicaCount:
  backend: 5
  frontend: 3

image:
  backend:
    repository: myregistry.azurecr.io/hub/backend
    tag: v1.0.0
  frontend:
    repository: myregistry.azurecr.io/hub/frontend  
    tag: v1.0.0

ingress:
  hosts:
    - host: hub.mycompany.com
  tls:
    - secretName: hub-tls
      hosts:
        - hub.mycompany.com

resources:
  backend:
    requests:
      cpu: 500m
      memory: 512Mi
    limits:
      cpu: 1000m
      memory: 1Gi

persistence:
  repositories:
    size: 2Ti
    storageClass: azure-files-premium

secrets:
  jwt:
    secret: "super-secure-jwt-key"
  database:
    password: "secure-db-password"  
  github:
    oauth:
      clientId: "your-github-client-id"
      clientSecret: "your-github-client-secret"
```

## Networking & Security

### Ingress Configuration
- **TLS Termination**: Automatic SSL with cert-manager
- **Path Routing**: `/api/*` → Backend, `/*` → Frontend
- **Security Headers**: XSS protection, content type sniffing protection
- **Rate Limiting**: Configurable via nginx annotations

### Network Policies
- Backend: Can communicate with database and Redis only
- Frontend: Can communicate with backend only  
- Database/Redis: Accept connections from backend only
- All pods: Allow DNS resolution and ingress from nginx

### Security Features
- Non-root containers
- Read-only root filesystems where possible
- Capability dropping (remove ALL)
- Resource limits enforcement
- Pod security policies/Pod security standards

## Monitoring & Observability

### Health Checks
- **Liveness Probes**: Detect and restart failed containers
- **Readiness Probes**: Control traffic routing to healthy pods
- **Startup Probes**: Handle slow-starting containers

### Metrics & Monitoring
```yaml
# Optional ServiceMonitor for Prometheus
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: hub-backend-metrics
spec:
  selector:
    matchLabels:
      app: hub-backend
  endpoints:
  - port: http
    path: /metrics
```

### Logging
- Structured JSON logging in production
- Log aggregation via Fluentd/Fluent Bit
- Centralized logging in Azure Log Analytics

## Scaling & Performance

### Horizontal Pod Autoscaling
- **Backend**: 3-10 replicas based on CPU/memory
- **Frontend**: 2-5 replicas based on CPU/memory
- **Metrics**: CPU 70%, Memory 80% thresholds

### Vertical Scaling
- Monitor resource usage and adjust requests/limits
- Use VPA (Vertical Pod Autoscaler) for automatic sizing

### Storage Scaling
- Repository storage: Expandable PVCs
- Database: Point-in-time recovery, automated backups
- Redis: Clustering for high availability

## Troubleshooting

### Common Issues

#### Pods Not Starting
```bash
# Check pod events
kubectl describe pod <pod-name> -n hub

# Check logs
kubectl logs <pod-name> -n hub --previous

# Check resource constraints
kubectl top pods -n hub
```

#### Database Connection Issues
```bash
# Test database connectivity
kubectl exec -it deployment/hub-backend -n hub -- nc -zv postgresql 5432

# Check secrets
kubectl get secret hub-secrets -n hub -o yaml
```

#### Ingress Issues
```bash
# Check ingress status
kubectl describe ingress hub-ingress -n hub

# Check ingress controller logs
kubectl logs -n ingress-nginx deployment/ingress-nginx-controller
```

### Debug Commands

```bash
# Port forward to access services locally
kubectl port-forward -n hub svc/hub-backend-service 8080:8080
kubectl port-forward -n hub svc/hub-frontend-service 3000:3000

# Execute commands in containers
kubectl exec -it deployment/hub-backend -n hub -- /bin/sh
kubectl exec -it deployment/postgresql -n hub -- psql -U hub

# View resource usage
kubectl top pods -n hub
kubectl top nodes
```

## Backup & Recovery

### Database Backups
```bash
# Manual backup
kubectl exec deployment/postgresql -n hub -- pg_dump -U hub hub > backup.sql

# Restore
kubectl exec -i deployment/postgresql -n hub -- psql -U hub hub < backup.sql
```

### Repository Backups
- Use Azure Files snapshots
- Regular rsync to backup storage
- Git bare repository mirroring

## Production Checklist

- [ ] **Security**: Update all secrets with secure values
- [ ] **TLS**: Configure proper SSL certificates  
- [ ] **DNS**: Update DNS records to point to ingress
- [ ] **Monitoring**: Set up alerting and monitoring
- [ ] **Backups**: Configure automated backup procedures
- [ ] **Resource Limits**: Set appropriate CPU/memory limits
- [ ] **Scaling**: Configure HPA with proper thresholds
- [ ] **Network Policies**: Enable and test network security
- [ ] **Image Security**: Scan images for vulnerabilities
- [ ] **Access Control**: Configure RBAC appropriately

## Support

For deployment issues:
1. Check the troubleshooting section above
2. Review pod logs and events
3. Validate configuration against this guide
4. Open an issue with deployment details and error messages

---

**Note**: Replace `hub.example.com` with your actual domain and update registry URLs with your container registry.