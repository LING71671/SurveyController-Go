package app

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/domain"
	"github.com/LING71671/SurveyController-go/internal/provider/wjx"
	"github.com/LING71671/SurveyController-go/internal/runner"
)

func TestRunWJXHTTPPlanSubmitsThroughPipeline(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, validWJXRunConfig()), RunPlanOverrides{Target: 2, Concurrency: 1})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}
	executor := &recordingWJXHTTPExecutor{
		response: wjx.HTTPSubmissionResponse{StatusCode: http.StatusOK, Body: "提交成功"},
	}

	report, err := RunWJXHTTPPlan(context.Background(), plan, WJXHTTPRunOptions{
		Seed:     7,
		Executor: executor,
		Survey:   testWJXSurvey(),
	})
	if err != nil {
		t.Fatalf("RunWJXHTTPPlan() error = %v", err)
	}
	if report.Successes != 2 || report.Failures != 0 || report.Completed != 2 {
		t.Fatalf("report = %+v, want two successful submissions", report)
	}
	if len(executor.calls) != 2 {
		t.Fatalf("executor calls = %d, want 2", len(executor.calls))
	}
	if got := executor.calls[0].Form.Get("q1"); got != "1" {
		t.Fatalf("first form q1 = %q, want mapped option value", got)
	}
}

func TestRunWJXHTTPPlanStopsOnVerification(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, validWJXRunConfig()), RunPlanOverrides{Target: 5, Concurrency: 1})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}
	plan.FailureThreshold = 1
	plan.FailStopEnabled = true
	executor := &recordingWJXHTTPExecutor{
		response: wjx.HTTPSubmissionResponse{StatusCode: http.StatusOK, Body: "请完成验证码"},
	}

	report, err := RunWJXHTTPPlan(context.Background(), plan, WJXHTTPRunOptions{
		Seed:     7,
		Executor: executor,
		Survey:   testWJXSurvey(),
	})
	if err != nil {
		t.Fatalf("RunWJXHTTPPlan() error = %v", err)
	}
	if !report.StopRequested || !report.FailureThreshold || report.Failures != 1 {
		t.Fatalf("report = %+v, want verification stop", report)
	}
}

func TestRunWJXHTTPPlanRejectsInvalidInputs(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, validWJXRunConfig()), RunPlanOverrides{})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}

	tests := []struct {
		name    string
		plan    func() runner.Plan
		options WJXHTTPRunOptions
		want    string
	}{
		{
			name:    "executor",
			options: WJXHTTPRunOptions{Survey: testWJXSurvey()},
			want:    "executor",
		},
		{
			name: "provider",
			plan: func() runner.Plan {
				other := plan
				other.Provider = "mock"
				return other
			},
			options: WJXHTTPRunOptions{Survey: testWJXSurvey(), Executor: &recordingWJXHTTPExecutor{}},
			want:    "wjx provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPlan := plan
			if tt.plan != nil {
				testPlan = tt.plan()
			}
			_, err := RunWJXHTTPPlan(context.Background(), testPlan, tt.options)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("RunWJXHTTPPlan() error = %v, want %q", err, tt.want)
			}
		})
	}
}

type recordingWJXHTTPExecutor struct {
	response wjx.HTTPSubmissionResponse
	calls    []wjx.HTTPSubmissionDraft
}

func (e *recordingWJXHTTPExecutor) ExecuteHTTPSubmission(ctx context.Context, draft wjx.HTTPSubmissionDraft) (wjx.HTTPSubmissionResponse, error) {
	if err := ctx.Err(); err != nil {
		return wjx.HTTPSubmissionResponse{}, err
	}
	e.calls = append(e.calls, draft)
	return e.response, nil
}

func validWJXRunConfig() string {
	return `schema_version: 1
survey:
  url: "https://www.wjx.cn/vm/example.aspx"
  provider: "wjx"
run:
  target: 1
  concurrency: 1
  mode: http
  failure_threshold: 1
  fail_stop_enabled: true
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`
}

func testWJXSurvey() domain.SurveyDefinition {
	return domain.SurveyDefinition{
		Provider: domain.ProviderWJX,
		Title:    "WJX HTTP",
		URL:      "https://www.wjx.cn/vm/example.aspx",
		Questions: []domain.QuestionDefinition{
			{
				ID:    "q1",
				Title: "Single",
				Kind:  domain.QuestionKindSingle,
				Options: []domain.OptionDefinition{
					{ID: "a", Label: "A", Value: "1"},
				},
			},
		},
	}
}
