package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCompileRunPlanFromFileDetectsProviderAndAppliesOverrides(t *testing.T) {
	path := writeRunConfig(t, `schema_version: 1
survey:
  url: "https://www.wjx.cn/vm/example.aspx"
run:
  target: 1
  concurrency: 1
  mode: http
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	plan, err := CompileRunPlanFromFile(path, RunPlanOverrides{Target: 3, Concurrency: 2})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}
	if plan.Provider != "wjx" || plan.Target != 3 || plan.Concurrency != 2 {
		t.Fatalf("plan = %+v, want detected provider and overrides", plan)
	}
}

func TestRunMockPlanReportsSuccess(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, validMockRunConfig()), RunPlanOverrides{Target: 2, Concurrency: 1})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}

	report, err := RunMockPlan(context.Background(), plan, MockRunOptions{Seed: 7})
	if err != nil {
		t.Fatalf("RunMockPlan() error = %v", err)
	}
	if report.Successes != 2 || report.Failures != 0 || report.Completed != 2 {
		t.Fatalf("report = %+v, want successful mock run", report)
	}
	if report.Goroutines <= 0 || report.TotalAllocDelta == 0 {
		t.Fatalf("report metrics = %+v, want runtime metrics", report)
	}
}

func TestRunMockPlanSupportsFailureInjection(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, validMockRunConfig()), RunPlanOverrides{Target: 5, Concurrency: 1})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}
	plan.FailureThreshold = 1
	plan.FailStopEnabled = true

	report, err := RunMockPlan(context.Background(), plan, MockRunOptions{Seed: 7, FailEvery: 2})
	if err != nil {
		t.Fatalf("RunMockPlan() error = %v", err)
	}
	if !report.FailureThreshold || report.Successes != 1 || report.Failures != 1 {
		t.Fatalf("report = %+v, want threshold failure after injected error", report)
	}
}

func validMockRunConfig() string {
	return `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: http
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`
}

func writeRunConfig(tb testing.TB, body string) string {
	tb.Helper()
	path := filepath.Join(tb.TempDir(), "survey.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		tb.Fatalf("write config: %v", err)
	}
	return path
}
