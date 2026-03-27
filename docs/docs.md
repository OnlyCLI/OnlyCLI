---
layout: page
title: "Documentation"
description: "Complete documentation for OnlyCLI. Generate native CLIs from OpenAPI specs with auto-pagination, streaming, GJSON transforms, and 35x fewer tokens than MCP."
permalink: /docs/
---

# OnlyCLI documentation

**OnlyCLI** turns an **OpenAPI 3.x** document into a **native Go CLI** (Cobra + standard library HTTP). You run a generator once, commit the project or rebuild in CI, and ship a **single binary** per API—no runtime interpreter, no MCP server.

This page covers installation, `onlycli generate`, and everything the **generated CLI** supports: output formats, transforms, pagination, bodies, auth, config, retries, templates, and more.

---

## Installation

### Go install (recommended for Go developers)

```bash
go install github.com/onlycli/onlycli/cmd/onlycli@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`, then:

```bash
onlycli --help
```

### Binary download (install script)

Fetches the latest release from GitHub (review any script before piping to your shell):

```bash
curl -sSfL https://raw.githubusercontent.com/onlycli/onlycli/main/install.sh | sh
```

The script installs `onlycli` to `INSTALL_DIR` (default `/usr/local/bin`).

### Homebrew

The upstream project ships **tarballs and checksums** via **[GoReleaser](https://goreleaser.com/)** and GitHub Releases. A **Homebrew tap** is not bundled by default; you can publish one with GoReleaser’s `brews` configuration or maintain a private tap that points at release archives.

### Docker

```bash
docker run --rm -v "$(pwd)":/work ghcr.io/onlycli/onlycli generate \
  --spec /work/openapi.yaml \
  --name myapi \
  --out /work/myapi-cli
```

Pin an image tag in CI for reproducible builds.

**Requirements:** [Go](https://go.dev/dl/) **1.22+** on the machine where you **build** the generated CLI (the generator host needs Go; end users of your binary do not need Go).

---

## Quick start

```bash
# 1) Generate a project from a spec (URL or file path)
onlycli generate \
  --spec https://petstore3.swagger.io/api/v3/openapi.json \
  --name petstore \
  --out ./petstore-cli

# 2) Build the CLI
cd petstore-cli
go mod tidy
go build -o petstore .

# 3) Call an operation (Petstore example)
./petstore pets list --format table
```

For a private or corporate API, replace `--spec` with `./openapi.yaml` and set auth as described below.

---

## CLI reference: `onlycli generate`

| Flag | Required | Description |
|------|----------|-------------|
| `--spec` | Yes | Path or URL to OpenAPI **3.x** YAML/JSON. |
| `--name` | Yes | CLI/binary name (also used in default module path and env var prefixes). |
| `--out` | Yes | Output directory for the generated Go module. |
| `--auth` | No | Force auth style: `bearer`, `basic`, `apikey`. If omitted, OnlyCLI infers from the spec when possible. |
| `--module` | No | Go module path (default: `github.com/<name>/<name>-cli`). |

Example with explicit auth and module:

```bash
onlycli generate \
  --spec ./api/openapi.yaml \
  --name acme \
  --auth bearer \
  --module github.com/acme/acme-cli \
  --out ./acme-cli
```

After generation:

```bash
cd acme-cli && go mod tidy && go build -o acme .
```

---

## Generated CLI features

Global flags apply to **every** operation command (inherited from the root command):

| Flag | Default | Purpose |
|------|---------|---------|
| `--format` | `json` | Response shape (see below). |
| `--transform` | (empty) | [GJSON](https://github.com/tidwall/gjson) path applied to JSON before formatting. |
| `--template` | (empty) | [Go `text/template`](https://pkg.go.dev/text/template) executed on decoded JSON. |
| `--profile` | (empty) | Select a named configuration profile. |
| `--page-limit` | `0` | Max pages for **GET** requests (`0` = single response). Auto-detects Link header, cursor, offset, and page-number schemes. |
| `--stream` | `false` | Stream response line-by-line (SSE / NDJSON). |
| `--dry-run` | `false` | Print method, URL, headers, and body preview; do not send HTTP. |
| `--max-retries` | `0` | Retry count for **429** and **5xx** (`0` = no retries). Uses exponential backoff and honors `Retry-After`. |
| `--verbose` | `false` | Log request/response details to stderr. |

---

### Output formats (`--format`)

The body is interpreted as JSON unless you use `raw`. Supported values:

| Value | When to use | Example |
|-------|--------------|---------|
| `json` | Compact JSON (default, pipes well) | `./acme system get-health --format json` |
| `pretty` | Indented JSON for humans | `./acme system get-health --format pretty` |
| `yaml` | JSON decoded and re-encoded as YAML | `./acme system get-health --format yaml` |
| `jsonl` | JSON **array** split into one JSON object per line | `./acme items list --format jsonl` |
| `table` | Tab-separated table when the top-level JSON is an **array of objects** | `./acme items list --format table` |
| `csv` | CSV when the top-level JSON is an **array of objects** | `./acme items list --format csv` |
| `raw` | Unmodified response bytes (skips JSON transform pipeline) | `./acme export blob --format raw` |

`table` and `csv` fall back to compact JSON if the payload is not an array of objects.

```bash
# Human-readable JSON
./petstore pets list --format pretty

# Spreadsheet-friendly
./petstore pets list --format csv > pets.csv
```

---

### Data transformation (`--transform` + GJSON)

`--transform` uses **GJSON** syntax to extract or reshape JSON **before** `--format` or `--template` run. It does not apply to `raw` format.

```bash
# Array of pets → only the "name" field from each element
./petstore pets list --transform '#.name'

# First element’s id
./petstore pets list --transform '0.id'

# Nested path
./github users get-authenticated --transform 'login'
```

See the [GJSON path syntax](https://github.com/tidwall/gjson#path-syntax) for modifiers (`@flatten`, `@join`, etc.).

---

### Pagination (`--page-limit`)

Setting `--page-limit` > `0` on **GET** requests enables auto-pagination. OnlyCLI detects the pagination scheme automatically from each response:

| Scheme | Detection |
|--------|-----------|
| **Link header** | `Link: <url>; rel="next"` header |
| **Cursor (body)** | Response fields like `next_cursor`, `next_page_token`, `end_cursor` in top-level or nested objects (`pagination`, `meta`, `page_info`, etc.) |
| **Cursor URL (body)** | Response fields like `next`, `next_url`, `next_page` containing an absolute URL |
| **Page number** | Increments `page` query parameter; stops on empty results |
| **Offset** | Similar to page number but uses `offset` parameter |

Responses are **merged**: JSON arrays are concatenated; wrapped responses (e.g. `{"data":[...]}`) are unwrapped and merged.

```bash
# Link header pagination (e.g. GitHub)
./github repos list-for-user --username octocat --page-limit 5 --format json

# Cursor-based APIs (auto-detected from response body)
./slack conversations-list --page-limit 10

# Page-number APIs
./myapi items list --page 1 --page-limit 20
```

---

### Streaming (`--stream`)

The `--stream` flag reads the response body line-by-line instead of buffering it, suitable for **Server-Sent Events** (SSE) and **NDJSON** streams.

- **SSE** (`text/event-stream`): parses `data:` lines, skips comments and control fields, stops on `[DONE]`.
- **NDJSON** (`application/x-ndjson`): each line is treated as a standalone JSON object.

Each line is formatted according to `--format` (default: compact JSON).

```bash
# SSE streaming endpoint
./myapi events stream --stream --format pretty

# NDJSON log tail
./myapi logs tail --stream --format json
```

---

### Body input (`--data`, `@file`, body field flags, nested args)

Operations with a JSON request body get:

1. **`--data`** — inline JSON, **`@path`** to a file, or **`@-`** for stdin.  
2. **Body field flags** — one flag per simple schema property (e.g. `--title`, `--body`).  
3. **Nested properties** — dotted flag names mapping to JSON paths (e.g. `metadata.labels` → `--metadata.labels`).

```bash
# Inline JSON
./github issues create --owner org --repo web --data '{"title":"bug","body":"steps"}'

# From file
./github issues create --owner org --repo web --data @issue.json

# From stdin
cat issue.json | ./github issues create --owner org --repo web --data @-

# Body field flags (when generated for the operation)
./github issues create --owner org --repo web --title "bug" --body "steps to reproduce"
```

If `--data` is a JSON **object**, its keys form the base body and any **non-empty body field flags** are **merged on top** (flags override or add nested paths). If `--data` is not an object, field flags cannot be merged with it (you will get an error if both are used in conflicting ways).

---

### Authentication (bearer, basic, apikey, OAuth2)

The generator wires auth from your OpenAPI **security schemes**:

| Style | Typical setup |
|-------|----------------|
| **Bearer** | Token in env (e.g. `GITHUB_TOKEN`) or `config set token ...` |
| **Basic** | Base64 or raw credentials per generated client conventions |
| **API key** | Token in env or config, sent as configured in the spec |
| **OAuth2** | If the spec defines OAuth2, `auth login` supports **device code** flow (`--client-id` only) or **client credentials** (`--client-id` and `--client-secret`) |

```bash
# Bearer via environment (name depends on --name / spec)
export GITHUB_TOKEN=ghp_xxxxxxxxxxxx

# Or persist for the default profile
./github config set token ghp_xxxxxxxxxxxx

# OAuth2 device flow (when generated for your spec)
./myapi auth login --client-id "$CLIENT_ID"
```

Use `./YOUR_CLI auth status` and `./YOUR_CLI auth logout` to inspect or clear stored tokens.

---

### Configuration (`config` subcommands, profiles)

```bash
# Keys: token, base_url, auth_type (exact set may vary slightly by project)
./github config set token ghp_xxx
./github config set base_url https://api.github.com

# Scoped to a profile
./github config set staging.token ghp_xxx --profile staging
# or
./github config set staging.token ghp_xxx

./github config get token
./github config list
./github config use-profile staging

# Any command
./github repos list-for-authenticated-user --profile staging
```

---

### Retry (`--max-retries`)

Retries on **429** and **5xx** responses. Waits using **`Retry-After`** when present, otherwise exponential backoff (capped).

```bash
./github search repos --q "language:go" --max-retries 5
```

---

### Dry-run (`--dry-run`)

Prints the HTTP method, full URL, headers, and a body preview (up to 2048 bytes). Exits without sending the request (`StatusCode` 0, empty body).

```bash
./github repos get --owner microsoft --repo vscode --dry-run
```

---

### Template output (`--template`)

When `--template` is non-empty, the (possibly transformed) JSON is unmarshaled into `interface{}` and executed with Go’s `text/template`. Built-in functions: `join`, `upper`, `lower`.

```bash
# One line per repo in a search result array
./github search repos --q "language:go stars:>100" --format json \
  --template '{{range .}}{{.full_name}} stars={{.stargazers_count}}{{"\n"}}{{end}}'

# Single object summary
./github users get-by-username --username octocat \
  --template '{{.login}} has {{.public_repos}} public repos'
```

`--template` runs **after** `--transform`; use an empty `--transform` to template the full JSON.

---

### Shell completions

When the OpenAPI schema lists **`enum`** values for a parameter, the generator registers **Cobra shell completion** for that flag (so shells with Cobra completion enabled suggest valid values).

Because templates differ by release, check `./YOUR_CLI --help` for any top-level **`completion`** command. If you need scripted completions, follow the [Cobra shell completion guide](https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md) and add a `completion` subcommand to your generated project if it is not already present.

---

## Advanced usage

### Regenerate when the spec changes

In CI, trigger on `openapi.yaml` changes:

```yaml
# .github/workflows/cli.yml
on:
  push:
    paths: ['openapi.yaml']
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
      - run: go install github.com/onlycli/onlycli/cmd/onlycli@latest
      - run: onlycli generate --spec openapi.yaml --name myapi --out ./cli
      - run: cd cli && go mod tidy && go build -o myapi .
```

### LLM agents

Generated CLIs expose rich **`--help`** trees and stable flags—enough for agents to discover operations on demand. Many teams add a short **`SKILL.md`** (or similar) beside the binary summarizing top workflows; that file is **not** emitted by the current generator, so maintain it in your repo or docs pipeline if you want a single “agent onboarding” artifact.

### Base URL overrides

Override the server URL from the spec with **`YOURCLI_BASE_URL`** (prefix depends on generated env naming) or `config set base_url`.

### Verbose debugging

```bash
./github repos get --owner foo --repo bar --verbose
```

Logs request and response lines to **stderr**; JSON output remains on **stdout** for piping.

---

## FAQ

### Which OpenAPI versions work?

The generator builds a **v3** model from your document. Use **OpenAPI 3.0 / 3.1** YAML or JSON. Convert older Swagger 2.0 specs with tooling such as `swagger2openapi` before generating.

### Where is config stored?

Generated CLIs use a small JSON config file under the user config directory (see `runtime/config.go` in your generated tree for the exact path). Use `config list` to inspect.

### Does the CLI support SOAP or gRPC?

No. Only **HTTP + JSON** (REST-style) as described in OpenAPI is in scope.

### How do I add a Homebrew formula?

Package the built binary or tarball from GitHub Releases. GoReleaser can publish to a Homebrew tap with a few lines of YAML; that is optional and maintained separately from OnlyCLI itself.

### What about MCP?

OnlyCLI intentionally avoids MCP for calling your API: you ship a **binary** (plus optional agent notes such as `SKILL.md`) instead of a long-lived tool server. See the [blog]({{ '/blog/' | relative_url }}) for discussion.

### How is pagination detected?

OnlyCLI tries multiple schemes in priority order: **Link header** (`rel="next"`), **body-based next URL** (fields like `next`, `next_url`), **body-based cursor** (fields like `next_cursor`, `next_page_token` in top-level or nested `pagination`/`meta` objects), and **page-number increment** (auto-increments `page` param, stops on empty results). The first matching scheme is used for each page.

### Is Windows supported?

Yes. Generated Go code targets **Linux, macOS, and Windows** (amd64/arm64). Use PowerShell-friendly quoting for JSON bodies.

---

## Related links

- [OnlyCLI on GitHub](https://github.com/onlycli/onlycli)
- [Getting started (blog)]({{ '/blog/getting-started-with-onlycli/' | relative_url }})
- [OnlyCLI vs Stainless]({{ '/compare/vs-stainless/' | relative_url }})
- [OnlyCLI vs Restish]({{ '/compare/vs-restish/' | relative_url }})
