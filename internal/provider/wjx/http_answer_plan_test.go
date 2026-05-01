package wjx

import (
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/answerplan"
	"github.com/LING71671/SurveyController-go/internal/domain"
)

func TestBuildHTTPAnswers(t *testing.T) {
	survey := testAnswerPlanSurvey()

	got, err := BuildHTTPAnswers(survey, answerplan.Plan{
		Answers: []answerplan.QuestionAnswer{
			{QuestionID: "q1", OptionIDs: []string{"b"}},
			{QuestionID: "q2", OptionIDs: []string{"a", "c"}},
			{QuestionID: "q3", OptionIDs: []string{"score5"}},
			{QuestionID: "q4", OptionIDs: []string{"city2"}},
		},
	})
	if err != nil {
		t.Fatalf("BuildHTTPAnswers() returned error: %v", err)
	}

	want := map[string]string{
		"q1": "2",
		"q2": "A,C",
		"q3": "5",
		"q4": "shanghai",
	}
	if len(got) != len(want) {
		t.Fatalf("len(answers) = %d, want %d: %+v", len(got), len(want), got)
	}
	for key, value := range want {
		if got[key] != value {
			t.Fatalf("answers[%q] = %q, want %q", key, got[key], value)
		}
	}
}

func TestBuildHTTPSubmissionDraftFromAnswerPlan(t *testing.T) {
	draft, err := BuildHTTPSubmissionDraftFromAnswerPlan(testAnswerPlanSurvey(), answerplan.Plan{
		Answers: []answerplan.QuestionAnswer{
			{QuestionID: "q1", OptionIDs: []string{"a"}},
			{QuestionID: "q2", OptionIDs: []string{"b", "c"}},
			{QuestionID: "q3", Value: "4"},
		},
	})
	if err != nil {
		t.Fatalf("BuildHTTPSubmissionDraftFromAnswerPlan() returned error: %v", err)
	}

	if draft.SurveyID != "answerplan" {
		t.Fatalf("SurveyID = %q, want answerplan", draft.SurveyID)
	}
	if draft.Form.Get("q1") != "1" || draft.Form.Get("q2") != "B,C" || draft.Form.Get("q3") != "4" {
		t.Fatalf("Form = %+v, want mapped answers", draft.Form)
	}
}

func TestBuildHTTPAnswersSupportsDirectValues(t *testing.T) {
	survey := testAnswerPlanSurvey()

	got, err := BuildHTTPAnswers(survey, answerplan.Plan{
		Answers: []answerplan.QuestionAnswer{
			{QuestionID: "q1", Value: "2"},
			{QuestionID: "q2", Value: "A,C"},
			{QuestionID: "q3", Value: "5"},
		},
	})
	if err != nil {
		t.Fatalf("BuildHTTPAnswers() returned error: %v", err)
	}
	if got["q1"] != "2" || got["q2"] != "A,C" || got["q3"] != "5" {
		t.Fatalf("answers = %+v, want direct values", got)
	}
}

func TestBuildHTTPAnswersRejectsInvalidPlan(t *testing.T) {
	survey := testAnswerPlanSurvey()
	tests := []struct {
		name   string
		survey domain.SurveyDefinition
		plan   answerplan.Plan
		want   string
	}{
		{
			name:   "empty plan",
			survey: survey,
			want:   "answer plan",
		},
		{
			name:   "missing question id",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: " ", Value: "1"},
			}},
			want: "question id",
		},
		{
			name:   "undefined question",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "missing", Value: "1"},
			}},
			want: "not defined",
		},
		{
			name:   "duplicate question",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q1", Value: "1"},
				{QuestionID: "q1", Value: "2"},
			}},
			want: "duplicate",
		},
		{
			name:   "duplicate survey question definition",
			survey: appendQuestion(survey, survey.Questions[0]),
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q1", Value: "1"},
			}},
			want: "defined more than once",
		},
		{
			name:   "duplicate option definition",
			survey: replaceQuestion(survey, 0, duplicateOptionQuestion()),
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q1", OptionIDs: []string{"a"}},
			}},
			want: "defined more than once",
		},
		{
			name:   "unsupported kind",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q5", Value: "hello"},
			}},
			want: "not supported",
		},
		{
			name:   "single multiple options",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q1", OptionIDs: []string{"a", "b"}},
			}},
			want: "expects one option",
		},
		{
			name:   "unknown option",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q1", OptionIDs: []string{"missing"}},
			}},
			want: "not defined",
		},
		{
			name:   "duplicate option",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q2", OptionIDs: []string{"a", "a"}},
			}},
			want: "more than once",
		},
		{
			name:   "empty direct value",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q3", Value: " "},
			}},
			want: "answer value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := BuildHTTPAnswers(tt.survey, tt.plan)
			if err == nil {
				t.Fatalf("BuildHTTPAnswers() returned nil error, want %q", tt.want)
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("BuildHTTPAnswers() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func appendQuestion(survey domain.SurveyDefinition, question domain.QuestionDefinition) domain.SurveyDefinition {
	survey.Questions = append(append([]domain.QuestionDefinition(nil), survey.Questions...), question)
	return survey
}

func replaceQuestion(survey domain.SurveyDefinition, index int, question domain.QuestionDefinition) domain.SurveyDefinition {
	survey.Questions = append([]domain.QuestionDefinition(nil), survey.Questions...)
	survey.Questions[index] = question
	return survey
}

func duplicateOptionQuestion() domain.QuestionDefinition {
	return domain.QuestionDefinition{
		ID:    "q1",
		Title: "Single",
		Kind:  domain.QuestionKindSingle,
		Options: []domain.OptionDefinition{
			{ID: "a", Label: "A", Value: "1"},
			{ID: "a", Label: "A again", Value: "2"},
		},
	}
}

func testAnswerPlanSurvey() domain.SurveyDefinition {
	return domain.SurveyDefinition{
		Provider: domain.ProviderWJX,
		Title:    "Answer Plan",
		URL:      "https://www.wjx.cn/vm/answerplan.aspx",
		Questions: []domain.QuestionDefinition{
			{
				ID:    "q1",
				Title: "Single",
				Kind:  domain.QuestionKindSingle,
				Options: []domain.OptionDefinition{
					{ID: "a", Label: "A", Value: "1"},
					{ID: "b", Label: "B", Value: "2"},
				},
			},
			{
				ID:    "q2",
				Title: "Multiple",
				Kind:  domain.QuestionKindMultiple,
				Options: []domain.OptionDefinition{
					{ID: "a", Label: "A", Value: "A"},
					{ID: "b", Label: "B", Value: "B"},
					{ID: "c", Label: "C", Value: "C"},
				},
			},
			{
				ID:    "q3",
				Title: "Rating",
				Kind:  domain.QuestionKindRating,
				Options: []domain.OptionDefinition{
					{ID: "score4", Label: "4", Value: "4"},
					{ID: "score5", Label: "5", Value: "5"},
				},
			},
			{
				ID:    "q4",
				Title: "Dropdown",
				Kind:  domain.QuestionKindDropdown,
				Options: []domain.OptionDefinition{
					{ID: "city1", Label: "Beijing", Value: "beijing"},
					{ID: "city2", Label: "Shanghai", Value: "shanghai"},
				},
			},
			{
				ID:    "q5",
				Title: "Text",
				Kind:  domain.QuestionKindText,
			},
		},
	}
}
