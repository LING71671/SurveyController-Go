package config

import (
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/domain"
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
			Options:  map[string]any{},
		})
	}
	if err := cfg.Validate(); err != nil {
		return RunConfig{}, err
	}
	return cfg, nil
}
