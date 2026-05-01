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
		}
	}
	return cloned
}
