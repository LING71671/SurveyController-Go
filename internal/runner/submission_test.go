package runner

import (
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/engine"
	"github.com/LING71671/SurveyController-Go/internal/logging"
	"github.com/LING71671/SurveyController-Go/internal/provider"
)

func TestRecordSubmissionResultCountsSuccess(t *testing.T) {
	state := NewRunState(StateOptions{Target: 1})

	state.RecordSubmissionResult(1, engine.SubmissionResult{
		State:    provider.SubmissionStateSuccess,
		Message:  "done",
		Success:  true,
		Terminal: true,
	})

	snapshot := state.Snapshot()
	if snapshot.Successes != 1 || snapshot.Failures != 0 {
		t.Fatalf("counts = %d/%d, want 1/0", snapshot.Successes, snapshot.Failures)
	}
	if snapshot.Workers[1].Successes != 1 || snapshot.Workers[1].Message != "" {
		t.Fatalf("worker progress = %+v, want success recorded", snapshot.Workers[1])
	}
}

func TestRecordSubmissionResultCountsFailure(t *testing.T) {
	state := NewRunState(StateOptions{FailureThreshold: 2})

	state.RecordSubmissionResult(2, engine.SubmissionResult{
		State:    provider.SubmissionStateFailure,
		Message:  "submit failed",
		Terminal: true,
		Error:    apperr.New(apperr.CodeSubmitFailed, "submit failed"),
	})

	snapshot := state.Snapshot()
	if snapshot.Successes != 0 || snapshot.Failures != 1 {
		t.Fatalf("counts = %d/%d, want 0/1", snapshot.Successes, snapshot.Failures)
	}
	if snapshot.Workers[2].Failures != 1 || snapshot.Workers[2].Message != "submit failed" {
		t.Fatalf("worker progress = %+v, want failure message", snapshot.Workers[2])
	}
	if snapshot.LastFailureCode != apperr.CodeSubmitFailed || snapshot.Workers[2].ErrorCode != apperr.CodeSubmitFailed {
		t.Fatalf("failure codes = %q/%q, want %q", snapshot.LastFailureCode, snapshot.Workers[2].ErrorCode, apperr.CodeSubmitFailed)
	}
	if snapshot.LastFailureReason != string(apperr.CodeSubmitFailed) || snapshot.Workers[2].FailureReason != string(apperr.CodeSubmitFailed) {
		t.Fatalf("failure reasons = %q/%q, want %q", snapshot.LastFailureReason, snapshot.Workers[2].FailureReason, apperr.CodeSubmitFailed)
	}
	if snapshot.StopRequested {
		t.Fatalf("StopRequested = true, want false")
	}
}

func TestRecordSubmissionResultRequestsStop(t *testing.T) {
	state := NewRunState(StateOptions{Target: 10, FailureThreshold: 10})

	state.RecordSubmissionResult(3, engine.SubmissionResult{
		State:      provider.SubmissionStateVerificationRequired,
		Message:    "verification required",
		Terminal:   true,
		ShouldStop: true,
		Error:      apperr.New(apperr.CodeVerificationNeeded, "verification required"),
	})

	snapshot := state.Snapshot()
	if snapshot.Failures != 1 || !snapshot.StopRequested || snapshot.StopReason != "verification required" {
		t.Fatalf("snapshot = %+v, want failure and explicit stop", snapshot)
	}
	if snapshot.LastFailureCode != apperr.CodeVerificationNeeded || snapshot.StopCode != apperr.CodeVerificationNeeded {
		t.Fatalf("snapshot codes = %q/%q, want %q", snapshot.LastFailureCode, snapshot.StopCode, apperr.CodeVerificationNeeded)
	}
	if snapshot.LastFailureReason != string(apperr.CodeVerificationNeeded) || snapshot.StopFailureReason != string(apperr.CodeVerificationNeeded) {
		t.Fatalf("snapshot failure reasons = %q/%q, want %q", snapshot.LastFailureReason, snapshot.StopFailureReason, apperr.CodeVerificationNeeded)
	}
	if !state.ShouldStop() {
		t.Fatalf("ShouldStop() = false, want true")
	}
}

func TestRecordSubmissionResultRecordsProgressForUnknown(t *testing.T) {
	state := NewRunState(StateOptions{})

	state.RecordSubmissionResult(4, engine.SubmissionResult{
		State:   provider.SubmissionStateUnknown,
		Message: "waiting for completion",
	})

	snapshot := state.Snapshot()
	if snapshot.Successes != 0 || snapshot.Failures != 0 {
		t.Fatalf("counts = %d/%d, want 0/0", snapshot.Successes, snapshot.Failures)
	}
	if snapshot.Workers[4].Message != "waiting for completion" {
		t.Fatalf("worker progress = %+v, want progress message", snapshot.Workers[4])
	}
}

func TestEventForSubmissionResult(t *testing.T) {
	tests := []struct {
		name      string
		result    engine.SubmissionResult
		wantType  logging.EventType
		wantLevel logging.Level
	}{
		{
			name: "success",
			result: engine.SubmissionResult{
				State:              provider.SubmissionStateSuccess,
				Message:            "done",
				Success:            true,
				Terminal:           true,
				CompletionDetected: true,
			},
			wantType:  logging.EventSubmissionSuccess,
			wantLevel: logging.LevelInfo,
		},
		{
			name: "verification",
			result: engine.SubmissionResult{
				State:      provider.SubmissionStateVerificationRequired,
				Message:    "captcha",
				Terminal:   true,
				ShouldStop: true,
				Error:      apperr.New(apperr.CodeVerificationNeeded, "captcha"),
			},
			wantType:  logging.EventVerificationNeeded,
			wantLevel: logging.LevelWarn,
		},
		{
			name: "failure",
			result: engine.SubmissionResult{
				State:             provider.SubmissionStateFailure,
				Message:           "failed",
				Terminal:          true,
				ShouldRotateProxy: true,
				Error:             apperr.New(apperr.CodeSubmitFailed, "failed"),
			},
			wantType:  logging.EventSubmissionFailure,
			wantLevel: logging.LevelError,
		},
		{
			name: "progress",
			result: engine.SubmissionResult{
				State:   provider.SubmissionStateUnknown,
				Message: "pending",
			},
			wantType:  logging.EventWorkerProgress,
			wantLevel: logging.LevelInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := EventForSubmissionResult(7, tt.result)
			if event.Type != tt.wantType || event.Level != tt.wantLevel || event.WorkerID != 7 {
				t.Fatalf("event = %+v, want type %q level %q worker 7", event, tt.wantType, tt.wantLevel)
			}
			if event.Time.IsZero() {
				t.Fatalf("event time is zero: %+v", event)
			}
			if event.Fields["state"] != string(tt.result.State) {
				t.Fatalf("state field = %v, want %q", event.Fields["state"], tt.result.State)
			}
			if event.Fields["should_rotate_proxy"] != tt.result.ShouldRotateProxy {
				t.Fatalf("should_rotate_proxy field = %v, want %v", event.Fields["should_rotate_proxy"], tt.result.ShouldRotateProxy)
			}
			if code, ok := apperr.CodeOf(tt.result.Error); ok {
				if event.Fields["error_code"] != string(code) || event.Fields["failure_reason"] != string(code) {
					t.Fatalf("error fields = %+v, want %q", event.Fields, code)
				}
			}
		})
	}
}
