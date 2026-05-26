package main

import (
	"flag"
	"fmt"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/pass"
	"github.com/spf13/cobra"
)

func newPassCmd() *cobra.Command {
	gofs := flag.NewFlagSet("pass", flag.ContinueOnError)
	input := gofs.String("i", "", "The input mitmproxy dump file or HAR dump file (from DevTools)")
	gofs.StringVar(input, "input", "", "The input mitmproxy dump file or HAR dump file (from DevTools)")
	output := gofs.String("o", "", "The output swagger schema file (yaml). If it exists, new endpoints will be added")
	gofs.StringVar(output, "output", "", "The output swagger schema file (yaml). If it exists, new endpoints will be added")
	apiPrefix := gofs.String("p", "", "The api prefix")
	gofs.StringVar(apiPrefix, "api-prefix", "", "The api prefix")
	examples := gofs.Bool("e", false, "Include examples in the schema. This might expose sensitive information.")
	gofs.BoolVar(examples, "examples", false, "Include examples in the schema. This might expose sensitive information.")
	headers := gofs.Bool("hd", false, "Include headers in the schema. This might expose sensitive information.")
	gofs.BoolVar(headers, "headers", false, "Include headers in the schema. This might expose sensitive information.")
	format := gofs.String("f", "", "Override the input file format auto-detection.")
	gofs.StringVar(format, "format", "", "Override the input file format auto-detection.")
	paramRegex := gofs.String("r", "[0-9]+", "Regex to match parameters in the API paths. Path segments that match this regex will be turned into parameter placeholders.")
	gofs.StringVar(paramRegex, "param-regex", "[0-9]+", "Regex to match parameters in the API paths. Path segments that match this regex will be turned into parameter placeholders.")
	suppressParams := gofs.Bool("s", false, "Do not include API paths that have the original parameter values, only the ones with placeholders.")
	gofs.BoolVar(suppressParams, "suppress-params", false, "Do not include API paths that have the original parameter values, only the ones with placeholders.")

	cmd := &cobra.Command{
		Use:   "pass",
		Short: "Run a Pass over a capture file and update the OpenAPI schema",
		RunE: func(cmd *cobra.Command, args []string) error {
			if *input == "" || *output == "" || *apiPrefix == "" {
				return fmt.Errorf("required flags: -i, -o, -p")
			}
			err := pass.Run(pass.Options{
				Input:          *input,
				Output:         *output,
				APIPrefix:      *apiPrefix,
				Examples:       *examples,
				Headers:        *headers,
				Format:         *format,
				ParamRegex:     *paramRegex,
				SuppressParams: *suppressParams,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "Done!")
			return nil
		},
	}

	cmd.Flags().AddGoFlagSet(gofs)

	return cmd
}
