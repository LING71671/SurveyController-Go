package runner

import (
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/answer"
	"github.com/LING71671/SurveyController-Go/internal/config"
	"github.com/LING71671/SurveyController-Go/internal/engine"
)

type Plan struct {
	Mode             engine.Mode
	Provider         string
	URL              string
	Target           int
	Concurrency      int
	FailureThreshold int
	FailStopEnabled  bool
	Headless         bool
	SubmitInterval   config.DurationRange
	AnswerDuration   config.DurationRange
	TimedMode        config.TimedModeConfig
	Proxy            config.ProxyConfig
	ReverseFill      config.ReverseFillConfig
	RandomUA         config.RandomUAConfig
	Questions        []QuestionPlan
}

type QuestionPlan struct {
	ID            string
	Kind          string
	Required      bool
	Options       map[string]any
	Weights       []answer.OptionWeight
	MatrixWeights map[string][]answer.OptionWeight
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
		Mode:             cfg.Run.Mode,
		Provider:         strings.TrimSpace(cfg.Survey.Provider),
		URL:              strings.TrimSpace(cfg.Survey.URL),
		Target:           cfg.Run.Target,
		Concurrency:      cfg.Run.Concurrency,
		FailureThreshold: cfg.Run.FailureThreshold,
		FailStopEnabled:  cfg.Run.FailStopEnabled,
		Headless:         cfg.Run.Headless,
		SubmitInterval:   cfg.Run.SubmitInterval,
		AnswerDuration:   cfg.Run.AnswerDuration,
		TimedMode:        cfg.Run.TimedMode,
		Proxy:            cfg.Proxy,
		ReverseFill:      cfg.ReverseFill,
		RandomUA:         cloneRandomUAConfig(cfg.RandomUA),
		Questions:        make([]QuestionPlan, 0, len(cfg.Questions)),
	}
	for _, question := range cfg.Questions {
		questionID := strings.TrimSpace(question.ID)
		if questionID == "" {
			return Plan{}, fmt.Errorf("question id is required")
		}
		weights, err := config.QuestionOptionWeights(question)
		if err != nil {
			return Plan{}, fmt.Errorf("question %q option weights: %w", questionID, err)
		}
		matrixWeights, err := config.QuestionMatrixWeights(question)
		if err != nil {
			return Plan{}, fmt.Errorf("question %q matrix weights: %w", questionID, err)
		}
		plan.Questions = append(plan.Questions, QuestionPlan{
			ID:            questionID,
			Kind:          strings.TrimSpace(question.Kind),
			Required:      question.Required,
			Options:       cloneOptions(question.Options),
			Weights:       weights,
			MatrixWeights: matrixWeights,
		})
	}
	if err := New().ValidatePlan(plan); err != nil {
		return Plan{}, err
	}
	return plan, nil
}

func (r *Runner) ValidatePlan(plan Plan) error {
	mode, err := engine.ParseMode(plan.Mode.String())
	if err != nil {
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
	if err := engine.ValidateConcurrency(mode, plan.Concurrency); err != nil {
		return err
	}
	if plan.FailureThreshold < 0 {
		return fmt.Errorf("failure threshold must not be negative")
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

func cloneRandomUAConfig(cfg config.RandomUAConfig) config.RandomUAConfig {
	cloned := cfg
	if len(cfg.Keys) > 0 {
		cloned.Keys = append([]string(nil), cfg.Keys...)
	}
	if len(cfg.Ratios) > 0 {
		cloned.Ratios = make(map[string]int, len(cfg.Ratios))
		for key, value := range cfg.Ratios {
			cloned.Ratios[key] = value
		}
	}
	return cloned
}
