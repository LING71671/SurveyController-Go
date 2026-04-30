package credamo

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

	if provider.ID() != domain.ProviderCredamo {
		t.Fatalf("ID() = %q, want %q", provider.ID(), domain.ProviderCredamo)
	}
	if !provider.MatchURL("https://www.credamo.com/answer.html#/s/demo") {
		t.Fatalf("MatchURL(credamo) = false, want true")
	}
	if !provider.Capabilities().CanParse(engine.ModeBrowser) {
		t.Fatalf("CanParse(browser) = false, want true")
	}
}

func TestParseSnapshot(t *testing.T) {
	survey := parseFixture(t, "snapshot.json")

	if survey.Provider != domain.ProviderCredamo {
		t.Fatalf("Provider = %q, want credamo", survey.Provider)
	}
	if survey.Title != "Credamo DOM 样例" {
		t.Fatalf("Title = %q, want fixture title", survey.Title)
	}
	if len(survey.Questions) != 5 {
		t.Fatalf("len(Questions) = %d, want 5", len(survey.Questions))
	}

	first := survey.Questions[0]
	if first.Number != 1 || first.Kind != domain.QuestionKindSingle || !first.Required {
		t.Fatalf("first question = %+v, want required single Q1", first)
	}
	if len(first.Options) != 2 || first.Options[0].Label != "学习" {
		t.Fatalf("first options = %+v, want parsed options", first.Options)
	}

	forced := survey.Questions[2]
	if forced.Title != "本题检测是否认真作答，请选 非常不满意" {
		t.Fatalf("forced title = %q, want stripped type tag", forced.Title)
	}
	if forced.ProviderRaw["forced_option_index"] != 0 || forced.ProviderRaw["forced_option_text"] != "非常不满意" {
		t.Fatalf("forced ProviderRaw = %+v, want forced first option", forced.ProviderRaw)
	}

	arithmetic := survey.Questions[3]
	if arithmetic.ProviderRaw["forced_option_index"] != 1 || arithmetic.ProviderRaw["forced_option_text"] != "200" {
		t.Fatalf("arithmetic ProviderRaw = %+v, want option 200", arithmetic.ProviderRaw)
	}

	text := survey.Questions[4]
	forcedTexts := text.ProviderRaw["forced_texts"].([]string)
	if len(forcedTexts) != 1 || forcedTexts[0] != "你好" {
		t.Fatalf("forced_texts = %+v, want 你好", forcedTexts)
	}
	if text.Kind != domain.QuestionKindText {
		t.Fatalf("text kind = %q, want text", text.Kind)
	}
}

func TestParseSnapshotDoesNotDedupeReusedDOMIDAcrossPages(t *testing.T) {
	survey, err := ParseSnapshot(strings.NewReader(`{
  "title": "reused id",
  "questions": [
    {"question_id": "question-0", "question_num": "Q8", "title": "Q8 请问100+100等于多少", "question_kind": "single", "option_texts": ["300", "200"], "page": 2},
    {"question_id": "question-0", "question_num": "Q10", "title": "Q10 本题检测，请输入：你好", "question_kind": "text", "text_inputs": 1, "page": 4}
  ]
}`), "https://www.credamo.com/answer.html#/s/demo")
	if err != nil {
		t.Fatalf("ParseSnapshot() returned error: %v", err)
	}
	if len(survey.Questions) != 2 {
		t.Fatalf("len(Questions) = %d, want 2 for reused DOM id on different pages", len(survey.Questions))
	}
}

func TestParseSnapshotDedupesSameQuestion(t *testing.T) {
	survey, err := ParseSnapshot(strings.NewReader(`{
  "title": "duplicate",
  "questions": [
    {"question_id": "q1", "question_num": "Q1", "title": "Q1 请选择", "question_kind": "single", "option_texts": ["A", "B"], "page": 1},
    {"question_id": "q1", "question_num": "Q1", "title": "Q1 请选择", "question_kind": "single", "option_texts": ["A", "B"], "page": 1}
  ]
}`), "https://www.credamo.com/answer.html#/s/demo")
	if err != nil {
		t.Fatalf("ParseSnapshot() returned error: %v", err)
	}
	if len(survey.Questions) != 1 {
		t.Fatalf("len(Questions) = %d, want deduped question", len(survey.Questions))
	}
}

func TestParseSnapshotRejectsEmptyQuestionList(t *testing.T) {
	_, err := ParseSnapshot(strings.NewReader(`{"title":"empty","questions":[]}`), "https://www.credamo.com/answer.html#/s/demo")
	if !apperr.IsCode(err, apperr.CodeParseFailed) {
		t.Fatalf("ParseSnapshot(empty) error = %v, want parse_failed", err)
	}
}

func TestParseSnapshotRejectsInvalidJSON(t *testing.T) {
	_, err := ParseSnapshot(strings.NewReader(`{`), "https://www.credamo.com/answer.html#/s/demo")
	if !apperr.IsCode(err, apperr.CodeParseFailed) {
		t.Fatalf("ParseSnapshot(invalid JSON) error = %v, want parse_failed", err)
	}
}

func parseFixture(t *testing.T, name string) domain.SurveyDefinition {
	t.Helper()
	path := filepath.Join("testdata", name)
	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer file.Close()

	survey, err := ParseSnapshot(file, "https://www.credamo.com/answer.html#/s/demo")
	if err != nil {
		t.Fatalf("ParseSnapshot(%s) returned error: %v", name, err)
	}
	return survey
}
