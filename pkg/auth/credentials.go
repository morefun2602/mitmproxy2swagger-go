package auth

import (
	"strings"
)

const (
	kindCookie  = "cookie"
	kindBearer  = "bearer"
	kindAPIKey  = "apiKey"
	kindHeader  = "header"
)

var skipHeaderNames = map[string]struct{}{
	"accept":            {},
	"accept-language":   {},
	"accept-encoding":   {},
	"content-type":      {},
	"content-length":    {},
	"user-agent":        {},
	"host":              {},
	"connection":        {},
	"origin":            {},
	"referer":           {},
	"cache-control":     {},
	"pragma":            {},
	"sec-fetch-site":    {},
	"sec-fetch-mode":    {},
	"sec-fetch-dest":    {},
	"sec-ch-ua":         {},
	"sec-ch-ua-mobile":  {},
	"sec-ch-ua-platform": {},
}

type credentialKey struct {
	kind string
	name string
}

func classifyAuthHeader(name string, values []string) (credentialKey, string, bool) {
	lower := strings.ToLower(strings.TrimSpace(name))
	if _, skip := skipHeaderNames[lower]; skip {
		return credentialKey{}, "", false
	}
	if lower == "cookie" {
		return credentialKey{kind: kindCookie, name: "Cookie"}, sampleCookieNames(values), true
	}
	if lower == "authorization" {
		val := firstValue(values)
		if strings.HasPrefix(strings.ToLower(val), "bearer ") {
			return credentialKey{kind: kindBearer, name: "Authorization"}, "Bearer [REDACTED]", true
		}
		return credentialKey{kind: kindHeader, name: "Authorization"}, "[REDACTED]", true
	}
	if lower == "x-api-key" || lower == "api-key" || lower == "x-apikey" {
		return credentialKey{kind: kindAPIKey, name: lower}, "[REDACTED]", true
	}
	if strings.Contains(lower, "api-key") || strings.Contains(lower, "apikey") {
		return credentialKey{kind: kindAPIKey, name: lower}, "[REDACTED]", true
	}
	if strings.Contains(lower, "auth") || strings.Contains(lower, "token") {
		return credentialKey{kind: kindHeader, name: lower}, "[REDACTED]", true
	}
	return credentialKey{}, "", false
}

func firstValue(values []string) string {
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func sampleCookieNames(values []string) string {
	val := firstValue(values)
	if val == "" {
		return ""
	}
	var names []string
	for _, part := range strings.Split(val, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		n, _, _ := strings.Cut(part, "=")
		n = strings.TrimSpace(n)
		if n != "" {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return "[REDACTED]"
	}
	return strings.Join(names, "; ")
}
