package open_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	captureopen "github.com/morefun2602/mitmproxy2swagger-go/internal/capture/open"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func capturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(repoRoot(t), "testdata", "captures", name)
}

func TestOpenReaderExplicitHAR(t *testing.T) {
	reader, err := captureopen.OpenReader(capturePath(t, "minimal.har"), "har", nil)
	if err != nil {
		t.Fatal(err)
	}
	if reader.Name() != "har" {
		t.Fatalf("Name() = %q, want har", reader.Name())
	}
}

func TestOpenReaderExplicitFlowNotImplemented(t *testing.T) {
	_, err := captureopen.OpenReader(capturePath(t, "test_flows"), "flow", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "flow dump Capture Reader is not implemented yet") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenReaderExplicitMitmproxyAlias(t *testing.T) {
	_, err := captureopen.OpenReader(capturePath(t, "test_flows"), "mitmproxy", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOpenReaderUnknownFormat(t *testing.T) {
	_, err := captureopen.OpenReader(capturePath(t, "minimal.har"), "xml", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), `unknown format "xml"`) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestOpenReaderAutoDetectHAR(t *testing.T) {
	reader, err := captureopen.OpenReader(capturePath(t, "minimal.har"), "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if reader.Name() != "har" {
		t.Fatalf("Name() = %q, want har", reader.Name())
	}
}

func TestOpenReaderAutoDetectFlowNotImplemented(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "capture.flow")
	content := []byte("9:\x00\x01status_code\x00regular")
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := captureopen.OpenReader(path, "", nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "flow dump Capture Reader is not implemented yet") {
		t.Fatalf("unexpected error: %v", err)
	}
}
