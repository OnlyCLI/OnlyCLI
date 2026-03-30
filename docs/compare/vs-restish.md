---
layout: page
title: "OnlyCLI vs Restish - CLI Generation vs Runtime API Client"
description: "Compare OnlyCLI's compile-time CLI generation with Restish's runtime API discovery. Different approaches to API CLIs."
permalink: /compare/vs-restish/
---

# OnlyCLI vs Restish: generated CLI vs universal runtime client

**OnlyCLI** and **[Restish](https://github.com/rest-sh/restish)** (the MIT-licensed HTTP CLI from the [rest-sh](https://github.com/rest-sh) project) both make REST APIs easier from the terminal—but they use **opposite architectures**.

---

## The key difference: compile time vs runtime

| | OnlyCLI | Restish |
|---|---------|---------|
| **Model** | **Static code generation** from OpenAPI: each operation becomes a real subcommand in **your** binary. | **Single generic binary**: discovers shape from **OpenAPI at runtime** (or similar), then dispatches requests. |
| **When work happens** | **Build time** — parse spec once, emit Go, compile. | **Run time** — fetch/parse spec (or profile) when you invoke the tool. |

OnlyCLI gives you a **dedicated executable** whose `--help` tree matches **one** API contract. Restish gives you **one tool** that can talk to **many** APIs after configuration or discovery.

---

## Feature comparison

| Feature | OnlyCLI | Restish |
|--------|---------|---------|
| **Architecture** | **Compiled** CLI per API (OpenAPI → Go + Cobra) | **Runtime** discovery / profiles; generic command router |
| **Binary size** | **Small per API** — only the endpoints you generated | **Larger generic** client — one binary for all configured services |
| **Startup time** | **Immediate** — no spec fetch or discovery on cold start | **Discovery / profile load** can add latency on first use |
| **Offline support** | **Full** for calling the API — spec already compiled in (network still needed for HTTP) | **Limited** if profiles or specs must be fetched remotely |
| **Format support** | `json`, `pretty`, `yaml`, `jsonl`, `raw`, `table`, `csv` + `--template` | JSON-focused output modes; human-friendly defaults |
| **Auth** | Bearer, Basic, API key from spec + env/config; **OAuth2** when defined (`auth login`) | Flexible auth plugins and profile auth |
| **Pagination** | **Auto** — Link header, cursor, page-number, offset (auto-detected via `--page-limit`) | Pagination behavior depends on resource / plugin patterns |
| **Streaming** | `--stream` for SSE and NDJSON | Not a built-in feature |
| **Compression** | `Accept-Encoding: gzip, deflate`; gunzip responses | Typically supports compressed responses in modern builds |
| **GJSON** | **`--transform`** ([gjson](https://github.com/tidwall/gjson)) on JSON before output | Often paired with external `jq`; different native story |
| **Profiles** | **Multi-profile** config (`config set`, `config use-profile`, `--profile`) | Strong **profile** model for many services |
| **Content types** | **JSON-first** (`application/json` bodies and Accept) | **CBOR, MessagePack, Ion**, and JSON (broader content negotiation) |
| **Hypermedia** | **No** HAL / Siren / JSON:API traversal | **HAL, Siren, JSON:API** and related hypermedia features |

---

## When to choose OnlyCLI

- You want a **named binary** for **your** API to hand to customers or internal teams.
- You ship **versioned CLIs** next to **versioned OpenAPI** in CI (regenerate on spec change).
- You need **predictable commands** for **automation and LLM agents** (stable `--help` / flags; optional agent-facing notes you maintain).
- You value **table/CSV/jsonl**, **GJSON transforms**, **`@file` bodies**, **dry-run**, **retries**, and **streaming** (`--stream` for SSE and NDJSON) in a **single-purpose** tool.
- You want the agent economics of a local binary; see the <a href="{{ '/token-cost/' | relative_url }}">CLI vs MCP token-cost benchmark</a> and <a href="{{ '/blog/why-cli-beats-mcp-for-llm-agents/' | relative_url }}">why native CLI beats MCP</a> for the deeper case.

### Example: generate and run against GitHub’s OpenAPI

```bash
onlycli generate \
  --spec https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml \
  --name github \
  --auth bearer \
  --out ./github-cli

cd github-cli && go mod tidy && go build -o github .

export GITHUB_TOKEN=ghp_your_token_here
./github users get-authenticated --format pretty
```

---

## When to choose Restish

- You **hop between many APIs** day to day and want **one** client with profiles.
- You rely on **hypermedia** (HAL, Siren, JSON:API) or **non-JSON** payloads.
- You prefer **no codegen step** — pull a spec URL, explore, and call.
- You want a **REPL-style** workflow for ad-hoc exploration rather than a frozen command tree.

### Example: runtime-oriented workflow (conceptual)

```bash
# Illustrative — exact commands follow Restish docs for your version
restish api add mysvc https://api.example.com/openapi.json
restish mysvc some-operation --param value
```

---

## Performance comparison

These are **qualitative** expectations; measure on your own hardware and networks for benchmarks.

### Cold start time

- **OnlyCLI-generated CLI:** process start is a **native Go binary** with **no spec parse** — typically **milliseconds**.
- **Restish:** may pay for **profile loading**, **plugin init**, or **remote spec retrieval** before the first request, so cold paths can be **noticeably slower** than a tiny generated binary.

### Binary size

- **OnlyCLI:** each binary includes **only** the generated command tree and runtime — often **tens of MB** uncompressed for large APIs, still **one purpose per file**.
- **Restish:** one **general-purpose** binary tends to be **larger** than a minimal generated CLI for a **tiny** spec, but **smaller in total** than maintaining **dozens** of generated CLIs if you truly need universal coverage.

### Memory usage

- **OnlyCLI:** steady-state memory is mostly **HTTP client + response buffers**; no embedded spec parser at runtime.
- **Restish:** may hold **larger runtime structures** (profiles, discovery caches, hypermedia state) depending on configuration.

---

## Summary

| Priority | Better fit |
|----------|------------|
| One binary per API, CI regeneration, agent-friendly commands | **OnlyCLI** |
| One binary for many APIs, hypermedia, alternative content types | **Restish** |

---

## Further reading

- [OnlyCLI documentation]({{ '/docs/' | relative_url }}) — full install and generated CLI reference.
- [CLI vs MCP token-cost benchmark]({{ '/token-cost/' | relative_url }}) — why binaries are lighter for agent workflows.
- [OnlyCLI vs Stainless]({{ '/compare/vs-stainless/' | relative_url }}) — another comparison for API CLI buyers.
- [Restish on GitHub](https://github.com/rest-sh/restish) — source, releases, and issue tracker.
