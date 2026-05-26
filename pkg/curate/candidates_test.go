package curate

import (
	"regexp"
	"testing"
)

func TestFindSlugMergeCandidates_groupsAppInfoPaths(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	templates := []string{
		"ignore:/v1/apps/info/calendar",
		"ignore:/v1/apps/info/woa-meeting-assistant",
		"ignore:/v1/chats/{id}/messages",
	}
	groups := findSlugMergeCandidates(templates, numericRe)
	if len(groups) != 1 {
		t.Fatalf("got %d groups, want 1", len(groups))
	}
	if groups[0].prefix != "/v1/apps/info" {
		t.Fatalf("prefix = %q, want /v1/apps/info", groups[0].prefix)
	}
	if len(groups[0].entries) != 2 {
		t.Fatalf("got %d entries, want 2", len(groups[0].entries))
	}
}

func TestFindSlugMergeCandidates_doesNotGroupUnrelatedPrefixes(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	templates := []string{
		"ignore:/v1/apps/info/calendar",
		"ignore:/v3/recent/chats/delta",
		"ignore:/v1/session/self/status",
		"ignore:/v2/gpt/assistants/doc_reader",
	}
	groups := findSlugMergeCandidates(templates, numericRe)
	if len(groups) != 0 {
		t.Fatalf("got %d groups, want 0 (unrelated prefixes): %v", len(groups), groups)
	}
}

func TestFindSlugMergeCandidates_excludesNumericLastSegment(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	templates := []string{
		"ignore:/v1/chats/11874305/messages",
		"ignore:/v1/chats/82378352/messages",
	}
	groups := findSlugMergeCandidates(templates, numericRe)
	if len(groups) != 0 {
		t.Fatalf("got %d groups, want 0 (numeric suffix handled by --auto)", len(groups))
	}
}

func TestFindSlugMergeCandidates_multiplePrefixGroups(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	templates := []string{
		"ignore:/v1/apps/info/calendar",
		"ignore:/v1/apps/info/AK20211012PHTVVY",
		"ignore:/v3/apps/info/calendar",
		"ignore:/v3/apps/info/meetingRoom",
	}
	groups := findSlugMergeCandidates(templates, numericRe)
	if len(groups) != 2 {
		t.Fatalf("got %d groups, want 2", len(groups))
	}
}
