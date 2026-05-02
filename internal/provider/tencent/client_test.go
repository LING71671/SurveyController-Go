package tencent

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/httpclient"
)

func TestExtractIdentifiers(t *testing.T) {
	ids, err := ExtractIdentifiers("https://wj.qq.com/s2/123456/abc_DEF-9/?foo=bar")
	if err != nil {
		t.Fatalf("ExtractIdentifiers() returned error: %v", err)
	}
	if ids.SurveyID != "123456" || ids.Hash != "abc_DEF-9" {
		t.Fatalf("ids = %+v, want survey id and hash", ids)
	}
}

func TestExtractIdentifiersRejectsInvalidURL(t *testing.T) {
	if _, err := ExtractIdentifiers("https://example.com/s2/123/hash"); err == nil {
		t.Fatal("ExtractIdentifiers(wrong host) returned nil error, want failure")
	}
	if _, err := ExtractIdentifiers("https://wj.qq.com/s2/not-enough"); err == nil {
		t.Fatal("ExtractIdentifiers(invalid path) returned nil error, want failure")
	}
}

func TestParseFromClientRequestsSessionMetaAndQuestions(t *testing.T) {
	client := newFakeAPIClient(map[string]fakeAPIResponse{
		"session:": {
			body: `{"code":"OK","data":{"session":"ok"}}`,
		},
		"meta:zhs": {
			body: `{"code":"OK","data":{"id":"123456","title":"腾讯问卷 API Client 样例","description":"desc","url":"https://wj.qq.com/s2/123456/hash/"}}`,
		},
		"questions:zhs": {
			body: `{"code":"OK","data":{"questions":[{"id":"q1","title":"请选择","type":"radio","required":true,"options":[{"id":"a","label":"A","value":"a"}]}]}}`,
		},
	})

	survey, err := ParseFromClient(context.Background(), client, "https://wj.qq.com/s2/123456/hash/", APIClientOptions{
		Locales: []string{"zhs"},
		Now:     fixedNow,
	})
	if err != nil {
		t.Fatalf("ParseFromClient() returned error: %v", err)
	}
	if survey.Provider != domain.ProviderTencent || survey.ID != "123456" || survey.Title != "腾讯问卷 API Client 样例" {
		t.Fatalf("survey = %+v, want parsed Tencent survey", survey)
	}
	if len(survey.Questions) != 1 || survey.Questions[0].Kind != domain.QuestionKindSingle || !survey.Questions[0].Required {
		t.Fatalf("questions = %+v, want required single question", survey.Questions)
	}
	wantOrder := []string{"session:", "meta:zhs", "questions:zhs"}
	if strings.Join(client.calls, ",") != strings.Join(wantOrder, ",") {
		t.Fatalf("calls = %+v, want %+v", client.calls, wantOrder)
	}
	firstURL, _ := url.Parse(client.urls[0])
	if firstURL.Query().Get("_") != "1700000000000" || firstURL.Query().Get("hash") != "hash" {
		t.Fatalf("first query = %s, want timestamp and hash", firstURL.RawQuery)
	}
	if client.headers[0].Get("Origin") != "https://wj.qq.com" {
		t.Fatalf("Origin = %q, want wj.qq.com", client.headers[0].Get("Origin"))
	}
}

func TestParseFromClientFallsBackLocales(t *testing.T) {
	client := newFakeAPIClient(map[string]fakeAPIResponse{
		"session:": {
			body: `{"code":"OK","data":{"session":"ok"}}`,
		},
		"meta:zhs": {
			body: `{"code":"OK","data":{"id":"123456","title":"bad locale"}}`,
		},
		"questions:zhs": {
			body: `{"code":"ERR","message":"locale unavailable","data":{}}`,
		},
		"meta:en": {
			body: `{"code":"OK","data":{"id":"123456","title":"English title"}}`,
		},
		"questions:en": {
			body: `{"code":"OK","data":{"questions":[{"id":"q1","title":"Name","type":"fill_blank"}]}}`,
		},
	})

	survey, err := ParseFromClient(context.Background(), client, "https://wj.qq.com/s2/123456/hash/", APIClientOptions{
		Locales: []string{"zhs", "en"},
		Now:     fixedNow,
	})
	if err != nil {
		t.Fatalf("ParseFromClient() returned error: %v", err)
	}
	if survey.Title != "English title" || survey.Questions[0].Kind != domain.QuestionKindText {
		t.Fatalf("survey = %+v, want fallback locale result", survey)
	}
	wantOrder := []string{"session:", "meta:zhs", "questions:zhs", "meta:en", "questions:en"}
	if strings.Join(client.calls, ",") != strings.Join(wantOrder, ",") {
		t.Fatalf("calls = %+v, want %+v", client.calls, wantOrder)
	}
}

func TestParseFromClientDetectsLoginRequired(t *testing.T) {
	client := newFakeAPIClient(map[string]fakeAPIResponse{
		"session:": {
			status: http.StatusUnauthorized,
			body:   `login required`,
		},
	})

	_, err := ParseFromClient(context.Background(), client, "https://wj.qq.com/s2/123456/hash/", APIClientOptions{
		Locales: []string{"zhs"},
		Now:     fixedNow,
	})
	if !apperr.IsCode(err, apperr.CodeLoginRequired) {
		t.Fatalf("ParseFromClient(login) error = %v, want login_required", err)
	}
}

func TestParseFromClientDetectsLoginLocation(t *testing.T) {
	client := newFakeAPIClient(map[string]fakeAPIResponse{
		"session:": {
			status: http.StatusFound,
			header: http.Header{
				"Location": []string{"https://wj.qq.com/r/login.html"},
			},
			body: `redirect`,
		},
	})

	_, err := ParseFromClient(context.Background(), client, "https://wj.qq.com/s2/123456/hash/", APIClientOptions{
		Locales: []string{"zhs"},
		Now:     fixedNow,
	})
	if !apperr.IsCode(err, apperr.CodeLoginRequired) {
		t.Fatalf("ParseFromClient(login location) error = %v, want login_required", err)
	}
}

func TestParseFromClientReportsAPIError(t *testing.T) {
	client := newFakeAPIClient(map[string]fakeAPIResponse{
		"session:": {
			body: `{"code":"ERR","message":"bad session","data":{}}`,
		},
	})

	_, err := ParseFromClient(context.Background(), client, "https://wj.qq.com/s2/123456/hash/", APIClientOptions{
		Locales: []string{"zhs"},
		Now:     fixedNow,
	})
	if !apperr.IsCode(err, apperr.CodeParseFailed) {
		t.Fatalf("ParseFromClient(api error) error = %v, want parse_failed", err)
	}
}

type fakeAPIResponse struct {
	status int
	header http.Header
	body   string
}

type fakeAPIClient struct {
	responses map[string]fakeAPIResponse
	calls     []string
	urls      []string
	headers   []http.Header
}

func newFakeAPIClient(responses map[string]fakeAPIResponse) *fakeAPIClient {
	return &fakeAPIClient{responses: responses}
}

func (c *fakeAPIClient) Do(ctx context.Context, options httpclient.RequestOptions) (*http.Response, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	parsed, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	endpoint := parts[len(parts)-1]
	locale := parsed.Query().Get("locale")
	key := endpoint + ":" + locale
	c.calls = append(c.calls, key)
	c.urls = append(c.urls, options.URL)
	c.headers = append(c.headers, options.Header.Clone())
	response := c.responses[key]
	status := response.status
	if status == 0 {
		status = http.StatusOK
	}
	header := response.header
	if header == nil {
		header = http.Header{}
	}
	body := response.body
	if body == "" {
		body = `{"code":"OK","data":{}}`
	}
	return &http.Response{
		StatusCode: status,
		Header:     header,
		Body:       io.NopCloser(strings.NewReader(body)),
	}, nil
}

func fixedNow() time.Time {
	return time.UnixMilli(1700000000000)
}
