package answerplan

import "strings"

type Plan struct {
	Answers []QuestionAnswer
}

type QuestionAnswer struct {
	QuestionID string
	OptionIDs  []string
	Value      string
}

func (p Plan) Empty() bool {
	return len(p.Answers) == 0
}

func (a QuestionAnswer) NormalizedQuestionID() string {
	return strings.TrimSpace(a.QuestionID)
}

func (a QuestionAnswer) HasOptionIDs() bool {
	return len(a.OptionIDs) > 0
}

func (a QuestionAnswer) DirectValue() string {
	return strings.TrimSpace(a.Value)
}
