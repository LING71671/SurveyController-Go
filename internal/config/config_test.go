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
