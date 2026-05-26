package main

import (
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/auth"
	"github.com/spf13/cobra"
)

func newAuthCmd() *cobra.Command {
	var (
		capture  string
		apiPrefix string
		format   string
		output   string
	)

	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Observe and apply API authentication (Auth Observation)",
	}

	observeCmd := &cobra.Command{
		Use:   "observe",
		Short: "Scan a capture and write auth-observations.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return auth.RunObserve(auth.Options{
				Capture:   capture,
				APIPrefix: apiPrefix,
				Format:    format,
				Output:    output,
			})
		},
	}
	observeFlags := observeCmd.Flags()
	observeFlags.StringVarP(&capture, "input", "i", "", "HAR or flow capture file")
	observeFlags.StringVarP(&apiPrefix, "api-prefix", "p", "", "API base URL prefix")
	observeFlags.StringVarP(&format, "format", "f", "", "Override capture format (har or flow)")
	observeFlags.StringVarP(&output, "output", "o", "", "Output auth-observations.yaml path")

	var (
		schemaPath   string
		observations string
		force        bool
		noAuthTags   bool
		cookieName   string
	)
	applyCmd := &cobra.Command{
		Use:   "apply",
		Short: "Write securitySchemes and security from verified auth-observations.yaml",
		RunE: func(cmd *cobra.Command, args []string) error {
			return auth.RunApply(auth.ApplyOptions{
				Schema:           schemaPath,
				Observations:     observations,
				Force:            force,
				TagAuthEndpoints: !noAuthTags,
				CookieName:       cookieName,
			})
		},
	}
	applyFlags := applyCmd.Flags()
	applyFlags.StringVarP(&schemaPath, "schema", "s", "", "OpenAPI schema.yaml to update")
	applyFlags.StringVar(&observations, "observations", "", "auth-observations.yaml (default: same dir as schema)")
	applyFlags.BoolVar(&force, "force", false, "Overwrite existing securitySchemes entries")
	applyFlags.BoolVar(&noAuthTags, "no-tag-auth-paths", false, "Do not add tags: [auth] on suggested_auth_paths")
	applyFlags.StringVar(&cookieName, "cookie-name", "", "Primary cookie name for sessionCookie scheme (default: inferred from observed sample)")

	cmd.AddCommand(observeCmd, applyCmd)
	return cmd
}
