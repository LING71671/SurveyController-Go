package app

import (
	"strings"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/runner"
)

func TestPreviewWJXHTTPSubmissionBuildsDraft(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, validWJXRunConfig()), RunPlanOverrides{})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}

	preview, err := PreviewWJXHTTPSubmission(plan, WJXHTTPPreviewOptions{
		Seed:   7,
		Survey: testWJXSurvey(),
	})
	if err != nil {
		t.Fatalf("PreviewWJXHTTPSubmission() error = %v", err)
	}

	if preview.Provider != "wjx" || preview.Mode != "http" || preview.Method != "POST" {
		t.Fatalf("preview = %+v, want wjx http POST", preview)
	}
	if preview.Endpoint != "https://www.wjx.cn/joinnew/processjq.ashx" || preview.SurveyID != "example" {
		t.Fatalf("preview endpoint/id = %q/%q, want WJX process endpoint and survey id", preview.Endpoint, preview.SurveyID)
	}
	if preview.AnswerCount != 1 || preview.Form["q1"][0] != "1" || preview.Form["curID"][0] != "example" {
		t.Fatalf("preview form = %+v, want generated answer draft", preview.Form)
	}
	if preview.Header["Referer"][0] != "https://www.wjx.cn/vm/example.aspx" {
		t.Fatalf("preview header = %+v, want referer", preview.Header)
	}
}

func TestPreviewWJXHTTPSubmissionRejectsInvalidInputs(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, validWJXRunConfig()), RunPlanOverrides{})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}

	tests := []struct {
		name string
		want string
	}{
		{
			name: "provider",
			want: "wjx provider",
		},
		{
			name: "survey",
			want: "url mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPlan := plan
			survey := testWJXSurvey()
			if tt.name == "provider" {
				testPlan.Provider = "mock"
			}
			if tt.name == "survey" {
				survey.URL = ""
			}

			_, err := PreviewWJXHTTPSubmission(testPlan, WJXHTTPPreviewOptions{Survey: survey})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("PreviewWJXHTTPSubmission() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestValidateWJXHTTPPreviewRejectsIncompatiblePlan(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, validWJXRunConfig()), RunPlanOverrides{})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}

	tests := []struct {
		name string
		mut  func(*runnerPlanForTest, *surveyForTest)
		want string
	}{
		{
			name: "mode",
			mut: func(p *runnerPlanForTest, s *surveyForTest) {
				p.Mode = "hybrid"
			},
			want: "http mode",
		},
		{
			name: "url",
			mut: func(p *runnerPlanForTest, s *surveyForTest) {
				s.URL = "https://www.wjx.cn/vm/other.aspx"
			},
			want: "url mismatch",
		},
		{
			name: "question",
			mut: func(p *runnerPlanForTest, s *surveyForTest) {
				p.Questions[0].ID = "missing"
			},
			want: "not present",
		},
		{
			name: "kind",
			mut: func(p *runnerPlanForTest, s *surveyForTest) {
				p.Questions[0].Kind = "multiple"
			},
			want: "kind mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPlan := plan
			testPlan.Questions = append([]runner.QuestionPlan(nil), plan.Questions...)
			testSurvey := testWJXSurvey()
			tt.mut((*runnerPlanForTest)(&testPlan), (*surveyForTest)(&testSurvey))

			err := ValidateWJXHTTPPreview(testPlan, testSurvey)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("ValidateWJXHTTPPreview() error = %v, want %q", err, tt.want)
			}
		})
	}
}

func TestPreviewWJXHTTPSubmissionClonesDraftMaps(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, validWJXRunConfig()), RunPlanOverrides{})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}

	first, err := PreviewWJXHTTPSubmission(plan, WJXHTTPPreviewOptions{Survey: testWJXSurvey()})
	if err != nil {
		t.Fatalf("PreviewWJXHTTPSubmission() error = %v", err)
	}
	first.Form["q1"][0] = "mutated"
	first.Header["Referer"][0] = "mutated"

	second, err := PreviewWJXHTTPSubmission(plan, WJXHTTPPreviewOptions{Survey: testWJXSurvey()})
	if err != nil {
		t.Fatalf("PreviewWJXHTTPSubmission() error = %v", err)
	}
	if second.Form["q1"][0] != "1" || second.Header["Referer"][0] != "https://www.wjx.cn/vm/example.aspx" {
		t.Fatalf("second preview = %+v, want independent maps", second)
	}
}

type runnerPlanForTest = runner.Plan
type surveyForTest = domain.SurveyDefinition
