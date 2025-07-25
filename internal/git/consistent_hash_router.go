package git

import (
	"crypto/md5"
	"sort"
	"strconv"
	"sync"
)

// ConsistentHashRouter implements consistent hashing for repository distribution
type ConsistentHashRouter struct {
	hashRing    map[uint32]string
	nodes       map[string]int
	sortedHashes []uint32
	virtualNodes int
	mutex        sync.RWMutex
}

// NewConsistentHashRouter creates a new consistent hash router
func NewConsistentHashRouter() *ConsistentHashRouter {
	return &ConsistentHashRouter{
		hashRing:     make(map[uint32]string),
		nodes:        make(map[string]int),
		virtualNodes: 100, // Number of virtual nodes per physical node
	}
}

// AddNode adds a node to the hash ring
func (chr *ConsistentHashRouter) AddNode(nodeID string, weight int) {
	chr.mutex.Lock()
	defer chr.mutex.Unlock()

	chr.nodes[nodeID] = weight
	
	// Add virtual nodes based on weight
	virtualCount := chr.virtualNodes * weight
	for i := 0; i < virtualCount; i++ {
		virtualNodeKey := chr.getVirtualNodeKey(nodeID, i)
		hash := chr.hashKey(virtualNodeKey)
		chr.hashRing[hash] = nodeID
	}
	
	chr.updateSortedHashes()
}

// RemoveNode removes a node from the hash ring
func (chr *ConsistentHashRouter) RemoveNode(nodeID string) {
	chr.mutex.Lock()
	defer chr.mutex.Unlock()

	weight, exists := chr.nodes[nodeID]
	if !exists {
		return
	}
	
	delete(chr.nodes, nodeID)
	
	// Remove virtual nodes
	virtualCount := chr.virtualNodes * weight
	for i := 0; i < virtualCount; i++ {
		virtualNodeKey := chr.getVirtualNodeKey(nodeID, i)
		hash := chr.hashKey(virtualNodeKey)
		delete(chr.hashRing, hash)
	}
	
	chr.updateSortedHashes()
}

// GetNode returns the node responsible for the given key
func (chr *ConsistentHashRouter) GetNode(key string) string {
	chr.mutex.RLock()
	defer chr.mutex.RUnlock()

	if len(chr.hashRing) == 0 {
		return ""
	}

	hash := chr.hashKey(key)
	idx := chr.searchRing(hash)
	return chr.hashRing[chr.sortedHashes[idx]]
}

// GetNodes returns multiple nodes for replication
func (chr *ConsistentHashRouter) GetNodes(key string, count int) []string {
	chr.mutex.RLock()
	defer chr.mutex.RUnlock()

	if len(chr.hashRing) == 0 || count <= 0 {
		return []string{}
	}

	hash := chr.hashKey(key)
	idx := chr.searchRing(hash)
	
	nodes := make([]string, 0, count)
	seen := make(map[string]bool)
	
	// Start from the found position and walk around the ring
	for i := 0; i < len(chr.sortedHashes) && len(nodes) < count; i++ {
		ringIdx := (idx + i) % len(chr.sortedHashes)
		nodeID := chr.hashRing[chr.sortedHashes[ringIdx]]
		
		if !seen[nodeID] {
			nodes = append(nodes, nodeID)
			seen[nodeID] = true
		}
	}
	
	return nodes
}

// GetAllNodes returns all active nodes
func (chr *ConsistentHashRouter) GetAllNodes() []string {
	chr.mutex.RLock()
	defer chr.mutex.RUnlock()

	nodes := make([]string, 0, len(chr.nodes))
	for nodeID := range chr.nodes {
		nodes = append(nodes, nodeID)
	}
	
	return nodes
}

// GetNodeCount returns the number of active nodes
func (chr *ConsistentHashRouter) GetNodeCount() int {
	chr.mutex.RLock()
	defer chr.mutex.RUnlock()
	
	return len(chr.nodes)
}

// Helper methods

func (chr *ConsistentHashRouter) getVirtualNodeKey(nodeID string, index int) string {
	return nodeID + ":" + strconv.Itoa(index)
}

func (chr *ConsistentHashRouter) hashKey(key string) uint32 {
	hasher := md5.New()
	hasher.Write([]byte(key))
	hashBytes := hasher.Sum(nil)
	
	// Use first 4 bytes for uint32 hash
	hash := uint32(hashBytes[0])<<24 |
		uint32(hashBytes[1])<<16 |
		uint32(hashBytes[2])<<8 |
		uint32(hashBytes[3])
	
	return hash
}

func (chr *ConsistentHashRouter) updateSortedHashes() {
	chr.sortedHashes = make([]uint32, 0, len(chr.hashRing))
	for hash := range chr.hashRing {
		chr.sortedHashes = append(chr.sortedHashes, hash)
	}
	sort.Slice(chr.sortedHashes, func(i, j int) bool {
		return chr.sortedHashes[i] < chr.sortedHashes[j]
	})
}

func (chr *ConsistentHashRouter) searchRing(hash uint32) int {
	// Binary search for the first hash >= target hash
	idx := sort.Search(len(chr.sortedHashes), func(i int) bool {
		return chr.sortedHashes[i] >= hash
	})
	
	// If we went past the end, wrap around to the beginning
	if idx == len(chr.sortedHashes) {
		idx = 0
	}
	
	return idx
}

// DistributionStats returns statistics about key distribution
func (chr *ConsistentHashRouter) DistributionStats() map[string]int {
	chr.mutex.RLock()
	defer chr.mutex.RUnlock()

	stats := make(map[string]int)
	for nodeID := range chr.nodes {
		stats[nodeID] = 0
	}
	
	// Count virtual nodes per physical node
	for _, nodeID := range chr.hashRing {
		stats[nodeID]++
	}
	
	return stats
}

// RebalanceNodes redistributes virtual nodes to maintain balance
func (chr *ConsistentHashRouter) RebalanceNodes() {
	chr.mutex.Lock()
	defer chr.mutex.Unlock()

	// Clear existing ring
	chr.hashRing = make(map[uint32]string)
	
	// Re-add all nodes
	for nodeID, weight := range chr.nodes {
		virtualCount := chr.virtualNodes * weight
		for i := 0; i < virtualCount; i++ {
			virtualNodeKey := chr.getVirtualNodeKey(nodeID, i)
			hash := chr.hashKey(virtualNodeKey)
			chr.hashRing[hash] = nodeID
		}
	}
	
	chr.updateSortedHashes()
}