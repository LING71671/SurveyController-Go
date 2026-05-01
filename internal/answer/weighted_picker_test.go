package answer

import (
	"math/rand"
	"strings"
	"testing"
)

func TestWeightedPickerPick(t *testing.T) {
	picker, err := NewWeightedPicker([]OptionWeight{
		{OptionID: "a", Weight: 0},
		{OptionID: "b", Weight: 1},
	})
	if err != nil {
		t.Fatalf("NewWeightedPicker() returned error: %v", err)
	}

	got, err := picker.Pick(rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("Pick() returned error: %v", err)
	}
	if got != "b" {
		t.Fatalf("Pick() = %q, want b", got)
	}
	if picker.Len() != 2 {
		t.Fatalf("Len() = %d, want 2", picker.Len())
	}
}

func TestWeightedPickerEvenWeights(t *testing.T) {
	picker, err := NewWeightedPicker([]OptionWeight{
		{OptionID: "a"},
		{OptionID: "b"},
		{OptionID: "c"},
	})
	if err != nil {
		t.Fatalf("NewWeightedPicker() returned error: %v", err)
	}

	got, err := picker.Pick(rand.New(rand.NewSource(2)))
	if err != nil {
		t.Fatalf("Pick() returned error: %v", err)
	}
	if got == "" {
		t.Fatal("Pick() returned empty option id")
	}
}

func TestWeightedPickerCopiesWeights(t *testing.T) {
	weights := []OptionWeight{
		{OptionID: "a", Weight: 0},
		{OptionID: "b", Weight: 1},
	}
	picker, err := NewWeightedPicker(weights)
	if err != nil {
		t.Fatalf("NewWeightedPicker() returned error: %v", err)
	}
	weights[1].OptionID = "mutated"
	weights[1].Weight = 0

	got, err := picker.Pick(rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("Pick() returned error: %v", err)
	}
	if got != "b" {
		t.Fatalf("Pick() = %q, want copied b", got)
	}
}

func TestWeightedPickerRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		weights []OptionWeight
		want    string
	}{
		{name: "empty", want: "required"},
		{name: "missing id", weights: []OptionWeight{{Weight: 1}}, want: "option id"},
		{name: "negative", weights: []OptionWeight{{OptionID: "a", Weight: -1}}, want: "negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewWeightedPicker(tt.weights)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("NewWeightedPicker() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestWeightedPickerRejectsNilRNG(t *testing.T) {
	picker, err := NewWeightedPicker([]OptionWeight{{OptionID: "a", Weight: 1}})
	if err != nil {
		t.Fatalf("NewWeightedPicker() returned error: %v", err)
	}

	if _, err := picker.Pick(nil); err == nil || !strings.Contains(err.Error(), "rng") {
		t.Fatalf("Pick(nil) error = %v, want rng error", err)
	}
}
