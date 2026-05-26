package flow_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture/flow"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture/har"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func capturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "testdata", "captures", name)
}

func TestDumpHeuristicFlowCapture(t *testing.T) {
	flowPath := capturePath(t, "msgpack_flows")
	harPath := capturePath(t, "minimal.har")

	flowScore := flow.DumpHeuristic(flowPath)
	if flowScore <= flow.DumpHeuristic(harPath) {
		t.Fatalf("expected msgpack flow score (%d) > minimal.har flow score (%d)", flowScore, flow.DumpHeuristic(harPath))
	}

	harScore := har.ArchiveHeuristic(harPath)
	if harScore <= flowScore {
		t.Fatalf("expected har score (%d) > flow score (%d) on HAR capture", harScore, flowScore)
	}
}

func TestDumpHeuristicBeatsHARHeuristic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "capture.flow")
	// Non-printable flow dump-like prefix; path and content cues boost flow score above HAR.
	content := []byte("9:\x00\x01status_code\x00regular")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	flowScore := flow.DumpHeuristic(path)
	harScore := har.ArchiveHeuristic(path)
	if flowScore <= harScore {
		t.Fatalf("expected flow score (%d) > har score (%d)", flowScore, harScore)
	}
}

func TestDumpHeuristicPathHints(t *testing.T) {
	if flow.DumpHeuristic("/tmp/my_flows") < flow.DumpHeuristic("/tmp/data") {
		t.Fatal("expected path containing 'flow' to score higher")
	}
}
