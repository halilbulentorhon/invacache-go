package inmemory

import "time"

type Entry[V any] struct {
	ExpiresAt time.Time
	Value     V
	prev      *Entry[V]
	next      *Entry[V]
	Key       string
}

func (e *Entry[V]) IsExpired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}
