package curate

import (
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture"
)

func collectSamplePaths(reader capture.Reader, apiPrefix string, want map[string]struct{}, limit int) ([]string, error) {
	if limit < 1 {
		limit = 3
	}
	seen := make(map[string]struct{})
	var out []string
	err := reader.Each(func(req capture.CapturedRequest) error {
		matchURL, ok := req.MatchingURL(apiPrefix)
		if !ok {
			return nil
		}
		path := stripQueryString(matchURL)
		path = strings.TrimPrefix(path, apiPrefix)
		if path == "" {
			path = "/"
		}
		if _, need := want[path]; !need {
			return nil
		}
		if _, dup := seen[path]; dup {
			return nil
		}
		seen[path] = struct{}{}
		out = append(out, path)
		if len(out) >= limit {
			return errStopIteration
		}
		return nil
	})
	if err == errStopIteration {
		return out, nil
	}
	return out, err
}

type stopError struct{}

func (stopError) Error() string { return "stop" }

var errStopIteration = stopError{}

func stripQueryString(rawURL string) string {
	if i := strings.Index(rawURL, "?"); i >= 0 {
		return rawURL[:i]
	}
	return rawURL
}
