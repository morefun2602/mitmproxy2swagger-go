package schema_test

import (
	"strings"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
)

func TestMarshalUsesTwoSpaceIndent(t *testing.T) {
	data, err := schema.Marshal(map[string]any{
		"info": map[string]any{
			"title": "example",
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "    title:") {
			t.Fatalf("expected 2-space indent, got 4-space: %q", line)
		}
	}
	if !strings.Contains(string(data), "  title:") {
		t.Fatalf("expected 2-space indented title, got:\n%s", data)
	}
}
