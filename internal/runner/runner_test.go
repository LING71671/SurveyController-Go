package runner

import (
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/config"
	"github.com/LING71671/SurveyController-Go/internal/engine"
)

func TestCompilePlan(t *testing.T) {
	cfg := config.DefaultRunConfig()
	cfg.Survey.URL = " https://example.com/survey "
	cfg.Survey.Provider = " mock "
	cfg.Run.Target = 3
	cfg.Run.Concurrency = 2
	cfg.Run.Mode = engine.ModeBrowser
	cfg.Run.FailureThreshold = 4
	cfg.Run.FailStopEnabled = true
	cfg.Run.Headless = false
	cfg.Run.SubmitInterval = config.DurationRange{MinSeconds: 1, MaxSeconds: 2}
	cfg.Run.AnswerDuration = config.DurationRange{MinSeconds: 10, MaxSeconds: 20}
	cfg.Run.TimedMode = config.TimedModeConfig{Enabled: true, RefreshIntervalSeconds: 5}
	cfg.Proxy = config.ProxyConfig{Enabled: true, Source: "custom", CustomAPI: "https://proxy.example/api", OccupyMinutes: 2}
	cfg.ReverseFill = config.ReverseFillConfig{Enabled: true, SourcePath: "samples.xlsx", Format: "wjx_text", StartRow: 2}
	cfg.RandomUA = config.RandomUAConfig{
		Enabled: true,
		Keys:    []string{"wechat", "pc"},
		Ratios:  map[string]int{"wechat": 60, "pc": 40},
	}
	cfg.Questions = []config.QuestionConfig{
		{
			ID:       " q1 ",
			Kind:     "single",
			Required: true,
			Options: map[string]any{
				"weight": 1,
			},
		},
	}

	plan, err := CompilePlan(cfg)
	if err != nil {
		t.Fatalf("CompilePlan() returned error: %v", err)
	}
	if plan.URL != "https://example.com/survey" {
		t.Fatalf("URL = %q, want trimmed url", plan.URL)
	}
	if plan.Provider != "mock" {
		t.Fatalf("Provider = %q, want mock", plan.Provider)
	}
	if plan.Mode != engine.ModeBrowser {
		t.Fatalf("Mode = %q, want browser", plan.Mode)
	}
	if plan.Target != 3 || plan.Concurrency != 2 {
		t.Fatalf("target/concurrency = %d/%d, want 3/2", plan.Target, plan.Concurrency)
	}
	if plan.FailureThreshold != 4 || !plan.FailStopEnabled || plan.Headless {
		t.Fatalf("runtime flags = %+v, want failure threshold 4 fail-stop true headless false", plan)
	}
	if plan.SubmitInterval.MinSeconds != 1 || plan.AnswerDuration.MaxSeconds != 20 || !plan.TimedMode.Enabled {
		t.Fatalf("runtime timing = %+v/%+v/%+v, want configured timing", plan.SubmitInterval, plan.AnswerDuration, plan.TimedMode)
	}
	if !plan.Proxy.Enabled || plan.Proxy.Source != "custom" || plan.Proxy.OccupyMinutes != 2 {
		t.Fatalf("Proxy = %+v, want custom proxy", plan.Proxy)
	}
	if !plan.ReverseFill.Enabled || plan.ReverseFill.Format != "wjx_text" {
		t.Fatalf("ReverseFill = %+v, want configured reverse fill", plan.ReverseFill)
	}
	if !plan.RandomUA.Enabled || plan.RandomUA.Ratios["wechat"] != 60 {
		t.Fatalf("RandomUA = %+v, want configured random ua", plan.RandomUA)
	}
	if len(plan.Questions) != 1 || plan.Questions[0].ID != "q1" {
		t.Fatalf("Questions = %+v, want q1", plan.Questions)
	}
}

func TestCompilePlanClonesQuestionOptions(t *testing.T) {
	cfg := config.DefaultRunConfig()
	cfg.Survey.URL = "https://example.com/survey"
	cfg.Survey.Provider = "mock"
	cfg.Questions = []config.QuestionConfig{
		{
			ID: "q1",
			Options: map[string]any{
				"weight": 1,
			},
		},
	}

	plan, err := CompilePlan(cfg)
	if err != nil {
		t.Fatalf("CompilePlan() returned error: %v", err)
	}
	cfg.Questions[0].Options["weight"] = 2
	if plan.Questions[0].Options["weight"] != 1 {
		t.Fatalf("compiled options changed after source mutation")
	}
}

func TestCompilePlanParsesQuestionWeights(t *testing.T) {
	cfg := config.DefaultRunConfig()
	cfg.Survey.URL = "https://example.com/survey"
	cfg.Survey.Provider = "mock"
	cfg.Questions = []config.QuestionConfig{
		{
			ID:   "q1",
			Kind: "single",
			Options: map[string]any{
				"weights": []any{
					map[string]any{"option_id": "a", "weight": 2},
					map[string]any{"option_id": "b", "weight": 1},
				},
			},
		},
		{
			ID:   "m1",
			Kind: "matrix",
			Options: map[string]any{
				"matrix_weights": []any{
					map[string]any{
						"row_id": "row1",
						"weights": []any{
							map[string]any{"option_id": "x", "weight": 1},
						},
					},
				},
			},
		},
	}

	plan, err := CompilePlan(cfg)
	if err != nil {
		t.Fatalf("CompilePlan() returned error: %v", err)
	}
	if len(plan.Questions[0].Weights) != 2 || plan.Questions[0].Weights[0].OptionID != "a" || plan.Questions[0].Weights[0].Weight != 2 {
		t.Fatalf("Weights = %+v, want parsed option weights", plan.Questions[0].Weights)
	}
	rowWeights := plan.Questions[1].MatrixWeights["row1"]
	if len(rowWeights) != 1 || rowWeights[0].OptionID != "x" || rowWeights[0].Weight != 1 {
		t.Fatalf("MatrixWeights = %+v, want parsed row weights", plan.Questions[1].MatrixWeights)
	}
}

func TestCompilePlanRejectsInvalidQuestionWeights(t *testing.T) {
	cfg := config.DefaultRunConfig()
	cfg.Survey.URL = "https://example.com/survey"
	cfg.Survey.Provider = "mock"
	cfg.Questions = []config.QuestionConfig{
		{
			ID: "q1",
			Options: map[string]any{
				"weights": []any{
					map[string]any{"option_id": "a", "weight": -1},
				},
			},
		},
	}

	_, err := CompilePlan(cfg)
	if err == nil || !strings.Contains(err.Error(), `question "q1" option weights`) {
		t.Fatalf("CompilePlan() error = %v, want question weight context", err)
	}
}

func TestCompilePlanClonesRandomUAConfig(t *testing.T) {
	cfg := config.DefaultRunConfig()
	cfg.Survey.URL = "https://example.com/survey"
	cfg.Survey.Provider = "mock"
	cfg.RandomUA = config.RandomUAConfig{
		Enabled: true,
		Keys:    []string{"wechat"},
		Ratios:  map[string]int{"wechat": 100},
	}

	plan, err := CompilePlan(cfg)
	if err != nil {
		t.Fatalf("CompilePlan() returned error: %v", err)
	}
	cfg.RandomUA.Keys[0] = "pc"
	cfg.RandomUA.Ratios["wechat"] = 1
	if plan.RandomUA.Keys[0] != "wechat" || plan.RandomUA.Ratios["wechat"] != 100 {
		t.Fatalf("compiled random ua changed after source mutation: %+v", plan.RandomUA)
	}
}

func TestCompilePlanRejectsInvalidQuestion(t *testing.T) {
	cfg := config.DefaultRunConfig()
	cfg.Survey.URL = "https://example.com/survey"
	cfg.Survey.Provider = "mock"
	cfg.Questions = []config.QuestionConfig{{ID: " "}}

	_, err := CompilePlan(cfg)
	if err == nil || !strings.Contains(err.Error(), "question id") {
		t.Fatalf("CompilePlan() error = %v, want question id error", err)
	}
}

func TestValidatePlanRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		plan Plan
		want string
	}{
		{name: "provider", plan: Plan{Mode: engine.ModeHybrid, URL: "https://example.com", Target: 1, Concurrency: 1}, want: "provider"},
		{name: "url", plan: Plan{Mode: engine.ModeHybrid, Provider: "mock", Target: 1, Concurrency: 1}, want: "url"},
		{name: "target", plan: Plan{Mode: engine.ModeHybrid, Provider: "mock", URL: "https://example.com", Concurrency: 1}, want: "target"},
		{name: "concurrency", plan: Plan{Mode: engine.ModeHybrid, Provider: "mock", URL: "https://example.com", Target: 1}, want: "concurrency"},
		{name: "max concurrency", plan: Plan{Mode: engine.ModeHybrid, Provider: "mock", URL: "https://example.com", Target: 1, Concurrency: DefaultMaxWorkerConcurrency + 1}, want: "concurrency"},
		{name: "browser concurrency profile", plan: Plan{Mode: engine.ModeBrowser, Provider: "mock", URL: "https://example.com", Target: 1, Concurrency: engine.BrowserWorkerConcurrencyLimit + 1}, want: "browser mode"},
		{name: "failure threshold", plan: Plan{Mode: engine.ModeHybrid, Provider: "mock", URL: "https://example.com", Target: 1, Concurrency: 1, FailureThreshold: -1}, want: "failure threshold"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New().ValidatePlan(tt.plan)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ValidatePlan() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestValidatePlanAllowsLightConcurrencyProfiles(t *testing.T) {
	for _, mode := range []engine.Mode{engine.ModeHTTP, engine.ModeHybrid} {
		t.Run(mode.String(), func(t *testing.T) {
			plan := Plan{
				Mode:        mode,
				Provider:    "mock",
				URL:         "https://example.com",
				Target:      1,
				Concurrency: DefaultMaxWorkerConcurrency,
			}
			if err := New().ValidatePlan(plan); err != nil {
				t.Fatalf("ValidatePlan() returned error: %v", err)
			}
		})
	}
}
