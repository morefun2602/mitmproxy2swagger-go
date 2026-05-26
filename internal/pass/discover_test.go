package pass

import (
	"regexp"
	"testing"
)

func TestPathToRegex(t *testing.T) {
	re := pathToRegex("/users/{id}/items")
	if !re.MatchString("/users/42/items") {
		t.Fatal("expected parameterized template to match literal path")
	}
	if re.MatchString("/users/42/items/extra") {
		t.Fatal("expected template match to be anchored")
	}
}

func TestBuildSuggestedTemplates(t *testing.T) {
	paramRegex := regexp.MustCompile("^[0-9]+$")
	out := buildSuggestedTemplates([]string{"/users/42/profile"}, paramRegex, false)

	want := []string{"ignore:/users/{id}/profile", "ignore:/users/42/profile"}
	if len(out) != len(want) {
		t.Fatalf("got %d templates, want %d: %v", len(out), len(want), out)
	}
	for i, entry := range want {
		if out[i] != entry {
			t.Fatalf("out[%d] = %q, want %q (full: %v)", i, out[i], entry, out)
		}
	}
}

func TestBuildSuggestedTemplatesSuppressParams(t *testing.T) {
	paramRegex := regexp.MustCompile("^[0-9]+$")
	out := buildSuggestedTemplates([]string{"/users/42"}, paramRegex, true)

	if len(out) != 1 || out[0] != "ignore:/users/{id}" {
		t.Fatalf("unexpected templates with suppress-params: %v", out)
	}
}

func TestDiscoverPathDedupes(t *testing.T) {
	runner := &passRunner{seenNew: make(map[string]struct{})}
	runner.discoverPath("/items")
	runner.discoverPath("/items")
	runner.discoverPath("/users")

	if len(runner.newPathTemplates) != 2 {
		t.Fatalf("got %d paths, want 2: %v", len(runner.newPathTemplates), runner.newPathTemplates)
	}
}

func TestMatchPathTemplate(t *testing.T) {
	doc := newTestDocument(t)
	doc.XPathTemplates = []string{"/items/{id}"}
	runner := newPassRunner(doc, Options{}, "https://api.example.com/v1", regexp.MustCompile("^[0-9]+$"))

	if idx := runner.matchPathTemplate("/items/7"); idx != 0 {
		t.Fatalf("matchPathTemplate() = %d, want 0", idx)
	}
	if idx := runner.matchPathTemplate("/missing"); idx != -1 {
		t.Fatalf("matchPathTemplate() = %d, want -1", idx)
	}
}
