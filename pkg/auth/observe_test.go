package auth_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/auth"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func TestRunObserve_authSample(t *testing.T) {
	root := repoRoot(t)
	capture := filepath.Join(root, "testdata", "captures", "auth_sample.har")
	dir := t.TempDir()
	out := filepath.Join(dir, "auth-observations.yaml")

	err := auth.RunObserve(auth.Options{
		Capture:   capture,
		APIPrefix: "https://api.example.com/v1",
		Format:    "har",
		Output:    out,
	})
	if err != nil {
		t.Fatal(err)
	}

	file, err := auth.LoadObservationsFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if file.SuccessRequestCount != 3 {
		t.Fatalf("success count = %d, want 3", file.SuccessRequestCount)
	}
	if len(file.Observed) < 2 {
		t.Fatalf("observed = %d, want at least cookie and bearer", len(file.Observed))
	}
	var cookieCov, bearerCov float64
	for _, o := range file.Observed {
		switch o.Kind {
		case "cookie":
			cookieCov = o.Coverage
			if o.SampleValue == "" || o.SampleValue == "[REDACTED]" {
				t.Fatalf("cookie sample should list names: %q", o.SampleValue)
			}
		case "bearer":
			bearerCov = o.Coverage
		}
	}
	if cookieCov < 0.99 {
		t.Fatalf("cookie coverage = %v", cookieCov)
	}
	if bearerCov < 0.3 {
		t.Fatalf("bearer coverage = %v", bearerCov)
	}
	foundTokenPath := false
	for _, s := range file.SuggestedAuthPaths {
		if s.Path == "/users/token" {
			foundTokenPath = true
		}
	}
	if !foundTokenPath {
		t.Fatalf("expected /users/token in suggestions: %v", file.SuggestedAuthPaths)
	}
}

func TestRunObserve_minimalNoAuth(t *testing.T) {
	root := repoRoot(t)
	capture := filepath.Join(root, "testdata", "captures", "minimal.har")
	out := filepath.Join(t.TempDir(), "auth-observations.yaml")

	err := auth.RunObserve(auth.Options{
		Capture:   capture,
		APIPrefix: "https://api.example.com/v1",
		Format:    "har",
		Output:    out,
	})
	if err != nil {
		t.Fatal(err)
	}
	file, err := auth.LoadObservationsFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(file.Observed) != 0 {
		t.Fatalf("expected no auth headers in minimal.har, got %v", file.Observed)
	}
}
