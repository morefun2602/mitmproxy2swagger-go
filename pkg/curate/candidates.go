package curate

import (
	"regexp"
	"sort"
	"strings"
)

type mergeCandidate struct {
	prefix  string
	entries []templateEntry
}

func findSlugMergeCandidates(templates []string, numericRe *regexp.Regexp) []mergeCandidate {
	return findSuffixSlugMergeCandidates(templates, numericRe)
}

// findSuffixSlugMergeCandidates groups literal paths that share the same prefix
// and differ only in the last non-numeric segment (suffix slug).
func findSuffixSlugMergeCandidates(templates []string, numericRe *regexp.Regexp) []mergeCandidate {
	byPrefix := make(map[string][]templateEntry)
	for _, raw := range templates {
		e := parseTemplateEntry(raw)
		if isParametricPath(e.path) {
			continue
		}
		prefix, ok := pathPrefixKey(e.path, numericRe)
		if !ok {
			continue
		}
		byPrefix[prefix] = append(byPrefix[prefix], e)
	}

	keys := make([]string, 0, len(byPrefix))
	for k := range byPrefix {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var out []mergeCandidate
	for _, prefix := range keys {
		group := dedupeEntries(byPrefix[prefix])
		if len(group) < 2 {
			continue
		}
		if allPathsIdentical(group) {
			continue
		}
		if !hasDistinctLastSegments(group) {
			continue
		}
		out = append(out, mergeCandidate{prefix: prefix, entries: group})
	}
	return out
}

func pathPrefixKey(path string, numericRe *regexp.Regexp) (string, bool) {
	segs := nonEmptySegments(path)
	if len(segs) < 2 {
		return "", false
	}
	last := segs[len(segs)-1]
	if numericRe.MatchString(last) {
		return "", false
	}
	return "/" + strings.Join(segs[:len(segs)-1], "/"), true
}

func nonEmptySegments(path string) []string {
	parts := splitPath(path)
	out := make([]string, 0, len(parts))
	for _, s := range parts {
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func hasDistinctLastSegments(entries []templateEntry) bool {
	lasts := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		segs := nonEmptySegments(e.path)
		if len(segs) == 0 {
			continue
		}
		lasts[segs[len(segs)-1]] = struct{}{}
	}
	return len(lasts) >= 2
}

func allPathsIdentical(entries []templateEntry) bool {
	if len(entries) == 0 {
		return true
	}
	first := entries[0].path
	for _, e := range entries[1:] {
		if e.path != first {
			return false
		}
	}
	return true
}

func splitPath(path string) []string {
	return strings.Split(path, "/")
}

func joinPath(segments []string) string {
	return strings.Join(segments, "/")
}
