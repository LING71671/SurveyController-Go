package wjx

import (
	"fmt"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/domain"
)

type HTTPAnswerPlan struct {
	Answers []HTTPQuestionAnswer
}

type HTTPQuestionAnswer struct {
	QuestionID string
	OptionIDs  []string
	Value      string
}

func BuildHTTPSubmissionDraftFromAnswerPlan(survey domain.SurveyDefinition, plan HTTPAnswerPlan) (HTTPSubmissionDraft, error) {
	answers, err := BuildHTTPAnswers(survey, plan)
	if err != nil {
		return HTTPSubmissionDraft{}, err
	}
	return BuildHTTPSubmissionDraft(survey.URL, answers)
}

func BuildHTTPAnswers(survey domain.SurveyDefinition, plan HTTPAnswerPlan) (map[string]string, error) {
	if len(plan.Answers) == 0 {
		return nil, fmt.Errorf("answer plan is required")
	}

	questions := indexQuestions(survey.Questions)
	answers := make(map[string]string, len(plan.Answers))
	for _, planned := range plan.Answers {
		questionID := strings.TrimSpace(planned.QuestionID)
		if questionID == "" {
			return nil, fmt.Errorf("question id is required")
		}
		question, ok := questions[questionID]
		if !ok {
			return nil, fmt.Errorf("question %q is not defined", questionID)
		}
		if _, exists := answers[questionID]; exists {
			return nil, fmt.Errorf("question %q has duplicate answers", questionID)
		}

		value, err := buildHTTPAnswerValue(question, planned)
		if err != nil {
			return nil, fmt.Errorf("question %q: %w", questionID, err)
		}
		answers[questionID] = value
	}
	return answers, nil
}

func indexQuestions(questions []domain.QuestionDefinition) map[string]domain.QuestionDefinition {
	index := make(map[string]domain.QuestionDefinition, len(questions))
	for _, question := range questions {
		id := strings.TrimSpace(question.ID)
		if id != "" {
			index[id] = question
		}
	}
	return index
}

func buildHTTPAnswerValue(question domain.QuestionDefinition, planned HTTPQuestionAnswer) (string, error) {
	switch question.Kind {
	case domain.QuestionKindSingle, domain.QuestionKindDropdown:
		return buildSingleHTTPAnswerValue(question, planned)
	case domain.QuestionKindMultiple:
		return buildMultipleHTTPAnswerValue(question, planned)
	case domain.QuestionKindRating:
		return buildRatingHTTPAnswerValue(question, planned)
	default:
		return "", fmt.Errorf("kind %q is not supported for HTTP answer plan", question.Kind)
	}
}

func buildSingleHTTPAnswerValue(question domain.QuestionDefinition, planned HTTPQuestionAnswer) (string, error) {
	if len(planned.OptionIDs) > 1 {
		return "", fmt.Errorf("single answer expects one option")
	}
	if len(planned.OptionIDs) == 1 {
		return optionValue(question, planned.OptionIDs[0])
	}
	return directAnswerValue(planned.Value)
}

func buildMultipleHTTPAnswerValue(question domain.QuestionDefinition, planned HTTPQuestionAnswer) (string, error) {
	if len(planned.OptionIDs) == 0 {
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

		value, err := optionValue(question, id)
		if err != nil {
			return "", err
		}
		values = append(values, value)
	}
	return strings.Join(values, ","), nil
}

func buildRatingHTTPAnswerValue(question domain.QuestionDefinition, planned HTTPQuestionAnswer) (string, error) {
	if len(planned.OptionIDs) > 1 {
		return "", fmt.Errorf("rating answer expects one option")
	}
	if len(planned.OptionIDs) == 1 {
		return optionValue(question, planned.OptionIDs[0])
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

func optionValue(question domain.QuestionDefinition, optionID string) (string, error) {
	optionID = strings.TrimSpace(optionID)
	if optionID == "" {
		return "", fmt.Errorf("option id is required")
	}
	for _, option := range question.Options {
		if strings.TrimSpace(option.ID) != optionID {
			continue
		}
		value := strings.TrimSpace(option.Value)
		if value == "" {
			value = strings.TrimSpace(option.ID)
		}
		if value == "" {
			return "", fmt.Errorf("option %q value is required", optionID)
		}
		return value, nil
	}
	return "", fmt.Errorf("option %q is not defined", optionID)
}
