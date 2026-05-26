package curate

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
)

func runApplySuggestions(opts Options) error {
	path := opts.ApplySuggestions
	if path == "" {
		return fmt.Errorf("required flag: --apply-suggestions")
	}

	paramRegex := opts.ParamRegex
	if paramRegex == "" {
		paramRegex = `[0-9]+`
	}
	numericRe, err := regexp.Compile("^" + paramRegex + "$")
	if err != nil {
		return fmt.Errorf("invalid path parameter regex: %w", err)
	}

	doc, err := schema.Load(opts.Schema)
	if err != nil {
		return err
	}

	file, err := loadSuggestionsFile(path)
	if err != nil {
		return err
	}

	before := len(doc.XPathTemplates)
	doc.XPathTemplates = applySuggestions(doc.XPathTemplates, file.Suggestions, numericRe)
	after := len(doc.XPathTemplates)

	if err := doc.Save(opts.Schema); err != nil {
		return err
	}
	fmt.Printf("curate --apply-suggestions: %d -> %d x-path-templates (%d merges applied)\n", before, after, len(file.Suggestions))
	return nil
}

func applySuggestions(templates []string, suggestions []TemplateSuggestion, numericRe *regexp.Regexp) []string {
	drop := make(map[string]struct{})
	for _, s := range suggestions {
		for _, r := range s.Replaces {
			drop[r] = struct{}{}
		}
	}

	kept := make([]string, 0, len(templates)+len(suggestions))
	for _, t := range templates {
		if _, remove := drop[t]; remove {
			continue
		}
		kept = append(kept, t)
	}
	for _, s := range suggestions {
		if strings.TrimSpace(s.ProposedTemplate) == "" {
			continue
		}
		kept = append(kept, s.ProposedTemplate)
	}
	return AutoTemplates(schema.DedupeStrings(kept), numericRe)
}
