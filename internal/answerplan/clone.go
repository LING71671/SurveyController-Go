package answerplan

func Clone(plan Plan) Plan {
	if len(plan.Answers) == 0 {
		return Plan{}
	}
	cloned := Plan{
		Answers: make([]QuestionAnswer, len(plan.Answers)),
	}
	for i, answer := range plan.Answers {
		cloned.Answers[i] = QuestionAnswer{
			QuestionID: answer.QuestionID,
			OptionIDs:  append([]string(nil), answer.OptionIDs...),
			Value:      answer.Value,
			Rows:       cloneRows(answer.Rows),
		}
	}
	return cloned
}

func cloneRows(rows []RowAnswer) []RowAnswer {
	if len(rows) == 0 {
		return nil
	}
	cloned := make([]RowAnswer, len(rows))
	for i, row := range rows {
		cloned[i] = RowAnswer{
			RowID:     row.RowID,
			OptionIDs: append([]string(nil), row.OptionIDs...),
			Value:     row.Value,
		}
	}
	return cloned
}
