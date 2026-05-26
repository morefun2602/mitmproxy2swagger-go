package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/golden"
)

func main() {
	repoRoot := flag.String("repo", ".", "repository root (default: current directory)")
	caseFilter := flag.String("case", "", "comma-separated case id(s); default: all cases in testdata/cases.yaml")
	verify := flag.Bool("verify", false, "compare generated output with committed golden files instead of overwriting them")
	flag.Parse()

	root, err := filepath.Abs(*repoRoot)
	if err != nil {
		exitErr(err)
	}

	casesPath := filepath.Join(root, "testdata", "cases.yaml")
	goldenDir := filepath.Join(root, "testdata", "golden")

	cases, err := golden.LoadCases(casesPath)
	if err != nil {
		exitErr(err)
	}

	var ids []string
	if *caseFilter != "" {
		for _, id := range strings.Split(*caseFilter, ",") {
			if id = strings.TrimSpace(id); id != "" {
				ids = append(ids, id)
			}
		}
	}
	cases, err = golden.FilterCases(cases, ids)
	if err != nil {
		exitErr(err)
	}

	fmt.Printf("Golden dir: %s\n", goldenDir)
	fmt.Printf("Cases:      %d\n", len(cases))
	if *verify {
		fmt.Println("Mode:       verify")
	} else {
		fmt.Println("Mode:       generate")
	}
	fmt.Println()

	for _, c := range cases {
		if *verify && caseUsesFlow(c) {
			fmt.Printf("[%s] SKIP (flow dump reader not implemented)\n", c.ID)
			continue
		}
		if *verify {
			if err := golden.VerifyOne(root, goldenDir, c); err != nil {
				exitErr(err)
			}
			continue
		}
		if err := golden.GenerateOne(root, goldenDir, c); err != nil {
			exitErr(err)
		}
	}

	fmt.Println("\nDone.")
}

func caseUsesFlow(c golden.Case) bool {
	for i := 0; i < len(c.ExtraArgs); i++ {
		switch c.ExtraArgs[i] {
		case "--format", "-f":
			if i+1 < len(c.ExtraArgs) && c.ExtraArgs[i+1] == "flow" {
				return true
			}
		}
	}
	return false
}

func exitErr(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}
