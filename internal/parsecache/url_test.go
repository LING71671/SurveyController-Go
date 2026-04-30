package parsecache

import "testing"

func TestNormalizeURL(t *testing.T) {
	got, err := NormalizeURL("HTTPS://Example.COM:443/path?b=2&a=1#fragment")
	if err != nil {
		t.Fatalf("NormalizeURL() returned error: %v", err)
	}
	want := "https://example.com/path?a=1&b=2"
	if got != want {
		t.Fatalf("NormalizeURL() = %q, want %q", got, want)
	}
}

func TestNormalizeURLSortsRepeatedQueryValues(t *testing.T) {
	got, err := NormalizeURL("https://example.com/path?tag=b&tag=a")
	if err != nil {
		t.Fatalf("NormalizeURL() returned error: %v", err)
	}
	want := "https://example.com/path?tag=a&tag=b"
	if got != want {
		t.Fatalf("NormalizeURL() = %q, want %q", got, want)
	}
}

func TestNormalizeURLRequiresSchemeAndHost(t *testing.T) {
	if _, err := NormalizeURL("/relative"); err == nil {
		t.Fatal("NormalizeURL(relative) returned nil error, want failure")
	}
}

func TestFingerprintMatchesEquivalentURLs(t *testing.T) {
	a, err := Fingerprint("https://example.com:443/path?b=2&a=1")
	if err != nil {
		t.Fatalf("Fingerprint(a) returned error: %v", err)
	}
	b, err := Fingerprint("HTTPS://EXAMPLE.COM/path?a=1&b=2#ignored")
	if err != nil {
		t.Fatalf("Fingerprint(b) returned error: %v", err)
	}
	if a != b {
		t.Fatalf("fingerprints differ for equivalent urls")
	}
}
