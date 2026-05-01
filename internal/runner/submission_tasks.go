package runner

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/engine"
)

type AnswerPlanSubmitter interface {
	Submit(ctx context.Context, plan answerplan.Plan) (engine.SubmissionResult, error)
}

func SubmissionTasksFromAnswerPlans(submitter AnswerPlanSubmitter, plans []answerplan.Plan) ([]SubmissionTask, error) {
	if submitter == nil {
		return nil, fmt.Errorf("answer plan submitter is required")
	}
	if len(plans) == 0 {
		return nil, fmt.Errorf("answer plans are required")
	}

	tasks := make([]SubmissionTask, 0, len(plans))
	for _, plan := range plans {
		taskPlan := answerplan.Clone(plan)
		tasks = append(tasks, func(ctx context.Context, workerID int) (engine.SubmissionResult, error) {
			_ = workerID
			return submitter.Submit(ctx, taskPlan)
		})
	}
	return tasks, nil
}

func SubmissionTasksFromPlan(rng *rand.Rand, plan Plan, submitter AnswerPlanSubmitter) ([]SubmissionTask, error) {
	if err := New().ValidatePlan(plan); err != nil {
		return nil, err
	}
	if submitter == nil {
		return nil, fmt.Errorf("answer plan submitter is required")
	}
	builder, err := CompileAnswerPlanBuilder(plan.Questions)
	if err != nil {
		return nil, err
	}

	tasks := make([]SubmissionTask, 0, plan.Target)
	for i := 0; i < plan.Target; i++ {
		answerPlan, err := builder.Build(rng)
		if err != nil {
			return nil, fmt.Errorf("answer plan %d: %w", i+1, err)
		}
		taskPlan := answerPlan
		tasks = append(tasks, func(ctx context.Context, workerID int) (engine.SubmissionResult, error) {
			_ = workerID
			return submitter.Submit(ctx, taskPlan)
		})
	}
	return tasks, nil
}
