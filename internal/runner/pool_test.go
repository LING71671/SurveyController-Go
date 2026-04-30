package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/LING71671/SurveyController-go/internal/logging"
)

func TestWorkerPoolRecordsSuccessAndFailure(t *testing.T) {
	events := make(chan logging.RunEvent, 16)
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 2, Target: 3, Events: events})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}
	tasks := []Task{
		func(context.Context, int) error { return nil },
		func(context.Context, int) error { return errors.New("failed") },
		func(context.Context, int) error { return nil },
	}

	snapshot := pool.Run(context.Background(), tasks)
	if snapshot.Successes != 2 || snapshot.Failures != 1 {
		t.Fatalf("snapshot counts = %d/%d, want 2/1", snapshot.Successes, snapshot.Failures)
	}
	if !hasEvent(events, logging.EventRunStarted) || !hasEvent(events, logging.EventRunFinished) {
		t.Fatalf("events did not include run start and finish")
	}
}

func TestWorkerPoolHonorsConcurrency(t *testing.T) {
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 2})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}
	var current int32
	var maxSeen int32
	task := func(context.Context, int) error {
		now := atomic.AddInt32(&current, 1)
		for {
			seen := atomic.LoadInt32(&maxSeen)
			if now <= seen || atomic.CompareAndSwapInt32(&maxSeen, seen, now) {
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&current, -1)
		return nil
	}

	snapshot := pool.Run(context.Background(), []Task{task, task, task, task})
	if snapshot.Successes != 4 {
		t.Fatalf("Successes = %d, want 4", snapshot.Successes)
	}
	if maxSeen > 2 {
		t.Fatalf("max concurrency = %d, want <= 2", maxSeen)
	}
}

func TestWorkerPoolStopsAtTarget(t *testing.T) {
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 1, Target: 2})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}
	var calls int32
	task := func(context.Context, int) error {
		atomic.AddInt32(&calls, 1)
		return nil
	}

	snapshot := pool.Run(context.Background(), []Task{task, task, task, task})
	if snapshot.Successes != 2 {
		t.Fatalf("Successes = %d, want 2", snapshot.Successes)
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want 2", calls)
	}
}

func TestWorkerPoolStopsOnContextCancel(t *testing.T) {
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 1})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	var calls int32
	task := func(context.Context, int) error {
		atomic.AddInt32(&calls, 1)
		cancel()
		return nil
	}

	snapshot := pool.Run(ctx, []Task{task, task, task})
	if snapshot.Successes != 1 {
		t.Fatalf("Successes = %d, want 1", snapshot.Successes)
	}
}

func TestNewWorkerPoolRejectsInvalidOptions(t *testing.T) {
	if _, err := NewWorkerPool(PoolOptions{}); err == nil {
		t.Fatal("NewWorkerPool(empty) returned nil error, want failure")
	}
	if _, err := NewWorkerPool(PoolOptions{Concurrency: 1, Target: -1}); err == nil {
		t.Fatal("NewWorkerPool(negative target) returned nil error, want failure")
	}
}

func hasEvent(events <-chan logging.RunEvent, eventType logging.EventType) bool {
	for {
		select {
		case event := <-events:
			if event.Type == eventType {
				return true
			}
		default:
			return false
		}
	}
}
