package inmemory

import (
	"sync"
	"testing"
	"time"

	"github.com/halilbulentorhon/invacache-go/backend/option"
	"github.com/halilbulentorhon/invacache-go/constant"
)

func TestNewInMemoryShard(t *testing.T) {
	shard := newInMemoryShard[string](10)
	if shard.capacity != 10 {
		t.Errorf("expected capacity 10, got %d", shard.capacity)
	}
	if shard.count != 0 {
		t.Errorf("expected count 0, got %d", shard.count)
	}
	if shard.items == nil {
		t.Error("items map should be initialized")
	}
	if shard.head == nil || shard.tail == nil {
		t.Error("head and tail should be initialized")
	}
	if shard.head.next != shard.tail {
		t.Error("head.next should point to tail")
	}
	if shard.tail.prev != shard.head {
		t.Error("tail.prev should point to head")
	}
}

func TestShardGetNonExistentKey(t *testing.T) {
	shard := newInMemoryShard[string](10)

	_, err := shard.get("nonexistent")
	if err == nil {
		t.Fatal("expected error for non-existent key")
	}
	if err.Error() != constant.ErrKeyNotFound+": nonexistent" {
		t.Errorf("expected 'key not found' error, got: %v", err)
	}
}

func TestShardSetAndGet(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	value, err := shard.get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}

	if shard.count != 1 {
		t.Errorf("expected count 1, got %d", shard.count)
	}
}

func TestShardSetWithTTL(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.set("key1", "value1", option.WithTTL(100*time.Millisecond))
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	value, err := shard.get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}

	time.Sleep(150 * time.Millisecond)

	_, err = shard.get("key1")
	if err == nil {
		t.Fatal("expected error for expired key")
	}

	if shard.count != 0 {
		t.Errorf("expected count 0 after expiration, got %d", shard.count)
	}
}

func TestShardSetWithNoExpiration(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.set("key1", "value1", option.WithNoExpiration())
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	time.Sleep(50 * time.Millisecond)

	value, err := shard.get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value1" {
		t.Errorf("expected 'value1', got '%s'", value)
	}
}

func TestShardUpdateExistingKey(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	err = shard.set("key1", "value2")
	if err != nil {
		t.Fatalf("unexpected error updating key: %v", err)
	}

	value, err := shard.get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if value != "value2" {
		t.Errorf("expected 'value2', got '%s'", value)
	}

	if shard.count != 1 {
		t.Errorf("expected count 1, got %d", shard.count)
	}
}

func TestShardDelete(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	err = shard.delete("key1")
	if err != nil {
		t.Fatalf("unexpected error deleting key: %v", err)
	}

	_, err = shard.get("key1")
	if err == nil {
		t.Fatal("expected error for deleted key")
	}

	if shard.count != 0 {
		t.Errorf("expected count 0, got %d", shard.count)
	}
}

func TestShardDeleteNonExistentKey(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.delete("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error deleting non-existent key: %v", err)
	}
}

func TestShardCapacityEviction(t *testing.T) {
	shard := newInMemoryShard[string](2)

	err := shard.set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = shard.set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	err = shard.set("key3", "value3")
	if err != nil {
		t.Fatalf("unexpected error setting key3: %v", err)
	}

	_, err = shard.get("key1")
	if err == nil {
		t.Fatal("key1 should have been evicted")
	}

	_, err = shard.get("key2")
	if err != nil {
		t.Fatalf("unexpected error getting key2: %v", err)
	}

	_, err = shard.get("key3")
	if err != nil {
		t.Fatalf("unexpected error getting key3: %v", err)
	}

	if shard.count != 2 {
		t.Errorf("expected count 2, got %d", shard.count)
	}
}

func TestShardLRUBehavior(t *testing.T) {
	shard := newInMemoryShard[string](2)

	err := shard.set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = shard.set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	_, err = shard.get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key1: %v", err)
	}

	err = shard.set("key3", "value3")
	if err != nil {
		t.Fatalf("unexpected error setting key3: %v", err)
	}

	_, err = shard.get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key1: %v", err)
	}

	_, err = shard.get("key2")
	if err == nil {
		t.Fatal("key2 should have been evicted")
	}
}

func TestShardSweepExpired(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.set("key1", "value1", option.WithTTL(50*time.Millisecond))
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = shard.set("key2", "value2", option.WithTTL(200*time.Millisecond))
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	err = shard.set("key3", "value3")
	if err != nil {
		t.Fatalf("unexpected error setting key3: %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	expiredCount := shard.sweepExpired()
	if expiredCount != 1 {
		t.Errorf("expected 1 expired entry, got %d", expiredCount)
	}

	_, err = shard.get("key1")
	if err == nil {
		t.Fatal("key1 should have been swept")
	}

	_, err = shard.get("key2")
	if err != nil {
		t.Fatalf("unexpected error getting key2: %v", err)
	}

	_, err = shard.get("key3")
	if err != nil {
		t.Fatalf("unexpected error getting key3: %v", err)
	}

	if shard.count != 2 {
		t.Errorf("expected count 2, got %d", shard.count)
	}
}

func TestShardSweepExpiredNoExpired(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = shard.set("key2", "value2", option.WithTTL(1*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	expiredCount := shard.sweepExpired()
	if expiredCount != 0 {
		t.Errorf("expected 0 expired entries, got %d", expiredCount)
	}

	if shard.count != 2 {
		t.Errorf("expected count 2, got %d", shard.count)
	}
}

func TestShardConcurrentAccess(t *testing.T) {
	shard := newInMemoryShard[int](100)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "key" + string(rune(i%10))
			shard.mu.Lock()
			err := shard.set(key, i)
			shard.mu.Unlock()
			if err != nil {
				t.Errorf("unexpected error setting key: %v", err)
			}
		}(i)
	}

	wg.Wait()

	for i := 0; i < 10; i++ {
		key := "key" + string(rune(i))
		shard.mu.Lock()
		_, err := shard.get(key)
		shard.mu.Unlock()
		if err != nil {
			t.Errorf("unexpected error getting key %s: %v", key, err)
		}
	}
}

func TestShardLinkedListOperations(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = shard.set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	err = shard.set("key3", "value3")
	if err != nil {
		t.Fatalf("unexpected error setting key3: %v", err)
	}

	_, err = shard.get("key1")
	if err != nil {
		t.Fatalf("unexpected error getting key1: %v", err)
	}

	if shard.head.next.Key != "key1" {
		t.Errorf("expected key1 to be at head, got %s", shard.head.next.Key)
	}

	if shard.tail.prev.Key != "key2" {
		t.Errorf("expected key2 to be at tail, got %s", shard.tail.prev.Key)
	}
}

func TestShardRemoveTail(t *testing.T) {
	shard := newInMemoryShard[string](10)

	err := shard.set("key1", "value1")
	if err != nil {
		t.Fatalf("unexpected error setting key1: %v", err)
	}

	err = shard.set("key2", "value2")
	if err != nil {
		t.Fatalf("unexpected error setting key2: %v", err)
	}

	tailEntry := shard.removeTail()
	if tailEntry == nil {
		t.Fatal("expected non-nil tail entry")
	}
	if tailEntry.Key != "key1" {
		t.Errorf("expected key1, got %s", tailEntry.Key)
	}

	if shard.count != 2 {
		t.Errorf("expected count 2, got %d", shard.count)
	}
}

func TestShardRemoveTailEmpty(t *testing.T) {
	shard := newInMemoryShard[string](10)

	tailEntry := shard.removeTail()
	if tailEntry != nil {
		t.Fatal("expected nil tail entry for empty shard")
	}
}

func TestShardDifferentTypes(t *testing.T) {
	stringShard := newInMemoryShard[string](10)
	intShard := newInMemoryShard[int](10)

	err := stringShard.set("str", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = intShard.set("int", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	strValue, err := stringShard.get("str")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strValue != "hello" {
		t.Errorf("expected 'hello', got '%s'", strValue)
	}

	intValue, err := intShard.get("int")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intValue != 42 {
		t.Errorf("expected 42, got %d", intValue)
	}
}
