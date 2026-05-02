package provider

import (
	"context"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/domain"
)

func TestCapabilitiesSupportsMode(t *testing.T) {
	capabilities := Capabilities{RunBrowser: true, SupportsHybrid: true}

	if !capabilities.Supports(modeStub(" Browser ")) {
		t.Fatalf("Supports(browser) = false, want true")
	}
	if capabilities.Supports(modeStub("http")) {
		t.Fatalf("Supports(http) = true, want false")
	}
	if capabilities.Supports(modeStub("magic")) {
		t.Fatalf("Supports(magic) = true, want false")
	}
}

func TestCapabilitiesCanParseAndSubmit(t *testing.T) {
	capabilities := Capabilities{
		ParseHTTP:      true,
		SubmitBrowser:  true,
		SupportsHybrid: true,
	}

	if !capabilities.CanParse(modeStub("http")) {
		t.Fatalf("CanParse(http) = false, want true")
	}
	if capabilities.CanSubmit(modeStub("http")) {
		t.Fatalf("CanSubmit(http) = true, want false")
	}
	if !capabilities.CanSubmit(modeStub("hybrid")) {
		t.Fatalf("CanSubmit(hybrid) = false, want true")
	}
}

func TestRequireSubmitCapability(t *testing.T) {
	tests := []struct {
		name         string
		provider     Provider
		mode         ModeValue
		wantErr      bool
		wantContains string
	}{
		{
			name: "http allowed",
			provider: stubProvider{
				id:           domain.ProviderWJX,
				capabilities: Capabilities{SubmitHTTP: true},
			},
			mode: modeStub("http"),
		},
		{
			name: "http requires submit http",
			provider: stubProvider{
				id:           domain.ProviderWJX,
				capabilities: Capabilities{ParseHTTP: true},
			},
			mode:         modeStub("http"),
			wantErr:      true,
			wantContains: "http submit",
		},
		{
			name: "browser requires submit browser",
			provider: stubProvider{
				id:           domain.ProviderWJX,
				capabilities: Capabilities{RunBrowser: true},
			},
			mode:         modeStub("browser"),
			wantErr:      true,
			wantContains: "browser submit",
		},
		{
			name: "hybrid allowed with submit http",
			provider: stubProvider{
				id:           domain.ProviderWJX,
				capabilities: Capabilities{SupportsHybrid: true, SubmitHTTP: true},
			},
			mode: modeStub("hybrid"),
		},
		{
			name: "hybrid requires supports hybrid",
			provider: stubProvider{
				id:           domain.ProviderWJX,
				capabilities: Capabilities{SubmitHTTP: true},
			},
			mode:         modeStub("hybrid"),
			wantErr:      true,
			wantContains: "hybrid submit",
		},
		{
			name:         "nil provider",
			mode:         modeStub("http"),
			wantErr:      true,
			wantContains: "provider is required",
		},
		{
			name: "invalid mode",
			provider: stubProvider{
				id:           domain.ProviderWJX,
				capabilities: Capabilities{SubmitHTTP: true},
			},
			mode:         modeStub("magic"),
			wantErr:      true,
			wantContains: "unsupported submit mode",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RequireSubmitCapability(tt.provider, tt.mode)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("RequireSubmitCapability() returned nil error, want %q", tt.wantContains)
				}
				if !strings.Contains(err.Error(), tt.wantContains) {
					t.Fatalf("RequireSubmitCapability() error = %v, want %q", err, tt.wantContains)
				}
				if !apperr.IsCode(err, apperr.CodeProviderUnsupported) {
					t.Fatalf("RequireSubmitCapability() error = %v, want provider_unsupported", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("RequireSubmitCapability() returned error: %v", err)
			}
		})
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
	id           ProviderID
	host         string
	capabilities Capabilities
}

type modeStub string

func (m modeStub) String() string {
	return string(m)
}

func (p stubProvider) ID() ProviderID {
	return p.id
}

func (p stubProvider) MatchURL(rawURL string) bool {
	return MatchHost(rawURL, p.host)
}

func (p stubProvider) Capabilities() Capabilities {
	if p.capabilities != (Capabilities{}) {
		return p.capabilities
	}
	return Capabilities{ParseHTTP: true}
}

func (p stubProvider) Parse(context.Context, string) (SurveyDefinition, error) {
	return SurveyDefinition{}, nil
}
