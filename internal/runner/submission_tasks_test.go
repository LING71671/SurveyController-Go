package runner

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/engine"
	"github.com/LING71671/SurveyController-go/internal/provider"
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

	if got := submitter.plans[0].Answers[0].QuestionID; got != "q1" {
		t.Fatalf("first submitted plan id = %q, want q1", got)
	}
	if got := submitter.plans[1].Answers[0].QuestionID; got != "q2" {
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

type recordingAnswerPlanSubmitter struct {
	err   error
	plans []answerplan.Plan
}

func (s *recordingAnswerPlanSubmitter) Submit(ctx context.Context, plan answerplan.Plan) (engine.SubmissionResult, error) {
	if err := ctx.Err(); err != nil {
		return engine.SubmissionResult{}, err
	}
	s.plans = append(s.plans, answerplan.Clone(plan))
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
