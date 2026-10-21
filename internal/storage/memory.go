package storage

import (
	"context"
	"sync"
	"time"
)

type entry struct {
	count      int
	windowEnd  time.Time
	blockedEnd time.Time
}

// Memory is an in-process limiter.Store, useful for tests and as proof that
// the persistence strategy is swappable.
type Memory struct {
	mu   sync.Mutex
	data map[string]*entry
	now  func() time.Time
}

func NewMemory() *Memory {
	return &Memory{data: map[string]*entry{}, now: time.Now}
}

func (m *Memory) Hit(_ context.Context, key string, limit int, blockFor time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.now()
	e, ok := m.data[key]
	if !ok {
		e = &entry{}
		m.data[key] = e
	}
	if now.Before(e.blockedEnd) {
		return false, nil
	}
	if now.After(e.windowEnd) {
		e.count = 0
		e.windowEnd = now.Add(time.Second)
	}
	e.count++
	if e.count > limit {
		e.blockedEnd = now.Add(blockFor)
		return false, nil
	}
	return true, nil
}
