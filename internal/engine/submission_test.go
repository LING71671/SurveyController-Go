package engine

import (
	"context"
	"errors"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/browser"
	"github.com/LING71671/SurveyController-Go/internal/provider"
)

func TestDetectSubmissionUsesDetector(t *testing.T) {
	page := &browser.FakePage{}
	detector := stubSubmissionDetector{
		detection: provider.SubmissionDetection{
			State:              provider.SubmissionStateSuccess,
			Message:            "done",
			CompletionDetected: true,
		},
	}

	result, err := DetectSubmission(context.Background(), detector, page)
	if err != nil {
		t.Fatalf("DetectSubmission() returned error: %v", err)
	}
	if !result.Success || !result.Terminal || result.Error != nil {
		t.Fatalf("DetectSubmission() = %+v, want successful terminal result without error", result)
	}
	if !result.CompletionDetected || result.Message != "done" {
		t.Fatalf("DetectSubmission() = %+v, want completion message preserved", result)
	}
}

func TestDetectSubmissionRejectsNilDetector(t *testing.T) {
	if _, err := DetectSubmission(context.Background(), nil, &browser.FakePage{}); err == nil {
		t.Fatal("DetectSubmission() returned nil error, want nil detector failure")
	}
}

func TestDetectSubmissionReturnsDetectorError(t *testing.T) {
	wantErr := errors.New("detector failed")
	_, err := DetectSubmission(context.Background(), stubSubmissionDetector{err: wantErr}, &browser.FakePage{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("DetectSubmission() error = %v, want %v", err, wantErr)
	}
}

func TestResultFromDetectionMapsStates(t *testing.T) {
	tests := []struct {
		name               string
		detection          provider.SubmissionDetection
		wantSuccess        bool
		wantTerminal       bool
		wantStop           bool
		wantRotate         bool
		wantErrCode        apperr.Code
		wantErr            bool
		wantDefaultState   provider.SubmissionState
		wantDefaultMessage string
	}{
		{
			name: "unknown",
			detection: provider.SubmissionDetection{
				ShouldRotateProxy: true,
			},
			wantRotate:         true,
			wantDefaultState:   provider.SubmissionStateUnknown,
			wantDefaultMessage: string(provider.SubmissionStateUnknown),
		},
		{
			name: "success",
			detection: provider.SubmissionDetection{
				State:              provider.SubmissionStateSuccess,
				CompletionDetected: true,
			},
			wantSuccess:        true,
			wantTerminal:       true,
			wantDefaultState:   provider.SubmissionStateSuccess,
			wantDefaultMessage: string(provider.SubmissionStateSuccess),
		},
		{
			name: "failure",
			detection: provider.SubmissionDetection{
				State:   provider.SubmissionStateFailure,
				Message: "submit failed",
			},
			wantTerminal:       true,
			wantErr:            true,
			wantErrCode:        apperr.CodeSubmitFailed,
			wantDefaultState:   provider.SubmissionStateFailure,
			wantDefaultMessage: "submit failed",
		},
		{
			name: "verification",
			detection: provider.SubmissionDetection{
				State: provider.SubmissionStateVerificationRequired,
			},
			wantTerminal:       true,
			wantStop:           true,
			wantErr:            true,
			wantErrCode:        apperr.CodeVerificationNeeded,
			wantDefaultState:   provider.SubmissionStateVerificationRequired,
			wantDefaultMessage: string(provider.SubmissionStateVerificationRequired),
		},
		{
			name: "login",
			detection: provider.SubmissionDetection{
				State: provider.SubmissionStateLoginRequired,
			},
			wantTerminal:       true,
			wantStop:           true,
			wantErr:            true,
			wantErrCode:        apperr.CodeLoginRequired,
			wantDefaultState:   provider.SubmissionStateLoginRequired,
			wantDefaultMessage: string(provider.SubmissionStateLoginRequired),
		},
		{
			name: "device quota",
			detection: provider.SubmissionDetection{
				State:             provider.SubmissionStateDeviceQuotaLimited,
				ShouldRotateProxy: true,
			},
			wantTerminal:       true,
			wantStop:           true,
			wantRotate:         true,
			wantErr:            true,
			wantErrCode:        apperr.CodeDeviceQuotaLimited,
			wantDefaultState:   provider.SubmissionStateDeviceQuotaLimited,
			wantDefaultMessage: string(provider.SubmissionStateDeviceQuotaLimited),
		},
		{
			name: "rate limited",
			detection: provider.SubmissionDetection{
				State: provider.SubmissionStateRateLimited,
			},
			wantTerminal:       true,
			wantStop:           true,
			wantErr:            true,
			wantErrCode:        apperr.CodeRateLimited,
			wantDefaultState:   provider.SubmissionStateRateLimited,
			wantDefaultMessage: string(provider.SubmissionStateRateLimited),
		},
		{
			name: "explicit stop",
			detection: provider.SubmissionDetection{
				State:      provider.SubmissionStateFailure,
				ShouldStop: true,
			},
			wantTerminal:       true,
			wantStop:           true,
			wantErr:            true,
			wantErrCode:        apperr.CodeSubmitFailed,
			wantDefaultState:   provider.SubmissionStateFailure,
			wantDefaultMessage: string(provider.SubmissionStateFailure),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ResultFromDetection(tt.detection)
			if result.State != tt.wantDefaultState {
				t.Fatalf("State = %q, want %q", result.State, tt.wantDefaultState)
			}
			if result.Message != tt.wantDefaultMessage {
				t.Fatalf("Message = %q, want %q", result.Message, tt.wantDefaultMessage)
			}
			if result.Success != tt.wantSuccess {
				t.Fatalf("Success = %v, want %v", result.Success, tt.wantSuccess)
			}
			if result.Terminal != tt.wantTerminal {
				t.Fatalf("Terminal = %v, want %v", result.Terminal, tt.wantTerminal)
			}
			if result.ShouldStop != tt.wantStop {
				t.Fatalf("ShouldStop = %v, want %v", result.ShouldStop, tt.wantStop)
			}
			if result.ShouldRotateProxy != tt.wantRotate {
				t.Fatalf("ShouldRotateProxy = %v, want %v", result.ShouldRotateProxy, tt.wantRotate)
			}
			if tt.wantErr {
				if !apperr.IsCode(result.Error, tt.wantErrCode) {
					t.Fatalf("Error = %v, want code %q", result.Error, tt.wantErrCode)
				}
			} else if result.Error != nil {
				t.Fatalf("Error = %v, want nil", result.Error)
			}
		})
	}
}

func TestResultFromDetectionPreservesRawProviderData(t *testing.T) {
	raw := map[string]any{"captcha": true}
	result := ResultFromDetection(provider.SubmissionDetection{
		State:       provider.SubmissionStateVerificationRequired,
		ProviderRaw: raw,
	})
	if result.ProviderRaw["captcha"] != true {
		t.Fatalf("ProviderRaw = %+v, want captcha=true", result.ProviderRaw)
	}
}

type stubSubmissionDetector struct {
	detection provider.SubmissionDetection
	err       error
}

func (s stubSubmissionDetector) DetectSubmissionState(ctx context.Context, page browser.Page) (provider.SubmissionDetection, error) {
	if s.err != nil {
		return provider.SubmissionDetection{}, s.err
	}
	if err := browser.MapContextError(ctx.Err()); err != nil {
		return provider.SubmissionDetection{}, err
	}
	_, _ = page.HTML(ctx)
	return s.detection, nil
}
