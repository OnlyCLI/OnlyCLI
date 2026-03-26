---
layout: post
title: "Getting Started with OnlyCLI: Generate a CLI in 5 Minutes"
description: "Learn how to install OnlyCLI and generate a native CLI from any OpenAPI spec in under 5 minutes."
date: 2026-03-25
author: OnlyCLI Team
tags:
  - tutorial
  - getting-started
---

**OnlyCLI** turns an OpenAPI 3.x document into a **native Go CLI**: real subcommands, flags for every parameter, JSON-first output, and optional extras like table mode and GJSON transforms. This tutorial walks from zero to a working binary using the small **Petstore** spec shipped in the OnlyCLI repository, then points you toward your own APIs.

## Prerequisites

Before you start, you need:

- **Go 1.22 or newer** on your PATH (`go version`). The OnlyCLI module may track a newer toolchain in `go.mod`; if `go install` complains, upgrade Go to match.
- **Network access** to fetch the spec (if you use a URL) and Go modules during `go mod tidy`.
- A terminal and a directory where you are happy to create a new folder (for example `~/cli-lab`).

You do **not** need Docker for this quick path, though it is available if you prefer a containerized generator.

## Installation: Three Ways

Pick one install path. They all produce the same `onlycli` generator binary.

### 1. `go install` (recommended for Go users)

```bash
go install github.com/onlycli/onlycli/cmd/onlycli@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`, then confirm:

```bash
onlycli --help
```

### 2. Binary download (install script)

If you do not want to compile from source, use the project install script (review it first if your policy requires):

```bash
curl -sSfL https://raw.githubusercontent.com/onlycli/onlycli/main/install.sh | sh
```

Re-open your shell or adjust `PATH` if the installer tells you where it placed the binary.

### 3. Docker

Run the generator inside a container and mount your working directory:

```bash
docker run --rm -v "$(pwd)":/work ghcr.io/onlycli/onlycli \
  generate \
  --spec /work/api.yaml \
  --name myapi \
  --out /work/myapi-cli
```

Replace `/work/api.yaml` with the path **inside the container** that corresponds to your host file. The rest of this guide assumes a local `onlycli` on your host for readability.

## Your First CLI: Generate from the Petstore Spec

We will use the official Swagger Petstore spec so your commands match the docs exactly. From any empty project directory:

```bash
mkdir -p ~/cli-lab && cd ~/cli-lab

onlycli generate \
  --spec https://raw.githubusercontent.com/swagger-api/swagger-petstore/master/src/main/resources/openapi.yaml \
  --name petstore \
  --out ./petstore-cli
```

What just happened:

- `--spec` accepts a **local path** or **HTTPS URL** to YAML or JSON OpenAPI 3.x.
- `--name` becomes the root command name (your binary will feel like `./petstore ...`).
- `--out` is a fresh directory; OnlyCLI writes `main.go`, `commands/`, `runtime/`, `go.mod`, and related files.

If generation succeeds, `cd` into the output and sync modules:

```bash
cd petstore-cli
go mod tidy
```

## Build and Run

Compile a single static binary:

```bash
go build -o petstore .
```

Smoke-test the root help:

```bash
./petstore --help
```

You should see a short description and top-level subcommands that mirror your spec’s **tags** (for this file, groups like `pets` and `store`).

### Call a read-only endpoint

The sample spec defines **list pets** under the `pets` tag. List pets filtered by status:

```bash
./petstore pets list --status available
```

Because `https://petstore.example.com` is a placeholder host in the sample spec, this command may fail DNS or TLS in the real world—that is fine for learning the **CLI shape**. For a request that hits the network successfully, point `--spec` at an API you control or at a public sandbox whose base URL resolves.

### Create a resource (POST with JSON body)

Create expects a body mapped from the OpenAPI schema. The Petstore `create` operation accepts a `Pet` object; the generated CLI exposes flags derived from that schema:

```bash
./petstore pets create --id 42 --name "Rex" --tag "demo"
```

Adjust flags to match your spec’s `requestBody` fields; `./petstore pets create --help` lists exactly what this build accepts.

## Exploring the Generated CLI

### Subcommands and groups

OpenAPI **tags** become **command groups**, and each **operation** becomes a leaf command. Explore the tree without memorizing it:

```bash
./petstore --help
./petstore pets --help
./petstore pets list --help
```

### Flags instead of handwritten URLs

Path parameters (for example `petId`) appear as required flags:

```bash
./petstore pets show-pet-by-id --pet-id 1 --help
```

Query parameters show up as optional flags (here, `limit` and `status` on `pets list`). Enum values from the spec often feed shell completion when you install completions for your shell.

### Output and debugging

By default you get **compact JSON** on stdout. For human-readable JSON during exploration:

```bash
./petstore pets list --status available --format pretty
```

Many generated projects also support `--verbose` on the root command for request tracing—check `./petstore --help` on your binary.

## Next Steps: Point at Your Own API

Once the Petstore flow makes sense, swap the spec for **your** OpenAPI document:

```bash
onlycli generate \
  --spec ./openapi.yaml \
  --name myplatform \
  --auth bearer \
  --out ./myplatform-cli

cd myplatform-cli && go mod tidy && go build -o myplatform .
```

Use `--auth bearer`, `--auth basic`, or `--auth apikey` when you want to force a style; otherwise OnlyCLI infers what it can from `securitySchemes`. Generated CLIs typically read tokens from an **environment variable** named after your CLI (for example `MYPLATFORM_TOKEN`)—see the generator summary printed at the end of `onlycli generate` and the `README` in the output folder.

### CI: regenerate when the spec changes

In GitHub Actions, a minimal pattern is: checkout → setup Go → `go install` OnlyCLI → `onlycli generate` → `go test` or `go build`. That keeps your distributed binary aligned with the same OpenAPI your documentation and server already use.

## Recap

You installed **OnlyCLI**, generated a **Petstore** CLI, built a **native binary**, and learned how **tags**, **operations**, and **parameters** map to **groups**, **commands**, and **flags**. From here, every endpoint in your OpenAPI file becomes a discoverable shell command—no hand-maintained curl scripts required.

---

*Next: learn output modes in [Mastering OnlyCLI Output]({{ '/blog/advanced-output-formatting/' | relative_url }}) or compare HTTP tools in [OnlyCLI vs curl vs HTTPie]({{ '/blog/onlycli-vs-curl-httpie/' | relative_url }}).*
