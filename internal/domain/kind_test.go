package domain

import "testing"

func TestParseProviderID(t *testing.T) {
	tests := []struct {
		raw  string
		want ProviderID
	}{
		{raw: "wjx", want: ProviderWJX},
		{raw: " Tencent ", want: ProviderTencent},
		{raw: "credamo", want: ProviderCredamo},
	}

	for _, tt := range tests {
		got, err := ParseProviderID(tt.raw)
		if err != nil {
			t.Fatalf("ParseProviderID(%q) returned error: %v", tt.raw, err)
		}
		if got != tt.want {
			t.Fatalf("ParseProviderID(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestParseProviderIDRejectsUnknown(t *testing.T) {
	if _, err := ParseProviderID("other"); err == nil {
		t.Fatal("ParseProviderID(other) returned nil error, want failure")
	}
}

func TestParseQuestionKind(t *testing.T) {
	tests := []struct {
		raw  string
		want QuestionKind
	}{
		{raw: "single", want: QuestionKindSingle},
		{raw: " Multiple ", want: QuestionKindMultiple},
		{raw: "dropdown", want: QuestionKindDropdown},
		{raw: "text", want: QuestionKindText},
		{raw: "textarea", want: QuestionKindTextarea},
		{raw: "rating", want: QuestionKindRating},
		{raw: "matrix", want: QuestionKindMatrix},
		{raw: "ranking", want: QuestionKindRanking},
	}

	for _, tt := range tests {
		got, err := ParseQuestionKind(tt.raw)
		if err != nil {
			t.Fatalf("ParseQuestionKind(%q) returned error: %v", tt.raw, err)
		}
		if got != tt.want {
			t.Fatalf("ParseQuestionKind(%q) = %q, want %q", tt.raw, got, tt.want)
		}
	}
}

func TestParseQuestionKindRejectsUnknown(t *testing.T) {
	if _, err := ParseQuestionKind("captcha"); err == nil {
		t.Fatal("ParseQuestionKind(captcha) returned nil error, want failure")
	}
}
