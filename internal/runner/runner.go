package runner

import (
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/config"
	"github.com/LING71671/SurveyController-go/internal/engine"
)

type Plan struct {
	Mode        engine.Mode
	Provider    string
	URL         string
	Target      int
	Concurrency int
	Questions   []QuestionPlan
}

type QuestionPlan struct {
	ID       string
	Kind     string
	Required bool
	Options  map[string]any
}

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func CompilePlan(cfg config.RunConfig) (Plan, error) {
	if err := cfg.Validate(); err != nil {
		return Plan{}, err
	}

	plan := Plan{
		Mode:        cfg.Run.Mode,
		Provider:    strings.TrimSpace(cfg.Survey.Provider),
		URL:         strings.TrimSpace(cfg.Survey.URL),
		Target:      cfg.Run.Target,
		Concurrency: cfg.Run.Concurrency,
		Questions:   make([]QuestionPlan, 0, len(cfg.Questions)),
	}
	for _, question := range cfg.Questions {
		if strings.TrimSpace(question.ID) == "" {
			return Plan{}, fmt.Errorf("question id is required")
		}
		plan.Questions = append(plan.Questions, QuestionPlan{
			ID:       strings.TrimSpace(question.ID),
			Kind:     strings.TrimSpace(question.Kind),
			Required: question.Required,
			Options:  cloneOptions(question.Options),
		})
	}
	if err := New().ValidatePlan(plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (r *Runner) ValidatePlan(plan Plan) error {
	if _, err := engine.ParseMode(plan.Mode.String()); err != nil {
		return err
	}
	if strings.TrimSpace(plan.Provider) == "" {
		return fmt.Errorf("provider is required")
	}
	if strings.TrimSpace(plan.URL) == "" {
		return fmt.Errorf("url is required")
	}
	if plan.Target <= 0 {
		return fmt.Errorf("target must be greater than 0")
	}
	if plan.Concurrency <= 0 {
		return fmt.Errorf("concurrency must be greater than 0")
	}
	return nil
}

func cloneOptions(options map[string]any) map[string]any {
	if len(options) == 0 {
		return nil
	}
	cloned := make(map[string]any, len(options))
	for key, value := range options {
		cloned[key] = value
	}
	return cloned
}
