package provider

import (
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/apperr"
)

func RequireSubmitCapability(p Provider, mode ModeValue) error {
	if p == nil {
		return apperr.New(apperr.CodeProviderUnsupported, "provider is required")
	}

	normalized, ok := normalizeMode(mode)
	if !ok {
		return apperr.New(apperr.CodeProviderUnsupported, fmt.Sprintf("unsupported submit mode %q", modeString(mode)))
	}

	capabilities := p.Capabilities()
	if capabilities.CanSubmit(modeStringValue(normalized)) {
		return nil
	}

	return apperr.New(
		apperr.CodeProviderUnsupported,
		fmt.Sprintf("provider %q does not support %s submit capability", p.ID(), normalized),
	)
}

type modeStringValue string

func (m modeStringValue) String() string {
	return string(m)
}

func modeString(mode ModeValue) string {
	if mode == nil {
		return ""
	}
	return strings.TrimSpace(mode.String())
}
