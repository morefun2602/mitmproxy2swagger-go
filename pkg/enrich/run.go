package enrich

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture"
	captureopen "github.com/morefun2602/mitmproxy2swagger-go/pkg/capture/open"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
	"github.com/openai/openai-go/v3"
)

const defaultConcurrency = 10

// EndpointProgressFunc is called once per Endpoint before enrichment work starts.
type EndpointProgressFunc func(current, total int, method, pathTemplate string)

// Options configures an Enrichment run.
type Options struct {
	Capture        string
	SchemaPath     string
	Output         string
	APIPrefix      string
	Format         string
	Samples        int
	Force          bool
	Redact         RedactMode
	EmitPromptsDir string
	Model          string
	BaseURL        string
	APIKey         string
	Concurrency    int
	Reader         capture.Reader
	Client         *openai.Client
	OnProgress     EndpointProgressFunc
}

type enrichJob struct {
	pathIdx      int
	method       string
	pathTemplate string
	userPrompt   string
}

func Run(ctx context.Context, opts Options) error {
	if opts.Samples < 1 {
		opts.Samples = 1
	}
	if opts.Redact == "" {
		opts.Redact = RedactStrict
	}

	doc, err := schema.Load(opts.SchemaPath)
	if err != nil {
		return fmt.Errorf("load schema: %w", err)
	}

	reader := opts.Reader
	if reader == nil {
		reader, err = captureopen.OpenReader(opts.Capture, opts.Format, nil)
		if err != nil {
			return err
		}
	}

	samples, err := endpointSamples(doc, reader, opts.APIPrefix, opts.Samples, opts.Redact)
	if err != nil {
		return err
	}

	var client openai.Client
	if opts.EmitPromptsDir == "" {
		var err error
		client, err = newEnrichmentClient(opts)
		if err != nil {
			return err
		}
	}

	jobs, skipped, promptCount, err := collectEnrichJobs(doc, samples, opts)
	if err != nil {
		return err
	}
	total := countEndpoints(doc)

	if opts.EmitPromptsDir != "" {
		if promptCount == 0 {
			return fmt.Errorf("no endpoints in schema paths (run Pass twice: first pass → curation → second pass)")
		}
		return nil
	}

	skipped += runEnrichmentJobs(ctx, opts, client, doc, jobs, total)

	if skipped > 0 {
		fmt.Fprintf(os.Stderr, "enrich: skipped %d/%d endpoint(s); saved partial result to %s\n", skipped, total, opts.Output)
	}

	return doc.Save(opts.Output)
}

func collectEnrichJobs(doc *schema.Document, samples map[endpointKey][]RequestSample, opts Options) (jobs []enrichJob, skipped int, promptCount int, err error) {
	total := countEndpoints(doc)
	progress := 0
	for i, item := range doc.Paths {
		pathTemplate := fmt.Sprint(item.Key)
		ops, ok := item.Value.(map[string]any)
		if !ok {
			continue
		}
		for method, raw := range ops {
			if !isHTTPMethod(method) {
				continue
			}
			op, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			key := endpointKey{pathTemplate: pathTemplate, method: strings.ToLower(method)}
			userPrompt, buildErr := buildUserPrompt(pathTemplate, method, op, samples[key])
			if buildErr != nil {
				fmt.Fprintf(os.Stderr, "warn: enrich %s %s: %v (skipped)\n", strings.ToUpper(method), pathTemplate, buildErr)
				skipped++
				continue
			}
			if opts.EmitPromptsDir != "" {
				if err := writePromptFile(opts.EmitPromptsDir, pathTemplate, method, userPrompt); err != nil {
					return nil, skipped, promptCount, err
				}
				promptCount++
				progress++
				if opts.OnProgress != nil {
					opts.OnProgress(progress, total, strings.ToUpper(method), pathTemplate)
				}
				continue
			}
			jobs = append(jobs, enrichJob{
				pathIdx:      i,
				method:       method,
				pathTemplate: pathTemplate,
				userPrompt:   userPrompt,
			})
		}
	}
	return jobs, skipped, promptCount, nil
}

func runEnrichmentJobs(ctx context.Context, opts Options, client openai.Client, doc *schema.Document, jobs []enrichJob, total int) int {
	if len(jobs) == 0 {
		return 0
	}

	conc := opts.Concurrency
	if conc <= 0 {
		conc = defaultConcurrency
	}
	if conc > len(jobs) {
		conc = len(jobs)
	}

	sem := make(chan struct{}, conc)
	var wg sync.WaitGroup
	var skipped atomic.Int32
	var progress atomic.Int32
	var mu sync.Mutex

	for _, job := range jobs {
		wg.Add(1)
		go func(job enrichJob) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			if ctx.Err() != nil {
				skipped.Add(1)
				return
			}

			cur := int(progress.Add(1))
			if opts.OnProgress != nil {
				opts.OnProgress(cur, total, strings.ToUpper(job.method), job.pathTemplate)
			}

			result, err := callEnrichmentLLM(ctx, client, opts.Model, systemPrompt, job.userPrompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "warn: enrich %s %s: %v (skipped)\n", strings.ToUpper(job.method), job.pathTemplate, err)
				skipped.Add(1)
				return
			}

			mu.Lock()
			ops, ok := doc.Paths[job.pathIdx].Value.(map[string]any)
			if !ok {
				mu.Unlock()
				skipped.Add(1)
				return
			}
			op, ok := ops[job.method].(map[string]any)
			if !ok {
				mu.Unlock()
				skipped.Add(1)
				return
			}
			applyEnrichment(op, result, opts.Force)
			mu.Unlock()
		}(job)
	}

	wg.Wait()
	return int(skipped.Load())
}

func countEndpoints(doc *schema.Document) int {
	n := 0
	for _, item := range doc.Paths {
		ops, ok := item.Value.(map[string]any)
		if !ok {
			continue
		}
		for method, raw := range ops {
			if !isHTTPMethod(method) {
				continue
			}
			if _, ok := raw.(map[string]any); ok {
				n++
			}
		}
	}
	return n
}

func isHTTPMethod(method string) bool {
	switch strings.ToLower(method) {
	case "get", "post", "put", "patch", "delete", "head", "options", "trace":
		return true
	default:
		return false
	}
}

func writePromptFile(dir, pathTemplate, method, prompt string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	name := sanitizeFilename(pathTemplate) + "__" + strings.ToLower(method) + ".txt"
	return os.WriteFile(filepath.Join(dir, name), []byte(prompt), 0o644)
}

func sanitizeFilename(s string) string {
	s = strings.TrimPrefix(s, "/")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "{", "")
	s = strings.ReplaceAll(s, "}", "")
	if s == "" {
		return "root"
	}
	return s
}
