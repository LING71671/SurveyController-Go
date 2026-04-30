package provider

import (
	"testing"

	"github.com/LING71671/SurveyController-go/internal/engine"
)

func TestCapabilitiesSupportsMode(t *testing.T) {
	capabilities := Capabilities{
		Engines: []engine.Mode{engine.ModeBrowser, engine.ModeHybrid},
	}

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
