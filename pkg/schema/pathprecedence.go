package schema

import (
	"sort"
	"strings"
)

func pathForPrecedence(template string) string {
	return strings.TrimPrefix(template, "ignore:")
}

// PathTemplatePrecedenceLess reports whether a should be tried before b when matching.
// More specific templates (fewer placeholders, longer path) come first.
func PathTemplatePrecedenceLess(a, b string) bool {
	pa, pb := pathForPrecedence(a), pathForPrecedence(b)
	ca, cb := strings.Count(pa, "{"), strings.Count(pb, "{")
	if ca != cb {
		return ca < cb
	}
	if len(pa) != len(pb) {
		return len(pa) > len(pb)
	}
	return pa < pb
}

// SortPathTemplates sorts in place for matching precedence (specific before wildcard).
func SortPathTemplates(templates []string) {
	sort.Slice(templates, func(i, j int) bool {
		return PathTemplatePrecedenceLess(templates[i], templates[j])
	})
}
