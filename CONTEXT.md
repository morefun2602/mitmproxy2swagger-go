# mitmproxy2swagger-go

从 HTTP 抓包文件反向生成 OpenAPI 3.0 Schema 的 Go CLI；在 Python v0.14.0 行为对齐已完成的基础上，作为**独立产品**持续演进（如 LLM 语义增强）。

## 语言

**产品演进**：
Python v0.14.0 移植已完成；后续不再以与 Python 版逐字段对齐为首要目标。Pass 流水线与 Golden Schema 仍用于回归，新产品能力（如 Enrichment）在 Go 版独立设计。
_Avoid_: 仍以「忠实移植」作为新功能决策的首要约束

**移植基准（历史）**：
Pass 行为曾以 `alufers/mitmproxy2swagger` **v0.14.0** 为验收对照；该阶段已结束。
_Avoid_: 将 v0.14.0 作为 Enrichment 等新能力的验收标准

**忠实移植（历史阶段）**：
Go 版在 CLI Pass 参数、两遍工作流、输出格式（含 `x-path-templates`）上曾与 Python 版行为等价。
_Avoid_: 用此描述当前产品目标（已 superseded by **产品演进**）

### 输入

**Capture**（抓包文件）：
含若干 HTTP 请求/响应记录的输入文件；工具从中读取流量并生成或更新 Schema。
_Avoid_: 流量文件、dump 文件（作统称时）

**Flow dump**：
mitmproxy 导出的 flow 格式 Capture；对应 CLI `--format flow`。
_Avoid_: mitmproxy 文件、.flow 文件（作概念名时）

**HAR archive**：
浏览器 DevTools 导出的 HAR 格式 Capture；对应 CLI `--format har`。
_Avoid_: HAR 文件（作概念名时）、DevTools 导出

**Captured Request**：
从 Capture 迭代出的单条 HTTP 请求及其响应；是处理流水线中的最小工作单元。
_Avoid_: 一条 capture、flow entry、HAR entry

**Capture Reader**：
读取 Capture 并产出 Captured Request 流的抽象；Flow dump 与 HAR archive 各有一种实现（对应 Python 的 `MitmproxyCaptureReader` / `HarCaptureReader`）。Flow dump 实现为**纯 Go**，不依赖 Python 运行时。
_Avoid_: 解析器、导入器（作 Capture Reader 的同义词时）；通过 subprocess 调用 Python mitmproxy

**Format Detection**（格式检测）：
未指定 `--format` 时，对输入文件做启发式打分（HAR vs Flow dump），选择对应的 Capture Reader；可用 `--format flow|har` 覆盖。
_Avoid_: 类型猜测、扩展名判断（作唯一手段时）

### 工作流

**Pass**（遍）：
对同一 Capture 与同一 Schema 文件执行的一次完整 CLI 运行。
_Avoid_: 轮次、阶段（作 Pass 的同义词时）

**First Pass**（第一遍）：
从 Capture 发现 URL，将候选 Path Template 写入 Schema 的 `x-path-templates`（默认带 `ignore:` 前缀）；尚未生成完整 endpoint 详情。
_Avoid_: 扫描阶段、发现模式

**Second Pass**（第二遍）：
在用户去掉目标 Path Template 上的 `ignore:` 并可能调整占位符后再次运行；匹配的 Captured Request 写入 `paths` 下的 endpoint（parameters、requestBody、responses 等）。
_Avoid_: 生成阶段、填充模式

**Path Template**（路径模板）：
OpenAPI `paths` 的键或 `x-path-templates` 中的条目，如 `/users/{id}`；路径段中的 `{param}` 为占位符。
_Avoid_: URL 模板、路由模式

**`ignore:` 前缀**：
加在 `x-path-templates` 条目前缀，表示该 Path Template 尚未启用；First Pass 默认添加，用户去掉后 Second Pass 才会匹配并生成 endpoint。
_Avoid_: 注释标记、跳过标记

**Curation**（人工筛选）：
First Pass 与 Second Pass 之间，用户对 **Schema** 的最终把关：去掉目标 Path Template 上的 `ignore:`、删除不需要的模板、调整占位符名。工作流中的必要环节，非可选；可由 **Curation Assist** 缩减工作量，但不替代人工确认。
_Avoid_: 手动配置、预处理；将 `curate` 子命令等同于已完成 Curation

**Curation Assist**（Curation 辅助）：
通过 `mitmproxy2swagger curate` 子命令，在 **Curation** 之前对 `x-path-templates` 做确定性合并与排序，并可（后续）产出 **Template Suggestion** 供人工采纳；不写入 `paths`、不生成 **Endpoint**。
_Avoid_: 自动 Curation、AI Pass（作 Curation Assist 的同义词时）

**Template Clustering**（模板聚类）：
**Curation Assist** 中 `--auto` 的保守合并：去掉已被参数化模板覆盖的 **Literal Path**、将 HAR 中仅数字段不同的路径骨架合并为 `{id}` 类占位符，并按 **Template Precedence** 重排；不合并字母 slug 等同形状不同语义段。
_Avoid_: 路径去重（作 Template Clustering 的同义词时）

**Template Suggestion**（模板建议）：
**Curation Assist** 中 `--llm-suggest`（计划能力）产出的合并建议，写入 sidecar YAML（非 **Schema** 主文件）；含 proposed template、被替换条目、样例路径与理由；用户编辑后通过 `--apply-suggestions` 写回 `x-path-templates`。
_Avoid_: 自动合并（作 Template Suggestion 的同义词时）；`suggest:` 前缀条目（与 `ignore:` 语义混用）

**Template Precedence**（模板优先级）：
多个 Path Template 均可匹配同一路径时，合并列表（`paths` 键 + `x-path-templates` 条目）中**位置靠前者优先**；首个 regex 命中即生效（greedy）。
_Avoid_: 最长匹配、最佳匹配

### 输出

**Schema**：
CLI `-o` 输出的 OpenAPI 3.0 YAML 文件；每次 Pass 读取并更新同一文件。
_Avoid_: Swagger 文件、spec 文件（作 Schema 的同义词时）

**Endpoint**：
Schema 中 `paths` 下某 Path Template 与 HTTP method 对应的 operation（含 parameters、requestBody、responses）。
_Avoid_: 路由、API 定义（作 Endpoint 的同义词时）

**API Prefix**：
CLI `-p` 指定的 API 基址 URL（不含尾部 `/`）；仅处理与之匹配的 Captured Request。
_Avoid_: base URL、server URL（作过滤条件时）

**Swagger**：
历史产品名（mitmproxy2swagger）或引用 Python 版原文时使用；不作为 Schema 的同义词。
_Avoid_: 用 Swagger 指代 OpenAPI 3.0 Schema

**CLI Contract**（CLI 契约）：
用户可见的命令行接口：二进制名 `mitmproxy2swagger`。Pass 通过 `pass` 子命令调用；Pass 参数名、短选项与 Python v0.14.0 Pass 一致。Go 扩展能力通过其他子命令暴露（如 `curate`、`enrich`、`version`、`completion`）。无子命令时 CLI 报错并提示可用子命令。
_Avoid_: m2s、mitmproxy2swagger-go（作用户-facing 命令名时）；将 enrich / curate 子命令参数与 Pass 参数混为一谈

**Enrichment**（语义增强）：
在 Pass 产出 **Schema** 之后，结合 **HAR archive** 与 LLM 为 **Endpoint** 补全语义字段（summary、description、tags、参数说明等）的后处理步骤；通过 `mitmproxy2swagger enrich` 子命令触发，非 Pass 的一部分。
_Avoid_: 增强、AI 文档（作 Enrichment 的同义词时）；在 First Pass / Second Pass 内做 LLM 调用

**Enriched Schema**：
经 **Enrichment** 写回后的 OpenAPI 3.0 Schema；主产物仍为 YAML，可交给 Redoc 等工具渲染。
_Avoid_: AI spec、智能文档（作 Enriched Schema 的同义词时）

**Enrichment Merge**（增强合并）：
Enrichment 写入 Enriched Schema 的策略：默认与 **Schema Merge** 相同（set-if-not-exists）；`--force` 时可覆盖已有语义字段。
_Avoid_: 全量重写、覆盖更新（作 Enrichment Merge 的同义词时）

**Redaction**（脱敏）：
任何将 **Captured Request** 送 LLM 的步骤（**Enrichment**、**Template Suggestion**）之前，对敏感 header/body 字段的替换；默认 strict，可通过 `--redact=off` 关闭。
_Avoid_: 匿名化、清洗（作 Redaction 的同义词时）

**Prefix Match**（前缀匹配）：
判断 Captured Request 是否属于当前 API Prefix 的过滤步骤。Flow dump 在 URL 直接不匹配时，会用 Host 相关头替换 netloc 后重试；HAR archive 仅做 URL 的 strict prefix 匹配。
_Avoid_: URL 过滤、域名匹配（作 Prefix Match 的同义词时）

**Schema Merge**（Schema 合并）：
对已有 Schema 文件的增量更新策略：仅当目标键不存在时写入（set-if-not-exists）；不覆盖已有 endpoint 字段。可对不同 Capture 多次 Pass，结果安全合并。
_Avoid_: 全量重建、覆盖更新

**Authentication**（认证）：
调用方证明身份的方式；因 Capture 来源系统不同（浏览器会话、移动端、服务间 Bearer 等），**Authentication** 机制须按每次 **Capture** 推断或配置，不能写死为单一方案。
_Avoid_: 授权、鉴权（作 Authentication 的同义词时）；把某一客户的抓包结论当作全局默认值

**Observed Credential**（观测凭证）：
在某份 **Capture** 中，对 **API Prefix** 下成功请求统计后，高频出现的认证相关 header/cookie（如 `Cookie`、`Authorization`、`X-Api-Key`）；表示「抓包中常见」，不自动等价于 OpenAPI **Security Requirement** 的「必选」。
_Avoid_: Security Scheme（作 Observed Credential 的同义词时）

**Verified Authentication**（已验证认证）：
经手工或探测（如去掉某 header 后请求失败）确认、可写入 **Security Scheme** 的 **Authentication** 方式；优先级高于仅凭 **Observed Credential** 的猜测。
_Avoid_: Observed Credential（作 Verified Authentication 的同义词时）

**Session Cookie**（会话 Cookie）：
一种 **Authentication** 形态：通过 `Cookie` 头传递会话（常含多个 name=value）。部分 Capture（如桌面客户端访问 woa API）上可为 **Verified Authentication**；其它系统可能以 Bearer 为主。
_Avoid_: Bearer Token（作 Session Cookie 的同义词时）

**Security Scheme**（安全方案）：
OpenAPI `components.securitySchemes` 中声明的认证机制类型（如 apiKey in cookie）；描述「调用时需要什么凭证」，不是具体 **Endpoint** 路径。
_Avoid_: 鉴权接口（作 Security Scheme 的同义词时）

**Security Requirement**（安全要求）：
OpenAPI 根级或 operation 上的 `security` 字段，引用一个或多个 **Security Scheme**；未列出或为空数组表示该 operation 不要求 **Authentication**。
_Avoid_: CORS、TLS（作 Security Requirement 的同义词时）

**Auth Endpoint**（鉴权 Endpoint）：
用于建立、刷新或校验会话/凭证的 **Endpoint**（如 `users/token`、`session/self/status`）；属于 **paths** 中的普通 API，与 **Security Scheme** 互补：前者是「鉴权 API」，后者是「其它 API 怎么带凭证」。是否属于 **Auth Endpoint** 由人工确认；**Auth Observation** 可给出 `suggested_auth_paths` 弱提示（path 含 auth/token/session 等），不自动写入 schema。
_Avoid_: Security Scheme（作 Auth Endpoint 的同义词时）；将 permission/precheck 等业务权限接口自动等同 **Auth Endpoint**

**Auth Observation**（鉴权观测）：
通过 `mitmproxy2swagger auth` 子命令，对 **Capture** 在 **API Prefix** 下统计 **Observed Credential**（覆盖率、样例 path、脱敏样例值），写入 sidecar（如 `auth-observations.yaml`）；不自动写入 **Security Scheme**。人工探测后填入 `verified` 与 `combination`（and / or / single），第二阶段 `auth apply` 再生成 `components.securitySchemes` 与 `security`。
_Avoid_: 自动鉴权（作 Auth Observation 的同义词时）；与 **Curation Assist** 混为同一子命令

### 路径推断

**Param Regex**：
CLI `-r` 指定的正则；用于识别路径中应参数化的段（默认 `[0-9]+`）。
_Avoid_: 参数正则表达式（冗长表述）

**Suggested Path Template**（建议路径模板）：
First Pass 根据 Param Regex 将数字段等替换为 `{id}`、`{id1}` 等占位符后生成的 Path Template；默认仍带 `ignore:` 前缀。
_Avoid_: 参数化路径、模板建议

**Literal Path**（字面量路径）：
Capture 中观察到的原始路径（含具体值，如 `/users/42/profile`）；First Pass 默认与 Suggested Path Template 一并写入 `x-path-templates`，除非启用 `--suppress-params`。
_Avoid_: 原始路径、真实路径（作 Literal Path 的同义词时）

### Body 推断

**Payload**：
Captured Request 的请求体或响应体内容（原始或解析后）。
_Avoid_: body 数据、消息体（作 Payload 的同义词时）

**Inferred Schema**：
从 Payload 解析结果经 `value_to_schema` 推导出的 OpenAPI schema；用于 requestBody 与 response content。
_Avoid_: 自动 schema、推导 schema

**Generic Keys**：
对象键全部为数字或全部为 UUID 时的结构特征；Inferred Schema 使用 `additionalProperties` 而非固定 `properties`。
_Avoid_: 动态键、通配键

### 验收

**Golden Schema**：
Pass 在固定 Capture 与 CLI 参数下生成的基准 Schema；用于 Pass 回归，验证与 v0.14.0 行为仍一致。**Enrichment** 不参与 Golden Schema 对比。
_Avoid_: 参考输出、预期 YAML（作 Golden Schema 的同义词时）；用 Golden Schema 验收 Enriched Schema

**Enrichment Acceptance**（增强验收）：
Enrichment 的自动化测试策略：CI 用 mock LLM 固定响应验证 pipeline（Redaction、Enrichment Merge、写回）；可选结构断言；真实 LLM 质量靠 `--emit-prompts` 等人工 spot-check，不做 byte-level golden。
_Avoid_: Golden Schema（作 Enrichment 验收的同义词时）

**Curation Assist Acceptance**（Curation 辅助验收）：
`curate --auto` 用固定 Schema + HAR 做确定性测试（条目数下降、保留字母 slug 字面量、**Template Precedence** 顺序）；**Template Suggestion** 用 mock LLM，不做 byte-level golden。
_Avoid_: Golden Schema（作 Curation Assist 验收的同义词时）

## 示例对话

> **开发者**：我跑完 First Pass，Schema 里 `x-path-templates` 全是 `ignore:/users/42` 这种 Literal Path，正常吗？  
> **领域专家**：正常。First Pass 只做发现，默认带 `ignore:`。你看有没有 Suggested Path Template 如 `ignore:/users/{id}`；Curation 时去掉要生成的路径上的 `ignore:`，再跑 Second Pass 才会写 Endpoint。  
>  
> **开发者**：同一份 Capture，Flow dump 能匹配 `https://api.example.com`，HAR 却匹配不到，是不是 Go 版 bug？  
> **领域专家**：先对照 v0.14.0。Flow dump 的 Prefix Match 会用 Host 头替换 netloc；HAR 只做 strict URL 匹配。这是格式差异，不是回归。用 Golden Schema 对比 Python 输出确认。  
>  
> **开发者**：Second Pass 后想更新某个 Endpoint 的 response schema，再跑一遍行吗？  
> **领域专家**：Schema Merge 是只增不改。已有字段不会被覆盖；要更新就手动删该 Endpoint 再跑 Pass。  
>  
> **开发者**：Pass 生成的 summary 是 `GET infos` 这种机械名字，能自动改成可读描述吗？  
> **领域专家**：Pass 不负责语义命名。Second Pass 与 Curation 完成后，跑 `mitmproxy2swagger pass` 产出 **Schema**，再跑 `mitmproxy2swagger enrich` 做 **Enrichment**，得到 **Enriched Schema**；默认 **Enrichment Merge** 不覆盖你手写的 description，需要时用 `--force`。  
>  
> **开发者**：First Pass 后 `x-path-templates` 有几百条，每个 chat id 一行，能少手工删吗？  
> **领域专家**：那是 **Curation** 的规模问题，不是 **Enrichment**。先跑 `mitmproxy2swagger curate --auto` 做 **Template Clustering**（去掉冗余 **Literal Path**、合并数字段骨架并排好 **Template Precedence**）。字母 slug 等长尾再用 **Template Suggestion**（sidecar + 人工 `--apply-suggestions`）。你仍要在 **Curation** 里去掉要生成的路径上的 `ignore:`，再 **Second Pass**。  
>  
> **开发者**：鉴权怎么进 OpenAPI？  
> **领域专家**：分两层。**Security Scheme** 描述「怎么调」；**Auth Endpoint** 写在 `paths`。机制因 **Capture** 而异：先列 **Observed Credential**，再经探测升为 **Verified Authentication**，最后才写入 `components.securitySchemes` 与 `security`。HAR 里常见但探测非必选的头（如某环境的 `x-pop-token`）只进说明，不进必选 **Security Requirement**。

