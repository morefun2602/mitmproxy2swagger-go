# Issue tracker：GitHub

本仓库的 issue 与 PRD 均托管为 GitHub Issues。所有操作使用 `gh` CLI。

## 约定

- **创建 issue**：`gh issue create --title "..." --body "..."`。多行正文请使用 heredoc。
- **读取 issue**：`gh issue view <number> --comments`，配合 `jq` 过滤评论，并获取标签。
- **列出 issue**：`gh issue list --state open --json number,title,body,labels,comments --jq '[.[] | {number, title, body, labels: [.labels[].name], comments: [.comments[].body]}]'`，按需添加 `--label` 和 `--state` 过滤条件。
- **评论 issue**：`gh issue comment <number> --body "..."`
- **添加 / 移除标签**：`gh issue edit <number> --add-label "..."` / `--remove-label "..."`
- **关闭 issue**：`gh issue close <number> --comment "..."`

在 clone 目录内运行 `gh` 时，会自动从 `git remote -v` 推断仓库，无需额外指定。

## 当技能说「发布到 issue tracker」

创建一条 GitHub issue。

## 当技能说「获取相关 ticket」

运行 `gh issue view <number> --comments`。
