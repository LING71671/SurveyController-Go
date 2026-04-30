package answer

import (
	"fmt"
	"math/rand"
	"strings"
)

type TextRule struct {
	Words     []string
	MinWords  int
	MaxWords  int
	Separator string
}

func RandomText(rng *rand.Rand, rule TextRule) (string, error) {
	if rng == nil {
		return "", fmt.Errorf("rng is required")
	}
	if len(rule.Words) == 0 {
		return "", fmt.Errorf("words are required")
	}
	min := rule.MinWords
	max := rule.MaxWords
	if min <= 0 {
		min = 1
	}
	if max == 0 {
		max = min
	}
	if min > max {
		return "", fmt.Errorf("min words must not be greater than max words")
	}
	separator := rule.Separator
	if separator == "" {
		separator = " "
	}

	count := min
	if max > min {
		count += rng.Intn(max - min + 1)
	}
	words := make([]string, 0, count)
	for len(words) < count {
		candidate := strings.TrimSpace(rule.Words[rng.Intn(len(rule.Words))])
		if candidate != "" {
			words = append(words, candidate)
		}
	}
	return strings.Join(words, separator), nil
}

func RandomInt(rng *rand.Rand, min int, max int) (int, error) {
	if rng == nil {
		return 0, fmt.Errorf("rng is required")
	}
	if min > max {
		return 0, fmt.Errorf("min must not be greater than max")
	}
	if min == max {
		return min, nil
	}
	return min + rng.Intn(max-min+1), nil
}
