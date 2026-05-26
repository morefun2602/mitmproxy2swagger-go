# 领域文档

工程技能在探索代码库时，应如何消费本仓库的领域文档。

## 探索前请先阅读

- 根目录的 **`CONTEXT.md`**，或
- 若存在根目录 **`CONTEXT-MAP.md`** — 它指向各上下文的 `CONTEXT.md`。请阅读与当前主题相关的每一份。
- **`docs/adr/`** — 阅读与你即将修改的区域相关的 ADR。在多上下文仓库中，还需查看 `src/<context>/docs/adr/` 中的上下文专属决策。

若上述文件尚不存在，**静默继续**。不要指出缺失，也不要主动建议提前创建。生产者技能（`/grill-with-docs`）会在术语或决策真正确定后按需创建。

## 文件结构

单上下文仓库（大多数仓库，包括本仓库）：

```
/
├── CONTEXT.md
├── docs/adr/
│   ├── 0001-event-sourced-orders.md
│   └── 0002-postgres-for-write-model.md
└── src/
```

多上下文仓库（根目录存在 `CONTEXT-MAP.md` 时）：

```
/
├── CONTEXT-MAP.md
├── docs/adr/                          ← 系统级决策
└── src/
    ├── ordering/
    │   ├── CONTEXT.md
    │   └── docs/adr/                  ← 上下文专属决策
    └── billing/
        ├── CONTEXT.md
        └── docs/adr/
```

## 使用术语表中的词汇

当你的输出涉及领域概念（issue 标题、重构提案、假设、测试名称等）时，请使用 `CONTEXT.md` 中定义的术语。不要偏离术语表明确避免的同义词。

若你需要的概念尚未出现在术语表中，这是一个信号 — 要么你在发明项目未使用的语言（请重新考虑），要么确实存在缺口（记录它，供 `/grill-with-docs` 补充）。

## 标记 ADR 冲突

若你的输出与现有 ADR 矛盾，请明确说明，而非静默覆盖：

> _与 ADR-0007（event-sourced orders）矛盾 — 但值得重新讨论，因为…_
