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
