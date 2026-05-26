package enrich

import "testing"

func TestApplyEnrichmentMerge(t *testing.T) {
	op := map[string]any{
		"summary": "GET items",
	}
	result := &EnrichmentResult{
		Summary:     "List items",
		Description: "Returns items for the tenant.",
		Tags:        []string{"catalog"},
	}
	applyEnrichment(op, result, false)
	if op["summary"] != "GET items" {
		t.Fatalf("summary overwritten without force: %#v", op["summary"])
	}
	if op["description"] != "Returns items for the tenant." {
		t.Fatalf("description not added: %#v", op["description"])
	}
}

func TestApplyEnrichmentForce(t *testing.T) {
	op := map[string]any{
		"summary": "GET items",
	}
	result := &EnrichmentResult{
		Summary: "List catalog items",
	}
	applyEnrichment(op, result, true)
	if op["summary"] != "List catalog items" {
		t.Fatalf("summary not overwritten with force: %#v", op["summary"])
	}
}

func TestParseEnrichmentResult(t *testing.T) {
	raw := []byte(`{"summary":"List items","description":"desc","tags":[],"operationId":"","parameterDescriptions":{},"requestBodyDescription":""}`)
	got, err := parseEnrichmentResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary != "List items" || got.Description != "desc" {
		t.Fatalf("unexpected parse: %#v", got)
	}
}

func TestParseEnrichmentResult_markdownFence(t *testing.T) {
	raw := []byte("```json\n{\"summary\":\"List items\",\"description\":\"desc\",\"tags\":[],\"operationId\":\"\",\"parameterDescriptions\":{},\"requestBodyDescription\":\"\"}\n```")
	got, err := parseEnrichmentResult(raw)
	if err != nil {
		t.Fatal(err)
	}
	if got.Summary != "List items" {
		t.Fatalf("unexpected parse: %#v", got)
	}
}
