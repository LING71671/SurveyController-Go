package parsecache

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCachePutGet(t *testing.T) {
	cache, err := New[string](NewMemoryStore[string](), time.Minute)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := cache.Put(context.Background(), "https://example.com:443/path?b=2&a=1", "value"); err != nil {
		t.Fatalf("Put() returned error: %v", err)
	}
	got, ok, err := cache.Get(context.Background(), "HTTPS://EXAMPLE.COM/path?a=1&b=2#ignored")
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if !ok || got != "value" {
		t.Fatalf("Get() = %q, %v, want value true", got, ok)
	}
}

func TestCacheExpiresEntries(t *testing.T) {
	store := NewMemoryStore[string]()
	cache, err := New[string](store, time.Minute)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	now := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return now }

	if err := cache.Put(context.Background(), "https://example.com", "value"); err != nil {
		t.Fatalf("Put() returned error: %v", err)
	}
	cache.now = func() time.Time { return now.Add(time.Minute) }
	_, ok, err := cache.Get(context.Background(), "https://example.com/")
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if ok {
		t.Fatalf("Get() ok = true, want expired entry miss")
	}
	if len(store.Snapshot()) != 0 {
		t.Fatalf("expired entry was not deleted")
	}
}

func TestCacheWithZeroTTLDoesNotExpire(t *testing.T) {
	cache, err := New[string](NewMemoryStore[string](), 0)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	now := time.Date(2026, 4, 30, 10, 0, 0, 0, time.UTC)
	cache.now = func() time.Time { return now }

	if err := cache.Put(context.Background(), "https://example.com", "value"); err != nil {
		t.Fatalf("Put() returned error: %v", err)
	}
	cache.now = func() time.Time { return now.Add(365 * 24 * time.Hour) }
	_, ok, err := cache.Get(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	if !ok {
		t.Fatalf("Get() ok = false, want non-expiring hit")
	}
}

func TestMemoryStoreSnapshotClonesMap(t *testing.T) {
	store := NewMemoryStore[string]()
	entry := Entry[string]{Key: "key", URL: "https://example.com", Value: "value"}
	if err := store.Put(context.Background(), entry); err != nil {
		t.Fatalf("Put() returned error: %v", err)
	}
	snapshot := store.Snapshot()
	delete(snapshot, "key")

	next := store.Snapshot()
	if len(next) != 1 {
		t.Fatalf("snapshot mutation affected store")
	}
}

func TestCachePropagatesContextCancellation(t *testing.T) {
	cache, err := New[string](NewMemoryStore[string](), time.Minute)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = cache.Put(ctx, "https://example.com", "value")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Put() error = %v, want context.Canceled", err)
	}
}

func TestNewRejectsInvalidOptions(t *testing.T) {
	if _, err := New[string](nil, time.Minute); err == nil {
		t.Fatal("New(nil store) returned nil error, want failure")
	}
	if _, err := New[string](NewMemoryStore[string](), -time.Second); err == nil {
		t.Fatal("New(negative ttl) returned nil error, want failure")
	}
}
