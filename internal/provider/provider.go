package provider

import (
	"context"

	"github.com/hungryM0/SurveyController-go/internal/engine"
)

type SurveyDefinition struct {
	Provider  string
	Title     string
	Questions []Question
}

type Question struct {
	Number  int
	Title   string
	Type    string
	Options []string
	Rows    []string
}

type Capabilities struct {
	Engines []engine.Mode
}

func (c Capabilities) Supports(mode engine.Mode) bool {
	normalized, err := engine.ParseMode(mode.String())
	if err != nil {
		return false
	}
	for _, candidate := range c.Engines {
		if candidate == normalized {
			return true
		}
	}
	return false
}

type Provider interface {
	Name() string
	MatchURL(rawURL string) bool
	Capabilities() Capabilities
	Parse(ctx context.Context, rawURL string) (SurveyDefinition, error)
}
