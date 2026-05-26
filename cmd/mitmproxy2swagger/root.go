package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewRoot() *cobra.Command {
	root := &cobra.Command{
		Use:   "mitmproxy2swagger",
		Short: "Reverse-engineer OpenAPI schemas from HTTP capture files",
		Long:  "Generate and enrich OpenAPI 3.0 YAML from mitmproxy flow dumps or HAR archives.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf(`required subcommand: pass, curate, auth, enrich, version, or completion

Try:
  mitmproxy2swagger pass -i capture.har -o schema.yaml -p https://api.example.com/v1
  mitmproxy2swagger curate --auto -o schema.yaml
  mitmproxy2swagger auth observe -i capture.har -p https://api.example.com/v1 -o auth-observations.yaml
  mitmproxy2swagger enrich -i capture.har -s schema.yaml -o enriched.yaml -p https://api.example.com/v1`)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.AddCommand(
		newPassCmd(),
		newCurateCmd(),
		newAuthCmd(),
		newEnrichCmd(),
		newVersionCmd(),
		newCompletionCmd(),
	)

	return root
}
