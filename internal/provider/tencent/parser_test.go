package tencent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/apperr"
	"github.com/LING71671/SurveyController-go/internal/domain"
	"github.com/LING71671/SurveyController-go/internal/engine"
)

func TestProviderMetadata(t *testing.T) {
	provider := Provider{}

	if provider.ID() != domain.ProviderTencent {
		t.Fatalf("ID() = %q, want %q", provider.ID(), domain.ProviderTencent)
	}
	if !provider.MatchURL("https://wj.qq.com/s2/example") {
		t.Fatalf("MatchURL(wj.qq.com) = false, want true")
	}
	if provider.MatchURL("https://www.wjx.cn/vm/example.aspx") {
		t.Fatalf("MatchURL(wjx.cn) = true, want false")
	}
	if !provider.Capabilities().CanParse(engine.ModeHTTP) {
		t.Fatalf("CanParse(http) = false, want true")
	}
}

func TestParseAPI(t *testing.T) {
	survey := parseFixture(t, "survey_api.json")

	if survey.Provider != domain.ProviderTencent {
		t.Fatalf("Provider = %q, want tencent", survey.Provider)
	}
	if survey.ID != "tencent-minimal" {
		t.Fatalf("ID = %q, want fixture id", survey.ID)
	}
	if survey.Title != "腾讯问卷 API 样例" {
		t.Fatalf("Title = %q, want fixture title", survey.Title)
	}
	if len(survey.Questions) != 6 {
		t.Fatalf("len(Questions) = %d, want 6", len(survey.Questions))
	}

	first := survey.Questions[0]
	if first.ID != "q1" || first.Kind != domain.QuestionKindSingle || !first.Required {
		t.Fatalf("first question = %+v, want required single q1", first)
	}
	if len(first.Options) != 2 || first.Options[0].Label != "学习" || first.Options[0].Value != "study" {
		t.Fatalf("first options = %+v, want parsed options", first.Options)
	}

	matrix := survey.Questions[5]
	if matrix.Kind != domain.QuestionKindMatrix || len(matrix.Rows) != 2 || len(matrix.Options) != 2 {
		t.Fatalf("matrix question = %+v, want matrix rows and options", matrix)
	}
	if matrix.ProviderRaw["type"] != "matrix_radio" {
		t.Fatalf("matrix ProviderRaw type = %v, want matrix_radio", matrix.ProviderRaw["type"])
	}
}

func TestParseAPISupportsSurveyEnvelope(t *testing.T) {
	survey, err := ParseAPI(strings.NewReader(`{
  "code": 0,
  "survey": {
    "id": "survey-envelope",
    "title": "survey envelope",
    "questions": [
      {"id": "q1", "title": "name", "kind": "fill_blank"}
    ]
  }
}`), "https://wj.qq.com/s2/envelope")
	if err != nil {
		t.Fatalf("ParseAPI() returned error: %v", err)
	}
	if survey.ID != "survey-envelope" || survey.Questions[0].Kind != domain.QuestionKindText {
		t.Fatalf("survey = %+v, want survey envelope text question", survey)
	}
}

func TestParseAPILoginRequired(t *testing.T) {
	_, err := parseFixtureErr(t, "login_required.json")
	if !apperr.IsCode(err, apperr.CodeLoginRequired) {
		t.Fatalf("ParseAPI(login required) error = %v, want login_required", err)
	}
}

func TestParseAPIRejectsUnknownType(t *testing.T) {
	_, err := ParseAPI(strings.NewReader(`{
  "code": 0,
  "data": {
    "title": "bad type",
    "questions": [
      {"id": "q1", "title": "captcha", "type": "captcha"}
    ]
  }
}`), "https://wj.qq.com/s2/bad")
	if !apperr.IsCode(err, apperr.CodeParseFailed) {
		t.Fatalf("ParseAPI(unknown type) error = %v, want parse_failed", err)
	}
}

func TestParseAPIRejectsInvalidSurvey(t *testing.T) {
	_, err := ParseAPI(strings.NewReader(`{
  "code": 0,
  "data": {
    "questions": [
      {"id": "q1", "title": "missing survey title", "type": "radio"}
    ]
  }
}`), "https://wj.qq.com/s2/missing-title")
	if !apperr.IsCode(err, apperr.CodeParseFailed) {
		t.Fatalf("ParseAPI(invalid survey) error = %v, want parse_failed", err)
	}
}

func parseFixture(t *testing.T, name string) domain.SurveyDefinition {
	t.Helper()
	survey, err := parseFixtureErr(t, name)
	if err != nil {
		t.Fatalf("ParseAPI(%s) returned error: %v", name, err)
	}
	return survey
}

func parseFixtureErr(t *testing.T, name string) (domain.SurveyDefinition, error) {
	t.Helper()
	path := filepath.Join("testdata", name)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer file.Close()

	return ParseAPI(file, "https://wj.qq.com/s2/example")
}
