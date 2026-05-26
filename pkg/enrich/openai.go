package enrich

import (
	"context"
	"fmt"
	"strings"

	"github.com/invopop/jsonschema"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var enrichmentResultSchema = generateEnrichmentSchema()

func generateEnrichmentSchema() any {
	reflector := jsonschema.Reflector{
		AllowAdditionalProperties: false,
		DoNotReference:            true,
	}
	return reflector.Reflect(EnrichmentResult{})
}

func newEnrichmentClient(opts Options) (openai.Client, error) {
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

func callEnrichmentLLM(ctx context.Context, client openai.Client, model, systemPrompt, userPrompt string) (*EnrichmentResult, error) {
	schemaParam := openai.ResponseFormatJSONSchemaJSONSchemaParam{
		Name:        "enrichment_result",
		Description: openai.String("OpenAPI operation enrichment fields"),
		Schema:      enrichmentResultSchema,
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
	return parseEnrichmentResult([]byte(chat.Choices[0].Message.Content))
}
