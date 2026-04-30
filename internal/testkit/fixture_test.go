package testkit

import (
	"path/filepath"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/domain"
)

func TestLoadSurveyFixture(t *testing.T) {
	path := filepath.Join("testdata", "providers", "wjx", "survey.json")
	survey := LoadSurveyFixture(t, path)

	if survey.Provider != domain.ProviderWJX {
		t.Fatalf("Provider = %q, want %q", survey.Provider, domain.ProviderWJX)
	}
	if len(survey.Questions) != 2 {
		t.Fatalf("len(Questions) = %d, want 2", len(survey.Questions))
	}
}

func TestLoadSurveyFixtureValidatesAllProviderSamples(t *testing.T) {
	paths := []string{
		filepath.Join("testdata", "providers", "wjx", "survey.json"),
		filepath.Join("testdata", "providers", "tencent", "survey.json"),
		filepath.Join("testdata", "providers", "credamo", "survey.json"),
	}

	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			LoadSurveyFixture(t, path)
		})
	}
}

func TestAssertSurveyEqual(t *testing.T) {
	survey := domain.SurveyDefinition{
		Provider: domain.ProviderWJX,
		Title:    "title",
		Questions: []domain.QuestionDefinition{
			{
				ID:    "q1",
				Title: "question",
				Kind:  domain.QuestionKindSingle,
			},
		},
	}

	AssertSurveyEqual(t, survey, survey)
}
