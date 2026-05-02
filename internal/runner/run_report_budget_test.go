package runner

import (
	"errors"
	"strings"
	"testing"
)

func TestRunReportBudgetCheckPasses(t *testing.T) {
	budget := RunReportBudget{
		MinThroughput:          100,
		MaxHeapAllocDelta:      2048,
		MaxGoroutines:          2,
		ExpectFailureThreshold: BoolBudget(false),
	}
	report := RunPlanReport{
		ThroughputPerSec: 150,
		HeapAllocDelta:   1024,
		Goroutines:       1,
		FailureThreshold: false,
	}

	if err := budget.Check(report); err != nil {
		t.Fatalf("Check() returned error: %v", err)
	}
}

func TestRunReportBudgetCheckReportsViolations(t *testing.T) {
	budget := RunReportBudget{
		MinThroughput:          100,
		MaxHeapAllocDelta:      500,
		MaxGoroutines:          1,
		ExpectFailureThreshold: BoolBudget(true),
	}
	report := RunPlanReport{
		ThroughputPerSec: 25,
		HeapAllocDelta:   700,
		Goroutines:       3,
		FailureThreshold: false,
	}

	err := budget.Check(report)
	var budgetErr RunReportBudgetError
	if !errors.As(err, &budgetErr) {
		t.Fatalf("Check() error = %T %[1]v, want RunReportBudgetError", err)
	}
	if len(budgetErr.Violations) != 4 {
		t.Fatalf("violations = %+v, want 4 entries", budgetErr.Violations)
	}
	for _, want := range []string{"throughput_per_second", "heap_alloc_delta_bytes", "goroutines", "failure_threshold_reached"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error = %q, want %q", err.Error(), want)
		}
	}
}

func TestRunReportBudgetRejectsInvalidBudget(t *testing.T) {
	tests := []struct {
		name   string
		budget RunReportBudget
		want   string
	}{
		{name: "throughput", budget: RunReportBudget{MinThroughput: -1}, want: "throughput"},
		{name: "goroutines", budget: RunReportBudget{MaxGoroutines: -1}, want: "goroutines"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.budget.Check(RunPlanReport{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Check() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestRunReportBudgetTreatsNegativeHeapDeltaAsViolation(t *testing.T) {
	err := RunReportBudget{MaxHeapAllocDelta: 1}.Check(RunPlanReport{HeapAllocDelta: -1})
	var budgetErr RunReportBudgetError
	if !errors.As(err, &budgetErr) {
		t.Fatalf("Check() error = %T %[1]v, want RunReportBudgetError", err)
	}
	if budgetErr.Violations[0].Field != "heap_alloc_delta_bytes" {
		t.Fatalf("violations = %+v, want heap delta", budgetErr.Violations)
	}
}

func TestBoolBudget(t *testing.T) {
	value := BoolBudget(true)
	if value == nil || !*value {
		t.Fatalf("BoolBudget(true) = %v, want pointer to true", value)
	}
}
