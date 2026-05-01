package wjx

import (
	"context"
	"fmt"

	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/engine"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

type HTTPSubmissionPipeline struct {
	schema   HTTPAnswerSchema
	executor HTTPSubmissionExecutor
}

func NewHTTPSubmissionPipeline(p provider.Provider, mode provider.ModeValue, survey provider.SurveyDefinition, executor HTTPSubmissionExecutor) (HTTPSubmissionPipeline, error) {
	if err := provider.RequireSubmitCapability(p, mode); err != nil {
		return HTTPSubmissionPipeline{}, err
	}
	if executor == nil {
		return HTTPSubmissionPipeline{}, fmt.Errorf("http submission executor is required")
	}
	schema, err := CompileHTTPAnswerSchema(survey)
	if err != nil {
		return HTTPSubmissionPipeline{}, err
	}
	return HTTPSubmissionPipeline{
		schema:   schema,
		executor: executor,
	}, nil
}

func (p HTTPSubmissionPipeline) Submit(ctx context.Context, plan answerplan.Plan) (engine.SubmissionResult, error) {
	if p.executor == nil {
		return engine.SubmissionResult{}, fmt.Errorf("http submission executor is required")
	}

	draft, err := p.schema.BuildSubmissionDraft(plan)
	if err != nil {
		return engine.SubmissionResult{}, err
	}
	response, err := ExecuteHTTPSubmission(ctx, p.executor, draft)
	if err != nil {
		return engine.SubmissionResult{}, err
	}
	return engine.ResultFromDetection(DetectHTTPSubmissionResponse(response)), nil
}
