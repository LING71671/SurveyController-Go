package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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
	for _, want := range []string{"config validate", "doctor", "version"} {
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

func TestRunDoctorBrowserRejectsExtraArgs(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"doctor", "browser", "extra"}, &stdout, &stderr)
	if code != exitUsage {
		t.Fatalf("run(doctor browser extra) exit code = %d, want %d", code, exitUsage)
	}
	if !strings.Contains(stderr.String(), "accepts no arguments") {
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
