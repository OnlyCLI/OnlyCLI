VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
BINARY_NAME := onlycli

LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)

.PHONY: build test test-integration lint fmt vet install clean generate-example check

build:
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY_NAME) ./cmd/onlycli

test:
	go test ./... -race -coverprofile=coverage.out

test-integration:
	go test ./... -race -run TestIntegration -timeout 300s

lint:
	golangci-lint run ./...

fmt:
	gofmt -w .
	go mod tidy

vet:
	go vet ./...

install:
	go install ./cmd/onlycli/

clean:
	rm -rf bin/ dist/ coverage.*

generate-example: build
	rm -rf examples/petstore examples/github
	./bin/$(BINARY_NAME) generate \
		--spec https://raw.githubusercontent.com/swagger-api/swagger-petstore/master/src/main/resources/openapi.yaml \
		--name petstore \
		--auth apikey \
		--out examples/petstore
	./bin/$(BINARY_NAME) generate \
		--spec https://raw.githubusercontent.com/github/rest-api-description/main/descriptions/api.github.com/api.github.com.yaml \
		--name github \
		--auth bearer \
		--out examples/github

check: fmt vet lint test
