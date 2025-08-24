package repo

import (
	"context"
	"time"
)

// Cache defines the interface for caching operations used by the weather API
type Cache interface {
	// Get retrieves a value from the cache by key
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in the cache with the specified TTL
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error

	// Delete removes a key from the cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in the cache
	Exists(ctx context.Context, key string) (bool, error)

	// SetNX sets a key only if it doesn't exist (atomic operation)
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)

	// GetTTL returns the remaining TTL for a key
	GetTTL(ctx context.Context, key string) (time.Duration, error)

	// Clear removes all keys from the cache (use with caution)
	Clear(ctx context.Context) error

	// Close closes the cache connection
	Close() error
}

// KVStore defines the interface for the underlying key-value storage
type KVStore interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error)
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Clear(ctx context.Context) error
	Close() error
}

// RequestCache implements Cache interface with request-specific optimizations
type RequestCache struct {
	store  KVStore
	prefix string
}

// NewRequestCache creates a new RequestCache instance
func NewRequestCache(store KVStore, prefix string) Cache {
	return &RequestCache{
		store:  store,
		prefix: prefix,
	}
}

// Get retrieves a value from the cache
func (c *RequestCache) Get(ctx context.Context, key string) ([]byte, error) {
	return c.store.Get(ctx, c.prefixKey(key))
}

// Set stores a value in the cache with TTL
func (c *RequestCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.store.Set(ctx, c.prefixKey(key), value, ttl)
}

// Delete removes a key from the cache
func (c *RequestCache) Delete(ctx context.Context, key string) error {
	return c.store.Delete(ctx, c.prefixKey(key))
}

// Exists checks if a key exists in the cache
func (c *RequestCache) Exists(ctx context.Context, key string) (bool, error) {
	return c.store.Exists(ctx, c.prefixKey(key))
}

// SetNX sets a key only if it doesn't exist
func (c *RequestCache) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	return c.store.SetNX(ctx, c.prefixKey(key), value, ttl)
}

// GetTTL returns the remaining TTL for a key
func (c *RequestCache) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	return c.store.GetTTL(ctx, c.prefixKey(key))
}

// Clear removes all keys from the cache
func (c *RequestCache) Clear(ctx context.Context) error {
	return c.store.Clear(ctx)
}

// Close closes the cache connection
func (c *RequestCache) Close() error {
	return c.store.Close()
}

func (c *RequestCache) prefixKey(key string) string {
	if c.prefix == "" {
		return key
	}
	return c.prefix + ":" + key
}
