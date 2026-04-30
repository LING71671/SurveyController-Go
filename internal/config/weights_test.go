package config

import (
	"encoding/json"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestQuestionOptionWeightsParsesYAMLWeights(t *testing.T) {
	cfg := loadConfigYAML(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: hybrid
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 2
        - option_id: b
          weight: 0.5
`)

	weights, err := QuestionOptionWeights(cfg.Questions[0])
	if err != nil {
		t.Fatalf("QuestionOptionWeights() returned error: %v", err)
	}
	if len(weights) != 2 || weights[0].OptionID != "a" || weights[0].Weight != 2 || weights[1].Weight != 0.5 {
		t.Fatalf("weights = %+v, want parsed YAML weights", weights)
	}
}

func TestQuestionOptionWeightsParsesJSONNumberWeights(t *testing.T) {
	var question QuestionConfig
	decoder := json.NewDecoder(strings.NewReader(`{
		"id": "q1",
		"options": {
			"weights": [
				{"option_id": "a", "weight": 3}
			]
		}
	}`))
	decoder.UseNumber()
	if err := decoder.Decode(&question); err != nil {
		t.Fatalf("decode question json: %v", err)
	}

	weights, err := QuestionOptionWeights(question)
	if err != nil {
		t.Fatalf("QuestionOptionWeights() returned error: %v", err)
	}
	if len(weights) != 1 || weights[0].OptionID != "a" || weights[0].Weight != 3 {
		t.Fatalf("weights = %+v, want parsed JSON number weight", weights)
	}
}

func TestQuestionOptionWeightsMissingReturnsNil(t *testing.T) {
	weights, err := QuestionOptionWeights(QuestionConfig{ID: "q1", Options: map[string]any{}})
	if err != nil {
		t.Fatalf("QuestionOptionWeights() returned error: %v", err)
	}
	if weights != nil {
		t.Fatalf("weights = %+v, want nil", weights)
	}
}

func TestQuestionOptionWeightsRejectsInvalidWeights(t *testing.T) {
	tests := []struct {
		name     string
		question QuestionConfig
		want     string
	}{
		{
			name:     "empty",
			question: QuestionConfig{Options: map[string]any{"weights": []any{}}},
			want:     "must not be empty",
		},
		{
			name: "missing option id",
			question: QuestionConfig{Options: map[string]any{"weights": []any{
				map[string]any{"weight": 1},
			}}},
			want: "option_id",
		},
		{
			name: "negative",
			question: QuestionConfig{Options: map[string]any{"weights": []any{
				map[string]any{"option_id": "a", "weight": -1},
			}}},
			want: "negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := QuestionOptionWeights(tt.question)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("QuestionOptionWeights() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestQuestionMatrixWeightsParsesRows(t *testing.T) {
	cfg := loadConfigYAML(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: hybrid
questions:
  - id: m1
    kind: matrix
    options:
      matrix_weights:
        - row_id: row1
          weights:
            - option_id: a
              weight: 1
            - option_id: b
              weight: 2
`)

	matrix, err := QuestionMatrixWeights(cfg.Questions[0])
	if err != nil {
		t.Fatalf("QuestionMatrixWeights() returned error: %v", err)
	}
	rowWeights := matrix["row1"]
	if len(rowWeights) != 2 || rowWeights[1].OptionID != "b" || rowWeights[1].Weight != 2 {
		t.Fatalf("matrix weights = %+v, want row option weights", matrix)
	}
}

func loadConfigYAML(t *testing.T, body string) RunConfig {
	t.Helper()
	cfg := DefaultRunConfig()
	if err := yaml.Unmarshal([]byte(body), &cfg); err != nil {
		t.Fatalf("decode config yaml: %v", err)
	}
	return cfg
}
