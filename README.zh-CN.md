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
- `enrich` — LLM 语义增强（见 [ADR-0003](docs/adr/0003-llm-enrichment-subcommand.md)）
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
| `pkg/enrich` | `Run`、`Options`、`EnrichmentResult`、`RedactMode` |
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
