# 纯 Go 读取 Flow dump

Python v0.14.0 通过 `mitmproxy` Python 包的 `FlowReader` 解析 Flow dump。Go 版必须在纯 Go 中实现 Flow dump 的 Capture Reader，不依赖 Python 运行时。验收以 Python v0.14.0 在相同 Capture 与 CLI 参数下生成的 Golden Schema 为准。

## Considered Options

- **纯 Go 解析 Flow dump（选用）** — 符合 Go 移植目标；可用 Python `testdata/` 做 Golden Schema 回归。
- **subprocess 调用 Python mitmproxy** — 仍绑定 Python 环境，非真正的 Go 替代。
- **Go 版暂不支持 Flow dump** — 与忠实移植、CLI 契约及 v0.14.0 基准冲突。

## Consequences

- Flow dump 格式解析是移植的主要技术风险，需优先投入并与 Golden Schema 测试同步推进。
- HAR 读取可独立实现；Flow 与 HAR 共享 Capture Reader 抽象与后续 Pass 流水线。
