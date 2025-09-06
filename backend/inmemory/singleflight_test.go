package inmemory

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDo_ReturnsValue(t *testing.T) {
	var g singleFlight[int]
	v, ttl, err := g.Do("a", func() (int, time.Duration, error) {
		return 42, 150 * time.Millisecond, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v != 42 {
		t.Fatalf("value=%d", v)
	}
	if ttl != 150*time.Millisecond {
		t.Fatalf("ttl=%v", ttl)
	}
}

func TestDo_DeduplicatesConcurrent(t *testing.T) {
	var g singleFlight[string]
	var calls int32
	start := make(chan struct{})
	const n = 30
	var wg sync.WaitGroup
	wg.Add(n)
	vals := make([]string, n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			v, _, err := g.Do("k", func() (string, time.Duration, error) {
				if atomic.AddInt32(&calls, 1) == 1 {
					time.Sleep(100 * time.Millisecond)
				}
				return "ok", 0, nil
			})
			vals[i], errs[i] = v, err
		}(i)
	}
	close(start)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("calls=%d", got)
	}
	for i := range vals {
		if errs[i] != nil || vals[i] != "ok" {
			t.Fatalf("i=%d v=%q err=%v", i, vals[i], errs[i])
		}
	}
}

func TestDo_ErrorPropagation(t *testing.T) {
	var g singleFlight[int]
	var calls int32
	start := make(chan struct{})
	const n = 16
	var wg sync.WaitGroup
	wg.Add(n)
	errs := make([]error, n)

	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			<-start
			_, _, err := g.Do("e", func() (int, time.Duration, error) {
				atomic.AddInt32(&calls, 1)
				time.Sleep(50 * time.Millisecond)
				return 0, 0, errors.New("boom")
			})
			errs[i] = err
		}(i)
	}
	close(start)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	if got := atomic.LoadInt32(&calls); got != 1 {
		t.Fatalf("calls=%d", got)
	}
	for i := range errs {
		if errs[i] == nil || errs[i].Error() != "boom" {
			t.Fatalf("i=%d err=%v", i, errs[i])
		}
	}
}

func TestDo_DifferentKeysConcurrent(t *testing.T) {
	var g singleFlight[int]
	var a, b int32
	start := make(chan struct{})
	var wg sync.WaitGroup
	const perKey = 12
	wg.Add(2 * perKey)

	for i := 0; i < perKey; i++ {
		go func() {
			defer wg.Done()
			<-start
			_, _, _ = g.Do("A", func() (int, time.Duration, error) {
				if atomic.AddInt32(&a, 1) == 1 {
					time.Sleep(60 * time.Millisecond)
				}
				return 1, 0, nil
			})
		}()
		go func() {
			defer wg.Done()
			<-start
			_, _, _ = g.Do("B", func() (int, time.Duration, error) {
				if atomic.AddInt32(&b, 1) == 1 {
					time.Sleep(60 * time.Millisecond)
				}
				return 2, 0, nil
			})
		}()
	}
	close(start)

	done := make(chan struct{})
	go func() { wg.Wait(); close(done) }()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout")
	}

	if atomic.LoadInt32(&a) != 1 {
		t.Fatalf("a=%d", a)
	}
	if atomic.LoadInt32(&b) != 1 {
		t.Fatalf("b=%d", b)
	}
}

func TestDo_SequentialCallsReexecutes(t *testing.T) {
	var g singleFlight[int]
	var n int32

	v1, _, err1 := g.Do("x", func() (int, time.Duration, error) {
		return int(atomic.AddInt32(&n, 1)), 0, nil
	})
	if err1 != nil || v1 != 1 {
		t.Fatalf("v1=%d err1=%v", v1, err1)
	}

	v2, _, err2 := g.Do("x", func() (int, time.Duration, error) {
		return int(atomic.AddInt32(&n, 1)), 0, nil
	})
	if err2 != nil || v2 != 2 {
		t.Fatalf("v2=%d err2=%v", v2, err2)
	}
}
