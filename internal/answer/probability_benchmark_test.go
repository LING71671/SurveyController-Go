package answer

import (
	"math/rand"
	"testing"
)

var benchmarkPickedOption string

func BenchmarkPickOne(b *testing.B) {
	weights := []OptionWeight{
		{OptionID: "a", Weight: 1},
		{OptionID: "b", Weight: 2},
		{OptionID: "c", Weight: 3},
		{OptionID: "d", Weight: 4},
	}
	rng := rand.New(rand.NewSource(1))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optionID, err := PickOne(rng, weights)
		if err != nil {
			b.Fatalf("PickOne() returned error: %v", err)
		}
		benchmarkPickedOption = optionID
	}
}

func BenchmarkPickOneEvenWeights(b *testing.B) {
	weights := []OptionWeight{
		{OptionID: "a"},
		{OptionID: "b"},
		{OptionID: "c"},
		{OptionID: "d"},
	}
	rng := rand.New(rand.NewSource(1))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optionID, err := PickOne(rng, weights)
		if err != nil {
			b.Fatalf("PickOne() returned error: %v", err)
		}
		benchmarkPickedOption = optionID
	}
}

func BenchmarkWeightedPickerPick(b *testing.B) {
	picker, err := NewWeightedPicker([]OptionWeight{
		{OptionID: "a", Weight: 1},
		{OptionID: "b", Weight: 2},
		{OptionID: "c", Weight: 3},
		{OptionID: "d", Weight: 4},
	})
	if err != nil {
		b.Fatalf("NewWeightedPicker() returned error: %v", err)
	}
	rng := rand.New(rand.NewSource(1))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optionID, err := picker.Pick(rng)
		if err != nil {
			b.Fatalf("Pick() returned error: %v", err)
		}
		benchmarkPickedOption = optionID
	}
}

func BenchmarkWeightedPickerPickEvenWeights(b *testing.B) {
	picker, err := NewWeightedPicker([]OptionWeight{
		{OptionID: "a"},
		{OptionID: "b"},
		{OptionID: "c"},
		{OptionID: "d"},
	})
	if err != nil {
		b.Fatalf("NewWeightedPicker() returned error: %v", err)
	}
	rng := rand.New(rand.NewSource(1))

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optionID, err := picker.Pick(rng)
		if err != nil {
			b.Fatalf("Pick() returned error: %v", err)
		}
		benchmarkPickedOption = optionID
	}
}
