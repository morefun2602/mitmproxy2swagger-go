# `curate` 子命令（Curation Assist）

First Pass 默认会为每个观察到的 Literal Path 与 Suggested Path Template 各写一条 `ignore:` 条目，大型 HAR 下 `x-path-templates` 可达数百行，人工 **Curation** 成本高。新增 `mitmproxy2swagger curate` 子命令，在 **Curation** 之前对 `x-path-templates` 做 **Curation Assist**：确定性 **Template Clustering**（`--auto`）与 LLM **Template Suggestion**（`--llm-suggest` 写 sidecar、`--apply-suggestions` 写回），不修改 **First Pass** 行为与 Golden Schema。

**Considered Options**

- **改 First Pass 默认** — 一次到位，但 breaking Golden、模糊 Pass 与 Curation 边界。
- **独立 `curate`（选用）** — opt-in；`pkg/curate` 可库化；先交付 `--auto`，再交付 LLM 建议与 apply。
- **仅 Makefile 脚本** — 难复用、难测。

**Consequences**

- 推荐：`pass1` → `curate --auto` →（可选）`curate --llm-suggest` → 编辑 `template-suggestions.yaml` → `curate --apply-suggestions` → 人工去 `ignore:` → `pass2` → `enrich`（见 Makefile `example-*` target）。
- `--auto`：去掉已被参数化模板覆盖的字面量、合并仅数字段不同的路径骨架、按 **Template Precedence** 排序。
- `--llm-suggest` / `--apply-suggestions`：仅对**共享前缀、末段 slug 不同**的字面量路径分组后送 LLM；`proposed_template` 须 regex 匹配组内全部路径否则丢弃；sidecar 含 `proposed_template`、`replaces`、`sample_paths`、`reason`；与 `enrich` 共用 `OPENAI_*` 与 `--model` / `--base-url`。
- 术语见 CONTEXT.md：**Curation Assist**、**Template Clustering**、**Template Suggestion**。
