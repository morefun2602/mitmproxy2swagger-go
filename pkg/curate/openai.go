package curate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

type templateSuggestionLLM struct {
	ProposedTemplate string `json:"proposed_template"`
	Confidence       string `json:"confidence"`
	Reason           string `json:"reason"`
}

var templateSuggestionSchema = generateTemplateSuggestionSchema()

func generateTemplateSuggestionSchema() any {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	return reflector.Reflect(templateSuggestionLLM{})
}

func newCurateClient(opts Options) (openai.Client, error) {
	if opts.Client != nil {
		return *opts.Client, nil
	}
	if opts.APIKey == "" {
		return openai.Client{}, fmt.Errorf("missing API key (set --api-key or OPENAI_API_KEY)")
	}
	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	return openai.NewClient(
		option.WithAPIKey(opts.APIKey),
		option.WithBaseURL(baseURL),
	), nil
}

func defaultModel(model string) string {
	if strings.TrimSpace(model) == "" {
		return "gpt-4o-mini"
	}
	return model
}

func callTemplateSuggestionLLM(ctx context.Context, client openai.Client, model, systemPrompt, userPrompt string) (*templateSuggestionLLM, error) {
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "template_suggestion",
		Description: openai.String("Path template merge suggestion"),
		Schema:      templateSuggestionSchema,
		Strict:      openai.Bool(true),
	}

	chat, err := client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: defaultModel(model),
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemPrompt),
			openai.UserMessage(userPrompt),
		},
		ResponseFormat: openai.ChatCompletionNewParamsResponseFormatUnion{
			OfJSONSchema: &openai.ResponseFormatJSONSchemaParam{JSONSchema: schemaParam},
		},
	})
	if err != nil {
		return nil, err
	}
	if len(chat.Choices) == 0 {
		return nil, fmt.Errorf("LLM API returned no choices")
	}
	var out templateSuggestionLLM
	if err := json.Unmarshal([]byte(chat.Choices[0].Message.Content), &out); err != nil {
		return nil, fmt.Errorf("parse LLM JSON: %w", err)
	}
	return &out, nil
}

const templateSuggestionSystemPrompt = `你是 API 路径模板专家。根据多条观察到的 HTTP 路径，判断是否应合并为一条 OpenAPI path template（使用 {param} 占位符）。

规则：
- 仅当路径骨架相同、仅个别段为不同标识（如应用 key、资源 slug）时建议合并。
- 占位符名使用 camelCase 且语义清晰（如 appKey、fileId、chatId）。
- proposed_template 只输出路径（以 / 开头），不要带 ignore: 前缀。
- confidence 为 high、medium 或 low 之一。
- reason 用一句中文说明。
- 若不应合并，proposed_template 设为空字符串，confidence 为 low，reason 说明原因。`

func buildTemplateSuggestionUserPrompt(group mergeCandidate, samplePaths []string) string {
	var b strings.Builder
	b.WriteString("以下路径是否应合并为一条 path template？\n\n")
	b.WriteString("本组路径共享前缀：")
	b.WriteString(group.prefix)
	b.WriteString("\n仅末段为不同 slug 标识；若实际不属于同一资源前缀，必须返回空 proposed_template。\n\n观察到的路径：\n")
	for _, e := range group.entries {
		b.WriteString("- ")
		b.WriteString(e.path)
		b.WriteByte('\n')
	}
	if len(samplePaths) > 0 {
		b.WriteString("\nHAR 中的样例路径：\n")
		for _, p := range samplePaths {
			b.WriteString("- ")
			b.WriteString(p)
			b.WriteByte('\n')
		}
	}
	return b.String()
}
