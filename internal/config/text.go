package config

import (
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/answer"
)

func QuestionTextAnswerRule(question QuestionConfig) (answer.TextAnswerRule, bool, error) {
	raw, ok := question.Options["text"]
	if !ok {
		return answer.TextAnswerRule{}, false, nil
	}
	rawMap, err := asMap(raw, "options.text")
	if err != nil {
		return answer.TextAnswerRule{}, true, err
	}

	words, err := optionalStringList(rawMap["words"], "options.text.words")
	if err != nil {
		return answer.TextAnswerRule{}, true, err
	}
	values, err := optionalStringList(rawMap["values"], "options.text.values")
	if err != nil {
		return answer.TextAnswerRule{}, true, err
	}
	prefixes, err := optionalStringList(rawMap["prefixes"], "options.text.prefixes")
	if err != nil {
		return answer.TextAnswerRule{}, true, err
	}
	minWords, err := optionalInt(rawMap["min_words"], "options.text.min_words")
	if err != nil {
		return answer.TextAnswerRule{}, true, err
	}
	maxWords, err := optionalInt(rawMap["max_words"], "options.text.max_words")
	if err != nil {
		return answer.TextAnswerRule{}, true, err
	}
	length, err := optionalInt(rawMap["length"], "options.text.length")
	if err != nil {
		return answer.TextAnswerRule{}, true, err
	}
	slots, err := optionalStringSlots(rawMap["slots"])
	if err != nil {
		return answer.TextAnswerRule{}, true, err
	}

	rule := answer.TextAnswerRule{
		Mode: answer.TextAnswerMode(optionalString(rawMap["mode"])),
		Words: answer.TextRule{
			Words:     words,
			MinWords:  minWords,
			MaxWords:  maxWords,
			Separator: optionalString(rawMap["separator"]),
		},
		Digits: answer.DigitsRule{
			Length: length,
			Prefix: optionalString(rawMap["prefix"]),
		},
		Phone: answer.PhoneRule{
			Prefixes: prefixes,
		},
		Template: answer.TemplateRule{
			Template: optionalString(rawMap["template"]),
			Slots:    slots,
		},
		Values: values,
	}
	if err := answer.ValidateTextAnswerRule(rule); err != nil {
		return answer.TextAnswerRule{}, true, fmt.Errorf("options.text: %w", err)
	}
	return rule, true, nil
}

func optionalString(raw any) string {
	value, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(value)
}

func optionalInt(raw any, name string) (int, error) {
	if raw == nil {
		return 0, nil
	}
	value, err := numeric(raw, name)
	if err != nil {
		return 0, err
	}
	return int(value), nil
}

func optionalStringList(raw any, name string) ([]string, error) {
	if raw == nil {
		return nil, nil
	}
	if value, ok := raw.(string); ok {
		value = strings.TrimSpace(value)
		if value == "" {
			return nil, nil
		}
		return []string{value}, nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be a string or list", name)
	}
	values := make([]string, 0, len(items))
	for index, item := range items {
		value, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%s[%d] must be a string", name, index)
		}
		value = strings.TrimSpace(value)
		if value != "" {
			values = append(values, value)
		}
	}
	return values, nil
}

func optionalStringSlots(raw any) (map[string][]string, error) {
	if raw == nil {
		return nil, nil
	}
	rawMap, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("options.text.slots must be an object")
	}
	slots := make(map[string][]string, len(rawMap))
	for key, value := range rawMap {
		key = strings.TrimSpace(key)
		values, err := optionalStringList(value, "options.text.slots."+key)
		if err != nil {
			return nil, err
		}
		if key != "" && len(values) > 0 {
			slots[key] = values
		}
	}
	return slots, nil
}
