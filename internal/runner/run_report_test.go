package runner

import (
	"testing"

	"github.com/LING71671/SurveyController-go/internal/engine"
)

func TestNewRunPlanReportSummarizesSnapshot(t *testing.T) {
	plan := Plan{
		Provider:    "mock",
		URL:         "https://example.com/survey",
		Mode:        engine.ModeHTTP,
		Target:      4,
		Concurrency: 3,
	}
	snapshot := StateSnapshot{
		Successes:         3,
		Failures:          1,
		StopRequested:     true,
		StopReason:        "failure_threshold",
		StopFailureReason: "validation required",
		Workers: map[int]WorkerProgress{
			1: {ID: 1, Status: WorkerStatusStopped},
			2: {ID: 2, Status: WorkerStatusStopped},
			3: {ID: 3, Status: WorkerStatusStopped},
		},
	}

	report := NewRunPlanReport(plan, snapshot)

	if report.Provider != "mock" || report.URL != "https://example.com/survey" || report.Mode != "http" {
		t.Fatalf("report identity = %+v, want provider/url/mode copied", report)
	}
	if report.Target != 4 || report.Concurrency != 3 || report.WorkerCount != 3 {
		t.Fatalf("report sizing = %+v, want target/concurrency/workers", report)
	}
	if report.Successes != 3 || report.Failures != 1 || report.Completed != 4 {
		t.Fatalf("report totals = %+v, want successes/failures/completed", report)
	}
	if report.CompletionRate != 1 || report.SuccessRate != 0.75 {
		t.Fatalf("report rates = completion %.4f success %.4f, want 1 and 0.75", report.CompletionRate, report.SuccessRate)
	}
	if !report.StopRequested || report.StopReason != "failure_threshold" || report.StopFailureReason != "validation required" {
		t.Fatalf("report stop = %+v, want stop details copied", report)
	}
	if !report.HasFailures() {
		t.Fatalf("HasFailures() = false, want true")
	}
	if report.TargetReached() {
		t.Fatalf("TargetReached() = true, want false when successes < target")
	}
}

func TestNewRunPlanReportHandlesPartialRun(t *testing.T) {
	report := NewRunPlanReport(Plan{Target: 10}, StateSnapshot{Successes: 2, Failures: 1})

	if report.Completed != 3 {
		t.Fatalf("Completed = %d, want 3", report.Completed)
	}
	if report.CompletionRate != 0.3 {
		t.Fatalf("CompletionRate = %.4f, want 0.3", report.CompletionRate)
	}
	if report.SuccessRate != 0.6667 {
		t.Fatalf("SuccessRate = %.4f, want rounded 0.6667", report.SuccessRate)
	}
}

func TestRunPlanReportTargetReached(t *testing.T) {
	report := NewRunPlanReport(Plan{Target: 2}, StateSnapshot{Successes: 2})

	if !report.TargetReached() {
		t.Fatalf("TargetReached() = false, want true")
	}
	if report.HasFailures() {
		t.Fatalf("HasFailures() = true, want false")
	}
}

func TestRunPlanReportAvoidsDivideByZero(t *testing.T) {
	report := NewRunPlanReport(Plan{}, StateSnapshot{})

	if report.CompletionRate != 0 || report.SuccessRate != 0 {
		t.Fatalf("rates = %.4f/%.4f, want zeros", report.CompletionRate, report.SuccessRate)
	}
}
