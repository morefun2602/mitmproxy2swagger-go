package main

import (
	"fmt"
	"os"
)

func main() {
	root := NewRoot()
	if err := root.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
