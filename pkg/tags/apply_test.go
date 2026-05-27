package tags_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/tags"
)

func TestRunApply_prefixAndOperationOverride(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	tagsPath := filepath.Join(dir, "tags.yaml")

	schemaYAML := `openapi: 3.0.0
info:
  title: test
  version: 1.0.0
paths:
  /v2/recent/chats:
    get:
      summary: list
      tags: [聊天, 会话]
      responses:
        "200":
          description: ok
  /v1/health:
    get:
      summary: health
      tags: [系统监控]
      responses:
        "200":
          description: ok
`
	if err := os.WriteFile(schemaPath, []byte(schemaYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	tagsYAML := `tags:
  - name: IM
    description: 即时通讯
  - name: 系统
x-tagGroups:
  - name: 平台
    tags: [系统]
prefixRules:
  - prefix: /v2/recent
    tags: [IM]
  - prefix: /v1
    tags: [系统]
operations:
  GET /v1/health: [系统]
`
	if err := os.WriteFile(tagsPath, []byte(tagsYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := tags.RunApply(tags.ApplyOptions{
		Schema:   schemaPath,
		TagsFile: tagsPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Updated != 2 {
		t.Fatalf("updated = %d, want 2", res.Updated)
	}

	doc, err := schema.Load(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	chats, _ := doc.PathOperations("/v2/recent/chats")
	get, _ := chats["get"].(map[string]any)
	if got := tagSlice(get); len(got) != 1 || got[0] != "IM" {
		t.Fatalf("recent chats tags = %#v", got)
	}

	health, _ := doc.PathOperations("/v1/health")
	hget, _ := health["get"].(map[string]any)
	if got := tagSlice(hget); len(got) != 1 || got[0] != "系统" {
		t.Fatalf("health tags = %#v", got)
	}
	if len(doc.Tags) != 2 {
		t.Fatalf("top-level tags = %#v", doc.Tags)
	}
	if len(doc.XTagGroups) != 1 {
		t.Fatalf("x-tagGroups = %#v", doc.XTagGroups)
	}
}

func TestRunApply_longestPrefixWins(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	tagsPath := filepath.Join(dir, "tags.yaml")

	schemaYAML := `openapi: 3.0.0
info:
  title: test
  version: 1.0.0
paths:
  /v3/chats/{id}/messages:
    get:
      responses:
        "200":
          description: ok
`
	if err := os.WriteFile(schemaPath, []byte(schemaYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	tagsYAML := `prefixRules:
  - prefix: /v3
    tags: [宽]
  - prefix: /v3/chats
    tags: [IM]
`
	if err := os.WriteFile(tagsPath, []byte(tagsYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := tags.RunApply(tags.ApplyOptions{Schema: schemaPath, TagsFile: tagsPath})
	if err != nil {
		t.Fatal(err)
	}

	doc, _ := schema.Load(schemaPath)
	ops, _ := doc.PathOperations("/v3/chats/{id}/messages")
	get, _ := ops["get"].(map[string]any)
	if got := tagSlice(get); len(got) != 1 || got[0] != "IM" {
		t.Fatalf("tags = %#v, want [IM]", got)
	}
}

func TestRunApply_mergeKeepsFirstWhenCombined(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	tagsPath := filepath.Join(dir, "tags.yaml")

	schemaYAML := `openapi: 3.0.0
info:
  title: test
  version: 1.0.0
paths:
  /items:
    get:
      tags: [旧标签]
      responses:
        "200":
          description: ok
`
	if err := os.WriteFile(schemaPath, []byte(schemaYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	tagsYAML := `prefixRules:
  - prefix: /items
    tags: [新标签]
`
	if err := os.WriteFile(tagsPath, []byte(tagsYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := tags.RunApply(tags.ApplyOptions{
		Schema:   schemaPath,
		TagsFile: tagsPath,
		Merge:    true,
	})
	if err != nil {
		t.Fatal(err)
	}

	doc, _ := schema.Load(schemaPath)
	ops, _ := doc.PathOperations("/items")
	get, _ := ops["get"].(map[string]any)
	got := tagSlice(get)
	if len(got) != 1 || got[0] != "新标签" {
		t.Fatalf("merge should prefer sidecar tag, got %#v", got)
	}
}

func TestRunApply_strictUnmatched(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	tagsPath := filepath.Join(dir, "tags.yaml")

	schemaYAML := `openapi: 3.0.0
info:
  title: test
  version: 1.0.0
paths:
  /unknown:
    get:
      responses:
        "200":
          description: ok
`
	if err := os.WriteFile(schemaPath, []byte(schemaYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	tagsYAML := `prefixRules:
  - prefix: /v1
    tags: [系统]
`
	if err := os.WriteFile(tagsPath, []byte(tagsYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := tags.RunApply(tags.ApplyOptions{
		Schema:   schemaPath,
		TagsFile: tagsPath,
		Strict:   true,
	})
	if err == nil {
		t.Fatal("expected strict error for unmatched operation")
	}
}

func tagSlice(op map[string]any) []string {
	raw, _ := op["tags"].([]any)
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		if s, ok := v.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
