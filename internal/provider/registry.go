package provider

import (
	"fmt"
)

type Registry struct {
	providers map[ProviderID]Provider
}

func NewRegistry(providers ...Provider) (*Registry, error) {
	registry := &Registry{
		providers: map[ProviderID]Provider{},
	}
	for _, provider := range providers {
		if err := registry.Register(provider); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (r *Registry) Register(provider Provider) error {
	if provider == nil {
		return fmt.Errorf("provider is nil")
	}
	id := provider.ID()
	if id == "" {
		return fmt.Errorf("provider id is required")
	}
	if _, exists := r.providers[id]; exists {
		return fmt.Errorf("provider %q is already registered", id)
	}
	r.providers[id] = provider
	return nil
}

func (r *Registry) Get(id ProviderID) (Provider, bool) {
	provider, ok := r.providers[id]
	return provider, ok
}

func (r *Registry) MatchURL(rawURL string) (Provider, bool) {
	for _, provider := range r.providers {
		if provider.MatchURL(rawURL) {
			return provider, true
		}
	}
	return nil, false
}

func (r *Registry) IDs() []ProviderID {
	ids := make([]ProviderID, 0, len(r.providers))
	for id := range r.providers {
		ids = append(ids, id)
	}
	return ids
}
