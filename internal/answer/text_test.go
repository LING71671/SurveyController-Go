package answer

import (
	"math/rand"
	"strings"
	"testing"
)

func TestRandomText(t *testing.T) {
	got, err := RandomText(rand.New(rand.NewSource(1)), TextRule{
		Words:     []string{"alpha", "beta", "gamma"},
		MinWords:  2,
		MaxWords:  2,
		Separator: "-",
	})
	if err != nil {
		t.Fatalf("RandomText() returned error: %v", err)
	}
	parts := strings.Split(got, "-")
	if len(parts) != 2 {
		t.Fatalf("RandomText() = %q, want two words", got)
	}
}

func TestRandomTextRejectsInvalidRules(t *testing.T) {
	tests := []struct {
		name string
		rule TextRule
		want string
	}{
		{name: "words", rule: TextRule{}, want: "words"},
		{name: "blank words", rule: TextRule{Words: []string{" ", "\t"}}, want: "words"},
		{name: "range", rule: TextRule{Words: []string{"a"}, MinWords: 3, MaxWords: 2}, want: "min words"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RandomText(rand.New(rand.NewSource(1)), tt.rule)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("RandomText() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestRandomDigits(t *testing.T) {
	got, err := RandomDigits(rand.New(rand.NewSource(1)), DigitsRule{Length: 8, Prefix: "42"})
	if err != nil {
		t.Fatalf("RandomDigits() returned error: %v", err)
	}
	if len(got) != 8 || !strings.HasPrefix(got, "42") {
		t.Fatalf("RandomDigits() = %q, want 8 digits with prefix 42", got)
	}
	for _, char := range got {
		if char < '0' || char > '9' {
			t.Fatalf("RandomDigits() = %q, want digits only", got)
		}
	}
}

func TestRandomDigitsRejectsInvalidRules(t *testing.T) {
	tests := []struct {
		name string
		rule DigitsRule
		want string
	}{
		{name: "length", rule: DigitsRule{}, want: "length"},
		{name: "prefix chars", rule: DigitsRule{Length: 4, Prefix: "ab"}, want: "prefix"},
		{name: "prefix too long", rule: DigitsRule{Length: 2, Prefix: "123"}, want: "prefix length"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RandomDigits(rand.New(rand.NewSource(1)), tt.rule)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("RandomDigits() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestRandomPhoneLike(t *testing.T) {
	got, err := RandomPhoneLike(rand.New(rand.NewSource(1)), PhoneRule{Prefixes: []string{"177"}})
	if err != nil {
		t.Fatalf("RandomPhoneLike() returned error: %v", err)
	}
	if len(got) != 11 || !strings.HasPrefix(got, "177") {
		t.Fatalf("RandomPhoneLike() = %q, want 11 digits with prefix 177", got)
	}
}

func TestRandomPhoneLikeUsesDefaultPrefixes(t *testing.T) {
	got, err := RandomPhoneLike(rand.New(rand.NewSource(1)), PhoneRule{})
	if err != nil {
		t.Fatalf("RandomPhoneLike(default) returned error: %v", err)
	}
	if len(got) != 11 {
		t.Fatalf("RandomPhoneLike(default) = %q, want 11 digits", got)
	}
}

func TestRandomPhoneLikeRejectsInvalidPrefix(t *testing.T) {
	_, err := RandomPhoneLike(rand.New(rand.NewSource(1)), PhoneRule{Prefixes: []string{"phone"}})
	if err == nil || !strings.Contains(err.Error(), "prefix") {
		t.Fatalf("RandomPhoneLike() error = %v, want prefix error", err)
	}
}

func TestRandomTemplateText(t *testing.T) {
	got, err := RandomTemplateText(rand.New(rand.NewSource(1)), TemplateRule{
		Template: "from {city} as {role}",
		Slots: map[string][]string{
			"city": {"shanghai", "hangzhou"},
			"role": {"student"},
		},
	})
	if err != nil {
		t.Fatalf("RandomTemplateText() returned error: %v", err)
	}
	if !strings.HasPrefix(got, "from ") || strings.Contains(got, "{") {
		t.Fatalf("RandomTemplateText() = %q, want rendered template", got)
	}
}

func TestRandomTemplateTextRejectsInvalidRules(t *testing.T) {
	tests := []struct {
		name string
		rule TemplateRule
		want string
	}{
		{name: "template", rule: TemplateRule{}, want: "template"},
		{name: "slot", rule: TemplateRule{Template: "hello {name}"}, want: "slot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := RandomTemplateText(rand.New(rand.NewSource(1)), tt.rule)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("RandomTemplateText() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestRandomInt(t *testing.T) {
	got, err := RandomInt(rand.New(rand.NewSource(1)), 10, 20)
	if err != nil {
		t.Fatalf("RandomInt() returned error: %v", err)
	}
	if got < 10 || got > 20 {
		t.Fatalf("RandomInt() = %d, want within [10,20]", got)
	}
}

func TestRandomIntRejectsInvalidRange(t *testing.T) {
	_, err := RandomInt(rand.New(rand.NewSource(1)), 5, 4)
	if err == nil || !strings.Contains(err.Error(), "min") {
		t.Fatalf("RandomInt() error = %v, want min range error", err)
	}
}
