package doctor

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCheckBrowserReportsSupportedEnvironment(t *testing.T) {
	report := CheckBrowser(context.Background(), BrowserOptions{
		GOOS: "windows",
		Env:  map[string]string{"PLAYWRIGHT_BROWSERS_PATH": "C:\\pw"},
		LookPath: func(file string) (string, error) {
			if file == "msedge" {
				return `C:\Program Files\Microsoft\Edge\Application\msedge.exe`, nil
			}
			return "", errors.New("not found")
		},
	})

	if !report.OK() {
		t.Fatalf("report.OK() = false, want true: %+v", report.Checks)
	}
	assertCheck(t, report, "operating_system", StatusOK, "windows")
	assertCheck(t, report, "playwright_browsers_path", StatusOK, "set")
	assertCheck(t, report, "system_browser", StatusOK, "msedge")
}

func TestCheckBrowserWarnsForOptionalMissingPieces(t *testing.T) {
	report := CheckBrowser(context.Background(), BrowserOptions{
		GOOS: "linux",
		Env:  map[string]string{},
		LookPath: func(file string) (string, error) {
			return "", errors.New("not found")
		},
	})

	if !report.OK() {
		t.Fatalf("report.OK() = false, want true for warnings: %+v", report.Checks)
	}
	assertCheck(t, report, "playwright_browsers_path", StatusWarn, "not set")
	assertCheck(t, report, "system_browser", StatusWarn, "no common system browser")
	assertCheck(t, report, "proxy_connectivity", StatusWarn, "not configured")
}

func TestCheckBrowserFailsUnsupportedOS(t *testing.T) {
	report := CheckBrowser(context.Background(), BrowserOptions{
		GOOS: "plan9",
		Env:  map[string]string{},
		LookPath: func(file string) (string, error) {
			return "", errors.New("not found")
		},
	})

	if report.OK() {
		t.Fatalf("report.OK() = true, want false")
	}
	assertCheck(t, report, "operating_system", StatusFail, "not supported")
}

func TestCheckBrowserFailsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	report := CheckBrowser(ctx, BrowserOptions{
		GOOS:     "linux",
		Env:      map[string]string{},
		LookPath: func(file string) (string, error) { return "", errors.New("not found") },
	})

	if report.OK() {
		t.Fatalf("report.OK() = true, want false")
	}
	assertCheck(t, report, "context", StatusFail, "canceled")
}

func assertCheck(t *testing.T, report Report, name string, status Status, messagePart string) {
	t.Helper()
	for _, check := range report.Checks {
		if check.Name != name {
			continue
		}
		if check.Status != status {
			t.Fatalf("%s status = %q, want %q", name, check.Status, status)
		}
		if !strings.Contains(check.Message, messagePart) {
			t.Fatalf("%s message = %q, want substring %q", name, check.Message, messagePart)
		}
		return
	}
	t.Fatalf("missing check %q in %+v", name, report.Checks)
}
