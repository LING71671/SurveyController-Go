package browser

import (
	"context"
	"fmt"

	"github.com/LING71671/SurveyController-go/internal/apperr"
)

type SessionOptions struct {
	Headless      bool
	ProxyURL      string
	UserAgent     string
	TimeoutMillis int
}

type BrowserPool interface {
	NewSession(ctx context.Context, options SessionOptions) (BrowserSession, error)
	Close(ctx context.Context) error
}

type BrowserSession interface {
	Page() Page
	Close(ctx context.Context) error
}

type Page interface {
	Navigate(ctx context.Context, rawURL string) error
	Click(ctx context.Context, selector string) error
	Fill(ctx context.Context, selector string, value string) error
	HTML(ctx context.Context) (string, error)
	Evaluate(ctx context.Context, script string) (string, error)
}

func MapContextError(err error) error {
	if err == nil {
		return nil
	}
	if err == context.Canceled {
		return apperr.Wrap(apperr.CodeUserCancelled, "browser operation cancelled", err)
	}
	if err == context.DeadlineExceeded {
		return apperr.Wrap(apperr.CodePageLoadFailed, "browser operation timed out", err)
	}
	return err
}

func ValidateSelector(selector string) error {
	if selector == "" {
		return fmt.Errorf("selector is required")
	}
	return nil
}
