package runner

import (
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/config"
	"github.com/LING71671/SurveyController-go/internal/engine"
)

func TestCompilePlan(t *testing.T) {
	cfg := config.DefaultRunConfig()
	cfg.Survey.URL = " https://example.com/survey "
	cfg.Survey.Provider = " mock "
	cfg.Run.Target = 3
	cfg.Run.Concurrency = 2
	cfg.Run.Mode = engine.ModeBrowser
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
