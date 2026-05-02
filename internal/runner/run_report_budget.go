package runner

import (
	"fmt"
	"strings"
)

type RunReportBudget struct {
	MinThroughput          float64
	MaxHeapAllocDelta      uint64
	MaxGoroutines          int
	ExpectFailureThreshold *bool
}

type RunReportBudgetViolation struct {
	Field string
	Got   string
	Want  string
}

type RunReportBudgetError struct {
	Violations []RunReportBudgetViolation
}

func (e RunReportBudgetError) Error() string {
	if len(e.Violations) == 0 {
		return "run report budget failed"
	}
	parts := make([]string, 0, len(e.Violations))
	for _, violation := range e.Violations {
		parts = append(parts, fmt.Sprintf("%s got %s, want %s", violation.Field, violation.Got, violation.Want))
	}
	return "run report budget failed: " + strings.Join(parts, "; ")
}

func (b RunReportBudget) Validate() error {
	if b.MinThroughput < 0 {
		return fmt.Errorf("minimum throughput must not be negative")
	}
	if b.MaxGoroutines < 0 {
		return fmt.Errorf("maximum goroutines must not be negative")
	}
	return nil
}

func (b RunReportBudget) Check(report RunPlanReport) error {
	if err := b.Validate(); err != nil {
		return err
	}

	violations := make([]RunReportBudgetViolation, 0, 4)
	if b.MinThroughput > 0 && report.ThroughputPerSec < b.MinThroughput {
		violations = append(violations, budgetViolation(
			"throughput_per_second",
			formatFloat(report.ThroughputPerSec),
			">= "+formatFloat(b.MinThroughput),
		))
	}
	if b.MaxHeapAllocDelta > 0 {
		if report.HeapAllocDelta < 0 {
			violations = append(violations, budgetViolation(
				"heap_alloc_delta_bytes",
				fmt.Sprintf("%d", report.HeapAllocDelta),
				">= 0",
			))
		} else if uint64(report.HeapAllocDelta) > b.MaxHeapAllocDelta {
			violations = append(violations, budgetViolation(
				"heap_alloc_delta_bytes",
				fmt.Sprintf("%d", report.HeapAllocDelta),
				"<= "+fmt.Sprintf("%d", b.MaxHeapAllocDelta),
			))
		}
	}
	if b.MaxGoroutines > 0 && report.Goroutines > b.MaxGoroutines {
		violations = append(violations, budgetViolation(
			"goroutines",
			fmt.Sprintf("%d", report.Goroutines),
			"<= "+fmt.Sprintf("%d", b.MaxGoroutines),
		))
	}
	if b.ExpectFailureThreshold != nil && report.FailureThreshold != *b.ExpectFailureThreshold {
		violations = append(violations, budgetViolation(
			"failure_threshold_reached",
			fmt.Sprintf("%t", report.FailureThreshold),
			fmt.Sprintf("%t", *b.ExpectFailureThreshold),
		))
	}

	if len(violations) > 0 {
		return RunReportBudgetError{Violations: violations}
	}
	return nil
}

func budgetViolation(field string, got string, want string) RunReportBudgetViolation {
	return RunReportBudgetViolation{Field: field, Got: got, Want: want}
}

func BoolBudget(value bool) *bool {
	return &value
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%.4f", value)
}
