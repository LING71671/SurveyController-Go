package answer

import (
	"fmt"
	"math/rand"
	"sort"
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
	if len(weights) == 0 {
		return nil, fmt.Errorf("weights are required")
	}

	normalized := make([]OptionWeight, 0, len(weights))
	total := 0.0
	for _, item := range weights {
		if item.OptionID == "" {
			return nil, fmt.Errorf("option id is required")
		}
		if item.Weight < 0 {
			return nil, fmt.Errorf("weight for option %q must not be negative", item.OptionID)
		}
		total += item.Weight
		normalized = append(normalized, item)
	}
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
	normalized, err := NormalizeWeights(weights)
	if err != nil {
		return "", err
	}

	point := rng.Float64()
	accumulated := 0.0
	for _, item := range normalized {
		accumulated += item.Weight
		if point <= accumulated {
			return item.OptionID, nil
		}
	}
	return normalized[len(normalized)-1].OptionID, nil
}

func PickMany(rng *rand.Rand, weights []OptionWeight, rule SelectionRule) (SelectionResult, error) {
	if rng == nil {
		return SelectionResult{}, fmt.Errorf("rng is required")
	}
	normalized, err := NormalizeWeights(weights)
	if err != nil {
		return SelectionResult{}, err
	}
	min, max, err := normalizeRule(rule, len(normalized))
	if err != nil {
		return SelectionResult{}, err
	}

	selected := map[string]bool{}
	for _, item := range normalized {
		if rng.Float64() <= item.Weight {
			selected[item.OptionID] = true
		}
	}

	for len(selected) < min {
		id, err := PickOne(rng, normalized)
		if err != nil {
			return SelectionResult{}, err
		}
		selected[id] = true
	}

	if len(selected) > max {
		ids := keys(selected)
		rng.Shuffle(len(ids), func(i, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})
		selected = map[string]bool{}
		for _, id := range ids[:max] {
			selected[id] = true
		}
	}

	return SelectionResult{OptionIDs: keys(selected)}, nil
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

func keys(values map[string]bool) []string {
	ids := make([]string, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
