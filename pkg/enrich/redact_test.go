package enrich

import (
	"testing"
)

func TestRedactHeadersStrict(t *testing.T) {
	headers := map[string][]string{
		"Authorization": {"Bearer secret"},
		"Accept":        {"application/json"},
		"X-Auth-Token":  {"abc"},
	}
	out := redactHeaders(headers, RedactStrict)
	if out["Accept"][0] != "application/json" {
		t.Fatalf("Accept header changed: %#v", out["Accept"])
	}
	if out["Authorization"][0] != "[REDACTED]" {
		t.Fatalf("Authorization not redacted: %#v", out["Authorization"])
	}
	if out["X-Auth-Token"][0] != "[REDACTED]" {
		t.Fatalf("token header not redacted")
	}
}

func TestRedactBodyStrict(t *testing.T) {
	body := []byte(`{"token":"secret","name":"widget"}`)
	out := redactBodyText(body, RedactStrict)
	if out == string(body) {
		t.Fatal("expected body mutation")
	}
	if contains(out, "secret") {
		t.Fatalf("token value leaked: %s", out)
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexSubstring(s, sub) >= 0)
}

func indexSubstring(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
