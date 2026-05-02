package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/domain"
	"github.com/LING71671/SurveyController-go/internal/logging"
	"github.com/LING71671/SurveyController-go/internal/provider/wjx"
	"github.com/LING71671/SurveyController-go/internal/runner"
)

type WJXHTTPRunOptions struct {
	Seed     int64
	Events   chan<- logging.RunEvent
	Executor wjx.HTTPSubmissionExecutor
	Survey   domain.SurveyDefinition
}

func RunWJXHTTPPlan(ctx context.Context, plan runner.Plan, options WJXHTTPRunOptions) (runner.RunPlanReport, error) {
	if strings.TrimSpace(plan.Provider) != domain.ProviderWJX.String() {
		return runner.RunPlanReport{}, fmt.Errorf("wjx http run requires wjx provider")
	}
	if options.Executor == nil {
		return runner.RunPlanReport{}, fmt.Errorf("wjx http submission executor is required")
	}
	if err := options.Survey.Validate(); err != nil {
		return runner.RunPlanReport{}, fmt.Errorf("wjx survey: %w", err)
	}

	pipeline, err := wjx.NewHTTPSubmissionPipeline(wjx.Provider{}, plan.Mode, options.Survey, options.Executor)
	if err != nil {
		return runner.RunPlanReport{}, err
	}
	return runPlanWithSubmitter(ctx, plan, options.Seed, pipeline, options.Events)
}
