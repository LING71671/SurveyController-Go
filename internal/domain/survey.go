package domain

import (
	"fmt"
	"strings"
)

type SurveyDefinition struct {
	Provider    ProviderID           `json:"provider"`
	ProviderRaw map[string]any       `json:"provider_raw,omitempty"`
	ID          string               `json:"id,omitempty"`
	Title       string               `json:"title"`
	Description string               `json:"description,omitempty"`
	URL         string               `json:"url,omitempty"`
	Pages       []PageDefinition     `json:"pages,omitempty"`
	Questions   []QuestionDefinition `json:"questions"`
}

type PageDefinition struct {
	Number      int            `json:"number"`
	Title       string         `json:"title,omitempty"`
	Description string         `json:"description,omitempty"`
	QuestionIDs []string       `json:"question_ids,omitempty"`
	ProviderRaw map[string]any `json:"provider_raw,omitempty"`
}

type QuestionDefinition struct {
	ID          string                `json:"id"`
	Number      int                   `json:"number"`
	Title       string                `json:"title"`
	Description string                `json:"description,omitempty"`
	Kind        QuestionKind          `json:"kind"`
	Required    bool                  `json:"required,omitempty"`
	Options     []OptionDefinition    `json:"options,omitempty"`
	Rows        []OptionDefinition    `json:"rows,omitempty"`
	Conditions  []ConditionDefinition `json:"conditions,omitempty"`
	ProviderRaw map[string]any        `json:"provider_raw,omitempty"`
}

type OptionDefinition struct {
	ID          string         `json:"id"`
	Label       string         `json:"label"`
	Value       string         `json:"value,omitempty"`
	Exclusive   bool           `json:"exclusive,omitempty"`
	ProviderRaw map[string]any `json:"provider_raw,omitempty"`
}

type ConditionDefinition struct {
	SourceQuestionID string            `json:"source_question_id"`
	Operator         ConditionOperator `json:"operator"`
	Values           []string          `json:"values,omitempty"`
	ProviderRaw      map[string]any    `json:"provider_raw,omitempty"`
}

type ConditionOperator string

const (
	ConditionOperatorEquals    ConditionOperator = "equals"
	ConditionOperatorNotEquals ConditionOperator = "not_equals"
	ConditionOperatorContains  ConditionOperator = "contains"
	ConditionOperatorNotEmpty  ConditionOperator = "not_empty"
	ConditionOperatorAlways    ConditionOperator = "always"
)

func (s SurveyDefinition) Validate() error {
	if s.Provider == ProviderUnknown {
		return fmt.Errorf("survey provider is required")
	}
	if strings.TrimSpace(s.Title) == "" {
		return fmt.Errorf("survey title is required")
	}
	for i, question := range s.Questions {
		if err := question.Validate(); err != nil {
			return fmt.Errorf("question %d: %w", i+1, err)
		}
	}
	return nil
}

func (q QuestionDefinition) Validate() error {
	if strings.TrimSpace(q.ID) == "" {
		return fmt.Errorf("id is required")
	}
	if strings.TrimSpace(q.Title) == "" {
		return fmt.Errorf("title is required")
	}
	if !q.Kind.Valid() {
		return fmt.Errorf("kind %q is unsupported", q.Kind)
	}
	return nil
}
