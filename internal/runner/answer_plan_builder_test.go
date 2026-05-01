package runner

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/answer"
)

func TestBuildAnswerPlan(t *testing.T) {
	got, err := BuildAnswerPlan(rand.New(rand.NewSource(1)), []QuestionPlan{
		{
			ID:   "q1",
			Kind: "single",
			Weights: []answer.OptionWeight{
				{OptionID: "a", Weight: 0},
				{OptionID: "b", Weight: 1},
			},
		},
		{
			ID:   "q2",
			Kind: "multiple",
			Options: map[string]any{
				"min_selected": 2,
				"max_selected": 2,
			},
			Weights: []answer.OptionWeight{
				{OptionID: "a", Weight: 1},
				{OptionID: "b", Weight: 1},
				{OptionID: "c", Weight: 1},
			},
		},
		{
			ID:   "q3",
			Kind: "rating",
			Weights: []answer.OptionWeight{
				{OptionID: "score5", Weight: 1},
			},
		},
		{
			ID:   "q4",
			Kind: "dropdown",
			Weights: []answer.OptionWeight{
				{OptionID: "city2", Weight: 1},
			},
		},
	})
	if err != nil {
		t.Fatalf("BuildAnswerPlan() returned error: %v", err)
	}

	if len(got.Answers) != 4 {
		t.Fatalf("len(Answers) = %d, want 4", len(got.Answers))
	}
	if got.Answers[0].QuestionID != "q1" || got.Answers[0].OptionIDs[0] != "b" {
		t.Fatalf("first answer = %+v, want q1=b", got.Answers[0])
	}
	if got.Answers[1].QuestionID != "q2" || len(got.Answers[1].OptionIDs) != 2 {
		t.Fatalf("second answer = %+v, want two selected options", got.Answers[1])
	}
	if got.Answers[2].OptionIDs[0] != "score5" || got.Answers[3].OptionIDs[0] != "city2" {
		t.Fatalf("rating/dropdown answers = %+v/%+v, want score5/city2", got.Answers[2], got.Answers[3])
	}
}

func TestBuildAnswerPlanSupportsMinMaxAliases(t *testing.T) {
	got, err := BuildAnswerPlan(rand.New(rand.NewSource(2)), []QuestionPlan{
		{
			ID:   "q1",
			Kind: "multiple",
			Options: map[string]any{
				"min": float64(1),
				"max": int64(1),
			},
			Weights: []answer.OptionWeight{
				{OptionID: "a", Weight: 1},
				{OptionID: "b", Weight: 1},
			},
		},
	})
	if err != nil {
		t.Fatalf("BuildAnswerPlan() returned error: %v", err)
	}
	if len(got.Answers[0].OptionIDs) != 1 {
		t.Fatalf("OptionIDs = %+v, want one selected option", got.Answers[0].OptionIDs)
	}
}

func TestBuildAnswerPlans(t *testing.T) {
	questions := []QuestionPlan{
		{
			ID:   "q1",
			Kind: "single",
			Weights: []answer.OptionWeight{
				{OptionID: "a", Weight: 1},
				{OptionID: "b", Weight: 1},
			},
		},
	}

	plans, err := BuildAnswerPlans(rand.New(rand.NewSource(3)), questions, 3)
	if err != nil {
		t.Fatalf("BuildAnswerPlans() returned error: %v", err)
	}
	if len(plans) != 3 {
		t.Fatalf("len(plans) = %d, want 3", len(plans))
	}
	plans[0].Answers[0].QuestionID = "mutated"
	if plans[1].Answers[0].QuestionID != "q1" {
		t.Fatalf("plans are not independent: %+v", plans)
	}
}

func TestBuildAnswerPlansRejectsInvalidInput(t *testing.T) {
	questions := []QuestionPlan{{
		ID:      "q1",
		Kind:    "single",
		Weights: []answer.OptionWeight{{OptionID: "a", Weight: 1}},
	}}

	if _, err := BuildAnswerPlans(rand.New(rand.NewSource(1)), questions, 0); err == nil || !strings.Contains(err.Error(), "count") {
		t.Fatalf("BuildAnswerPlans(count=0) error = %v, want count error", err)
	}
	if _, err := BuildAnswerPlans(nil, questions, 1); err == nil || !strings.Contains(err.Error(), "answer plan 1") || !strings.Contains(err.Error(), "rng") {
		t.Fatalf("BuildAnswerPlans(nil rng) error = %v, want indexed rng error", err)
	}
}

func TestBuildAnswerPlanRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		name      string
		rng       *rand.Rand
		questions []QuestionPlan
		want      string
	}{
		{name: "rng", questions: []QuestionPlan{{ID: "q1", Kind: "single"}}, want: "rng"},
		{name: "questions", rng: rand.New(rand.NewSource(1)), want: "questions"},
		{name: "question id", rng: rand.New(rand.NewSource(1)), questions: []QuestionPlan{{Kind: "single"}}, want: "question id"},
		{name: "kind", rng: rand.New(rand.NewSource(1)), questions: []QuestionPlan{{ID: "q1", Kind: "text"}}, want: "not supported"},
		{name: "weights", rng: rand.New(rand.NewSource(1)), questions: []QuestionPlan{{ID: "q1", Kind: "single"}}, want: "weights"},
		{
			name: "multiple rule",
			rng:  rand.New(rand.NewSource(1)),
			questions: []QuestionPlan{{
				ID:      "q1",
				Kind:    "multiple",
				Options: map[string]any{"min": 2, "max": 1},
				Weights: []answer.OptionWeight{{OptionID: "a", Weight: 1}},
			}},
			want: "min must not be greater",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildAnswerPlan(tt.rng, tt.questions)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("BuildAnswerPlan() error = %v, want %q", err, tt.want)
			}
		})
	}
}
