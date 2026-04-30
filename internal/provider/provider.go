package provider

import (
	"context"

	"github.com/LING71671/SurveyController-go/internal/domain"
	"github.com/LING71671/SurveyController-go/internal/engine"
)

type SurveyDefinition = domain.SurveyDefinition
type QuestionDefinition = domain.QuestionDefinition
type ProviderID = domain.ProviderID

type Capabilities struct {
	ParseHTTP       bool
	ParseBrowser    bool
	RunBrowser      bool
	SubmitHTTP      bool
	SubmitBrowser   bool
	SupportsHybrid  bool
	RequiresLoginOK bool
}

func (c Capabilities) Supports(mode engine.Mode) bool {
	normalized, err := engine.ParseMode(mode.String())
	if err != nil {
		return false
	}
	switch normalized {
	case engine.ModeHybrid:
		return c.SupportsHybrid
	case engine.ModeBrowser:
		return c.RunBrowser || c.SubmitBrowser || c.ParseBrowser
	case engine.ModeHTTP:
		return c.SubmitHTTP || c.ParseHTTP
	default:
		return false
	}
}

func (c Capabilities) CanParse(mode engine.Mode) bool {
	normalized, err := engine.ParseMode(mode.String())
	if err != nil {
		return false
	}
	switch normalized {
	case engine.ModeHybrid:
		return c.SupportsHybrid && (c.ParseHTTP || c.ParseBrowser)
	case engine.ModeBrowser:
		return c.ParseBrowser
	case engine.ModeHTTP:
		return c.ParseHTTP
	default:
		return false
	}
}

func (c Capabilities) CanSubmit(mode engine.Mode) bool {
	normalized, err := engine.ParseMode(mode.String())
	if err != nil {
		return false
	}
	switch normalized {
	case engine.ModeHybrid:
		return c.SupportsHybrid && (c.SubmitHTTP || c.SubmitBrowser)
	case engine.ModeBrowser:
		return c.SubmitBrowser
	case engine.ModeHTTP:
		return c.SubmitHTTP
	default:
		return false
	}
}

type Provider interface {
	ID() ProviderID
	MatchURL(rawURL string) bool
	Capabilities() Capabilities
	Parse(ctx context.Context, rawURL string) (SurveyDefinition, error)
}
