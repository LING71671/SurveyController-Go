package parsecache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Entry[T any] struct {
	Key       string
	URL       string
	Value     T
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (e Entry[T]) Expired(now time.Time) bool {
	return !e.ExpiresAt.IsZero() && !now.Before(e.ExpiresAt)
}

type Store[T any] interface {
	Get(ctx context.Context, key string) (Entry[T], bool, error)
	Put(ctx context.Context, entry Entry[T]) error
	Delete(ctx context.Context, key string) error
}

type Cache[T any] struct {
	store Store[T]
	ttl   time.Duration
	now   func() time.Time
}

func New[T any](store Store[T], ttl time.Duration) (*Cache[T], error) {
	if store == nil {
		return nil, fmt.Errorf("store is required")
	}
	if ttl < 0 {
		return nil, fmt.Errorf("ttl must not be negative")
	}
	return &Cache[T]{
		store: store,
		ttl:   ttl,
		now:   time.Now,
	}, nil
}

func (c *Cache[T]) Get(ctx context.Context, rawURL string) (T, bool, error) {
	var zero T
	key, normalized, err := cacheKey(rawURL)
	if err != nil {
		return zero, false, err
	}
	entry, ok, err := c.store.Get(ctx, key)
	if err != nil || !ok {
		return zero, ok, err
	}
	if entry.URL != normalized || entry.Expired(c.now()) {
		if err := c.store.Delete(ctx, key); err != nil {
			return zero, false, err
		}
		return zero, false, nil
	}
	return entry.Value, true, nil
}

func (c *Cache[T]) Put(ctx context.Context, rawURL string, value T) error {
	key, normalized, err := cacheKey(rawURL)
	if err != nil {
		return err
	}
	now := c.now()
	entry := Entry[T]{
		Key:       key,
		URL:       normalized,
		Value:     value,
		CreatedAt: now,
	}
	if c.ttl > 0 {
		entry.ExpiresAt = now.Add(c.ttl)
	}
	return c.store.Put(ctx, entry)
}

func cacheKey(rawURL string) (string, string, error) {
	normalized, err := NormalizeURL(rawURL)
	if err != nil {
		return "", "", err
	}
	key, err := Fingerprint(normalized)
	if err != nil {
		return "", "", err
	}
	return key, normalized, nil
}

type MemoryStore[T any] struct {
	mu      sync.RWMutex
	entries map[string]Entry[T]
}

func NewMemoryStore[T any]() *MemoryStore[T] {
	return &MemoryStore[T]{
		entries: map[string]Entry[T]{},
	}
}

func (s *MemoryStore[T]) Get(ctx context.Context, key string) (Entry[T], bool, error) {
	if err := ctx.Err(); err != nil {
		var zero Entry[T]
		return zero, false, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.entries[key]
	return entry, ok, nil
}

func (s *MemoryStore[T]) Put(ctx context.Context, entry Entry[T]) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.entries[entry.Key] = entry
	return nil
}

func (s *MemoryStore[T]) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.entries, key)
	return nil
}

func (s *MemoryStore[T]) Snapshot() map[string]Entry[T] {
	s.mu.RLock()
	defer s.mu.RUnlock()

	snapshot := make(map[string]Entry[T], len(s.entries))
	for key, entry := range s.entries {
		snapshot[key] = entry
	}
	return snapshot
}
