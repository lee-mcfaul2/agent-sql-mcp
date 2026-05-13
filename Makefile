.PHONY: build embed-schemas test test-unit test-integration test-conformance lint run-local clean

BIN := bin/sql-mcp

embed-schemas:
	mkdir -p cmd/sql-mcp/embedded
	cp schemas/*.json cmd/sql-mcp/embedded/

build: embed-schemas
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
