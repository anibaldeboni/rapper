package processor

import (
	"fmt"
	"sort"
	"strings"

	"github.com/anibaldeboni/rapper/internal/web"
)

type MissingFieldsError struct {
	Missing     []string
	Suggestions map[string][]string
}

func (e *MissingFieldsError) Error() string {
	quoted := make([]string, len(e.Missing))
	for i, field := range e.Missing {
		quoted[i] = fmt.Sprintf("%q", field)
	}

	message := "missing csv fields: " + strings.Join(quoted, ", ")
	if len(e.Suggestions) == 0 {
		return message
	}

	suggestions := make([]string, 0, len(e.Suggestions))
	for _, field := range e.Missing {
		values := e.Suggestions[field]
		if len(values) == 0 {
			continue
		}
		suggestions = append(suggestions, fmt.Sprintf("%q: %s", field, quoteFields(values)))
	}
	if len(suggestions) == 0 {
		return message
	}

	return message + "; did you mean " + strings.Join(suggestions, ", ") + "?"
}

func validateHeaders(headers, required []string) error {
	headerSet := make(map[string]struct{}, len(headers))
	for _, header := range headers {
		headerSet[header] = struct{}{}
	}

	missingSet := make(map[string]struct{})
	for _, field := range required {
		if _, ok := headerSet[field]; !ok {
			missingSet[field] = struct{}{}
		}
	}
	if len(missingSet) == 0 {
		return nil
	}

	missing := make([]string, 0, len(missingSet))
	for field := range missingSet {
		missing = append(missing, field)
	}
	sort.Strings(missing)

	candidates := make([]string, 0, len(headers))
	for _, header := range headers {
		if header != "" {
			candidates = append(candidates, header)
		}
	}

	var suggestions map[string][]string
	if len(candidates) > 0 {
		suggestions = make(map[string][]string, len(missing))
		for _, field := range missing {
			suggestions[field] = web.SuggestFields(field, candidates)
		}
	}

	return &MissingFieldsError{Missing: missing, Suggestions: suggestions}
}

func quoteFields(fields []string) string {
	quoted := make([]string, len(fields))
	for i, field := range fields {
		quoted[i] = fmt.Sprintf("%q", field)
	}
	return strings.Join(quoted, ", ")
}
