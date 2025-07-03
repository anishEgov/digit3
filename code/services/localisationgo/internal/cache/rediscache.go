package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
)

type RedisMessageCache struct {
	client *redis.Client
}

func NewRedisMessageCache(client *redis.Client) ports.MessageCache {
	return &RedisMessageCache{
		client: client,
	}
}

func (c *RedisMessageCache) GetMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error) {
	key := buildCacheKey(tenantID, module, locale)
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, ports.ErrCacheMiss
	} else if err != nil {
		return nil, err
	}

	var messages []domain.Message
	err = json.Unmarshal([]byte(val), &messages)
	if err != nil {
		return nil, err
	}
	return messages, nil
}

func (c *RedisMessageCache) SetMessages(ctx context.Context, tenantID, module, locale string, messages []domain.Message) error {
	key := buildCacheKey(tenantID, module, locale)
	val, err := json.Marshal(messages)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, val, 24*time.Hour).Err()
}

func (c *RedisMessageCache) Invalidate(ctx context.Context, tenantID, module, locale string) error {
	key := buildCacheKey(tenantID, module, locale)
	return c.client.Del(ctx, key).Err()
}

func (c *RedisMessageCache) BustCache(ctx context.Context, tenantID, module, locale string) error {
	// Busting cache for a specific module and locale combination
	if module != "" && locale != "" {
		key := buildCacheKey(tenantID, module, locale)
		return c.client.Del(ctx, key).Err()
	}

	// Busting cache for an entire tenant (potentially slow, use with caution)
	pattern := tenantID + ":*:*"
	iter := c.client.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		if err := c.client.Del(ctx, iter.Val()).Err(); err != nil {
			return err
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	return nil
}
