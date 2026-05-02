package browser

import (
	"context"
	"errors"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
)

func TestMapContextError(t *testing.T) {
	if err := MapContextError(nil); err != nil {
		t.Fatalf("MapContextError(nil) = %v, want nil", err)
	}
	if err := MapContextError(context.Canceled); !apperr.IsCode(err, apperr.CodeUserCancelled) {
		t.Fatalf("MapContextError(canceled) = %v, want user_cancelled", err)
	}
	if err := MapContextError(context.DeadlineExceeded); !apperr.IsCode(err, apperr.CodePageLoadFailed) {
		t.Fatalf("MapContextError(deadline) = %v, want page_load_failed", err)
	}
	plain := errors.New("plain")
	if err := MapContextError(plain); !errors.Is(err, plain) {
		t.Fatalf("MapContextError(plain) = %v, want original", err)
	}
}

func TestValidateSelector(t *testing.T) {
	if err := ValidateSelector("#submit"); err != nil {
		t.Fatalf("ValidateSelector() returned error: %v", err)
	}
	if err := ValidateSelector(""); err == nil {
		t.Fatal("ValidateSelector(empty) returned nil error, want failure")
	}
}
