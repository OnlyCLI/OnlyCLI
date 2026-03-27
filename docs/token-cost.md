---
layout: page
title: "Token Cost: CLI vs MCP"
description: "See how much your AI agents spend on MCP tool schemas. OnlyCLI-generated CLIs use 96-99% fewer tokens."
permalink: /token-cost/
---

<style>
.cost-hero {
  text-align: center;
  padding: 2.5rem 1rem 2rem;
}
.cost-hero h1 {
  font-size: 2rem;
  margin-bottom: 0.5rem;
}
.cost-hero .big-number {
  font-size: 4rem;
  font-weight: 800;
  color: #3fb950;
  line-height: 1.1;
}
.cost-hero .big-label {
  font-size: 1.1rem;
  color: #8b949e;
  margin-top: 0.25rem;
}
.cost-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 1.5rem;
  max-width: 800px;
  margin: 2rem auto;
}
@media (max-width: 640px) {
  .cost-grid { grid-template-columns: 1fr; }
}
.cost-card {
  background: #161b22;
  border: 1px solid #30363d;
  border-radius: 10px;
  padding: 1.5rem;
}
.cost-card h3 {
  margin-top: 0;
  font-size: 1rem;
  color: #8b949e;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.cost-card .number {
  font-size: 2.5rem;
  font-weight: 700;
  line-height: 1.2;
}
.cost-card .unit {
  font-size: 0.9rem;
  color: #8b949e;
}
.cost-card.mcp .number { color: #f85149; }
.cost-card.cli .number { color: #3fb950; }
.cost-table {
  width: 100%;
  max-width: 800px;
  margin: 2rem auto;
  border-collapse: collapse;
}
.cost-table th, .cost-table td {
  padding: 0.75rem 1rem;
  text-align: left;
  border-bottom: 1px solid #30363d;
}
.cost-table th {
  color: #8b949e;
  font-size: 0.85rem;
  text-transform: uppercase;
}
.cost-table .highlight { color: #3fb950; font-weight: 600; }
.cost-table .warn { color: #f85149; font-weight: 600; }
.section-title {
  text-align: center;
  margin: 3rem 0 1.5rem;
}
.cost-cta {
  text-align: center;
  margin: 3rem 0 2rem;
}
.cost-cta a {
  display: inline-block;
  padding: 0.75rem 2rem;
  background: #238636;
  color: #fff;
  border-radius: 6px;
  text-decoration: none;
  font-weight: 600;
}
.cost-cta a:hover {
  background: #2ea043;
}
</style>

<div class="cost-hero">
  <h1>Stop paying the MCP token tax</h1>
  <div class="big-number">35x</div>
  <div class="big-label">fewer tokens per task with OnlyCLI vs MCP</div>
</div>

<div class="cost-grid">
  <div class="cost-card mcp">
    <h3>MCP (GitHub, 93 tools)</h3>
    <div class="number">55,000</div>
    <div class="unit">tokens loaded on every turn</div>
  </div>
  <div class="cost-card cli">
    <h3>OnlyCLI (GitHub, 1,107 commands)</h3>
    <div class="number">~200</div>
    <div class="unit">tokens for --help discovery</div>
  </div>
  <div class="cost-card mcp">
    <h3>3 MCP services connected</h3>
    <div class="number">72%</div>
    <div class="unit">of 200K context window gone on idle</div>
  </div>
  <div class="cost-card cli">
    <h3>SKILL.md agent summary</h3>
    <div class="number">~400</div>
    <div class="unit">tokens for full command discovery</div>
  </div>
</div>

---

<h2 class="section-title">Real task: "What languages does octocat/Hello-World use?"</h2>

<table class="cost-table">
  <thead>
    <tr>
      <th>Approach</th>
      <th>Tokens</th>
      <th>Cost (Claude Sonnet)</th>
      <th>Ratio</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>MCP (GitHub server)</td>
      <td class="warn">44,026</td>
      <td class="warn">$0.132</td>
      <td>---</td>
    </tr>
    <tr>
      <td>OnlyCLI</td>
      <td class="highlight">1,365</td>
      <td class="highlight">$0.004</td>
      <td class="highlight">32x cheaper</td>
    </tr>
  </tbody>
</table>

---

<h2 class="section-title">Cost at scale</h2>

<table class="cost-table">
  <thead>
    <tr>
      <th>Daily requests</th>
      <th>MCP overhead / month</th>
      <th>CLI overhead / month</th>
      <th>Monthly savings</th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td>100</td>
      <td class="warn">$510</td>
      <td class="highlight">~$0</td>
      <td class="highlight">>99%</td>
    </tr>
    <tr>
      <td>1,000</td>
      <td class="warn">$5,100</td>
      <td class="highlight">~$12</td>
      <td class="highlight">99.8%</td>
    </tr>
    <tr>
      <td>10,000</td>
      <td class="warn">$51,000</td>
      <td class="highlight">~$120</td>
      <td class="highlight">99.8%</td>
    </tr>
  </tbody>
</table>

<p style="text-align:center; color:#8b949e; font-size:0.9rem;">
  MCP overhead: schema injection on every completion request (default behavior).<br>
  CLI overhead: occasional <code>--help</code> calls at conversation start.
</p>

---

<h2 class="section-title">Why is MCP so expensive?</h2>

Each MCP tool definition costs **550--1,400 tokens**. When an agent connects to an MCP server, the host injects **every** tool's JSON Schema into the system prompt---whether the model uses zero tools or ten. There is no standard lazy-loading mechanism across providers.

With a CLI, the "schema" is the `--help` text: ~80 tokens for one subcommand, read on demand. The agent pulls what it needs instead of carrying everything.

| | MCP | CLI (OnlyCLI) |
|---|---|---|
| Discovery model | Push all schemas every turn | Pull `--help` on demand |
| Per-tool cost | 550--1,400 tokens | 80--150 tokens (only when read) |
| Idle cost | Full catalog in every request | Zero |
| Scaling | Linear with tool count | Constant (only read what you need) |

---

<h2 class="section-title">The ecosystem agrees</h2>

Multiple independent projects have measured the same gap:

- **mcp2cli**: 96--99% token savings via lazy CLI discovery
- **CLIHub**: 92--98% savings converting MCP servers to CLIs
- **Anthropic Tool Search**: ~85% reduction (Anthropic-only, still loads full schemas on use)
- **Vensas benchmark**: MCP 4--32x more expensive per task

OnlyCLI goes further: instead of wrapping MCP at runtime, it generates a **native, compiled CLI** from your OpenAPI spec. No runtime dependency. No MCP server. Just a binary.

<div class="cost-cta">
  <a href="{{ '/docs/' | relative_url }}">Get started with OnlyCLI</a>
</div>
