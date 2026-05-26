package curate_test

import (
	"regexp"
	"strings"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/curate"
)

func TestAutoTemplates_dropsLiteralsCoveredByParametric(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	in := []string{
		"ignore:/v1/items/{id}/reviews",
		"ignore:/v1/items/11874305/reviews",
		"ignore:/v1/items/82378352/reviews",
	}
	out := curate.AutoTemplates(in, numericRe)

	if len(out) != 1 {
		t.Fatalf("got %d entries, want 1: %v", len(out), out)
	}
	if out[0] != "ignore:/v1/items/{id}/reviews" {
		t.Fatalf("unexpected entry: %q", out[0])
	}
}

func TestAutoTemplates_clustersNumericLiterals(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	in := []string{
		"ignore:/v1/items/11874305/reviews",
		"ignore:/v1/items/82378352/reviews",
	}
	out := curate.AutoTemplates(in, numericRe)

	if len(out) != 1 {
		t.Fatalf("got %d entries, want 1: %v", len(out), out)
	}
	if out[0] != "ignore:/v1/items/{id}/reviews" {
		t.Fatalf("unexpected entry: %q", out[0])
	}
}

func TestAutoTemplates_keepsAlphabeticSlugLiterals(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	in := []string{
		"ignore:/v1/catalog/info/electronics",
		"ignore:/v1/catalog/info/AK20211012PHTVVY",
	}
	out := curate.AutoTemplates(in, numericRe)

	if len(out) != 2 {
		t.Fatalf("got %d entries, want 2: %v", len(out), out)
	}
}

func TestAutoTemplates_parametricBeforeLiteralPrecedence(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	in := []string{
		"ignore:/v1/items/11874305/reviews",
		"ignore:/v1/items/{id}/reviews",
	}
	out := curate.AutoTemplates(in, numericRe)

	if len(out) != 1 {
		t.Fatalf("got %d entries, want 1: %v", len(out), out)
	}
	if !strings.Contains(out[0], "{id}") {
		t.Fatalf("expected parametric template: %q", out[0])
	}
}

func TestAutoTemplates_largeSubsetReduction(t *testing.T) {
	numericRe := regexp.MustCompile(`^[0-9]+$`)
	in := []string{
		"ignore:/v1/items/{id}/reviews",
		"ignore:/v1/items/11874305/reviews",
		"ignore:/v1/items/82378352/reviews",
		"ignore:/v1/items/83397588/reviews",
		"ignore:/v1/items/{id}/status",
		"ignore:/v1/items/11874305/status",
		"ignore:/v1/items/82378352/status",
		"ignore:/v1/catalog/info/electronics",
		"ignore:/v1/catalog/info/AK20211012PHTVVY",
		"ignore:/v1/accounts/{id}",
		"ignore:/v1/accounts/49106",
	}
	out := curate.AutoTemplates(in, numericRe)

	if len(out) >= len(in) {
		t.Fatalf("expected fewer templates: %d -> %d", len(in), len(out))
	}
	if len(out) > 6 {
		t.Fatalf("expected substantial reduction, got %d: %v", len(out), out)
	}
}
