package builtin

import (
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/domain"
)

func TestDetectProvider(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want domain.ProviderID
	}{
		{name: "wjx", url: "https://www.wjx.cn/vm/example.aspx", want: domain.ProviderWJX},
		{name: "tencent", url: "https://wj.qq.com/s2/example", want: domain.ProviderTencent},
		{name: "credamo", url: "https://www.credamo.com/s/example", want: domain.ProviderCredamo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := DetectProvider(tt.url)
			if !ok || got != tt.want {
				t.Fatalf("DetectProvider(%q) = (%q, %v), want (%q, true)", tt.url, got, ok, tt.want)
			}
		})
	}
}

func TestDetectProviderRejectsUnknownURL(t *testing.T) {
	got, ok := DetectProvider("https://example.com/survey")
	if ok || got != "" {
		t.Fatalf("DetectProvider(unknown) = (%q, %v), want empty false", got, ok)
	}
}

func TestNewRegistryContainsBuiltins(t *testing.T) {
	registry, err := NewRegistry()
	if err != nil {
		t.Fatalf("NewRegistry() returned error: %v", err)
	}
	for _, id := range []domain.ProviderID{domain.ProviderWJX, domain.ProviderTencent, domain.ProviderCredamo} {
		if _, ok := registry.Get(id); !ok {
			t.Fatalf("registry missing provider %q", id)
		}
	}
}
