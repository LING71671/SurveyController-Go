package runner

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/engine"
	"github.com/LING71671/SurveyController-go/internal/logging"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

type RunPlanOptions struct {
	RNG       *rand.Rand
	Submitter AnswerPlanSubmitter
	Events    chan<- logging.RunEvent
}

func RunPlanSubmissions(ctx context.Context, plan Plan, options RunPlanOptions) (StateSnapshot, error) {
	if ctx == nil {
		return StateSnapshot{}, fmt.Errorf("context is required")
	}
	if options.RNG == nil {
		return StateSnapshot{}, fmt.Errorf("rng is required")
	}
	if options.Submitter == nil {
		return StateSnapshot{}, fmt.Errorf("answer plan submitter is required")
	}
	if err := New().ValidatePlan(plan); err != nil {
		return StateSnapshot{}, err
	}
	builder, err := CompileAnswerPlanBuilder(plan.Questions)
	if err != nil {
		return StateSnapshot{}, err
	}
	pool, err := NewWorkerPool(PoolOptions{
		Concurrency:      plan.Concurrency,
		Target:           plan.Target,
		FailureThreshold: runFailureThreshold(plan),
		Events:           options.Events,
	})
	if err != nil {
		return StateSnapshot{}, err
	}
	return pool.RunGeneratedSubmissions(ctx, plan.Target, func(index int) (SubmissionTask, error) {
		answerPlan, err := builder.Build(options.RNG)
		if err != nil {
			return nil, fmt.Errorf("answer plan %d: %w", index+1, err)
		}
		taskPlan := answerPlan
		return func(ctx context.Context, workerID int) (engine.SubmissionResult, error) {
			_ = workerID
			return options.Submitter.Submit(ctx, taskPlan)
		}, nil
	})
}

func runFailureThreshold(plan Plan) int {
	if !plan.FailStopEnabled {
		return 0
	}
	return plan.FailureThreshold
}

type MockAnswerPlanSubmitter struct{}

func (MockAnswerPlanSubmitter) Submit(ctx context.Context, plan answerplan.Plan) (engine.SubmissionResult, error) {
	if err := ctx.Err(); err != nil {
		return engine.SubmissionResult{}, err
	}
	if plan.Empty() {
		return engine.SubmissionResult{
			State:    provider.SubmissionStateFailure,
			Message:  "answer plan is empty",
			Terminal: true,
			Error:    fmt.Errorf("answer plan is empty"),
		}, nil
	}
	return engine.SubmissionResult{
		State:              provider.SubmissionStateSuccess,
		Message:            "mock submission succeeded",
		Success:            true,
		Terminal:           true,
		CompletionDetected: true,
	}, nil
}
