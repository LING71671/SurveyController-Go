package wjx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/engine"
)

func TestProviderMetadata(t *testing.T) {
	provider := Provider{}

	if provider.ID() != domain.ProviderWJX {
		t.Fatalf("ID() = %q, want %q", provider.ID(), domain.ProviderWJX)
	}
	if !provider.MatchURL("https://www.wjx.cn/vm/example.aspx") {
		t.Fatalf("MatchURL(wjx.cn) = false, want true")
	}
	if !provider.Capabilities().CanParse(engine.ModeHTTP) {
		t.Fatalf("CanParse(http) = false, want true")
	}
}

func TestParseHTML(t *testing.T) {
	survey := parseFixture(t, "survey.html")

	if survey.Provider != domain.ProviderWJX {
		t.Fatalf("Provider = %q, want wjx", survey.Provider)
	}
	if survey.Title != "问卷星 HTML 样例" {
		t.Fatalf("Title = %q, want fixture title", survey.Title)
	}
	if len(survey.Questions) != 5 {
		t.Fatalf("len(Questions) = %d, want 5", len(survey.Questions))
	}

	first := survey.Questions[0]
	if first.ID != "q1" || first.Kind != domain.QuestionKindSingle || !first.Required {
		t.Fatalf("first question = %+v, want required single q1", first)
	}
	if len(first.Options) != 2 || first.Options[0].Label != "浏览器" {
		t.Fatalf("first options = %+v, want parsed options", first.Options)
	}

	text := survey.Questions[2]
	if text.Kind != domain.QuestionKindText || len(text.Options) != 0 {
		t.Fatalf("text question = %+v, want text without options", text)
	}

	matrix := survey.Questions[4]
	if matrix.ID != "q5" || matrix.Kind != domain.QuestionKindMatrix {
		t.Fatalf("matrix question = %+v, want q5 matrix", matrix)
	}
	if len(matrix.Rows) != 2 || matrix.Rows[0].ID != "q5_r1" || matrix.Rows[1].Label != "性能" {
		t.Fatalf("matrix rows = %+v, want parsed data-row entries", matrix.Rows)
	}
	if len(matrix.Options) != 2 || matrix.Options[0].Value != "1" || matrix.Options[1].Value != "5" {
		t.Fatalf("matrix options = %+v, want parsed matrix columns", matrix.Options)
	}
}

func TestParseHTMLBlockedStates(t *testing.T) {
	tests := []struct {
		name string
		file string
		code apperr.Code
	}{
		{name: "paused", file: "paused.html", code: apperr.CodeParseFailed},
		{name: "closed", file: "closed.html", code: apperr.CodeParseFailed},
		{name: "verification", file: "verification.html", code: apperr.CodeVerificationNeeded},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("testdata", tt.file)
			file, err := os.Open(path)
			if err != nil {
				t.Fatalf("open fixture: %v", err)
			}
			defer file.Close()

			_, err = ParseHTML(file, "https://www.wjx.cn/vm/example.aspx")
			if !apperr.IsCode(err, tt.code) {
				t.Fatalf("ParseHTML() error = %v, want code %q", err, tt.code)
			}
		})
	}
}

func TestParseHTMLRejectsInvalidQuestionKind(t *testing.T) {
	_, err := ParseHTML(strings.NewReader(`
<!doctype html>
<html>
  <body>
    <h1 data-survey-title>title</h1>
    <div data-question="q1" data-kind="captcha">
      <div data-question-title>bad</div>
    </div>
  </body>
</html>`), "https://www.wjx.cn/vm/example.aspx")
	if !apperr.IsCode(err, apperr.CodeParseFailed) {
		t.Fatalf("ParseHTML() error = %v, want parse_failed", err)
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

	survey, err := ParseHTML(file, "https://www.wjx.cn/vm/example.aspx")
	if err != nil {
		t.Fatalf("ParseHTML(%s) returned error: %v", name, err)
	}
	return survey
}
