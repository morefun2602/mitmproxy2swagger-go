# LLM Enrichment 作为 `enrich` 子命令

在 Pass 产出 Schema 之后，通过 `mitmproxy2swagger pass` 生成/更新 Schema，再通过 `mitmproxy2swagger enrich` 读取 Schema + HAR archive，调用 OpenAI-compatible LLM，为 Endpoint 补全 operation/参数层语义字段，输出 Enriched Schema。Enrichment 为后处理，不嵌入 Pass；默认 Enrichment Merge（set-if-not-exists），`--force` 可覆盖；HAR 样本默认 `--samples 1`，可增大；送 LLM 前默认 strict Redaction。

**Considered Options**

- **Pass 内嵌 LLM** — 破坏 Pass 确定性，Golden 难维护。
- **独立二进制** — 产品族分裂；用户更倾向同一 CLI。
- **`enrich` 子命令（选用）** — Pass 与 Enrichment 边界清晰；CLI 形态见 [ADR-0005](./0005-cobra-cli-pass-subcommand.md)。

**Consequences**

- 实现 `pkg/enrich`；LLM 集成细节见 [ADR-0004](./0004-enrichment-openai-go-structured-outputs.md)。CI 用 httptest mock OpenAI API，不依赖 API key。
- v1 范围：operation + parameters/requestBody 说明；property 级 description 后续迭代。
