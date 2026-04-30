package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/config"
	"gopkg.in/yaml.v3"
)

func TestRunVersion(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"version"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(version) exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), "surveyctl v0.1.0") {
		t.Fatalf("version output = %q, want surveyctl version", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunHelpListsV02CommandStubs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"help"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(help) exit code = %d, want %d", code, exitOK)
	}
	for _, want := range []string{"config validate", "config generate", "doctor", "run", "version"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("help output = %q, want %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunConfigValidateFile(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
run:
  target: 1
  concurrency: 1
  mode: hybrid
questions: []
`)

	code := run([]string{"config", "validate", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(config validate) exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), path) {
		t.Fatalf("stdout = %q, want validated path", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunConfigGenerateFromFixtures(t *testing.T) {
	tests := []struct {
		name     string
		provider string
		fixture  string
		url      string
	}{
		{
			name:     "wjx",
			provider: "wjx",
			fixture:  filepath.Join("..", "..", "internal", "provider", "wjx", "testdata", "survey.html"),
			url:      "https://www.wjx.cn/vm/example.aspx",
		},
		{
			name:     "tencent",
			provider: "tencent",
			fixture:  filepath.Join("..", "..", "internal", "provider", "tencent", "testdata", "survey_api.json"),
			url:      "https://wj.qq.com/s2/example",
		},
		{
			name:     "credamo",
			provider: "credamo",
			fixture:  filepath.Join("..", "..", "internal", "provider", "credamo", "testdata", "snapshot.json"),
			url:      "https://www.credamo.com/s/example",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := run([]string{"config", "generate", "--provider", tt.provider, "--fixture", tt.fixture, "--url", tt.url}, &stdout, &stderr)
			if code != exitOK {
				t.Fatalf("run(config generate) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
			}
			var cfg config.RunConfig
			if err := yaml.Unmarshal(stdout.Bytes(), &cfg); err != nil {
				t.Fatalf("decode generated yaml: %v; output=%q", err, stdout.String())
			}
			if cfg.Survey.Provider != tt.provider || cfg.Survey.URL != tt.url {
				t.Fatalf("Survey = %+v, want provider %q url %q", cfg.Survey, tt.provider, tt.url)
			}
			if len(cfg.Questions) == 0 {
				t.Fatalf("generated config has no questions: %+v", cfg)
			}
			if err := cfg.Validate(); err != nil {
				t.Fatalf("generated config did not validate: %v", err)
			}
			if stderr.Len() != 0 {
				t.Fatalf("stderr = %q, want empty", stderr.String())
			}
		})
	}
}

func TestRunConfigGenerateRequiresProvider(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"config", "generate", "--fixture", "survey.html", "--url", "https://example.com"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(config generate missing provider) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "requires --provider") {
		t.Fatalf("stderr = %q, want provider requirement", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunConfigGenerateRejectsUnsupportedProvider(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"config", "generate", "--provider", "nope", "--fixture", "survey.html", "--url", "https://example.com"}, &stdout, &stderr)
	if code != exitFailure {
		t.Fatalf("run(config generate unsupported provider) exit code = %d, want %d", code, exitFailure)
	}
	if !strings.Contains(stderr.String(), "unsupported provider") {
		t.Fatalf("stderr = %q, want unsupported provider", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunDryRunPrintsPlanSummary(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 2
  concurrency: 2
  mode: browser
  failure_threshold: 3
  headless: false
questions:
  - id: q1
    kind: single
`)

	code := run([]string{"run", "--dry-run", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(dry-run) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	for _, want := range []string{"dry-run plan:", "provider: mock", "mode: browser", "target: 2", "questions: 1", "submissions: 0"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunDryRunJSONPrintsPlanSummary(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: hybrid
questions: []
`)

	code := run([]string{"run", "--dry-run", "--json", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(dry-run json) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	var summary map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
		t.Fatalf("json output decode failed: %v; output=%q", err, stdout.String())
	}
	if summary["provider"] != "mock" || summary["mode"] != "hybrid" || summary["question_count"] != float64(0) {
		t.Fatalf("summary = %+v, want provider/mode/question count", summary)
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunDryRunDetectsProviderFromURL(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://www.wjx.cn/vm/example.aspx"
run:
  target: 1
  concurrency: 1
  mode: hybrid
questions: []
`)

	code := run([]string{"run", "--dry-run", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(dry-run detect provider) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if !strings.Contains(stdout.String(), "provider: wjx") {
		t.Fatalf("stdout = %q, want detected wjx provider", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunDryRunKeepsExplicitProvider(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://www.wjx.cn/vm/example.aspx"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: hybrid
questions: []
`)

	code := run([]string{"run", "--dry-run", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(dry-run explicit provider) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if !strings.Contains(stdout.String(), "provider: mock") {
		t.Fatalf("stdout = %q, want explicit mock provider", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunRequiresDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "survey.yaml"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(without dry-run) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "requires --dry-run") {
		t.Fatalf("stderr = %q, want dry-run requirement", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunDryRunReportsInvalidConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
run:
  target: 1
  concurrency: 1
  mode: hybrid
questions: []
`)

	code := run([]string{"run", "--dry-run", path}, &stdout, &stderr)
	if code != exitFailure {
		t.Fatalf("run(dry-run invalid) exit code = %d, want %d", code, exitFailure)
	}
	if !strings.Contains(stderr.String(), "survey.url must match a built-in provider") {
		t.Fatalf("stderr = %q, want provider detection error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunDoctorPlaceholder(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"doctor"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(doctor) exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), "doctor checks: ok") {
		t.Fatalf("stdout = %q, want doctor ok", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunDoctorBrowser(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"doctor", "browser"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(doctor browser) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	for _, want := range []string{"browser doctor:", "operating_system", "proxy_connectivity"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunDoctorBrowserProbe(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"doctor", "browser", "--probe"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(doctor browser --probe) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	for _, want := range []string{"browser doctor:", "browser_launch_probe", "not configured"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunDoctorBrowserRejectsExtraArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"doctor", "browser", "extra"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(doctor browser extra) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "unknown doctor browser argument") {
		t.Fatalf("stderr = %q, want extra argument error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunConfigValidateRejectsExtraArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"config", "validate", "one.yaml", "two.yaml"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(config validate extra args) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "at most one path") {
		t.Fatalf("stderr = %q, want argument error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunConfigValidateReportsInvalidConfig(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: ""
run:
  target: 1
  concurrency: 1
  mode: hybrid
`)

	code := run([]string{"config", "validate", path}, &stdout, &stderr)
	if code != exitFailure {
		t.Fatalf("run(config validate invalid) exit code = %d, want %d", code, exitFailure)
	}
	if !strings.Contains(stderr.String(), "survey.url is required") {
		t.Fatalf("stderr = %q, want validation error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"nope"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(unknown) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "unknown command") {
		t.Fatalf("stderr = %q, want unknown command message", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "survey.yaml")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return path
}
