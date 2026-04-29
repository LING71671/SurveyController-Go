package runner

import (
	"fmt"
	"strings"

	"github.com/hungryM0/SurveyController-go/internal/engine"
)

type Plan struct {
	Engine   engine.Mode
	Provider string
	URL      string
}

type Runner struct{}

func New() *Runner {
	return &Runner{}
}

func (r *Runner) ValidatePlan(plan Plan) error {
	if _, err := engine.ParseMode(plan.Engine.String()); err != nil {
		return err
	}
	if strings.TrimSpace(plan.Provider) == "" {
		return fmt.Errorf("provider is required")
	}
	if strings.TrimSpace(plan.URL) == "" {
		return fmt.Errorf("url is required")
	}
	return nil
}
