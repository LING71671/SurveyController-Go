package wjx

import (
	"context"
	"net/http"
	"sync"
)

type DryRunHTTPSubmissionExecutor struct {
	mu    sync.Mutex
	calls []HTTPSubmissionDraft
}

func (e *DryRunHTTPSubmissionExecutor) ExecuteHTTPSubmission(ctx context.Context, draft HTTPSubmissionDraft) (HTTPSubmissionResponse, error) {
	if err := ctx.Err(); err != nil {
		return HTTPSubmissionResponse{}, err
	}
	if err := validateHTTPSubmissionDraft(draft); err != nil {
		return HTTPSubmissionResponse{}, err
	}

	e.mu.Lock()
	e.calls = append(e.calls, cloneHTTPSubmissionDraft(draft))
	e.mu.Unlock()

	header := http.Header{}
	header.Set("X-SurveyController-Dry-Run", "true")
	return HTTPSubmissionResponse{
		StatusCode: http.StatusAccepted,
		Header:     header,
		Body:       "submission accepted by local dry-run executor",
	}, nil
}

func (e *DryRunHTTPSubmissionExecutor) Drafts() []HTTPSubmissionDraft {
	if e == nil {
		return nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()

	drafts := make([]HTTPSubmissionDraft, 0, len(e.calls))
	for _, draft := range e.calls {
		drafts = append(drafts, cloneHTTPSubmissionDraft(draft))
	}
	return drafts
}
