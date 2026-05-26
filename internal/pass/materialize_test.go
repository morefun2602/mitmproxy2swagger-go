package pass

import (
	"regexp"
	"strings"
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/schema"
)

type fakeRequest struct {
	url           string
	method        string
	reqHeaders    map[string][]string
	reqBody       []byte
	status        int
	reason        string
	respHeaders   map[string][]string
	respBody      []byte
}

func (f fakeRequest) URL() string { return f.url }

func (f fakeRequest) MatchingURL(prefix string) (string, bool) {
	if strings.HasPrefix(f.url, prefix) {
		return f.url, true
	}
	return "", false
}

func (f fakeRequest) Method() string { return f.method }

func (f fakeRequest) RequestHeaders() map[string][]string { return f.reqHeaders }

func (f fakeRequest) RequestBody() []byte { return f.reqBody }

func (f fakeRequest) ResponseStatusCode() int { return f.status }

func (f fakeRequest) ResponseReason() string { return f.reason }

func (f fakeRequest) ResponseHeaders() map[string][]string { return f.respHeaders }

func (f fakeRequest) ResponseBody() []byte { return f.respBody }

func newTestDocument(t *testing.T) *schema.Document {
	t.Helper()
	doc := schema.New("test")
	doc.EnsureDefaults("https://api.example.com/v1")
	return doc
}

func newTestRunner(t *testing.T, doc *schema.Document, templates ...string) *passRunner {
	t.Helper()
	doc.XPathTemplates = append(doc.XPathTemplates, templates...)
	return newPassRunner(doc, Options{}, "https://api.example.com/v1", regexp.MustCompile("^[0-9]+$"))
}

func TestParseRequestBodyJSON(t *testing.T) {
	val, contentType, ok := parseRequestBody([]byte(`{"name":"widget"}`))
	if !ok {
		t.Fatal("expected JSON body to parse")
	}
	if contentType != "application/json" {
		t.Fatalf("contentType = %q", contentType)
	}
	m, ok := val.(map[string]any)
	if !ok || m["name"] != "widget" {
		t.Fatalf("unexpected value: %#v", val)
	}
}

func TestParseRequestBodyMsgpack(t *testing.T) {
	// msgpack fixmap: {"a": 1}
	body := []byte{0x81, 0xa1, 0x61, 0x01}
	val, contentType, ok := parseRequestBody(body)
	if !ok {
		t.Fatal("expected msgpack body to parse")
	}
	if contentType != "application/msgpack" {
		t.Fatalf("contentType = %q", contentType)
	}
	m, ok := val.(map[string]any)
	if !ok || m["a"] == nil {
		t.Fatalf("unexpected value: %#v", val)
	}
}

func TestParseResponseBodyJSON(t *testing.T) {
	val, contentType, ok := parseResponseBody([]byte(`[1,2,3]`))
	if !ok || contentType != "application/json" {
		t.Fatalf("parseResponseBody() = %#v, %q, %v", val, contentType, ok)
	}
}

func TestMaterializeEndpointSchemaMerge(t *testing.T) {
	doc := newTestDocument(t)
	runner := newTestRunner(t, doc, "/items")

	req := fakeRequest{
		url:      "https://api.example.com/v1/items",
		method:   "GET",
		status:   200,
		reason:   "OK",
		respBody: []byte(`{"first":true}`),
	}

	if err := runner.materializeEndpoint(req, req.url, "/items", 0); err != nil {
		t.Fatal(err)
	}

	ops, ok := doc.PathOperations("/items")
	if !ok {
		t.Fatal("expected /items path")
	}
	methodDoc, _ := ops["get"].(map[string]any)
	responses, _ := methodDoc["responses"].(map[string]any)
	firstResp, _ := responses["200"].(map[string]any)
	firstContent, _ := firstResp["content"].(map[string]any)
	firstJSON, _ := firstContent["application/json"].(map[string]any)
	firstSchema, _ := firstJSON["schema"].(map[string]any)
	firstProps, _ := firstSchema["properties"].(map[string]any)
	if firstProps["first"] == nil {
		t.Fatalf("missing first response schema: %#v", firstSchema)
	}

	req.respBody = []byte(`{"second":true}`)
	if err := runner.materializeEndpoint(req, req.url, "/items", 0); err != nil {
		t.Fatal(err)
	}

	methodDoc, _ = ops["get"].(map[string]any)
	responses, _ = methodDoc["responses"].(map[string]any)
	secondResp, _ := responses["200"].(map[string]any)
	secondContent, _ := secondResp["content"].(map[string]any)
	secondJSON, _ := secondContent["application/json"].(map[string]any)
	secondSchema, _ := secondJSON["schema"].(map[string]any)
	secondProps, _ := secondSchema["properties"].(map[string]any)
	if secondProps["first"] == nil {
		t.Fatal("Schema Merge must not overwrite existing response schema")
	}
	if secondProps["second"] != nil {
		t.Fatal("Schema Merge must not add new keys to existing response schema")
	}
}

func TestMaterializeEndpointRequestBody(t *testing.T) {
	doc := newTestDocument(t)
	runner := newTestRunner(t, doc, "/items")

	req := fakeRequest{
		url:     "https://api.example.com/v1/items",
		method:  "POST",
		status:  201,
		reason:  "Created",
		reqBody: []byte(`{"name":"widget"}`),
	}

	if err := runner.materializeEndpoint(req, req.url, "/items", 0); err != nil {
		t.Fatal(err)
	}

	ops, ok := doc.PathOperations("/items")
	if !ok {
		t.Fatal("expected /items path")
	}
	methodDoc, _ := ops["post"].(map[string]any)
	body, ok := methodDoc["requestBody"].(map[string]any)
	if !ok {
		t.Fatalf("expected requestBody, got %#v", methodDoc)
	}
	content, _ := body["content"].(map[string]any)
	jsonContent, _ := content["application/json"].(map[string]any)
	if jsonContent["schema"] == nil {
		t.Fatal("expected requestBody schema")
	}
}

var _ capture.CapturedRequest = fakeRequest{}
