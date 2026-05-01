package wjx

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

func TestExecuteHTTPSubmissionUsesExecutor(t *testing.T) {
	draft := mustDraft(t)
	executor := &recordingHTTPSubmissionExecutor{
		response: HTTPSubmissionResponse{
			StatusCode: http.StatusOK,
			Header:     http.Header{"X-Test": []string{"ok"}},
			Body:       "success",
		},
	}

	response, err := ExecuteHTTPSubmission(context.Background(), executor, draft)
	if err != nil {
		t.Fatalf("ExecuteHTTPSubmission() returned error: %v", err)
	}
	if response.StatusCode != http.StatusOK || response.Body != "success" || response.Header.Get("X-Test") != "ok" {
		t.Fatalf("response = %+v, want configured mock response", response)
	}
	if len(executor.calls) != 1 || executor.calls[0].SurveyID != draft.SurveyID {
		t.Fatalf("calls = %+v, want one cloned draft call", executor.calls)
	}
}

func TestExecuteHTTPSubmissionRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		executor HTTPSubmissionExecutor
		draft    HTTPSubmissionDraft
		want     string
	}{
		{name: "executor", draft: mustDraft(t), want: "executor"},
		{name: "method", executor: &recordingHTTPSubmissionExecutor{}, draft: HTTPSubmissionDraft{Endpoint: "https://example.com", SurveyID: "s", Form: mapValues("q1", "a")}, want: "method"},
		{name: "endpoint", executor: &recordingHTTPSubmissionExecutor{}, draft: HTTPSubmissionDraft{Method: http.MethodPost, SurveyID: "s", Form: mapValues("q1", "a")}, want: "endpoint"},
		{name: "form", executor: &recordingHTTPSubmissionExecutor{}, draft: HTTPSubmissionDraft{Method: http.MethodPost, Endpoint: "https://example.com", SurveyID: "s"}, want: "form"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ExecuteHTTPSubmission(context.Background(), tt.executor, tt.draft)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ExecuteHTTPSubmission() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestExecuteHTTPSubmissionReturnsExecutorError(t *testing.T) {
	wantErr := errors.New("offline")
	_, err := ExecuteHTTPSubmission(context.Background(), &recordingHTTPSubmissionExecutor{err: wantErr}, mustDraft(t))
	if !errors.Is(err, wantErr) {
		t.Fatalf("ExecuteHTTPSubmission() error = %v, want executor error", err)
	}
}

func TestExecuteHTTPSubmissionClonesDraftAndResponse(t *testing.T) {
	draft := mustDraft(t)
	executor := &recordingHTTPSubmissionExecutor{
		response: HTTPSubmissionResponse{
			StatusCode: http.StatusAccepted,
			Header:     http.Header{"X-Original": []string{"value"}},
			Body:       "accepted",
		},
	}

	response, err := ExecuteHTTPSubmission(context.Background(), executor, draft)
	if err != nil {
		t.Fatalf("ExecuteHTTPSubmission() returned error: %v", err)
	}
	draft.Form.Set("q1", "mutated")
	draft.Header.Set("Referer", "mutated")
	response.Header.Set("X-Original", "mutated")

	if executor.calls[0].Form.Get("q1") != "a" || executor.calls[0].Header.Get("Referer") != "https://www.wjx.cn/vm/example.aspx" {
		t.Fatalf("recorded draft was mutated: %+v", executor.calls[0])
	}
	if executor.response.Header.Get("X-Original") != "value" {
		t.Fatalf("executor response header was mutated: %+v", executor.response.Header)
	}
}

type recordingHTTPSubmissionExecutor struct {
	response HTTPSubmissionResponse
	err      error
	calls    []HTTPSubmissionDraft
}

func (e *recordingHTTPSubmissionExecutor) ExecuteHTTPSubmission(ctx context.Context, draft HTTPSubmissionDraft) (HTTPSubmissionResponse, error) {
	if err := ctx.Err(); err != nil {
		return HTTPSubmissionResponse{}, err
	}
	e.calls = append(e.calls, cloneHTTPSubmissionDraft(draft))
	if e.err != nil {
		return HTTPSubmissionResponse{}, e.err
	}
	return cloneHTTPSubmissionResponse(e.response), nil
}

func mustDraft(t *testing.T) HTTPSubmissionDraft {
	t.Helper()
	draft, err := BuildHTTPSubmissionDraft("https://www.wjx.cn/vm/example.aspx", map[string]string{"q1": "a"})
	if err != nil {
		t.Fatalf("BuildHTTPSubmissionDraft() returned error: %v", err)
	}
	return draft
}

func mapValues(key string, value string) url.Values {
	return url.Values{key: []string{value}}
}
