package apperr

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeConfigInvalid       Code = "config_invalid"
	CodeProviderUnsupported Code = "provider_unsupported"
	CodeParseFailed         Code = "parse_failed"
	CodeBrowserStartFailed  Code = "browser_start_failed"
	CodePageLoadFailed      Code = "page_load_failed"
	CodeFillFailed          Code = "fill_failed"
	CodeSubmitFailed        Code = "submit_failed"
	CodeVerificationNeeded  Code = "verification_required"
	CodeLoginRequired       Code = "login_required"
	CodeDeviceQuotaLimited  Code = "device_quota_limited"
	CodeRateLimited         Code = "rate_limited"
	CodeProxyUnavailable    Code = "proxy_unavailable"
	CodeSampleExhausted     Code = "sample_exhausted"
	CodeUserCancelled       Code = "user_cancelled"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Wrap(code Code, message string, err error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	if e.Err == nil {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func IsCode(err error, code Code) bool {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return false
	}
	return appErr.Code == code
}

func CodeOf(err error) (Code, bool) {
	var appErr *Error
	if !errors.As(err, &appErr) {
		return "", false
	}
	if appErr.Code == "" {
		return "", false
	}
	return appErr.Code, true
}
