package engine

import (
	"context"
	"fmt"

	"github.com/LING71671/SurveyController-Go/internal/answerplan"
)

type AnswerPlanSubmitter interface {
	Submit(ctx context.Context, plan answerplan.Plan) (SubmissionResult, error)
}

func SubmitAnswerPlan(ctx context.Context, submitter AnswerPlanSubmitter, plan answerplan.Plan) (SubmissionResult, error) {
	if ctx == nil {
		return SubmissionResult{}, fmt.Errorf("context is required")
	}
	if submitter == nil {
		return SubmissionResult{}, fmt.Errorf("answer plan submitter is required")
	}
	if err := ctx.Err(); err != nil {
		return SubmissionResult{}, err
	}
	return submitter.Submit(ctx, plan)
}
