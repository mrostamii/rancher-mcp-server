# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /build

RUN apk add --no-cache ca-certificates

COPY go.mod go.sum ./
RUN go mod download

ARG VERSION=dev
COPY . .
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X github.com/mrostamii/rancher-mcp-server/internal/cmd.version=${VERSION}" -o rancher-mcp-server ./cmd/rancher-mcp-server

# Runtime stage
FROM alpine:3.19
RUN apk add --no-cache ca-certificates
WORKDIR /app

COPY --from=builder /build/rancher-mcp-server .

ENTRYPOINT ["/app/rancher-mcp-server"]
