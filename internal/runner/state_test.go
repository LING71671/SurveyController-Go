package runner

import (
	"sync"
	"testing"
)

func TestRunStateRecordsSuccessAndFailure(t *testing.T) {
	state := NewRunState(StateOptions{Target: 2, FailureThreshold: 3})

	state.RecordSuccess(1)
	state.RecordFailure(2, "failed")
	snapshot := state.Snapshot()

	if snapshot.Successes != 1 || snapshot.Failures != 1 {
		t.Fatalf("snapshot counts = %d/%d, want 1/1", snapshot.Successes, snapshot.Failures)
	}
	if snapshot.Workers[1].Successes != 1 {
		t.Fatalf("worker 1 successes = %d, want 1", snapshot.Workers[1].Successes)
	}
	if snapshot.Workers[2].Failures != 1 || snapshot.Workers[2].Message != "failed" {
		t.Fatalf("worker 2 progress = %+v, want failure message", snapshot.Workers[2])
	}
}

func TestRunStateSnapshotClonesWorkers(t *testing.T) {
	state := NewRunState(StateOptions{})
	state.SetWorkerStatus(1, WorkerStatusRunning, "working")

	snapshot := state.Snapshot()
	snapshot.Workers[1] = WorkerProgress{ID: 1, Status: WorkerStatusStopped}

	next := state.Snapshot()
	if next.Workers[1].Status != WorkerStatusRunning {
		t.Fatalf("internal worker map was mutated through snapshot")
	}
}

func TestRunStateShouldStop(t *testing.T) {
	state := NewRunState(StateOptions{Target: 2, FailureThreshold: 2})
	if state.ShouldStop() {
		t.Fatalf("ShouldStop() = true before any result, want false")
	}

	state.RecordSuccess(1)
	state.RecordSuccess(1)
	if !state.ShouldStop() {
		t.Fatalf("ShouldStop() = false after target reached, want true")
	}

	state = NewRunState(StateOptions{Target: 10, FailureThreshold: 1})
	state.RecordFailure(1, "failed")
	if !state.ShouldStop() {
		t.Fatalf("ShouldStop() = false after failure threshold reached, want true")
	}

	state = NewRunState(StateOptions{Target: 10, FailureThreshold: 10})
	state.RequestStop("verification required")
	if !state.ShouldStop() {
		t.Fatalf("ShouldStop() = false after explicit stop, want true")
	}
	snapshot := state.Snapshot()
	if !snapshot.StopRequested || snapshot.StopReason != "verification required" {
		t.Fatalf("explicit stop snapshot = %+v, want stop reason", snapshot)
	}
}

func TestRunStateIsConcurrentSafe(t *testing.T) {
	state := NewRunState(StateOptions{Target: 100})
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			state.RecordSuccess(workerID)
		}(i % 4)
	}
	wg.Wait()

	snapshot := state.Snapshot()
	if snapshot.Successes != 100 {
		t.Fatalf("Successes = %d, want 100", snapshot.Successes)
	}
	if len(snapshot.Workers) != 4 {
		t.Fatalf("len(Workers) = %d, want 4", len(snapshot.Workers))
	}
}
