package services

import (
	"context"
	"fmt"
	"time"

	"github.com/a5c-ai/hub/internal/config"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisService handles Redis connections and operations
type RedisService struct {
	client *redis.Client
	config config.Redis
	logger *logrus.Logger
}

// NewRedisService creates a new Redis service
func NewRedisService(cfg config.Redis, logger *logrus.Logger) (*RedisService, error) {
	if !cfg.Enabled {
		logger.Info("Redis is disabled, Redis service will not be initialized")
		return &RedisService{
			config: cfg,
			logger: logger,
		}, nil
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:       fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:   cfg.Password,
		DB:         cfg.DB,
		MaxRetries: cfg.MaxRetries,
		PoolSize:   cfg.PoolSize,
		// Connection and read timeouts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		// Pool timeouts
		PoolTimeout: 4 * time.Second,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"host": cfg.Host,
		"port": cfg.Port,
		"db":   cfg.DB,
	}).Info("Successfully connected to Redis")

	return &RedisService{
		client: rdb,
		config: cfg,
		logger: logger,
	}, nil
}

// GetClient returns the Redis client
func (s *RedisService) GetClient() *redis.Client {
	return s.client
}

// IsEnabled returns whether Redis is enabled
func (s *RedisService) IsEnabled() bool {
	return s.config.Enabled && s.client != nil
}

// HealthCheck performs a health check on the Redis connection
func (s *RedisService) HealthCheck(ctx context.Context) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	_, err := s.client.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

// Close closes the Redis connection
func (s *RedisService) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

// Queue Operations

// RPush adds items to the right (end) of a list
func (s *RedisService) RPush(ctx context.Context, key string, values ...interface{}) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	return s.client.RPush(ctx, key, values...).Err()
}

// LPop removes and returns the first item from a list
func (s *RedisService) LPop(ctx context.Context, key string) (string, error) {
	if !s.IsEnabled() {
		return "", fmt.Errorf("Redis is not enabled")
	}

	return s.client.LPop(ctx, key).Result()
}

// BLPop is a blocking version of LPop that waits for items
func (s *RedisService) BLPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("Redis is not enabled")
	}

	return s.client.BLPop(ctx, timeout, keys...).Result()
}

// LLen returns the length of a list
func (s *RedisService) LLen(ctx context.Context, key string) (int64, error) {
	if !s.IsEnabled() {
		return 0, fmt.Errorf("Redis is not enabled")
	}

	return s.client.LLen(ctx, key).Result()
}

// LRange returns a range of items from a list
func (s *RedisService) LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("Redis is not enabled")
	}

	return s.client.LRange(ctx, key, start, stop).Result()
}

// LRem removes items from a list
func (s *RedisService) LRem(ctx context.Context, key string, count int64, value interface{}) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	return s.client.LRem(ctx, key, count, value).Err()
}

// Sorted Set Operations for Priority Queues

// ZAdd adds items to a sorted set with scores (for priority)
func (s *RedisService) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	return s.client.ZAdd(ctx, key, members...).Err()
}

// ZPopMax removes and returns the highest scored item from a sorted set
func (s *RedisService) ZPopMax(ctx context.Context, key string, count ...int64) ([]redis.Z, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("Redis is not enabled")
	}

	return s.client.ZPopMax(ctx, key, count...).Result()
}

// BZPopMax is a blocking version of ZPopMax
func (s *RedisService) BZPopMax(ctx context.Context, timeout time.Duration, keys ...string) (*redis.ZWithKey, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("Redis is not enabled")
	}

	return s.client.BZPopMax(ctx, timeout, keys...).Result()
}

// ZCard returns the number of items in a sorted set
func (s *RedisService) ZCard(ctx context.Context, key string) (int64, error) {
	if !s.IsEnabled() {
		return 0, fmt.Errorf("Redis is not enabled")
	}

	return s.client.ZCard(ctx, key).Result()
}

// ZRange returns a range of items from a sorted set
func (s *RedisService) ZRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("Redis is not enabled")
	}

	return s.client.ZRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores returns a range of items with scores from a sorted set
func (s *RedisService) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]redis.Z, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("Redis is not enabled")
	}

	return s.client.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRem removes items from a sorted set
func (s *RedisService) ZRem(ctx context.Context, key string, members ...interface{}) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	return s.client.ZRem(ctx, key, members...).Err()
}

// Hash Operations

// HSet sets fields in a hash
func (s *RedisService) HSet(ctx context.Context, key string, values ...interface{}) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	return s.client.HSet(ctx, key, values...).Err()
}

// HGet gets a field from a hash
func (s *RedisService) HGet(ctx context.Context, key, field string) (string, error) {
	if !s.IsEnabled() {
		return "", fmt.Errorf("Redis is not enabled")
	}

	return s.client.HGet(ctx, key, field).Result()
}

// HGetAll gets all fields from a hash
func (s *RedisService) HGetAll(ctx context.Context, key string) (map[string]string, error) {
	if !s.IsEnabled() {
		return nil, fmt.Errorf("Redis is not enabled")
	}

	return s.client.HGetAll(ctx, key).Result()
}

// HDel deletes fields from a hash
func (s *RedisService) HDel(ctx context.Context, key string, fields ...string) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	return s.client.HDel(ctx, key, fields...).Err()
}

// General Operations

// Set sets a key-value pair
func (s *RedisService) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	return s.client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value by key
func (s *RedisService) Get(ctx context.Context, key string) (string, error) {
	if !s.IsEnabled() {
		return "", fmt.Errorf("Redis is not enabled")
	}

	return s.client.Get(ctx, key).Result()
}

// Del deletes keys
func (s *RedisService) Del(ctx context.Context, keys ...string) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	return s.client.Del(ctx, keys...).Err()
}

// Exists checks if keys exist
func (s *RedisService) Exists(ctx context.Context, keys ...string) (int64, error) {
	if !s.IsEnabled() {
		return 0, fmt.Errorf("Redis is not enabled")
	}

	return s.client.Exists(ctx, keys...).Result()
}

// Expire sets expiration for a key
func (s *RedisService) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if !s.IsEnabled() {
		return fmt.Errorf("Redis is not enabled")
	}

	return s.client.Expire(ctx, key, expiration).Err()
}

// Transaction Operations

// TxPipeline creates a new transaction pipeline
func (s *RedisService) TxPipeline() redis.Pipeliner {
	if !s.IsEnabled() {
		return nil
	}

	return s.client.TxPipeline()
}