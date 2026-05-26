package curate_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/curate"
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

func mockTemplateSuggestServer(t *testing.T, proposed, confidence, reason string) *httptest.Server {
	t.Helper()
	content, err := json.Marshal(map[string]string{
		"proposed_template": proposed,
		"confidence":        confidence,
		"reason":            reason,
	})
	if err != nil {
		t.Fatal(err)
	}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/chat/completions" {
			http.NotFound(w, r)
			return
		}
		resp := map[string]any{
			"id":     "chatcmpl-test",
			"object": "chat.completion",
			"choices": []map[string]any{{
				"message": map[string]any{
					"role":    "assistant",
					"content": string(content),
				},
				"finish_reason": "stop",
			}},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
}

func TestLLMSuggestVerticalSlice(t *testing.T) {
	root := repoRoot(t)
	capture := filepath.Join(root, "testdata", "captures", "minimal.har")
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	suggestionsPath := filepath.Join(dir, "template-suggestions.yaml")

	doc := schema.New("test.har")
	doc.XPathTemplates = []string{
		"ignore:/items/foo",
		"ignore:/items/bar",
	}
	if err := doc.Save(schemaPath); err != nil {
		t.Fatal(err)
	}

	srv := mockTemplateSuggestServer(t, "/items/{itemKey}", "high", "同一资源集合下的条目标识")
	t.Cleanup(srv.Close)

	client := openai.NewClient(
		option.WithBaseURL(srv.URL+"/v1"),
		option.WithAPIKey("test-key"),
	)

	err := curate.Run(context.Background(), curate.Options{
		Schema:         schemaPath,
		LLMSuggest:     true,
		Capture:        capture,
		APIPrefix:      "https://api.example.com/v1",
		Format:         "har",
		SuggestionsOut: suggestionsPath,
		Client:         &client,
		MaxGroups:      10,
	})
	if err != nil {
		t.Fatal(err)
	}

	file, err := curate.LoadSuggestionsFile(suggestionsPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(file.Suggestions) < 1 {
		t.Fatalf("expected suggestions, got %d", len(file.Suggestions))
	}
	if file.Suggestions[0].ProposedTemplate != "ignore:/items/{itemKey}" {
		t.Fatalf("unexpected proposed: %q", file.Suggestions[0].ProposedTemplate)
	}
}

func TestLLMSuggestRejectsMismatchedProposed(t *testing.T) {
	root := repoRoot(t)
	capture := filepath.Join(root, "testdata", "captures", "minimal.har")
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	suggestionsPath := filepath.Join(dir, "template-suggestions.yaml")

	doc := schema.New("test.har")
	doc.XPathTemplates = []string{
		"ignore:/items/foo",
		"ignore:/items/bar",
	}
	if err := doc.Save(schemaPath); err != nil {
		t.Fatal(err)
	}

	srv := mockTemplateSuggestServer(t, "/v1/apps/info/{appKey}", "high", "错误合并")
	t.Cleanup(srv.Close)

	client := openai.NewClient(
		option.WithBaseURL(srv.URL+"/v1"),
		option.WithAPIKey("test-key"),
	)

	err := curate.Run(context.Background(), curate.Options{
		Schema:         schemaPath,
		LLMSuggest:     true,
		Capture:        capture,
		APIPrefix:      "https://api.example.com/v1",
		Format:         "har",
		SuggestionsOut: suggestionsPath,
		Client:         &client,
		MaxGroups:      10,
	})
	if err != nil {
		t.Fatal(err)
	}

	file, err := curate.LoadSuggestionsFile(suggestionsPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(file.Suggestions) != 0 {
		t.Fatalf("expected no suggestions after validation failure, got %d", len(file.Suggestions))
	}
}
