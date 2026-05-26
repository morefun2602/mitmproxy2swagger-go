package enrich

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/morefun2602/mitmproxy2swagger-go/pkg/schema"
)

func applyEnrichment(op map[string]any, result *EnrichmentResult, force bool) {
	if result == nil {
		return
	}
	setField(op, "summary", result.Summary, force)
	setField(op, "description", result.Description, force)
	if len(result.Tags) > 0 {
		setField(op, "tags", result.Tags, force)
	}
	setField(op, "operationId", result.OperationID, force)

	if len(result.ParameterDescriptions) > 0 {
		applyParameterDescriptions(op, result.ParameterDescriptions, force)
	}
	if result.RequestBodyDescription != "" {
		applyRequestBodyDescription(op, result.RequestBodyDescription, force)
	}
}

func setField(m map[string]any, key string, value any, force bool) {
	if value == nil {
		return
	}
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return
		}
	case []string:
		if len(v) == 0 {
			return
		}
	}
	if force {
		m[key] = value
		return
	}
	schema.SetKeyIfNotExists(m, key, value)
}

func applyParameterDescriptions(op map[string]any, descriptions map[string]string, force bool) {
	raw, ok := op["parameters"]
	if !ok {
		return
	}
	params, ok := raw.([]any)
	if !ok {
		return
	}
	for i, item := range params {
		param, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name, _ := param["name"].(string)
		desc, ok := descriptions[name]
		if !ok || strings.TrimSpace(desc) == "" {
			continue
		}
		if force {
			param["description"] = desc
		} else if _, exists := param["description"]; !exists {
			param["description"] = desc
		}
		params[i] = param
	}
	op["parameters"] = params
}

func applyRequestBodyDescription(op map[string]any, description string, force bool) {
	raw, ok := op["requestBody"]
	if !ok {
		return
	}
	body, ok := raw.(map[string]any)
	if !ok {
		return
	}
	if force {
		body["description"] = description
	} else {
		schema.SetKeyIfNotExists(body, "description", description)
	}
	op["requestBody"] = body
}

func buildUserPrompt(pathTemplate, method string, op map[string]any, samples []RequestSample) (string, error) {
	payload := map[string]any{
		"pathTemplate": pathTemplate,
		"method":       strings.ToUpper(method),
		"operation":    op,
		"samples":      samples,
	}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(`根据 HTTP 样本为下列 OpenAPI operation 补全语义字段。

返回单个 JSON 对象，字段如下（除 operationId 外均使用简体中文）：
- summary (string)：简短中文摘要
- description (string)：中文说明
- tags (string array, optional)：中文标签
- operationId (string, optional)：英文 camelCase 标识符
- parameterDescriptions (object)：参数名 → 中文说明
- requestBodyDescription (string, optional)：中文说明

Operation context:
%s`, string(data)), nil
}

const systemPrompt = `你是 API 文档助手。根据 HTTP 流量样本推断准确、简洁的 OpenAPI 语义字段。summary、description、tags、parameterDescriptions、requestBodyDescription 必须使用简体中文；operationId 使用英文 camelCase。不要编造样本未支持的 endpoint 或字段。`
