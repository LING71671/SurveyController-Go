package linkextract

import (
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/provider/builtin"
)

var httpURLRE = regexp.MustCompile("(?i)https?://[^\\s<>\"'`]+")

type Candidate struct {
	Raw      string            `json:"raw"`
	URL      string            `json:"url"`
	Provider domain.ProviderID `json:"provider"`
}

func Extract(text string) []Candidate {
	text = html.UnescapeString(strings.TrimSpace(text))
	if text == "" {
		return nil
	}

	matches := httpURLRE.FindAllString(text, -1)
	candidates := make([]Candidate, 0, len(matches))
	seen := map[string]struct{}{}
	for _, match := range matches {
		raw := strings.TrimSpace(match)
		normalized, ok := normalizeURL(raw)
		if !ok {
			continue
		}
		providerID, ok := builtin.DetectProvider(normalized)
		if !ok {
			continue
		}
		key := strings.ToLower(normalized)
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}
		candidates = append(candidates, Candidate{
			Raw:      raw,
			URL:      normalized,
			Provider: providerID,
		})
	}
	return candidates
}

func First(text string) (Candidate, bool) {
	candidates := Extract(text)
	if len(candidates) == 0 {
		return Candidate{}, false
	}
	return candidates[0], true
}

func normalizeURL(raw string) (string, bool) {
	candidate := strings.TrimSpace(raw)
	for candidate != "" {
		candidate = strings.TrimRight(candidate, trailingURLPunctuation)
		parsed, err := url.Parse(candidate)
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != "" {
			return parsed.String(), true
		}
		next := strings.TrimRight(candidate, trailingURLPunctuation)
		if next == candidate {
			return "", false
		}
		candidate = next
	}
	return "", false
}

const trailingURLPunctuation = ".,;:!?)]}>\"'`，。；：！？）】》、"
