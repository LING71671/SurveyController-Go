package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/logging"
	"github.com/LING71671/SurveyController-Go/internal/provider/wjx"
	"github.com/LING71671/SurveyController-Go/internal/runner"
)

type WJXHTTPRunOptions struct {
	Seed     int64
	Events   chan<- logging.RunEvent
	Executor wjx.HTTPSubmissionExecutor
	Survey   domain.SurveyDefinition
}

type WJXHTTPDryRunResult struct {
	Report runner.RunPlanReport       `json:"report"`
	Drafts []WJXHTTPSubmissionPreview `json:"drafts"`
}

func RunWJXHTTPDryRun(ctx context.Context, plan runner.Plan, options WJXHTTPRunOptions) (WJXHTTPDryRunResult, error) {
	if err := validateWJXHTTPPlanCompatibility(plan, options.Survey, "dry-run"); err != nil {
		return WJXHTTPDryRunResult{}, err
	}

	executor := &wjx.DryRunHTTPSubmissionExecutor{}
	report, err := RunWJXHTTPPlan(ctx, plan, WJXHTTPRunOptions{
		Seed:     options.Seed,
		Events:   options.Events,
		Executor: executor,
		Survey:   options.Survey,
	})
	if err != nil {
		return WJXHTTPDryRunResult{}, err
	}
	return WJXHTTPDryRunResult{
		Report: report,
		Drafts: previewsFromWJXHTTPDrafts(plan, executor.Drafts()),
	}, nil
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

func previewsFromWJXHTTPDrafts(plan runner.Plan, drafts []wjx.HTTPSubmissionDraft) []WJXHTTPSubmissionPreview {
	if len(drafts) == 0 {
		return nil
	}
	previews := make([]WJXHTTPSubmissionPreview, 0, len(drafts))
	for _, draft := range drafts {
		previews = append(previews, previewFromWJXHTTPDraft(plan, draft, wjxHTTPDraftAnswerCount(draft)))
	}
	return previews
}

func wjxHTTPDraftAnswerCount(draft wjx.HTTPSubmissionDraft) int {
	count := 0
	for key := range draft.Form {
		switch strings.TrimSpace(key) {
		case "", "curID", "submittype":
			continue
		default:
			count++
		}
	}
	return count
}
