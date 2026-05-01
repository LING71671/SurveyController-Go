package wjx

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type HTTPSubmissionResponse struct {
	StatusCode int
	Header     http.Header
	Body       string
}

type HTTPSubmissionExecutor interface {
	ExecuteHTTPSubmission(ctx context.Context, draft HTTPSubmissionDraft) (HTTPSubmissionResponse, error)
}

func ExecuteHTTPSubmission(ctx context.Context, executor HTTPSubmissionExecutor, draft HTTPSubmissionDraft) (HTTPSubmissionResponse, error) {
	if executor == nil {
		return HTTPSubmissionResponse{}, fmt.Errorf("http submission executor is required")
	}
	if err := validateHTTPSubmissionDraft(draft); err != nil {
		return HTTPSubmissionResponse{}, err
	}
	response, err := executor.ExecuteHTTPSubmission(ctx, cloneHTTPSubmissionDraft(draft))
	if err != nil {
		return HTTPSubmissionResponse{}, err
	}
	return cloneHTTPSubmissionResponse(response), nil
}

func validateHTTPSubmissionDraft(draft HTTPSubmissionDraft) error {
	if strings.TrimSpace(draft.Method) == "" {
		return fmt.Errorf("submission method is required")
	}
	if strings.TrimSpace(draft.Endpoint) == "" {
		return fmt.Errorf("submission endpoint is required")
	}
	if strings.TrimSpace(draft.SurveyID) == "" {
		return fmt.Errorf("survey id is required")
	}
	if len(draft.Form) == 0 {
		return fmt.Errorf("submission form is required")
	}
	return nil
}

func cloneHTTPSubmissionDraft(draft HTTPSubmissionDraft) HTTPSubmissionDraft {
	return HTTPSubmissionDraft{
		Method:   draft.Method,
		Endpoint: draft.Endpoint,
		Header:   cloneHTTPHeader(draft.Header),
		Form:     cloneURLValues(draft.Form),
		SurveyID: draft.SurveyID,
	}
}

func cloneHTTPSubmissionResponse(response HTTPSubmissionResponse) HTTPSubmissionResponse {
	return HTTPSubmissionResponse{
		StatusCode: response.StatusCode,
		Header:     cloneHTTPHeader(response.Header),
		Body:       response.Body,
	}
}

func cloneHTTPHeader(header http.Header) http.Header {
	if len(header) == 0 {
		return nil
	}
	cloned := make(http.Header, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func cloneURLValues(values url.Values) url.Values {
	if len(values) == 0 {
		return nil
	}
	cloned := make(url.Values, len(values))
	for key, items := range values {
		cloned[key] = append([]string(nil), items...)
	}
	return cloned
}
