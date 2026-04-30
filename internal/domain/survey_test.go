package domain

import (
	"strings"
	"testing"
)

func TestSurveyDefinitionValidate(t *testing.T) {
	survey := SurveyDefinition{
		Provider: ProviderWJX,
		Title:    "满意度调查",
		ProviderRaw: map[string]any{
			"activity_id": "123",
		},
		Questions: []QuestionDefinition{
			{
				ID:       "q1",
				Number:   1,
				Title:    "请选择",
				Kind:     QuestionKindSingle,
				Required: true,
				Options: []OptionDefinition{
					{ID: "o1", Label: "A", Value: "1"},
				},
				ProviderRaw: map[string]any{
					"field_id": "1",
				},
			},
		},
	}

	if err := survey.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
	if survey.ProviderRaw["activity_id"] != "123" {
		t.Fatalf("ProviderRaw was not preserved")
	}
	if survey.Questions[0].ProviderRaw["field_id"] != "1" {
		t.Fatalf("question ProviderRaw was not preserved")
	}
}

func TestSurveyDefinitionValidateRejectsMissingProvider(t *testing.T) {
	survey := SurveyDefinition{Title: "title"}

	err := survey.Validate()
	if err == nil || !strings.Contains(err.Error(), "provider") {
		t.Fatalf("Validate() error = %v, want provider error", err)
	}
}

func TestQuestionDefinitionValidateRejectsInvalidQuestion(t *testing.T) {
	tests := []struct {
		name string
		q    QuestionDefinition
		want string
	}{
		{
			name: "id",
			q: QuestionDefinition{
				Title: "title",
				Kind:  QuestionKindSingle,
			},
			want: "id is required",
		},
		{
			name: "title",
			q: QuestionDefinition{
				ID:   "q1",
				Kind: QuestionKindSingle,
			},
			want: "title is required",
		},
		{
			name: "kind",
			q: QuestionDefinition{
				ID:    "q1",
				Title: "title",
				Kind:  QuestionKind("other"),
			},
			want: "unsupported",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.q.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want %q", err, tt.want)
			}
		})
	}
}
