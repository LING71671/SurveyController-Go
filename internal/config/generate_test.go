package config

import (
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/domain"
	"github.com/LING71671/SurveyController-go/internal/engine"
)

func TestFromSurveyDefinitionBuildsDefaultRunConfig(t *testing.T) {
	survey := domain.SurveyDefinition{
		Provider: domain.ProviderWJX,
		Title:    "Customer survey",
		URL:      " https://www.wjx.cn/vm/example.aspx ",
		Questions: []domain.QuestionDefinition{
			{
				ID:       " q1 ",
				Number:   1,
				Title:    "Choose one",
				Kind:     domain.QuestionKindSingle,
				Required: true,
			},
			{
				ID:     "q2",
				Number: 2,
				Title:  "Comment",
				Kind:   domain.QuestionKindText,
			},
		},
	}

	cfg, err := FromSurveyDefinition(survey)
	if err != nil {
		t.Fatalf("FromSurveyDefinition() returned error: %v", err)
	}
	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
	}
	if cfg.Survey.URL != "https://www.wjx.cn/vm/example.aspx" || cfg.Survey.Provider != "wjx" {
		t.Fatalf("Survey = %+v, want trimmed url and wjx provider", cfg.Survey)
	}
	if cfg.Run.Mode != engine.ModeHybrid || cfg.Run.Target != 1 || cfg.Run.Concurrency != 1 {
		t.Fatalf("Run defaults = %+v, want hybrid target 1 concurrency 1", cfg.Run)
	}
	if len(cfg.Questions) != 2 {
		t.Fatalf("len(Questions) = %d, want 2", len(cfg.Questions))
	}
	if cfg.Questions[0].ID != "q1" || cfg.Questions[0].Kind != "single" || !cfg.Questions[0].Required {
		t.Fatalf("first question = %+v, want mapped fields", cfg.Questions[0])
	}
	if cfg.Questions[0].Options == nil {
		t.Fatalf("first question options = nil, want editable empty map")
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("generated config did not validate: %v", err)
	}
}

func TestFromSurveyDefinitionRejectsInvalidSurveyDefinition(t *testing.T) {
	_, err := FromSurveyDefinition(domain.SurveyDefinition{
		Provider: domain.ProviderWJX,
		URL:      "https://www.wjx.cn/vm/example.aspx",
	})
	if err == nil || !strings.Contains(err.Error(), "survey title") {
		t.Fatalf("FromSurveyDefinition(invalid survey) error = %v, want survey title error", err)
	}
}

func TestFromSurveyDefinitionRejectsMissingURL(t *testing.T) {
	_, err := FromSurveyDefinition(domain.SurveyDefinition{
		Provider: domain.ProviderTencent,
		Title:    "Tencent survey",
		Questions: []domain.QuestionDefinition{
			{
				ID:     "q1",
				Number: 1,
				Title:  "Choose one",
				Kind:   domain.QuestionKindSingle,
			},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "survey.url") {
		t.Fatalf("FromSurveyDefinition(missing url) error = %v, want survey.url error", err)
	}
}
