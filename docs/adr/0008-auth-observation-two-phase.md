# 鉴权观测两阶段（`auth observe` / `auth apply`）

不同 Capture 的 **Authentication** 机制不同（Session Cookie、Bearer、自定义 header 等），且 HAR 中「高频出现」不等于「服务端必选」。采用两阶段：**Auth Observation** 写 sidecar 统计 **Observed Credential**；人工 curl/剔除验证后填 `verified` 与 `combination`；第二阶段 `auth apply` 再写入 `components.securitySchemes` 与 `security`。

**Considered Options**

- **Pass 内自动写 security** — 易把可选头标成必选，且破坏 Golden。
- **仅 `-hd` 写 header parameter** — 非 OpenAPI security 模型，且易泄露凭证。
- **单阶段启发式直接写 schema** — 跨系统误判率高。
- **两阶段 sidecar + `auth` 子命令（选用）** — 与 `curate` / Template Suggestion 模式一致。

**Consequences**

- `mitmproxy2swagger auth observe -i <capture> -p <prefix> -o auth-observations.yaml`。
- `mitmproxy2swagger auth apply -s schema.yaml [--observations auth-observations.yaml]`：读 `verified` + `combination`，写 `components.securitySchemes` 与根级 `security`；对 `suggested_auth_paths` 中已出现在 `paths` 的项打 `tags: [auth]`（`--no-tag-auth-paths` 可关）。
- sidecar 含 `observed`、`verified`（人工）、`combination`（`single` / `and` / `or`）、`suggested_auth_paths`（弱启发式，非权威）。
- `verified` 取值：`cookie`、`bearer`、或 header 名（如 `x-csrftoken`）；`header:<name>` 亦可。
- 不改 Pass 默认输出；Auth Endpoint 与业务接口共用同一 **Security Scheme**（见 grill B1）。
- 术语见 CONTEXT.md：**Auth Observation**、**Observed Credential**、**Verified Authentication**。
