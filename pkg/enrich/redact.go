package enrich

import (
	"encoding/json"
	"strings"
)

var sensitiveHeaderNames = map[string]struct{}{
	"authorization": {},
	"cookie":        {},
	"set-cookie":    {},
	"x-api-key":     {},
	"x-auth-token":  {},
}

var sensitiveBodyKeys = map[string]struct{}{
	"password":      {},
	"token":         {},
	"access_token":  {},
	"refresh_token": {},
	"secret":        {},
	"api_key":       {},
	"apikey":        {},
}

type RedactMode string

const (
	RedactStrict RedactMode = "strict"
	RedactOff    RedactMode = "off"
)

func redactHeaders(headers map[string][]string, mode RedactMode) map[string][]string {
	if mode == RedactOff || len(headers) == 0 {
		return headers
	}
	out := make(map[string][]string, len(headers))
	for name, values := range headers {
		key := strings.ToLower(name)
		if _, ok := sensitiveHeaderNames[key]; ok || strings.Contains(key, "token") {
			out[name] = []string{"[REDACTED]"}
			continue
		}
		out[name] = values
	}
	return out
}

func redactBodyText(body []byte, mode RedactMode) string {
	if len(body) == 0 {
		return ""
	}
	if mode == RedactOff {
		return string(body)
	}
	var parsed any
	if err := json.Unmarshal(body, &parsed); err != nil {
		return string(body)
	}
	redactValue(parsed, mode)
	data, err := json.Marshal(parsed)
	if err != nil {
		return string(body)
	}
	return string(data)
}

func redactValue(v any, mode RedactMode) {
	if mode == RedactOff {
		return
	}
	switch t := v.(type) {
	case map[string]any:
		for key, val := range t {
			if _, ok := sensitiveBodyKeys[strings.ToLower(key)]; ok {
				t[key] = "[REDACTED]"
				continue
			}
			redactValue(val, mode)
		}
	case []any:
		for i := range t {
			redactValue(t[i], mode)
		}
	}
}
