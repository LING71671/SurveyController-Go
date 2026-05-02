package runner

import (
	"context"
	"errors"
	"math/rand"
	"strings"
	"sync"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/answer"
	"github.com/LING71671/SurveyController-Go/internal/answerplan"
	"github.com/LING71671/SurveyController-Go/internal/engine"
	"github.com/LING71671/SurveyController-Go/internal/provider"
)

func TestSubmissionTasksFromAnswerPlans(t *testing.T) {
	submitter := &recordingAnswerPlanSubmitter{}
	plans := []answerplan.Plan{
		{Answers: []answerplan.QuestionAnswer{{QuestionID: "q1", Value: "1"}}},
		{Answers: []answerplan.QuestionAnswer{{QuestionID: "q2", Value: "2"}}},
	}

	tasks, err := SubmissionTasksFromAnswerPlans(submitter, plans)
	if err != nil {
		t.Fatalf("SubmissionTasksFromAnswerPlans() returned error: %v", err)
	}
	plans[0].Answers[0].QuestionID = "mutated"

	for _, task := range tasks {
		result, err := task(context.Background(), 7)
		if err != nil {
			t.Fatalf("task returned error: %v", err)
		}
		if !result.Success {
			t.Fatalf("task result = %+v, want success", result)
		}
	}

	submittedPlans := submitter.submittedPlans()
	if got := submittedPlans[0].Answers[0].QuestionID; got != "q1" {
		t.Fatalf("first submitted plan id = %q, want q1", got)
	}
	if got := submittedPlans[1].Answers[0].QuestionID; got != "q2" {
		t.Fatalf("second submitted plan id = %q, want q2", got)
	}
}

func TestWorkerPoolRunSubmissionsWithAnswerPlanTasks(t *testing.T) {
	tasks, err := SubmissionTasksFromAnswerPlans(&recordingAnswerPlanSubmitter{}, []answerplan.Plan{
		{Answers: []answerplan.QuestionAnswer{{QuestionID: "q1", Value: "1"}}},
		{Answers: []answerplan.QuestionAnswer{{QuestionID: "q2", Value: "2"}}},
	})
	if err != nil {
		t.Fatalf("SubmissionTasksFromAnswerPlans() returned error: %v", err)
	}
	pool, err := NewWorkerPool(PoolOptions{Concurrency: 2, Target: 2})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}

	snapshot := pool.RunSubmissions(context.Background(), tasks)
	if snapshot.Successes != 2 || snapshot.Failures != 0 {
		t.Fatalf("snapshot counts = %d/%d, want 2/0", snapshot.Successes, snapshot.Failures)
	}
}

func TestSubmissionTasksFromPlan(t *testing.T) {
	submitter := &recordingAnswerPlanSubmitter{}
	tasks, err := SubmissionTasksFromPlan(rand.New(rand.NewSource(11)), validSubmissionTaskPlan(3), submitter)
	if err != nil {
		t.Fatalf("SubmissionTasksFromPlan() returned error: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("len(tasks) = %d, want 3", len(tasks))
	}

	pool, err := NewWorkerPool(PoolOptions{Concurrency: 2, Target: 3})
	if err != nil {
		t.Fatalf("NewWorkerPool() returned error: %v", err)
	}
	snapshot := pool.RunSubmissions(context.Background(), tasks)
	if snapshot.Successes != 3 || snapshot.Failures != 0 {
		t.Fatalf("snapshot counts = %d/%d, want 3/0", snapshot.Successes, snapshot.Failures)
	}
	submittedPlans := submitter.submittedPlans()
	if len(submittedPlans) != 3 {
		t.Fatalf("submitted plans = %d, want 3", len(submittedPlans))
	}
	for _, submitted := range submittedPlans {
		if len(submitted.Answers) != 1 || submitted.Answers[0].QuestionID != "q1" {
			t.Fatalf("submitted plan = %+v, want q1 answer", submitted)
		}
	}
}

func TestSubmissionTasksFromPlanRejectsInvalidInput(t *testing.T) {
	validPlan := validSubmissionTaskPlan(1)
	tests := []struct {
		name      string
		rng       *rand.Rand
		plan      Plan
		submitter AnswerPlanSubmitter
		want      string
	}{
		{name: "invalid plan", rng: rand.New(rand.NewSource(1)), plan: Plan{}, submitter: &recordingAnswerPlanSubmitter{}, want: "provider"},
		{name: "rng", plan: validPlan, submitter: &recordingAnswerPlanSubmitter{}, want: "rng"},
		{name: "questions", rng: rand.New(rand.NewSource(1)), plan: planWithoutQuestions(validPlan), submitter: &recordingAnswerPlanSubmitter{}, want: "questions"},
		{name: "submitter", rng: rand.New(rand.NewSource(1)), plan: validPlan, want: "submitter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SubmissionTasksFromPlan(tt.rng, tt.plan, tt.submitter)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("SubmissionTasksFromPlan() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSubmissionTasksFromAnswerPlansRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		submitter AnswerPlanSubmitter
		plans     []answerplan.Plan
		want      string
	}{
		{name: "submitter", plans: []answerplan.Plan{{}}, want: "submitter"},
		{name: "plans", submitter: &recordingAnswerPlanSubmitter{}, want: "plans"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SubmissionTasksFromAnswerPlans(tt.submitter, tt.plans)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("SubmissionTasksFromAnswerPlans() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSubmissionTasksFromAnswerPlansReturnsSubmitterError(t *testing.T) {
	wantErr := errors.New("submit failed")
	tasks, err := SubmissionTasksFromAnswerPlans(&recordingAnswerPlanSubmitter{err: wantErr}, []answerplan.Plan{{}})
	if err != nil {
		t.Fatalf("SubmissionTasksFromAnswerPlans() returned error: %v", err)
	}

	_, err = tasks[0](context.Background(), 1)
	if !errors.Is(err, wantErr) {
		t.Fatalf("task error = %v, want %v", err, wantErr)
	}
}

func validSubmissionTaskPlan(target int) Plan {
	return Plan{
		Mode:        engine.ModeHTTP,
		Provider:    "mock",
		URL:         "https://example.com/survey",
		Target:      target,
		Concurrency: 2,
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

func planWithoutQuestions(plan Plan) Plan {
	plan.Questions = nil
	return plan
}

type recordingAnswerPlanSubmitter struct {
	mu    sync.Mutex
	err   error
	plans []answerplan.Plan
}

func (s *recordingAnswerPlanSubmitter) Submit(ctx context.Context, plan answerplan.Plan) (engine.SubmissionResult, error) {
	if err := ctx.Err(); err != nil {
		return engine.SubmissionResult{}, err
	}
	s.mu.Lock()
	s.plans = append(s.plans, answerplan.Clone(plan))
	s.mu.Unlock()
	if s.err != nil {
		return engine.SubmissionResult{}, s.err
	}
	return engine.SubmissionResult{
		State:    provider.SubmissionStateSuccess,
		Message:  "ok",
		Success:  true,
		Terminal: true,
	}, nil
}

func (s *recordingAnswerPlanSubmitter) submittedPlans() []answerplan.Plan {
	s.mu.Lock()
	defer s.mu.Unlock()

	plans := make([]answerplan.Plan, 0, len(s.plans))
	for _, plan := range s.plans {
		plans = append(plans, answerplan.Clone(plan))
	}
	return plans
}
