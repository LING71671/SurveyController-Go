package answer

import (
	"math"
	"math/rand"
	"strings"
	"testing"
)

func TestNormalizeWeights(t *testing.T) {
	got, err := NormalizeWeights([]OptionWeight{
		{OptionID: "a", Weight: 2},
		{OptionID: "b", Weight: 1},
	})
	if err != nil {
		t.Fatalf("NormalizeWeights() returned error: %v", err)
	}

	if math.Abs(got[0].Weight-0.666666) > 0.0001 {
		t.Fatalf("first weight = %f, want about 0.666666", got[0].Weight)
	}
	if math.Abs(got[1].Weight-0.333333) > 0.0001 {
		t.Fatalf("second weight = %f, want about 0.333333", got[1].Weight)
	}
}

func TestNormalizeWeightsUsesEvenWeightsWhenTotalIsZero(t *testing.T) {
	got, err := NormalizeWeights([]OptionWeight{
		{OptionID: "a"},
		{OptionID: "b"},
	})
	if err != nil {
		t.Fatalf("NormalizeWeights() returned error: %v", err)
	}
	if got[0].Weight != 0.5 || got[1].Weight != 0.5 {
		t.Fatalf("weights = %+v, want even weights", got)
	}
}

func TestNormalizeWeightsRejectsInvalidInput(t *testing.T) {
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
			_, err := NormalizeWeights(tt.weights)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("NormalizeWeights() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestPickOneUsesWeightedChoice(t *testing.T) {
	got, err := PickOne(rand.New(rand.NewSource(1)), []OptionWeight{
		{OptionID: "a", Weight: 0},
		{OptionID: "b", Weight: 1},
	})
	if err != nil {
		t.Fatalf("PickOne() returned error: %v", err)
	}
	if got != "b" {
		t.Fatalf("PickOne() = %q, want b", got)
	}
}

func TestPickManyHonorsMinMaxAndUniqueness(t *testing.T) {
	got, err := PickMany(rand.New(rand.NewSource(7)), []OptionWeight{
		{OptionID: "a", Weight: 1},
		{OptionID: "b", Weight: 1},
		{OptionID: "c", Weight: 1},
	}, SelectionRule{Min: 2, Max: 2})
	if err != nil {
		t.Fatalf("PickMany() returned error: %v", err)
	}
	if len(got.OptionIDs) != 2 {
		t.Fatalf("len(OptionIDs) = %d, want 2", len(got.OptionIDs))
	}
	if got.OptionIDs[0] == got.OptionIDs[1] {
		t.Fatalf("OptionIDs = %+v, want unique options", got.OptionIDs)
	}
}

func TestPickManyRejectsInvalidRule(t *testing.T) {
	_, err := PickMany(rand.New(rand.NewSource(1)), []OptionWeight{
		{OptionID: "a", Weight: 1},
	}, SelectionRule{Min: 2, Max: 1})
	if err == nil || !strings.Contains(err.Error(), "min") {
		t.Fatalf("PickMany() error = %v, want min rule error", err)
	}
}
