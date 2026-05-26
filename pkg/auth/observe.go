package auth

import (
	"fmt"
	"sort"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture"
	captureopen "github.com/morefun2602/mitmproxy2swagger-go/pkg/capture/open"
)

// Options configures auth observe.
type Options struct {
	Capture          string
	APIPrefix        string
	Format           string
	Output           string
	AuthPathKeywords []string
	Reader           capture.Reader
}

type credStats struct {
	key         credentialKey
	sampleValue string
	count       int
	paths       map[string]struct{}
}

// RunObserve scans a capture and writes auth-observations.yaml.
func RunObserve(opts Options) error {
	if opts.Capture == "" || opts.APIPrefix == "" {
		return fmt.Errorf("required flags for auth observe: -i, -p")
	}
	if opts.Output == "" {
		return fmt.Errorf("required flag: -o")
	}

	reader := opts.Reader
	var err error
	if reader == nil {
		reader, err = captureopen.OpenReader(opts.Capture, opts.Format, nil)
		if err != nil {
			return err
		}
	}

	apiPrefix := strings.TrimSuffix(opts.APIPrefix, "/")
	stats := make(map[credentialKey]*credStats)
	allPaths := make(map[string]struct{})
	successCount := 0

	err = reader.Each(func(req capture.CapturedRequest) error {
		matchURL, ok := req.MatchingURL(apiPrefix)
		if !ok {
			return nil
		}
		status := req.ResponseStatusCode()
		if status < 200 || status >= 300 {
			return nil
		}
		successCount++

		path := stripQuery(matchURL)
		path = strings.TrimPrefix(path, apiPrefix)
		if path == "" {
			path = "/"
		}
		allPaths[path] = struct{}{}

		for name, values := range req.RequestHeaders() {
			key, sample, interesting := classifyAuthHeader(name, values)
			if !interesting {
				continue
			}
			st, exists := stats[key]
			if !exists {
				st = &credStats{
					key:         key,
					sampleValue: sample,
					paths:       make(map[string]struct{}),
				}
				stats[key] = st
			}
			st.count++
			if key.kind == kindCookie && len(sample) > len(st.sampleValue) {
				st.sampleValue = sample
			}
			if len(st.paths) < 5 {
				st.paths[path] = struct{}{}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	observed := buildObserved(stats, successCount)
	suggested := suggestAuthPaths(allPaths, opts.AuthPathKeywords)

	file := &ObservationsFile{
		APIPrefix:           apiPrefix,
		Capture:             opts.Capture,
		SuccessRequestCount: successCount,
		Observed:            observed,
		Verified:            []string{},
		Combination:         "",
		SuggestedAuthPaths:  suggested,
	}

	if err := saveObservationsFile(opts.Output, file); err != nil {
		return err
	}
	fmt.Printf("auth observe: %d success requests, %d observed credentials, %d suggested auth paths -> %s\n",
		successCount, len(observed), len(suggested), opts.Output)
	return nil
}

func buildObserved(stats map[credentialKey]*credStats, total int) []ObservedCredential {
	if total == 0 {
		return nil
	}
	keys := make([]credentialKey, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].kind != keys[j].kind {
			return keys[i].kind < keys[j].kind
		}
		return keys[i].name < keys[j].name
	})

	out := make([]ObservedCredential, 0, len(keys))
	for _, k := range keys {
		st := stats[k]
		paths := make([]string, 0, len(st.paths))
		for p := range st.paths {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		cov := float64(st.count) / float64(total)
		out = append(out, ObservedCredential{
			Kind:         k.kind,
			Name:         k.name,
			Coverage:     roundCoverage(cov),
			SamplePaths:  paths,
			SampleValue:  st.sampleValue,
			RequestCount: st.count,
		})
	}
	return out
}

func roundCoverage(c float64) float64 {
	return float64(int(c*1000+0.5)) / 1000
}

func stripQuery(rawURL string) string {
	if i := strings.Index(rawURL, "?"); i >= 0 {
		return rawURL[:i]
	}
	return rawURL
}
