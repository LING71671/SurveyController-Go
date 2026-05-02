package wjx

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/answerplan"
	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/engine"
	"github.com/LING71671/SurveyController-Go/internal/provider"
)

func TestHTTPSubmissionPipelineSubmitSuccess(t *testing.T) {
	executor := &pipelineRecordingExecutor{
		response: HTTPSubmissionResponse{
			StatusCode: http.StatusOK,
			Body:       "提交成功",
		},
	}
	pipeline, err := testHTTPSubmissionPipeline(executor, provider.Capabilities{SubmitHTTP: true})
	if err != nil {
		t.Fatalf("NewHTTPSubmissionPipeline() returned error: %v", err)
	}

	result, err := pipeline.Submit(context.Background(), testPipelineAnswerPlan())
	if err != nil {
		t.Fatalf("Submit() returned error: %v", err)
	}

	if !result.Success || result.State != provider.SubmissionStateSuccess || !result.CompletionDetected {
		t.Fatalf("Submit() = %+v, want successful result", result)
	}
	if len(executor.calls) != 1 || executor.calls[0].Form.Get("q1") != "2" {
		t.Fatalf("executor calls = %+v, want mapped draft", executor.calls)
	}
}

func TestHTTPSubmissionPipelineAcceptsDryRunExecutor(t *testing.T) {
	executor := &DryRunHTTPSubmissionExecutor{}
	pipeline, err := testHTTPSubmissionPipeline(executor, provider.Capabilities{SubmitHTTP: true})
	if err != nil {
		t.Fatalf("NewHTTPSubmissionPipeline() returned error: %v", err)
	}

	result, err := pipeline.Submit(context.Background(), testPipelineAnswerPlan())
	if err != nil {
		t.Fatalf("Submit() returned error: %v", err)
	}

	if !result.Success || result.State != provider.SubmissionStateSuccess {
		t.Fatalf("Submit() = %+v, want successful dry-run result", result)
	}
	if drafts := executor.Drafts(); len(drafts) != 1 || drafts[0].Form.Get("q1") != "2" {
		t.Fatalf("dry-run drafts = %+v, want recorded mapped draft", drafts)
	}
}

func TestHTTPSubmissionPipelineRequiresSubmitCapability(t *testing.T) {
	executor := &pipelineRecordingExecutor{
		response: HTTPSubmissionResponse{StatusCode: http.StatusOK, Body: "提交成功"},
	}
	_, err := testHTTPSubmissionPipeline(executor, provider.Capabilities{ParseHTTP: true})
	if !apperr.IsCode(err, apperr.CodeProviderUnsupported) {
		t.Fatalf("NewHTTPSubmissionPipeline() error = %v, want provider_unsupported", err)
	}
	if len(executor.calls) != 0 {
		t.Fatalf("executor was called despite capability gate failure: %+v", executor.calls)
	}
}

func TestHTTPSubmissionPipelineRejectsMissingExecutor(t *testing.T) {
	_, err := testHTTPSubmissionPipeline(nil, provider.Capabilities{SubmitHTTP: true})
	if err == nil || !strings.Contains(err.Error(), "executor") {
		t.Fatalf("NewHTTPSubmissionPipeline() error = %v, want executor error", err)
	}
}

func TestHTTPSubmissionPipelineReturnsExecutorError(t *testing.T) {
	wantErr := errors.New("offline")
	pipeline, err := testHTTPSubmissionPipeline(&pipelineRecordingExecutor{err: wantErr}, provider.Capabilities{SubmitHTTP: true})
	if err != nil {
		t.Fatalf("NewHTTPSubmissionPipeline() returned error: %v", err)
	}
	_, err = pipeline.Submit(context.Background(), testPipelineAnswerPlan())
	if !errors.Is(err, wantErr) {
		t.Fatalf("Submit() error = %v, want %v", err, wantErr)
	}
}

func TestHTTPSubmissionPipelineMapsStopDetections(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		state   provider.SubmissionState
		errCode apperr.Code
	}{
		{name: "verification", body: "请完成验证码", state: provider.SubmissionStateVerificationRequired, errCode: apperr.CodeVerificationNeeded},
		{name: "rate limited", body: "操作频繁，请稍后再试", state: provider.SubmissionStateRateLimited, errCode: apperr.CodeRateLimited},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline, err := testHTTPSubmissionPipeline(&pipelineRecordingExecutor{
				response: HTTPSubmissionResponse{StatusCode: http.StatusOK, Body: tt.body},
			}, provider.Capabilities{SubmitHTTP: true})
			if err != nil {
				t.Fatalf("NewHTTPSubmissionPipeline() returned error: %v", err)
			}
			result, err := pipeline.Submit(context.Background(), testPipelineAnswerPlan())
			if err != nil {
				t.Fatalf("Submit() returned error: %v", err)
			}
			if result.State != tt.state || !result.ShouldStop || !apperr.IsCode(result.Error, tt.errCode) {
				t.Fatalf("Submit() = %+v, want state %q stop with %q", result, tt.state, tt.errCode)
			}
		})
	}
}

func TestHTTPSubmissionPipelineRejectsInvalidSurvey(t *testing.T) {
	_, err := NewHTTPSubmissionPipeline(
		pipelineProvider{capabilities: provider.Capabilities{SubmitHTTP: true}},
		engine.ModeHTTP,
		domain.SurveyDefinition{
			Provider: domain.ProviderWJX,
			Title:    "invalid",
			URL:      "https://www.wjx.cn/vm/invalid.aspx",
			Questions: []domain.QuestionDefinition{
				{ID: "q1", Title: "One", Kind: domain.QuestionKindSingle},
				{ID: "q1", Title: "Duplicate", Kind: domain.QuestionKindSingle},
			},
		},
		&pipelineRecordingExecutor{},
	)
	if err == nil || !strings.Contains(err.Error(), "defined more than once") {
		t.Fatalf("NewHTTPSubmissionPipeline() error = %v, want schema error", err)
	}
}

func testHTTPSubmissionPipeline(executor HTTPSubmissionExecutor, capabilities provider.Capabilities) (HTTPSubmissionPipeline, error) {
	return NewHTTPSubmissionPipeline(
		pipelineProvider{
			capabilities: capabilities,
		},
		engine.ModeHTTP,
		testPipelineSurvey(),
		executor,
	)
}

type pipelineProvider struct {
	capabilities provider.Capabilities
}

func (pipelineProvider) ID() provider.ProviderID {
	return domain.ProviderWJX
}

func (pipelineProvider) MatchURL(rawURL string) bool {
	return provider.MatchHostSuffix(rawURL, "wjx.cn")
}

func (p pipelineProvider) Capabilities() provider.Capabilities {
	return p.capabilities
}

func (pipelineProvider) Parse(context.Context, string) (provider.SurveyDefinition, error) {
	return provider.SurveyDefinition{}, nil
}

type pipelineRecordingExecutor struct {
	response HTTPSubmissionResponse
	err      error
	calls    []HTTPSubmissionDraft
}

func (e *pipelineRecordingExecutor) ExecuteHTTPSubmission(ctx context.Context, draft HTTPSubmissionDraft) (HTTPSubmissionResponse, error) {
	if err := ctx.Err(); err != nil {
		return HTTPSubmissionResponse{}, err
	}
	e.calls = append(e.calls, cloneHTTPSubmissionDraft(draft))
	if e.err != nil {
		return HTTPSubmissionResponse{}, e.err
	}
	return cloneHTTPSubmissionResponse(e.response), nil
}

func testPipelineSurvey() domain.SurveyDefinition {
	return domain.SurveyDefinition{
		Provider: domain.ProviderWJX,
		Title:    "Pipeline",
		URL:      "https://www.wjx.cn/vm/pipeline.aspx",
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
				Title: "Rating",
				Kind:  domain.QuestionKindRating,
				Options: []domain.OptionDefinition{
					{ID: "score5", Label: "5", Value: "5"},
				},
			},
		},
	}
}

func testPipelineAnswerPlan() answerplan.Plan {
	return answerplan.Plan{
		Answers: []answerplan.QuestionAnswer{
			{QuestionID: "q1", OptionIDs: []string{"b"}},
			{QuestionID: "q2", OptionIDs: []string{"score5"}},
		},
	}
}
