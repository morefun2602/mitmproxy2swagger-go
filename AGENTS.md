## Agent skills

### Issue tracker

Issue 托管在 GitHub（`morefun2602/mitmproxy2swagger-go`）。详见 `docs/agents/issue-tracker.md`。

### Triage labels

使用五个标准 triage 标签，名称与默认一致。详见 `docs/agents/triage-labels.md`。

### Domain docs

单上下文布局：根目录 `CONTEXT.md` + `docs/adr/`。详见 `docs/agents/domain.md`。

## 开发环境

本仓库使用 GVM 安装的 Go 工具链，**不要**使用 Homebrew 或其他路径下的 `go`：

```
/Users/bytedance/.gvm/gos/go1.26.1/bin/go
```

终端中执行 `go build`、`go test`、`go run` 等命令时，应使用该二进制，或先 `gvm use go1.26.1`。若 `GOROOT` 指向其他版本会导致编译失败。
