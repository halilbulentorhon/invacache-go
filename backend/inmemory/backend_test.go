package inmemory

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/halilbulentorhon/invacache-go/backend"
	"github.com/halilbulentorhon/invacache-go/backend/option"
	"github.com/halilbulentorhon/invacache-go/config"
	"github.com/halilbulentorhon/invacache-go/constant"
)

func TestNewInMemoryBackend(t *testing.T) {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      4,
				Capacity:        100,
				SweeperInterval: 1 * time.Minute,
			},
		},
	}

	cache, err := NewInMemoryBackend[string](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	if cache == nil {
		t.Fatal("cache should not be nil")
	}
}

func TestNewInMemoryBackendWithDefaults(t *testing.T) {
	cfg := config.InvaCacheConfig{}

	cache, err := NewInMemoryBackend[int](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	if cache == nil {
		t.Fatal("cache should not be nil")
	}
}

func TestGetNonExistentKey(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	_, err := cache.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
	if err.Error() != constant.ErrKeyNotFound+": nonexistent" {
		t.Errorf("expected 'key not found' error, got: %v", err)
	}
}

func TestSetAndGet(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}
}

func TestSetWithTTL(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1", option.WithTTL(100*time.Millisecond))
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}

	time.Sleep(150 * time.Millisecond)

	_, err = cache.Get("key1")
	if err == nil {
		t.Fatal("expected error for expired key")
	}
}

func TestSetWithNoExpiration(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1", option.WithNoExpiration())
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}
}

func TestUpdateExistingKey(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	err = cache.Set("key1", "value2")
	if err != nil {
		t.Fatalf("unexpected error updating key: %v", err)
	}

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value2" {
		t.Errorf("expected 'value2', got '%s'", value)
	}
}

func TestDelete(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	err = cache.Delete("key1")
	if err != nil {
		t.Fatalf("unexpected error deleting key: %v", err)
	}

	_, err = cache.Get("key1")
	if err == nil {
		t.Fatal("expected error for deleted key")
	}
}

func TestDeleteNonExistentKey(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Delete("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error deleting non-existent key: %v", err)
	}
}

func TestGetOrLoadWithExistingKey(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	value, err := cache.GetOrLoad("key1", func(key string) (string, time.Duration, error) {
		t.Fatal("loader should not be called for existing key")
		return "", 0, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}
}

func TestGetOrLoadWithNonExistentKey(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	value, err := cache.GetOrLoad("key1", func(key string) (string, time.Duration, error) {
		return "loaded_value", 1 * time.Minute, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "loaded_value" {
		t.Errorf("expected 'loaded_value', got '%s'", value)
	}

	cachedValue, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting cached value: %v", err)
	}
	if cachedValue != "loaded_value" {
		t.Errorf("expected 'loaded_value', got '%s'", cachedValue)
	}
}

func TestGetOrLoadWithLoaderError(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	_, err := cache.GetOrLoad("key1", func(key string) (string, time.Duration, error) {
		return "", 0, errors.New("loader error")
	})
	if err == nil {
		t.Fatal("expected error from loader")
	}
	if err.Error() != "loader error" {
		t.Errorf("expected 'loader error', got: %v", err)
	}
}

func TestGetOrLoadConcurrent(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	var callCount int
	var mu sync.Mutex

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			value, err := cache.GetOrLoad("key1", func(key string) (string, time.Duration, error) {
				mu.Lock()
				callCount++
				mu.Unlock()
				time.Sleep(10 * time.Millisecond)
				return "loaded_value", 1 * time.Minute, nil
			})
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if value != "loaded_value" {
				t.Errorf("expected 'loaded_value', got '%s'", value)
			}
		}()
	}

	wg.Wait()

	mu.Lock()
	if callCount != 1 {
		t.Errorf("expected loader to be called once, got %d times", callCount)
	}
	mu.Unlock()
}

func TestCapacityEviction(t *testing.T) {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      1,
				Capacity:        2,
				SweeperInterval: 1 * time.Minute,
			},
		},
	}

	cache, err := NewInMemoryBackend[string](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	err = cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = cache.Set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	err = cache.Set("key3", "value3")
	if err != nil {
		t.Fatalf("unexpected error setting key3: %v", err)
	}

	_, err = cache.Get("key1")
	if err == nil {
		t.Fatal("key1 should have been evicted")
	}

	_, err = cache.Get("key2")
	if err != nil {
		t.Fatalf("unexpected error getting key2: %v", err)
	}

	_, err = cache.Get("key3")
	if err != nil {
		t.Fatalf("unexpected error getting key3: %v", err)
	}
}

func TestLRUBehavior(t *testing.T) {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      1,
				Capacity:        2,
				SweeperInterval: 1 * time.Minute,
			},
		},
	}

	cache, err := NewInMemoryBackend[string](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	err = cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = cache.Set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	_, err = cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key1: %v", err)
	}

	err = cache.Set("key3", "value3")
	if err != nil {
		t.Fatalf("unexpected error setting key3: %v", err)
	}

	_, err = cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key1: %v", err)
	}

	_, err = cache.Get("key2")
	if err == nil {
		t.Fatal("key2 should have been evicted")
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache := createTestCache[int](t)
	defer cache.Close()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key" + string(rune(i%10))
			err := cache.Set(key, i)
			if err != nil {
				t.Errorf("unexpected error setting key: %v", err)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		key := "key" + string(rune(i))
		_, err := cache.Get(key)
		if err != nil {
			t.Errorf("unexpected error getting key %s: %v", key, err)
		}
	}
}

func TestDifferentTypes(t *testing.T) {
	stringCache := createTestCache[string](t)
	defer stringCache.Close()

	intCache := createTestCache[int](t)
	defer intCache.Close()

	err := stringCache.Set("str", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = intCache.Set("int", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	strValue, err := stringCache.Get("str")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strValue != "hello" {
		t.Errorf("expected 'hello', got '%s'", strValue)
	}

	intValue, err := intCache.Get("int")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intValue != 42 {
		t.Errorf("expected 42, got %d", intValue)
	}
}

func TestClose(t *testing.T) {
	cache := createTestCache[string](t)

	err := cache.Close()
	if err != nil {
		t.Fatalf("unexpected error closing cache: %v", err)
	}
}

func TestGetInvalidatorConfig(t *testing.T) {
	tests := []struct {
		name     string
		cfgType  string
		expected bool
	}{
		{
			name:     "couchbase config",
			cfgType:  constant.CouchbaseInvalidationConfigType,
			expected: true,
		},
		{
			name:     "redis config",
			cfgType:  constant.RedisInvalidationConfigType,
			expected: true,
		},
		{
			name:     "unknown config",
			cfgType:  "unknown",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.InvalidationConfig{
				Type:         tt.cfgType,
				DriverConfig: map[string]any{"host": "localhost"},
			}

			result := getInvalidatorConfig(cfg)
			if tt.expected && result == nil {
				t.Error("expected non-nil result")
			}
			if !tt.expected && result != nil {
				t.Error("expected nil result")
			}
		})
	}
}

func TestClear(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = cache.Set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	err = cache.Set("key3", "value3")
	if err != nil {
		t.Fatalf("unexpected error setting key3: %v", err)
	}

	err = cache.Clear()
	if err != nil {
		t.Fatalf("unexpected error clearing cache: %v", err)
	}

	_, err = cache.Get("key1")
	if err == nil {
		t.Error("key1 should not exist after clear")
	}

	_, err = cache.Get("key2")
	if err == nil {
		t.Error("key2 should not exist after clear")
	}

	_, err = cache.Get("key3")
	if err == nil {
		t.Error("key3 should not exist after clear")
	}
}

func TestClearEmpty(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Clear()
	if err != nil {
		t.Fatalf("unexpected error clearing empty cache: %v", err)
	}
}

func TestClearAndReuse(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = cache.Clear()
	if err != nil {
		t.Fatalf("unexpected error clearing cache: %v", err)
	}

	err = cache.Set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error after clear: %v", err)
	}

	value, err := cache.Get("key2")
	if err != nil {
		t.Fatalf("unexpected error getting key2: %v", err)
	}
	if value != "value2" {
		t.Errorf("expected 'value2', got '%s'", value)
	}

	_, err = cache.Get("key1")
	if err == nil {
		t.Error("key1 should not exist after clear")
	}
}

func TestDeleteWithOption(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	err = cache.Delete("key1", option.WithDeleteInvalidation())
	if err != nil {
		t.Fatalf("unexpected error deleting key: %v", err)
	}

	_, err = cache.Get("key1")
	if err == nil {
		t.Fatal("expected error for deleted key")
	}
}

func TestSetWithInvalidationOption(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1", option.WithInvalidation())
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}
}

func TestClearWithOption(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = cache.Clear(option.WithClearInvalidation())
	if err != nil {
		t.Fatalf("unexpected error clearing cache: %v", err)
	}

	_, err = cache.Get("key1")
	if err == nil {
		t.Error("key1 should not exist after clear")
	}
}

func TestIsClearEvent(t *testing.T) {
	if !isClearEvent(constant.EmptyString) {
		t.Error("empty string should be a clear event")
	}

	if isClearEvent("somekey") {
		t.Error("non-empty string should not be a clear event")
	}
}

func TestHandleInvalidationMessageClear(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	be := cache.(*inMemoryBackend[string])

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = be.handleInvalidationMessage(constant.EmptyString)
	if err != nil {
		t.Fatalf("unexpected error handling clear message: %v", err)
	}

	_, err = cache.Get("key1")
	if err == nil {
		t.Error("key1 should not exist after clear event")
	}
}

func TestHandleInvalidationMessageDelete(t *testing.T) {
	cache := createTestCache[string](t)
	defer cache.Close()

	be := cache.(*inMemoryBackend[string])

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = cache.Set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = be.handleInvalidationMessage("key1")
	if err != nil {
		t.Fatalf("unexpected error handling delete message: %v", err)
	}

	_, err = cache.Get("key1")
	if err == nil {
		t.Error("key1 should not exist after delete event")
	}

	value, err := cache.Get("key2")
	if err != nil {
		t.Fatalf("unexpected error getting key2: %v", err)
	}
	if value != "value2" {
		t.Errorf("expected 'value2', got '%s'", value)
	}
}

func TestSetWithDefaultTTL(t *testing.T) {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      4,
				Capacity:        100,
				SweeperInterval: 1 * time.Minute,
				Ttl:             "100ms",
			},
		},
	}

	cache, err := NewInMemoryBackend[string](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	err = cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}

	time.Sleep(150 * time.Millisecond)

	_, err = cache.Get("key1")
	if err == nil {
		t.Fatal("expected error for expired key with default TTL")
	}
}

func TestSetWithDefaultTTLOverriddenByOption(t *testing.T) {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      4,
				Capacity:        100,
				SweeperInterval: 1 * time.Minute,
				Ttl:             "50ms",
			},
		},
	}

	cache, err := NewInMemoryBackend[string](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	err = cache.Set("key1", "value1", option.WithTTL(200*time.Millisecond))
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("key should still be valid, default TTL was overridden: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}
}

func TestSetWithDefaultTTLAndNoExpiration(t *testing.T) {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      4,
				Capacity:        100,
				SweeperInterval: 1 * time.Minute,
				Ttl:             "50ms",
			},
		},
	}

	cache, err := NewInMemoryBackend[string](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	err = cache.Set("key1", "value1", option.WithNoExpiration())
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("key should not expire with NoExpiration option: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}
}

func TestNoExpirationFlagIgnoresDefaultTTL(t *testing.T) {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      4,
				Capacity:        100,
				SweeperInterval: 1 * time.Minute,
				Ttl:             "100ms",
			},
		},
	}

	cache, err := NewInMemoryBackend[string](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	err = cache.Set("key1", "value1", option.WithNoExpiration())
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = cache.Set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	time.Sleep(150 * time.Millisecond)

	value1, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("key1 should not expire with NoExpiration: %v", err)
	}
	if value1 != "value1" {
		t.Errorf("expected 'value1', got '%s'", value1)
	}

	_, err = cache.Get("key2")
	if err == nil {
		t.Fatal("key2 should have expired with default TTL")
	}
}

func TestUpdateWithNoExpiration(t *testing.T) {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      4,
				Capacity:        100,
				SweeperInterval: 1 * time.Minute,
				Ttl:             "50ms",
			},
		},
	}

	cache, err := NewInMemoryBackend[string](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	err = cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	time.Sleep(30 * time.Millisecond)

	err = cache.Set("key1", "value2", option.WithNoExpiration())
	if err != nil {
		t.Fatalf("unexpected error updating key: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("key should not expire after NoExpiration update: %v", err)
	}
	if value != "value2" {
		t.Errorf("expected 'value2', got '%s'", value)
	}
}

func TestDefaultTTLAppliedWhenNoOption(t *testing.T) {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      4,
				Capacity:        100,
				SweeperInterval: 1 * time.Minute,
				Ttl:             "100ms",
			},
		},
	}

	cache, err := NewInMemoryBackend[string](cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer cache.Close()

	err = cache.Set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	value, err := cache.Get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}

	time.Sleep(70 * time.Millisecond)

	_, err = cache.Get("key1")
	if err == nil {
		t.Fatal("key should have expired with default TTL")
	}
}

func createTestCache[V any](t *testing.T) backend.Cache[V] {
	cfg := config.InvaCacheConfig{
		Backend: &config.BackendConfig{
			InMemory: &config.InMemoryConfig{
				ShardCount:      4,
				Capacity:        100,
				SweeperInterval: 1 * time.Minute,
			},
		},
	}

	cache, err := NewInMemoryBackend[V](cfg)
	if err != nil {
		t.Fatalf("unexpected error creating cache: %v", err)
	}

	return cache
}
