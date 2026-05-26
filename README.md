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
