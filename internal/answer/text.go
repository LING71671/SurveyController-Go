package answer

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
)

type TextRule struct {
	Words     []string
	MinWords  int
	MaxWords  int
	Separator string
}

type DigitsRule struct {
	Length int
	Prefix string
}

type PhoneRule struct {
	Prefixes []string
}

type TemplateRule struct {
	Template string
	Slots    map[string][]string
}

var templateSlotRE = regexp.MustCompile(`\{([A-Za-z0-9_]+)\}`)

func RandomText(rng *rand.Rand, rule TextRule) (string, error) {
	if rng == nil {
		return "", fmt.Errorf("rng is required")
	}
	ruleWords := cleanWords(rule.Words)
	if len(ruleWords) == 0 {
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
		words = append(words, ruleWords[rng.Intn(len(ruleWords))])
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

func RandomDigits(rng *rand.Rand, rule DigitsRule) (string, error) {
	if rng == nil {
		return "", fmt.Errorf("rng is required")
	}
	if rule.Length <= 0 {
		return "", fmt.Errorf("length must be positive")
	}
	prefix := strings.TrimSpace(rule.Prefix)
	if !isDigits(prefix) {
		return "", fmt.Errorf("prefix must contain only digits")
	}
	if len(prefix) > rule.Length {
		return "", fmt.Errorf("prefix length must not exceed length")
	}
	builder := strings.Builder{}
	builder.Grow(rule.Length)
	builder.WriteString(prefix)
	for builder.Len() < rule.Length {
		builder.WriteByte(byte('0' + rng.Intn(10)))
	}
	return builder.String(), nil
}

func RandomPhoneLike(rng *rand.Rand, rule PhoneRule) (string, error) {
	if rng == nil {
		return "", fmt.Errorf("rng is required")
	}
	prefixes := cleanWords(rule.Prefixes)
	if len(prefixes) == 0 {
		prefixes = []string{"130", "131", "132", "155", "156", "185", "186"}
	}
	for _, prefix := range prefixes {
		if len(prefix) >= 11 || !isDigits(prefix) {
			return "", fmt.Errorf("phone prefix %q must be digits shorter than 11", prefix)
		}
	}
	prefix := prefixes[rng.Intn(len(prefixes))]
	return RandomDigits(rng, DigitsRule{Length: 11, Prefix: prefix})
}

func RandomTemplateText(rng *rand.Rand, rule TemplateRule) (string, error) {
	if rng == nil {
		return "", fmt.Errorf("rng is required")
	}
	template := strings.TrimSpace(rule.Template)
	if template == "" {
		return "", fmt.Errorf("template is required")
	}
	result := templateSlotRE.ReplaceAllStringFunc(template, func(match string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(match, "{"), "}")
		values := cleanWords(rule.Slots[name])
		if len(values) == 0 {
			return match
		}
		return values[rng.Intn(len(values))]
	})
	missing := templateSlotRE.FindAllString(result, -1)
	if len(missing) > 0 {
		return "", fmt.Errorf("template slot %s has no values", missing[0])
	}
	return result, nil
}

func cleanWords(words []string) []string {
	cleaned := make([]string, 0, len(words))
	for _, word := range words {
		word = strings.TrimSpace(word)
		if word != "" {
			cleaned = append(cleaned, word)
		}
	}
	return cleaned
}

func isDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
