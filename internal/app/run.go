package app

import (
	"context"
	"fmt"
	"math/rand"
	"runtime"
	"strings"
	"time"

	"github.com/LING71671/SurveyController-go/internal/config"
	"github.com/LING71671/SurveyController-go/internal/logging"
	"github.com/LING71671/SurveyController-go/internal/provider/builtin"
	"github.com/LING71671/SurveyController-go/internal/runner"
)

type RunPlanOverrides struct {
	Target      int
	Concurrency int
}

type MockRunOptions struct {
	Seed      int64
	FailEvery int
	Events    chan<- logging.RunEvent
}

func CompileRunPlanFromFile(path string, overrides RunPlanOverrides) (runner.Plan, error) {
	cfg, err := config.LoadRunConfig(path)
	if err != nil {
		return runner.Plan{}, err
	}
	if strings.TrimSpace(cfg.Survey.Provider) == "" {
		providerID, ok := builtin.DetectProvider(cfg.Survey.URL)
		if !ok {
			return runner.Plan{}, fmt.Errorf("provider is required or survey.url must match a built-in provider: %s", cfg.Survey.URL)
		}
		cfg.Survey.Provider = providerID.String()
	}
	plan, err := runner.CompilePlan(cfg)
	if err != nil {
		return runner.Plan{}, err
	}
	return ApplyRunPlanOverrides(plan, overrides)
}

func ApplyRunPlanOverrides(plan runner.Plan, overrides RunPlanOverrides) (runner.Plan, error) {
	if overrides.Target > 0 {
		plan.Target = overrides.Target
	}
	if overrides.Concurrency > 0 {
		plan.Concurrency = overrides.Concurrency
	}
	if err := runner.New().ValidatePlan(plan); err != nil {
		return runner.Plan{}, err
	}
	return plan, nil
}

func RunMockPlan(ctx context.Context, plan runner.Plan, options MockRunOptions) (runner.RunPlanReport, error) {
	return runPlanWithSubmitter(ctx, plan, options.Seed, mockSubmitter(options.FailEvery), options.Events)
}

func mockSubmitter(failEvery int) runner.AnswerPlanSubmitter {
	if failEvery > 0 {
		return &runner.FailureInjectingMockSubmitter{FailEvery: failEvery}
	}
	return runner.MockAnswerPlanSubmitter{}
}

func runPlanWithSubmitter(ctx context.Context, plan runner.Plan, seed int64, submitter runner.AnswerPlanSubmitter, events chan<- logging.RunEvent) (runner.RunPlanReport, error) {
	if ctx == nil {
		return runner.RunPlanReport{}, fmt.Errorf("context is required")
	}
	if submitter == nil {
		return runner.RunPlanReport{}, fmt.Errorf("answer plan submitter is required")
	}
	if seed == 0 {
		seed = 1
	}

	beforeRuntime := sampleRuntime()
	startedAt := time.Now()
	snapshot, err := runner.RunPlanSubmissions(ctx, plan, runner.RunPlanOptions{
		RNG:       rand.New(rand.NewSource(seed)),
		Submitter: submitter,
		Events:    events,
	})
	elapsed := time.Since(startedAt)
	afterRuntime := sampleRuntime()
	if err != nil {
		return runner.RunPlanReport{}, err
	}
	return runner.NewTimedRunPlanReport(plan, snapshot, elapsed).WithResourceMetrics(runtimeMetrics(beforeRuntime, afterRuntime)), nil
}

type runtimeSample struct {
	heapAlloc  uint64
	totalAlloc uint64
	goroutines int
}

func sampleRuntime() runtimeSample {
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return runtimeSample{
		heapAlloc:  stats.HeapAlloc,
		totalAlloc: stats.TotalAlloc,
		goroutines: runtime.NumGoroutine(),
	}
}

func runtimeMetrics(before runtimeSample, after runtimeSample) runner.RunResourceMetrics {
	return runner.RunResourceMetrics{
		Goroutines:      after.goroutines,
		HeapAllocBytes:  after.heapAlloc,
		HeapAllocDelta:  int64(after.heapAlloc) - int64(before.heapAlloc),
		TotalAllocDelta: subtractUint64(after.totalAlloc, before.totalAlloc),
	}
}

func subtractUint64(after uint64, before uint64) uint64 {
	if after < before {
		return 0
	}
	return after - before
}
