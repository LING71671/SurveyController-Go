package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/config"
	"github.com/LING71671/SurveyController-go/internal/doctor"
	"github.com/LING71671/SurveyController-go/internal/version"
)

const usage = `surveyctl is the SurveyController-go command line tool.

Usage:
  surveyctl version
  surveyctl config validate [path]
  surveyctl doctor [browser]
  surveyctl help

Commands:
  version          Print build version
  config validate  Validate a run configuration file
  doctor           Run local environment checks
  help             Print this help message
`

const configUsage = `Usage:
  surveyctl config validate [path]
`

const doctorUsage = `Usage:
  surveyctl doctor [browser [--probe]]
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
	default:
		return usageError(fmt.Sprintf("unknown config command %q", args[0]), configUsage)
	}
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
