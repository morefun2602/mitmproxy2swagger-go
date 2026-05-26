package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootRequiresSubcommand(t *testing.T) {
	root := NewRoot()
	var out, errOut bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errOut)
	root.SetArgs(nil)

	_, err := root.ExecuteC()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "required subcommand") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "pass") {
		t.Fatalf("expected pass hint in error: %v", err)
	}
}

func TestPassRequiresFlags(t *testing.T) {
	root := NewRoot()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"pass"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "required flags") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnrichRequiresFlags(t *testing.T) {
	root := NewRoot()
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"enrich"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "required flags") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVersion(t *testing.T) {
	root := NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"version"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	if !strings.Contains(got, "mitmproxy2swagger dev") {
		t.Fatalf("unexpected version output: %q", got)
	}
}

func TestCompletionBash(t *testing.T) {
	root := NewRoot()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&bytes.Buffer{})
	root.SetArgs([]string{"completion", "bash"})

	if err := root.Execute(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "mitmproxy2swagger") {
		t.Fatal("expected bash completion script")
	}
}
