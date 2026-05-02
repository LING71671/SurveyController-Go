package browser

import (
	"context"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
)

func TestFakePoolSessionAndPage(t *testing.T) {
	pool := NewFakePool()
	session, err := pool.NewSession(context.Background(), SessionOptions{Headless: true})
	if err != nil {
		t.Fatalf("NewSession() returned error: %v", err)
	}
	fakeSession := session.(*FakeSession)
	page := fakeSession.FakePage()
	page.SetHTML("<html></html>")
	page.SetEvaluateResult("document.title", "title")

	if err := session.Page().Navigate(context.Background(), "https://example.com"); err != nil {
		t.Fatalf("Navigate() returned error: %v", err)
	}
	if err := session.Page().Click(context.Background(), "#start"); err != nil {
		t.Fatalf("Click() returned error: %v", err)
	}
	if err := session.Page().Fill(context.Background(), "#name", "alice"); err != nil {
		t.Fatalf("Fill() returned error: %v", err)
	}
	html, err := session.Page().HTML(context.Background())
	if err != nil {
		t.Fatalf("HTML() returned error: %v", err)
	}
	if html != "<html></html>" {
		t.Fatalf("HTML() = %q, want fixture html", html)
	}
	result, err := session.Page().Evaluate(context.Background(), "document.title")
	if err != nil {
		t.Fatalf("Evaluate() returned error: %v", err)
	}
	if result != "title" {
		t.Fatalf("Evaluate() = %q, want title", result)
	}

	calls := page.Calls()
	if len(calls) != 4 {
		t.Fatalf("len(Calls) = %d, want 4", len(calls))
	}
	if calls[2].Name != "fill" || calls[2].Selector != "#name" || calls[2].Value != "alice" {
		t.Fatalf("fill call = %+v, want recorded fill", calls[2])
	}
}

func TestFakeCloseIsIdempotent(t *testing.T) {
	pool := NewFakePool()
	session, err := pool.NewSession(context.Background(), SessionOptions{})
	if err != nil {
		t.Fatalf("NewSession() returned error: %v", err)
	}
	if err := session.Close(context.Background()); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
	if err := session.Close(context.Background()); err != nil {
		t.Fatalf("second Close() returned error: %v", err)
	}
	if !session.(*FakeSession).Closed() {
		t.Fatalf("Closed() = false, want true")
	}
	if err := pool.Close(context.Background()); err != nil {
		t.Fatalf("pool Close() returned error: %v", err)
	}
	if err := pool.Close(context.Background()); err != nil {
		t.Fatalf("second pool Close() returned error: %v", err)
	}
}

func TestFakePoolRejectsSessionAfterClose(t *testing.T) {
	pool := NewFakePool()
	if err := pool.Close(context.Background()); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
	if _, err := pool.NewSession(context.Background(), SessionOptions{}); err == nil {
		t.Fatal("NewSession() after Close returned nil error, want failure")
	}
}

func TestFakeOperationsHonorContextCancel(t *testing.T) {
	pool := NewFakePool()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := pool.NewSession(ctx, SessionOptions{}); !apperr.IsCode(err, apperr.CodeUserCancelled) {
		t.Fatalf("NewSession(canceled) error = %v, want user_cancelled", err)
	}

	session, err := pool.NewSession(context.Background(), SessionOptions{})
	if err != nil {
		t.Fatalf("NewSession() returned error: %v", err)
	}
	if err := session.Page().Navigate(ctx, "https://example.com"); !apperr.IsCode(err, apperr.CodeUserCancelled) {
		t.Fatalf("Navigate(canceled) error = %v, want user_cancelled", err)
	}
}

func TestFakePageCallsSnapshot(t *testing.T) {
	page := &FakePage{}
	if err := page.Click(context.Background(), "#one"); err != nil {
		t.Fatalf("Click() returned error: %v", err)
	}
	calls := page.Calls()
	calls[0].Selector = "#mutated"

	next := page.Calls()
	if next[0].Selector != "#one" {
		t.Fatalf("calls snapshot mutation affected fake page")
	}
}
