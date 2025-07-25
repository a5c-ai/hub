package git

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrLockTimeout     = errors.New("lock acquisition timeout")
	ErrLockNotFound    = errors.New("lock not found")
	ErrLockNotOwned    = errors.New("lock not owned by this client")
	ErrLockExpired     = errors.New("lock has expired")
)

// DistributedLock represents a distributed lock
type DistributedLock struct {
	Key        string    `json:"key"`
	Value      string    `json:"value"`
	ExpireTime time.Time `json:"expire_time"`
	NodeID     string    `json:"node_id"`
}

// DistributedLockManager manages distributed locks across git storage nodes
type DistributedLockManager struct {
	locks      map[string]*DistributedLock
	locksMutex sync.RWMutex
	nodeID     string
	cleanup    chan struct{}
}

// NewDistributedLockManager creates a new distributed lock manager
func NewDistributedLockManager() *DistributedLockManager {
	manager := &DistributedLockManager{
		locks:   make(map[string]*DistributedLock),
		nodeID:  generateNodeID(),
		cleanup: make(chan struct{}),
	}
	
	// Start cleanup routine for expired locks
	go manager.cleanupExpiredLocks()
	
	return manager
}

// AcquireLock attempts to acquire a distributed lock
func (dlm *DistributedLockManager) AcquireLock(ctx context.Context, key string, ttl time.Duration) (*DistributedLock, error) {
	lockValue := uuid.New().String()
	expireTime := time.Now().Add(ttl)
	
	lock := &DistributedLock{
		Key:        key,
		Value:      lockValue,
		ExpireTime: expireTime,
		NodeID:     dlm.nodeID,
	}
	
	// Try to acquire lock with timeout
	timeout := time.After(30 * time.Second) // Default timeout for lock acquisition
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-timeout:
			return nil, ErrLockTimeout
		case <-ticker.C:
			if dlm.tryAcquireLock(lock) {
				return lock, nil
			}
		}
	}
}

// ReleaseLock releases a distributed lock
func (dlm *DistributedLockManager) ReleaseLock(lock *DistributedLock) error {
	if lock == nil {
		return ErrLockNotFound
	}
	
	dlm.locksMutex.Lock()
	defer dlm.locksMutex.Unlock()
	
	existingLock, exists := dlm.locks[lock.Key]
	if !exists {
		return ErrLockNotFound
	}
	
	// Verify lock ownership
	if existingLock.Value != lock.Value || existingLock.NodeID != lock.NodeID {
		return ErrLockNotOwned
	}
	
	// Check if lock has expired
	if time.Now().After(existingLock.ExpireTime) {
		delete(dlm.locks, lock.Key)
		return ErrLockExpired
	}
	
	delete(dlm.locks, lock.Key)
	return nil
}

// RenewLock extends the TTL of an existing lock
func (dlm *DistributedLockManager) RenewLock(lock *DistributedLock, ttl time.Duration) error {
	if lock == nil {
		return ErrLockNotFound
	}
	
	dlm.locksMutex.Lock()
	defer dlm.locksMutex.Unlock()
	
	existingLock, exists := dlm.locks[lock.Key]
	if !exists {
		return ErrLockNotFound
	}
	
	// Verify lock ownership
	if existingLock.Value != lock.Value || existingLock.NodeID != lock.NodeID {
		return ErrLockNotOwned
	}
	
	// Check if lock has expired
	if time.Now().After(existingLock.ExpireTime) {
		delete(dlm.locks, lock.Key)
		return ErrLockExpired
	}
	
	// Extend expiration time
	existingLock.ExpireTime = time.Now().Add(ttl)
	lock.ExpireTime = existingLock.ExpireTime
	
	return nil
}

// IsLocked checks if a key is currently locked
func (dlm *DistributedLockManager) IsLocked(key string) bool {
	dlm.locksMutex.Lock()
	defer dlm.locksMutex.Unlock()
	
	lock, exists := dlm.locks[key]
	if !exists {
		return false
	}
	
	// Check if lock has expired
	if time.Now().After(lock.ExpireTime) {
		// Remove expired lock immediately
		delete(dlm.locks, key)
		return false
	}
	
	return true
}

// GetLockInfo returns information about a lock
func (dlm *DistributedLockManager) GetLockInfo(key string) (*DistributedLock, error) {
	dlm.locksMutex.RLock()
	defer dlm.locksMutex.RUnlock()
	
	lock, exists := dlm.locks[key]
	if !exists {
		return nil, ErrLockNotFound
	}
	
	// Check if lock has expired
	if time.Now().After(lock.ExpireTime) {
		return nil, ErrLockExpired
	}
	
	// Return a copy to prevent external modification
	return &DistributedLock{
		Key:        lock.Key,
		Value:      lock.Value,
		ExpireTime: lock.ExpireTime,
		NodeID:     lock.NodeID,
	}, nil
}

// ListLocks returns all active locks
func (dlm *DistributedLockManager) ListLocks() []*DistributedLock {
	dlm.locksMutex.RLock()
	defer dlm.locksMutex.RUnlock()
	
	locks := make([]*DistributedLock, 0, len(dlm.locks))
	now := time.Now()
	
	for _, lock := range dlm.locks {
		if now.Before(lock.ExpireTime) {
			locks = append(locks, &DistributedLock{
				Key:        lock.Key,
				Value:      lock.Value,
				ExpireTime: lock.ExpireTime,
				NodeID:     lock.NodeID,
			})
		}
	}
	
	return locks
}

// ForceReleaseLock forcibly releases a lock (admin operation)
func (dlm *DistributedLockManager) ForceReleaseLock(key string) error {
	dlm.locksMutex.Lock()
	defer dlm.locksMutex.Unlock()
	
	_, exists := dlm.locks[key]
	if !exists {
		return ErrLockNotFound
	}
	
	delete(dlm.locks, key)
	return nil
}

// Close shuts down the lock manager
func (dlm *DistributedLockManager) Close() {
	close(dlm.cleanup)
}

// Private methods

func (dlm *DistributedLockManager) tryAcquireLock(lock *DistributedLock) bool {
	dlm.locksMutex.Lock()
	defer dlm.locksMutex.Unlock()
	
	existingLock, exists := dlm.locks[lock.Key]
	if exists {
		// Check if existing lock has expired
		if time.Now().After(existingLock.ExpireTime) {
			// Lock has expired, remove it and acquire new one
			delete(dlm.locks, lock.Key)
		} else {
			// Lock is still active
			return false
		}
	}
	
	// Acquire the lock
	dlm.locks[lock.Key] = lock
	return true
}

func (dlm *DistributedLockManager) cleanupExpiredLocks() {
	ticker := time.NewTicker(30 * time.Second) // Cleanup every 30 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-dlm.cleanup:
			return
		case <-ticker.C:
			dlm.performCleanup()
		}
	}
}

func (dlm *DistributedLockManager) performCleanup() {
	dlm.locksMutex.Lock()
	defer dlm.locksMutex.Unlock()
	
	now := time.Now()
	expiredKeys := make([]string, 0)
	
	for key, lock := range dlm.locks {
		if now.After(lock.ExpireTime) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	for _, key := range expiredKeys {
		delete(dlm.locks, key)
	}
	
	if len(expiredKeys) > 0 {
		fmt.Printf("Cleaned up %d expired locks\n", len(expiredKeys))
	}
}

func generateNodeID() string {
	return uuid.New().String()
}

// LockScope represents different scopes for locking
type LockScope string

const (
	LockScopeRepository LockScope = "repo"
	LockScopeBranch     LockScope = "branch"
	LockScopeFile       LockScope = "file"
	LockScopeGlobal     LockScope = "global"
)

// GenerateLockKey creates a standardized lock key
func GenerateLockKey(scope LockScope, parts ...string) string {
	key := string(scope)
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// LockWithScope is a convenience method for acquiring scoped locks
func (dlm *DistributedLockManager) LockWithScope(ctx context.Context, scope LockScope, ttl time.Duration, parts ...string) (*DistributedLock, error) {
	key := GenerateLockKey(scope, parts...)
	return dlm.AcquireLock(ctx, key, ttl)
}

// Repository-specific lock helpers

// LockRepository acquires a repository-level lock
func (dlm *DistributedLockManager) LockRepository(ctx context.Context, repoPath string, ttl time.Duration) (*DistributedLock, error) {
	return dlm.LockWithScope(ctx, LockScopeRepository, ttl, repoPath)
}

// LockBranch acquires a branch-level lock
func (dlm *DistributedLockManager) LockBranch(ctx context.Context, repoPath, branchName string, ttl time.Duration) (*DistributedLock, error) {
	return dlm.LockWithScope(ctx, LockScopeBranch, ttl, repoPath, branchName)
}

// LockFile acquires a file-level lock
func (dlm *DistributedLockManager) LockFile(ctx context.Context, repoPath, filePath string, ttl time.Duration) (*DistributedLock, error) {
	return dlm.LockWithScope(ctx, LockScopeFile, ttl, repoPath, filePath)
}

// LockGlobal acquires a global lock
func (dlm *DistributedLockManager) LockGlobal(ctx context.Context, operation string, ttl time.Duration) (*DistributedLock, error) {
	return dlm.LockWithScope(ctx, LockScopeGlobal, ttl, operation)
}