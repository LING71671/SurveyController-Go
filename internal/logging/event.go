package logging

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

type EventType string

const (
	EventRunStarted         EventType = "run_started"
	EventWorkerStarted      EventType = "worker_started"
	EventWorkerProgress     EventType = "worker_progress"
	EventSubmissionSuccess  EventType = "submission_success"
	EventSubmissionFailure  EventType = "submission_failure"
	EventProviderWarning    EventType = "provider_warning"
	EventVerificationNeeded EventType = "verification_required"
	EventRunPaused          EventType = "run_paused"
	EventRunStopped         EventType = "run_stopped"
	EventRunFinished        EventType = "run_finished"
)

type Level string

const (
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type Format string

const (
	FormatText      Format = "text"
	FormatJSONLines Format = "jsonl"
)

type RunEvent struct {
	Time     time.Time      `json:"time"`
	Type     EventType      `json:"type"`
	Level    Level          `json:"level"`
	Message  string         `json:"message,omitempty"`
	WorkerID int            `json:"worker_id,omitempty"`
	Fields   map[string]any `json:"fields,omitempty"`
}

func NewEvent(eventType EventType, message string) RunEvent {
	return RunEvent{
		Time:    time.Now().UTC(),
		Type:    eventType,
		Level:   LevelInfo,
		Message: message,
	}
}

type EventWriter struct {
	out    io.Writer
	format Format
}

func NewEventWriter(out io.Writer, format Format) EventWriter {
	if format == "" {
		format = FormatText
	}
	return EventWriter{
		out:    out,
		format: format,
	}
}

func (w EventWriter) WriteEvent(event RunEvent) error {
	switch w.format {
	case FormatText:
		_, err := fmt.Fprintln(w.out, event.Text())
		return err
	case FormatJSONLines:
		encoder := json.NewEncoder(w.out)
		encoder.SetEscapeHTML(false)
		return encoder.Encode(event)
	default:
		return fmt.Errorf("unsupported log format %q", w.format)
	}
}

func (e RunEvent) Text() string {
	parts := make([]string, 0, 5)
	if !e.Time.IsZero() {
		parts = append(parts, e.Time.UTC().Format(time.RFC3339))
	}
	if e.Level != "" {
		parts = append(parts, strings.ToUpper(string(e.Level)))
	}
	if e.Type != "" {
		parts = append(parts, string(e.Type))
	}
	if e.WorkerID > 0 {
		parts = append(parts, fmt.Sprintf("worker=%d", e.WorkerID))
	}
	if e.Message != "" {
		parts = append(parts, e.Message)
	}
	if len(e.Fields) > 0 {
		parts = append(parts, formatFields(e.Fields))
	}
	return strings.Join(parts, " ")
}

func formatFields(fields map[string]any) string {
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, key := range keys {
		pairs = append(pairs, fmt.Sprintf("%s=%v", key, fields[key]))
	}
	return strings.Join(pairs, " ")
}
