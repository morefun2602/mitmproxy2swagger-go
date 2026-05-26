package curate

import (
	"context"
	"fmt"
	"regexp"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
	"github.com/openai/openai-go/v3"
)

// Options configures a curate run.
type Options struct {
	Schema           string
	Auto             bool
	LLMSuggest       bool
	ApplySuggestions string
	Capture          string
	APIPrefix        string
	Format           string
	SuggestionsOut   string
	ParamRegex       string
	EmitPromptsDir   string
	Model            string
	BaseURL          string
	APIKey           string
	MaxGroups        int
	Client           *openai.Client
}

// Run updates the schema and/or writes suggestion sidecars.
func Run(ctx context.Context, opts Options) error {
	if opts.Schema == "" {
		return fmt.Errorf("required flag: -o (schema path)")
	}

	modes := 0
	if opts.Auto {
		modes++
	}
	if opts.LLMSuggest {
		modes++
	}
	if opts.ApplySuggestions != "" {
		modes++
	}
	if modes != 1 {
		return fmt.Errorf("exactly one of --auto, --llm-suggest, or --apply-suggestions is required")
	}

	if opts.Auto {
		return runAuto(opts)
	}
	if opts.LLMSuggest {
		return runLLMSuggest(ctx, opts)
	}
	return runApplySuggestions(opts)
}

func runAuto(opts Options) error {
	paramRegex := opts.ParamRegex
	if paramRegex == "" {
		paramRegex = `[0-9]+`
	}
	numericRe, err := regexp.Compile("^" + paramRegex + "$")
	if err != nil {
		return fmt.Errorf("invalid path parameter regex: %w", err)
	}

	doc, err := schema.Load(opts.Schema)
	if err != nil {
		return err
	}

	before := len(doc.XPathTemplates)
	doc.XPathTemplates = AutoTemplates(doc.XPathTemplates, numericRe)
	after := len(doc.XPathTemplates)

	if err := doc.Save(opts.Schema); err != nil {
		return err
	}

	fmt.Printf("curate --auto: %d -> %d x-path-templates\n", before, after)
	return nil
}
