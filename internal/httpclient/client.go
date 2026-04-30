package httpclient

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

const DefaultUserAgent = "SurveyController-go/0.1"

type Options struct {
	Timeout       time.Duration
	UserAgent     string
	DefaultHeader http.Header
	ProxyURL      string
	Transport     http.RoundTripper
	Jar           http.CookieJar
}

type Client struct {
	httpClient    *http.Client
	userAgent     string
	defaultHeader http.Header
}

type RequestOptions struct {
	Method string
	URL    string
	Header http.Header
	Body   io.Reader
}

func New(options Options) (*Client, error) {
	transport := options.Transport
	if transport == nil {
		var err error
		transport, err = newTransport(options.ProxyURL)
		if err != nil {
			return nil, err
		}
	}

	jar := options.Jar
	if jar == nil {
		var err error
		jar, err = cookiejar.New(nil)
		if err != nil {
			return nil, err
		}
	}

	userAgent := strings.TrimSpace(options.UserAgent)
	if userAgent == "" {
		userAgent = DefaultUserAgent
	}

	return &Client{
		httpClient: &http.Client{
			Timeout:   options.Timeout,
			Transport: transport,
			Jar:       jar,
		},
		userAgent:     userAgent,
		defaultHeader: cloneHeader(options.DefaultHeader),
	}, nil
}

func (c *Client) Do(ctx context.Context, options RequestOptions) (*http.Response, error) {
	if c == nil {
		return nil, fmt.Errorf("client is nil")
	}
	method := strings.TrimSpace(options.Method)
	if method == "" {
		method = http.MethodGet
	}
	req, err := http.NewRequestWithContext(ctx, method, options.URL, options.Body)
	if err != nil {
		return nil, err
	}
	for key, values := range c.defaultHeader {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	for key, values := range options.Header {
		req.Header.Del(key)
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	if req.Header.Get("User-Agent") == "" {
		req.Header.Set("User-Agent", c.userAgent)
	}
	return c.httpClient.Do(req)
}

func (c *Client) Get(ctx context.Context, rawURL string) (*http.Response, error) {
	return c.Do(ctx, RequestOptions{Method: http.MethodGet, URL: rawURL})
}

func newTransport(proxyRawURL string) (http.RoundTripper, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if strings.TrimSpace(proxyRawURL) == "" {
		return transport, nil
	}
	proxyURL, err := url.Parse(proxyRawURL)
	if err != nil {
		return nil, fmt.Errorf("parse proxy url: %w", err)
	}
	transport.Proxy = http.ProxyURL(proxyURL)
	return transport, nil
}

func cloneHeader(header http.Header) http.Header {
	if len(header) == 0 {
		return nil
	}
	cloned := make(http.Header, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}
