package pass

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/internal/capture"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/schema"
	"github.com/morefun2602/mitmproxy2swagger-go/internal/swaggerutil"
	"github.com/vmihailenco/msgpack/v5"
)

func (r *passRunner) materializeEndpoint(req capture.CapturedRequest, matchURL, path string, pathTemplateIndex int) error {
	pathTemplate := r.pathTemplates[pathTemplateIndex]
	r.doc.SetPathIfNotExists(pathTemplate, map[string]any{})

	ops, ok := r.doc.PathOperations(pathTemplate)
	if !ok {
		return nil
	}

	method := strings.ToLower(req.Method())
	status := req.ResponseStatusCode()

	schema.SetKeyIfNotExists(ops, method, map[string]any{
		"summary":   swaggerutil.PathTemplateToEndpointName(method, pathTemplate),
		"responses": map[string]any{},
	})

	methodDoc, _ := ops[method].(map[string]any)

	if r.opts.Headers {
		headerParams := swaggerutil.RequestToHeaders(req.RequestHeaders())
		if len(headerParams) > 0 {
			schema.SetKeyIfNotExists(methodDoc, "parameters", headerParams)
		}
	}
	params := swaggerutil.URLToParams(matchURL, pathTemplate)
	if len(params) > 0 {
		schema.SetKeyIfNotExists(methodDoc, "parameters", params)
	}

	if method != "get" && method != "head" {
		if body := req.RequestBody(); body != nil {
			if bodyVal, contentType, ok := parseRequestBody(body); ok {
				content := map[string]any{
					"content": map[string]any{
						contentType: map[string]any{
							"schema": swaggerutil.ValueToSchema(bodyVal),
						},
					},
				}
				if r.opts.Examples {
					content["content"].(map[string]any)[contentType].(map[string]any)["example"] = swaggerutil.LimitExampleSize(bodyVal)
				}
				schema.SetKeyIfNotExists(methodDoc, "requestBody", content)
			}
		}
	}

	if respBody := req.ResponseBody(); respBody != nil {
		if parsed, contentType, ok := parseResponseBody(respBody); ok {
			resp := map[string]any{
				"description": req.ResponseReason(),
				"content": map[string]any{
					contentType: map[string]any{
						"schema": swaggerutil.ValueToSchema(parsed),
					},
				},
			}
			if r.opts.Examples {
				resp["content"].(map[string]any)[contentType].(map[string]any)["example"] = swaggerutil.LimitExampleSize(parsed)
			}
			if r.opts.Headers {
				if h := swaggerutil.ResponseToHeaders(req.ResponseHeaders()); len(h) > 0 {
					resp["headers"] = h
				}
			}
			responses, _ := methodDoc["responses"].(map[string]any)
			if responses == nil {
				responses = map[string]any{}
				methodDoc["responses"] = responses
			}
			schema.SetKeyIfNotExists(responses, fmt.Sprint(status), resp)
		}
	}

	responses, _ := methodDoc["responses"].(map[string]any)
	if responses != nil && len(responses) == 0 {
		responses["200"] = map[string]any{
			"description": "OK",
			"content":     map[string]any{},
		}
	}

	return nil
}

func parseRequestBody(body []byte) (any, string, bool) {
	var jsonVal any
	if err := json.Unmarshal(body, &jsonVal); err == nil {
		return jsonVal, "application/json", true
	}
	var msgpackVal any
	if err := msgpack.Unmarshal(body, &msgpackVal); err == nil {
		return msgpackVal, "application/msgpack", true
	}
	form, err := url.ParseQuery(string(body))
	if err == nil && len(form) > 0 {
		out := map[string]any{}
		for key, values := range form {
			if len(values) > 0 {
				out[key] = values[0]
			}
		}
		if len(out) > 0 {
			return out, "application/x-www-form-urlencoded", true
		}
	}
	return nil, "", false
}

func parseResponseBody(body []byte) (any, string, bool) {
	var jsonVal any
	if err := json.Unmarshal(body, &jsonVal); err == nil {
		return jsonVal, "application/json", true
	}
	var msgpackVal any
	if err := msgpack.Unmarshal(body, &msgpackVal); err == nil {
		return msgpackVal, "application/msgpack", true
	}
	return nil, "", false
}
