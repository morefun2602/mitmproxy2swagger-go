# 公共 API 迁至 `pkg/`

Go 的 `internal/` 规则禁止模块外 import。为使外部项目可调用 Pass / Enrichment 流水线，将库代码从 `internal/{pass,enrich,capture,schema,swaggerutil}` 迁至 `pkg/` 下同名包；`internal/golden` 保留（仅回归测试与 `cmd/generategolden`、`cmd/stripignore`）。

**Considered Options**

- **保留 `internal/`，仅文档说明 fork 方式** — 外部无法 `import`，不符合「作为库」目标。
- **迁至 `pkg/`（选用）** — 标准 Go 公共 API 布局；import 路径 breaking，本仓库一次性改完。
- **拆分子 module** — 对本仓库规模过度；未改 `go.mod` module path。

**Consequences**

- 外部应使用 `github.com/morefun2602/mitmproxy2swagger-go/pkg/pass` 等路径；勿再引用已删除的 `internal/pass` 等。
- CLI（`cmd/mitmproxy2swagger`）与 `internal/golden` 改为 import `pkg/*`。
- README「As a library」/「作为库使用」、AGENTS.md 代码布局已更新。
- 未承诺稳定：`internal/golden`、`cmd/*`；未新增 `pkg/` 根级聚合 facade。
