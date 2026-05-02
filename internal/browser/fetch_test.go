package browser

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
)

func TestFetchHTMLNavigatesReadsAndClosesSession(t *testing.T) {
	pool := &recordingPool{
		session: &recordingSession{
			page: &recordingPage{html: "<html>survey</html>"},
		},
	}

	html, err := FetchHTML(context.Background(), pool, "https://example.com/survey", SessionOptions{
		Headless:      true,
		TimeoutMillis: 1000,
	})
	if err != nil {
		t.Fatalf("FetchHTML() returned error: %v", err)
	}
	if html != "<html>survey</html>" {
		t.Fatalf("FetchHTML() html = %q, want fixture html", html)
	}
	if pool.options.TimeoutMillis != 1000 || !pool.options.Headless {
		t.Fatalf("session options = %+v, want forwarded options", pool.options)
	}
	if pool.session.page.navigatedURL != "https://example.com/survey" {
		t.Fatalf("navigatedURL = %q, want survey URL", pool.session.page.navigatedURL)
	}
	if !pool.session.closed {
		t.Fatal("session was not closed")
	}
}

func TestFetchHTMLRejectsInvalidURL(t *testing.T) {
	if _, err := FetchHTML(context.Background(), &recordingPool{}, "", SessionOptions{}); err == nil {
		t.Fatal("FetchHTML(empty URL) returned nil error, want failure")
	}
	if _, err := FetchHTML(context.Background(), &recordingPool{}, "ftp://example.com/survey", SessionOptions{}); err == nil {
		t.Fatal("FetchHTML(ftp URL) returned nil error, want failure")
	}
}

func TestFetchHTMLMapsSessionStartFailure(t *testing.T) {
	startErr := errors.New("start failed")
	pool := &recordingPool{newSessionErr: startErr}

	_, err := FetchHTML(context.Background(), pool, "https://example.com/survey", SessionOptions{})
	if !apperr.IsCode(err, apperr.CodeBrowserStartFailed) {
		t.Fatalf("FetchHTML() error = %v, want browser_start_failed", err)
	}
}

func TestFetchHTMLMapsNavigateFailure(t *testing.T) {
	pool := &recordingPool{
		session: &recordingSession{
			page: &recordingPage{navigateErr: errors.New("navigation failed")},
		},
	}

	_, err := FetchHTML(context.Background(), pool, "https://example.com/survey", SessionOptions{})
	if !apperr.IsCode(err, apperr.CodePageLoadFailed) {
		t.Fatalf("FetchHTML() error = %v, want page_load_failed", err)
	}
	if !pool.session.closed {
		t.Fatal("session was not closed after navigate failure")
	}
}

func TestFetchHTMLMapsHTMLFailure(t *testing.T) {
	pool := &recordingPool{
		session: &recordingSession{
			page: &recordingPage{htmlErr: errors.New("html failed")},
		},
	}

	_, err := FetchHTML(context.Background(), pool, "https://example.com/survey", SessionOptions{})
	if !apperr.IsCode(err, apperr.CodeParseFailed) {
		t.Fatalf("FetchHTML() error = %v, want parse_failed", err)
	}
	if !pool.session.closed {
		t.Fatal("session was not closed after HTML failure")
	}
}

func TestFetchHTMLMapsCloseFailure(t *testing.T) {
	pool := &recordingPool{
		session: &recordingSession{
			page:     &recordingPage{html: "<html></html>"},
			closeErr: errors.New("close failed"),
		},
	}

	_, err := FetchHTML(context.Background(), pool, "https://example.com/survey", SessionOptions{})
	if !apperr.IsCode(err, apperr.CodeBrowserStartFailed) {
		t.Fatalf("FetchHTML() error = %v, want browser_start_failed", err)
	}
}

func TestFetchHTMLMapsCanceledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := FetchHTML(ctx, &recordingPool{}, "https://example.com/survey", SessionOptions{})
	if !apperr.IsCode(err, apperr.CodeUserCancelled) {
		t.Fatalf("FetchHTML() error = %v, want user_cancelled", err)
	}
}

func TestFetchHTMLAppliesTimeout(t *testing.T) {
	pool := &recordingPool{
		session: &recordingSession{
			page: &recordingPage{navigateDelay: 20 * time.Millisecond},
		},
	}

	_, err := FetchHTML(context.Background(), pool, "https://example.com/survey", SessionOptions{
		TimeoutMillis: 1,
	})
	if !apperr.IsCode(err, apperr.CodePageLoadFailed) {
		t.Fatalf("FetchHTML() error = %v, want page_load_failed", err)
	}
	if !pool.session.closed {
		t.Fatal("session was not closed after timeout")
	}
}

type recordingPool struct {
	session       *recordingSession
	options       SessionOptions
	newSessionErr error
}

func (p *recordingPool) NewSession(ctx context.Context, options SessionOptions) (BrowserSession, error) {
	if err := MapContextError(ctx.Err()); err != nil {
		return nil, err
	}
	p.options = options
	if p.newSessionErr != nil {
		return nil, p.newSessionErr
	}
	if p.session == nil {
		p.session = &recordingSession{page: &recordingPage{}}
	}
	return p.session, nil
}

func (p *recordingPool) Close(ctx context.Context) error {
	return MapContextError(ctx.Err())
}

type recordingSession struct {
	page     *recordingPage
	closed   bool
	closeErr error
}

func (s *recordingSession) Page() Page {
	return s.page
}

func (s *recordingSession) Close(ctx context.Context) error {
	s.closed = true
	if err := MapContextError(ctx.Err()); err != nil {
		return err
	}
	return s.closeErr
}

type recordingPage struct {
	navigatedURL  string
	html          string
	navigateErr   error
	navigateDelay time.Duration
	htmlErr       error
}

func (p *recordingPage) Navigate(ctx context.Context, rawURL string) error {
	if p.navigateDelay > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(p.navigateDelay):
		}
	}
	if err := MapContextError(ctx.Err()); err != nil {
		return err
	}
	p.navigatedURL = rawURL
	return p.navigateErr
}

func (p *recordingPage) Click(ctx context.Context, selector string) error {
	return nil
}

func (p *recordingPage) Fill(ctx context.Context, selector string, value string) error {
	return nil
}

func (p *recordingPage) HTML(ctx context.Context) (string, error) {
	if err := MapContextError(ctx.Err()); err != nil {
		return "", err
	}
	return p.html, p.htmlErr
}

func (p *recordingPage) Evaluate(ctx context.Context, script string) (string, error) {
	return "", nil
}
