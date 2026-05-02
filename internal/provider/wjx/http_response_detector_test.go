package wjx

import (
	"net/http"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/provider"
)

func TestDetectHTTPSubmissionResponse(t *testing.T) {
	tests := []struct {
		name           string
		response       HTTPSubmissionResponse
		wantState      provider.SubmissionState
		wantMessage    string
		wantCompletion bool
		wantStop       bool
	}{
		{
			name: "success",
			response: HTTPSubmissionResponse{
				StatusCode: http.StatusOK,
				Body:       "提交成功",
			},
			wantState:      provider.SubmissionStateSuccess,
			wantMessage:    "submission accepted",
			wantCompletion: true,
		},
		{
			name: "verification",
			response: HTTPSubmissionResponse{
				StatusCode: http.StatusOK,
				Body:       "请完成智能验证后继续",
			},
			wantState:   provider.SubmissionStateVerificationRequired,
			wantMessage: "verification required",
			wantStop:    true,
		},
		{
			name: "login",
			response: HTTPSubmissionResponse{
				StatusCode: http.StatusOK,
				Body:       "请先登录后再提交",
			},
			wantState:   provider.SubmissionStateLoginRequired,
			wantMessage: "login required",
			wantStop:    true,
		},
		{
			name: "device quota",
			response: HTTPSubmissionResponse{
				StatusCode: http.StatusOK,
				Body:       "每台设备只能填写一次",
			},
			wantState:   provider.SubmissionStateDeviceQuotaLimited,
			wantMessage: "device quota limited",
			wantStop:    true,
		},
		{
			name: "rate limited status",
			response: HTTPSubmissionResponse{
				StatusCode: http.StatusTooManyRequests,
				Header:     http.Header{"Retry-After": []string{"30"}},
				Body:       "too many requests",
			},
			wantState:   provider.SubmissionStateRateLimited,
			wantMessage: "rate limited",
			wantStop:    true,
		},
		{
			name: "validation failure",
			response: HTTPSubmissionResponse{
				StatusCode: http.StatusBadRequest,
				Body:       "参数错误",
			},
			wantState:   provider.SubmissionStateFailure,
			wantMessage: "invalid submission",
			wantStop:    true,
		},
		{
			name: "unknown",
			response: HTTPSubmissionResponse{
				StatusCode: http.StatusAccepted,
				Header:     http.Header{"Content-Type": []string{"text/plain"}},
				Body:       "queued",
			},
			wantState:   provider.SubmissionStateUnknown,
			wantMessage: "submission state unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectHTTPSubmissionResponse(tt.response)
			if got.State != tt.wantState {
				t.Fatalf("State = %q, want %q", got.State, tt.wantState)
			}
			if got.Message != tt.wantMessage {
				t.Fatalf("Message = %q, want %q", got.Message, tt.wantMessage)
			}
			if got.CompletionDetected != tt.wantCompletion {
				t.Fatalf("CompletionDetected = %v, want %v", got.CompletionDetected, tt.wantCompletion)
			}
			if got.ShouldStop != tt.wantStop {
				t.Fatalf("ShouldStop = %v, want %v", got.ShouldStop, tt.wantStop)
			}
			if got.ProviderRaw["status_code"] != tt.response.StatusCode {
				t.Fatalf("ProviderRaw status_code = %+v, want %d", got.ProviderRaw["status_code"], tt.response.StatusCode)
			}
		})
	}
}

func TestDetectHTTPSubmissionResponsePrioritizesSafetyBlocks(t *testing.T) {
	got := DetectHTTPSubmissionResponse(HTTPSubmissionResponse{
		StatusCode: http.StatusOK,
		Body:       "提交成功，但是需要验证码",
	})

	if got.State != provider.SubmissionStateVerificationRequired {
		t.Fatalf("State = %q, want verification required", got.State)
	}
	if !got.ShouldStop || got.CompletionDetected {
		t.Fatalf("DetectHTTPSubmissionResponse() = %+v, want stop without completion", got)
	}
}
