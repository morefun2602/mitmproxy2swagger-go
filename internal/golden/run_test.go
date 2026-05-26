package golden_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/golden"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/pass"
)

func TestRunCasePreservesOpenAPI(t *testing.T) {
	root := repoRoot(t)
	actual, err := golden.RunCase(root, golden.Case{
		ID:        "minimal",
		Capture:   "testdata/captures/minimal.har",
		APIPrefix: "https://api.example.com/v1",
		ExtraArgs: []string{"--format", "har"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(string(actual), "openapi: 3.0.0") {
		t.Fatalf("expected openapi 3.0.0 in output:\n%s", actual)
	}
}

func containsString(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func TestStripIgnorePreservesOpenAPI(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.yaml")
	content := `openapi: 3.0.0
info:
  title: example
  version: 1.0.0
x-path-templates:
  - ignore:/items
`
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := golden.StripIgnorePrefixes(path); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(string(got), "openapi: 3.0.0") {
		t.Fatalf("openapi lost after strip:\n%s", got)
	}
}

func TestPassSecondPassPreservesOpenAPI(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.yaml")
	opts := pass.Options{
		Input:     filepath.Join(repoRoot(t), "testdata", "captures", "minimal.har"),
		Output:    out,
		APIPrefix: "https://api.example.com/v1",
		Format:    "har",
	}
	if err := pass.Run(opts); err != nil {
		t.Fatal(err)
	}
	if err := golden.StripIgnorePrefixes(out); err != nil {
		t.Fatal(err)
	}
	if err := pass.Run(opts); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !containsString(string(got), "openapi: 3.0.0") {
		t.Fatalf("openapi lost after second pass:\n%s", got)
	}
}
