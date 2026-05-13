.PHONY: build test test-unit test-integration test-conformance lint run-local clean

BIN := bin/sql-mcp

build:
	@mkdir -p bin
	go build -o $(BIN) ./cmd/sql-mcp

test: test-unit

test-unit:
	go test ./internal/...

test-integration:
	go test -tags=integration ./tests/integration/...

test-conformance:
	go test -tags=integration ./tests/conformance/...

lint:
	golangci-lint run ./...

run-local:
	docker compose -f deploy/dev/compose.yaml up --build

clean:
	rm -rf bin/ out/ coverage.out
