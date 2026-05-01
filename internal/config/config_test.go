package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/engine"
)

func TestDefaultRunConfigIsValidAfterSurveyURL(t *testing.T) {
	cfg := DefaultRunConfig()
	cfg.Survey.URL = "https://example.com/survey"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate() returned error: %v", err)
	}
	if cfg.SchemaVersion != CurrentSchemaVersion {
		t.Fatalf("SchemaVersion = %d, want %d", cfg.SchemaVersion, CurrentSchemaVersion)
	}
	if cfg.Run.Mode != engine.ModeHybrid {
		t.Fatalf("Run.Mode = %q, want %q", cfg.Run.Mode, engine.ModeHybrid)
	}
	if !cfg.Run.Headless || !cfg.Run.FailStopEnabled || cfg.Run.FailureThreshold != 1 {
		t.Fatalf("default runtime = %+v, want headless fail-stop threshold 1", cfg.Run)
	}
	if cfg.Proxy.Source != "default" || cfg.Proxy.OccupyMinutes != 1 {
		t.Fatalf("default proxy = %+v, want default source and occupy minute 1", cfg.Proxy)
	}
	if cfg.ReverseFill.Format != "auto" || cfg.ReverseFill.StartRow != 1 {
		t.Fatalf("default reverse fill = %+v, want auto start row 1", cfg.ReverseFill)
	}
}

func TestLoadRunConfigReadsYAML(t *testing.T) {
	path := writeRunConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 3
  concurrency: 2
  mode: browser
  failure_threshold: 5
  fail_stop_enabled: true
  headless: false
  submit_interval:
    min_seconds: 1
    max_seconds: 3
  answer_duration:
    min_seconds: 10
    max_seconds: 20
  timed_mode:
    enabled: true
    refresh_interval_seconds: 5
proxy:
  enabled: true
  source: custom
  custom_api: "https://proxy.example/api"
  area_code: "110000"
  occupy_minutes: 2
reverse_fill:
  enabled: true
  source_path: "samples.xlsx"
  format: wjx_sequence
  start_row: 2
random_ua:
  enabled: true
  keys: ["wechat", "pc"]
  ratios:
    wechat: 60
    pc: 40
questions: []
`)

	cfg, err := LoadRunConfig(path)
	if err != nil {
		t.Fatalf("LoadRunConfig() returned error: %v", err)
	}
	if cfg.Survey.Provider != "mock" {
		t.Fatalf("Survey.Provider = %q, want mock", cfg.Survey.Provider)
	}
	if cfg.Run.Target != 3 {
		t.Fatalf("Run.Target = %d, want 3", cfg.Run.Target)
	}
	if cfg.Run.Mode != engine.ModeBrowser {
		t.Fatalf("Run.Mode = %q, want %q", cfg.Run.Mode, engine.ModeBrowser)
	}
	if cfg.Run.FailureThreshold != 5 || cfg.Run.Headless {
		t.Fatalf("Run = %+v, want failure threshold 5 and headless false", cfg.Run)
	}
	if cfg.Run.SubmitInterval.MinSeconds != 1 || cfg.Run.AnswerDuration.MaxSeconds != 20 {
		t.Fatalf("Run duration ranges = %+v/%+v, want configured values", cfg.Run.SubmitInterval, cfg.Run.AnswerDuration)
	}
	if !cfg.Proxy.Enabled || cfg.Proxy.Source != "custom" || cfg.Proxy.OccupyMinutes != 2 {
		t.Fatalf("Proxy = %+v, want custom proxy settings", cfg.Proxy)
	}
	if !cfg.ReverseFill.Enabled || cfg.ReverseFill.Format != "wjx_sequence" || cfg.ReverseFill.StartRow != 2 {
		t.Fatalf("ReverseFill = %+v, want configured settings", cfg.ReverseFill)
	}
	if !cfg.RandomUA.Enabled || cfg.RandomUA.Ratios["wechat"] != 60 {
		t.Fatalf("RandomUA = %+v, want configured ratios", cfg.RandomUA)
	}
}

func TestLoadRunConfigAppliesDefaultsForOmittedRuntimeFields(t *testing.T) {
	path := writeRunConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: hybrid
questions: []
`)

	cfg, err := LoadRunConfig(path)
	if err != nil {
		t.Fatalf("LoadRunConfig() returned error: %v", err)
	}
	if !cfg.Run.Headless || !cfg.Run.FailStopEnabled || cfg.Run.FailureThreshold != 1 {
		t.Fatalf("Run defaults = %+v, want headless fail-stop threshold 1", cfg.Run)
	}
	if cfg.Proxy.Source != "default" || cfg.Proxy.OccupyMinutes != 1 {
		t.Fatalf("Proxy defaults = %+v, want default source occupy minute 1", cfg.Proxy)
	}
	if cfg.ReverseFill.Format != "auto" || cfg.ReverseFill.StartRow != 1 {
		t.Fatalf("ReverseFill defaults = %+v, want auto start row 1", cfg.ReverseFill)
	}
	if cfg.RandomUA.Ratios["pc"] != 34 {
		t.Fatalf("RandomUA defaults = %+v, want pc ratio", cfg.RandomUA)
	}
}

func TestRunConfigValidationRejectsBadSchemaVersion(t *testing.T) {
	cfg := DefaultRunConfig()
	cfg.SchemaVersion = 99
	cfg.Survey.URL = "https://example.com/survey"

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "unsupported schema_version") {
		t.Fatalf("Validate() error = %v, want schema version error", err)
	}
}

func TestRunConfigValidationRejectsMissingURL(t *testing.T) {
	cfg := DefaultRunConfig()

	err := cfg.Validate()
	if err == nil || !strings.Contains(err.Error(), "survey.url is required") {
		t.Fatalf("Validate() error = %v, want survey url error", err)
	}
}

func TestRuntimeConfigValidationRejectsInvalidValues(t *testing.T) {
	tests := []struct {
		name string
		cfg  RuntimeConfig
		want string
	}{
		{
			name: "target",
			cfg: RuntimeConfig{
				Target:      0,
				Concurrency: 1,
				Mode:        engine.ModeHybrid,
			},
			want: "run.target",
		},
		{
			name: "concurrency",
			cfg: RuntimeConfig{
				Target:      1,
				Concurrency: 0,
				Mode:        engine.ModeHybrid,
			},
			want: "run.concurrency",
		},
		{
			name: "mode",
			cfg: RuntimeConfig{
				Target:      1,
				Concurrency: 1,
				Mode:        engine.Mode("magic"),
			},
			want: "run.mode",
		},
		{
			name: "browser concurrency profile",
			cfg: RuntimeConfig{
				Target:      1,
				Concurrency: engine.BrowserWorkerConcurrencyLimit + 1,
				Mode:        engine.ModeBrowser,
			},
			want: "run.concurrency",
		},
		{
			name: "failure threshold",
			cfg: RuntimeConfig{
				Target:           1,
				Concurrency:      1,
				Mode:             engine.ModeHybrid,
				FailureThreshold: -1,
			},
			want: "run.failure_threshold",
		},
		{
			name: "submit interval",
			cfg: RuntimeConfig{
				Target:         1,
				Concurrency:    1,
				Mode:           engine.ModeHybrid,
				SubmitInterval: DurationRange{MinSeconds: 5, MaxSeconds: 1},
			},
			want: "run.submit_interval",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestRuntimeConfigValidationAllowsLightConcurrencyProfiles(t *testing.T) {
	for _, mode := range []engine.Mode{engine.ModeHTTP, engine.ModeHybrid} {
		t.Run(mode.String(), func(t *testing.T) {
			cfg := RuntimeConfig{
				Target:      1,
				Concurrency: engine.LightWorkerConcurrencyBaseline,
				Mode:        mode,
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("Validate() returned error: %v", err)
			}
		})
	}
}

func TestRunConfigValidationRejectsInvalidNestedRuntimeSettings(t *testing.T) {
	tests := []struct {
		name string
		edit func(*RunConfig)
		want string
	}{
		{
			name: "proxy custom api",
			edit: func(cfg *RunConfig) {
				cfg.Proxy.Enabled = true
				cfg.Proxy.Source = "custom"
			},
			want: "proxy: custom_api",
		},
		{
			name: "reverse fill source",
			edit: func(cfg *RunConfig) {
				cfg.ReverseFill.Enabled = true
			},
			want: "reverse_fill: source_path",
		},
		{
			name: "random ua ratio",
			edit: func(cfg *RunConfig) {
				cfg.RandomUA.Ratios = map[string]int{"pc": -1}
			},
			want: "random_ua: ratio",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := DefaultRunConfig()
			cfg.Survey.URL = "https://example.com/survey"
			tt.edit(&cfg)

			err := cfg.Validate()
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("Validate() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestValidateFileRejectsUnknownFields(t *testing.T) {
	path := writeRunConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
run:
  target: 1
  concurrency: 1
  mode: hybrid
unknown: true
`)

	err := ValidateFile(path)
	if err == nil || !strings.Contains(err.Error(), "field unknown not found") {
		t.Fatalf("ValidateFile() error = %v, want unknown field error", err)
	}
}

func writeRunConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "survey.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
