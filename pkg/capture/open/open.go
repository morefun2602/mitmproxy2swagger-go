package open

import (
	"fmt"
	"os"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture/flow"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture/har"
)

// OpenReader selects a Capture Reader from format flags and heuristics.
func OpenReader(inputPath, format string, progress capture.ProgressFunc) (capture.Reader, error) {
	switch format {
	case "har":
		return har.NewCaptureReader(inputPath, progress), nil
	case "flow", "mitmproxy":
		return nil, fmt.Errorf("flow dump Capture Reader is not implemented yet")
	case "":
		harScore := har.ArchiveHeuristic(inputPath)
		flowScore := flow.DumpHeuristic(inputPath)
		if os.Getenv("MITMPROXY2SWAGGER_DEBUG") != "" {
			fmt.Printf("har score: %d\n", harScore)
			fmt.Printf("flow score: %d\n", flowScore)
		}
		if harScore > flowScore {
			return har.NewCaptureReader(inputPath, progress), nil
		}
		return nil, fmt.Errorf("flow dump Capture Reader is not implemented yet")
	default:
		return nil, fmt.Errorf("unknown format %q", format)
	}
}
