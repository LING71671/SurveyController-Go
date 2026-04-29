package engine

import (
	"fmt"
	"strings"
)

type Mode string

const (
	ModeHybrid  Mode = "hybrid"
	ModeBrowser Mode = "browser"
	ModeHTTP    Mode = "http"
)

func ParseMode(raw string) (Mode, error) {
	switch Mode(strings.ToLower(strings.TrimSpace(raw))) {
	case "", ModeHybrid:
		return ModeHybrid, nil
	case ModeBrowser:
		return ModeBrowser, nil
	case ModeHTTP:
		return ModeHTTP, nil
	default:
		return "", fmt.Errorf("unsupported engine mode %q", raw)
	}
}

func (m Mode) String() string {
	if m == "" {
		return string(ModeHybrid)
	}
	return string(m)
}

func (m Mode) Valid() bool {
	_, err := ParseMode(m.String())
	return err == nil
}
