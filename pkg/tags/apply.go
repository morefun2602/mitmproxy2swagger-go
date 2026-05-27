package tags

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
)

// ApplyOptions configures tags apply.
type ApplyOptions struct {
	Schema   string
	TagsFile string
	Output   string
	Merge    bool
	Strict   bool
}

// ApplyResult reports how many operations were updated or left unmatched.
type ApplyResult struct {
	Updated    int
	Unmatched  int
	UnmatchedKeys []string
}

// RunApply writes operation tags and optional top-level tags / x-tagGroups from a tags sidecar.
func RunApply(opts ApplyOptions) (*ApplyResult, error) {
	if opts.Schema == "" {
		return nil, fmt.Errorf("schema path is required")
	}
	if opts.TagsFile == "" {
		opts.TagsFile = filepath.Join(filepath.Dir(opts.Schema), "tags.yaml")
	}
	out := opts.Output
	if out == "" {
		out = opts.Schema
	}

	cfg, err := LoadTagsFile(opts.TagsFile)
	if err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("%s: %w", opts.TagsFile, err)
	}

	doc, err := schema.Load(opts.Schema)
	if err != nil {
		return nil, err
	}

	resolver := newResolver(cfg)
	result := &ApplyResult{}

	for _, item := range doc.Paths {
		pathTemplate := fmt.Sprint(item.Key)
		ops, ok := item.Value.(map[string]any)
		if !ok {
			continue
		}
		for method, raw := range ops {
			if !isHTTPMethod(method) {
				continue
			}
			op, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			tag, matched := resolver.resolve(strings.ToUpper(method), pathTemplate)
			if !matched {
				result.Unmatched++
				result.UnmatchedKeys = append(result.UnmatchedKeys, strings.ToUpper(method)+" "+pathTemplate)
				continue
			}
			applyOperationTag(op, tag, opts.Merge)
			result.Updated++
		}
	}

	if opts.Strict && result.Unmatched > 0 {
		return result, fmt.Errorf("%d operation(s) unmatched in %s (see --help); first: %s",
			result.Unmatched, opts.TagsFile, result.UnmatchedKeys[0])
	}

	if len(cfg.Tags) > 0 {
		doc.Tags = tagDefsToOpenAPI(cfg.Tags)
	}
	if len(cfg.XTagGroups) > 0 {
		doc.XTagGroups = tagGroupsToOpenAPI(cfg.XTagGroups)
	}

	if err := doc.Save(out); err != nil {
		return nil, err
	}
	return result, nil
}

type resolver struct {
	operations map[string]string // "GET /path" -> primary tag
	prefixes   []prefixEntry
}

type prefixEntry struct {
	prefix string
	tag    string
}

func newResolver(cfg *TagsFile) *resolver {
	r := &resolver{
		operations: make(map[string]string),
	}
	for key, tags := range cfg.Operations {
		method, path, ok := parseOperationKey(key)
		if !ok {
			continue
		}
		r.operations[method+" "+path] = primaryTag(tags)
	}
	for _, rule := range cfg.PrefixRules {
		r.prefixes = append(r.prefixes, prefixEntry{
			prefix: rule.Prefix,
			tag:    primaryTag(rule.Tags),
		})
	}
	sort.Slice(r.prefixes, func(i, j int) bool {
		return len(r.prefixes[i].prefix) > len(r.prefixes[j].prefix)
	})
	return r
}

func (r *resolver) resolve(method, pathTemplate string) (tag string, ok bool) {
	key := method + " " + pathTemplate
	if t, ok := r.operations[key]; ok {
		return t, true
	}
	for _, pe := range r.prefixes {
		if strings.HasPrefix(pathTemplate, pe.prefix) {
			return pe.tag, true
		}
	}
	return "", false
}

func primaryTag(tags []string) string {
	for _, t := range tags {
		if strings.TrimSpace(t) != "" {
			return strings.TrimSpace(t)
		}
	}
	return ""
}

func applyOperationTag(op map[string]any, tag string, merge bool) {
	if tag == "" {
		return
	}
	if merge {
		existing := tagsFromOperation(op)
		combined := schema.DedupeStrings(append([]string{tag}, existing...))
		if len(combined) > 0 {
			op["tags"] = stringSliceToAny([]string{combined[0]})
		}
		return
	}
	op["tags"] = stringSliceToAny([]string{tag})
}

func tagsFromOperation(op map[string]any) []string {
	raw, ok := op["tags"].([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
			out = append(out, s)
		}
	}
	return out
}

func stringSliceToAny(in []string) []any {
	out := make([]any, len(in))
	for i, s := range in {
		out[i] = s
	}
	return out
}

func tagDefsToOpenAPI(defs []TagDef) []any {
	out := make([]any, len(defs))
	for i, d := range defs {
		m := map[string]any{"name": d.Name}
		if d.Description != "" {
			m["description"] = d.Description
		}
		out[i] = m
	}
	return out
}

func tagGroupsToOpenAPI(groups []TagGroup) []any {
	out := make([]any, len(groups))
	for i, g := range groups {
		tags := make([]any, len(g.Tags))
		for j, t := range g.Tags {
			tags[j] = t
		}
		out[i] = map[string]any{
			"name": g.Name,
			"tags": tags,
		}
	}
	return out
}

func isHTTPMethod(method string) bool {
	switch strings.ToLower(method) {
	case "get", "put", "post", "delete", "options", "head", "patch", "trace":
		return true
	default:
		return false
	}
}

// PrintApplySummary writes a short summary to stderr.
func PrintApplySummary(res *ApplyResult, outPath string) {
	if res == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "tags apply: updated %d operation(s) → %s\n", res.Updated, outPath)
	if res.Unmatched > 0 {
		fmt.Fprintf(os.Stderr, "tags apply: %d operation(s) unmatched (prefix/operations rules)\n", res.Unmatched)
	}
}
