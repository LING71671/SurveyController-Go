package wjx

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/domain"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

var benchmarkHTTPPathDetection provider.SubmissionDetection

func BenchmarkHTTPPathLightTasks(b *testing.B) {
	for _, taskCount := range []int{1, 1000, 5000} {
		b.Run(fmt.Sprintf("tasks_%d", taskCount), func(b *testing.B) {
			benchmarkHTTPPathLightTasks(b, taskCount)
		})
	}
}

func benchmarkHTTPPathLightTasks(b *testing.B, taskCount int) {
	b.Helper()
	ctx := context.Background()
	schema, err := CompileHTTPAnswerSchema(benchmarkSurvey())
	if err != nil {
		b.Fatalf("CompileHTTPAnswerSchema() returned error: %v", err)
	}
	plan := benchmarkAnswerPlan()
	executor := staticHTTPSubmissionExecutor{
		response: HTTPSubmissionResponse{
			StatusCode: http.StatusOK,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       "提交成功",
		},
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var detection provider.SubmissionDetection
		for j := 0; j < taskCount; j++ {
			draft, err := schema.BuildSubmissionDraft(plan)
			if err != nil {
				b.Fatalf("BuildSubmissionDraft() returned error: %v", err)
			}
			response, err := ExecuteHTTPSubmission(ctx, executor, draft)
			if err != nil {
				b.Fatalf("ExecuteHTTPSubmission() returned error: %v", err)
			}
			detection = DetectHTTPSubmissionResponse(response)
			if detection.State != provider.SubmissionStateSuccess {
				b.Fatalf("state = %q, want success", detection.State)
			}
		}
		benchmarkHTTPPathDetection = detection
	}
}

type staticHTTPSubmissionExecutor struct {
	response HTTPSubmissionResponse
}

func (e staticHTTPSubmissionExecutor) ExecuteHTTPSubmission(ctx context.Context, draft HTTPSubmissionDraft) (HTTPSubmissionResponse, error) {
	if err := ctx.Err(); err != nil {
		return HTTPSubmissionResponse{}, err
	}
	_ = draft
	return e.response, nil
}

func benchmarkSurvey() domain.SurveyDefinition {
	return domain.SurveyDefinition{
		Provider: domain.ProviderWJX,
		Title:    "Benchmark",
		URL:      "https://www.wjx.cn/vm/benchmark.aspx",
		Questions: []domain.QuestionDefinition{
			{
				ID:    "q1",
				Title: "Single",
				Kind:  domain.QuestionKindSingle,
				Options: []domain.OptionDefinition{
					{ID: "a", Label: "A", Value: "1"},
					{ID: "b", Label: "B", Value: "2"},
				},
			},
			{
				ID:    "q2",
				Title: "Multiple",
				Kind:  domain.QuestionKindMultiple,
				Options: []domain.OptionDefinition{
					{ID: "a", Label: "A", Value: "A"},
					{ID: "b", Label: "B", Value: "B"},
					{ID: "c", Label: "C", Value: "C"},
				},
			},
			{
				ID:    "q3",
				Title: "Rating",
				Kind:  domain.QuestionKindRating,
				Options: []domain.OptionDefinition{
					{ID: "score4", Label: "4", Value: "4"},
					{ID: "score5", Label: "5", Value: "5"},
				},
			},
			{
				ID:    "q4",
				Title: "Dropdown",
				Kind:  domain.QuestionKindDropdown,
				Options: []domain.OptionDefinition{
					{ID: "city1", Label: "Beijing", Value: "beijing"},
					{ID: "city2", Label: "Shanghai", Value: "shanghai"},
				},
			},
		},
	}
}

func benchmarkAnswerPlan() answerplan.Plan {
	return answerplan.Plan{
		Answers: []answerplan.QuestionAnswer{
			{QuestionID: "q1", OptionIDs: []string{"b"}},
			{QuestionID: "q2", OptionIDs: []string{"a", "c"}},
			{QuestionID: "q3", OptionIDs: []string{"score5"}},
			{QuestionID: "q4", OptionIDs: []string{"city2"}},
		},
	}
}
