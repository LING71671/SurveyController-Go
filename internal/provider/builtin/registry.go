package builtin

import (
	"github.com/LING71671/SurveyController-Go/internal/provider"
	"github.com/LING71671/SurveyController-Go/internal/provider/credamo"
	"github.com/LING71671/SurveyController-Go/internal/provider/tencent"
	"github.com/LING71671/SurveyController-Go/internal/provider/wjx"
)

func NewRegistry() (*provider.Registry, error) {
	return provider.NewRegistry(
		wjx.Provider{},
		tencent.Provider{},
		credamo.Provider{},
	)
}

func DetectProvider(rawURL string) (provider.ProviderID, bool) {
	registry, err := NewRegistry()
	if err != nil {
		return "", false
	}
	matched, ok := registry.MatchURL(rawURL)
	if !ok {
		return "", false
	}
	return matched.ID(), true
}
