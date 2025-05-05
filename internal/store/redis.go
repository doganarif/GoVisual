package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/go-redis/redis/v8"
)

// RedisStore implements the Store interface with Redis as backend
type RedisStore struct {
	client    *redis.Client
	keyPrefix string
	capacity  int
	ttl       time.Duration
	ctx       context.Context
}

// NewRedisStore creates a new Redis-backed store
func NewRedisStore(connStr string, capacity int, ttlSeconds int) (*RedisStore, error) {
	if capacity <= 0 {
		capacity = 100
	}

	if ttlSeconds <= 0 {
		ttlSeconds = 86400 // Default to 24 hours
	}

	// Parse the Redis connection string
	opts, err := redis.ParseURL(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis connection string: %w", err)
	}

	// Create Redis client
	client := redis.NewClient(opts)
	ctx := context.Background()

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{
		client:    client,
		keyPrefix: "govisual:",
		capacity:  capacity,
		ttl:       time.Duration(ttlSeconds) * time.Second,
		ctx:       ctx,
	}, nil
}

// Add adds a new request log to the store
func (s *RedisStore) Add(reqLog *model.RequestLog) {
	// Convert log to JSON
	data, err := json.Marshal(reqLog)
	if err != nil {
		reqLog.Error = fmt.Sprintf("Failed to marshal log: %v", err)
		return
	}

	// Key for the log
	key := s.keyPrefix + reqLog.ID

	// Store log as a JSON string
	if err := s.client.Set(s.ctx, key, data, s.ttl).Err(); err != nil {
		log.Printf("Failed to store log in Redis: %v", err)
		return
	}

	// Add to sorted set for time-ordered access
	score := float64(reqLog.Timestamp.UnixNano())
	if err := s.client.ZAdd(s.ctx, s.keyPrefix+"logs", &redis.Z{
		Score:  score,
		Member: reqLog.ID,
	}).Err(); err != nil {
		log.Printf("Failed to add log ID to sorted set: %v", err)
	}

	// Clean up old logs
	s.cleanup()
}

// cleanup removes old logs to maintain the capacity limit
func (s *RedisStore) cleanup() {
	// Get the current number of logs
	count, err := s.client.ZCard(s.ctx, s.keyPrefix+"logs").Result()
	if err != nil {
		log.Printf("Failed to count logs: %v", err)
		return
	}

	if count <= int64(s.capacity) {
		return
	}

	// Get the oldest log IDs that exceed capacity
	oldestIDs, err := s.client.ZRange(s.ctx, s.keyPrefix+"logs", 0, count-int64(s.capacity)-1).Result()
	if err != nil {
		log.Printf("Failed to get oldest log IDs: %v", err)
		return
	}

	if len(oldestIDs) == 0 {
		return
	}

	// Create a pipeline for batch operations
	pipe := s.client.Pipeline()

	// Remove from sorted set
	pipe.ZRem(s.ctx, s.keyPrefix+"logs", oldestIDs)

	// Remove each log
	for _, id := range oldestIDs {
		pipe.Del(s.ctx, s.keyPrefix+id)
	}

	// Execute pipeline
	if _, err := pipe.Exec(s.ctx); err != nil {
		log.Printf("Failed to clean up old logs: %v", err)
	}
}

// Get retrieves a specific request log by its ID
func (s *RedisStore) Get(id string) (*model.RequestLog, bool) {
	key := s.keyPrefix + id

	// Get log data from Redis
	data, err := s.client.Get(s.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		log.Printf("Failed to get log from Redis: %v", err)
		return nil, false
	}

	// Unmarshal data
	var reqLog model.RequestLog
	if err := json.Unmarshal(data, &reqLog); err != nil {
		log.Printf("Failed to unmarshal log data: %v", err)
		return nil, false
	}

	return &reqLog, true
}

// GetAll returns all stored request logs
func (s *RedisStore) GetAll() []*model.RequestLog {
	// Get all log IDs from the sorted set, in reverse order (newest first)
	ids, err := s.client.ZRevRange(s.ctx, s.keyPrefix+"logs", 0, -1).Result()
	if err != nil {
		log.Printf("Failed to get log IDs: %v", err)
		return nil
	}

	return s.getLogs(ids)
}

// GetLatest returns the n most recent request logs
func (s *RedisStore) GetLatest(n int) []*model.RequestLog {
	// Get the n newest log IDs
	ids, err := s.client.ZRevRange(s.ctx, s.keyPrefix+"logs", 0, int64(n-1)).Result()
	if err != nil {
		log.Printf("Failed to get latest log IDs: %v", err)
		return nil
	}

	return s.getLogs(ids)
}

// getLogs retrieves logs by their IDs
func (s *RedisStore) getLogs(ids []string) []*model.RequestLog {
	if len(ids) == 0 {
		return nil
	}

	// Use a pipeline for batch retrieval
	pipe := s.client.Pipeline()
	cmds := make(map[string]*redis.StringCmd)

	// Queue up the get commands
	for _, id := range ids {
		cmds[id] = pipe.Get(s.ctx, s.keyPrefix+id)
	}

	// Execute pipeline
	_, err := pipe.Exec(s.ctx)
	if err != nil && err != redis.Nil {
		log.Printf("Failed to execute pipeline: %v", err)
		return nil
	}

	// Process results
	logs := make([]*model.RequestLog, 0, len(ids))
	idToLog := make(map[string]*model.RequestLog)

	for id, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			if err != redis.Nil {
				log.Printf("Failed to get log %s: %v", id, err)
			}
			continue
		}

		var reqLog model.RequestLog
		if err := json.Unmarshal(data, &reqLog); err != nil {
			log.Printf("Failed to unmarshal log data for %s: %v", id, err)
			continue
		}

		logs = append(logs, &reqLog)
		idToLog[id] = &reqLog
	}

	// Sort logs by timestamp (newest first)
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	return logs
}

// Clear removes all logs from the store
func (s *RedisStore) Clear() error {
	// Get all log IDs
	ids, err := s.client.ZRange(s.ctx, s.keyPrefix+"logs", 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get log IDs: %w", err)
	}

	// Create a pipeline for batch operations
	pipe := s.client.Pipeline()

	// Remove from sorted set
	pipe.ZRem(s.ctx, s.keyPrefix+"logs", ids)

	// Remove each log from Redis
	for _, id := range ids {
		pipe.Unlink(s.ctx, s.keyPrefix+id)
	}

	// Execute pipeline
	if _, err := pipe.Exec(s.ctx); err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}

	// Delete the sorted set
	if err := s.client.Del(s.ctx, s.keyPrefix+"logs").Err(); err != nil {
		return fmt.Errorf("failed to delete sorted set: %w", err)
	}

	return nil
}

// Close closes the Redis client connection
func (s *RedisStore) Close() error {
	return s.client.Close()
}
