package curate

import "strings"

// validateProposedMatchesGroup reports whether proposed matches every path in the group.
func validateProposedMatchesGroup(proposed string, entries []templateEntry) bool {
	path := strings.TrimPrefix(proposed, "ignore:")
	if path == "" {
		return false
	}
	re := pathTemplateToRegex(path)
	for _, e := range entries {
		if !re.MatchString(e.path) {
			return false
		}
	}
	return true
}
