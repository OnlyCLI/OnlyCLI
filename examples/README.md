# Examples

This directory holds sample CLI projects produced by [onlycli](../README.md) from live OpenAPI specs. Use them to see how generated code is laid out and to try the tool end to end.

## Petstore

Generated from the official Swagger Petstore spec:

```bash
./bin/onlycli generate \
  --spec https://raw.githubusercontent.com/swagger-api/swagger-petstore/master/src/main/resources/openapi.yaml \
  --name petstore \
  --auth apikey \
  --out examples/petstore
```

Build and run (from repository root):

```bash
cd examples/petstore
go mod tidy
go build -o petstore .
./petstore --help
```

Auth uses an API key; set `PETSTORE_TOKEN` when calling protected operations.

## GitHub

Generated from the official GitHub REST API spec (1100+ commands):

```bash
./bin/onlycli generate \
  --spec https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml \
  --name github \
  --auth bearer \
  --out examples/github
```

Build and run (from repository root):

```bash
cd examples/github
go mod tidy
go build -o github-cli .
./github-cli --help
```

Auth uses a bearer token; set `GITHUB_TOKEN` when calling the API.

## Regenerating

These directories contain generated code. Re-generate them from the repository root with:

```bash
make generate-example
```

That builds onlycli and fetches the latest specs online to regenerate both examples.
