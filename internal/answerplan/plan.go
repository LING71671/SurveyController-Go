package answerplan

import "strings"

type Plan struct {
	Answers []QuestionAnswer
}

type QuestionAnswer struct {
	QuestionID string
	OptionIDs  []string
	Value      string
	Rows       []RowAnswer
}

type RowAnswer struct {
	RowID     string
	OptionIDs []string
	Value     string
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

func (a QuestionAnswer) HasRows() bool {
	return len(a.Rows) > 0
}

func (r RowAnswer) NormalizedRowID() string {
	return strings.TrimSpace(r.RowID)
}

func (r RowAnswer) HasOptionIDs() bool {
	return len(r.OptionIDs) > 0
}

func (r RowAnswer) DirectValue() string {
	return strings.TrimSpace(r.Value)
}
