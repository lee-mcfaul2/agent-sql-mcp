# syntax=docker/dockerfile:1.7

# ---------- Stage 1: build ----------
FROM golang:1.25-alpine AS builder
WORKDIR /src
RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN mkdir -p cmd/sql-mcp/embedded && cp schemas/*.json cmd/sql-mcp/embedded/
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -trimpath -ldflags='-s -w' -o /out/sql-mcp ./cmd/sql-mcp

# ---------- Stage 2: runtime ----------
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /out/sql-mcp /usr/local/bin/sql-mcp
EXPOSE 8443
USER nonroot:nonroot
ENTRYPOINT ["/usr/local/bin/sql-mcp"]
