package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/pass"
)

func main() {
	input := flag.String("i", "", "The input mitmproxy dump file or HAR dump file (from DevTools)")
	flag.StringVar(input, "input", "", "The input mitmproxy dump file or HAR dump file (from DevTools)")
	output := flag.String("o", "", "The output swagger schema file (yaml). If it exists, new endpoints will be added")
	flag.StringVar(output, "output", "", "The output swagger schema file (yaml). If it exists, new endpoints will be added")
	apiPrefix := flag.String("p", "", "The api prefix")
	flag.StringVar(apiPrefix, "api-prefix", "", "The api prefix")
	examples := flag.Bool("e", false, "Include examples in the schema. This might expose sensitive information.")
	flag.BoolVar(examples, "examples", false, "Include examples in the schema. This might expose sensitive information.")
	headers := flag.Bool("hd", false, "Include headers in the schema. This might expose sensitive information.")
	flag.BoolVar(headers, "headers", false, "Include headers in the schema. This might expose sensitive information.")
	format := flag.String("f", "", "Override the input file format auto-detection.")
	flag.StringVar(format, "format", "", "Override the input file format auto-detection.")
	paramRegex := flag.String("r", "[0-9]+", "Regex to match parameters in the API paths. Path segments that match this regex will be turned into parameter placeholders.")
	flag.StringVar(paramRegex, "param-regex", "[0-9]+", "Regex to match parameters in the API paths. Path segments that match this regex will be turned into parameter placeholders.")
	suppressParams := flag.Bool("s", false, "Do not include API paths that have the original parameter values, only the ones with placeholders.")
	flag.BoolVar(suppressParams, "suppress-params", false, "Do not include API paths that have the original parameter values, only the ones with placeholders.")
	flag.Parse()

	if *input == "" || *output == "" || *apiPrefix == "" {
		fmt.Fprintln(os.Stderr, "required flags: -i, -o, -p")
		flag.Usage()
		os.Exit(1)
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
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Done!")
}
