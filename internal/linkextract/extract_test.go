package linkextract

import (
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/domain"
)

func TestExtractFindsSupportedSurveyLinks(t *testing.T) {
	text := `扫码文本：
问卷星 https://www.wjx.cn/vm/example.aspx?from=qr。
腾讯 https://wj.qq.com/s2/123456/hash?source=copy
Credamo https://www.credamo.com/answer.html#/s/demo）`

	got := Extract(text)
	if len(got) != 3 {
		t.Fatalf("Extract() returned %d candidates, want 3: %+v", len(got), got)
	}
	assertCandidate(t, got[0], domain.ProviderWJX, "https://www.wjx.cn/vm/example.aspx?from=qr")
	assertCandidate(t, got[1], domain.ProviderTencent, "https://wj.qq.com/s2/123456/hash?source=copy")
	assertCandidate(t, got[2], domain.ProviderCredamo, "https://www.credamo.com/answer.html#/s/demo")
}

func TestExtractSkipsUnsupportedAndInvalidURLs(t *testing.T) {
	text := `noise https://example.com/survey https:// / not-url`

	got := Extract(text)
	if len(got) != 0 {
		t.Fatalf("Extract() = %+v, want no supported candidates", got)
	}
}

func TestExtractUnescapesHTMLAndDeduplicates(t *testing.T) {
	text := `href="https://www.wjx.cn/vm/example.aspx?x=1&amp;y=2" copy https://www.wjx.cn/vm/example.aspx?x=1&y=2`

	got := Extract(text)
	if len(got) != 1 {
		t.Fatalf("Extract() returned %d candidates, want 1: %+v", len(got), got)
	}
	assertCandidate(t, got[0], domain.ProviderWJX, "https://www.wjx.cn/vm/example.aspx?x=1&y=2")
}

func TestFirstReturnsFirstCandidate(t *testing.T) {
	got, ok := First(`unsupported https://example.com then https://wj.qq.com/s2/123/hash`)
	if !ok {
		t.Fatal("First() returned ok=false, want true")
	}
	assertCandidate(t, got, domain.ProviderTencent, "https://wj.qq.com/s2/123/hash")
}

func TestFirstReturnsFalseForEmptyInput(t *testing.T) {
	if got, ok := First("   "); ok {
		t.Fatalf("First(empty) = (%+v, true), want false", got)
	}
}

func assertCandidate(t *testing.T, got Candidate, wantProvider domain.ProviderID, wantURL string) {
	t.Helper()
	if got.Provider != wantProvider || got.URL != wantURL {
		t.Fatalf("candidate = %+v, want provider %q url %q", got, wantProvider, wantURL)
	}
	if got.Raw == "" {
		t.Fatal("candidate raw is empty")
	}
}
