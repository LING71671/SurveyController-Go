package provider

import (
	"net/url"
	"strings"
)

func MatchHost(rawURL string, hosts ...string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return false
	}
	for _, candidate := range hosts {
		if host == strings.ToLower(strings.TrimSpace(candidate)) {
			return true
		}
	}
	return false
}

func MatchHostSuffix(rawURL string, suffixes ...string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := strings.ToLower(parsed.Hostname())
	if host == "" {
		return false
	}
	for _, suffix := range suffixes {
		normalized := strings.ToLower(strings.TrimPrefix(strings.TrimSpace(suffix), "."))
		if host == normalized || strings.HasSuffix(host, "."+normalized) {
			return true
		}
	}
	return false
}
