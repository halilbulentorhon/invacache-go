package inmemory

import (
	"testing"
	"time"
)

func TestEntryIsExpiredWithZeroTime(t *testing.T) {
	entry := &Entry[string]{
		Key:       "test",
		Value:     "value",
		ExpiresAt: time.Time{},
	}

	if entry.IsExpired() {
		t.Error("entry with zero expiration time should not be expired")
	}
}

func TestEntryIsExpiredWithFutureTime(t *testing.T) {
	entry := &Entry[string]{
		Key:       "test",
		Value:     "value",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	if entry.IsExpired() {
		t.Error("entry with future expiration time should not be expired")
	}
}

func TestEntryIsExpiredWithPastTime(t *testing.T) {
	entry := &Entry[string]{
		Key:       "test",
		Value:     "value",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	if !entry.IsExpired() {
		t.Error("entry with past expiration time should be expired")
	}
}

func TestEntryIsExpiredWithVeryRecentPastTime(t *testing.T) {
	entry := &Entry[string]{
		Key:       "test",
		Value:     "value",
		ExpiresAt: time.Now().Add(-1 * time.Millisecond),
	}

	time.Sleep(2 * time.Millisecond)

	if !entry.IsExpired() {
		t.Error("entry with recent past expiration time should be expired")
	}
}

func TestEntryCreation(t *testing.T) {
	now := time.Now()
	entry := &Entry[int]{
		Key:       "number",
		Value:     42,
		ExpiresAt: now,
	}

	if entry.Key != "number" {
		t.Errorf("expected key 'number', got '%s'", entry.Key)
	}
	if entry.Value != 42 {
		t.Errorf("expected value 42, got %d", entry.Value)
	}
	if !entry.ExpiresAt.Equal(now) {
		t.Errorf("expected expiration time %v, got %v", now, entry.ExpiresAt)
	}
}

func TestEntryLinkedListPointers(t *testing.T) {
	entry1 := &Entry[string]{Key: "first", Value: "value1"}
	entry2 := &Entry[string]{Key: "second", Value: "value2"}
	entry3 := &Entry[string]{Key: "third", Value: "value3"}

	entry1.next = entry2
	entry2.prev = entry1
	entry2.next = entry3
	entry3.prev = entry2

	if entry1.next != entry2 {
		t.Error("entry1.next should point to entry2")
	}
	if entry2.prev != entry1 {
		t.Error("entry2.prev should point to entry1")
	}
	if entry2.next != entry3 {
		t.Error("entry2.next should point to entry3")
	}
	if entry3.prev != entry2 {
		t.Error("entry3.prev should point to entry2")
	}
}

func TestEntryWithDifferentTypes(t *testing.T) {
	stringEntry := &Entry[string]{
		Key:   "text",
		Value: "hello world",
	}
	if stringEntry.Value != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", stringEntry.Value)
	}

	intEntry := &Entry[int]{
		Key:   "number",
		Value: 100,
	}
	if intEntry.Value != 100 {
		t.Errorf("expected 100, got %d", intEntry.Value)
	}

	type CustomType struct {
		Name string
		Age  int
	}
	customEntry := &Entry[CustomType]{
		Key:   "person",
		Value: CustomType{Name: "Alice", Age: 25},
	}
	if customEntry.Value.Name != "Alice" || customEntry.Value.Age != 25 {
		t.Errorf("expected {Alice 25}, got %+v", customEntry.Value)
	}
}

func TestEntryIsExpiredEdgeCase(t *testing.T) {
	now := time.Now()
	entry := &Entry[string]{
		Key:       "edge",
		Value:     "test",
		ExpiresAt: now,
	}

	time.Sleep(1 * time.Nanosecond)

	if !entry.IsExpired() {
		t.Error("entry should be expired when time passes expiration exactly")
	}
}
