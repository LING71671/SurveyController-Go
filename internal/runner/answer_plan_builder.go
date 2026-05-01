package runner

import (
	"fmt"
	"math"
	"math/rand"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/answer"
	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/domain"
)

func BuildAnswerPlan(rng *rand.Rand, questions []QuestionPlan) (answerplan.Plan, error) {
	if rng == nil {
		return answerplan.Plan{}, fmt.Errorf("rng is required")
	}
	if len(questions) == 0 {
		return answerplan.Plan{}, fmt.Errorf("questions are required")
	}

	plan := answerplan.Plan{
		Answers: make([]answerplan.QuestionAnswer, 0, len(questions)),
	}
	for _, question := range questions {
		answer, err := buildQuestionAnswer(rng, question)
		if err != nil {
			return answerplan.Plan{}, err
		}
		plan.Answers = append(plan.Answers, answer)
	}
	return plan, nil
}

func buildQuestionAnswer(rng *rand.Rand, question QuestionPlan) (answerplan.QuestionAnswer, error) {
	questionID := strings.TrimSpace(question.ID)
	if questionID == "" {
		return answerplan.QuestionAnswer{}, fmt.Errorf("question id is required")
	}

	kind, err := domain.ParseQuestionKind(question.Kind)
	if err != nil {
		return answerplan.QuestionAnswer{}, fmt.Errorf("question %q kind: %w", questionID, err)
	}

	switch kind {
	case domain.QuestionKindSingle, domain.QuestionKindDropdown, domain.QuestionKindRating:
		optionID, err := answer.PickOne(rng, question.Weights)
		if err != nil {
			return answerplan.QuestionAnswer{}, fmt.Errorf("question %q pick one: %w", questionID, err)
		}
		return answerplan.QuestionAnswer{QuestionID: questionID, OptionIDs: []string{optionID}}, nil
	case domain.QuestionKindMultiple:
		selected, err := answer.PickMany(rng, question.Weights, selectionRuleFromOptions(question.Options))
		if err != nil {
			return answerplan.QuestionAnswer{}, fmt.Errorf("question %q pick many: %w", questionID, err)
		}
		return answerplan.QuestionAnswer{QuestionID: questionID, OptionIDs: selected.OptionIDs}, nil
	default:
		return answerplan.QuestionAnswer{}, fmt.Errorf("question %q kind %q is not supported for answer plan builder", questionID, kind)
	}
}

func selectionRuleFromOptions(options map[string]any) answer.SelectionRule {
	return answer.SelectionRule{
		Min: intOption(options, "min_selected", "min"),
		Max: intOption(options, "max_selected", "max"),
	}
}

func intOption(options map[string]any, keys ...string) int {
	for _, key := range keys {
		raw, ok := options[key]
		if !ok {
			continue
		}
		value, ok := asNonNegativeInt(raw)
		if ok {
			return value
		}
	}
	return 0
}

func asNonNegativeInt(raw any) (int, bool) {
	switch value := raw.(type) {
	case int:
		return clampNonNegative(value), true
	case int64:
		return clampNonNegative64(value), true
	case float64:
		if math.Trunc(value) != value {
			return 0, false
		}
		return clampNonNegative64(int64(value)), true
	default:
		return 0, false
	}
}

func clampNonNegative(value int) int {
	if value < 0 {
		return 0
	}
	return value
}

func clampNonNegative64(value int64) int {
	if value < 0 {
		return 0
	}
	maxInt := int64(int(^uint(0) >> 1))
	if value > maxInt {
		return int(maxInt)
	}
	return int(value)
}
