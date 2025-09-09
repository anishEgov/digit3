package cache

import (
    "context"
    "sync"
    "errors"
)

type InMemoryCache struct {
    mu    sync.RWMutex
    store map[string]interface{}
}

func NewInMemoryCache() *InMemoryCache {
    return &InMemoryCache{
        store: make(map[string]interface{}),
    }
}

func (c *InMemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.store[key]
    if !ok {
        return nil, errors.New("key not found")
    }
    return val, nil
}

func (c *InMemoryCache) Set(ctx context.Context, key string, value interface{}) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.store[key] = value
    return nil
}

func (c *InMemoryCache) Delete(ctx context.Context, key string) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.store, key)
    return nil
} 