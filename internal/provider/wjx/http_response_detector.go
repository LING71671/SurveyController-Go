package wjx

import (
	"net/http"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/provider"
)

type httpResponseSignal struct {
	state   provider.SubmissionState
	message string
	stop    bool
}

func DetectHTTPSubmissionResponse(response HTTPSubmissionResponse) provider.SubmissionDetection {
	statusCode := response.StatusCode
	body := normalizeResponseText(response.Body)
	contentType := strings.TrimSpace(response.Header.Get("Content-Type"))

	signal := detectHTTPResponseSignal(statusCode, body, response.Header)
	detection := provider.SubmissionDetection{
		State:      signal.state,
		Message:    signal.message,
		ShouldStop: signal.stop,
		ProviderRaw: map[string]any{
			"status_code":  statusCode,
			"content_type": contentType,
		},
	}
	if detection.State == provider.SubmissionStateSuccess {
		detection.CompletionDetected = true
	}
	return detection
}

func detectHTTPResponseSignal(statusCode int, body string, header http.Header) httpResponseSignal {
	switch {
	case statusCode == http.StatusTooManyRequests || header.Get("Retry-After") != "":
		return terminalHTTPResponseSignal(provider.SubmissionStateRateLimited, "rate limited")
	case containsAny(body, "过于频繁", "操作频繁", "too many requests", "rate limit", "稍后再试"):
		return terminalHTTPResponseSignal(provider.SubmissionStateRateLimited, "rate limited")
	case containsAny(body, "验证码", "智能验证", "滑块验证", "验证失败", "需要验证", "captcha"):
		return terminalHTTPResponseSignal(provider.SubmissionStateVerificationRequired, "verification required")
	case containsAny(body, "请先登录", "登录后", "login required", "sign in"):
		return terminalHTTPResponseSignal(provider.SubmissionStateLoginRequired, "login required")
	case containsAny(body, "每台设备", "每个设备", "每个ip", "每个 ip", "已经填写过", "已填写过", "只能填写一次", "重复提交"):
		return terminalHTTPResponseSignal(provider.SubmissionStateDeviceQuotaLimited, "device quota limited")
	case statusCode >= 500:
		return terminalHTTPResponseSignal(provider.SubmissionStateFailure, "server rejected submission")
	case statusCode == http.StatusBadRequest || statusCode == http.StatusUnprocessableEntity:
		return terminalHTTPResponseSignal(provider.SubmissionStateFailure, "invalid submission")
	case containsAny(body, "参数错误", "提交失败", "非法", "无效", "invalid", "failed"):
		return terminalHTTPResponseSignal(provider.SubmissionStateFailure, "submission rejected")
	case statusCode >= 200 && statusCode < 300 && containsAny(body, "提交成功", "答卷提交成功", "已完成", "success", "succeeded"):
		return httpResponseSignal{state: provider.SubmissionStateSuccess, message: "submission accepted"}
	default:
		return httpResponseSignal{state: provider.SubmissionStateUnknown, message: "submission state unknown"}
	}
}

func terminalHTTPResponseSignal(state provider.SubmissionState, message string) httpResponseSignal {
	return httpResponseSignal{
		state:   state,
		message: message,
		stop:    true,
	}
}

func normalizeResponseText(body string) string {
	return strings.ToLower(strings.Join(strings.Fields(body), " "))
}

func containsAny(value string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(value, strings.ToLower(needle)) {
			return true
		}
	}
	return false
}
