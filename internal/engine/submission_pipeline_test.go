package engine

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

func TestSubmitAnswerPlanCallsSubmitter(t *testing.T) {
	submitter := &recordingAnswerPlanSubmitter{
		result: SubmissionResult{
			State:   provider.SubmissionStateSuccess,
			Message: "ok",
			Success: true,
		},
	}
	plan := answerplan.Plan{Answers: []answerplan.QuestionAnswer{{QuestionID: "q1", OptionIDs: []string{"a"}}}}

	result, err := SubmitAnswerPlan(context.Background(), submitter, plan)
	if err != nil {
		t.Fatalf("SubmitAnswerPlan() error = %v", err)
	}
	if !result.Success || result.Message != "ok" {
		t.Fatalf("result = %+v, want success ok", result)
	}
	if submitter.calls != 1 || submitter.lastPlan.Answers[0].QuestionID != "q1" {
		t.Fatalf("submitter = %+v, want one call with q1", submitter)
	}
}

func TestSubmitAnswerPlanRejectsInvalidInputs(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		submitter AnswerPlanSubmitter
		want      string
	}{
		{name: "context", submitter: &recordingAnswerPlanSubmitter{}, want: "context"},
		{name: "submitter", ctx: context.Background(), want: "submitter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SubmitAnswerPlan(tt.ctx, tt.submitter, answerplan.Plan{})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("SubmitAnswerPlan() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSubmitAnswerPlanHonorsCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := SubmitAnswerPlan(ctx, &recordingAnswerPlanSubmitter{}, answerplan.Plan{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SubmitAnswerPlan() error = %v, want context.Canceled", err)
	}
}

type recordingAnswerPlanSubmitter struct {
	result   SubmissionResult
	err      error
	calls    int
	lastPlan answerplan.Plan
}

func (s *recordingAnswerPlanSubmitter) Submit(ctx context.Context, plan answerplan.Plan) (SubmissionResult, error) {
	if err := ctx.Err(); err != nil {
		return SubmissionResult{}, err
	}
	s.calls++
	s.lastPlan = answerplan.Clone(plan)
	return s.result, s.err
}
