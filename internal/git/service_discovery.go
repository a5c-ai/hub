package git

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// ServiceDiscovery handles discovery of git storage nodes in Kubernetes
type ServiceDiscovery struct {
	kubeClient   kubernetes.Interface
	namespace    string
	serviceName  string
	logger       *logrus.Logger
	nodes        map[string]*StorageNode
	nodesMutex   sync.RWMutex
	stopCh       chan struct{}
	updateCh     chan []*StorageNode
}

// NewServiceDiscovery creates a new service discovery instance
func NewServiceDiscovery(namespace, serviceName string, logger *logrus.Logger) (*ServiceDiscovery, error) {
	// Create Kubernetes client from in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	sd := &ServiceDiscovery{
		kubeClient:  kubeClient,
		namespace:   namespace,
		serviceName: serviceName,
		logger:      logger,
		nodes:       make(map[string]*StorageNode),
		stopCh:      make(chan struct{}),
		updateCh:    make(chan []*StorageNode, 10),
	}

	// Start discovery loop
	go sd.discoveryLoop()

	return sd, nil
}

// GetNodes returns the current list of discovered storage nodes
func (sd *ServiceDiscovery) GetNodes() []*StorageNode {
	sd.nodesMutex.RLock()
	defer sd.nodesMutex.RUnlock()

	nodes := make([]*StorageNode, 0, len(sd.nodes))
	for _, node := range sd.nodes {
		nodes = append(nodes, node)
	}

	// Sort nodes by ID for consistent ordering
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})

	return nodes
}

// GetHealthyNodes returns only healthy storage nodes
func (sd *ServiceDiscovery) GetHealthyNodes() []*StorageNode {
	sd.nodesMutex.RLock()
	defer sd.nodesMutex.RUnlock()

	var healthyNodes []*StorageNode
	for _, node := range sd.nodes {
		if node.Healthy {
			healthyNodes = append(healthyNodes, node)
		}
	}

	// Sort nodes by ID for consistent ordering
	sort.Slice(healthyNodes, func(i, j int) bool {
		return healthyNodes[i].ID < healthyNodes[j].ID
	})

	return healthyNodes
}

// GetNodeByID returns a specific node by ID
func (sd *ServiceDiscovery) GetNodeByID(nodeID string) (*StorageNode, bool) {
	sd.nodesMutex.RLock()
	defer sd.nodesMutex.RUnlock()

	node, exists := sd.nodes[nodeID]
	return node, exists
}

// Subscribe returns a channel that receives updates when nodes change
func (sd *ServiceDiscovery) Subscribe() <-chan []*StorageNode {
	return sd.updateCh
}

// Close stops the service discovery
func (sd *ServiceDiscovery) Close() {
	close(sd.stopCh)
}

// Private methods

func (sd *ServiceDiscovery) discoveryLoop() {
	ticker := time.NewTicker(30 * time.Second) // Discovery interval
	defer ticker.Stop()

	// Initial discovery
	sd.performDiscovery()

	for {
		select {
		case <-sd.stopCh:
			return
		case <-ticker.C:
			sd.performDiscovery()
		}
	}
}

func (sd *ServiceDiscovery) performDiscovery() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get endpoints for the headless service
	endpoints, err := sd.kubeClient.CoreV1().Endpoints(sd.namespace).Get(ctx, sd.serviceName, metav1.GetOptions{})
	if err != nil {
		sd.logger.WithError(err).WithFields(logrus.Fields{
			"namespace":    sd.namespace,
			"service_name": sd.serviceName,
		}).Error("Failed to get service endpoints")
		return
	}

	discoveredNodes := make(map[string]*StorageNode)

	for _, subset := range endpoints.Subsets {
		for _, address := range subset.Addresses {
			if address.TargetRef != nil && address.TargetRef.Kind == "Pod" {
				nodeID := sd.generateNodeID(address.TargetRef.Name, sd.namespace)
				nodeAddress := fmt.Sprintf("http://%s:8080", address.IP)

				// Check if this is a new node or existing node
				existingNode, exists := sd.nodes[nodeID]
				
				node := &StorageNode{
					ID:       nodeID,
					Address:  nodeAddress,
					Weight:   1, // Default weight
					Healthy:  true,
					LastSeen: time.Now(),
				}

				// Preserve existing node data if available
				if exists {
					node.Weight = existingNode.Weight
					// Health will be updated by health checks
				}

				// Perform health check
				healthy := sd.checkNodeHealth(ctx, nodeAddress)
				node.Healthy = healthy

				discoveredNodes[nodeID] = node

				sd.logger.WithFields(logrus.Fields{
					"node_id":  nodeID,
					"address":  nodeAddress,
					"healthy":  healthy,
				}).Debug("Discovered git storage node")
			}
		}
	}

	// Update nodes and notify subscribers
	sd.updateNodes(discoveredNodes)
}

func (sd *ServiceDiscovery) updateNodes(discoveredNodes map[string]*StorageNode) {
	sd.nodesMutex.Lock()
	defer sd.nodesMutex.Unlock()

	// Check for changes
	changed := false
	if len(sd.nodes) != len(discoveredNodes) {
		changed = true
	} else {
		for nodeID, node := range discoveredNodes {
			existingNode, exists := sd.nodes[nodeID]
			if !exists || existingNode.Address != node.Address || existingNode.Healthy != node.Healthy {
				changed = true
				break
			}
		}
	}

	if changed {
		sd.nodes = discoveredNodes

		// Notify subscribers
		nodeList := make([]*StorageNode, 0, len(discoveredNodes))
		for _, node := range discoveredNodes {
			nodeList = append(nodeList, node)
		}

		// Sort for consistent ordering
		sort.Slice(nodeList, func(i, j int) bool {
			return nodeList[i].ID < nodeList[j].ID
		})

		select {
		case sd.updateCh <- nodeList:
		default:
			// Channel is full, skip this update
		}

		sd.logger.WithField("node_count", len(nodeList)).Info("Updated git storage nodes")
	}
}

func (sd *ServiceDiscovery) checkNodeHealth(ctx context.Context, nodeAddress string) bool {
	// TODO: Implement actual health check via HTTP request to /health endpoint
	// For now, assume all discovered nodes are healthy
	return true
}

func (sd *ServiceDiscovery) generateNodeID(podName, namespace string) string {
	return fmt.Sprintf("%s.%s", podName, namespace)
}

// DistributedConfig adapter for service discovery
func (sd *ServiceDiscovery) GetDistributedConfig() *DistributedConfig {
	nodes := sd.GetHealthyNodes()
	
	storageNodes := make([]StorageNode, len(nodes))
	for i, node := range nodes {
		storageNodes[i] = StorageNode{
			ID:      node.ID,
			Address: node.Address,
			Weight:  node.Weight,
		}
	}

	return &DistributedConfig{
		Enabled:             len(nodes) > 0,
		NodeID:              "", // Will be set by the service
		StorageNodes:        storageNodes,
		ReplicationCount:    3,
		ConsistentHashing:   true,
		HealthCheckInterval: 30 * time.Second,
	}
}

// K8sNodeSelector helps select appropriate nodes for git storage workloads
type K8sNodeSelector struct {
	kubeClient kubernetes.Interface
	logger     *logrus.Logger
}

// NewK8sNodeSelector creates a new node selector
func NewK8sNodeSelector(logger *logrus.Logger) (*K8sNodeSelector, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &K8sNodeSelector{
		kubeClient: kubeClient,
		logger:     logger,
	}, nil
}

// GetOptimalNodes returns nodes suitable for git storage based on criteria
func (ns *K8sNodeSelector) GetOptimalNodes(ctx context.Context, criteria NodeSelectionCriteria) ([]string, error) {
	nodes, err := ns.kubeClient.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}

	var optimalNodes []string

	for _, node := range nodes.Items {
		if ns.isNodeSuitable(&node, criteria) {
			optimalNodes = append(optimalNodes, node.Name)
		}
	}

	return optimalNodes, nil
}

// NodeSelectionCriteria defines criteria for selecting optimal nodes
type NodeSelectionCriteria struct {
	MinCPU              string            `json:"min_cpu"`
	MinMemory           string            `json:"min_memory"`
	MinStorage          string            `json:"min_storage"`
	RequiredLabels      map[string]string `json:"required_labels"`
	PreferredLabels     map[string]string `json:"preferred_labels"`
	AvoidLabels         map[string]string `json:"avoid_labels"`
	RequiredTaints      []string          `json:"required_taints"`
	ToleratedTaints     []string          `json:"tolerated_taints"`
	StorageClass        string            `json:"storage_class"`
	AvailabilityZone    string            `json:"availability_zone"`
	MaxPodsPerNode      int               `json:"max_pods_per_node"`
}

func (ns *K8sNodeSelector) isNodeSuitable(node *v1.Node, criteria NodeSelectionCriteria) bool {
	// Check node readiness
	if !ns.isNodeReady(node) {
		return false
	}

	// Check required labels
	for key, value := range criteria.RequiredLabels {
		if nodeValue, exists := node.Labels[key]; !exists || nodeValue != value {
			return false
		}
	}

	// Check avoid labels
	for key, value := range criteria.AvoidLabels {
		if nodeValue, exists := node.Labels[key]; exists && nodeValue == value {
			return false
		}
	}

	// Check availability zone if specified
	if criteria.AvailabilityZone != "" {
		zone, exists := node.Labels["topology.kubernetes.io/zone"]
		if !exists || zone != criteria.AvailabilityZone {
			return false
		}
	}

	// Check taints
	if !ns.checkTaints(node, criteria) {
		return false
	}

	// TODO: Add resource capacity checks (CPU, Memory, Storage)
	// This would require parsing resource quantities and comparing

	return true
}

func (ns *K8sNodeSelector) isNodeReady(node *v1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == v1.NodeReady {
			return condition.Status == v1.ConditionTrue
		}
	}
	return false
}

func (ns *K8sNodeSelector) checkTaints(node *v1.Node, criteria NodeSelectionCriteria) bool {
	// Check if required taints are present
	for _, requiredTaint := range criteria.RequiredTaints {
		found := false
		for _, taint := range node.Spec.Taints {
			if taint.Key == requiredTaint {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// Check if any non-tolerated taints are present
	for _, taint := range node.Spec.Taints {
		if taint.Effect == v1.TaintEffectNoSchedule || taint.Effect == v1.TaintEffectNoExecute {
			tolerated := false
			for _, toleratedTaint := range criteria.ToleratedTaints {
				if taint.Key == toleratedTaint {
					tolerated = true
					break
				}
			}
			if !tolerated {
				return false
			}
		}
	}

	return true
}

// GetNodeMetrics returns metrics for a node (requires metrics-server)
func (ns *K8sNodeSelector) GetNodeMetrics(ctx context.Context, nodeName string) (*NodeMetrics, error) {
	// This would typically use the metrics API
	// For now, return mock metrics
	return &NodeMetrics{
		NodeName:    nodeName,
		CPUUsage:    "50m",
		MemoryUsage: "1Gi",
		StorageUsage: map[string]string{
			"ephemeral-storage": "10Gi",
		},
		NetworkRx: "100MB",
		NetworkTx: "100MB",
		Timestamp: time.Now(),
	}, nil
}

// NodeMetrics represents resource usage metrics for a node
type NodeMetrics struct {
	NodeName     string            `json:"node_name"`
	CPUUsage     string            `json:"cpu_usage"`
	MemoryUsage  string            `json:"memory_usage"`
	StorageUsage map[string]string `json:"storage_usage"`
	NetworkRx    string            `json:"network_rx"`
	NetworkTx    string            `json:"network_tx"`
	Timestamp    time.Time         `json:"timestamp"`
}

// ConfigMapBasedDiscovery provides an alternative discovery method using ConfigMaps
type ConfigMapBasedDiscovery struct {
	kubeClient  kubernetes.Interface
	namespace   string
	configMap   string
	logger      *logrus.Logger
}

// NewConfigMapBasedDiscovery creates a configmap-based discovery instance
func NewConfigMapBasedDiscovery(namespace, configMap string, logger *logrus.Logger) (*ConfigMapBasedDiscovery, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &ConfigMapBasedDiscovery{
		kubeClient: kubeClient,
		namespace:  namespace,
		configMap:  configMap,
		logger:     logger,
	}, nil
}

// GetStorageNodes reads storage nodes from a ConfigMap
func (cmd *ConfigMapBasedDiscovery) GetStorageNodes(ctx context.Context) ([]*StorageNode, error) {
	cm, err := cmd.kubeClient.CoreV1().ConfigMaps(cmd.namespace).Get(ctx, cmd.configMap, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap %s: %w", cmd.configMap, err)
	}

	nodesData, exists := cm.Data["storage-nodes"]
	if !exists {
		return nil, fmt.Errorf("storage-nodes key not found in configmap %s", cmd.configMap)
	}

	var nodes []*StorageNode
	if err := json.Unmarshal([]byte(nodesData), &nodes); err != nil {
		return nil, fmt.Errorf("failed to unmarshal storage nodes: %w", err)
	}

	return nodes, nil
}

// UpdateStorageNodes updates the storage nodes in the ConfigMap
func (cmd *ConfigMapBasedDiscovery) UpdateStorageNodes(ctx context.Context, nodes []*StorageNode) error {
	nodesData, err := json.Marshal(nodes)
	if err != nil {
		return fmt.Errorf("failed to marshal storage nodes: %w", err)
	}

	cm, err := cmd.kubeClient.CoreV1().ConfigMaps(cmd.namespace).Get(ctx, cmd.configMap, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get configmap %s: %w", cmd.configMap, err)
	}

	if cm.Data == nil {
		cm.Data = make(map[string]string)
	}
	cm.Data["storage-nodes"] = string(nodesData)

	_, err = cmd.kubeClient.CoreV1().ConfigMaps(cmd.namespace).Update(ctx, cm, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update configmap %s: %w", cmd.configMap, err)
	}

	return nil
}