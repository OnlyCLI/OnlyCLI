---
layout: page
title: "OnlyCLI vs Stainless CLI - Feature Comparison"
description: "Compare OnlyCLI and Stainless CLI for generating command-line tools from OpenAPI specs. See which is right for your project."
permalink: /compare/vs-stainless/
---

# OnlyCLI vs Stainless CLI: Which OpenAPI CLI generator fits your stack?

Both **OnlyCLI** and **Stainless** can turn API descriptions into command-line tools. They solve related problems with different philosophies:

- **OnlyCLI** is an open-source, standalone generator. You point it at any OpenAPI 3.x document; it emits a native Go/Cobra project you own, build, and ship—no hosted platform required.
- **Stainless** is a commercial platform oriented around **SDK ecosystems**. Its CLI is typically produced alongside language SDKs (for example Go) from Stainless-managed configuration and workflows, not as a fully isolated “spec in, binary out” loop.

This page compares capabilities so you can decide when each approach makes sense.

---

## How each tool fits in your workflow

### OnlyCLI: spec-first, self-contained

OnlyCLI reads YAML or JSON OpenAPI 3.x from a path or URL, generates a CLI project, and you compile a single binary. The output includes HTTP client code, subcommands aligned with your spec, and standard Go tooling for distribution. For LLM agents, the stable command tree and `--help` text are the primary discovery surface; you can layer a hand-written `SKILL.md` if you want a compact onboarding doc.

### Stainless: platform-first, SDK-centric

Stainless targets teams that want maintained SDKs plus a CLI in one pipeline. Configuration and generation are tied to Stainless’s model (including repository and edition settings). The CLI is described as wrapping generated SDK behavior and ships with platform features such as an interactive explorer.

---

## Feature comparison: OnlyCLI vs Stainless CLI

| Feature | OnlyCLI | Stainless CLI |
|--------|---------|----------------|
| **License** | MIT (open source) | Commercial platform / subscription |
| **Dependencies** | Standalone generator; generated CLI is plain Go + modules | CLI generation is tied to the **Go SDK** target and Stainless config |
| **Input** | Any **OpenAPI 3.x** (file or URL) | **Stainless configuration** + workflow (not “drop any spec” in isolation) |
| **Output formats** | `json`, `pretty`, `yaml`, `table`, `csv`, `jsonl`, `raw` | Typically `json`, `yaml`, `raw`, plus **interactive explore** |
| **GJSON transform** | Yes (`--transform`, [tidwall/gjson](https://github.com/tidwall/gjson) paths) | Not a first-class parallel to OnlyCLI’s pipe-friendly GJSON |
| **Auto-pagination** | Yes for **GET** with `Link: ...; rel="next"` (merged JSON arrays when pages are arrays) | Yes (documented as automatic pagination) |
| **OAuth2** | When the spec defines OAuth2: `auth login` with device flow or client credentials (`--client-id`, optional `--client-secret`) | Supported via platform/SDK auth patterns |
| **`@file` body input** | Yes: `--data @path.json`, `--data @-` for stdin | SDK/CLI patterns vary; not the same flag surface as OnlyCLI |
| **Nested request args** | Yes: body fields map to dotted JSON paths (e.g. `metadata.labels`) | Nested flags / resource-oriented args (e.g. `--name.full-name`) |
| **Multi-profile** | Yes: `config` + `--profile`, `config use-profile` | Profile / environment patterns depend on generated project |
| **Dry-run** | Yes (`--dry-run`: print method, URL, headers, body preview) | Varies by generated CLI |
| **Retry** | Yes (`--max-retries` with backoff + `Retry-After`) | Platform / client may include retries |
| **HTTP/2** | Yes (`ForceAttemptHTTP2` on the transport) | Go-based clients commonly use HTTP/2 where available |
| **Compression** | Requests advertise `gzip, deflate`; responses can be gunzipped | Typically supported in generated stacks |
| **Shell completions** | OpenAPI **enum** flags get Cobra completion callbacks; wire shell scripts via Cobra docs as needed | Bash, zsh, fish, PowerShell (documented) |
| **Interactive TUI** | No dedicated explorer | **Explore**-style interactive mode |
| **SDK generation** | **CLI only** (no multi-language SDK emit) | **SDKs + CLI** as part of the product |
| **Man pages** | Can be produced from Cobra help via your docs pipeline; not bundled by OnlyCLI itself | Documented as part of Stainless-generated CLI |
| **Homebrew / install** | **GitHub releases** + `install.sh`; [GoReleaser](https://goreleaser.com/) builds archives; Homebrew tap is up to your release story | Stainless automates distribution for customers in their workflow |

---

## When to choose OnlyCLI

- You want **MIT-licensed** code and **no vendor lock-in**.
- You already have (or only care about) **standard OpenAPI 3.x**—including public specs from GitHub, vendor bundles, or CI-generated documents.
- You need a **dedicated binary** per API for **developers, CI/CD, or LLM agents** (stable `--help` and flags; optional `SKILL.md` you maintain).
- You care about **CLI ergonomics**: multiple output shapes (`table`, `csv`, `jsonl`), **GJSON** transforms, **`@file` bodies**, and **dry-run** without adopting a full SDK platform.

---

## When to choose Stainless

- You want **official SDKs** (Go, and other targets) **and** a CLI from one managed pipeline.
- You prefer a **hosted / commercial** solution with Stainless handling updates, editions, and repository integration.
- Your team values an **interactive explorer** (TUI-style) for browsing resources and operations.
- You are already standardizing on **Stainless config** rather than raw spec-only workflows.

---

## Code comparison: the same API with both approaches

Below, the goal is identical: a **Petstore-style** `POST /pets` that accepts JSON `{ "id": 1, "name": "doggie", "tag": "nice" }`. The *invocation style* differs because OnlyCLI generates flags from your OpenAPI `requestBody` schema, while Stainless uses its resource-oriented CLI layout tied to the generated SDK.

### OnlyCLI: generate from OpenAPI, then call the binary

```bash
# 1) Generate a CLI project from the spec
onlycli generate \
  --spec https://petstore3.swagger.io/api/v3/openapi.json \
  --name petstore \
  --out ./petstore-cli

cd petstore-cli && go mod tidy && go build -o petstore .

# 2) Call the operation (example matches typical OpenAPI pet schemas)
./petstore pets create --id 1 --name doggie --tag nice --format pretty
```

You can also pass a full JSON body from disk:

```bash
echo '{"id":1,"name":"doggie","tag":"nice"}' > body.json
./petstore pets create --data @body.json
```

### Stainless: configure targets, generate via Stainless, then use the shipped CLI

Stainless does not mirror “single `onlycli generate --spec URL`” one-to-one; you add a **`cli` target** (and typically a **`go` SDK target**) in your Stainless configuration, run Stainless generation in your normal pipeline, then install the produced binary. Conceptually:

```yaml
# stainless.yaml (illustrative — real keys depend on your Stainless edition / docs)
organization: your-org
project: petstore

targets:
  go:
    package_name: petstore
    production_repo: github.com/your-org/petstore-go
  cli:
    binary_name: petstore
    # CLI wraps the Go SDK; resource/method names follow Stainless output
```

After generation and build (per [Stainless CLI documentation](https://www.stainless.com/docs/cli/)):

```bash
# Illustrative invocation shape (exact subcommands come from generated output)
petstore pets create --name.full-name doggie
```

**Takeaway:** OnlyCLI optimizes for **direct OpenAPI → your repo → `go build`**. Stainless optimizes for **Stainless-managed SDK + CLI** with explorer, man pages, and platform integration.

---

## Further reading

- [OnlyCLI documentation]({{ '/docs/' | relative_url }}) — installation, flags, and generated CLI behavior.
- [Stainless CLI docs](https://www.stainless.com/docs/cli/) — official Stainless CLI generator overview.

---

## Summary

| If you need… | Lean toward |
|--------------|-------------|
| Open source, any OpenAPI 3.x, minimal moving parts | **OnlyCLI** |
| Multi-language SDKs + managed generation + interactive explorer | **Stainless** |

Both can produce professional CLIs; the decision is whether your center of gravity is **a standalone spec-driven binary you own** (OnlyCLI) or **a commercial SDK + CLI platform** (Stainless).
