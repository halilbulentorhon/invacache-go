package inmemory

import (
	"fmt"
	"sync"
	"time"

	"github.com/halilbulentorhon/invacache-go/backend/option"
	"github.com/halilbulentorhon/invacache-go/constant"
)

type inMemoryShard[V any] struct {
	items      map[string]*Entry[V]
	head, tail *Entry[V]
	mu         sync.RWMutex
	count      int
	capacity   int
	defaultTTL time.Duration
}

func newInMemoryShard[V any](capacity int, defaultTTL time.Duration) inMemoryShard[V] {
	head := &Entry[V]{}
	tail := &Entry[V]{}
	head.next = tail
	tail.prev = head

	return inMemoryShard[V]{
		items:      make(map[string]*Entry[V]),
		head:       head,
		tail:       tail,
		capacity:   capacity,
		defaultTTL: defaultTTL,
	}
}

func (s *inMemoryShard[V]) get(key string) (V, error) {
	entry, exists := s.items[key]
	if !exists {
		var zero V
		return zero, fmt.Errorf("%s: %s", constant.ErrKeyNotFound, key)
	}

	if entry.IsExpired() {
		s.removeEntry(entry)
		delete(s.items, key)
		s.count--
		var zero V
		return zero, fmt.Errorf("%s: %s", constant.ErrKeyNotFound, key)
	}

	s.moveToHead(entry)
	return entry.Value, nil
}

func (s *inMemoryShard[V]) set(key string, value V, options ...option.OptFnc) error {
	cfg := option.ApplyOptions(options)

	var expiresAt time.Time
	if !cfg.NoExpiration {
		ttl := cfg.TTL
		if ttl == 0 {
			ttl = s.defaultTTL
		}
		if ttl > 0 {
			expiresAt = time.Now().Add(ttl)
		}
	}

	if existingEntry, exists := s.items[key]; exists {
		existingEntry.Value = value
		existingEntry.ExpiresAt = expiresAt
		s.moveToHead(existingEntry)
		return nil
	}

	newEntry := &Entry[V]{
		Value:     value,
		ExpiresAt: expiresAt,
		Key:       key,
	}

	for s.count >= s.capacity {
		tailEntry := s.removeTail()
		if tailEntry != nil {
			delete(s.items, tailEntry.Key)
			s.count--
		}
	}

	s.items[key] = newEntry
	s.addToHead(newEntry)
	s.count++

	return nil
}

func (s *inMemoryShard[V]) delete(key string) error {
	if entry, exists := s.items[key]; exists {
		s.removeEntry(entry)
		delete(s.items, key)
		s.count--
	}

	return nil
}

func (s *inMemoryShard[V]) sweepExpired() int {
	var expiredKeys []string

	for key, entry := range s.items {
		if entry.IsExpired() {
			expiredKeys = append(expiredKeys, key)
		}
	}

	for _, key := range expiredKeys {
		if entry := s.items[key]; entry != nil {
			s.removeEntry(entry)
			delete(s.items, key)
			s.count--
		}
	}

	return len(expiredKeys)
}

func (s *inMemoryShard[V]) addToHead(entry *Entry[V]) {
	entry.prev = s.head
	entry.next = s.head.next
	s.head.next.prev = entry
	s.head.next = entry
}

func (s *inMemoryShard[V]) removeEntry(entry *Entry[V]) {
	entry.prev.next = entry.next
	entry.next.prev = entry.prev
}

func (s *inMemoryShard[V]) moveToHead(entry *Entry[V]) {
	s.removeEntry(entry)
	s.addToHead(entry)
}

func (s *inMemoryShard[V]) removeTail() *Entry[V] {
	lastEntry := s.tail.prev
	if lastEntry == s.head {
		return nil
	}
	s.removeEntry(lastEntry)
	return lastEntry
}

func (s *inMemoryShard[V]) clear() {
	s.items = make(map[string]*Entry[V])
	head := &Entry[V]{}
	tail := &Entry[V]{}
	head.next = tail
	tail.prev = head
	s.head = head
	s.tail = tail
	s.count = 0
}
