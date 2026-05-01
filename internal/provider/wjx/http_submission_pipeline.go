package wjx

import (
	"context"
	"fmt"

	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/engine"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

type HTTPSubmissionPipeline struct {
	Provider provider.Provider
	Mode     provider.ModeValue
	Schema   HTTPAnswerSchema
	Executor HTTPSubmissionExecutor
}

func (p HTTPSubmissionPipeline) Submit(ctx context.Context, plan answerplan.Plan) (engine.SubmissionResult, error) {
	if err := provider.RequireSubmitCapability(p.Provider, p.Mode); err != nil {
		return engine.SubmissionResult{}, err
	}
	if p.Executor == nil {
		return engine.SubmissionResult{}, fmt.Errorf("http submission executor is required")
	}

	draft, err := p.Schema.BuildSubmissionDraft(plan)
	if err != nil {
		return engine.SubmissionResult{}, err
	}
	response, err := ExecuteHTTPSubmission(ctx, p.Executor, draft)
	if err != nil {
		return engine.SubmissionResult{}, err
	}
	return engine.ResultFromDetection(DetectHTTPSubmissionResponse(response)), nil
}
