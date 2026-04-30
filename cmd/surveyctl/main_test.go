package main

import (
	"bytes"
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

func TestRunConfigValidatePlaceholder(t *testing.T) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	code := run([]string{"config", "validate", "example.yaml"}, &stdout, &stderr)
	if code != exitOK {
		t.Fatalf("run(config validate) exit code = %d, want %d", code, exitOK)
	}
	if !strings.Contains(stdout.String(), "example.yaml") {
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
	if !strings.Contains(stdout.String(), "doctor checks placeholder: ok") {
		t.Fatalf("stdout = %q, want doctor placeholder", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
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
