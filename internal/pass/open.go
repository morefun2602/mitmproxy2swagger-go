package pass

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture/har"
)

func osIsNotExist(err error) bool {
	return errors.Is(err, os.ErrNotExist)
}

// OpenReader selects a Capture Reader from format flags and heuristics.
func OpenReader(inputPath, format string, progress capture.ProgressFunc) (capture.Reader, error) {
	switch format {
	case "har":
		return har.NewCaptureReader(inputPath, progress), nil
	case "flow", "mitmproxy":
		return nil, fmt.Errorf("flow dump Capture Reader is not implemented yet")
	case "":
		harScore := har.ArchiveHeuristic(inputPath)
		flowScore := flowDumpHeuristic(inputPath)
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

func flowDumpHeuristic(path string) int {
	val := 0
	if strings.Contains(path, "flow") {
		val++
	}
	if strings.Contains(path, "mitmproxy") {
		val++
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return val
	}
	if len(data) > 2048 {
		data = data[:2048]
	}
	if len(data) == 0 {
		return val
	}
	compact := strings.NewReplacer("\r", "", "\n", "").Replace(string(data))
	if !isPrintable(compact) {
		val += 50
	}
	if data[0] >= '0' && data[0] <= '9' {
		val += 5
	}
	if strings.Contains(string(data), "status_code") {
		val += 5
	}
	if strings.Contains(string(data), "regular") {
		val += 10
	}
	return val
}

func isPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) {
			return false
		}
	}
	return true
}
