## Agent skills

### Issue tracker

Issue 托管在 GitHub（`morefun2602/mitmproxy2swagger-go`）。详见 `docs/agents/issue-tracker.md`。

### Triage labels

使用五个标准 triage 标签，名称与默认一致。详见 `docs/agents/triage-labels.md`。

### Domain docs

单上下文布局：根目录 `CONTEXT.md` + `docs/adr/`。详见 `docs/agents/domain.md`。

## 代码布局

- **`pkg/`** — 可被外部 Go 模块 import 的公共 API：`pass`、`enrich`、`curate`、`auth`、`capture`（含 `har`/`flow`/`open`）、`schema`、`swaggerutil`。
- **`internal/golden`** — 仅本模块：golden 生成/对比、`generategolden` 与 `stripignore` 工具。
- **`cmd/mitmproxy2swagger`** — CLI，import `pkg/pass`、`pkg/curate` 与 `pkg/enrich`。

详见 [ADR-0006](docs/adr/0006-public-api-in-pkg.md)、[ADR-0007](docs/adr/0007-curate-curation-assist.md)、[ADR-0008](docs/adr/0008-auth-observation-two-phase.md)。

## 开发环境

需要 Go **1.26.1** 或更高版本（见 `go.mod`）。使用 `PATH` 中的 `go` 即可：

```bash
go version   # 应 >= go1.26.1
go test ./...
go run ./cmd/mitmproxy2swagger ...
```

若本机安装了多个 Go 版本，请通过 `asdf`、`gvm`、`goenv` 等工具切换，或设置 `GO=/path/to/go` / `make GO=/path/to/go`。**不要**在仓库内写死用户特定的 Go 路径。

若 `GOROOT` 与 `go` 二进制版本不一致，可能导致编译失败；此时修正环境变量或改用匹配的 `go` 即可。
