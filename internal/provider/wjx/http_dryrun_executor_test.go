package wjx

import (
	"context"
	"errors"
	"net/http"
	"testing"
)

func TestDryRunHTTPSubmissionExecutorRecordsDraft(t *testing.T) {
	executor := &DryRunHTTPSubmissionExecutor{}
	draft, err := BuildHTTPSubmissionDraft("https://www.wjx.cn/vm/example.aspx", map[string]string{"q1": "a"})
	if err != nil {
		t.Fatalf("BuildHTTPSubmissionDraft() error = %v", err)
	}

	response, err := executor.ExecuteHTTPSubmission(context.Background(), draft)
	if err != nil {
		t.Fatalf("ExecuteHTTPSubmission() error = %v", err)
	}
	if response.StatusCode != http.StatusAccepted || response.Header.Get("X-SurveyController-Dry-Run") != "true" {
		t.Fatalf("response = %+v, want accepted dry-run response", response)
	}

	drafts := executor.Drafts()
	if len(drafts) != 1 || drafts[0].Form.Get("q1") != "a" {
		t.Fatalf("drafts = %+v, want recorded draft", drafts)
	}
}

func TestDryRunHTTPSubmissionExecutorClonesDrafts(t *testing.T) {
	executor := &DryRunHTTPSubmissionExecutor{}
	draft, err := BuildHTTPSubmissionDraft("https://www.wjx.cn/vm/example.aspx", map[string]string{"q1": "a"})
	if err != nil {
		t.Fatalf("BuildHTTPSubmissionDraft() error = %v", err)
	}
	if _, err := executor.ExecuteHTTPSubmission(context.Background(), draft); err != nil {
		t.Fatalf("ExecuteHTTPSubmission() error = %v", err)
	}

	first := executor.Drafts()
	first[0].Form.Set("q1", "mutated")
	first[0].Header.Set("Referer", "mutated")

	second := executor.Drafts()
	if second[0].Form.Get("q1") != "a" || second[0].Header.Get("Referer") != "https://www.wjx.cn/vm/example.aspx" {
		t.Fatalf("second drafts = %+v, want independent clone", second)
	}
}

func TestDryRunHTTPSubmissionExecutorRejectsInvalidDraft(t *testing.T) {
	executor := &DryRunHTTPSubmissionExecutor{}
	_, err := executor.ExecuteHTTPSubmission(context.Background(), HTTPSubmissionDraft{})
	if err == nil {
		t.Fatalf("ExecuteHTTPSubmission() error = nil, want invalid draft error")
	}
	if len(executor.Drafts()) != 0 {
		t.Fatalf("drafts = %+v, want no recorded invalid draft", executor.Drafts())
	}
}

func TestDryRunHTTPSubmissionExecutorHonorsCancelledContext(t *testing.T) {
	executor := &DryRunHTTPSubmissionExecutor{}
	draft, err := BuildHTTPSubmissionDraft("https://www.wjx.cn/vm/example.aspx", map[string]string{"q1": "a"})
	if err != nil {
		t.Fatalf("BuildHTTPSubmissionDraft() error = %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = executor.ExecuteHTTPSubmission(ctx, draft)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("ExecuteHTTPSubmission() error = %v, want context.Canceled", err)
	}
}
