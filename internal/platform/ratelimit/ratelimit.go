// Package ratelimit is a small in-memory fixed-window limiter. It holds no
// external state, consistent with the single-binary constraint (N1). Suitable for
// login throttling on one process; not a distributed limiter.
package ratelimit

import (
	"sync"
	"time"
)

type window struct {
	count int
	reset time.Time
}

// sweepEvery bounds map growth: expired windows are purged once this many
// mutating operations have happened.
const sweepEvery = 256

// Limiter counts failures per key within a fixed time window.
type Limiter struct {
	mu       sync.Mutex
	maxFails int
	window   time.Duration
	now      func() time.Time
	keys     map[string]*window
	ops      int
}

// New builds a limiter allowing maxFails failures per window per key.
func New(maxFails int, w time.Duration) *Limiter {
	return &Limiter{
		maxFails: maxFails,
		window:   w,
		now:      time.Now,
		keys:     make(map[string]*window),
	}
}

// Allowed reports whether key is currently under its failure budget.
func (l *Limiter) Allowed(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	w := l.current(key)
	return w.count < l.maxFails
}

// Fail records a failed attempt for key.
func (l *Limiter) Fail(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.current(key).count++
}

// Reset clears key's counter, e.g. after a successful attempt.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.keys, key)
}

func (l *Limiter) current(key string) *window {
	now := l.now()

	l.ops++
	if l.ops%sweepEvery == 0 {
		for k, w := range l.keys {
			if now.After(w.reset) {
				delete(l.keys, k)
			}
		}
	}

	w, ok := l.keys[key]
	if !ok || now.After(w.reset) {
		w = &window{reset: now.Add(l.window)}
		l.keys[key] = w
	}
	return w
}
