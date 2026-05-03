package answerplan

import "testing"

func TestClone(t *testing.T) {
	plan := Plan{
		Answers: []QuestionAnswer{
			{
				QuestionID: "q1",
				OptionIDs:  []string{"a", "b"},
				Value:      "x",
				Rows:       []RowAnswer{{RowID: "r1", OptionIDs: []string{"c"}, Value: "y"}},
			},
		},
	}

	cloned := Clone(plan)
	plan.Answers[0].QuestionID = "mutated"
	plan.Answers[0].OptionIDs[0] = "mutated"
	plan.Answers[0].Value = "mutated"
	plan.Answers[0].Rows[0].RowID = "mutated"
	plan.Answers[0].Rows[0].OptionIDs[0] = "mutated"
	plan.Answers[0].Rows[0].Value = "mutated"

	if cloned.Answers[0].QuestionID != "q1" || cloned.Answers[0].OptionIDs[0] != "a" || cloned.Answers[0].Value != "x" {
		t.Fatalf("Clone() = %+v, want independent copy", cloned)
	}
	if cloned.Answers[0].Rows[0].RowID != "r1" || cloned.Answers[0].Rows[0].OptionIDs[0] != "c" || cloned.Answers[0].Rows[0].Value != "y" {
		t.Fatalf("Clone() rows = %+v, want independent row copy", cloned.Answers[0].Rows)
	}
}
