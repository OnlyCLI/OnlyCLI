# OnlyCLI architecture

## High-level overview

OnlyCLI ingests an OpenAPI 3.x document (YAML or JSON), builds an internal **APISpec** intermediate representation (IR) that groups operations into commands, then renders that IR through Go templates into a standalone Go module: a Cobra-based CLI, registration wiring, and a small HTTP runtime client. Teams can add an optional, hand-maintained **SKILL.md** beside the generated CLI for LLM-oriented documentation. The onlycli binary itself is a thin Cobra wrapper around parse → IR → codegen steps.

## Data flow

```
OpenAPI Spec (YAML/JSON)
     |
     v
[parser] libopenapi -> high-level model
     |
     v
[model] APISpec IR (groups, commands, parameters)
     |
     +--> [codegen] Go templates -> .go source files
               |
               +--> main.go, commands/*.go, runtime/*.go, go.mod
```

## Packages

### `internal/model`

Defines IR types such as **APISpec**, **CommandGroup**, **Command**, and **Parameter**, plus naming utilities used by both parser and codegen (for example **ToKebabCase**, **StripGroupPrefix**, and helpers that produce stable Go identifiers and file names from OpenAPI names and tags).

### `internal/parser`

Wraps **libopenapi**, walks paths and operations, and populates the IR. Responsibilities include deriving command names from **operationId** with sensible fallbacks, grouping operations by **tags**, detecting security schemes for auth hints, and normalizing parameters (path, query, header, body) into the IR.

### `internal/codegen`

Template engine for the generated project. Templates are embedded with **`go:embed`**. Each template executes with typed or map data derived from the IR; emitted Go files pass through **`go/format`** before write. Outputs include **main.go**, **commands/root.go** (root command), one `commands/{group}.go` per tag group, `commands/{group}_{operation}.go` per operation, **commands/register.go** (command tree wiring), **runtime/client.go** (HTTP client and auth), and **runtime/output.go** (response printing helpers), plus **go.mod** from a dedicated template.

### `cmd/onlycli`

The **onlycli** tool: Cobra root with **`generate`** (read spec from file or URL, write output tree) and **`version`**. Spec bytes are loaded from a local path or HTTP(S) URL, then passed into the parser and downstream generators.

## OpenAPI to CLI mapping

| OpenAPI concept | Generated CLI behavior |
|-----------------|-------------------------|
| Operation | One Cobra **command** under a group (subcommand of root). |
| Tags (first tag) | **Command group** (parent command); operations sharing a tag live under the same group. |
| Path parameters | **Flags** (for example `--id`); substituted into the URL when invoking the HTTP client. |
| Query parameters | **Flags**; appended to the request URL. |
| Request body | **`--data`** (and related body handling) when the operation defines a body. |
| Summary / description | **Command short/long help** (`--help` text on the generated command). |

## Generated project structure

After `onlycli generate`, a typical layout looks like this:

```
<out>/
├── main.go                 # delegates to commands.Execute()
├── go.mod                  # module path from --module or default
├── commands/
│   ├── root.go             # root Cobra command
│   ├── register.go         # attaches groups and leaf commands to root
│   ├── <group>.go          # one file per tag / group
│   └── <group>_<op>.go     # one file per operation (leaf commands)
└── runtime/
    ├── client.go           # base URL, auth, DoRequest helpers
    └── output.go           # print / format responses
```

File names for groups and operations follow kebab-style or derived names from the IR naming helpers so they stay valid Go file names and identifiers after transformation.

## Design decisions

### Flags instead of positional arguments for path parameters

Path (and query) inputs are exposed as **named flags** rather than positional args so invocations are **self-documenting** and easier for humans and **LLMs** to construct correctly: every value is labeled, order is not ambiguous, and `--help` lists all inputs explicitly.

### Static code generation instead of a runtime-only CLI

Emitting a **real Go project** makes the result **distributable** as a normal binary, **predictable** in behavior and dependencies, and free of a runtime "interpret the spec on every invocation" layer. Users get idiomatic Cobra code they can inspect, fork, and ship.

### `go/format` instead of `goimports` for generated Go

Generated files are formatted with **`go/format`** because the new module often **does not exist on disk yet** in a form `goimports` can resolve. Import paths are controlled in templates; `format.Source` guarantees valid Go layout without requiring a full module graph for external tooling.

### Conditional imports in templates

Templates use **conditional import blocks** (or minimal import sets) so generated files only import packages that are **actually used** for that operation (for example `fmt` / `strings` / `io` / `os` only when body or path substitution needs them). That avoids **unused import** compile errors across the wide variety of OpenAPI shapes.
