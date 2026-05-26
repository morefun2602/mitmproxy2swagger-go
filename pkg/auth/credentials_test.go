package auth

import "testing"

func TestClassifyAuthHeader_cookieNamesOnly(t *testing.T) {
	key, sample, ok := classifyAuthHeader("Cookie", []string{"session=secret; csrf=abc"})
	if !ok || key.kind != kindCookie {
		t.Fatalf("got ok=%v key=%+v", ok, key)
	}
	if sample != "session; csrf" {
		t.Fatalf("sample = %q", sample)
	}
}

func TestClassifyAuthHeader_bearer(t *testing.T) {
	_, sample, ok := classifyAuthHeader("Authorization", []string{"Bearer xyz"})
	if !ok || sample != "Bearer [REDACTED]" {
		t.Fatalf("got ok=%v sample=%q", ok, sample)
	}
}
