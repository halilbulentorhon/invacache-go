package inmemory

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/halilbulentorhon/invacache-go/backend/option"
	"github.com/halilbulentorhon/invacache-go/config"
	"github.com/halilbulentorhon/invacache-go/constant"
)

func TestBackendBasicOperations(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   50,
			ShardCount: 4,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	err := cache.Set("user:123", "alice")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	value, err := cache.Get("user:123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "alice" {
		t.Errorf("expected 'alice', got '%s'", value)
	}

	err = cache.Delete("user:123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = cache.Get("user:123")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestBackendMultiShardDistribution(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   20,
			ShardCount: 4,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	keys := []string{"apple", "banana", "cherry", "date", "elderberry", "fig", "grape", "honey"}

	for _, key := range keys {
		cache.Set(key, "fruit_"+key)
	}

	for _, key := range keys {
		value, err := cache.Get(key)
		if err != nil {
			t.Errorf("unexpected error for key %s: %v", key, err)
		}
		expected := "fruit_" + key
		if value != expected {
			t.Errorf("expected '%s', got '%s'", expected, value)
		}
	}
}

func TestBackendGetOrLoadCacheHit(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 2,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("existing", "cached_value")

	loader := func(key string) (string, time.Duration, error) {
		return "should_not_be_called", 0, nil
	}

	value, err := cache.GetOrLoad("existing", loader)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "cached_value" {
		t.Errorf("expected 'cached_value', got '%s'", value)
	}
}

func TestBackendGetOrLoadCacheMiss(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 2,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	loader := func(key string) (string, time.Duration, error) {
		return "loaded_" + key, 50 * time.Millisecond, nil
	}

	value, err := cache.GetOrLoad("missing_key", loader)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "loaded_missing_key" {
		t.Errorf("expected 'loaded_missing_key', got '%s'", value)
	}

	cachedValue, err := cache.Get("missing_key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cachedValue != "loaded_missing_key" {
		t.Errorf("expected cached value, got '%s'", cachedValue)
	}

	time.Sleep(60 * time.Millisecond)

	_, err = cache.Get("missing_key")
	if err == nil {
		t.Error("key should have expired")
	}
}

func TestBackendGetOrLoadWithError(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 2,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	loader := func(key string) (string, time.Duration, error) {
		return "", 0, errors.New("database connection failed")
	}

	_, err := cache.GetOrLoad("error_key", loader)
	if err == nil {
		t.Error("expected loader error")
	}
	if err.Error() != "database connection failed" {
		t.Errorf("expected 'database connection failed', got '%s'", err.Error())
	}

	_, err = cache.Get("error_key")
	if err == nil {
		t.Error("key should not exist after failed load")
	}
}

func TestBackendConfigurationDefaults(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:        -1,
			ShardCount:      0,
			SweeperInterval: -1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("test", "value")
	value, err := cache.Get("test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "value" {
		t.Errorf("expected 'value', got '%s'", value)
	}
}

func TestBackendCapacityDistribution(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 3,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	backend := cache.(*inMemoryBackend[string])

	expectedCapacities := []int{3, 3, 4}
	for i := range backend.shards {
		if backend.shards[i].capacity != expectedCapacities[i] {
			t.Errorf("shard %d: expected capacity %d, got %d", i, expectedCapacities[i], backend.shards[i].capacity)
		}
	}
}

func TestBackendConcurrentAccess(t *testing.T) {
	cache := NewInMemoryBackend[int](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   2000,
			ShardCount: 4,
		},
	})
	defer cache.(*inMemoryBackend[int]).Close()

	var wg sync.WaitGroup
	numWorkers := 10
	numOps := 50

	for worker := 0; worker < numWorkers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for op := 0; op < numOps; op++ {
				key := fmt.Sprintf("worker_%d_op_%d", workerID, op)
				value := workerID*1000 + op

				cache.Set(key, value)

				retrieved, err := cache.Get(key)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if retrieved != value {
					t.Errorf("expected %d, got %d", value, retrieved)
					return
				}
			}
		}(worker)
	}

	wg.Wait()
}

func TestBackendSweepBehavior(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:        20,
			ShardCount:      2,
			SweeperInterval: 25 * time.Millisecond,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("expire_soon", "value1", option.WithTTL(20*time.Millisecond))
	cache.Set("permanent", "value2")

	time.Sleep(45 * time.Millisecond)

	_, err := cache.Get("expire_soon")
	if err == nil {
		t.Error("expired key should have been swept")
	}

	value, err := cache.Get("permanent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "value2" {
		t.Errorf("expected 'value2', got '%s'", value)
	}
}

func TestBackendGracefulShutdown(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 2,
		},
	})

	cache.Set("test", "value")

	err := cache.(*inMemoryBackend[string]).Close()
	if err != nil {
		t.Errorf("unexpected error during close: %v", err)
	}

	value, err := cache.Get("test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "value" {
		t.Errorf("expected 'value', got '%s'", value)
	}
}

func TestBackendWithTTL(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 2,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("temp", "temporary", option.WithTTL(40*time.Millisecond))

	value, err := cache.Get("temp")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "temporary" {
		t.Errorf("expected 'temporary', got '%s'", value)
	}

	time.Sleep(50 * time.Millisecond)

	_, err = cache.Get("temp")
	if err == nil {
		t.Error("expected error for expired key")
	}
}

func TestBackendNonExistentKey(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 2,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	_, err := cache.Get("does_not_exist")
	if err == nil {
		t.Error("expected error for non-existent key")
	}
	if !strings.Contains(err.Error(), constant.ErrKeyNotFound) {
		t.Errorf("expected key not found error, got: %v", err)
	}
}

func TestBackendDifferentTypes(t *testing.T) {
	type Order struct {
		Customer string
		ID       int
		Total    float64
	}

	orderCache := NewInMemoryBackend[Order](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   20,
			ShardCount: 4,
		},
	})
	defer orderCache.(*inMemoryBackend[Order]).Close()

	order := Order{ID: 456, Customer: "Bob", Total: 99.99}
	orderCache.Set("order:456", order)

	retrieved, err := orderCache.Get("order:456")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved.ID != 456 || retrieved.Customer != "Bob" || retrieved.Total != 99.99 {
		t.Errorf("expected {456 Bob 99.99}, got %+v", retrieved)
	}
}

func TestBackendShardRouting(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   40,
			ShardCount: 4,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	backend := cache.(*inMemoryBackend[string])

	keys := []string{"test1", "test2", "test3", "test4", "test5", "test6"}
	shardUsage := make(map[int]bool)

	for _, key := range keys {
		cache.Set(key, "value")
		shard := backend.getShard(key)

		for i := range backend.shards {
			if &backend.shards[i] == shard {
				shardUsage[i] = true
				break
			}
		}
	}

	if len(shardUsage) == 0 {
		t.Error("no shards were used")
	}
}
