package runner

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/LING71671/SurveyController-go/internal/apperr"
	"github.com/LING71671/SurveyController-go/internal/engine"
	"github.com/LING71671/SurveyController-go/internal/logging"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

func TestWorkerPoolRecordsSuccessAndFailure(t *testing.T) {
	events := make(chan logging.RunEvent, 16)
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 2, Target: 3, Events: events})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}
	tasks := []Task{
		func(context.Context, int) error { return nil },
		func(context.Context, int) error { return apperr.New(apperr.CodeFillFailed, "failed") },
		func(context.Context, int) error { return nil },
	}

	snapshot := pool.Run(context.Background(), tasks)
	if snapshot.Successes != 2 || snapshot.Failures != 1 {
		t.Fatalf("snapshot counts = %d/%d, want 2/1", snapshot.Successes, snapshot.Failures)
	}
	if snapshot.LastFailureCode != apperr.CodeFillFailed {
		t.Fatalf("LastFailureCode = %q, want %q", snapshot.LastFailureCode, apperr.CodeFillFailed)
	}
	if !hasEvent(events, logging.EventRunStarted) || !hasEvent(events, logging.EventRunFinished) {
		t.Fatalf("events did not include run start and finish")
	}
}

func TestWorkerPoolEventsIncludeTimestamps(t *testing.T) {
	events := make(chan logging.RunEvent, 16)
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 1, Target: 1, Events: events})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}

	pool.RunSubmissions(context.Background(), []SubmissionTask{
		submissionResultTask(engine.SubmissionResult{
			State:    provider.SubmissionStateSuccess,
			Message:  "done",
			Success:  true,
			Terminal: true,
		}),
	})

	for {
		select {
		case event := <-events:
			if event.Time.IsZero() {
				t.Fatalf("event %q has zero timestamp: %+v", event.Type, event)
			}
		default:
			return
		}
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

func TestWorkerPoolSupportsThousandLightweightWorkers(t *testing.T) {
	pool, err := NewWorkerPool(PoolOptions{Concurrency: DefaultMaxWorkerConcurrency, Target: DefaultMaxWorkerConcurrency})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}
	tasks := make([]Task, 0, DefaultMaxWorkerConcurrency)
	for i := 0; i < DefaultMaxWorkerConcurrency; i++ {
		tasks = append(tasks, func(context.Context, int) error {
			return nil
		})
	}

	snapshot := pool.Run(context.Background(), tasks)
	if snapshot.Successes != DefaultMaxWorkerConcurrency || snapshot.Failures != 0 {
		t.Fatalf("snapshot counts = %d/%d, want %d/0", snapshot.Successes, snapshot.Failures, DefaultMaxWorkerConcurrency)
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

func TestWorkerPoolStartsNoMoreWorkersThanTasks(t *testing.T) {
	events := make(chan logging.RunEvent, 16)
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 10, Target: 2, Events: events})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}

	snapshot := pool.RunSubmissions(context.Background(), []SubmissionTask{
		submissionResultTask(engine.SubmissionResult{State: provider.SubmissionStateSuccess, Success: true}),
		submissionResultTask(engine.SubmissionResult{State: provider.SubmissionStateSuccess, Success: true}),
	})

	if snapshot.Successes != 2 || len(snapshot.Workers) != 2 {
		t.Fatalf("snapshot = %+v, want two successes from two workers", snapshot)
	}
	if got := countEvents(events, logging.EventWorkerStarted); got != 2 {
		t.Fatalf("worker started events = %d, want 2", got)
	}
}

func TestWorkerPoolHandlesEmptyTaskListWithoutWorkers(t *testing.T) {
	events := make(chan logging.RunEvent, 8)
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 10, Events: events})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}

	snapshot := pool.RunSubmissions(context.Background(), nil)
	if snapshot.Successes != 0 || snapshot.Failures != 0 || len(snapshot.Workers) != 0 {
		t.Fatalf("snapshot = %+v, want no work recorded", snapshot)
	}
	if got := countEvents(events, logging.EventWorkerStarted); got != 0 {
		t.Fatalf("worker started events = %d, want 0", got)
	}
}

func TestWorkerPoolRunSubmissionsRecordsResults(t *testing.T) {
	events := make(chan logging.RunEvent, 16)
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 1, Target: 3, Events: events})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}
	tasks := []SubmissionTask{
		submissionResultTask(engine.SubmissionResult{
			State:    provider.SubmissionStateSuccess,
			Message:  "done",
			Success:  true,
			Terminal: true,
		}),
		submissionResultTask(engine.SubmissionResult{
			State:   provider.SubmissionStateUnknown,
			Message: "waiting",
		}),
		submissionResultTask(engine.SubmissionResult{
			State:    provider.SubmissionStateFailure,
			Message:  "failed",
			Terminal: true,
			Error:    apperr.New(apperr.CodeSubmitFailed, "failed"),
		}),
	}

	snapshot := pool.RunSubmissions(context.Background(), tasks)
	if snapshot.Successes != 1 || snapshot.Failures != 1 {
		t.Fatalf("snapshot counts = %d/%d, want 1/1", snapshot.Successes, snapshot.Failures)
	}
	if snapshot.Workers[1].Message != "worker stopped" {
		t.Fatalf("worker progress = %+v, want stopped message", snapshot.Workers[1])
	}
	if !hasEvent(events, logging.EventSubmissionSuccess) {
		t.Fatalf("events did not include submission success")
	}
	if !hasEvent(events, logging.EventWorkerProgress) {
		t.Fatalf("events did not include worker progress")
	}
	if !hasEvent(events, logging.EventSubmissionFailure) {
		t.Fatalf("events did not include submission failure")
	}
}

func TestWorkerPoolRunSubmissionsStopsOnSubmissionSignal(t *testing.T) {
	events := make(chan logging.RunEvent, 16)
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 1, Target: 10, FailureThreshold: 10, Events: events})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}
	var calls int32
	tasks := []SubmissionTask{
		func(context.Context, int) (engine.SubmissionResult, error) {
			atomic.AddInt32(&calls, 1)
			return engine.SubmissionResult{
				State:      provider.SubmissionStateVerificationRequired,
				Message:    "captcha",
				Terminal:   true,
				ShouldStop: true,
				Error:      apperr.New(apperr.CodeVerificationNeeded, "captcha"),
			}, nil
		},
		func(context.Context, int) (engine.SubmissionResult, error) {
			atomic.AddInt32(&calls, 1)
			return engine.SubmissionResult{State: provider.SubmissionStateSuccess, Success: true}, nil
		},
	}

	snapshot := pool.RunSubmissions(context.Background(), tasks)
	if calls != 1 {
		t.Fatalf("calls = %d, want 1", calls)
	}
	if snapshot.Failures != 1 || !snapshot.StopRequested || snapshot.StopReason != "captcha" {
		t.Fatalf("snapshot = %+v, want stop after verification", snapshot)
	}
	if !hasEvent(events, logging.EventVerificationNeeded) {
		t.Fatalf("events did not include verification needed")
	}
}

func TestWorkerPoolRunSubmissionsRecordsTaskError(t *testing.T) {
	events := make(chan logging.RunEvent, 16)
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 1, Events: events})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}

	snapshot := pool.RunSubmissions(context.Background(), []SubmissionTask{
		func(context.Context, int) (engine.SubmissionResult, error) {
			return engine.SubmissionResult{}, apperr.Wrap(apperr.CodeBrowserStartFailed, "browser crashed", errors.New("spawn"))
		},
	})

	if snapshot.Successes != 0 || snapshot.Failures != 1 {
		t.Fatalf("counts = %d/%d, want 0/1", snapshot.Successes, snapshot.Failures)
	}
	if snapshot.LastFailureCode != apperr.CodeBrowserStartFailed {
		t.Fatalf("LastFailureCode = %q, want %q", snapshot.LastFailureCode, apperr.CodeBrowserStartFailed)
	}
	event, ok := findEvent(events, logging.EventSubmissionFailure)
	if !ok {
		t.Fatalf("events did not include submission failure")
	}
	if event.Fields["error_code"] != string(apperr.CodeBrowserStartFailed) {
		t.Fatalf("error_code field = %v, want %q", event.Fields["error_code"], apperr.CodeBrowserStartFailed)
	}
}

func TestNewWorkerPoolRejectsInvalidOptions(t *testing.T) {
	if _, err := NewWorkerPool(PoolOptions{}); err == nil {
		t.Fatal("NewWorkerPool(empty) returned nil error, want failure")
	}
	if _, err := NewWorkerPool(PoolOptions{Concurrency: DefaultMaxWorkerConcurrency + 1}); err == nil {
		t.Fatal("NewWorkerPool(too much concurrency) returned nil error, want failure")
	}
	if _, err := NewWorkerPool(PoolOptions{Concurrency: 1, Target: -1}); err == nil {
		t.Fatal("NewWorkerPool(negative target) returned nil error, want failure")
	}
}

func submissionResultTask(result engine.SubmissionResult) SubmissionTask {
	return func(context.Context, int) (engine.SubmissionResult, error) {
		return result, nil
	}
}

func hasEvent(events <-chan logging.RunEvent, eventType logging.EventType) bool {
	_, ok := findEvent(events, eventType)
	return ok
}

func findEvent(events <-chan logging.RunEvent, eventType logging.EventType) (logging.RunEvent, bool) {
	for {
		select {
		case event := <-events:
			if event.Type == eventType {
				return event, true
			}
		default:
			return logging.RunEvent{}, false
		}
	}
}

func countEvents(events <-chan logging.RunEvent, eventType logging.EventType) int {
	count := 0
	for {
		select {
		case event := <-events:
			if event.Type == eventType {
				count++
			}
		default:
			return count
		}
	}
}
