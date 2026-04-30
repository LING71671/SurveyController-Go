package provider

import (
	"context"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/domain"
	"github.com/LING71671/SurveyController-go/internal/engine"
)

func TestCapabilitiesSupportsMode(t *testing.T) {
	capabilities := Capabilities{RunBrowser: true, SupportsHybrid: true}

	if !capabilities.Supports(engine.Mode(" Browser ")) {
		t.Fatalf("Supports(browser) = false, want true")
	}
	if capabilities.Supports(engine.ModeHTTP) {
		t.Fatalf("Supports(http) = true, want false")
	}
	if capabilities.Supports(engine.Mode("magic")) {
		t.Fatalf("Supports(magic) = true, want false")
	}
}

func TestCapabilitiesCanParseAndSubmit(t *testing.T) {
	capabilities := Capabilities{
		ParseHTTP:      true,
		SubmitBrowser:  true,
		SupportsHybrid: true,
	}

	if !capabilities.CanParse(engine.ModeHTTP) {
		t.Fatalf("CanParse(http) = false, want true")
	}
	if capabilities.CanSubmit(engine.ModeHTTP) {
		t.Fatalf("CanSubmit(http) = true, want false")
	}
	if !capabilities.CanSubmit(engine.ModeHybrid) {
		t.Fatalf("CanSubmit(hybrid) = false, want true")
	}
}

func TestRegistryGetAndMatchURL(t *testing.T) {
	provider := stubProvider{
		id:   domain.ProviderWJX,
		host: "www.wjx.cn",
	}
	registry, err := NewRegistry(provider)
	if err != nil {
		t.Fatalf("NewRegistry() returned error: %v", err)
	}

	got, ok := registry.Get(domain.ProviderWJX)
	if !ok || got.ID() != domain.ProviderWJX {
		t.Fatalf("Get(wjx) = (%v, %v), want provider", got, ok)
	}

	got, ok = registry.MatchURL("https://www.wjx.cn/vm/demo.aspx")
	if !ok || got.ID() != domain.ProviderWJX {
		t.Fatalf("MatchURL(wjx url) = (%v, %v), want provider", got, ok)
	}
}

func TestRegistryRejectsDuplicateProvider(t *testing.T) {
	_, err := NewRegistry(
		stubProvider{id: domain.ProviderWJX},
		stubProvider{id: domain.ProviderWJX},
	)
	if err == nil {
		t.Fatal("NewRegistry(duplicate) returned nil error, want failure")
	}
}

func TestMatchHostAndSuffix(t *testing.T) {
	if !MatchHost("https://wj.qq.com/s2/123", "wj.qq.com") {
		t.Fatalf("MatchHost() = false, want true")
	}
	if !MatchHostSuffix("https://sub.wjx.cn/vm/demo.aspx", "wjx.cn") {
		t.Fatalf("MatchHostSuffix() = false, want true")
	}
	if MatchHostSuffix("not a url", "wjx.cn") {
		t.Fatalf("MatchHostSuffix(invalid) = true, want false")
	}
}

type stubProvider struct {
	id   ProviderID
	host string
}

func (p stubProvider) ID() ProviderID {
	return p.id
}

func (p stubProvider) MatchURL(rawURL string) bool {
	return MatchHost(rawURL, p.host)
}

func (p stubProvider) Capabilities() Capabilities {
	return Capabilities{ParseHTTP: true}
}

func (p stubProvider) Parse(context.Context, string) (SurveyDefinition, error) {
	return SurveyDefinition{}, nil
}
