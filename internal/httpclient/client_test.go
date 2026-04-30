package httpclient

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestClientSendsDefaultAndRequestHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("User-Agent"); got != "custom-agent" {
			t.Fatalf("User-Agent = %q, want custom-agent", got)
		}
		if got := r.Header.Get("X-Default"); got != "default" {
			t.Fatalf("X-Default = %q, want default", got)
		}
		if got := r.Header.Get("X-Request"); got != "request" {
			t.Fatalf("X-Request = %q, want request", got)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(Options{
		UserAgent: "custom-agent",
		DefaultHeader: http.Header{
			"X-Default": []string{"default"},
		},
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	resp, err := client.Do(context.Background(), RequestOptions{
		URL: server.URL,
		Header: http.Header{
			"X-Request": []string{"request"},
		},
	})
	if err != nil {
		t.Fatalf("Do() returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("StatusCode = %d, want 204", resp.StatusCode)
	}
}

func TestClientCookieJarPersistsCookies(t *testing.T) {
	var seenCookie bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("sid"); err == nil && cookie.Value == "abc" {
			seenCookie = true
		}
		http.SetCookie(w, &http.Cookie{Name: "sid", Value: "abc"})
	}))
	defer server.Close()

	client, err := New(Options{})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	for i := 0; i < 2; i++ {
		resp, err := client.Get(context.Background(), server.URL)
		if err != nil {
			t.Fatalf("Get() returned error: %v", err)
		}
		resp.Body.Close()
	}
	if !seenCookie {
		t.Fatalf("server did not receive persisted cookie")
	}
}

func TestClientUsesInjectedTransport(t *testing.T) {
	client, err := New(Options{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(strings.NewReader("ok")),
				Header:     http.Header{},
				Request:    req,
			}, nil
		}),
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	resp, err := client.Get(context.Background(), "https://example.com")
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("StatusCode = %d, want 201", resp.StatusCode)
	}
}

func TestClientHonorsContextCancel(t *testing.T) {
	client, err := New(Options{
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			<-req.Context().Done()
			return nil, req.Context().Err()
		}),
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = client.Get(ctx, "https://example.com")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Get() error = %v, want context.Canceled", err)
	}
}

func TestClientHonorsTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client, err := New(Options{Timeout: 1 * time.Millisecond})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	resp, err := client.Get(context.Background(), server.URL)
	if resp != nil {
		resp.Body.Close()
	}
	if err == nil {
		t.Fatalf("Get() returned nil error, want timeout")
	}
}

func TestNewRejectsInvalidProxyURL(t *testing.T) {
	_, err := New(Options{ProxyURL: "://bad"})
	if err == nil || !strings.Contains(err.Error(), "proxy") {
		t.Fatalf("New() error = %v, want proxy parse error", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
