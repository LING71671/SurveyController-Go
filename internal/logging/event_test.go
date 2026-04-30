package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRunEventTextIsStable(t *testing.T) {
	event := RunEvent{
		Time:     time.Date(2026, 4, 30, 8, 0, 0, 0, time.UTC),
		Type:     EventWorkerProgress,
		Level:    LevelInfo,
		Message:  "submitted one response",
		WorkerID: 2,
		Fields: map[string]any{
			"done":   3,
			"target": 10,
		},
	}

	got := event.Text()
	for _, want := range []string{
		"2026-04-30T08:00:00Z",
		"INFO",
		"worker_progress",
		"worker=2",
		"submitted one response",
		"done=3 target=10",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("Text() = %q, want %q", got, want)
		}
	}
}

func TestEventWriterWritesJSONLines(t *testing.T) {
	var out bytes.Buffer
	writer := NewEventWriter(&out, FormatJSONLines)
	event := RunEvent{
		Time:    time.Date(2026, 4, 30, 8, 0, 0, 0, time.UTC),
		Type:    EventRunStarted,
		Level:   LevelInfo,
		Message: "run started",
		Fields: map[string]any{
			"target": 1,
		},
	}

	if err := writer.WriteEvent(event); err != nil {
		t.Fatalf("WriteEvent() returned error: %v", err)
	}
	if got := out.String(); !strings.HasSuffix(got, "\n") {
		t.Fatalf("jsonl output = %q, want trailing newline", got)
	}

	var decoded RunEvent
	if err := json.Unmarshal(bytes.TrimSpace(out.Bytes()), &decoded); err != nil {
		t.Fatalf("jsonl output could not be decoded: %v", err)
	}
	if decoded.Type != EventRunStarted {
		t.Fatalf("decoded.Type = %q, want %q", decoded.Type, EventRunStarted)
	}
	if decoded.Fields["target"].(float64) != 1 {
		t.Fatalf("decoded target = %v, want 1", decoded.Fields["target"])
	}
}

func TestEventWriterWritesText(t *testing.T) {
	var out bytes.Buffer
	writer := NewEventWriter(&out, FormatText)

	if err := writer.WriteEvent(RunEvent{Type: EventRunFinished, Level: LevelInfo, Message: "done"}); err != nil {
		t.Fatalf("WriteEvent() returned error: %v", err)
	}
	if got := out.String(); got != "INFO run_finished done\n" {
		t.Fatalf("text output = %q, want stable line", got)
	}
}

func TestEventWriterRejectsUnknownFormat(t *testing.T) {
	var out bytes.Buffer
	writer := NewEventWriter(&out, Format("xml"))

	err := writer.WriteEvent(RunEvent{Type: EventRunStarted})
	if err == nil || !strings.Contains(err.Error(), "unsupported log format") {
		t.Fatalf("WriteEvent() error = %v, want unsupported format", err)
	}
}
