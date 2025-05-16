package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
)

// Redis key expiration time
const keyExpiration = 24 * time.Hour

// RedisCacheImpl implements MessageCache using Redis
type RedisCacheImpl struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis cache
func NewRedisCache(client *redis.Client) ports.MessageCache {
	return &RedisCacheImpl{
		client: client,
	}
}

// buildKey creates a cache key from tenant, module, and locale
func buildKey(tenantID, module, locale string) string {
	// For empty module queries, use a special prefix to avoid conflicts
	if module == "" {
		return fmt.Sprintf("messages:all:%s:%s", tenantID, locale)
	}
	return fmt.Sprintf("messages:%s:%s:%s", tenantID, module, locale)
}

// SetMessages adds messages to the cache
func (c *RedisCacheImpl) SetMessages(ctx context.Context, tenantID, module, locale string, messages []domain.Message) error {
	if len(messages) == 0 {
		return nil
	}

	key := buildKey(tenantID, module, locale)

	// Serialize messages to JSON
	messagesJSON, err := json.Marshal(messages)
	if err != nil {
		return err
	}

	// Store in Redis with expiration
	return c.client.Set(ctx, key, messagesJSON, keyExpiration).Err()
}

// GetMessages retrieves messages from the cache
func (c *RedisCacheImpl) GetMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error) {
	key := buildKey(tenantID, module, locale)

	// Get data from Redis
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			// Key does not exist, return empty slice
			return []domain.Message{}, nil
		}
		return nil, err
	}

	// Deserialize JSON to messages
	var messages []domain.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, err
	}

	return messages, nil
}

// Invalidate removes cached messages for a specific tenant+module+locale
func (c *RedisCacheImpl) Invalidate(ctx context.Context, tenantID, module, locale string) error {
	key := buildKey(tenantID, module, locale)
	return c.client.Del(ctx, key).Err()
}

// BustCache clears the entire cache by deleting all message keys
func (c *RedisCacheImpl) BustCache(ctx context.Context) error {
	// Use SCAN to find all message keys
	var cursor uint64
	var keys []string

	for {
		var batch []string
		var err error

		// Scan for keys with the pattern 'messages:*'
		batch, cursor, err = c.client.Scan(ctx, cursor, "messages:*", 100).Result()
		if err != nil {
			return err
		}

		keys = append(keys, batch...)

		// If cursor is 0, we've scanned all keys
		if cursor == 0 {
			break
		}
	}

	// Delete the keys if any found
	if len(keys) > 0 {
		return c.client.Del(ctx, keys...).Err()
	}

	return nil
}
