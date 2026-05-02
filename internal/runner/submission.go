package runner

import (
	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/engine"
	"github.com/LING71671/SurveyController-Go/internal/logging"
	"github.com/LING71671/SurveyController-Go/internal/provider"
)

func (s *RunState) RecordSubmissionResult(workerID int, result engine.SubmissionResult) {
	switch {
	case result.Success:
		s.RecordSuccess(workerID)
	case result.Error != nil:
		code, _ := apperr.CodeOf(result.Error)
		s.RecordFailureWithCode(workerID, result.Message, code)
	default:
		s.SetWorkerStatus(workerID, WorkerStatusRunning, result.Message)
	}

	if result.ShouldStop {
		code, _ := apperr.CodeOf(result.Error)
		s.RequestStopWithCode(result.Message, code)
	}
}

func EventForSubmissionResult(workerID int, result engine.SubmissionResult) logging.RunEvent {
	event := logging.NewEvent(logging.EventWorkerProgress, result.Message)
	event.WorkerID = workerID
	event.Fields = map[string]any{
		"state":               string(result.State),
		"terminal":            result.Terminal,
		"completion_detected": result.CompletionDetected,
		"should_stop":         result.ShouldStop,
		"should_rotate_proxy": result.ShouldRotateProxy,
	}
	if code, ok := apperr.CodeOf(result.Error); ok {
		event.Fields["error_code"] = string(code)
		event.Fields["failure_reason"] = string(code)
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

func errorCode(err error) apperr.Code {
	code, _ := apperr.CodeOf(err)
	return code
}

func addErrorFields(event *logging.RunEvent, err error) {
	code, ok := apperr.CodeOf(err)
	if !ok {
		return
	}
	if event.Fields == nil {
		event.Fields = map[string]any{}
	}
	event.Fields["error_code"] = string(code)
	event.Fields["failure_reason"] = string(code)
}
