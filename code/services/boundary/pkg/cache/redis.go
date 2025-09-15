package cache

import (
    "context"
    "github.com/go-redis/redis/v8"
    "errors"
    "fmt"
)

type RedisCache struct {
    client *redis.Client
}

func NewRedisCache(addr, password string, db int) *RedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: password,
        DB:       db,
    })
    return &RedisCache{client: rdb}
}

func (c *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
    val, err := c.client.Get(ctx, key).Result()
    if err == redis.Nil {
        return nil, errors.New("key not found")
    } else if err != nil {
        return nil, fmt.Errorf("redis get error: %w", err)
    }
    return val, nil
}

func (c *RedisCache) Set(ctx context.Context, key string, value interface{}) error {
    strVal, ok := value.(string)
    if !ok {
        return errors.New("redis cache only supports string values")
    }
    return c.client.Set(ctx, key, strVal, 0).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
    return c.client.Del(ctx, key).Err()
} 