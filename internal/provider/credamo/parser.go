package credamo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/LING71671/SurveyController-go/internal/apperr"
	"github.com/LING71671/SurveyController-go/internal/domain"
	"github.com/LING71671/SurveyController-go/internal/provider"
)

var (
	spaceRE                  = regexp.MustCompile(`\s+`)
	questionNumberRE         = regexp.MustCompile(`(?i)^\s*(?:Q|题目?)\s*(\d+)\b`)
	leadingTypeTagRE         = regexp.MustCompile(`^(?:(?:\[[^\]]+\]|【[^】]+】)\s*)+`)
	forceSelectCommandRE     = regexp.MustCompile(`请(?:务必|一定|必须|直接)?\s*选(?:择)?`)
	forceSelectIndexRE       = regexp.MustCompile(`^第?\s*(\d{1,3})\s*(?:个|项|选项|分|星)?$`)
	forceSelectSplitRE       = regexp.MustCompile(`[。；;！？!\n\r]`)
	forceSelectCleanRE       = regexp.MustCompile(`[\s` + "`" + `'"“”‘’【】\[\]\(\)（）<>《》,，、。；;:：!?！？]`)
	forceSelectLabelTargetRE = regexp.MustCompile(`^([A-Za-z])(?:项|选项|答案)?$`)
	forceSelectOptionLabelRE = regexp.MustCompile(`^(?:第\s*)?[\(（【\[]?\s*([A-Za-z])\s*[\)）】\]]?(?:$|[\.．、:：\-\s]|[\p{Han}])`)
	arithmeticExprRE         = regexp.MustCompile(`(?:^|[^\d.])(\d+(?:\.\d+)?(?:\s*[+\-*/×xX÷]\s*\d+(?:\.\d+)?)+)(?:$|[^\d.])`)
	optionNumberRE           = regexp.MustCompile(`-?\d+(?:\.\d+)?`)
	forceTextRE              = regexp.MustCompile(`请(?:务必|一定|必须|直接)?\s*(?:输入|填写|填入|写入)\s*[：:\s]*["“'‘]?([^"”'’\s，,。；;！!？?）)]+)`)
)

type Provider struct{}

func (Provider) ID() provider.ProviderID {
	return domain.ProviderCredamo
}

func (Provider) MatchURL(rawURL string) bool {
	return provider.MatchHostSuffix(rawURL, "credamo.com")
}

func (Provider) Capabilities() provider.Capabilities {
	return provider.Capabilities{
		ParseBrowser: true,
	}
}

func (p Provider) Parse(ctx context.Context, rawURL string) (provider.SurveyDefinition, error) {
	_ = ctx
	_ = rawURL
	return provider.SurveyDefinition{}, apperr.New(apperr.CodeProviderUnsupported, "credamo provider requires DOM snapshot JSON for this parser prototype")
}

type snapshot struct {
	ID          string             `json:"id"`
	Title       string             `json:"title"`
	Description string             `json:"description"`
	URL         string             `json:"url"`
	Questions   []snapshotQuestion `json:"questions"`
}

type snapshotQuestion struct {
	QuestionID    string   `json:"question_id"`
	QuestionNum   string   `json:"question_num"`
	Title         string   `json:"title"`
	TitleFullText string   `json:"title_full_text"`
	TitleText     string   `json:"title_text"`
	TipText       string   `json:"tip_text"`
	BodyText      string   `json:"body_text"`
	OptionTexts   []string `json:"option_texts"`
	InputTypes    []string `json:"input_types"`
	TextInputs    int      `json:"text_inputs"`
	Required      bool     `json:"required"`
	ProviderType  string   `json:"provider_type"`
	QuestionKind  string   `json:"question_kind"`
	Page          int      `json:"page"`
}

func ParseSnapshot(r io.Reader, rawURL string) (domain.SurveyDefinition, error) {
	var input snapshot
	decoder := json.NewDecoder(r)
	if err := decoder.Decode(&input); err != nil {
		return domain.SurveyDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "parse credamo dom snapshot json", err)
	}

	survey := domain.SurveyDefinition{
		Provider:    domain.ProviderCredamo,
		ID:          normalizeText(input.ID),
		Title:       firstNonEmpty(input.Title, "Credamo 见数问卷"),
		Description: normalizeText(input.Description),
		URL:         firstNonEmpty(input.URL, rawURL),
		Questions:   []domain.QuestionDefinition{},
		ProviderRaw: map[string]any{
			"source": "credamo_dom_snapshot",
		},
	}

	seen := map[string]struct{}{}
	for index, raw := range input.Questions {
		question, err := normalizeQuestion(raw, index+1)
		if err != nil {
			return domain.SurveyDefinition{}, err
		}
		key := dedupeKey(question)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		survey.Questions = append(survey.Questions, question)
	}
	if len(survey.Questions) == 0 {
		return domain.SurveyDefinition{}, apperr.New(apperr.CodeParseFailed, "credamo snapshot contains no questions")
	}
	if err := survey.Validate(); err != nil {
		return domain.SurveyDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "validate credamo survey", err)
	}
	return survey, nil
}

func normalizeQuestion(raw snapshotQuestion, fallbackNumber int) (domain.QuestionDefinition, error) {
	rawTitle := firstNonEmpty(raw.TitleFullText, raw.Title)
	number := normalizeQuestionNumber(raw.QuestionNum, fallbackNumber)
	title := rawTitle
	if matches := questionNumberRE.FindStringSubmatch(rawTitle); len(matches) == 2 {
		number = normalizeQuestionNumber(matches[1], fallbackNumber)
		title = normalizeText(rawTitle[len(matches[0]):])
		title = normalizeText(leadingTypeTagRE.ReplaceAllString(title, ""))
		if title == "" {
			title = rawTitle
		}
	}
	if title == "" {
		title = fmt.Sprintf("Q%d", number)
	}

	optionTexts := compactTexts(raw.OptionTexts)
	kind := inferKind(raw, optionTexts)
	options := makeOptions(optionTexts)
	forcedIndex, forcedText := extractForcedOption(rawTitle, optionTexts, raw.TitleText, raw.TipText)
	if forcedIndex == nil {
		forcedIndex, forcedText = extractArithmeticOption(rawTitle, optionTexts, raw.TitleText, raw.TipText)
	}
	forcedTexts := extractForcedTexts(rawTitle, raw.TitleText, raw.TipText)

	page := raw.Page
	if page <= 0 {
		page = 1
	}
	providerType := firstNonEmpty(raw.ProviderType, raw.QuestionKind, kind.String())
	questionID := firstNonEmpty(raw.QuestionID, strconv.Itoa(number))
	question := domain.QuestionDefinition{
		ID:       questionID,
		Number:   number,
		Title:    title,
		Kind:     kind,
		Required: raw.Required,
		Options:  options,
		ProviderRaw: map[string]any{
			"provider_type":        providerType,
			"provider_page_id":     strconv.Itoa(page),
			"text_inputs":          maxInt(raw.TextInputs, 0),
			"is_multi_text":        raw.TextInputs > 1,
			"forced_option_index":  forcedOptionIndexValue(forcedIndex),
			"forced_option_text":   forcedText,
			"forced_texts":         forcedTexts,
			"original_question_id": raw.QuestionID,
		},
	}
	if err := question.Validate(); err != nil {
		return domain.QuestionDefinition{}, apperr.Wrap(apperr.CodeParseFailed, "validate credamo question", err)
	}
	return question, nil
}

func inferKind(raw snapshotQuestion, optionTexts []string) domain.QuestionKind {
	kind := strings.ToLower(firstNonEmpty(raw.QuestionKind, raw.ProviderType))
	inputTypes := map[string]struct{}{}
	for _, inputType := range raw.InputTypes {
		inputTypes[strings.ToLower(normalizeText(inputType))] = struct{}{}
	}
	switch {
	case kind == "multiple":
		return domain.QuestionKindMultiple
	case kind == "dropdown":
		return domain.QuestionKindDropdown
	case kind == "scale":
		return domain.QuestionKindRating
	case kind == "order":
		return domain.QuestionKindRanking
	case kind == "single":
		return domain.QuestionKindSingle
	case kind == "text" || kind == "multi_text":
		return domain.QuestionKindText
	case hasInputType(inputTypes, "checkbox"):
		return domain.QuestionKindMultiple
	case hasInputType(inputTypes, "radio"):
		return domain.QuestionKindSingle
	case raw.TextInputs > 0 && len(optionTexts) == 0:
		return domain.QuestionKindText
	case len(optionTexts) >= 2:
		return domain.QuestionKindSingle
	default:
		return domain.QuestionKindText
	}
}

func hasInputType(inputTypes map[string]struct{}, name string) bool {
	_, ok := inputTypes[name]
	return ok
}

func extractForcedOption(title string, optionTexts []string, extras ...string) (*int, string) {
	if len(optionTexts) == 0 {
		return nil, ""
	}
	fragments := compactTexts(append([]string{title}, extras...))
	for _, fragment := range fragments {
		matches := forceSelectCommandRE.FindAllStringIndex(fragment, -1)
		for _, match := range matches {
			tail := strings.Trim(fragment[match[1]:], " ：:，,、")
			if tail == "" {
				continue
			}
			sentence := strings.Trim(forceSelectSplitRE.Split(tail, 2)[0], " ：:，,、")
			compact := normalizeForceSelectText(sentence)
			if compact == "" {
				continue
			}
			if index, text, ok := matchOptionByText(compact, optionTexts); ok {
				return &index, text
			}
			if label := forceSelectLabelTargetRE.FindStringSubmatch(compact); len(label) == 2 {
				target := strings.ToUpper(label[1])
				for index, optionText := range optionTexts {
					if extractOptionLabel(optionText) == target {
						found := index
						return &found, normalizeText(optionText)
					}
				}
			}
			if indexMatch := forceSelectIndexRE.FindStringSubmatch(sentence); len(indexMatch) == 2 {
				index, err := strconv.Atoi(indexMatch[1])
				if err == nil && index >= 1 && index <= len(optionTexts) {
					found := index - 1
					return &found, normalizeText(optionTexts[found])
				}
			}
		}
	}
	return nil, ""
}

func extractArithmeticOption(title string, optionTexts []string, extras ...string) (*int, string) {
	for _, fragment := range compactTexts(append([]string{title}, extras...)) {
		for _, match := range arithmeticExprRE.FindAllStringSubmatch(fragment, -1) {
			if len(match) < 2 {
				continue
			}
			result, ok := evalArithmetic(match[1])
			if !ok {
				continue
			}
			for index, optionText := range optionTexts {
				value, ok := extractNumericOptionValue(optionText)
				if ok && math.Abs(value-result) < 1e-9 {
					found := index
					return &found, normalizeText(optionText)
				}
			}
		}
	}
	return nil, ""
}

func extractForcedTexts(title string, extras ...string) []string {
	seen := map[string]struct{}{}
	var values []string
	for _, fragment := range compactTexts(append([]string{title}, extras...)) {
		for _, match := range forceTextRE.FindAllStringSubmatch(fragment, -1) {
			if len(match) < 2 {
				continue
			}
			text := normalizeText(match[1])
			if text == "" {
				continue
			}
			if _, ok := seen[text]; ok {
				continue
			}
			seen[text] = struct{}{}
			values = append(values, text)
		}
	}
	return values
}

func matchOptionByText(compactSentence string, optionTexts []string) (int, string, bool) {
	bestIndex := -1
	bestText := ""
	bestLength := -1
	for index, optionText := range optionTexts {
		text := normalizeText(optionText)
		normalized := normalizeForceSelectText(text)
		if normalized == "" || isDigits(normalized) {
			continue
		}
		if strings.Contains(compactSentence, normalized) && len(normalized) > bestLength {
			bestIndex = index
			bestText = text
			bestLength = len(normalized)
		}
	}
	return bestIndex, bestText, bestIndex >= 0
}

func evalArithmetic(expression string) (float64, bool) {
	tokens, ok := tokenizeArithmetic(expression)
	if !ok || len(tokens) == 0 {
		return 0, false
	}
	values := []float64{tokens[0].value}
	ops := []rune{}
	for i := 1; i < len(tokens); i++ {
		op := tokens[i].op
		value := tokens[i].value
		switch op {
		case '*':
			values[len(values)-1] *= value
		case '/':
			if math.Abs(value) < 1e-12 {
				return 0, false
			}
			values[len(values)-1] /= value
		case '+', '-':
			ops = append(ops, op)
			values = append(values, value)
		default:
			return 0, false
		}
	}
	result := values[0]
	for i, op := range ops {
		if op == '+' {
			result += values[i+1]
		} else {
			result -= values[i+1]
		}
	}
	return result, true
}

type arithmeticToken struct {
	op    rune
	value float64
}

func tokenizeArithmetic(expression string) ([]arithmeticToken, bool) {
	text := strings.NewReplacer("×", "*", "x", "*", "X", "*", "÷", "/").Replace(expression)
	text = strings.ReplaceAll(text, " ", "")
	if text == "" {
		return nil, false
	}
	var tokens []arithmeticToken
	op := '+'
	start := 0
	for i, r := range text {
		if i == 0 && (r == '+' || r == '-') {
			continue
		}
		if strings.ContainsRune("+-*/", r) {
			value, err := strconv.ParseFloat(text[start:i], 64)
			if err != nil {
				return nil, false
			}
			tokens = append(tokens, arithmeticToken{op: op, value: value})
			op = r
			start = i + 1
		}
	}
	value, err := strconv.ParseFloat(text[start:], 64)
	if err != nil {
		return nil, false
	}
	tokens = append(tokens, arithmeticToken{op: op, value: value})
	return tokens, true
}

func extractNumericOptionValue(optionText string) (float64, bool) {
	match := optionNumberRE.FindString(optionText)
	if match == "" {
		return 0, false
	}
	value, err := strconv.ParseFloat(match, 64)
	return value, err == nil
}

func extractOptionLabel(optionText string) string {
	match := forceSelectOptionLabelRE.FindStringSubmatch(normalizeText(optionText))
	if len(match) != 2 {
		return ""
	}
	return strings.ToUpper(match[1])
}

func dedupeKey(question domain.QuestionDefinition) string {
	pageID := fmt.Sprint(question.ProviderRaw["provider_page_id"])
	return fmt.Sprintf("page:%s|id:%s|num:%d|title:%s", pageID, question.ID, question.Number, question.Title)
}

func forcedOptionIndexValue(index *int) any {
	if index == nil {
		return nil
	}
	return *index
}

func makeOptions(texts []string) []domain.OptionDefinition {
	options := make([]domain.OptionDefinition, 0, len(texts))
	for index, text := range texts {
		id := strconv.Itoa(index + 1)
		options = append(options, domain.OptionDefinition{
			ID:    id,
			Label: text,
			Value: id,
			ProviderRaw: map[string]any{
				"index": index,
			},
		})
	}
	return options
}

func normalizeQuestionNumber(raw string, fallback int) int {
	match := regexp.MustCompile(`\d+`).FindString(raw)
	if match == "" {
		return maxInt(fallback, 1)
	}
	value, err := strconv.Atoi(match)
	if err != nil {
		return maxInt(fallback, 1)
	}
	return maxInt(value, 1)
}

func compactTexts(values []string) []string {
	result := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		text := normalizeText(value)
		if text == "" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		result = append(result, text)
	}
	return result
}

func normalizeText(value string) string {
	return strings.TrimSpace(spaceRE.ReplaceAllString(value, " "))
}

func normalizeForceSelectText(value string) string {
	return strings.ToLower(forceSelectCleanRE.ReplaceAllString(normalizeText(value), ""))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		text := normalizeText(value)
		if text != "" {
			return text
		}
	}
	return ""
}

func isDigits(value string) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
