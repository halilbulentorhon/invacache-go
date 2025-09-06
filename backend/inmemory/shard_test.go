package inmemory

import (
	"github.com/halilbulentorhon/invacache-go/backend/option"
	"github.com/halilbulentorhon/invacache-go/config"
	"github.com/halilbulentorhon/invacache-go/constant"
	"strings"
	"testing"
	"time"
)

func TestShardBasicOperations(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	err := cache.Set("key1", "value1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	value, err := cache.Get("key1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}

	err = cache.Delete("key1")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	_, err = cache.Get("key1")
	if err == nil {
		t.Error("expected error after delete")
	}
}

func TestShardExpiration(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("temp", "value", option.WithTTL(30*time.Millisecond))

	value, err := cache.Get("temp")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "value" {
		t.Errorf("expected 'value', got '%s'", value)
	}

	time.Sleep(40 * time.Millisecond)

	_, err = cache.Get("temp")
	if err == nil {
		t.Error("expected error for expired key")
	}
	if !strings.Contains(err.Error(), constant.ErrKeyNotFound) {
		t.Errorf("expected key not found error, got: %v", err)
	}
}

func TestShardCapacityEviction(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   3,
			ShardCount: 1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	_, err := cache.Get("key1")
	if err != nil {
		t.Error("key1 should exist")
	}

	cache.Set("key4", "value4")

	_, err = cache.Get("key2")
	if err == nil {
		t.Error("key2 should have been evicted")
	}

	_, err = cache.Get("key1")
	if err != nil {
		t.Error("key1 should still exist after access")
	}
}

func TestShardLRUBehavior(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   3,
			ShardCount: 1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("first", "1")
	cache.Set("second", "2")
	cache.Set("third", "3")

	cache.Get("first")
	cache.Get("second")

	cache.Set("fourth", "4")

	_, err := cache.Get("third")
	if err == nil {
		t.Error("third should have been evicted as LRU")
	}

	_, err = cache.Get("first")
	if err != nil {
		t.Error("first should still exist")
	}
}

func TestShardUpdateExisting(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("key", "old_value")
	cache.Set("key", "new_value")

	value, err := cache.Get("key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "new_value" {
		t.Errorf("expected 'new_value', got '%s'", value)
	}
}

func TestShardUpdateWithTTL(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("key", "value1")
	cache.Set("key", "value2", option.WithTTL(30*time.Millisecond))

	value, err := cache.Get("key")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "value2" {
		t.Errorf("expected 'value2', got '%s'", value)
	}

	time.Sleep(40 * time.Millisecond)

	_, err = cache.Get("key")
	if err == nil {
		t.Error("key should have expired")
	}
}

func TestShardZeroTTL(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	cache.Set("permanent", "value", option.WithTTL(0))

	time.Sleep(10 * time.Millisecond)

	value, err := cache.Get("permanent")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != "value" {
		t.Errorf("expected 'value', got '%s'", value)
	}
}

func TestShardDeleteNonExistent(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	err := cache.Delete("nonexistent")
	if err != nil {
		t.Errorf("unexpected error for deleting non-existent key: %v", err)
	}
}

func TestShardSweepExpired(t *testing.T) {
	backend := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:        10,
			ShardCount:      1,
			SweeperInterval: 100 * time.Millisecond,
		},
	}).(*inMemoryBackend[string])
	defer backend.Close()

	backend.Set("expire1", "value1", option.WithTTL(30*time.Millisecond))
	backend.Set("expire2", "value2", option.WithTTL(30*time.Millisecond))
	backend.Set("keep", "value3")

	time.Sleep(40 * time.Millisecond)

	shard := &backend.shards[0]
	shard.mu.Lock()
	expiredCount := shard.sweepExpired()
	shard.mu.Unlock()

	if expiredCount != 2 {
		t.Errorf("expected 2 expired entries, got %d", expiredCount)
	}

	_, err := backend.Get("keep")
	if err != nil {
		t.Error("keep should still exist")
	}
}

func TestShardDifferentTypes(t *testing.T) {
	intCache := NewInMemoryBackend[int](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 1,
		},
	})
	defer intCache.(*inMemoryBackend[int]).Close()

	intCache.Set("number", 42)
	value, err := intCache.Get("number")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if value != 42 {
		t.Errorf("expected 42, got %d", value)
	}

	type Product struct {
		Name  string
		ID    int
		Price float64
	}

	productCache := NewInMemoryBackend[Product](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   10,
			ShardCount: 1,
		},
	})
	defer productCache.(*inMemoryBackend[Product]).Close()

	product := Product{ID: 123, Name: "Widget", Price: 19.99}
	productCache.Set("product:123", product)

	retrieved, err := productCache.Get("product:123")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if retrieved.ID != 123 || retrieved.Name != "Widget" || retrieved.Price != 19.99 {
		t.Errorf("expected {123 Widget 19.99}, got %+v", retrieved)
	}
}

func TestShardLinkedListIntegrity(t *testing.T) {
	cache := NewInMemoryBackend[string](config.InvaCacheConfig{
		BackendName: "inmemory",
		InMemory: config.InMemoryConfig{
			Capacity:   5,
			ShardCount: 1,
		},
	})
	defer cache.(*inMemoryBackend[string]).Close()

	backend := cache.(*inMemoryBackend[string])
	shard := &backend.shards[0]

	keys := []string{"a", "b", "c", "d", "e"}
	for _, key := range keys {
		cache.Set(key, "value_"+key)
	}

	shard.mu.Lock()
	current := shard.head.next
	count := 0
	for current != shard.tail {
		count++
		if current.prev.next != current {
			shard.mu.Unlock()
			t.Error("forward link broken")
			return
		}
		if current.next.prev != current {
			shard.mu.Unlock()
			t.Error("backward link broken")
			return
		}
		current = current.next
	}
	shard.mu.Unlock()

	if count != 5 {
		t.Errorf("expected 5 entries in list, found %d", count)
	}
}
