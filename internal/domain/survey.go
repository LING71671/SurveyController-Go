package domain

import (
	"fmt"
	"strings"
)

type SurveyDefinition struct {
	Provider    ProviderID
	ProviderRaw map[string]any
	ID          string
	Title       string
	Description string
	URL         string
	Pages       []PageDefinition
	Questions   []QuestionDefinition
}

type PageDefinition struct {
	Number      int
	Title       string
	Description string
	QuestionIDs []string
	ProviderRaw map[string]any
}

type QuestionDefinition struct {
	ID          string
	Number      int
	Title       string
	Description string
	Kind        QuestionKind
	Required    bool
	Options     []OptionDefinition
	Rows        []OptionDefinition
	Conditions  []ConditionDefinition
	ProviderRaw map[string]any
}

type OptionDefinition struct {
	ID          string
	Label       string
	Value       string
	Exclusive   bool
	ProviderRaw map[string]any
}

type ConditionDefinition struct {
	SourceQuestionID string
	Operator         ConditionOperator
	Values           []string
	ProviderRaw      map[string]any
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
