package app

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/provider/wjx"
	"github.com/LING71671/SurveyController-Go/internal/runner"
)

type WJXHTTPPreviewOptions struct {
	Seed   int64
	Survey domain.SurveyDefinition
}

type WJXHTTPSubmissionPreview struct {
	Provider    string              `json:"provider"`
	Mode        string              `json:"mode"`
	Method      string              `json:"method"`
	Endpoint    string              `json:"endpoint"`
	SurveyID    string              `json:"survey_id"`
	Header      map[string][]string `json:"header"`
	Form        map[string][]string `json:"form"`
	AnswerCount int                 `json:"answer_count"`
}

func PreviewWJXHTTPSubmission(plan runner.Plan, options WJXHTTPPreviewOptions) (WJXHTTPSubmissionPreview, error) {
	if err := ValidateWJXHTTPPreview(plan, options.Survey); err != nil {
		return WJXHTTPSubmissionPreview{}, err
	}
	seed := options.Seed
	if seed == 0 {
		seed = 1
	}

	answerPlan, err := runner.BuildAnswerPlan(rand.New(rand.NewSource(seed)), plan.Questions)
	if err != nil {
		return WJXHTTPSubmissionPreview{}, err
	}
	draft, err := wjx.BuildHTTPSubmissionDraftFromAnswerPlan(options.Survey, answerPlan)
	if err != nil {
		return WJXHTTPSubmissionPreview{}, err
	}
	return previewFromWJXHTTPDraft(plan, draft, len(answerPlan.Answers)), nil
}

func ValidateWJXHTTPPreview(plan runner.Plan, survey domain.SurveyDefinition) error {
	return validateWJXHTTPPlanCompatibility(plan, survey, "preview")
}

func validateWJXHTTPPlanCompatibility(plan runner.Plan, survey domain.SurveyDefinition, operation string) error {
	if strings.TrimSpace(plan.Provider) != domain.ProviderWJX.String() {
		return fmt.Errorf("wjx http %s requires wjx provider", operation)
	}
	if plan.Mode.String() != "http" {
		return fmt.Errorf("wjx http %s requires http mode", operation)
	}
	if strings.TrimSpace(plan.URL) != strings.TrimSpace(survey.URL) {
		return fmt.Errorf("wjx http %s url mismatch: plan %q, survey %q", operation, plan.URL, survey.URL)
	}
	if err := survey.Validate(); err != nil {
		return fmt.Errorf("wjx survey: %w", err)
	}
	questions, err := wjxPreviewQuestionIndex(survey.Questions)
	if err != nil {
		return err
	}
	for _, question := range plan.Questions {
		id := strings.TrimSpace(question.ID)
		if id == "" {
			return fmt.Errorf("wjx http preview question id is required")
		}
		surveyQuestion, ok := questions[id]
		if !ok {
			return fmt.Errorf("wjx http preview question %q is not present in survey fixture", id)
		}
		kind := strings.TrimSpace(question.Kind)
		if kind == "" {
			return fmt.Errorf("wjx http preview question %q kind is required", id)
		}
		if surveyQuestion.Kind.String() != kind {
			return fmt.Errorf("wjx http preview question %q kind mismatch: plan %q, survey %q", id, kind, surveyQuestion.Kind)
		}
	}
	return nil
}

func wjxPreviewQuestionIndex(questions []domain.QuestionDefinition) (map[string]domain.QuestionDefinition, error) {
	index := make(map[string]domain.QuestionDefinition, len(questions))
	for _, question := range questions {
		id := strings.TrimSpace(question.ID)
		if id == "" {
			return nil, fmt.Errorf("wjx survey question id is required")
		}
		if _, exists := index[id]; exists {
			return nil, fmt.Errorf("wjx survey question %q is defined more than once", id)
		}
		index[id] = question
	}
	return index, nil
}

func previewFromWJXHTTPDraft(plan runner.Plan, draft wjx.HTTPSubmissionDraft, answerCount int) WJXHTTPSubmissionPreview {
	return WJXHTTPSubmissionPreview{
		Provider:    strings.TrimSpace(plan.Provider),
		Mode:        plan.Mode.String(),
		Method:      draft.Method,
		Endpoint:    draft.Endpoint,
		SurveyID:    draft.SurveyID,
		Header:      cloneHeaderMap(draft.Header),
		Form:        cloneValuesMap(draft.Form),
		AnswerCount: answerCount,
	}
}

func cloneHeaderMap(header http.Header) map[string][]string {
	if len(header) == 0 {
		return nil
	}
	cloned := make(map[string][]string, len(header))
	for key, values := range header {
		cloned[key] = append([]string(nil), values...)
	}
	return cloned
}

func cloneValuesMap(values url.Values) map[string][]string {
	if len(values) == 0 {
		return nil
	}
	cloned := make(map[string][]string, len(values))
	for key, items := range values {
		cloned[key] = append([]string(nil), items...)
	}
	return cloned
}
