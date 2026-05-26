package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/golden"
)

func main() {
	schemaPath := flag.String("schema", "", "OpenAPI schema YAML to strip ignore: prefixes from x-path-templates")
	flag.Parse()
	if *schemaPath == "" {
		fmt.Fprintln(os.Stderr, "usage: stripignore -schema schema.yaml")
		os.Exit(1)
	}
	if err := golden.StripIgnorePrefixes(*schemaPath); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
