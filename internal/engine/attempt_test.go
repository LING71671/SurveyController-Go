package engine

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestExecutionAttemptCommitUsesResourceOrder(t *testing.T) {
	var calls []string
	attempt := NewExecutionAttempt(2)
	mustAddResource(t, attempt, recordingResource{name: "proxy", calls: &calls})
	mustAddResource(t, attempt, recordingResource{name: "sample", calls: &calls})

	if err := attempt.Commit(context.Background()); err != nil {
		t.Fatalf("Commit() returned error: %v", err)
	}

	want := []string{"commit:proxy", "commit:sample"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	snapshot := attempt.Snapshot()
	if !snapshot.Finalized || !snapshot.Committed || snapshot.RolledBack {
		t.Fatalf("snapshot = %+v, want committed finalized attempt", snapshot)
	}
}

func TestExecutionAttemptRollbackUsesReverseResourceOrder(t *testing.T) {
	var calls []string
	attempt := NewExecutionAttempt(3)
	mustAddResource(t, attempt, recordingResource{name: "browser", calls: &calls})
	mustAddResource(t, attempt, recordingResource{name: "proxy", calls: &calls})
	mustAddResource(t, attempt, recordingResource{name: "sample", calls: &calls})

	if err := attempt.Rollback(context.Background()); err != nil {
		t.Fatalf("Rollback() returned error: %v", err)
	}

	want := []string{"rollback:sample", "rollback:proxy", "rollback:browser"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	snapshot := attempt.Snapshot()
	if !snapshot.Finalized || snapshot.Committed || !snapshot.RolledBack {
		t.Fatalf("snapshot = %+v, want rolled back finalized attempt", snapshot)
	}
}

func TestExecutionAttemptFinalizeUsesSubmissionResult(t *testing.T) {
	var successCalls []string
	success := NewExecutionAttempt(1)
	mustAddResource(t, success, recordingResource{name: "sample", calls: &successCalls})
	if err := success.Finalize(context.Background(), SubmissionResult{Success: true}); err != nil {
		t.Fatalf("Finalize(success) returned error: %v", err)
	}
	if !reflect.DeepEqual(successCalls, []string{"commit:sample"}) {
		t.Fatalf("success calls = %v, want commit", successCalls)
	}

	var failureCalls []string
	failure := NewExecutionAttempt(1)
	mustAddResource(t, failure, recordingResource{name: "sample", calls: &failureCalls})
	if err := failure.Finalize(context.Background(), SubmissionResult{}); err != nil {
		t.Fatalf("Finalize(failure) returned error: %v", err)
	}
	if !reflect.DeepEqual(failureCalls, []string{"rollback:sample"}) {
		t.Fatalf("failure calls = %v, want rollback", failureCalls)
	}
}

func TestExecutionAttemptCommitFailureRollsBackCommittedResources(t *testing.T) {
	var calls []string
	commitErr := errors.New("commit failed")
	attempt := NewExecutionAttempt(3)
	mustAddResource(t, attempt, recordingResource{name: "proxy", calls: &calls})
	mustAddResource(t, attempt, recordingResource{name: "sample", calls: &calls, commitErr: commitErr})
	mustAddResource(t, attempt, recordingResource{name: "browser", calls: &calls})

	err := attempt.Commit(context.Background())
	if !errors.Is(err, commitErr) {
		t.Fatalf("Commit() error = %v, want commit error", err)
	}

	want := []string{"commit:proxy", "commit:sample", "rollback:proxy"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("calls = %v, want %v", calls, want)
	}
	snapshot := attempt.Snapshot()
	if !snapshot.Finalized || snapshot.Committed || !snapshot.RolledBack {
		t.Fatalf("snapshot = %+v, want rolled back finalized attempt", snapshot)
	}
}

func TestExecutionAttemptRollbackJoinsErrors(t *testing.T) {
	firstErr := errors.New("first")
	secondErr := errors.New("second")
	attempt := NewExecutionAttempt(2)
	mustAddResource(t, attempt, recordingResource{name: "first", rollbackErr: firstErr})
	mustAddResource(t, attempt, recordingResource{name: "second", rollbackErr: secondErr})

	err := attempt.Rollback(context.Background())
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) {
		t.Fatalf("Rollback() error = %v, want both rollback errors", err)
	}
}

func TestExecutionAttemptRejectsInvalidTransitions(t *testing.T) {
	attempt := NewExecutionAttempt(-1)
	if snapshot := attempt.Snapshot(); snapshot.ResourceCount != 0 {
		t.Fatalf("ResourceCount = %d, want 0", snapshot.ResourceCount)
	}
	if err := attempt.AddResource(nil); err == nil {
		t.Fatalf("AddResource(nil) returned nil error, want failure")
	}

	mustAddResource(t, attempt, recordingResource{name: "sample"})
	if err := attempt.Commit(context.Background()); err != nil {
		t.Fatalf("Commit() returned error: %v", err)
	}
	if err := attempt.AddResource(recordingResource{name: "late"}); err == nil {
		t.Fatalf("AddResource(after finalize) returned nil error, want failure")
	}
	if err := attempt.Rollback(context.Background()); err == nil {
		t.Fatalf("Rollback(after finalize) returned nil error, want failure")
	}
}

func mustAddResource(t *testing.T, attempt *ExecutionAttempt, resource AttemptResource) {
	t.Helper()
	if err := attempt.AddResource(resource); err != nil {
		t.Fatalf("AddResource() returned error: %v", err)
	}
}

type recordingResource struct {
	name        string
	calls       *[]string
	commitErr   error
	rollbackErr error
}

func (r recordingResource) ResourceName() string {
	return r.name
}

func (r recordingResource) Commit(context.Context) error {
	r.record("commit")
	return r.commitErr
}

func (r recordingResource) Rollback(context.Context) error {
	r.record("rollback")
	return r.rollbackErr
}

func (r recordingResource) record(action string) {
	if r.calls == nil {
		return
	}
	*r.calls = append(*r.calls, action+":"+r.name)
}
