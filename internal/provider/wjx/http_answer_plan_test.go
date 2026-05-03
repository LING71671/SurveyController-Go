package wjx

import (
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/answerplan"
	"github.com/LING71671/SurveyController-Go/internal/domain"
)

func TestBuildHTTPAnswers(t *testing.T) {
	survey := testAnswerPlanSurvey()

	got, err := BuildHTTPAnswers(survey, answerplan.Plan{
		Answers: []answerplan.QuestionAnswer{
			{QuestionID: "q1", OptionIDs: []string{"b"}},
			{QuestionID: "q2", OptionIDs: []string{"a", "c"}},
			{QuestionID: "q3", OptionIDs: []string{"score5"}},
			{QuestionID: "q4", OptionIDs: []string{"city2"}},
			{QuestionID: "q5", Value: "local text"},
			{QuestionID: "q6", Rows: []answerplan.RowAnswer{
				{RowID: "row1", OptionIDs: []string{"agree"}},
				{RowID: "row2", OptionIDs: []string{"neutral"}},
			}},
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
		"q5": "local text",
		"q6": "row1:5;row2:3",
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

func TestHTTPAnswerSchemaBuildSubmissionDraft(t *testing.T) {
	schema, err := CompileHTTPAnswerSchema(testAnswerPlanSurvey())
	if err != nil {
		t.Fatalf("CompileHTTPAnswerSchema() returned error: %v", err)
	}

	draft, err := schema.BuildSubmissionDraft(answerplan.Plan{
		Answers: []answerplan.QuestionAnswer{
			{QuestionID: "q1", OptionIDs: []string{"a"}},
			{QuestionID: "q2", OptionIDs: []string{"b", "c"}},
			{QuestionID: "q3", Value: "4"},
		},
	})
	if err != nil {
		t.Fatalf("BuildSubmissionDraft() returned error: %v", err)
	}

	if draft.Endpoint != "https://www.wjx.cn/joinnew/processjq.ashx" {
		t.Fatalf("Endpoint = %q, want process endpoint", draft.Endpoint)
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

func TestBuildHTTPAnswersSupportsMatrixRows(t *testing.T) {
	survey := testAnswerPlanSurvey()

	got, err := BuildHTTPAnswers(survey, answerplan.Plan{
		Answers: []answerplan.QuestionAnswer{
			{QuestionID: "q6", Rows: []answerplan.RowAnswer{
				{RowID: "row1", OptionIDs: []string{"agree"}},
				{RowID: "row2", Value: "4"},
			}},
		},
	})
	if err != nil {
		t.Fatalf("BuildHTTPAnswers() returned error: %v", err)
	}
	if got["q6"] != "row1:5;row2:4" {
		t.Fatalf("answers[q6] = %q, want row mapped matrix answer", got["q6"])
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
			name:   "duplicate matrix row definition",
			survey: replaceQuestion(survey, 5, duplicateMatrixRowQuestion()),
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q6", Rows: []answerplan.RowAnswer{{RowID: "row1", OptionIDs: []string{"agree"}}}},
			}},
			want: "defined more than once",
		},
		{
			name:   "unsupported kind",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q7", Value: "hello"},
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
		{
			name:   "text option ids",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q5", OptionIDs: []string{"a"}},
			}},
			want: "direct value",
		},
		{
			name:   "text row answers",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q5", Rows: []answerplan.RowAnswer{{RowID: "row1", Value: "x"}}},
			}},
			want: "row answers",
		},
		{
			name:   "matrix top-level options",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q6", OptionIDs: []string{"agree"}},
			}},
			want: "row answers",
		},
		{
			name:   "matrix unknown row",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q6", Rows: []answerplan.RowAnswer{{RowID: "missing", OptionIDs: []string{"agree"}}}},
			}},
			want: "not defined",
		},
		{
			name:   "matrix duplicate row",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q6", Rows: []answerplan.RowAnswer{
					{RowID: "row1", OptionIDs: []string{"agree"}},
					{RowID: "row1", OptionIDs: []string{"neutral"}},
				}},
			}},
			want: "more than once",
		},
		{
			name:   "matrix row multiple options",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q6", Rows: []answerplan.RowAnswer{{RowID: "row1", OptionIDs: []string{"agree", "neutral"}}}},
			}},
			want: "expects one option",
		},
		{
			name:   "matrix row unknown option",
			survey: survey,
			plan: answerplan.Plan{Answers: []answerplan.QuestionAnswer{
				{QuestionID: "q6", Rows: []answerplan.RowAnswer{{RowID: "row1", OptionIDs: []string{"missing"}}}},
			}},
			want: "not defined",
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

func duplicateMatrixRowQuestion() domain.QuestionDefinition {
	return domain.QuestionDefinition{
		ID:    "q6",
		Title: "Matrix",
		Kind:  domain.QuestionKindMatrix,
		Rows: []domain.OptionDefinition{
			{ID: "row1", Label: "UX"},
			{ID: "row1", Label: "UX again"},
		},
		Options: []domain.OptionDefinition{
			{ID: "agree", Label: "Agree", Value: "5"},
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
			{
				ID:    "q6",
				Title: "Matrix",
				Kind:  domain.QuestionKindMatrix,
				Rows: []domain.OptionDefinition{
					{ID: "row1", Label: "UX"},
					{ID: "row2", Label: "Performance"},
				},
				Options: []domain.OptionDefinition{
					{ID: "agree", Label: "Agree", Value: "5"},
					{ID: "neutral", Label: "Neutral", Value: "3"},
				},
			},
			{
				ID:    "q7",
				Title: "Unsupported",
				Kind:  domain.QuestionKindUnknown,
			},
		},
	}
}
