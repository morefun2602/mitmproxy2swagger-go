package enrich

import "encoding/json"

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
	var out EnrichmentResult
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
