# OnlyCLI

[![CI](https://github.com/onlycli/onlycli/actions/workflows/ci.yml/badge.svg)](https://github.com/onlycli/onlycli/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/onlycli/onlycli)](https://goreportcard.com/report/github.com/onlycli/onlycli)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://pkg.go.dev/badge/github.com/onlycli/onlycli.svg)](https://pkg.go.dev/github.com/onlycli/onlycli)

Turn any OpenAPI spec into a native CLI binary. No MCP, no bloat, no runtime dependencies.

MCP injects all tool schemas into every LLM conversation, burning 4-32x more tokens and suffering 28% connection failures ([source](https://scalekit.com)). OnlyCLI takes a different approach: generate a static, distributable CLI binary with stable `--help` and completion that LLM agents can learn from; you can optionally maintain a `SKILL.md` for agent-oriented discovery.

## Quick Start

### Install

```bash
# Go install
go install github.com/onlycli/onlycli/cmd/onlycli@latest

# Or use the install script
curl -sSfL https://raw.githubusercontent.com/onlycli/onlycli/main/install.sh | sh

# Or with Docker
docker run --rm -v $(pwd):/work ghcr.io/onlycli/onlycli generate --spec /work/api.yaml --name myapi --out /work/myapi-cli
```

### Generate a CLI

```bash
onlycli generate \
  --spec https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml \
  --name github \
  --auth bearer \
  --out ./github-cli

cd github-cli && go mod tidy && go build -o github .
```

### Use the Generated CLI

```bash
export GITHUB_TOKEN=ghp_xxxxx

./github repos get --owner microsoft --repo vscode
./github issues list-for-repo --owner microsoft --repo vscode --state open --per-page 5
./github pulls create --owner user --repo myrepo --data '{"title":"fix","head":"feat","base":"main"}'
./github search repos --q "language:go stars:>1000" --sort stars
```

All output is plain JSON. Pipe to `jq` for filtering:

```bash
./github repos list-for-user --username octocat | jq '.[].full_name'
```

## How It Works

```
OpenAPI Spec (YAML/JSON)
     |
     v
 [libopenapi]  -- Parse into typed model
     |
     v
 [IR Model]    -- APISpec, CommandGroups, Commands, Parameters
     |
     +---> [codegen]   -- Go templates --> .go source files
               |
               +--> main.go
               +--> commands/ (root, groups, operations, register)
               +--> runtime/ (HTTP client, output formatting)
               +--> go.mod
```

Each OpenAPI operation becomes a Cobra CLI command, grouped by tag. The mapping:

| OpenAPI Element | CLI Mapping |
|---|---|
| `tags` | Subcommand groups (`github repos`, `github issues`) |
| `operationId` | Command name, kebab-cased with group prefix stripped |
| Path parameters | Required `--flags` |
| Query parameters | Optional `--flags` |
| `requestBody` | `--data` flag (supports stdin via `--data -`) |
| `description` | `--help` text |
| `responses` | Plain JSON output |

## Flags

| Flag | Required | Description |
|---|---|---|
| `--spec` | Yes | Path or URL to OpenAPI spec (YAML/JSON) |
| `--name` | Yes | Name of the generated CLI binary |
| `--out` | Yes | Output directory for the generated project |
| `--auth` | No | Auth type: `bearer`, `basic`, `apikey` (auto-detected from spec) |
| `--module` | No | Go module path (default: `github.com/<name>/<name>-cli`) |

## Generated CLI Features

- Subcommand tree grouped by OpenAPI tags
- All parameters as self-documenting `--flags`
- Plain JSON output by default, `--pretty` for formatted
- Authentication via environment variable
- Shell completion (`<cli> completion bash/zsh/fish/powershell`)
- Cross-platform (Linux, macOS, Windows; amd64, arm64)
- LLM-friendly CLI surface: predictable `--help`, flags, and shell completion (optional hand-maintained `SKILL.md` for agent docs)

## Examples

See the [examples/](examples/) directory for pre-generated CLI projects:

- **[petstore](examples/petstore/)** - Classic Petstore API (2 groups, 6 commands)
- **[github](examples/github/)** - GitHub REST API subset (7 groups, 20 commands)

## CI Integration

Rebuild the CLI when your API spec changes:

```yaml
# .github/workflows/build-cli.yml
on:
  push:
    paths: ['openapi.yaml']
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.23' }
      - run: go install github.com/onlycli/onlycli/cmd/onlycli@latest
      - run: onlycli generate --spec openapi.yaml --name myapi --out ./cli
      - run: cd cli && go mod tidy && go build -o myapi .
```

## Development

```bash
git clone https://github.com/onlycli/onlycli.git
cd onlycli
make test          # Run all tests
make lint          # Run linter
make check         # Full pre-commit check (fmt + vet + lint + test)
make build         # Build binary to bin/onlycli
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for the full contributing guide and [ARCHITECTURE.md](ARCHITECTURE.md) for codebase structure.

## License

[MIT](LICENSE)
