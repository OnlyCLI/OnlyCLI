FROM golang:1.23-alpine AS builder

ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN mkdir -p /src/bin && CGO_ENABLED=0 go build \
	-ldflags "-X main.version=${VERSION} -X main.commit=${COMMIT} -X main.date=${DATE}" \
	-o /src/bin/onlycli \
	./cmd/onlycli

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

COPY --from=builder /src/bin/onlycli /usr/local/bin/onlycli

ENTRYPOINT ["onlycli"]
