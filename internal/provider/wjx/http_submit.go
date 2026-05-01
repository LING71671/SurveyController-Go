package wjx

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
)

const processJQPath = "/joinnew/processjq.ashx"

type HTTPSubmissionDraft struct {
	Method   string
	Endpoint string
	Header   http.Header
	Form     url.Values
	SurveyID string
}

func BuildHTTPSubmissionDraft(rawSurveyURL string, answers map[string]string) (HTTPSubmissionDraft, error) {
	if !(Provider{}).MatchURL(rawSurveyURL) {
		return HTTPSubmissionDraft{}, fmt.Errorf("wjx url is required")
	}
	surveyID, err := ExtractSurveyID(rawSurveyURL)
	if err != nil {
		return HTTPSubmissionDraft{}, err
	}
	if len(answers) == 0 {
		return HTTPSubmissionDraft{}, fmt.Errorf("answers are required")
	}

	endpoint, err := processEndpoint(rawSurveyURL)
	if err != nil {
		return HTTPSubmissionDraft{}, err
	}
	form := url.Values{}
	form.Set("curID", surveyID)
	form.Set("submittype", "1")
	for _, key := range sortedKeys(answers) {
		value := strings.TrimSpace(answers[key])
		if strings.TrimSpace(key) == "" {
			return HTTPSubmissionDraft{}, fmt.Errorf("answer key is required")
		}
		form.Set(strings.TrimSpace(key), value)
	}

	return HTTPSubmissionDraft{
		Method:   http.MethodPost,
		Endpoint: endpoint,
		Header: http.Header{
			"Content-Type": []string{"application/x-www-form-urlencoded"},
			"Referer":      []string{strings.TrimSpace(rawSurveyURL)},
		},
		Form:     form,
		SurveyID: surveyID,
	}, nil
}

func ExtractSurveyID(rawSurveyURL string) (string, error) {
	parsed, err := url.Parse(rawSurveyURL)
	if err != nil {
		return "", fmt.Errorf("parse wjx url: %w", err)
	}
	base := path.Base(parsed.EscapedPath())
	if base == "." || base == "/" || base == "" {
		return "", fmt.Errorf("survey id is required")
	}
	id := strings.TrimSuffix(base, path.Ext(base))
	id, err = url.PathUnescape(id)
	if err != nil {
		return "", fmt.Errorf("decode survey id: %w", err)
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return "", fmt.Errorf("survey id is required")
	}
	return id, nil
}

func processEndpoint(rawSurveyURL string) (string, error) {
	parsed, err := url.Parse(rawSurveyURL)
	if err != nil {
		return "", fmt.Errorf("parse wjx url: %w", err)
	}
	parsed.Path = processJQPath
	parsed.RawPath = ""
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func sortedKeys(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
