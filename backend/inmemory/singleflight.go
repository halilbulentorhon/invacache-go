package inmemory

import (
	"fmt"
	"sync"
	"time"
)

type call[V any] struct {
	value V
	err   error
	ttl   time.Duration
	wg    sync.WaitGroup
}

type singleFlight[V any] struct {
	m  map[string]*call[V]
	mu sync.Mutex
}

func (g *singleFlight[V]) Do(key string, fn func() (V, time.Duration, error)) (V, time.Duration, error) {
	g.mu.Lock()
	if g.m == nil {
		g.m = make(map[string]*call[V])
	}
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()
		return c.value, c.ttl, c.err
	}

	c := new(call[V])
	c.wg.Add(1)
	g.m[key] = c
	g.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			var zero V
			c.value, c.ttl, c.err = zero, 0, fmt.Errorf("singleflight panic: %v", r)
		}
		c.wg.Done()
		g.mu.Lock()
		delete(g.m, key)
		g.mu.Unlock()
	}()

	v, ttl, err := fn()
	c.value, c.ttl, c.err = v, ttl, err
	return c.value, c.ttl, c.err
}
