package har_test

import (
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture/flow"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture/har"
)

func TestArchiveHeuristicWinsForHARCapture(t *testing.T) {
	harPath := testdataPath(t, "captures", "minimal.har")

	harScore := har.ArchiveHeuristic(harPath)
	flowScore := flow.DumpHeuristic(harPath)

	if harScore <= flowScore {
		t.Fatalf("expected har score (%d) > flow score (%d) for HAR capture", harScore, flowScore)
	}
	if harScore < 100 {
		t.Fatalf("expected high HAR score, got %d", harScore)
	}
}

func TestArchiveHeuristicSuffixBonus(t *testing.T) {
	if har.ArchiveHeuristic("capture.har") <= har.ArchiveHeuristic("capture.json") {
		t.Fatal("expected .har suffix to increase score")
	}
}
