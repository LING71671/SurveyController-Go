package browser

import (
	"context"
	"fmt"
	"sync"
)

type FakePool struct {
	mu       sync.Mutex
	closed   bool
	sessions []*FakeSession
}

func NewFakePool() *FakePool {
	return &FakePool{}
}

func (p *FakePool) NewSession(ctx context.Context, options SessionOptions) (BrowserSession, error) {
	if err := MapContextError(ctx.Err()); err != nil {
		return nil, err
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, fmt.Errorf("browser pool is closed")
	}
	session := &FakeSession{
		page: &FakePage{},
	}
	p.sessions = append(p.sessions, session)
	return session, nil
}

func (p *FakePool) Close(ctx context.Context) error {
	if err := MapContextError(ctx.Err()); err != nil {
		return err
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil
	}
	p.closed = true
	for _, session := range p.sessions {
		if err := session.Close(ctx); err != nil {
			return err
		}
	}
	return nil
}

type FakeSession struct {
	mu     sync.Mutex
	closed bool
	page   *FakePage
}

func (s *FakeSession) Page() Page {
	return s.page
}

func (s *FakeSession) FakePage() *FakePage {
	return s.page
}

func (s *FakeSession) Close(ctx context.Context) error {
	if err := MapContextError(ctx.Err()); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.closed = true
	return nil
}

func (s *FakeSession) Closed() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closed
}

type FakePage struct {
	mu      sync.Mutex
	html    string
	calls   []Call
	results map[string]string
}

type Call struct {
	Name     string
	Selector string
	Value    string
}

func (p *FakePage) SetHTML(html string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.html = html
}

func (p *FakePage) SetEvaluateResult(script string, result string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.results == nil {
		p.results = map[string]string{}
	}
	p.results[script] = result
}

func (p *FakePage) Navigate(ctx context.Context, rawURL string) error {
	if err := MapContextError(ctx.Err()); err != nil {
		return err
	}
	p.record(Call{Name: "navigate", Value: rawURL})
	return nil
}

func (p *FakePage) Click(ctx context.Context, selector string) error {
	if err := MapContextError(ctx.Err()); err != nil {
		return err
	}
	if err := ValidateSelector(selector); err != nil {
		return err
	}
	p.record(Call{Name: "click", Selector: selector})
	return nil
}

func (p *FakePage) Fill(ctx context.Context, selector string, value string) error {
	if err := MapContextError(ctx.Err()); err != nil {
		return err
	}
	if err := ValidateSelector(selector); err != nil {
		return err
	}
	p.record(Call{Name: "fill", Selector: selector, Value: value})
	return nil
}

func (p *FakePage) HTML(ctx context.Context) (string, error) {
	if err := MapContextError(ctx.Err()); err != nil {
		return "", err
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.html, nil
}

func (p *FakePage) Evaluate(ctx context.Context, script string) (string, error) {
	if err := MapContextError(ctx.Err()); err != nil {
		return "", err
	}
	p.mu.Lock()
	defer p.mu.Unlock()

	p.calls = append(p.calls, Call{Name: "evaluate", Value: script})
	return p.results[script], nil
}

func (p *FakePage) Calls() []Call {
	p.mu.Lock()
	defer p.mu.Unlock()

	return append([]Call(nil), p.calls...)
}

func (p *FakePage) record(call Call) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.calls = append(p.calls, call)
}
