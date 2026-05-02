package runner

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/answer"
	"github.com/LING71671/SurveyController-Go/internal/answerplan"
	"github.com/LING71671/SurveyController-Go/internal/engine"
	"github.com/LING71671/SurveyController-Go/internal/logging"
	"github.com/LING71671/SurveyController-Go/internal/provider"
)

func TestRunPlanSubmissionsWithMockSubmitter(t *testing.T) {
	events := make(chan logging.RunEvent, 64)
	snapshot, err := RunPlanSubmissions(context.Background(), runnableTestPlan(3), RunPlanOptions{
		RNG:       rand.New(rand.NewSource(1)),
		Submitter: MockAnswerPlanSubmitter{},
		Events:    events,
	})
	if err != nil {
		t.Fatalf("RunPlanSubmissions() returned error: %v", err)
	}
	if snapshot.Successes != 3 || snapshot.Failures != 0 {
		t.Fatalf("snapshot counts = %d/%d, want 3/0", snapshot.Successes, snapshot.Failures)
	}
	if !hasEvent(events, logging.EventRunStarted) || !hasEvent(events, logging.EventSubmissionSuccess) || !hasEvent(events, logging.EventRunFinished) {
		t.Fatalf("events did not include run start, success, and finish")
	}
}

func TestRunPlanSubmissionsHonorsFailStopFlag(t *testing.T) {
	plan := runnableTestPlan(3)
	plan.FailStopEnabled = false
	plan.FailureThreshold = 1

	snapshot, err := RunPlanSubmissions(context.Background(), plan, RunPlanOptions{
		RNG:       rand.New(rand.NewSource(1)),
		Submitter: failingAnswerPlanSubmitter{},
	})
	if err != nil {
		t.Fatalf("RunPlanSubmissions() returned error: %v", err)
	}
	if snapshot.Failures != 3 || snapshot.StopRequested {
		t.Fatalf("snapshot = %+v, want all failures without stop request", snapshot)
	}
}

func TestFailureInjectingMockSubmitter(t *testing.T) {
	submitter := &FailureInjectingMockSubmitter{FailEvery: 2}
	plan := answerplan.Plan{Answers: []answerplan.QuestionAnswer{{QuestionID: "q1", OptionIDs: []string{"a"}}}}

	first, err := submitter.Submit(context.Background(), plan)
	if err != nil {
		t.Fatalf("first Submit() returned error: %v", err)
	}
	if !first.Success || first.State != provider.SubmissionStateSuccess {
		t.Fatalf("first result = %+v, want success", first)
	}

	second, err := submitter.Submit(context.Background(), plan)
	if err != nil {
		t.Fatalf("second Submit() returned error: %v", err)
	}
	if second.Success || second.State != provider.SubmissionStateFailure || second.Error == nil {
		t.Fatalf("second result = %+v, want injected failure", second)
	}
}

func TestRunPlanSubmissionsWithFailureInjectingMockSubmitter(t *testing.T) {
	plan := runnableTestPlan(5)
	plan.Concurrency = 1
	plan.FailureThreshold = 1
	plan.FailStopEnabled = true

	snapshot, err := RunPlanSubmissions(context.Background(), plan, RunPlanOptions{
		RNG:       rand.New(rand.NewSource(1)),
		Submitter: &FailureInjectingMockSubmitter{FailEvery: 2},
	})
	if err != nil {
		t.Fatalf("RunPlanSubmissions() returned error: %v", err)
	}
	if snapshot.Successes != 1 || snapshot.Failures != 1 {
		t.Fatalf("snapshot counts = %d/%d, want 1/1", snapshot.Successes, snapshot.Failures)
	}
	if !snapshot.FailureThresholdReached() {
		t.Fatalf("FailureThresholdReached() = false, want true")
	}
}

func TestRunPlanSubmissionsRejectsInvalidInput(t *testing.T) {
	validPlan := runnableTestPlan(1)
	tests := []struct {
		name    string
		ctx     context.Context
		plan    Plan
		options RunPlanOptions
		want    string
	}{
		{name: "context", plan: validPlan, options: RunPlanOptions{RNG: rand.New(rand.NewSource(1)), Submitter: MockAnswerPlanSubmitter{}}, want: "context"},
		{name: "rng", ctx: context.Background(), plan: validPlan, options: RunPlanOptions{Submitter: MockAnswerPlanSubmitter{}}, want: "rng"},
		{name: "submitter", ctx: context.Background(), plan: validPlan, options: RunPlanOptions{RNG: rand.New(rand.NewSource(1))}, want: "submitter"},
		{name: "plan", ctx: context.Background(), plan: Plan{}, options: RunPlanOptions{RNG: rand.New(rand.NewSource(1)), Submitter: MockAnswerPlanSubmitter{}}, want: "provider"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RunPlanSubmissions(tt.ctx, tt.plan, tt.options)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("RunPlanSubmissions() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func runnableTestPlan(target int) Plan {
	return Plan{
		Mode:             engine.ModeHTTP,
		Provider:         "mock",
		URL:              "https://example.com/survey",
		Target:           target,
		Concurrency:      2,
		FailureThreshold: 1,
		FailStopEnabled:  true,
		Questions: []QuestionPlan{
			{
				ID:   "q1",
				Kind: "single",
				Weights: []answer.OptionWeight{
					{OptionID: "a", Weight: 1},
					{OptionID: "b", Weight: 1},
				},
			},
		},
	}
}

type failingAnswerPlanSubmitter struct{}

func (failingAnswerPlanSubmitter) Submit(ctx context.Context, plan answerplan.Plan) (engine.SubmissionResult, error) {
	if err := ctx.Err(); err != nil {
		return engine.SubmissionResult{}, err
	}
	_ = plan
	return engine.SubmissionResult{
		State:    provider.SubmissionStateFailure,
		Message:  "mock failure",
		Terminal: true,
		Error:    fmt.Errorf("mock failure"),
	}, nil
}
