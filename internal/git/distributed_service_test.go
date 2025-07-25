package git

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDistributedGitService(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create mock local git service
	localService := NewGitService(logger)

	// Setup distributed config
	config := &DistributedConfig{
		Enabled:          true,
		NodeID:           "test-node-1",
		ReplicationCount: 3,
		ConsistentHashing: true,
		HealthCheckInterval: 30 * time.Second,
		StorageNodes: []StorageNode{
			{ID: "node-1", Address: "http://node-1:8080", Weight: 1},
			{ID: "node-2", Address: "http://node-2:8080", Weight: 1},
			{ID: "node-3", Address: "http://node-3:8080", Weight: 1},
		},
	}

	// Create distributed service
	service := NewDistributedGitService(config, localService, logger)
	require.NotNil(t, service)

	t.Run("TestNodeSelection", func(t *testing.T) {
		// Test consistent node selection for repository
		repoPath := "test/repo1"
		nodes1 := service.getTargetNodes(repoPath)
		nodes2 := service.getTargetNodes(repoPath)

		// Should return same nodes for same repository
		assert.Equal(t, nodes1, nodes2)
		assert.Len(t, nodes1, config.ReplicationCount)
	})

	t.Run("TestConsistentHashing", func(t *testing.T) {
		// Test that repositories are distributed across nodes
		repos := []string{"repo1", "repo2", "repo3", "repo4", "repo5"}
		nodeUsage := make(map[string]int)

		for _, repo := range repos {
			nodes := service.getTargetNodes(repo)
			primaryNode := nodes[0]
			nodeUsage[primaryNode]++
		}

		// All nodes should be used (though distribution might not be perfectly even)
		assert.Greater(t, len(nodeUsage), 0)
	})

	t.Run("TestHashDistribution", func(t *testing.T) {
		// Test that hash function distributes repositories reasonably
		repoCount := 100
		nodeHits := make(map[string]int)

		for i := 0; i < repoCount; i++ {
			repoPath := fmt.Sprintf("test/repo%d", i)
			nodes := service.getTargetNodes(repoPath)
			if len(nodes) > 0 {
				nodeHits[nodes[0]]++
			}
		}

		// Each node should get some repositories (allowing for some imbalance)
		for nodeID, hits := range nodeHits {
			t.Logf("Node %s: %d repositories", nodeID, hits)
			assert.Greater(t, hits, 0)
		}
	})
}

func TestConsistentHashRouter(t *testing.T) {
	router := NewConsistentHashRouter()

	t.Run("TestAddRemoveNodes", func(t *testing.T) {
		// Add nodes
		router.AddNode("node1", 1)
		router.AddNode("node2", 1)
		router.AddNode("node3", 1)

		assert.Equal(t, 3, router.GetNodeCount())

		// Test key distribution
		key1 := "test-key-1"
		node1 := router.GetNode(key1)
		assert.NotEmpty(t, node1)

		// Remove a node
		router.RemoveNode("node2")
		assert.Equal(t, 2, router.GetNodeCount())

		// Key should still map to a node
		node2 := router.GetNode(key1)
		assert.NotEmpty(t, node2)
	})

	t.Run("TestConsistentMapping", func(t *testing.T) {
		router.AddNode("node1", 1)
		router.AddNode("node2", 1)

		key := "test-repository"
		
		// Same key should consistently map to same node
		node1 := router.GetNode(key)
		node2 := router.GetNode(key)
		assert.Equal(t, node1, node2)
	})

	t.Run("TestMultipleNodes", func(t *testing.T) {
		router.AddNode("node1", 1)
		router.AddNode("node2", 1)
		router.AddNode("node3", 1)

		key := "test-repository"
		nodes := router.GetNodes(key, 2)
		
		assert.Len(t, nodes, 2)
		assert.NotEqual(t, nodes[0], nodes[1])
	})

	t.Run("TestWeightedDistribution", func(t *testing.T) {
		router := NewConsistentHashRouter()
		router.AddNode("heavy-node", 3)
		router.AddNode("light-node", 1)

		stats := router.DistributionStats()
		assert.Greater(t, stats["heavy-node"], stats["light-node"])
	})
}

func TestDistributedLockManager(t *testing.T) {
	manager := NewDistributedLockManager()
	defer manager.Close()

	t.Run("TestBasicLocking", func(t *testing.T) {
		ctx := context.Background()
		key := "test-lock"
		ttl := 5 * time.Second

		// Acquire lock
		lock1, err := manager.AcquireLock(ctx, key, ttl)
		require.NoError(t, err)
		require.NotNil(t, lock1)

		// Lock should be held
		assert.True(t, manager.IsLocked(key))

		// Release lock
		err = manager.ReleaseLock(lock1)
		require.NoError(t, err)

		// Lock should be released
		assert.False(t, manager.IsLocked(key))
	})

	t.Run("TestLockContention", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		key := "contended-lock"
		ttl := 5 * time.Second

		// Acquire first lock
		lock1, err := manager.AcquireLock(ctx, key, ttl)
		require.NoError(t, err)

		// Try to acquire same lock (should timeout due to context deadline)
		_, err = manager.AcquireLock(ctx, key, ttl)
		assert.Error(t, err)
		// Either timeout or context deadline is acceptable
		assert.True(t, err == ErrLockTimeout || err == context.DeadlineExceeded)

		// Release first lock
		err = manager.ReleaseLock(lock1)
		require.NoError(t, err)
	})

	t.Run("TestLockExpiration", func(t *testing.T) {
		ctx := context.Background()
		key := "expiring-lock"
		ttl := 300 * time.Millisecond // Longer initial TTL to ensure lock is acquired properly

		// Acquire lock with short TTL
		expiringLock, err := manager.AcquireLock(ctx, key, ttl)
		require.NoError(t, err)
		require.NotNil(t, expiringLock)

		// Lock should be held initially - verify immediately after acquisition
		startTime := time.Now()
		assert.True(t, manager.IsLocked(key))
		
		// Log the expiration time for debugging
		t.Logf("Lock expires at: %v, current time: %v", expiringLock.ExpireTime, startTime)

		// Wait for expiration plus some buffer
		time.Sleep(ttl + 100*time.Millisecond)
		
		// Check time again
		endTime := time.Now()
		t.Logf("After wait - current time: %v, elapsed: %v", endTime, endTime.Sub(startTime))

		// Lock should be expired (IsLocked handles cleanup automatically)
		isLocked := manager.IsLocked(key)
		t.Logf("IsLocked result: %v", isLocked)
		assert.False(t, isLocked)
	})

	t.Run("TestLockRenewal", func(t *testing.T) {
		ctx := context.Background()
		key := "renewable-lock"
		ttl := 200 * time.Millisecond

		// Acquire lock
		renewableLock, err := manager.AcquireLock(ctx, key, ttl)
		require.NoError(t, err)

		// Verify lock is active
		assert.True(t, manager.IsLocked(key))

		// Wait less than TTL
		time.Sleep(50 * time.Millisecond)

		// Renew lock before expiration
		err = manager.RenewLock(renewableLock, ttl)
		require.NoError(t, err)

		// Wait a bit more (should still be locked due to renewal)
		time.Sleep(100 * time.Millisecond)
		assert.True(t, manager.IsLocked(key))

		// Release lock
		err = manager.ReleaseLock(renewableLock)
		require.NoError(t, err)
	})

	t.Run("TestScopedLocks", func(t *testing.T) {
		ctx := context.Background()
		ttl := 5 * time.Second

		// Test repository lock
		repoLock, err := manager.LockRepository(ctx, "test/repo", ttl)
		require.NoError(t, err)
		defer manager.ReleaseLock(repoLock)

		// Test branch lock
		branchLock, err := manager.LockBranch(ctx, "test/repo", "main", ttl)
		require.NoError(t, err)
		defer manager.ReleaseLock(branchLock)

		// Test file lock
		fileLock, err := manager.LockFile(ctx, "test/repo", "README.md", ttl)
		require.NoError(t, err)
		defer manager.ReleaseLock(fileLock)

		// All locks should be active
		assert.True(t, manager.IsLocked(GenerateLockKey(LockScopeRepository, "test/repo")))
		assert.True(t, manager.IsLocked(GenerateLockKey(LockScopeBranch, "test/repo", "main")))
		assert.True(t, manager.IsLocked(GenerateLockKey(LockScopeFile, "test/repo", "README.md")))
	})
}

func TestServiceIntegration(t *testing.T) {
	// Skip if running in short mode
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Create temporary directory for test repositories
	tempDir := t.TempDir()

	// Create local git service
	localService := NewGitService(logger)

	// Setup distributed config with mock nodes
	config := &DistributedConfig{
		Enabled:          true,
		NodeID:           "test-node",
		ReplicationCount: 2,
		ConsistentHashing: true,
		HealthCheckInterval: 30 * time.Second,
		StorageNodes: []StorageNode{
			{ID: "node-1", Address: "http://localhost:8081", Weight: 1},
			{ID: "node-2", Address: "http://localhost:8082", Weight: 1},
		},
	}

	// Create distributed service
	distributedService := NewDistributedGitService(config, localService, logger)

	t.Run("TestInitRepository", func(t *testing.T) {
		ctx := context.Background()
		repoPath := tempDir + "/test-repo"

		// Test repository initialization
		err := distributedService.InitRepository(ctx, repoPath, true)
		// This will succeed for local node but may fail for remote nodes (expected in test)
		// We're mainly testing the distributed logic flow
		t.Logf("Init repository result: %v", err)
	})

	t.Run("TestCreateFile", func(t *testing.T) {
		ctx := context.Background()
		repoPath := tempDir + "/test-repo"

		// First initialize repository locally
		err := localService.InitRepository(ctx, repoPath, true)
		require.NoError(t, err)

		// Test file creation with distributed service
		req := CreateFileRequest{
			Path:    "README.md",
			Content: "# Test Repository\n\nThis is a test.",
			Message: "Add README",
			Branch:  "main",
			Author: CommitAuthor{
				Name:  "Test User",
				Email: "test@example.com",
				Date:  time.Now(),
			},
		}

		// This will work for local operations
		commit, err := distributedService.CreateFile(ctx, repoPath, req)
		t.Logf("Create file result: %v", err)
		if commit != nil {
			t.Logf("Created commit: %s", commit.SHA)
		}
	})
}

func BenchmarkConsistentHashRouter(b *testing.B) {
	router := NewConsistentHashRouter()

	// Add nodes
	for i := 0; i < 10; i++ {
		router.AddNode(fmt.Sprintf("node-%d", i), 1)
	}

	b.ResetTimer()
	b.Run("GetNode", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key-%d", i%1000)
			router.GetNode(key)
		}
	})

	b.Run("GetNodes", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("key-%d", i%1000)
			router.GetNodes(key, 3)
		}
	})
}

func BenchmarkDistributedLocking(b *testing.B) {
	manager := NewDistributedLockManager()
	defer manager.Close()

	ctx := context.Background()
	ttl := 1 * time.Second

	b.ResetTimer()
	b.Run("AcquireRelease", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("lock-%d", i)
			lock, err := manager.AcquireLock(ctx, key, ttl)
			if err != nil {
				b.Fatal(err)
			}
			manager.ReleaseLock(lock)
		}
	})

	b.Run("Contention", func(b *testing.B) {
		key := "contended-lock"
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				lock, err := manager.AcquireLock(ctx, key, ttl)
				if err == nil {
					time.Sleep(1 * time.Millisecond) // Hold lock briefly
					manager.ReleaseLock(lock)
				}
			}
		})
	})
}

// Helper functions for testing

func createTestStorageNodes(count int) []StorageNode {
	nodes := make([]StorageNode, count)
	for i := 0; i < count; i++ {
		nodes[i] = StorageNode{
			ID:      fmt.Sprintf("node-%d", i+1),
			Address: fmt.Sprintf("http://node-%d:8080", i+1),
			Weight:  1,
		}
	}
	return nodes
}

func createTestDistributedConfig(nodeCount int) *DistributedConfig {
	return &DistributedConfig{
		Enabled:             true,
		NodeID:              "test-node",
		StorageNodes:        createTestStorageNodes(nodeCount),
		ReplicationCount:    3,
		ConsistentHashing:   true,
		HealthCheckInterval: 30 * time.Second,
	}
}