package curate

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
)

type templateEntry struct {
	ignore bool
	path   string
}

func parseTemplateEntry(raw string) templateEntry {
	if strings.HasPrefix(raw, "ignore:") {
		return templateEntry{ignore: true, path: strings.TrimPrefix(raw, "ignore:")}
	}
	return templateEntry{ignore: false, path: raw}
}

func formatTemplateEntry(e templateEntry) string {
	if e.ignore {
		return "ignore:" + e.path
	}
	return e.path
}

func isParametricPath(path string) bool {
	return strings.Contains(path, "{") && strings.Contains(path, "}")
}

// AutoTemplates applies conservative Template Clustering to x-path-templates entries.
func AutoTemplates(templates []string, numericRe *regexp.Regexp) []string {
	if len(templates) == 0 {
		return templates
	}

	entries := make([]templateEntry, 0, len(templates))
	for _, raw := range templates {
		entries = append(entries, parseTemplateEntry(raw))
	}

	var parametric []templateEntry
	var literals []templateEntry
	for _, e := range entries {
		if isParametricPath(e.path) {
			parametric = append(parametric, e)
		} else {
			literals = append(literals, e)
		}
	}

	parametric = dedupeEntries(parametric)

	var paramRegexes []*regexp.Regexp
	for _, p := range parametric {
		paramRegexes = append(paramRegexes, pathTemplateToRegex(p.path))
	}

	var uncovered []templateEntry
	for _, lit := range literals {
		if literalCoveredByParametric(lit.path, paramRegexes) {
			continue
		}
		uncovered = append(uncovered, lit)
	}

	generalized := make(map[string]templateEntry)
	for _, lit := range uncovered {
		gpath := generalizeNumericPath(lit.path, numericRe)
		if existing, ok := generalized[gpath]; ok {
			if lit.ignore {
				existing.ignore = true
			}
			generalized[gpath] = existing
			continue
		}
		generalized[gpath] = templateEntry{ignore: lit.ignore, path: gpath}
	}

	gkeys := make([]string, 0, len(generalized))
	for k := range generalized {
		gkeys = append(gkeys, k)
	}
	sort.Strings(gkeys)

	out := make([]templateEntry, 0, len(parametric)+len(generalized))
	out = append(out, parametric...)
	for _, k := range gkeys {
		out = append(out, generalized[k])
	}

	out = dedupeEntries(out)

	result := make([]string, len(out))
	for i, e := range out {
		result[i] = formatTemplateEntry(e)
	}
	schema.SortPathTemplates(result)
	return result
}

func literalCoveredByParametric(path string, paramRegexes []*regexp.Regexp) bool {
	for _, re := range paramRegexes {
		if re.MatchString(path) {
			return true
		}
	}
	return false
}

func generalizeNumericPath(path string, numericRe *regexp.Regexp) string {
	segments := strings.Split(path, "/")
	paramID := 0
	for i, seg := range segments {
		if seg != "" && numericRe.MatchString(seg) {
			name := "id"
			if paramID > 0 {
				name = fmt.Sprintf("id%d", paramID)
			}
			segments[i] = "{" + name + "}"
			paramID++
		}
	}
	return strings.Join(segments, "/")
}

func dedupeEntries(entries []templateEntry) []templateEntry {
	seen := make(map[string]templateEntry, len(entries))
	order := make([]string, 0, len(entries))
	for _, e := range entries {
		key := formatTemplateEntry(e)
		if prev, ok := seen[key]; ok {
			if e.ignore {
				prev.ignore = true
			}
			seen[key] = prev
			continue
		}
		seen[key] = e
		order = append(order, key)
	}
	out := make([]templateEntry, 0, len(order))
	for _, key := range order {
		out = append(out, seen[key])
	}
	return out
}

