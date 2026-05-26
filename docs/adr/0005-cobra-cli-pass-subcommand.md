# CLI 使用 Cobra 与显式 `pass` 子命令

主二进制 `mitmproxy2swagger` 改用 `github.com/spf13/cobra`（v1.6.1）组织命令树：Pass 为 `pass` 子命令（不再支持无子命令直接跑 Pass）；`enrich` 保持子命令；新增 `version`（`-ldflags` 可注入 Version/Commit）与 `completion`。无子命令时返回短错误并提示 `pass` / `enrich`。Pass / enrich 的 flag 名与短选项 1:1 保留。

**Considered Options**

- **Root `Run` = Pass** — 零 breaking，与旧 Python 用法一致；子命令扩展 awkward。
- **显式 `pass` + Cobra（选用）** — 命令边界清晰；breaking，需更新脚本与文档。
- **`internal/cli` 包** — 对本仓库规模过度。

**Consequences**

- `cmd/mitmproxy2swagger/` 拆分 `root.go`、`pass.go`、`enrich.go`、`version.go`、`completion.go`；库逻辑在 `pkg/pass` / `pkg/enrich`（见 [ADR-0006](./0006-public-api-in-pkg.md)）。
- 更新 **CLI Contract**（CONTEXT.md）、README、Makefile；CLI 行为见 `cmd/mitmproxy2swagger/*_test.go`。
- 详见 [ADR-0003](./0003-llm-enrichment-subcommand.md) 中 enrich invocation 示例。
