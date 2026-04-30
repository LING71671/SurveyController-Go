package domain

import (
	"fmt"
	"strings"
)

type ProviderID string

const (
	ProviderUnknown ProviderID = ""
	ProviderWJX     ProviderID = "wjx"
	ProviderTencent ProviderID = "tencent"
	ProviderCredamo ProviderID = "credamo"
)

func (id ProviderID) String() string {
	return string(id)
}

func ParseProviderID(raw string) (ProviderID, error) {
	switch ProviderID(strings.ToLower(strings.TrimSpace(raw))) {
	case ProviderWJX:
		return ProviderWJX, nil
	case ProviderTencent:
		return ProviderTencent, nil
	case ProviderCredamo:
		return ProviderCredamo, nil
	default:
		return ProviderUnknown, fmt.Errorf("unsupported provider %q", raw)
	}
}

type QuestionKind string

const (
	QuestionKindUnknown  QuestionKind = ""
	QuestionKindSingle   QuestionKind = "single"
	QuestionKindMultiple QuestionKind = "multiple"
	QuestionKindDropdown QuestionKind = "dropdown"
	QuestionKindText     QuestionKind = "text"
	QuestionKindTextarea QuestionKind = "textarea"
	QuestionKindRating   QuestionKind = "rating"
	QuestionKindMatrix   QuestionKind = "matrix"
	QuestionKindRanking  QuestionKind = "ranking"
)

func (k QuestionKind) String() string {
	return string(k)
}

func (k QuestionKind) Valid() bool {
	_, err := ParseQuestionKind(k.String())
	return err == nil
}

func ParseQuestionKind(raw string) (QuestionKind, error) {
	switch QuestionKind(strings.ToLower(strings.TrimSpace(raw))) {
	case QuestionKindSingle:
		return QuestionKindSingle, nil
	case QuestionKindMultiple:
		return QuestionKindMultiple, nil
	case QuestionKindDropdown:
		return QuestionKindDropdown, nil
	case QuestionKindText:
		return QuestionKindText, nil
	case QuestionKindTextarea:
		return QuestionKindTextarea, nil
	case QuestionKindRating:
		return QuestionKindRating, nil
	case QuestionKindMatrix:
		return QuestionKindMatrix, nil
	case QuestionKindRanking:
		return QuestionKindRanking, nil
	default:
		return QuestionKindUnknown, fmt.Errorf("unsupported question kind %q", raw)
	}
}
