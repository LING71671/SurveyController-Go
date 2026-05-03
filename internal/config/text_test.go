package config

import (
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/answer"
)

func TestQuestionTextAnswerRuleParsesModes(t *testing.T) {
	tests := []struct {
		name string
		raw  map[string]any
		want answer.TextAnswerMode
	}{
		{
			name: "fixed",
			raw: map[string]any{
				"text": map[string]any{
					"mode":   "fixed",
					"values": []any{"alpha", "beta"},
				},
			},
			want: answer.TextAnswerModeFixed,
		},
		{
			name: "words",
			raw: map[string]any{
				"text": map[string]any{
					"mode":      "words",
					"words":     []any{"alpha", "beta"},
					"min_words": 2,
					"max_words": 3,
					"separator": "-",
				},
			},
			want: answer.TextAnswerModeWords,
		},
		{
			name: "digits",
			raw: map[string]any{
				"text": map[string]any{
					"mode":   "digits",
					"length": 6,
					"prefix": "42",
				},
			},
			want: answer.TextAnswerModeDigits,
		},
		{
			name: "phone",
			raw: map[string]any{
				"text": map[string]any{
					"mode":     "phone",
					"prefixes": []any{"177"},
				},
			},
			want: answer.TextAnswerModePhone,
		},
		{
			name: "template",
			raw: map[string]any{
				"text": map[string]any{
					"mode":     "template",
					"template": "from {city}",
					"slots": map[string]any{
						"city": []any{"shanghai"},
					},
				},
			},
			want: answer.TextAnswerModeTemplate,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok, err := QuestionTextAnswerRule(QuestionConfig{ID: "q1", Options: tt.raw})
			if err != nil {
				t.Fatalf("QuestionTextAnswerRule() returned error: %v", err)
			}
			if !ok {
				t.Fatal("QuestionTextAnswerRule() ok=false, want true")
			}
			if got.Mode != tt.want {
				t.Fatalf("Mode = %q, want %q", got.Mode, tt.want)
			}
		})
	}
}

func TestQuestionTextAnswerRuleReturnsFalseWhenMissing(t *testing.T) {
	got, ok, err := QuestionTextAnswerRule(QuestionConfig{ID: "q1"})
	if err != nil {
		t.Fatalf("QuestionTextAnswerRule() returned error: %v", err)
	}
	if ok || got.Mode != "" {
		t.Fatalf("QuestionTextAnswerRule() = (%+v, %v), want empty false", got, ok)
	}
}

func TestQuestionTextAnswerRuleRejectsInvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		options map[string]any
		want    string
	}{
		{
			name:    "object",
			options: map[string]any{"text": "bad"},
			want:    "object",
		},
		{
			name: "mode",
			options: map[string]any{
				"text": map[string]any{"mode": "magic"},
			},
			want: "unsupported",
		},
		{
			name: "words type",
			options: map[string]any{
				"text": map[string]any{
					"mode":  "words",
					"words": 1,
				},
			},
			want: "words",
		},
		{
			name: "length type",
			options: map[string]any{
				"text": map[string]any{
					"mode":   "digits",
					"length": "bad",
				},
			},
			want: "length",
		},
		{
			name: "template slot",
			options: map[string]any{
				"text": map[string]any{
					"mode":     "template",
					"template": "hello {name}",
				},
			},
			want: "slot",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := QuestionTextAnswerRule(QuestionConfig{ID: "q1", Options: tt.options})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("QuestionTextAnswerRule() error = %v, want %q", err, tt.want)
			}
		})
	}
}
