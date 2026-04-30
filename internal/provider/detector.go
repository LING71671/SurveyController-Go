package provider

import (
	"context"

	"github.com/LING71671/SurveyController-go/internal/apperr"
	"github.com/LING71671/SurveyController-go/internal/browser"
)

type SubmissionState string

const (
	SubmissionStateUnknown              SubmissionState = "unknown"
	SubmissionStateSuccess              SubmissionState = "success"
	SubmissionStateFailure              SubmissionState = "failure"
	SubmissionStateVerificationRequired SubmissionState = "verification_required"
	SubmissionStateLoginRequired        SubmissionState = "login_required"
	SubmissionStateDeviceQuotaLimited   SubmissionState = "device_quota_limited"
)

type SubmissionDetection struct {
	State              SubmissionState `json:"state"`
	Message            string          `json:"message,omitempty"`
	CompletionDetected bool            `json:"completion_detected,omitempty"`
	ShouldStop         bool            `json:"should_stop,omitempty"`
	ShouldRotateProxy  bool            `json:"should_rotate_proxy,omitempty"`
	ProviderRaw        map[string]any  `json:"provider_raw,omitempty"`
}

type CompletionDetector interface {
	IsCompletion(ctx context.Context, page browser.Page) (bool, error)
}

type SubmissionDetector interface {
	DetectSubmissionState(ctx context.Context, page browser.Page) (SubmissionDetection, error)
}

type SubmissionSuccessSignalConsumer interface {
	ConsumeSubmissionSuccessSignal(ctx context.Context, page browser.Page) (bool, error)
}

func (s SubmissionState) Terminal() bool {
	switch s {
	case SubmissionStateSuccess,
		SubmissionStateFailure,
		SubmissionStateVerificationRequired,
		SubmissionStateLoginRequired,
		SubmissionStateDeviceQuotaLimited:
		return true
	default:
		return false
	}
}

func (s SubmissionState) ErrorCode() (apperr.Code, bool) {
	switch s {
	case SubmissionStateFailure:
		return apperr.CodeSubmitFailed, true
	case SubmissionStateVerificationRequired:
		return apperr.CodeVerificationNeeded, true
	case SubmissionStateLoginRequired:
		return apperr.CodeLoginRequired, true
	case SubmissionStateDeviceQuotaLimited:
		return apperr.CodeDeviceQuotaLimited, true
	default:
		return "", false
	}
}

func (d SubmissionDetection) Successful() bool {
	return d.State == SubmissionStateSuccess
}

func (d SubmissionDetection) Terminal() bool {
	return d.State.Terminal()
}

func (d SubmissionDetection) Error() error {
	code, ok := d.State.ErrorCode()
	if !ok {
		return nil
	}
	message := d.Message
	if message == "" {
		message = string(d.State)
	}
	return apperr.New(code, message)
}
