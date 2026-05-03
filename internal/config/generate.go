package config

import (
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/domain"
)

func FromSurveyDefinition(survey domain.SurveyDefinition) (RunConfig, error) {
	if err := survey.Validate(); err != nil {
		return RunConfig{}, fmt.Errorf("survey definition: %w", err)
	}

	cfg := DefaultRunConfig()
	cfg.Survey = SurveyConfig{
		URL:      strings.TrimSpace(survey.URL),
		Provider: survey.Provider.String(),
	}
	cfg.Questions = make([]QuestionConfig, 0, len(survey.Questions))
	for _, question := range survey.Questions {
		cfg.Questions = append(cfg.Questions, QuestionConfig{
			ID:       strings.TrimSpace(question.ID),
			Kind:     question.Kind.String(),
			Required: question.Required,
			Options:  defaultQuestionOptions(question),
		})
	}
	if err := cfg.Validate(); err != nil {
		return RunConfig{}, err
	}
	return cfg, nil
}

func defaultQuestionOptions(question domain.QuestionDefinition) map[string]any {
	options := map[string]any{}
	if len(question.Options) > 0 {
		options["weights"] = defaultOptionWeights(question.Options)
	}
	if isTextQuestion(question.Kind) {
		options["text"] = defaultTextAnswerSkeleton()
	}
	if len(question.Rows) > 0 && len(question.Options) > 0 {
		rows := make([]map[string]any, 0, len(question.Rows))
		for _, row := range question.Rows {
			rowID := strings.TrimSpace(row.ID)
			if rowID == "" {
				continue
			}
			rows = append(rows, map[string]any{
				"row_id":  rowID,
				"weights": defaultOptionWeights(question.Options),
			})
		}
		if len(rows) > 0 {
			options["matrix_weights"] = rows
		}
	}
	return options
}

func isTextQuestion(kind domain.QuestionKind) bool {
	return kind == domain.QuestionKindText || kind == domain.QuestionKindTextarea
}

func defaultTextAnswerSkeleton() map[string]any {
	return map[string]any{
		"mode":   "fixed",
		"values": []string{"sample answer"},
	}
}

func defaultOptionWeights(options []domain.OptionDefinition) []map[string]any {
	weights := make([]map[string]any, 0, len(options))
	for _, option := range options {
		optionID := strings.TrimSpace(option.ID)
		if optionID == "" {
			continue
		}
		weights = append(weights, map[string]any{
			"option_id": optionID,
			"weight":    1,
		})
	}
	return weights
}
