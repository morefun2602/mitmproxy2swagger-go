package enrich_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/enrich"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/golden"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/pass"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
}

func mockOpenAIServer(t *testing.T, result enrich.EnrichmentResult) *httptest.Server {
	t.Helper()
	content, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		resp := map[string]any{
			"id":      "chatcmpl-test",
			"object":  "chat.completion",
			"created": 0,
			"model":   "gpt-4o-mini",
			"choices": []map[string]any{{
				"index": 0,
				"message": map[string]any{
					"role":    "assistant",
					"content": string(content),
				},
				"finish_reason": "stop",
			}},
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(resp); err != nil {
			t.Fatal(err)
		}
	}))
}

func TestEnrichVerticalSlice(t *testing.T) {
	root := repoRoot(t)
	capture := filepath.Join(root, "testdata", "captures", "minimal.har")
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	outPath := filepath.Join(dir, "enriched.yaml")

	if err := pass.Run(pass.Options{
		Input:     capture,
		Output:    schemaPath,
		APIPrefix: "https://api.example.com/v1",
		Format:    "har",
	}); err != nil {
		t.Fatalf("pass: %v", err)
	}
	if err := golden.StripIgnorePrefixes(schemaPath); err != nil {
		t.Fatalf("curation: %v", err)
	}
	if err := pass.Run(pass.Options{
		Input:     capture,
		Output:    schemaPath,
		APIPrefix: "https://api.example.com/v1",
		Format:    "har",
	}); err != nil {
		t.Fatalf("second pass: %v", err)
	}

	srv := mockOpenAIServer(t, enrich.EnrichmentResult{
		Summary:     "List items",
		Description: "Returns a collection of items.",
		Tags:        []string{"items"},
		OperationID: "listItems",
	})
	t.Cleanup(srv.Close)

	client := openai.NewClient(
		option.WithBaseURL(srv.URL+"/v1"),
		option.WithAPIKey("test-key"),
	)

	var progressCalls, lastTotal int
	err := enrich.Run(context.Background(), enrich.Options{
		Capture:    capture,
		SchemaPath: schemaPath,
		Output:     outPath,
		APIPrefix:  "https://api.example.com/v1",
		Format:     "har",
		Samples:    1,
		Force:      true,
		Client:     &client,
		OnProgress: func(cur, total int, method, path string) {
			progressCalls++
			lastTotal = total
			if cur < 1 || cur > total || total < 1 || method == "" || path == "" {
				t.Errorf("OnProgress(%d, %d, %q, %q)", cur, total, method, path)
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if progressCalls != lastTotal || progressCalls < 1 {
		t.Fatalf("OnProgress calls = %d, total = %d", progressCalls, lastTotal)
	}

	doc, err := schema.Load(outPath)
	if err != nil {
		t.Fatal(err)
	}
	ops, ok := doc.PathOperations("/items")
	if !ok {
		t.Fatal("missing /items")
	}
	get, ok := ops["get"].(map[string]any)
	if !ok {
		t.Fatal("missing GET /items")
	}
	if get["summary"] != "List items" {
		t.Fatalf("summary = %#v", get["summary"])
	}
	if get["description"] == "" {
		t.Fatal("expected description")
	}
	if get["operationId"] != "listItems" {
		t.Fatalf("operationId = %#v", get["operationId"])
	}
}

func TestEnrichEmitPrompts(t *testing.T) {
	root := repoRoot(t)
	capture := filepath.Join(root, "testdata", "captures", "minimal.har")
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	outPath := filepath.Join(dir, "out.yaml")
	promptsDir := filepath.Join(dir, "prompts")

	if err := pass.Run(pass.Options{
		Input:     capture,
		Output:    schemaPath,
		APIPrefix: "https://api.example.com/v1",
		Format:    "har",
	}); err != nil {
		t.Fatal(err)
	}
	if err := golden.StripIgnorePrefixes(schemaPath); err != nil {
		t.Fatal(err)
	}
	if err := pass.Run(pass.Options{
		Input:     capture,
		Output:    schemaPath,
		APIPrefix: "https://api.example.com/v1",
		Format:    "har",
	}); err != nil {
		t.Fatal(err)
	}

	var progressCalls, lastTotal int
	err := enrich.Run(context.Background(), enrich.Options{
		Capture:        capture,
		SchemaPath:     schemaPath,
		Output:         outPath,
		APIPrefix:      "https://api.example.com/v1",
		Format:         "har",
		EmitPromptsDir: promptsDir,
		OnProgress: func(cur, total int, method, path string) {
			progressCalls++
			lastTotal = total
			if cur < 1 || cur > total || total < 1 || method == "" || path == "" {
				t.Errorf("OnProgress(%d, %d, %q, %q)", cur, total, method, path)
			}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if progressCalls != lastTotal || progressCalls < 1 {
		t.Fatalf("OnProgress calls = %d, total = %d", progressCalls, lastTotal)
	}
	entries, err := os.ReadDir(promptsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 1 {
		t.Fatal("expected prompt files")
	}
}
