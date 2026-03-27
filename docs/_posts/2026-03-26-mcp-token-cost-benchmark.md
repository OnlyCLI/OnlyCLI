---
layout: post
title: "MCP Token Trap: Why Your AI Agent Burns 35x More Tokens Than a CLI"
description: "Hard numbers on MCP vs CLI token costs. A GitHub MCP server loads 55,000 tokens before any work. OnlyCLI-generated CLIs need 200. Here's the math."
date: 2026-03-26
author: OnlyCLI Team
tags:
  - benchmark
  - mcp
  - token-cost
  - ai-agents
  - llm
---

Every time an LLM agent connects to an MCP server, the full tool catalog is injected into the context window. For a 93-tool GitHub MCP server, that is **55,000 tokens**---before the agent does anything. Connect three services (GitHub, Slack, Sentry) and you have burned **143,000 of your 200,000 token window**. Seventy-two percent gone. On idle.

This post presents the numbers, explains why, and shows how OnlyCLI-generated CLIs bring that cost down by **96--99%**.

## The numbers

These figures come from published benchmarks by independent researchers and our own measurements. All token counts use the Claude Sonnet tokenizer; GPT-4 counts are comparable within 10%.

### MCP: schema injection on every turn

| Metric | Value |
|--------|-------|
| Tokens per MCP tool definition | 550--1,400 |
| GitHub MCP server (93 tools) | ~55,000 tokens |
| 3 services loaded (GitHub + Slack + Sentry) | ~143,000 tokens |
| % of 200K context window consumed on idle | **72%** |
| Cost per request at Claude Sonnet pricing ($3/M input) | ~$0.17 for schema alone |

Every completion request carries this payload, whether the model calls zero tools or ten. At 1,000 requests per day, schema overhead alone costs **$170/day** or **$5,100/month**. At 10,000 requests, **$51,000/month**.

### CLI: on-demand discovery

| Metric | Value |
|--------|-------|
| `--help` output for root command | ~150--250 tokens |
| `--help` for a single subcommand | ~80--150 tokens |
| One command invocation + JSON output | 200--3,000 tokens |
| Agent skill file (`SKILL.md`) summary | ~300--500 tokens |
| Overhead injected on idle turns | **0 tokens** |

An agent using a CLI discovers capabilities by calling `--help` when needed. It does not carry a 55,000-token schema in every request. The discovery cost is paid once per conversation, not once per turn.

### Direct comparison: real task

Task: "What languages does the `octocat/Hello-World` repository use?"

| Approach | Tokens consumed | Cost (Claude Sonnet) |
|----------|---------------:|-----:|
| MCP (GitHub server loaded) | 44,026 | $0.132 |
| CLI (`./github repos get --owner octocat --repo Hello-World --transform languages_url`) | 1,365 | $0.004 |
| **Ratio** | **32x** | **32x** |

The MCP approach pays for all 93 tool definitions plus the JSON-RPC envelope. The CLI approach pays for the command string and the JSON response.

## Why MCP costs so much

MCP was designed for rich, interactive tool surfaces---IDE extensions, chat integrations, multi-step workflows. Its architecture assumes the host needs the full tool catalog available at all times:

1. **Full schema injection**: Every registered tool's JSON Schema is serialized into the system prompt. The model needs it to decide which tool to call.
2. **No lazy loading by default**: Until Anthropic shipped Tool Search (late 2025), there was no standard way to defer schema loading. Even Tool Search still pulls full schemas on demand and is Anthropic-only.
3. **Multiplicative scaling**: Each additional MCP server adds its complete catalog. Five servers with 20 tools each = 100 tool definitions = ~80,000 tokens of schema before any task.

## Why CLI costs so little

A generated CLI is a compiled binary. Its "schema" is the `--help` text, which the agent reads on demand:

```
$ ./github repos get --help
Get a repository

Usage:
  github repos get [flags]

Flags:
      --owner string   Repository owner
      --repo string    Repository name
```

That is ~80 tokens. The agent reads it once, then issues the command. On subsequent turns, it already knows the flags.

### The SKILL.md advantage

OnlyCLI projects can ship a `SKILL.md`---a compact, agent-facing summary of the top operations. A 20-command summary fits in ~400 tokens. The agent reads it at conversation start and has enough context to call any operation without ever loading a full schema.

Compare: 400 tokens (SKILL.md) vs 55,000 tokens (MCP schema). That is **137x fewer tokens** for the discovery step.

## Cost at scale

| Scale | MCP overhead/month | CLI overhead/month | Savings |
|-------|-------------------:|-------------------:|--------:|
| 100 req/day | $510 | ~$0 | >99% |
| 1,000 req/day | $5,100 | ~$12 | 99.8% |
| 10,000 req/day | $51,000 | ~$120 | 99.8% |

"CLI overhead" accounts for occasional `--help` calls during conversation starts. MCP overhead assumes the schema is carried on every completion request (the default behavior).

## The emerging consensus

We are not alone in this analysis. Multiple independent projects have converged on the same conclusion:

- **mcp2cli** (CyberCorsairs): Replaces full schema injection with lazy CLI discovery. Claims 96--99% token savings.
- **CLIHub** (Kagan Yilmaz): Converts MCP servers to CLIs. Measures 92--98% savings.
- **BuildMVPFast analysis**: Documents $81K/month overhead at scale for MCP-heavy architectures.
- **Vensas benchmark**: Finds MCP 4--32x more expensive per task depending on complexity.

The pattern is clear: for stateless REST API access, MCP's always-on schema model is a poor fit.

## When MCP is still worth it

MCP is justified when you need:

- **Stateful sessions**: Database connections, file handles, multi-step transactions.
- **Server push**: Real-time notifications from the tool to the model.
- **Vendor lock-in mitigation**: When the only integration a vendor ships is MCP.
- **Rich IDE integration**: Where the host runtime handles lifecycle and the user benefits from interactive exploration.

For calling `GET /repos/{owner}/{repo}` and reading JSON? That is what CLIs were built for.

## Try it yourself

Generate a CLI from any OpenAPI spec and compare your token costs:

```bash
# Install
go install github.com/onlycli/onlycli/cmd/onlycli@latest

# Generate from GitHub's spec
onlycli generate \
  --spec https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml \
  --name github --auth bearer --out ./github-cli

cd github-cli && go mod tidy && go build -o github .

# 1,107 commands, ~200 tokens to discover
./github --help
```

Every endpoint in the GitHub REST API becomes a typed command with flags, help text, and shell completion---at a fraction of the token cost of MCP.

---

*For a deeper technical comparison, see [Why Native CLI Beats MCP for LLM Agent Tool Use]({{ '/blog/why-cli-beats-mcp-for-llm-agents/' | relative_url }}). To get started, follow the [quick start guide]({{ '/blog/getting-started-with-onlycli/' | relative_url }}).*
