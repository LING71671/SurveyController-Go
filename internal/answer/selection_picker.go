package answer

import (
	"fmt"
	"math/rand"
	"sort"
)

type SelectionPicker struct {
	options []OptionWeight
	indexes map[string]int
	picker  WeightedPicker
	min     int
	max     int
}

func NewSelectionPicker(weights []OptionWeight, rule SelectionRule) (SelectionPicker, error) {
	normalized, err := NormalizeWeights(weights)
	if err != nil {
		return SelectionPicker{}, err
	}
	min, max, err := normalizeRule(rule, len(normalized))
	if err != nil {
		return SelectionPicker{}, err
	}
	picker, err := NewWeightedPicker(normalized)
	if err != nil {
		return SelectionPicker{}, err
	}
	indexes := make(map[string]int, len(normalized))
	for index, item := range normalized {
		indexes[item.OptionID] = index
	}
	return SelectionPicker{
		options: normalized,
		indexes: indexes,
		picker:  picker,
		min:     min,
		max:     max,
	}, nil
}

func (p SelectionPicker) Pick(rng *rand.Rand) (SelectionResult, error) {
	if rng == nil {
		return SelectionResult{}, fmt.Errorf("rng is required")
	}
	if len(p.options) == 0 {
		return SelectionResult{}, fmt.Errorf("selection picker is empty")
	}

	selected := make([]bool, len(p.options))
	ids := make([]string, 0, p.max)
	for index, item := range p.options {
		if rng.Float64() <= item.Weight {
			selected[index] = true
			ids = append(ids, item.OptionID)
		}
	}

	for len(ids) < p.min {
		id, err := p.picker.Pick(rng)
		if err != nil {
			return SelectionResult{}, err
		}
		index := p.indexes[id]
		if selected[index] {
			continue
		}
		selected[index] = true
		ids = append(ids, id)
	}

	if len(ids) > p.max {
		rng.Shuffle(len(ids), func(i, j int) {
			ids[i], ids[j] = ids[j], ids[i]
		})
		ids = ids[:p.max]
	}
	sort.Strings(ids)
	return SelectionResult{OptionIDs: ids}, nil
}

func (p SelectionPicker) Len() int {
	return len(p.options)
}

func (p SelectionPicker) Rule() SelectionRule {
	return SelectionRule{Min: p.min, Max: p.max}
}
