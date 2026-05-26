package enrich

import (
	"encoding/json"
	"strings"
)

// EnrichmentResult is the structured LLM output for one Endpoint.
type EnrichmentResult struct {
	Summary                string            `json:"summary"`
	Description            string            `json:"description"`
	Tags                   []string          `json:"tags"`
	OperationID            string            `json:"operationId"`
	ParameterDescriptions  map[string]string `json:"parameterDescriptions"`
	RequestBodyDescription string            `json:"requestBodyDescription"`
}

func parseEnrichmentResult(data []byte) (*EnrichmentResult, error) {
	jsonBytes := extractJSONPayload(data)
	var out EnrichmentResult
	if err := json.Unmarshal(jsonBytes, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func extractJSONPayload(data []byte) []byte {
	s := strings.TrimSpace(string(data))
	if s == "" {
		return data
	}
	if strings.HasPrefix(s, "```") {
		if idx := strings.Index(s, "\n"); idx >= 0 {
			s = s[idx+1:]
		}
		if end := strings.LastIndex(s, "```"); end >= 0 {
			s = s[:end]
		}
		s = strings.TrimSpace(s)
	}
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start >= 0 && end > start {
		return []byte(s[start : end+1])
	}
	return []byte(s)
}
