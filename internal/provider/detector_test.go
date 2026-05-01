package provider

import (
	"context"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/apperr"
	"github.com/LING71671/SurveyController-go/internal/browser"
)

func TestSubmissionStateTerminal(t *testing.T) {
	tests := []struct {
		state SubmissionState
		want  bool
	}{
		{state: SubmissionStateUnknown, want: false},
		{state: SubmissionStateSuccess, want: true},
		{state: SubmissionStateFailure, want: true},
		{state: SubmissionStateVerificationRequired, want: true},
		{state: SubmissionStateLoginRequired, want: true},
		{state: SubmissionStateDeviceQuotaLimited, want: true},
		{state: SubmissionStateRateLimited, want: true},
	}

	for _, tt := range tests {
		if got := tt.state.Terminal(); got != tt.want {
			t.Fatalf("%s.Terminal() = %v, want %v", tt.state, got, tt.want)
		}
	}
}

func TestSubmissionStateErrorCode(t *testing.T) {
	tests := []struct {
		state SubmissionState
		code  apperr.Code
		ok    bool
	}{
		{state: SubmissionStateUnknown, ok: false},
		{state: SubmissionStateSuccess, ok: false},
		{state: SubmissionStateFailure, code: apperr.CodeSubmitFailed, ok: true},
		{state: SubmissionStateVerificationRequired, code: apperr.CodeVerificationNeeded, ok: true},
		{state: SubmissionStateLoginRequired, code: apperr.CodeLoginRequired, ok: true},
		{state: SubmissionStateDeviceQuotaLimited, code: apperr.CodeDeviceQuotaLimited, ok: true},
		{state: SubmissionStateRateLimited, code: apperr.CodeRateLimited, ok: true},
	}

	for _, tt := range tests {
		code, ok := tt.state.ErrorCode()
		if ok != tt.ok || code != tt.code {
			t.Fatalf("%s.ErrorCode() = (%q, %v), want (%q, %v)", tt.state, code, ok, tt.code, tt.ok)
		}
	}
}

func TestSubmissionDetectionHelpers(t *testing.T) {
	success := SubmissionDetection{State: SubmissionStateSuccess, CompletionDetected: true}
	if !success.Successful() || !success.Terminal() || success.Error() != nil {
		t.Fatalf("success detection helpers returned unexpected values")
	}

	verification := SubmissionDetection{
		State:   SubmissionStateVerificationRequired,
		Message: "verification required",
	}
	if verification.Successful() || !verification.Terminal() {
		t.Fatalf("verification detection helpers returned unexpected values")
	}
	if err := verification.Error(); !apperr.IsCode(err, apperr.CodeVerificationNeeded) {
		t.Fatalf("verification Error() = %v, want verification_required", err)
	}
}

func TestDetectorInterfaces(t *testing.T) {
	var completion CompletionDetector = stubDetector{}
	var submission SubmissionDetector = stubDetector{}
	var successSignal SubmissionSuccessSignalConsumer = stubDetector{}
	page := &browser.FakePage{}

	ok, err := completion.IsCompletion(context.Background(), page)
	if err != nil || !ok {
		t.Fatalf("IsCompletion() = (%v, %v), want true nil", ok, err)
	}
	detection, err := submission.DetectSubmissionState(context.Background(), page)
	if err != nil || detection.State != SubmissionStateSuccess {
		t.Fatalf("DetectSubmissionState() = (%+v, %v), want success nil", detection, err)
	}
	consumed, err := successSignal.ConsumeSubmissionSuccessSignal(context.Background(), page)
	if err != nil || !consumed {
		t.Fatalf("ConsumeSubmissionSuccessSignal() = (%v, %v), want true nil", consumed, err)
	}
}

type stubDetector struct{}

func (stubDetector) IsCompletion(ctx context.Context, page browser.Page) (bool, error) {
	if err := browser.MapContextError(ctx.Err()); err != nil {
		return false, err
	}
	_, _ = page.HTML(ctx)
	return true, nil
}

func (stubDetector) DetectSubmissionState(ctx context.Context, page browser.Page) (SubmissionDetection, error) {
	if err := browser.MapContextError(ctx.Err()); err != nil {
		return SubmissionDetection{}, err
	}
	_, _ = page.HTML(ctx)
	return SubmissionDetection{State: SubmissionStateSuccess, CompletionDetected: true}, nil
}

func (stubDetector) ConsumeSubmissionSuccessSignal(ctx context.Context, page browser.Page) (bool, error) {
	if err := browser.MapContextError(ctx.Err()); err != nil {
		return false, err
	}
	_, _ = page.HTML(ctx)
	return true, nil
}
