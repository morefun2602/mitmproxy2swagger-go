package main

import (
	"fmt"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/tags"
	"github.com/spf13/cobra"
)

func newTagsCmd() *cobra.Command {
	var (
		schemaPath string
		tagsFile   string
		output     string
		merge      bool
		strict     bool
	)

	cmd := &cobra.Command{
		Use:   "tags",
		Short: "Apply tag grouping from a tags.yaml sidecar",
	}

	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Write operation tags and optional x-tagGroups from tags.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			if schemaPath == "" {
				return fmt.Errorf("required flag: -s / --schema")
			}
			res, err := tags.RunApply(tags.ApplyOptions{
				Schema:   schemaPath,
				TagsFile: tagsFile,
				Output:   output,
				Merge:    merge,
				Strict:   strict,
			})
			if err != nil {
				return err
			}
			out := output
			if out == "" {
				out = schemaPath
			}
			tags.PrintApplySummary(res, out)
			return nil
		},
	}

	flags := applyCmd.Flags()
	flags.StringVarP(&schemaPath, "schema", "s", "", "OpenAPI schema YAML to update (e.g. enriched.yaml)")
	flags.StringVarP(&tagsFile, "tags-file", "t", "", "tags.yaml sidecar (default: same dir as schema)")
	flags.StringVarP(&output, "output", "o", "", "Output path (default: overwrite schema)")
	flags.BoolVar(&merge, "merge", false, "Merge sidecar tag with existing tags (default: replace)")
	flags.BoolVar(&strict, "strict", false, "Fail if any operation has no matching rule")

	cmd.AddCommand(applyCmd)
	return cmd
}
