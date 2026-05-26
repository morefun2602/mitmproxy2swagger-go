package auth_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/auth"
	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
	"gopkg.in/yaml.v3"
)

func TestRunApply_cookieAndBearer(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	obsPath := filepath.Join(dir, "auth-observations.yaml")

	schemaYAML := `openapi: 3.0.0
info:
  title: test
  version: 1.0.0
paths:
  /users/token:
    get:
      responses:
        "200":
          description: ok
  /items:
    get:
      responses:
        "200":
          description: ok
`
	if err := os.WriteFile(schemaPath, []byte(schemaYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	obsYAML := `observed:
  - kind: cookie
    name: Cookie
    sample_value: session_id=abc; csrf=def
  - kind: bearer
    name: Authorization
verified:
  - cookie
  - bearer
combination: and
suggested_auth_paths:
  - path: /users/token
`
	if err := os.WriteFile(obsPath, []byte(obsYAML), 0o644); err != nil {
		t.Fatal(err)
	}

	err := auth.RunApply(auth.ApplyOptions{
		Schema:           schemaPath,
		Observations:     obsPath,
		TagAuthEndpoints: true,
	})
	if err != nil {
		t.Fatal(err)
	}

	doc, err := schema.Load(schemaPath)
	if err != nil {
		t.Fatal(err)
	}
	schemes, ok := doc.Components["securitySchemes"].(map[string]any)
	if !ok || len(schemes) < 2 {
		t.Fatalf("securitySchemes: %#v", doc.Components)
	}
	cookie, _ := schemes["sessionCookie"].(map[string]any)
	if cookie["in"] != "cookie" || cookie["name"] != "session_id" {
		t.Fatalf("sessionCookie: %#v", cookie)
	}
	if len(doc.Security) != 1 {
		t.Fatalf("security = %#v", doc.Security)
	}
	req, _ := doc.Security[0].(map[string]any)
	if _, ok := req["sessionCookie"]; !ok {
		t.Fatalf("expected sessionCookie in and requirement: %#v", req)
	}
	if _, ok := req["bearerAuth"]; !ok {
		t.Fatalf("expected bearerAuth in and requirement: %#v", req)
	}

	ops, ok := doc.PathOperations("/users/token")
	if !ok {
		t.Fatal("missing /users/token")
	}
	getOp, _ := ops["get"].(map[string]any)
	tags, _ := getOp["tags"].([]any)
	if len(tags) == 0 || fmt.Sprint(tags[0]) != "auth" {
		t.Fatalf("expected auth tag on /users/token: %#v", tags)
	}
}

func TestRunApply_emptyVerified(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	obsPath := filepath.Join(dir, "obs.yaml")
	_ = os.WriteFile(schemaPath, []byte("openapi: 3.0.0\ninfo:\n  title: t\n  version: 1\n"), 0o644)
	_ = os.WriteFile(obsPath, []byte("verified: []\nobserved: []\n"), 0o644)

	err := auth.RunApply(auth.ApplyOptions{Schema: schemaPath, Observations: obsPath})
	if err == nil || !strings.Contains(err.Error(), "verified is empty") {
		t.Fatalf("err = %v", err)
	}
}

func TestRunApply_orCombination(t *testing.T) {
	dir := t.TempDir()
	schemaPath := filepath.Join(dir, "schema.yaml")
	obsPath := filepath.Join(dir, "obs.yaml")
	_ = os.WriteFile(schemaPath, []byte(`openapi: 3.0.0
info: {title: t, version: "1"}
paths: {}
`), 0o644)
	_ = os.WriteFile(obsPath, []byte(`observed:
  - kind: cookie
    name: Cookie
    sample_value: a=1
  - kind: bearer
    name: Authorization
verified: [cookie, bearer]
combination: or
`), 0o644)

	if err := auth.RunApply(auth.ApplyOptions{Schema: schemaPath, Observations: obsPath}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(schemaPath)
	var raw map[string]any
	if err := yaml.Unmarshal(data, &raw); err != nil {
		t.Fatal(err)
	}
	sec, _ := raw["security"].([]any)
	if len(sec) != 2 {
		t.Fatalf("or security want 2 entries, got %#v", sec)
	}
}
