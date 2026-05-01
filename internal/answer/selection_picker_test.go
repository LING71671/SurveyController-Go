package answer

import (
	"math/rand"
	"strings"
	"testing"
)

func TestSelectionPickerPick(t *testing.T) {
	picker, err := NewSelectionPicker([]OptionWeight{
		{OptionID: "a", Weight: 1},
		{OptionID: "b", Weight: 1},
		{OptionID: "c", Weight: 1},
	}, SelectionRule{Min: 2, Max: 2})
	if err != nil {
		t.Fatalf("NewSelectionPicker() returned error: %v", err)
	}

	got, err := picker.Pick(rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("Pick() returned error: %v", err)
	}
	if len(got.OptionIDs) != 2 {
		t.Fatalf("len(OptionIDs) = %d, want 2: %+v", len(got.OptionIDs), got.OptionIDs)
	}
	if got.OptionIDs[0] > got.OptionIDs[1] {
		t.Fatalf("OptionIDs = %+v, want sorted IDs", got.OptionIDs)
	}
	if picker.Len() != 3 {
		t.Fatalf("Len() = %d, want 3", picker.Len())
	}
	if picker.Rule() != (SelectionRule{Min: 2, Max: 2}) {
		t.Fatalf("Rule() = %+v, want min/max 2", picker.Rule())
	}
}

func TestSelectionPickerEvenWeights(t *testing.T) {
	picker, err := NewSelectionPicker([]OptionWeight{
		{OptionID: "a"},
		{OptionID: "b"},
		{OptionID: "c"},
	}, SelectionRule{Min: 1, Max: 1})
	if err != nil {
		t.Fatalf("NewSelectionPicker() returned error: %v", err)
	}

	got, err := picker.Pick(rand.New(rand.NewSource(2)))
	if err != nil {
		t.Fatalf("Pick() returned error: %v", err)
	}
	if len(got.OptionIDs) != 1 || got.OptionIDs[0] == "" {
		t.Fatalf("OptionIDs = %+v, want one non-empty option", got.OptionIDs)
	}
}

func TestSelectionPickerCopiesWeights(t *testing.T) {
	weights := []OptionWeight{
		{OptionID: "a", Weight: 0},
		{OptionID: "b", Weight: 1},
	}
	picker, err := NewSelectionPicker(weights, SelectionRule{Min: 1, Max: 1})
	if err != nil {
		t.Fatalf("NewSelectionPicker() returned error: %v", err)
	}
	weights[1].OptionID = "mutated"
	weights[1].Weight = 0

	got, err := picker.Pick(rand.New(rand.NewSource(1)))
	if err != nil {
		t.Fatalf("Pick() returned error: %v", err)
	}
	if len(got.OptionIDs) != 1 || got.OptionIDs[0] != "b" {
		t.Fatalf("OptionIDs = %+v, want copied b", got.OptionIDs)
	}
}

func TestSelectionPickerRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name    string
		weights []OptionWeight
		rule    SelectionRule
		want    string
	}{
		{name: "empty", want: "required"},
		{name: "missing id", weights: []OptionWeight{{Weight: 1}}, want: "option id"},
		{name: "negative", weights: []OptionWeight{{OptionID: "a", Weight: -1}}, want: "negative"},
		{name: "invalid rule", weights: []OptionWeight{{OptionID: "a", Weight: 1}}, rule: SelectionRule{Min: 2, Max: 1}, want: "min must not be greater"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewSelectionPicker(tt.weights, tt.rule)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("NewSelectionPicker() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSelectionPickerRejectsNilRNG(t *testing.T) {
	picker, err := NewSelectionPicker([]OptionWeight{{OptionID: "a", Weight: 1}}, SelectionRule{})
	if err != nil {
		t.Fatalf("NewSelectionPicker() returned error: %v", err)
	}

	if _, err := picker.Pick(nil); err == nil || !strings.Contains(err.Error(), "rng") {
		t.Fatalf("Pick(nil) error = %v, want rng error", err)
	}
}
