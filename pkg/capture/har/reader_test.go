package har_test

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture/har"
)

func testdataPath(t *testing.T, parts ...string) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Join(filepath.Dir(file), "..", "..", "..")
	return filepath.Join(append([]string{root, "testdata"}, parts...)...)
}

func TestArchiveHeuristic(t *testing.T) {
	harPath := testdataPath(t, "captures", "minimal.har")
	flowPath := testdataPath(t, "captures", "test_flows")

	harScore := har.ArchiveHeuristic(harPath)
	flowScore := har.ArchiveHeuristic(flowPath)

	if harScore <= flowScore {
		t.Fatalf("expected HAR score (%d) > flow score (%d)", harScore, flowScore)
	}
	if harScore < 100 {
		t.Fatalf("expected high HAR score, got %d", harScore)
	}
}

func TestCaptureReaderEach(t *testing.T) {
	path := testdataPath(t, "captures", "minimal.har")
	reader := har.NewCaptureReader(path, nil)

	var count int
	var gotPrefixMatch bool
	err := reader.Each(func(req capture.CapturedRequest) error {
		count++
		switch {
		case req.Method() == "GET" && req.URL() == "https://api.example.com/v1/items":
			if body := string(req.ResponseBody()); body != `{"items":[]}` {
				t.Fatalf("unexpected GET response body: %q", body)
			}
		case req.Method() == "POST":
			if string(req.RequestBody()) != `{"name":"widget"}` {
				t.Fatalf("unexpected POST body: %q", req.RequestBody())
			}
			if string(req.ResponseBody()) != `{"id":1}` {
				t.Fatalf("unexpected base64 response body: %q", req.ResponseBody())
			}
			if req.ResponseStatusCode() != 201 {
				t.Fatalf("unexpected status: %d", req.ResponseStatusCode())
			}
		}
		if _, ok := req.MatchingURL("https://api.example.com/v1"); ok {
			gotPrefixMatch = true
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("expected 3 entries, got %d", count)
	}
	if !gotPrefixMatch {
		t.Fatal("expected at least one Prefix Match for api.example.com/v1")
	}
}

func TestMatchingURLStrict(t *testing.T) {
	path := testdataPath(t, "captures", "minimal.har")
	reader := har.NewCaptureReader(path, nil)

	var otherURL string
	_ = reader.Each(func(req capture.CapturedRequest) error {
		if req.URL() == "https://other.example.com/nope" {
			if _, ok := req.MatchingURL("https://api.example.com/v1"); ok {
				t.Fatal("HAR Prefix Match must not use Host fallback")
			}
			otherURL = req.URL()
		}
		return nil
	})
	if otherURL == "" {
		t.Fatal("expected other.example.com entry")
	}
}

func TestCaptureReaderName(t *testing.T) {
	reader := har.NewCaptureReader("x.har", nil)
	if reader.Name() != "har" {
		t.Fatalf("Name() = %q, want har", reader.Name())
	}
}
