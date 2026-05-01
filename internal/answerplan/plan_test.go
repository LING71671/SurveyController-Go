package answerplan

import "testing"

func TestPlanEmpty(t *testing.T) {
	if !(Plan{}).Empty() {
		t.Fatal("empty plan returned false")
	}
	if (Plan{Answers: []QuestionAnswer{{QuestionID: "q1"}}}).Empty() {
		t.Fatal("plan with answers returned true")
	}
}

func TestQuestionAnswerHelpers(t *testing.T) {
	answer := QuestionAnswer{
		QuestionID: " q1 ",
		OptionIDs:  []string{"a"},
		Value:      " 1 ",
	}
	if answer.NormalizedQuestionID() != "q1" {
		t.Fatalf("NormalizedQuestionID() = %q, want q1", answer.NormalizedQuestionID())
	}
	if !answer.HasOptionIDs() {
		t.Fatal("HasOptionIDs() = false, want true")
	}
	if answer.DirectValue() != "1" {
		t.Fatalf("DirectValue() = %q, want 1", answer.DirectValue())
	}
}
