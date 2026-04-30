package testkit

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/LING71671/SurveyController-go/internal/domain"
)

func LoadSurveyFixture(t testing.TB, path string) domain.SurveyDefinition {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read survey fixture %q: %v", path, err)
	}

	var survey domain.SurveyDefinition
	if err := json.Unmarshal(data, &survey); err != nil {
		t.Fatalf("parse survey fixture %q: %v", path, err)
	}
	if err := survey.Validate(); err != nil {
		t.Fatalf("validate survey fixture %q: %v", path, err)
	}
	return survey
}

func AssertSurveyEqual(t testing.TB, got, want domain.SurveyDefinition) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		gotJSON, _ := json.MarshalIndent(got, "", "  ")
		wantJSON, _ := json.MarshalIndent(want, "", "  ")
		t.Fatalf("survey mismatch\ngot:\n%s\nwant:\n%s", gotJSON, wantJSON)
	}
}
