package apperr

import (
	"errors"
	"strings"
	"testing"
)

func TestErrorIncludesCodeAndMessage(t *testing.T) {
	err := New(CodeConfigInvalid, "missing survey url")

	if got := err.Error(); !strings.Contains(got, "config_invalid") || !strings.Contains(got, "missing survey url") {
		t.Fatalf("Error() = %q, want code and message", got)
	}
}

func TestWrapPreservesCause(t *testing.T) {
	cause := errors.New("open file")
	err := Wrap(CodeConfigInvalid, "read config", cause)

	if !errors.Is(err, cause) {
		t.Fatalf("errors.Is(wrapped, cause) = false, want true")
	}
	if !IsCode(err, CodeConfigInvalid) {
		t.Fatalf("IsCode(wrapped, CodeConfigInvalid) = false, want true")
	}
}

func TestIsCodeRejectsDifferentCode(t *testing.T) {
	err := New(CodeLoginRequired, "login required")

	if IsCode(err, CodeConfigInvalid) {
		t.Fatalf("IsCode(login_required, config_invalid) = true, want false")
	}
}

func TestIsCodeRejectsPlainError(t *testing.T) {
	if IsCode(errors.New("plain"), CodeConfigInvalid) {
		t.Fatalf("IsCode(plain error) = true, want false")
	}
}
