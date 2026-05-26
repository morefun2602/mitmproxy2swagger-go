package swaggerutil

import (
	"net/url"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var verbs = map[string]struct{}{
	"add": {}, "create": {}, "delete": {}, "get": {}, "attach": {},
	"detach": {}, "update": {}, "push": {}, "extendedcreate": {}, "activate": {},
}

const (
	maxExampleArrayElements     = 10
	maxExampleObjectProperties  = 150
)

// PathTemplateToEndpointName generates a summary from method and path template.
func PathTemplateToEndpointName(method, pathTemplate string) string {
	pathTemplate = strings.Trim(pathTemplate, "/")
	segments := strings.Split(pathTemplate, "/")
	var params []string
	var kept []string
	for _, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			params = append(params, segment)
		} else {
			kept = append(kept, segment)
		}
	}
	for i, j := 0, len(kept)-1; i < j; i, j = i+1, j-1 {
		kept[i], kept[j] = kept[j], kept[i]
	}
	var nameParts []string
	for _, segment := range kept {
		if _, ok := verbs[strings.ToLower(segment)]; ok {
			nameParts = append([]string{strings.ToLower(segment)}, nameParts...)
		} else {
			nameParts = append([]string{strings.ToLower(segment)}, nameParts...)
			break
		}
	}
	if len(params) > 0 {
		param := strings.Trim(params[0], "{}")
		nameParts = append(nameParts, "by "+param)
	}
	return strings.ToUpper(method) + " " + strings.Join(nameParts, " ")
}

// URLToParams builds OpenAPI parameters from a URL and path template.
func URLToParams(rawURL, pathTemplate string) []map[string]any {
	pathTemplate = strings.Trim(pathTemplate, "/")
	segments := strings.Split(pathTemplate, "/")
	urlPath := strings.Split(strings.Split(rawURL, "?")[0], "?")[0]
	urlPath = strings.Trim(urlPath, "/")
	urlSegments := strings.Split(urlPath, "/")

	var params []map[string]any
	for idx, segment := range segments {
		if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
			paramType := "string"
			if idx < len(urlSegments) && isDigits(urlSegments[idx]) {
				paramType = "number"
			}
			params = append(params, map[string]any{
				"name":     strings.Trim(segment, "{}"),
				"in":       "path",
				"required": true,
				"schema":   map[string]any{"type": paramType},
			})
		}
	}

	parsed, err := url.Parse(rawURL)
	if err == nil && parsed.RawQuery != "" {
		query := parsed.Query()
		for _, key := range orderedQueryKeys(parsed.RawQuery) {
			values := query[key]
			paramType := "string"
			if len(values) > 0 && isDigits(values[0]) {
				paramType = "number"
			}
			params = append(params, map[string]any{
				"name":     key,
				"in":       "query",
				"required": false,
				"schema":   map[string]any{"type": paramType},
			})
		}
	}
	return params
}

func orderedQueryKeys(rawQuery string) []string {
	var keys []string
	seen := make(map[string]struct{})
	for _, part := range strings.Split(rawQuery, "&") {
		if part == "" {
			continue
		}
		key, _, _ := strings.Cut(part, "=")
		decoded, err := url.QueryUnescape(key)
		if err == nil {
			key = decoded
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		keys = append(keys, key)
	}
	return keys
}

func RequestToHeaders(headers map[string][]string) []map[string]any {
	var params []map[string]any
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		paramType := "string"
		if isDigits(values[0]) {
			paramType = "number"
		}
		params = append(params, map[string]any{
			"name":     key,
			"in":       "header",
			"required": false,
			"schema":   map[string]any{"type": paramType},
		})
	}
	return params
}

func ResponseToHeaders(headers map[string][]string) map[string]any {
	out := map[string]any{}
	for key, values := range headers {
		if len(values) == 0 {
			continue
		}
		paramType := "string"
		if isDigits(values[0]) {
			paramType = "number"
		}
		out[key] = map[string]any{
			"description": values[0],
			"schema":      map[string]any{"type": paramType},
		}
	}
	return out
}

// ValueToSchema infers an OpenAPI schema from a decoded JSON/msgpack value.
func ValueToSchema(value any) map[string]any {
	switch v := value.(type) {
	case float64:
		return map[string]any{"type": "number"}
	case float32:
		return map[string]any{"type": "number"}
	case int, int32, int64, uint, uint32, uint64:
		return map[string]any{"type": "number"}
	case bool:
		return map[string]any{"type": "boolean"}
	case string:
		return map[string]any{"type": "string"}
	case []any:
		if len(v) == 0 {
			return map[string]any{"type": "array", "items": map[string]any{}}
		}
		return map[string]any{"type": "array", "items": ValueToSchema(v[0])}
	case map[string]any:
		if len(v) == 0 {
			return map[string]any{"type": "object", "properties": map[string]any{}}
		}
		allNumeric := true
		allUUID := true
		for key := range v {
			if !isNumericString(key) {
				allNumeric = false
			}
			if !isUUID(key) {
				allUUID = false
			}
		}
		if (allNumeric || allUUID) && len(v) > 0 {
			for _, val := range v {
				return map[string]any{
					"type":                 "object",
					"additionalProperties": ValueToSchema(val),
				}
			}
		}
		props := map[string]any{}
		for key, val := range v {
			props[key] = ValueToSchema(val)
		}
		return map[string]any{"type": "object", "properties": props}
	case nil:
		return map[string]any{"type": "object", "nullable": true}
	default:
		return map[string]any{"type": "string"}
	}
}

func LimitExampleSize(example any) any {
	switch v := example.(type) {
	case []any:
		var out []any
		for _, el := range v {
			if len(out) >= maxExampleArrayElements {
				break
			}
			out = append(out, LimitExampleSize(el))
		}
		return out
	case map[string]any:
		out := map[string]any{}
		for key, val := range v {
			if len(out) >= maxExampleObjectProperties {
				break
			}
			out[key] = LimitExampleSize(val)
		}
		return out
	default:
		return example
	}
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isNumericString(key string) bool {
	_, err := strconv.ParseUint(key, 10, 64)
	return err == nil && key != ""
}

func isUUID(key string) bool {
	_, err := uuid.Parse(key)
	return err == nil
}
