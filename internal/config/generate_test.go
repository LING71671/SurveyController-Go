package config

import (
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/engine"
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
				Options: []domain.OptionDefinition{
					{ID: "a", Label: "A"},
					{ID: "b", Label: "B"},
				},
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
	weights, ok := cfg.Questions[0].Options["weights"].([]map[string]any)
	if !ok {
		t.Fatalf("weights = %#v, want generated option weight maps", cfg.Questions[0].Options["weights"])
	}
	if len(weights) != 2 || weights[0]["option_id"] != "a" || weights[0]["weight"] != 1 {
		t.Fatalf("weights = %+v, want default weights for options", weights)
	}
	textOptions, ok := cfg.Questions[1].Options["text"].(map[string]any)
	if !ok {
		t.Fatalf("text question options = %+v, want text answer skeleton", cfg.Questions[1].Options)
	}
	values, ok := textOptions["values"].([]string)
	if textOptions["mode"] != "fixed" || !ok || len(values) != 1 || values[0] != "sample answer" {
		t.Fatalf("text skeleton = %+v, want fixed sample answer", textOptions)
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("generated config did not validate: %v", err)
	}
}

func TestFromSurveyDefinitionBuildsTextAreaSkeleton(t *testing.T) {
	survey := domain.SurveyDefinition{
		Provider: domain.ProviderWJX,
		Title:    "Text area survey",
		URL:      "https://www.wjx.cn/vm/example.aspx",
		Questions: []domain.QuestionDefinition{
			{
				ID:    "q1",
				Title: "Long comment",
				Kind:  domain.QuestionKindTextarea,
			},
		},
	}

	cfg, err := FromSurveyDefinition(survey)
	if err != nil {
		t.Fatalf("FromSurveyDefinition() returned error: %v", err)
	}
	if _, ok := cfg.Questions[0].Options["text"].(map[string]any); !ok {
		t.Fatalf("textarea options = %+v, want text answer skeleton", cfg.Questions[0].Options)
	}
}

func TestFromSurveyDefinitionBuildsMatrixWeightSkeleton(t *testing.T) {
	survey := domain.SurveyDefinition{
		Provider: domain.ProviderTencent,
		Title:    "Matrix survey",
		URL:      "https://wj.qq.com/s2/example",
		Questions: []domain.QuestionDefinition{
			{
				ID:     "matrix1",
				Number: 1,
				Title:  "Matrix question",
				Kind:   domain.QuestionKindMatrix,
				Rows: []domain.OptionDefinition{
					{ID: "row1", Label: "Row 1"},
					{ID: "row2", Label: "Row 2"},
				},
				Options: []domain.OptionDefinition{
					{ID: "opt1", Label: "Option 1"},
					{ID: "opt2", Label: "Option 2"},
				},
			},
		},
	}

	cfg, err := FromSurveyDefinition(survey)
	if err != nil {
		t.Fatalf("FromSurveyDefinition() returned error: %v", err)
	}
	matrixWeights, ok := cfg.Questions[0].Options["matrix_weights"].([]map[string]any)
	if !ok {
		t.Fatalf("matrix_weights = %#v, want generated matrix weight maps", cfg.Questions[0].Options["matrix_weights"])
	}
	if len(matrixWeights) != 2 || matrixWeights[0]["row_id"] != "row1" {
		t.Fatalf("matrix_weights = %+v, want row weights", matrixWeights)
	}
	rowWeights, ok := matrixWeights[0]["weights"].([]map[string]any)
	if !ok {
		t.Fatalf("row weights = %#v, want option weight maps", matrixWeights[0]["weights"])
	}
	if len(rowWeights) != 2 || rowWeights[1]["option_id"] != "opt2" || rowWeights[1]["weight"] != 1 {
		t.Fatalf("row weights = %+v, want default option weights", rowWeights)
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
