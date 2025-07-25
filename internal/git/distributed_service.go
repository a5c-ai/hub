package git

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// StorageNode represents a git storage node in the distributed cluster
type StorageNode struct {
	ID       string `json:"id"`
	Address  string `json:"address"`
	Weight   int    `json:"weight"`
	Healthy  bool   `json:"healthy"`
	LastSeen time.Time `json:"last_seen"`
}

// DistributedConfig holds configuration for distributed git storage
type DistributedConfig struct {
	Enabled         bool          `mapstructure:"enabled"`
	NodeID          string        `mapstructure:"node_id"`
	StorageNodes    []StorageNode `mapstructure:"storage_nodes"`
	ReplicationCount int          `mapstructure:"replication_count"`
	ConsistentHashing bool        `mapstructure:"consistent_hashing"`
	HealthCheckInterval time.Duration `mapstructure:"health_check_interval"`
}

// DistributedGitService implements distributed git operations
type DistributedGitService struct {
	config         *DistributedConfig
	localService   GitService
	nodes          map[string]*StorageNode
	nodesMutex     sync.RWMutex
	router         *ConsistentHashRouter
	lockManager    *DistributedLockManager
	logger         *logrus.Logger
}

// NewDistributedGitService creates a new distributed git service
func NewDistributedGitService(config *DistributedConfig, localService GitService, logger *logrus.Logger) *DistributedGitService {
	service := &DistributedGitService{
		config:       config,
		localService: localService,
		nodes:        make(map[string]*StorageNode),
		router:       NewConsistentHashRouter(),
		lockManager:  NewDistributedLockManager(),
		logger:       logger,
	}

	// Initialize nodes
	for _, node := range config.StorageNodes {
		service.nodes[node.ID] = &node
		service.router.AddNode(node.ID, node.Weight)
	}

	// Start health check routine
	go service.healthCheckLoop()

	return service
}

// InitRepository initializes a repository with distributed storage
func (d *DistributedGitService) InitRepository(ctx context.Context, repoPath string, bare bool) error {
	if !d.config.Enabled {
		return d.localService.InitRepository(ctx, repoPath, bare)
	}

	// Get target nodes for this repository
	nodes := d.getTargetNodes(repoPath)
	
	// Acquire distributed lock for repository creation
	lockKey := fmt.Sprintf("init:%s", repoPath)
	lock, err := d.lockManager.AcquireLock(ctx, lockKey, 30*time.Second)
	if err != nil {
		return fmt.Errorf("failed to acquire lock for repository init: %w", err)
	}
	defer d.lockManager.ReleaseLock(lock)

	d.logger.WithFields(logrus.Fields{
		"repo_path": repoPath,
		"nodes":     nodes,
		"bare":      bare,
	}).Info("Initializing repository on distributed storage")

	// Initialize on all target nodes
	var lastErr error
	successCount := 0
	for _, nodeID := range nodes {
		if err := d.initRepositoryOnNode(ctx, nodeID, repoPath, bare); err != nil {
			d.logger.WithError(err).WithFields(logrus.Fields{
				"node_id": nodeID,
				"repo_path": repoPath,
			}).Error("Failed to initialize repository on node")
			lastErr = err
		} else {
			successCount++
		}
	}

	// Require at least one successful initialization
	if successCount == 0 {
		return fmt.Errorf("failed to initialize repository on any node: %w", lastErr)
	}

	// If we have replication, require majority success
	requiredSuccess := (len(nodes) / 2) + 1
	if successCount < requiredSuccess {
		return fmt.Errorf("failed to initialize repository on majority of nodes (%d/%d)", successCount, len(nodes))
	}

	return nil
}

// GetCommits retrieves commits with distributed load balancing
func (d *DistributedGitService) GetCommits(ctx context.Context, repoPath string, opts CommitOptions) ([]*Commit, error) {
	if !d.config.Enabled {
		return d.localService.GetCommits(ctx, repoPath, opts)
	}

	// Get primary node for read operations
	nodeID := d.getPrimaryNode(repoPath)
	
	d.logger.WithFields(logrus.Fields{
		"repo_path": repoPath,
		"node_id":   nodeID,
	}).Debug("Getting commits from distributed storage")

	return d.getCommitsFromNode(ctx, nodeID, repoPath, opts)
}

// CreateFile creates a file with distributed replication
func (d *DistributedGitService) CreateFile(ctx context.Context, repoPath string, req CreateFileRequest) (*Commit, error) {
	if !d.config.Enabled {
		return d.localService.CreateFile(ctx, repoPath, req)
	}

	// Acquire distributed lock for write operations
	lockKey := fmt.Sprintf("write:%s", repoPath)
	lock, err := d.lockManager.AcquireLock(ctx, lockKey, 30*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for file creation: %w", err)
	}
	defer d.lockManager.ReleaseLock(lock)

	// Get target nodes for this repository
	nodes := d.getTargetNodes(repoPath)
	
	d.logger.WithFields(logrus.Fields{
		"repo_path": repoPath,
		"file_path": req.Path,
		"nodes":     nodes,
	}).Info("Creating file on distributed storage")

	// Create on primary node first
	primaryNode := nodes[0]
	commit, err := d.createFileOnNode(ctx, primaryNode, repoPath, req)
	if err != nil {
		return nil, fmt.Errorf("failed to create file on primary node %s: %w", primaryNode, err)
	}

	// Replicate to other nodes asynchronously
	for i := 1; i < len(nodes); i++ {
		go func(nodeID string) {
			if err := d.replicateFileCreation(ctx, nodeID, repoPath, req); err != nil {
				d.logger.WithError(err).WithFields(logrus.Fields{
					"node_id": nodeID,
					"repo_path": repoPath,
					"file_path": req.Path,
				}).Error("Failed to replicate file creation")
			}
		}(nodes[i])
	}

	return commit, nil
}

// Helper methods for distributed operations

func (d *DistributedGitService) getTargetNodes(repoPath string) []string {
	if d.config.ConsistentHashing {
		return d.router.GetNodes(repoPath, d.config.ReplicationCount)
	}
	
	// Simple hash-based distribution
	hash := d.hashRepository(repoPath)
	nodeCount := len(d.nodes)
	if nodeCount == 0 {
		return []string{}
	}
	
	replicationCount := d.config.ReplicationCount
	if replicationCount > nodeCount {
		replicationCount = nodeCount
	}
	
	nodes := make([]string, 0, replicationCount)
	startIndex := hash % nodeCount
	
	// Get ordered list of node IDs
	nodeIDs := make([]string, 0, nodeCount)
	d.nodesMutex.RLock()
	for nodeID := range d.nodes {
		nodeIDs = append(nodeIDs, nodeID)
	}
	d.nodesMutex.RUnlock()
	
	// Select nodes starting from hash index
	for i := 0; i < replicationCount; i++ {
		nodeIndex := (startIndex + i) % nodeCount
		nodes = append(nodes, nodeIDs[nodeIndex])
	}
	
	return nodes
}

func (d *DistributedGitService) getPrimaryNode(repoPath string) string {
	nodes := d.getTargetNodes(repoPath)
	if len(nodes) == 0 {
		return ""
	}
	return nodes[0]
}

func (d *DistributedGitService) hashRepository(repoPath string) int {
	hasher := md5.New()
	hasher.Write([]byte(repoPath))
	hashBytes := hasher.Sum(nil)
	hashString := hex.EncodeToString(hashBytes)
	
	// Convert first 8 characters to int
	hashInt, _ := strconv.ParseInt(hashString[:8], 16, 64)
	return int(hashInt)
}

func (d *DistributedGitService) initRepositoryOnNode(ctx context.Context, nodeID, repoPath string, bare bool) error {
	// In a real implementation, this would make RPC calls to the specific node
	// For now, we'll use the local service as a placeholder
	if nodeID == d.config.NodeID {
		return d.localService.InitRepository(ctx, repoPath, bare)
	}
	
	// TODO: Implement RPC call to remote node
	d.logger.WithFields(logrus.Fields{
		"node_id": nodeID,
		"repo_path": repoPath,
	}).Info("Would initialize repository on remote node")
	
	return nil
}

func (d *DistributedGitService) getCommitsFromNode(ctx context.Context, nodeID, repoPath string, opts CommitOptions) ([]*Commit, error) {
	// In a real implementation, this would make RPC calls to the specific node
	if nodeID == d.config.NodeID {
		return d.localService.GetCommits(ctx, repoPath, opts)
	}
	
	// TODO: Implement RPC call to remote node
	d.logger.WithFields(logrus.Fields{
		"node_id": nodeID,
		"repo_path": repoPath,
	}).Info("Would get commits from remote node")
	
	return nil, fmt.Errorf("remote node operations not yet implemented")
}

func (d *DistributedGitService) createFileOnNode(ctx context.Context, nodeID, repoPath string, req CreateFileRequest) (*Commit, error) {
	// In a real implementation, this would make RPC calls to the specific node
	if nodeID == d.config.NodeID {
		return d.localService.CreateFile(ctx, repoPath, req)
	}
	
	// TODO: Implement RPC call to remote node
	d.logger.WithFields(logrus.Fields{
		"node_id": nodeID,
		"repo_path": repoPath,
	}).Info("Would create file on remote node")
	
	return nil, fmt.Errorf("remote node operations not yet implemented")
}

func (d *DistributedGitService) replicateFileCreation(ctx context.Context, nodeID, repoPath string, req CreateFileRequest) error {
	// Asynchronous replication
	_, err := d.createFileOnNode(ctx, nodeID, repoPath, req)
	return err
}

func (d *DistributedGitService) healthCheckLoop() {
	ticker := time.NewTicker(d.config.HealthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			d.performHealthChecks()
		}
	}
}

func (d *DistributedGitService) performHealthChecks() {
	d.nodesMutex.Lock()
	defer d.nodesMutex.Unlock()
	
	for nodeID, node := range d.nodes {
		healthy := d.checkNodeHealth(nodeID)
		if healthy != node.Healthy {
			d.logger.WithFields(logrus.Fields{
				"node_id": nodeID,
				"healthy": healthy,
			}).Info("Node health status changed")
			
			node.Healthy = healthy
			node.LastSeen = time.Now()
			
			// Update router
			if healthy {
				d.router.AddNode(nodeID, node.Weight)
			} else {
				d.router.RemoveNode(nodeID)
			}
		}
	}
}

func (d *DistributedGitService) checkNodeHealth(nodeID string) bool {
	// In a real implementation, this would check node health via RPC
	// For now, assume local node is always healthy
	if nodeID == d.config.NodeID {
		return true
	}
	
	// TODO: Implement health check RPC
	return true
}

// Implement remaining GitService interface methods with distributed logic

func (d *DistributedGitService) CloneRepository(ctx context.Context, sourceURL, destPath string, options CloneOptions) error {
	if !d.config.Enabled {
		return d.localService.CloneRepository(ctx, sourceURL, destPath, options)
	}
	// TODO: Implement distributed clone
	return d.localService.CloneRepository(ctx, sourceURL, destPath, options)
}

func (d *DistributedGitService) DeleteRepository(ctx context.Context, repoPath string) error {
	if !d.config.Enabled {
		return d.localService.DeleteRepository(ctx, repoPath)
	}
	// TODO: Implement distributed delete
	return d.localService.DeleteRepository(ctx, repoPath)
}

func (d *DistributedGitService) GetCommit(ctx context.Context, repoPath, sha string) (*Commit, error) {
	if !d.config.Enabled {
		return d.localService.GetCommit(ctx, repoPath, sha)
	}
	// TODO: Implement distributed get commit
	return d.localService.GetCommit(ctx, repoPath, sha)
}

func (d *DistributedGitService) GetCommitDiff(ctx context.Context, repoPath, fromSHA, toSHA string) (*Diff, error) {
	if !d.config.Enabled {
		return d.localService.GetCommitDiff(ctx, repoPath, fromSHA, toSHA)
	}
	// TODO: Implement distributed diff
	return d.localService.GetCommitDiff(ctx, repoPath, fromSHA, toSHA)
}

func (d *DistributedGitService) GetBranches(ctx context.Context, repoPath string) ([]*Branch, error) {
	if !d.config.Enabled {
		return d.localService.GetBranches(ctx, repoPath)
	}
	// TODO: Implement distributed branches
	return d.localService.GetBranches(ctx, repoPath)
}

func (d *DistributedGitService) GetBranch(ctx context.Context, repoPath, branchName string) (*Branch, error) {
	if !d.config.Enabled {
		return d.localService.GetBranch(ctx, repoPath, branchName)
	}
	// TODO: Implement distributed branch
	return d.localService.GetBranch(ctx, repoPath, branchName)
}

func (d *DistributedGitService) CreateBranch(ctx context.Context, repoPath, branchName, fromRef string) error {
	if !d.config.Enabled {
		return d.localService.CreateBranch(ctx, repoPath, branchName, fromRef)
	}
	// TODO: Implement distributed create branch
	return d.localService.CreateBranch(ctx, repoPath, branchName, fromRef)
}

func (d *DistributedGitService) DeleteBranch(ctx context.Context, repoPath, branchName string) error {
	if !d.config.Enabled {
		return d.localService.DeleteBranch(ctx, repoPath, branchName)
	}
	// TODO: Implement distributed delete branch
	return d.localService.DeleteBranch(ctx, repoPath, branchName)
}

func (d *DistributedGitService) GetTags(ctx context.Context, repoPath string) ([]*Tag, error) {
	if !d.config.Enabled {
		return d.localService.GetTags(ctx, repoPath)
	}
	// TODO: Implement distributed tags
	return d.localService.GetTags(ctx, repoPath)
}

func (d *DistributedGitService) GetTag(ctx context.Context, repoPath, tagName string) (*Tag, error) {
	if !d.config.Enabled {
		return d.localService.GetTag(ctx, repoPath, tagName)
	}
	// TODO: Implement distributed tag
	return d.localService.GetTag(ctx, repoPath, tagName)
}

func (d *DistributedGitService) CreateTag(ctx context.Context, repoPath, tagName, ref, message string) error {
	if !d.config.Enabled {
		return d.localService.CreateTag(ctx, repoPath, tagName, ref, message)
	}
	// TODO: Implement distributed create tag
	return d.localService.CreateTag(ctx, repoPath, tagName, ref, message)
}

func (d *DistributedGitService) DeleteTag(ctx context.Context, repoPath, tagName string) error {
	if !d.config.Enabled {
		return d.localService.DeleteTag(ctx, repoPath, tagName)
	}
	// TODO: Implement distributed delete tag
	return d.localService.DeleteTag(ctx, repoPath, tagName)
}

func (d *DistributedGitService) GetTree(ctx context.Context, repoPath, ref, path string) (*Tree, error) {
	if !d.config.Enabled {
		return d.localService.GetTree(ctx, repoPath, ref, path)
	}
	// TODO: Implement distributed tree
	return d.localService.GetTree(ctx, repoPath, ref, path)
}

func (d *DistributedGitService) GetBlob(ctx context.Context, repoPath, sha string) (*Blob, error) {
	if !d.config.Enabled {
		return d.localService.GetBlob(ctx, repoPath, sha)
	}
	// TODO: Implement distributed blob
	return d.localService.GetBlob(ctx, repoPath, sha)
}

func (d *DistributedGitService) GetFile(ctx context.Context, repoPath, ref, path string) (*File, error) {
	if !d.config.Enabled {
		return d.localService.GetFile(ctx, repoPath, ref, path)
	}
	// TODO: Implement distributed file
	return d.localService.GetFile(ctx, repoPath, ref, path)
}

func (d *DistributedGitService) UpdateFile(ctx context.Context, repoPath string, req UpdateFileRequest) (*Commit, error) {
	if !d.config.Enabled {
		return d.localService.UpdateFile(ctx, repoPath, req)
	}
	// TODO: Implement distributed update file
	return d.localService.UpdateFile(ctx, repoPath, req)
}

func (d *DistributedGitService) DeleteFile(ctx context.Context, repoPath string, req DeleteFileRequest) (*Commit, error) {
	if !d.config.Enabled {
		return d.localService.DeleteFile(ctx, repoPath, req)
	}
	// TODO: Implement distributed delete file
	return d.localService.DeleteFile(ctx, repoPath, req)
}

func (d *DistributedGitService) GetRepositoryInfo(ctx context.Context, repoPath string) (*RepositoryInfo, error) {
	if !d.config.Enabled {
		return d.localService.GetRepositoryInfo(ctx, repoPath)
	}
	// TODO: Implement distributed repo info
	return d.localService.GetRepositoryInfo(ctx, repoPath)
}

func (d *DistributedGitService) GetRepositoryStats(ctx context.Context, repoPath string) (*RepositoryStats, error) {
	if !d.config.Enabled {
		return d.localService.GetRepositoryStats(ctx, repoPath)
	}
	// TODO: Implement distributed repo stats
	return d.localService.GetRepositoryStats(ctx, repoPath)
}

func (d *DistributedGitService) CompareRefs(repoPath, base, head string) (*BranchComparison, error) {
	if !d.config.Enabled {
		return d.localService.CompareRefs(repoPath, base, head)
	}
	// TODO: Implement distributed compare refs
	return d.localService.CompareRefs(repoPath, base, head)
}

func (d *DistributedGitService) CanMerge(repoPath, base, head string) (bool, error) {
	if !d.config.Enabled {
		return d.localService.CanMerge(repoPath, base, head)
	}
	// TODO: Implement distributed can merge
	return d.localService.CanMerge(repoPath, base, head)
}

func (d *DistributedGitService) MergeBranches(repoPath, base, head string, mergeMethod, title, message string) (string, error) {
	if !d.config.Enabled {
		return d.localService.MergeBranches(repoPath, base, head, mergeMethod, title, message)
	}
	// TODO: Implement distributed merge
	return d.localService.MergeBranches(repoPath, base, head, mergeMethod, title, message)
}

func (d *DistributedGitService) GetBranchCommit(repoPath, branch string) (string, error) {
	if !d.config.Enabled {
		return d.localService.GetBranchCommit(repoPath, branch)
	}
	// TODO: Implement distributed branch commit
	return d.localService.GetBranchCommit(repoPath, branch)
}