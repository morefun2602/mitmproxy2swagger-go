package schema_test

import (
	"testing"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
)

func TestSortPathTemplates_specificBeforeWildcard(t *testing.T) {
	in := []string{
		"ignore:/v3/chats/{id}/messages/{id1}",
		"/v3/chats/{id}/messages/read_status",
		"/v1/health",
		"ignore:/v3/chats/{id}/messages",
	}
	schema.SortPathTemplates(in)

	want := []string{
		"/v1/health",
		"/v3/chats/{id}/messages/read_status",
		"ignore:/v3/chats/{id}/messages",
		"ignore:/v3/chats/{id}/messages/{id1}",
	}
	for i := range want {
		if in[i] != want[i] {
			t.Fatalf("index %d: got %q want %q (full: %#v)", i, in[i], want[i], in)
		}
	}
}
