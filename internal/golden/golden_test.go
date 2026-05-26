package golden_test

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/golden"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestStripIgnorePrefixes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.yaml")
	if err := os.WriteFile(path, []byte(`openapi: 3.0.0
x-path-templates:
  - ignore:/users/{id}
  - ignore:/post
`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := golden.StripIgnorePrefixes(path); err != nil {
		t.Fatal(err)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(got, []byte("ignore:")) {
		t.Fatalf("expected ignore: prefixes removed, got %q", got)
	}
	if !bytes.Contains(got, []byte("/users/{id}")) {
		t.Fatalf("expected curated templates in %q", got)
	}
}

func TestCompareYAMLMatchesSelf(t *testing.T) {
	data := []byte(`openapi: 3.0.0
info:
  title: ignored
  version: 1.0.0
paths:
  /items:
    get:
      summary: GET items
`)
	if err := golden.CompareYAML(data, bytes.Clone(data)); err != nil {
		t.Fatal(err)
	}
}

func TestCompareYAMLIgnoresTitle(t *testing.T) {
	expected := []byte(`openapi: 3.0.0
info:
  title: old/path.har Mitmproxy2Swagger
  version: 1.0.0
`)
	actual := []byte(`openapi: 3.0.0
info:
  title: new/path.har Mitmproxy2Swagger
  version: 1.0.0
`)
	if err := golden.CompareYAML(expected, actual); err != nil {
		t.Fatal(err)
	}
}

func TestCompareYAMLDetectsDiff(t *testing.T) {
	expected := []byte(`openapi: 3.0.0
info:
  version: 1.0.0
paths:
  /a:
    get:
      summary: GET a
`)
	actual := []byte(`openapi: 3.0.0
info:
  version: 1.0.0
paths:
  /b:
    get:
      summary: GET b
`)
	err := golden.CompareYAML(expected, actual)
	if err == nil {
		t.Fatal("expected mismatch error")
	}
}

func TestCasesCapturePathsStayInRepo(t *testing.T) {
	root := repoRoot(t)
	cases, err := golden.LoadCases(filepath.Join(root, "testdata", "cases.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		if filepath.IsAbs(c.Capture) {
			t.Fatalf("[%s] capture must be relative: %s", c.ID, c.Capture)
		}
		if strings.Contains(c.Capture, "..") {
			t.Fatalf("[%s] capture must not escape repository: %s", c.ID, c.Capture)
		}
		if _, err := golden.ResolveCapture(root, c.Capture); err != nil {
			t.Fatalf("[%s] %v", c.ID, err)
		}
	}
}

func TestVerifyOne_har_lisek(t *testing.T) {
	root := repoRoot(t)
	cases, err := golden.LoadCases(filepath.Join(root, "testdata", "cases.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	cases, err = golden.FilterCases(cases, []string{"har_lisek"})
	if err != nil {
		t.Fatal(err)
	}
	if len(cases) != 1 {
		t.Fatalf("expected 1 case, got %d", len(cases))
	}
	if strings.Contains(cases[0].Capture, "..") {
		t.Fatalf("case capture must not escape repository: %s", cases[0].Capture)
	}
	if err := golden.VerifyOne(root, filepath.Join(root, "testdata", "golden"), cases[0]); err != nil {
		t.Fatal(err)
	}
}
