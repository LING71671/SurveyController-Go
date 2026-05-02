package browser

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
)

func FetchHTML(ctx context.Context, pool BrowserPool, rawURL string, options SessionOptions) (string, error) {
	if pool == nil {
		return "", fmt.Errorf("browser pool is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}
	if err := validatePageURL(rawURL); err != nil {
		return "", err
	}

	runCtx := ctx
	var cancel context.CancelFunc
	if options.TimeoutMillis > 0 {
		runCtx, cancel = context.WithTimeout(ctx, time.Duration(options.TimeoutMillis)*time.Millisecond)
		defer cancel()
	}
	if err := MapContextError(runCtx.Err()); err != nil {
		return "", err
	}

	session, err := pool.NewSession(runCtx, options)
	if err != nil {
		return "", mapBrowserOperationError(apperr.CodeBrowserStartFailed, "create browser session", err)
	}

	html, fetchErr := fetchSessionHTML(runCtx, session, rawURL)
	closeErr := session.Close(context.WithoutCancel(ctx))
	if fetchErr != nil {
		return "", fetchErr
	}
	if closeErr != nil {
		return "", mapBrowserOperationError(apperr.CodeBrowserStartFailed, "close browser session", closeErr)
	}
	return html, nil
}

func fetchSessionHTML(ctx context.Context, session BrowserSession, rawURL string) (string, error) {
	page := session.Page()
	if err := page.Navigate(ctx, rawURL); err != nil {
		return "", mapBrowserOperationError(apperr.CodePageLoadFailed, "navigate browser page", err)
	}
	html, err := page.HTML(ctx)
	if err != nil {
		return "", mapBrowserOperationError(apperr.CodeParseFailed, "read browser page html", err)
	}
	return html, nil
}

func validatePageURL(rawURL string) error {
	if strings.TrimSpace(rawURL) == "" {
		return fmt.Errorf("url is required")
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("parse url: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https")
	}
	if parsed.Host == "" {
		return fmt.Errorf("url host is required")
	}
	return nil
}

func mapBrowserOperationError(code apperr.Code, message string, err error) error {
	if err == nil {
		return nil
	}
	if mapped := MapContextError(err); mapped != err {
		return mapped
	}
	var appErr *apperr.Error
	if errors.As(err, &appErr) {
		return err
	}
	return apperr.Wrap(code, message, err)
}
