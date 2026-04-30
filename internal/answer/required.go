package answer

import "sort"

func MergeRequiredOptions(selected []string, required []string) SelectionResult {
	seen := map[string]bool{}
	for _, id := range selected {
		if id != "" {
			seen[id] = true
		}
	}
	for _, id := range required {
		if id != "" {
			seen[id] = true
		}
	}
	return SelectionResult{OptionIDs: sortedIDs(seen)}
}

func sortedIDs(values map[string]bool) []string {
	ids := make([]string, 0, len(values))
	for id := range values {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}
