# Distributed Git Storage Implementation

## Overview

This document describes the implementation of Option 2: Kubernetes-Native Distributed Storage for the Hub git hosting service. This solution provides a highly available, scalable, and resilient git storage layer using Kubernetes StatefulSets and distributed storage technologies.

## Architecture

### Components

1. **DistributedGitService** - Main service orchestrating distributed git operations
2. **ConsistentHashRouter** - Load balances repositories across storage nodes using consistent hashing
3. **DistributedLockManager** - Ensures data consistency with distributed locking
4. **ServiceDiscovery** - Discovers and monitors git storage nodes in Kubernetes
5. **Git Storage StatefulSet** - Kubernetes-native storage pods with persistent volumes

### High-Level Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Hub Backend   │    │   Hub Backend   │    │   Hub Backend   │
│   (API Layer)   │    │   (API Layer)   │    │   (API Layer)   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │  Distributed Git Service  │
                    │  (Load Balancer/Router)   │
                    └─────────────┬─────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
  ┌───────┴───────┐      ┌───────┴───────┐      ┌───────┴───────┐
  │ Git Storage   │      │ Git Storage   │      │ Git Storage   │
  │   Node 1      │      │   Node 2      │      │   Node 3      │
  │ (StatefulSet) │      │ (StatefulSet) │      │ (StatefulSet) │
  └───────────────┘      └───────────────┘      └───────────────┘
```

### Key Features

- **High Availability**: Repositories are replicated across multiple nodes
- **Consistent Hashing**: Ensures balanced distribution and minimal reshuffling
- **Distributed Locking**: Prevents conflicts during concurrent operations
- **Auto-Discovery**: Kubernetes service discovery for dynamic node management
- **Health Monitoring**: Automatic health checks and failover
- **Horizontal Scaling**: Add/remove storage nodes dynamically

## Components Detail

### 1. DistributedGitService

**Location**: `internal/git/distributed_service.go`

The main orchestrator that:
- Routes git operations to appropriate storage nodes
- Handles repository initialization across multiple nodes
- Manages replication for write operations
- Implements fallback strategies for read operations

**Key Methods**:
- `InitRepository()` - Creates repository on multiple nodes
- `CreateFile()` - Writes to primary node, replicates to others
- `GetCommits()` - Reads from primary node with fallback

### 2. ConsistentHashRouter

**Location**: `internal/git/consistent_hash_router.go`

Implements consistent hashing for repository distribution:
- Virtual nodes for better distribution
- Weighted node support
- Minimal key redistribution on node changes
- O(log N) lookup performance

**Features**:
- 100 virtual nodes per physical node (configurable)
- Support for node weights
- Consistent repository-to-node mapping
- Automatic rebalancing on topology changes

### 3. DistributedLockManager

**Location**: `internal/git/distributed_lock_manager.go`

Provides distributed locking to ensure consistency:
- Repository-level locks for major operations
- Branch-level locks for concurrent push protection
- File-level locks for fine-grained control
- TTL-based expiration with renewal
- Different lock scopes (global, repository, branch, file)

**Lock Types**:
- Repository locks: For initialization, deletion
- Branch locks: For push operations, branch management
- File locks: For individual file operations
- Global locks: For cluster-wide operations

### 4. ServiceDiscovery

**Location**: `internal/git/service_discovery.go`

Kubernetes-native service discovery:
- Watches StatefulSet endpoints
- Health checking of storage nodes
- Automatic node registration/deregistration
- ConfigMap-based configuration fallback

## Kubernetes Deployment

### StatefulSet Configuration

**Location**: `k8s/git-storage-statefulset.yaml`

- 3 replicas for high availability
- Persistent volume claims per pod
- Azure premium storage for performance
- Init containers for setup
- Health probes for monitoring

**Storage**:
- 100GB per node (configurable)
- Azure managed-premium storage class
- ReadWriteOnce volumes per pod

### Load Balancing

**Location**: `k8s/git-storage-load-balancer.yaml`

- Istio VirtualService for advanced routing
- Repository-based consistent hashing
- Circuit breakers and retries
- Network policies for security

### Monitoring

**Location**: `k8s/git-storage-monitoring.yaml`

- Prometheus metrics collection
- Grafana dashboard for visualization
- Alerting for node failures
- Log aggregation with ELK stack

## Configuration

### Environment Variables

- `DISTRIBUTED_STORAGE_ENABLED`: Enable distributed storage
- `DISTRIBUTED_STORAGE_NODE_ID`: Unique node identifier
- `DISTRIBUTED_STORAGE_REPLICATION_COUNT`: Number of replicas
- `DISTRIBUTED_STORAGE_CONSISTENT_HASHING`: Enable consistent hashing
- `DISTRIBUTED_STORAGE_HEALTH_CHECK_INTERVAL`: Health check frequency

### Config File

```yaml
storage:
  repository_path: "/var/lib/hub/repositories"
  distributed:
    enabled: true
    node_id: "hub-git-storage-0.hub"
    replication_count: 3
    consistent_hashing: true
    health_check_interval: "30s"
    storage_nodes:
      - id: "node-1"
        address: "http://hub-git-storage-0:8080"
        weight: 1
      - id: "node-2" 
        address: "http://hub-git-storage-1:8080"
        weight: 1
      - id: "node-3"
        address: "http://hub-git-storage-2:8080"
        weight: 1
```

## Deployment Guide

### Prerequisites

1. Kubernetes cluster with at least 3 nodes
2. Azure storage classes configured
3. Istio service mesh (optional but recommended)
4. Prometheus and Grafana for monitoring

### Step 1: Deploy Storage Infrastructure

```bash
# Create namespace
kubectl apply -f k8s/namespace.yaml

# Create storage resources
kubectl apply -f k8s/storage.yaml

# Deploy git storage StatefulSet
kubectl apply -f k8s/git-storage-statefulset.yaml
```

### Step 2: Configure Load Balancing

```bash
# Deploy service mesh configuration
kubectl apply -f k8s/git-storage-load-balancer.yaml
```

### Step 3: Enable Monitoring

```bash
# Deploy monitoring stack
kubectl apply -f k8s/git-storage-monitoring.yaml
```

### Step 4: Update Backend Configuration

```bash
# Update backend deployment with distributed storage config
kubectl apply -f k8s/backend-deployment.yaml
```

## Operations

### Scaling

**Add Node**:
```bash
kubectl patch statefulset hub-git-storage -p '{"spec":{"replicas":4}}'
```

**Remove Node**:
```bash
kubectl patch statefulset hub-git-storage -p '{"spec":{"replicas":2}}'
```

### Monitoring

**Check Node Status**:
```bash
kubectl get pods -l app=hub-git-storage
```

**View Logs**:
```bash
kubectl logs hub-git-storage-0 -f
```

**Metrics**:
- Access Grafana dashboard at configured URL
- Monitor repository distribution
- Track operation performance
- Alert on failures

### Troubleshooting

**Common Issues**:

1. **Node not joining cluster**
   - Check service discovery logs
   - Verify network connectivity
   - Confirm ConfigMap configuration

2. **Lock contention**
   - Monitor lock wait times
   - Review concurrent operation patterns
   - Adjust lock TTLs if needed

3. **Uneven distribution**
   - Check consistent hash ring status
   - Verify node weights
   - Rebalance if necessary

**Debug Commands**:
```bash
# Check service endpoints
kubectl get endpoints hub-git-storage-headless

# View storage utilization
kubectl exec hub-git-storage-0 -- df -h /git-data

# Check lock status (via API)
curl http://hub-git-storage-0:8080/debug/locks
```

## Performance Characteristics

### Throughput
- **Read Operations**: Scales linearly with node count
- **Write Operations**: Limited by replication factor
- **Repository Creation**: Requires coordination across nodes

### Latency
- **Read Latency**: ~10-50ms (single node lookup)
- **Write Latency**: ~50-200ms (includes replication)
- **Lock Acquisition**: ~1-10ms (local operations)

### Capacity
- **Per Node**: 100GB default (configurable)
- **Total Cluster**: N × 100GB (where N = number of nodes)
- **Effective Storage**: Total ÷ Replication Factor

## Testing

### Unit Tests

```bash
go test ./internal/git -v
```

### Integration Tests

```bash
go test ./internal/git -v -tags=integration
```

### Load Testing

```bash
# Use k6 or similar tools to test at scale
k6 run tests/load/git-operations.js
```

## Security Considerations

1. **Network Policies**: Restrict inter-pod communication
2. **RBAC**: Limit service account permissions
3. **Encryption**: TLS for inter-node communication
4. **Audit Logging**: Track all git operations
5. **Secret Management**: Secure storage of credentials

## Future Enhancements

1. **Cross-Region Replication**: Geo-distributed storage
2. **Smart Caching**: LRU cache for frequently accessed repos
3. **Compression**: Repository-level compression
4. **Deduplication**: Cross-repository object sharing
5. **Backup Integration**: Automated backup to object storage

## Conclusion

This distributed git storage implementation provides a robust, scalable foundation for the Hub git hosting service. It addresses the requirements for better availability, distribution, and robustness while maintaining compatibility with the existing API layer.

The Kubernetes-native approach ensures operational simplicity while the distributed architecture provides the resilience needed for production workloads.