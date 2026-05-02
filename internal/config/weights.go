package config

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/answer"
)

func QuestionOptionWeights(question QuestionConfig) ([]answer.OptionWeight, error) {
	raw, ok := question.Options["weights"]
	if !ok || raw == nil {
		return nil, nil
	}
	return parseOptionWeights(raw, "options.weights")
}

func QuestionMatrixWeights(question QuestionConfig) (map[string][]answer.OptionWeight, error) {
	raw, ok := question.Options["matrix_weights"]
	if !ok || raw == nil {
		return nil, nil
	}
	rows, err := asList(raw, "options.matrix_weights")
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, fmt.Errorf("options.matrix_weights must not be empty")
	}

	result := make(map[string][]answer.OptionWeight, len(rows))
	for index, rawRow := range rows {
		name := fmt.Sprintf("options.matrix_weights[%d]", index)
		row, err := asMap(rawRow, name)
		if err != nil {
			return nil, err
		}
		rowID, err := requiredString(row["row_id"], name+".row_id")
		if err != nil {
			return nil, err
		}
		weights, err := parseOptionWeights(row["weights"], name+".weights")
		if err != nil {
			return nil, err
		}
		result[rowID] = weights
	}
	return result, nil
}

func parseOptionWeights(raw any, name string) ([]answer.OptionWeight, error) {
	items, err := asList(raw, name)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%s must not be empty", name)
	}

	weights := make([]answer.OptionWeight, 0, len(items))
	for index, item := range items {
		itemName := fmt.Sprintf("%s[%d]", name, index)
		fields, err := asMap(item, itemName)
		if err != nil {
			return nil, err
		}
		optionID, err := requiredString(fields["option_id"], itemName+".option_id")
		if err != nil {
			return nil, err
		}
		weight, err := numeric(fields["weight"], itemName+".weight")
		if err != nil {
			return nil, err
		}
		if weight < 0 {
			return nil, fmt.Errorf("%s.weight must not be negative", itemName)
		}
		weights = append(weights, answer.OptionWeight{
			OptionID: optionID,
			Weight:   weight,
		})
	}
	return weights, nil
}

func asList(raw any, name string) ([]any, error) {
	switch value := raw.(type) {
	case []any:
		return value, nil
	case []map[string]any:
		items := make([]any, 0, len(value))
		for _, item := range value {
			items = append(items, item)
		}
		return items, nil
	default:
		return nil, fmt.Errorf("%s must be a list", name)
	}
}

func asMap(raw any, name string) (map[string]any, error) {
	switch value := raw.(type) {
	case map[string]any:
		return value, nil
	case map[any]any:
		result := make(map[string]any, len(value))
		for key, item := range value {
			result[fmt.Sprint(key)] = item
		}
		return result, nil
	default:
		return nil, fmt.Errorf("%s must be an object", name)
	}
}

func requiredString(raw any, name string) (string, error) {
	value := strings.TrimSpace(fmt.Sprint(raw))
	if value == "" || value == "<nil>" {
		return "", fmt.Errorf("%s is required", name)
	}
	return value, nil
}

func numeric(raw any, name string) (float64, error) {
	switch value := raw.(type) {
	case int:
		return float64(value), nil
	case int8:
		return float64(value), nil
	case int16:
		return float64(value), nil
	case int32:
		return float64(value), nil
	case int64:
		return float64(value), nil
	case uint:
		return float64(value), nil
	case uint8:
		return float64(value), nil
	case uint16:
		return float64(value), nil
	case uint32:
		return float64(value), nil
	case uint64:
		return float64(value), nil
	case float32:
		return float64(value), nil
	case float64:
		return value, nil
	case json.Number:
		parsed, err := strconv.ParseFloat(value.String(), 64)
		if err != nil {
			return 0, fmt.Errorf("%s must be numeric", name)
		}
		return parsed, nil
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err != nil {
			return 0, fmt.Errorf("%s must be numeric", name)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("%s must be numeric", name)
	}
}
