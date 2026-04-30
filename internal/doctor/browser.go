package doctor

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type Status string

const (
	StatusOK   Status = "ok"
	StatusWarn Status = "warn"
	StatusFail Status = "fail"
)

type Check struct {
	Name    string
	Status  Status
	Message string
}

type Report struct {
	Checks []Check
}

func (r Report) OK() bool {
	for _, check := range r.Checks {
		if check.Status == StatusFail {
			return false
		}
	}
	return true
}

type BrowserOptions struct {
	GOOS          string
	Env           map[string]string
	LookPath      func(file string) (string, error)
	RunProbe      bool
	Probe         BrowserProbe
	ProbeHeadless bool
	ProbeBrowsers []string
}

type BrowserProbe func(ctx context.Context, request BrowserProbeRequest) BrowserProbeResult

type BrowserProbeRequest struct {
	Headless       bool
	PreferBrowsers []string
}

type BrowserProbeResult struct {
	OK        bool
	Browser   string
	ErrorKind string
	Message   string
	Elapsed   time.Duration
}

func CheckBrowser(ctx context.Context, options BrowserOptions) Report {
	if options.GOOS == "" {
		options.GOOS = runtime.GOOS
	}
	if options.Env == nil {
		options.Env = environ()
	}
	if options.LookPath == nil {
		options.LookPath = exec.LookPath
	}

	checks := []Check{
		checkContext(ctx),
		checkOperatingSystem(options.GOOS),
		checkPlaywrightBrowserPath(options.Env),
		checkSystemBrowser(options.GOOS, options.LookPath),
		checkProxyPlaceholder(),
	}
	if options.RunProbe {
		checks = append(checks, checkBrowserLaunchProbe(ctx, options))
	}
	return Report{Checks: checks}
}

func checkContext(ctx context.Context) Check {
	if err := ctx.Err(); err != nil {
		return Check{Name: "context", Status: StatusFail, Message: err.Error()}
	}
	return Check{Name: "context", Status: StatusOK, Message: "doctor context is active"}
}

func checkOperatingSystem(goos string) Check {
	switch goos {
	case "windows", "darwin", "linux":
		return Check{Name: "operating_system", Status: StatusOK, Message: fmt.Sprintf("%s is supported", goos)}
	default:
		return Check{Name: "operating_system", Status: StatusFail, Message: fmt.Sprintf("%s is not supported for browser mode", goos)}
	}
}

func checkPlaywrightBrowserPath(env map[string]string) Check {
	value := env["PLAYWRIGHT_BROWSERS_PATH"]
	if value == "" {
		return Check{
			Name:    "playwright_browsers_path",
			Status:  StatusWarn,
			Message: "PLAYWRIGHT_BROWSERS_PATH is not set; the default Playwright cache will be used",
		}
	}
	return Check{Name: "playwright_browsers_path", Status: StatusOK, Message: "PLAYWRIGHT_BROWSERS_PATH is set"}
}

func checkSystemBrowser(goos string, lookPath func(file string) (string, error)) Check {
	for _, name := range browserExecutableNames(goos) {
		path, err := lookPath(name)
		if err == nil {
			return Check{Name: "system_browser", Status: StatusOK, Message: fmt.Sprintf("found %s at %s", name, path)}
		}
	}
	return Check{
		Name:    "system_browser",
		Status:  StatusWarn,
		Message: "no common system browser executable found on PATH; a managed Playwright browser can still be installed later",
	}
}

func browserExecutableNames(goos string) []string {
	switch goos {
	case "windows":
		return []string{"msedge", "chrome", "chromium"}
	case "darwin":
		return []string{"Google Chrome", "Microsoft Edge", "chromium"}
	default:
		return []string{"google-chrome", "chromium", "chromium-browser", "microsoft-edge"}
	}
}

func checkProxyPlaceholder() Check {
	return Check{
		Name:    "proxy_connectivity",
		Status:  StatusWarn,
		Message: "proxy connectivity check is not configured yet",
	}
}

func checkBrowserLaunchProbe(ctx context.Context, options BrowserOptions) Check {
	if err := ctx.Err(); err != nil {
		return Check{Name: "browser_launch_probe", Status: StatusFail, Message: err.Error()}
	}
	if options.Probe == nil {
		return Check{
			Name:    "browser_launch_probe",
			Status:  StatusWarn,
			Message: "browser launch probe is not configured yet",
		}
	}
	result := options.Probe(ctx, BrowserProbeRequest{
		Headless:       options.ProbeHeadless,
		PreferBrowsers: append([]string(nil), options.ProbeBrowsers...),
	})
	if result.OK {
		message := "browser launch probe passed"
		if result.Browser != "" {
			message = fmt.Sprintf("browser launch probe passed with %s", result.Browser)
		}
		if result.Elapsed > 0 {
			message = fmt.Sprintf("%s in %s", message, result.Elapsed.Truncate(time.Millisecond))
		}
		return Check{Name: "browser_launch_probe", Status: StatusOK, Message: message}
	}
	message := result.Message
	if message == "" {
		message = "browser launch probe failed"
	}
	if result.ErrorKind != "" {
		message = fmt.Sprintf("%s: %s", result.ErrorKind, message)
	}
	return Check{Name: "browser_launch_probe", Status: StatusFail, Message: message}
}

func environ() map[string]string {
	env := map[string]string{}
	for _, entry := range os.Environ() {
		for i, ch := range entry {
			if ch == '=' {
				env[entry[:i]] = entry[i+1:]
				break
			}
		}
	}
	return env
}
