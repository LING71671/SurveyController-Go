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

type AnswerPlanBuilder struct {
	questions []compiledQuestionPlan
}

type compiledQuestionPlan struct {
	id      string
	kind    domain.QuestionKind
	picker  answer.WeightedPicker
	weights []answer.OptionWeight
	rule    answer.SelectionRule
}

func BuildAnswerPlan(rng *rand.Rand, questions []QuestionPlan) (answerplan.Plan, error) {
	if rng == nil {
		return answerplan.Plan{}, fmt.Errorf("rng is required")
	}
	builder, err := CompileAnswerPlanBuilder(questions)
	if err != nil {
		return answerplan.Plan{}, err
	}
	return builder.Build(rng)
}

func CompileAnswerPlanBuilder(questions []QuestionPlan) (AnswerPlanBuilder, error) {
	if len(questions) == 0 {
		return AnswerPlanBuilder{}, fmt.Errorf("questions are required")
	}
	builder := AnswerPlanBuilder{
		questions: make([]compiledQuestionPlan, 0, len(questions)),
	}
	for _, question := range questions {
		compiled, err := compileQuestionPlan(question)
		if err != nil {
			return AnswerPlanBuilder{}, err
		}
		builder.questions = append(builder.questions, compiled)
	}
	return builder, nil
}

func (b AnswerPlanBuilder) Build(rng *rand.Rand) (answerplan.Plan, error) {
	if rng == nil {
		return answerplan.Plan{}, fmt.Errorf("rng is required")
	}
	if len(b.questions) == 0 {
		return answerplan.Plan{}, fmt.Errorf("questions are required")
	}
	plan := answerplan.Plan{
		Answers: make([]answerplan.QuestionAnswer, 0, len(b.questions)),
	}
	for _, question := range b.questions {
		answer, err := question.buildAnswer(rng)
		if err != nil {
			return answerplan.Plan{}, err
		}
		plan.Answers = append(plan.Answers, answer)
	}
	return plan, nil
}

func (b AnswerPlanBuilder) BuildMany(rng *rand.Rand, count int) ([]answerplan.Plan, error) {
	if count <= 0 {
		return nil, fmt.Errorf("answer plan count must be greater than 0")
	}
	plans := make([]answerplan.Plan, 0, count)
	for i := 0; i < count; i++ {
		plan, err := b.Build(rng)
		if err != nil {
			return nil, fmt.Errorf("answer plan %d: %w", i+1, err)
		}
		plans = append(plans, answerplan.Clone(plan))
	}
	return plans, nil
}

func BuildAnswerPlans(rng *rand.Rand, questions []QuestionPlan, count int) ([]answerplan.Plan, error) {
	if count <= 0 {
		return nil, fmt.Errorf("answer plan count must be greater than 0")
	}
	builder, err := CompileAnswerPlanBuilder(questions)
	if err != nil {
		return nil, err
	}
	return builder.BuildMany(rng, count)
}

func compileQuestionPlan(question QuestionPlan) (compiledQuestionPlan, error) {
	questionID := strings.TrimSpace(question.ID)
	if questionID == "" {
		return compiledQuestionPlan{}, fmt.Errorf("question id is required")
	}

	kind, err := domain.ParseQuestionKind(question.Kind)
	if err != nil {
		return compiledQuestionPlan{}, fmt.Errorf("question %q kind: %w", questionID, err)
	}

	switch kind {
	case domain.QuestionKindSingle, domain.QuestionKindDropdown, domain.QuestionKindRating:
		picker, err := answer.NewWeightedPicker(question.Weights)
		if err != nil {
			return compiledQuestionPlan{}, fmt.Errorf("question %q weights: %w", questionID, err)
		}
		return compiledQuestionPlan{id: questionID, kind: kind, picker: picker}, nil
	case domain.QuestionKindMultiple:
		weights := append([]answer.OptionWeight(nil), question.Weights...)
		rule := selectionRuleFromOptions(question.Options)
		if _, err := answer.PickMany(rand.New(rand.NewSource(1)), weights, rule); err != nil {
			return compiledQuestionPlan{}, fmt.Errorf("question %q weights: %w", questionID, err)
		}
		return compiledQuestionPlan{id: questionID, kind: kind, weights: weights, rule: rule}, nil
	default:
		return compiledQuestionPlan{}, fmt.Errorf("question %q kind %q is not supported for answer plan builder", questionID, kind)
	}
}

func (q compiledQuestionPlan) buildAnswer(rng *rand.Rand) (answerplan.QuestionAnswer, error) {
	switch q.kind {
	case domain.QuestionKindSingle, domain.QuestionKindDropdown, domain.QuestionKindRating:
		optionID, err := q.picker.Pick(rng)
		if err != nil {
			return answerplan.QuestionAnswer{}, fmt.Errorf("question %q pick one: %w", q.id, err)
		}
		return answerplan.QuestionAnswer{QuestionID: q.id, OptionIDs: []string{optionID}}, nil
	case domain.QuestionKindMultiple:
		selected, err := answer.PickMany(rng, q.weights, q.rule)
		if err != nil {
			return answerplan.QuestionAnswer{}, fmt.Errorf("question %q pick many: %w", q.id, err)
		}
		return answerplan.QuestionAnswer{QuestionID: q.id, OptionIDs: selected.OptionIDs}, nil
	default:
		return answerplan.QuestionAnswer{}, fmt.Errorf("question %q kind %q is not supported for answer plan builder", q.id, q.kind)
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
