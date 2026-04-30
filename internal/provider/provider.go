package provider

import (
	"context"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/domain"
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

type ModeValue interface {
	String() string
}

func (c Capabilities) Supports(mode ModeValue) bool {
	normalized, ok := normalizeMode(mode)
	if !ok {
		return false
	}
	switch normalized {
	case "hybrid":
		return c.SupportsHybrid
	case "browser":
		return c.RunBrowser || c.SubmitBrowser || c.ParseBrowser
	case "http":
		return c.SubmitHTTP || c.ParseHTTP
	default:
		return false
	}
}

func (c Capabilities) CanParse(mode ModeValue) bool {
	normalized, ok := normalizeMode(mode)
	if !ok {
		return false
	}
	switch normalized {
	case "hybrid":
		return c.SupportsHybrid && (c.ParseHTTP || c.ParseBrowser)
	case "browser":
		return c.ParseBrowser
	case "http":
		return c.ParseHTTP
	default:
		return false
	}
}

func (c Capabilities) CanSubmit(mode ModeValue) bool {
	normalized, ok := normalizeMode(mode)
	if !ok {
		return false
	}
	switch normalized {
	case "hybrid":
		return c.SupportsHybrid && (c.SubmitHTTP || c.SubmitBrowser)
	case "browser":
		return c.SubmitBrowser
	case "http":
		return c.SubmitHTTP
	default:
		return false
	}
}

func normalizeMode(mode ModeValue) (string, bool) {
	if mode == nil {
		return "", false
	}
	switch normalized := strings.ToLower(strings.TrimSpace(mode.String())); normalized {
	case "":
		return "hybrid", true
	case "hybrid", "browser", "http":
		return normalized, true
	default:
		return "", false
	}
}

type Provider interface {
	ID() ProviderID
	MatchURL(rawURL string) bool
	Capabilities() Capabilities
	Parse(ctx context.Context, rawURL string) (SurveyDefinition, error)
}
