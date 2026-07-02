package store

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/doganarif/govisual/internal/model"
	"github.com/go-redis/redis/v8"
)

// RedisStore implements the Store interface with Redis as backend
type RedisStore struct {
	client      *redis.Client
	keyPrefix   string
	capacity    int
	ttl         time.Duration
	insertCount atomic.Uint64
}

// NewRedisStore creates a new Redis-backed store
func NewRedisStore(connStr string, capacity int, ttlSeconds int) (*RedisStore, error) {
	if capacity <= 0 {
		capacity = 100
	}
	if ttlSeconds <= 0 {
		ttlSeconds = 86400 // 24h
	}

	opts, err := redis.ParseURL(connStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis connection string: %w", err)
	}

	client := redis.NewClient(opts)

	pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisStore{
		client:    client,
		keyPrefix: "govisual:",
		capacity:  capacity,
		ttl:       time.Duration(ttlSeconds) * time.Second,
	}, nil
}

// opCtx returns a short-lived context for a single Redis call. Stores must not
// hang onto a context for their entire lifetime.
func (s *RedisStore) opCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}

func (s *RedisStore) Add(reqLog *model.RequestLog) error {
	data, err := json.Marshal(reqLog)
	if err != nil {
		return fmt.Errorf("redis marshal: %w", err)
	}

	ctx, cancel := s.opCtx()
	defer cancel()

	key := s.keyPrefix + reqLog.ID
	if err := s.client.Set(ctx, key, data, s.ttl).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	score := float64(reqLog.Timestamp.UnixNano())
	if err := s.client.ZAdd(ctx, s.keyPrefix+"logs", &redis.Z{
		Score:  score,
		Member: reqLog.ID,
	}).Err(); err != nil {
		return fmt.Errorf("redis zadd: %w", err)
	}

	if s.insertCount.Add(1)%cleanupEveryN == 0 {
		s.cleanup()
	}
	return nil
}

func (s *RedisStore) cleanup() {
	ctx, cancel := s.opCtx()
	defer cancel()

	count, err := s.client.ZCard(ctx, s.keyPrefix+"logs").Result()
	if err != nil {
		log.Printf("govisual: failed to count Redis logs: %v", err)
		return
	}
	if count <= int64(s.capacity) {
		return
	}

	oldestIDs, err := s.client.ZRange(ctx, s.keyPrefix+"logs", 0, count-int64(s.capacity)-1).Result()
	if err != nil {
		log.Printf("govisual: failed to get oldest Redis log IDs: %v", err)
		return
	}
	if len(oldestIDs) == 0 {
		return
	}

	pipe := s.client.Pipeline()
	// ZRem takes ...interface{}; convert from []string.
	members := make([]interface{}, len(oldestIDs))
	for i, id := range oldestIDs {
		members[i] = id
	}
	pipe.ZRem(ctx, s.keyPrefix+"logs", members...)
	for _, id := range oldestIDs {
		pipe.Del(ctx, s.keyPrefix+id)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		log.Printf("govisual: failed to clean up old Redis logs: %v", err)
	}
}

func (s *RedisStore) Get(id string) (*model.RequestLog, bool) {
	ctx, cancel := s.opCtx()
	defer cancel()

	data, err := s.client.Get(ctx, s.keyPrefix+id).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, false
		}
		log.Printf("govisual: failed to get log from Redis: %v", err)
		return nil, false
	}

	var reqLog model.RequestLog
	if err := json.Unmarshal(data, &reqLog); err != nil {
		log.Printf("govisual: failed to unmarshal Redis log: %v", err)
		return nil, false
	}
	return &reqLog, true
}

func (s *RedisStore) GetAll() []*model.RequestLog {
	ctx, cancel := s.opCtx()
	defer cancel()

	ids, err := s.client.ZRevRange(ctx, s.keyPrefix+"logs", 0, -1).Result()
	if err != nil {
		log.Printf("govisual: failed to get Redis log IDs: %v", err)
		return nil
	}
	return s.getLogs(ctx, ids)
}

func (s *RedisStore) GetLatest(n int) []*model.RequestLog {
	ctx, cancel := s.opCtx()
	defer cancel()

	ids, err := s.client.ZRevRange(ctx, s.keyPrefix+"logs", 0, int64(n-1)).Result()
	if err != nil {
		log.Printf("govisual: failed to get latest Redis log IDs: %v", err)
		return nil
	}
	return s.getLogs(ctx, ids)
}

func (s *RedisStore) getLogs(ctx context.Context, ids []string) []*model.RequestLog {
	if len(ids) == 0 {
		return nil
	}

	// Keep results aligned with ids so the ZRevRange order survives; a map
	// here would shuffle entries that share a timestamp.
	pipe := s.client.Pipeline()
	cmds := make([]*redis.StringCmd, len(ids))
	for i, id := range ids {
		cmds[i] = pipe.Get(ctx, s.keyPrefix+id)
	}
	if _, err := pipe.Exec(ctx); err != nil && err != redis.Nil {
		log.Printf("govisual: failed to execute Redis pipeline: %v", err)
		return nil
	}

	logs := make([]*model.RequestLog, 0, len(ids))
	var expired []interface{}
	for i, cmd := range cmds {
		data, err := cmd.Bytes()
		if err != nil {
			if err == redis.Nil {
				// The key expired but its ID is still indexed; prune it so
				// the sorted set doesn't diverge from the key space.
				expired = append(expired, ids[i])
			} else {
				log.Printf("govisual: failed to get Redis log %s: %v", ids[i], err)
			}
			continue
		}
		var reqLog model.RequestLog
		if err := json.Unmarshal(data, &reqLog); err != nil {
			log.Printf("govisual: failed to unmarshal Redis log %s: %v", ids[i], err)
			continue
		}
		logs = append(logs, &reqLog)
	}

	if len(expired) > 0 {
		if err := s.client.ZRem(ctx, s.keyPrefix+"logs", expired...).Err(); err != nil {
			log.Printf("govisual: failed to prune expired Redis log IDs: %v", err)
		}
	}
	return logs
}

func (s *RedisStore) Clear() error {
	ctx, cancel := s.opCtx()
	defer cancel()

	ids, err := s.client.ZRange(ctx, s.keyPrefix+"logs", 0, -1).Result()
	if err != nil {
		return fmt.Errorf("failed to get log IDs: %w", err)
	}

	pipe := s.client.Pipeline()
	if len(ids) > 0 {
		members := make([]interface{}, len(ids))
		for i, id := range ids {
			members[i] = id
		}
		pipe.ZRem(ctx, s.keyPrefix+"logs", members...)
		for _, id := range ids {
			pipe.Unlink(ctx, s.keyPrefix+id)
		}
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}

	if err := s.client.Del(ctx, s.keyPrefix+"logs").Err(); err != nil {
		return fmt.Errorf("failed to delete sorted set: %w", err)
	}
	return nil
}

func (s *RedisStore) Close() error {
	return s.client.Close()
}
