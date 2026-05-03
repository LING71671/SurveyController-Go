package app

import (
	"context"
	"fmt"
	"testing"

	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/runner"
)

var benchmarkWJXHTTPDryRunResult WJXHTTPDryRunResult

func BenchmarkRunWJXHTTPDryRun(b *testing.B) {
	for _, target := range []int{10, 1000} {
		b.Run(fmt.Sprintf("target_%d", target), func(b *testing.B) {
			benchmarkRunWJXHTTPDryRun(b, target)
		})
	}
}

func benchmarkRunWJXHTTPDryRun(b *testing.B, target int) {
	b.Helper()
	plan, err := CompileRunPlanFromFile(writeRunConfig(b, benchmarkWJXHTTPRunConfig()), RunPlanOverrides{
		Target:      target,
		Concurrency: target,
	})
	if err != nil {
		b.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}
	survey := benchmarkWJXHTTPSurvey()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result, err := RunWJXHTTPDryRun(context.Background(), plan, WJXHTTPRunOptions{
			Seed:   int64(i + 1),
			Survey: survey,
		})
		if err != nil {
			b.Fatalf("RunWJXHTTPDryRun() error = %v", err)
		}
		if result.Report.Successes != target || len(result.Drafts) != target {
			b.Fatalf("result = %+v drafts=%d, want %d successes and drafts", result.Report, len(result.Drafts), target)
		}
		benchmarkWJXHTTPDryRunResult = result
	}
}

func benchmarkWJXHTTPRunConfig() string {
	return `schema_version: 1
survey:
  url: "https://www.wjx.cn/vm/benchmark.aspx"
  provider: "wjx"
run:
  target: 1
  concurrency: 1
  mode: http
  failure_threshold: 1
  fail_stop_enabled: true
questions:
  - id: q1
    kind: single
    options:
      weights:
        - option_id: a
          weight: 1
  - id: q2
    kind: multiple
    options:
      min_selected: 2
      max_selected: 2
      weights:
        - option_id: a
          weight: 1
        - option_id: b
          weight: 1
        - option_id: c
          weight: 0
  - id: q3
    kind: text
    options:
      text:
        mode: fixed
        values:
          - benchmark answer
  - id: q4
    kind: rating
    options:
      weights:
        - option_id: score5
          weight: 1
  - id: q5
    kind: matrix
    options:
      matrix_weights:
        - row_id: row1
          weights:
            - option_id: agree
              weight: 1
        - row_id: row2
          weights:
            - option_id: neutral
              weight: 1
`
}

func benchmarkWJXHTTPSurvey() domain.SurveyDefinition {
	return domain.SurveyDefinition{
		Provider: domain.ProviderWJX,
		Title:    "WJX HTTP Benchmark",
		URL:      "https://www.wjx.cn/vm/benchmark.aspx",
		Questions: []domain.QuestionDefinition{
			{
				ID:    "q1",
				Title: "Single",
				Kind:  domain.QuestionKindSingle,
				Options: []domain.OptionDefinition{
					{ID: "a", Label: "A", Value: "1"},
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
				Title: "Text",
				Kind:  domain.QuestionKindText,
			},
			{
				ID:    "q4",
				Title: "Rating",
				Kind:  domain.QuestionKindRating,
				Options: []domain.OptionDefinition{
					{ID: "score5", Label: "5", Value: "5"},
				},
			},
			{
				ID:    "q5",
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
		},
	}
}

func TestBenchmarkWJXHTTPDryRunFixturesStayCompatible(t *testing.T) {
	plan, err := CompileRunPlanFromFile(writeRunConfig(t, benchmarkWJXHTTPRunConfig()), RunPlanOverrides{
		Target:      3,
		Concurrency: 2,
	})
	if err != nil {
		t.Fatalf("CompileRunPlanFromFile() error = %v", err)
	}

	result, err := RunWJXHTTPDryRun(context.Background(), plan, WJXHTTPRunOptions{
		Seed:   7,
		Survey: benchmarkWJXHTTPSurvey(),
	})
	if err != nil {
		t.Fatalf("RunWJXHTTPDryRun() error = %v", err)
	}
	assertBenchmarkDryRunReport(t, result.Report, 3)
	if len(result.Drafts) != 3 {
		t.Fatalf("drafts = %d, want 3", len(result.Drafts))
	}
	firstDraft := result.Drafts[0]
	if firstDraft.AnswerCount != 5 {
		t.Fatalf("AnswerCount = %d, want 5", firstDraft.AnswerCount)
	}
	if got := firstDraft.Form["q3"][0]; got != "benchmark answer" {
		t.Fatalf("text draft = %q, want direct mapped text answer", got)
	}
	if got := firstDraft.Form["q5"][0]; got != "row1:5;row2:3" {
		t.Fatalf("matrix draft = %q, want row-level mapped answer", got)
	}
}

func assertBenchmarkDryRunReport(t *testing.T, report runner.RunPlanReport, target int) {
	t.Helper()
	if report.Target != target || report.Successes != target || report.Failures != 0 || report.Completed != target {
		t.Fatalf("report = %+v, want successful target %d", report, target)
	}
	if report.TotalAllocDelta == 0 {
		t.Fatalf("report = %+v, want allocation metrics", report)
	}
}
