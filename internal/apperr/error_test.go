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

func TestCodeOf(t *testing.T) {
	err := Wrap(CodeSubmitFailed, "submit", errors.New("transport"))

	code, ok := CodeOf(err)
	if !ok || code != CodeSubmitFailed {
		t.Fatalf("CodeOf() = (%q, %v), want (%q, true)", code, ok, CodeSubmitFailed)
	}

	if code, ok := CodeOf(errors.New("plain")); ok || code != "" {
		t.Fatalf("CodeOf(plain) = (%q, %v), want empty false", code, ok)
	}
}
