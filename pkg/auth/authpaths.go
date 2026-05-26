package auth

import (
	"sort"
	"strings"
)

var defaultAuthPathKeywords = []string{
	"auth", "token", "session", "login", "oauth", "refresh",
}

func suggestAuthPaths(paths map[string]struct{}, keywords []string) []SuggestedAuthPath {
	if len(keywords) == 0 {
		keywords = defaultAuthPathKeywords
	}
	var out []SuggestedAuthPath
	seen := make(map[string]struct{})
	for path := range paths {
		lower := strings.ToLower(path)
		segments := strings.Split(strings.Trim(lower, "/"), "/")
		var matched string
		for _, seg := range segments {
			for _, kw := range keywords {
				if strings.Contains(seg, kw) {
					matched = kw
					break
				}
			}
			if matched != "" {
				break
			}
		}
		if matched == "" {
			continue
		}
		if _, dup := seen[path]; dup {
			continue
		}
		seen[path] = struct{}{}
		out = append(out, SuggestedAuthPath{
			Path:   path,
			Reason: "path segment contains \"" + matched + "\"",
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Path < out[j].Path })
	return out
}
