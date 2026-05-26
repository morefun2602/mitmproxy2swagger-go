# Enrichment 使用 openai-go 与 Structured Outputs

Enrichment 的 LLM 调用改用官方 `github.com/openai/openai-go/v3`（v3.22.0），通过 Chat Completions 的 `json_schema` response format（`strict: true`）约束输出形状，对齐 `EnrichmentResult`。保留 `--base-url`（`option.WithBaseURL`）以兼容 OpenAI-compatible 端点；删除手写 HTTP 客户端与 `LLMClient` interface。`Options.Client` 可注入 `*openai.Client`；默认 client 由 `openai.go` 的 `newEnrichmentClient` 构造。CI 用 `httptest` mock `/v1/chat/completions`，不依赖真实 API key。

**Considered Options**

- **手写 HTTP + 自由文本 JSON** — 需 `extractJSON` 容错，输出不稳定。
- **SDK + `LLMClient` adapter** — 多一层 indirection，测试仍要 mock interface。
- **官方 SDK + Structured Outputs + 可注入 Client（选用）** — schema 约束输出；测试指向 httptest base URL。

**Consequences**

- 新增 `pkg/enrich/openai.go` 集中 SDK 调用与 JSON Schema 定义。
- `parseEnrichmentResult([]byte)` 保留于 `llm.go`；删除 `extractJSON`。
- 默认 model 仍为 `gpt-4o-mini`（`--model` 可覆盖）。
