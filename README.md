# mitmproxy2swagger

[English](README.md) | [中文](README.zh-CN.md)

Go port of [mitmproxy2swagger](https://github.com/alufers/mitmproxy2swagger).

A CLI tool that reverse-engineers REST APIs from HTTP capture files and produces [OpenAPI 3.0](https://swagger.io/specification/) YAML schemas.

## Installation

Requires Go **1.26.1** or later.

**Install the binary:**

```bash
go install github.com/morefun2602/mitmproxy2swagger-go/cmd/mitmproxy2swagger@latest
```

**Or build from source:**

```bash
git clone https://github.com/morefun2602/mitmproxy2swagger-go.git
cd mitmproxy2swagger-go
go build -o mitmproxy2swagger ./cmd/mitmproxy2swagger
```

## Usage

Subcommands:

- `pass` — run a Pass over a capture file and update the OpenAPI schema
- `curate` — cluster `x-path-templates` (`--auto`), LLM merge suggestions (`--llm-suggest`), apply edited suggestions (`--apply-suggestions`); see [ADR-0007](docs/adr/0007-curate-curation-assist.md)
- `auth observe` / `auth apply` — two-phase authentication: observe credentials into `auth-observations.yaml`, then after curl verification write `components.securitySchemes` and `security`; see [ADR-0008](docs/adr/0008-auth-observation-two-phase.md)
- `enrich` — LLM semantic enrichment (see [ADR-0003](docs/adr/0003-llm-enrichment-subcommand.md))
- `tags apply` — apply `tags.yaml` sidecar for Redoc grouping (see below)
- `version` — print build version
- `completion` — generate shell completion scripts

The Pass workflow runs twice over the same capture file and schema:

1. **First pass** — discovers URL paths and writes candidate path templates to `x-path-templates` (prefixed with `ignore:` by default).
2. **Curation** — edit the schema: remove `ignore:` from the paths you want, and adjust placeholders if needed.
3. **Second pass** — reads the same capture again and writes matching endpoints under `paths`.

Existing schema content is merged with a set-if-not-exists policy: new keys are added, but existing endpoint fields are not overwritten.

### HAR capture

1. Export traffic from browser DevTools (**Network** → **Export HAR**).
2. Run the first pass:

   ```bash
   mitmproxy2swagger pass \
     -i capture.har \
     -o schema.yaml \
     -p https://api.example.com/v1
   ```

   `-p` / `--api-prefix` is the API base URL (no trailing slash). Only requests whose URL starts with this prefix are processed.

3. Open `schema.yaml` and remove `ignore:` from entries you want generated, for example:

   ```yaml
   x-path-templates:
     - ignore:/users/{id}
     - ignore:/users/42
   ```

   Templates closer to the top of the list take precedence when matching.

4. Run the second pass with the same flags:

   ```bash
   mitmproxy2swagger pass \
     -i capture.har \
     -o schema.yaml \
     -p https://api.example.com/v1
   ```

   To regenerate an endpoint from scratch, delete that operation from `paths` before running the second pass.

HAR files are auto-detected. You can force the format with `-f har`.

### Incremental update (existing enriched schema)

After the initial Pass → Curation → second Pass → `enrich`, if endpoints are **missing** or **individual operations** need better docs or structure, record a new HAR that covers only the missing flows and merge into the same `enriched.yaml` (no need to restart from a separate `schema.yaml`).

**Merge policy (see [CONTEXT.md](CONTEXT.md) — Schema Merge / Enrichment Merge):**

- **Pass**: set-if-not-exists for paths and HTTP methods; does **not** overwrite existing parameters, requestBody, or responses.
- **Enrich** (default, no `--force`): fills empty semantic fields only; does **not** overwrite existing summaries or descriptions. Still calls the LLM for every endpoint, but only empty fields are written.
- **Enrich `--force`**: rewrites semantic fields on **all** operations — usually avoid for incremental work.

**Recommended workflow:**

1. **Backup** — commit or copy `enriched.yaml`.
2. **Incremental HAR** — browse only missing APIs; export e.g. `testdata/local/incremental.har`.
3. **First pass** — append new `ignore:` templates to `enriched.yaml`:

   ```bash
   mitmproxy2swagger pass \
     -i incremental.har \
     -o enriched.yaml \
     -p https://api.example.com/v1
   ```

4. **Curation (manual)** — remove `ignore:` only from **new** `x-path-templates` entries; leave existing order alone. Avoid `curate --auto` on incremental runs (can change template precedence).
5. **Second pass** — same `-i` / `-o` / `-p` as step 3.
6. **Enrich** — without `--force`:

   ```bash
   mitmproxy2swagger enrich \
     -i incremental.har \
     -s enriched.yaml \
     -o enriched.yaml \
     -p https://api.example.com/v1
   ```

**Fixing individual operations:**

| Issue | Action |
|-------|--------|
| Docs only | Edit YAML, or clear semantic fields on that operation and re-run `enrich` (no `--force`) with a HAR that includes the request |
| Wrong structure | Delete the method (or path) from `paths`, ensure the template in `x-path-templates` is active (no `ignore:`), then second Pass + `enrich` |

Example Makefile targets: `example-incremental-pass1` → manual curation → `example-incremental-pass2` → `example-incremental-enrich` (requires `testdata/local/incremental.har` and existing `build/example/enriched.yaml`).

### Redoc grouping (`tags apply`)

LLM `enrich` tags are often too granular for Redoc. Maintain a `tags.yaml` sidecar (prefix rules + per-operation overrides), then run **after enrich, before Redoc**:

```bash
mitmproxy2swagger tags apply \
  -s build/example/enriched.yaml \
  -t build/example/tags.yaml
```

Replaces each operation’s `tags` with a single primary tag by default; `--merge` prefers the sidecar tag. Writes top-level `tags:` and `x-tagGroups` when present. See `build/example/tags.yaml` (alongside `enriched.yaml`). Use `make example-tags-apply` or `make example-redoc` (runs tags apply first).

## CLI flags (`pass`)

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--input` | `-i` | *(required)* | Input mitmproxy dump or HAR file |
| `--output` | `-o` | *(required)* | Output OpenAPI schema YAML |
| `--api-prefix` | `-p` | *(required)* | API base URL prefix |
| `--format` | `-f` | auto | Override input format (`har` or `flow`) |
| `--param-regex` | `-r` | `[0-9]+` | Regex for path segments to parameterize |
| `--examples` | `-e` | off | Include request/response examples (may expose sensitive data) |
| `--headers` | `-hd` | off | Include headers in the schema (may expose sensitive data) |
| `--suppress-params` | `-s` | off | Omit literal paths; keep parameterized templates only |

## Limitations

- **Flow dump** (`--format flow`, mitmproxy export) is **not implemented yet**. Use HAR captures for now.
- For HAR input, API prefix matching is strict URL prefix only (no Host-header fallback).

## As a library

Public packages live under `pkg/` and can be imported from other Go modules:

```go
import (
    "github.com/morefun2602/mitmproxy2swagger-go/pkg/pass"
    "github.com/morefun2602/mitmproxy2swagger-go/pkg/enrich"
)

err := pass.Run(pass.Options{
    Input:     "capture.har",
    Output:    "schema.yaml",
    APIPrefix: "https://api.example.com/v1",
    Format:    "har",
})
```

| Package | Main exports |
|---------|----------------|
| `pkg/pass` | `Run`, `Options` |
| `pkg/curate` | `Run`, `Options`, `AutoTemplates`, `LoadSuggestionsFile` |
| `pkg/auth` | `RunObserve`, `RunApply`, `Options`, `ApplyOptions`, `LoadObservationsFile` |
| `pkg/enrich` | `Run`, `Options`, `EnrichmentResult`, `RedactMode` |
| `pkg/tags` | `RunApply`, `ApplyOptions`, `LoadTagsFile` |
| `pkg/capture` | `Reader`, `CapturedRequest`, `ProgressFunc` |
| `pkg/capture/open` | `OpenReader` |
| `pkg/schema` | `Document`, `Load`, `Save` |
| `pkg/swaggerutil` | path/parameter inference helpers |

`internal/golden` and `cmd/*` are module-private (golden tests and binaries only). See [ADR-0006](docs/adr/0006-public-api-in-pkg.md).

## Development

```bash
go test ./...
go run ./cmd/generategolden -verify
```

- [CONTEXT.md](CONTEXT.md) — domain terminology and workflow concepts
- [docs/adr/](docs/adr/) — architecture decision records

## License

[MIT](LICENSE)
