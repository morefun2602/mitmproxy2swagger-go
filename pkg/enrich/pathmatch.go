package enrich

import (
	"regexp"
	"strings"
)

func pathToRegex(pathTemplate string) *regexp.Regexp {
	escaped := regexp.QuoteMeta(pathTemplate)
	escaped = strings.ReplaceAll(escaped, "\\{", "(?P<")
	escaped = strings.ReplaceAll(escaped, "\\}", ">[^/]+)")
	escaped = strings.ReplaceAll(escaped, "\\*", ".*")
	re, err := regexp.Compile("^" + escaped + "$")
	if err != nil {
		return regexp.MustCompile("^$")
	}
	return re
}

func stripQueryString(rawURL string) string {
	if i := strings.Index(rawURL, "?"); i >= 0 {
		return rawURL[:i]
	}
	return rawURL
}

func stripAPIPrefix(rawURL, prefix string) string {
	return strings.TrimPrefix(rawURL, prefix)
}
