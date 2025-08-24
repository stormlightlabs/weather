package repo

import (
	"context"
	"errors"
	"testing"
	"time"
)

// MockKVStore implements KVStore interface for testing
type MockKVStore struct {
	data        map[string][]byte
	ttls        map[string]time.Time
	shouldError bool
	errorMsg    string
}

// NewMockKVStore creates a new MockKVStore
func NewMockKVStore() *MockKVStore {
	return &MockKVStore{
		data: make(map[string][]byte),
		ttls: make(map[string]time.Time),
	}
}

func (m *MockKVStore) Get(ctx context.Context, key string) ([]byte, error) {
	if m.shouldError {
		return nil, errors.New(m.errorMsg)
	}

	if expiry, exists := m.ttls[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttls, key)
		return nil, errors.New("key not found")
	}

	value, exists := m.data[key]
	if !exists {
		return nil, errors.New("key not found")
	}
	return value, nil
}

func (m *MockKVStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	m.data[key] = value
	if ttl > 0 {
		m.ttls[key] = time.Now().Add(ttl)
	}
	return nil
}

func (m *MockKVStore) Delete(ctx context.Context, key string) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	delete(m.data, key)
	delete(m.ttls, key)
	return nil
}

func (m *MockKVStore) Exists(ctx context.Context, key string) (bool, error) {
	if m.shouldError {
		return false, errors.New(m.errorMsg)
	}

	if expiry, exists := m.ttls[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttls, key)
		return false, nil
	}

	_, exists := m.data[key]
	return exists, nil
}

func (m *MockKVStore) SetNX(ctx context.Context, key string, value []byte, ttl time.Duration) (bool, error) {
	if m.shouldError {
		return false, errors.New(m.errorMsg)
	}

	if expiry, exists := m.ttls[key]; exists && time.Now().After(expiry) {
		delete(m.data, key)
		delete(m.ttls, key)
	}

	if _, exists := m.data[key]; exists {
		return false, nil
	}

	m.data[key] = value
	if ttl > 0 {
		m.ttls[key] = time.Now().Add(ttl)
	}
	return true, nil
}

func (m *MockKVStore) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	if m.shouldError {
		return 0, errors.New(m.errorMsg)
	}

	expiry, exists := m.ttls[key]
	if !exists {
		return -1, nil
	}

	remaining := time.Until(expiry)
	if remaining <= 0 {
		delete(m.data, key)
		delete(m.ttls, key)
		return -1, nil
	}

	return remaining, nil
}

func (m *MockKVStore) Clear(ctx context.Context) error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}

	m.data = make(map[string][]byte)
	m.ttls = make(map[string]time.Time)
	return nil
}

func (m *MockKVStore) Close() error {
	if m.shouldError {
		return errors.New(m.errorMsg)
	}
	return nil
}

// SetError configures the mock to return errors
func (m *MockKVStore) SetError(shouldError bool, errorMsg string) {
	m.shouldError = shouldError
	m.errorMsg = errorMsg
}

// Consolidated test function following the project's pattern
func TestCache(t *testing.T) {
	t.Run("interface compliance", func(t *testing.T) {
		var _ Cache = (*RequestCache)(nil)
	})

	t.Run("RequestCache creation", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "test")

		if cache == nil {
			t.Error("NewRequestCache returned nil")
		}
	})

	t.Run("basic operations", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "test")
		ctx := context.Background()

		key := "weather:forecast:123"
		value := []byte(`{"temperature": 20.5}`)
		ttl := 5 * time.Minute

		err := cache.Set(ctx, key, value, ttl)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		retrieved, err := cache.Get(ctx, key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}

		if string(retrieved) != string(value) {
			t.Errorf("Retrieved value mismatch. Expected %s, got %s", string(value), string(retrieved))
		}
	})

	t.Run("key prefixing", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "weather")
		ctx := context.Background()

		key := "forecast:123"
		value := []byte("test data")

		err := cache.Set(ctx, key, value, time.Minute)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		_, exists := store.data["weather:"+key]
		if !exists {
			t.Error("Key was not properly prefixed in the underlying store")
		}

		retrieved, err := cache.Get(ctx, key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}

		if string(retrieved) != string(value) {
			t.Errorf("Retrieved value mismatch")
		}
	})

	t.Run("empty prefix", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "")
		ctx := context.Background()

		key := "test:key"
		value := []byte("test value")

		err := cache.Set(ctx, key, value, time.Minute)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		_, exists := store.data[key]
		if !exists {
			t.Error("Key should be stored without prefix when prefix is empty")
		}
	})

	t.Run("exists", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "test")
		ctx := context.Background()

		key := "exists:test"

		exists, err := cache.Exists(ctx, key)
		if err != nil {
			t.Errorf("Exists failed: %v", err)
		}
		if exists {
			t.Error("Key should not exist")
		}

		err = cache.Set(ctx, key, []byte("data"), time.Minute)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		exists, err = cache.Exists(ctx, key)
		if err != nil {
			t.Errorf("Exists failed: %v", err)
		}
		if !exists {
			t.Error("Key should exist")
		}
	})

	t.Run("delete", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "test")
		ctx := context.Background()

		key := "delete:test"
		value := []byte("to be deleted")

		err := cache.Set(ctx, key, value, time.Minute)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		exists, err := cache.Exists(ctx, key)
		if err != nil || !exists {
			t.Error("Key should exist before deletion")
		}

		err = cache.Delete(ctx, key)
		if err != nil {
			t.Errorf("Delete failed: %v", err)
		}

		exists, err = cache.Exists(ctx, key)
		if err != nil {
			t.Errorf("Exists check failed: %v", err)
		}
		if exists {
			t.Error("Key should not exist after deletion")
		}
	})

	t.Run("SetNX", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "test")
		ctx := context.Background()

		key := "setnx:test"
		value1 := []byte("first value")
		value2 := []byte("second value")

		success, err := cache.SetNX(ctx, key, value1, time.Minute)
		if err != nil {
			t.Errorf("SetNX failed: %v", err)
		}
		if !success {
			t.Error("First SetNX should succeed")
		}

		success, err = cache.SetNX(ctx, key, value2, time.Minute)
		if err != nil {
			t.Errorf("SetNX failed: %v", err)
		}
		if success {
			t.Error("Second SetNX should fail because key exists")
		}

		retrieved, err := cache.Get(ctx, key)
		if err != nil {
			t.Errorf("Get failed: %v", err)
		}
		if string(retrieved) != string(value1) {
			t.Error("Original value should be preserved")
		}
	})

	t.Run("TTL", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "test")
		ctx := context.Background()

		key := "ttl:test"
		value := []byte("ttl data")
		ttl := time.Hour

		err := cache.Set(ctx, key, value, ttl)
		if err != nil {
			t.Errorf("Set failed: %v", err)
		}

		remaining, err := cache.GetTTL(ctx, key)
		if err != nil {
			t.Errorf("GetTTL failed: %v", err)
		}

		if remaining <= 0 || remaining > ttl {
			t.Errorf("TTL should be positive and less than original TTL. Got: %v", remaining)
		}

		remaining, err = cache.GetTTL(ctx, "nonexistent")
		if err != nil {
			t.Errorf("GetTTL for nonexistent key failed: %v", err)
		}
		if remaining != -1 {
			t.Error("TTL for nonexistent key should be -1")
		}
	})

	t.Run("clear", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "test")
		ctx := context.Background()

		keys := []string{"clear:1", "clear:2", "clear:3"}
		for _, key := range keys {
			err := cache.Set(ctx, key, []byte("data"), time.Minute)
			if err != nil {
				t.Errorf("Set failed for key %s: %v", key, err)
			}
		}

		for _, key := range keys {
			exists, err := cache.Exists(ctx, key)
			if err != nil || !exists {
				t.Errorf("Key %s should exist before clear", key)
			}
		}

		err := cache.Clear(ctx)
		if err != nil {
			t.Errorf("Clear failed: %v", err)
		}

		for _, key := range keys {
			exists, err := cache.Exists(ctx, key)
			if err != nil {
				t.Errorf("Exists check failed for key %s: %v", key, err)
			}
			if exists {
				t.Errorf("Key %s should not exist after clear", key)
			}
		}
	})

	t.Run("close", func(t *testing.T) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "test")

		err := cache.Close()
		if err != nil {
			t.Errorf("Close failed: %v", err)
		}
	})

	t.Run("error handling & propogation", func(t *testing.T) {
		store := NewMockKVStore()
		store.SetError(true, "mock error")
		cache := NewRequestCache(store, "test")
		ctx := context.Background()

		_, err := cache.Get(ctx, "test")
		if err == nil {
			t.Error("Get should return error when store fails")
		}

		err = cache.Set(ctx, "test", []byte("data"), time.Minute)
		if err == nil {
			t.Error("Set should return error when store fails")
		}

		err = cache.Delete(ctx, "test")
		if err == nil {
			t.Error("Delete should return error when store fails")
		}

		_, err = cache.Exists(ctx, "test")
		if err == nil {
			t.Error("Exists should return error when store fails")
		}

		_, err = cache.SetNX(ctx, "test", []byte("data"), time.Minute)
		if err == nil {
			t.Error("SetNX should return error when store fails")
		}

		_, err = cache.GetTTL(ctx, "test")
		if err == nil {
			t.Error("GetTTL should return error when store fails")
		}

		err = cache.Clear(ctx)
		if err == nil {
			t.Error("Clear should return error when store fails")
		}

		err = cache.Close()
		if err == nil {
			t.Error("Close should return error when store fails")
		}
	})
}

func BenchmarkCache(b *testing.B) {
	b.Run("RequestCache Set", func(b *testing.B) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "bench")
		ctx := context.Background()
		value := []byte("benchmark data")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "benchmark:set:" + string(rune(i))
			_ = cache.Set(ctx, key, value, time.Minute)
		}
	})

	b.Run("RequestCache Get", func(b *testing.B) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "bench")
		ctx := context.Background()

		key := "benchmark:get"
		value := []byte("benchmark data")
		_ = cache.Set(ctx, key, value, time.Hour)

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, _ = cache.Get(ctx, key)
		}
	})

	b.Run("RequestCache SetNX", func(b *testing.B) {
		store := NewMockKVStore()
		cache := NewRequestCache(store, "bench")
		ctx := context.Background()
		value := []byte("benchmark data")

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "benchmark:setnx:" + string(rune(i))
			_, _ = cache.SetNX(ctx, key, value, time.Minute)
		}
	})
}
