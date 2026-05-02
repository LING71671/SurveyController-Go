package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/app"
	"github.com/LING71671/SurveyController-Go/internal/config"
	"github.com/LING71671/SurveyController-Go/internal/doctor"
	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/logging"
	"github.com/LING71671/SurveyController-Go/internal/provider/credamo"
	"github.com/LING71671/SurveyController-Go/internal/provider/tencent"
	"github.com/LING71671/SurveyController-Go/internal/provider/wjx"
	"github.com/LING71671/SurveyController-Go/internal/runner"
	"github.com/LING71671/SurveyController-Go/internal/version"
	"gopkg.in/yaml.v3"
)

const usage = `surveyctl is the SurveyController-go command line tool.

Usage:
  surveyctl version
  surveyctl config validate [path]
  surveyctl config generate --provider <id> --fixture <path> --url <url>
  surveyctl run --dry-run [path] [--json] [--target <n>] [--concurrency <n>]
  surveyctl run --mock [path] [--json] [--seed <n>] [--mock-fail-every <n>] [--events <text|jsonl>] [--target <n>] [--concurrency <n>] [--min-throughput <n>] [--max-heap-delta <bytes>] [--max-goroutines <n>] [--expect-failure-threshold <true|false>]
  surveyctl doctor [browser]
  surveyctl help

Commands:
  version          Print build version
  config validate  Validate a run configuration file
  config generate  Generate a run configuration from a local fixture
  run              Compile and preview a run plan
  doctor           Run local environment checks
  help             Print this help message
`

const configUsage = `Usage:
  surveyctl config validate [path]
  surveyctl config generate --provider <id> --fixture <path> --url <url>
`

const doctorUsage = `Usage:
  surveyctl doctor [browser [--probe]]
`

const runUsage = `Usage:
  surveyctl run --dry-run [path] [--json] [--target <n>] [--concurrency <n>]
  surveyctl run --mock [path] [--json] [--seed <n>] [--mock-fail-every <n>] [--events <text|jsonl>] [--target <n>] [--concurrency <n>] [--min-throughput <n>] [--max-heap-delta <bytes>] [--max-goroutines <n>] [--expect-failure-threshold <true|false>]
`

const (
	exitOK = iota
	exitFailure
	exitUsage
)

type cliError struct {
	code  int
	msg   string
	usage string
}

func (e *cliError) Error() string {
	return e.msg
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout io.Writer, stderr io.Writer) int {
	if err := execute(args, stdout); err != nil {
		var commandErr *cliError
		if errors.As(err, &commandErr) {
			fmt.Fprintln(stderr, commandErr.msg)
			if commandErr.usage != "" {
				fmt.Fprintln(stderr)
				fmt.Fprint(stderr, commandErr.usage)
			}
			return commandErr.code
		}
		fmt.Fprintf(stderr, "command failed: %v\n", err)
		return exitFailure
	}
	return exitOK
}

func execute(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stdout, usage)
		return nil
	}

	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	case "version", "-v", "--version":
		fmt.Fprintln(stdout, version.Info().String())
		return nil
	case "config":
		return runConfig(args[1:], stdout)
	case "run":
		return runRun(args[1:], stdout)
	case "doctor":
		return runDoctor(args[1:], stdout)
	default:
		return usageError(fmt.Sprintf("unknown command %q", args[0]), usage)
	}
}

func runConfig(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return usageError("missing config command", configUsage)
	}
	switch strings.ToLower(strings.TrimSpace(args[0])) {
	case "help", "-h", "--help":
		fmt.Fprint(stdout, configUsage)
		return nil
	case "validate":
		if len(args) > 2 {
			return usageError("config validate accepts at most one path", configUsage)
		}
		path := "survey.yaml"
		if len(args) == 2 && strings.TrimSpace(args[1]) != "" {
			path = args[1]
		}
		if err := config.ValidateFile(path); err != nil {
			return commandError(exitFailure, fmt.Sprintf("config validation failed: %v", err), "")
		}
		fmt.Fprintf(stdout, "config valid: %s\n", path)
		return nil
	case "generate":
		return runConfigGenerate(args[1:], stdout)
	default:
		return usageError(fmt.Sprintf("unknown config command %q", args[0]), configUsage)
	}
}

func runConfigGenerate(args []string, stdout io.Writer) error {
	var providerID string
	var fixturePath string
	var rawURL string
	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		switch strings.ToLower(arg) {
		case "":
			continue
		case "help", "-h", "--help":
			fmt.Fprint(stdout, configUsage)
			return nil
		case "--provider":
			value, next, err := readFlagValue(args, i, "--provider")
			if err != nil {
				return err
			}
			providerID = value
			i = next
		case "--fixture":
			value, next, err := readFlagValue(args, i, "--fixture")
			if err != nil {
				return err
			}
			fixturePath = value
			i = next
		case "--url":
			value, next, err := readFlagValue(args, i, "--url")
			if err != nil {
				return err
			}
			rawURL = value
			i = next
		default:
			return usageError(fmt.Sprintf("unknown config generate argument %q", arg), configUsage)
		}
	}
	if strings.TrimSpace(providerID) == "" {
		return usageError("config generate requires --provider", configUsage)
	}
	if strings.TrimSpace(fixturePath) == "" {
		return usageError("config generate requires --fixture", configUsage)
	}
	if strings.TrimSpace(rawURL) == "" {
		return usageError("config generate requires --url", configUsage)
	}

	cfg, err := generateConfigFromFixture(providerID, fixturePath, rawURL)
	if err != nil {
		return commandError(exitFailure, fmt.Sprintf("config generate failed: %v", err), "")
	}
	encoder := yaml.NewEncoder(stdout)
	if err := encoder.Encode(cfg); err != nil {
		_ = encoder.Close()
		return err
	}
	return encoder.Close()
}

func generateConfigFromFixture(providerID string, fixturePath string, rawURL string) (config.RunConfig, error) {
	id, err := domain.ParseProviderID(providerID)
	if err != nil {
		return config.RunConfig{}, err
	}
	file, err := os.Open(fixturePath)
	if err != nil {
		return config.RunConfig{}, fmt.Errorf("open fixture %q: %w", fixturePath, err)
	}
	defer file.Close()

	var survey domain.SurveyDefinition
	switch id {
	case domain.ProviderWJX:
		survey, err = wjx.ParseHTML(file, rawURL)
	case domain.ProviderTencent:
		survey, err = tencent.ParseAPI(file, rawURL)
	case domain.ProviderCredamo:
		survey, err = credamo.ParseSnapshot(file, rawURL)
	default:
		return config.RunConfig{}, fmt.Errorf("unsupported provider %q", providerID)
	}
	if err != nil {
		return config.RunConfig{}, err
	}
	survey.URL = strings.TrimSpace(rawURL)
	return config.FromSurveyDefinition(survey)
}

func readFlagValue(args []string, index int, flag string) (string, int, error) {
	next := index + 1
	if next >= len(args) || strings.TrimSpace(args[next]) == "" {
		return "", index, usageError(fmt.Sprintf("%s requires a value", flag), configUsage)
	}
	return args[next], next, nil
}

func runRun(args []string, stdout io.Writer) error {
	if len(args) == 0 {
		return usageError("run requires --dry-run or --mock", runUsage)
	}
	dryRun := false
	mockRun := false
	jsonOutput := false
	eventFormat := logging.Format("")
	overrides := app.RunPlanOverrides{}
	budget := runner.RunReportBudget{}
	budgetSet := false
	mockFailEvery := 0
	seed := int64(1)
	path := "survey.yaml"
	pathSet := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		normalized := strings.ToLower(strings.TrimSpace(arg))
		switch normalized {
		case "":
			continue
		case "help", "-h", "--help":
			fmt.Fprint(stdout, runUsage)
			return nil
		case "--dry-run":
			dryRun = true
		case "--mock":
			mockRun = true
		case "--json":
			jsonOutput = true
		case "--target":
			value, next, err := readRunFlagValue(args, i, "--target")
			if err != nil {
				return err
			}
			target, err := parsePositiveRunInt(value, "--target")
			if err != nil {
				return err
			}
			overrides.Target = target
			i = next
		case "--concurrency":
			value, next, err := readRunFlagValue(args, i, "--concurrency")
			if err != nil {
				return err
			}
			concurrency, err := parsePositiveRunInt(value, "--concurrency")
			if err != nil {
				return err
			}
			overrides.Concurrency = concurrency
			i = next
		case "--events":
			value, next, err := readRunFlagValue(args, i, "--events")
			if err != nil {
				return err
			}
			format, err := parseRunEventFormat(value)
			if err != nil {
				return err
			}
			eventFormat = format
			i = next
		case "--seed":
			value, next, err := readRunFlagValue(args, i, "--seed")
			if err != nil {
				return err
			}
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return usageError("--seed requires an integer", runUsage)
			}
			seed = parsed
			i = next
		case "--mock-fail-every":
			value, next, err := readRunFlagValue(args, i, "--mock-fail-every")
			if err != nil {
				return err
			}
			failEvery, err := parsePositiveRunInt(value, "--mock-fail-every")
			if err != nil {
				return err
			}
			mockFailEvery = failEvery
			i = next
		case "--min-throughput":
			value, next, err := readRunFlagValue(args, i, "--min-throughput")
			if err != nil {
				return err
			}
			minThroughput, err := parseNonNegativeRunFloat(value, "--min-throughput")
			if err != nil {
				return err
			}
			budget.MinThroughput = minThroughput
			budgetSet = true
			i = next
		case "--max-heap-delta":
			value, next, err := readRunFlagValue(args, i, "--max-heap-delta")
			if err != nil {
				return err
			}
			maxHeapDelta, err := parsePositiveRunUint64(value, "--max-heap-delta")
			if err != nil {
				return err
			}
			budget.MaxHeapAllocDelta = maxHeapDelta
			budgetSet = true
			i = next
		case "--max-goroutines":
			value, next, err := readRunFlagValue(args, i, "--max-goroutines")
			if err != nil {
				return err
			}
			maxGoroutines, err := parsePositiveRunInt(value, "--max-goroutines")
			if err != nil {
				return err
			}
			budget.MaxGoroutines = maxGoroutines
			budgetSet = true
			i = next
		case "--expect-failure-threshold":
			value, next, err := readRunFlagValue(args, i, "--expect-failure-threshold")
			if err != nil {
				return err
			}
			expectFailureThreshold, err := parseRunBool(value, "--expect-failure-threshold")
			if err != nil {
				return err
			}
			budget.ExpectFailureThreshold = runner.BoolBudget(expectFailureThreshold)
			budgetSet = true
			i = next
		default:
			if pathSet {
				return usageError("run accepts at most one path", runUsage)
			}
			path = arg
			pathSet = true
		}
	}
	if dryRun && mockRun {
		return usageError("run accepts only one of --dry-run or --mock", runUsage)
	}
	if !dryRun && !mockRun {
		return usageError("run requires --dry-run or --mock", runUsage)
	}
	if dryRun && eventFormat != "" {
		return usageError("run --events requires --mock", runUsage)
	}
	if jsonOutput && eventFormat != "" {
		return usageError("run --events cannot be combined with --json summary output", runUsage)
	}
	if dryRun && mockFailEvery > 0 {
		return usageError("run --mock-fail-every requires --mock", runUsage)
	}
	if dryRun && budgetSet {
		return usageError("run budget flags require --mock", runUsage)
	}

	plan, err := app.CompileRunPlanFromFile(path, overrides)
	if err != nil {
		action := "dry-run"
		if mockRun {
			action = "mock"
		}
		return commandError(exitFailure, fmt.Sprintf("run %s failed: %v", action, err), "")
	}

	if dryRun {
		if err := printDryRunPlan(stdout, path, plan, jsonOutput); err != nil {
			return err
		}
		return nil
	}
	var finishEvents func() (int, error)
	mockOptions := app.MockRunOptions{
		Seed:      seed,
		FailEvery: mockFailEvery,
	}
	if eventFormat != "" {
		mockOptions.Events, finishEvents = startRunEventStream(stdout, eventFormat, runEventBufferSize(plan))
	}
	report, err := app.RunMockPlan(contextBackground(), plan, mockOptions)
	if finishEvents != nil {
		eventCount, eventErr := finishEvents()
		if err == nil && eventErr != nil {
			err = eventErr
		}
		if err == nil {
			fmt.Fprintf(stdout, "events: %d\n", eventCount)
		}
	}
	if err != nil {
		return commandError(exitFailure, fmt.Sprintf("run mock failed: %v", err), "")
	}
	if err := printMockRunSummary(stdout, path, report, seed, jsonOutput); err != nil {
		return err
	}
	if budgetSet {
		if err := budget.Check(report); err != nil {
			return commandError(exitFailure, fmt.Sprintf("run mock budget failed: %v", err), "")
		}
	}
	return nil
}

func readRunFlagValue(args []string, index int, flag string) (string, int, error) {
	next := index + 1
	if next >= len(args) || strings.TrimSpace(args[next]) == "" {
		return "", index, usageError(fmt.Sprintf("%s requires a value", flag), runUsage)
	}
	return args[next], next, nil
}

func parsePositiveRunInt(value string, flag string) (int, error) {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return 0, usageError(fmt.Sprintf("%s requires a positive integer", flag), runUsage)
	}
	return parsed, nil
}

func parsePositiveRunUint64(value string, flag string) (uint64, error) {
	parsed, err := strconv.ParseUint(value, 10, 64)
	if err != nil || parsed == 0 {
		return 0, usageError(fmt.Sprintf("%s requires a positive integer", flag), runUsage)
	}
	return parsed, nil
}

func parseNonNegativeRunFloat(value string, flag string) (float64, error) {
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || parsed < 0 {
		return 0, usageError(fmt.Sprintf("%s requires a non-negative number", flag), runUsage)
	}
	return parsed, nil
}

func parseRunBool(value string, flag string) (bool, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, usageError(fmt.Sprintf("%s requires true or false", flag), runUsage)
	}
}

func parseRunEventFormat(value string) (logging.Format, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(logging.FormatText):
		return logging.FormatText, nil
	case string(logging.FormatJSONLines):
		return logging.FormatJSONLines, nil
	default:
		return "", usageError("--events must be text or jsonl", runUsage)
	}
}

type runEventStreamResult struct {
	count int
	err   error
}

func startRunEventStream(stdout io.Writer, format logging.Format, bufferSize int) (chan<- logging.RunEvent, func() (int, error)) {
	events := make(chan logging.RunEvent, bufferSize)
	done := make(chan runEventStreamResult, 1)
	go func() {
		writer := logging.NewEventWriter(stdout, format)
		count := 0
		for event := range events {
			if err := writer.WriteEvent(event); err != nil {
				done <- runEventStreamResult{count: count, err: err}
				return
			}
			count++
		}
		done <- runEventStreamResult{count: count}
	}()
	return events, func() (int, error) {
		close(events)
		result := <-done
		return result.count, result.err
	}
}

func runEventBufferSize(plan runner.Plan) int {
	size := plan.Target + plan.Concurrency + 4
	if size < 16 {
		return 16
	}
	if size > 4096 {
		return 4096
	}
	return size
}

func runDoctor(args []string, stdout io.Writer) error {
	if len(args) > 0 {
		switch strings.ToLower(strings.TrimSpace(args[0])) {
		case "help", "-h", "--help":
			fmt.Fprint(stdout, doctorUsage)
			return nil
		case "browser":
			runProbe, err := parseDoctorBrowserArgs(args[1:])
			if err != nil {
				return err
			}
			report := doctor.CheckBrowser(contextBackground(), doctor.BrowserOptions{RunProbe: runProbe, ProbeHeadless: true})
			printDoctorReport(stdout, "browser", report)
			if !report.OK() {
				return commandError(exitFailure, "doctor browser checks failed", "")
			}
			return nil
		default:
			return usageError(fmt.Sprintf("unknown doctor argument %q", args[0]), doctorUsage)
		}
	}
	fmt.Fprintln(stdout, "doctor checks: ok")
	fmt.Fprintln(stdout, "run `surveyctl doctor browser` for browser preflight checks")
	return nil
}

type dryRunPlanSummary struct {
	Path             string `json:"path"`
	Provider         string `json:"provider"`
	URL              string `json:"url"`
	Mode             string `json:"mode"`
	Target           int    `json:"target"`
	Concurrency      int    `json:"concurrency"`
	FailureThreshold int    `json:"failure_threshold"`
	FailStopEnabled  bool   `json:"fail_stop_enabled"`
	Headless         bool   `json:"headless"`
	QuestionCount    int    `json:"question_count"`
	ProxyEnabled     bool   `json:"proxy_enabled"`
	ReverseFill      bool   `json:"reverse_fill_enabled"`
	RandomUA         bool   `json:"random_ua_enabled"`
}

type mockRunSummary struct {
	Path              string  `json:"path"`
	Provider          string  `json:"provider"`
	URL               string  `json:"url"`
	Mode              string  `json:"mode"`
	Target            int     `json:"target"`
	Concurrency       int     `json:"concurrency"`
	Seed              int64   `json:"seed"`
	Successes         int     `json:"successes"`
	Failures          int     `json:"failures"`
	Completed         int     `json:"completed"`
	CompletionRate    float64 `json:"completion_rate"`
	SuccessRate       float64 `json:"success_rate"`
	DurationMS        int64   `json:"duration_ms"`
	ThroughputPerSec  float64 `json:"throughput_per_second"`
	Goroutines        int     `json:"goroutines"`
	HeapAllocBytes    uint64  `json:"heap_alloc_bytes"`
	HeapAllocDelta    int64   `json:"heap_alloc_delta_bytes"`
	TotalAllocDelta   uint64  `json:"total_alloc_delta_bytes"`
	StopRequested     bool    `json:"stop_requested"`
	FailureThreshold  bool    `json:"failure_threshold_reached"`
	StopReason        string  `json:"stop_reason,omitempty"`
	StopFailureReason string  `json:"stop_failure_reason,omitempty"`
	WorkerCount       int     `json:"worker_count"`
}

func printDryRunPlan(stdout io.Writer, path string, plan runner.Plan, jsonOutput bool) error {
	summary := dryRunPlanSummary{
		Path:             path,
		Provider:         plan.Provider,
		URL:              plan.URL,
		Mode:             plan.Mode.String(),
		Target:           plan.Target,
		Concurrency:      plan.Concurrency,
		FailureThreshold: plan.FailureThreshold,
		FailStopEnabled:  plan.FailStopEnabled,
		Headless:         plan.Headless,
		QuestionCount:    len(plan.Questions),
		ProxyEnabled:     plan.Proxy.Enabled,
		ReverseFill:      plan.ReverseFill.Enabled,
		RandomUA:         plan.RandomUA.Enabled,
	}
	if jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(summary)
	}
	fmt.Fprintln(stdout, "dry-run plan:")
	fmt.Fprintf(stdout, "  path: %s\n", summary.Path)
	fmt.Fprintf(stdout, "  provider: %s\n", summary.Provider)
	fmt.Fprintf(stdout, "  url: %s\n", summary.URL)
	fmt.Fprintf(stdout, "  mode: %s\n", summary.Mode)
	fmt.Fprintf(stdout, "  target: %d\n", summary.Target)
	fmt.Fprintf(stdout, "  concurrency: %d\n", summary.Concurrency)
	fmt.Fprintf(stdout, "  failure_threshold: %d\n", summary.FailureThreshold)
	fmt.Fprintf(stdout, "  fail_stop_enabled: %t\n", summary.FailStopEnabled)
	fmt.Fprintf(stdout, "  headless: %t\n", summary.Headless)
	fmt.Fprintf(stdout, "  questions: %d\n", summary.QuestionCount)
	fmt.Fprintf(stdout, "  proxy_enabled: %t\n", summary.ProxyEnabled)
	fmt.Fprintf(stdout, "  reverse_fill_enabled: %t\n", summary.ReverseFill)
	fmt.Fprintf(stdout, "  random_ua_enabled: %t\n", summary.RandomUA)
	fmt.Fprintln(stdout, "  submissions: 0 (dry run)")
	return nil
}

func printMockRunSummary(stdout io.Writer, path string, report runner.RunPlanReport, seed int64, jsonOutput bool) error {
	summary := mockRunSummary{
		Path:              path,
		Provider:          report.Provider,
		URL:               report.URL,
		Mode:              report.Mode,
		Target:            report.Target,
		Concurrency:       report.Concurrency,
		Seed:              seed,
		Successes:         report.Successes,
		Failures:          report.Failures,
		Completed:         report.Completed,
		CompletionRate:    report.CompletionRate,
		SuccessRate:       report.SuccessRate,
		DurationMS:        report.DurationMS,
		ThroughputPerSec:  report.ThroughputPerSec,
		Goroutines:        report.Goroutines,
		HeapAllocBytes:    report.HeapAllocBytes,
		HeapAllocDelta:    report.HeapAllocDelta,
		TotalAllocDelta:   report.TotalAllocDelta,
		StopRequested:     report.StopRequested,
		FailureThreshold:  report.FailureThreshold,
		StopReason:        report.StopReason,
		StopFailureReason: report.StopFailureReason,
		WorkerCount:       report.WorkerCount,
	}
	if jsonOutput {
		encoder := json.NewEncoder(stdout)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(summary)
	}
	fmt.Fprintln(stdout, "mock run:")
	fmt.Fprintf(stdout, "  path: %s\n", summary.Path)
	fmt.Fprintf(stdout, "  provider: %s\n", summary.Provider)
	fmt.Fprintf(stdout, "  url: %s\n", summary.URL)
	fmt.Fprintf(stdout, "  mode: %s\n", summary.Mode)
	fmt.Fprintf(stdout, "  target: %d\n", summary.Target)
	fmt.Fprintf(stdout, "  concurrency: %d\n", summary.Concurrency)
	fmt.Fprintf(stdout, "  seed: %d\n", summary.Seed)
	fmt.Fprintf(stdout, "  successes: %d\n", summary.Successes)
	fmt.Fprintf(stdout, "  failures: %d\n", summary.Failures)
	fmt.Fprintf(stdout, "  completed: %d\n", summary.Completed)
	fmt.Fprintf(stdout, "  completion_rate: %s\n", formatPercent(summary.CompletionRate))
	fmt.Fprintf(stdout, "  success_rate: %s\n", formatPercent(summary.SuccessRate))
	fmt.Fprintf(stdout, "  duration_ms: %d\n", summary.DurationMS)
	fmt.Fprintf(stdout, "  throughput_per_second: %.2f\n", summary.ThroughputPerSec)
	fmt.Fprintf(stdout, "  goroutines: %d\n", summary.Goroutines)
	fmt.Fprintf(stdout, "  heap_alloc_bytes: %d\n", summary.HeapAllocBytes)
	fmt.Fprintf(stdout, "  heap_alloc_delta_bytes: %d\n", summary.HeapAllocDelta)
	fmt.Fprintf(stdout, "  total_alloc_delta_bytes: %d\n", summary.TotalAllocDelta)
	fmt.Fprintf(stdout, "  stop_requested: %t\n", summary.StopRequested)
	fmt.Fprintf(stdout, "  failure_threshold_reached: %t\n", summary.FailureThreshold)
	fmt.Fprintf(stdout, "  workers: %d\n", summary.WorkerCount)
	fmt.Fprintln(stdout, "  network: disabled (mock)")
	return nil
}

func formatPercent(ratio float64) string {
	return fmt.Sprintf("%.2f%%", ratio*100)
}

func parseDoctorBrowserArgs(args []string) (bool, error) {
	runProbe := false
	for _, arg := range args {
		switch strings.ToLower(strings.TrimSpace(arg)) {
		case "":
			continue
		case "--probe":
			runProbe = true
		default:
			return false, usageError(fmt.Sprintf("unknown doctor browser argument %q", arg), doctorUsage)
		}
	}
	return runProbe, nil
}

func printDoctorReport(stdout io.Writer, name string, report doctor.Report) {
	fmt.Fprintf(stdout, "%s doctor:\n", name)
	for _, check := range report.Checks {
		fmt.Fprintf(stdout, "  [%s] %s: %s\n", check.Status, check.Name, check.Message)
	}
}

func contextBackground() context.Context {
	return context.Background()
}

func usageError(msg string, usage string) error {
	return commandError(exitUsage, msg, usage)
}

func commandError(code int, msg string, usage string) error {
	return &cliError{
		code:  code,
		msg:   msg,
		usage: usage,
	}
}
