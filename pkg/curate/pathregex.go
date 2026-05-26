package curate

import (
	"regexp"
	"strings"
)

func pathTemplateToRegex(path string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(path)
	escaped = strings.ReplaceAll(escaped, "\\{", "(?P<")
	escaped = strings.ReplaceAll(escaped, "\\}", ">[^/]+)")
	escaped = strings.ReplaceAll(escaped, "\\*", ".*")
	re, err := regexp.Compile("^" + escaped + "$")
	if err != nil {
		return regexp.MustCompile("^$")
	}
	return re
}
