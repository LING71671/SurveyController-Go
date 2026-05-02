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
	if strings.TrimSpace(plan.Provider) != domain.ProviderWJX.String() {
		return WJXHTTPSubmissionPreview{}, fmt.Errorf("wjx http preview requires wjx provider")
	}
	if err := options.Survey.Validate(); err != nil {
		return WJXHTTPSubmissionPreview{}, fmt.Errorf("wjx survey: %w", err)
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
