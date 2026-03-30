# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.3.0](https://github.com/OnlyCLI/OnlyCLI/compare/v0.2.0...v0.3.0) (2026-03-30)


### Features

* improve docs SEO and GEO discoverability ([d646a7c](https://github.com/OnlyCLI/OnlyCLI/commit/d646a7c1d10362ea9102309bcf580d1595908f3b))
* marketing overhaul, token cost content, and skillmd command ([2b48fe6](https://github.com/OnlyCLI/OnlyCLI/commit/2b48fe6bee51389697ebb6ded293e1c873850893))
* marketing overhaul, token cost content, and skillmd command ([c457615](https://github.com/OnlyCLI/OnlyCLI/commit/c4576159535e9c7d08def0ab2ba2d81ef1332cf8))


### Bug Fixes

* strengthen docs internal linking ([0a3dd2e](https://github.com/OnlyCLI/OnlyCLI/commit/0a3dd2e948e183cea072029d787b8a4e65f487ef))

## [Unreleased]

## [0.1.0] - 2026-03-26

### Added
- Core CLI tool: `onlycli generate` command to generate Go CLI projects from OpenAPI specs
- OpenAPI 3.0/3.1 parsing via libopenapi
- Code generation using Go text/template with go/format formatting
- Cobra-based generated CLIs with subcommand grouping by OpenAPI tags
- LLM agent-friendly CLI design with stable `--help` and completion support
- Support for path, query, and header parameters as CLI flags
- Request body support via --data flag (supports stdin with -)
- Authentication support: bearer, basic, apikey (auto-detected from spec)
- Cross-platform generated code (Linux, macOS, Windows)
- Shell completion in generated CLIs (via Cobra built-in)
- Version command with build info injection via ldflags
- CI/CD with GitHub Actions (test, lint, multi-OS build, release)
- Cross-platform binary releases via GoReleaser
- Comprehensive test suite: unit tests per package, integration tests with GitHub API spec
- Docker support for containerized usage
- curl-based install script for easy installation
