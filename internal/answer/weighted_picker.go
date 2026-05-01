package answer

import (
	"fmt"
	"math/rand"
)

type WeightedPicker struct {
	options []OptionWeight
	total   float64
}

func NewWeightedPicker(weights []OptionWeight) (WeightedPicker, error) {
	total, err := validateAndSumWeights(weights)
	if err != nil {
		return WeightedPicker{}, err
	}

	return WeightedPicker{
		options: append([]OptionWeight(nil), weights...),
		total:   total,
	}, nil
}

func (p WeightedPicker) Pick(rng *rand.Rand) (string, error) {
	if rng == nil {
		return "", fmt.Errorf("rng is required")
	}
	if len(p.options) == 0 {
		return "", fmt.Errorf("weighted picker is empty")
	}
	if p.total == 0 {
		return pickEven(rng, p.options), nil
	}

	point := rng.Float64() * p.total
	accumulated := 0.0
	for _, item := range p.options {
		accumulated += item.Weight
		if point <= accumulated {
			return item.OptionID, nil
		}
	}
	return p.options[len(p.options)-1].OptionID, nil
}

func (p WeightedPicker) Len() int {
	return len(p.options)
}
