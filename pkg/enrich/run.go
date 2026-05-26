package enrich

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/capture"
	captureopen "github.com/morefun2602/mitmproxy2swagger-go/pkg/capture/open"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
	"github.com/openai/openai-go/v3"
)

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
	Reader         capture.Reader
	Client         *openai.Client
	OnProgress     EndpointProgressFunc
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

	total := countEndpoints(doc)
	current := 0
	promptCount := 0
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
			current++
			if opts.OnProgress != nil {
				opts.OnProgress(current, total, strings.ToUpper(method), pathTemplate)
			}
			key := endpointKey{pathTemplate: pathTemplate, method: strings.ToLower(method)}
			userPrompt, err := buildUserPrompt(pathTemplate, method, op, samples[key])
			if err != nil {
				return err
			}
			if opts.EmitPromptsDir != "" {
				if err := writePromptFile(opts.EmitPromptsDir, pathTemplate, method, userPrompt); err != nil {
					return err
				}
				promptCount++
				continue
			}
			result, err := callEnrichmentLLM(ctx, client, opts.Model, systemPrompt, userPrompt)
			if err != nil {
				return fmt.Errorf("enrich %s %s: %w", strings.ToUpper(method), pathTemplate, err)
			}
			applyEnrichment(op, result, opts.Force)
			ops[method] = op
		}
		doc.Paths[i].Value = ops
	}

	if opts.EmitPromptsDir != "" {
		if promptCount == 0 {
			return fmt.Errorf("no endpoints in schema paths (run Pass twice: first pass → curation → second pass)")
		}
		return nil
	}

	return doc.Save(opts.Output)
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
