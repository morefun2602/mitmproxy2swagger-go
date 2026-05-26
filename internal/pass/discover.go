package pass

import (
	"fmt"
	"regexp"
	"strings"
)

func (r *passRunner) discoverPath(path string) {
	if _, exists := r.seenNew[path]; exists {
		return
	}
	r.seenNew[path] = struct{}{}
	r.newPathTemplates = append(r.newPathTemplates, path)
}

func buildSuggestedTemplates(paths []string, paramRegex *regexp.Regexp, suppressParams bool) []string {
	var out []string
	seen := make(map[string]struct{})

	isParam := func(segment string) bool {
		return paramRegex.MatchString(segment)
	}

	for _, path := range paths {
		segments := strings.Split(path, "/")
		hasParam := false
		for _, seg := range segments {
			if seg != "" && isParam(seg) {
				hasParam = true
				break
			}
		}

		if hasParam {
			newSegments := make([]string, 0, len(segments))
			paramID := 0
			for _, seg := range segments {
				if seg != "" && isParam(seg) {
					name := "id"
					if paramID > 0 {
						name = fmt.Sprintf("id%d", paramID)
					}
					newSegments = append(newSegments, "{"+name+"}")
					paramID++
				} else {
					newSegments = append(newSegments, seg)
				}
			}
			suggested := strings.Join(newSegments, "/")
			entry := "ignore:" + suggested
			if _, ok := seen[entry]; !ok {
				seen[entry] = struct{}{}
				out = append(out, entry)
			}
		}

		if !hasParam || !suppressParams {
			entry := "ignore:" + path
			if _, ok := seen[entry]; !ok {
				seen[entry] = struct{}{}
				out = append(out, entry)
			}
		}
	}
	return out
}

func pathToRegex(path string) *regexp.Regexp {
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

func stripQueryString(rawURL string) string {
	if i := strings.Index(rawURL, "?"); i >= 0 {
		return rawURL[:i]
	}
	return rawURL
}

func stripAPIPrefix(rawURL, prefix string) string {
	return strings.TrimPrefix(rawURL, prefix)
}
