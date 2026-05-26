package pass_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/pass"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
)

func TestPass_templatePrecedenceSpecificFirst(t *testing.T) {
	dir := t.TempDir()
	capture := filepath.Join(dir, "capture.har")
	har := `{
  "log": {
    "entries": [{
      "request": {
        "method": "GET",
        "url": "https://api.example.com/v1/items/1/detail",
        "headers": []
      },
      "response": {
        "status": 200,
        "content": { "text": "{}" }
      }
    }]
  }
}`
	if err := os.WriteFile(capture, []byte(har), 0o644); err != nil {
		t.Fatal(err)
	}

	tmpPath := filepath.Join(dir, "schema.yaml")
	schemaYAML := `openapi: 3.0.0
info:
  title: test
  version: 1.0.0
x-path-templates:
  - /items/{id}
  - /items/{id}/detail
`
	if err := os.WriteFile(tmpPath, []byte(schemaYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := pass.Run(pass.Options{
		Input:     capture,
		Output:    tmpPath,
		APIPrefix: "https://api.example.com/v1",
		Format:    "har",
	}); err != nil {
		t.Fatal(err)
	}

	doc, err := schema.Load(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := doc.PathOperations("/items/{id}"); ok {
		t.Fatal("expected /items/{id}/detail to win over /items/{id}")
	}
	if _, ok := doc.PathOperations("/items/{id}/detail"); !ok {
		t.Fatalf("expected specific path materialized, paths: %#v", doc.Paths)
	}
}
