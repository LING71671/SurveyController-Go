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

var benchmarkHTTPPathState provider.SubmissionState

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
	plan := benchmarkAnswerPlan()
	pipeline, err := NewHTTPSubmissionPipeline(
		benchmarkProvider{capabilities: provider.Capabilities{SubmitHTTP: true}},
		engineModeHTTP{},
		benchmarkSurvey(),
		staticHTTPSubmissionExecutor{
			response: HTTPSubmissionResponse{
				StatusCode: http.StatusOK,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Body:       "提交成功",
			},
		},
	)
	if err != nil {
		b.Fatalf("NewHTTPSubmissionPipeline() returned error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var state provider.SubmissionState
		for j := 0; j < taskCount; j++ {
			submissionResult, err := pipeline.Submit(ctx, plan)
			if err != nil {
				b.Fatalf("Submit() returned error: %v", err)
			}
			if submissionResult.State != provider.SubmissionStateSuccess {
				b.Fatalf("state = %q, want success", submissionResult.State)
			}
			state = submissionResult.State
		}
		benchmarkHTTPPathState = state
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

type benchmarkProvider struct {
	capabilities provider.Capabilities
}

func (benchmarkProvider) ID() provider.ProviderID {
	return domain.ProviderWJX
}

func (benchmarkProvider) MatchURL(rawURL string) bool {
	return provider.MatchHostSuffix(rawURL, "wjx.cn")
}

func (p benchmarkProvider) Capabilities() provider.Capabilities {
	return p.capabilities
}

func (benchmarkProvider) Parse(context.Context, string) (provider.SurveyDefinition, error) {
	return provider.SurveyDefinition{}, nil
}

type engineModeHTTP struct{}

func (engineModeHTTP) String() string {
	return "http"
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
