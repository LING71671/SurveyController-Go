package wjx

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/LING71671/SurveyController-Go/internal/apperr"
	"github.com/LING71671/SurveyController-Go/internal/domain"
	"github.com/LING71671/SurveyController-Go/internal/provider"
	"golang.org/x/net/html"
)

type Provider struct{}

func (Provider) ID() provider.ProviderID {
	return domain.ProviderWJX
}

func (Provider) MatchURL(rawURL string) bool {
	return provider.MatchHostSuffix(rawURL, "wjx.cn", "wjx.top")
}

func (Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		ParseHTTP:      true,
		ParseBrowser:   true,
		SubmitHTTP:     true,
		SupportsHybrid: true,
	}
}

func (p Provider) Parse(ctx context.Context, rawURL string) (provider.SurveyDefinition, error) {
	_ = ctx
	_ = rawURL
	return provider.SurveyDefinition{}, apperr.New(apperr.CodeProviderUnsupported, "wjx provider requires HTML content for this parser prototype")
}

func ParseHTML(r io.Reader, rawURL string) (domain.SurveyDefinition, error) {
	root, err := html.Parse(r)
	if err != nil {
		return domain.SurveyDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "parse wjx html", err)
	}
	if err := detectBlockedState(root); err != nil {
		return domain.SurveyDefinition{}, err
	}

	title := firstTextByAttr(root, "data-survey-title")
	if title == "" {
		title = firstTextByTag(root, "title")
	}
	survey := domain.SurveyDefinition{
		Provider:  domain.ProviderWJX,
		Title:     strings.TrimSpace(title),
		URL:       rawURL,
		Questions: []domain.QuestionDefinition{},
		ProviderRaw: map[string]any{
			"source": "wjx_html",
		},
	}

	nodes := findAll(root, func(n *html.Node) bool {
		return attr(n, "data-question") != ""
	})
	for index, node := range nodes {
		question, err := parseQuestion(node, index+1)
		if err != nil {
			return domain.SurveyDefinition{}, err
		}
		survey.Questions = append(survey.Questions, question)
	}
	if err := survey.Validate(); err != nil {
		return domain.SurveyDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "validate wjx survey", err)
	}
	return survey, nil
}

func detectBlockedState(root *html.Node) error {
	state := strings.ToLower(strings.TrimSpace(firstAttr(root, "data-page-state")))
	text := strings.ToLower(textContent(root))
	switch {
	case state == "paused" || strings.Contains(text, "暂停"):
		return apperr.New(apperr.CodeParseFailed, "survey is paused")
	case state == "closed" || strings.Contains(text, "未开放") || strings.Contains(text, "已结束"):
		return apperr.New(apperr.CodeParseFailed, "survey is not open")
	case state == "verification" || strings.Contains(text, "验证码") || strings.Contains(text, "验证"):
		return apperr.New(apperr.CodeVerificationNeeded, "verification is required")
	default:
		return nil
	}
}

func parseQuestion(node *html.Node, fallbackNumber int) (domain.QuestionDefinition, error) {
	id := strings.TrimSpace(attr(node, "data-question"))
	kind, err := parseKind(attr(node, "data-kind"))
	if err != nil {
		return domain.QuestionDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "parse wjx question kind", err)
	}

	number := fallbackNumber
	if rawNumber := attr(node, "data-number"); rawNumber != "" {
		parsed, err := strconv.Atoi(rawNumber)
		if err != nil {
			return domain.QuestionDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "parse wjx question number", err)
		}
		number = parsed
	}

	question := domain.QuestionDefinition{
		ID:       id,
		Number:   number,
		Title:    strings.TrimSpace(firstTextByAttr(node, "data-question-title")),
		Kind:     kind,
		Required: attr(node, "data-required") == "true",
		Options:  parseOptions(node),
		ProviderRaw: map[string]any{
			"data_question": id,
		},
	}
	if question.Title == "" {
		return domain.QuestionDefinition{}, apperr.New(apperr.CodeParseFailed, fmt.Sprintf("question %q title is required", id))
	}
	if err := question.Validate(); err != nil {
		return domain.QuestionDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "validate wjx question", err)
	}
	return question, nil
}

func parseKind(raw string) (domain.QuestionKind, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "single", "radio":
		return domain.QuestionKindSingle, nil
	case "multiple", "checkbox":
		return domain.QuestionKindMultiple, nil
	case "dropdown", "select":
		return domain.QuestionKindDropdown, nil
	case "text", "input":
		return domain.QuestionKindText, nil
	case "textarea":
		return domain.QuestionKindTextarea, nil
	case "rating", "scale":
		return domain.QuestionKindRating, nil
	default:
		return domain.ParseQuestionKind(raw)
	}
}

func parseOptions(node *html.Node) []domain.OptionDefinition {
	optionNodes := findAll(node, func(n *html.Node) bool {
		return attr(n, "data-option") != ""
	})
	options := make([]domain.OptionDefinition, 0, len(optionNodes))
	for _, optionNode := range optionNodes {
		id := strings.TrimSpace(attr(optionNode, "data-option"))
		label := strings.TrimSpace(textContent(optionNode))
		value := strings.TrimSpace(attr(optionNode, "data-value"))
		if value == "" {
			value = id
		}
		options = append(options, domain.OptionDefinition{
			ID:    id,
			Label: label,
			Value: value,
			ProviderRaw: map[string]any{
				"data_option": id,
			},
		})
	}
	return options
}

func firstTextByAttr(root *html.Node, name string) string {
	node := first(root, func(n *html.Node) bool {
		return hasAttr(n, name)
	})
	if node == nil {
		return ""
	}
	return textContent(node)
}

func firstTextByTag(root *html.Node, tag string) string {
	node := first(root, func(n *html.Node) bool {
		return n.Type == html.ElementNode && strings.EqualFold(n.Data, tag)
	})
	if node == nil {
		return ""
	}
	return textContent(node)
}

func firstAttr(root *html.Node, name string) string {
	node := first(root, func(n *html.Node) bool {
		return hasAttr(n, name)
	})
	if node == nil {
		return ""
	}
	return attr(node, name)
}

func first(root *html.Node, match func(*html.Node) bool) *html.Node {
	if root == nil {
		return nil
	}
	if match(root) {
		return root
	}
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if found := first(child, match); found != nil {
			return found
		}
	}
	return nil
}

func findAll(root *html.Node, match func(*html.Node) bool) []*html.Node {
	var nodes []*html.Node
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node == nil {
			return
		}
		if match(node) {
			nodes = append(nodes, node)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return nodes
}

func attr(node *html.Node, name string) string {
	if node == nil {
		return ""
	}
	for _, attribute := range node.Attr {
		if attribute.Key == name {
			return attribute.Val
		}
	}
	return ""
}

func hasAttr(node *html.Node, name string) bool {
	if node == nil {
		return false
	}
	for _, attribute := range node.Attr {
		if attribute.Key == name {
			return true
		}
	}
	return false
}

func textContent(node *html.Node) string {
	if node == nil {
		return ""
	}
	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(current *html.Node) {
		if current.Type == html.TextNode {
			builder.WriteString(current.Data)
			builder.WriteByte(' ')
		}
		for child := current.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(node)
	return strings.Join(strings.Fields(builder.String()), " ")
}
