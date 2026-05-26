package swaggerutil_test

import (
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/swaggerutil"
)

func TestPathTemplateToEndpointName(t *testing.T) {
	got := swaggerutil.PathTemplateToEndpointName("post", "/api/v1/things/{id}/create")
	want := "POST things create by id"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestValueToSchemaGenericKeys(t *testing.T) {
	schema := swaggerutil.ValueToSchema(map[string]any{
		"123": map[string]any{"a": "b"},
	})
	if _, ok := schema["additionalProperties"]; !ok {
		t.Fatalf("expected additionalProperties, got %#v", schema)
	}
	if _, ok := schema["properties"]; ok {
		t.Fatal("expected no properties for numeric keys")
	}
}
