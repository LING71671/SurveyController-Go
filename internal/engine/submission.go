package engine

import (
	"context"
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/browser"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

type SubmissionResult struct {
	State              provider.SubmissionState
	Message            string
	Success            bool
	Terminal           bool
	CompletionDetected bool
	ShouldStop         bool
	ShouldRotateProxy  bool
	ProviderRaw        map[string]any
	Error              error
}

func DetectSubmission(ctx context.Context, detector provider.SubmissionDetector, page browser.Page) (SubmissionResult, error) {
	if detector == nil {
		return SubmissionResult{}, fmt.Errorf("submission detector is required")
	}

	detection, err := detector.DetectSubmissionState(ctx, page)
	if err != nil {
		return SubmissionResult{}, err
	}
	return ResultFromDetection(detection), nil
}

func ResultFromDetection(detection provider.SubmissionDetection) SubmissionResult {
	state := detection.State
	if state == "" {
		state = provider.SubmissionStateUnknown
	}

	message := strings.TrimSpace(detection.Message)
	if message == "" {
		message = string(state)
	}

	normalized := provider.SubmissionDetection{
		State:              state,
		Message:            message,
		CompletionDetected: detection.CompletionDetected,
		ShouldStop:         detection.ShouldStop,
		ShouldRotateProxy:  detection.ShouldRotateProxy,
		ProviderRaw:        detection.ProviderRaw,
	}

	return SubmissionResult{
		State:              state,
		Message:            message,
		Success:            normalized.Successful(),
		Terminal:           normalized.Terminal(),
		CompletionDetected: normalized.CompletionDetected,
		ShouldStop:         normalized.ShouldStop || shouldStopForSubmissionState(state),
		ShouldRotateProxy:  normalized.ShouldRotateProxy,
		ProviderRaw:        normalized.ProviderRaw,
		Error:              normalized.Error(),
	}
}

func shouldStopForSubmissionState(state provider.SubmissionState) bool {
	switch state {
	case provider.SubmissionStateVerificationRequired,
		provider.SubmissionStateLoginRequired,
		provider.SubmissionStateDeviceQuotaLimited:
		return true
	default:
		return false
	}
}
