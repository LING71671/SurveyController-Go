package answer

import (
	"fmt"
	"math/rand"
)

type OptionWeight struct {
	OptionID string
	Weight   float64
}

type SelectionRule struct {
	Min int
	Max int
}

type SelectionResult struct {
	OptionIDs []string
}

func NormalizeWeights(weights []OptionWeight) ([]OptionWeight, error) {
	total, err := validateAndSumWeights(weights)
	if err != nil {
		return nil, err
	}

	normalized := append([]OptionWeight(nil), weights...)
	if total == 0 {
		even := 1.0 / float64(len(normalized))
		for i := range normalized {
			normalized[i].Weight = even
		}
		return normalized, nil
	}

	for i := range normalized {
		normalized[i].Weight = normalized[i].Weight / total
	}
	return normalized, nil
}

func PickOne(rng *rand.Rand, weights []OptionWeight) (string, error) {
	if rng == nil {
		return "", fmt.Errorf("rng is required")
	}
	total, err := validateAndSumWeights(weights)
	if err != nil {
		return "", err
	}
	if total == 0 {
		return pickEven(rng, weights), nil
	}

	point := rng.Float64() * total
	accumulated := 0.0
	for _, item := range weights {
		accumulated += item.Weight
		if point <= accumulated {
			return item.OptionID, nil
		}
	}
	return weights[len(weights)-1].OptionID, nil
}

func validateAndSumWeights(weights []OptionWeight) (float64, error) {
	if len(weights) == 0 {
		return 0, fmt.Errorf("weights are required")
	}

	total := 0.0
	for _, item := range weights {
		if item.OptionID == "" {
			return 0, fmt.Errorf("option id is required")
		}
		if item.Weight < 0 {
			return 0, fmt.Errorf("weight for option %q must not be negative", item.OptionID)
		}
		total += item.Weight
	}
	return total, nil
}

func pickEven(rng *rand.Rand, weights []OptionWeight) string {
	point := rng.Float64() * float64(len(weights))
	index := int(point)
	if index >= len(weights) {
		index = len(weights) - 1
	}
	return weights[index].OptionID
}

func PickMany(rng *rand.Rand, weights []OptionWeight, rule SelectionRule) (SelectionResult, error) {
	picker, err := NewSelectionPicker(weights, rule)
	if err != nil {
		return SelectionResult{}, err
	}
	return picker.Pick(rng)
}

func normalizeRule(rule SelectionRule, optionCount int) (int, int, error) {
	min := rule.Min
	max := rule.Max
	if min < 0 {
		return 0, 0, fmt.Errorf("min must not be negative")
	}
	if max == 0 {
		max = optionCount
	}
	if max < 0 {
		return 0, 0, fmt.Errorf("max must not be negative")
	}
	if min > max {
		return 0, 0, fmt.Errorf("min must not be greater than max")
	}
	if max > optionCount {
		return 0, 0, fmt.Errorf("max must not be greater than option count")
	}
	return min, max, nil
}
