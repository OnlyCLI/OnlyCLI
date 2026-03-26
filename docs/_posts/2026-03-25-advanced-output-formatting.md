---
layout: post
title: "Mastering OnlyCLI Output: Table, YAML, GJSON, and Go Templates"
description: "Deep dive into OnlyCLI's output formatting options including table view, YAML, GJSON transforms, and Go template customization."
date: 2026-03-25
author: OnlyCLI Team
tags:
  - tutorial
  - advanced
  - formatting
---

Every OnlyCLI-generated binary prints **structured responses** so scripts and agents can rely on predictable stdout. By default you get **compact JSON**, but the runtime also supports **pretty JSON**, **YAML**, **JSON Lines**, **raw bodies**, **tabular** views, **CSV**, **GJSON transforms**, and **Go templates**. This guide explains each option, when to use it, and how to combine flags without fighting the shell.

## Default JSON Output

With `--format json` (the default), the CLI **compacts** the response to a single line per object when possible. That is ideal for pipelines:

```bash
./github repos get --owner octocat --repo Hello-World | jq '.stargazers_count'
```

Compact JSON minimizes noise in logs and makes `jq` fast paths obvious.

## Pretty-Printed JSON (`--format pretty`)

For human inspection in the terminal, use **indented** JSON:

```bash
./github repos get --owner octocat --repo Hello-World --format pretty
```

Use this in demos and debugging; switch back to `json` or `jsonl` in automation to avoid multiline parsing surprises.

## YAML Output (`--format yaml`)

YAML reads well for configs and diffs. The runtime unmarshals JSON from the API and re-marshals as YAML:

```bash
./petstore pet find-pets-by-status --status available --format yaml
```

(Command names depend on the spec's tags and operation IDs. The examples below use the official Swagger Petstore and GitHub REST API specs.)

YAML is excellent when you paste results into tickets or merge them with existing `.yaml` files. Remember that **YAML is a superset** of JSON structurally here—the wire format from the server is still JSON; conversion happens client-side.

## Table Output (`--format table`)

When the response decodes to a **JSON array of objects**, the CLI can render a **tab-aligned table** with sorted column headers:

```bash
./github repos list-for-user --username octocat --format table
```

If the payload is **not** an array of objects (for example a single object or a primitive), the runtime falls back to compact JSON so you still see data instead of an empty grid.

## CSV for Data Pipelines (`--format csv`)

The same **array-of-objects** shape can become **CSV** for spreadsheets and `awk`:

```bash
./github repos list-for-user --username octocat --per-page 100 --format csv > repos.csv
```

Header row order follows sorted keys from the first object. Nested values are rendered in a cell-safe string form—reach for **`--transform`** when you need flat columns.

## JSONL for Streaming (`--format jsonl`)

**JSON Lines** is one JSON value per line. OnlyCLI’s `jsonl` mode expects the **top-level response to be a JSON array**; each element is written on its own line:

```bash
./github issues list-for-repo --owner microsoft --repo vscode --state open --format jsonl | head -5
```

This pattern plays well with **line-oriented** Unix tools (`while read`, `rg`, log aggregators) and with streaming consumers that dislike parsing a giant array in one shot.

## Raw Bodies (`--format raw`)

When you need the exact bytes from the server (non-JSON, debugging, or piping to a file), use:

```bash
./myapi documents download --id 123 --format raw > out.bin
```

Pair with `--verbose` when you need to inspect status and headers on stderr while keeping stdout clean.

## GJSON Transforms (`--transform`)

**GJSON** is a path language for JSON. The `--transform` flag applies a GJSON expression **before** formatting—so you can slice arrays, pick fields, and filter without installing `jq` (though `jq` remains excellent for complex logic).

### Select a single field

Assume an object with `name` and `stars`:

```bash
./github repos get --owner octocat --repo Hello-World --transform "name"
```

### Map over an array (`#.field`)

For an array of objects, `#` means “each element”:

```bash
./github repos list-for-user --username octocat --transform "#.name"
```

### Filter arrays

GJSON supports predicates inside paths. Example pattern for “objects where `stars` exceeds 100” (syntax mirrors GJSON filter expressions):

```bash
./github search repos --q 'language:go' --transform 'items.#(stargazers_count>100)#'
```

The **`items`** segment reflects GitHub’s **search** JSON shape (results live under an `items` array). For a **bare array** of repository objects, the predicate form looks like `'#(stargazers_count>100)#'` instead. Always inspect **`--format pretty`** once on a real response before locking in a path.

### Multiple fields as a new object

```bash
./github repos get --owner octocat --repo Hello-World --transform "{name,stargazers_count}"
```

This is useful for shrinking payloads before `--format table` or `--format csv`.

## Go Templates (`--template`)

When JSON path syntax is not enough, **`--template`** runs a **Go `text/template`** over the decoded JSON value. The data is passed as generic `interface{}` maps and slices—use dot paths and `range` like standard Go templates.

### Simple field access

```bash
./github repos get --owner octocat --repo Hello-World --template '{{.name}}'
```

### Range over an array

```bash
./github repos list-for-user --username octocat --template '{{range .}}{{.name}}{{"\n"}}{{end}}'
```

### Conditionals and formatting inside templates

You can branch on fields once you know their JSON types (strings, floats, booleans). Because `encoding/json` decodes numbers into **`float64`** when unmarshaling into `interface{}`, numeric comparisons in templates are easy to get wrong: a value you expect to be an integer is still a float at template time. For heavy logic, pipe JSON to **`jq`** instead. A simple string branch:

```bash
./github repos get --owner octocat --repo Hello-World \
  --template '{{if eq .name "Hello-World"}}match{{else}}other{{end}}'
```

If you must compare counts, prefer **`--transform`** to narrow the JSON first, or compare as floats with care (for example `{{if gt .stargazers_count 100.0}}` when that field exists on the object you pass to the template).

Templates execute **after** optional `--transform` in the pipeline implemented by your binary—check `runtime/output.go` in the generated project if you customize the toolchain.

> **Tip:** The `--format` flag’s inline help may list `json`, `pretty`, `yaml`, `jsonl`, and `raw`. Many builds also accept **`table`** and **`csv`** when the response body is a JSON **array of objects**; if a format is ignored, upgrade OnlyCLI and regenerate.

## Combining Transform and Format

Shrink, then render:

```bash
./github repos list-for-user --username octocat \
  --transform "#.{name,description}" \
  --format table
```

Or extract IDs as plain text for xargs:

```bash
./github issues list-for-repo --owner myorg --repo myrepo --state open \
  --transform "#.number" \
  --format json
```

Experiment with `--format pretty` first until the JSON shape is obvious, then tighten with `--transform`.

## Piping to Other Tools

You are not limited to built-in formats. The universal pattern:

```bash
./github search repos --q 'stars:>1000 language:go' --format json | jq '.items[] | .full_name'
```

Because stdout is stable and stderr carries `--verbose` diagnostics, you can **tee** logs without corrupting JSON:

```bash
./github repos list-for-user --username octocat --format json 2>trace.log | jq length
```

## Summary

| Goal | Flag / approach |
|------|------------------|
| Machine default | `--format json` |
| Readable JSON | `--format pretty` |
| Config-style | `--format yaml` |
| Terminal table | `--format table` |
| Spreadsheets | `--format csv` |
| Line streaming | `--format jsonl` |
| Non-JSON body | `--format raw` |
| Slice JSON | `--transform 'gjson expr'` |
| Custom layout | `--template '{{ ... }}'` |

Mastering these switches turns a generated CLI from a thin HTTP wrapper into a **swiss-army formatter** for APIs you touch every day.

---

*If you have not generated a CLI yet, follow [Getting Started with OnlyCLI]({{ '/blog/getting-started-with-onlycli/' | relative_url }}).*
