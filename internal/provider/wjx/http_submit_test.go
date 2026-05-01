package wjx

import (
	"net/http"
	"strings"
	"testing"
)

func TestBuildHTTPSubmissionDraft(t *testing.T) {
	draft, err := BuildHTTPSubmissionDraft("https://www.wjx.cn/vm/example.aspx", map[string]string{
		"q1": "a",
		"q2": "b,c",
	})
	if err != nil {
		t.Fatalf("BuildHTTPSubmissionDraft() returned error: %v", err)
	}
	if draft.Method != http.MethodPost {
		t.Fatalf("Method = %q, want POST", draft.Method)
	}
	if draft.Endpoint != "https://www.wjx.cn/joinnew/processjq.ashx" {
		t.Fatalf("Endpoint = %q, want processjq endpoint", draft.Endpoint)
	}
	if draft.SurveyID != "example" || draft.Form.Get("curID") != "example" {
		t.Fatalf("survey id/form = %q/%q, want example", draft.SurveyID, draft.Form.Get("curID"))
	}
	if draft.Form.Get("submittype") != "1" || draft.Form.Get("q1") != "a" || draft.Form.Get("q2") != "b,c" {
		t.Fatalf("Form = %+v, want submit type and answers", draft.Form)
	}
	if draft.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		t.Fatalf("Content-Type = %q, want form encoding", draft.Header.Get("Content-Type"))
	}
	if draft.Header.Get("Referer") != "https://www.wjx.cn/vm/example.aspx" {
		t.Fatalf("Referer = %q, want survey url", draft.Header.Get("Referer"))
	}
}

func TestBuildHTTPSubmissionDraftSupportsVJPath(t *testing.T) {
	draft, err := BuildHTTPSubmissionDraft("https://sub.wjx.top/vj/abc123.aspx?from=qr", map[string]string{"q1": "x"})
	if err != nil {
		t.Fatalf("BuildHTTPSubmissionDraft() returned error: %v", err)
	}
	if draft.SurveyID != "abc123" {
		t.Fatalf("SurveyID = %q, want abc123", draft.SurveyID)
	}
	if draft.Endpoint != "https://sub.wjx.top/joinnew/processjq.ashx" {
		t.Fatalf("Endpoint = %q, want query-free process endpoint", draft.Endpoint)
	}
}

func TestBuildHTTPSubmissionDraftRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		answers map[string]string
		want    string
	}{
		{name: "host", rawURL: "https://example.com/vm/demo.aspx", answers: map[string]string{"q1": "a"}, want: "wjx url"},
		{name: "survey id", rawURL: "https://www.wjx.cn/", answers: map[string]string{"q1": "a"}, want: "survey id"},
		{name: "answers", rawURL: "https://www.wjx.cn/vm/demo.aspx", want: "answers"},
		{name: "answer key", rawURL: "https://www.wjx.cn/vm/demo.aspx", answers: map[string]string{" ": "a"}, want: "answer key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildHTTPSubmissionDraft(tt.rawURL, tt.answers)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("BuildHTTPSubmissionDraft() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestExtractSurveyID(t *testing.T) {
	got, err := ExtractSurveyID("https://www.wjx.cn/vm/demo%201.aspx")
	if err != nil {
		t.Fatalf("ExtractSurveyID() returned error: %v", err)
	}
	if got != "demo 1" {
		t.Fatalf("ExtractSurveyID() = %q, want decoded id", got)
	}
}
