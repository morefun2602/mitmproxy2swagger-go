package pass_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/golden"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/pass"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/schema"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestPassTwoPassHAR(t *testing.T) {
	root := repoRoot(t)
	capture := filepath.Join(root, "testdata", "captures", "minimal.har")
	tmp, err := os.CreateTemp("", "pass-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	tmpPath := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpPath)

	opts := pass.Options{
		Input:     capture,
		Output:    tmpPath,
		APIPrefix: "https://api.example.com/v1",
		Format:    "har",
	}

	if err := pass.Run(opts); err != nil {
		t.Fatalf("first pass: %v", err)
	}
	if err := golden.StripIgnorePrefixes(tmpPath); err != nil {
		t.Fatalf("curation: %v", err)
	}
	if err := pass.Run(opts); err != nil {
		t.Fatalf("second pass: %v", err)
	}

	doc, err := schema.Load(tmpPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Paths) < 1 {
		t.Fatalf("expected paths, got %d", len(doc.Paths))
	}
	ops, ok := doc.PathOperations("/items")
	if !ok {
		t.Fatalf("paths: %#v", doc.Paths)
	}
	if _, ok := ops["get"]; !ok {
		t.Fatalf("expected GET /items, got %#v", ops)
	}
	if _, ok := ops["post"]; !ok {
		t.Fatalf("expected POST /items, got %#v", ops)
	}
}
