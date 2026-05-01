package runner

import (
	"context"
	"fmt"

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
