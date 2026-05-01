package runner

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/answer"
	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/engine"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

var benchmarkRunnerSnapshot StateSnapshot

func BenchmarkSubmissionTasksFromPlan(b *testing.B) {
	for _, target := range []int{1, 1000, 5000} {
		b.Run(fmt.Sprintf("target_%d", target), func(b *testing.B) {
			benchmarkSubmissionTasksFromPlan(b, target)
		})
	}
}

func benchmarkSubmissionTasksFromPlan(b *testing.B, target int) {
	b.Helper()
	plan := benchmarkRunnerPlan(target)
	submitter := benchmarkAnswerPlanSubmitter{}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tasks, err := SubmissionTasksFromPlan(rand.New(rand.NewSource(int64(i+1))), plan, submitter)
		if err != nil {
			b.Fatalf("SubmissionTasksFromPlan() returned error: %v", err)
		}
		if len(tasks) != target {
			b.Fatalf("len(tasks) = %d, want %d", len(tasks), target)
		}
	}
}

func BenchmarkWorkerPoolRunSubmissionsLightTasks(b *testing.B) {
	for _, target := range []int{1000, 5000} {
		b.Run(fmt.Sprintf("target_%d", target), func(b *testing.B) {
			benchmarkWorkerPoolRunSubmissionsLightTasks(b, target)
		})
	}
}

func benchmarkWorkerPoolRunSubmissionsLightTasks(b *testing.B, target int) {
	b.Helper()
	tasks, err := SubmissionTasksFromPlan(rand.New(rand.NewSource(1)), benchmarkRunnerPlan(target), benchmarkAnswerPlanSubmitter{})
	if err != nil {
		b.Fatalf("SubmissionTasksFromPlan() returned error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool, err := NewWorkerPool(PoolOptions{Concurrency: DefaultMaxWorkerConcurrency, Target: target})
		if err != nil {
			b.Fatalf("NewWorkerPool() returned error: %v", err)
		}
		snapshot := pool.RunSubmissions(context.Background(), tasks)
		if snapshot.Successes != target || snapshot.Failures != 0 {
			b.Fatalf("snapshot counts = %d/%d, want %d/0", snapshot.Successes, snapshot.Failures, target)
		}
		benchmarkRunnerSnapshot = snapshot
	}
}

func benchmarkRunnerPlan(target int) Plan {
	return Plan{
		Mode:        engine.ModeHTTP,
		Provider:    "mock",
		URL:         "https://example.com/survey",
		Target:      target,
		Concurrency: DefaultMaxWorkerConcurrency,
		Questions: []QuestionPlan{
			{
				ID:   "q1",
				Kind: "single",
				Weights: []answer.OptionWeight{
					{OptionID: "a", Weight: 1},
					{OptionID: "b", Weight: 1},
				},
			},
			{
				ID:   "q2",
				Kind: "multiple",
				Options: map[string]any{
					"min": 1,
					"max": 2,
				},
				Weights: []answer.OptionWeight{
					{OptionID: "a", Weight: 1},
					{OptionID: "b", Weight: 1},
					{OptionID: "c", Weight: 1},
				},
			},
			{
				ID:   "q3",
				Kind: "rating",
				Weights: []answer.OptionWeight{
					{OptionID: "score4", Weight: 1},
					{OptionID: "score5", Weight: 1},
				},
			},
		},
	}
}

type benchmarkAnswerPlanSubmitter struct{}

func (benchmarkAnswerPlanSubmitter) Submit(ctx context.Context, plan answerplan.Plan) (engine.SubmissionResult, error) {
	if err := ctx.Err(); err != nil {
		return engine.SubmissionResult{}, err
	}
	_ = plan
	return engine.SubmissionResult{
		State:    provider.SubmissionStateSuccess,
		Message:  "ok",
		Success:  true,
		Terminal: true,
	}, nil
}
