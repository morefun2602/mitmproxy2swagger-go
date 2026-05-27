# mitmproxy2swagger

[English](README.md) | [中文](README.zh-CN.md)

[mitmproxy2swagger](https://github.com/alufers/mitmproxy2swagger) 的 Go 实现。

从 HTTP 抓包文件反向生成 [OpenAPI 3.0](https://swagger.io/specification/) YAML Schema 的 CLI 工具。

## 安装

需要 Go **1.26.1** 或更高版本。

**安装二进制：**

```bash
go install github.com/morefun2602/mitmproxy2swagger-go/cmd/mitmproxy2swagger@latest
```

**或从源码构建：**

```bash
git clone https://github.com/morefun2602/mitmproxy2swagger-go.git
cd mitmproxy2swagger-go
go build -o mitmproxy2swagger ./cmd/mitmproxy2swagger
```

## 用法

子命令：

- `pass` — 对抓包运行 Pass，更新 OpenAPI Schema
- `curate` — 聚类 `x-path-templates`（`--auto`）、LLM 合并建议（`--llm-suggest`）、应用建议（`--apply-suggestions`）；见 [ADR-0007](docs/adr/0007-curate-curation-assist.md)
- `auth observe` — 扫描抓包中的鉴权相关 header，写出 `auth-observations.yaml`；见 [ADR-0008](docs/adr/0008-auth-observation-two-phase.md)
- `enrich` — LLM 语义增强（见 [ADR-0003](docs/adr/0003-llm-enrichment-subcommand.md)）
- `tags apply` — 从 `tags.yaml` 侧车写入 operation 分组与 `x-tagGroups`（见下文）
- `version` — 打印构建版本
- `completion` — 生成 shell 补全脚本

工具对同一份抓包文件和 Schema 文件执行两遍：

1. **第一遍** — 发现 URL 路径，将候选路径模板写入 `x-path-templates`（默认带 `ignore:` 前缀）。
2. **人工筛选（Curation）** — 编辑 Schema：去掉要生成 endpoint 的路径上的 `ignore:`，并按需调整占位符。
3. **第二遍** — 再次读取同一份抓包，将匹配到的请求写入 `paths` 下的 endpoint。

已有 Schema 采用「仅当键不存在时写入」的合并策略：会新增内容，但不会覆盖已有 endpoint 字段。

### HAR 抓包

1. 在浏览器 DevTools 中导出流量（**Network** → **Export HAR**）。
2. 运行第一遍：

   ```bash
   mitmproxy2swagger pass \
     -i capture.har \
     -o schema.yaml \
     -p https://api.example.com/v1
   ```

   `-p` / `--api-prefix` 为 API 基址 URL（不含尾部 `/`）。仅处理 URL 以此前缀开头的请求。

3. 打开 `schema.yaml`，去掉要生成的条目上的 `ignore:`，例如：

   ```yaml
   x-path-templates:
     - ignore:/users/{id}
     - ignore:/users/42
   ```

   列表中靠前的模板在匹配时优先级更高。

4. 使用相同参数运行第二遍：

   ```bash
   mitmproxy2swagger pass \
     -i capture.har \
     -o schema.yaml \
     -p https://api.example.com/v1
   ```

   若要重新生成某个 endpoint，请在第二遍之前从 `paths` 中删除对应 operation。

HAR 文件会自动检测格式。可用 `-f har` 强制指定。

### 增量更新（已有 Enriched Schema）

完成首轮 Pass → Curation → Second Pass → `enrich` 后，若发现**缺少部分接口**或**个别接口注释/结构需修正**，可重新录制**仅覆盖缺失场景**的 HAR，并在同一份 `enriched.yaml` 上增量合并（不必从 `schema.yaml` 重来）。

**合并策略（与 [CONTEXT.md](CONTEXT.md) 中 Schema Merge / Enrichment Merge 一致）：**

- **Pass**：仅当 path / HTTP method 不存在时写入（set-if-not-exists）；**不会**覆盖已有 operation 的 parameters、requestBody、responses。
- **Enrich**（默认，无 `--force`）：仅补全空的语义字段（summary、description、参数说明等）；**不会**覆盖已有中文注释。会对 schema 中所有 endpoint 调用 LLM，但只有空字段会被写入。
- **Enrich `--force`**：重写**全部** operation 的语义字段——增量场景下通常**不要**使用。

**推荐流程：**

1. **备份** — `git commit` 或复制 `enriched.yaml`。
2. **增量 HAR** — 在浏览器中只操作此前未抓到的接口，导出 `incremental.har`（可放在 `testdata/local/`，已 gitignore）。
3. **First Pass** — 向 `enriched.yaml` 追加新的 `x-path-templates`（带 `ignore:`）：

   ```bash
   mitmproxy2swagger pass \
     -i incremental.har \
     -o enriched.yaml \
     -p https://api.example.com/v1
   ```

4. **Curation（手工）** — 仅对**新增**的 `x-path-templates` 条目去掉 `ignore:`；已有条目与顺序尽量不动。一般**不要**对增量场景再跑 `curate --auto`（可能改变模板优先级，影响已生成 endpoint 的匹配）。
5. **Second Pass** — 将新模板物化到 `paths`（同上 `-i` / `-o` / `-p`）。
6. **Enrich** — 为**新** endpoint 补语义字段（不加 `--force`）：

   ```bash
   mitmproxy2swagger enrich \
     -i incremental.har \
     -s enriched.yaml \
     -o enriched.yaml \
     -p https://api.example.com/v1
   ```

**个别已有接口修正（可选）：**

| 问题 | 做法 |
|------|------|
| 仅注释/语义不清 | 直接编辑 `enriched.yaml`；或删除该 operation 上的 `summary` / `description` / 参数 `description` 后，用含该请求的 HAR 再跑 `enrich`（无 `--force`） |
| 结构不对（参数、body、responses） | 从 `paths` 删除该 method（或整段 path），确认 `x-path-templates` 中对应模板无 `ignore:`，再跑 Second Pass + `enrich` |

仓库示例可用 Makefile：`make example-incremental-pass1` → 手工 Curation → `make example-incremental-pass2` → `make example-incremental-enrich`（需 `testdata/local/incremental.har` 与已有 `build/example/enriched.yaml`）。

### Redoc 分组（`tags apply`）

`enrich` 生成的 `tags` 由 LLM 推断，Redoc 侧边栏可能过碎或重复。维护侧车 `tags.yaml`（路径前缀默认 + `METHOD /path` 覆盖），在 **enrich 之后、Redoc 之前** 运行：

```bash
mitmproxy2swagger tags apply \
  -s build/example/enriched.yaml \
  -t build/example/tags.yaml
```

- 默认**替换**每个 operation 的 `tags` 为侧车中的**单个**主标签；`--merge` 时优先采用侧车标签并与已有 tag 去重合并。
- 可定义顶层 `tags:` 与 Redoc `x-tagGroups`（见 `build/example/tags.yaml`，与 `enriched.yaml` 同目录）。
- 未匹配到规则的 operation 会打印警告；`--strict` 时失败退出。

示例：`make example-tags-apply`，或 `make example-redoc`（会自动先执行 tags apply）。

## CLI 参数（`pass`）

| 参数 | 短选项 | 默认值 | 说明 |
|------|--------|--------|------|
| `--input` | `-i` | *必填* | 输入 mitmproxy dump 或 HAR 文件 |
| `--output` | `-o` | *必填* | 输出 OpenAPI Schema YAML |
| `--api-prefix` | `-p` | *必填* | API 基址 URL 前缀 |
| `--format` | `-f` | 自动检测 | 覆盖输入格式（`har` 或 `flow`） |
| `--param-regex` | `-r` | `[0-9]+` | 路径段参数化所用的正则 |
| `--examples` | `-e` | 关闭 | 包含请求/响应示例（可能泄露敏感信息） |
| `--headers` | `-hd` | 关闭 | 在 Schema 中包含 headers（可能泄露敏感信息） |
| `--suppress-params` | `-s` | 关闭 | 不写原始字面量路径，仅保留参数化模板 |

## 已知限制

- **Flow dump**（`--format flow`，mitmproxy 导出）**尚未实现**。请暂时使用 HAR 抓包。
- HAR 输入下，API 前缀匹配仅做 URL 严格前缀匹配（不会用 Host 头回退）。

## 作为库使用

可被其他 Go 模块引用的公共包位于 `pkg/`：

```go
import (
    "github.com/morefun2602/mitmproxy2swagger-go/pkg/pass"
    "github.com/morefun2602/mitmproxy2swagger-go/pkg/enrich"
)

err := pass.Run(pass.Options{
    Input:     "capture.har",
    Output:    "schema.yaml",
    APIPrefix: "https://api.example.com/v1",
    Format:    "har",
})
```

| 包 | 主要导出 |
|----|----------|
| `pkg/pass` | `Run`、`Options` |
| `pkg/curate` | `Run`、`Options`、`AutoTemplates`、`LoadSuggestionsFile` |
| `pkg/auth` | `RunObserve`、`Options`、`LoadObservationsFile` |
| `pkg/enrich` | `Run`、`Options`、`EnrichmentResult`、`RedactMode` |
| `pkg/tags` | `RunApply`、`ApplyOptions`、`LoadTagsFile` |
| `pkg/capture` | `Reader`、`CapturedRequest`、`ProgressFunc` |
| `pkg/capture/open` | `OpenReader` |
| `pkg/schema` | `Document`、`Load`、`Save` |
| `pkg/swaggerutil` | 路径/参数推断辅助函数 |

`internal/golden` 与 `cmd/*` 仅本模块内使用（回归测试与二进制）。详见 [ADR-0006](docs/adr/0006-public-api-in-pkg.md)。

## 开发

```bash
go test ./...
go run ./cmd/generategolden -verify
```

- [CONTEXT.md](CONTEXT.md) — 领域术语与工作流概念
- [docs/adr/](docs/adr/) — 架构决策记录

## 许可证

[MIT](LICENSE)
