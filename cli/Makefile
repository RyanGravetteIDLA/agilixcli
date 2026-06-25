.PHONY: build test lint install clean

build:
	go build -o bin/agilix-buzz-pp-cli ./cmd/agilix-buzz-pp-cli

test:
	go test ./...

lint:
	golangci-lint run

install:
	go install ./cmd/agilix-buzz-pp-cli

clean:
	rm -rf bin/

build-mcp:
	go build -o bin/agilix-buzz-pp-mcp ./cmd/agilix-buzz-pp-mcp

install-mcp:
	go install ./cmd/agilix-buzz-pp-mcp

build-all: build build-mcp
