package main

import (
	"context"
	"fmt"
	"os"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/curate"
	"github.com/spf13/cobra"
)

func newCurateCmd() *cobra.Command {
	var (
		schemaPath       string
		capture          string
		apiPrefix        string
		format           string
		auto             bool
		llmSuggest       bool
		applySuggestions string
		suggestionsOut   string
		emitPrompts      string
		paramRegex       string
		baseURL          string
		apiKey           string
		model            string
		maxGroups        int
	)

	cmd := &cobra.Command{
		Use:   "curate",
		Short: "Assist Curation by clustering x-path-templates (Curation Assist)",
		Long: `Reduce x-path-templates noise before manual Curation.

  --auto                  Deterministic template clustering (numeric segments)
  --llm-suggest           LLM merge suggestions for slug-like paths (writes sidecar YAML)
  --apply-suggestions     Apply an edited suggestions file to the schema`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if schemaPath == "" {
				return fmt.Errorf("required flag: -o")
			}
			return curate.Run(context.Background(), curate.Options{
				Schema:           schemaPath,
				Auto:             auto,
				LLMSuggest:       llmSuggest,
				ApplySuggestions: applySuggestions,
				Capture:          capture,
				APIPrefix:        apiPrefix,
				Format:           format,
				SuggestionsOut:   suggestionsOut,
				ParamRegex:       paramRegex,
				EmitPromptsDir:   emitPrompts,
				BaseURL:          envOr(baseURL, os.Getenv("OPENAI_BASE_URL")),
				APIKey:           envOr(apiKey, os.Getenv("OPENAI_API_KEY")),
				Model:            envOr(model, os.Getenv("OPENAI_MODEL")),
				MaxGroups:        maxGroups,
			})
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&schemaPath, "output", "o", "", "OpenAPI schema YAML to update in place")
	flags.StringVarP(&capture, "input", "i", "", "HAR or flow capture (required for --llm-suggest)")
	flags.StringVarP(&apiPrefix, "api-prefix", "p", "", "API base URL prefix (required for --llm-suggest)")
	flags.StringVarP(&format, "format", "f", "", "Override capture format (har or flow)")
	flags.BoolVar(&auto, "auto", false, "Apply conservative template clustering")
	flags.BoolVar(&llmSuggest, "llm-suggest", false, "Write LLM template merge suggestions to a sidecar YAML")
	flags.StringVar(&applySuggestions, "apply-suggestions", "", "Apply edited template-suggestions.yaml to the schema")
	flags.StringVar(&suggestionsOut, "suggestions-out", "", "Output path for --llm-suggest (default: <schema-dir>/template-suggestions.yaml)")
	flags.StringVar(&emitPrompts, "emit-prompts", "", "Write LLM prompts without calling the API")
	flags.StringVarP(&paramRegex, "param-regex", "r", "[0-9]+", "Regex for numeric path segments (same as pass -r)")
	flags.StringVar(&baseURL, "base-url", "", "OpenAI-compatible API base URL")
	flags.StringVar(&apiKey, "api-key", "", "LLM API key")
	flags.StringVar(&model, "model", "", "LLM model name")
	flags.IntVar(&maxGroups, "max-groups", 50, "Max candidate groups to send to the LLM per run")

	return cmd
}
