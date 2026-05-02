package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/config"
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

func TestRunDryRunAppliesTargetConcurrencyOverrides(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: http
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	code := run([]string{"run", "--dry-run", "--target", "7", "--concurrency", "3", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(dry-run overrides) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	for _, want := range []string{"target: 7", "concurrency: 3"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunMockRunPrintsSummary(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 3
  concurrency: 2
  mode: http
  failure_threshold: 1
  fail_stop_enabled: true
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
        - option_id: b
          weight: 1
`)

	code := run([]string{"run", "--mock", "--seed", "7", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(mock) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	for _, want := range []string{"mock run:", "provider: mock", "target: 3", "successes: 3", "failures: 0", "completed: 3", "completion_rate: 100.00%", "success_rate: 100.00%", "duration_ms:", "throughput_per_second:", "goroutines:", "heap_alloc_bytes:", "total_alloc_delta_bytes:", "failure_threshold_reached: false", "network: disabled"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunMockRunAppliesTargetConcurrencyOverrides(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: http
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	code := run([]string{"run", "--mock", "--target", "4", "--concurrency", "2", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(mock overrides) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	for _, want := range []string{"target: 4", "concurrency: 2", "successes: 4", "workers: 2"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunMockRunJSONPrintsSummary(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 2
  concurrency: 2
  mode: http
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	code := run([]string{"run", "--mock", "--json", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(mock json) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	var summary map[string]any
	if err := json.Unmarshal(stdout.Bytes(), &summary); err != nil {
		t.Fatalf("json output decode failed: %v; output=%q", err, stdout.String())
	}
	if summary["successes"] != float64(2) || summary["failures"] != float64(0) || summary["completed"] != float64(2) || summary["completion_rate"] != float64(1) || summary["success_rate"] != float64(1) || summary["seed"] != float64(1) {
		t.Fatalf("summary = %+v, want mock success summary", summary)
	}
	if _, ok := summary["duration_ms"]; !ok {
		t.Fatalf("summary = %+v, want duration_ms", summary)
	}
	if _, ok := summary["throughput_per_second"]; !ok {
		t.Fatalf("summary = %+v, want throughput_per_second", summary)
	}
	if summary["failure_threshold_reached"] != false {
		t.Fatalf("summary = %+v, want failure threshold false", summary)
	}
	for _, key := range []string{"goroutines", "heap_alloc_bytes", "heap_alloc_delta_bytes", "total_alloc_delta_bytes"} {
		if _, ok := summary[key]; !ok {
			t.Fatalf("summary = %+v, want %s", summary, key)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunMockRunBudgetPasses(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 2
  concurrency: 1
  mode: http
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	code := run([]string{"run", "--mock", "--min-throughput", "0", "--max-heap-delta", "999999999", "--max-goroutines", "1000", "--expect-failure-threshold", "false", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(mock budget pass) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if !strings.Contains(stdout.String(), "mock run:") {
		t.Fatalf("stdout = %q, want mock summary", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunMockRunBudgetFailureStillPrintsSummary(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: http
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	code := run([]string{"run", "--mock", "--min-throughput", "999999999", path}, &stdout, &stderr)
	if code != exitFailure {
		t.Fatalf("run(mock budget failure) exit code = %d, want %d", code, exitFailure)
	}
	if !strings.Contains(stdout.String(), "mock run:") {
		t.Fatalf("stdout = %q, want diagnostic summary", stdout.String())
	}
	for _, want := range []string{"run mock budget failed", "throughput_per_second"} {
		if !strings.Contains(stderr.String(), want) {
			t.Fatalf("stderr = %q, want %q", stderr.String(), want)
		}
	}
}

func TestRunMockRunBudgetChecksFailureThreshold(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 5
  concurrency: 1
  mode: http
  failure_threshold: 1
  fail_stop_enabled: true
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	code := run([]string{"run", "--mock", "--mock-fail-every", "2", "--expect-failure-threshold", "true", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(mock failure threshold budget) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	if !strings.Contains(stdout.String(), "failure_threshold_reached: true") {
		t.Fatalf("stdout = %q, want threshold reached", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunRejectsInvalidBudgetFlags(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want string
	}{
		{name: "throughput", args: []string{"run", "--mock", "--min-throughput", "-1"}, want: "--min-throughput requires a non-negative number"},
		{name: "heap", args: []string{"run", "--mock", "--max-heap-delta", "0"}, want: "--max-heap-delta requires a positive integer"},
		{name: "bool", args: []string{"run", "--mock", "--expect-failure-threshold", "yes"}, want: "--expect-failure-threshold requires true or false"},
		{name: "dry-run", args: []string{"run", "--dry-run", "--min-throughput", "1"}, want: "budget flags require --mock"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdout bytes.Buffer
			var stderr bytes.Buffer

			code := run(tt.args, &stdout, &stderr)
			if code != exitUsage {
				t.Fatalf("run(invalid budget) exit code = %d, want %d", code, exitUsage)
			}
			if !strings.Contains(stderr.String(), tt.want) {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.want)
			}
			if stdout.Len() != 0 {
				t.Fatalf("stdout = %q, want empty", stdout.String())
			}
		})
	}
}

func TestRunRejectsInvalidTargetOverride(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "--mock", "--target", "0"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(invalid target override) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "--target requires a positive integer") {
		t.Fatalf("stderr = %q, want target override error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsInvalidConcurrencyOverride(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "--mock", "--concurrency", "nope"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(invalid concurrency override) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "--concurrency requires a positive integer") {
		t.Fatalf("stderr = %q, want concurrency override error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsConcurrencyOverrideAboveModeLimit(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: browser
questions: []
`)

	code := run([]string{"run", "--dry-run", "--concurrency", "17", path}, &stdout, &stderr)
	if code != exitFailure {
		t.Fatalf("run(too high concurrency override) exit code = %d, want %d", code, exitFailure)
	}
	if !strings.Contains(stderr.String(), "browser mode concurrency") {
		t.Fatalf("stderr = %q, want browser concurrency limit", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunMockRunFailureInjection(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 5
  concurrency: 1
  mode: http
  failure_threshold: 1
  fail_stop_enabled: true
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	code := run([]string{"run", "--mock", "--mock-fail-every", "2", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(mock fail every) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	for _, want := range []string{"successes: 1", "failures: 1", "failure_threshold_reached: true"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunMockRejectsInvalidFailEvery(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "--mock", "--mock-fail-every", "0"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(invalid fail every) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "--mock-fail-every requires a positive integer") {
		t.Fatalf("stderr = %q, want fail every error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunDryRunRejectsFailEvery(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "--dry-run", "--mock-fail-every", "2"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(dry-run fail every) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "requires --mock") {
		t.Fatalf("stderr = %q, want mock-only error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunMockRunEventsTextPrintsEventStream(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: http
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	code := run([]string{"run", "--mock", "--events", "text", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(mock events text) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	for _, want := range []string{"run_started", "worker_started", "submission_success", "run_finished", "events: 4", "mock run:"} {
		if !strings.Contains(stdout.String(), want) {
			t.Fatalf("stdout = %q, want %q", stdout.String(), want)
		}
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunMockRunEventsJSONLPrintsEventStream(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	path := writeConfig(t, `schema_version: 1
survey:
  url: "https://example.com/survey"
  provider: "mock"
run:
  target: 1
  concurrency: 1
  mode: http
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
`)

	code := run([]string{"run", "--mock", "--events", "jsonl", path}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(mock events jsonl) exit code = %d, want %d; stderr=%q", code, exitOK, stderr.String())
	}
	lines := strings.Split(strings.TrimSpace(stdout.String()), "\n")
	if len(lines) < 6 {
		t.Fatalf("stdout lines = %q, want event stream and summary", lines)
	}
	var firstEvent map[string]any
	if err := json.Unmarshal([]byte(lines[0]), &firstEvent); err != nil {
		t.Fatalf("first jsonl event decode failed: %v; line=%q", err, lines[0])
	}
	if firstEvent["type"] != "run_started" || firstEvent["level"] != "info" {
		t.Fatalf("firstEvent = %+v, want run_started info", firstEvent)
	}
	if !strings.Contains(stdout.String(), "events: 4") || !strings.Contains(stdout.String(), "mock run:") {
		t.Fatalf("stdout = %q, want event count and summary", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunMockRejectsEventsWithJSONSummary(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "--mock", "--json", "--events", "jsonl"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(mock json events) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "cannot be combined") {
		t.Fatalf("stderr = %q, want json/events conflict", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunDryRunRejectsEvents(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "--dry-run", "--events", "text"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(dry-run events) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "requires --mock") {
		t.Fatalf("stderr = %q, want dry-run events error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunMockRejectsInvalidEventFormat(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "--mock", "--events", "xml"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(invalid events) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "--events must be text or jsonl") {
		t.Fatalf("stderr = %q, want event format error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRejectsBothDryRunAndMock(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "--dry-run", "--mock"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(conflicting modes) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "only one") {
		t.Fatalf("stderr = %q, want conflict error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunMockRejectsInvalidSeed(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "--mock", "--seed", "nope"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(invalid seed) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "--seed requires an integer") {
		t.Fatalf("stderr = %q, want seed error", stderr.String())
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout = %q, want empty", stdout.String())
	}
}

func TestRunRequiresDryRun(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"run", "survey.yaml"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(without dry-run) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "requires --dry-run or --mock") {
		t.Fatalf("stderr = %q, want run mode requirement", stderr.String())
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
