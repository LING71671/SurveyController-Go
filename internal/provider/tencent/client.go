package tencent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/httpclient"
)

var (
	surveyURLRE         = regexp.MustCompile(`(?i)/s\d+/(\d+)/([A-Za-z0-9_-]+)/?`)
	loginRequiredTokens = []string{
		"open.weixin.qq.com/connect/confirm",
		"wj.qq.com/r/login.html",
		"/r/login.html",
		"need login",
		"login required",
		"require login",
		"未登录",
		"需登录",
		"需要登录",
	}
)

type HTTPClient interface {
	Do(ctx context.Context, options httpclient.RequestOptions) (*http.Response, error)
}

type APIClientOptions struct {
	Locales []string
	Now     func() time.Time
}

type Identifiers struct {
	SurveyID string
	Hash     string
}

func ParseFromClient(ctx context.Context, client HTTPClient, rawURL string, options APIClientOptions) (domain.SurveyDefinition, error) {
	if client == nil {
		return domain.SurveyDefinition{}, fmt.Errorf("tencent api client is required")
	}
	ids, err := ExtractIdentifiers(rawURL)
	if err != nil {
		return domain.SurveyDefinition{}, err
	}
	locales := options.Locales
	if len(locales) == 0 {
		locales = []string{"zhs", "zht", "zh", "en"}
	}
	now := options.Now
	if now == nil {
		now = time.Now
	}

	if _, err := requestEndpoint(ctx, client, ids, "session", "", rawURL, now); err != nil {
		return domain.SurveyDefinition{}, err
	}

	var lastErr error
	for _, locale := range locales {
		metaPayload, err := requestEndpoint(ctx, client, ids, "meta", locale, rawURL, now)
		if err != nil {
			if apperr.IsCode(err, apperr.CodeLoginRequired) {
				return domain.SurveyDefinition{}, err
			}
			lastErr = err
			continue
		}
		questionPayload, err := requestEndpoint(ctx, client, ids, "questions", locale, rawURL, now)
		if err != nil {
			if apperr.IsCode(err, apperr.CodeLoginRequired) {
				return domain.SurveyDefinition{}, err
			}
			lastErr = err
			continue
		}

		api, err := combineEndpointPayloads(metaPayload, questionPayload, ids, rawURL)
		if err != nil {
			lastErr = err
			continue
		}
		return buildSurvey(api, rawURL)
	}
	if lastErr != nil {
		return domain.SurveyDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "parse tencent api endpoints", lastErr)
	}
	return domain.SurveyDefinition{}, apperr.New(apperr.CodeParseFailed, "tencent api returned no usable locale")
}

func ExtractIdentifiers(rawURL string) (Identifiers, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return Identifiers{}, fmt.Errorf("parse tencent url: %w", err)
	}
	if parsed.Hostname() != "" && !strings.EqualFold(parsed.Hostname(), "wj.qq.com") {
		return Identifiers{}, fmt.Errorf("tencent url host must be wj.qq.com")
	}
	matches := surveyURLRE.FindStringSubmatch(parsed.EscapedPath())
	if len(matches) != 3 {
		return Identifiers{}, fmt.Errorf("tencent survey url must contain /s2/{survey_id}/{hash}")
	}
	return Identifiers{SurveyID: matches[1], Hash: matches[2]}, nil
}

func requestEndpoint(ctx context.Context, client HTTPClient, ids Identifiers, endpoint string, locale string, referer string, now func() time.Time) (json.RawMessage, error) {
	endpointURL, err := buildEndpointURL(ids, endpoint, locale, now)
	if err != nil {
		return nil, err
	}
	header := http.Header{}
	header.Set("Accept", "application/json, text/plain, */*")
	header.Set("Origin", "https://wj.qq.com")
	header.Set("Referer", buildSurveyPageURL(ids, referer))

	response, err := client.Do(ctx, httpclient.RequestOptions{
		Method: http.MethodGet,
		URL:    endpointURL,
		Header: header,
	})
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeParseFailed, "request tencent api", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, apperr.Wrap(apperr.CodeParseFailed, "read tencent api response", err)
	}
	if responseIndicatesLogin(response, body) {
		return nil, apperr.New(apperr.CodeLoginRequired, "tencent survey requires login")
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, apperr.New(apperr.CodeParseFailed, fmt.Sprintf("tencent api returned HTTP %d", response.StatusCode))
	}
	data, err := ensureEndpointOK(body, endpoint)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func buildEndpointURL(ids Identifiers, endpoint string, locale string, now func() time.Time) (string, error) {
	if strings.TrimSpace(ids.SurveyID) == "" || strings.TrimSpace(ids.Hash) == "" {
		return "", fmt.Errorf("tencent identifiers are required")
	}
	if strings.TrimSpace(endpoint) == "" {
		return "", fmt.Errorf("tencent endpoint is required")
	}
	values := url.Values{}
	values.Set("_", strconv.FormatInt(now().UnixMilli(), 10))
	values.Set("hash", ids.Hash)
	if strings.TrimSpace(locale) != "" {
		values.Set("locale", strings.TrimSpace(locale))
	}
	return fmt.Sprintf("https://wj.qq.com/api/v2/respondent/surveys/%s/%s?%s", ids.SurveyID, endpoint, values.Encode()), nil
}

func buildSurveyPageURL(ids Identifiers, fallback string) string {
	if ids.SurveyID != "" && ids.Hash != "" {
		return fmt.Sprintf("https://wj.qq.com/s2/%s/%s/", ids.SurveyID, ids.Hash)
	}
	return fallback
}

func ensureEndpointOK(body []byte, endpoint string) (json.RawMessage, error) {
	var envelope struct {
		Code    any             `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&envelope); err != nil {
		return nil, apperr.Wrap(apperr.CodeParseFailed, "parse tencent api envelope", err)
	}
	if payloadIndicatesLogin(envelope) {
		return nil, apperr.New(apperr.CodeLoginRequired, "tencent survey requires login")
	}
	if !endpointCodeOK(envelope.Code) {
		message := strings.TrimSpace(envelope.Message)
		if message == "" {
			message = fmt.Sprintf("tencent api returned non-ok code for %s", endpoint)
		}
		return nil, apperr.New(apperr.CodeParseFailed, message)
	}
	if len(envelope.Data) == 0 || string(envelope.Data) == "null" {
		return nil, apperr.New(apperr.CodeParseFailed, fmt.Sprintf("tencent api missing data for %s", endpoint))
	}
	return envelope.Data, nil
}

func endpointCodeOK(code any) bool {
	switch value := code.(type) {
	case nil:
		return false
	case string:
		normalized := strings.ToUpper(strings.TrimSpace(value))
		return normalized == "OK" || normalized == "0"
	case float64:
		return value == 0
	case json.Number:
		if value.String() == "0" {
			return true
		}
		parsed, err := value.Int64()
		return err == nil && parsed == 0
	default:
		return false
	}
}

func combineEndpointPayloads(metaPayload json.RawMessage, questionPayload json.RawMessage, ids Identifiers, rawURL string) (apiSurvey, error) {
	var meta struct {
		ID          string `json:"id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		URL         string `json:"url"`
	}
	if err := json.Unmarshal(metaPayload, &meta); err != nil {
		return apiSurvey{}, apperr.Wrap(apperr.CodeParseFailed, "parse tencent meta data", err)
	}
	var questions struct {
		Questions []apiQuestion `json:"questions"`
	}
	if err := json.Unmarshal(questionPayload, &questions); err != nil {
		return apiSurvey{}, apperr.Wrap(apperr.CodeParseFailed, "parse tencent questions data", err)
	}
	if len(questions.Questions) == 0 {
		return apiSurvey{}, apperr.New(apperr.CodeParseFailed, "tencent questions data is empty")
	}
	return apiSurvey{
		ID:          firstNonEmpty(meta.ID, ids.SurveyID),
		Title:       strings.TrimSpace(meta.Title),
		Description: strings.TrimSpace(meta.Description),
		URL:         firstNonEmpty(meta.URL, rawURL),
		Questions:   questions.Questions,
	}, nil
}

func responseIndicatesLogin(response *http.Response, body []byte) bool {
	if response == nil {
		return false
	}
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return true
	}
	if location := response.Header.Get("Location"); isLoginRequiredText(location) {
		return true
	}
	return isLoginRequiredText(string(body))
}

func payloadIndicatesLogin(value any) bool {
	data, err := json.Marshal(value)
	if err != nil {
		return false
	}
	return isLoginRequiredText(string(data))
}

func isLoginRequiredText(value string) bool {
	text := strings.ToLower(strings.TrimSpace(value))
	if text == "" {
		return false
	}
	for _, token := range loginRequiredTokens {
		if strings.Contains(text, strings.ToLower(token)) {
			return true
		}
	}
	return false
}
