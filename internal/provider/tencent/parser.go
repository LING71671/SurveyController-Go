package tencent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/provider"
)

type Provider struct{}

func (Provider) ID() provider.ProviderID {
	return domain.ProviderTencent
}

func (Provider) MatchURL(rawURL string) bool {
	return provider.MatchHost(rawURL, "wj.qq.com")
}

func (Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		ParseHTTP: true,
	}
}

func (p Provider) Parse(ctx context.Context, rawURL string) (provider.SurveyDefinition, error) {
	_ = ctx
	_ = rawURL
	return provider.SurveyDefinition{}, apperr.New(apperr.CodeProviderUnsupported, "tencent provider requires API JSON for this parser prototype")
}

type apiResponse struct {
	Code          int       `json:"code"`
	Message       string    `json:"message"`
	LoginRequired bool      `json:"login_required"`
	Data          apiSurvey `json:"data"`
	Survey        apiSurvey `json:"survey"`
}

type apiSurvey struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	URL         string        `json:"url"`
	Questions   []apiQuestion `json:"questions"`
}

type apiQuestion struct {
	ID       string      `json:"id"`
	Number   int         `json:"number"`
	Title    string      `json:"title"`
	Type     string      `json:"type"`
	Kind     string      `json:"kind"`
	Required bool        `json:"required"`
	Options  []apiOption `json:"options"`
	Rows     []apiOption `json:"rows"`
}

type apiOption struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Value string `json:"value"`
}

func ParseAPI(r io.Reader, rawURL string) (domain.SurveyDefinition, error) {
	var response apiResponse
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&response); err != nil {
		return domain.SurveyDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "parse tencent api json", err)
	}
	if response.LoginRequired || response.Code == 401 || response.Code == 403 {
		return domain.SurveyDefinition{}, apperr.New(apperr.CodeLoginRequired, "tencent survey requires login")
	}
	if response.Code != 0 {
		message := strings.TrimSpace(response.Message)
		if message == "" {
			message = fmt.Sprintf("tencent api returned code %d", response.Code)
		}
		return domain.SurveyDefinition{}, apperr.New(apperr.CodeParseFailed, message)
	}

	api := response.Data
	if api.Title == "" && response.Survey.Title != "" {
		api = response.Survey
	}
	return buildSurvey(api, rawURL)
}

func buildSurvey(api apiSurvey, rawURL string) (domain.SurveyDefinition, error) {
	survey := domain.SurveyDefinition{
		Provider:    domain.ProviderTencent,
		ID:          strings.TrimSpace(api.ID),
		Title:       strings.TrimSpace(api.Title),
		Description: strings.TrimSpace(api.Description),
		URL:         firstNonEmpty(api.URL, rawURL),
		Questions:   make([]domain.QuestionDefinition, 0, len(api.Questions)),
		ProviderRaw: map[string]any{
			"source": "tencent_api",
		},
	}
	for index, question := range api.Questions {
		parsed, err := parseQuestion(question, index+1)
		if err != nil {
			return domain.SurveyDefinition{}, err
		}
		survey.Questions = append(survey.Questions, parsed)
	}
	if err := survey.Validate(); err != nil {
		return domain.SurveyDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "validate tencent survey", err)
	}
	return survey, nil
}

func parseQuestion(api apiQuestion, fallbackNumber int) (domain.QuestionDefinition, error) {
	rawType := firstNonEmpty(api.Type, api.Kind)
	kind, err := parseKind(rawType)
	if err != nil {
		return domain.QuestionDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "parse tencent question kind", err)
	}
	number := api.Number
	if number == 0 {
		number = fallbackNumber
	}
	question := domain.QuestionDefinition{
		ID:       strings.TrimSpace(api.ID),
		Number:   number,
		Title:    strings.TrimSpace(api.Title),
		Kind:     kind,
		Required: api.Required,
		Options:  parseOptions(api.Options),
		Rows:     parseOptions(api.Rows),
		ProviderRaw: map[string]any{
			"type": rawType,
		},
	}
	if err := question.Validate(); err != nil {
		return domain.QuestionDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "validate tencent question", err)
	}
	return question, nil
}

func parseKind(raw string) (domain.QuestionKind, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "radio", "single", "single_choice":
		return domain.QuestionKindSingle, nil
	case "checkbox", "multiple", "multi", "multiple_choice":
		return domain.QuestionKindMultiple, nil
	case "select", "dropdown":
		return domain.QuestionKindDropdown, nil
	case "text", "input", "blank", "fill_blank":
		return domain.QuestionKindText, nil
	case "textarea", "long_text":
		return domain.QuestionKindTextarea, nil
	case "nps", "star", "rating", "score":
		return domain.QuestionKindRating, nil
	case "matrix", "matrix_radio", "matrix_single", "matrix_checkbox", "matrix_multiple", "matrix_star", "matrix_score":
		return domain.QuestionKindMatrix, nil
	case "ranking", "sort":
		return domain.QuestionKindRanking, nil
	default:
		return domain.ParseQuestionKind(raw)
	}
}

func parseOptions(apiOptions []apiOption) []domain.OptionDefinition {
	options := make([]domain.OptionDefinition, 0, len(apiOptions))
	for index, option := range apiOptions {
		id := strings.TrimSpace(option.ID)
		if id == "" {
			id = strconv.Itoa(index + 1)
		}
		value := strings.TrimSpace(option.Value)
		if value == "" {
			value = id
		}
		options = append(options, domain.OptionDefinition{
			ID:    id,
			Label: strings.TrimSpace(option.Label),
			Value: value,
			ProviderRaw: map[string]any{
				"index": index,
			},
		})
	}
	return options
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
