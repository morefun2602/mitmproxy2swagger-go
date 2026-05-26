package main

import (
	"context"
	"fmt"
	"os"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/enrich"
	"github.com/spf13/cobra"
)

func newEnrichCmd() *cobra.Command {
	var (
		capture     string
		schemaPath  string
		output      string
		apiPrefix   string
		format      string
		samples     int
		force       bool
		redact      string
		emitPrompts string
		baseURL     string
		apiKey      string
		model       string
	)

	cmd := &cobra.Command{
		Use:   "enrich",
		Short: "Enrich an OpenAPI schema with LLM-generated semantic fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			if capture == "" || schemaPath == "" || output == "" || apiPrefix == "" {
				return fmt.Errorf("required flags: -i, -s, -o, -p")
			}

			opts := enrich.Options{
				Capture:        capture,
				SchemaPath:     schemaPath,
				Output:         output,
				APIPrefix:      apiPrefix,
				Format:         format,
				Samples:        samples,
				Force:          force,
				Redact:         enrich.RedactMode(redact),
				EmitPromptsDir: emitPrompts,
				BaseURL:        envOr(baseURL, os.Getenv("OPENAI_BASE_URL")),
				APIKey:         envOr(apiKey, os.Getenv("OPENAI_API_KEY")),
				Model:          envOr(model, os.Getenv("OPENAI_MODEL")),
				OnProgress: func(cur, total int, method, path string) {
					fmt.Fprintf(cmd.OutOrStdout(), "[%d/%d] %s %s\n", cur, total, method, path)
				},
			}

			if err := enrich.Run(context.Background(), opts); err != nil {
				return err
			}
			if opts.EmitPromptsDir != "" {
				fmt.Fprintln(cmd.OutOrStdout(), "Prompts written.")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Done!")
			return nil
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&capture, "input", "i", "", "HAR or flow capture file")
	flags.StringVarP(&schemaPath, "schema", "s", "", "Input OpenAPI schema YAML from Pass")
	flags.StringVarP(&output, "output", "o", "", "Output enriched schema YAML")
	flags.StringVarP(&apiPrefix, "api-prefix", "p", "", "The api prefix")
	flags.StringVarP(&format, "format", "f", "", "Override capture format (har or flow)")
	flags.IntVar(&samples, "samples", 1, "Max Captured Request samples per Endpoint")
	flags.BoolVar(&force, "force", false, "Overwrite existing semantic fields")
	flags.StringVar(&redact, "redact", "strict", "Redaction mode before sending to LLM: strict or off")
	flags.StringVar(&emitPrompts, "emit-prompts", "", "Write LLM user prompts to this directory without calling the LLM")
	flags.StringVar(&baseURL, "base-url", "", "OpenAI-compatible API base URL (default $OPENAI_BASE_URL or https://api.openai.com/v1)")
	flags.StringVar(&apiKey, "api-key", "", "LLM API key (default $OPENAI_API_KEY)")
	flags.StringVar(&model, "model", "", "LLM model name (default $OPENAI_MODEL or gpt-4o-mini)")

	return cmd
}
