package provider_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/provider/credamo"
	"github.com/LING71671/SurveyController-Go/internal/provider/tencent"
	"github.com/LING71671/SurveyController-Go/internal/provider/wjx"
)

func TestParserFixturesProduceStandardSurveys(t *testing.T) {
	tests := []struct {
		name     string
		parse    func(t *testing.T) domain.SurveyDefinition
		provider domain.ProviderID
		title    string
		kinds    []domain.QuestionKind
	}{
		{
			name: "wjx_html",
			parse: func(t *testing.T) domain.SurveyDefinition {
				t.Helper()
				file := openFixture(t, "wjx", "testdata", "survey.html")
				defer file.Close()
				survey, err := wjx.ParseHTML(file, "https://www.wjx.cn/vm/example.aspx")
				if err != nil {
					t.Fatalf("ParseHTML() returned error: %v", err)
				}
				return survey
			},
			provider: domain.ProviderWJX,
			title:    "问卷星 HTML 样例",
			kinds: []domain.QuestionKind{
				domain.QuestionKindSingle,
				domain.QuestionKindMultiple,
				domain.QuestionKindText,
				domain.QuestionKindRating,
			},
		},
		{
			name: "tencent_api",
			parse: func(t *testing.T) domain.SurveyDefinition {
				t.Helper()
				file := openFixture(t, "tencent", "testdata", "survey_api.json")
				defer file.Close()
				survey, err := tencent.ParseAPI(file, "https://wj.qq.com/s2/example")
				if err != nil {
					t.Fatalf("ParseAPI() returned error: %v", err)
				}
				return survey
			},
			provider: domain.ProviderTencent,
			title:    "腾讯问卷 API 样例",
			kinds: []domain.QuestionKind{
				domain.QuestionKindSingle,
				domain.QuestionKindMultiple,
				domain.QuestionKindDropdown,
				domain.QuestionKindText,
				domain.QuestionKindRating,
				domain.QuestionKindMatrix,
			},
		},
		{
			name: "credamo_dom_snapshot",
			parse: func(t *testing.T) domain.SurveyDefinition {
				t.Helper()
				file := openFixture(t, "credamo", "testdata", "snapshot.json")
				defer file.Close()
				survey, err := credamo.ParseSnapshot(file, "https://www.credamo.com/answer.html#/s/demo")
				if err != nil {
					t.Fatalf("ParseSnapshot() returned error: %v", err)
				}
				return survey
			},
			provider: domain.ProviderCredamo,
			title:    "Credamo DOM 样例",
			kinds: []domain.QuestionKind{
				domain.QuestionKindSingle,
				domain.QuestionKindDropdown,
				domain.QuestionKindSingle,
				domain.QuestionKindSingle,
				domain.QuestionKindText,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			survey := tt.parse(t)
			if err := survey.Validate(); err != nil {
				t.Fatalf("Validate() returned error: %v", err)
			}
			if survey.Provider != tt.provider {
				t.Fatalf("Provider = %q, want %q", survey.Provider, tt.provider)
			}
			if survey.Title != tt.title {
				t.Fatalf("Title = %q, want %q", survey.Title, tt.title)
			}
			if len(survey.Questions) != len(tt.kinds) {
				t.Fatalf("len(Questions) = %d, want %d", len(survey.Questions), len(tt.kinds))
			}
			for index, wantKind := range tt.kinds {
				question := survey.Questions[index]
				if question.Kind != wantKind {
					t.Fatalf("question %d kind = %q, want %q", index+1, question.Kind, wantKind)
				}
				if question.ID == "" || question.Title == "" {
					t.Fatalf("question %d missing stable id/title: %+v", index+1, question)
				}
			}
		})
	}
}

func TestParserFixturesPreserveProviderSignals(t *testing.T) {
	t.Run("wjx required and option values", func(t *testing.T) {
		file := openFixture(t, "wjx", "testdata", "survey.html")
		defer file.Close()
		survey, err := wjx.ParseHTML(file, "https://www.wjx.cn/vm/example.aspx")
		if err != nil {
			t.Fatalf("ParseHTML() returned error: %v", err)
		}
		first := survey.Questions[0]
		if !first.Required || len(first.Options) != 2 || first.Options[0].Value != "browser" {
			t.Fatalf("first WJX question = %+v, want required browser option", first)
		}
	})

	t.Run("tencent matrix rows", func(t *testing.T) {
		file := openFixture(t, "tencent", "testdata", "survey_api.json")
		defer file.Close()
		survey, err := tencent.ParseAPI(file, "https://wj.qq.com/s2/example")
		if err != nil {
			t.Fatalf("ParseAPI() returned error: %v", err)
		}
		matrix := survey.Questions[5]
		if matrix.Kind != domain.QuestionKindMatrix || len(matrix.Rows) != 2 || len(matrix.Options) != 2 {
			t.Fatalf("Tencent matrix = %+v, want rows and options", matrix)
		}
	})

	t.Run("credamo forced answers", func(t *testing.T) {
		file := openFixture(t, "credamo", "testdata", "snapshot.json")
		defer file.Close()
		survey, err := credamo.ParseSnapshot(file, "https://www.credamo.com/answer.html#/s/demo")
		if err != nil {
			t.Fatalf("ParseSnapshot() returned error: %v", err)
		}
		forced := survey.Questions[2]
		if forced.ProviderRaw["forced_option_index"] != 0 || forced.ProviderRaw["forced_option_text"] != "非常不满意" {
			t.Fatalf("Credamo forced question = %+v, want forced first option", forced)
		}
		text := survey.Questions[4]
		forcedTexts, ok := text.ProviderRaw["forced_texts"].([]string)
		if !ok || len(forcedTexts) != 1 || forcedTexts[0] != "你好" {
			t.Fatalf("Credamo forced text = %+v, want 你好", text.ProviderRaw["forced_texts"])
		}
	})
}

func openFixture(t *testing.T, parts ...string) *os.File {
	t.Helper()
	file, err := os.Open(filepath.Join(parts...))
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	return file
}
