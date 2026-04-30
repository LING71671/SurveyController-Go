package runner

import (
	"github.com/LING71671/SurveyController-go/internal/engine"
	"github.com/LING71671/SurveyController-go/internal/logging"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

func (s *RunState) RecordSubmissionResult(workerID int, result engine.SubmissionResult) {
	switch {
	case result.Success:
		s.RecordSuccess(workerID)
	case result.Error != nil:
		s.RecordFailure(workerID, result.Message)
	default:
		s.SetWorkerStatus(workerID, WorkerStatusRunning, result.Message)
	}

	if result.ShouldStop {
		s.RequestStop(result.Message)
	}
}

func EventForSubmissionResult(workerID int, result engine.SubmissionResult) logging.RunEvent {
	event := logging.RunEvent{
		Type:     logging.EventWorkerProgress,
		Level:    logging.LevelInfo,
		WorkerID: workerID,
		Message:  result.Message,
		Fields: map[string]any{
			"state":               string(result.State),
			"terminal":            result.Terminal,
			"completion_detected": result.CompletionDetected,
			"should_stop":         result.ShouldStop,
			"should_rotate_proxy": result.ShouldRotateProxy,
		},
	}

	switch {
	case result.Success:
		event.Type = logging.EventSubmissionSuccess
		event.Message = messageOrDefault(result.Message, "submission succeeded")
	case result.State == provider.SubmissionStateVerificationRequired:
		event.Type = logging.EventVerificationNeeded
		event.Level = logging.LevelWarn
	case result.Error != nil:
		event.Type = logging.EventSubmissionFailure
		event.Level = logging.LevelError
	}

	return event
}

func messageOrDefault(message string, fallback string) string {
	if message == "" {
		return fallback
	}
	return message
}
