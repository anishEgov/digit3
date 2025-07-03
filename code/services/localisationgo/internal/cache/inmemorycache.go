package cache

import (
	"context"
	"log"
	"sync"

	"localisationgo/internal/core/domain"
	"localisationgo/internal/core/ports"
)

// InMemoryMessageCache is an in-memory implementation of the MessageCache interface
type InMemoryMessageCache struct {
	mu    sync.RWMutex
	cache map[string][]domain.Message
}

// NewInMemoryMessageCache creates a new InMemoryMessageCache
func NewInMemoryMessageCache() ports.MessageCache {
	return &InMemoryMessageCache{
		cache: make(map[string][]domain.Message),
	}
}

// GetMessages retrieves messages from the in-memory cache
func (c *InMemoryMessageCache) GetMessages(ctx context.Context, tenantID, module, locale string) ([]domain.Message, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := buildCacheKey(tenantID, module, locale)
	messages, found := c.cache[key]
	if !found {
		log.Printf("IN-MEMORY CACHE: MISS for key: %s", key)
		return nil, ports.ErrCacheMiss
	}

	log.Printf("IN-MEMORY CACHE: HIT for key: %s", key)
	return messages, nil
}

// SetMessages stores messages in the in-memory cache
func (c *InMemoryMessageCache) SetMessages(ctx context.Context, tenantID, module, locale string, messages []domain.Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := buildCacheKey(tenantID, module, locale)
	log.Printf("IN-MEMORY CACHE: SET for key: %s", key)
	c.cache[key] = messages
	return nil
}

// Invalidate removes a specific key from the in-memory cache
func (c *InMemoryMessageCache) Invalidate(ctx context.Context, tenantID, module, locale string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	key := buildCacheKey(tenantID, module, locale)
	log.Printf("IN-MEMORY CACHE: INVALIDATE for key: %s", key)
	delete(c.cache, key)
	return nil
}

// BustCache clears all cache entries for a given tenant
func (c *InMemoryMessageCache) BustCache(ctx context.Context, tenantID, module, locale string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Since this is a simple map, the easiest way to "bust" a tenant's cache
	// is to iterate and delete keys with the matching prefix.
	// For simplicity and because we don't have module/locale specific busting here,
	// this will bust the entire cache for the tenant.
	// A more complex implementation could handle module/locale busting.
	prefix := tenantID + ":"
	for key := range c.cache {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			delete(c.cache, key)
		}
	}
	return nil
}

// For consistency, we can use the same key-building logic as the Redis cache,
// even though it's not strictly necessary for a simple map.
func buildCacheKey(tenantID, module, locale string) string {
	return tenantID + ":" + module + ":" + locale
}
