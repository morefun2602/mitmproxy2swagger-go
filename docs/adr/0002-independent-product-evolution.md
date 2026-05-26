# 独立产品演进（移植阶段结束）

Python v0.14.0 行为对齐与 Golden Schema 回归已完成；后续不再以与 Python 版逐字段对齐为产品目标。Go 版作为独立 CLI 产品演进（如 LLM Enrichment），Pass 子集仍保持 v0.14.0 兼容以便回归，新能力通过子命令与后处理模块扩展。

**Considered Options**

- **继续以 v0.14.0 为 north star** — 限制 Enrichment 等产品能力。
- **独立产品演进（选用）** — Pass 回归保留；新功能在 Go 版单独设计。

**Consequences**

- 新功能需自带验收策略（Enrichment 不用 Golden Schema）。
- README / CONTEXT 区分 Pass CLI Contract 与 Go 扩展子命令。
