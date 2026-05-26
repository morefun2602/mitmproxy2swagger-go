package curate

import (
	"regexp"
	"testing"
)

func TestApplySuggestions_mergesReplaces(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	in := []string{
		"ignore:/v1/apps/info/calendar",
		"ignore:/v1/apps/info/woa-meeting-assistant",
		"ignore:/v1/health",
	}
	suggestions := []TemplateSuggestion{{
		ProposedTemplate: "ignore:/v1/apps/info/{appKey}",
		Replaces: []string{
			"ignore:/v1/apps/info/calendar",
			"ignore:/v1/apps/info/woa-meeting-assistant",
		},
	}}
	out := applySuggestions(in, suggestions, numericRe)
	if len(out) != 2 {
		t.Fatalf("got %d templates, want 2: %v", len(out), out)
	}
	found := false
	for _, e := range out {
		if e == "ignore:/v1/apps/info/{appKey}" {
			found = true
		}
	}
	if !found {
		t.Fatalf("missing merged template: %v", out)
	}
}
