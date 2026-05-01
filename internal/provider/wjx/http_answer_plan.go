package wjx

import (
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/domain"
)

type HTTPAnswerPlan = answerplan.Plan
type HTTPQuestionAnswer = answerplan.QuestionAnswer

type httpAnswerSchema struct {
	questions map[string]httpQuestionSpec
}

type httpQuestionSpec struct {
	id      string
	kind    domain.QuestionKind
	options map[string]string
}

func BuildHTTPSubmissionDraftFromAnswerPlan(survey domain.SurveyDefinition, plan answerplan.Plan) (HTTPSubmissionDraft, error) {
	answers, err := BuildHTTPAnswers(survey, plan)
	if err != nil {
		return HTTPSubmissionDraft{}, err
	}
	return BuildHTTPSubmissionDraft(survey.URL, answers)
}

func BuildHTTPAnswers(survey domain.SurveyDefinition, plan answerplan.Plan) (map[string]string, error) {
	if plan.Empty() {
		return nil, fmt.Errorf("answer plan is required")
	}

	schema, err := compileHTTPAnswerSchema(survey.Questions)
	if err != nil {
		return nil, err
	}

	answers := make(map[string]string, len(plan.Answers))
	for _, planned := range plan.Answers {
		questionID := planned.NormalizedQuestionID()
		if _, exists := answers[questionID]; exists {
			return nil, fmt.Errorf("question %q has duplicate answers", questionID)
		}

		value, err := schema.mapAnswer(planned)
		if err != nil {
			return nil, fmt.Errorf("question %q: %w", questionID, err)
		}
		answers[questionID] = value
	}
	return answers, nil
}

func compileHTTPAnswerSchema(questions []domain.QuestionDefinition) (httpAnswerSchema, error) {
	schema := httpAnswerSchema{
		questions: make(map[string]httpQuestionSpec, len(questions)),
	}
	for _, question := range questions {
		spec, err := compileHTTPQuestionSpec(question)
		if err != nil {
			return httpAnswerSchema{}, err
		}
		if spec.id != "" {
			if _, exists := schema.questions[spec.id]; exists {
				return httpAnswerSchema{}, fmt.Errorf("question %q is defined more than once", spec.id)
			}
			schema.questions[spec.id] = spec
		}
	}
	return schema, nil
}

func compileHTTPQuestionSpec(question domain.QuestionDefinition) (httpQuestionSpec, error) {
	spec := httpQuestionSpec{
		id:      strings.TrimSpace(question.ID),
		kind:    question.Kind,
		options: make(map[string]string, len(question.Options)),
	}
	for _, option := range question.Options {
		id, value, err := compileHTTPOptionValue(option)
		if err != nil {
			return httpQuestionSpec{}, fmt.Errorf("question %q option: %w", spec.id, err)
		}
		if id != "" {
			if _, exists := spec.options[id]; exists {
				return httpQuestionSpec{}, fmt.Errorf("question %q option %q is defined more than once", spec.id, id)
			}
			spec.options[id] = value
		}
	}
	return spec, nil
}

func compileHTTPOptionValue(option domain.OptionDefinition) (string, string, error) {
	id := strings.TrimSpace(option.ID)
	value := strings.TrimSpace(option.Value)
	if value == "" {
		value = id
	}
	if id != "" && value == "" {
		return "", "", fmt.Errorf("option %q value is required", id)
	}
	return id, value, nil
}

func (s httpAnswerSchema) mapAnswer(planned answerplan.QuestionAnswer) (string, error) {
	questionID := planned.NormalizedQuestionID()
	if questionID == "" {
		return "", fmt.Errorf("question id is required")
	}
	question, ok := s.questions[questionID]
	if !ok {
		return "", fmt.Errorf("question %q is not defined", questionID)
	}
	return question.mapAnswer(planned)
}

func (q httpQuestionSpec) mapAnswer(planned answerplan.QuestionAnswer) (string, error) {
	switch q.kind {
	case domain.QuestionKindSingle, domain.QuestionKindDropdown:
		return q.mapSingleAnswer(planned)
	case domain.QuestionKindMultiple:
		return q.mapMultipleAnswer(planned)
	case domain.QuestionKindRating:
		return q.mapRatingAnswer(planned)
	default:
		return "", fmt.Errorf("kind %q is not supported for HTTP answer plan", q.kind)
	}
}

func (q httpQuestionSpec) mapSingleAnswer(planned answerplan.QuestionAnswer) (string, error) {
	if len(planned.OptionIDs) > 1 {
		return "", fmt.Errorf("single answer expects one option")
	}
	if planned.HasOptionIDs() {
		return q.optionValue(planned.OptionIDs[0])
	}
	return directAnswerValue(planned.Value)
}

func (q httpQuestionSpec) mapMultipleAnswer(planned answerplan.QuestionAnswer) (string, error) {
	if !planned.HasOptionIDs() {
		return directAnswerValue(planned.Value)
	}

	seen := map[string]bool{}
	values := make([]string, 0, len(planned.OptionIDs))
	for _, optionID := range planned.OptionIDs {
		id := strings.TrimSpace(optionID)
		if id == "" {
			return "", fmt.Errorf("option id is required")
		}
		if seen[id] {
			return "", fmt.Errorf("option %q is selected more than once", id)
		}
		seen[id] = true

		value, err := q.optionValue(id)
		if err != nil {
			return "", err
		}
		values = append(values, value)
	}
	return strings.Join(values, ","), nil
}

func (q httpQuestionSpec) mapRatingAnswer(planned answerplan.QuestionAnswer) (string, error) {
	if len(planned.OptionIDs) > 1 {
		return "", fmt.Errorf("rating answer expects one option")
	}
	if planned.HasOptionIDs() {
		return q.optionValue(planned.OptionIDs[0])
	}
	return directAnswerValue(planned.Value)
}

func directAnswerValue(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("answer value is required")
	}
	return value, nil
}

func (q httpQuestionSpec) optionValue(optionID string) (string, error) {
	optionID = strings.TrimSpace(optionID)
	if optionID == "" {
		return "", fmt.Errorf("option id is required")
	}
	value, ok := q.options[optionID]
	if ok {
		return value, nil
	}
	return "", fmt.Errorf("option %q is not defined", optionID)
}
