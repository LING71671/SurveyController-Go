package parsecache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strings"
)

func NormalizeURL(rawURL string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("url must include scheme and host")
	}

	parsed.Scheme = strings.ToLower(parsed.Scheme)
	host := strings.ToLower(parsed.Hostname())
	port := parsed.Port()
	switch {
	case port == "":
		parsed.Host = host
	case (parsed.Scheme == "http" && port == "80") || (parsed.Scheme == "https" && port == "443"):
		parsed.Host = host
	default:
		parsed.Host = net.JoinHostPort(host, port)
	}

	parsed.Fragment = ""
	parsed.RawQuery = normalizeQuery(parsed.Query())
	if parsed.Path == "" {
		parsed.Path = "/"
	}
	return parsed.String(), nil
}

func Fingerprint(rawURL string) (string, error) {
	normalized, err := NormalizeURL(rawURL)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:]), nil
}

func normalizeQuery(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
		sort.Strings(values[key])
	}
	sort.Strings(keys)

	ordered := url.Values{}
	for _, key := range keys {
		for _, value := range values[key] {
			ordered.Add(key, value)
		}
	}
	return ordered.Encode()
}
